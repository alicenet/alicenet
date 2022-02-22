package dkgtasks

import (
	"context"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

// KeyshareSubmissionTask is the task for submitting Keyshare information
type KeyshareSubmissionTask struct {
	Start   uint64
	End     uint64
	State   *objects.DkgState
	Success bool
}

// asserting that KeyshareSubmissionTask struct implements interface interfaces.Task
var _ interfaces.Task = &KeyshareSubmissionTask{}

// NewKeyshareSubmissionTask creates a new task
func NewKeyshareSubmissionTask(state *objects.DkgState, start uint64, end uint64) *KeyshareSubmissionTask {
	return &KeyshareSubmissionTask{
		Start:   start,
		End:     end,
		State:   state,
		Success: false,
	}
}

// Initialize prepares for work to be done in KeyShareSubmission phase.
// Here, the G1 key share, G1 proof, and G2 key share are constructed
// and stored for submission.
func (t *KeyshareSubmissionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("KeyshareSubmissionTask Initialize()")

	// Generate the key shares
	g1KeyShare, g1Proof, g2KeyShare, err := math.GenerateKeyShare(t.State.SecretValue)
	if err != nil {
		return err
	}

	me := t.State.Account.Address

	t.State.Participants[me].KeyShareG1s = g1KeyShare
	t.State.Participants[me].KeyShareG1CorrectnessProofs = g1Proof
	t.State.Participants[me].KeyShareG2s = g2KeyShare

	return nil
}

// DoWork is the first attempt at the performing the KeyShareSubmission phase
func (t *KeyshareSubmissionTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	logger.Info("DoWork() ...")
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at the performing the KeyShareSubmission phase
func (t *KeyshareSubmissionTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	logger.Info("DoRetry() ...")
	return t.doTask(ctx, logger, eth)
}

func (t *KeyshareSubmissionTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("KeyshareSubmissionTask doTask()")

	// Setup
	me := t.State.Account

	txnOpts, err := eth.GetTransactionOpts(ctx, me)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// Submit Keyshares
	logger.Infof("submitting key shares: %v %v %v %v",
		me.Address,
		t.State.Participants[me.Address].KeyShareG1s,
		t.State.Participants[me.Address].KeyShareG1CorrectnessProofs,
		t.State.Participants[me.Address].KeyShareG2s)
	txn, err := eth.Contracts().Ethdkg().SubmitKeyShare(txnOpts,
		t.State.Participants[me.Address].KeyShareG1s,
		t.State.Participants[me.Address].KeyShareG1CorrectnessProofs,
		t.State.Participants[me.Address].KeyShareG2s)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "submitting keyshare failed: %v", err)
	}

	//TODO: add retry logic, add timeout

	// Waiting for receipt
	receipt, err := eth.Queue().QueueAndWait(ctx, txn)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "waiting for receipt failed: %v", err)
	}
	if receipt == nil {
		return dkg.LogReturnErrorf(logger, "missing submit keyshare receipt")
	}

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		return dkg.LogReturnErrorf(logger, "submit keyshare status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
	}

	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
func (t *KeyshareSubmissionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("KeyshareSubmissionTask ShouldRetry()")

	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	state := t.State

	me := state.Account
	callOpts := eth.GetCallOpts(ctx, me)

	phase, err := eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	if err != nil {
		logger.Infof("KeyshareSubmissionTask ShouldRetry GetETHDKGPhase error: %v", err)
		return true
	}

	// DisputeShareDistribution || KeyShareSubmission
	if phase != uint8(objects.DisputeShareDistribution) && phase != uint8(objects.KeyShareSubmission) {
		return false
	}

	// Check the key share submission status
	status, err := CheckKeyShare(ctx, eth.Contracts().Ethdkg(), logger, callOpts, me.Address, state.Participants[me.Address].KeyShareG1s)
	if err != nil {
		logger.Errorf("KeyshareSubmissionTask ShouldRetry CheckKeyShare error: %v", err)
		return true
	}

	if status == KeyShared || status == BadKeyShared {
		return false
	}

	return true
}

// DoDone creates a log entry saying task is complete
func (t *KeyshareSubmissionTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("KeyshareSubmissionTask done")
}
