package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/logging"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

type ReceiptResponse interface {
	IsReady() bool
	GetReceiptBlocking(ctx context.Context) (*types.Receipt, error)
}

type Watcher interface {
	Start() error
	Close()
	Subscribe(ctx context.Context, txn *types.Transaction, options *SubscribeOptions) (ReceiptResponse, error)
	Wait(ctx context.Context, receiptResponse ReceiptResponse) (*types.Receipt, error)
	SubscribeAndWait(ctx context.Context, txn *types.Transaction, options *SubscribeOptions) (*types.Receipt, error)
}

type SubscribeOptions struct {
	EnableAutoRetry bool   // if we should disable auto retry of a transaction in case it becomes stale
	MaxStaleBlocks  uint64 // how many blocks we should consider a transaction stale and mark it for retry
}

func NewSubscribeOptions(enableAutoRetry bool, maxStaleBlocks uint64) *SubscribeOptions {
	return &SubscribeOptions{EnableAutoRetry: enableAutoRetry, MaxStaleBlocks: maxStaleBlocks}
}

// Struct that has the data necessary by the Transaction FrontWatcher service. The
// transaction watcher service is responsible for check, retrieve and cache
// transaction receipts.
type FrontWatcher struct {
	backend          *WatcherBackend         // backend service responsible for check, retrieving and caching the receipts
	logger           *logrus.Entry           // logger used to log the message for the transaction watcher
	closeMainContext context.CancelFunc      // function used to cancel the main context in the backend service
	requestChannel   chan<- SubscribeRequest // channel used to send request to the backend service to retrieve transactions
}

var _ Watcher = &FrontWatcher{}

// Creates a new transaction watcher struct
func NewWatcher(client layer1.Client, txConfirmationBlocks uint64, database *db.Database, statusDisplay bool, txPollingTime time.Duration) *FrontWatcher {
	requestChannel := make(chan SubscribeRequest, 100)
	// main context that will cancel all workers and go routine
	mainCtx, cf := context.WithCancel(context.Background())

	logger := logging.GetLogger("transaction")

	logger.Info("Creating transaction watcher")

	backend := newWatcherBackend(mainCtx, requestChannel, client, logger, database, statusDisplay, txPollingTime)

	transactionWatcher := &FrontWatcher{
		requestChannel:   requestChannel,
		closeMainContext: cf,
		backend:          backend,
		logger:           logger.WithField("Component", "TransactionWatcher"),
	}
	return transactionWatcher
}

// WatcherFromNetwork creates a transaction Watcher from a given ethereum network.
func WatcherFromNetwork(network layer1.Client, database *db.Database, statusDisplay bool, txPollingTime time.Duration) *FrontWatcher {
	watcher := NewWatcher(network, network.GetFinalityDelay(), database, statusDisplay, txPollingTime)
	err := watcher.Start()
	if err != nil {
		panic(fmt.Sprintf("couldn't start transaction watcher: %v", err))
	}
	return watcher
}

// Start the transaction watcher service
func (f *FrontWatcher) Start() error {
	err := f.backend.LoadState()
	if err != nil {
		f.logger.Tracef("could not find previous State: %v", err)
		if err != badger.ErrKeyNotFound {
			return fmt.Errorf("could not find previous State: %v", err)
		}
	}
	f.logger.Info("loaded state for transaction watcher")
	go f.backend.Loop()
	return nil
}

// Close the transaction watcher service
func (f *FrontWatcher) Close() {
	f.logger.Warn("Closing transaction watcher")
	close(f.requestChannel)
	f.closeMainContext()
}

// Subscribe a transaction to be watched by the transaction watcher service. If
// a transaction was accepted by the watcher service, a response struct is
// returned. The response struct is where the receipt going to be written once
// available. The final tx hash in the receipt can be different from the initial
// txn sent. This can happen if the txn got stale and the watcher did a
// transaction replace with higher fees.
func (w *FrontWatcher) Subscribe(ctx context.Context, txn *types.Transaction, options *SubscribeOptions) (ReceiptResponse, error) {
	w.logger.WithField("Txn", txn.Hash().Hex()).Debug("Subscribing for a transaction")
	req := NewSubscribeRequest(txn, options)
	select {
	case w.requestChannel <- req:
	case <-ctx.Done():
		return nil, &ErrInvalidTransactionRequest{fmt.Sprintf("context cancelled reqChannel: %v", ctx.Err())}
	}
	return req.Listen(ctx)
}

// function that wait for a transaction receipt. This is blocking function that
// will wait for a receipt to be received
func (w *FrontWatcher) Wait(ctx context.Context, receiptResponse ReceiptResponse) (*types.Receipt, error) {
	return receiptResponse.GetReceiptBlocking(ctx)
}

// Queue a transaction and wait for its receipt
func (w *FrontWatcher) SubscribeAndWait(ctx context.Context, txn *types.Transaction, options *SubscribeOptions) (*types.Receipt, error) {
	receiptResponse, err := w.Subscribe(ctx, txn, options)
	if err != nil {
		return nil, err
	}
	return w.Wait(ctx, receiptResponse)
}
