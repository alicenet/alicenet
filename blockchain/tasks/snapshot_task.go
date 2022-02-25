package tasks

import (
	"context"
	"errors"
	"sync"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/sirupsen/logrus"
)

// SnapshotTask pushes a snapshot to Ethereum
type SnapshotTask struct {
	sync.RWMutex
	acct        accounts.Account
	BlockHeader *objs.BlockHeader
	rawBclaims  []byte
	rawSigGroup []byte
}

func NewSnapshotTask(account accounts.Account) *SnapshotTask {
	return &SnapshotTask{
		acct: account,
	}
}

func (t *SnapshotTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, _ interface{}) error {

	if t.BlockHeader == nil {
		return errors.New("BlockHeader must be assigned before initializing")
	}

	rawBClaims, err := t.BlockHeader.BClaims.MarshalBinary()
	if err != nil {
		logger.Errorf("Unable to marshal block header: %v", err)
		return err
	}

	t.Lock()
	defer t.Unlock()

	t.rawBclaims = rawBClaims
	t.rawSigGroup = t.BlockHeader.SigGroup

	return nil
}
func (t *SnapshotTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *SnapshotTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *SnapshotTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {

	t.RLock()
	defer t.RUnlock()

	txnOpts, err := eth.GetTransactionOpts(ctx, t.acct)
	if err != nil {
		logger.Warnf("Failed to generate transaction options: %v", err)
		return nil
	}

	txn, err := eth.Contracts().Snapshots().Snapshot(txnOpts, t.rawSigGroup, t.rawBclaims)
	if err != nil {
		logger.Warnf("Snapshot failed: %v", err)
		return nil
	} else {
		rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
		if err != nil {
			logger.Warnf("Snapshot failed to retreive receipt: %v", err)
			return nil
		}

		if rcpt.Status != 1 {
			logger.Warnf("Snapshot receipt status != 1")
			return nil
		}

		logger.Info("Snapshot succeeded")
	}

	return nil
}

func (t *SnapshotTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.RLock()
	defer t.RUnlock()

	opts := eth.GetCallOpts(ctx, t.acct)

	epoch, err := eth.Contracts().Snapshots().GetEpoch(opts)
	if err != nil {
		logger.Errorf("Failed to determine current epoch: %v", err)
		return true
	}

	height, err := eth.Contracts().Snapshots().GetMadnetHeightFromSnapshot(opts, epoch)
	if err != nil {
		logger.Errorf("Failed to determine height: %v", err)
		return true
	}

	// This means the block height we want to snapshot is older than (or same as) what's already been snapshotted
	if t.BlockHeader.BClaims.Height != 0 && t.BlockHeader.BClaims.Height < uint32(height.Uint64()) {
		return false
	}

	return true
}

func (*SnapshotTask) DoDone(logger *logrus.Entry) {
}
