package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/objects"
	"github.com/MadBase/MadNet/logging"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"reflect"
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
	IsRunning bool             `json:"isRunning"`
	Task      interfaces.ITask `json:"-"`
}

type TaskResponse struct {
	Id  string
	Err error
}

type innerBlock struct {
	Start       uint64
	End         uint64
	WrappedTask *objects.InstanceWrapper
}

type TasksScheduler struct {
	Schedule               map[string]*TaskRequest `json:"schedule"`
	LastHeightSeen         uint64                  `json:"last_height_seen"`
	eth                    interfaces.Ethereum     `json:"-"`
	adminHandler           interfaces.AdminHandler `json:"-"`
	marshaller             *objects.TypeRegistry   `json:"-"`
	cancelChan             chan bool               `json:"-"`
	lastFinalizedBlockChan <-chan uint64           `json:"-"`
	taskRequestChan        <-chan interfaces.ITask `json:"-"`
	taskResponseChan       chan TaskResponse       `json:"-"`
	logger                 *logrus.Entry           `json:"-"`
	//TODO: add db for recovery (Should be a separate db or a shared one with the monitor???)
}

type innerSequentialSchedule struct {
	Ranges map[string]*innerBlock
}

func NewTasksScheduler(adminHandler interfaces.AdminHandler, eth interfaces.Ethereum, lastFinalizedBlockChan <-chan uint64, taskRequestChan <-chan interfaces.ITask) *TasksScheduler {
	tr := &objects.TypeRegistry{}

	//TODO: refactor, import cycle not allowed. Move dkgtasks inside tasks package???
	tr.RegisterInstanceType(&dkgtasks.CompletionTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeShareDistributionTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeMissingShareDistributionTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeMissingKeySharesTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeMissingGPKjTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeGPKjTask{})
	tr.RegisterInstanceType(&dkgtasks.GPKjSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.KeyshareSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.MPKSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.PlaceHolder{})
	tr.RegisterInstanceType(&dkgtasks.RegisterTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeMissingRegistrationTask{})
	tr.RegisterInstanceType(&dkgtasks.ShareDistributionTask{})
	tr.RegisterInstanceType(&SnapshotTask{})

	s := &TasksScheduler{
		Schedule:               make(map[string]*TaskRequest),
		eth:                    eth,
		adminHandler:           adminHandler,
		marshaller:             tr,
		cancelChan:             make(chan bool, 1),
		lastFinalizedBlockChan: lastFinalizedBlockChan,
		taskRequestChan:        taskRequestChan,
		taskResponseChan:       make(chan TaskResponse, 100),
	}

	logger := logging.GetLogger("tasks_scheduler").WithField("Schedule", s.Schedule)
	s.logger = logger

	return s
}

func (s *TasksScheduler) Start() error {
	//TODO: load the last saved state for recovery

	go s.eventLoop()
	return nil
}

func (s *TasksScheduler) eventLoop() {
	ctx, cf := context.WithCancel(context.Background())

	for {
		select {
		case <-s.cancelChan:
			s.logger.Warnf("Received cancel request for event loop.")
			//TODO: save the state for recovery
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
		case taskResponse := <-s.taskResponseChan:
			//TODO: process task response
		default:
			toStart, expired, unresponsive := s.findTasks()
			err := s.startTasks(ctx, toStart)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to processBlock %d", s.LastHeightSeen)
			}

			err = s.killExpiredTasks(ctx, expired)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to killExpiredTasks %d", s.LastHeightSeen)
			}

			err = s.removeUnresponsiveTasks(ctx, unresponsive)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to removeUnresponsiveTasks %d", s.LastHeightSeen)
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

		if start >= end {
			return ErrWrongParams
		}

		if end <= s.LastHeightSeen {
			return ErrTaskExpired
		}

		id := uuid.NewRandom()
		s.Schedule[id.String()] = &TaskRequest{Id: id.String(), Start: start, End: end, Task: task}
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

			//TODO: run the task in a go routine
			task.IsRunning = true
		}

	}

	return nil
}

func (s *TasksScheduler) killExpiredTasks(ctx context.Context, tasks []*TaskRequest) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.logger.Debug("Looking for killing expired tasks")

		for i := 0; i < len(tasks); i++ {
			task := tasks[i]
			taskName, _ := objects.GetNameType(task)
			s.logger.Infof("Task name: %s is about to be killed", taskName)
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
	for _, request := range s.Schedule {
		request.Task.GetExecutionData().Close()
	}
	s.Schedule = make(map[string]*TaskRequest)
}

func (s *TasksScheduler) findTasks() ([]*TaskRequest, []*TaskRequest, []*TaskRequest) {
	toStart := make([]*TaskRequest, 0)
	expired := make([]*TaskRequest, 0)
	unresponsive := make([]*TaskRequest, 0)

	for _, taskRequest := range s.Schedule {
		if taskRequest.End+heightToleranceBeforeRemoving <= s.LastHeightSeen {
			unresponsive = append(unresponsive, taskRequest)
			continue
		}

		if taskRequest.End <= s.LastHeightSeen {
			expired = append(expired, taskRequest)
			continue
		}

		if taskRequest.Start <= s.LastHeightSeen && taskRequest.End > s.LastHeightSeen && !taskRequest.IsRunning {
			toStart = append(toStart, taskRequest)
			continue
		}
	}
	return toStart, expired, unresponsive
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

func (s *TasksScheduler) MarshalJSON() ([]byte, error) {

	ws := &innerSequentialSchedule{Ranges: make(map[string]*innerBlock)}

	for k, v := range s.Schedule {
		wt, err := s.marshaller.WrapInstance(v.Task)
		if err != nil {
			return []byte{}, err
		}
		ws.Ranges[k] = &innerBlock{Start: v.Start, End: v.End, WrappedTask: wt}
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
	for k, v := range aa.Ranges {
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

		s.Schedule[k] = &TaskRequest{Start: v.Start, End: v.End, Task: t.(interfaces.ITask)}
	}

	return nil
}
