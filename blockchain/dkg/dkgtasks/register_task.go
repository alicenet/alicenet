package dkgtasks

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

// RegisterTask contains required state for safely performing a registration
type RegisterTask struct {
	sync.Mutex              // TODO Update tasks to not be a Mutex, just use State as Mutex
	OriginalRegistrationEnd uint64
	State                   *objects.DkgState
	Success                 bool
}

// NewRegisterTask creates a background task that attempts to register with ETHDKG
func NewRegisterTask(state *objects.DkgState) *RegisterTask {
	return &RegisterTask{
		OriginalRegistrationEnd: state.RegistrationEnd, // If these quit being equal, this task should be abandoned
		State:                   state,
	}
}

// This is not exported and does not lock so can only be called from within task. Return value indicates whether task has been initialized.
func (t *RegisterTask) Initialize(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {

	priv, pub, err := math.GenerateKeys()
	if err != nil {
		return err
	}

	t.State.TransportPrivateKey = priv
	t.State.TransportPublicKey = pub

	return nil
}

// DoWork is the first attempt at registering with ethdkg
func (t *RegisterTask) DoWork(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at registering with ethdkg
func (t *RegisterTask) DoRetry(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *RegisterTask) doTask(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {

	t.Lock()
	defer t.Unlock()

	// Is there any point in running? Make sure we're both initialized and within block range
	block, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return err
	}

	if block >= t.State.RegistrationEnd {
		return fmt.Errorf("at block %v but registration ends at %v", block, t.State.RegistrationEnd)
	}

	// Setup
	c := eth.Contracts()
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		logger.Errorf("getting txn opts failed: %v", err)
		return err
	}

	// Register
	logger.Infof("Registering  publicKey (%v) with ETHDKG", FormatPublicKey(t.State.TransportPublicKey))
	logger.Debugf("registering on block %v with public key: %v", block, FormatPublicKey(t.State.TransportPublicKey))
	txn, err := c.Ethdkg().Register(txnOpts, t.State.TransportPublicKey)
	if err != nil {
		logger.Errorf("registering failed: %v", err)
		return err
	}
	eth.Queue().QueueTransaction(ctx, txn)

	// Waiting for receipt
	receipt, err := eth.Queue().WaitTransaction(ctx, txn)
	if err != nil {
		logger.Errorf("waiting for receipt failed: %v", err)
		return err
	}
	if receipt == nil {
		logger.Error("missing registration receipt")
		return errors.New("registration receipt is nil")
	}

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		message := fmt.Sprintf("registration status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
		logger.Error(message)
		return errors.New(message)
	}

	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
// -- the registration open hasn't moved, i.e. ETHDKG has not restarted
func (t *RegisterTask) ShouldRetry(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) bool {

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
	lastBlock, err := c.Ethdkg().TREGISTRATIONEND(callOpts)
	if err != nil {
		return true
	}

	// We save registration star
	if lastBlock.Uint64() != t.OriginalRegistrationEnd {
		logger.Infof("aborting registration due to restart")
		return false
	}

	// Check to see if we are already registered
	ethdkg := eth.Contracts().Ethdkg()
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

	t.State.Lock()
	defer t.State.Unlock()

	t.State.Registration = t.Success
}
