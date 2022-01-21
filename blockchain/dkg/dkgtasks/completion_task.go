package dkgtasks

import (
	"context"
	"fmt"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

// CompletionTask contains required state for safely performing a registration
type CompletionTask struct {
	Start   uint64
	End     uint64
	State   *objects.DkgState
	Success bool
}

// asserting that CompletionTask struct implements interface interfaces.Task
var _ interfaces.Task = &CompletionTask{}

// NewCompletionTask creates a background task that attempts to call Complete on ethdkg
func NewCompletionTask(state *objects.DkgState, start uint64, end uint64) *CompletionTask {
	return &CompletionTask{
		Start:   start,
		End:     end,
		State:   state,
		Success: false,
	}
}

// Initialize prepares for work to be done in the Completion phase
func (t *CompletionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {

	dkgState, validState := state.(*objects.DkgState)
	if !validState {
		panic(fmt.Errorf("%w invalid state type", objects.ErrCanNotContinue))
	}

	t.State = dkgState

	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("StateLocation", fmt.Sprintf("%p", t.State)).Info("CompletionTask Initialize()...")

	if t.State.Phase != objects.DisputeGPKJSubmission {
		return fmt.Errorf("%w because it's not in DisputeGPKJSubmission phase", objects.ErrCanNotContinue)
	}

	return nil
}

// DoWork is the first attempt
func (t *CompletionTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts
func (t *CompletionTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *CompletionTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("CompletionTask doTask()")

	// Setup
	c := eth.Contracts()
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// Register
	txn, err := c.Ethdkg().Complete(txnOpts)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "completion failed: %v", err)
	}

	logger.Info("CompletionTask sent completed call")

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

	logger.Info("CompletionTask complete!")

	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
// -- the registration open hasn't moved, i.e. ETHDKG has not restarted
func (t *CompletionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("CompletionTask ShouldRetry()")

	if t.State.Phase != objects.DisputeGPKJSubmission ||
		t.State.Phase == objects.Completion {
		logger.WithFields(logrus.Fields{
			"t.State.Phase":      t.State.Phase,
			"t.State.PhaseStart": t.State.PhaseStart,
		}).Info("CompletionTask ShouldRetry - will not retry")
		return false
	}

	var shouldRetry bool = GeneralTaskShouldRetry(ctx, t.State.Account, logger, eth,
		t.State.TransportPublicKey, t.Start, t.End)

	logger.WithFields(logrus.Fields{
		"shouldRetry": shouldRetry,
		"t.Start":     t.Start,
		"t.End":       t.End,
	}).Info("CompletionTask ShouldRetry")

	// This wraps the retry logic for every phase, _except_ registration
	return shouldRetry
}

// DoDone creates a log entry saying task is complete
func (t *CompletionTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("CompletionTask done")
}
