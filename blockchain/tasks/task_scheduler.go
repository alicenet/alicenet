package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/objects"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/other_tasks"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

var (
	ErrNothingScheduled = errors.New("nothing schedule for time")
	ErrNotScheduled     = errors.New("scheduled task not found")
	ErrWrongParams      = errors.New("wrong start/end height for the task")
	ErrTaskExpired      = errors.New("the task is already expired")
)

const (
	heightToleranceBeforeRemoving uint64 = 50
)

type TaskRequest struct {
	Id        string           `json:"id"`
	Start     uint64           `json:"start"`
	End       uint64           `json:"end"`
	IsRunning bool             `json:"is_running"`
	Task      interfaces.ITask `json:"-"`
}

type TaskResponse struct {
	Id  string
	Err error
}

type innerBlock struct {
	Id          string
	Start       uint64
	End         uint64
	IsRunning   bool
	WrappedTask *objects.InstanceWrapper
}

type TasksScheduler struct {
	Schedule               map[string]*TaskRequest `json:"schedule"`
	LastHeightSeen         uint64                  `json:"last_height_seen"`
	database               *db.Database            `json:"-"`
	adminHandler           interfaces.AdminHandler `json:"-"`
	marshaller             *objects.TypeRegistry   `json:"-"`
	cancelChan             chan bool               `json:"-"`
	lastFinalizedBlockChan <-chan uint64           `json:"-"`
	taskRequestChan        <-chan interfaces.ITask `json:"-"`
	taskResponseChan       chan TaskResponse       `json:"-"`
	taskKillChan           <-chan string           `json:"-"`
	logger                 *logrus.Entry           `json:"-"`
}

type innerSequentialSchedule struct {
	Schedule map[string]*innerBlock
}

func NewTasksScheduler(database *db.Database, adminHandler interfaces.AdminHandler, lastFinalizedBlockChan <-chan uint64, taskRequestChan <-chan interfaces.ITask, taskKillChan <-chan string) *TasksScheduler {
	tr := &objects.TypeRegistry{}

	//TODO: refactor, import cycle not allowed. Move dkgtasks inside tasks package???
	tr.RegisterInstanceType(&dkgtasks.CompletionTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeShareDistributionTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeMissingShareDistributionTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeMissingKeySharesTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeMissingGPKjTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeGPKjTask{})
	tr.RegisterInstanceType(&dkgtasks.GPKjSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.KeyShareSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.MPKSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.PlaceHolder{})
	tr.RegisterInstanceType(&dkgtasks.RegisterTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeMissingRegistrationTask{})
	tr.RegisterInstanceType(&dkgtasks.ShareDistributionTask{})
	tr.RegisterInstanceType(&other_tasks.SnapshotTask{})

	s := &TasksScheduler{
		Schedule:               make(map[string]*TaskRequest),
		database:               database,
		adminHandler:           adminHandler,
		marshaller:             tr,
		cancelChan:             make(chan bool, 1),
		lastFinalizedBlockChan: lastFinalizedBlockChan,
		taskRequestChan:        taskRequestChan,
		taskResponseChan:       make(chan TaskResponse, 100),
		taskKillChan:           taskKillChan,
	}

	logger := logging.GetLogger("tasks_scheduler").WithField("Schedule", s.Schedule)
	s.logger = logger

	return s
}

func (s *TasksScheduler) Start() error {
	err := s.LoadState()
	if err != nil {
		s.logger.Warnf("could not find previous State: %v", err)
		if err != badger.ErrKeyNotFound {
			return err
		}
	}

	s.logger.Info(strings.Repeat("-", 80))
	s.logger.Infof("Current Tasks: %d", len(s.Schedule))
	for id, task := range s.Schedule {
		taskName, _ := objects.GetNameType(task)
		s.logger.Infof("...ID: %s Name: %s Between: %d and %d isRunning %t", id, taskName, task.Start, task.End, task.IsRunning)
	}
	s.logger.Info(strings.Repeat("-", 80))

	go s.eventLoop()
	return nil
}

func (s *TasksScheduler) eventLoop() {
	ctx, cf := context.WithCancel(context.Background())

	for {
		select {
		case <-s.cancelChan:
			s.logger.Warnf("Received cancel request for event loop.")
			s.purge()
			cf()
			return
		case block := <-s.lastFinalizedBlockChan:
			s.LastHeightSeen = block
		case taskRequest := <-s.taskRequestChan:
			err := s.schedule(ctx, taskRequest)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to schedule task request %d", s.LastHeightSeen)
			}
			err = s.PersistState()
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to schedule task request %d", s.LastHeightSeen)
			}
		case taskResponse := <-s.taskResponseChan:
			err := s.processTaskResponse(ctx, taskResponse)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to processTaskResponse %v", taskResponse)
			}
			err = s.PersistState()
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to schedule task request %d", s.LastHeightSeen)
			}
		case taskToKill := <-s.taskKillChan:
			err := s.killTaskByName(ctx, taskToKill)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to killTaskByName %v", taskToKill)
			}
		default:
			toStart, expired, unresponsive := s.findTasks()
			err := s.startTasks(ctx, toStart)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to processBlock %d", s.LastHeightSeen)
			}

			err = s.killTasks(ctx, expired)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to killExpiredTasks %d", s.LastHeightSeen)
			}

			err = s.removeUnresponsiveTasks(ctx, unresponsive)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to removeUnresponsiveTasks %d", s.LastHeightSeen)
			}
			err = s.PersistState()
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to schedule task request %d", s.LastHeightSeen)
			}
		}
	}
}

