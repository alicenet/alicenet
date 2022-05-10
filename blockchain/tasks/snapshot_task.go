package tasks

import (
	"context"
	"errors"
	"math/big"
	dangerousRand "math/rand"
	"strings"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/sirupsen/logrus"
)

// SnapshotTask pushes a snapshot to Ethereum
type SnapshotTask struct {
	*Task
}

// asserting that SnapshotTask struct implements interface interfaces.Task
var _ interfaces.ITask = &SnapshotTask{}

type SnapshotState struct {
	sync.RWMutex
	account     accounts.Account
	blockHeader *objs.BlockHeader
	rawBclaims  []byte
	rawSigGroup []byte
	consensusDb *db.Database
}

// asserting that SnapshotState struct implements interface interfaces.ITaskState
var _ interfaces.ITaskState = &SnapshotState{}

func NewSnapshotTask(account accounts.Account, db *db.Database, bh *objs.BlockHeader, start uint64, end uint64) *SnapshotTask {
	return &SnapshotTask{
		Task: NewTask(&SnapshotState{
			account:     account,
			blockHeader: bh,
			consensusDb: db,
		}, start, end),
	}
}

func (t *SnapshotTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("CompletionTask Initialize()...")

	taskState, ok := t.State.(*SnapshotState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	rawBClaims, err := taskState.blockHeader.BClaims.MarshalBinary()
	if err != nil {
		logger.Errorf("Unable to marshal block header for snapshot: %v", err)
		return err
	}

	taskState.rawBclaims = rawBClaims
	taskState.rawSigGroup = taskState.blockHeader.SigGroup

	return nil
}

func (t *SnapshotTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *SnapshotTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *SnapshotTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {

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

	txnOpts, err := eth.GetTransactionOpts(ctx, taskState.account)
	if err != nil {
		// if it failed here, it means that we are not willing to pay the tx costs based on config or we
		// failed to retrieve tx fee data from the ethereum node
		logger.Debugf("Failed to generate transaction options: %v", err)
		return err
	}
	txn, err := eth.Contracts().Snapshots().Snapshot(txnOpts, taskState.rawSigGroup, taskState.rawBclaims)
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

func (t *SnapshotTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.RLock()
	defer t.RUnlock()

	taskState, ok := t.State.(*SnapshotState)
	if !ok {
		logger.Error("Invalid convertion of taskState object")
		return true
	}

	var height uint32
	err := taskState.consensusDb.View(func(txn *badger.Txn) error {
		bh, err := taskState.consensusDb.GetLastSnapshot(txn)
		if err != nil {
			return err
		}
		if bh == nil {
			return errors.New("invalid snapshot bh was read from the db")
		}
		height = bh.BClaims.Height
		return nil
	})
	if err != nil {
		logger.Debugf("Snapshot for height %v was not found on from db: %v", taskState.blockHeader.BClaims.Height, err)
		return true
	}

	// This means the block height we want to snapshot is older than (or same as) what's already been snapshotted
	if taskState.blockHeader.BClaims.Height != 0 && taskState.blockHeader.BClaims.Height < height {
		return false
	}

	return true
}

func (*SnapshotTask) DoDone(logger *logrus.Entry) {
}

func (t *SnapshotTask) GetExecutionData() interfaces.ITaskExecutionData {
	return t.Task
}
