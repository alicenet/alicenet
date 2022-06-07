package transaction

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/blockchain/transaction/interfaces"
	"github.com/MadBase/MadNet/blockchain/transaction/objects"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

//
var (
	ErrUnknownRequest = errors.New("unknown request type")
)

type ErrRecoverable struct {
	message string
}

func (e *ErrRecoverable) Error() string {
	return e.message
}

type ErrTransactionStale struct {
	message string
}

func (e *ErrTransactionStale) Error() string {
	return e.message
}

type ErrNonRecoverable struct {
	message string
}

func (e *ErrNonRecoverable) Error() string {
	return e.message
}

type ErrInvalidMonitorRequest struct {
	message string
}

func (e *ErrInvalidMonitorRequest) Error() string {
	return e.message
}

type ErrTransactionAlreadyQueued struct {
	message string
}

func (e *ErrTransactionAlreadyQueued) Error() string {
	return e.message
}

type ErrInvalidTransactionRequest struct {
	message string
}

func (e *ErrInvalidTransactionRequest) Error() string {
	return e.message
}

// Internal struct to keep track of transactions that are being monitoring
type info struct {
	ctx                     context.Context                           // ctx used for calling the monitoring a certain tx
	txn                     *types.Transaction                        // Transaction object
	selector                objects.FuncSelector                      // 4 bytes that identify the function being called by the tx
	functionSignature       string                                    // function signature as we see on the smart contracts
	startedMonitoringHeight uint64                                    // ethereum height where we first added the tx to be watched. Mainly used to see if a tx was dropped
	receiptResponseChannel  *ResponseChannel[objects.ReceiptResponse] // channel where the response will be sent
}

// Internal struct to keep track of the receipts
type receipt struct {
	receipt           *types.Receipt // receipt object
	retrievedAtHeight uint64         // block where receipt was added to the cache
}

// Internal struct to keep track of what blocks we already checked during monitoring
type block struct {
	height uint64      // block height
	hash   common.Hash // block header hash
}

// Compare if 2 blockInfo structs are equal by comparing the height and block
// hash. Return true in case they are equal, false otherwise.
func (a *block) Equal(b *block) bool {
	return bytes.Equal(a.hash[:], b.hash[:]) && a.height == b.height
}

// Type to do request against the tx receipt monitoring system. Ctx and response channel should be set
type subscribeRequest struct {
	ctx             context.Context                     // tx ctx used for tx monitoring cancellation
	txn             *types.Transaction                  // the transaction that should watched
	responseChannel *ResponseChannel[SubscribeResponse] // channel where we going to send the request response
}

// Constrain interface used by the Response channel generics
type transferable interface {
	SubscribeResponse | objects.ReceiptResponse
}

// Type that it's going to be used to reply a request
type SubscribeResponse struct {
	TxnHash                common.Hash                               // Hash of the txs which this response belongs
	Err                    error                                     // errors that happened when processing the monitor request
	ReceiptResponseChannel *ResponseChannel[objects.ReceiptResponse] // non blocking channel where the result from the tx/receipt monitoring will be send
}

// A response channel is basically a non-blocking channel that can only be
// written and closed once. The internal channel is closed after the first
// message is sent. Additional tries to send a message result in no-op. The
// writes to the internal channel are non-blocking calls. If for some reason the
// internal channel is full, the message is dropped and log is recorded. Only
// first attempt to close the Response channel will result in the closing.
// Additional calls are no-op.
type ResponseChannel[T transferable] struct {
	writeOnce sync.Once
	channel   chan *T       // internal channel
	isClosed  bool          // flag to check if a channel is closed or not
	logger    *logrus.Entry // logger using for logging error when trying to write a response more than once
}

// Create a new response channel.
func NewResponseChannel[T transferable](logger *logrus.Entry) *ResponseChannel[T] {
	return &ResponseChannel[T]{channel: make(chan *T, 1), logger: logger}
}

