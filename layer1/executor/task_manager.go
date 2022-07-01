package executor

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

type TasksManager struct {
	Transactions map[string]*types.Transaction `json:"transactions"`
	txWatcher    transaction.Watcher           `json:"-"`
	database     *db.Database                  `json:"-"`
	logger       *logrus.Entry                 `json:"-"`
}

// Creates a new TasksManager instance
func NewTaskManager(txWatcher transaction.Watcher, database *db.Database, logger *logrus.Entry) (*TasksManager, error) {
	taskManager := &TasksManager{
		Transactions: map[string]*types.Transaction{},
		txWatcher:    txWatcher,
		database:     database,
		logger:       logger,
	}

	err := taskManager.loadState()
	if err != nil {
		taskManager.logger.Warnf("could not find previous State: %v", err)
		if err != badger.ErrKeyNotFound {
			return nil, err
		}
	}

	return taskManager, nil
}

// main function to manage a task. It basically an abstraction to handle the
// task execution in a separate process.
func (tm *TasksManager) ManageTask(mainCtx context.Context, task tasks.Task, name string, taskId string, database *db.Database, logger *logrus.Entry, eth layer1.Client, taskResponseChan tasks.TaskResponseChan) {
	err := tm.processTask(mainCtx, task, name, taskId, database, logger, eth, taskResponseChan)
	task.Finish(err)
}

func (tm *TasksManager) processTask(mainCtx context.Context, task tasks.Task, name string, taskId string, database *db.Database, logger *logrus.Entry, eth layer1.Client, taskResponseChan tasks.TaskResponseChan) error {
	taskCtx, cf := context.WithCancel(mainCtx)
	defer cf()
	defer task.Close()
	err := task.Initialize(taskCtx, cf, database, logger, eth, name, taskId, taskResponseChan)
	if err != nil {
		return err
	}
	retryDelay := constants.MonitorRetryDelay
	isComplete := false
	if txn, present := tm.Transactions[task.GetId()]; present {
		isComplete, err = tm.checkCompletion(taskCtx, task, txn)
		if err != nil {
			return err
		}
	} else {
		err = prepareTask(taskCtx, task, retryDelay)
		if err != nil {
			// unrecoverable errors or ctx.done
			return err
		}
	}

	if !isComplete {
		err = tm.executeTask(taskCtx, task, retryDelay)
		if err != nil {
			// unrecoverable errors, staleTx errors or ctx.done
			return err
		}
	}

	// We got a successful receipt, removing from state
	delete(tm.Transactions, task.GetId())
	err = tm.persistState()
	if err != nil {
		return err
	}
	return nil
}

// prepareTask executes task preparation. We keep retrying until the task is
// killed, we get an unrecoverable error or we succeed
func prepareTask(ctx context.Context, task tasks.Task, retryDelay time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		taskErr := task.Prepare(ctx)
		// no errors or unrecoverable errors
		if taskErr == nil {
			return nil
		}
		if !taskErr.IsRecoverable() {
			return taskErr
		}
		err := sleepWithContext(ctx, retryDelay)
		if err != nil {
			return err
		}
	}
}

// executeTask executes task business logic. We keep retrying until the task is
// killed, we get an unrecoverable error or we succeed
func (tm *TasksManager) executeTask(ctx context.Context, task tasks.Task, retryDelay time.Duration) error {
	logger := task.GetLogger()
	for {
		hasToExecute, err := shouldExecute(ctx, task)
		if err != nil {
			return err
		}
		if hasToExecute {
			txn, taskErr := task.Execute(ctx)
			if taskErr != nil {
				if taskErr.IsRecoverable() {
					logger.Tracef("got a recoverable error during task.execute: %v", taskErr.Error())
					err := sleepWithContext(ctx, retryDelay)
					if err != nil {
						return err
					}
					continue
				}
				logger.Debugf("got a unrecoverable error during task.execute finishing execution err: %v", taskErr.Error())
				return taskErr
			}
			if txn != nil {
				logger.Debugf("got a successful txn: %v", txn.Hash().Hex())
				tm.Transactions[task.GetId()] = txn
				err := tm.persistState()
				if err != nil {
					return err
				}

				isComplete, err := tm.checkCompletion(ctx, task, txn)
				if err != nil {
					return err
				}
				if isComplete {
					return nil
				}
				continue
			} else {
				logger.Debug("Task returned no transaction, finishing")
				return nil
			}
		}
		return nil
	}
}

