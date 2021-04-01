package dkgtasks

import (
	"context"
	"math/big"
	"sync"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/sirupsen/logrus"
)

// RegisterTask contains required state for safely performing a registration
type RegisterTask struct {
	sync.Mutex
	Account   accounts.Account
	LastBlock uint64
	PublicKey [2]*big.Int
}

// NewRegisterTask creates a background task that attempts to register with ETHDKG
func NewRegisterTask(acct accounts.Account, publicKey [2]*big.Int, lastBlock uint64) *RegisterTask {
	return &RegisterTask{
		Account:   acct,
		PublicKey: blockchain.CloneBigInt2(publicKey),
		LastBlock: lastBlock,
	}
}

// DoWork is the first attempt at registering with ethdkg
func (t *RegisterTask) DoWork(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at registering with ethdkg
func (t *RegisterTask) DoRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	return t.doTask(ctx, logger, eth)
}

func (t *RegisterTask) doTask(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	logger.Infof("Registering  publicKey (%v) with ETHDKG", FormatPublicKey(t.PublicKey))

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
	logger.Infof("registering public key: %v", FormatPublicKey(t.PublicKey))
	logger.Infof("txnOpts:%v", txnOpts.From)
	txn, err := c.Ethdkg.Register(txnOpts, t.PublicKey)
	if err != nil {
		logger.Errorf("registering failed: %v", err)
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
func (t *RegisterTask) ShouldRetry(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {

	t.Lock()
	defer t.Unlock()

	c := eth.Contracts()
	callOpts := eth.GetCallOpts(ctx, t.Account)

	currentBlock, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return true
	}

	// Definitely past quitting time
	if currentBlock > t.LastBlock {
		return false
	}

	// Check if the registration window has moved, quit if it has
	lastBlock, err := c.Ethdkg.TREGISTRATIONEND(callOpts)
	if err != nil {
		return true
	}

	if lastBlock.Uint64() != t.LastBlock {
		logger.Infof("aborting registration due to restart")
		return false
	}

	// Check to see if we are already registered
	ethdkg := eth.Contracts().Ethdkg
	status, err := CheckRegistration(ctx, ethdkg, logger, callOpts, t.Account.Address, t.PublicKey)
	if err != nil {
		logger.Warnf("could not check if we're registered: %v", err)
		return true
	}

	if status == Registered || status == BadRegistration {
		return false
	}

	return true
}

// DoDone just creates a log entry saying task is complete
func (t *RegisterTask) DoDone(logger *logrus.Logger) {
	logger.Infof("done")
}
