package executor

import (
	"context"
	"errors"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/test/mocks"
	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/big"
	"os"
	"testing"
)

func getTaskManager(t *testing.T) (*TasksManager, *mocks.MockClient, *db.Database, *taskResponseChan, *mocks.MockWatcher) {
	db := mocks.NewTestDB()
	client := mocks.NewMockClient()
	client.ExtractTransactionSenderFunc.SetDefaultReturn(common.Address{}, nil)
	client.GetTxMaxStaleBlocksFunc.SetDefaultReturn(10)
	hdr := &types.Header{
		Number: big.NewInt(1),
	}
	client.GetHeaderByNumberFunc.SetDefaultReturn(hdr, nil)
	client.GetBlockBaseFeeAndSuggestedGasTipFunc.SetDefaultReturn(big.NewInt(100), big.NewInt(1), nil)
	client.GetDefaultAccountFunc.SetDefaultReturn(accounts.Account{Address: common.Address{}})
	client.GetTransactionByHashFunc.SetDefaultReturn(nil, false, nil)
	client.GetFinalityDelayFunc.SetDefaultReturn(10)

	logger := logging.GetLogger("test")

	txWatcher := mocks.NewMockWatcher()
	taskManager, err := NewTaskManager(txWatcher, db, logger.WithField("Component", "schedule"))
	assert.Nil(t, err)

	taskRespChan := &taskResponseChan{trChan: make(chan tasks.TaskResponse, 100)}
	return taskManager, client, db, taskRespChan, txWatcher
}

func Test_TaskManager_HappyPath(t *testing.T) {
	manager, client, db, taskRespChan, txWatcher := getTaskManager(t)
	defer taskRespChan.close()

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(20),
	}

	receiptResponse := mocks.NewMockReceiptResponse()
	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)

	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)

	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)

	task := mocks.NewMockTask()
	task.PrepareFunc.SetDefaultReturn(nil)
	txn := types.NewTx(&types.LegacyTx{
		Nonce:    1,
		Value:    big.NewInt(1),
		Gas:      1,
		GasPrice: big.NewInt(1),
		Data:     []byte{52, 66, 175, 92},
	})
	task.ExecuteFunc.SetDefaultReturn(txn, nil)
	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
	task.GetLoggerFunc.SetDefaultReturn(manager.logger)

	mainCtx := context.Background()
	manager.ManageTask(mainCtx, task, "", "123", db, manager.logger, client, taskRespChan)

	mockrequire.CalledOnce(t, task.PrepareFunc)
	mockrequire.CalledOnce(t, task.ExecuteFunc)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(nil))
	assert.Emptyf(t, manager.Transactions, "Expected transactions to be empty")
}

func Test_TaskManager_TaskErrorRecoverable(t *testing.T) {
	manager, client, db, taskRespChan, txWatcher := getTaskManager(t)
	defer taskRespChan.close()

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(20),
	}

	receiptResponse := mocks.NewMockReceiptResponse()
	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)

	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)

	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)

	taskErr := tasks.NewTaskErr("Recoverable error", true)
	task := mocks.NewMockTask()
	task.PrepareFunc.PushReturn(taskErr)
	task.PrepareFunc.PushReturn(nil)
	txn := types.NewTx(&types.LegacyTx{
		Nonce:    1,
		Value:    big.NewInt(1),
		Gas:      1,
		GasPrice: big.NewInt(1),
		Data:     []byte{52, 66, 175, 92},
	})
	task.ExecuteFunc.SetDefaultReturn(txn, nil)
	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
	task.GetLoggerFunc.SetDefaultReturn(manager.logger)

	mainCtx := context.Background()
	manager.ManageTask(mainCtx, task, "", "123", db, manager.logger, client, taskRespChan)

	mockrequire.CalledN(t, task.PrepareFunc, 2)
	mockrequire.CalledOnce(t, task.ExecuteFunc)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(nil))
	assert.Emptyf(t, manager.Transactions, "Expected transactions to be empty")
}

