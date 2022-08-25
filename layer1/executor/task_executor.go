package executor

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
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

type TaskExecutor struct {
	sync.RWMutex
	TxsBackup map[string]*types.Transaction `json:"transactionsBackup"`
	closeChan chan struct{}                 `json:"-"`
	closeOnce sync.Once                     `json:"-"`
	txWatcher transaction.Watcher           `json:"-"`
	database  *db.Database                  `json:"-"`
	logger    *logrus.Entry                 `json:"-"`
}

// newTaskExecutor creates a new TaskExecutor instance and recover the previous state from DB.
func newTaskExecutor(txWatcher transaction.Watcher, database *db.Database, logger *logrus.Entry) (*TaskExecutor, error) {
	taskExecutor := &TaskExecutor{
		TxsBackup: make(map[string]*types.Transaction),
		closeChan: make(chan struct{}),
		closeOnce: sync.Once{},
		txWatcher: txWatcher,
		database:  database,
		logger:    logger,
	}

	err := taskExecutor.loadState()
	if err != nil {
		taskExecutor.logger.Warnf("could not find previous State: %v", err)
		if !errors.Is(err, badger.ErrKeyNotFound) {
			return nil, err
		}
	}

	return taskExecutor, nil
}

// close a TaskExecutor. This only can be done once.
func (te *TaskExecutor) close() {
	te.Lock()
	defer te.Unlock()
	te.closeOnce.Do(func() {
		te.logger.Warn("Closing Task Executor")
		close(te.closeChan)
	})
}

// isClosed return true if the task was closed.
func (te *TaskExecutor) isClosed() bool {
	te.RLock()
	defer te.RUnlock()
	select {
	case <-te.closeChan:
		return true
	default:
		return false
	}
}

// triggered when onError occurs during the execution.
func (te *TaskExecutor) onError(err error) error {
	te.logger.WithError(err).Errorf("An unercoverable error occured %v", err)
	te.close()
	return err
}

// addTxBackup adds txn to the backup map for recovery.
func (te *TaskExecutor) addTxBackup(uuid string, tx *types.Transaction) error {
	te.Lock()
	defer te.Unlock()
	te.TxsBackup[uuid] = tx
	return te.persistState()
}

// removeTxBackup removes txn from backup map.
func (te *TaskExecutor) removeTxBackup(uuid string) error {
	te.Lock()
	defer te.Unlock()
	delete(te.TxsBackup, uuid)
	return te.persistState()
}

// getTxBackup retrieves txn from backup map.
func (te *TaskExecutor) getTxBackup(uuid string) (*types.Transaction, bool) {
	te.RLock()
	defer te.RUnlock()
	txn, present := te.TxsBackup[uuid]
	return txn, present
}

// handleTaskExecution It is basically an abstraction to handle the task execution in a separate process.
func (te *TaskExecutor) handleTaskExecution(task tasks.Task, name string, taskId string, start uint64, end uint64, allowMultiExecution bool, subscribeOptions *transaction.SubscribeOptions, database *db.Database, logger *logrus.Entry, eth layer1.Client, contracts layer1.AllSmartContracts, taskResponseChan tasks.InternalTaskResponseChan) {
	err := te.processTask(task, name, taskId, start, end, allowMultiExecution, subscribeOptions, database, logger, eth, contracts, taskResponseChan)
	// Clean up in case the task was killed
	if task.WasKilled() {
		task.GetLogger().Trace("task was externally killed, removing tx backup")
		te.removeTxBackup(task.GetId())
	}
	task.Finish(err)
}

// processTask processes all the stages of the task execution.
func (te *TaskExecutor) processTask(task tasks.Task, name string, taskId string, start uint64, end uint64, allowMultiExecution bool, subscribeOptions *transaction.SubscribeOptions, database *db.Database, logger *logrus.Entry, eth layer1.Client, contracts layer1.AllSmartContracts, taskResponseChan tasks.InternalTaskResponseChan) error {
	taskCtx, cf := context.WithCancel(context.Background())
	defer cf()

	err := task.Initialize(database, logger, eth, contracts, name, taskId, start, end, allowMultiExecution, subscribeOptions, taskResponseChan)
	if err != nil {
		return err
	}
	retryDelay := constants.MonitorRetryDelay
	isComplete := false
	if txn, present := te.getTxBackup(task.GetId()); present {
		isComplete, err = te.checkCompletion(taskCtx, task, txn)
		if err != nil {
			return err
		}
	} else {
		err = te.prepareTask(taskCtx, task, retryDelay)
		if err != nil {
			// unrecoverable errors or ctx.done
			return err
		}
	}

	if !isComplete {
		err = te.executeTask(taskCtx, task, retryDelay)
		if err != nil {
			// unrecoverable errors, staleTx errors or ctx.done
			return err
		}
	}

	task.GetLogger().Trace("Task finished successfully, removing tx backup")
	err = te.removeTxBackup(task.GetId())
	if err != nil {
		return err
	}
	return nil
}

// prepareTask executes task preparation. We keep retrying until the task is
// killed, we get an unrecoverable error, or we succeed.
func (te *TaskExecutor) prepareTask(ctx context.Context, task tasks.Task, retryDelay time.Duration) error {
	for {
		if task.WasKilled() {
			return tasks.ErrTaskKilled
		}
		if te.isClosed() {
			return tasks.ErrTaskExecutionMechanismClosed
		}

		taskErr := task.Prepare(ctx)
		// no errors or unrecoverable errors
		if taskErr == nil {
			return nil
		}
		if !taskErr.IsRecoverable() {
			return taskErr
		}
		err := te.sleepOrExit(task, retryDelay)
		if err != nil {
			return err
		}
	}
}

