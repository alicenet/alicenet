package dkgtasks

import (
	"context"
	"math/big"
	"sync"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/sirupsen/logrus"
)

// CompletionTask contains required state for safely performing a registration
type CompletionTask struct {
	sync.Mutex
	publicKey       [2]*big.Int
	registrationEnd uint64
	lastBlock       uint64
}

// NewCompletionTask creates a background task that attempts to call Complete on ethdkg
func NewCompletionTask(publicKey [2]*big.Int, registrationEnd uint64, lastBlock uint64) *CompletionTask {
	return &CompletionTask{
		publicKey: blockchain.CloneBigInt2(publicKey),
		lastBlock: lastBlock,
	}
}

// DoWork is the first attempt
func (t *CompletionTask) DoWork(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts
func (t *CompletionTask) DoRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	return t.doTask(ctx, logger, eth)
}

func (t *CompletionTask) doTask(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {

	t.Lock()
	defer t.Unlock()

	// Setup
	c := eth.Contracts()
	txnOpts, err := eth.GetTransactionOpts(ctx, eth.GetDefaultAccount())
	if err != nil {
		logger.Errorf("getting txn opts failed: %v", err)
		return false
	}

	// Register
	txn, err := c.Ethdkg.SuccessfulCompletion(txnOpts)
	if err != nil {
		logger.Errorf("completion failed: %v", err)
		return false
	}

	// Waiting for receipt
	receipt, err := eth.WaitForReceipt(ctx, txn)
	if err != nil {
		logger.Errorf("waiting for receipt failed: %v", err)
		return false
	}
	if receipt == nil {
		logger.Error("missing completion receipt")
		return false
	}

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		logger.Errorf("completion status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
		return false
	}

	return true
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
// -- the registration open hasn't moved, i.e. ETHDKG has not restarted
func (t *CompletionTask) ShouldRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {

	t.Lock()
	defer t.Unlock()

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, logger,
		eth, t.publicKey,
		t.registrationEnd, t.lastBlock)
}

// DoDone creates a log entry saying task is complete
func (t *CompletionTask) DoDone(logger *logrus.Logger) {
	logger.Infof("done")
}
