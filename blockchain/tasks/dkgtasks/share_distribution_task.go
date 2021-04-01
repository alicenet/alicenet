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
	Account         accounts.Account
	RegistrationEnd uint64
	LastBlock       uint64
	PublicKey       [2]*big.Int
	EncryptedShares []*big.Int
	Commitments     [][2]*big.Int
}

// NewShareDistributionTask creates a new task
func NewShareDistributionTask(acct accounts.Account, publicKey [2]*big.Int, encryptedShares []*big.Int, commitments [][2]*big.Int, registrationEnd uint64, lastBlock uint64) *ShareDistributionTask {
	return &ShareDistributionTask{
		Account:         acct,
		RegistrationEnd: registrationEnd,
		LastBlock:       lastBlock,
		Commitments:     commitments,
		EncryptedShares: blockchain.CloneBigIntSlice(encryptedShares),
		PublicKey:       blockchain.CloneBigInt2(publicKey),
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
	logger.Infof("me:%v", t.Account.Address.Hex())
	txnOpts, err := eth.GetTransactionOpts(ctx, t.Account)
	if err != nil {
		logger.Errorf("getting txn opts failed: %v", err)
		return false
	}

	// Distribute shares
	logger.Infof("# shares:%d # commitments:%d", len(t.EncryptedShares), len(t.Commitments))
	txn, err := c.Ethdkg.DistributeShares(txnOpts, t.EncryptedShares, t.Commitments)
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
	generalRetry := GeneralTaskShouldRetry(ctx, t.Account, logger, eth, t.PublicKey,
		t.RegistrationEnd, t.LastBlock)

	// If it's generally good to retry, let's try to be more specific
	if generalRetry {
		callOpts := eth.GetCallOpts(ctx, t.Account)
		distributionHash, err := eth.Contracts().Ethdkg.ShareDistributionHashes(callOpts, t.Account.Address)
		if err != nil {
			return true
		}

		// TODO can I prove this is the correct share distribution hash?
		logger.Infof("DistributionHash: %x", distributionHash)
	}

	return generalRetry
}

// DoDone creates a log entry saying task is complete
func (t *ShareDistributionTask) DoDone(logger *logrus.Logger) {
	logger.Infof("done")
}
