package dkgtasks

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

// RegisterTask contains required state for safely performing a registration
type RegisterTask struct {
	Start   uint64
	End     uint64
	State   *objects.DkgState
	Success bool
	TxOpts  *bind.TransactOpts
	TxHash  common.Hash
}

// asserting that RegisterTask struct implements interface interfaces.Task
var _ interfaces.Task = &RegisterTask{}

// NewRegisterTask creates a background task that attempts to register with ETHDKG
func NewRegisterTask(state *objects.DkgState, start uint64, end uint64) *RegisterTask {
	return &RegisterTask{
		Start:   start,
		End:     end,
		State:   state,
		Success: false,
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
			logger.Errorf("getting txn opts failed: %v", err)
			return err
		}

		nonce, err := eth.GetGethClient().PendingNonceAt(ctx, t.State.Account.Address)
		if err != nil {
			logger.Errorf("getting acct nonce 2: %v", err)
			return err
		}

		txnOpts.Nonce = big.NewInt(int64(nonce))

		logger.WithFields(logrus.Fields{
			"GasFeeCap": txnOpts.GasFeeCap,
			"GasTipCap": txnOpts.GasTipCap,
			"Nonce":     txnOpts.Nonce,
		}).Info("registering fees")

		t.TxOpts = txnOpts
	}

	// Register
	logger.Infof("Registering  publicKey (%v) with ETHDKG", FormatPublicKey(t.State.TransportPublicKey))
	logger.Debugf("registering on block %v with public key: %v", block, FormatPublicKey(t.State.TransportPublicKey))
	txn, err := eth.Contracts().Ethdkg().Register(t.TxOpts, t.State.TransportPublicKey)
	if err != nil {
		logger.Errorf("registering failed: %v", err)

		if err.Error() == "nonce too low" {
			nonce, err := eth.GetGethClient().PendingNonceAt(ctx, t.State.Account.Address)
			if err != nil {
				logger.Errorf("getting acct nonce 2: %v", err)
				return err
			}

			t.TxOpts.Nonce = big.NewInt(int64(nonce))
		}

		return err
	}

	t.TxHash = txn.Hash()

	logger.WithFields(logrus.Fields{
		"GasFeeCap":  t.TxOpts.GasFeeCap,
		"GasFeeCap2": txn.GasFeeCap(),
		"GasTipCap":  t.TxOpts.GasTipCap,
		"GasTipCap2": txn.GasTipCap(),
		"Nonce":      t.TxOpts.Nonce,
		// "Nonce2":     txn.Nonce,
		"txn.Hash()": txn.Hash().Hex(),
		"t.TxHash":   t.TxHash.Hex(),
		//"emptyHashEq": t.TxHash == emptyHash,
	}).Info("registering fees 2")

	timeOutCtx, cancelFunc := context.WithTimeout(ctx, 30*time.Second)
	defer cancelFunc()

	eth.Queue().QueueTransaction(ctx, txn)

	// Waiting for receipt with timeout
	// If we reach the timeout, we return an error
	// to evaluate if we should retry with a higher fee and tip
	receipt, err := eth.Queue().WaitTransaction(timeOutCtx, txn)
	if err != nil {
		logger.Errorf("waiting for receipt failed: %v", err)
		return err
	}

	if receipt == nil {
		//logger.Error("missing registration receipt")
		return errors.New("missing registration receipt")
	}

	// Wait for finalityDelay to avoid fork rollback
	err = dkg.WaitConfirmations(t.TxHash, ctx, logger, eth)
	if err != nil {
		logger.Errorf("waiting confirmations failed: %v", err)
		return err
	}

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

	if needsRegistration {
		if t.TxOpts == nil {
			txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
			if err != nil {
				logger.Errorf("getting txn opts failed: %v", err)
				return true
			}

			t.TxOpts = txnOpts
		}

		// increase FeeCap and TipCap
		l := logger.WithFields(logrus.Fields{
			"gasFeeCap": t.TxOpts.GasFeeCap,
			"gasTipCap": t.TxOpts.GasTipCap,
			"Nonce":     t.TxOpts.Nonce,
		})

		// calculate 10% increase in GasFeeCap and GasTipCap
		increasedFeeCap, increasedTipCap := dkg.IncreaseFeeAndTipCap(t.TxOpts.GasFeeCap, t.TxOpts.GasTipCap, big.NewInt(10))
		t.TxOpts.GasFeeCap = increasedFeeCap
		t.TxOpts.GasTipCap = increasedTipCap

		l.WithFields(logrus.Fields{
			"gasFeeCap10pc": t.TxOpts.GasFeeCap,
			"gasTipCap10pc": t.TxOpts.GasTipCap,
		}).Info("Retrying register with higher fee/tip caps")
	} else {
		var emptyHash common.Hash
		if t.TxHash != emptyHash {
			// Wait for finalityDelay to avoid fork rollback
			err = dkg.WaitConfirmations(t.TxHash, ctx, logger, eth)
			if err != nil {
				logger.Errorf("register.ShouldRetry() error waitingConfirmations: %v", err)
				return true
			}
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
