package objects

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	ErrOverlappingSchedule = errors.New("overlapping schedule range")
	ErrNothingScheduled    = errors.New("nothing schedule for time")
	ErrNotScheduled        = errors.New("scheduled task not found")
)

type Block struct {
	Start     uint64          `json:"start"`
	End       uint64          `json:"end"`
	Task      interfaces.Task `json:"-"`
	IsRunning bool            `json:"isRunning"`
}

type innerBlock struct {
	Start       uint64
	End         uint64
	WrappedTask *InstanceWrapper
}

type SequentialSchedule struct {
	Ranges       map[string]*Block       `json:"ranges"`
	adminHandler interfaces.AdminHandler `json:"-"`
	marshaller   *TypeRegistry           `json:"-"`
}

type innerSequentialSchedule struct {
	Ranges map[string]*innerBlock
}

func NewSequentialSchedule(m *TypeRegistry, adminHandler interfaces.AdminHandler) *SequentialSchedule {
	return &SequentialSchedule{Ranges: make(map[string]*Block), adminHandler: adminHandler, marshaller: m}
}

func (s *SequentialSchedule) Initialize(typeRegistry *TypeRegistry, adminHandler interfaces.AdminHandler) {
	s.adminHandler = adminHandler
	s.marshaller = typeRegistry
}

func (s *SequentialSchedule) Schedule(start uint64, end uint64, thing interfaces.Task) uuid.UUID {
	id := uuid.New()
	s.Ranges[id.String()] = &Block{Start: start, End: end, Task: thing}
	return id
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

func (s *SequentialSchedule) SetRunning(taskId uuid.UUID, running bool) error {
	block, present := s.Ranges[taskId.String()]
	if !present {
		return ErrNotScheduled
	}

	block.IsRunning = running
	return nil
}

func (s *SequentialSchedule) IsRunning(taskId uuid.UUID) (bool, error) {
	block, present := s.Ranges[taskId.String()]
	if !present {
		return false, ErrNotScheduled
	}

	return block.IsRunning, nil
}

func (s *SequentialSchedule) Find(now uint64) (uuid.UUID, error) {

	for taskId, block := range s.Ranges {
		if block.Start <= now && block.End > now {
			return uuid.Parse(taskId)
		}
	}
	return uuid.UUID{}, ErrNothingScheduled
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
	for _, block := range s.Ranges {
		name, _ := GetNameType(block.Task)
		logger.Infof("Schedule %p Task %v Range %v and %v", s, name, block.Start, block.End)
	}
}

func (ss *SequentialSchedule) MarshalJSON() ([]byte, error) {

	ws := &innerSequentialSchedule{Ranges: make(map[string]*innerBlock)}

	for k, v := range ss.Ranges {
		wt, err := ss.marshaller.WrapInstance(v.Task)
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

func (ss *SequentialSchedule) UnmarshalJSON(raw []byte) error {
	aa := &innerSequentialSchedule{}

	err := json.Unmarshal(raw, aa)
	if err != nil {
		return err
	}

	adminInterface := reflect.TypeOf((*interfaces.AdminClient)(nil)).Elem()

	ss.Ranges = make(map[string]*Block)
	for k, v := range aa.Ranges {
		t, err := ss.marshaller.UnwrapInstance(v.WrappedTask)
		if err != nil {
			return err
		}

		// Marshalling service handlers is mostly non-sense, so
		isAdminClient := reflect.TypeOf(t).Implements(adminInterface)
		if isAdminClient {
			adminClient := t.(interfaces.AdminClient)
			adminClient.SetAdminHandler(ss.adminHandler)
		}

		ss.Ranges[k] = &Block{Start: v.Start, End: v.End, Task: t.(interfaces.Task)}
	}

	return nil
}
