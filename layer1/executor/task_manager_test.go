package executor

// import (
// 	"io/ioutil"
// 	"os"
// 	"testing"
// 	"time"

// 	"github.com/alicenet/alicenet/consensus/db"
// 	"github.com/alicenet/alicenet/constants"
// 	"github.com/alicenet/alicenet/layer1/executor/tasks"
// 	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
// 	"github.com/alicenet/alicenet/layer1/transaction"
// 	"github.com/alicenet/alicenet/test/mocks"
// 	"github.com/dgraph-io/badger/v2"
// 	"github.com/stretchr/testify/assert"
// )

// func getTaskScheduler(t *testing.T) (*TaskManager, chan tasks.TaskRequest, *mocks.MockClient) {
// 	db := mocks.NewTestDB()
// 	client := mocks.NewMockClient()
// 	adminHandlers := mocks.NewMockAdminHandler()
// 	txWatcher := transaction.NewWatcher(client, 12, db, false, constants.TxPollingTime)
// 	requestChan := make(chan tasks.TaskRequest, constants.TaskSchedulerBufferSize)
// 	tasksScheduler, err := NewTaskHandler(db, client, adminHandlers, requestChan, txWatcher)
// 	assert.Nil(t, err)
// 	t.Cleanup(func() {
// 		txWatcher.Close()
// 		tasksScheduler.Close()
// 		close(requestChan)
// 		db.DB().Close()
// 	})
// 	return tasksScheduler, requestChan, client
// }

// // Auxiliary function to get how many tasks we have inside the scheduler. This
// // function creates a copy of the scheduler to get the len without race
// // conditions.
// func getScheduleLen(t *testing.T, scheduler *TaskManager) int {
// 	newScheduler := &TaskManager{Schedule: make(map[string]ManagerRequestInfo), marshaller: getTaskRegistry(), database: scheduler.database}
// 	err := scheduler.persistState()
// 	assert.Nil(t, err)
// 	err = newScheduler.loadState()
// 	assert.Nil(t, err)
// 	return len(newScheduler.Schedule)
// }

// func TestTasksScheduler_Schedule_NilTask(t *testing.T) {
// 	scheduler, tasksChan, _ := getTaskScheduler(t)
// 	err := scheduler.start()
// 	assert.Nil(t, err)

// 	request := tasks.NewScheduleTaskRequest(nil)
// 	tasksChan <- request

// 	<-time.After(10 * time.Millisecond)
// 	assert.Emptyf(t, scheduler.Schedule, "Expected zero tasks scheduled")
// }

// func TestTasksScheduler_Schedule_WrongStartDate(t *testing.T) {

// 	scheduler, tasksChan, _ := getTaskScheduler(t)
// 	err := scheduler.start()
// 	assert.Nil(t, err)

// 	task := dkg.NewCompletionTask(2, 1)
// 	request := tasks.NewScheduleTaskRequest(task)
// 	tasksChan <- request

// 	scheduler.Close()
// 	<-time.After(constants.TaskSchedulerProcessingTime + 10*time.Millisecond)

// 	assert.Equalf(t, 0, getScheduleLen(t, scheduler), "Expected zero tasks scheduled")
// }

// func TestTasksScheduler_Schedule_WrongEndDate(t *testing.T) {

// 	scheduler, tasksChan, client := getTaskScheduler(t)
// 	client.GetFinalizedHeightFunc.SetDefaultReturn(12, nil)
// 	err := scheduler.start()
// 	assert.Nil(t, err)

// 	task := dkg.NewCompletionTask(2, 3)
// 	request := tasks.NewScheduleTaskRequest(task)
// 	tasksChan <- request

// 	time.After(20 * time.Millisecond)
// 	assert.Equalf(t, 0, getScheduleLen(t, scheduler), "Expected zero tasks scheduled")
// }

// func TestTasksScheduler_ScheduleAndKillTasks_Success(t *testing.T) {