// send a unique response and close the internal channel. Additional calls to
// this function will be no-op
func (rc *ResponseChannel[T]) SendResponse(response *T) {
	if !rc.isClosed {
		select {
		case rc.channel <- response:
		default:
			rc.logger.Debugf("Failed to write request to channel")
		}
		rc.CloseChannel()
	}
}

// Close the internal channel. Additional calls will be no-op
func (rc *ResponseChannel[T]) CloseChannel() {
	rc.writeOnce.Do(func() {
		rc.isClosed = true
		close(rc.channel)
	})
}

// Profile to keep track of gas metrics in the overall system
type Profile struct {
	AverageGas   uint64
	MinimumGas   uint64
	MaximumGas   uint64
	TotalCount   uint64
	TotalGas     uint64
	TotalSuccess uint64
}

// Internal struct used to send work requests to the workers that will retrieve
// the receipts
type MonitorWorkRequest struct {
	txn    info   // TransactionInfo object that contains the state that will be used to retrieve the receipt from the blockchain
	height uint64 // Current height of the blockchain head
}

// Internal struct used by the workers to communicate the result from the receipt retrieval work
type MonitorWorkResponse struct {
	txnHash common.Hash    // hash of transaction
	err     error          // any error found during the receipt retrieve (can be NonRecoverable, Recoverable or TransactionState errors)
	receipt *types.Receipt // receipt retrieved (can be nil) if a receipt was not found or it's not ready yet
}

// Backend struct used to monitor Ethereum transactions and retrieve their receipts
type WatcherBackend struct {
	mainCtx            context.Context                  // main context for the background services
	lastProcessedBlock *block                           // Last ethereum block that we checked for receipts
	monitoredTxns      map[common.Hash]info             // Map of transactions whose receipts we're looking for
	receiptCache       map[common.Hash]receipt          // Receipts retrieved from transactions
	aggregates         map[objects.FuncSelector]Profile // Struct to keep track of the gas metrics used by the system
	client             ethereum.Network                 // An interface with the ethereum functionality we need
	knownSelectors     interfaces.ISelectorMap          // Map with signature -> name
	logger             *logrus.Entry                    // Logger to log messages
	requestChannel     <-chan subscribeRequest          // Channel used to send request to this backend service
}

func (b *WatcherBackend) Loop() {

	poolingTime := time.After(constants.TxPollingTime)
	for {
		select {
		case req, ok := <-b.requestChannel:
			if !ok {
				b.logger.Debugf("request channel closed, exiting")
				return
			}
			if req.responseChannel == nil {
				b.logger.Debug("Invalid request for txn without a response channel, ignoring")
				continue
			}
			b.queue(req)

		case <-poolingTime:
			b.collectReceipts()
			poolingTime = time.After(constants.TxPollingTime)
		}
	}
}

func (b *WatcherBackend) queue(req subscribeRequest) {

	if req.txn == nil {
		req.responseChannel.SendResponse(&SubscribeResponse{Err: &ErrInvalidMonitorRequest{"invalid request, missing txn object"}})
		return
	}
	if req.ctx == nil {
		req.responseChannel.SendResponse(&SubscribeResponse{Err: &ErrInvalidMonitorRequest{"invalid request, missing ctx"}})
		return
	}

	txnHash := req.txn.Hash()
	receiptResponseChannel := NewResponseChannel[objects.ReceiptResponse](b.logger)

	// if we already have the records of the receipt for this tx we try to send the
	// receipt back
	if receipt, ok := b.receiptCache[txnHash]; ok {
		receiptResponseChannel.SendResponse(&objects.ReceiptResponse{Receipt: receipt.receipt, TxnHash: txnHash})
	} else {
		if _, ok := b.monitoredTxns[txnHash]; ok {
			req.responseChannel.SendResponse(&SubscribeResponse{Err: &ErrTransactionAlreadyQueued{"invalid request, tx already queued, try to get receipt later!"}})
			return
		}

		selector := ExtractSelector(req.txn.Data())
		//todo: replace this with a generated mapping from the bindings
		sig := b.knownSelectors.Signature(selector)

		logEntry := b.logger.WithField("Transaction", txnHash).
			WithField("Function", sig).
			WithField("Selector", fmt.Sprintf("%x", selector))

		b.monitoredTxns[txnHash] = info{
			ctx:                    req.ctx,
			txn:                    req.txn,
			selector:               selector,
			functionSignature:      sig,
			receiptResponseChannel: receiptResponseChannel,
		}
		logEntry.Debug("Transaction queued")
	}
	req.responseChannel.SendResponse(&SubscribeResponse{ReceiptResponseChannel: receiptResponseChannel})
}

