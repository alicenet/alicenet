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
	mutex sync.RWMutex
	// Unique Id of the task
	ID string `json:"id"`
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
	killChan         chan struct{}            `json:"-"`
	killOnce         sync.Once                `json:"-"`
	database         *db.Database             `json:"-"`
	logger           *logrus.Entry            `json:"-"`
	client           layer1.Client            `json:"-"`
	contracts        layer1.AllSmartContracts `json:"-"`
	taskResponseChan InternalTaskResponseChan `json:"-"`
}

// NewBaseTask creates a new Base task. BaseTask should be the base of any task.
// This function is called outside the scheduler to create the object to be
// scheduled.
func NewBaseTask(
	start uint64,
	end uint64,
	allowMultiExecution bool,
	subscribeOptions *transaction.SubscribeOptions,
) *BaseTask {
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
func (bt *BaseTask) Initialize(
	database *db.Database,
	logger *logrus.Entry,
	eth layer1.Client,
	contracts layer1.AllSmartContracts,
	name string,
	id string,
	start uint64,
	end uint64,
	allowMultiExecution bool,
	subscribeOptions *transaction.SubscribeOptions,
	taskResponseChan InternalTaskResponseChan,
) error {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()
	if bt.isInitialized {
		return errors.New("trying to initialize task twice")
	}

	bt.ID = id
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
	bt.killChan = make(chan struct{})
	bt.killOnce = sync.Once{}
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
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()
	return bt.ID
}

// GetStart gets the start date of a task. Returns 0 if a task does not have a
// start date (started immediately).
func (bt *BaseTask) GetStart() uint64 {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()
	return bt.Start
}

// GetEnd gets the end date in blocks of a task. In case 0, the task does not
// have an end block.
func (bt *BaseTask) GetEnd() uint64 {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()
	return bt.End
}

// GetName get the name of the task.
func (bt *BaseTask) GetName() string {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()
	return bt.Name
}

// GetAllowMultiExecution returns if a task type allows multiple execution.
func (bt *BaseTask) GetAllowMultiExecution() bool {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()
	return bt.AllowMultiExecution
}

// GetSubscribeOptions gets the transactionWatcher subscribeOptions specific for
// a task.
func (bt *BaseTask) GetSubscribeOptions() *transaction.SubscribeOptions {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()

	if bt.SubscribeOptions == nil {
		return nil
	}
	subscribeOptionsClone := *bt.SubscribeOptions
	return &subscribeOptionsClone
}

// GetClient returns the layer1 client implemented by the task.
func (bt *BaseTask) GetClient() layer1.Client {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()
	return bt.client
}

// GetContractsHandler returns the handler that has access to all different
// layer1 smart contracts.
func (bt *BaseTask) GetContractsHandler() layer1.AllSmartContracts {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()
	return bt.contracts
}

// GetLogger returns the task logger.
func (bt *BaseTask) GetLogger() *logrus.Entry {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()
	return bt.logger
}

// GetDB returns the database where the task can save and load its state.
func (bt *BaseTask) GetDB() *db.Database {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()
	return bt.database
}

// Kill a running task. This only can be done once.
func (bt *BaseTask) Kill() {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()
	bt.killOnce.Do(func() {
		bt.logger.Warnf("Closing task %s-%s", bt.Name, bt.ID)
		close(bt.killChan)
	})
}

// KillChan returns a channel that is closed when the Task was
// killed.
func (bt *BaseTask) KillChan() <-chan struct{} {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()
	return bt.killChan
}

// WasKilled return true if the task was killed.
func (bt *BaseTask) WasKilled() bool {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()
	select {
	case <-bt.killChan:
		return true
	default:
		return false
	}
}

// Finish executes the cleanup logic once a task finishes.
func (bt *BaseTask) Finish(err error) {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, ErrTaskKilled) ||
			errors.Is(err, ErrTaskExecutionMechanismClosed) {
			bt.logger.WithError(err).Debug("finishing task execution, it was aborted")
		} else {
			bt.logger.WithError(err).Error("got an error when executing task")
		}
	} else {
		bt.logger.Info("task is done")
	}
	if bt.taskResponseChan != nil {
		bt.taskResponseChan.Add(bt.ID, err)
	}
}

func (bt *BaseTask) Lock() {
	bt.mutex.Lock()
}

func (bt *BaseTask) Unlock() {
	bt.mutex.Unlock()
}
