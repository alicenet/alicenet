package snapshots

import (
	"context"
	"fmt"

	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/snapshots/state"
	"github.com/alicenet/alicenet/utils"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// SnapshotTask pushes a snapshot to Ethereum
type SnapshotTask struct {
	*tasks.BaseTask
	Height          uint64
	NumOfValidators int
	ValidatorIndex  int
}

// asserting that SnapshotTask struct implements interface tasks.Task
var _ tasks.Task = &SnapshotTask{}

func NewSnapshotTask(height uint64, numOfValidators int, validatorIndex int) *SnapshotTask {
	snapshotTask := &SnapshotTask{
		BaseTask:        tasks.NewBaseTask(0, 0, false, nil),
		Height:          height,
		NumOfValidators: numOfValidators,
		ValidatorIndex:  validatorIndex,
	}
	return snapshotTask
}

// Prepare prepares for work to be done in the SnapshotTask
func (t *SnapshotTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithFields(
		logrus.Fields{
			"method":          "Prepare()",
			"AliceNetHeight":  t.Height,
			"numOfValidators": t.NumOfValidators,
			"EthdkgIndex":     t.ValidatorIndex,
		},
	)
	logger.Debug("preparing task")

	snapshotState, err := state.GetSnapshotState(t.GetDB())
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorDuringPreparation, err), false)
	}

	rawBClaims, err := snapshotState.BlockHeader.BClaims.MarshalBinary()
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf("unable to marshal block header for snapshot: %v", err), false)
	}

	client := t.GetClient()
	opts, err := client.GetCallOpts(ctx, snapshotState.Account)
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingCallOpts, err), true)
	}

	height, err := ethereum.GetContracts().Snapshots().GetCommittedHeightFromLatestSnapshot(opts)
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf("failed to determine committed height: %v", err), true)
	}

	desperationFactor, err := ethereum.GetContracts().Snapshots().GetSnapshotDesperationFactor(opts)
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf("failed to determine desperation factor: %v", err), true)
	}

	desperationDelay, err := ethereum.GetContracts().Snapshots().GetSnapshotDesperationDelay(opts)
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf("failed to determine desperation delay: %v", err), true)
	}

	// TODO: ask Hunter panic?
	randHash, err := snapshotState.BlockHeader.BClaims.BlockHash()
	if err != nil {
		return tasks.NewTaskErr("failed to compute randHash", false)
	}

	snapshotState.RawBClaims = rawBClaims
	snapshotState.RawSigGroup = snapshotState.BlockHeader.SigGroup
	snapshotState.LastSnapshotHeight = int(height.Int64())
	snapshotState.DesperationDelay = int(desperationDelay.Int64())
	snapshotState.DesperationFactor = int(desperationFactor.Int64())
	snapshotState.RandomSeedHash = randHash

	err = state.SaveSnapshotState(t.GetDB(), snapshotState)
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorDuringPreparation, err), false)
	}

	return nil
}

// Execute executes the task business logic
func (t *SnapshotTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithFields(
		logrus.Fields{
			"method":          "Execute()",
			"AliceNetHeight":  t.Height,
			"numOfValidators": t.NumOfValidators,
			"EthdkgIndex":     t.ValidatorIndex,
		},
	)
	logger.Debug("initiate execution")

	snapshotState, err := state.GetSnapshotState(t.GetDB())
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()

	isLeading, err := utils.AmILeading(
		client,
		ctx,
		logger,
		snapshotState.LastSnapshotHeight,
		snapshotState.RandomSeedHash,
		t.NumOfValidators,
		t.ValidatorIndex,
		snapshotState.DesperationFactor,
		snapshotState.DesperationDelay,
	)
	if err != nil {
		return nil, tasks.NewTaskErr("error getting eth height for leader election", true)
	}
	if !isLeading {
		return nil, tasks.NewTaskErr("not leading MPK submission yet", true)
	}

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

// ShouldExecute checks if it makes sense to execute the task
func (t *SnapshotTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithFields(
		logrus.Fields{
			"method":          "ShouldExecute()",
			"AliceNetHeight":  t.Height,
			"numOfValidators": t.NumOfValidators,
			"EthdkgIndex":     t.ValidatorIndex,
		},
	)
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