func (b *WatcherBackend) collectReceipts() {

	lenMonitoredTxns := len(b.monitoredTxns)

	// If there's no tx to be monitored just return
	if lenMonitoredTxns == 0 {
		b.logger.Tracef("no transaction to watch")
		return
	}

	networkCtx, cf := context.WithTimeout(b.mainCtx, constants.TxNetworkTimeout)
	defer cf()

	blockHeader, err := b.client.GetHeaderByNumber(networkCtx, nil)
	if err != nil {
		b.logger.Debugf("error getting latest block number from ethereum node: %v", err)
		return
	}
	blockInfo := &block{
		blockHeader.Number.Uint64(),
		blockHeader.Hash(),
	}

	if b.lastProcessedBlock.Equal(blockInfo) {
		b.logger.Tracef("already processed block: %v with hash: %v", blockInfo.height, blockInfo.hash.Hex())
		return
	}
	b.logger.Tracef("processing block: %v with hash: %v", blockInfo.height, blockInfo.hash.Hex())

	txnOpts, err := b.client.GetTransactionOpts(networkCtx, b.client.GetDefaultAccount())
	if err != nil {
		b.logger.Debugf("error getting txnOpts from ethereum node: %v", err)
		return
	}

	var expiredTxs []common.Hash
	finishedTxs := make(map[common.Hash]MonitorWorkResponse)

	numWorkers := utils.Min(utils.Max(uint64(lenMonitoredTxns)/4, 128), 1)
	requestWorkChannel := make(chan MonitorWorkRequest, lenMonitoredTxns+3)
	responseWorkChannel := make(chan MonitorWorkResponse, lenMonitoredTxns+3)

	for txn, txnInfo := range b.monitoredTxns {
		select {
		case <-txnInfo.ctx.Done():
			// the go-routine who wanted this information has stopped caring. This most
			// likely indicates a failure, and cancellation of polling prevents a memory
			// leak
			b.logger.Debugf("context for tx %v is finished, marking it to be excluded!", txn.Hex())
			expiredTxs = append(expiredTxs, txnInfo.txn.Hash())
		default:
			// if this is the first time seeing a tx or we have a reorg and
			// startedMonitoring is now greater than the current ethereum block height
			if txnInfo.startedMonitoringHeight == 0 || txnInfo.startedMonitoringHeight > blockInfo.height {
				txnInfo.startedMonitoringHeight = blockInfo.height
				b.monitoredTxns[txn] = txnInfo
			}
			requestWorkChannel <- MonitorWorkRequest{txnInfo, blockInfo.height}
		}
	}

	// close the request channel, so the workers know when to finish
	close(requestWorkChannel)

	workerPool := NewWorkerPool(b.mainCtx, b.client, txnOpts, b.logger, requestWorkChannel, responseWorkChannel)

	// spawn the workers and wait for all to complete
	go workerPool.ExecuteWork(numWorkers)

	for workResponse := range responseWorkChannel {
		select {
		case <-b.mainCtx.Done():
			// main thread was killed
			return
		default:
			logEntry := b.logger.WithFields(logrus.Fields{"txn": workResponse.txnHash})
			if workResponse.err != nil {
				if _, ok := workResponse.err.(*ErrRecoverable); !ok {
					logEntry.Debugf("Couldn't get tx receipt, err: %v", workResponse.err)
					finishedTxs[workResponse.txnHash] = workResponse
				} else {
					logEntry.Tracef("Retrying! Got a recoverable error when trying to get receipt, err: %v", workResponse.err)
				}
			} else {
				if workResponse.receipt != nil {
					logEntry.WithFields(
						logrus.Fields{
							"mined":          workResponse.receipt.BlockNumber,
							"current height": blockInfo.height,
						},
					).Debug("Successfully got receipt")
					b.receiptCache[workResponse.txnHash] = receipt{receipt: workResponse.receipt, retrievedAtHeight: blockInfo.height}
					finishedTxs[workResponse.txnHash] = workResponse
				}
			}
		}
	}

	// Cleaning finished and failed transactions
	for txnHash, workResponse := range finishedTxs {
		if txnInfo, ok := b.monitoredTxns[txnHash]; ok {
			txnInfo.receiptResponseChannel.SendResponse(&objects.ReceiptResponse{TxnHash: workResponse.txnHash, Receipt: workResponse.receipt, Err: workResponse.err})
			delete(b.monitoredTxns, txnHash)
		}
	}

	// Cleaning expired transactions
	for _, txnHash := range expiredTxs {
		if txnInfo, ok := b.monitoredTxns[txnHash]; ok {
			txnInfo.receiptResponseChannel.CloseChannel()
			delete(b.monitoredTxns, txnHash)
		}
	}

	var expiredReceipts []common.Hash
	// Marking expired receipts
	for receiptTxnHash, receiptInfo := range b.receiptCache {
		if blockInfo.height >= receiptInfo.retrievedAtHeight+constants.TxReceiptCacheMaxBlocks {
			expiredReceipts = append(expiredReceipts, receiptTxnHash)
		}
	}

	// being paranoic and excluding the expired receipts in another loop
	for _, receiptTxHash := range expiredReceipts {
		delete(b.receiptCache, receiptTxHash)
	}

	b.lastProcessedBlock = blockInfo
}

