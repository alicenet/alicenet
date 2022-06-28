package executor

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/snapshots"
	monitorInterfaces "github.com/alicenet/alicenet/layer1/monitor/interfaces"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	ErrNotScheduled = errors.New("scheduled task not found")
	ErrWrongParams  = errors.New("wrong start/end height for the task")
	ErrTaskExpired  = errors.New("the task is already expired")
	ErrTaskIsNil    = errors.New("the task we're trying to schedule is nil")
)

type TaskRequestInfo struct {
	Id        string
	Name      string
	Start     uint64
	End       uint64
	Task      tasks.Task
	isRunning bool
}

type taskRequestInner struct {
	Id          string                      `json:"id"`
	Name        string                      `json:"name"`
	Start       uint64                      `json:"start"`
	End         uint64                      `json:"end"`
	WrappedTask *marshaller.InstanceWrapper `json:"wrappedTask"`
}

type TasksScheduler struct {
	Schedule         map[string]TaskRequestInfo     `json:"schedule"`
	LastHeightSeen   uint64                         `json:"last_height_seen"`
	mainCtx          context.Context                `json:"-"`
	mainCtxCf        context.CancelFunc             `json:"-"`
	eth              layer1.Client                  `json:"-"`
	database         *db.Database                   `json:"-"`
	adminHandler     monitorInterfaces.AdminHandler `json:"-"`
	marshaller       *marshaller.TypeRegistry       `json:"-"`
	cancelChan       chan bool                      `json:"-"`
	taskRequestChan  <-chan tasks.TaskRequest       `json:"-"`
	taskResponseChan *taskResponseChan              `json:"-"`
	logger           *logrus.Entry                  `json:"-"`
	tasksManager     *TasksManager                  `json:"-"`
	txWatcher        *transaction.FrontWatcher      `json:"-"`
}

type taskResponseChan struct {
	writeOnce sync.Once
	trChan    chan tasks.TaskResponse
	isClosed  bool
}

func (tr *taskResponseChan) close() {
	tr.writeOnce.Do(func() {
		tr.isClosed = true
		close(tr.trChan)
	})
}

func (tr *taskResponseChan) Add(taskResponse tasks.TaskResponse) {
	if !tr.isClosed {
		tr.trChan <- taskResponse
	}
}

var _ tasks.TaskResponseChan = &taskResponseChan{}

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
		"isRunning": taskReq.isRunning,
	})
	return logEntry
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

func NewTasksScheduler(database *db.Database, eth layer1.Client, adminHandler monitorInterfaces.AdminHandler, taskRequestChan <-chan tasks.TaskRequest, txWatcher *transaction.FrontWatcher) (*TasksScheduler, error) {
	tr := GetTaskRegistry()

	// main context that will cancel all workers and go routine
	mainCtx, cf := context.WithCancel(context.Background())

	s := &TasksScheduler{
		Schedule:         make(map[string]TaskRequestInfo),
		mainCtx:          mainCtx,
		mainCtxCf:        cf,
		database:         database,
		eth:              eth,
		adminHandler:     adminHandler,
		marshaller:       tr,
		cancelChan:       make(chan bool, 1),
		taskRequestChan:  taskRequestChan,
		taskResponseChan: &taskResponseChan{trChan: make(chan tasks.TaskResponse, 100)},
		txWatcher:        txWatcher,
	}

	logger := logging.GetLogger("tasks")
	s.logger = logger.WithField("Component", "schedule")

	tasksManager, err := NewTaskManager(txWatcher, database, logger.WithField("Component", "manager"))
	if err != nil {
		return nil, err
	}
	s.tasksManager = tasksManager

	return s, nil
}

func (s *TasksScheduler) Start() error {
	err := s.loadState()
	if err != nil {
		s.logger.Warnf("could not find previous State: %v", err)
		if err != badger.ErrKeyNotFound {
			return err
		}
	}

	s.logger.Info(strings.Repeat("-", 80))
	s.logger.Infof("Current Tasks: %d", len(s.Schedule))
	for id, task := range s.Schedule {
		s.logger.Infof("...ID: %s Name: %s Between: %d and %d", id, task.Name, task.Start, task.End)
	}
	s.logger.Info(strings.Repeat("-", 80))

	go s.eventLoop()
	return nil
}

