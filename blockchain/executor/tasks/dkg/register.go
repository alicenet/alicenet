package dkg

import (
	"fmt"
	"math/big"

	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	executorInterfaces "github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	dkgUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
	"github.com/MadBase/MadNet/blockchain/transaction"
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
		Task: objects.NewTask(constants.RegisterTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the RegisterTask
// We construct our TransportPrivateKey and TransportPublicKey
// which will be used in the ShareDistribution phase for secure communication.
// These keys are *not* used otherwise.
// Also get the list of existing validators from the pool to assert accusation
// in later phases
func (t *RegisterTask) Prepare() *executorInterfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Tracef("preparing task")

	dkgState := &state.DkgState{}
	var isRecoverable bool
	err := t.GetDB().Update(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		if err != nil {
			isRecoverable = false
			return err
		}

		if dkgState.TransportPrivateKey == nil ||
			dkgState.TransportPrivateKey.Cmp(big.NewInt(0)) == 0 {

			logger.Infof("generating private-public transport keys")
			priv, pub, err := state.GenerateKeys()
			if err != nil {
				isRecoverable = true
				return err
			}
			dkgState.TransportPrivateKey = priv
			dkgState.TransportPublicKey = pub

			err = dkgState.PersistState(txn)
			if err != nil {
				isRecoverable = false
				return err
			}
		} else {
			logger.Infof("private-public transport keys already defined")
		}
		return nil
	})

	if err != nil {
		return executorInterfaces.NewTaskErr(fmt.Sprintf("error during the preparation: %v", err), isRecoverable)
	}

	return nil
}

// Execute executes the task business logic
func (t *RegisterTask) Execute() ([]*types.Transaction, *executorInterfaces.TaskErr) {
	logger := t.GetLogger().WithField("method", "RegisterTask.Execute()")
	logger.Trace("initiate execution")

	eth := t.GetClient()
	ctx := t.GetCtx()
	block, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return nil, executorInterfaces.NewTaskErr(fmt.Sprintf("failed to get current height : %v", err), true)
	}

	dkgState := &state.DkgState{}
	err = t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, executorInterfaces.NewTaskErr(fmt.Sprintf("error loading dkgState: %v", err), false)
	}

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, executorInterfaces.NewTaskErr(fmt.Sprintf("getting txn opts failed: %v", err), true)
	}

	// Register
	logger.Infof("Registering  publicKey (%v) with ETHDKG", dkgUtils.FormatPublicKey(dkgState.TransportPublicKey))
	logger.Debugf("registering on block %v with public key: %v", block, dkgUtils.FormatPublicKey(dkgState.TransportPublicKey))
	txn, err := eth.Contracts().Ethdkg().Register(txnOpts, dkgState.TransportPublicKey)
	if err != nil {
		return nil, executorInterfaces.NewTaskErr(fmt.Sprintf("registering failed: %v", err), true)
	}

	return []*types.Transaction{txn}, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *RegisterTask) ShouldExecute() *executorInterfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "RegisterTask.ShouldExecute()")
	logger.Trace("should execute task")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return executorInterfaces.NewTaskErr(fmt.Sprintf("could not get dkgState with error %v", err), false)
	}

	eth := t.GetClient()
	ctx := t.GetCtx()
	if dkgState.Phase != state.RegistrationOpen {
		return executorInterfaces.NewTaskErr(fmt.Sprintf("phase %v different from RegistrationOpen", dkgState.Phase), false)
	}

	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return executorInterfaces.NewTaskErr(fmt.Sprintf("failed getting call options: %v", err), true)
	}

	status, err := state.CheckRegistration(eth.Contracts().Ethdkg(), logger, callOpts, dkgState.Account.Address, dkgState.TransportPublicKey)
	logger.Infof("registration status: %v", status)
	if err != nil {
		return executorInterfaces.NewTaskErr(fmt.Sprintf("failed to check registration %v", err), true)
	}
	if status == state.Registered || status == state.BadRegistration {
		return executorInterfaces.NewTaskErr(fmt.Sprintf("registration already occurred %v", status), false)
	}

	return nil
}
