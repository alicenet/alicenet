package blockchain

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/logging"
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

// Internal struct to keep track of transactions that are being monitoring
type TransactionInfo struct {
	ctx                     context.Context         // ctx used for calling the monitoring a certain tx
	txn                     *types.Transaction      // Transaction object
	selector                interfaces.FuncSelector // 4 bytes that identify the function being called by the tx
	functionSignature       string                  // function signature as we see on the smart contracts
	startedMonitoringHeight uint64                  // ethereum height where we first added the tx to be watched. Mainly used to see if a tx was dropped
	receiptResponseChannel  *ReceiptResponseChannel // channel where the response will be sent
}

// Internal struct to keep track of what blocks we already checked during monitoring
type BlockInfo struct {
	height uint64      // block height
	hash   common.Hash // block header hash
}

// Compare if 2 blockInfo structs are equal by comparing the height and block
// hash. Return true in case they are equal, false otherwise.
func (a *BlockInfo) Equal(b *BlockInfo) bool {
	return bytes.Equal(a.hash[:], b.hash[:]) && a.height == b.height
}

// Type to do request against the tx receipt monitoring system. Ctx and response channel should be set
type MonitorRequest struct {
	ctx             context.Context        // tx ctx used for tx monitoring cancellation
	txn             *types.Transaction     // the transaction that should watched
	responseChannel chan<- MonitorResponse // channel where we going to send the channel where the response will be sent
}

// Type that it's going to be used to reply a request
type MonitorResponse struct {
	receiptResponseChannel <-chan ReceiptResponse // channel where the result from the tx/receipt monitoring will be send
	err                    error                  // errors that happened when processing the monitor request
}

// Response of the monitoring system
type ReceiptResponse struct {
	txnHash common.Hash    // Hash of the txs which this response belongs
	err     error          // response error that happened during processing
	receipt *types.Receipt // tx receipt after txConfirmationBlocks of a tx that was not queued in txGroup
}

// Try to send a request monitor response in a channel. In case the channel is full, we just skip and log the error
func SendMonitorResponse(responseChannel chan<- MonitorResponse, monitorResponse MonitorResponse, logger *logrus.Entry) {
	select {
	case responseChannel <- monitorResponse:
	default:
		logger.Debugf("Failed to write request response Channel")
	}
}

type ReceiptResponseChannel struct {
	writeOnce sync.Once
	channel   chan ReceiptResponse
	isClosed  bool
	logger    *logrus.Entry
}

func NewReceiptResponseChannel(logger *logrus.Entry) *ReceiptResponseChannel {
	return &ReceiptResponseChannel{channel: make(chan ReceiptResponse, 1), logger: logger}
}

// send a receipt and close the receiptResponseChannel. This function should
// only be called once. Additional calls will be no-op
func (rc *ReceiptResponseChannel) SendReceipt(receipt ReceiptResponse) {
	if !rc.isClosed {
		select {
		case rc.channel <- receipt:
		default:
			rc.logger.Debugf("Failed to write request response Channel")
		}
		rc.CloseChannel()
	}
}

// Close the internal channel. This function should only be called once.
// Additional calls will be no-op
func (rc *ReceiptResponseChannel) CloseChannel() {
	rc.writeOnce.Do(func() {
		rc.isClosed = true
		close(rc.channel)
	})
}

// TransactionProfile to keep track of gas metrics in the overall system
type TransactionProfile struct {
	AverageGas   uint64
	MinimumGas   uint64
	MaximumGas   uint64
	TotalCount   uint64
	TotalGas     uint64
	TotalSuccess uint64
}

type MonitorWorkRequest struct {
	txn    TransactionInfo
	height uint64
}

type MonitorWorkResponse struct {
	txnHash common.Hash
	err     error
	receipt *types.Receipt
}

// Backend structs used to monitor Ethereum transactions and retrieve their receipts
type Behind struct {
	mainCtx              context.Context                                // main context for the background services
	lastProcessedBlock   *BlockInfo                                     // Last ethereum block that we checked for receipts
	monitoredTxns        map[common.Hash]TransactionInfo                // Map of transactions whose receipts we're looking for
	receiptCache         map[common.Hash]*types.Receipt                 // Receipts retrieved from transactions
	txConfirmationBlocks uint64                                         // number of ethereum blocks that we should wait to consider a receipt valid
	aggregates           map[interfaces.FuncSelector]TransactionProfile // Struct to keep track of the gas metrics used by the system
	client               interfaces.GethClient                          // An interface with the Geth functionality we need
	knownSelectors       interfaces.SelectorMap                         // Map with signature -> name
	logger               *logrus.Entry                                  // Logger to log messages
	requestChannel       <-chan MonitorRequest                          // Channel used to send request to this backend service
}