func (s *TasksScheduler) Close() {
	s.logger.Warn("Closing scheduler")
	s.cancelChan <- true
}

func (s *TasksScheduler) eventLoop() {
	processingTime := time.After(constants.TaskSchedulerProcessingTime)

	for {
		select {
		case <-s.cancelChan:
			s.logger.Warn("Received cancel request for event loop.")
			s.mainCtxCf()
			s.taskResponseChan.close()
			return

		case taskRequest := <-s.taskRequestChan:
			switch taskRequest.Action {
			case tasks.Kill:
				taskName, _ := marshaller.GetNameType(taskRequest.Task)
				s.logger.Tracef("received request to kill all tasks type: %v", taskName)
				err := s.killTaskByName(s.mainCtx, taskName)
				if err != nil {
					s.logger.WithError(err).Errorf("Failed to killTaskByName %v", taskName)
				}
			case tasks.Schedule:
				s.logger.Trace("received request for a task")
				err := s.schedule(s.mainCtx, taskRequest.Task)
				if err != nil {
					// if we are not synchronized, don't log expired task as errors, since we will
					// be replaying the events from far way in the past
					if errors.Is(err, ErrTaskExpired) && !s.adminHandler.IsSynchronized() {
						s.logger.WithError(err).Debugf("Failed to schedule task request %d", s.LastHeightSeen)
					} else {
						s.logger.WithError(err).Errorf("Failed to schedule task request %d", s.LastHeightSeen)
					}
				}
			}
			err := s.persistState()
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to persist state %d on task request", s.LastHeightSeen)
			}

		case taskResponse := <-s.taskResponseChan.trChan:
			s.logger.Trace("received a task response")
			err := s.processTaskResponse(s.mainCtx, taskResponse)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to processTaskResponse %v", taskResponse)
			}
			err = s.persistState()
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to persist state %d on task response", s.LastHeightSeen)
			}
		case <-processingTime:
			s.logger.Trace("processing latest height")
			networkCtx, networkCf := context.WithTimeout(s.mainCtx, constants.TaskSchedulerNetworkTimeout)
			height, err := s.eth.GetFinalizedHeight(networkCtx)
			networkCf()
			if err != nil {
				s.logger.WithError(err).Debug("Failed to retrieve the latest height from eth node")
				processingTime = time.After(constants.TaskSchedulerProcessingTime)
				continue
			}
			s.LastHeightSeen = height

			toStart, expired, unresponsive := s.findTasks()
			err = s.startTasks(s.mainCtx, toStart)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to startTasks %d", s.LastHeightSeen)
			}
			err = s.persistState()
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to persist state %d", s.LastHeightSeen)
			}

			err = s.killTasks(s.mainCtx, expired)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to killExpiredTasks %d", s.LastHeightSeen)
			}

			err = s.removeUnresponsiveTasks(s.mainCtx, unresponsive)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to removeUnresponsiveTasks %d", s.LastHeightSeen)
			}
			err = s.persistState()
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to persist state %d", s.LastHeightSeen)
			}
			processingTime = time.After(constants.TaskSchedulerProcessingTime)
		}
	}
}

func (s *TasksScheduler) schedule(ctx context.Context, task tasks.Task) error {
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

		if end != 0 && end <= s.LastHeightSeen {
			return ErrTaskExpired
		}

		id := uuid.New()
		taskName, _ := marshaller.GetNameType(task)
		taskReq := TaskRequestInfo{Id: id.String(), Name: taskName, Start: start, End: end, Task: task}
		s.Schedule[id.String()] = taskReq
		GetTaskLoggerComplete(taskReq).Debug("Received task request")
	}
	return nil
}

