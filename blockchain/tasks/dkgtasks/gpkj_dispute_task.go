package dkgtasks

import (
	"context"
	"math/big"
	"sync"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/sirupsen/logrus"
)

// GPKJDisputeTask contains required state for performing a group accusation
type GPKJDisputeTask struct {
	sync.Mutex
	Account           accounts.Account
	RregistrationEnd  uint64
	LastBlock         uint64
	PublicKey         [2]*big.Int
	Inverse           []*big.Int
	HonestIndicies    []*big.Int
	DishonestIndicies []*big.Int
	RegistrationEnd   uint64
}

// NewGPKJDisputeTask creates a background task that attempts perform a group accusation if necessary
func NewGPKJDisputeTask(
	acct accounts.Account,
	publicKey [2]*big.Int,
	inverse []*big.Int,
	honestIndicies []*big.Int,
	dishonestIndicies []*big.Int,
	registrationEnd uint64, lastBlock uint64) *GPKJDisputeTask {
	return &GPKJDisputeTask{
		Account:           acct,
		PublicKey:         blockchain.CloneBigInt2(publicKey),
		Inverse:           blockchain.CloneBigIntSlice(inverse),
		HonestIndicies:    blockchain.CloneBigIntSlice(honestIndicies),
		DishonestIndicies: blockchain.CloneBigIntSlice(dishonestIndicies),
		RegistrationEnd:   registrationEnd,
		LastBlock:         lastBlock,
	}
}

// DoWork is the first attempt at registering with ethdkg
func (t *GPKJDisputeTask) DoWork(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	logger.Info("DoWork() ...")

	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at registering with ethdkg
func (t *GPKJDisputeTask) DoRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	logger.Info("DoRetry() ...")

	return t.doTask(ctx, logger, eth)
}

func (t *GPKJDisputeTask) doTask(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {

	t.Lock()
	defer t.Unlock()

	// Setup
	c := eth.Contracts()
	txnOpts, err := eth.GetTransactionOpts(ctx, t.Account)
	if err != nil {
		logger.Errorf("getting txn opts failed: %v", err)
		return false
	}

	// Perform group accusation
	txn, err := c.Ethdkg.GroupAccusationGPKj(txnOpts, t.Inverse, t.HonestIndicies, t.DishonestIndicies)
	if err != nil {
		logger.Errorf("group accusation failed: %v", err)
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
func (t *GPKJDisputeTask) ShouldRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {

	t.Lock()
	defer t.Unlock()

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, t.Account, logger,
		eth, t.PublicKey,
		t.RegistrationEnd, t.LastBlock)
}

// DoDone creates a log entry saying task is complete
func (t *GPKJDisputeTask) DoDone(logger *logrus.Logger) {
	logger.Infof("done")
}
