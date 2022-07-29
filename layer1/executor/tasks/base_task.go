package tasks

import (
	"context"
	"errors"
	"sync"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/sirupsen/logrus"
)

type BaseTask struct {
	sync.RWMutex
	// Unique Id of the task
	Id string `json:"id"`
	// Task name/type
	Name string `json:"name"`
	// If this task can be executed in parallel with other tasks of the same type/name
	AllowMultiExecution bool `json:"allowMultiExecution"`
	// Subscription options (if the task should be retried, finality delay, etc)
	SubscribeOptions *transaction.SubscribeOptions `json:"subscribeOptions"`
	// Which block the task should be started. In case the start is 0 the task is
	// started immediately.
	Start uint64 `json:"start"`
	// Which block the task should be ended. In case the end is 0 the task runs
	// forever (until the task succeeds, or it's killed, be careful when using this).
	// Otherwise, the task will end at the specified block.
	End              uint64                   `json:"end"`
	isInitialized    bool                     `json:"-"`
	wasKilled        bool                     `json:"-"`
	ctx              context.Context          `json:"-"`
	cancelFunc       context.CancelFunc       `json:"-"`
	database         *db.Database             `json:"-"`
	logger           *logrus.Entry            `json:"-"`
	client           layer1.Client            `json:"-"`
	contracts        layer1.AllSmartContracts `json:"-"`
	taskResponseChan InternalTaskResponseChan `json:"-"`
}

// NewBaseTask creates a new Base task. BaseTask should be the base of any task.
// This function is called outside the scheduler to create the object to be
// scheduled.
func NewBaseTask(start uint64, end uint64, allowMultiExecution bool, subscribeOptions *transaction.SubscribeOptions) *BaseTask {
	return &BaseTask{
		Start:               start,
		End:                 end,
		AllowMultiExecution: allowMultiExecution,
		SubscribeOptions:    subscribeOptions,
	}
}

// Initialize initializes the task after its creation. It should be only called
// by the task scheduler during task spawn as separated go routine. This
// function all the parameters for task execution and control by the scheduler.
func (bt *BaseTask) Initialize(ctx context.Context, cancelFunc context.CancelFunc, database *db.Database, logger *logrus.Entry, eth layer1.Client, contracts layer1.AllSmartContracts, name string, id string, start uint64, end uint64, allowMultiExecution bool, subscribeOptions *transaction.SubscribeOptions, taskResponseChan InternalTaskResponseChan) error {
	bt.Lock()
	defer bt.Unlock()
	if bt.isInitialized {
		return errors.New("trying to initialize task twice")
	}

	bt.Id = id
	bt.Name = name
	bt.Start = start
	bt.End = end
	bt.AllowMultiExecution = allowMultiExecution

	if subscribeOptions == nil {
		bt.SubscribeOptions = subscribeOptions
	} else {
		subscribeOptionsClone := *subscribeOptions
		bt.SubscribeOptions = &subscribeOptionsClone
	}
	bt.ctx = ctx
	bt.cancelFunc = cancelFunc
	bt.database = database
	bt.logger = logger
	bt.client = eth
	bt.contracts = contracts
	bt.taskResponseChan = taskResponseChan
	bt.isInitialized = true

	return nil
}

// GetId gets the task unique ID.
func (bt *BaseTask) GetId() string {
	bt.RLock()
	defer bt.RUnlock()
	return bt.Id
}

// GetStart gets the start date of a task. Returns 0 if a task does not have a
// start date (started immediately).
func (bt *BaseTask) GetStart() uint64 {
	bt.RLock()
	defer bt.RUnlock()
	return bt.Start
}

// GetEnd gets the end date in blocks of a task. In case 0, the task does not
// have an end block.
func (bt *BaseTask) GetEnd() uint64 {
	bt.RLock()
	defer bt.RUnlock()
	return bt.End
}

// GetName get the name of the task.
func (bt *BaseTask) GetName() string {
	bt.RLock()
	defer bt.RUnlock()
	return bt.Name
}

// GetAllowMultiExecution returns if a task type allows multiple execution.
func (bt *BaseTask) GetAllowMultiExecution() bool {
	bt.RLock()
	defer bt.RUnlock()
	return bt.AllowMultiExecution
}

// GetSubscribeOptions gets the transactionWatcher subscribeOptions specific for
// a task.
func (bt *BaseTask) GetSubscribeOptions() *transaction.SubscribeOptions {
	bt.RLock()
	defer bt.RUnlock()

	if bt.SubscribeOptions == nil {
		return nil
	}
	subscribeOptionsClone := *bt.SubscribeOptions
	return &subscribeOptionsClone
}

// GetCtx get the context to be used by a task.
func (bt *BaseTask) GetCtx() context.Context {
	bt.RLock()
	defer bt.RUnlock()
	return bt.ctx
}

// WasKilled returns true if the task was killed otherwise false.
func (bt *BaseTask) WasKilled() bool {
	bt.RLock()
	defer bt.RUnlock()
	return bt.wasKilled
}

// GetClient returns the layer1 client implemented by the task.
func (bt *BaseTask) GetClient() layer1.Client {
	bt.RLock()
	defer bt.RUnlock()
	return bt.client
}

// GetContractsHandler returns the handler that has access to all different
// layer1 smart contracts.
func (bt *BaseTask) GetContractsHandler() layer1.AllSmartContracts {
	bt.RLock()
	defer bt.RUnlock()
	return bt.contracts
}

// GetLogger returns the task logger.
func (bt *BaseTask) GetLogger() *logrus.Entry {
	bt.RLock()
	defer bt.RUnlock()
	return bt.logger
}

// GetDB returns the database where the task can save and load its state.
func (bt *BaseTask) GetDB() *db.Database {
	bt.RLock()
	defer bt.RUnlock()
	return bt.database
}

// Close closes a running task. It set a bool flag and call the cancelFunc in
// case it's different from nil.
func (bt *BaseTask) Close() {
	bt.Lock()
	defer bt.Unlock()
	if bt.cancelFunc != nil {
		bt.cancelFunc()
	}
	bt.wasKilled = true
}

// Finish executes the cleanup logic once a task finishes.
func (bt *BaseTask) Finish(err error) {
	bt.Lock()
	defer bt.Unlock()
	if err != nil {
		if bt.wasKilled {
			bt.logger.WithError(err).Debug("cancelling task execution, task was killed")
		} else {
			bt.logger.WithError(err).Error("got an error when executing task")
		}
	} else {
		bt.logger.Info("task is done")
	}
	if bt.taskResponseChan != nil {
		bt.taskResponseChan.Add(bt.Id, err)
	}
}
