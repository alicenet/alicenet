package executor

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/snapshots"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/executor/marshaller"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	monitorInterfaces "github.com/alicenet/alicenet/layer1/monitor/interfaces"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

var (
	ErrNotScheduled                   = errors.New("scheduled task not found")
	ErrWrongParams                    = errors.New("wrong start/end height for the task")
	ErrTaskExpired                    = errors.New("the task is already expired")
	ErrTaskNotAllowMultipleExecutions = errors.New("a task of the same type is already scheduled and allowed multiple execution for this type is false")
	ErrTaskIsNil                      = errors.New("the task we're trying to schedule is nil")
)

// InternalTaskState is an enumeration indicating the possible states of a task
type InternalTaskState int

const (
	NotStarted InternalTaskState = iota
	Running
	Killed
)

func (state InternalTaskState) String() string {
	return [...]string{
		"NotStarted",
		"Running",
		"Killed",
	}[state]
}

type BaseRequest struct {
	Id            string            `json:"id"`
	Name          string            `json:"name"`
	Start         uint64            `json:"start"`
	End           uint64            `json:"end"`
	InternalState InternalTaskState `json:"internalState"`
}

type TaskRequestInfo struct {
	BaseRequest
	Task tasks.Task
	*TaskSharedResponse
	killedAt uint64
}

type taskRequestInner struct {
	BaseRequest
	WrappedTask *marshaller.InstanceWrapper `json:"wrappedTask"`
}

type TaskSharedResponse struct {
	doneChan chan struct{}
	err      error // error in case the task failed
}

func newTaskResponse() *TaskSharedResponse {
	return &TaskSharedResponse{doneChan: make(chan struct{})}
}

// IsReady function to check if a receipt is ready
func (r *TaskSharedResponse) IsReady() bool {
	select {
	case <-r.doneChan:
		return true
	default:
		return false
	}
}

// GetTaskResponseBlocking blocking function to get the execution status of a task.
// This function will block until the task is finished and the final result is returned.
func (r *TaskSharedResponse) GetTaskResponseBlocking(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.doneChan:
		return r.err
	}
}

// writeResponse function to write the receipt or error from a transaction being watched.
func (r *TaskSharedResponse) writeResponse(err error) {
	if !r.IsReady() {
		r.err = err
		close(r.doneChan)
	}
}

type TaskManager struct {
	Schedule       map[string]TaskRequestInfo     `json:"schedule"`
	LastHeightSeen uint64                         `json:"lastHeightSeen"`
	mainCtx        context.Context                `json:"-"`
	eth            layer1.Client                  `json:"-"`
	database       *db.Database                   `json:"-"`
	adminHandler   monitorInterfaces.AdminHandler `json:"-"`
	marshaller     *marshaller.TypeRegistry       `json:"-"`
	requestChan    <-chan internalRequest         `json:"-"`
	responseChan   *taskResponseChan              `json:"-"`
	logger         *logrus.Entry                  `json:"-"`
	taskExecutor   *TaskExecutor                  `json:"-"`
}

func newTaskManager(mainCtx context.Context, eth layer1.Client, database *db.Database, adminHandler monitorInterfaces.AdminHandler, requestChan <-chan internalRequest, txWatcher *transaction.FrontWatcher) (*TaskManager, error) {
	taskManager := &TaskManager{
		Schedule:     make(map[string]TaskRequestInfo),
		mainCtx:      mainCtx,
		database:     database,
		eth:          eth,
		adminHandler: adminHandler,
		marshaller:   GetTaskRegistry(),
		requestChan:  requestChan,
		responseChan: &taskResponseChan{trChan: make(chan InternalTaskResponse, 100)},
	}

	logger := logging.GetLogger("tasks")
	taskManager.logger = logger.WithField("Component", "schedule")

	err := taskManager.loadState()
	taskManager.logger.Warnf("could not find previous State: %v", err)
	if err != badger.ErrKeyNotFound {
		return nil, err
	}

	tasksExecutor, err := newTaskExecutor(txWatcher, database, logger.WithField("Component", "executor"))
	if err != nil {
		return nil, err
	}
	taskManager.taskExecutor = tasksExecutor

	return taskManager, nil
}

type taskResponseChan struct {
	writeOnce sync.Once
	trChan    chan InternalTaskResponse
	isClosed  bool
}

func (tr *taskResponseChan) close() {
	tr.writeOnce.Do(func() {
		tr.isClosed = true
		close(tr.trChan)
	})
}

func (tr *taskResponseChan) Add(taskResponse InternalTaskResponse) {
	if !tr.isClosed {
		tr.trChan <- taskResponse
	}
}

