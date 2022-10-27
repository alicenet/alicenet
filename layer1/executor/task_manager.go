package executor

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/alicenet/alicenet/layer1/executor/tasks/accusations"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/snapshots"

	"github.com/alicenet/alicenet/consensus/db"
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
	Responses      map[string]ManagerResponseInfo `json:"responses"`
	LastHeightSeen uint64                         `json:"lastHeightSeen"`
	mainCtx        context.Context                `json:"-"`
	eth            layer1.Client                  `json:"-"`
	contracts      layer1.AllSmartContracts       `json:"-"`
	database       *db.Database                   `json:"-"`
	adminHandler   monitorInterfaces.AdminHandler `json:"-"`
	marshaller     *marshaller.TypeRegistry       `json:"-"`
	closeChan      chan struct{}                  `json:"-"`
	closeOnce      sync.Once                      `json:"-"`
	requestChan    <-chan managerRequest          `json:"-"`
	responseChan   *executorResponseChan          `json:"-"`
	logger         *logrus.Entry                  `json:"-"`
	taskExecutor   *TaskExecutor                  `json:"-"`
}

// newTaskManager creates a new TaskManager instance and recover the previous state from DB.
func newTaskManager(mainCtx context.Context, eth layer1.Client, contracts layer1.AllSmartContracts, database *db.Database, logger *logrus.Entry, adminHandler monitorInterfaces.AdminHandler, requestChan <-chan managerRequest, txWatcher transaction.Watcher) (*TaskManager, error) {
	taskManager := &TaskManager{
		Schedule:     make(map[string]ManagerRequestInfo),
		Responses:    make(map[string]ManagerResponseInfo),
		mainCtx:      mainCtx,
		database:     database,
		eth:          eth,
		contracts:    contracts,
		adminHandler: adminHandler,
		marshaller:   getTaskRegistry(),
		closeChan:    make(chan struct{}),
		closeOnce:    sync.Once{},
		requestChan:  requestChan,
		responseChan: &executorResponseChan{erChan: make(chan ExecutorResponse, 100)},
		logger:       logger,
	}

	err := taskManager.recoverState()
	if err != nil {
		taskManager.logger.Warnf("could not find previous State: %v", err)
		if !errors.Is(err, badger.ErrKeyNotFound) {
			return nil, err
		}
	}

	tasksExecutor, err := newTaskExecutor(txWatcher, database, logger.WithField("Component", "TaskExecutor"))
	if err != nil {
		return nil, err
	}
	taskManager.taskExecutor = tasksExecutor

	return taskManager, nil
}

// start the infinite loop TaskManager execution.
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

// close the TaskManager execution.
func (tm *TaskManager) close() {
	tm.closeOnce.Do(func() {
		tm.logger.Warn("Closing task manager")
		close(tm.closeChan)
		tm.responseChan.close()
	})
}

// eventLoop where the interaction with all the pieces is developed.
func (tm *TaskManager) eventLoop() {
	processingTime := time.After(tasks.ManagerProcessingTime)

	for {
		select {
		case <-tm.closeChan:
			tm.logger.Warn("Received closing request")
			return

		case taskRequest, ok := <-tm.requestChan:
			if !ok {
				tm.logger.Warnf("Received a request on a closed channel %v", taskRequest)
				return
			}
			if taskRequest.response == nil {
				tm.logger.Warn("Received a request with nil response channel")
				return
			}

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

		case taskResponse, ok := <-tm.responseChan.erChan:
			if !ok {
				tm.logger.Warnf("Received a taskResponse on a closed channel %v", taskResponse)
				return
			}

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
			networkCtx, networkCf := context.WithTimeout(tm.mainCtx, tasks.ManagerNetworkTimeout)
			height, err := tm.eth.GetFinalizedHeight(networkCtx)
			networkCf()
			if err != nil {
				tm.logger.WithError(err).Debug("Failed to retrieve the latest height from eth node")
				processingTime = time.After(tasks.ManagerProcessingTime)
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

			tm.cleanResponses()
			err = tm.persistState()
			if err != nil {
				tm.logger.WithError(err).Errorf("Failed to persist after kill tasks state %d", tm.LastHeightSeen)
			}
			processingTime = time.After(tasks.ManagerProcessingTime)
		}
	}
}

// schedule the task after its validation.
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

		if resp, exists := tm.Responses[id]; exists {
			return nil, resp.HandlerResponse
		}

		taskName, _, present := tm.marshaller.NameTypeIsPresent(task)
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

		if !task.GetAllowMultiExecution() && len(tm.findTasksByName(taskName)) > 0 {
			return ErrTaskNotAllowMultipleExecutions, nil
		}

		taskReq := ManagerRequestInfo{
			BaseRequest: BaseRequest{
				Id:                  id,
				Name:                taskName,
				Start:               start,
				End:                 end,
				AllowMultiExecution: task.GetAllowMultiExecution(),
				SubscribeOptions:    task.GetSubscribeOptions(),
				InternalState:       NotStarted,
			},
			Task: task,
		}
		tm.Schedule[id] = taskReq

		taskResp := ManagerResponseInfo{
			HandlerResponse: newHandlerResponse(),
		}
		tm.Responses[id] = taskResp
		getTaskLoggerComplete(tm.logger, taskReq).Debug("Received task request")

		return nil, taskResp.HandlerResponse
	}
}

