package executor

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/snapshots"
	"reflect"
	"strings"
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

var _ TaskResponse = &HandlerResponse{}

type TaskManager struct {
	Schedule       map[string]ManagerRequestInfo  `json:"schedule"`
	LastHeightSeen uint64                         `json:"lastHeightSeen"`
	mainCtx        context.Context                `json:"-"`
	eth            layer1.Client                  `json:"-"`
	database       *db.Database                   `json:"-"`
	adminHandler   monitorInterfaces.AdminHandler `json:"-"`
	marshaller     *marshaller.TypeRegistry       `json:"-"`
	requestChan    <-chan managerRequest          `json:"-"`
	responseChan   *executorResponseChan          `json:"-"`
	logger         *logrus.Entry                  `json:"-"`
	taskExecutor   *TaskExecutor                  `json:"-"`
}

func newTaskManager(mainCtx context.Context, eth layer1.Client, database *db.Database, adminHandler monitorInterfaces.AdminHandler, requestChan <-chan managerRequest, txWatcher *transaction.FrontWatcher) (*TaskManager, error) {
	taskManager := &TaskManager{
		Schedule:     make(map[string]ManagerRequestInfo),
		mainCtx:      mainCtx,
		database:     database,
		eth:          eth,
		adminHandler: adminHandler,
		marshaller:   getTaskRegistry(),
		requestChan:  requestChan,
		responseChan: &executorResponseChan{erChan: make(chan ExecutorResponse, 100)},
	}

	logger := logging.GetLogger("tasks")
	taskManager.logger = logger.WithField("Component", "schedule")

	err := taskManager.loadState()
	if err != nil {
		taskManager.logger.Warnf("could not find previous State: %v", err)
		if err != badger.ErrKeyNotFound {
			return nil, err
		}
	}

	tasksExecutor, err := newTaskExecutor(txWatcher, database, logger.WithField("Component", "executor"))
	if err != nil {
		return nil, err
	}
	taskManager.taskExecutor = tasksExecutor

	return taskManager, nil
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
			response := &managerResponse{Err: nil}
			switch taskRequest.action {
			case KillByType:
				err := tm.killTaskByType(tm.mainCtx, taskRequest.task)
				if err != nil {
					tm.logger.WithError(err).Errorf("Failed to killTaskByType %v", taskRequest.task)
					response.Err = err
				}
			case KillById:
				err := tm.killTaskById(tm.mainCtx, taskRequest.id)
				if err != nil {
					tm.logger.WithError(err).Errorf("Failed to killTaskById %v", taskRequest.id)
					response.Err = err
				}
			case Schedule:
				tm.logger.Trace("received request for a task")
				err, sharedResponse := tm.schedule(tm.mainCtx, taskRequest.task, taskRequest.id)
				if err != nil {
					// if we are not synchronized, don't log expired task as errors, since we will
					// be replaying the events from far way in the past
					if errors.Is(err, ErrTaskExpired) && !tm.adminHandler.IsSynchronized() {
						tm.logger.WithError(err).Debugf("Failed to schedule task request %d", tm.LastHeightSeen)
					} else {
						tm.logger.WithError(err).Errorf("Failed to schedule task request %d", tm.LastHeightSeen)
						response.Err = err
					}
				} else {
					response.HandlerResponse = sharedResponse
				}
			}
			taskRequest.response.sendResponse(response)
			err := tm.persistState()
			if err != nil {
				tm.logger.WithError(err).Errorf("Failed to persist state %d on task request", tm.LastHeightSeen)
			}

		case taskResponse := <-tm.responseChan.erChan:
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

func (tm *TaskManager) schedule(ctx context.Context, task tasks.Task, id string) (error, *HandlerResponse) {
	select {
	case <-ctx.Done():
		return ctx.Err(), nil
	default:
		if task == nil {
			return ErrTaskIsNil, nil
		}

		if id == "" {
			return ErrTaskIdEmpty, nil
		}

		taskName, _, present := tm.marshaller.GetNameType(task)
		if !present {
			return ErrTaskTypeNotInRegistry, nil
		}

		start := task.GetStart()
		end := task.GetEnd()

		if start != 0 && end != 0 && start >= end {
			return ErrWrongParams, nil
		}

		if end != 0 && end <= tm.LastHeightSeen {
			return ErrTaskExpired, nil
		}

		if scheduledTask, exists := tm.Schedule[id]; exists {
			return nil, scheduledTask.HandlerResponse
		}

		if !task.GetAllowMultiExecution() && len(tm.findTasksByName(taskName)) > 0 {
			return ErrTaskNotAllowMultipleExecutions, nil
		}

		taskResp := newHandlerResponse()
		taskReq := ManagerRequestInfo{
			BaseRequest: BaseRequest{
				Id:            id,
				Name:          taskName,
				Start:         start,
				End:           end,
				InternalState: NotStarted,
			},
			Task:            task,
			HandlerResponse: taskResp,
		}
		tm.Schedule[id] = taskReq
		getTaskLoggerComplete(taskReq).Debug("Received task request")

		return nil, taskResp
	}
}

func (tm *TaskManager) processTaskResponse(ctx context.Context, executorResponse ExecutorResponse) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		logger := tm.logger
		task, present := tm.Schedule[executorResponse.Id]
		if !present {
			tm.logger.Warnf("received an internal response for non existing task with id %s", executorResponse.Id)
			return nil
		}

		task.HandlerResponse.writeResponse(executorResponse.Err)
		task.ExecutorResponse = executorResponse
		logger = getTaskLoggerComplete(task)
		if executorResponse.Err != nil {
			if !errors.Is(executorResponse.Err, context.Canceled) {
				logger.Errorf("Task executed with error: %v", executorResponse.Err)
			} else {
				logger.Debug("Task got killed")
			}
		} else {
			logger.Info("Task successfully executed")
		}
		err := tm.remove(executorResponse.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tm *TaskManager) startTasks(ctx context.Context, tasks []ManagerRequestInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tm.logger.Debug("Looking for starting tasks")
		for i := 0; i < len(tasks); i++ {
			task := tasks[i]
			logEntry := getTaskLogger(task.Task)
			logEntry = logEntry.WithField("taskId", task.Id).WithField("taskName", task.Name)
			getTaskLoggerComplete(task).Info("task is about to start")

			go tm.taskExecutor.manageTask(ctx, task.Task, task.Name, task.Id, tm.database, logEntry, tm.eth, tm.responseChan)

			task.InternalState = Running
			tm.Schedule[task.Id] = task
		}

	}

	return nil
}

