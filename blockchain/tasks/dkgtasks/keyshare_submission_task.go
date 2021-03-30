package dkgtasks

import (
	"context"
	"math/big"
	"sync"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/sirupsen/logrus"
)

// KeyshareSubmissionTask contains required state for safely performing a registration
type KeyshareSubmissionTask struct {
	sync.Mutex
	registrationEnd uint64
	lastBlock       uint64
	publicKey       [2]*big.Int
	keyshareG1      [2]*big.Int
	keyshareG1Proof [2]*big.Int
	keyshareG2      [4]*big.Int
}

// NewKeyshareSubmissionTask creates a background task that attempts to register with ETHDKG
func NewKeyshareSubmissionTask(
	publicKey [2]*big.Int,
	keyshareG1 [2]*big.Int, keyshareG1Proof [2]*big.Int,
	keyshareG2 [4]*big.Int,
	registrationEnd uint64, lastBlock uint64) *KeyshareSubmissionTask {
	return &KeyshareSubmissionTask{
		publicKey:       blockchain.CloneBigInt2(publicKey),
		keyshareG1:      blockchain.CloneBigInt2(keyshareG1),
		keyshareG1Proof: blockchain.CloneBigInt2(keyshareG1Proof),
		keyshareG2:      blockchain.CloneBigInt4(keyshareG2),
		registrationEnd: registrationEnd,
		lastBlock:       lastBlock,
	}
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

	// Setup
	c := eth.Contracts()
	me := eth.GetDefaultAccount()
	txnOpts, err := eth.GetTransactionOpts(ctx, me)
	if err != nil {
		logger.Errorf("getting txn opts failed: %v", err)
		return false
	}

	// Submit Keyshares
	txn, err := c.Ethdkg.SubmitKeyShare(txnOpts, me.Address, t.keyshareG1, t.keyshareG1Proof, t.keyshareG2)
	if err != nil {
		logger.Errorf("submitting keyshare failed: %v", err)
		return false
	}

	// Waiting for receipt
	receipt, err := eth.WaitForReceipt(ctx, txn)
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

	// This wraps the retry logic for the general case
	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth,
		t.publicKey, t.registrationEnd, t.lastBlock)

	// If it's generally good to retry, let's try to be more specific
	if generalRetry {
		me := eth.GetDefaultAccount()
		callOpts := eth.GetCallOpts(ctx, me)
		status, err := CheckKeyShare(ctx, eth.Contracts().Ethdkg, logger, callOpts, me.Address, t.keyshareG1)
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