func (s *TasksScheduler) Close() {
	close(s.taskResponseChan)
	s.cancelChan <- true
}

func (s *TasksScheduler) schedule(ctx context.Context, task interfaces.ITask) error {
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
		s.Schedule[id.String()] = &TaskRequest{Id: id.String(), Start: start, End: end, Task: task}
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

func (s *TasksScheduler) startTasks(ctx context.Context, tasks []*TaskRequest) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.logger.Debug("Looking for starting tasks")
		for i := 0; i < len(tasks); i++ {
			task := tasks[i]
			taskName, _ := objects.GetNameType(task)
			s.logger.Infof("Task name: %s is about to start", taskName)
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

func (s *TasksScheduler) killTasks(ctx context.Context, tasks []*TaskRequest) error {
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

func (s *TasksScheduler) removeUnresponsiveTasks(ctx context.Context, tasks []*TaskRequest) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.logger.Debug("Looking for removing unresponsive tasks")

		for i := 0; i < len(tasks); i++ {
			task := tasks[i]
			taskName, _ := objects.GetNameType(task)
			s.logger.Infof("Task name: %s is about to be removed", taskName)

			err := s.remove(task.Id)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to remove unresponsive task id: %s", task.Id)
			}
		}

	}

	return nil
}

func (s *TasksScheduler) purge() {
	s.Schedule = make(map[string]*TaskRequest)
}

func (s *TasksScheduler) findTasks() ([]*TaskRequest, []*TaskRequest, []*TaskRequest) {
	toStart := make([]*TaskRequest, 0)
	expired := make([]*TaskRequest, 0)
	unresponsive := make([]*TaskRequest, 0)

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

func (s *TasksScheduler) findTasksByName(taskName string) []*TaskRequest {
	tasks := make([]*TaskRequest, 0)

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

func (s *TasksScheduler) PersistState() error {
	rawData, err := json.Marshal(s)
	if err != nil {
		return err
	}

	err = s.database.Update(func(txn *badger.Txn) error {
		keyLabel := fmt.Sprintf("%x", getStateKey())
		s.logger.WithField("Key", keyLabel).Infof("Saving state")
		if err := utils.SetValue(txn, getStateKey(), rawData); err != nil {
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

func (s *TasksScheduler) LoadState() error {

	if err := s.database.View(func(txn *badger.Txn) error {
		keyLabel := fmt.Sprintf("%x", getStateKey())
		s.logger.WithField("Key", keyLabel).Infof("Looking up state")
		rawData, err := utils.GetValue(txn, getStateKey())
		if err != nil {
			return err
		}
		// TODO: Cleanup loaded obj, this is a memory / storage leak
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

func getStateKey() []byte {
	return []byte("schedulerStateKey")
}

func (s *TasksScheduler) MarshalJSON() ([]byte, error) {

	ws := &innerSequentialSchedule{Schedule: make(map[string]*innerBlock)}

	for k, v := range s.Schedule {
		wt, err := s.marshaller.WrapInstance(v.Task)
		if err != nil {
			return []byte{}, err
		}
		ws.Schedule[k] = &innerBlock{Id: v.Id, Start: v.Start, End: v.End, IsRunning: v.IsRunning, WrappedTask: wt}
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

	adminInterface := reflect.TypeOf((*interfaces.AdminClient)(nil)).Elem()

	s.Schedule = make(map[string]*TaskRequest)
	for k, v := range aa.Schedule {
		t, err := s.marshaller.UnwrapInstance(v.WrappedTask)
		if err != nil {
			return err
		}

		// Marshalling service handlers is mostly non-sense, so
		isAdminClient := reflect.TypeOf(t).Implements(adminInterface)
		if isAdminClient {
			adminClient := t.(interfaces.AdminClient)
			adminClient.SetAdminHandler(s.adminHandler)
		}

		s.Schedule[k] = &TaskRequest{Id: v.Id, Start: v.Start, End: v.End, IsRunning: v.IsRunning, Task: t.(interfaces.ITask)}
	}

	return nil
}
