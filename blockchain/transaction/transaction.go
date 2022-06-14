package transaction

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/bridge/bindings"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/constants/dbprefix"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
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

type IReceiptResponse interface {
	IsReady() bool
	GetReceipt(ctx context.Context) (*types.Receipt, error)
}

type IWatcher interface {
	Start() error
	Close()
	Subscribe(ctx context.Context, txn *types.Transaction, disableAutoRetry bool) (IReceiptResponse, error)
	Wait(ctx context.Context, receiptResponse IReceiptResponse) (*types.Receipt, error)
	SubscribeAndWait(ctx context.Context, txn *types.Transaction, disableAutoRetry bool) (*types.Receipt, error)
}

type FuncSelector [4]byte

// Internal struct to keep track of transactions that are being monitoring
type info struct {
	txn               *types.Transaction `json:"txn"`               // Transaction object
	fromAddress       common.Address     `json:"fromAddress"`       // address of the transaction signer
	selector          FuncSelector       `json:"selector"`          // 4 bytes that identify the function being called by the tx
	functionSignature string             `json:"functionSignature"` // function signature as we see on the smart contracts
	retryGroup        common.Hash        `json:"retryGroup"`        // internal group Id to keep track of all tx that were created during the retry of a tx
	disableAutoRetry  bool               `json:"disableAutoRetry"`  // whether we should disable the auto retry of a transaction
	monitoringHeight  uint64             `json:"monitoringHeight"`  // ethereum height where we first added the tx to be watched or did a tx retry.
	retryAmount       uint64             `json:"retryAmount"`       // counter to indicate how many times we tried to retry a transaction
	notFoundBlocks    uint64             `json:"notFoundBlocks"`    // counter to indicate approximate number of blocks that we could not find a tx
	logger            *logrus.Entry      `json:"-"`                 // logger to log transaction info
}

// Internal struct to keep track of transactions retries groups
type group struct {
	internalGroup   []common.Hash    `json:"internalGroup"` // slice where we keep track of all tx in a group
	receiptResponse *ReceiptResponse `json:"-"`             // struct used to send/share the receipt
}

// creates a new group
func newGroup() group {
	return group{receiptResponse: newReceiptResponse()}
}

// add a new hash to the group
func (g *group) add(txHash common.Hash) {
	g.internalGroup = append(g.internalGroup, txHash)
}

// remove a hash from the group
func (g *group) remove(txHash common.Hash) error {
	index := -1
	lastIndex := len(g.internalGroup) - 1
	if lastIndex == -1 {
		return fmt.Errorf("invalid removal, empty group %v", txHash.Hex())
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

// check if a group is empty
func (g *group) isEmpty() bool {
	return len(g.internalGroup) == 0
}

// send a receipt inc ase this group has an unique tx or we have the receipt
func (g *group) sendReceipt(logger *logrus.Entry, receipt *types.Receipt, err error) {
	if g.isEmpty() {
		logger.Trace("empty group, cannot send receipt")
		return
	}
	if err != nil {
		logger.Tracef("sending group err %v", err)
	}
	if receipt != nil {
		logger.Tracef(
			"sending response group with receipt status %v mined at block %v",
			receipt.Status,
			receipt.BlockHash,
		)
	}
	// if this is the unique tx in the retry group or we have the receipt, we are good to send the response
	if len(g.internalGroup) == 1 || receipt != nil {
		logger.Trace("sending tx")
		// in case we are recovering the group from a serialization during a crash, receiptResponse will be nil
		if g.receiptResponse == nil {
			g.receiptResponse = newReceiptResponse()
		}
		g.receiptResponse.writeReceipt(receipt, err)
	} else {
		logger.Tracef("not sending tx since group has more than one txn, group.len: %v", len(g.internalGroup))
	}
}

// making sure that struct conforms the interface
var _ IReceiptResponse = &ReceiptResponse{}

// Struct to send and share a receipt retrieved by the watcher
type ReceiptResponse struct {
	doneChan chan struct{}
	err      error          // response error that happened during processing
	receipt  *types.Receipt // tx receipt after txConfirmationBlocks of a tx that was not queued in txGroup
}

func newReceiptResponse() *ReceiptResponse {
	return &ReceiptResponse{doneChan: make(chan struct{}, 1)}
}

// Function to check if a receipt is ready
func (r *ReceiptResponse) IsReady() bool {
	select {
	case <-r.doneChan:
		return true
	default:
		return false
	}
}

// todo: revisit this
// blocking function to get the receipt from a transaction. This function will
// block until the receipt is available and sent by the transaction watcher
// service.
func (r *ReceiptResponse) GetReceiptBlock(ctx context.Context) (*types.Receipt, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("error waiting for receipt: %v", ctx.Err())
		case <-r.doneChan:
			return r.receipt, r.err
		}
	}
}

