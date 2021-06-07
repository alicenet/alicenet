package dkgtasks

import (
	"context"
	"sync"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/sirupsen/logrus"
)

// RegisterTask contains required state for safely performing a registration
type RegisterTask struct {
	sync.Mutex              // TODO Do I need this? It might be sufficient to only use a RWMutex on `State`
	OriginalRegistrationEnd uint64
	State                   *dkg.EthDKGState
}

// NewRegisterTask creates a background task that attempts to register with ETHDKG
func NewRegisterTask(state *dkg.EthDKGState) *RegisterTask {
	return &RegisterTask{
		OriginalRegistrationEnd: state.RegistrationEnd, // If these quit being equal, this task should be abandoned
		State:                   state,
	}
}

// This is not exported and does not lock so can only be called from within task. Return value indicates whether task has been initialized.
func (t *RegisterTask) init(ctx context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	if t.State.TransportPrivateKey == nil {
		priv, pub, err := math.GenerateKeys()
		if err != nil {
			return false
		}

		t.State.TransportPrivateKey = priv
		t.State.TransportPublicKey = pub
	}

	return true
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

	t.Lock()
	defer t.Unlock()

	// Is there any point in running? Make sure we're both initialized and within block range
	if !t.init(ctx, logger, eth) {
		return false
	}

	block, err := eth.GetCurrentHeight(ctx)
	if err != nil || block >= t.State.RegistrationEnd {
		return false
	}

	// Setup
	c := eth.Contracts()
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		logger.Errorf("getting txn opts failed: %v", err)
		return false
	}

	// Register
	logger.Infof("Registering  publicKey (%v) with ETHDKG", FormatPublicKey(t.State.TransportPublicKey))
	logger.Debugf("registering on block %v with public key: %v", block, FormatPublicKey(t.State.TransportPublicKey))
	txn, err := c.Ethdkg.Register(txnOpts, t.State.TransportPublicKey)
	if err != nil {
		logger.Errorf("registering failed: %v", err)
		return false
	}
	eth.Queue().QueueTransaction(ctx, txn)

	// Waiting for receipt
	receipt, err := eth.Queue().WaitTransaction(ctx, txn)
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
	callOpts := eth.GetCallOpts(ctx, t.State.Account)

	currentBlock, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return true
	}

	// Definitely past quitting time
	if currentBlock >= t.State.RegistrationEnd {
		return false
	}

	// Check if the registration window has moved, quit if it has
	lastBlock, err := c.Ethdkg.TREGISTRATIONEND(callOpts)
	if err != nil {
		return true
	}

	// We save registration star
	if lastBlock.Uint64() != t.OriginalRegistrationEnd {
		logger.Infof("aborting registration due to restart")
		return false
	}

	// Check to see if we are already registered
	ethdkg := eth.Contracts().Ethdkg
	status, err := CheckRegistration(ctx, ethdkg, logger, callOpts, t.State.Account.Address, t.State.TransportPublicKey)
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