// checks if a task is complete. The function is going to subscribe a
// transaction in the transactionWatcher, and it will wait until it gets the
// receipt, the task is killed, or shouldExecute returns false.
func (tm *TasksManager) checkCompletion(ctx context.Context, task tasks.Task, txn *types.Transaction) (bool, error) {
	var err error
	var receipt *types.Receipt
	logger := task.GetLogger()
	receiptResponse, err := tm.txWatcher.Subscribe(ctx, txn, task.GetSubscribeOptions())
	if err != nil {
		// if we get an error here, it means that we have a corrupted txn we should
		// retry a transaction
		logger.Errorf("failed to subscribe tx with error: %s", err.Error())
		return false, nil
	}

	for {
		select {
		case <-ctx.Done():
			return true, ctx.Err()
		case <-time.After(constants.TaskManagerPoolingTime):
		}

		if receiptResponse.IsReady() {
			receipt, err = receiptResponse.GetReceiptBlocking(ctx)
			if err != nil {
				var txnStaleError *transaction.ErrTransactionStale
				if errors.As(err, &txnStaleError) {
					logger.Info("got a stale transaction and couldn't retry, finishing execution")
					return false, err
				}
				logger.Warnf("got a error while waiting for receipt %v, retrying execution", err)
				return false, nil
			}

			if receipt.Status == types.ReceiptStatusSuccessful {
				logger.Debug("got a successful receipt")
				return true, nil
			} else {
				logger.Warn("got a reverted receipt, retrying")
				return false, nil
			}
		}

		logger.Trace("receipt is not ready yet")
		hasToExecute, err := shouldExecute(ctx, task)
		if err != nil {
			return false, err
		}
		if !hasToExecute {
			return true, nil
		}
	}
}

// checks if a task should be executed. In case of recoverable errors the
// function is going to retry `constants.MonitorRetryCount` times. If it
// exhausted the number of retries, it sends true since there's no information.
// The function returns false in case of unrecoverable errors, or if it
// shouldn't execute a task. In case of no errors, the return value is true
// (default case to t.ShouldExecute to return that a task should be executed).
func shouldExecute(ctx context.Context, task tasks.Task) (bool, error) {
	logger := task.GetLogger()
	for i := uint64(1); i <= constants.MonitorRetryCount; i++ {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
		if hasToExecute, err := task.ShouldExecute(ctx); err != nil {
			if err.IsRecoverable() {
				logger.Tracef("got a recoverable error during task.ShouldExecute: %v", err)
				if err := sleepWithContext(ctx, constants.MonitorRetryDelay); err != nil {
					return false, err
				}
				continue
			}
			logger.Debugf("got a non recoverable error during task.ShouldExecute: %v", err)
			return false, err
		} else {
			logger.Tracef("should execute BaseTask: %v", hasToExecute)
			return hasToExecute, nil
		}
	}

	return true, nil
}

// persist task manager state to disk
func (tm *TasksManager) persistState() error {
	logger := logging.GetLogger("staterecover").WithField("State", "taskManager")
	rawData, err := json.Marshal(tm)
	if err != nil {
		return err
	}

	err = tm.database.Update(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskManagerState()
		logger.WithField("Key", string(key)).Debug("Saving state in the database")
		if err := utils.SetValue(txn, key, rawData); err != nil {
			logger.Error("Failed to set Value")
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := tm.database.Sync(); err != nil {
		logger.Error("Failed to set sync")
		return err
	}

	return nil
}

// load task's manager state from database
func (tm *TasksManager) loadState() error {
	logger := logging.GetLogger("staterecover").WithField("State", "taskManager")
	if err := tm.database.View(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskManagerState()
		logger.WithField("Key", string(key)).Debug("Loading state from database")
		rawData, err := utils.GetValue(txn, key)
		if err != nil {
			return err
		}

		err = json.Unmarshal(rawData, tm)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// synchronizing db state to disk
	if err := tm.database.Sync(); err != nil {
		logger.Error("Failed to set sync")
		return err
	}

	return nil

}

// sleeps a certain amount of time also checking the context. It fails in case
// the context is cancelled.
func sleepWithContext(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}
