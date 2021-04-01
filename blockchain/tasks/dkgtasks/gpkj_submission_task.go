package dkgtasks

import (
	"context"
	"math/big"
	"sync"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/sirupsen/logrus"
)

// GPKSubmissionTask contains required state for safely performing a registration
type GPKSubmissionTask struct {
	sync.Mutex
	Account         accounts.Account
	RegistrationEnd uint64
	LastBlock       uint64
	PublicKey       [2]*big.Int
	GroupPublicKey  [4]*big.Int
	Signature       [2]*big.Int
}

// NewGPKSubmissionTask creates a background task that attempts to register with ETHDKG
func NewGPKSubmissionTask(
	acct accounts.Account,
	publicKey [2]*big.Int,
	groupPublicKey [4]*big.Int,
	signature [2]*big.Int,
	registrationEnd uint64, lastBlock uint64) *GPKSubmissionTask {
	return &GPKSubmissionTask{
		Account:         acct,
		PublicKey:       blockchain.CloneBigInt2(publicKey),
		GroupPublicKey:  blockchain.CloneBigInt4(groupPublicKey),
		Signature:       blockchain.CloneBigInt2(signature),
		RegistrationEnd: registrationEnd,
		LastBlock:       lastBlock,
	}
}

// DoWork is the first attempt at registering with ethdkg
func (t *GPKSubmissionTask) DoWork(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at registering with ethdkg
func (t *GPKSubmissionTask) DoRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	return t.doTask(ctx, logger, eth)
}

func (t *GPKSubmissionTask) doTask(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	t.Lock()
	defer t.Unlock()

	// Setup
	c := eth.Contracts()
	txnOpts, err := eth.GetTransactionOpts(ctx, t.Account)
	if err != nil {
		logger.Errorf("getting txn opts failed: %v", err)
		return false
	}

	// Register
	txn, err := c.Ethdkg.SubmitGPKj(txnOpts, t.GroupPublicKey, t.Signature)
	if err != nil {
		logger.Errorf("submitting gpkj failed: %v", err)
		return false
	}

	// Waiting for receipt
	receipt, err := eth.WaitForReceipt(ctx, txn)
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
func (t *GPKSubmissionTask) ShouldRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {

	t.Lock()
	defer t.Unlock()

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, t.Account, logger, eth,
		t.PublicKey, t.RegistrationEnd, t.LastBlock)
}

// DoDone creates a log entry saying task is complete
func (t *GPKSubmissionTask) DoDone(logger *logrus.Logger) {
	logger.Infof("done")
}
