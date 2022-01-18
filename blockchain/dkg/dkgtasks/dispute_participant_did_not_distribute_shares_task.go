package dkgtasks

import (
	"context"
	"fmt"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

// DisputeParticipantDidNotDistributeSharesTask stores the data required to dispute shares
type DisputeParticipantDidNotDistributeSharesTask struct {
	OriginalRegistrationEnd uint64
	State                   *objects.DkgState
	Success                 bool
}

// NewDisputeParticipantDidNotDistributeSharesTask creates a new task
func NewDisputeParticipantDidNotDistributeSharesTask(state *objects.DkgState) *DisputeParticipantDidNotDistributeSharesTask {
	return &DisputeParticipantDidNotDistributeSharesTask{
		OriginalRegistrationEnd: state.RegistrationEnd, // If these quit being equal, this task should be abandoned
		State:                   state,
	}
}

// Initialize begins the setup phase for DisputeParticipantDidNotDistributeSharesTask.
// It determines if the shares previously distributed are valid.
// If any are invalid, disputes will be issued.
func (t *DisputeParticipantDidNotDistributeSharesTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	dkgState, validState := state.(*objects.DkgState)
	if !validState {
		panic(fmt.Errorf("%w invalid state type", objects.ErrCanNotContinue))
	}

	t.State = dkgState

	logger.Info("Initializing DisputeParticipantDidNotDistributeSharesTask...")

	return nil
}

// DoWork is the first attempt at disputing distributed shares
func (t *DisputeParticipantDidNotDistributeSharesTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is subsequent attempts at disputing distributed shares
func (t *DisputeParticipantDidNotDistributeSharesTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *DisputeParticipantDidNotDistributeSharesTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("Working DisputeParticipantDidNotDistributeSharesTask...")

	t.Success = true
	return nil
}

// ShouldRetry checks if it makes sense to try again
func (t *DisputeParticipantDidNotDistributeSharesTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.State.Lock()
	defer t.State.Unlock()

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, t.State.Account, logger, eth,
		t.State.TransportPublicKey, t.OriginalRegistrationEnd, t.State.DisputeShareDistributionEnd)
}

// DoDone creates a log entry saying task is complete
func (t *DisputeParticipantDidNotDistributeSharesTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Info("done")

	t.State.DisputeShareDistribution = t.Success
	t.State.Phase = objects.DisputeShareDistribution
}