// processTaskResponse from the TaskExecutor and writes it to the HandlerResponse.
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

		taskResp, present := tm.Responses[executorResponse.Id]
		if !present {
			tm.logger.Warnf("response structure doesn't exist for a received response with id %s", executorResponse.Id)
			return nil
		}

		taskResp.ReceivedOnBlock = tm.LastHeightSeen
		taskResp.ExecutorResponse = executorResponse
		taskResp.HandlerResponse.writeResponse(executorResponse.Err)
		tm.Responses[executorResponse.Id] = taskResp

		logger = getTaskLoggerComplete(tm.logger, task)
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

// startTasks spawning a go routine to handle Task execution using the TaskExecutor.
func (tm *TaskManager) startTasks(ctx context.Context, tasks []ManagerRequestInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tm.logger.Debug("Looking for starting tasks")
		for i := 0; i < len(tasks); i++ {
			task := tasks[i]
			logEntry := getTaskLoggerComplete(tm.logger, task)
			logEntry.Info("task is about to start")

			go tm.taskExecutor.handleTaskExecution(ctx, task.Task, task.Name, task.Id, task.Start, task.End, task.AllowMultiExecution, task.SubscribeOptions, tm.database, logEntry, tm.eth, tm.contracts, tm.responseChan)

			task.InternalState = Running
			tm.Schedule[task.Id] = task
		}

	}

	return nil
}

// killTaskByType sends task to kill after the request validation.
func (tm *TaskManager) killTaskByType(ctx context.Context, task tasks.Task) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if task == nil {
			return ErrTaskIsNil
		}

		taskName, _, present := tm.marshaller.NameTypeIsPresent(task)
		if !present {
			return ErrTaskTypeNotInRegistry
		}

		tm.logger.Tracef("received request to kill all tasks type: %v", taskName)
		return tm.killTasks(ctx, tm.findTasksByName(taskName))
	}
}

// killTaskById sends task to kill after the request validation.
func (tm *TaskManager) killTaskById(ctx context.Context, id string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if id == "" {
			return ErrTaskIdEmpty
		}

		task, present := tm.Schedule[id]
		if !present {
			return ErrNotScheduled
		}

		tm.logger.Tracef("received request to kill task with id: %s", id)
		return tm.killTask(task)
	}
}

