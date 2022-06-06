package executor

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/MadBase/MadNet/constants/dbprefix"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	executorInterfaces "github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/marshaller"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/snapshots"
	monitorInterfaces "github.com/MadBase/MadNet/blockchain/monitor/interfaces"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
)

var (
	ErrNotScheduled = errors.New("scheduled task not found")
	ErrWrongParams  = errors.New("wrong start/end height for the task")
	ErrTaskExpired  = errors.New("the task is already expired")
)

const (
	heightToleranceBeforeRemoving uint64 = 50
)

type TaskRequestInfo struct {
	Id        string                   `json:"id"`
	Start     uint64                   `json:"start"`
	End       uint64                   `json:"end"`
	IsRunning bool                     `json:"-"`
	Task      executorInterfaces.ITask `json:"-"`
}

type TaskResponse struct {
	Id  string
	Err error
}

type innerBlock struct {
	Id          string
	Start       uint64
	End         uint64
	WrappedTask *marshaller.InstanceWrapper
}

type TasksScheduler struct {
	Schedule         map[string]TaskRequestInfo      `json:"schedule"`
	LastHeightSeen   uint64                          `json:"last_height_seen"`
	eth              ethereum.Network                `json:"-"`
	database         *db.Database                    `json:"-"`
	adminHandler     monitorInterfaces.IAdminHandler `json:"-"`
	marshaller       *marshaller.TypeRegistry        `json:"-"`
	cancelChan       chan bool                       `json:"-"`
	taskRequestChan  <-chan executorInterfaces.ITask `json:"-"`
	taskResponseChan *taskResponseChan               `json:"-"`
	taskKillChan     <-chan string                   `json:"-"`
	logger           *logrus.Entry                   `json:"-"`
}

type taskResponseChan struct {
	writeOnce sync.Once
	trChan    chan TaskResponse
	isClosed  bool
}

type ITaskResponseChan interface {
	Add(TaskResponse)
}

func (tr *taskResponseChan) close() {
	tr.writeOnce.Do(func() {
		tr.isClosed = true
		close(tr.trChan)
	})
}

func (tr *taskResponseChan) Add(taskResponse TaskResponse) {
	if !tr.isClosed {
		tr.trChan <- taskResponse
	}
}

type innerSequentialSchedule struct {
	Schedule map[string]*innerBlock
}

func NewTasksScheduler(database *db.Database, eth ethereum.Network, adminHandler monitorInterfaces.IAdminHandler, taskRequestChan <-chan executorInterfaces.ITask, taskKillChan <-chan string) *TasksScheduler {
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

	s := &TasksScheduler{
		Schedule:         make(map[string]TaskRequestInfo),
		database:         database,
		eth:              eth,
		adminHandler:     adminHandler,
		marshaller:       tr,
		cancelChan:       make(chan bool, 1),
		taskRequestChan:  taskRequestChan,
		taskResponseChan: &taskResponseChan{trChan: make(chan TaskResponse, 100)},
		taskKillChan:     taskKillChan,
	}

	logger := logging.GetLogger("tasks_scheduler").WithField("Schedule", s.Schedule)
	s.logger = logger

	return s
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
		s.logger.Infof("...ID: %s Name: %s Between: %d and %d", id, task.Task.GetExecutionData().GetName(), task.Start, task.End)
	}
	s.logger.Info(strings.Repeat("-", 80))

	go s.eventLoop()
	return nil
}

func (s *TasksScheduler) Close() {
	s.cancelChan <- true
}

func (s *TasksScheduler) eventLoop() {
	ctx, cf := context.WithCancel(context.Background())
	processingTime := time.After(3 * time.Second)

	for {
		select {
		case <-s.cancelChan:
			s.logger.Warnf("Received cancel request for event loop.")
			s.purge()
			cf()
			s.taskResponseChan.close()
			return
		case taskRequest := <-s.taskRequestChan:
			err := s.schedule(ctx, taskRequest)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to schedule task request %d", s.LastHeightSeen)
			}
			err = s.persistState()
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to persist state %d on task request", s.LastHeightSeen)
			}
		case taskResponse := <-s.taskResponseChan.trChan:
			err := s.processTaskResponse(ctx, taskResponse)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to processTaskResponse %v", taskResponse)
			}
			err = s.persistState()
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to persist state %d on task response", s.LastHeightSeen)
			}
		case taskToKill := <-s.taskKillChan:
			err := s.killTaskByName(ctx, taskToKill)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to killTaskByName %v", taskToKill)
			}
		case <-processingTime:
			networkCtx, networkCf := context.WithTimeout(ctx, 1*time.Second)
			height, err := s.eth.GetCurrentHeight(networkCtx)
			networkCf()
			if err != nil {
				s.logger.WithError(err).Debug("Failed to retrieve the latest height from eth node")
				continue
			}
			s.LastHeightSeen = height

			toStart, expired, unresponsive := s.findTasks()
			err = s.startTasks(ctx, toStart)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to startTasks %d", s.LastHeightSeen)
			}

			err = s.killTasks(ctx, expired)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to killExpiredTasks %d", s.LastHeightSeen)
			}

			err = s.removeUnresponsiveTasks(ctx, unresponsive)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to removeUnresponsiveTasks %d", s.LastHeightSeen)
			}
			err = s.persistState()
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to persist state %d", s.LastHeightSeen)
			}
		}
	}
}

