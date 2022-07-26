package executor

import (
	"context"
	"errors"
	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	taskMocks "github.com/alicenet/alicenet/layer1/executor/tasks/mocks"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
	"time"
)

func getTaskHandler(t *testing.T) (*Handler, *mocks.MockClient, *mocks.MockAllSmartContracts, *mocks.MockWatcher, accounts.Account) {
	db := mocks.NewTestDB()
	client := mocks.NewMockClient()
	client.GetFinalizedHeightFunc.SetDefaultReturn(0, nil)
	adminHandlers := mocks.NewMockAdminHandler()
	adminHandlers.IsSynchronizedFunc.SetDefaultReturn(true)
	txWatcher := mocks.NewMockWatcher()
	contracts := mocks.NewMockAllSmartContracts()

	account := accounts.Account{
		Address: common.HexToAddress("546F99F244b7B58B855330AE0E2BC1b30b41302F"),
		URL: accounts.URL{
			Scheme: "http",
			Path:   "",
		},
	}

	taskHandler, err := NewTaskHandler(db, client, contracts, adminHandlers, txWatcher)
	require.Nil(t, err)
	t.Cleanup(func() {
		taskHandler.Close()
		db.DB().Close()
	})

	taskHandler.Start()
	return taskHandler.(*Handler), client, contracts, txWatcher, account
}

// getTaskManagerCopy creates a copy of the manager from the DB without race
// conditions.
func getTaskManagerCopy(t *testing.T, manager *TaskManager) *TaskManager {
	newManager := &TaskManager{Schedule: make(map[string]ManagerRequestInfo), Responses: make(map[string]ManagerResponseInfo), marshaller: getTaskRegistry(), database: manager.database}
	<-time.After(10 * time.Millisecond)
	err := newManager.loadState()
	if err != nil {
		require.Equal(t, badger.ErrKeyNotFound, err)
	}
	return newManager
}

// getScheduleLen returns the amount of tasks is in the TaskManager.Schedule
func getScheduleLen(t *testing.T, manager *TaskManager) int {
	newManager := getTaskManagerCopy(t, manager)
	return len(newManager.Schedule)
}

func TestTasksScheduler_Schedule_NilTask(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t)
	ctx := context.Background()

	_, err := handler.ScheduleTask(ctx, nil, "")
	require.Equal(t, ErrTaskIsNil, err)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
}

func TestTasksScheduler_Schedule_NotRegistredTask(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t)
	ctx := context.Background()

	_, err := handler.ScheduleTask(ctx, taskMocks.NewMockTask(), "")
	require.Equal(t, ErrTaskTypeNotInRegistry, err)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
}

func TestTasksScheduler_Schedule_WrongStartDate(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t)
	ctx := context.Background()

	task := dkg.NewCompletionTask(2, 1)
	_, err := handler.ScheduleTask(ctx, task, "")
	require.Equal(t, ErrWrongParams, err)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
}

func TestTasksScheduler_Schedule_WrongEndDate(t *testing.T) {
	handler, client, _, _, _ := getTaskHandler(t)
	ctx := context.Background()
	client.GetFinalizedHeightFunc.SetDefaultReturn(12, nil)
	<-time.After(constants.TaskManagerProcessingTime)

	task := dkg.NewCompletionTask(2, 3)
	_, err := handler.ScheduleTask(ctx, task, "")
	require.Equal(t, ErrTaskExpired, err)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
}

func TestTasksScheduler_Schedule_MultiExecutionNotAllowed(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t)
	ctx := context.Background()

	task := dkg.NewCompletionTask(10, 40)
	resp, err := handler.ScheduleTask(ctx, task, "")
	require.Nil(t, err)
	require.NotNil(t, resp)

	_, err = handler.ScheduleTask(ctx, task, "")
	require.Equal(t, ErrTaskNotAllowMultipleExecutions, err)
	require.Equal(t, 1, getScheduleLen(t, handler.manager))
}

func TestTasksScheduler_KillById_EmptyId(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t)
	ctx := context.Background()

	_, err := handler.KillTaskById(ctx, "")
	require.Equal(t, ErrTaskIdEmpty, err)
}

func TestTasksScheduler_KillById_NotFound(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t)
	ctx := context.Background()

	_, err := handler.KillTaskById(ctx, "123")
	require.Equal(t, ErrNotScheduled, err)
}

func TestTasksScheduler_ScheduleAndKillById(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t)
	ctx := context.Background()

	task := dkg.NewCompletionTask(10, 40)
	taskId := uuid.New().String()
	resp, err := handler.ScheduleTask(ctx, task, taskId)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, getScheduleLen(t, handler.manager))

	_, err = handler.KillTaskById(ctx, taskId)
	require.Nil(t, err)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
}