// function to write the receipt or error from a transaction being watched.
func (r *ReceiptResponse) writeReceipt(receipt *types.Receipt, err error) {
	if receipt == nil && err == nil {
		return
	}
	if !r.IsReady() {
		r.receipt = receipt
		r.err = err
		close(r.doneChan)
	}

}

// Internal struct to keep track of the receipts
type receipt struct {
	receipt           *types.Receipt `json:"receipt"` // receipt object
	retrievedAtHeight uint64         `json:"-"`       // block height where receipt was added to the cache
}

// Internal struct to keep track of what blocks we already checked during monitoring
type block struct {
	height uint64      `json:"height"` // block height
	hash   common.Hash `json:"hash"`   // block header hash
}

// Compare if 2 blockInfo structs are equal by comparing the height and block
// hash. Return true in case they are equal, false otherwise.
func (a *block) Equal(b *block) bool {
	return bytes.Equal(a.hash[:], b.hash[:]) && a.height == b.height
}

// Type to do subscription request against the tx watcher system. SubscribeResponseChannel should be set
type SubscribeRequest struct {
	txn              *types.Transaction        // the transaction that should watched
	disableAutoRetry bool                      // whether we should disable the auto retry of a transaction
	responseChannel  *SubscribeResponseChannel // channel where we going to send the request response
}

// creates a new subscribe request
func newSubscribeRequest(txn *types.Transaction, disableAutoRetry bool) SubscribeRequest {
	return SubscribeRequest{txn: txn, responseChannel: NewResponseChannel(), disableAutoRetry: disableAutoRetry}
}

// blocking function to listen for the response of a subscribe request
func (a SubscribeRequest) Listen(ctx context.Context) (*ReceiptResponse, error) {
	select {
	case subscribeResponse := <-a.responseChannel.channel:
		return subscribeResponse.Response, subscribeResponse.Err
	case <-ctx.Done():
		return nil, &ErrInvalidTransactionRequest{fmt.Sprintf("context cancelled: %v", ctx.Err())}
	}
}

// Type that it's going to be used to reply a subscription request
type SubscribeResponse struct {
	Err      error            // errors that happened when processing the subscription request
	Response *ReceiptResponse // struct where the receipt from the tx monitoring will be send
}

// A response channel is basically a non-blocking channel that can only be
// written and closed once.
type SubscribeResponseChannel struct {
	writeOnce sync.Once
	channel   chan *SubscribeResponse // internal channel
}

// Create a new response channel.
func NewResponseChannel() *SubscribeResponseChannel {
	return &SubscribeResponseChannel{channel: make(chan *SubscribeResponse, 1)}
}

// send a unique response and close the internal channel. Additional calls to
// this function will be no-op
func (rc *SubscribeResponseChannel) sendResponse(response *SubscribeResponse) {
	rc.writeOnce.Do(func() {
		rc.channel <- response
		close(rc.channel)
	})
	// todo: panic if the channel is called more than once
}

// Profile to keep track of gas metrics in the overall system
type Profile struct {
	AverageGas   uint64 `json:"averageGas"`
	MinimumGas   uint64 `json:"minimumGas"`
	MaximumGas   uint64 `json:"maximumGas"`
	TotalGas     uint64 `json:"totalGas"`
	TotalCount   uint64 `json:"totalCount"`
	TotalSuccess uint64 `json:"totalSuccess"`
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
	err        error               // any error found during the receipt retrieve (can be NonRecoverable, Recoverable or TransactionStale errors)
	receipt    *types.Receipt      // receipt retrieved (can be nil) if a receipt was not found or it's not ready yet
}

// Internal struct to keep track of retried transaction by the workers
type retriedTransaction struct {
	txn *types.Transaction // new transaction after the retry attempt
	err error              // error that happened during the transaction retry
}