// Structs that keep track of the state needed by the worker pool service. The
// workerPool spawn multiple go routines (workers) to check and retrieve the
// receipts.
type WorkerPool struct {
	wg                  *sync.WaitGroup
	ctx                 context.Context            // Main context passed by the parent, used to cancel worker and the pool
	client              ethereum.Network           // An interface with the Geth functionality we need
	txnOpts             *bind.TransactOpts         // Transaction options in case we need to retry a stale transaction
	logger              *logrus.Entry              // Logger to log messages
	requestWorkChannel  <-chan MonitorWorkRequest  // Channel where will be send the work requests
	responseWorkChannel chan<- MonitorWorkResponse // Channel where the work response will be send
}

// Creates a new WorkerPool service
func NewWorkerPool(ctx context.Context, client ethereum.Network, txnOpts *bind.TransactOpts, logger *logrus.Entry, requestWorkChannel <-chan MonitorWorkRequest, responseWorkChannel chan<- MonitorWorkResponse) *WorkerPool {
	return &WorkerPool{new(sync.WaitGroup), ctx, client, txnOpts, logger, requestWorkChannel, responseWorkChannel}
}

// Function to spawn the workers and wait for the job to be done.
func (w *WorkerPool) ExecuteWork(numWorkers uint64) {
	for i := uint64(0); i < numWorkers; i++ {
		w.wg.Add(1)
		go w.worker()
	}
	w.wg.Wait()
	close(w.responseWorkChannel)
}

