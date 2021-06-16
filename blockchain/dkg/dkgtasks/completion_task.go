package dkgtasks

import (
	"context"
	"sync"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

// CompletionTask contains required state for safely performing a registration
type CompletionTask struct {
	sync.Mutex
	OriginalRegistrationEnd uint64
	State                   *objects.DkgState
	Success                 bool
}

// NewCompletionTask creates a background task that attempts to call Complete on ethdkg
func NewCompletionTask(state *objects.DkgState) *CompletionTask {
	return &CompletionTask{
		OriginalRegistrationEnd: state.RegistrationEnd,
		State:                   state,
	}
}

func (t *CompletionTask) Initialize(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {
	return nil
}

// DoWork is the first attempt
func (t *CompletionTask) DoWork(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts
func (t *CompletionTask) DoRetry(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *CompletionTask) doTask(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {

	t.Lock()
	defer t.Unlock()

	// Setup
	c := eth.Contracts()
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// Register
	txn, err := c.Ethdkg().SuccessfulCompletion(txnOpts)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "completion failed: %v", err)
	}

	logger.Info("Completion completed")

	// Waiting for receipt
	receipt, err := eth.Queue().QueueAndWait(ctx, txn)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "waiting for receipt failed: %v", err)
	}
	if receipt == nil {
		return dkg.LogReturnErrorf(logger, "missing completion receipt")
	}

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		return dkg.LogReturnErrorf(logger, "completion status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
	}

	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
// -- the registration open hasn't moved, i.e. ETHDKG has not restarted
func (t *CompletionTask) ShouldRetry(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) bool {

	t.Lock()
	defer t.Unlock()

	state := t.State

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, state.Account, logger, eth,
		state.TransportPublicKey, t.OriginalRegistrationEnd, state.CompleteEnd)
}

// DoDone creates a log entry saying task is complete
func (t *CompletionTask) DoDone(logger *logrus.Logger) {
	logger.Infof("done")

	t.State.Lock()
	defer t.State.Unlock()

	t.State.Complete = t.Success
}
