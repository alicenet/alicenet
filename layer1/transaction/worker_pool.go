package transaction

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/layer1"
	goEthereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// MonitorWorkRequest is an internal struct used to send work requests to the
// workers that will retrieve the receipts
type MonitorWorkRequest struct {
	txn    info   // Info object that contains the state that will be used to retrieve the receipt from the blockchain
	height uint64 // Current height of the blockchain head
}

// MonitorWorkResponse is an internal struct used by the workers to communicate the result
// from the receipt retrieval work
type MonitorWorkResponse struct {
	txnHash    common.Hash         // hash of transaction
	retriedTxn *retriedTransaction // transaction info object from the analyzed transaction
	err        error               // any error found during the receipt retrieve (can be NonRecoverable, Recoverable or TransactionStale errors)
	receipt    *types.Receipt      // receipt retrieved (can be nil) if a receipt was not found or it's not ready yet
}

// retriedTransaction is an internal struct to keep track of retried transaction by the workers
type retriedTransaction struct {
	txn *types.Transaction // new transaction after the retry attempt
	err error              // error that happened during the transaction retry
}

// WorkerPool is a Struct that keep track of the state needed by the worker pool service.
// The WorkerPool spawn multiple go routines (workers) to check and retrieve the
// receipts.
type WorkerPool struct {
	wg                  *sync.WaitGroup
	ctx                 context.Context            // Main context passed by the parent, used to cancel worker and the pool
	client              layer1.Client              // An interface with the Geth functionality we need
	baseFee             *big.Int                   // Base fee of the current block in case we need to retry a stale transaction
	tipCap              *big.Int                   // Miner tip cap of the current block in case we need to retry a stale transaction
	logger              *logrus.Entry              // Logger to log messages
	requestWorkChannel  <-chan MonitorWorkRequest  // Channel where will be sent the work requests
	responseWorkChannel chan<- MonitorWorkResponse // Channel where the work response will be sent
}

// NewWorkerPool creates a new WorkerPool service
func NewWorkerPool(ctx context.Context, client layer1.Client, baseFee *big.Int, tipCap *big.Int, logger *logrus.Entry, requestWorkChannel <-chan MonitorWorkRequest, responseWorkChannel chan<- MonitorWorkResponse) *WorkerPool {
	return &WorkerPool{new(sync.WaitGroup), ctx, client, baseFee, tipCap, logger, requestWorkChannel, responseWorkChannel}
}

// ExecuteWork is a function to spawn the workers and wait for the job to be done.
func (w *WorkerPool) ExecuteWork(numWorkers uint64) {
	for i := uint64(0); i < numWorkers; i++ {
		w.wg.Add(1)
		go w.worker()
	}
	w.wg.Wait()
	close(w.responseWorkChannel)
}

// worker is a unit of work. A worker is spawned as go routine. A worker check and retrieve
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
			txnHash := monitoredTx.Txn.Hash()
			for i := uint64(1); i <= constants.TxWorkerMaxWorkRetries; i++ {
				select {
				case <-ctx.Done():
					// worker context timed out or parent was cancelled, should return
					return
				default:
				}
				rcpt, err := w.getReceipt(ctx, monitoredTx, currentHeight, txnHash)
				finalResp, retry := w.handleResponse(ctx, monitoredTx, txnHash, rcpt, err, i)
				if !retry {
					select {
					case w.responseWorkChannel <- finalResp:
					default:
					}
					break
				}
			}
		}
	}
}

// handleResponse handles the receipt for the txn and decides how to proceed based on it information
func (w *WorkerPool) handleResponse(ctx context.Context, monitoredTx info, txnHash common.Hash, rcpt *types.Receipt, err error, iteration uint64) (MonitorWorkResponse, bool) {
	if err != nil {
		switch err.(type) {
		case *ErrRecoverable:
			// retry on recoverable error `constants.TxWorkerMaxWorkRetries` times
			if iteration < constants.TxWorkerMaxWorkRetries {
				return MonitorWorkResponse{}, true
			}
		case *ErrTransactionStale:
			// try to replace a transaction if the conditions are met
			if monitoredTx.EnableAutoRetry {
				defaultAccount := w.client.GetDefaultAccount()
				if bytes.Equal(monitoredTx.FromAddress[:], defaultAccount.Address[:]) {
					newTxn, retryTxErr := w.client.RetryTransaction(ctx, monitoredTx.Txn, w.baseFee, w.tipCap)
					return MonitorWorkResponse{txnHash: txnHash, retriedTxn: &retriedTransaction{txn: newTxn, err: retryTxErr}}, false
				}
			}
		default:
		}
		// send recoverable errors after constants.TxWorkerMaxWorkRetries, txNotFound or
		// other errors back to main
		return MonitorWorkResponse{txnHash: txnHash, err: err}, false
	} else {
		// send receipt (even if it is nil) back to main thread
		return MonitorWorkResponse{txnHash: txnHash, receipt: rcpt}, false
	}
}

// getReceipt is an internal function used by the workers to check/retrieve the receipts for a given transaction
func (w *WorkerPool) getReceipt(ctx context.Context, monitoredTx info, currentHeight uint64, txnHash common.Hash) (*types.Receipt, error) {
	txnHex := txnHash.Hex()
	blockTimeSpan := currentHeight - monitoredTx.MonitoringHeight
	_, isPending, err := w.client.GetTransactionByHash(ctx, txnHash)
	if err != nil {
		// if we couldn't locate a tx after NotFoundMaxBlocks blocks, and we are still
		// failing in getting the tx data, probably means that it was dropped
		if errors.Is(err, goEthereum.NotFound) {
			return nil, &ErrTxNotFound{fmt.Sprintf("could not find tx %v in the height %v!", txnHex, currentHeight)}
		}
		// probably a network error, should retry
		return nil, &ErrRecoverable{fmt.Sprintf("error getting tx: %v: %v", txnHex, err)}
	}
	if isPending {
		// We multiply MaxStaleBlocks by the number of times that we tried to retry a tx
		// to add an increasing delay between successful retry attempts.
		// startedMonitoringHeight is restarted at every retry attempt. Most of the time
		// after a successful retry, the tx being replaced will fall in the branch above
		// (err tx not found). But in case of an edge case, where tx replacing and tx
		// replaced are both valid (e.g. sending tx to different nodes) we will continue
		// to retry both, until we have a valid tx for this nonce.
		maxPendingBlocks := monitoredTx.MaxStaleBlocks * (monitoredTx.RetryAmount + 1)
		// after first retry we increase the delay between retries
		if monitoredTx.RetryAmount > 0 {
			maxPendingBlocks *= constants.TxBackOffDelayStaleTxMultiplier
		}
		if blockTimeSpan >= maxPendingBlocks {
			return nil, &ErrTransactionStale{fmt.Sprintf("error tx: %v is stale on the memory pool for more than %v blocks!", txnHex, maxPendingBlocks)}
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
