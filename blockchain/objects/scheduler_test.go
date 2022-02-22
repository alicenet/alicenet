package objects_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/consensus/admin"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSchedule(t *testing.T) {
	m := &objects.TypeRegistry{}
	s := objects.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var err error
	var task interfaces.Task

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

func TestPurge(t *testing.T) {
	m := &objects.TypeRegistry{}
	s := objects.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var err error
	var task interfaces.Task

	_, err = s.Schedule(1, 2, task)
	assert.Nil(t, err)

	_, err = s.Schedule(3, 4, task)
	assert.Nil(t, err)

	_, err = s.Schedule(5, 6, task)
	assert.Nil(t, err)

	_, err = s.Schedule(7, 8, task)
	assert.Nil(t, err)

	assert.Equal(t, 4, s.Length())

	s.Purge()

	assert.Equal(t, 0, s.Length())
}

func TestPurgePrior(t *testing.T) {
	m := &objects.TypeRegistry{}
	s := objects.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var err error
	var task interfaces.Task

	_, err = s.Schedule(1, 2, task)
	assert.Nil(t, err)

	_, err = s.Schedule(3, 4, task)
	assert.Nil(t, err)

	_, err = s.Schedule(5, 6, task)
	assert.Nil(t, err)

	_, err = s.Schedule(7, 8, task)
	assert.Nil(t, err)

	assert.Equal(t, 4, s.Length())

	s.PurgePrior(7)

	assert.Equal(t, 1, s.Length())
}

func TestFailSchedule(t *testing.T) {
	m := &objects.TypeRegistry{}
	s := objects.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var err error
	var task interfaces.Task

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

func TestFailSchedule2(t *testing.T) {
	m := &objects.TypeRegistry{}
	s := objects.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var err error
	var task interfaces.Task

	_, err = s.Schedule(1, 2, task)
	assert.Nil(t, err)

	_, err = s.Schedule(2, 3, task)
	t.Log("error:", err)
	assert.Nil(t, err)
}

func TestFailSchedule3(t *testing.T) {
	m := &objects.TypeRegistry{}
	s := objects.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var err error
	var task interfaces.Task

	_, err = s.Schedule(7, 15, task)
	assert.Nil(t, err)

	_, err = s.Schedule(15, 17, task)
	t.Log("error:", err)
	assert.Nil(t, err)

	_, err = s.Schedule(15, 21, task)
	t.Log("error:", err)
	assert.NotNil(t, err)

	_, err = s.Schedule(1, 7, task)
	t.Log("error:", err)
	assert.Nil(t, err)

	_, err = s.Schedule(1, 8, task)
	t.Log("error:", err)
	assert.NotNil(t, err)
}

func TestFind(t *testing.T) {
	m := &objects.TypeRegistry{}
	s := objects.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var task interfaces.Task

	id, err := s.Schedule(5, 15, task)
	assert.Nil(t, err)

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
	m := &objects.TypeRegistry{}
	s := objects.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var task interfaces.Task

	_, err := s.Schedule(5, 15, task)
	assert.Nil(t, err)

	_, err = s.Find(4)
	assert.Equal(t, objects.ErrNothingScheduled, err)
}

func TestFailFind2(t *testing.T) {
	m := &objects.TypeRegistry{}
	s := objects.NewSequentialSchedule(m, nil)
	assert.NotNil(t, s, "Scheduler should not be nil")

	var task interfaces.Task

	_, err := s.Schedule(5, 15, task)
	assert.Nil(t, err)

	_, err = s.Find(15)
	assert.Equal(t, objects.ErrNothingScheduled, err)
}

func TestRemove(t *testing.T) {
	acct := accounts.Account{}
	state := objects.NewDkgState(acct)
	task := dkgtasks.NewPlaceHolder(state)
	m := &objects.TypeRegistry{}
	s := objects.NewSequentialSchedule(m, nil)

	taskID, err := s.Schedule(5, 15, task)
	assert.Nil(t, err)

	assert.Equal(t, 1, s.Length())

	s.Remove(taskID)

	assert.Equal(t, 0, s.Length())
}

func TestFailRemove(t *testing.T) {
	acct := accounts.Account{}
	state := objects.NewDkgState(acct)
	task := dkgtasks.NewPlaceHolder(state)

	m := &objects.TypeRegistry{}
	s := objects.NewSequentialSchedule(m, nil)

	// Schedule something but don't bother saving the id
	_, err := s.Schedule(5, 15, task)
	assert.Nil(t, err)
	assert.Equal(t, 1, s.Length())

	// Make up a random id
	taskID := uuid.NewRandom()
	err = s.Remove(taskID)
	assert.Equal(t, objects.ErrNotScheduled, err)

	// Nothing should have been removed
	assert.Equal(t, 1, s.Length())
}

func TestRetreive(t *testing.T) {
	acct := accounts.Account{}
	state := objects.NewDkgState(acct)
	task := dkgtasks.NewPlaceHolder(state)
	m := &objects.TypeRegistry{}
	s := objects.NewSequentialSchedule(m, nil)

	// tasks.RegisterTask(task)

	// Schedule something
	taskID, err := s.Schedule(5, 15, task)
	assert.Nil(t, err)

	_, err = s.Retrieve(taskID)
	assert.Nil(t, err)
}

func TestFailRetrieve(t *testing.T) {
	acct := accounts.Account{}
	state := objects.NewDkgState(acct)
	task := dkgtasks.NewPlaceHolder(state)
	m := &objects.TypeRegistry{}
	s := objects.NewSequentialSchedule(m, nil)

	// tasks.RegisterTask(task)

	// Schedule something
	_, err := s.Schedule(5, 15, task)
	assert.Nil(t, err)

	_, err = s.Retrieve(uuid.NewRandom())
	assert.Equal(t, objects.ErrNotScheduled, err)
}

func TestMarshal(t *testing.T) {
	task := &adminTaskMock{}
	m := &objects.TypeRegistry{}
	m.RegisterInstanceType(&objects.Block{})
	m.RegisterInstanceType(task)
	s := objects.NewSequentialSchedule(m, nil)

	// Schedule something
	taskID, err := s.Schedule(5, 15, task)
	assert.Nil(t, err)
	assert.NotNil(t, taskID)
	assert.Equal(t, 1, s.Length())

	// Marshal the schedule
	raw, err := json.Marshal(s)
	assert.Nil(t, err)

	// Unmarshal the schedule
	ns := objects.NewSequentialSchedule(m, nil)
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

func (ph *adminTaskMock) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
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