// 	scheduler, tasksChan, _ := getTaskScheduler(t)
// 	err := scheduler.start()
// 	assert.Nil(t, err)

// 	completionTask := dkg.NewCompletionTask(2, 3)
// 	request := tasks.NewScheduleTaskRequest(completionTask)
// 	tasksChan <- request

// 	completionTask2 := dkg.NewCompletionTask(3, 4)
// 	request = tasks.NewScheduleTaskRequest(completionTask2)
// 	tasksChan <- request

// 	registerTask := dkg.NewRegisterTask(2, 5)
// 	request = tasks.NewScheduleTaskRequest(registerTask)
// 	tasksChan <- request

// 	<-time.After(10 * time.Millisecond)
// 	assert.Equalf(t, 3, getScheduleLen(t, scheduler), "Expected 3 task scheduled")

// 	request = tasks.NewKillTaskRequest(&dkg.CompletionTask{})
// 	tasksChan <- request
// 	time.After(10 * time.Millisecond)
// 	assert.Equalf(t, 1, getScheduleLen(t, scheduler), "Expected 1 task after Completion tasks have been killed")

// 	request = tasks.NewKillTaskRequest(&dkg.DisputeMissingGPKjTask{})
// 	tasksChan <- request

// 	<-time.After(10 * time.Millisecond)
// 	assert.Equalf(t, 1, getScheduleLen(t, scheduler), "There should be 1 tasks left still, due there were no DisputeMissing task scheduled")

// 	request = tasks.NewKillTaskRequest(&dkg.RegisterTask{})
// 	tasksChan <- request

// 	<-time.After(10 * time.Millisecond)
// 	assert.Equalf(t, 0, getScheduleLen(t, scheduler), "All the tasks should have been removed")
// }

// func TestTasksScheduler_ScheduleRunAndKillTask_Success(t *testing.T) {

// 	scheduler, tasksChan, client := getTaskScheduler(t)
// 	err := scheduler.start()
// 	assert.Nil(t, err)

// 	completionTask := dkg.NewCompletionTask(1, 40)
// 	request := tasks.NewScheduleTaskRequest(completionTask)
// 	tasksChan <- request

// 	scheduler.Close()
// 	<-time.After(constants.TaskSchedulerProcessingTime + 1500*time.Millisecond)
// 	client.GetFinalizedHeightFunc.SetDefaultReturn(10, nil)
// 	scheduler, err = NewTaskHandler(scheduler.database, client, scheduler.adminHandler, tasksChan, scheduler.txWatcher)
// 	assert.Nil(t, err)
// 	err = scheduler.start()
// 	<-time.After(constants.TaskSchedulerProcessingTime + 10*time.Millisecond)

// 	assert.Equalf(t, 0, getScheduleLen(t, scheduler), "All the tasks should have been removed")
// }

// func TestTasksScheduler_ScheduleDuplicatedTask_Success(t *testing.T) {

// 	scheduler, tasksChan, client := getTaskScheduler(t)
// 	err := scheduler.start()
// 	assert.Nil(t, err)

// 	completionTask := dkg.NewCompletionTask(1, 40)
// 	request := tasks.NewScheduleTaskRequest(completionTask)
// 	tasksChan <- request
// 	completionTask2 := dkg.NewCompletionTask(1, 40)
// 	request = tasks.NewScheduleTaskRequest(completionTask2)
// 	tasksChan <- request

// 	scheduler.Close()
// 	<-time.After(constants.TaskSchedulerProcessingTime + 10*time.Millisecond)

// 	client.GetFinalizedHeightFunc.SetDefaultReturn(10, nil)
// 	scheduler, err = NewTaskHandler(scheduler.database, client, scheduler.adminHandler, tasksChan, scheduler.txWatcher)
// 	assert.Nil(t, err)
// 	err = scheduler.start()
// 	<-time.After(constants.TaskSchedulerProcessingTime + 10*time.Millisecond)

