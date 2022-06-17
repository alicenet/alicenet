package executor

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/MadBase/MadNet/constants/dbprefix"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"time"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/constants"
	"github.com/sirupsen/logrus"
)

type TasksManager struct {
	Transactions map[string]*types.Transaction `json:"transactions"`
	txWatcher    *transaction.Watcher
	database     *db.Database  `json:"-"`
	logger       *logrus.Entry `json:"-"`
}

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

func (t *TasksManager) ManageTask(mainCtx context.Context, task interfaces.ITask, database *db.Database, logger *logrus.Entry, eth ethereum.Network, taskResponseChan interfaces.ITaskResponseChan) {
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

	err = prepareTask(taskCtx, task, retryDelay)
	if err != nil {
		// unrecoverable errors, recoverable errors but we exhausted all attempts or
		// ctx.done
		return
	}

	err = t.executeTask(taskCtx, task, retryDelay)
}

// prepareTask executes task preparation
func prepareTask(ctx context.Context, task interfaces.ITask, retryDelay time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		taskErr := task.Prepare(ctx)
		// no errors or unrecoverable errors break
		if taskErr == nil || !taskErr.IsRecoverable() {
			return taskErr
		}
		err := sleepWithContext(ctx, retryDelay)
		if err != nil {
			return err
		}
	}
}

// executeTask executes task business logic
func (t *TasksManager) executeTask(ctx context.Context, task interfaces.ITask, retryDelay time.Duration) error {
	if shouldExecute(ctx, task) {
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
				t.Transactions[task.GetId()] = txn
				err := t.persistState()
				if err != nil {
					return err
				}

				isComplete, err := t.checkCompletion(ctx, task, txn)
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

func (t *TasksManager) checkCompletion(ctx context.Context, task interfaces.ITask, txn *types.Transaction) (bool, error) {
	var err error
	var receipt *types.Receipt
	isComplete := false
	receiptResponse, err := t.txWatcher.Subscribe(ctx, txn, task.GetSubscribeOptions())
	if err != nil {
		// if we get an error here, it means that we have a corrupted txn we should
		// retry a transaction
		t.logger.Errorf("failed to subscribe tx with error: %s", err.Error())
		return isComplete, nil
	}

	for {
		select {
		case <-ctx.Done():
			isComplete = true
			return isComplete, ctx.Err()
		case <-time.After(1 * time.Second):
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

		if !shouldExecute(task.GetCtx(), task) {
			isComplete = true
			return isComplete, nil
		}
	}

	return isComplete, nil
}

func shouldExecute(ctx context.Context, task interfaces.ITask) bool {
	for i := uint64(1); i <= constants.MonitorRetryCount; i++ {
		if err := task.ShouldExecute(ctx); err != nil {
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

func (t *TasksManager) persistState() error {
	rawData, err := json.Marshal(t)
	if err != nil {
		return err
	}

	err = t.database.Update(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskManagerState()
		t.logger.WithField("Key", string(key)).Infof("Saving state")
		if err := utils.SetValue(txn, key, rawData); err != nil {
			t.logger.Error("Failed to set Value")
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := t.database.Sync(); err != nil {
		t.logger.Error("Failed to set sync")
		return err
	}

	return nil
}

func (t *TasksManager) loadState() error {

	if err := t.database.View(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskManagerState()
		t.logger.WithField("Key", string(key)).Infof("Looking up state")
		rawData, err := utils.GetValue(txn, key)
		if err != nil {
			return err
		}

		err = json.Unmarshal(rawData, t)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// synchronizing db state to disk
	if err := t.database.Sync(); err != nil {
		t.logger.Error("Failed to set sync")
		return err
	}

	return nil

}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}
