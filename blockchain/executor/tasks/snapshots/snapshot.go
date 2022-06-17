package snapshots

import (
	"context"
	"fmt"
	dangerousRand "math/rand"
	"time"

	"github.com/MadBase/MadNet/blockchain/executor/tasks/snapshots/state"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
)

// SnapshotTask pushes a snapshot to Ethereum
type SnapshotTask struct {
	*objects.Task
}

// asserting that SnapshotTask struct implements interface interfaces.Task
var _ interfaces.ITask = &SnapshotTask{}

func NewSnapshotTask(start uint64, end uint64) *SnapshotTask {
	snapshotTask := &SnapshotTask{
		Task: objects.NewTask(constants.SnapshotTaskName, start, end, false, nil),
	}
	return snapshotTask
}

// Prepare prepares for work to be done in the SnapshotTask
func (t *SnapshotTask) Prepare(ctx context.Context) *interfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debugf("preparing task")

	snapshotState := &state.SnapshotState{}
	err := t.GetDB().Update(func(txn *badger.Txn) error {
		err := snapshotState.LoadState(txn)
		if err != nil {
			return err
		}

		rawBClaims, err := snapshotState.BlockHeader.BClaims.MarshalBinary()
		if err != nil {
			logger.Errorf("unable to marshal block header for snapshot: %v", err)
			return err
		}

		snapshotState.RawBClaims = rawBClaims
		snapshotState.RawSigGroup = snapshotState.BlockHeader.SigGroup

		err = snapshotState.PersistState(txn)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// all errors are not recoverable
		return interfaces.NewTaskErr(fmt.Sprintf(constants.ErrorDuringPreparation, err), false)
	}

	return nil
}

// Execute executes the task business logic
func (t *SnapshotTask) Execute(ctx context.Context) (*types.Transaction, *interfaces.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	snapshotState := &state.SnapshotState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := snapshotState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()

	// todo: remove this after leader election
	dangerousRand.Seed(time.Now().UnixNano())
	n := dangerousRand.Intn(60) // n will be between 0 and 60
	select {
	case <-ctx.Done():
		return nil, interfaces.NewTaskErr(fmt.Sprintf("task killed by ctx: %v", ctx.Err()), false)
	// wait some random time
	case <-time.After(time.Duration(n) * time.Second):
	}
	/////////////////////////////////////////////

	txnOpts, err := client.GetTransactionOpts(ctx, snapshotState.Account)
	if err != nil {
		// if it failed here, it means that we are not willing to pay the tx costs based on config or we
		// failed to retrieve tx fee data from the ethereum node
		return nil, interfaces.NewTaskErr(fmt.Sprintf(constants.FailedGettingTxnOpts, err), true)
	}
	logger.Info("trying to commit snapshot")
	txn, err := client.Contracts().Snapshots().Snapshot(txnOpts, snapshotState.RawSigGroup, snapshotState.RawBClaims)
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf("failed to send snapshot: %v", err), true)
	}

	return txn, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *SnapshotTask) ShouldExecute(ctx context.Context) *interfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	snapshotState := &state.SnapshotState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := snapshotState.LoadState(txn)
		return err
	})
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()
	opts, err := client.GetCallOpts(ctx, snapshotState.Account)
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf(constants.FailedGettingCallOpts, err), true)
	}

	height, err := client.Contracts().Snapshots().GetAliceNetHeightFromLatestSnapshot(opts)
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf("failed to determine height: %v", err), true)
	}

	// This means the block height we want to snapshot is older than (or same as) what's already been snapshotted
	if snapshotState.BlockHeader.BClaims.Height != 0 && snapshotState.BlockHeader.BClaims.Height < uint32(height.Uint64()) {
		return interfaces.NewTaskErr(fmt.Sprint("block height we want to snapshot is older than (or same as) what's already been snapshotted"), false)
	}

	return nil
}
