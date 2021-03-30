package dkgtasks

import (
	"context"
	"math/big"
	"sync"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/sirupsen/logrus"
)

// ShareDistributionTask stores the data required safely distribute shares
type ShareDistributionTask struct {
	sync.Mutex
	registrationEnd uint64
	lastBlock       uint64
	publicKey       [2]*big.Int
	encryptedShares []*big.Int
	commitments     [][2]*big.Int
}

// NewShareDistributionTask creates a new task
func NewShareDistributionTask(publicKey [2]*big.Int, encryptedShares []*big.Int, commitments [][2]*big.Int, registrationEnd uint64, lastBlock uint64) *ShareDistributionTask {
	return &ShareDistributionTask{
		registrationEnd: registrationEnd,
		lastBlock:       lastBlock,
		commitments:     commitments,
		encryptedShares: blockchain.CloneBigIntSlice(encryptedShares),
		publicKey:       blockchain.CloneBigInt2(publicKey),
	}
}

// DoWork is the first attempt at distributing shares via ethdkg
func (t *ShareDistributionTask) DoWork(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is subsequent attempts at distributing shares via ethdkg
func (t *ShareDistributionTask) DoRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	return t.doTask(ctx, logger, eth)
}

func (t *ShareDistributionTask) doTask(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {

	t.Lock()
	defer t.Unlock()

	// Setup
	c := eth.Contracts()
	me := eth.GetDefaultAccount()
	logger.Infof("me:%v", me.Address.Hex())
	txnOpts, err := eth.GetTransactionOpts(ctx, me)
	if err != nil {
		logger.Errorf("getting txn opts failed: %v", err)
		return false
	}

	// Distribute shares
	logger.Infof("# shares:%d # commitments:%d", len(t.encryptedShares), len(t.commitments))
	txn, err := c.Ethdkg.DistributeShares(txnOpts, t.encryptedShares, t.commitments)
	if err != nil {
		logger.Errorf("distributing shares failed: %v", err)
		return false
	}

	// Waiting for receipt
	receipt, err := eth.WaitForReceipt(ctx, txn)
	if err != nil {
		logger.Errorf("waiting for receipt failed: %v", err)
		return false
	}
	if receipt == nil {
		logger.Error("missing distribute shares receipt")
		return false
	}

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		logger.Errorf("receipt status indicates failure: %v", receipt.Status)
		return false
	}

	return true
}

// ShouldRetry checks if it makes sense to try again
func (t *ShareDistributionTask) ShouldRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {

	t.Lock()
	defer t.Unlock()

	// This wraps the retry logic for the general case
	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.publicKey,
		t.registrationEnd, t.lastBlock)

	// If it's generally good to retry, let's try to be more specific
	if generalRetry {
		me := eth.GetDefaultAccount()
		callOpts := eth.GetCallOpts(ctx, me)
		distributionHash, err := eth.Contracts().Ethdkg.ShareDistributionHashes(callOpts, me.Address)
		if err != nil {
			return true
		}

		// TODO an I prove this is the correct share distribution hash?
		logger.Infof("DistributionHash: %x", distributionHash)
	}

	return generalRetry
}

// DoDone creates a log entry saying task is complete
func (t *ShareDistributionTask) DoDone(logger *logrus.Logger) {
	logger.Infof("done")
}
