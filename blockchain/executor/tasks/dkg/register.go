package dkg

import (
	"context"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/tasks/dkg/math"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/objects"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/utils"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/sirupsen/logrus"
)

// RegisterTask contains required state for safely performing a registration
type RegisterTask struct {
	*objects.Task
}

// asserting that RegisterTask struct implements interface interfaces.Task
var _ interfaces.ITask = &RegisterTask{}

// NewRegisterTask creates a background task that attempts to register with ETHDKG
func NewRegisterTask(state *objects.DkgState, start uint64, end uint64) *RegisterTask {
	return &RegisterTask{
		Task: objects.NewTask(state, RegisterTaskName, start, end),
	}
}

// Initialize begins the setup phase for Register.
// We construct our TransportPrivateKey and TransportPublicKey
// which will be used in the ShareDistribution phase for secure communication.
// These keys are *not* used otherwise.
// Also get the list of existing validators from the pool to assert accusation
// in later phases
func (t *RegisterTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Infof("RegisterTask Initialize()")

	taskState, ok := t.State.(*objects.DkgState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	if taskState.TransportPrivateKey == nil ||
		taskState.TransportPrivateKey.Cmp(big.NewInt(0)) == 0 {

		logger.Infof("RegisterTask Initialize(): generating private-public transport keys")
		priv, pub, err := math.GenerateKeys()
		if err != nil {
			return err
		}
		taskState.TransportPrivateKey = priv
		taskState.TransportPublicKey = pub
	} else {
		logger.Infof("RegisterTask Initialize(): private-public transport keys already defined")
	}

	return nil
}

// DoWork is the first attempt at registering with ethdkg
func (t *RegisterTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at registering with ethdkg
func (t *RegisterTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *RegisterTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	taskState, ok := t.State.(*objects.DkgState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	// Is there any point in running? Make sure we're both initialized and within block range
	block, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return err
	}

	logger.Info("RegisterTask doTask()")

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, taskState.Account)
	if err != nil {
		return utils.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// If the TxOpts exists, meaning the Tx replacement timeout was reached,
	// we increase the Gas to have priority for the next blocks
	if t.TxOpts != nil && t.TxOpts.Nonce != nil {
		logger.Info("txnOpts Replaced")
		txnOpts.Nonce = t.TxOpts.Nonce
		txnOpts.GasFeeCap = t.TxOpts.GasFeeCap
		txnOpts.GasTipCap = t.TxOpts.GasTipCap
	}

	// Register
	logger.Infof("Registering  publicKey (%v) with ETHDKG", FormatPublicKey(taskState.TransportPublicKey))
	logger.Debugf("registering on block %v with public key: %v", block, FormatPublicKey(taskState.TransportPublicKey))
	txn, err := eth.Contracts().Ethdkg().Register(txnOpts, taskState.TransportPublicKey)
	if err != nil {
		logger.Errorf("registering failed: %v", err)
		return err
	}
	t.TxOpts.TxHashes = append(t.TxOpts.TxHashes, txn.Hash())
	t.TxOpts.GasFeeCap = txn.GasFeeCap()
	t.TxOpts.GasTipCap = txn.GasTipCap()
	t.TxOpts.Nonce = big.NewInt(int64(txn.Nonce()))

	logger.WithFields(logrus.Fields{
		"GasFeeCap": t.TxOpts.GasFeeCap,
		"GasTipCap": t.TxOpts.GasTipCap,
		"Nonce":     t.TxOpts.Nonce,
	}).Info("registering fees")

	// Queue transaction
	eth.TransactionWatcher().SubscribeTransaction(ctx, txn)

	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
// -- the registration open hasn't moved, i.e. ETHDKG has not restarted
func (t *RegisterTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("RegisterTask ShouldRetry")
	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	taskState, ok := t.State.(*objects.DkgState)
	if !ok {
		logger.Error("RegisterTask ShouldRetry invalid convertion of taskState object")
		return false
	}

	if taskState.Phase != objects.RegistrationOpen {
		return false
	}

	callOpts, err := eth.GetCallOpts(ctx, taskState.Account)
	if err != nil {
		logger.Errorf("RegisterTask ShouldRetry failed getting call options: %v", err)
		return true
	}

	var needsRegistration bool
	status, err := CheckRegistration(eth.Contracts().Ethdkg(), logger, callOpts, taskState.Account.Address, taskState.TransportPublicKey)
	logger.Infof("registration status: %v", status)
	if err != nil {
		needsRegistration = true
	} else {
		if status != Registered && status != BadRegistration {
			needsRegistration = true
		}
	}

	return needsRegistration
}

// DoDone just creates a log entry saying task is complete
func (t *RegisterTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("RegisterTask done")
}

func (t *RegisterTask) GetExecutionData() interfaces.ITaskExecutionData {
	return t.Task
}