func TestTasksScheduler_ScheduleAndKillById_RunningTask(t *testing.T) {
	handler, client, contracts, _, acc := getTaskHandler(t)
	client.GetFinalizedHeightFunc.SetDefaultReturn(12, nil)
	ctx := context.Background()

	dkgState := state.NewDkgState(acc)
	dkgState.OnRegistrationOpened(
		10,
		40,
		40,
		1,
	)
	publicKey := [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	dkgState.TransportPublicKey = publicKey

	err := state.SaveDkgState(handler.manager.database, dkgState)
	require.Nil(t, err)

	ethDkgMock := mocks.NewMockIETHDKG()
	ethDkgMock.RegisterFunc.SetDefaultReturn(nil, errors.New("network error"))
	ethDkgMock.GetNonceFunc.SetDefaultReturn(big.NewInt(1), nil)
	participantState := bindings.Participant{
		PublicKey: publicKey,
		Nonce:     uint64(1),
	}
	ethDkgMock.GetParticipantInternalStateFunc.SetDefaultReturn(participantState, nil)

	ethereumContracts := mocks.NewMockEthereumContracts()
	ethereumContracts.EthdkgFunc.SetDefaultReturn(ethDkgMock)
	contracts.EthereumContractsFunc.SetDefaultReturn(ethereumContracts)

	task := dkg.NewRegisterTask(10, 40)
	taskId := uuid.New().String()
	resp, err := handler.ScheduleTask(ctx, task, taskId)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, getScheduleLen(t, handler.manager))

	isRunning := false
	for !isRunning {
		select {
		case <-time.After(constants.TaskManagerProcessingTime):
			t.Fatal("didnt process task in time")
		default:
		}
		taskManagerCopy := getTaskManagerCopy(t, handler.manager)
		taskCopy := taskManagerCopy.Schedule[taskId]
		isRunning = taskCopy.InternalState == Running
	}

	_, err = handler.KillTaskById(ctx, taskId)
	require.Nil(t, err)

	for !resp.IsReady() {
		select {
		case <-time.After(constants.TaskManagerProcessingTime):
			t.Fatal("didnt process task in time")
		default:
		}
	}

	blockingResp := resp.GetResponseBlocking(ctx)
	require.NotNil(t, blockingResp)
	require.Equal(t, context.Canceled, blockingResp)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
}

func TestTasksScheduler_KillByType_Nil(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t)
	ctx := context.Background()

	_, err := handler.KillTaskByType(ctx, nil)
	require.Equal(t, ErrTaskIsNil, err)
}

func TestTasksScheduler_KillByType_NotInRegistry(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t)
	ctx := context.Background()

	_, err := handler.KillTaskByType(ctx, taskMocks.NewMockTask())
	require.Equal(t, ErrTaskTypeNotInRegistry, err)
}

func TestTasksScheduler_ScheduleAndKillByType(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t)
	ctx := context.Background()

	task1 := dkg.NewCompletionTask(10, 40)
	task1.AllowMultiExecution = true
	resp, err := handler.ScheduleTask(ctx, task1, "")
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, getScheduleLen(t, handler.manager))

	task2 := dkg.NewCompletionTask(10, 40)
	task2.AllowMultiExecution = true
	resp, err = handler.ScheduleTask(ctx, task1, "")
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 2, getScheduleLen(t, handler.manager))

	_, err = handler.KillTaskByType(ctx, &dkg.CompletionTask{})
	require.Nil(t, err)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
}

func TestTasksScheduler_ScheduleKillCloseAndRecover(t *testing.T) {
	handler, client, contracts, _, acc := getTaskHandler(t)
	client.GetFinalizedHeightFunc.SetDefaultReturn(12, nil)
	ctx := context.Background()

	dkgState := state.NewDkgState(acc)
	dkgState.OnRegistrationOpened(
		10,
		40,
		40,
		1,
	)
	publicKey := [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	dkgState.TransportPublicKey = publicKey

	err := state.SaveDkgState(handler.manager.database, dkgState)
	require.Nil(t, err)

	ethDkgMock := mocks.NewMockIETHDKG()
	ethDkgMock.RegisterFunc.SetDefaultReturn(nil, errors.New("network error"))
	ethDkgMock.GetNonceFunc.SetDefaultReturn(big.NewInt(1), nil)
	participantState := bindings.Participant{
		PublicKey: publicKey,
		Nonce:     uint64(1),
	}
	ethDkgMock.GetParticipantInternalStateFunc.SetDefaultReturn(participantState, nil)

	ethereumContracts := mocks.NewMockEthereumContracts()
	ethereumContracts.EthdkgFunc.SetDefaultReturn(ethDkgMock)
	contracts.EthereumContractsFunc.SetDefaultReturn(ethereumContracts)

	task := dkg.NewRegisterTask(10, 40)
	taskId := uuid.New().String()
	resp, err := handler.ScheduleTask(ctx, task, taskId)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, getScheduleLen(t, handler.manager))

	isRunning := false
	for !isRunning {
		select {
		case <-time.After(constants.TaskManagerProcessingTime):
			t.Fatal("didnt process task in time")
		default:
		}
		taskManagerCopy := getTaskManagerCopy(t, handler.manager)
		taskCopy := taskManagerCopy.Schedule[taskId]
		isRunning = taskCopy.InternalState == Running
	}

	handler.Close()
	newHandler, err := NewTaskHandler(handler.manager.database, handler.manager.eth, handler.manager.contracts, handler.manager.adminHandler, handler.manager.taskExecutor.txWatcher)
	require.Nil(t, err)
	newHandler.Start()

	resp, err = newHandler.ScheduleTask(ctx, task, taskId)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, getScheduleLen(t, handler.manager))

	_, err = handler.KillTaskById(ctx, taskId)
	require.Nil(t, err)

	for !resp.IsReady() {
		select {
		case <-time.After(constants.TaskManagerProcessingTime):
			t.Fatal("didnt process task in time")
		default:
		}
	}

	blockingResp := resp.GetResponseBlocking(ctx)
	require.NotNil(t, blockingResp)
	require.Equal(t, context.Canceled, blockingResp)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))

	newHandler.Close()
	newHandler2, err := NewTaskHandler(handler.manager.database, handler.manager.eth, handler.manager.contracts, handler.manager.adminHandler, handler.manager.taskExecutor.txWatcher)
	require.Nil(t, err)
	newHandler2.Start()

	resp, err = newHandler2.ScheduleTask(ctx, task, taskId)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
	require.NotNil(t, blockingResp)
	require.Equal(t, context.Canceled, blockingResp)
}
