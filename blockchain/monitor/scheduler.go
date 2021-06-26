package monitor

import (
	"encoding/json"
	"errors"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
)

var (
	ErrOverlappingSchedule = errors.New("overlapping schedule range")
	ErrNothingScheduled    = errors.New("nothing schedule for time")
	ErrNotScheduled        = errors.New("scheduled task not found")
)

type Block struct {
	Start uint64
	End   uint64
	Task  interfaces.Task
}

type innerBlock struct {
	Start       uint64
	End         uint64
	WrappedTask tasks.TaskWrapper
}

func (b *Block) MarshalJSON() ([]byte, error) {

	wrappedTask, err := tasks.WrapTask(b.Task)
	if err != nil {
		return nil, err
	}

	raw, err := json.Marshal(&innerBlock{
		Start:       b.Start,
		End:         b.End,
		WrappedTask: wrappedTask,
	})

	return raw, err
}

func (b *Block) UnmarshalJSON(raw []byte) error {
	aa := &innerBlock{}

	err := json.Unmarshal(raw, aa)
	if err != nil {
		return err
	}

	b.Start = aa.Start
	b.End = aa.End

	b.Task, err = tasks.UnwrapTask(aa.WrappedTask)
	if err != nil {
		return err
	}

	return nil
}

type SequentialSchedule struct {
	Ranges map[string]*Block
}

func NewSequentialSchedule() *SequentialSchedule {
	return &SequentialSchedule{Ranges: make(map[string]*Block)}
}

func (s *SequentialSchedule) Schedule(start uint64, end uint64, thing interfaces.Task) (uuid.UUID, error) {

	for _, block := range s.Ranges {
		if start <= block.End && block.Start <= end {
			return nil, ErrOverlappingSchedule
		}
	}

	id := uuid.NewRandom()

	s.Ranges[id.String()] = &Block{Start: start, End: end, Task: thing}

	return id, nil
}

func (s *SequentialSchedule) Purge() {
	for taskID := range s.Ranges {
		delete(s.Ranges, taskID)
	}
}

func (s *SequentialSchedule) PurgePrior(now uint64) {
	for taskID, block := range s.Ranges {
		if block.Start <= now && block.End <= now {
			delete(s.Ranges, taskID)
		}
	}
}

func (s *SequentialSchedule) Find(now uint64) (uuid.UUID, error) {

	for taskId, block := range s.Ranges {
		if block.Start <= now && block.End >= now {
			return uuid.Parse(taskId), nil
		}
	}
	return nil, ErrNothingScheduled
}

func (s *SequentialSchedule) Retrieve(taskId uuid.UUID) (interfaces.Task, error) {
	block, present := s.Ranges[taskId.String()]
	if !present {
		return nil, ErrNotScheduled
	}

	return block.Task, nil
}

func (s *SequentialSchedule) Length() int {
	return len(s.Ranges)
}

func (s *SequentialSchedule) Remove(taskId uuid.UUID) error {
	id := taskId.String()

	_, present := s.Ranges[id]
	if !present {
		return ErrNotScheduled
	}

	delete(s.Ranges, id)

	return nil
}

func (s *SequentialSchedule) Status(logger *logrus.Entry) {
	for id, block := range s.Ranges {
		str, err := block.MarshalJSON()
		if err != nil {
			logger.Errorf("id:%v unable to marshal block: %v", id, err)
		}
		logger.Infof("id:%v block:%+v", id, string(str))
	}
}