func (tm *TaskManager) killTaskByType(ctx context.Context, task tasks.Task) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if task == nil {
			return ErrTaskIsNil
		}

		taskName, _, present := tm.marshaller.GetNameType(task)
		if !present {
			return ErrTaskTypeNotInRegistry
		}

		tm.logger.Tracef("received request to kill all tasks type: %v", taskName)
		return tm.killTasks(ctx, tm.findTasksByName(taskName))
	}
}

func (tm *TaskManager) killTaskById(ctx context.Context, id string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if id == "" {
			return ErrTaskIdEmpty
		}

		tm.logger.Tracef("received request to kill task with id: %s", id)
		return tm.remove(id)
	}
}

func (tm *TaskManager) killTasks(ctx context.Context, tasks []ManagerRequestInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		for i := 0; i < len(tasks); i++ {
			task := tasks[i]
			getTaskLoggerComplete(task).Info("Task is about to be killed")
			if task.InternalState == Running {
				task.Task.Close()
				task.InternalState = Killed
				task.killedAt = tm.LastHeightSeen
			} else if task.InternalState == Killed {
				getTaskLoggerComplete(task).Error("Task already killed")
			} else {
				getTaskLoggerComplete(task).Trace("Task is not running yet, pruning directly")
				err := tm.remove(task.Id)
				if err != nil {
					tm.logger.WithError(err).Errorf("Failed to kill task id: %s", task.Id)
				}
			}
		}
	}

	return nil
}

func (tm *TaskManager) findTasks() ([]ManagerRequestInfo, []ManagerRequestInfo) {
	toStart := make([]ManagerRequestInfo, 0)
	expired := make([]ManagerRequestInfo, 0)
	unresponsive := make([]ManagerRequestInfo, 0)

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

func (tm *TaskManager) findTasksByName(taskName string) []ManagerRequestInfo {
	tm.logger.Tracef("trying to find tasks by name %s", taskName)
	tasks := make([]ManagerRequestInfo, 0)

	for _, taskRequest := range tm.Schedule {
		if taskRequest.Name == taskName {
			tasks = append(tasks, taskRequest)
		}
	}
	tm.logger.Tracef("found %v tasks with name %s", len(tasks), taskName)
	return tasks
}

func (tm *TaskManager) findRunningTasksByName(taskName string) []ManagerRequestInfo {
	tm.logger.Tracef("finding running tasks by name %s", taskName)
	tasks := make([]ManagerRequestInfo, 0)

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
			err := tm.remove(task.Id)
			if err != nil {
				return err
			}
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

	ws := &innerSequentialSchedule{Schedule: make(map[string]*requestStored)}

	for k, v := range tm.Schedule {
		wt, err := tm.marshaller.WrapInstance(v.Task)
		if err != nil {
			return []byte{}, err
		}
		ws.Schedule[k] = &requestStored{BaseRequest: v.BaseRequest, WrappedTask: wt}
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

	tm.Schedule = make(map[string]ManagerRequestInfo)
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

		handlerResponse := newHandlerResponse()
		if v.ExecutorResponse.Err != nil {
			handlerResponse.writeResponse(v.ExecutorResponse.Err)
		}
		tm.Schedule[k] = ManagerRequestInfo{
			BaseRequest:      v.BaseRequest,
			ExecutorResponse: v.ExecutorResponse,
			HandlerResponse:  handlerResponse,
			Task:             t.(tasks.Task),
		}
	}

	return nil
}

func getTaskLogger(task tasks.Task) *logrus.Entry {
	logger := logging.GetLogger("tasks")
	logEntry := logger.WithFields(logrus.Fields{
		"Component": "task",
		"taskStart": task.GetStart(),
		"taskEnd":   task.GetEnd(),
	})
	return logEntry
}

func getTaskLoggerComplete(taskReq ManagerRequestInfo) *logrus.Entry {
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

func getTaskRegistry() *marshaller.TypeRegistry {
	// registry the task type here
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
