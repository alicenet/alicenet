package monitor

import (
	"errors"

	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/pborman/uuid"
)

var (
	ErrOverlappingSchedule = errors.New("Overlapping schedule range")
	ErrNothingScheduled    = errors.New("Nothing schedule for time")
	ErrNotScheduled        = errors.New("Scheduled task not found")
)

type Block struct {
	Start uint64
	End   uint64
	Task  tasks.Task
}

type SequentialSchedule struct {
	Ranges map[string]Block
}

func NewSequentialSchedule() *SequentialSchedule {
	return &SequentialSchedule{Ranges: make(map[string]Block)}
}

func (s *SequentialSchedule) Schedule(start uint64, end uint64, thing tasks.Task) (uuid.UUID, error) {

	for _, block := range s.Ranges {
		if start <= block.End && block.Start <= end {
			return nil, ErrOverlappingSchedule
		}
	}

	id := uuid.NewRandom()

	s.Ranges[id.String()] = Block{Start: start, End: end, Task: thing}

	return id, nil
}

func (s *SequentialSchedule) Find(now uint64) (uuid.UUID, error) {

	for taskId, block := range s.Ranges {
		if block.Start <= now && block.End >= now {
			return uuid.Parse(taskId), nil
		}
	}
	return nil, ErrNothingScheduled
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
