package dkgtasks

import (
	"context"
	"math/big"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/sirupsen/logrus"
)

// GPKJDisputeTask contains required state for safely performing a registration
type GPKJDisputeTask struct {
	eth               blockchain.Ethereum
	acct              accounts.Account
	logger            *logrus.Logger
	registrationEnd   uint64
	lastBlock         uint64
	publicKey         [2]*big.Int
	inverse           []*big.Int
	honestIndicies    []*big.Int
	dishonestIndicies []*big.Int
	success           bool
	count             int
}

// NewGPKJDisputeTask creates a background task that attempts to register with ETHDKG
func NewGPKJDisputeTask(logger *logrus.Logger, eth blockchain.Ethereum, acct accounts.Account,
	publicKey [2]*big.Int,
	inverse []*big.Int,
	honestIndicies []*big.Int,
	dishonestIndicies []*big.Int,
	registrationEnd uint64, lastBlock uint64) *GPKJDisputeTask {
	return &GPKJDisputeTask{
		logger:            logger,
		eth:               eth,
		acct:              acct,
		publicKey:         blockchain.CloneBigInt2(publicKey),
		inverse:           blockchain.CloneBigIntSlice(inverse),
		honestIndicies:    blockchain.CloneBigIntSlice(honestIndicies),
		dishonestIndicies: blockchain.CloneBigIntSlice(dishonestIndicies),
		registrationEnd:   registrationEnd,
		lastBlock:         lastBlock,
	}
}

// DoWork is the first attempt at registering with ethdkg
func (t *GPKJDisputeTask) DoWork(ctx context.Context) {
	t.logger.Info("DoWork() ...")
	t.count = 1

	t.doTask(ctx)
}

// DoRetry is all subsequent attempts at registering with ethdkg
func (t *GPKJDisputeTask) DoRetry(ctx context.Context) {
	t.logger.Info("DoRetry() ...")
	t.count++

	t.doTask(ctx)
}

func (t *GPKJDisputeTask) doTask(ctx context.Context) {

	// Setup
	c := t.eth.Contracts()
	txnOpts, err := t.eth.GetTransactionOpts(ctx, t.acct)
	if err != nil {
		t.logger.Errorf("getting txn opts failed: %v", err)
		return
	}

	// Register
	txn, err := c.Ethdkg.GroupAccusationGPKj(txnOpts, t.inverse, t.honestIndicies, t.dishonestIndicies)
	if err != nil {
		t.logger.Errorf("registering failed: %v", err)
		return
	}

	// Waiting for receipt
	receipt, err := t.eth.WaitForReceipt(ctx, txn)
	if err != nil {
		t.logger.Errorf("waiting for receipt failed: %v", err)
		return
	}
	if receipt == nil {
		t.logger.Error("missing registration receipt")
		return
	}

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		t.logger.Errorf("registration status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
		return
	}

	t.success = true
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
// -- the registration open hasn't moved, i.e. ETHDKG has not restarted
func (t *GPKJDisputeTask) ShouldRetry(ctx context.Context) bool {

	// If we were successful we should not try again
	if t.success {
		return false
	}

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, t.logger,
		t.eth, t.acct, t.publicKey,
		t.registrationEnd, t.lastBlock)
}

// DoDone creates a log entry saying task is complete
func (t *GPKJDisputeTask) DoDone() {
	t.logger.Infof("DoDone() ... tries:%v", t.count)
}
