package tasks_test

import (
	"context"
	"encoding/json"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/dkgtasks"
	objects2 "github.com/MadBase/MadNet/blockchain/tasks/dkg/objects"
	"testing"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/consensus/admin"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSchedule(t *testing.T) {
	m := &objects2.TypeRegistry{}
	s := tasks.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var task interfaces.ITask

	s.Schedule(1, 2, task)

	s.Schedule(3, 4, task)

	s.Schedule(5, 6, task)

	s.Schedule(7, 8, task)

	assert.Equal(t, 4, s.Length())
}

func TestPurge(t *testing.T) {
	m := &objects2.TypeRegistry{}
	s := tasks.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var task interfaces.ITask

	s.Schedule(1, 2, task)

	s.Schedule(3, 4, task)

	s.Schedule(5, 6, task)

	s.Schedule(7, 8, task)

	assert.Equal(t, 4, s.Length())

	s.Purge()

	assert.Equal(t, 0, s.Length())
}

func TestPurgePrior(t *testing.T) {
	m := &objects2.TypeRegistry{}
	s := tasks.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var task interfaces.ITask

	s.Schedule(1, 2, task)

	s.Schedule(3, 4, task)

	s.Schedule(5, 6, task)

	s.Schedule(7, 8, task)

	assert.Equal(t, 4, s.Length())

	s.PurgePrior(7)

	assert.Equal(t, 1, s.Length())
}

func TestFailSchedule(t *testing.T) {
	m := &objects2.TypeRegistry{}
	s := tasks.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var task interfaces.ITask

	s.Schedule(5, 15, task)
	s.Schedule(4, 6, task)
	s.Schedule(6, 14, task)
	s.Schedule(14, 16, task)
	s.Schedule(4, 16, task)

	assert.Equal(t, 1, s.Length())
}

func TestFailSchedule2(t *testing.T) {
	m := &objects2.TypeRegistry{}
	s := tasks.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var err error
	var task interfaces.ITask

	s.Schedule(1, 2, task)

	s.Schedule(2, 3, task)
	assert.Nil(t, err)
}

func TestFailSchedule3(t *testing.T) {
	m := &objects2.TypeRegistry{}
	s := tasks.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var err error
	var task interfaces.ITask

	s.Schedule(7, 15, task)

	s.Schedule(15, 17, task)

	assert.Nil(t, err)

	s.Schedule(15, 21, task)

	assert.NotNil(t, err)

	s.Schedule(1, 7, task)

	assert.Nil(t, err)

	s.Schedule(1, 8, task)

	assert.NotNil(t, err)
}

func TestFind(t *testing.T) {
	m := &objects2.TypeRegistry{}
	s := tasks.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var task interfaces.ITask

	id := s.Schedule(5, 15, task)

	taskID, err := s.Find(10)
	assert.Nil(t, err)
	assert.Equal(t, id, taskID)

	taskID, err = s.Find(14)
	assert.Nil(t, err)
	assert.Equal(t, id, taskID)

	taskID, err = s.Find(5)
	assert.Nil(t, err)
	assert.Equal(t, id, taskID)

	_, err = s.Find(15)
	assert.NotNil(t, err)

	_, err = s.Find(4)
	assert.NotNil(t, err)

}

func TestFailFind(t *testing.T) {
	m := &objects2.TypeRegistry{}
	s := tasks.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var task interfaces.ITask

	s.Schedule(5, 15, task)

	_, err := s.Find(4)
	assert.Equal(t, tasks.ErrNothingScheduled, err)
}

func TestFailFind2(t *testing.T) {
	m := &objects2.TypeRegistry{}
	s := tasks.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var task interfaces.ITask

	s.Schedule(5, 15, task)

	_, err := s.Find(15)
	assert.Equal(t, tasks.ErrNothingScheduled, err)
}

func TestRemove(t *testing.T) {
	acct := accounts.Account{}
	state := objects2.NewDkgState(acct)
	task := dkgtasks.NewPlaceHolder(state)
	m := &objects2.TypeRegistry{}
	s := tasks.NewSequentialSchedule(m, nil)

	taskID := s.Schedule(5, 15, task)

	assert.Equal(t, 1, s.Length())

	err := s.Remove(taskID)
	assert.Nil(t, err)

	assert.Equal(t, 0, s.Length())
}

func TestFailRemove(t *testing.T) {
	acct := accounts.Account{}
	state := objects2.NewDkgState(acct)
	task := dkgtasks.NewPlaceHolder(state)

	m := &objects2.TypeRegistry{}
	s := tasks.NewSequentialSchedule(m, nil)

	// Schedule something but don't bother saving the id
	s.Schedule(5, 15, task)
	assert.Equal(t, 1, s.Length())

	// Make up a random id
	taskID := uuid.NewRandom()
	err := s.Remove(taskID)
	assert.Equal(t, tasks.ErrNotScheduled, err)

	// Nothing should have been removed
	assert.Equal(t, 1, s.Length())
}

func TestRetreive(t *testing.T) {
	acct := accounts.Account{}
	state := objects2.NewDkgState(acct)
	task := dkgtasks.NewPlaceHolder(state)
	m := &objects2.TypeRegistry{}
	s := tasks.NewSequentialSchedule(m, nil)

	// tasks.RegisterTask(task)

	// Schedule something
	taskID := s.Schedule(5, 15, task)

	_, err := s.Retrieve(taskID)
	assert.Nil(t, err)
}

func TestFailRetrieve(t *testing.T) {
	acct := accounts.Account{}
	state := objects2.NewDkgState(acct)
	task := dkgtasks.NewPlaceHolder(state)
	m := &objects2.TypeRegistry{}
	s := tasks.NewSequentialSchedule(m, nil)

	// tasks.RegisterTask(task)

	// Schedule something
	s.Schedule(5, 15, task)

	_, err := s.Retrieve(uuid.NewRandom())
	assert.Equal(t, tasks.ErrNotScheduled, err)
}

func TestMarshal(t *testing.T) {
	task := &adminTaskMock{}
	m := &objects2.TypeRegistry{}
	m.RegisterInstanceType(&tasks.Block{})
	m.RegisterInstanceType(task)
	s := tasks.NewSequentialSchedule(m, nil)

	// Schedule something
	taskID := s.Schedule(5, 15, task)
	assert.NotNil(t, taskID)
	assert.Equal(t, 1, s.Length())

	// Marshal the schedule
	raw, err := json.Marshal(s)
	assert.Nil(t, err)

	// Unmarshal the schedule
	ns := tasks.NewSequentialSchedule(m, nil)
	err = json.Unmarshal(raw, &ns)
	assert.Nil(t, err)
	assert.Equal(t, 1, ns.Length())

	// Make sure the schedule and task  are correct
	block, present := ns.Ranges[taskID.String()]
	assert.True(t, present)
	assert.Equal(t, uint64(5), block.Start)
	assert.Equal(t, uint64(15), block.End)

	// Confirm task survived marshalling
	_, err = s.Retrieve(taskID)
	assert.Nil(t, err)
}

type adminTaskMock struct {
}

var _ interfaces.ITask = &adminTaskMock{}

func (ph *adminTaskMock) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return nil
}
func (ph *adminTaskMock) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return nil
}

func (ph *adminTaskMock) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return nil
}

func (ph *adminTaskMock) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	return false
}

func (ph *adminTaskMock) DoDone(logger *logrus.Entry) {
}

func (ph *adminTaskMock) SetAdminHandler(adminHandler *admin.Handlers) {
}

func (ph *adminTaskMock) GetExecutionData() interfaces.ITaskExecutionData {
	return nil
}
