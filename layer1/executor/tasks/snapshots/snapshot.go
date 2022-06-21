package snapshots

import (
	"context"
	"fmt"
	dangerousRand "math/rand"
	"time"

	"github.com/MadBase/MadNet/layer1/ethereum"
	"github.com/MadBase/MadNet/layer1/executor/tasks"
	"github.com/MadBase/MadNet/layer1/executor/tasks/snapshots/state"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/MadBase/MadNet/layer1/executor/constants"
)

// SnapshotTask pushes a snapshot to Ethereum
type SnapshotTask struct {
	*tasks.BaseTask
}

// asserting that SnapshotTask struct implements interface tasks.Task
var _ tasks.Task = &SnapshotTask{}

func NewSnapshotTask(start uint64, end uint64) *SnapshotTask {
	snapshotTask := &SnapshotTask{
		BaseTask: tasks.NewBaseTask(constants.SnapshotTaskName, start, end, false, nil),
	}
	return snapshotTask
}

// Prepare prepares for work to be done in the SnapshotTask
func (t *SnapshotTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debugf("preparing task")

	snapshotState, err := state.GetSnapshotState(t.GetDB())
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf(constants.ErrorDuringPreparation, err), false)
	}

	rawBClaims, err := snapshotState.BlockHeader.BClaims.MarshalBinary()
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf("unable to marshal block header for snapshot: %v", err), false)
	}

	snapshotState.RawBClaims = rawBClaims
	snapshotState.RawSigGroup = snapshotState.BlockHeader.SigGroup

	err = state.SaveSnapshotState(t.GetDB(), snapshotState)
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf(constants.ErrorDuringPreparation, err), false)
	}

	return nil
}

// Execute executes the task business logic
func (t *SnapshotTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	snapshotState, err := state.GetSnapshotState(t.GetDB())
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()

	// todo: remove this after leader election
	dangerousRand.Seed(time.Now().UnixNano())
	n := dangerousRand.Intn(60) // n will be between 0 and 60
	select {
	case <-ctx.Done():
		return nil, tasks.NewTaskErr(fmt.Sprintf("task killed by ctx: %v", ctx.Err()), false)
	// wait some random time
	case <-time.After(time.Duration(n) * time.Second):
	}
	/////////////////////////////////////////////

	txnOpts, err := client.GetTransactionOpts(ctx, snapshotState.Account)
	if err != nil {
		// if it failed here, it means that we are not willing to pay the tx costs based on config or we
		// failed to retrieve tx fee data from the ethereum node
		return nil, tasks.NewTaskErr(fmt.Sprintf(constants.FailedGettingTxnOpts, err), true)
	}
	logger.Info("trying to commit snapshot")
	txn, err := ethereum.GetContracts().Snapshots().Snapshot(txnOpts, snapshotState.RawSigGroup, snapshotState.RawBClaims)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("failed to send snapshot: %v", err), true)
	}

	return txn, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *SnapshotTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	snapshotState, err := state.GetSnapshotState(t.GetDB())
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()
	opts, err := client.GetCallOpts(ctx, snapshotState.Account)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(constants.FailedGettingCallOpts, err), true)
	}

	height, err := ethereum.GetContracts().Snapshots().GetAliceNetHeightFromLatestSnapshot(opts)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf("failed to determine height: %v", err), true)
	}

	// This means the block height we want to snapshot is older than (or same as) what's already been snapshotted
	if snapshotState.BlockHeader.BClaims.Height != 0 && snapshotState.BlockHeader.BClaims.Height < uint32(height.Uint64()) {
		logger.Debugf(
			"block height we want to snapshot height:%v is older than (or same as) what's already been snapshotted height:%v",
			snapshotState.BlockHeader.BClaims.Height,
			uint32(height.Uint64()),
		)
		return false, nil
	}

	return true, nil
}