// Backend struct used to monitor Ethereum transactions and retrieve their receipts
type WatcherBackend struct {
	mainCtx            context.Context          `json:"-"`             // main context for the background services
	lastProcessedBlock *block                   `json:"-"`             // Last ethereum block that we checked for receipts
	monitoredTxns      map[common.Hash]info     `json:"monitoredTxns"` // Map of transactions whose receipts we're looking for
	receiptCache       map[common.Hash]receipt  `json:"receiptCache"`  // Receipts retrieved from transactions. The keys are are txGroup hashes
	aggregates         map[FuncSelector]Profile `json:"aggregates"`    // Struct to keep track of the gas metrics used by the system
	retryGroups        map[common.Hash]group    `json:"retryGroups"`   // Map of groups of transactions that were retried
	client             ethereum.Network         `json:"-"`             // An interface with the ethereum functionality we need
	logger             *logrus.Entry            `json:"-"`             // Logger to log messages
	requestChannel     <-chan SubscribeRequest  `json:"-"`             // Channel used to send request to this backend service
	database           *db.Database             `json:"-"`             // database where we are going to persist and load state
	metricsDisplay     bool                     `json:"-"`             // flag to display the metrics in the logs. The metrics are still collect even if this flag is false.
}

// Creates a new watcher backend
func newWatcherBackend(mainCtx context.Context, requestChannel <-chan SubscribeRequest, client ethereum.Network, logger *logrus.Logger, database *db.Database, metricsDisplay bool) *WatcherBackend {
	return &WatcherBackend{
		mainCtx:            mainCtx,
		requestChannel:     requestChannel,
		client:             client,
		logger:             logger.WithField("Component", "TransactionWatcherBackend"),
		database:           database,
		monitoredTxns:      make(map[common.Hash]info),
		receiptCache:       make(map[common.Hash]receipt),
		aggregates:         make(map[FuncSelector]Profile),
		retryGroups:        make(map[common.Hash]group),
		lastProcessedBlock: &block{0, common.HexToHash("")},
		metricsDisplay:     metricsDisplay,
	}
}