// 	assert.Equalf(t, 1, getScheduleLen(t, scheduler), "Expected to have 1 task")
// 	for _, task := range scheduler.Schedule {
// 		assert.NotEqualf(t, task.InternalState, Running, "this task shouldn't be running due to duplication")
// 	}
// }

// func TestTasksScheduler_ScheduleAndKillExpiredAndUnresponsiveTasks_Success(t *testing.T) {

// 	scheduler, tasksChan, client := getTaskScheduler(t)
// 	err := scheduler.start()
// 	assert.Nil(t, err)

// 	completionTask := dkg.NewCompletionTask(50, 90)
// 	request := tasks.NewScheduleTaskRequest(completionTask)
// 	tasksChan <- request
// 	completionTask2 := dkg.NewCompletionTask(1, 10)
// 	request = tasks.NewScheduleTaskRequest(completionTask2)
// 	tasksChan <- request
// 	completionTask3 := dkg.NewCompletionTask(110, 150)
// 	request = tasks.NewScheduleTaskRequest(completionTask3)
// 	tasksChan <- request

// 	scheduler.Close()
// 	<-time.After(constants.TaskSchedulerProcessingTime + 1500*time.Millisecond)
// 	client.GetFinalizedHeightFunc.SetDefaultReturn(100, nil)
// 	// trick to set client mock parameters without having race conditions
// 	scheduler, err = NewTaskHandler(scheduler.database, client, scheduler.adminHandler, tasksChan, scheduler.txWatcher)
// 	assert.Nil(t, err)
// 	err = scheduler.start()
// 	<-time.After(constants.TaskSchedulerProcessingTime + 10*time.Millisecond)

// 	assert.Equalf(t, 1, getScheduleLen(t, scheduler), "Expected to have 1 task")
// }

// func TestTasksScheduler_Recovery_Success(t *testing.T) {
// 	dir, err := ioutil.TempDir("", "db-test")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer func() {
// 		if err := os.RemoveAll(dir); err != nil {
// 			t.Fatal(err)
// 		}
// 	}()
// 	opts := badger.DefaultOptions(dir)
// 	rawDB, err := badger.Open(opts)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer rawDB.Close()

// 	db := &db.Database{}
// 	db.Init(rawDB)

// 	client := mocks.NewMockClient()
// 	adminHandlers := mocks.NewMockAdminHandler()
// 	txWatcher := transaction.NewWatcher(client, 12, db, false, constants.TxPollingTime)
// 	tasksChan := make(chan tasks.TaskRequest, constants.TaskSchedulerBufferSize)
// 	scheduler, err := NewTaskHandler(db, client, adminHandlers, tasksChan, txWatcher)
// 	assert.Nil(t, err)
// 	err = scheduler.start()

// 	completionTask := dkg.NewCompletionTask(50, 90)
// 	request := tasks.NewScheduleTaskRequest(completionTask)
// 	tasksChan <- request
// 	completionTask2 := dkg.NewCompletionTask(1, 10)
// 	request2 := tasks.NewScheduleTaskRequest(completionTask2)
// 	tasksChan <- request2
// 	completionTask3 := dkg.NewCompletionTask(110, 150)
// 	request3 := tasks.NewScheduleTaskRequest(completionTask3)
// 	tasksChan <- request3

// 	<-time.After(10 * time.Millisecond)

// 	assert.Equalf(t, 3, getScheduleLen(t, scheduler), "Expected to have 3 tasks")

// 	scheduler.Close()
// 	close(tasksChan)

// 	tasksChan = make(chan tasks.TaskRequest, constants.TaskSchedulerBufferSize)
// 	scheduler, err = NewTaskHandler(db, client, adminHandlers, tasksChan, txWatcher)
// 	assert.Nil(t, err)
// 	err = scheduler.start()
// 	assert.Nil(t, err)
// 	assert.Equalf(t, 3, getScheduleLen(t, scheduler), "Expected to have 3 tasks")

// 	scheduler.Close()

