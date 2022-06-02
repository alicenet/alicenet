package executor

import (
	"context"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	dkgtasks "github.com/MadBase/MadNet/blockchain/executor/tasks/dkg"
	"github.com/MadBase/MadNet/blockchain/testutils"
	"github.com/MadBase/MadNet/consensus/admin"
	"github.com/MadBase/MadNet/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	lastHeightSeen uint64 = 12
	taskGroupName         = "ethdkg_group"
)

func getTaskScheduler(t *testing.T, lastFinalizedBlockChan chan uint64, taskRequestChan chan interfaces.ITask, taskKillChan chan string) *TasksScheduler {

	consDB := mocks.NewTestDB()
	consAdminHandlers := &admin.Handlers{}
	ecdsaPrivateKeys, _ := testutils.InitializePrivateKeysAndAccounts(5)
	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 1000*time.Millisecond)

	if lastFinalizedBlockChan == nil {
		lastFinalizedBlockChan = make(chan uint64, 100)
	}
	if taskRequestChan == nil {
		taskRequestChan = make(chan interfaces.ITask, 100)
	}
	if taskKillChan == nil {
		taskKillChan = make(chan string, 100)
	}
	return NewTasksScheduler(consDB, eth, consAdminHandlers, taskRequestChan, taskKillChan)
}

func TestTasksScheduler_Close(t *testing.T) {
	s := getTaskScheduler(t, nil, nil, nil)
	s.Close()

	_, ok := <-s.taskResponseChan.trChan
	if ok {
		assert.Fail(t, "Expected taskResponseChan to be closed")
	}
}

func TestTasksScheduler_Schedule_WrongExecutionData(t *testing.T) {

	ctx, _ := context.WithCancel(context.Background())
	s := getTaskScheduler(t, nil, nil, nil)

	taskInvalidParams := dkgtasks.CompletionTask{
		Task: &objects.Task{
			Start: 2,
			End:   1,
		},
	}
	err := s.schedule(ctx, &taskInvalidParams)
	assert.Equal(t, ErrWrongParams, err)

	taskExpired := dkgtasks.CompletionTask{
		Task: &objects.Task{
			End: 1,
		},
	}
	s.LastHeightSeen = lastHeightSeen
	err = s.schedule(ctx, &taskExpired)
	assert.Equal(t, ErrTaskExpired, err)
}

func TestTasksScheduler_Schedule_Success(t *testing.T) {

	ctx, _ := context.WithCancel(context.Background())
	s := getTaskScheduler(t, nil, nil, nil)

	s.LastHeightSeen = lastHeightSeen
	task := dkgtasks.CompletionTask{
		Task: &objects.Task{
			Start: 10,
			End:   20,
		},
	}

	assert.Emptyf(t, s.Schedule, "Expected Schedule map to be empty")
	err := s.schedule(ctx, &task)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(s.Schedule))

	for k, _ := range s.Schedule {
		taskRequest := s.Schedule[k]
		taskRequest.Start = task.Task.Start
		taskRequest.End = task.Task.End
	}
}

func TestTasksScheduler_ProcessTaskResponse_RemoveTaskWithErrNotNil(t *testing.T) {

	s := getTaskScheduler(t, nil, nil, nil)

	ctx, _ := context.WithCancel(context.Background())
	s.Schedule["1"] = TaskRequestInfo{"First", 1, 1, true, &dkgtasks.CompletionTask{Task: &objects.Task{Start: 10, End: 20}}}
	assert.NotEmptyf(t, s.Schedule, "Expected one task request scheduled")

	taskResponse := TaskResponse{
		Id:  "1",
		Err: ErrTaskExpired,
	}

	err := s.processTaskResponse(ctx, taskResponse)
	assert.Nil(t, err)
	assert.Emptyf(t, s.Schedule, "Expected no tasks")
}

func TestTasksScheduler_ProcessTaskResponse_RemoveNotScheduledTask(t *testing.T) {

	s := getTaskScheduler(t, nil, nil, nil)

	ctx, _ := context.WithCancel(context.Background())
	s.Schedule["1"] = TaskRequestInfo{"First", 1, 1, true, &dkgtasks.CompletionTask{Task: &objects.Task{Start: 10, End: 20}}}
	assert.NotEmptyf(t, s.Schedule, "Expected one task request scheduled")

	taskResponse := TaskResponse{
		Id:  "2",
		Err: ErrTaskExpired,
	}

	err := s.processTaskResponse(ctx, taskResponse)
	assert.Equal(t, err, ErrNotScheduled)
	assert.NotEmptyf(t, s.Schedule, "Expected one task to still be in scheduled")
}

