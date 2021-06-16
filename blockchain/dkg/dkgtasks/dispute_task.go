package dkgtasks

import (
	"context"
	"sync"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

// DisputeTask stores the data required to dispute shares
type DisputeTask struct {
	sync.Mutex
	OriginalRegistrationEnd uint64
	State                   *objects.DkgState
	Success                 bool
}

// NewDisputeTask creates a new task
func NewDisputeTask(state *objects.DkgState) *DisputeTask {
	return &DisputeTask{
		OriginalRegistrationEnd: state.RegistrationEnd, // If these quit being equal, this task should be abandoned
		State:                   state,
	}
}

// This is not exported and does not lock so can only be called from within task. Return value indicates whether task has been initialized.
func (t *DisputeTask) Initialize(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {
	// TODO Figure out the best place for this function to be invoked
	return nil
}

// DoWork is the first attempt at distributing shares via ethdkg
func (t *DisputeTask) DoWork(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {
	logger.Info("DoWork() ...")
	return t.doTask(ctx, logger, eth)
}

// DoRetry is subsequent attempts at distributing shares via ethdkg
func (t *DisputeTask) DoRetry(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {
	logger.Info("DoRetry() ...")
	return t.doTask(ctx, logger, eth)
}

func (t *DisputeTask) doTask(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {

	t.Lock()
	defer t.Unlock()

	// TODO Implement
	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
func (t *DisputeTask) ShouldRetry(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) bool {

	t.Lock()
	defer t.Unlock()

	state := t.State

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, state.Account, logger, eth,
		state.TransportPublicKey, t.OriginalRegistrationEnd, state.DisputeEnd)
}

// DoDone creates a log entry saying task is complete
func (t *DisputeTask) DoDone(logger *logrus.Logger) {
	logger.Infof("done")
}
