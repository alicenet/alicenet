package dkgtasks

import (
	"context"
	"math/big"
	"sync"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// DisputeTask stores the data required to dispute shares
type KeyshareSubmissionTask struct {
	sync.Mutex
	OriginalRegistrationEnd uint64
	State                   *objects.DkgState
}

// NewKeyshareSubmissionTask creates a new task
func NewKeyshareSubmissionTask(state *objects.DkgState) *KeyshareSubmissionTask {
	return &KeyshareSubmissionTask{
		OriginalRegistrationEnd: state.RegistrationEnd, // If these quit being equal, this task should be abandoned
		State:                   state,
	}
}

// This is not exported and does not lock so can only be called from within task. Return value indicates whether task has been initialized.
func (t *KeyshareSubmissionTask) init(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	if t.State.KeyShareG1s == nil {

		// Generate the key shares
		g1KeyShare, g1Proof, g2KeyShare, err := math.GenerateKeyShare(t.State.SecretValue)
		if err != nil {
			return false
		}

		// t.State.KeyShareG1s[state.Account.Address]
		me := t.State.Account.Address

		t.State.KeyShareG1s = make(map[common.Address][2]*big.Int)
		t.State.KeyShareG1s[me] = g1KeyShare
		t.State.KeyShareG1CorrectnessProofs = make(map[common.Address][2]*big.Int)
		t.State.KeyShareG1CorrectnessProofs[me] = g1Proof
		t.State.KeyShareG2s = make(map[common.Address][4]*big.Int)
		t.State.KeyShareG2s[me] = g2KeyShare
	}

	return true
}

// DoWork is the first attempt at registering with ethdkg
func (t *KeyshareSubmissionTask) DoWork(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	logger.Info("DoWork() ...")
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at registering with ethdkg
func (t *KeyshareSubmissionTask) DoRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	logger.Info("DoRetry() ...")
	return t.doTask(ctx, logger, eth)
}

func (t *KeyshareSubmissionTask) doTask(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	t.Lock()
	defer t.Unlock()

	// Is there any point in running? Make sure we're both initialized and within block range
	if !t.init(ctx, logger, eth) {
		return false
	}

	// Setup
	me := t.State.Account

	txnOpts, err := eth.GetTransactionOpts(ctx, me)
	if err != nil {
		logger.Errorf("getting txn opts failed: %v", err)
		return false
	}

	// Submit Keyshares
	txn, err := eth.Contracts().Ethdkg().SubmitKeyShare(txnOpts, me.Address,
		t.State.KeyShareG1s[me.Address],
		t.State.KeyShareG1CorrectnessProofs[me.Address],
		t.State.KeyShareG2s[me.Address])
	if err != nil {
		logger.Errorf("submitting keyshare failed: %v", err)
		return false
	}
	eth.Queue().QueueTransaction(ctx, txn)

	// Waiting for receipt
	receipt, err := eth.Queue().WaitTransaction(ctx, txn)
	if err != nil {
		logger.Errorf("waiting for receipt failed: %v", err)
		return false
	}
	if receipt == nil {
		logger.Error("missing submit keyshare receipt")
		return false
	}

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		logger.Errorf("submit keyshare status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
		return false
	}

	return true
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
// -- the registration open hasn't moved, i.e. ETHDKG has not restarted
func (t *KeyshareSubmissionTask) ShouldRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	t.Lock()
	defer t.Unlock()

	state := t.State

	// This wraps the retry logic for the general case
	generalRetry := GeneralTaskShouldRetry(ctx, state.Account, logger, eth,
		state.TransportPublicKey, t.OriginalRegistrationEnd, state.KeyShareSubmissionEnd)

	// If it's generally good to retry, let's try to be more specific
	if generalRetry {
		me := state.Account
		callOpts := eth.GetCallOpts(ctx, me)
		status, err := CheckKeyShare(ctx, eth.Contracts().Ethdkg(), logger, callOpts, me.Address, state.KeyShareG1s[me.Address])
		if err != nil {
			return true
		}
		logger.Infof("Key shared status: %v", status)

		// If we have already shared a key, there is no reason to retry. Regardless of whether it's right or wrong.
		if status == KeyShared || status == BadKeyShared {
			return false
		}
	}

	return generalRetry
}

// DoDone creates a log entry saying task is complete
func (t *KeyshareSubmissionTask) DoDone(logger *logrus.Logger) {
	logger.Infof("done")
}
