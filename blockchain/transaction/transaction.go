package transaction

import (
	"context"
	"fmt"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/logging"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

type IReceiptResponse interface {
	IsReady() bool
	GetReceiptBlocking(ctx context.Context) (*types.Receipt, error)
}

type IWatcher interface {
	Start() error
	Close()
	Subscribe(ctx context.Context, txn *types.Transaction, disableAutoRetry bool) (IReceiptResponse, error)
	Wait(ctx context.Context, receiptResponse IReceiptResponse) (*types.Receipt, error)
	SubscribeAndWait(ctx context.Context, txn *types.Transaction, disableAutoRetry bool) (*types.Receipt, error)
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
	req := NewSubscribeRequest(txn, disableAutoRetry)
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
	return receiptResponse.GetReceiptBlocking(ctx)
}

// Queue a transaction and wait for its receipt
func (w *Watcher) SubscribeAndWait(ctx context.Context, txn *types.Transaction, disableAutoRetry bool) (*types.Receipt, error) {
	receiptResponse, err := w.Subscribe(ctx, txn, disableAutoRetry)
	if err != nil {
		return nil, err
	}
	return w.Wait(ctx, receiptResponse)
}
