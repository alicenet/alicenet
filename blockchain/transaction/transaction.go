package transaction

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/blockchain/transaction/interfaces"
	"github.com/MadBase/MadNet/blockchain/transaction/objects"
	"github.com/MadBase/MadNet/bridge/mapping/signatures"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	goEthereum "github.com/ethereum/go-ethereum"
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

type ErrTxNotFound struct {
	message string
}

func (e *ErrTxNotFound) Error() string {
	return e.message
}

type ErrInvalidMonitorRequest struct {
	message string
}

func (e *ErrInvalidMonitorRequest) Error() string {
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
	// Transaction object
	txn *types.Transaction
	// address of the transaction signer
	fromAddress common.Address
	// 4 bytes that identify the function being called by the tx
	selector objects.FuncSelector
	// function signature as we see on the smart contracts
	functionSignature string
	// internal group Id to keep track of all tx that were created during the retry of a tx
	retryGroup common.Hash
	// whether we should disable the auto retry of a transaction
	disableAutoRetry bool
	// ethereum height where we first added the tx to be watched or did a tx retry.
	monitoringHeight uint64
	// counter to indicate how many times we tried to retry a transaction
	retryAmount uint64
	// counter to indicate approximate number of blocks that we could not find a tx
	notFoundBlocks uint64
	// logger to log transaction info
	logger *logrus.Entry
}

type ReceiptResponse struct {
	writeOnce sync.Once
	mu        sync.RWMutex
	isReady   bool
	err       error          // response error that happened during processing
	receipt   *types.Receipt // tx receipt after txConfirmationBlocks of a tx that was not queued in txGroup
}

func newReceiptResponse() *ReceiptResponse {
	return &ReceiptResponse{}
}

func (r *ReceiptResponse) IsReady() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.isReady
}

func (r *ReceiptResponse) GetReceipt() (*types.Receipt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.isReady {
		return r.receipt, r.err
	}
	return nil, nil
}

func (r *ReceiptResponse) writeReceipt(receipt *types.Receipt, err error) {
	if receipt == nil && err != nil {
		return
	}
	r.writeOnce.Do(func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.receipt = receipt
		r.err = err
		r.isReady = true
	})
}

type group struct {
	internalGroup   []common.Hash    //
	receiptResponse *ReceiptResponse // struct used to send/share the receipt
}

func newGroup() group {
	return group{receiptResponse: newReceiptResponse()}
}

func (g *group) add(txHash common.Hash) {
	g.internalGroup = append(g.internalGroup, txHash)
}

func (g *group) remove(txHash common.Hash) error {
	index := -1
	lastIndex := len(g.internalGroup) - 1
	if lastIndex == -1 {
		return fmt.Errorf("invalid removal, empty group", txHash.Hex())
	}
	for i, internalInfo := range g.internalGroup {
		if bytes.Equal(internalInfo.Bytes(), txHash.Bytes()) {
			index = i
		}
	}
	if index == -1 {
		return fmt.Errorf("txInfo %v not found", txHash.Hex())
	}
	if index != lastIndex {
		// copy the last element in the index that we want to delete
		g.internalGroup[index] = g.internalGroup[lastIndex]
	}
	// drop the last index
	g.internalGroup = g.internalGroup[:lastIndex]
	return nil
}

func (g *group) sendReceipt(logger, receipt *types.Receipt, err error) {

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
	txn              *types.Transaction                  // the transaction that should watched
	disableAutoRetry bool                                // whether we should disable the auto retry of a transaction
	responseChannel  *ResponseChannel[SubscribeResponse] // channel where we going to send the request response
}

// Constrain interface used by the Response channel generics
type transferable interface {
	SubscribeResponse
}