func (s *TasksScheduler) processTaskResponse(ctx context.Context, taskResponse tasks.TaskResponse) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		logger := s.logger
		task, present := s.Schedule[taskResponse.Id]
		if present {
			logger = GetTaskLoggerComplete(task)
		}
		if taskResponse.Err != nil {
			if taskResponse.Err != context.Canceled {
				logger.Errorf("Task executed with error: %v", taskResponse.Err)
			} else {
				logger.Debug("Task got killed")
			}
		} else {
			logger.Info("Task successfully executed")
		}
		err := s.remove(taskResponse.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *TasksScheduler) startTasks(ctx context.Context, tasks []TaskRequestInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.logger.Debug("Looking for starting tasks")
		for i := 0; i < len(tasks); i++ {
			task := tasks[i]
			logEntry := GetTaskLogger(task.Task)
			logEntry = logEntry.WithField("taskId", task.Id).WithField("taskName", task.Name)
			GetTaskLoggerComplete(task).Info("task is about to start")

			go s.tasksManager.ManageTask(ctx, task.Task, task.Name, task.Id, s.database, logEntry, s.eth, s.taskResponseChan)

			task.isRunning = true
			s.Schedule[task.Id] = task
		}

	}

	return nil
}

func (s *TasksScheduler) killTaskByName(ctx context.Context, taskName string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.logger.Debugf("Looking for killing tasks by name %s", taskName)
		return s.killTasks(ctx, s.findTasksByName(taskName))
	}
}

func (s *TasksScheduler) killTasks(ctx context.Context, tasks []TaskRequestInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		for i := 0; i < len(tasks); i++ {
			task := tasks[i]
			GetTaskLoggerComplete(task).Info("Task is about to be killed")
			if task.isRunning {
				task.Task.Close()
			} else {
				GetTaskLoggerComplete(task).Trace("Task is not running yet, pruning directly")
				err := s.remove(task.Id)
				if err != nil {
					s.logger.WithError(err).Errorf("Failed to kill task id: %s", task.Id)
				}
			}
		}
	}

	return nil
}

func (s *TasksScheduler) removeUnresponsiveTasks(ctx context.Context, tasks []TaskRequestInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.logger.Debug("Looking for removing unresponsive tasks")

		for i := 0; i < len(tasks); i++ {
			task := tasks[i]
			GetTaskLoggerComplete(task).Info("Task is about to be removed for being unresponsive or expired")

			err := s.remove(task.Id)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to remove unresponsive task id: %s", task.Id)
			}
		}
	}

	return nil
}

func (s *TasksScheduler) findTasks() ([]TaskRequestInfo, []TaskRequestInfo, []TaskRequestInfo) {
	toStart := make([]TaskRequestInfo, 0)
	expired := make([]TaskRequestInfo, 0)
	unresponsive := make([]TaskRequestInfo, 0)
	multiExecutionCheck := make(map[string]bool)

	for _, taskRequest := range s.Schedule {
		if taskRequest.End != 0 && taskRequest.End+constants.TaskSchedulerHeightToleranceBeforeRemoving <= s.LastHeightSeen {
			s.logger.Tracef("marking task as unresponsive %s", taskRequest.Task.GetId())
			unresponsive = append(unresponsive, taskRequest)
			continue
		}

		if taskRequest.End != 0 && taskRequest.End <= s.LastHeightSeen {
			s.logger.Tracef("marking task as expired %s", taskRequest.Task.GetId())
			expired = append(expired, taskRequest)
			continue
		}

		if ((taskRequest.Start == 0 && taskRequest.End == 0) ||
			(taskRequest.Start != 0 && taskRequest.Start <= s.LastHeightSeen && taskRequest.End == 0) ||
			(taskRequest.Start <= s.LastHeightSeen && taskRequest.End > s.LastHeightSeen)) && !taskRequest.isRunning {

			if taskRequest.Task.GetAllowMultiExecution() {
				multiExecutionCheck[taskRequest.Name] = true
				toStart = append(toStart, taskRequest)
			} else {
				if alreadyPicked := multiExecutionCheck[taskRequest.Name]; !alreadyPicked && len(s.findRunningTasksByName(taskRequest.Name)) == 0 {
					multiExecutionCheck[taskRequest.Name] = true
					toStart = append(toStart, taskRequest)
				} else {
					s.logger.Debugf("trying to start more than 1 task instance when this is not allowed id: %s, name: %s", taskRequest.Id, taskRequest.Name)
				}
			}
			continue
		}
	}
	return toStart, expired, unresponsive
}