// Unit of work. A worker is spawned as go routine. A worker check and retrieve
// receipts for multiple transactions. The worker will be executing while
// there's transactions to be checked or there's a timeout (set by
// constants.TxWorkerTimeout)
func (w *WorkerPool) worker() {
	ctx, cf := context.WithTimeout(w.ctx, constants.TxWorkerTimeout)
	defer cf()
	defer w.wg.Done()
	// iterating over a closed channel
	for work := range w.requestWorkChannel {
		select {
		case <-ctx.Done():
			// worker context timed out or parent was cancelled, should return
			return
		default:
			monitoredTx := work.txn
			currentHeight := work.height
			txnHash := monitoredTx.txn.Hash()
		Loop:
			for i := uint64(1); i <= constants.TxWorkerMaxWorkRetries; i++ {
				select {
				case <-monitoredTx.ctx.Done():
					// the go-routine who wanted this information has stopped caring. This most
					// likely indicates a failure, and cancellation of polling prevents a memory
					// leak
					w.responseWorkChannel <- MonitorWorkResponse{txnHash: txnHash, err: &ErrNonRecoverable{fmt.Sprintf("context for tx %v is finished!", txnHash.Hex())}}
					//should continue getting other tx work
					break
				case <-ctx.Done():
					// worker context timed out or parent was cancelled, should return
					return
				default:
					rcpt, err := w.getReceipt(ctx, monitoredTx, currentHeight, txnHash)
					if err != nil {
						// retry on recoverable error `constants.TxWorkerMaxWorkRetries` times
						if _, ok := err.(*ErrRecoverable); ok && i < constants.TxWorkerMaxWorkRetries {
							continue Loop
						}
						// send nonRecoverable errors back to main or recoverable errors after constants.TxWorkerMaxWorkRetries
						w.responseWorkChannel <- MonitorWorkResponse{txnHash: txnHash, err: err}
					} else {
						// send receipt (even if it nil) back to main thread
						w.responseWorkChannel <- MonitorWorkResponse{txnHash: txnHash, receipt: rcpt}
					}
					//should continue getting other tx work
					break Loop
				}
			}
		}
	}
}

// Internal function used by the workers to check/retrieve the receipts for a given transaction
func (w *WorkerPool) getReceipt(ctx context.Context, monitoredTx info, currentHeight uint64, txnHash common.Hash) (*types.Receipt, error) {
	txnHex := txnHash.Hex()
	blockTimeSpan := currentHeight - monitoredTx.startedMonitoringHeight
	_, isPending, err := w.client.GetTransactionByHash(ctx, txnHash)
	if err != nil {
		// if we couldn't locate a tx after NotFoundMaxBlocks blocks and we are still
		// failing in getting the tx data, probably means that it was dropped
		if blockTimeSpan >= w.client.GetTxNotFoundMaxBlocks() {
			return nil, &ErrNonRecoverable{fmt.Sprintf("could not find tx %v and %v blocks have passed!", txnHex, w.client.GetTxNotFoundMaxBlocks())}
		}
		// probably a network error, should retry
		return nil, &ErrRecoverable{fmt.Sprintf("error getting tx: %v: %v", txnHex, err)}
	}
	if isPending {
		if blockTimeSpan >= w.client.GetTxMaxStaleBlocks() {
			return nil, &ErrTransactionStale{fmt.Sprintf("error tx: %v is stale on the memory pool for more than %v blocks, please retry!", txnHex, w.client.GetTxMaxStaleBlocks())}
		}
	} else {
		// tx is not pending, so check for receipt
		rcpt, err := w.client.GetTransactionReceipt(ctx, txnHash)
		if err != nil {
			// if we couldn't locate a tx receipt after NotFoundMaxBlocks blocks and we are still
			// failing in getting the tx data, probably means that it was dropped
			if blockTimeSpan >= w.client.GetTxNotFoundMaxBlocks() {
				return nil, &ErrNonRecoverable{fmt.Sprintf("could not find receipt for tx %v and %v blocks have passed!", txnHex, w.client.GetTxNotFoundMaxBlocks())}
			}
			// 1. probably a network error, should retry
			// 2. in case receipt not found after tx not Pending check, we had an edge case,
			// probably tx was mined (isPending == false), then we had a chain re-org, now
			// the tx is back to the memPool or was dropped, we should retry, and the logic
			// above should see if the tx was dropped or not in the next iteration
			return nil, &ErrRecoverable{fmt.Sprintf("error getting receipt: %v: %v", txnHex, err)}
		}

		if currentHeight >= rcpt.BlockNumber.Uint64()+w.client.GetFinalityDelay() {
			return rcpt, nil
		}
	}
	return nil, nil
}