// Type that it's going to be used to reply a request
type SubscribeResponse struct {
	TxnHash  common.Hash      // Hash of the txs which this response belongs
	Err      error            // errors that happened when processing the monitor request
	Response *ReceiptResponse // channel where the result from the tx/receipt monitoring will be send
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

// Check if a channel is closed
func (rc *ResponseChannel[T]) IsChannelClosed() bool {
	return rc.isClosed
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
	txn    info   // Info object that contains the state that will be used to retrieve the receipt from the blockchain
	height uint64 // Current height of the blockchain head
}

// Internal struct used by the workers to communicate the result from the receipt retrieval work
type MonitorWorkResponse struct {
	txnHash    common.Hash         // hash of transaction
	retriedTxn *retriedTransaction // transaction info object from the analyzed transaction
	err        error               // any error found during the receipt retrieve (can be NonRecoverable, Recoverable or TransactionState errors)
	receipt    *types.Receipt      // receipt retrieved (can be nil) if a receipt was not found or it's not ready yet
}

// Internal struct to keep track of retried transaction by the workers
type retriedTransaction struct {
	txn *types.Transaction // new transaction after the retry attempt
	err error              // error that happened during the transaction retry
}

// Backend struct used to monitor Ethereum transactions and retrieve their receipts
type WatcherBackend struct {
	// main context for the background services
	mainCtx context.Context
	// Last ethereum block that we checked for receipts
	lastProcessedBlock *block
	// Map of transactions whose receipts we're looking for
	monitoredTxns map[common.Hash]info
	// Receipts retrieved from transactions
	receiptCache map[common.Hash]receipt
	// Struct to keep track of the gas metrics used by the system
	aggregates map[objects.FuncSelector]Profile
	// Map of groups of transactions that were retried
	retryGroups map[common.Hash]group
	// An interface with the ethereum functionality we need
	client ethereum.Network
	// Logger to log messages
	logger *logrus.Entry
	// Channel used to send request to this backend service
	requestChannel <-chan subscribeRequest
}

func (b *WatcherBackend) Loop() {
	// 30 * time.Second
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
			resp, err := b.queue(req)
			req.responseChannel.SendResponse(&SubscribeResponse{Err: err, Response: resp})

		case <-poolingTime:
			b.collectReceipts()
			poolingTime = time.After(constants.TxPollingTime)
			// todo status
		}
	}
}

func (wb *WatcherBackend) queue(req subscribeRequest) (*ReceiptResponse, error) {

	if req.txn == nil {
		return nil, &ErrInvalidMonitorRequest{"invalid request, missing txn object"}
	}
	txnHash := req.txn.Hash()
	fromAddr, err := wb.client.ExtractTransactionSender(req.txn)
	if err != nil {
		// faulty transaction
		return nil, &ErrInvalidMonitorRequest{fmt.Sprintf("cannot extract fromAddr from transaction: %v", txnHash)}
	}

	var receiptResponse *ReceiptResponse
	// if we already have the records of the receipt for this tx, we try to send the
	// receipt back
	if receipt, ok := wb.receiptCache[txnHash]; ok {
		receiptResponse = newReceiptResponse()
		receiptResponse.writeReceipt(receipt.receipt, nil)
	} else {
		var txInfo info
		if _, ok = wb.monitoredTxns[txnHash]; ok {
			txInfo = wb.monitoredTxns[txnHash]
		} else {
			selector, err := ExtractSelector(req.txn.Data())
			if err != nil {
				return nil, &ErrInvalidTransactionRequest{
					fmt.Sprintf("invalid request, transaction data is not present %v, err %v!", txnHash.Hex(), err),
				}
			}
			sig := signatures.FunctionMapping[selector]

			logEntry := wb.logger.WithField("Transaction", txnHash).
				WithField("Function", sig).
				WithField("Selector", fmt.Sprintf("%x", selector))

			txInfo = info{
				txn:               req.txn,
				fromAddress:       fromAddr,
				selector:          selector,
				functionSignature: sig,
				retryGroup:        txnHash,
				disableAutoRetry:  req.disableAutoRetry,
				logger:            logEntry,
			}
			wb.monitoredTxns[txnHash] = txInfo
			txGroup := newGroup()
			txGroup.add(txnHash)
			wb.retryGroups[txnHash] = txGroup
			logEntry.Debug("Transaction queued")
		}
		receiptResponse = wb.retryGroups[txInfo.retryGroup].receiptResponse
	}
	return receiptResponse, nil
}

