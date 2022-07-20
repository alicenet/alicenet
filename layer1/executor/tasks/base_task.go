package tasks

import (
	"context"
	"errors"
	"github.com/alicenet/alicenet/layer1/executor"
	"sync"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/sirupsen/logrus"
)

type BaseTask struct {
	sync.RWMutex
	Name                string                        `json:"name"`
	AllowMultiExecution bool                          `json:"allowMultiExecution"`
	SubscribeOptions    *transaction.SubscribeOptions `json:"subscribeOptions,omitempty"`
	Id                  string                        `json:"id"`
	Start               uint64                        `json:"start"`
	End                 uint64                        `json:"end"`
	isInitialized       bool                          `json:"-"`
	wasKilled           bool                          `json:"-"`
	ctx                 context.Context               `json:"-"`
	cancelFunc          context.CancelFunc            `json:"-"`
	database            *db.Database                  `json:"-"`
	logger              *logrus.Entry                 `json:"-"`
	client              layer1.Client                 `json:"-"`
	taskResponseChan    InternalTaskResponseChan      `json:"-"`
}

func NewBaseTask(start uint64, end uint64, allowMultiExecution bool, subscribeOptions *transaction.SubscribeOptions) *BaseTask {
	ctx, cf := context.WithCancel(context.Background())

	return &BaseTask{
		Start:               start,
		End:                 end,
		AllowMultiExecution: allowMultiExecution,
		SubscribeOptions:    subscribeOptions,
		ctx:                 ctx,
		cancelFunc:          cf,
	}
}

// Initialize default implementation for the ITask interface
func (bt *BaseTask) Initialize(ctx context.Context, cancelFunc context.CancelFunc, database *db.Database, logger *logrus.Entry, eth layer1.Client, name string, id string, taskResponseChan InternalTaskResponseChan) error {
	bt.Lock()
	defer bt.Unlock()
	if bt.isInitialized {
		return errors.New("trying to initialize task twice!")
	}

	bt.Name = name
	bt.Id = id
	bt.ctx = ctx
	bt.cancelFunc = cancelFunc
	bt.database = database
	bt.logger = logger
	bt.client = eth
	bt.taskResponseChan = taskResponseChan
	bt.isInitialized = true

	return nil
}

// GetId default implementation for the ITask interface
func (bt *BaseTask) GetId() string {
	bt.RLock()
	defer bt.RUnlock()
	return bt.Id
}

// GetStart default implementation for the ITask interface
func (bt *BaseTask) GetStart() uint64 {
	bt.RLock()
	defer bt.RUnlock()
	return bt.Start
}

// GetEnd default implementation for the ITask interface
func (bt *BaseTask) GetEnd() uint64 {
	bt.RLock()
	defer bt.RUnlock()
	return bt.End
}

// GetName default implementation for the ITask interface
func (bt *BaseTask) GetName() string {
	bt.RLock()
	defer bt.RUnlock()
	return bt.Name
}

// GetAllowMultiExecution default implementation for the ITask interface
func (bt *BaseTask) GetAllowMultiExecution() bool {
	bt.RLock()
	defer bt.RUnlock()
	return bt.AllowMultiExecution
}

func (bt *BaseTask) GetSubscribeOptions() *transaction.SubscribeOptions {
	bt.RLock()
	defer bt.RUnlock()
	return bt.SubscribeOptions
}

// GetCtx default implementation for the ITask interface
func (bt *BaseTask) GetCtx() context.Context {
	bt.RLock()
	defer bt.RUnlock()
	return bt.ctx
}

// GetCtx default implementation for the ITask interface
func (bt *BaseTask) WasKilled() bool {
	bt.RLock()
	defer bt.RUnlock()
	return bt.wasKilled
}

// GetEth default implementation for the ITask interface
func (bt *BaseTask) GetClient() layer1.Client {
	bt.RLock()
	defer bt.RUnlock()
	return bt.client
}

// GetLogger default implementation for the ITask interface
func (bt *BaseTask) GetLogger() *logrus.Entry {
	bt.RLock()
	defer bt.RUnlock()
	return bt.logger
}

func (bt *BaseTask) GetDB() *db.Database {
	bt.RLock()
	defer bt.RUnlock()
	return bt.database
}

// Close default implementation for the ITask interface
func (bt *BaseTask) Close() {
	bt.Lock()
	defer bt.Unlock()
	if bt.cancelFunc != nil {
		bt.cancelFunc()
	}
	bt.wasKilled = true
}

// Finish default implementation for the ITask interface
func (bt *BaseTask) Finish(err error) {
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			bt.logger.WithError(err).Error("got an error when executing task")
		} else {
			bt.logger.WithError(err).Debug("cancelling task execution")
		}
	} else {
		bt.logger.Info("task is done")
	}
	if bt.taskResponseChan != nil {
		bt.taskResponseChan.Add(executor.ExecutorResponse{Id: bt.Id, Err: err})
	}
}
