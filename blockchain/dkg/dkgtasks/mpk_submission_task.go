package dkgtasks

import (
	"context"
	"math/big"
	"sync"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/sirupsen/logrus"
)

// MPKSubmissionTask contains required state for safely performing a registration
type MPKSubmissionTask struct {
	sync.Mutex
	Account         accounts.Account
	RegistrationEnd uint64
	LastBlock       uint64
	PublicKey       [2]*big.Int
	MasterPublicKey [4]*big.Int
}

// NewMPKSubmissionTask creates a background task that attempts to register with ETHDKG
func NewMPKSubmissionTask(
	acct accounts.Account,
	publicKey [2]*big.Int,
	masterPublicKey [4]*big.Int,
	registrationEnd uint64, lastBlock uint64) *MPKSubmissionTask {
	return &MPKSubmissionTask{
		Account:         acct,
		PublicKey:       blockchain.CloneBigInt2(publicKey),
		MasterPublicKey: blockchain.CloneBigInt4(masterPublicKey),
		RegistrationEnd: registrationEnd,
		LastBlock:       lastBlock,
	}
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

	// Setup
	c := eth.Contracts()
	me := eth.GetDefaultAccount()
	txnOpts, err := eth.GetTransactionOpts(ctx, me)
	if err != nil {
		logger.Errorf("getting txn opts failed: %v", err)
		return false
	}

	// Register
	txn, err := c.Ethdkg.SubmitMasterPublicKey(txnOpts, t.MasterPublicKey)
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
	return GeneralTaskShouldRetry(ctx, t.Account, logger, eth,
		t.PublicKey, t.RegistrationEnd, t.LastBlock)
}

// DoDone creates a log entry saying task is complete
func (t *MPKSubmissionTask) DoDone(logger *logrus.Logger) {
	logger.Infof("done")
}
