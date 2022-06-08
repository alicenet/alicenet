package dkg

import (
	"fmt"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	dkgUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
)

// RegisterTask contains required state for safely performing a registration
type RegisterTask struct {
	*objects.Task
}

// asserting that RegisterTask struct implements interface interfaces.Task
var _ interfaces.ITask = &RegisterTask{}

// NewRegisterTask creates a background task that attempts to register with ETHDKG
func NewRegisterTask(start uint64, end uint64) *RegisterTask {
	return &RegisterTask{
		Task: objects.NewTask(constants.RegisterTaskName, start, end, false),
	}
}

// Prepare prepares for work to be done in the RegisterTask
// We construct our TransportPrivateKey and TransportPublicKey
// which will be used in the ShareDistribution phase for secure communication.
// These keys are *not* used otherwise.
// Also get the list of existing validators from the pool to assert accusation
// in later phases
func (t *RegisterTask) Prepare() error {
	logger := t.GetLogger()
	logger.Infof("RegisterTask Prepare()...")

	dkgState := &state.DkgState{}
	err := t.GetDB().Update(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		if err != nil {
			return err
		}

		if dkgState.TransportPrivateKey == nil ||
			dkgState.TransportPrivateKey.Cmp(big.NewInt(0)) == 0 {

			logger.Infof("RegisterTask Prepare(): generating private-public transport keys")
			priv, pub, err := state.GenerateKeys()
			if err != nil {
				return err
			}
			dkgState.TransportPrivateKey = priv
			dkgState.TransportPublicKey = pub

			err = dkgState.PersistState(txn)
			if err != nil {
				return err
			}
		} else {
			logger.Infof("RegisterTask Prepare(): private-public transport keys already defined")
		}

		return nil
	})

	if err != nil {
		return dkgUtils.LogReturnErrorf(logger, "RegisterTask.Prepare(): error during the preparation: %v", err)
	}

	return nil
}

// Execute executes the task business logic
func (t *RegisterTask) Execute() ([]*types.Transaction, error) {
	logger := t.GetLogger()
	logger.Info("RegisterTask Execute()")

	// Is there any point in running? Make sure we're both initialized and within block range
	eth := t.GetEth()
	ctx := t.GetCtx()
	block, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return nil, err
	}

	dkgState := &state.DkgState{}
	err = t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("RegisterTask.Execute(): error loading dkgState: %v", err)
	}

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// Register
	logger.Infof("Registering  publicKey (%v) with ETHDKG", dkgUtils.FormatPublicKey(dkgState.TransportPublicKey))
	logger.Debugf("registering on block %v with public key: %v", block, dkgUtils.FormatPublicKey(dkgState.TransportPublicKey))
	txn, err := eth.Contracts().Ethdkg().Register(txnOpts, dkgState.TransportPublicKey)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "registering failed: %v", err)
	}

	return []*types.Transaction{txn}, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *RegisterTask) ShouldExecute() bool {
	logger := t.GetLogger()
	logger.Info("RegisterTask ShouldExecute")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		logger.Errorf("could not get dkgState with error %v", err)
		return true
	}

	eth := t.GetEth()
	ctx := t.GetCtx()
	if dkgState.Phase != state.RegistrationOpen {
		return false
	}

	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		logger.Errorf("RegisterTask ShouldExecute failed getting call options: %v", err)
		return true
	}

	var needsRegistration bool
	status, err := state.CheckRegistration(eth.Contracts().Ethdkg(), logger, callOpts, dkgState.Account.Address, dkgState.TransportPublicKey)
	logger.Infof("registration status: %v", status)
	if err != nil {
		needsRegistration = true
	} else {
		if status != state.Registered && status != state.BadRegistration {
			needsRegistration = true
		}
	}

	return needsRegistration
}
