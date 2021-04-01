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
	Account         accounts.Account
	RegistrationEnd uint64
	LastBlock       uint64
	PublicKey       [2]*big.Int
}

// NewDisputeTask creates a new task
func NewDisputeTask(acct accounts.Account, publicKey [2]*big.Int, registrationEnd uint64, lastBlock uint64) *DisputeTask {
	return &DisputeTask{
		Account:         acct,
		RegistrationEnd: registrationEnd,
		LastBlock:       lastBlock,
		PublicKey:       blockchain.CloneBigInt2(publicKey),
	}
}

// DoWork is the first attempt at distributing shares via ethdkg
func (t *DisputeTask) DoWork(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	logger.Info("DoWork() ...")
	return t.doTask(ctx)
}

// DoRetry is subsequent attempts at distributing shares via ethdkg
func (t *DisputeTask) DoRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	logger.Info("DoRetry() ...")
	return t.doTask(ctx)
}

func (t *DisputeTask) doTask(ctx context.Context) bool {
	return true
}

// ShouldRetry checks if it makes sense to try again
func (t *DisputeTask) ShouldRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, t.Account, logger, eth,
		t.PublicKey, t.RegistrationEnd, t.LastBlock)
}

// DoDone creates a log entry saying task is complete
func (t *DisputeTask) DoDone(logger *logrus.Logger) {
	logger.Infof("done")
}
