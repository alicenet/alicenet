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
	eth       blockchain.Ethereum
	acct      accounts.Account
	logger    *logrus.Logger
	lastBlock uint64
	publicKey [2]*big.Int
}

// NewRegisterTask creates a background task that attempts to register with ETHDKG
func NewRegisterTask(logger *logrus.Logger, eth blockchain.Ethereum, acct accounts.Account, publicKey [2]*big.Int, lastBlock uint64) *RegisterTask {
	logger.Infof("Registering  publicKey (%v) with ETHDKG", FormatPublicKey(publicKey))
	return &RegisterTask{
		logger:    logger,
		eth:       eth,
		acct:      acct,
		publicKey: blockchain.CloneBigInt2(publicKey),
		lastBlock: lastBlock,
	}
}

// DoWork is the first attempt at registering with ethdkg
func (t *RegisterTask) DoWork(ctx context.Context) bool {
	return t.doTask(ctx)
}

// DoRetry is all subsequent attempts at registering with ethdkg
func (t *RegisterTask) DoRetry(ctx context.Context) bool {
	return t.doTask(ctx)
}

func (t *RegisterTask) doTask(ctx context.Context) bool {
	t.Lock()
	defer t.Unlock()

	// Setup
	c := t.eth.Contracts()
	txnOpts, err := t.eth.GetTransactionOpts(ctx, t.acct)
	if err != nil {
		t.logger.Errorf("getting txn opts failed: %v", err)
		return false
	}

	// Register
	t.logger.Infof("registering public key: %v", FormatPublicKey(t.publicKey))
	txn, err := c.Ethdkg.Register(txnOpts, t.publicKey)
	if err != nil {
		t.logger.Errorf("registering failed: %v", err)
		return false
	}

	// Waiting for receipt
	receipt, err := t.eth.WaitForReceipt(ctx, txn)
	if err != nil {
		t.logger.Errorf("waiting for receipt failed: %v", err)
		return false
	}
	if receipt == nil {
		t.logger.Error("missing registration receipt")
		return false
	}

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		t.logger.Errorf("registration status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
		return false
	}

	return true
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
// -- the registration open hasn't moved, i.e. ETHDKG has not restarted
func (t *RegisterTask) ShouldRetry(ctx context.Context) bool {

	t.Lock()
	defer t.Unlock()

	c := t.eth.Contracts()
	callOpts := t.eth.GetCallOpts(ctx, t.acct)

	currentBlock, err := t.eth.GetCurrentHeight(ctx)
	if err != nil {
		return true
	}

	// Definitely past quitting time
	if currentBlock > t.lastBlock {
		return false
	}

	// Check if the registration window has moved, quit if it has
	lastBlock, err := c.Ethdkg.TREGISTRATIONEND(callOpts)
	if err != nil {
		return true
	}

	if lastBlock.Uint64() != t.lastBlock {
		t.logger.Infof("aborting registration due to restart")
		return false
	}

	// Check to see if we are already registered
	ethdkg := t.eth.Contracts().Ethdkg
	status, err := CheckRegistration(ctx, ethdkg, t.logger, callOpts, t.acct.Address, t.publicKey)
	if err != nil {
		t.logger.Warnf("could not check if we're registered: %v", err)
		return true
	}

	if status == Registered || status == BadRegistration {
		return false
	}

	return true
}

// DoDone just creates a log entry saying task is complete
func (t *RegisterTask) DoDone() {
	t.logger.Infof("done")
}
