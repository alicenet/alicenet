package snapshots

import (
	dangerousRand "math/rand"
	"strings"
	"time"

	dkgUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
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
func (t *SnapshotTask) Prepare() *executorInterfaces.TaskErr {
	logger := t.GetLogger()
	logger.Info("CompletionTask Initialize()...")

	snapshotState := &state.SnapshotState{}
	err := t.GetDB().Update(func(txn *badger.Txn) error {
		err := snapshotState.LoadState(txn)
		if err != nil {
			return err
		}

		rawBClaims, err := snapshotState.BlockHeader.BClaims.MarshalBinary()
		if err != nil {
			logger.Errorf("Unable to marshal block header for snapshot: %v", err)
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
		return dkgUtils.LogReturnErrorf(logger, "SnapshotTask.Prepare(): error during the preparation: %v", err)
	}

	return nil
}

// Execute executes the task business logic
func (t *SnapshotTask) Execute() ([]*types.Transaction, *executorInterfaces.TaskErr) {
	logger := t.GetLogger()
	logger.Info("SnapshotTask Execute()...")

	snapshotState := &state.SnapshotState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := snapshotState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "could not get snapshotState with error %v", err)
	}

	eth := t.GetClient()
	ctx := t.GetCtx()
	dangerousRand.Seed(time.Now().UnixNano())
	n := dangerousRand.Intn(60) // n will be between 0 and 60
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	// wait some random time
	case <-time.After(time.Duration(n) * time.Second):
	}

	// someone else already did the snapshot
	if !t.ShouldExecute() {
		logger.Debug("Snapshot already sent! Exiting!")
		return nil, nil
	}

	txnOpts, err := eth.GetTransactionOpts(ctx, snapshotState.Account)
	if err != nil {
		// if it failed here, it means that we are not willing to pay the tx costs based on config or we
		// failed to retrieve tx fee data from the ethereum node
		logger.Debugf("Failed to generate transaction options: %v", err)
		return nil, err
	}
	txn, err := eth.Contracts().Snapshots().Snapshot(txnOpts, snapshotState.RawSigGroup, snapshotState.RawBClaims)
	if err != nil {
		logger.Debugf("Failed to send snapshot: %v", err)
		// TODO: ?we should ignore any revert on contract execution and wait the confirmation delay to try again
		if strings.Contains(err.Error(), "reverted") {
			return nil, nil
		}
		return nil, err
	}

	logger.Info("Snapshot tx succeeded!")
	return []*types.Transaction{txn}, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *SnapshotTask) ShouldExecute() (bool, *executorInterfaces.TaskErr) {
	logger := t.GetLogger()
	logger.Info("SnapshotTask ShouldExecute()...")

	snapshotState := &state.SnapshotState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := snapshotState.LoadState(txn)
		return err
	})
	if err != nil {
		logger.Errorf("could not get snapshotState with error %v", err)
		return true
	}

	eth := t.GetClient()
	ctx := t.GetCtx()
	opts, err := eth.GetCallOpts(ctx, snapshotState.Account)
	if err != nil {
		logger.Errorf("SnapshotsTask.ShouldExecute() failed to get call options: %v", err)
		return true
	}

	height, err := eth.Contracts().Snapshots().GetAliceNetHeightFromLatestSnapshot(opts)
	if err != nil {
		logger.Errorf("Failed to determine height: %v", err)
		return true
	}

	// This means the block height we want to snapshot is older than (or same as) what's already been snapshotted
	if snapshotState.BlockHeader.BClaims.Height != 0 && snapshotState.BlockHeader.BClaims.Height < uint32(height.Uint64()) {
		return false
	}

	return true
}
