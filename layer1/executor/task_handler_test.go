package executor

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	taskMocks "github.com/alicenet/alicenet/layer1/executor/tasks/mocks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/snapshots"
	snapshotState "github.com/alicenet/alicenet/layer1/executor/tasks/snapshots/state"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func getTaskHandler(t *testing.T, doCleanup bool) (*Handler, *mocks.MockClient, *mocks.MockAllSmartContracts, *mocks.MockWatcher, accounts.Account) {
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

	if doCleanup {
		t.Cleanup(func() {
			taskHandler.Close()
		})
	}

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

func TestTasksHandlerAndManager_Schedule_NilTask(t *testing.T) {
	t.Parallel()
	handler, _, _, _, _ := getTaskHandler(t, true)
	handler.Start()

	_, err := handler.ScheduleTask(nil, "")
	require.Equal(t, ErrTaskIsNil, err)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
}

func TestTasksHandlerAndManager_Schedule_NotRegisteredTask(t *testing.T) {
	t.Parallel()
	handler, _, _, _, _ := getTaskHandler(t, true)
	handler.Start()

	_, err := handler.ScheduleTask(taskMocks.NewMockTask(), "")
	require.Equal(t, ErrTaskTypeNotInRegistry, err)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
}

func TestTasksHandlerAndManager_Schedule_WrongStartDate(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t, true)
	handler.Start()

	task := dkg.NewCompletionTask(2, 1)
	_, err := handler.ScheduleTask(task, "")
	require.Equal(t, ErrWrongParams, err)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
}

func TestTasksHandlerAndManager_Schedule_WrongEndDate(t *testing.T) {
	handler, client, _, _, _ := getTaskHandler(t, true)
	client.GetFinalizedHeightFunc.SetDefaultReturn(12, nil)
	handler.Start()
	<-time.After(tasks.ManagerProcessingTime)

	task := dkg.NewCompletionTask(2, 3)
	_, err := handler.ScheduleTask(task, "")
	require.Equal(t, ErrTaskExpired, err)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
}

func TestTasksHandlerAndManager_Schedule_MultiExecutionNotAllowed(t *testing.T) {
	t.Parallel()
	handler, _, _, _, _ := getTaskHandler(t, true)
	handler.Start()

	task := dkg.NewCompletionTask(10, 40)
	resp, err := handler.ScheduleTask(task, "")
	require.Nil(t, err)
	require.NotNil(t, resp)

	_, err = handler.ScheduleTask(task, "")
	require.Equal(t, ErrTaskNotAllowMultipleExecutions, err)
	require.Equal(t, 1, getScheduleLen(t, handler.manager))
}

func TestTasksHandlerAndManager_KillById_EmptyId(t *testing.T) {
	t.Parallel()
	handler, _, _, _, _ := getTaskHandler(t, true)
	handler.Start()

	_, err := handler.KillTaskById("")
	require.Equal(t, ErrTaskIdEmpty, err)
}

func TestTasksHandlerAndManager_KillById_NotFound(t *testing.T) {
	t.Parallel()
	handler, _, _, _, _ := getTaskHandler(t, true)
	handler.Start()

	_, err := handler.KillTaskById("123")
	require.Equal(t, ErrNotScheduled, err)
}

func TestTasksHandlerAndManager_ScheduleAndKillById(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t, true)
	handler.Start()

	task := dkg.NewCompletionTask(10, 40)
	taskId := uuid.New().String()
	resp, err := handler.ScheduleTask(task, taskId)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, getScheduleLen(t, handler.manager))

	_, err = handler.KillTaskById(taskId)
	require.Nil(t, err)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
}

func TestTasksHandlerAndManager_ScheduleAndKillById_RunningTask(t *testing.T) {
	handler, client, contracts, _, acc := getTaskHandler(t, true)
	client.GetFinalizedHeightFunc.SetDefaultReturn(12, nil)
	handler.Start()
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
	resp, err := handler.ScheduleTask(task, taskId)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, getScheduleLen(t, handler.manager))

	isRunning := false
	failTime := time.After(tasks.ManagerProcessingTime)
	for !isRunning {
		select {
		case <-failTime:
			t.Fatal("didnt process task in time")
		default:
		}
		taskManagerCopy := getTaskManagerCopy(t, handler.manager)
		taskCopy := taskManagerCopy.Schedule[taskId]
		isRunning = taskCopy.InternalState == Running
	}

	_, err = handler.KillTaskById(taskId)
	require.Nil(t, err)

	failTime = time.After(tasks.ManagerProcessingTime)
	for !resp.IsReady() {
		select {
		case <-failTime:
			t.Fatal("didnt process task in time")
		default:
		}
	}

	blockingResp := resp.GetResponseBlocking(context.Background())
	require.NotNil(t, blockingResp)
	require.Equal(t, tasks.ErrTaskKilled, blockingResp)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
}

