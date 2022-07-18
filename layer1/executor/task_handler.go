package executor

import (
	"context"
	"github.com/sirupsen/logrus"
	"sync"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	monitorInterfaces "github.com/alicenet/alicenet/layer1/monitor/interfaces"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/google/uuid"
)

// TaskAction is an enumeration indicating the actions that the scheduler
// can do with a task during a request:
type TaskAction int

// The possible actions that the scheduler can do with a task during a request:
// * KillByType          - To kill/prune a task by type immediately
// * KillById            - To kill/prune a task by id immediately
// * Schedule            - To schedule a new task
const (
	KillByType TaskAction = iota
	KillById
	Schedule
)

func (action TaskAction) String() string {
	return [...]string{
		"KillByType",
		"KillById",
		"Schedule",
	}[action]
}

type TaskManagerResponse struct {
	SharedResponse *TaskSharedResponse
	Err            error
}

// TaskManagerResponseChannel a response channel is basically a non-blocking channel that can
// only be written and closed once.
type TaskManagerResponseChannel struct {
	writeOnce sync.Once
	channel   chan *TaskManagerResponse // internal channel
}

// NewResponseChannel creates a new response channel.
func NewResponseChannel() *TaskManagerResponseChannel {
	return &TaskManagerResponseChannel{channel: make(chan *TaskManagerResponse, 1)}
}

// sendResponse sends a unique response and close the internal channel. Additional calls to
// this function will be no-op
func (rc *TaskManagerResponseChannel) sendResponse(response *TaskManagerResponse) {
	rc.writeOnce.Do(func() {
		rc.channel <- response
		close(rc.channel)
	})
}

func (rc *TaskManagerResponseChannel) listen(ctx context.Context) (*TaskSharedResponse, error) {
	// wait for request to be processed
	select {
	case response := <-rc.channel:
		return response.SharedResponse, response.Err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type InternalTaskResponse struct {
	Id  string
	Err error
}

type internalRequest struct {
	task     tasks.Task
	id       string
	action   TaskAction
	response *TaskManagerResponseChannel
}

var _ tasks.TaskHandler = &Handler{}

type Handler struct {
	manager          *TaskManager
	logger           *logrus.Entry
	closeMainContext context.CancelFunc
	requestChannel   chan internalRequest
}

func NewTaskHandler(database *db.Database, eth layer1.Client, adminHandler monitorInterfaces.AdminHandler, txWatcher *transaction.FrontWatcher) (tasks.TaskHandler, error) {
	// main context that will cancel all workers and go routines
	mainCtx, cf := context.WithCancel(context.Background())

	// Setup tasks scheduler
	requestChan := make(chan internalRequest, constants.TaskSchedulerBufferSize)

	taskManager, err := newTaskManager(mainCtx, eth, database, adminHandler, requestChan, txWatcher)
	if err != nil {
		cf()
		return nil, err
	}

	handler := &Handler{
		manager:          taskManager,
		closeMainContext: cf,
		requestChannel:   requestChan,
	}

	return handler, nil
}

func (i *Handler) Start() {
	i.logger.Info("Starting task handler")
	i.manager.start()
}

func (i *Handler) Close() {
	i.logger.Warn("Closing task handler")
	close(i.requestChannel)
	i.closeMainContext()
}

// ScheduleTask sends the task to the backend
func (i *Handler) ScheduleTask(ctx context.Context, task tasks.Task, id string) (*TaskSharedResponse, error) {
	// In case the id field is not specified, create it
	if id == "" {
		id = uuid.New().String()
	}
	req := internalRequest{task: task, id: id, action: Schedule, response: NewResponseChannel()}
	err := i.waitForRequestProcessing(ctx, req)
	if err != nil {
		return nil, err
	}
	return req.response.listen(ctx)
}

func (i *Handler) KillTaskByType(ctx context.Context, taskType tasks.Task) (*TaskSharedResponse, error) {
	req := internalRequest{task: taskType, action: KillByType, response: NewResponseChannel()}
	err := i.waitForRequestProcessing(ctx, req)
	if err != nil {
		return nil, err
	}
	return req.response.listen(ctx)
}

func (i *Handler) KillTaskById(ctx context.Context, id string) (*TaskSharedResponse, error) {
	req := internalRequest{id: id, action: KillById, response: NewResponseChannel()}
	err := i.waitForRequestProcessing(ctx, req)
	if err != nil {
		return nil, err
	}
	return req.response.listen(ctx)
}

func (i *Handler) waitForRequestProcessing(ctx context.Context, req internalRequest) error {
	// wait for request to be accepted
	select {
	case i.requestChannel <- req:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