// 	<-time.After(10 * time.Millisecond)
// 	close(tasksChan)
// }

///////////////////
//import (
//	"context"
//	"errors"
//	"io/ioutil"
//	"math/big"
//	"os"
//	"testing"
//
//	"github.com/alicenet/alicenet/consensus/db"
//	"github.com/alicenet/alicenet/layer1/executor/tasks"
//	"github.com/alicenet/alicenet/layer1/transaction"
//	"github.com/alicenet/alicenet/logging"
//	"github.com/alicenet/alicenet/test/mocks"
//	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
//	"github.com/dgraph-io/badger/v2"
//	"github.com/ethereum/go-ethereum/accounts"
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/core/types"
//	"github.com/stretchr/testify/assert"
//)
//
//func getTaskManager(t *testing.T) (*TasksManager, *mocks.MockClient, *db.Database, *taskResponseChan, *mocks.MockWatcher, *mocks.MockAllSmartContracts) {
//	db := mocks.NewTestDB()
//	client := mocks.NewMockClient()
//	client.ExtractTransactionSenderFunc.SetDefaultReturn(common.Address{}, nil)
//	client.GetTxMaxStaleBlocksFunc.SetDefaultReturn(10)
//	hdr := &types.Header{
//		Number: big.NewInt(1),
//	}
//	client.GetHeaderByNumberFunc.SetDefaultReturn(hdr, nil)
//	client.GetBlockBaseFeeAndSuggestedGasTipFunc.SetDefaultReturn(big.NewInt(100), big.NewInt(1), nil)
//	client.GetDefaultAccountFunc.SetDefaultReturn(accounts.Account{Address: common.Address{}})
//	client.GetTransactionByHashFunc.SetDefaultReturn(nil, false, nil)
//	client.GetFinalityDelayFunc.SetDefaultReturn(10)
//
//	logger := logging.GetLogger("test")
//
//	txWatcher := mocks.NewMockWatcher()
//	contracts := mocks.NewMockAllSmartContracts()
//
//	taskManager, err := NewTaskManager(txWatcher, db, logger.WithField("Component", "schedule"))
//	assert.Nil(t, err)
//
//	taskRespChan := &taskResponseChan{trChan: make(chan tasks.TaskResponse, 100)}
//	return taskManager, client, db, taskRespChan, txWatcher, contracts
//}
//
//func Test_TaskExecutor_HappyPath(t *testing.T) {
//	manager, client, db, taskRespChan, txWatcher, contracts := getTaskManager(t)
//	defer taskRespChan.close()
//
//	receipt := &types.Receipt{
//		Status:      types.ReceiptStatusSuccessful,
//		BlockNumber: big.NewInt(20),
//	}
//
//	receiptResponse := mocks.NewMockReceiptResponse()
//	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
//	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)
//
//	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)
//
//	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)
//
//	task := mocks.NewMockTask()
//	task.PrepareFunc.SetDefaultReturn(nil)
//	txn := types.NewTx(&types.LegacyTx{
//		Nonce:    1,
//		Value:    big.NewInt(1),
//		Gas:      1,
//		GasPrice: big.NewInt(1),
//		Data:     []byte{52, 66, 175, 92},
//	})
//	task.ExecuteFunc.SetDefaultReturn(txn, nil)
//	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
//	task.GetLoggerFunc.SetDefaultReturn(manager.logger)
//
//	mainCtx := context.Background()
//	manager.ManageTask(mainCtx, task, "", "123", db, manager.logger, client, contracts, taskRespChan)
//
//	mockrequire.CalledOnce(t, task.PrepareFunc)
//	mockrequire.CalledOnce(t, task.ExecuteFunc)
//	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(nil))
//	assert.Emptyf(t, manager.TxsBackup, "Expected transactions to be empty")
//}
//
//func Test_TaskManager_TaskErrorRecoverable(t *testing.T) {
//	manager, client, db, taskRespChan, txWatcher, contracts := getTaskManager(t)
//	defer taskRespChan.close()
//
//	receipt := &types.Receipt{
//		Status:      types.ReceiptStatusSuccessful,
//		BlockNumber: big.NewInt(20),
//	}
//
//	receiptResponse := mocks.NewMockReceiptResponse()
//	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
//	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)
//
//	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)
//
//	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)
//
//	taskErr := tasks.NewTaskErr("Recoverable error", true)
//	task := mocks.NewMockTask()
//	task.PrepareFunc.PushReturn(taskErr)
//	task.PrepareFunc.PushReturn(nil)
//	txn := types.NewTx(&types.LegacyTx{
//		Nonce:    1,
//		Value:    big.NewInt(1),
//		Gas:      1,
//		GasPrice: big.NewInt(1),
//		Data:     []byte{52, 66, 175, 92},
//	})
//	task.ExecuteFunc.SetDefaultReturn(txn, nil)
//	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
//	task.GetLoggerFunc.SetDefaultReturn(manager.logger)
//
//	mainCtx := context.Background()
//	manager.ManageTask(mainCtx, task, "", "123", db, manager.logger, client, contracts, taskRespChan)
//
//	mockrequire.CalledN(t, task.PrepareFunc, 2)
//	mockrequire.CalledOnce(t, task.ExecuteFunc)
//	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(nil))
//	assert.Emptyf(t, manager.TxsBackup, "Expected transactions to be empty")
//}
//
//func Test_TaskManager_UnrecoverableError(t *testing.T) {
//	manager, client, db, taskRespChan, txWatcher, contracts := getTaskManager(t)
//	defer taskRespChan.close()
//
//	receipt := &types.Receipt{
//		Status:      types.ReceiptStatusSuccessful,
//		BlockNumber: big.NewInt(20),
//	}
//
//	receiptResponse := mocks.NewMockReceiptResponse()
//	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
//	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)
//
//	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)
//
//	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)
//
//	task := mocks.NewMockTask()
//	taskErr := tasks.NewTaskErr("Unrecoverable error", false)
//
//	task.PrepareFunc.SetDefaultReturn(taskErr)
//	mainCtx := context.Background()
//	manager.ManageTask(mainCtx, task, "", "123", db, manager.logger, client, contracts, taskRespChan)
//
//	mockrequire.CalledOnce(t, task.PrepareFunc)
//	mockrequire.NotCalled(t, task.ExecuteFunc)
//	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(taskErr))
//	assert.Emptyf(t, manager.TxsBackup, "Expected transactions to be empty")
//}
//
//func Test_TaskManager_TaskInTasksManagerTransactions(t *testing.T) {
//	manager, client, db, taskRespChan, txWatcher, contracts := getTaskManager(t)
//	defer taskRespChan.close()
//
//	receipt := &types.Receipt{
//		Status:      types.ReceiptStatusSuccessful,
//		BlockNumber: big.NewInt(20),
//	}
//
//	receiptResponse := mocks.NewMockReceiptResponse()
//	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
//	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)
//
//	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)
//
//	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)
//
//	task := mocks.NewMockTask()
//	task.PrepareFunc.SetDefaultReturn(nil)
//	txn := types.NewTx(&types.LegacyTx{
//		Nonce:    1,
//		Value:    big.NewInt(1),
//		Gas:      1,
//		GasPrice: big.NewInt(1),
//		Data:     []byte{52, 66, 175, 92},
//	})
//	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
//	task.GetLoggerFunc.SetDefaultReturn(manager.logger)
//	taskId := "123"
//	task.GetIdFunc.SetDefaultReturn(taskId)
//
//	mainCtx := context.Background()
//	manager.TxsBackup[task.GetId()] = txn
//	manager.ManageTask(mainCtx, task, "", taskId, db, manager.logger, client, contracts, taskRespChan)
//
//	mockrequire.NotCalled(t, task.PrepareFunc)
//	mockrequire.NotCalled(t, task.ExecuteFunc)
//	assert.Emptyf(t, manager.TxsBackup, "Expected transactions to be empty")
//}
//
//func Test_TaskManager_ExecuteWithErrors(t *testing.T) {
//	manager, client, db, taskRespChan, txWatcher, contracts := getTaskManager(t)
//	defer taskRespChan.close()
//
//	receipt := &types.Receipt{
//		Status:      types.ReceiptStatusSuccessful,
//		BlockNumber: big.NewInt(20),
//	}
//
//	receiptResponse := mocks.NewMockReceiptResponse()
//	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
//	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)
//
//	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)
//
//	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)
//
//	task := mocks.NewMockTask()
//	task.PrepareFunc.SetDefaultReturn(nil)
//	task.ExecuteFunc.PushReturn(nil, tasks.NewTaskErr("Recoverable error", true))
//	unrecoverableError := tasks.NewTaskErr("Unrecoverable error", false)
//	task.ExecuteFunc.PushReturn(nil, unrecoverableError)
//	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
//	task.GetLoggerFunc.SetDefaultReturn(manager.logger)
//
//	mainCtx := context.Background()
//	manager.ManageTask(mainCtx, task, "", "123", db, manager.logger, client, contracts, taskRespChan)
//
//	mockrequire.CalledOnce(t, task.PrepareFunc)
//	mockrequire.CalledN(t, task.ExecuteFunc, 2)
//	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(unrecoverableError))
//	assert.Emptyf(t, manager.TxsBackup, "Expected transactions to be empty")
//}
//
//func Test_TaskManager_ReceiptWithErrorAndFailure(t *testing.T) {
//	manager, client, db, taskRespChan, txWatcher, contracts := getTaskManager(t)
//	defer taskRespChan.close()
//
//	receipt := &types.Receipt{
//		Status:      types.ReceiptStatusFailed,
//		BlockNumber: big.NewInt(20),
//	}
//
//	receiptResponse := mocks.NewMockReceiptResponse()
//
//	task := mocks.NewMockTask()
//	task.PrepareFunc.SetDefaultReturn(nil)
//	txn := types.NewTx(&types.LegacyTx{
//		Nonce:    1,
//		Value:    big.NewInt(1),
//		Gas:      1,
//		GasPrice: big.NewInt(1),
//		Data:     []byte{52, 66, 175, 92},
//	})
//	task.ShouldExecuteFunc.PushReturn(true, nil)
//	task.ExecuteFunc.SetDefaultReturn(txn, nil)
//	task.GetLoggerFunc.SetDefaultReturn(manager.logger)
//
//	receiptResponse.IsReadyFunc.PushReturn(false)
//	task.ShouldExecuteFunc.PushReturn(true, nil)
//
//	receiptResponse.IsReadyFunc.PushReturn(true)
//	receiptResponse.GetReceiptBlockingFunc.PushReturn(nil, errors.New("got error while getting receipt"))
//	task.ShouldExecuteFunc.PushReturn(true, nil)
//	txWatcher.SubscribeFunc.PushReturn(receiptResponse, nil)
//
//	receiptResponse.IsReadyFunc.PushReturn(true)
//	receiptResponse.GetReceiptBlockingFunc.PushReturn(receipt, nil)
//	task.ShouldExecuteFunc.PushReturn(true, nil)
//	txWatcher.SubscribeFunc.PushReturn(receiptResponse, nil)
//
//	receiptResponse.IsReadyFunc.PushReturn(false)
//	txWatcher.SubscribeFunc.PushReturn(receiptResponse, nil)
//	task.ShouldExecuteFunc.PushReturn(false, nil)
//
//	mainCtx := context.Background()
//	manager.ManageTask(mainCtx, task, "", "123", db, manager.logger, client, contracts, taskRespChan)
//
//	mockrequire.CalledOnce(t, task.PrepareFunc)
//	mockrequire.CalledN(t, task.ExecuteFunc, 3)
//	mockrequire.CalledN(t, receiptResponse.IsReadyFunc, 4)
//	mockrequire.CalledN(t, receiptResponse.GetReceiptBlockingFunc, 2)
//	mockrequire.CalledN(t, task.ShouldExecuteFunc, 5)
//
//	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(nil))
//	assert.Emptyf(t, manager.TxsBackup, "Expected transactions to be empty")
//}
//
//func Test_TaskManager_RecoveringTaskManager(t *testing.T) {
//	dir, err := ioutil.TempDir("", "db-test")
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer func() {
//		if err := os.RemoveAll(dir); err != nil {
//			t.Fatal(err)
//		}
//	}()
//	opts := badger.DefaultOptions(dir)
//	rawDB, err := badger.Open(opts)
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer rawDB.Close()
//
//	db := &db.Database{}
//	db.Init(rawDB)
//
//	client := mocks.NewMockClient()
//	client.ExtractTransactionSenderFunc.SetDefaultReturn(common.Address{}, nil)
//	client.GetTxMaxStaleBlocksFunc.SetDefaultReturn(10)
//	hdr := &types.Header{
//		Number: big.NewInt(1),
//	}
//	client.GetHeaderByNumberFunc.SetDefaultReturn(hdr, nil)
//	client.GetBlockBaseFeeAndSuggestedGasTipFunc.SetDefaultReturn(big.NewInt(100), big.NewInt(1), nil)
//	client.GetDefaultAccountFunc.SetDefaultReturn(accounts.Account{Address: common.Address{}})
//	client.GetTransactionByHashFunc.SetDefaultReturn(nil, false, nil)
//	client.GetFinalityDelayFunc.SetDefaultReturn(10)
//
//	logger := logging.GetLogger("test")
//
//	txWatcher := mocks.NewMockWatcher()
//	manager, err := NewTaskManager(txWatcher, db, logger.WithField("Component", "schedule"))
//	assert.Nil(t, err)
//
//	taskRespChan := &taskResponseChan{trChan: make(chan tasks.TaskResponse, 100)}
//	defer taskRespChan.close()
//
//	receipt := &types.Receipt{
//		Status:      types.ReceiptStatusSuccessful,
//		BlockNumber: big.NewInt(20),
//	}
//
//	receiptResponse := mocks.NewMockReceiptResponse()
//	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
//	receiptResponse.GetReceiptBlockingFunc.PushReturn(nil, &transaction.ErrTransactionStale{})
//	receiptResponse.GetReceiptBlockingFunc.PushReturn(receipt, nil)
//
//	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)
//
//	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)
//
//	contracts := mocks.NewMockAllSmartContracts()
//
//	task := mocks.NewMockTask()
//	task.PrepareFunc.SetDefaultReturn(nil)
//	txn := types.NewTx(&types.LegacyTx{
//		Nonce:    1,
//		Value:    big.NewInt(1),
//		Gas:      1,
//		GasPrice: big.NewInt(1),
//		Data:     []byte{52, 66, 175, 92},
//	})
//	task.ExecuteFunc.SetDefaultReturn(txn, nil)
//	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
//	task.GetLoggerFunc.SetDefaultReturn(manager.logger)
//
//	mainCtx := context.Background()
//	manager.ManageTask(mainCtx, task, "", "123", db, manager.logger, client, contracts, taskRespChan)
//
//	assert.Equalf(t, 1, len(manager.TxsBackup), "Expected one transaction (stale status)")
//	manager, err = NewTaskManager(txWatcher, db, logger.WithField("Component", "schedule"))
//	assert.Nil(t, err)
//	manager.ManageTask(mainCtx, task, "", "123", db, manager.logger, client, contracts, taskRespChan)
//
//	mockrequire.CalledOnce(t, task.PrepareFunc)
//	mockrequire.CalledOnce(t, task.ExecuteFunc)
//	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(nil))
//	assert.Emptyf(t, manager.TxsBackup, "Expected transactions to be empty")
//
//}
