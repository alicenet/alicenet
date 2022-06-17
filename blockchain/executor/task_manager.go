package executor

import (
	"context"
	"time"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/constants"
	"github.com/sirupsen/logrus"
)

func ManageTask(ctx context.Context, task interfaces.ITask, database *db.Database, logger *logrus.Entry, eth ethereum.Network, taskResponseChan interfaces.ITaskResponseChan, txWatcher *transaction.Watcher) {
	taskCtx, taskCancelFunc := context.WithCancel(ctx)
	defer taskCancelFunc()
	taskLogger := logger.WithField("TaskName", task.GetName())
	err := task.Initialize(taskCtx, taskCancelFunc, database, taskLogger, eth, task.GetId(), taskResponseChan)
	if err != nil {
		task.Finish(err)
	}
	retryCount := int(constants.MonitorRetryCount)
	retryDelay := constants.MonitorRetryDelay

	err = prepareTask(ctx, task, retryCount, retryDelay)
	if err != nil {
		// unrecoverable errors, recoverable errors but we exhausted all attempts or
		// ctx.done
		task.Finish(err)
	}

	err = executeTask(ctx, task, retryCount, retryDelay, txWatcher)
	task.Finish(err)
}

// prepareTask executes task preparation
func prepareTask(ctx context.Context, task interfaces.ITask, retryCount int, retryDelay time.Duration) error {
	var taskErr *interfaces.TaskErr
	for count := 0; count < retryCount; count++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		taskErr = task.Prepare(ctx)
		// no errors or unrecoverable errors break
		if taskErr == nil || !taskErr.IsRecoverable() {
			break
		}
		err := sleepWithContext(ctx, retryDelay)
		if err != nil {
			return err
		}
	}
	// return taskErr in case is nil, nonRecoverable or Recoverable after exhausted
	// all retry attempts
	return taskErr
}

// executeTask executes task business logic
func executeTask(ctx context.Context, task interfaces.ITask, retryCount int, retryDelay time.Duration, txWatcher *transaction.Watcher) error {
	var taskErr *interfaces.TaskErr
	for count := 0; count < retryCount; count++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		txn, taskErr := task.Execute(ctx)
		if taskErr != nil {
			if !taskErr.IsRecoverable() {
				err := sleepWithContext(ctx, retryDelay)
				if err != nil {
					return err
				}
				continue
			}
			return taskErr
		}
		if txn == nil {
			// there was no transaction to be executed
			return nil
		}

		respChan, err := txWatcher.Subscribe(ctx, txn, task.GetSubscribeOptions())
		if err != nil {
			// if we get an error here, it means that we have a corrupted txn we should
			// retry a transaction
			task.GetLogger().Errorf("failed to subscribe tx with error: %s", err.Error())
			continue
		}
	}

	return taskErr
}

func shouldExecute(ctx context.Context, task interfaces.ITask) bool {
	// Make sure we're in the right block range to continue
	currentBlock, err := task.GetClient().GetCurrentHeight(ctx)
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
				if err := sleepWithContext(ctx, constants.MonitorRetryDelay); err != nil {
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