var _ internalTaskResponseChan = &taskResponseChan{}

type innerSequentialSchedule struct {
	Schedule map[string]*taskRequestInner
}

func GetTaskLogger(task tasks.Task) *logrus.Entry {
	logger := logging.GetLogger("tasks")
	logEntry := logger.WithFields(logrus.Fields{
		"Component": "task",
		"taskStart": task.GetStart(),
		"taskEnd":   task.GetEnd(),
	})
	return logEntry
}

func GetTaskLoggerComplete(taskReq TaskRequestInfo) *logrus.Entry {
	logger := logging.GetLogger("tasks")
	logEntry := logger.WithFields(logrus.Fields{
		"Component": "task",
		"taskName":  taskReq.Name,
		"taskStart": taskReq.Task.GetStart(),
		"taskEnd":   taskReq.Task.GetEnd(),
		"taskId":    taskReq.Id,
		"state":     taskReq.InternalState,
	})
	return logEntry
}

func (tm *TaskManager) start() {
	tm.logger.Info("Starting task manager")
	tm.logger.Info(strings.Repeat("-", 80))
	tm.logger.Infof("Current Tasks: %d", len(tm.Schedule))
	for id, task := range tm.Schedule {
		tm.logger.Infof("...ID: %s Name: %s Between: %d and %d", id, task.Name, task.Start, task.End)
	}
	tm.logger.Info(strings.Repeat("-", 80))

	go tm.eventLoop()
}

func (tm *TaskManager) eventLoop() {
	processingTime := time.After(constants.TaskSchedulerProcessingTime)

	for {
		select {
		case <-tm.mainCtx.Done():
			tm.logger.Warn("Received closing context request")
			tm.responseChan.close()
			return

		case taskRequest := <-tm.requestChan:
			response := &TaskManagerResponse{Err: nil}
			switch taskRequest.action {
			case Kill:
				if taskRequest.task != nil {
					taskName, _ := marshaller.GetNameType(taskRequest.task)
					tm.logger.Tracef("received request to kill all tasks type: %v", taskName)
					err := tm.killTaskByName(tm.mainCtx, taskName)
					if err != nil {
						tm.logger.WithError(err).Errorf("Failed to killTaskByName %v", taskName)
						response = &TaskManagerResponse{Err: err}
					}
				} else {
					tm.logger.Tracef("received request to kill task with ID: %v", taskRequest.id)
					// todo: check if Id is empty
					err := tm.remove(taskRequest.id)
					if err != nil {
						tm.logger.WithError(err).Errorf("Failed to killTaskById %v", taskRequest.id)
						response = &TaskManagerResponse{Err: err}
					}
				}
			case Schedule:
				tm.logger.Trace("received request for a task")
				err := tm.schedule(tm.mainCtx, taskRequest.task, taskRequest.id)
				if err != nil {
					// if we are not synchronized, don't log expired task as errors, since we will
					// be replaying the events from far way in the past
					if errors.Is(err, ErrTaskExpired) && !tm.adminHandler.IsSynchronized() {
						tm.logger.WithError(err).Debugf("Failed to schedule task request %d", tm.LastHeightSeen)
					} else {
						tm.logger.WithError(err).Errorf("Failed to schedule task request %d", tm.LastHeightSeen)
						response = &TaskManagerResponse{Err: err}
					}
				}
			}
			taskRequest.response.sendResponse(response)
			err := tm.persistState()
			if err != nil {
				tm.logger.WithError(err).Errorf("Failed to persist state %d on task request", tm.LastHeightSeen)
			}

		case taskResponse := <-tm.responseChan.trChan:
			tm.logger.Trace("received a task response")
			err := tm.processTaskResponse(tm.mainCtx, taskResponse)
			if err != nil {
				tm.logger.WithError(err).Errorf("Failed to processTaskResponse %v", taskResponse)
			}
			err = tm.persistState()
			if err != nil {
				tm.logger.WithError(err).Errorf("Failed to persist state %d on task response", tm.LastHeightSeen)
			}
		case <-processingTime:
			tm.logger.Trace("processing latest height")
			networkCtx, networkCf := context.WithTimeout(tm.mainCtx, constants.TaskSchedulerNetworkTimeout)
			height, err := tm.eth.GetFinalizedHeight(networkCtx)
			networkCf()
			if err != nil {
				tm.logger.WithError(err).Debug("Failed to retrieve the latest height from eth node")
				processingTime = time.After(constants.TaskSchedulerProcessingTime)
				continue
			}
			tm.LastHeightSeen = height

			toStart, expired := tm.findTasks()
			err = tm.startTasks(tm.mainCtx, toStart)
			if err != nil {
				tm.logger.WithError(err).Errorf("Failed to startTasks %d", tm.LastHeightSeen)
			}
			err = tm.persistState()
			if err != nil {
				tm.logger.WithError(err).Errorf("Failed to persist state after start tasks %d", tm.LastHeightSeen)
			}

			err = tm.killTasks(tm.mainCtx, expired)
			if err != nil {
				tm.logger.WithError(err).Errorf("Failed to killExpiredTasks %d", tm.LastHeightSeen)
			}
			err = tm.persistState()
			if err != nil {
				tm.logger.WithError(err).Errorf("Failed to persist after kill tasks state %d", tm.LastHeightSeen)
			}
			processingTime = time.After(constants.TaskSchedulerProcessingTime)
		}
	}
}