func TestTasksHandlerAndManager_KillByType_Nil(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t, true)
	handler.Start()

	_, err := handler.KillTaskByType(nil)
	require.Equal(t, ErrTaskIsNil, err)
}

func TestTasksHandlerAndManager_KillByType_NotInRegistry(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t, true)
	handler.Start()

	_, err := handler.KillTaskByType(taskMocks.NewMockTask())
	require.Equal(t, ErrTaskTypeNotInRegistry, err)
}

func TestTasksHandlerAndManager_ScheduleAndKillByType(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t, true)
	handler.Start()

	task1 := dkg.NewCompletionTask(10, 40)
	task1.AllowMultiExecution = true
	resp, err := handler.ScheduleTask(task1, "")
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, getScheduleLen(t, handler.manager))

	task2 := dkg.NewCompletionTask(10, 40)
	task2.AllowMultiExecution = true
	resp, err = handler.ScheduleTask(task1, "")
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 2, getScheduleLen(t, handler.manager))

	_, err = handler.KillTaskByType(&dkg.CompletionTask{})
	require.Nil(t, err)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
}

func TestTasksHandlerAndManager_ScheduleKillCloseAndRecover(t *testing.T) {
	handler, client, contracts, _, acc := getTaskHandler(t, false)
	handler.Start()
	client.GetFinalizedHeightFunc.SetDefaultReturn(12, nil)

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
	task.AllowMultiExecution = true
	task.SubscribeOptions = &transaction.SubscribeOptions{
		EnableAutoRetry: true,
		MaxStaleBlocks:  14,
	}
	taskId := uuid.New().String()
	resp, err := handler.ScheduleTask(task, taskId)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, getScheduleLen(t, handler.manager))

	isRunning := false
	failTime := time.After(tasks.ManagerProcessingTime)
	for !isRunning {
		select {
		case <-failTime:
			t.Fatal("didnt process task in time")
		default:
		}
		taskManagerCopy := getTaskManagerCopy(t, handler.manager)
		taskCopy := taskManagerCopy.Schedule[taskId]
		isRunning = taskCopy.InternalState == Running
	}

	handler.Close()
	newHandler, err := NewTaskHandler(handler.manager.database, handler.manager.eth, handler.manager.contracts, handler.manager.adminHandler, handler.manager.taskExecutor.txWatcher)
	recoveredTask := newHandler.(*Handler).manager.Schedule[taskId]
	require.Equal(t, task.ID, recoveredTask.Id)
	require.Equal(t, task.Name, recoveredTask.Name)
	require.Equal(t, task.Start, recoveredTask.Start)
	require.Equal(t, task.End, recoveredTask.End)
	require.Equal(t, task.AllowMultiExecution, recoveredTask.AllowMultiExecution)
	require.Equal(t, task.SubscribeOptions.MaxStaleBlocks, recoveredTask.SubscribeOptions.MaxStaleBlocks)
	require.Equal(t, task.SubscribeOptions.EnableAutoRetry, recoveredTask.SubscribeOptions.EnableAutoRetry)
	require.Nil(t, err)
	newHandler.Start()

	resp, err = newHandler.ScheduleTask(task, taskId)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, getScheduleLen(t, newHandler.(*Handler).manager))

	_, err = newHandler.KillTaskById(taskId)
	require.Nil(t, err)

	failTime = time.After(tasks.ManagerProcessingTime)
	for !resp.IsReady() {
		select {
		case <-failTime:
			t.Fatal("didnt process task in time")
		default:
		}
	}

	blockingResp := resp.GetResponseBlocking(context.Background())
	require.NotNil(t, blockingResp)
	require.Equal(t, ErrTaskKilledBeforeExecution, blockingResp)
	require.Equal(t, 0, getScheduleLen(t, newHandler.(*Handler).manager))

	newHandler.Close()
	newHandler2, err := NewTaskHandler(newHandler.(*Handler).manager.database, newHandler.(*Handler).manager.eth, newHandler.(*Handler).manager.contracts, newHandler.(*Handler).manager.adminHandler, newHandler.(*Handler).manager.taskExecutor.txWatcher)
	require.Nil(t, err)
	newHandler2.Start()

	resp, err = newHandler2.ScheduleTask(task, taskId)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 0, getScheduleLen(t, newHandler2.(*Handler).manager))

	failTime = time.After(tasks.ManagerProcessingTime)
	for !resp.IsReady() {
		select {
		case <-failTime:
			t.Fatal("didnt process task in time")
		default:
		}
	}

	require.NotNil(t, blockingResp)
	require.Equal(t, ErrTaskKilledBeforeExecution, blockingResp)

	newHandler2.Close()
	handler.manager.database.DB().Close()
}

