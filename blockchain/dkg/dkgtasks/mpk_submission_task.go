package dkgtasks

import (
	"context"
	"math/big"
	"sync"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

// DisputeTask stores the data required to dispute shares
type MPKSubmissionTask struct {
	sync.Mutex
	OriginalRegistrationEnd uint64
	State                   *objects.DkgState
	Success                 bool
}

// NewDisputeTask creates a new task
func NewMPKSubmissionTask(state *objects.DkgState) *MPKSubmissionTask {
	return &MPKSubmissionTask{
		OriginalRegistrationEnd: state.RegistrationEnd, // If these quit being equal, this task should be abandoned
		State:                   state,
	}
}

func (t *MPKSubmissionTask) Initialize(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {

	state := t.State

	logger.Infof("MPK0==nil")
	g1KeyShares := make([][2]*big.Int, state.NumberOfValidators)
	g2KeyShares := make([][4]*big.Int, state.NumberOfValidators)

	for idx, participant := range state.Participants {

		// Bringing these in from state but could directly query contract
		g1KeyShares[idx] = state.KeyShareG1s[participant.Address]
		g2KeyShares[idx] = state.KeyShareG2s[participant.Address]

		logger.Infof("INIT idx:%v pidx:%v address:%v g1:%v g2:%v", idx, participant.Index, participant.Address.Hex(), g1KeyShares[idx], g2KeyShares[idx])
	}

	mpk, err := math.GenerateMasterPublicKey(g1KeyShares, g2KeyShares)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "Failed to generate master public key:%v", err)
	}

	// Master public key is all we generate here so save it
	state.MasterPublicKey = mpk

	return nil
}

// DoWork is the first attempt at registering with ethdkg
func (t *MPKSubmissionTask) DoWork(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {
	logger.Info("DoWork() ...")
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at registering with ethdkg
func (t *MPKSubmissionTask) DoRetry(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {
	logger.Info("DoRetry() ...")
	return t.doTask(ctx, logger, eth)
}

func (t *MPKSubmissionTask) doTask(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {
	t.Lock()
	defer t.Unlock()

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// Register
	logger.Infof("submitting master public key:%v", t.State.MasterPublicKey)
	txn, err := eth.Contracts().Ethdkg().SubmitMasterPublicKey(txnOpts, t.State.MasterPublicKey)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "submitting master public key failed: %v", err)
	}

	// Waiting for receipt
	receipt, err := eth.Queue().QueueAndWait(ctx, txn)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "waiting for receipt failed: %v", err)
	}
	if receipt == nil {
		return dkg.LogReturnErrorf(logger, "missing registration receipt")
	}

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		dkg.LogReturnErrorf(logger, "registration status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
	}
	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
// -- the registration open hasn't moved, i.e. ETHDKG has not restarted
func (t *MPKSubmissionTask) ShouldRetry(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) bool {
	t.Lock()
	defer t.Unlock()

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, t.State.Account, logger, eth,
		t.State.TransportPublicKey, t.OriginalRegistrationEnd, t.State.MPKSubmissionEnd)
}

// DoDone creates a log entry saying task is complete
func (t *MPKSubmissionTask) DoDone(logger *logrus.Logger) {
	logger.Infof("done")

	t.State.Lock()
	defer t.State.Unlock()

	t.State.MPKSubmission = t.Success
}
