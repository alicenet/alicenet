package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"reflect"
)

var (
	ErrOverlappingSchedule = errors.New("overlapping schedule range")
	ErrNothingScheduled    = errors.New("nothing schedule for time")
	ErrNotScheduled        = errors.New("scheduled task not found")
)

type Block struct {
	Start     uint64           `json:"start"`
	End       uint64           `json:"end"`
	Task      interfaces.ITask `json:"-"`
	IsRunning bool             `json:"isRunning"`
}

type innerBlock struct {
	Start       uint64
	End         uint64
	WrappedTask *objects.InstanceWrapper
}

type Scheduler struct {
	Ranges           map[string]*Block       `json:"ranges"`
	eth              interfaces.Ethereum     `json:"-"`
	adminHandler     interfaces.AdminHandler `json:"-"`
	marshaller       *objects.TypeRegistry   `json:"-"`
	cancelChan       chan bool               `json:"-"`
	currentBlockChan <-chan uint64           `json:"-"`
	tasksChan        <-chan TaskToSchedule   `json:"-"`
	logger           *logrus.Entry           `json:"-"`
	//TODO: add db for recovery (Should be a separate db or a shared one with the monitor???)
}

type TaskToSchedule struct {
	Start uint64
	End   uint64
	Task  interfaces.ITask
}

type innerSequentialSchedule struct {
	Ranges map[string]*innerBlock
}

func NewTasksScheduler(adminHandler interfaces.AdminHandler, eth interfaces.Ethereum, currentBlockChan <-chan uint64, tasksChan <-chan TaskToSchedule) *Scheduler {
	tr := &objects.TypeRegistry{}

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

	s := &Scheduler{
		Ranges:           make(map[string]*Block),
		eth:              eth,
		adminHandler:     adminHandler,
		marshaller:       tr,
		cancelChan:       make(chan bool, 1),
		currentBlockChan: currentBlockChan,
		tasksChan:        tasksChan,
	}

	return s
}

func (s *Scheduler) Start() error {
	//TODO: load the last saved state for recovery

	go s.eventLoop()
	return nil
}

func (s *Scheduler) eventLoop() {
	ctx, cf := context.WithCancel(context.Background())

	for {
		select {
		case <-s.cancelChan:
			s.logger.Warnf("Received cancel request for event loop.")
			//TODO: save the state for recovery
			s.purge()
			cf()
			return
		case task := <-s.tasksChan:
			go s.schedule(task.Start, task.End, task.Task)
		case currentBlock := <-s.currentBlockChan:
			err := s.processBlock(ctx, currentBlock)
			s.logger.WithError(err).Errorf("Failed to processBlock %d", currentBlock)
		}
	}
}

func (s *Scheduler) Close() {
	s.cancelChan <- true
}

func (s *Scheduler) processBlock(ctx context.Context, currentBlock uint64) error {
	select {
	case <-ctx.Done():
		s.logger.Debugf("context ended with error: %s", ctx.Err())
		s.purge()
		return nil
	default:
		s.logger.Debug("Looking for scheduled task")
		uuid, err := s.find(currentBlock)
		if err == nil {
			isRunning, err := s.isRunning(uuid)
			if err != nil {
				s.logger.WithError(err).Error("Failed to execute isRunning")
			}

			if !isRunning {
				task, err := s.retrieve(uuid)
				if err != nil {
					s.logger.WithError(err).Error("Failed to execute retrieve")
				}

				taskName, _ := objects.GetNameType(task)
				s.logger.Infof("Task name: %s", taskName)

				//onFinishCB := func() {
				//	err := monitorState.Schedule.SetRunning(uuid, false)
				//	if err != nil {
				//		logEntry.WithError(err).Error("Failed to set task to not running")
				//	}
				//	err = monitorState.Schedule.Remove(uuid)
				//	if err != nil {
				//		logEntry.WithError(err).Error("Failed to remove task from schedule")
				//	}
				//}
				//
				//err = StartTask(log, wg, eth, task, persistMonitorCB, onFinishCB)
				//if err != nil {
				//	return err
				//}
				//err = monitorState.Schedule.SetRunning(uuid, true)
				//if err != nil {
				//	return err
				//}

				//TODO: spawn go routine executing a task using a new Task Manager
				//log := s.logger.WithFields(logrus.Fields{
				//	"TaskID":   uuid.String(),
				//	"TaskName": taskName})
			}

		} else if err == objects.ErrNothingScheduled {
			s.logger.Debug("No tasks scheduled")
		} else {
			s.logger.Warnf("Error retrieving scheduled task: %v", err)
		}
	}

	return nil
}

func (s *Scheduler) schedule(start uint64, end uint64, task interfaces.ITask) uuid.UUID {
	id := uuid.NewRandom()
	s.Ranges[id.String()] = &Block{Start: start, End: end, Task: task}
	return id
}

func (s *Scheduler) purge() {
	for taskID := range s.Ranges {
		delete(s.Ranges, taskID)
	}
}

func (s *Scheduler) purgePrior(now uint64) {
	for taskID, block := range s.Ranges {
		if block.Start <= now && block.End <= now {
			delete(s.Ranges, taskID)
		}
	}
}

func (s *Scheduler) setRunning(taskId uuid.UUID, running bool) error {
	block, present := s.Ranges[taskId.String()]
	if !present {
		return ErrNotScheduled
	}

	block.IsRunning = running
	return nil
}

func (s *Scheduler) isRunning(taskId uuid.UUID) (bool, error) {
	block, present := s.Ranges[taskId.String()]
	if !present {
		return false, ErrNotScheduled
	}

	return block.IsRunning, nil
}

func (s *Scheduler) find(now uint64) (uuid.UUID, error) {

	for taskId, block := range s.Ranges {
		if block.Start <= now && block.End > now {
			return uuid.Parse(taskId), nil
		}
	}
	return nil, ErrNothingScheduled
}

func (s *Scheduler) retrieve(taskId uuid.UUID) (interfaces.ITask, error) {
	block, present := s.Ranges[taskId.String()]
	if !present {
		return nil, ErrNotScheduled
	}

	return block.Task, nil
}

func (s *Scheduler) length() int {
	return len(s.Ranges)
}

func (s *Scheduler) remove(taskId uuid.UUID) error {
	id := taskId.String()

	_, present := s.Ranges[id]
	if !present {
		return ErrNotScheduled
	}

	delete(s.Ranges, id)

	return nil
}

func (s *Scheduler) status(logger *logrus.Entry) {
	for _, block := range s.Ranges {
		name, _ := objects.GetNameType(block.Task)
		logger.Infof("Schedule %p Task %v Range %v and %v", s, name, block.Start, block.End)
	}
}

func (s *Scheduler) MarshalJSON() ([]byte, error) {

	ws := &innerSequentialSchedule{Ranges: make(map[string]*innerBlock)}

	for k, v := range s.Ranges {
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

func (s *Scheduler) UnmarshalJSON(raw []byte) error {
	aa := &innerSequentialSchedule{}

	err := json.Unmarshal(raw, aa)
	if err != nil {
		return err
	}

	adminInterface := reflect.TypeOf((*interfaces.AdminClient)(nil)).Elem()

	s.Ranges = make(map[string]*Block)
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

		s.Ranges[k] = &Block{Start: v.Start, End: v.End, Task: t.(interfaces.ITask)}
	}

	return nil
}
