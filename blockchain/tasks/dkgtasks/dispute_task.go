package dkgtasks

import (
	"context"
	"math/big"
	"sync"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/sirupsen/logrus"
)

// DisputeTask stores the data required to dispute shares
type DisputeTask struct {
	sync.Mutex
	eth             blockchain.Ethereum
	acct            accounts.Account
	logger          *logrus.Logger
	registrationEnd uint64
	lastBlock       uint64
	publicKey       [2]*big.Int
}

// NewDisputeTask creates a new task
func NewDisputeTask(logger *logrus.Logger, eth blockchain.Ethereum, acct accounts.Account, publicKey [2]*big.Int, registrationEnd uint64, lastBlock uint64) *DisputeTask {
	return &DisputeTask{
		logger:          logger,
		eth:             eth,
		acct:            acct,
		registrationEnd: registrationEnd,
		lastBlock:       lastBlock,
		publicKey:       blockchain.CloneBigInt2(publicKey),
	}
}

// DoWork is the first attempt at distributing shares via ethdkg
func (t *DisputeTask) DoWork(ctx context.Context) bool {
	t.logger.Info("DoWork() ...")
	return t.doTask(ctx)
}

// DoRetry is subsequent attempts at distributing shares via ethdkg
func (t *DisputeTask) DoRetry(ctx context.Context) bool {
	t.logger.Info("DoRetry() ...")
	return t.doTask(ctx)
}

func (t *DisputeTask) doTask(ctx context.Context) bool {
	return true
}

// ShouldRetry checks if it makes sense to try again
func (t *DisputeTask) ShouldRetry(ctx context.Context) bool {

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, t.logger,
		t.eth, t.acct, t.publicKey,
		t.registrationEnd, t.lastBlock)
}

// DoDone creates a log entry saying task is complete
func (t *DisputeTask) DoDone() {
	t.logger.Infof("done")
}
