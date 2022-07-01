package executor

import (
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func getTaskScheduler(t *testing.T) (*TasksScheduler, chan tasks.TaskRequest, *mocks.MockClient) {
	db := mocks.NewTestDB()
	client := mocks.NewMockClient()
	adminHandlers := mocks.NewMockAdminHandler()
	txWatcher := transaction.NewWatcher(client, 12, db, false, constants.TxPollingTime)
	taskRequestChan := make(chan tasks.TaskRequest, constants.TaskSchedulerBufferSize)
	tasksScheduler, err := NewTasksScheduler(db, client, adminHandlers, taskRequestChan, txWatcher)
	assert.Nil(t, err)
	return tasksScheduler, taskRequestChan, client
}

func TestTasksScheduler_Schedule_NilTask(t *testing.T) {
	scheduler, tasksChan, _ := getTaskScheduler(t)
	defer close(tasksChan)
	err := scheduler.Start()
	assert.Nil(t, err)
	defer scheduler.Close()

	request := tasks.NewScheduleTaskRequest(nil)
	tasksChan <- request

	select {
	case <-time.After(10 * time.Millisecond):
	}
	assert.Emptyf(t, scheduler.Schedule, "Expected zero tasks scheduled")
}

func TestTasksScheduler_Schedule_WrongExecutionData(t *testing.T) {

	scheduler, tasksChan, _ := getTaskScheduler(t)
	defer close(tasksChan)
	err := scheduler.Start()
	assert.Nil(t, err)
	defer scheduler.Close()

	task := dkg.NewCompletionTask(2, 1)
	request := tasks.NewScheduleTaskRequest(task)
	tasksChan <- request

	scheduler.LastHeightSeen = 12
	task = dkg.NewCompletionTask(2, 3)
	request = tasks.NewScheduleTaskRequest(task)
	tasksChan <- request

	select {
	case <-time.After(20 * time.Millisecond):
	}
	assert.Emptyf(t, scheduler.Schedule, "Expected zero tasks scheduled")
}

func TestTasksScheduler_ScheduleAndKillTasks_Success(t *testing.T) {

	scheduler, tasksChan, _ := getTaskScheduler(t)
	defer close(tasksChan)
	err := scheduler.Start()
	assert.Nil(t, err)
	defer scheduler.Close()

	completionTask := dkg.NewCompletionTask(2, 3)
	request := tasks.NewScheduleTaskRequest(completionTask)
	tasksChan <- request

	completionTask2 := dkg.NewCompletionTask(3, 4)
	request = tasks.NewScheduleTaskRequest(completionTask2)
	tasksChan <- request

	registerTask := dkg.NewRegisterTask(2, 5)
	request = tasks.NewScheduleTaskRequest(registerTask)
	tasksChan <- request

	select {
	case <-time.After(10 * time.Millisecond):
	}
	assert.Equalf(t, 3, len(scheduler.Schedule), "Expected 3 task scheduled")

	request = tasks.NewKillTaskRequest(&dkg.CompletionTask{})
	tasksChan <- request
	select {
	case <-time.After(10 * time.Millisecond):
	}
	assert.Equalf(t, 1, len(scheduler.Schedule), "Expected 1 task after Completion tasks have been killed")

	request = tasks.NewKillTaskRequest(&dkg.DisputeMissingGPKjTask{})
	tasksChan <- request
	select {
	case <-time.After(10 * time.Millisecond):
	}
	assert.Equalf(t, 1, len(scheduler.Schedule), "There should be 1 tasks left still, due there were no DisputeMissing task scheduled")

	request = tasks.NewKillTaskRequest(&dkg.RegisterTask{})
	tasksChan <- request
	select {
	case <-time.After(10 * time.Millisecond):
	}
	assert.Equalf(t, 0, len(scheduler.Schedule), "All the tasks should have been removed")
}

func TestTasksScheduler_ScheduleRunAndKillTask_Success(t *testing.T) {

	scheduler, tasksChan, client := getTaskScheduler(t)
	defer close(tasksChan)
	err := scheduler.Start()
	assert.Nil(t, err)
	defer scheduler.Close()

	completionTask := dkg.NewCompletionTask(1, 40)
	request := tasks.NewScheduleTaskRequest(completionTask)
	tasksChan <- request

	client.GetFinalizedHeightFunc.SetDefaultReturn(10, nil)

	select {
	case <-time.After(constants.TaskSchedulerProcessingTime + 10*time.Millisecond):
	}

	assert.Equalf(t, 0, len(scheduler.Schedule), "All the tasks should have been removed")
}

func TestTasksScheduler_ScheduleDuplicatedTask_Success(t *testing.T) {

	scheduler, tasksChan, client := getTaskScheduler(t)
	defer close(tasksChan)
	err := scheduler.Start()
	assert.Nil(t, err)
	defer scheduler.Close()

	completionTask := dkg.NewCompletionTask(1, 40)
	request := tasks.NewScheduleTaskRequest(completionTask)
	tasksChan <- request
	completionTask2 := dkg.NewCompletionTask(1, 40)
	request = tasks.NewScheduleTaskRequest(completionTask2)
	tasksChan <- request

	client.GetFinalizedHeightFunc.SetDefaultReturn(10, nil)

	select {
	case <-time.After(constants.TaskSchedulerProcessingTime + 10*time.Millisecond):
	}

	assert.Equalf(t, 1, len(scheduler.Schedule), "Expected to have 1 task")
	for _, task := range scheduler.Schedule {
		assert.Falsef(t, task.isRunning, "this task shouldn't be running due to duplication")
	}
}

func TestTasksScheduler_ScheduleAndKillExpiredAndUnresponsiveTasks_Success(t *testing.T) {

	scheduler, tasksChan, client := getTaskScheduler(t)
	defer close(tasksChan)
	err := scheduler.Start()
	assert.Nil(t, err)
	defer scheduler.Close()

	completionTask := dkg.NewCompletionTask(50, 90)
	request := tasks.NewScheduleTaskRequest(completionTask)
	tasksChan <- request
	completionTask2 := dkg.NewCompletionTask(1, 10)
	request = tasks.NewScheduleTaskRequest(completionTask2)
	tasksChan <- request
	completionTask3 := dkg.NewCompletionTask(110, 150)
	request = tasks.NewScheduleTaskRequest(completionTask3)
	tasksChan <- request

	client.GetFinalizedHeightFunc.SetDefaultReturn(100, nil)

	select {
	case <-time.After(constants.TaskSchedulerProcessingTime + 10*time.Millisecond):
	}

	assert.Equalf(t, 1, len(scheduler.Schedule), "Expected to have 1 task")
}

func TestTasksScheduler_Recovery_Success(t *testing.T) {
	dir, err := ioutil.TempDir("", "db-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	rawDB, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer rawDB.Close()

	db := &db.Database{}
	db.Init(rawDB)

	client := mocks.NewMockClient()
	adminHandlers := mocks.NewMockAdminHandler()
	txWatcher := transaction.NewWatcher(client, 12, db, false, constants.TxPollingTime)
	tasksChan := make(chan tasks.TaskRequest, constants.TaskSchedulerBufferSize)
	scheduler, err := NewTasksScheduler(db, client, adminHandlers, tasksChan, txWatcher)
	assert.Nil(t, err)
	err = scheduler.Start()

	completionTask := dkg.NewCompletionTask(50, 90)
	request := tasks.NewScheduleTaskRequest(completionTask)
	tasksChan <- request
	completionTask2 := dkg.NewCompletionTask(1, 10)
	request2 := tasks.NewScheduleTaskRequest(completionTask2)
	tasksChan <- request2
	completionTask3 := dkg.NewCompletionTask(110, 150)
	request3 := tasks.NewScheduleTaskRequest(completionTask3)
	tasksChan <- request3

	select {
	case <-time.After(10 * time.Millisecond):
	}

	assert.Equalf(t, 3, len(scheduler.Schedule), "Expected to have 3 tasks")

	scheduler.Close()
	close(tasksChan)

	tasksChan = make(chan tasks.TaskRequest, constants.TaskSchedulerBufferSize)
	scheduler, err = NewTasksScheduler(db, client, adminHandlers, tasksChan, txWatcher)
	assert.Nil(t, err)
	err = scheduler.Start()
	assert.Nil(t, err)
	assert.Equalf(t, 3, len(scheduler.Schedule), "Expected to have 3 tasks")

	scheduler.Close()
	select {
	case <-time.After(10 * time.Millisecond):
	}
	close(tasksChan)
}
