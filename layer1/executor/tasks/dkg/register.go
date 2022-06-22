package dkg

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/MadBase/MadNet/layer1/ethereum"
	"github.com/MadBase/MadNet/layer1/executor/constants"
	"github.com/MadBase/MadNet/layer1/executor/tasks"
	"github.com/MadBase/MadNet/layer1/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/layer1/executor/tasks/dkg/utils"
	"github.com/MadBase/MadNet/layer1/transaction"
)

// RegisterTask contains required state for safely performing a registration
type RegisterTask struct {
	*tasks.BaseTask
}

// asserting that RegisterTask struct implements interface tasks.Task
var _ tasks.Task = &RegisterTask{}

// NewRegisterTask creates a background task that attempts to register with ETHDKG
func NewRegisterTask(start uint64, end uint64) *RegisterTask {
	return &RegisterTask{
		BaseTask: tasks.NewBaseTask(constants.RegisterTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the RegisterTask
// We construct our TransportPrivateKey and TransportPublicKey
// which will be used in the ShareDistribution phase for secure communication.
// These keys are *not* used otherwise.
// Also get the list of existing validators from the pool to assert accusation
// in later phases
func (t *RegisterTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debugf("preparing task")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf(constants.ErrorDuringPreparation, err), false)
	}

	if dkgState.TransportPrivateKey == nil ||
		dkgState.TransportPrivateKey.Cmp(big.NewInt(0)) == 0 {

		logger.Debug("generating private-public transport keys")
		// If this function fails, probably we got a bad random value. We can retry
		// later to get a new value.
		priv, pub, err := state.GenerateKeys()
		if err != nil {
			return tasks.NewTaskErr(fmt.Sprintf("failed to generate keys: %v", err), true)
		}
		dkgState.TransportPrivateKey = priv
		dkgState.TransportPublicKey = pub

		err = state.SaveDkgState(t.GetDB(), dkgState)
		if err != nil {
			return tasks.NewTaskErr(fmt.Sprintf(constants.ErrorDuringPreparation, err), false)
		}
	} else {
		logger.Debug("private-public transport keys already defined")
	}

	return nil
}

// Execute executes the task business logic
func (t *RegisterTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	eth := t.GetClient()
	block, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("failed to get current height : %v", err), true)
	}

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(constants.FailedGettingTxnOpts, err), true)
	}

	// Register
	logger.Infof("Registering  publicKey (%v) with ETHDKG", utils.FormatPublicKey(dkgState.TransportPublicKey))
	logger.Debugf("registering on block %v with public key: %v", block, utils.FormatPublicKey(dkgState.TransportPublicKey))
	txn, err := ethereum.GetContracts().Ethdkg().Register(txnOpts, dkgState.TransportPublicKey)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("registering failed: %v", err), true)
	}

	return txn, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *RegisterTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	if dkgState.Phase != state.RegistrationOpen {
		logger.Debugf("phase %v different from RegistrationOpen", dkgState.Phase)
		return false, nil
	}

	client := t.GetClient()
	callOpts, err := client.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(constants.FailedGettingCallOpts, err), true)
	}

	status, err := state.CheckRegistration(ethereum.GetContracts().Ethdkg(), logger, callOpts, dkgState.Account.Address, dkgState.TransportPublicKey)
	logger.Debugf("registration status: %v", status)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf("failed to check registration %v", err), true)
	}
	if status == state.Registered || status == state.BadRegistration {
		logger.Debug("registration already occurred")
		return false, nil
	}

	return true, nil
}