func TestTasksHandlerAndManager_ScheduleAndRecover_RunningSnapshotTask(t *testing.T) {
	handler, client, contracts, _, acc := getTaskHandler(t, false)
	client.GetFinalizedHeightFunc.SetDefaultReturn(12, nil)
	handler.Start()

	bh := &objs.BlockHeader{
		BClaims: &objs.BClaims{
			ChainID:    1337,
			Height:     1,
			TxCount:    0,
			PrevBlock:  crypto.Hasher([]byte("")),
			TxRoot:     crypto.Hasher([]byte("")),
			StateRoot:  crypto.Hasher([]byte("")),
			HeaderRoot: crypto.Hasher([]byte("")),
		},
		TxHshLst: make([][]byte, 0),
		GroupKey: make([]byte, 0),
		SigGroup: make([]byte, 0),
	}
	ssState := &snapshotState.SnapshotState{
		Account:     acc,
		BlockHeader: bh,
	}
	err := snapshotState.SaveSnapshotState(handler.manager.database, ssState)
	require.Nil(t, err)

	ssContracts := mocks.NewMockISnapshots()
	ssContracts.GetCommittedHeightFromLatestSnapshotFunc.SetDefaultReturn(big.NewInt(0), nil)
	ssContracts.GetSnapshotDesperationFactorFunc.SetDefaultReturn(big.NewInt(40), nil)
	ssContracts.GetSnapshotDesperationDelayFunc.SetDefaultReturn(big.NewInt(10), nil)
	ssContracts.SnapshotFunc.SetDefaultReturn(nil, errors.New("network error"))
	ssContracts.GetAliceNetHeightFromLatestSnapshotFunc.SetDefaultReturn(big.NewInt(0), nil)

	ethereumContracts := mocks.NewMockEthereumContracts()
	ethereumContracts.SnapshotsFunc.SetDefaultReturn(ssContracts)
	contracts.EthereumContractsFunc.SetDefaultReturn(ethereumContracts)

	task := snapshots.NewSnapshotTask(0, 5, 1)
	taskId := uuid.New().String()
	resp, err := handler.ScheduleTask(task, taskId)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, getScheduleLen(t, handler.manager))

	isRunning := false
	failTime := time.After(tasks.ManagerProcessingTime)
	for !isRunning {
		select {
		case <-failTime:
			t.Fatal("didnt process task in time")
		default:
		}
		taskManagerCopy := getTaskManagerCopy(t, handler.manager)
		taskCopy := taskManagerCopy.Schedule[taskId]
		isRunning = taskCopy.InternalState == Running
	}

	handler.Close()
	newHandler, err := NewTaskHandler(handler.manager.database, handler.manager.eth, handler.manager.contracts, handler.manager.adminHandler, handler.manager.taskExecutor.txWatcher)
	recoveredTask := newHandler.(*Handler).manager.Schedule[taskId]
	require.Equal(t, task.ID, recoveredTask.Id)
	require.Equal(t, task.Name, recoveredTask.Name)
	require.Equal(t, task.Start, recoveredTask.Start)
	require.Equal(t, task.End, recoveredTask.End)
	require.Equal(t, task.AllowMultiExecution, recoveredTask.AllowMultiExecution)
	require.Equal(t, task.NumOfValidators, recoveredTask.Task.(*snapshots.SnapshotTask).NumOfValidators)
	require.Equal(t, task.ValidatorIndex, recoveredTask.Task.(*snapshots.SnapshotTask).ValidatorIndex)
	require.Equal(t, task.Height, recoveredTask.Task.(*snapshots.SnapshotTask).Height)
	require.Nil(t, err)
	newHandler.Start()

	isRunning = false
	failTime = time.After(tasks.ManagerProcessingTime)
	for !isRunning {
		select {
		case <-failTime:
			t.Fatal("didnt process task in time")
		default:
		}
		taskManagerCopy := getTaskManagerCopy(t, newHandler.(*Handler).manager)
		taskCopy := taskManagerCopy.Schedule[taskId]
		isRunning = taskCopy.InternalState == Running
	}

	newHandler.Close()
}

func TestHandlerManagerAndExecutor_ErrorOnCreation(t *testing.T) {
	t.Parallel()
	db := mocks.NewTestDB()
	client := mocks.NewMockClient()
	adminHandlers := mocks.NewMockAdminHandler()
	txWatcher := mocks.NewMockWatcher()
	contracts := mocks.NewMockAllSmartContracts()

	err := db.Update(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskExecutorState()
		if err := utils.SetValue(txn, key, []byte("corrupted data")); err != nil {
			return err
		}
		return nil
	})
	require.Nil(t, err)

	taskHandler, err := NewTaskHandler(db, client, contracts, adminHandlers, txWatcher)

	require.Nil(t, taskHandler)
	require.NotNil(t, err)
}

func TestTasksHandlerAndManager_SendNilResponseAndCloseRequestChan(t *testing.T) {
	handler, _, _, _, _ := getTaskHandler(t, true)
	handler.Start()

	task1 := dkg.NewCompletionTask(10, 40)
	req := managerRequest{task: task1, id: "task-id", action: Schedule, response: nil}
	handler.requestChannel <- req
	<-time.After(100 * time.Millisecond)
	require.Equal(t, 0, getScheduleLen(t, handler.manager))
}