func (wb *WatcherBackend) LoadState() error {
	if err := wb.database.View(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTransactionWatcherState()
		wb.logger.WithField("Key", string(key)).Tracef("Looking up state")
		rawData, err := utils.GetValue(txn, key)
		if err != nil {
			return fmt.Errorf("failed to get value %v", err)
		}
		err = json.Unmarshal(rawData, wb)
		if err != nil {
			return fmt.Errorf("failed to unmarshal %v", err)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (wb *WatcherBackend) PersistState() error {
	rawData, err := json.Marshal(wb)
	if err != nil {
		return fmt.Errorf("failed to marshal %v", err)
	}
	err = wb.database.Update(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTransactionWatcherState()
		wb.logger.WithField("Key", string(key)).Tracef("Saving state")
		if err := utils.SetValue(txn, key, rawData); err != nil {
			return fmt.Errorf("failed to set Value %v", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// synchronizing db state to disk
	if err := wb.database.Sync(); err != nil {
		return fmt.Errorf("Failed to set sync %v", err)
	}
	return nil
}

func (wb *WatcherBackend) Loop() {
	poolingTime := time.After(constants.TxPollingTime)
	statusTime := time.After(constants.TxStatusTime)
	for {
		select {
		case req, ok := <-wb.requestChannel:
			if !ok {
				wb.logger.Debugf("request channel closed, exiting")
				return
			}
			if req.responseChannel == nil {
				wb.logger.Debug("Invalid request for txn without a response channel, ignoring")
				continue
			}
			resp, err := wb.queue(req)
			req.responseChannel.sendResponse(&SubscribeResponse{Err: err, Response: resp})

		case <-poolingTime:
			wb.collectReceipts()
			poolingTime = time.After(constants.TxPollingTime)
			err := wb.PersistState()
			if err != nil {
				wb.logger.Errorf("Failed to persist state on the database %v", err)
			}
		case <-statusTime:
			for selector, profile := range wb.aggregates {
				sig := bindings.FunctionMapping[selector]
				wb.logger.WithField("Selector", fmt.Sprintf("%x", selector)).
					WithField("Function", sig).
					WithField("Profile", fmt.Sprintf("%+v", profile)).
					Info("Status")
			}
			statusTime = time.After(constants.TxStatusTime)
		}
	}
}

func (wb *WatcherBackend) queue(req SubscribeRequest) (*ReceiptResponse, error) {
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
		var txGroupHash common.Hash
		if _, ok = wb.monitoredTxns[txnHash]; ok {
			txGroupHash = wb.monitoredTxns[txnHash].retryGroup
		} else if _, ok = wb.retryGroups[txnHash]; ok {
			txGroupHash = txnHash
		} else {
			selector, err := ExtractSelector(req.txn.Data())
			if err != nil {
				return nil, &ErrInvalidTransactionRequest{
					fmt.Sprintf("invalid request, transaction data is not present %v, err %v!", txnHash.Hex(), err),
				}
			}
			sig := bindings.FunctionMapping[selector]

			logEntry := wb.logger.WithField("Transaction", txnHash).
				WithField("Function", sig).
				WithField("Selector", fmt.Sprintf("%x", selector))

			wb.monitoredTxns[txnHash] = info{
				txn:               req.txn,
				fromAddress:       fromAddr,
				selector:          selector,
				functionSignature: sig,
				retryGroup:        txnHash,
				disableAutoRetry:  req.disableAutoRetry,
				logger:            logEntry,
			}
			txGroup := newGroup()
			txGroup.add(txnHash)
			wb.retryGroups[txnHash] = txGroup
			logEntry.Debug("Transaction queued")
			txGroupHash = txnHash
		}
		receiptResponse = wb.retryGroups[txGroupHash].receiptResponse
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
		txInfo, ok := wb.monitoredTxns[workResponse.txnHash]
		logEntry := txInfo.logger
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
						txn:               workResponse.retriedTxn.txn,
						fromAddress:       txInfo.fromAddress,
						selector:          txInfo.selector,
						functionSignature: txInfo.functionSignature,
						retryGroup:        txInfo.retryGroup,
						disableAutoRetry:  txInfo.disableAutoRetry,
						logger:            txInfo.logger,
					}
					// update retry group
					txGroup := wb.retryGroups[txInfo.retryGroup]
					txGroup.add(newTxnHash)
					wb.retryGroups[txInfo.retryGroup] = txGroup
					logEntry.Tracef("successfully replaced a tx with %v", newTxnHash)
					txInfo.retryAmount++
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
				wb.receiptCache[txInfo.retryGroup] = receipt{receipt: workResponse.receipt, retrievedAtHeight: blockInfo.height}
				finishedTxs[workResponse.txnHash] = workResponse
			}
		}
		wb.monitoredTxns[workResponse.txnHash] = txInfo
	}

	// Cleaning finished and failed transactions
	for txnHash, workResponse := range finishedTxs {
		if txnInfo, ok := wb.monitoredTxns[txnHash]; ok {
			if txGroup, ok := wb.retryGroups[txnInfo.retryGroup]; ok {
				logger := txnInfo.logger.WithFields(logrus.Fields{
					"group": txnInfo.retryGroup,
				})

				if workResponse.receipt != nil {
					rcpt := workResponse.receipt
					var profile Profile
					if _, present := wb.aggregates[txnInfo.selector]; present {
						profile = wb.aggregates[txnInfo.selector]
					} else {
						profile = Profile{}
					}
					// Update transaction profile
					profile.AverageGas = (profile.AverageGas*profile.TotalCount + rcpt.GasUsed) / (profile.TotalCount + 1)
					if profile.MaximumGas < rcpt.GasUsed {
						profile.MaximumGas = rcpt.GasUsed
					}
					if profile.MinimumGas == 0 || profile.MinimumGas > rcpt.GasUsed {
						profile.MinimumGas = rcpt.GasUsed
					}
					profile.TotalCount++
					profile.TotalGas += rcpt.GasUsed
					if rcpt.Status == uint64(1) {
						profile.TotalSuccess++
					}
					wb.aggregates[txnInfo.selector] = profile
				}
				txGroup.sendReceipt(logger, workResponse.receipt, workResponse.err)
				err = txGroup.remove(txnHash)
				if err != nil {
					logger.Debugf("Failed to remove txn from group: %v", err)
				} else {
					if txGroup.isEmpty() {
						logger.Tracef("empty group removing")
						delete(wb.retryGroups, txnInfo.retryGroup)
					}
				}
			} else {
				txnInfo.logger.Debugf("Failed to find a group for txn")
			}
			delete(wb.monitoredTxns, txnHash)
		} else {
			wb.logger.Debugf("Failed to find txn to remove: %v", txnHash.Hex())
		}
	}

	var expiredReceipts []common.Hash
	// Marking expired receipts and restarting the height of state recovered receipts
	for receiptTxnHash, receiptInfo := range wb.receiptCache {
		if receiptInfo.retrievedAtHeight == 0 || receiptInfo.retrievedAtHeight > blockInfo.height {
			receiptInfo.retrievedAtHeight = blockInfo.height
		}
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
	requestChannel   chan<- SubscribeRequest // channel used to send request to the backend service to retrieve transactions
}

var _ IWatcher = &Watcher{}

// Creates a new transaction watcher struct
func NewWatcher(client ethereum.Network, txConfirmationBlocks uint64, database *db.Database, statusDisplay bool) *Watcher {
	requestChannel := make(chan SubscribeRequest, 100)
	// main context that will cancel all workers and go routine
	mainCtx, cf := context.WithCancel(context.Background())

	logger := logging.GetLogger("transaction")

	backend := newWatcherBackend(mainCtx, requestChannel, client, logger, database, statusDisplay)

	transactionWatcher := &Watcher{
		requestChannel:   requestChannel,
		closeMainContext: cf,
		backend:          backend,
		logger:           logger.WithField("Component", "TransactionWatcher"),
	}
	return transactionWatcher
}

// WatcherFromNetwork creates a transaction Watcher from a given ethereum network.
func WatcherFromNetwork(network ethereum.Network, database *db.Database, statusDisplay bool) *Watcher {
	watcher := NewWatcher(network, network.GetFinalityDelay(), database, statusDisplay)
	watcher.Start()
	return watcher
}

// Start the transaction watcher service
func (f *Watcher) Start() error {
	err := f.backend.LoadState()
	if err != nil {
		f.logger.Tracef("could not find previous State: %v", err)
		if err != badger.ErrKeyNotFound {
			return fmt.Errorf("could not find previous State: %v", err)
		}
	}
	go f.backend.Loop()
	return nil
}

// Close the transaction watcher service
func (f *Watcher) Close() {
	f.logger.Debug("closing request channel...")
	close(f.requestChannel)
	f.closeMainContext()
}

// Subscribe a transaction to be watched by the transaction watcher service. If
// a transaction was accepted by the watcher service, a response struct is
// returned. The response struct is where the receipt going to be written once
// available. The final tx hash in the receipt can be different from the initial
// txn sent. This can happen if the txn got stale and the watcher did a
// transaction replace with higher fees.
func (w *Watcher) Subscribe(ctx context.Context, txn *types.Transaction, disableAutoRetry bool) (IReceiptResponse, error) {
	w.logger.WithField("Txn", txn.Hash().Hex()).Debug("Subscribing for a transaction")
	req := newSubscribeRequest(txn, disableAutoRetry)
	select {
	case w.requestChannel <- req:
	case <-ctx.Done():
		return nil, &ErrInvalidTransactionRequest{fmt.Sprintf("context cancelled reqChannel: %v", ctx.Err())}
	}
	return req.Listen(ctx)
}

// function that wait for a transaction receipt. This is blocking function that
// will wait for a receipt to be received
func (w *Watcher) Wait(ctx context.Context, receiptResponse IReceiptResponse) (*types.Receipt, error) {
	return receiptResponse.GetReceipt(ctx)
}

// Queue a transaction and wait for its receipt
func (w *Watcher) SubscribeAndWait(ctx context.Context, txn *types.Transaction, disableAutoRetry bool) (*types.Receipt, error) {
	receiptResponse, err := w.Subscribe(ctx, txn, disableAutoRetry)
	if err != nil {
		return nil, err
	}
	return w.Wait(ctx, receiptResponse)
}

func ExtractSelector(data []byte) (FuncSelector, error) {
	var selector [4]byte
	if len(data) < 4 {
		return selector, fmt.Errorf("couldn't extract selector for data: %v", data)
	}
	for idx := 0; idx < 4; idx++ {
		selector[idx] = data[idx]
	}
	return selector, nil
}
