package tasks

import (
	"context"
	"errors"
	dangerousRand "math/rand"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/core/types"
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

// asserting that SnapshotTask struct implements interface interfaces.Task
var _ interfaces.Task = &SnapshotTask{}

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

	var txn *types.Transaction
	retryReceiptWaitCount := uint64(0)
	maximumReceiptRetryAmount := uint64(6)
	isToWaitFinalityDelay := false
	for {
		dangerousRand.Seed(time.Now().UnixNano())
		n := dangerousRand.Intn(60) // n will be between 0 and 60
		select {
		case <-ctx.Done():
			return ctx.Err()
		// wait some random time
		case <-time.After(time.Duration(n) * time.Second):
		}

		var err error

		if !isToWaitFinalityDelay {
			retryReceiptWaitCount, err = t.tryCommitSnapshot(ctx, logger, eth, txn, retryReceiptWaitCount, maximumReceiptRetryAmount)
		}

		if err != nil {
			select {
			case <-ctx.Done():
				return err
			default:
			}
			logger.Debugf("Retrying snapshot after failed tx")
			continue
		}

		isToWaitFinalityDelay = true
		logger.Debugf("Waiting for finality delay")

		err = waitFinalityDelay(ctx, logger, eth)

		if err != nil {
			select {
			case <-ctx.Done():
				return err
			default:
			}
			logger.Debugf("Retrying snapshot after everything: %v", err)
			continue
		}

		// someone else already did the snapshot
		if !t.ShouldRetry(ctx, logger, eth) {
			logger.Debug("Snapshot already sent! Exiting!")
			return nil
		}

		txn = nil
		retryReceiptWaitCount = 0
		isToWaitFinalityDelay = false
	}
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

	height, err := eth.Contracts().Snapshots().GetAliceNetHeightFromSnapshot(opts, epoch)
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

func (*SnapshotTask) GetExecutionData() interface{} {
	return nil
}

func (t *SnapshotTask) tryCommitSnapshot(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, txn *types.Transaction, retryReceiptWaitCount uint64, maximumReceiptRetryAmount uint64) (uint64, error) {
	// do the actual snapshot
	txnOpts, err := eth.GetTransactionOpts(ctx, t.acct)
	if err != nil {
		logger.Debugf("Failed to generate transaction options: %v", err)
		return 0, err
	}
	if txn == nil || retryReceiptWaitCount > maximumReceiptRetryAmount {
		txn, err = eth.Contracts().Snapshots().Snapshot(txnOpts, t.rawSigGroup, t.rawBclaims)
		if err != nil {
			logger.Errorf("Failed to send snapshot: %v", err)
			// we don't return the Error because we want to wait the finality delay
			// to retry because maybe another validator did the tx
			return 0, nil
		}
	}

	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	if err != nil {
		logger.Debugf("Snapshot failed to retrieve receipt: %v", err)
		retryReceiptWaitCount++
		return retryReceiptWaitCount, err
	}

	if rcpt.Status != 1 {
		logger.Debugf("Snapshot receipt status != 1")
		// we don't return the Error because we want to wait the finality delay
		// to retry because maybe another validator did the tx
		return 0, nil
	}

	logger.Info("Snapshot tx succeeded! Waiting for confirmation delay!")
	return 0, nil
}

func waitFinalityDelay(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	// check/wait for finality delay
	subctx, cf := context.WithTimeout(ctx, 5*time.Second)
	defer cf()
	initialHeight, err := eth.GetCurrentHeight(subctx)
	if err != nil {
		logger.Debugf("Error to get current eth height")
		return err
	}

	finalityDelay := eth.GetFinalityDelay()
	isDone := false
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * 5):
		}

		isDone, err = checkCurrentHeight(ctx, logger, eth, initialHeight, finalityDelay)
		if err != nil {
			select {
			case <-ctx.Done():
				return err
			default:
			}
			logger.Debugf("Finality delay failed: %v", err)
			continue
		} else if isDone {
			return nil
		}
	}
}

func checkCurrentHeight(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, initialHeight uint64, finalityDelay uint64) (bool, error) {
	subctx, cf := context.WithTimeout(ctx, 5*time.Second)
	defer cf()
	testHeight, err := eth.GetCurrentHeight(subctx)
	if err != nil {
		logger.Debugf("Error to get test eth height")
		return false, err
	}
	logger.Debugf("Waiting for finality delay initial block: %v current block: %v", initialHeight, testHeight)

	if testHeight > initialHeight+finalityDelay {
		return true, nil
	}
	return false, nil
}
