package executor

import (
	"context"
	"sync"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/executor/marshaller"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/snapshots"
	monitorInterfaces "github.com/alicenet/alicenet/layer1/monitor/interfaces"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/logging"
	"github.com/google/uuid"
)

type TaskSender interface {
	ScheduleTask(ctx context.Context, task tasks.Task) error
	KillTaskByType(ctx context.Context, task tasks.Task) error
	KillTaskById(ctx context.Context, id string) error
	Start()
	Close()
}

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

type SharedTaskResponse struct {
	doneChan chan struct{}
	err      error // error in case the task failed
}

func newSharedResponse() *SharedTaskResponse {
	return &SharedTaskResponse{doneChan: make(chan struct{})}
}

// Function to check if a receipt is ready
func (r *SharedTaskResponse) IsReady() bool {
	select {
	case <-r.doneChan:
		return true
	default:
		return false
	}
}

// blocking function to get the execution status of a task. This function will
// block until the task is finished and the final result is returned.
func (r *SharedTaskResponse) GetTaskResponseBlocking(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.doneChan:
		return r.err
	}
}

// function to write the receipt or error from a transaction being watched.
func (r *SharedTaskResponse) writeResponse(err error) {
	if !r.IsReady() {
		r.err = err
		close(r.doneChan)
	}
}

type SchedulerResponse struct {
	SharedResponse *SharedTaskResponse
	Err            error
}

// A response channel is basically a non-blocking channel that can only be
// written and closed once.
type SchedulerResponseChannel struct {
	writeOnce sync.Once
	channel   chan *SchedulerResponse // internal channel
}

// Create a new response channel.
func NewResponseChannel() *SchedulerResponseChannel {
	return &SchedulerResponseChannel{channel: make(chan *SchedulerResponse, 1)}
}

// send a unique response and close the internal channel. Additional calls to
// this function will be no-op
func (rc *SchedulerResponseChannel) sendResponse(response *SchedulerResponse) {
	rc.writeOnce.Do(func() {
		rc.channel <- response
		close(rc.channel)
	})
}

func (rc *SchedulerResponseChannel) listen(ctx context.Context) (*SharedTaskResponse, error) {
	// wait for request to be processed
	select {
	case response := <-rc.channel:
		return response.SharedResponse, response.Err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type internalTaskResponse struct {
	Id  string
	Err error
}

type internalRequest struct {
	task     tasks.Task
	id       string
	action   TaskAction
	response *SchedulerResponseChannel
}

var _ TaskSender = &Sender{}

type Sender struct {
	taskScheduler  *TasksSchedulerBackend
	requestChannel chan internalRequest
}

func NewTasksScheduler(database *db.Database, eth layer1.Client, adminHandler monitorInterfaces.AdminHandler, txWatcher *transaction.FrontWatcher) (TaskSender, error) {

	// main context that will cancel all workers and go routine
	mainCtx, cf := context.WithCancel(context.Background())

	// Setup tasks scheduler
	taskRequestChan := make(chan internalRequest, constants.TaskSchedulerBufferSize)

	s := &TasksSchedulerBackend{
		Schedule:         make(map[string]TaskRequestInfo),
		mainCtx:          mainCtx,
		mainCtxCf:        cf,
		database:         database,
		eth:              eth,
		adminHandler:     adminHandler,
		marshaller:       GetTaskRegistry(),
		cancelChan:       make(chan bool, 1),
		taskRequestChan:  taskRequestChan,
		taskResponseChan: &taskResponseChan{trChan: make(chan internalTaskResponse, 100)},
		txWatcher:        txWatcher,
	}

	logger := logging.GetLogger("tasks")
	s.logger = logger.WithField("Component", "schedule")

	tasksManager, err := NewTaskManager(txWatcher, database, logger.WithField("Component", "manager"))
	if err != nil {
		return nil, err
	}
	s.tasksManager = tasksManager

	sender := &Sender{
		taskScheduler:  s,
		requestChannel: taskRequestChan,
	}

	return sender, nil
}

func (r *Sender) Start() {
	r.taskScheduler.Start()
}

func (r *Sender) Close() {
	r.taskScheduler.Close()
	close(r.requestChannel)
}

func (r *Sender) ScheduleTask(ctx context.Context, task tasks.Task, id string) (*SharedTaskResponse, error) {
	newId := uuid.New().String()
	// In case the id field is specified, use it instead
	if id != "" {
		newId = id
	}
	req := internalRequest{task: task, id: newId, action: Schedule, response: NewResponseChannel()}
	err := r.waitForRequestProcessing(ctx, req)
	if err != nil {
		return nil, err
	}
	return req.response.listen(ctx)
}

func (r *Sender) KillTaskByType(ctx context.Context, taskType tasks.Task) error {
	req := internalRequest{task: taskType, action: Kill, response: NewResponseChannel()}
	return r.waitForRequestProcessing(ctx, req)
}

func (r *Sender) KillTaskById(ctx context.Context, id string) error {
	req := internalRequest{id: id, action: Kill, response: NewResponseChannel()}
	return r.waitForRequestProcessing(ctx, req)
}

func (r *Sender) waitForRequestProcessing(ctx context.Context, req internalRequest) error {
	// wait for request to be accepted
	select {
	case r.requestChannel <- req:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func GetTaskRegistry() *marshaller.TypeRegistry {
	// registry the type here
	tr := &marshaller.TypeRegistry{}
	tr.RegisterInstanceType(&dkg.CompletionTask{})
	tr.RegisterInstanceType(&dkg.DisputeShareDistributionTask{})
	tr.RegisterInstanceType(&dkg.DisputeMissingShareDistributionTask{})
	tr.RegisterInstanceType(&dkg.DisputeMissingKeySharesTask{})
	tr.RegisterInstanceType(&dkg.DisputeMissingGPKjTask{})
	tr.RegisterInstanceType(&dkg.DisputeGPKjTask{})
	tr.RegisterInstanceType(&dkg.GPKjSubmissionTask{})
	tr.RegisterInstanceType(&dkg.KeyShareSubmissionTask{})
	tr.RegisterInstanceType(&dkg.MPKSubmissionTask{})
	tr.RegisterInstanceType(&dkg.RegisterTask{})
	tr.RegisterInstanceType(&dkg.DisputeMissingRegistrationTask{})
	tr.RegisterInstanceType(&dkg.ShareDistributionTask{})
	tr.RegisterInstanceType(&snapshots.SnapshotTask{})
	return tr
}
