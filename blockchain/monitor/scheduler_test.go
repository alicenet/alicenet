package monitor_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/monitor"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type dumbTask struct {
	Val int
}

// DoDone(*logrus.Logger)
// DoRetry(context.Context, *logrus.Logger, blockchain.Ethereum) bool
// DoWork(context.Context, *logrus.Logger, blockchain.Ethereum) bool
// ShouldRetry(context.Context, *logrus.Logger, blockchain.Ethereum) bool

func (task *dumbTask) DoDone(logger *logrus.Logger) {
	logger.Infof("DoDone():%v", task.Val)
}

func (task *dumbTask) DoRetry(context context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	logger.Infof("DoRetry():%v", task.Val)
	return true
}

func (task *dumbTask) DoWork(context context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	logger.Infof("DoWork():%v", task.Val)
	return true
}

func (task *dumbTask) ShouldRetry(context context.Context, logger *logrus.Logger, eth blockchain.Ethereum) bool {
	logger.Infof("ShouldRetry():%v", task.Val)
	return true
}

func TestSchedule(t *testing.T) {
	s := monitor.NewSequentialSchedule()
	assert.NotNil(t, s, "Scheduler should not be nil")

	var err error
	var task tasks.Task

	_, err = s.Schedule(1, 2, task)
	assert.Nil(t, err)

	_, err = s.Schedule(3, 4, task)
	assert.Nil(t, err)

	_, err = s.Schedule(5, 6, task)
	assert.Nil(t, err)

	_, err = s.Schedule(7, 8, task)
	assert.Nil(t, err)

	assert.Equal(t, 4, s.Length())
}

func TestFailSchedule(t *testing.T) {
	s := monitor.NewSequentialSchedule()
	assert.NotNil(t, s, "Scheduler should not be nil")

	var err error
	var task tasks.Task

	_, err = s.Schedule(5, 15, task)
	assert.Nil(t, err)

	_, err = s.Schedule(4, 6, task)
	assert.NotNil(t, err)

	_, err = s.Schedule(6, 14, task)
	assert.NotNil(t, err)

	_, err = s.Schedule(14, 16, task)
	assert.NotNil(t, err)

	_, err = s.Schedule(4, 16, task)
	assert.NotNil(t, err)

	assert.Equal(t, 1, s.Length())
}

func TestFind(t *testing.T) {
	s := monitor.NewSequentialSchedule()
	assert.NotNil(t, s, "Scheduler should not be nil")

	var task tasks.Task

	id, err := s.Schedule(5, 15, task)
	assert.Nil(t, err)

	taskID, err := s.Find(10)
	assert.Nil(t, err)
	assert.Equal(t, id, taskID)
}

func TestFailFind(t *testing.T) {
	s := monitor.NewSequentialSchedule()
	assert.NotNil(t, s, "Scheduler should not be nil")

	var task tasks.Task

	_, err := s.Schedule(5, 15, task)
	assert.Nil(t, err)

	_, err = s.Find(4)
	assert.Equal(t, monitor.ErrNothingScheduled, err)
}

func TestRemove(t *testing.T) {
	s := monitor.NewSequentialSchedule()

	task := new(dumbTask)

	taskID, err := s.Schedule(5, 15, task)
	assert.Nil(t, err)

	assert.Equal(t, 1, s.Length())

	s.Remove(taskID)

	assert.Equal(t, 0, s.Length())
}

func TestFailRemove(t *testing.T) {
	s := monitor.NewSequentialSchedule()

	task := new(dumbTask)

	// Schedule something but don't bother saving the id
	_, err := s.Schedule(5, 15, task)
	assert.Nil(t, err)
	assert.Equal(t, 1, s.Length())

	// Make up a random id
	taskID := uuid.NewRandom()
	err = s.Remove(taskID)
	assert.Equal(t, monitor.ErrNotScheduled, err)

	// Nothing should have been removed
	assert.Equal(t, 1, s.Length())
}

func TestRetreive(t *testing.T) {
	s := monitor.NewSequentialSchedule()

	var task tasks.Task = &dumbTask{Val: 4}

	tasks.RegisterTask(task)

	// Schedule something
	taskID, err := s.Schedule(5, 15, task)
	assert.Nil(t, err)

	_, err = s.Retrieve(taskID)
	assert.Nil(t, err)
}

func TestFailRetrieve(t *testing.T) {

	s := monitor.NewSequentialSchedule()

	var task tasks.Task = &dumbTask{Val: 4}

	tasks.RegisterTask(task)

	// Schedule something
	_, err := s.Schedule(5, 15, task)
	assert.Nil(t, err)

	_, err = s.Retrieve(uuid.NewRandom())
	assert.Equal(t, monitor.ErrNotScheduled, err)
}

func TestMarshal(t *testing.T) {

	s := monitor.NewSequentialSchedule()

	var task tasks.Task = &dumbTask{Val: 4}

	tasks.RegisterTask(task)

	// Schedule something
	taskID, err := s.Schedule(5, 15, task)
	assert.Nil(t, err)
	assert.NotNil(t, taskID)
	assert.Equal(t, 1, s.Length())

	// Marshal the schedule
	raw, err := json.Marshal(s)
	assert.Nil(t, err)

	// Unmarshal the schedule
	ns := &monitor.SequentialSchedule{}
	err = json.Unmarshal(raw, &ns)
	assert.Nil(t, err)
	assert.NotNil(t, taskID)
	assert.Equal(t, 1, s.Length())

	// Confirm task survived marshalling
	_, err = s.Retrieve(taskID)
	assert.Nil(t, err)
}