func TestTasksScheduler_ProcessTaskResponse_RemoveTaskWithErrNil(t *testing.T) {

	s := getTaskScheduler(t, nil, nil, nil)

	ctx, _ := context.WithCancel(context.Background())
	s.Schedule["1"] = TaskRequestInfo{"First", 1, 1, true, &dkgtasks.CompletionTask{Task: &objects.Task{Start: 10, End: 20}}}
	assert.NotEmptyf(t, s.Schedule, "Expected one task request scheduled")

	taskResponse := TaskResponse{
		Id:  "1",
		Err: nil,
	}

	err := s.processTaskResponse(ctx, taskResponse)
	assert.Nil(t, err)
	assert.Emptyf(t, s.Schedule, "Expected no tasks")
}

func TestTasksScheduler_ProcessTaskResponse_RemoveTaskWithErrNilNotScheduledTask(t *testing.T) {

	s := getTaskScheduler(t, nil, nil, nil)

	ctx, _ := context.WithCancel(context.Background())
	s.Schedule["1"] = TaskRequestInfo{"First", 1, 1, true, &dkgtasks.CompletionTask{Task: &objects.Task{Start: 10, End: 20}}}
	assert.NotEmptyf(t, s.Schedule, "Expected one task request scheduled")

	taskResponse := TaskResponse{
		Id:  "2",
		Err: nil,
	}

	err := s.processTaskResponse(ctx, taskResponse)
	assert.Equal(t, err, ErrNotScheduled)
	assert.NotEmptyf(t, s.Schedule, "Expected one task to still be in scheduled")
}

func TestTasksScheduler_StartTask_Success(t *testing.T) {

	s := getTaskScheduler(t, nil, nil, nil)

	ctx, _ := context.WithCancel(context.Background())
	taskRequestList := []TaskRequestInfo{
		{"First", 1, 2, false, &dkgtasks.CompletionTask{Task: &objects.Task{Start: 10, End: 20}}},
		{"First", 1, 2, false, &dkgtasks.CompletionTask{Task: &objects.Task{Start: 10, End: 20}}},
	}
	err := s.startTasks(ctx, taskRequestList)
	assert.Nil(t, err)
	for _, task := range taskRequestList {
		task.IsRunning = true
		assert.Truef(t, task.IsRunning, "Expecting task to be running")
	}
}

func TestTasksScheduler_KillTaskByName(t *testing.T) {

	s := getTaskScheduler(t, nil, nil, nil)

	ctx, _ := context.WithCancel(context.Background())
	taskRequestList := []*TaskRequestInfo{
		{"First", 1, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Start: 10, End: 20, Name: taskGroupName}}},
		{"Second", 1, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Start: 10, End: 20, Name: taskGroupName}}},
		{"Third", 1, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Start: 10, End: 20, Name: "other"}}},
	}
	for _, taskRequest := range taskRequestList {
		err := s.schedule(ctx, taskRequest.Task)
		assert.Nil(t, err)
	}
	assert.Equalf(t, 3, len(s.Schedule), "Expected 3 tasks scheduled")
	time.Sleep(time.Second)
	err := s.killTaskByName(ctx, taskGroupName)
	assert.Nil(t, err)
	killedTaskList := s.findTasksByName(taskGroupName)
	assert.Emptyf(t, killedTaskList, "Expected no tasks with this name to be running")
}

func TestTasksScheduler_KillTasks(t *testing.T) {

	s := getTaskScheduler(t, nil, nil, nil)

	ctx, _ := context.WithCancel(context.Background())
	taskRequestList := []TaskRequestInfo{
		{"First", 1, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Id: "1", Start: 10, End: 20, Name: taskGroupName}}},
		{"Second", 1, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Id: "2", Start: 10, End: 20, Name: taskGroupName}}},
		{"Third", 1, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Id: "3", Start: 10, End: 20, Name: "other"}}},
	}
	err := s.killTasks(ctx, taskRequestList)
	assert.Nil(t, err)
	killedTaskList := s.findTasksByName(taskGroupName)
	assert.Emptyf(t, killedTaskList, "Expected no tasks with this name to be running")
}

func TestTasksScheduler_RemoveUnresponsiveTasks_WithScheduledAndNoScheduledTask(t *testing.T) {

	s := getTaskScheduler(t, nil, nil, nil)

	ctx, _ := context.WithCancel(context.Background())
	taskRequestList := []TaskRequestInfo{
		{"First", 1, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Id: "1", Start: 10, End: 20, Name: taskGroupName}}},
		{"Second", 1, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Id: "2", Start: 10, End: 20, Name: taskGroupName}}},
		{"Third", 1, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Id: "3", Start: 10, End: 20, Name: "other"}}},
	}
	// Not scheduling the first task on purpose
	for i := 1; i < len(taskRequestList); i++ {
		err := s.schedule(ctx, taskRequestList[i].Task)
		assert.Nil(t, err)
	}
	err := s.removeUnresponsiveTasks(ctx, taskRequestList)
	assert.Nil(t, err)
	assert.Emptyf(t, s.Schedule, "Expected no tasks")
}

