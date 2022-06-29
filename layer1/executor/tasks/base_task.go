package tasks

import (
	"context"
	"errors"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/sirupsen/logrus"
)

// TaskAction is an enumeration indicating the actions that the scheduler
// can do with a task during a request:
type TaskAction int

// The possible actions that the scheduler can do with a task during a request:
// * Kill          - To kill/prune a task type immediately
// * Schedule      - To schedule a new task
const (
	Kill TaskAction = iota
	Schedule
)

func (action TaskAction) String() string {
	return [...]string{
		"Kill",
		"Schedule",
	}[action]
}

type TaskResponse struct {
	Id  string
	Err error
}

type TaskRequest struct {
	Action TaskAction
	Task   Task
}

func NewScheduleTaskRequest(task Task) TaskRequest {
	return TaskRequest{Action: Schedule, Task: task}
}

func NewKillTaskRequest(task Task) TaskRequest {
	return TaskRequest{Action: Kill, Task: task}
}

type BaseTask struct {
	Name                string                        `json:"name"`
	AllowMultiExecution bool                          `json:"allowMultiExecution"`
	SubscribeOptions    *transaction.SubscribeOptions `json:"subscribeOptions,omitempty"`
	Id                  string                        `json:"id"`
	Start               uint64                        `json:"start"`
	End                 uint64                        `json:"end"`
	isInitialized       bool                          `json:"-"`
	ctx                 context.Context               `json:"-"`
	cancelFunc          context.CancelFunc            `json:"-"`
	database            *db.Database                  `json:"-"`
	logger              *logrus.Entry                 `json:"-"`
	client              layer1.Client                 `json:"-"`
	taskResponseChan    TaskResponseChan              `json:"-"`
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
func (bt *BaseTask) Initialize(ctx context.Context, cancelFunc context.CancelFunc, database *db.Database, logger *logrus.Entry, eth layer1.Client, name string, id string, taskResponseChan TaskResponseChan) error {
	if !bt.isInitialized {
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
	return errors.New("trying to initialize task twice!")
}

// GetId default implementation for the ITask interface
func (bt *BaseTask) GetId() string {
	return bt.Id
}

// GetStart default implementation for the ITask interface
func (bt *BaseTask) GetStart() uint64 {
	return bt.Start
}

// GetEnd default implementation for the ITask interface
func (bt *BaseTask) GetEnd() uint64 {
	return bt.End
}

// GetName default implementation for the ITask interface
func (bt *BaseTask) GetName() string {
	return bt.Name
}

// GetAllowMultiExecution default implementation for the ITask interface
func (bt *BaseTask) GetAllowMultiExecution() bool {
	return bt.AllowMultiExecution
}

func (bt *BaseTask) GetSubscribeOptions() *transaction.SubscribeOptions {
	return bt.SubscribeOptions
}

// GetCtx default implementation for the ITask interface
func (bt *BaseTask) GetCtx() context.Context {
	return bt.ctx
}

// GetEth default implementation for the ITask interface
func (bt *BaseTask) GetClient() layer1.Client {
	return bt.client
}

// GetLogger default implementation for the ITask interface
func (bt *BaseTask) GetLogger() *logrus.Entry {
	return bt.logger
}

// Close default implementation for the ITask interface
func (bt *BaseTask) Close() {
	if bt.cancelFunc != nil {
		bt.cancelFunc()
	}
}

// Finish default implementation for the ITask interface
func (bt *BaseTask) Finish(err error) {
	if err != nil {
		if err != context.Canceled {
			bt.logger.WithError(err).Error("got an error when executing task")
		} else {
			bt.logger.WithError(err).Debug("cancelling task execution")
		}
	} else {
		bt.logger.Info("task is done")
	}
	if bt.taskResponseChan != nil {
		bt.taskResponseChan.Add(TaskResponse{Id: bt.Id, Err: err})
	}
}

func (bt *BaseTask) GetDB() *db.Database {
	return bt.database
}
