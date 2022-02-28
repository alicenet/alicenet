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
			State:             state,
			Start:             start,
			End:               end,
			Success:           false,
			TxReplacementOpts: &TxReplacementOpts{},
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
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	if t.TxReplacementOpts != nil && t.TxReplacementOpts.Nonce != nil {
		logger.Info("txnOpts Replaced")
		txnOpts.Nonce = t.TxReplacementOpts.Nonce
		txnOpts.GasFeeCap = t.TxReplacementOpts.GasFeeCap
		txnOpts.GasTipCap = t.TxReplacementOpts.GasTipCap
	}

	logger.WithFields(logrus.Fields{
		"GasFeeCap": txnOpts.GasFeeCap,
		"GasTipCap": txnOpts.GasTipCap,
		"Nonce":     txnOpts.Nonce,
	}).Info("registering fees")

	// Register
	logger.Infof("Registering  publicKey (%v) with ETHDKG", FormatPublicKey(t.State.TransportPublicKey))
	logger.Debugf("registering on block %v with public key: %v", block, FormatPublicKey(t.State.TransportPublicKey))
	txn, err := eth.Contracts().Ethdkg().Register(txnOpts, t.State.TransportPublicKey)
	if err != nil {
		logger.Errorf("registering failed: %v", err)
		return err
	}
	t.TxReplacementOpts.TxHash = txn.Hash()
	t.TxReplacementOpts.GasFeeCap = txn.GasFeeCap()
	t.TxReplacementOpts.GasTipCap = txn.GasTipCap()
	t.TxReplacementOpts.Nonce = big.NewInt(int64(txn.Nonce()))

	logger.WithFields(logrus.Fields{
		"GasFeeCap": t.TxReplacementOpts.GasFeeCap,
		"GasTipCap": t.TxReplacementOpts.GasTipCap,
		"Nonce":     t.TxReplacementOpts.Nonce,
		"Hash":      t.TxReplacementOpts.TxHash.Hex(),
	}).Info("registering fees2")

	// Queue transaction
	eth.Queue().QueueTransaction(ctx, txn)

	// Waiting for receipt with timeout
	// If we reach the timeout, we return an error
	// to evaluate if we should retry with a higher fee and tip
	//receipt, err := eth.Queue().WaitTransaction(ctx, txn)
	//if err != nil {
	//	logger.Errorf("waiting for receipt failed: %v", err)
	//	return err
	//}
	//
	//logger.Info("RECEIPT: NO ERROR")
	//
	//if receipt == nil {
	//	logger.Error("missing registration receipt")
	//	return errors.New("missing registration receipt")
	//}
	//
	//logger.Info("RECEIPT: NOT NIL")
	//
	//// Check receipt to confirm we were successful
	//if receipt.Status != uint64(1) {
	//	logger.Errorf("registration status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
	//	return dkg.LogReturnErrorf(logger, "registration status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
	//}
	//logger.Info("RECEIPT: GOOD STATUS")

	logger.WithFields(logrus.Fields{
		"GasFeeCap": t.TxReplacementOpts.GasFeeCap,
		"GasTipCap": t.TxReplacementOpts.GasTipCap,
		"Nonce":     t.TxReplacementOpts.Nonce,
		"Hash":      t.TxReplacementOpts.TxHash.Hex(),
	}).Info("registering fees3")

	// Wait for finalityDelay to avoid fork rollback
	//err = dkg.WaitConfirmations(t.TxReplacementOpts.TxHash, ctx, logger, eth)
	//if err != nil {
	//	logger.Errorf("waiting confirmations failed: %v", err)
	//	return err
	//}

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
