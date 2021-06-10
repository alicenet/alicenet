package dkgtasks

import (
	"context"
	"math/big"
	"sync"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

// DisputeTask stores the data required to dispute shares
type MPKSubmissionTask struct {
	sync.Mutex
	OriginalRegistrationEnd uint64
	State                   *objects.DkgState
}

// NewDisputeTask creates a new task
func NewMPKSubmissionTask(state *objects.DkgState) *MPKSubmissionTask {
	return &MPKSubmissionTask{
		OriginalRegistrationEnd: state.RegistrationEnd, // If these quit being equal, this task should be abandoned
		State:                   state,
	}
}

// MPKSubmissionTask contains required state for safely performing a registration
// type MPKSubmissionTask struct {
// 	sync.Mutex
// 	Account         accounts.Account
// 	RegistrationEnd uint64
// 	LastBlock       uint64
// 	PublicKey       [2]*big.Int
// 	MasterPublicKey [4]*big.Int
// }

// // NewMPKSubmissionTask creates a background task that attempts to register with ETHDKG
// func NewMPKSubmissionTask(
// 	acct accounts.Account,
// 	publicKey [2]*big.Int,
// 	masterPublicKey [4]*big.Int,
// 	registrationEnd uint64, lastBlock uint64) *MPKSubmissionTask {
// 	return &MPKSubmissionTask{
// 		Account:         acct,
// 		PublicKey:       blockchain.CloneBigInt2(publicKey),
// 		MasterPublicKey: blockchain.CloneBigInt4(masterPublicKey),
// 		RegistrationEnd: registrationEnd,
// 		LastBlock:       lastBlock,
// 	}
// }

func (t *MPKSubmissionTask) init(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {

	state := t.State

	if state.MasterPublicKey[0] == nil {
		g1KeyShares := make([][2]*big.Int, state.NumberOfValidators)
		g2KeyShares := make([][4]*big.Int, state.NumberOfValidators)

		for idx, participant := range state.Participants {
			// Bringing these in from state but could just directly query contract
			g1KeyShares[idx] = state.KeyShareG1s[participant.Address]
			g2KeyShares[idx] = state.KeyShareG2s[participant.Address]
		}

		mpk, err := math.GenerateMasterPublicKey(g1KeyShares, g2KeyShares)
		if err != nil {
			logger.Errorf("Failed to generate master public key:%v", err)
			return false
		}

		// Master public key is all we generate here so save it
		state.MasterPublicKey = mpk
	}

	return true
}

// DoWork is the first attempt at registering with ethdkg
func (t *MPKSubmissionTask) DoWork(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	logger.Info("DoWork() ...")
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at registering with ethdkg
func (t *MPKSubmissionTask) DoRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	logger.Info("DoRetry() ...")
	return t.doTask(ctx, logger, eth)
}

func (t *MPKSubmissionTask) doTask(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	t.Lock()
	defer t.Unlock()

	if !t.init(ctx, logger, eth) {
		return false
	}

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		logger.Errorf("getting txn opts failed: %v", err)
		return false
	}

	// Register
	logger.Infof("submitting master public key:%v", t.State.MasterPublicKey)
	txn, err := eth.Contracts().Ethdkg().SubmitMasterPublicKey(txnOpts, t.State.MasterPublicKey)
	if err != nil {
		logger.Errorf("submitting master public key failed: %v", err)
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
		logger.Error("missing registration receipt")
		return false
	}

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		logger.Errorf("registration status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
		return false
	}

	return true
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
// -- the registration open hasn't moved, i.e. ETHDKG has not restarted
func (t *MPKSubmissionTask) ShouldRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	t.Lock()
	defer t.Unlock()

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, t.State.Account, logger, eth,
		t.State.TransportPublicKey, t.OriginalRegistrationEnd, t.State.MPKSubmissionEnd)
}

// DoDone creates a log entry saying task is complete
func (t *MPKSubmissionTask) DoDone(logger *logrus.Logger) {
	logger.Infof("done")
}