func (s *TasksScheduler) schedule(ctx context.Context, task executorInterfaces.ITask) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		start := task.GetExecutionData().GetStart()
		end := task.GetExecutionData().GetEnd()

		if start != 0 && end != 0 && start >= end {
			return ErrWrongParams
		}

		if end != 0 && end <= s.LastHeightSeen {
			return ErrTaskExpired
		}

		id := uuid.NewRandom()
		s.Schedule[id.String()] = TaskRequestInfo{Id: id.String(), Start: start, End: end, Task: task}
	}
	return nil
}

func (s *TasksScheduler) processTaskResponse(ctx context.Context, taskResponse TaskResponse) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if taskResponse.Err != nil {
			s.logger.Debugf("Task id: %s executed with error: %v", taskResponse.Id, taskResponse.Err)
		} else {
			s.logger.Infof("Task id: %s executed with succesfully", taskResponse.Id)
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
			s.logger.Infof("Task id: %s name: %s is about to start", task.Id, task.Task.GetExecutionData().GetName())
			taskCtx, taskCancelFunc := context.WithCancel(ctx)
			task.Task.GetExecutionData().SetContext(taskCtx, taskCancelFunc)
			task.Task.GetExecutionData().SetId(task.Id)

			//TODO: run the task in a go routine
			task.IsRunning = true
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
			s.logger.Infof("Task name: %s and id: %s is about to be killed", task.Task.GetExecutionData().GetName(), task.Id)
			task.Task.GetExecutionData().Close()
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
			s.logger.Infof("Task name: %s is about to be removed", task.Task.GetExecutionData().GetName())

			err := s.remove(task.Id)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to remove unresponsive task id: %s", task.Id)
			}
		}

	}

	return nil
}

func (s *TasksScheduler) purge() {
	s.Schedule = make(map[string]TaskRequestInfo)
}

func (s *TasksScheduler) findTasks() ([]TaskRequestInfo, []TaskRequestInfo, []TaskRequestInfo) {
	toStart := make([]TaskRequestInfo, 0)
	expired := make([]TaskRequestInfo, 0)
	unresponsive := make([]TaskRequestInfo, 0)

	for _, taskRequest := range s.Schedule {
		if taskRequest.End != 0 && taskRequest.End+heightToleranceBeforeRemoving <= s.LastHeightSeen {
			unresponsive = append(unresponsive, taskRequest)
			continue
		}

		if taskRequest.End != 0 && taskRequest.End <= s.LastHeightSeen {
			expired = append(expired, taskRequest)
			continue
		}

		if ((taskRequest.Start == 0 && taskRequest.End == 0) ||
			(taskRequest.Start != 0 && taskRequest.Start <= s.LastHeightSeen && taskRequest.End == 0) ||
			(taskRequest.Start <= s.LastHeightSeen && taskRequest.End > s.LastHeightSeen)) && !taskRequest.IsRunning {
			toStart = append(toStart, taskRequest)
			continue
		}
	}
	return toStart, expired, unresponsive
}

func (s *TasksScheduler) findTasksByName(taskName string) []TaskRequestInfo {
	tasks := make([]TaskRequestInfo, 0)

	for _, taskRequest := range s.Schedule {
		if taskRequest.Task.GetExecutionData().GetName() == taskName {
			tasks = append(tasks, taskRequest)
		}
	}
	return tasks
}

func (s *TasksScheduler) length() int {
	return len(s.Schedule)
}

func (s *TasksScheduler) remove(id string) error {
	_, present := s.Schedule[id]
	if !present {
		return ErrNotScheduled
	}

	delete(s.Schedule, id)

	return nil
}

func (s *TasksScheduler) persistState() error {
	rawData, err := json.Marshal(s)
	if err != nil {
		return err
	}

	err = s.database.Update(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskSchedulerState()
		s.logger.WithField("Key", string(key)).Infof("Saving state")
		if err := utils.SetValue(txn, key, rawData); err != nil {
			s.logger.Error("Failed to set Value")
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := s.database.Sync(); err != nil {
		s.logger.Error("Failed to set sync")
		return err
	}

	return nil
}

func (s *TasksScheduler) loadState() error {

	if err := s.database.View(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskSchedulerState()
		s.logger.WithField("Key", string(key)).Infof("Looking up state")
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

	return nil

}

func (s *TasksScheduler) MarshalJSON() ([]byte, error) {

	ws := &innerSequentialSchedule{Schedule: make(map[string]*innerBlock)}

	for k, v := range s.Schedule {
		wt, err := s.marshaller.WrapInstance(v.Task)
		if err != nil {
			return []byte{}, err
		}
		ws.Schedule[k] = &innerBlock{Id: v.Id, Start: v.Start, End: v.End, WrappedTask: wt}
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

	adminInterface := reflect.TypeOf((*monitorInterfaces.IAdminClient)(nil)).Elem()

	s.Schedule = make(map[string]TaskRequestInfo)
	for k, v := range aa.Schedule {
		t, err := s.marshaller.UnwrapInstance(v.WrappedTask)
		if err != nil {
			return err
		}

		// Marshalling service handlers is mostly non-sense, so
		isAdminClient := reflect.TypeOf(t).Implements(adminInterface)
		if isAdminClient {
			adminClient := t.(monitorInterfaces.IAdminClient)
			adminClient.SetAdminHandler(s.adminHandler)
		}

		s.Schedule[k] = TaskRequestInfo{Id: v.Id, Start: v.Start, End: v.End, Task: t.(executorInterfaces.ITask)}
	}

	return nil
}
