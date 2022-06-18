package executor

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/MadBase/MadNet/constants/dbprefix"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/constants"
	"github.com/sirupsen/logrus"
)

type TasksManager struct {
	Transactions map[string]*types.Transaction `json:"transactions"`
	txWatcher    *transaction.Watcher          `json:"-"`
	database     *db.Database                  `json:"-"`
	logger       *logrus.Entry                 `json:"-"`
}

// Creates a new TasksManager instance
func NewTaskManager(txWatcher *transaction.Watcher, database *db.Database, logger *logrus.Entry) (*TasksManager, error) {
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
func (tm *TasksManager) ManageTask(mainCtx context.Context, task interfaces.ITask, database *db.Database, logger *logrus.Entry, eth ethereum.Network, taskResponseChan interfaces.ITaskResponseChan) {
	var err error
	taskCtx, cf := context.WithCancel(mainCtx)
	defer cf()
	defer task.Close()
	defer task.Finish(err)

	taskLogger := logger.WithField("TaskName", task.GetName())
	err = task.Initialize(taskCtx, cf, database, taskLogger, eth, task.GetId(), taskResponseChan)
	if err != nil {
		return
	}
	retryDelay := constants.MonitorRetryDelay

	isComplete := false
	if txn, present := tm.Transactions[task.GetId()]; present {
		isComplete, err = tm.checkCompletion(taskCtx, task, txn)
		if err != nil {
			return
		}
	} else {
		err = prepareTask(taskCtx, task, retryDelay)
		if err != nil {
			// unrecoverable errors or ctx.done
			return
		}
	}

	if !isComplete {
		err = tm.executeTask(taskCtx, task, retryDelay)
		if err != nil {
			// unrecoverable errors, staleTx errors or ctx.done
			return
		}
	}

	// We got a successful receipt, removing from state
	delete(tm.Transactions, task.GetId())
	err = tm.persistState()
	if err != nil {
		return
	}
}

// prepareTask executes task preparation. We keep retrying until the task is
// killed, we get an unrecoverable error or we succeed
func prepareTask(ctx context.Context, task interfaces.ITask, retryDelay time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		taskErr := task.Prepare(ctx)
		// no errors or unrecoverable errors
		if taskErr == nil || !taskErr.IsRecoverable() {
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
func (tm *TasksManager) executeTask(ctx context.Context, task interfaces.ITask, retryDelay time.Duration) error {
	hasToExecute, err := shouldExecute(task.GetCtx(), task)
	if err != nil {
		return err
	}
	if hasToExecute {
		for {
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
			if txn != nil {
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
			} else {
				return nil
			}
		}
	}

	return nil
}

// checks if a task is complete. The function is going to subscribe a
// transaction in the transactionWatcher and it will wait until it gets the
// receipt, the task is killed, or shouldExecute returns false.
func (tm *TasksManager) checkCompletion(ctx context.Context, task interfaces.ITask, txn *types.Transaction) (bool, error) {
	var err error
	var receipt *types.Receipt
	isComplete := false
	receiptResponse, err := tm.txWatcher.Subscribe(ctx, txn, task.GetSubscribeOptions())
	if err != nil {
		// if we get an error here, it means that we have a corrupted txn we should
		// retry a transaction
		tm.logger.Errorf("failed to subscribe tx with error: %s", err.Error())
		return isComplete, nil
	}

	for {
		select {
		case <-ctx.Done():
			isComplete = true
			return isComplete, ctx.Err()
		case <-time.After(constants.TaskManagerPoolingTime):
		}

		if receiptResponse.IsReady() {
			receipt, err = receiptResponse.GetReceiptBlocking(ctx)
			if err != nil {
				var txnStaleError *transaction.ErrTransactionStale
				if errors.As(err, &txnStaleError) {
					return isComplete, err
				}
				break
			}

			if receipt.Status == types.ReceiptStatusSuccessful {
				isComplete = true
				return isComplete, nil
			}
		}

		hasToExecute, err := shouldExecute(task.GetCtx(), task)
		if err != nil {
			return isComplete, err
		}
		if !hasToExecute {
			isComplete = true
			return isComplete, nil
		}
	}

	return isComplete, nil
}

// checks if a task should be executed. In case of recoverable errors the
// function is going to retry `constants.MonitorRetryCount` times. If it
// exhausted the number of retries, it sends true since there's no information.
// The function returns false in case of unrecoverable errors, or if it
// shouldn't execute a task. In case of no errors, the return value is true
// (default case to t.ShouldExecute to return that a task should be executed).
func shouldExecute(ctx context.Context, task interfaces.ITask) (bool, error) {
	for i := uint64(1); i <= constants.MonitorRetryCount; i++ {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
		if err := task.ShouldExecute(ctx); err != nil {
			if err.IsRecoverable() {
				task.GetLogger().Tracef("got a recoverable error during task should execute: %v", err)
				if err := sleepWithContext(ctx, constants.MonitorRetryDelay); err != nil {
					return false, err
				}
				continue
			}
			task.GetLogger().Debugf("got a non recoverable error during task should execute: %v", err)
			return false, nil
		}
	}

	return true, nil
}

// persist task manager state to disk
func (tm *TasksManager) persistState() error {
	rawData, err := json.Marshal(tm)
	if err != nil {
		return err
	}

	err = tm.database.Update(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskManagerState()
		tm.logger.WithField("Key", string(key)).Infof("Saving state")
		if err := utils.SetValue(txn, key, rawData); err != nil {
			tm.logger.Error("Failed to set Value")
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := tm.database.Sync(); err != nil {
		tm.logger.Error("Failed to set sync")
		return err
	}

	return nil
}

// load task's manager state from database
func (tm *TasksManager) loadState() error {

	if err := tm.database.View(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskManagerState()
		tm.logger.WithField("Key", string(key)).Infof("Looking up state")
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
		tm.logger.Error("Failed to set sync")
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