func (s *TasksScheduler) findTasksByName(taskName string) []TaskRequestInfo {
	s.logger.Tracef("trying to find tasks by name %s", taskName)
	tasks := make([]TaskRequestInfo, 0)

	for _, taskRequest := range s.Schedule {
		if taskRequest.Name == taskName {
			tasks = append(tasks, taskRequest)
		}
	}
	s.logger.Tracef("found %v tasks with name %s", len(tasks), taskName)
	return tasks
}

func (s *TasksScheduler) findRunningTasksByName(taskName string) []TaskRequestInfo {
	s.logger.Tracef("finding running tasks by name %s", taskName)
	tasks := make([]TaskRequestInfo, 0)

	for _, taskRequest := range s.Schedule {
		if taskRequest.Name == taskName && taskRequest.isRunning {
			tasks = append(tasks, taskRequest)
		}
	}
	s.logger.Tracef("found %v running tasks with name %s", len(tasks), taskName)
	return tasks
}

func (s *TasksScheduler) remove(id string) error {
	s.logger.Tracef("trying to remove task with id %s", id)
	_, present := s.Schedule[id]
	if !present {
		return ErrNotScheduled
	}

	delete(s.Schedule, id)

	return nil
}

func (s *TasksScheduler) persistState() error {
	logger := logging.GetLogger("staterecover").WithField("State", "taskScheduler")
	rawData, err := json.Marshal(s)
	if err != nil {
		return err
	}

	err = s.database.Update(func(txn *badger.Txn) error {
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

	if err := s.database.Sync(); err != nil {
		logger.Error("Failed to set sync")
		return err
	}

	return nil
}

func (s *TasksScheduler) loadState() error {
	logger := logging.GetLogger("staterecover").WithField("State", "taskScheduler")
	if err := s.database.View(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskSchedulerState()
		logger.WithField("Key", string(key)).Debug("Loading state from database")
		rawData, err := utils.GetValue(txn, key)
		if err != nil {
			return err
		}

		err = json.Unmarshal(rawData, s)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// synchronizing db state to disk
	if err := s.database.Sync(); err != nil {
		logger.Error("Failed to set sync")
		return err
	}

	return nil

}

func (s *TasksScheduler) MarshalJSON() ([]byte, error) {

	ws := &innerSequentialSchedule{Schedule: make(map[string]*taskRequestInner)}

	for k, v := range s.Schedule {
		wt, err := s.marshaller.WrapInstance(v.Task)
		if err != nil {
			return []byte{}, err
		}
		ws.Schedule[k] = &taskRequestInner{Id: v.Id, Name: v.Name, Start: v.Start, End: v.End, WrappedTask: wt}
	}

	raw, err := json.Marshal(&ws)
	if err != nil {
		return []byte{}, err
	}

	return raw, nil
}

func (s *TasksScheduler) UnmarshalJSON(raw []byte) error {
	aa := &innerSequentialSchedule{}

	err := json.Unmarshal(raw, aa)
	if err != nil {
		return err
	}

	adminInterface := reflect.TypeOf((*monitorInterfaces.AdminClient)(nil)).Elem()

	s.Schedule = make(map[string]TaskRequestInfo)
	for k, v := range aa.Schedule {
		t, err := s.marshaller.UnwrapInstance(v.WrappedTask)
		if err != nil {
			return err
		}

		// Marshalling service handlers is mostly non-sense, so
		isAdminClient := reflect.TypeOf(t).Implements(adminInterface)
		if isAdminClient {
			adminClient := t.(monitorInterfaces.AdminClient)
			adminClient.SetAdminHandler(s.adminHandler)
		}

		s.Schedule[k] = TaskRequestInfo{Id: v.Id, Name: v.Name, Start: v.Start, End: v.End, Task: t.(tasks.Task)}
	}

	return nil
}
