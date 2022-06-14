package executor

import (
	"context"
	"time"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/constants"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

func ManageTask(ctx context.Context, task interfaces.ITask, database *db.Database, logger *logrus.Entry, eth ethereum.Network, taskResponseChan interfaces.ITaskResponseChan, txWatcher *transaction.Watcher) {
	taskCtx, taskCancelFunc := context.WithCancel(ctx)
	taskLogger := logger.WithField("TaskName", task.GetName())

	task.Initialize(taskCtx, taskCancelFunc, database, taskLogger, eth, task.GetId(), taskResponseChan)
	defer task.Close()

	retryCount := int(constants.MonitorRetryCount)
	retryDelay := constants.MonitorRetryDelay

	err := prepareTask(task, retryCount, retryDelay)
	if err != nil {
		task.Finish(err)
	}

	err = executeTask(task, retryCount, retryDelay, txWatcher)
	task.Finish(err)
}

// prepareTask executes task preparation
func prepareTask(task interfaces.ITask, retryCount int, retryDelay time.Duration) error {
	var count int
	var err error
	ctx := task.GetCtx()

Loop:
	for count < retryCount {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = task.Prepare()
			if err != nil {
				err = sleepWithContext(ctx, retryDelay)
				if err != nil {
					return err
				}
				count++
				break
			}
			break Loop
		}
	}

	return err
}

// executeTask executes task business logic
func executeTask(task interfaces.ITask, retryCount int, retryDelay time.Duration, txWatcher *transaction.Watcher) error {
	var count int
	var success bool
	var err error
	var txns []*types.Transaction
	ctx := task.GetCtx()

	for !success && shouldExecute(task) { //todo:
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			txns, err = task.Execute()
			if err != nil {
				err = sleepWithContext(ctx, retryDelay)
				if err != nil {
					return err
				}
				count++
				break
			}

			success, err = watchForTransactions(task.GetCtx(), txns, txWatcher, task.GetAllowTxFeeAutoReplacement())
			if err != nil {
				task.GetLogger().Errorf("failed to get receipts with error: %s", err.Error())
			}
		}
	}

	return err
}

func watchForTransactions(ctx context.Context, txns []*types.Transaction, txWatcher *transaction.Watcher, allowTxFeeAutoReplacement bool) (bool, error) {
	if txns == nil || len(txns) == 0 {
		return true, nil
	}

	for _, txn := range txns {
		respChan, err := txWatcher.Subscribe(ctx, txn, allowTxFeeAutoReplacement)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func shouldExecute(task interfaces.ITask) bool {
	// Make sure we're in the right block range to continue
	currentBlock, err := task.GetEth().GetCurrentHeight(task.GetCtx())
	if err != nil {
		// This probably means an endpoint issue, so we have to try again
		task.GetLogger().Warnf("could not check current height of chain: %v", err)
		return true
	}

	end := task.GetEnd()
	if end > 0 && end < currentBlock {
		return false
	}
	retryCount := int(constants.MonitorRetryCount)
	for i := 1; i <= retryCount; i++ {
		if err := task.ShouldExecute(); err != nil {
			if err.IsRecoverable() {
				task.GetLogger().Tracef("got a recoverable error during task should execute: %v", err)
				if err := sleepWithContext(task.GetCtx(), constants.MonitorRetryDelay); err != nil {
					return false
				}
				continue
			}
			task.GetLogger().Debugf("got a non recoverable error during task should execute: %v", err)
			return false
		}
	}

	return true
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}