func Test_TaskManager_UnrecoverableError(t *testing.T) {
	manager, client, db, taskRespChan, txWatcher := getTaskManager(t)
	defer taskRespChan.close()

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(20),
	}

	receiptResponse := mocks.NewMockReceiptResponse()
	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)

	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)

	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)

	task := mocks.NewMockTask()
	taskErr := tasks.NewTaskErr("Unrecoverable error", false)

	task.PrepareFunc.SetDefaultReturn(taskErr)
	mainCtx := context.Background()
	manager.ManageTask(mainCtx, task, "", "123", db, manager.logger, client, taskRespChan)

	mockrequire.CalledOnce(t, task.PrepareFunc)
	mockrequire.NotCalled(t, task.ExecuteFunc)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(taskErr))
	assert.Emptyf(t, manager.Transactions, "Expected transactions to be empty")
}

func Test_TaskManager_TaskInTasksManagerTransactions(t *testing.T) {
	manager, client, db, taskRespChan, txWatcher := getTaskManager(t)
	defer taskRespChan.close()

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(20),
	}

	receiptResponse := mocks.NewMockReceiptResponse()
	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)

	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)

	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)

	task := mocks.NewMockTask()
	task.PrepareFunc.SetDefaultReturn(nil)
	txn := types.NewTx(&types.LegacyTx{
		Nonce:    1,
		Value:    big.NewInt(1),
		Gas:      1,
		GasPrice: big.NewInt(1),
		Data:     []byte{52, 66, 175, 92},
	})
	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
	task.GetLoggerFunc.SetDefaultReturn(manager.logger)
	taskId := "123"
	task.GetIdFunc.SetDefaultReturn(taskId)

	mainCtx := context.Background()
	manager.Transactions[task.GetId()] = txn
	manager.ManageTask(mainCtx, task, "", taskId, db, manager.logger, client, taskRespChan)

	mockrequire.NotCalled(t, task.PrepareFunc)
	mockrequire.NotCalled(t, task.ExecuteFunc)
	assert.Emptyf(t, manager.Transactions, "Expected transactions to be empty")
}

func Test_TaskManager_ExecuteWithErrors(t *testing.T) {
	manager, client, db, taskRespChan, txWatcher := getTaskManager(t)
	defer taskRespChan.close()

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(20),
	}

	receiptResponse := mocks.NewMockReceiptResponse()
	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)

	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)

	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)

	task := mocks.NewMockTask()
	task.PrepareFunc.SetDefaultReturn(nil)
	task.ExecuteFunc.PushReturn(nil, tasks.NewTaskErr("Recoverable error", true))
	unrecoverableError := tasks.NewTaskErr("Unrecoverable error", false)
	task.ExecuteFunc.PushReturn(nil, unrecoverableError)
	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
	task.GetLoggerFunc.SetDefaultReturn(manager.logger)

	mainCtx := context.Background()
	manager.ManageTask(mainCtx, task, "", "123", db, manager.logger, client, taskRespChan)

	mockrequire.CalledOnce(t, task.PrepareFunc)
	mockrequire.CalledN(t, task.ExecuteFunc, 2)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(unrecoverableError))
	assert.Emptyf(t, manager.Transactions, "Expected transactions to be empty")
}