func (b *Behind) Loop() {

	poolingTime := time.After(constants.TxPollingTime)
	for {
		select {
		case req, ok := <-b.requestChannel:
			if !ok {
				b.logger.Debugf("request channel closed, exiting")
				return
			}
			b.logger.Debugf("received request: %v channel open: %v", req, ok)
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
	b.logger.Debug("finished")
}

func (b *Behind) queue(req MonitorRequest) {

	if req.txn == nil {
		SendMonitorResponse(req.responseChannel, MonitorResponse{err: &ErrInvalidMonitorRequest{"invalid request, missing txn object"}}, b.logger)
		return
	}
	if req.ctx == nil {
		SendMonitorResponse(req.responseChannel, MonitorResponse{err: &ErrInvalidMonitorRequest{"invalid request, missing ctx"}}, b.logger)
		return
	}

	txnHash := req.txn.Hash()
	receiptResponseChannel := NewReceiptResponseChannel()

	// if we already have the records of the receipt for this tx we try to send the
	// receipt back
	if receipt, ok := b.receiptCache[txnHash]; ok {
		receiptResponseChannel.SendReceipt(ReceiptResponse{receipt: receipt, txnHash: txnHash})
	} else {
		if _, ok := b.monitoredTxns[txnHash]; ok {
			SendMonitorResponse(req.responseChannel, MonitorResponse{err: &ErrTransactionAlreadyQueued{"invalid request, tx already queued, try to get receipt later!"}}, b.logger)
			return
		}

		selector := ExtractSelector(req.txn.Data())
		//todo: replace this with a generated mapping from the bindings
		sig := b.knownSelectors.Signature(selector)

		logEntry := b.logger.WithField("Transaction", txnHash).
			WithField("Function", sig).
			WithField("Selector", fmt.Sprintf("%x", selector))

		b.monitoredTxns[txnHash] = TransactionInfo{
			ctx:                    req.ctx,
			txn:                    req.txn,
			selector:               selector,
			functionSignature:      sig,
			receiptResponseChannel: receiptResponseChannel,
		}
		logEntry.Debug("Transaction queued")
	}
	SendMonitorResponse(req.responseChannel, MonitorResponse{receiptResponseChannel: receiptResponseChannel.channel}, b.logger)
}

func (b *Behind) collectReceipts() {

	lenMonitoredTxns := len(b.monitoredTxns)

	// If there's no tx to be monitored just return
	if lenMonitoredTxns == 0 {
		b.logger.Tracef("TxMonitor: no transaction to watch")
		return
	}

	networkCtx, cf := context.WithTimeout(b.mainCtx, constants.TxNetworkTimeout)
	defer cf()

	blockHeader, err := b.client.HeaderByNumber(networkCtx, nil)
	if err != nil {
		b.logger.Debugf("TxMonitor: error getting latest block number from ethereum node: %v", err)
		return
	}
	blockInfo := &BlockInfo{
		blockHeader.Number.Uint64(),
		blockHeader.Hash(),
	}

	if b.lastProcessedBlock.Equal(blockInfo) {
		b.logger.Tracef("TxMonitor: block: %v with hash: %v already processed", blockInfo.height, blockInfo.hash.Hex())
		return
	}

	var expiredTxs []common.Hash
	var finishedTxs map[common.Hash]MonitorWorkResponse

	numWorkers := min(max(uint64(lenMonitoredTxns)/4, 128), 1)
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
			// if this is the first time seeing a tx
			if txnInfo.startedMonitoringHeight == 0 {
				txnInfo.startedMonitoringHeight = blockInfo.height
			}
			requestWorkChannel <- MonitorWorkRequest{txnInfo, blockInfo.height}
		}
	}

	// close the request channel, so the workers know when to finish
	close(requestWorkChannel)

	workerPool := NewWorkerPool(b.mainCtx, b.client, b.logger, b.txConfirmationBlocks, requestWorkChannel, responseWorkChannel)

	// spawn the workers and wait for all to complete
	go workerPool.ExecuteWork(numWorkers)

	for workResponse := range responseWorkChannel {
		select {
		case <-b.mainCtx.Done():
			// main thread was killed
			return
		default:
			if workResponse.err != nil {
				if _, ok := err.(*ErrRecoverable); !ok {
					b.logger.Debugf("Couldn't get tx receipt for tx:%v cause: %v", workResponse.txnHash, workResponse.err)
					finishedTxs[workResponse.txnHash] = workResponse
				}
			} else if workResponse.receipt != nil {
				b.logger.Debugf("Successfully got receipt for tx:%v", workResponse.txnHash)
				b.receiptCache[workResponse.txnHash] = workResponse.receipt
				finishedTxs[workResponse.txnHash] = workResponse
			}
		}
	}

	//todo: clean the cache
	// Cleaning finished and failed transactions
	for txnHash, workResponse := range finishedTxs {
		if txnInfo, ok := b.monitoredTxns[txnHash]; ok {
			txnInfo.receiptResponseChannel.SendReceipt(ReceiptResponse{txnHash: workResponse.txnHash, receipt: workResponse.receipt, err: workResponse.err})
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

	b.lastProcessedBlock = blockInfo
}

type WorkerPool struct {
	wg                   *sync.WaitGroup
	ctx                  context.Context            //
	ethClient            interfaces.GethClient      // An interface with the Geth functionality we need
	logger               *logrus.Entry              // Logger to log messages
	txConfirmationBlocks uint64                     //
	requestWorkChannel   <-chan MonitorWorkRequest  //
	responseWorkChannel  chan<- MonitorWorkResponse //
}

func NewWorkerPool(ctx context.Context, ethClient interfaces.GethClient, logger *logrus.Entry, txConfirmationBlocks uint64, requestWorkChannel <-chan MonitorWorkRequest, responseWorkChannel chan<- MonitorWorkResponse) *WorkerPool {
	return &WorkerPool{new(sync.WaitGroup), ctx, ethClient, logger, txConfirmationBlocks, requestWorkChannel, responseWorkChannel}
}

// Function to spawn the workers and wait for the job to be done. This function
// sends a message in the completionChannel once all workers have finished
func (w *WorkerPool) ExecuteWork(numWorkers uint64) {
	for i := uint64(0); i < numWorkers; i++ {
		w.wg.Add(1)
		go w.worker()
	}
	w.wg.Wait()
	close(w.responseWorkChannel)
}

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
							continue
						}
						// send nonRecoverable errors back to main or recoverable errors after constants.TxWorkerMaxWorkRetries
						w.responseWorkChannel <- MonitorWorkResponse{txnHash: txnHash, err: err}
					} else {
						// send receipt (even if it nil) back to main thread
						w.responseWorkChannel <- MonitorWorkResponse{txnHash: txnHash, receipt: rcpt}
					}
					//should continue getting other tx work
					break
				}
			}
		}
	}
}