func (tm *TaskManager) schedule(ctx context.Context, task tasks.Task, id string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if task == nil {
			return ErrTaskIsNil
		}

		start := task.GetStart()
		end := task.GetEnd()

		if start != 0 && end != 0 && start >= end {
			return ErrWrongParams
		}

		if end != 0 && end <= tm.LastHeightSeen {
			return ErrTaskExpired
		}

		if _, present := tm.Schedule[id]; present {
			// todo: return the sharedTaskResponse here
			return nil
		}

		taskName, _ := marshaller.GetNameType(task)
		if len(tm.findTasksByName(taskName)) > 0 && !task.GetAllowMultiExecution() {
			return ErrTaskNotAllowMultipleExecutions
		}

		taskReq := TaskRequestInfo{BaseRequest: BaseRequest{Id: id, Name: taskName, Start: start, End: end}, Task: task}
		tm.Schedule[id] = taskReq
		GetTaskLoggerComplete(taskReq).Debug("Received task request")
	}
	return nil
}

func (tm *TaskManager) processTaskResponse(ctx context.Context, taskResponse InternalTaskResponse) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		logger := tm.logger
		task, present := tm.Schedule[taskResponse.Id]
		if present {
			logger = GetTaskLoggerComplete(task)
		}
		// todo: send task response here
		if taskResponse.Err != nil {
			if !errors.Is(taskResponse.Err, context.Canceled) {
				logger.Errorf("Task executed with error: %v", taskResponse.Err)
			} else {
				logger.Debug("Task got killed")
			}
		} else {
			logger.Info("Task successfully executed")
		}
		err := tm.remove(taskResponse.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tm *TaskManager) startTasks(ctx context.Context, tasks []TaskRequestInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tm.logger.Debug("Looking for starting tasks")
		for i := 0; i < len(tasks); i++ {
			task := tasks[i]
			logEntry := GetTaskLogger(task.Task)
			logEntry = logEntry.WithField("taskId", task.Id).WithField("taskName", task.Name)
			GetTaskLoggerComplete(task).Info("task is about to start")

			go tm.taskExecutor.manageTask(ctx, task.Task, task.Name, task.Id, tm.database, logEntry, tm.eth, tm.responseChan)

			task.InternalState = Running
			tm.Schedule[task.Id] = task
		}

	}

	return nil
}

func (tm *TaskManager) killTaskByName(ctx context.Context, taskName string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tm.logger.Debugf("Looking for killing tasks by name %s", taskName)
		return tm.killTasks(ctx, tm.findTasksByName(taskName))
	}
}

func (tm *TaskManager) killTasks(ctx context.Context, tasks []TaskRequestInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		for i := 0; i < len(tasks); i++ {
			task := tasks[i]
			GetTaskLoggerComplete(task).Info("Task is about to be killed")
			if task.InternalState == Running {
				task.Task.Close()
				task.InternalState = Killed
				task.killedAt = tm.LastHeightSeen
			} else if task.InternalState == Killed {
				GetTaskLoggerComplete(task).Error("Task already killed")
			} else {
				GetTaskLoggerComplete(task).Trace("Task is not running yet, pruning directly")
				err := tm.remove(task.Id)
				if err != nil {
					tm.logger.WithError(err).Errorf("Failed to kill task id: %s", task.Id)
				}
			}
		}
	}

	return nil
}