// Struct that has the data necessary by the Transaction Watcher service. The
// transaction watcher service is responsible for check, retrieve and cache
// transaction receipts.
type Watcher struct {
	backend          *WatcherBackend         // backend service responsible for check, retrieving and caching the receipts
	logger           *logrus.Entry           // logger used to log the message for the transaction watcher
	closeMainContext context.CancelFunc      // function used to cancel the main context in the backend service
	requestChannel   chan<- subscribeRequest // channel used to send request to the backend service to retrieve transactions
}

var _ interfaces.IWatcher = &Watcher{}

// Creates a new transaction watcher struct
func NewWatcher(client ethereum.Network, selectMap interfaces.ISelectorMap, txConfirmationBlocks uint64) *Watcher {
	requestChannel := make(chan subscribeRequest, 100)
	// main context that will cancel all workers and go routine
	mainCtx, cf := context.WithCancel(context.Background())

	logger := logging.GetLogger("transaction")

	backend := &WatcherBackend{
		mainCtx:            mainCtx,
		requestChannel:     requestChannel,
		client:             client,
		logger:             logger.WithField("Component", "TransactionWatcherBackend"),
		monitoredTxns:      make(map[common.Hash]info),
		receiptCache:       make(map[common.Hash]receipt),
		aggregates:         make(map[objects.FuncSelector]Profile),
		knownSelectors:     selectMap,
		lastProcessedBlock: &block{0, common.HexToHash("")},
	}

	transactionWatcher := &Watcher{
		requestChannel:   requestChannel,
		closeMainContext: cf,
		backend:          backend,
		logger:           logger.WithField("Component", "TransactionWatcher"),
	}
	return transactionWatcher
}

// WatcherFromNetwork creates a transaction Watcher from a given ethereum network.
func WatcherFromNetwork(network ethereum.Network) *Watcher {
	watcher := NewWatcher(network, NewKnownSelectors(), network.GetFinalityDelay())
	watcher.StartLoop()
	return watcher
}

// Start the transaction watcher service
func (f *Watcher) StartLoop() {
	go f.backend.Loop()
}

// Close the transaction watcher service
func (f *Watcher) Close() {
	f.logger.Debug("closing request channel...")
	close(f.requestChannel)
	f.closeMainContext()
}

// Subscribe a transaction to be watched by the transaction watcher service. If a
// transaction was accepted to be watched, a response channel is returned. The
// response channel is where the receipt going to be sent by the tx watcher
// backend.
func (tw *Watcher) Subscribe(ctx context.Context, txn *types.Transaction) (<-chan *objects.ReceiptResponse, error) {
	tw.logger.WithField("Txn", txn.Hash().Hex()).Debug("Subscribing for a transaction")
	respChannel := NewResponseChannel[SubscribeResponse](tw.logger)
	defer respChannel.CloseChannel()
	req := subscribeRequest{ctx: ctx, txn: txn, responseChannel: respChannel}

	select {
	case tw.requestChannel <- req:
	case <-ctx.Done():
		return nil, &ErrInvalidTransactionRequest{fmt.Sprintf("context cancelled reqChannel: %v", ctx.Err())}
	}

	select {
	case requestResponse := <-req.responseChannel.channel:
		return requestResponse.ReceiptResponseChannel.channel, requestResponse.Err
	case <-ctx.Done():
		return nil, &ErrInvalidTransactionRequest{fmt.Sprintf("context cancelled: %v", ctx.Err())}
	}
}

// function that wait for a transaction receipt. This is blocking function that will wait for a response in the input IResponse channel
func (f *Watcher) Wait(ctx context.Context, receiptResponseChannel <-chan *objects.ReceiptResponse) (*types.Receipt, error) {
	select {
	case receiptResponse := <-receiptResponseChannel:
		return receiptResponse.Receipt, receiptResponse.Err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Queue a transaction and wait for its receipt
func (f *Watcher) SubscribeAndWait(ctx context.Context, txn *types.Transaction) (*types.Receipt, error) {
	receiptResponseChannel, err := f.Subscribe(ctx, txn)
	if err != nil {
		return nil, err
	}
	return f.Wait(ctx, receiptResponseChannel)
}

func (f *Watcher) Status(ctx context.Context) error {
	f.logger.Error("Status function not implemented yet")
	return nil
}