// iterates a list of Task to killTasks
func (tm *TaskManager) killTasks(ctx context.Context, tasks []ManagerRequestInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		for i := 0; i < len(tasks); i++ {
			task := tasks[i]
			err := tm.killTask(task)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// killTask in a way depending on its state.
func (tm *TaskManager) killTask(task ManagerRequestInfo) error {
	getTaskLoggerComplete(tm.logger, task).Info("Task is about to be killed")
	if task.InternalState == Running {
		task.Task.Close()
		task.InternalState = Killed
		task.killedAt = tm.LastHeightSeen
		tm.Schedule[task.Id] = task
	} else if task.InternalState == Killed {
		getTaskLoggerComplete(tm.logger, task).Error("Task already killed")
	} else {
		getTaskLoggerComplete(tm.logger, task).Trace("Task is not running yet, pruning directly")

		taskResp, present := tm.Responses[task.Id]
		if !present {
			tm.logger.Warnf("response structure doesn't exist for a received killing request with id %s", task.Id)
		}

		taskResp.ReceivedOnBlock = tm.LastHeightSeen
		executorResponse := ExecutorResponse{
			Id:  task.Id,
			Err: ErrTaskKilledBeforeExecution,
		}
		taskResp.ExecutorResponse = executorResponse
		taskResp.HandlerResponse.writeResponse(executorResponse.Err)
		tm.Responses[executorResponse.Id] = taskResp

		err := tm.remove(task.Id)
		if err != nil {
			tm.logger.WithError(err).Errorf("Failed to kill task id: %s", task.Id)
			return err
		}
	}
	return nil
}

// cleanResponses after constants.TaskManagerResponseToleranceBeforeRemoving amount of blocks.
func (tm *TaskManager) cleanResponses() {
	for id, resp := range tm.Responses {
		if resp.ReceivedOnBlock != 0 && resp.ReceivedOnBlock+tasks.ManagerResponseToleranceBeforeRemoving <= tm.LastHeightSeen {
			delete(tm.Responses, id)
		}
	}
}

// findTasks for killing and starting.
func (tm *TaskManager) findTasks() ([]ManagerRequestInfo, []ManagerRequestInfo) {
	toStart := make([]ManagerRequestInfo, 0)
	expired := make([]ManagerRequestInfo, 0)
	unresponsive := make([]ManagerRequestInfo, 0)

	for _, taskRequest := range tm.Schedule {
		if taskRequest.InternalState == Killed && taskRequest.killedAt+tasks.ManagerHeightToleranceBeforeRemoving <= tm.LastHeightSeen {
			tm.logger.Errorf("marking task as unresponsive %s", taskRequest.Id)
			unresponsive = append(unresponsive, taskRequest)
			continue
		}

		if taskRequest.End != 0 && taskRequest.End <= tm.LastHeightSeen {
			tm.logger.Tracef("marking task as expired %s", taskRequest.Id)
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

// findTasksByName in the Schedule.
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

// remove Task from Schedule.
func (tm *TaskManager) remove(id string) error {
	tm.logger.Tracef("removing task with id %s", id)
	_, present := tm.Schedule[id]
	if !present {
		return ErrNotScheduled
	}

	delete(tm.Schedule, id)

	return nil
}

// persistState TaskManager to database.
func (tm *TaskManager) persistState() error {
	logger := logging.GetLogger("staterecover").WithField("State", "manager")
	rawData, err := json.Marshal(tm)
	if err != nil {
		return err
	}

	err = tm.database.Update(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskManagerState()
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

// loadState TaskManager from database.
func (tm *TaskManager) loadState() error {
	logger := logging.GetLogger("staterecover").WithField("State", "manager")
	if err := tm.database.View(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskManagerState()
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

	// synchronizing db state to disk
	if err := tm.database.Sync(); err != nil {
		logger.Error("Failed to set sync")
		return err
	}

	return nil
}

// recoverState TaskManager modifying the Schedule and Responses in order to operate properly.
func (tm *TaskManager) recoverState() error {
	err := tm.loadState()
	if err != nil {
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

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (tm *TaskManager) MarshalJSON() ([]byte, error) {

	ws := &taskManagerBackup{Schedule: make(map[string]requestStored), Responses: make(map[string]responseStored), LastHeightSeen: tm.LastHeightSeen}

	for k, v := range tm.Schedule {
		wt, err := func() (*marshaller.InstanceWrapper, error) {
			v.Task.Lock()
			defer v.Task.Unlock()
			return tm.marshaller.WrapInstance(v.Task)
		}()

		if err != nil {
			return []byte{}, err
		}

		ws.Schedule[k] = requestStored{BaseRequest: v.BaseRequest, WrappedTask: wt, killedAt: v.killedAt}
	}

	for k, v := range tm.Responses {
		responseStored := responseStored{
			ReceivedOnBlock: v.ReceivedOnBlock,
		}

		if v.Err != nil {
			responseStored.ErrMsg = v.Err.Error()
		}
		ws.Responses[k] = responseStored
	}

	raw, err := json.Marshal(&ws)
	if err != nil {
		return []byte{}, err
	}

	return raw, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (tm *TaskManager) UnmarshalJSON(raw []byte) error {
	aa := &taskManagerBackup{}

	err := json.Unmarshal(raw, aa)
	if err != nil {
		return err
	}

	adminInterface := reflect.TypeOf((*monitorInterfaces.AdminClient)(nil)).Elem()

	tm.Schedule = make(map[string]ManagerRequestInfo)
	tm.Responses = make(map[string]ManagerResponseInfo)
	tm.LastHeightSeen = aa.LastHeightSeen
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

		tm.Schedule[k] = ManagerRequestInfo{
			BaseRequest: v.BaseRequest,
			Task:        t.(tasks.Task),
			killedAt:    v.killedAt,
		}
	}

	for k, v := range aa.Responses {
		resp := ManagerResponseInfo{
			ReceivedOnBlock: v.ReceivedOnBlock,
			HandlerResponse: newHandlerResponse(),
		}

		if v.ReceivedOnBlock != 0 {
			resp.ExecutorResponse = ExecutorResponse{
				Id:  k,
				Err: errors.New(v.ErrMsg),
			}
			resp.HandlerResponse.writeResponse(resp.ExecutorResponse.Err)
		}
		tm.Responses[k] = resp
	}

	return nil
}

// getTaskLoggerComplete with all the fields.
func getTaskLoggerComplete(logger *logrus.Entry, taskReq ManagerRequestInfo) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"Component": "task",
		"taskName":  taskReq.Name,
		"taskStart": taskReq.Start,
		"taskEnd":   taskReq.End,
		"taskId":    taskReq.Id,
		"state":     taskReq.InternalState,
	})
}

// getTaskRegistry all the Tasks we can handle in the request.
// If you want to create a new task register its instance type here.
func getTaskRegistry() *marshaller.TypeRegistry {
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
	tr.RegisterInstanceType(&dkg.InitializeTask{})
	tr.RegisterInstanceType(&dkg.DisputeMissingRegistrationTask{})
	tr.RegisterInstanceType(&dkg.ShareDistributionTask{})
	tr.RegisterInstanceType(&snapshots.SnapshotTask{})
	tr.RegisterInstanceType(&accusations.MultipleProposalAccusationTask{})
	return tr
}