func (tm *TaskManager) findTasks() ([]TaskRequestInfo, []TaskRequestInfo) {
	toStart := make([]TaskRequestInfo, 0)
	expired := make([]TaskRequestInfo, 0)
	unresponsive := make([]TaskRequestInfo, 0)
	multiExecutionCheck := make(map[string]bool)

	for _, taskRequest := range tm.Schedule {
		if taskRequest.InternalState == Killed && taskRequest.killedAt+constants.TaskSchedulerHeightToleranceBeforeRemoving <= tm.LastHeightSeen {
			tm.logger.Errorf("marking task as unresponsive %s", taskRequest.Task.GetId())
			unresponsive = append(unresponsive, taskRequest)
			continue
		}

		if taskRequest.End != 0 && taskRequest.End <= tm.LastHeightSeen {
			tm.logger.Tracef("marking task as expired %s", taskRequest.Task.GetId())
			expired = append(expired, taskRequest)
			continue
		}

		if ((taskRequest.Start == 0 && taskRequest.End == 0) ||
			(taskRequest.Start != 0 && taskRequest.Start <= tm.LastHeightSeen && taskRequest.End == 0) ||
			(taskRequest.Start <= tm.LastHeightSeen && taskRequest.End > tm.LastHeightSeen)) && taskRequest.InternalState == NotStarted {

			toStart = append(toStart, taskRequest)
			continue
		}
	}
	if len(unresponsive) > 0 {
		panic("found unresponsive tasks")
	}
	return toStart, expired
}

func (tm *TaskManager) findTasksByName(taskName string) []TaskRequestInfo {
	tm.logger.Tracef("trying to find tasks by name %s", taskName)
	tasks := make([]TaskRequestInfo, 0)

	for _, taskRequest := range tm.Schedule {
		if taskRequest.Name == taskName {
			tasks = append(tasks, taskRequest)
		}
	}
	tm.logger.Tracef("found %v tasks with name %s", len(tasks), taskName)
	return tasks
}

func (tm *TaskManager) findRunningTasksByName(taskName string) []TaskRequestInfo {
	tm.logger.Tracef("finding running tasks by name %s", taskName)
	tasks := make([]TaskRequestInfo, 0)

	for _, taskRequest := range tm.Schedule {
		if taskRequest.Name == taskName && taskRequest.InternalState == Running {
			tasks = append(tasks, taskRequest)
		}
	}
	tm.logger.Tracef("found %v running tasks with name %s", len(tasks), taskName)
	return tasks
}

func (tm *TaskManager) remove(id string) error {
	tm.logger.Tracef("removing task with id %s", id)
	_, present := tm.Schedule[id]
	if !present {
		return ErrNotScheduled
	}

	delete(tm.Schedule, id)

	return nil
}

func (tm *TaskManager) persistState() error {
	logger := logging.GetLogger("staterecover").WithField("State", "manager")
	rawData, err := json.Marshal(tm)
	if err != nil {
		return err
	}

	err = tm.database.Update(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskSchedulerState()
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

func (tm *TaskManager) loadState() error {
	logger := logging.GetLogger("staterecover").WithField("State", "manager")
	if err := tm.database.View(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskSchedulerState()
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

	// If the tasks were running, we mark them as not started so the scheduled can
	// start them again
	for _, task := range tm.Schedule {
		if task.InternalState == Running {
			task.InternalState = NotStarted
			tm.Schedule[task.Id] = task
		} else if task.InternalState == Killed {
			tm.remove(task.Id)
		}
	}

	// synchronizing db state to disk
	if err := tm.database.Sync(); err != nil {
		logger.Error("Failed to set sync")
		return err
	}

	return nil

}

func (tm *TaskManager) MarshalJSON() ([]byte, error) {

	ws := &innerSequentialSchedule{Schedule: make(map[string]*taskRequestInner)}

	for k, v := range tm.Schedule {
		wt, err := tm.marshaller.WrapInstance(v.Task)
		if err != nil {
			return []byte{}, err
		}
		ws.Schedule[k] = &taskRequestInner{BaseRequest: v.BaseRequest, WrappedTask: wt}
	}

	raw, err := json.Marshal(&ws)
	if err != nil {
		return []byte{}, err
	}

	return raw, nil
}

func (tm *TaskManager) UnmarshalJSON(raw []byte) error {
	aa := &innerSequentialSchedule{}

	err := json.Unmarshal(raw, aa)
	if err != nil {
		return err
	}

	adminInterface := reflect.TypeOf((*monitorInterfaces.AdminClient)(nil)).Elem()

	tm.Schedule = make(map[string]TaskRequestInfo)
	for k, v := range aa.Schedule {
		t, err := tm.marshaller.UnwrapInstance(v.WrappedTask)
		if err != nil {
			return err
		}

		// Marshalling service handlers is mostly non-sense, so
		isAdminClient := reflect.TypeOf(t).Implements(adminInterface)
		if isAdminClient {
			adminClient := t.(monitorInterfaces.AdminClient)
			adminClient.SetAdminHandler(tm.adminHandler)
		}

		tm.Schedule[k] = TaskRequestInfo{BaseRequest: v.BaseRequest, Task: t.(tasks.Task)}
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