func (w *WorkerPool) getReceipt(ctx context.Context, monitoredTx TransactionInfo, currentHeight uint64, txnHash common.Hash) (*types.Receipt, error) {
	txnHex := txnHash.Hex()
	blockTimeSpan := currentHeight - monitoredTx.startedMonitoringHeight
	_, isPending, err := w.ethClient.TransactionByHash(ctx, txnHash)
	if err != nil {
		// if we couldn't locate a tx after NotFoundMaxBlocks blocks and we are still
		// failing in getting the tx data, probably means that it was dropped
		if blockTimeSpan >= constants.TxNotFoundMaxBlocks {
			return nil, &ErrNonRecoverable{fmt.Sprintf("could not find tx %v and %v blocks have passed!", txnHex, constants.TxNotFoundMaxBlocks)}
		}
		// probably a network error, should retry
		return nil, &ErrRecoverable{fmt.Sprintf("error getting tx: %v: %v", txnHex, err)}
	}
	if isPending {
		if blockTimeSpan >= constants.TxMaxStaleBlocks {
			return nil, &ErrTransactionStale{fmt.Sprintf("error tx: %v is stale on the memory pool for more than %v blocks, please retry!", txnHex, constants.TxMaxStaleBlocks)}
		}
	} else {
		// tx is not pending, so check for receipt
		rcpt, err := w.ethClient.TransactionReceipt(ctx, txnHash)
		if err != nil {
			// if we couldn't locate a tx receipt after NotFoundMaxBlocks blocks and we are still
			// failing in getting the tx data, probably means that it was dropped
			if blockTimeSpan >= constants.TxNotFoundMaxBlocks {
				return nil, &ErrNonRecoverable{fmt.Sprintf("could not find receipt for tx %v and %v blocks have passed!", txnHex, constants.TxNotFoundMaxBlocks)}
			}
			// 1. probably a network error, should retry
			// 2. in case receipt not found after tx not Pending check, we had an edge case,
			// probably tx was mined (isPending == false), then we had a chain re-org, now
			// the tx is back to the memPool or was dropped, we should retry, and the logic
			// above should see if the tx was dropped or not in the next iteration
			return nil, &ErrRecoverable{fmt.Sprintf("error getting receipt: %v: %v", txnHex, err)}
		}

		if currentHeight >= rcpt.BlockNumber.Uint64()+w.txConfirmationBlocks {
			return rcpt, nil
		}
	}
	return nil, nil
}