// executeTask executes task business logic. We keep retrying until the task is
// killed, we get an unrecoverable error, or we succeed.
func (te *TaskExecutor) executeTask(ctx context.Context, task tasks.Task, retryDelay time.Duration) error {
	logger := task.GetLogger()
	for {
		hasToExecute, err := te.shouldExecute(ctx, task)
		if err != nil {
			return err
		}
		if !hasToExecute {
			return nil
		}
		txn, taskErr := task.Execute(ctx)
		if taskErr != nil {
			if taskErr.IsRecoverable() {
				logger.Tracef("got a recoverable error during task.execute: %v", taskErr.Error())
				err := te.sleepOrExit(task, retryDelay)
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
			err := te.addTxBackup(task.GetId(), txn)
			if err != nil {
				return err
			}

			isComplete, err := te.checkCompletion(ctx, task, txn)
			if err != nil {
				return err
			}
			if isComplete {
				return nil
			}
			continue
		}
		logger.Debug("Task returned no transaction, finishing")
		return nil
	}
}

// checkCompletion checks if a task is complete. The function is going to subscribe a
// transaction in the txWatcher, and it will wait until it gets the receipt,
// the task is killed, or shouldExecute returns false.
func (te *TaskExecutor) checkCompletion(ctx context.Context, task tasks.Task, txn *types.Transaction) (bool, error) {
	var err error
	var receipt *types.Receipt
	logger := task.GetLogger()

	receiptResponse, err := te.txWatcher.Subscribe(ctx, txn, task.GetSubscribeOptions())
	if err != nil {
		// if we get an error here, it means that we have a corrupted txn we should
		// retry a transaction
		logger.Errorf("failed to subscribe tx with error: %s", err.Error())
		return false, nil
	}

	for {
		select {
		case <-task.KillChan():
			return true, tasks.ErrTaskKilled
		case <-te.closeChan:
			return true, tasks.ErrTaskExecutionMechanismClosed
		case <-time.After(tasks.ExecutorPoolingTime):
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
		hasToExecute, err := te.shouldExecute(ctx, task)
		if err != nil {
			return false, err
		}
		if !hasToExecute {
			return true, nil
		}
	}
}

// shouldExecute checks if a task should be executed. In case of recoverable errors the
// function is going to retry `constants.MonitorRetryCount` times. If it
// exhausted the number of retries, it sends true since there's no information.
// The function returns false in case of unrecoverable errors, or if it
// shouldn't execute a task. In case of no errors, the return value is true
// (default case to t.ShouldExecute to return that a task should be executed).
func (te *TaskExecutor) shouldExecute(ctx context.Context, task tasks.Task) (bool, error) {
	logger := task.GetLogger()
	for i := uint64(1); i <= constants.MonitorRetryCount; i++ {
		if task.WasKilled() {
			return false, tasks.ErrTaskKilled
		}
		if te.isClosed() {
			return false, tasks.ErrTaskExecutionMechanismClosed
		}

		if hasToExecute, err := task.ShouldExecute(ctx); err != nil {
			if err.IsRecoverable() {
				logger.Tracef("got a recoverable error during task.ShouldExecute: %v", err)
				if err := te.sleepOrExit(task, constants.MonitorRetryDelay); err != nil {
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

// persistState TaskExecutor to database.
func (te *TaskExecutor) persistState() error {
	logger := logging.GetLogger("staterecover").WithField("State", "task_executor")
	rawData, err := json.Marshal(te)
	if err != nil {
		return te.onError(err)
	}

	err = te.database.Update(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskExecutorState()
		logger.WithField("Key", string(key)).Debug("Saving state in the database")
		if err := utils.SetValue(txn, key, rawData); err != nil {
			logger.Error("Failed to set Value")
			return err
		}
		return nil
	})
	if err != nil {
		return te.onError(err)
	}

	if err = te.database.Sync(); err != nil {
		logger.Error("Failed to set sync")
		return te.onError(err)
	}

	return nil
}

// loadState TaskExecutor from database.
func (te *TaskExecutor) loadState() error {
	logger := logging.GetLogger("staterecover").WithField("State", "task_executor")
	if err := te.database.View(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskExecutorState()
		logger.WithField("Key", string(key)).Debug("Loading state from database")
		rawData, err := utils.GetValue(txn, key)
		if err != nil {
			return err
		}

		err = json.Unmarshal(rawData, te)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return te.onError(err)
	}

	// synchronizing db state to disk
	if err := te.database.Sync(); err != nil {
		logger.Error("Failed to set sync")
		return te.onError(err)
	}

	return nil

}

// sleepOrClose sleeps a certain amount of time.
// It fails in case the task is closed.
func (te *TaskExecutor) sleepOrExit(task tasks.Task, delay time.Duration) error {
	select {
	case <-task.KillChan():
		return tasks.ErrTaskKilled
	case <-te.closeChan:
		return tasks.ErrTaskExecutionMechanismClosed
	case <-time.After(delay):
		return nil
	}
}