func (wb *WatcherBackend) collectReceipts() {

	lenMonitoredTxns := len(wb.monitoredTxns)

	// If there's no tx to be monitored just return
	if lenMonitoredTxns == 0 {
		wb.logger.Tracef("no transaction to watch")
		return
	}

	networkCtx, cf := context.WithTimeout(wb.mainCtx, constants.TxNetworkTimeout)
	defer cf()

	blockHeader, err := wb.client.GetHeaderByNumber(networkCtx, nil)
	if err != nil {
		wb.logger.Debugf("error getting latest block number from ethereum node: %v", err)
		return
	}
	blockInfo := &block{
		blockHeader.Number.Uint64(),
		blockHeader.Hash(),
	}

	if wb.lastProcessedBlock.Equal(blockInfo) {
		wb.logger.Tracef("already processed block: %v with hash: %v", blockInfo.height, blockInfo.hash.Hex())
		return
	}
	wb.logger.Tracef("processing block: %v with hash: %v", blockInfo.height, blockInfo.hash.Hex())

	baseFee, tipCap, err := wb.client.GetBlockBaseFeeAndSuggestedGasTip(networkCtx)
	if err != nil {
		wb.logger.Debugf("error getting baseFee and suggested gas tip from ethereum node: %v", err)
		return
	}

	finishedTxs := make(map[common.Hash]MonitorWorkResponse)

	numWorkers := utils.Min(utils.Max(uint64(lenMonitoredTxns)/4, 128), 1)
	requestWorkChannel := make(chan MonitorWorkRequest, lenMonitoredTxns+3)
	responseWorkChannel := make(chan MonitorWorkResponse, lenMonitoredTxns+3)

	for txn, txnInfo := range wb.monitoredTxns {
		// if this is the first time seeing a tx or we have a reorg and
		// startedMonitoring is now greater than the current ethereum block height
		if txnInfo.monitoringHeight == 0 || txnInfo.monitoringHeight > blockInfo.height {
			txnInfo.monitoringHeight = blockInfo.height
			wb.monitoredTxns[txn] = txnInfo
		}
		requestWorkChannel <- MonitorWorkRequest{txnInfo, blockInfo.height}
	}

	// close the request channel, so the workers know when to finish
	close(requestWorkChannel)

	workerPool := NewWorkerPool(wb.mainCtx, wb.client, baseFee, tipCap, wb.logger, requestWorkChannel, responseWorkChannel)

	// spawn the workers and wait for all to complete
	go workerPool.ExecuteWork(numWorkers)

	for workResponse := range responseWorkChannel {
		select {
		case <-wb.mainCtx.Done():
			// main thread was killed
			return
		default:
		}
		logEntry := wb.logger.WithFields(logrus.Fields{"txn": workResponse.txnHash.Hex()})
		txInfo, ok := wb.monitoredTxns[workResponse.txnHash]
		if !ok {
			// invalid tx, should not happen, but well if it happens we continue
			logEntry.Trace("got a invalid tx with hash from workers")
			continue
		}
		if workResponse.err != nil {
			err := workResponse.err
			switch err.(type) {
			case *ErrRecoverable:
				logEntry.Tracef("Retrying! Got a recoverable error when trying to get receipt, err: %v", workResponse.err)
			case *ErrTxNotFound:
				// since we only analyze a tx once per new block, the notFoundBlocks counter
				// should have approx the amount of blocks that we failed on finding the tx
				txInfo.notFoundBlocks++
				if txInfo.notFoundBlocks >= wb.client.GetTxNotFoundMaxBlocks() {
					logEntry.Debugf("Couldn't get tx receipt, err: %v", workResponse.err)
					finishedTxs[workResponse.txnHash] = workResponse
				}
				logEntry.Tracef("Retrying, couldn't get info, num attempts: %v, err: %v", txInfo.notFoundBlocks, workResponse.err)
			case *ErrTransactionStale:
				// If we get this error it means that we should not retry or we cannot retry
				// automatically, should forward the error to the subscribers
				logEntry.Debugf("Stale transaction, err: %v", workResponse.err)
				finishedTxs[workResponse.txnHash] = workResponse
			}
		} else {
			if workResponse.retriedTxn != nil {
				// restart the monitoringHeight, so we don't retry the tx in the next block
				txInfo.monitoringHeight = 0
				if workResponse.retriedTxn.err == nil && workResponse.retriedTxn.txn != nil {
					newTxnHash := workResponse.retriedTxn.txn.Hash()
					wb.monitoredTxns[newTxnHash] = info{
						txn:                    workResponse.retriedTxn.txn,
						fromAddress:            txInfo.fromAddress,
						selector:               txInfo.selector,
						functionSignature:      txInfo.functionSignature,
						retryGroup:             txInfo.retryGroup,
						receiptResponseChannel: txInfo.receiptResponseChannel,
					}
					// increase the number of tx in the retry group
					wb.retryGroups[wb.retryGroupId]++
					txInfo.retryAmount++
					logEntry.Tracef("successfully replaced a tx with %v", newTxnHash)
				} else {
					logEntry.Debugf("could not replace tx error %v", workResponse.retriedTxn.err)
				}
			}
			if workResponse.receipt != nil {
				logEntry.WithFields(
					logrus.Fields{
						"mined":          workResponse.receipt.BlockNumber,
						"current height": blockInfo.height,
					},
				).Debug("Successfully got receipt")
				wb.receiptCache[workResponse.txnHash] = receipt{receipt: workResponse.receipt, retrievedAtHeight: blockInfo.height}
				finishedTxs[workResponse.txnHash] = workResponse
			}
		}
		wb.monitoredTxns[workResponse.txnHash] = txInfo
	}

	// Cleaning finished and failed transactions
	for txnHash, workResponse := range finishedTxs {
		if txnInfo, ok := wb.monitoredTxns[txnHash]; ok {
			// if this is the unique tx in the retry group or we have the receipt, we are good to send the response
			if wb.retryGroups[txnInfo.retryGroup] == 1 || workResponse.receipt != nil {
				if workResponse.err != nil {
					wb.logger.Tracef("sending response for tx: %v, err %v", txnHash.Hex(), workResponse.err)
				}
				if workResponse.receipt != nil {
					wb.logger.Tracef(
						"sending response for tx: %v, with receipt status %v mined at block %v",
						txnHash.Hex(),
						workResponse.receipt.Status,
						workResponse.receipt.BlockHash,
					)
				}
				if !txnInfo.receiptResponseChannel.isClosed {
					wb.logger.Tracef("sending response to channel for tx %v", txnHash.Hex())
					txnInfo.receiptResponseChannel.SendResponse(&objects.ReceiptResponse{TxnHash: workResponse.txnHash, Receipt: workResponse.receipt, Err: workResponse.err})
				}
			}
			if wb.retryGroups[txnInfo.retryGroup] >= 1 {
				wb.retryGroups[txnInfo.retryGroup]--
				wb.logger.Tracef("removing tx entry: %v from retry group with has now %v members", txnHash.Hex(), wb.retryGroups[txnInfo.retryGroup])
			}
			delete(wb.monitoredTxns, txnHash)
		}
	}

	var expiredReceipts []common.Hash
	// Marking expired receipts
	for receiptTxnHash, receiptInfo := range wb.receiptCache {
		if blockInfo.height >= receiptInfo.retrievedAtHeight+constants.TxReceiptCacheMaxBlocks {
			expiredReceipts = append(expiredReceipts, receiptTxnHash)
		}
	}
	for _, receiptTxHash := range expiredReceipts {
		wb.logger.Tracef("cleaning %v from receipt cache", receiptTxHash.Hex())
		delete(wb.receiptCache, receiptTxHash)
	}

	wb.lastProcessedBlock = blockInfo
}

