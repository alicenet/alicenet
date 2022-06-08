package executor

import (
	"context"
	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/constants"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"time"
)

func ManageTask(ctx context.Context, task interfaces.ITask, database *db.Database, logger *logrus.Entry, eth ethereum.Network, taskResponseChan interfaces.ITaskResponseChan) {
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

	txns, err := executeTask(task, retryCount, retryDelay)
	if err != nil {
		task.Finish(err)
	}

	if len(txns) > 0 {
		//TODO: add interaction with txWatcher
	}
}

// prepareTask executes task preparation
func prepareTask(task interfaces.ITask, retryCount int, retryDelay time.Duration) error {
	var count int
	var err error
	ctx := task.GetCtx()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		err = task.Prepare()
		for err != nil && count < retryCount {
			err = sleepWithContext(ctx, retryDelay)
			if err != nil {
				return err
			}

			err = task.Prepare()
			count++
		}
	}

	return err
}

// executeTask executes task business logic
func executeTask(task interfaces.ITask, retryCount int, retryDelay time.Duration) ([]*types.Transaction, error) {
	var count int
	var err error
	txns := make([]*types.Transaction, 0)
	ctx := task.GetCtx()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if shouldExecute(task) {
			txns, err = task.Execute()
			for err != nil && count < retryCount && shouldExecute(task) {
				err = sleepWithContext(ctx, retryDelay)
				if err != nil {
					return nil, err
				}

				txns, err = task.Execute()
				count++
			}
		}
	}

	return txns, err
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

	return task.ShouldExecute()
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}