type TxnQueueDetail struct {
	backend          *Behind
	logger           *logrus.Entry
	closeMainContext context.CancelFunc
	requestChannel   chan<- MonitorRequest
}

func NewTxnQueue(client interfaces.GethClient, sm interfaces.SelectorMap, txConfirmationBlocks uint64) *TxnQueueDetail {
	reqChannel := make(chan MonitorRequest, 100)
	// main context that will cancel all workers and go routine
	mainCtx, cf := context.WithCancel(context.Background())

	b := &Behind{
		mainCtx:              mainCtx,
		requestChannel:       reqChannel,
		client:               client,
		logger:               logging.GetLogger("ethereum").WithField("Component", "behind"),
		monitoredTxns:        make(map[common.Hash]TransactionInfo),
		receiptCache:         make(map[common.Hash]*types.Receipt),
		aggregates:           make(map[interfaces.FuncSelector]TransactionProfile),
		knownSelectors:       sm,
		lastProcessedBlock:   &BlockInfo{0, common.HexToHash("")},
		txConfirmationBlocks: txConfirmationBlocks,
	}

	q := &TxnQueueDetail{
		requestChannel:   reqChannel,
		closeMainContext: cf,
		backend:          b,
		logger:           logging.GetLogger("ethereum").WithField("Component", "infront"),
	}
	return q
}

func (f *TxnQueueDetail) StartLoop() {
	go f.backend.Loop()
}

func (f *TxnQueueDetail) Close() {
	f.logger.Debug("closing request channel...")
	close(f.requestChannel)
	f.closeMainContext()
}

func (f *TxnQueueDetail) QueueTransaction(ctx context.Context, txn *types.Transaction) (<-chan ReceiptResponse, error) {
	f.logger.WithField("Txn", txn.Hash().Hex()).Debug("Queueing")
	respChannel := make(chan MonitorResponse, 1)
	defer close(respChannel)
	req := MonitorRequest{ctx: ctx, txn: txn, responseChannel: respChannel}
	f.requestChannel <- req
	select {
	case txResponseChannel, ok := <-req.responseChannel:
		if !ok {
			return nil, errors.New("response channel closed")
		}
		return txResponseChannel.receiptResponseChannel, txResponseChannel.err
	case <-ctx.Done():
		f.logger.Infof("context cancelled: %v", ctx.Err())
		return nil, ctx.Err()
	}
}

func (f *TxnQueueDetail) WaitTransaction(ctx context.Context, txResponseChannel <-chan ReceiptResponse) (*types.Receipt, error) {
	select {
	case resp, ok := <-txResponseChannel:
		if !ok {
			return nil, errors.New("tx response channel closed")
		}
		return resp.receipt, resp.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (f *TxnQueueDetail) QueueAndWait(ctx context.Context, txn *types.Transaction) (*types.Receipt, error) {
	txResponseChannel, err := f.QueueTransaction(ctx, txn)
	if err != nil {
		return nil, err
	}
	return f.WaitTransaction(ctx, txResponseChannel)
}

func isTestRun() bool {
	return flag.Lookup("test.v") != nil
}

func max(a uint64, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func min(a uint64, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}

// func (f *TxnQueueDetail) QueueGroupTransaction(ctx context.Context, groupId uint64, txn *types.Transaction) (*Response, error) {
// 	if groupId == 0 {
// 		return nil, errors.New("groupTx should be greater than 0!")
// 	}
// 	respChannel := make(chan chan *Response, 1)
// 	defer close(respChannel)
// 	req := &Request{ctx: ctx, command: "queue", txn: txn, txGroup: groupId, respCh: respChannel}
// 	response, err := f.requestWait(ctx, req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return response, nil
// }

// func (f *TxnQueueDetail) WaitGroupTransactions(ctx context.Context, txns []*types.Transaction) ([]*types.Receipt, error) {
// 	respChannel := make(chan *Response, 1)
// 	defer close(respChannel)
// 	req := &Request{ctx: ctx, command: "wait", txGroup: groupId, respCh: respChannel}
// 	resp, err := f.requestWait(ctx, req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if resp.rcpts == nil {
// 		return nil, errors.New(fmt.Sprintf("no receipts were found for group: %v", groupId))
// 	}

// 	return resp.rcpts, nil
// }
