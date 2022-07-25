package executor

import (
	"context"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	taskMocks "github.com/alicenet/alicenet/layer1/executor/tasks/mocks"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func getTaskHandler(t *testing.T) (*Handler, *mocks.MockClient, layer1.AllSmartContracts) {
	db := mocks.NewTestDB()
	client := mocks.NewMockClient()
	adminHandlers := mocks.NewMockAdminHandler()
	adminHandlers.IsSynchronizedFunc.SetDefaultReturn(true)
	txWatcher := mocks.NewMockWatcher()
	contracts := mocks.NewMockAllSmartContracts()

	taskHandler, err := NewTaskHandler(db, client, contracts, adminHandlers, txWatcher)
	require.Nil(t, err)
	t.Cleanup(func() {
		txWatcher.Close()
		taskHandler.Close()
		db.DB().Close()
	})

	taskHandler.Start()
	return taskHandler.(*Handler), client, contracts
}

// Auxiliary function to get how many tasks we have inside the manager. This
// function creates a copy of the manager to get the len without race
// conditions.
func getSchedulerLen(t *testing.T, manager *TaskManager) int {
	newManager := &TaskManager{Schedule: make(map[string]ManagerRequestInfo), Responses: make(map[string]ManagerResponseInfo), marshaller: getTaskRegistry(), database: manager.database}
	err := manager.persistState()
	require.Nil(t, err)
	err = newManager.loadState()
	require.Nil(t, err)
	return len(newManager.Schedule)
}

func TestTasksScheduler_Schedule_NilTask(t *testing.T) {
	handler, _, _ := getTaskHandler(t)
	ctx := context.Background()

	_, err := handler.ScheduleTask(ctx, nil, "")
	require.Equal(t, ErrTaskIsNil, err)
	require.Equal(t, 0, getSchedulerLen(t, handler.manager))
}

func TestTasksScheduler_Schedule_NotRegistredTask(t *testing.T) {
	handler, _, _ := getTaskHandler(t)
	ctx := context.Background()

	_, err := handler.ScheduleTask(ctx, taskMocks.NewMockTask(), "")
	require.Equal(t, ErrTaskTypeNotInRegistry, err)
	require.Equal(t, 0, getSchedulerLen(t, handler.manager))
}

func TestTasksScheduler_Schedule_WrongStartDate(t *testing.T) {
	handler, _, _ := getTaskHandler(t)
	ctx := context.Background()

	task := dkg.NewCompletionTask(2, 1)
	_, err := handler.ScheduleTask(ctx, task, "")
	require.Equal(t, ErrWrongParams, err)
	require.Equal(t, 0, getSchedulerLen(t, handler.manager))
}

func TestTasksScheduler_Schedule_WrongEndDate(t *testing.T) {
	handler, client, _ := getTaskHandler(t)
	ctx := context.Background()
	client.GetFinalizedHeightFunc.SetDefaultReturn(12, nil)
	<-time.After(constants.TaskManagerProcessingTime)

	task := dkg.NewCompletionTask(2, 3)
	_, err := handler.ScheduleTask(ctx, task, "")
	require.Equal(t, ErrTaskExpired, err)
	require.Equal(t, 0, getSchedulerLen(t, handler.manager))
}

func TestTasksScheduler_Schedule_MultiExecutionNotAllowed(t *testing.T) {
	handler, _, _ := getTaskHandler(t)
	ctx := context.Background()

	task := dkg.NewCompletionTask(10, 40)
	resp, err := handler.ScheduleTask(ctx, task, "")
	require.Nil(t, err)
	require.NotNil(t, resp)

	_, err = handler.ScheduleTask(ctx, task, "")
	require.Equal(t, ErrTaskNotAllowMultipleExecutions, err)
}