// Structs that keep track of the state needed by the worker pool service. The
// workerPool spawn multiple go routines (workers) to check and retrieve the
// receipts.
type WorkerPool struct {
	wg                  *sync.WaitGroup
	ctx                 context.Context            // Main context passed by the parent, used to cancel worker and the pool
	client              ethereum.Network           // An interface with the Geth functionality we need
	baseFee             *big.Int                   // Base fee of the current block in case we need to retry a stale transaction
	tipCap              *big.Int                   // Miner tip cap of the current block in case we need to retry a stale transaction
	logger              *logrus.Entry              // Logger to log messages
	requestWorkChannel  <-chan MonitorWorkRequest  // Channel where will be send the work requests
	responseWorkChannel chan<- MonitorWorkResponse // Channel where the work response will be send
}

// Creates a new WorkerPool service
func NewWorkerPool(ctx context.Context, client ethereum.Network, baseFee *big.Int, tipCap *big.Int, logger *logrus.Entry, requestWorkChannel <-chan MonitorWorkRequest, responseWorkChannel chan<- MonitorWorkResponse) *WorkerPool {
	return &WorkerPool{new(sync.WaitGroup), ctx, client, baseFee, tipCap, logger, requestWorkChannel, responseWorkChannel}
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
		RetryLoop:
			for i := uint64(1); i <= constants.TxWorkerMaxWorkRetries; i++ {
			RetrySelect:
				select {
				case <-ctx.Done():
					// worker context timed out or parent was cancelled, should return
					return
				default:
					rcpt, err := w.getReceipt(ctx, monitoredTx, currentHeight, txnHash)
					if err != nil {
						switch err.(type) {
						case *ErrRecoverable:
							// retry on recoverable error `constants.TxWorkerMaxWorkRetries` times
							if i < constants.TxWorkerMaxWorkRetries {
								continue RetryLoop
							}
						case *ErrTransactionStale:
							if !monitoredTx.disableAutoRetry {
								defaultAccount := w.client.GetDefaultAccount()
								if bytes.Equal(monitoredTx.fromAddress[:], defaultAccount.Address[:]) {
									newTxn, retryTxErr := w.client.RetryTransaction(ctx, monitoredTx.txn, w.baseFee, w.tipCap)
									w.responseWorkChannel <- MonitorWorkResponse{txnHash: txnHash, retriedTxn: &retriedTransaction{txn: newTxn, err: retryTxErr}}
									break RetrySelect
								}
							}
						}
						// send recoverable errors after constants.TxWorkerMaxWorkRetries,txNotFound or
						// other errors back to main
						w.responseWorkChannel <- MonitorWorkResponse{txnHash: txnHash, err: err}
					} else {
						// send receipt (even if it nil) back to main thread
						w.responseWorkChannel <- MonitorWorkResponse{txnHash: txnHash, receipt: rcpt}
					}
					//should continue getting other tx work
					break RetryLoop
				}
			}
		}
	}
}

