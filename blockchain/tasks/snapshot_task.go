package tasks

import (
	"context"
	"errors"
	dangerousRand "math/rand"
	"strings"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/sirupsen/logrus"
)

// SnapshotTask pushes a snapshot to Ethereum
type SnapshotTask struct {
	sync.RWMutex
	acct        accounts.Account
	blockHeader *objs.BlockHeader
	rawBclaims  []byte
	rawSigGroup []byte
	consensusDb *db.Database
}

// asserting that SnapshotTask struct implements interface interfaces.Task
var _ interfaces.Task = &SnapshotTask{}

func NewSnapshotTask(account accounts.Account, db *db.Database, bh *objs.BlockHeader) *SnapshotTask {
	return &SnapshotTask{
		acct:        account,
		blockHeader: bh,
		consensusDb: db,
	}
}

func (t *SnapshotTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, _ interface{}) error {

	rawBClaims, err := t.blockHeader.BClaims.MarshalBinary()
	if err != nil {
		logger.Errorf("Unable to marshal block header for snapshot: %v", err)
		return err
	}

	t.Lock()
	defer t.Unlock()

	t.rawBclaims = rawBClaims
	t.rawSigGroup = t.blockHeader.SigGroup

	return nil
}

func (t *SnapshotTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *SnapshotTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *SnapshotTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {

	t.Lock()
	defer t.Unlock()

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

	txnOpts, err := eth.GetTransactionOpts(ctx, t.acct)
	if err != nil {
		// if it failed here, it means that we are not willing to pay the tx costs based on config or we
		// failed to retrieve tx fee data from the ethereum node
		logger.Debugf("Failed to generate transaction options: %v", err)
		return err
	}
	txn, err := eth.Contracts().Snapshots().Snapshot(txnOpts, t.rawSigGroup, t.rawBclaims)
	if err != nil {
		logger.Debugf("Failed to send snapshot: %v", err)
		// TODO: ?we should ignore any revert on contract execution and wait the confirmation delay to try again
		if strings.Contains(err.Error(), "reverted") {
			return nil
		}
		return err
	}

	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	if err != nil {
		logger.Debugf("Snapshot failed to retrieve receipt: %v", err)
		return err
	}

	if rcpt.Status != 1 {
		logger.Debugf("Snapshot receipt status != 1")
		return err
	}

	logger.Info("Snapshot tx succeeded!")
	return nil
}

func (t *SnapshotTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.RLock()
	defer t.RUnlock()

	var height uint32
	err := t.consensusDb.View(func(txn *badger.Txn) error {
		bh, err := t.consensusDb.GetLastSnapshot(txn)
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
		logger.Debugf("Snapshot for height %v was not found on from db: %v", t.blockHeader.BClaims.Height, err)
		return true
	}

	// This means the block height we want to snapshot is older than (or same as) what's already been snapshotted
	if t.blockHeader.BClaims.Height != 0 && t.blockHeader.BClaims.Height < height {
		return false
	}

	return true
}

func (*SnapshotTask) DoDone(logger *logrus.Entry) {
}

func (*SnapshotTask) GetExecutionData() interface{} {
	return nil
}