func Test_TaskManager_ReceiptWithErrorAndFailure(t *testing.T) {
	manager, client, db, taskRespChan, txWatcher := getTaskManager(t)
	defer taskRespChan.close()

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusFailed,
		BlockNumber: big.NewInt(20),
	}

	receiptResponse := mocks.NewMockReceiptResponse()

	task := mocks.NewMockTask()
	task.PrepareFunc.SetDefaultReturn(nil)
	txn := types.NewTx(&types.LegacyTx{
		Nonce:    1,
		Value:    big.NewInt(1),
		Gas:      1,
		GasPrice: big.NewInt(1),
		Data:     []byte{52, 66, 175, 92},
	})
	task.ShouldExecuteFunc.PushReturn(true, nil)
	task.ExecuteFunc.SetDefaultReturn(txn, nil)
	task.GetLoggerFunc.SetDefaultReturn(manager.logger)

	receiptResponse.IsReadyFunc.PushReturn(false)
	task.ShouldExecuteFunc.PushReturn(true, nil)

	receiptResponse.IsReadyFunc.PushReturn(true)
	receiptResponse.GetReceiptBlockingFunc.PushReturn(nil, errors.New("got error while getting receipt"))
	task.ShouldExecuteFunc.PushReturn(true, nil)
	txWatcher.SubscribeFunc.PushReturn(receiptResponse, nil)

	receiptResponse.IsReadyFunc.PushReturn(true)
	receiptResponse.GetReceiptBlockingFunc.PushReturn(receipt, nil)
	task.ShouldExecuteFunc.PushReturn(true, nil)
	txWatcher.SubscribeFunc.PushReturn(receiptResponse, nil)

	receiptResponse.IsReadyFunc.PushReturn(false)
	txWatcher.SubscribeFunc.PushReturn(receiptResponse, nil)
	task.ShouldExecuteFunc.PushReturn(false, nil)

	mainCtx := context.Background()
	manager.ManageTask(mainCtx, task, "", "123", db, manager.logger, client, taskRespChan)

	mockrequire.CalledOnce(t, task.PrepareFunc)
	mockrequire.CalledN(t, task.ExecuteFunc, 3)
	mockrequire.CalledN(t, receiptResponse.IsReadyFunc, 4)
	mockrequire.CalledN(t, receiptResponse.GetReceiptBlockingFunc, 2)
	mockrequire.CalledN(t, task.ShouldExecuteFunc, 5)

	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(nil))
	assert.Emptyf(t, manager.Transactions, "Expected transactions to be empty")
}

func Test_TaskManager_RecoveringTaskManager(t *testing.T) {
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
	client.ExtractTransactionSenderFunc.SetDefaultReturn(common.Address{}, nil)
	client.GetTxMaxStaleBlocksFunc.SetDefaultReturn(10)
	hdr := &types.Header{
		Number: big.NewInt(1),
	}
	client.GetHeaderByNumberFunc.SetDefaultReturn(hdr, nil)
	client.GetBlockBaseFeeAndSuggestedGasTipFunc.SetDefaultReturn(big.NewInt(100), big.NewInt(1), nil)
	client.GetDefaultAccountFunc.SetDefaultReturn(accounts.Account{Address: common.Address{}})
	client.GetTransactionByHashFunc.SetDefaultReturn(nil, false, nil)
	client.GetFinalityDelayFunc.SetDefaultReturn(10)

	logger := logging.GetLogger("test")

	txWatcher := mocks.NewMockWatcher()
	manager, err := NewTaskManager(txWatcher, db, logger.WithField("Component", "schedule"))
	assert.Nil(t, err)

	taskRespChan := &taskResponseChan{trChan: make(chan tasks.TaskResponse, 100)}
	defer taskRespChan.close()

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(20),
	}

	receiptResponse := mocks.NewMockReceiptResponse()
	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
	receiptResponse.GetReceiptBlockingFunc.PushReturn(nil, &transaction.ErrTransactionStale{})
	receiptResponse.GetReceiptBlockingFunc.PushReturn(receipt, nil)

	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)

	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)

	task := mocks.NewMockTask()
	task.PrepareFunc.SetDefaultReturn(nil)
	txn := types.NewTx(&types.LegacyTx{
		Nonce:    1,
		Value:    big.NewInt(1),
		Gas:      1,
		GasPrice: big.NewInt(1),
		Data:     []byte{52, 66, 175, 92},
	})
	task.ExecuteFunc.SetDefaultReturn(txn, nil)
	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
	task.GetLoggerFunc.SetDefaultReturn(manager.logger)

	mainCtx := context.Background()
	manager.ManageTask(mainCtx, task, "", "123", db, manager.logger, client, taskRespChan)

	assert.Equalf(t, 1, len(manager.Transactions), "Expected one transaction (stale status)")
	manager, err = NewTaskManager(txWatcher, db, logger.WithField("Component", "schedule"))
	assert.Nil(t, err)
	manager.ManageTask(mainCtx, task, "", "123", db, manager.logger, client, taskRespChan)

	mockrequire.CalledOnce(t, task.PrepareFunc)
	mockrequire.CalledOnce(t, task.ExecuteFunc)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(nil))
	assert.Emptyf(t, manager.Transactions, "Expected transactions to be empty")

}
