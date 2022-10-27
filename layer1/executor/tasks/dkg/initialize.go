package dkg

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
)

// InitializeTask contains required state for safely performing a registration.
type InitializeTask struct {
	*tasks.BaseTask
}

// asserting that InitializeTask struct implements interface tasks.Task.
var _ tasks.Task = &InitializeTask{}

// NewInitializeTask creates a background task initializes ETHDKG.
func NewInitializeTask() *InitializeTask {
	return &InitializeTask{
		BaseTask: tasks.NewBaseTask(0, 0, false, nil),
	}
}

// Prepare prepares for work to be done in the InitializeTask
func (t *InitializeTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debugf("preparing task")

	return nil
}

// Execute executes the task business logic.
func (t *InitializeTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	eth := t.GetClient()
	txnOpts, err := eth.GetTransactionOpts(ctx, eth.GetDefaultAccount())
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingTxnOpts, err), true)
	}

	txn, err := t.GetContractsHandler().EthereumContracts().ValidatorPool().InitializeETHDKG(txnOpts)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("ETHDKG initialization failed: %v", err), true)
	}

	return txn, nil
}

// ShouldExecute checks if it makes sense to execute the task.
func (t *InitializeTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()
	callOpts, err := client.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingCallOpts, err), true)
	}

	isETHDKGRunning, err := t.GetContractsHandler().EthereumContracts().Ethdkg().IsETHDKGRunning(callOpts)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf("failed to check ETHDKG running state %v", err), true)
	}

	return !isETHDKGRunning, nil
}
