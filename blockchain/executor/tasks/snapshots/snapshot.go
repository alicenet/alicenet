package snapshots

import (
	"context"
	"math/big"
	dangerousRand "math/rand"
	"strings"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/sirupsen/logrus"
)

// SnapshotTask pushes a snapshot to Ethereum
type SnapshotTask struct {
	*objects.Task
}

// asserting that SnapshotTask struct implements interface interfaces.Task
var _ interfaces.ITask = &SnapshotTask{}

type SnapshotState struct {
	sync.RWMutex
	Account     accounts.Account
	RawBClaims  []byte
	RawSigGroup []byte
	BlockHeader *objs.BlockHeader
}

// asserting that SnapshotState struct implements interface interfaces.ITaskState
var _ interfaces.ITaskState = &SnapshotState{}

func NewSnapshotTask(account accounts.Account, bh *objs.BlockHeader, start uint64, end uint64, ctx context.Context, cancel context.CancelFunc) *SnapshotTask {
	snapshotTask := &SnapshotTask{
		Task: objects.NewTask(&SnapshotState{
			Account:     account,
			BlockHeader: bh,
		}, constants.SnapshotTaskName, start, end),
	}
	snapshotTask.SetContext(ctx, cancel)

	return snapshotTask
}

func (t *SnapshotTask) Initialize(ctx context.Context, logger *logrus.Entry, eth ethereum.Network) error {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("CompletionTask Initialize()...")

	taskState, ok := t.State.(*SnapshotState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	rawBClaims, err := taskState.BlockHeader.BClaims.MarshalBinary()
	if err != nil {
		logger.Errorf("Unable to marshal block header for snapshot: %v", err)
		return err
	}

	taskState.RawBClaims = rawBClaims
	taskState.RawSigGroup = taskState.BlockHeader.SigGroup

	return nil
}

func (t *SnapshotTask) DoWork(ctx context.Context, logger *logrus.Entry, eth ethereum.Network) error {
	return t.doTask(ctx, logger, eth)
}

func (t *SnapshotTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth ethereum.Network) error {
	return t.doTask(ctx, logger, eth)
}

func (t *SnapshotTask) doTask(ctx context.Context, logger *logrus.Entry, eth ethereum.Network) error {

	t.State.Lock()
	defer t.State.Unlock()

	taskState, ok := t.State.(*SnapshotState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	dangerousRand.Seed(time.Now().UnixNano())
	n := dangerousRand.Intn(60) // n will be between 0 and 60
	select {
	case <-ctx.Done():
		return ctx.Err()
	// wait some random time
	case <-time.After(time.Duration(n) * time.Second):
	}

	// someone else already did the snapshot
	if !t.ShouldRetry(ctx, logger, eth) {
		logger.Debug("Snapshot already sent! Exiting!")
		return nil
	}

	txnOpts, err := eth.GetTransactionOpts(ctx, taskState.Account)
	if err != nil {
		// if it failed here, it means that we are not willing to pay the tx costs based on config or we
		// failed to retrieve tx fee data from the ethereum node
		logger.Debugf("Failed to generate transaction options: %v", err)
		return err
	}
	txn, err := eth.Contracts().Snapshots().Snapshot(txnOpts, taskState.RawSigGroup, taskState.RawBClaims)
	if err != nil {
		logger.Debugf("Failed to send snapshot: %v", err)
		// TODO: ?we should ignore any revert on contract execution and wait the confirmation delay to try again
		if strings.Contains(err.Error(), "reverted") {
			return nil
		}
		return err
	}

	t.TxOpts.TxHashes = append(t.TxOpts.TxHashes, txn.Hash())
	t.TxOpts.GasFeeCap = txn.GasFeeCap()
	t.TxOpts.GasTipCap = txn.GasTipCap()
	t.TxOpts.Nonce = big.NewInt(int64(txn.Nonce()))

	logger.Info("Snapshot tx succeeded!")
	return nil
}

func (t *SnapshotTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth ethereum.Network) bool {

	t.RLock()
	defer t.RUnlock()

	taskState, ok := t.State.(*SnapshotState)
	if !ok {
		logger.Error("Invalid conversion of taskState object")
		return true
	}

	opts, err := eth.GetCallOpts(ctx, taskState.Account)
	if err != nil {
		logger.Errorf("SnapshotsTask.ShouldRetry() failed to get call options: %v", err)
		return true
	}

	height, err := eth.Contracts().Snapshots().GetAliceNetHeightFromLatestSnapshot(opts)
	if err != nil {
		logger.Errorf("Failed to determine height: %v", err)
		return true
	}

	// This means the block height we want to snapshot is older than (or same as) what's already been snapshotted
	if taskState.BlockHeader.BClaims.Height != 0 && taskState.BlockHeader.BClaims.Height < uint32(height.Uint64()) {
		return false
	}

	return true
}

func (*SnapshotTask) DoDone(logger *logrus.Entry) {
}

func (t *SnapshotTask) GetExecutionData() interfaces.ITaskExecutionData {
	return t.Task
}
