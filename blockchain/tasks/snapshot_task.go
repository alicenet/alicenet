package tasks

import (
	"context"
	"math/big"
	"sync"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/sirupsen/logrus"
)

// SnapshotTask pushes a snapshot to Ethereum
type SnapshotTask struct {
	sync.Mutex
	acct        accounts.Account
	epoch       *big.Int
	eth         blockchain.Ethereum
	logger      *logrus.Logger
	rawBclaims  []byte
	rawSigGroup []byte
}

// NewSnapshotTask creates a new task
func NewSnapshotTask(acct accounts.Account, logger *logrus.Logger, eth blockchain.Ethereum, epoch *big.Int, rawBclaims []byte, rawSigGroup []byte) *SnapshotTask {
	return &SnapshotTask{
		acct:        acct,
		epoch:       epoch,
		eth:         eth,
		logger:      logger,
		rawBclaims:  rawBclaims,
		rawSigGroup: rawSigGroup,
	}
}

// DoWork is the first attempt at distributing shares via ethdkg
func (t *SnapshotTask) DoWork(ctx context.Context) bool {
	t.logger.Info("DoWork() ...")
	return t.doTask(ctx)
}

// DoRetry is subsequent attempts at distributing shares via ethdkg
func (t *SnapshotTask) DoRetry(ctx context.Context) bool {
	t.logger.Info("DoRetry() ...")
	return t.doTask(ctx)
}

func (t *SnapshotTask) doTask(ctx context.Context) bool {

	t.Lock()
	defer t.Unlock()

	c := t.eth.Contracts()

	// Do the mechanics
	txnOpts, err := t.eth.GetTransactionOpts(ctx, t.acct)
	if err != nil {
		t.logger.Errorf("Could not create transaction for snapshot: %v", err)
		return false
	}

	txn, err := c.Validators.Snapshot(txnOpts, t.rawSigGroup, t.rawBclaims)
	if err != nil {
		t.logger.Errorf("Failed to take snapshot: %v", err)
		return false
	}
	t.eth.Queue().QueueTransaction(ctx, txn)

	receipt, err := t.eth.Queue().WaitTransaction(ctx, txn)
	if err != nil {
		t.logger.Errorf("Failed to retrieve snapshot receipt: %v", err)
		return false
	}

	if receipt == nil {
		t.logger.Error("missing snapshot receipt")
		return false
	}

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		t.logger.Errorf("snapshot status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
		return false
	}

	return true
}

// ShouldRetry checks if it makes sense to try again
func (t *SnapshotTask) ShouldRetry(ctx context.Context) bool {
	t.Lock()
	defer t.Unlock()

	c := t.eth.Contracts()

	callOpts := t.eth.GetCallOpts(ctx, t.acct)
	height, err := c.Validators.GetHeightFromSnapshot(callOpts, t.epoch)
	if err != nil {
		return true
	}

	t.logger.Infof("height of epoch %v: %v", t.epoch, height)

	return false // TODO check to see if snapshot exists
}

// DoDone creates a log entry saying task is complete
func (t *SnapshotTask) DoDone() {
	t.logger.Infof("done")
}
