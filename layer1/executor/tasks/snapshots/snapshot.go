package snapshots

import (
	"context"
	"fmt"
	dangerousRand "math/rand"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/snapshots/state"
)

// SnapshotTask pushes a snapshot to Ethereum.
type SnapshotTask struct {
	*tasks.BaseTask
	Height uint64
}

// asserting that SnapshotTask struct implements interface tasks.Task.
var _ tasks.Task = &SnapshotTask{}

func NewSnapshotTask(start, end, height uint64) *SnapshotTask {
	snapshotTask := &SnapshotTask{
		BaseTask: tasks.NewBaseTask(start, end, false, nil),
		Height:   height,
	}
	return snapshotTask
}

// Prepare prepares for work to be done in the SnapshotTask.
func (t *SnapshotTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()").WithField("AliceNetHeight", t.Height)
	logger.Debugf("preparing task")

	snapshotState, err := state.GetSnapshotState(t.GetDB())
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorDuringPreparation, err), false)
	}

	rawBClaims, err := snapshotState.BlockHeader.BClaims.MarshalBinary()
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf("unable to marshal block header for snapshot: %v", err), false)
	}

	snapshotState.RawBClaims = rawBClaims
	snapshotState.RawSigGroup = snapshotState.BlockHeader.SigGroup

	err = state.SaveSnapshotState(t.GetDB(), snapshotState)
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorDuringPreparation, err), false)
	}

	return nil
}

// Execute executes the task business logic.
func (t *SnapshotTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()").WithField("AliceNetHeight", t.Height)
	logger.Debug("initiate execution")

	snapshotState, err := state.GetSnapshotState(t.GetDB())
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()

	// todo: remove this after leader election
	dangerousRand.Seed(time.Now().UnixNano())
	n := dangerousRand.Intn(60) // n will be between 0 and 60
	select {
	case <-ctx.Done():
		return nil, tasks.NewTaskErr(ctx.Err().Error(), false)
	// wait some random time
	case <-time.After(time.Duration(n) * time.Second):
	}
	/////////////////////////////////////////////

	txnOpts, err := client.GetTransactionOpts(ctx, snapshotState.Account)
	if err != nil {
		// if it failed here, it means that we are not willing to pay the tx costs based on config or we
		// failed to retrieve tx fee data from the ethereum node
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingTxnOpts, err), true)
	}
	logger.Info("trying to commit snapshot")
	txn, err := t.GetContractsHandler().EthereumContracts().Snapshots().Snapshot(txnOpts, snapshotState.RawSigGroup, snapshotState.RawBClaims)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("failed to send snapshot: %v", err), true)
	}

	return txn, nil
}

// ShouldExecute checks if it makes sense to execute the task.
func (t *SnapshotTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()").WithField("AliceNetHeight", t.Height)
	logger.Debug("should execute task")

	snapshotState, err := state.GetSnapshotState(t.GetDB())
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()
	opts, err := client.GetCallOpts(ctx, snapshotState.Account)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingCallOpts, err), true)
	}

	height, err := t.GetContractsHandler().EthereumContracts().Snapshots().GetAliceNetHeightFromLatestSnapshot(opts)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf("failed to determine height: %v", err), true)
	}

	// This means the block height we want to snapshot is older than (or same as) what's already been snapshotted
	if snapshotState.BlockHeader.BClaims.Height != 0 && snapshotState.BlockHeader.BClaims.Height <= uint32(height.Uint64()) {
		logger.Debugf(
			"block height we want to snapshot height:%v is older than (or same as) what's already been snapshotted height:%v",
			snapshotState.BlockHeader.BClaims.Height,
			uint32(height.Uint64()),
		)
		return false, nil
	}

	return true, nil
}