// Internal function used by the workers to check/retrieve the receipts for a given transaction
func (w *WorkerPool) getReceipt(ctx context.Context, monitoredTx info, currentHeight uint64, txnHash common.Hash) (*types.Receipt, error) {
	txnHex := txnHash.Hex()
	blockTimeSpan := currentHeight - monitoredTx.monitoringHeight
	_, isPending, err := w.client.GetTransactionByHash(ctx, txnHash)
	if err != nil {
		// if we couldn't locate a tx after NotFoundMaxBlocks blocks and we are still
		// failing in getting the tx data, probably means that it was dropped
		if errors.Is(err, goEthereum.NotFound) {
			return nil, &ErrTxNotFound{fmt.Sprintf("could not find tx %v in the height %v!", txnHex, currentHeight)}
		}
		// probably a network error, should retry
		return nil, &ErrRecoverable{fmt.Sprintf("error getting tx: %v: %v", txnHex, err)}
	}
	if isPending {
		// We multiply MaxStaleBlocks by the number of times that we tried to retry a tx
		// to add a increasing delay between successful retry attempts.
		// startedMonitoringHeight is restarted at every retry attempt. Most of the time
		// after a successful retry, the tx being replaced will fall in the branch above
		// (err tx not found). But in case of an edge case, where tx replacing and tx
		// replaced are both valid (e.g sending tx to different nodes) we will continue
		// to retry both, until we have a valid tx for this nonce.
		maxPendingBlocks := w.client.GetTxMaxStaleBlocks() * (monitoredTx.retryAmount + 1)
		if blockTimeSpan >= maxPendingBlocks {
			return nil, &ErrTransactionStale{fmt.Sprintf("error tx: %v is stale on the memory pool for more than %v blocks!", txnHex, w.client.GetTxMaxStaleBlocks())}
		}
	} else {
		// tx is not pending, so check for receipt
		rcpt, err := w.client.GetTransactionReceipt(ctx, txnHash)
		if err != nil {
			// if can locate a tx (branch logic above), but we cannot locate a tx receipt
			// after NotFoundMaxBlocks blocks, there's definitely something wrong
			if errors.Is(err, goEthereum.NotFound) {
				return nil, &ErrTxNotFound{fmt.Sprintf("could not find receipt for tx %v in the height %v!", txnHex, currentHeight)}
			}
			// probably a network error, should retry
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
func NewWatcher(client ethereum.Network, txConfirmationBlocks uint64) *Watcher {
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
		retryGroups:        make(map[common.Hash]group),
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
	watcher := NewWatcher(network, network.GetFinalityDelay())
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

// Subscribe a transaction to be watched by the transaction watcher service. If
// a transaction was accepted by the watcher service, a response channel is
// returned. The response channel is where the receipt going to be sent. The
// final tx hash in the receipt can be different from the initial txn sent. This
// can happen if the txn got stale and the watcher did a transaction replace
// with higher fees.
func (w *Watcher) Subscribe(ctx context.Context, txn *types.Transaction) (<-chan *objects.ReceiptResponse, error) {
	w.logger.WithField("Txn", txn.Hash().Hex()).Debug("Subscribing for a transaction")
	respChannel := NewResponseChannel[SubscribeResponse](w.logger)
	defer respChannel.CloseChannel()
	req := subscribeRequest{txn: txn, responseChannel: respChannel}

	select {
	case w.requestChannel <- req:
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

// function that wait for a transaction receipt. This is blocking function that
// will wait for a response in the input receiptResponseChannel
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

func ExtractSelector(data []byte) (objects.FuncSelector, error) {
	var selector [4]byte
	if len(data) < 4 {
		return selector, fmt.Errorf("couldn't extract selector for data: %v", data)
	}
	for idx := 0; idx < 4; idx++ {
		selector[idx] = data[idx]
	}
	return selector, nil
}
