package dkgtasks

import (
	"context"
	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
	"math/big"
)

// RegisterTask contains required state for safely performing a registration
type RegisterTask struct {
	*DkgTask
}

// asserting that RegisterTask struct implements interface interfaces.Task
var _ interfaces.Task = &RegisterTask{}

// asserting that RegisterTask struct implements DkgTaskIfase
var _ DkgTaskIfase = &RegisterTask{}

// NewRegisterTask creates a background task that attempts to register with ETHDKG
func NewRegisterTask(state *objects.DkgState, start uint64, end uint64) *RegisterTask {
	return &RegisterTask{
		DkgTask: &DkgTask{
			State:   state,
			Start:   start,
			End:     end,
			Success: false,
		},
	}
}

// Initialize begins the setup phase for Register.
// We construct our TransportPrivateKey and TransportPublicKey
// which will be used in the ShareDistribution phase for secure communication.
// These keys are *not* used otherwise.
// Also get the list of existing validators from the pool to assert accusation
// in later phases
func (t *RegisterTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("RegisterTask Initialize()")

	callOpts := eth.GetCallOpts(ctx, t.State.Account)
	validatorAddresses, err := dkg.GetValidatorAddressesFromPool(callOpts, eth, logger)

	if err != nil {
		return dkg.LogReturnErrorf(logger, "RegisterTask.Initialize(): Unable to get validatorAddresses from ValidatorPool: %v", err)
	}

	t.State.ValidatorAddresses = validatorAddresses
	t.State.NumberOfValidators = len(validatorAddresses)

	priv, pub, err := math.GenerateKeys()
	if err != nil {
		return err
	}
	t.State.TransportPrivateKey = priv
	t.State.TransportPublicKey = pub
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

	// Is there any point in running? Make sure we're both initialized and within block range
	block, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return err
	}

	logger.Info("RegisterTask doTask()")

	// Setup
	if t.TxOpts == nil {
		txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
		}

		t.TxOpts = txnOpts
	}

	logger.WithFields(logrus.Fields{
		"GasFeeCap": t.TxOpts.GasFeeCap,
		"GasTipCap": t.TxOpts.GasTipCap,
		"Nonce":     t.TxOpts.Nonce,
		"TxHash":    t.TxHash.Hex(),
	}).Info("registering fees")

	// Register
	logger.Infof("Registering  publicKey (%v) with ETHDKG", FormatPublicKey(t.State.TransportPublicKey))
	logger.Debugf("registering on block %v with public key: %v", block, FormatPublicKey(t.State.TransportPublicKey))
	txn, err := eth.Contracts().Ethdkg().Register(t.TxOpts, t.State.TransportPublicKey)
	if err != nil {
		logger.Errorf("registering failed: %v", err)
		return err
	}
	t.TxHash = txn.Hash()
	t.TxOpts.GasFeeCap = txn.GasFeeCap()
	t.TxOpts.GasTipCap = txn.GasTipCap()
	t.TxOpts.Nonce = big.NewInt(int64(txn.Nonce()))

	logger.WithFields(logrus.Fields{
		"GasFeeCap":  txn.GasFeeCap(),
		"GasTipCap":  txn.GasTipCap(),
		"Nonce":      txn.Nonce(),
		"txn.Hash()": txn.Hash(),
		"TxHash":     t.TxHash.Hex(),
	}).Info("registering fees2")

	// Queue transaction
	eth.Queue().QueueTransaction(ctx, txn)
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

	if t.State.Phase != objects.RegistrationOpen {
		return false
	}

	callOpts := eth.GetCallOpts(ctx, t.State.Account)

	var needsRegistration bool
	status, err := CheckRegistration(eth.Contracts().Ethdkg(), logger, callOpts, t.State.Account.Address, t.State.TransportPublicKey)
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

func (t *RegisterTask) GetDkgTask() *DkgTask {
	return t.DkgTask
}

func (t *RegisterTask) SetDkgTask(dkgTask *DkgTask) {
	t.DkgTask = dkgTask
}