func TestTasksScheduler_Purge(t *testing.T) {

	s := getTaskScheduler(t, nil, nil, nil)

	ctx, _ := context.WithCancel(context.Background())
	taskRequestList := []TaskRequestInfo{
		{"First", 1, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Id: "1", Start: 10, End: 20, Name: taskGroupName}}},
		{"Second", 1, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Id: "2", Start: 10, End: 20, Name: taskGroupName}}},
		{"Third", 1, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Id: "3", Start: 10, End: 20, Name: "other"}}},
	}
	// Not scheduling the first task on purpose
	for _, taskRequest := range taskRequestList {
		err := s.schedule(ctx, taskRequest.Task)
		assert.Nil(t, err)
	}
	s.purge()
	assert.Emptyf(t, s.Schedule, "Expected no tasks")
}

func TestTasksScheduler_FindTasks(t *testing.T) {

	s := getTaskScheduler(t, nil, nil, nil)

	ctx, _ := context.WithCancel(context.Background())
	taskRequestList := []TaskRequestInfo{
		{"First", 1, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Id: "1", Start: 10, End: 20, Name: taskGroupName}}},
		{"Second", 1, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Id: "2", Start: 10, End: 20, Name: taskGroupName}}},
		{"Third", 1, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Id: "3", Start: 10, End: 20, Name: "other"}}},
	}
	// Not scheduling the first task on purpose
	for _, taskRequest := range taskRequestList {
		err := s.schedule(ctx, taskRequest.Task)
		assert.Nil(t, err)
	}
	s.purge()
	assert.Emptyf(t, s.Schedule, "Expected no tasks")
}

//---

func TestTasksScheduler_PersistState_Success(t *testing.T) {

	s := getTaskScheduler(t, nil, nil, nil)

	s.LastHeightSeen = 0
	err := s.PersistState()
	assert.Nil(t, err)

	err = s.LoadState()
	assert.Nil(t, err)

	lastHeightSeenBeforeAfter := s.LastHeightSeen
	s.LastHeightSeen = lastHeightSeen
	err = s.PersistState()
	err = s.LoadState()
	lastHeightSeenAfter := s.LastHeightSeen
	assert.Nil(t, err)
	assert.NotEqualf(t, lastHeightSeenBeforeAfter, lastHeightSeenAfter, "Expected TaskScheduler to be different")
}

func TestTasksScheduler_LoadState_MissingKey(t *testing.T) {

	s := getTaskScheduler(t, nil, nil, nil)

	err := s.LoadState()
	assert.NotNil(t, err)
}

func TestTasksScheduler_Start_EmptySchedule(t *testing.T) {

	s := getTaskScheduler(t, nil, nil, nil)

	err := s.PersistState()
	assert.Nil(t, err)

	assert.Emptyf(t, s.Schedule, "Scheduled map expected to be empty")
	err = s.Start()
	assert.Nil(t, err)
	assert.Emptyf(t, s.Schedule, "Scheduled map expected to still be empty")
}

func TestTasksScheduler_EventLoop_PurgeToEmptyTheSchedulerMap(t *testing.T) {

	taskRequestChan := make(chan interfaces.ITask, 100)
	s := getTaskScheduler(t, nil, taskRequestChan, nil)

	s.Schedule["1"] = TaskRequestInfo{"First", 1, 1, true, &dkgtasks.CompletionTask{Task: &objects.Task{Start: 10, End: 20}}}
	s.Schedule["2"] = TaskRequestInfo{"Second", 2, 2, true, &dkgtasks.CompletionTask{Task: &objects.Task{Start: 10, End: 20}}}
	assert.Equal(t, 2, len(s.Schedule))
	s.purge()

	assert.Equal(t, 0, len(s.Schedule))
}

func TestTasksScheduler_EventLoop_LastFinalizedBlock(t *testing.T) {

	lastFinalizedBlockChan := make(chan uint64, 100)
	s := getTaskScheduler(t, lastFinalizedBlockChan, nil, nil)

	assert.Equal(t, uint64(0), s.LastHeightSeen)
	lastFinalizedBlockChan <- lastHeightSeen
	s.cancelChan <- true
	s.eventLoop()

	select {
	case <-lastFinalizedBlockChan:
		assert.Equal(t, uint64(0), s.LastHeightSeen)
	case <-time.After(time.Second):
		assert.Equal(t, lastHeightSeen, s.LastHeightSeen)
	}
}

func TestTasksScheduler_Schedule_ScheduleTask(t *testing.T) {

	taskRequestChan := make(chan interfaces.ITask, 100)
	s := getTaskScheduler(t, nil, taskRequestChan, nil)

	s.LastHeightSeen = lastHeightSeen
	task := dkgtasks.CompletionTask{
		Task: &objects.Task{
			Start: 10,
			End:   20,
		},
	}

	err := s.PersistState()
	assert.Nil(t, err)
	s2 := s
	assert.Equal(t, s2, s)

	taskRequestChan <- &task
	s.cancelChan <- true
	err = s.LoadState()
	assert.Nil(t, err)

	s.eventLoop()

	select {
	case <-time.After(time.Second):
		assert.Equalf(t, s2, s, "Expected initial state to be the same")
	}
}
