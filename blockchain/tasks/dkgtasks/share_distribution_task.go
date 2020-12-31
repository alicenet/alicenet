package dkgtasks

import (
	"context"
	"math/big"
	"sync"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/sirupsen/logrus"
)

// ShareDistributionTask stores the data required safely distribute shares
type ShareDistributionTask struct {
	sync.Mutex
	eth             blockchain.Ethereum
	acct            accounts.Account
	logger          *logrus.Logger
	registrationEnd uint64
	lastBlock       uint64
	publicKey       [2]*big.Int
	encryptedShares []*big.Int
	commitments     [][2]*big.Int
}

// NewShareDistributionTask creates a new task
func NewShareDistributionTask(logger *logrus.Logger, eth blockchain.Ethereum, acct accounts.Account, publicKey [2]*big.Int, encryptedShares []*big.Int, commitments [][2]*big.Int, registrationEnd uint64, lastBlock uint64) *ShareDistributionTask {
	return &ShareDistributionTask{
		logger:          logger,
		eth:             eth,
		acct:            acct,
		registrationEnd: registrationEnd,
		lastBlock:       lastBlock,
		commitments:     commitments,
		encryptedShares: blockchain.CloneBigIntSlice(encryptedShares),
		publicKey:       blockchain.CloneBigInt2(publicKey),
	}
}

// DoWork is the first attempt at distributing shares via ethdkg
func (t *ShareDistributionTask) DoWork(ctx context.Context) bool {
	return t.doTask(ctx)
}

// DoRetry is subsequent attempts at distributing shares via ethdkg
func (t *ShareDistributionTask) DoRetry(ctx context.Context) bool {
	return t.doTask(ctx)
}

func (t *ShareDistributionTask) doTask(ctx context.Context) bool {

	t.Lock()
	defer t.Unlock()

	// Setup
	c := t.eth.Contracts()
	t.logger.Infof("me:%v", t.acct.Address.Hex())
	txnOpts, err := t.eth.GetTransactionOpts(ctx, t.acct)
	if err != nil {
		t.logger.Errorf("getting txn opts failed: %v", err)
		return false
	}

	// Distribute shares
	t.logger.Infof("# shares:%d # commitments:%d", len(t.encryptedShares), len(t.commitments))
	txn, err := c.Ethdkg.DistributeShares(txnOpts, t.encryptedShares, t.commitments)
	if err != nil {
		t.logger.Errorf("distributing shares failed: %v", err)
		return false
	}

	// Waiting for receipt
	receipt, err := t.eth.WaitForReceipt(ctx, txn)
	if err != nil {
		t.logger.Errorf("waiting for receipt failed: %v", err)
		return false
	}
	if receipt == nil {
		t.logger.Error("missing distribute shares receipt")
		return false
	}

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		t.logger.Errorf("receipt status indicates failure: %v", receipt.Status)
		return false
	}

	return true
}

// ShouldRetry checks if it makes sense to try again
func (t *ShareDistributionTask) ShouldRetry(ctx context.Context) bool {

	t.Lock()
	defer t.Unlock()

	// This wraps the retry logic for the general case
	generalRetry := GeneralTaskShouldRetry(ctx, t.logger,
		t.eth, t.acct, t.publicKey,
		t.registrationEnd, t.lastBlock)

	// If it's generally good to retry, let's try to be more specific
	if generalRetry {
		callOpts := t.eth.GetCallOpts(ctx, t.acct)
		distributionHash, err := t.eth.Contracts().Ethdkg.ShareDistributionHashes(callOpts, t.acct.Address)
		if err != nil {
			return true
		}

		// TODO an I prove this is the correct share distribution hash?
		t.logger.Infof("DistributionHash: %x", distributionHash)
	}

	return generalRetry
}

// DoDone creates a log entry saying task is complete
func (t *ShareDistributionTask) DoDone() {
	t.logger.Infof("done")
}
