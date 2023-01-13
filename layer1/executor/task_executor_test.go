package executor

import (
	"errors"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	taskMocks "github.com/alicenet/alicenet/layer1/executor/tasks/mocks"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/alicenet/alicenet/utils"
	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTaskExecutor(
	t *testing.T,
) (*TaskExecutor, *mocks.MockClient, *db.Database, *executorResponseChan, *mocks.MockWatcher) {
	db := mocks.NewTestDB()
	client := mocks.NewMockClient()
	client.ExtractTransactionSenderFunc.SetDefaultReturn(common.Address{}, nil)
	client.GetTxMaxStaleBlocksFunc.SetDefaultReturn(10)
	hdr := &types.Header{
		Number: big.NewInt(1),
	}
	client.GetHeaderByNumberFunc.SetDefaultReturn(hdr, nil)
	client.GetBlockBaseFeeAndSuggestedGasTipFunc.SetDefaultReturn(
		big.NewInt(100),
		big.NewInt(1),
		nil,
	)
	client.GetBlockBaseFeeAndSuggestedGasTipFunc.SetDefaultReturn(
		big.NewInt(100),
		big.NewInt(1),
		nil,
	)
	client.GetDefaultAccountFunc.SetDefaultReturn(accounts.Account{Address: common.Address{}})
	client.GetTransactionByHashFunc.SetDefaultReturn(nil, false, nil)
	client.GetFinalityDelayFunc.SetDefaultReturn(10)

	logger := logging.GetLogger("test")

	txWatcher := mocks.NewMockWatcher()
	taskExecutor, err := newTaskExecutor(txWatcher, db, logger.WithField("Component", "schedule"))
	assert.Nil(t, err)

	taskRespChan := &executorResponseChan{erChan: make(chan ExecutorResponse, 100)}

	t.Cleanup(func() {
		taskRespChan.close()
	})
	return taskExecutor, client, db, taskRespChan, txWatcher
}

func Test_TaskExecutor_HappyPath(t *testing.T) {
	executor, client, db, taskRespChan, txWatcher := getTaskExecutor(t)

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(20),
	}

	receiptResponse := mocks.NewMockReceiptResponse()
	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)
	txWatcher.SubscribeFunc.PushReturn(nil, errors.New("subscribe error"))
	txWatcher.SubscribeFunc.PushReturn(receiptResponse, nil)

	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)

	task := taskMocks.NewMockTask()
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
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)

	executor.handleTaskExecution(
		task,
		"",
		"123",
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.CalledOnce(t, task.PrepareFunc)
	mockrequire.CalledN(t, task.ExecuteFunc, 2)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(nil))
	assert.Emptyf(t, executor.TxsBackup, "Expected transactions to be empty")
}

func Test_TaskExecutor_TaskErrorRecoverable(t *testing.T) {
	executor, client, db, taskRespChan, txWatcher := getTaskExecutor(t)

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
	task := taskMocks.NewMockTask()
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
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)

	executor.handleTaskExecution(
		task,
		"",
		"123",
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.CalledN(t, task.PrepareFunc, 2)
	mockrequire.CalledOnce(t, task.ExecuteFunc)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(nil))
	assert.Emptyf(t, executor.TxsBackup, "Expected transactions to be empty")
}

func Test_TaskExecutor_UnrecoverableError(t *testing.T) {
	executor, client, db, taskRespChan, txWatcher := getTaskExecutor(t)

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(20),
	}

	receiptResponse := mocks.NewMockReceiptResponse()
	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)

	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)

	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)

	task := taskMocks.NewMockTask()
	taskErr := tasks.NewTaskErr("Unrecoverable error", false)

	task.PrepareFunc.SetDefaultReturn(taskErr)
	executor.handleTaskExecution(
		task,
		"",
		"123",
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.CalledOnce(t, task.PrepareFunc)
	mockrequire.NotCalled(t, task.ExecuteFunc)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(taskErr))
	assert.Emptyf(t, executor.TxsBackup, "Expected transactions to be empty")
}

func Test_TaskExecutor_TaskInTasksExecutorTransactions(t *testing.T) {
	executor, client, db, taskRespChan, txWatcher := getTaskExecutor(t)

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(20),
	}

	receiptResponse := mocks.NewMockReceiptResponse()
	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)

	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)

	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)

	task := taskMocks.NewMockTask()
	task.PrepareFunc.SetDefaultReturn(nil)
	txn := types.NewTx(&types.LegacyTx{
		Nonce:    1,
		Value:    big.NewInt(1),
		Gas:      1,
		GasPrice: big.NewInt(1),
		Data:     []byte{52, 66, 175, 92},
	})
	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)
	taskId := "123"
	task.GetIdFunc.SetDefaultReturn(taskId)

	executor.TxsBackup[task.GetId()] = txn
	executor.handleTaskExecution(
		task,
		"",
		taskId,
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.NotCalled(t, task.PrepareFunc)
	mockrequire.NotCalled(t, task.ExecuteFunc)
	assert.Emptyf(t, executor.TxsBackup, "Expected transactions to be empty")
}

func Test_TaskExecutor_ExecuteWithErrors(t *testing.T) {
	executor, client, db, taskRespChan, txWatcher := getTaskExecutor(t)

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(20),
	}

	receiptResponse := mocks.NewMockReceiptResponse()
	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)

	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)

	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)

	task := taskMocks.NewMockTask()
	task.PrepareFunc.SetDefaultReturn(nil)
	task.ExecuteFunc.PushReturn(nil, tasks.NewTaskErr("Recoverable error", true))
	unrecoverableError := tasks.NewTaskErr("Unrecoverable error", false)
	task.ExecuteFunc.PushReturn(nil, unrecoverableError)
	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)

	executor.handleTaskExecution(
		task,
		"",
		"123",
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.CalledOnce(t, task.PrepareFunc)
	mockrequire.CalledN(t, task.ExecuteFunc, 2)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(unrecoverableError))
	assert.Emptyf(t, executor.TxsBackup, "Expected transactions to be empty")
}

func Test_TaskExecutor_ReceiptWithErrorAndFailure(t *testing.T) {
	executor, client, db, taskRespChan, txWatcher := getTaskExecutor(t)

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusFailed,
		BlockNumber: big.NewInt(20),
	}

	receiptResponse := mocks.NewMockReceiptResponse()

	task := taskMocks.NewMockTask()
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
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)

	receiptResponse.IsReadyFunc.PushReturn(false)
	task.ShouldExecuteFunc.PushReturn(true, nil)

	receiptResponse.IsReadyFunc.PushReturn(true)
	receiptResponse.GetReceiptBlockingFunc.PushReturn(
		nil,
		errors.New("got error while getting receipt"),
	)
	task.ShouldExecuteFunc.PushReturn(true, nil)
	txWatcher.SubscribeFunc.PushReturn(receiptResponse, nil)

	receiptResponse.IsReadyFunc.PushReturn(true)
	receiptResponse.GetReceiptBlockingFunc.PushReturn(receipt, nil)
	task.ShouldExecuteFunc.PushReturn(true, nil)
	txWatcher.SubscribeFunc.PushReturn(receiptResponse, nil)

	receiptResponse.IsReadyFunc.PushReturn(false)
	txWatcher.SubscribeFunc.PushReturn(receiptResponse, nil)
	task.ShouldExecuteFunc.PushReturn(false, nil)

	executor.handleTaskExecution(
		task,
		"",
		"123",
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.CalledOnce(t, task.PrepareFunc)
	mockrequire.CalledN(t, task.ExecuteFunc, 3)
	mockrequire.CalledN(t, receiptResponse.IsReadyFunc, 4)
	mockrequire.CalledN(t, receiptResponse.GetReceiptBlockingFunc, 2)
	mockrequire.CalledN(t, task.ShouldExecuteFunc, 5)

	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(nil))
	assert.Emptyf(t, executor.TxsBackup, "Expected transactions to be empty")
}

func Test_TaskExecutor_Recovering(t *testing.T) {
	dir, err := os.MkdirTemp("", "db-test")
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
	client.GetBlockBaseFeeAndSuggestedGasTipFunc.SetDefaultReturn(
		big.NewInt(100),
		big.NewInt(1),
		nil,
	)
	client.GetDefaultAccountFunc.SetDefaultReturn(accounts.Account{Address: common.Address{}})
	client.GetTransactionByHashFunc.SetDefaultReturn(nil, false, nil)
	client.GetFinalityDelayFunc.SetDefaultReturn(10)

	logger := logging.GetLogger("test")

	txWatcher := mocks.NewMockWatcher()
	executor, err := newTaskExecutor(txWatcher, db, logger.WithField("Component", "schedule"))
	assert.Nil(t, err)

	taskRespChan := &executorResponseChan{erChan: make(chan ExecutorResponse, 100)}
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

	task := taskMocks.NewMockTask()
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
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)

	executor.handleTaskExecution(
		task,
		"",
		"123",
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	assert.Equalf(t, 1, len(executor.TxsBackup), "Expected one transaction (stale status)")
	executor, err = newTaskExecutor(txWatcher, db, logger.WithField("Component", "schedule"))
	assert.Nil(t, err)
	executor.handleTaskExecution(
		task,
		"",
		"123",
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.CalledOnce(t, task.PrepareFunc)
	mockrequire.CalledOnce(t, task.ExecuteFunc)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(nil))
	assert.Emptyf(t, executor.TxsBackup, "Expected transactions to be empty")

}

func Test_TaskExecutor_TxFromBackupStale(t *testing.T) {
	executor, client, db, taskRespChan, txWatcher := getTaskExecutor(t)

	receiptResponse := mocks.NewMockReceiptResponse()
	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
	receiptResponse.GetReceiptBlockingFunc.PushReturn(nil, &transaction.ErrTransactionStale{})

	txn := types.NewTx(&types.LegacyTx{
		Nonce:    1,
		Value:    big.NewInt(1),
		Gas:      1,
		GasPrice: big.NewInt(1),
		Data:     []byte{52, 66, 175, 92},
	})

	taskId := "123"
	task := taskMocks.NewMockTask()
	task.GetIdFunc.SetDefaultReturn(taskId)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)
	executor.TxsBackup[taskId] = txn
	txWatcher.SubscribeFunc.PushReturn(receiptResponse, nil)

	executor.handleTaskExecution(
		task,
		"",
		"123",
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.CalledOnce(t, txWatcher.SubscribeFunc, 1)
	mockrequire.CalledOnceWith(
		t,
		task.FinishFunc,
		mockrequire.Values(&transaction.ErrTransactionStale{}),
	)
}

func Test_TaskExecutor_DbCorrupted(t *testing.T) {
	database := mocks.NewTestDB()

	err := database.Update(func(txn *badger.Txn) error {
		key := dbprefix.PrefixTaskExecutorState()
		rawData := []byte("corrupted data")
		if err := utils.SetValue(txn, key, rawData); err != nil {
			return err
		}
		return nil
	})
	require.Nil(t, err)

	logger := logging.GetLogger("test")
	txWatcher := mocks.NewMockWatcher()
	taskExecutor, err := newTaskExecutor(
		txWatcher,
		database,
		logger.WithField("Component", "schedule"),
	)
	assert.Nil(t, taskExecutor)
	assert.NotNil(t, err)
}

func Test_TaskExecutor_InitError(t *testing.T) {
	executor, client, db, taskRespChan, _ := getTaskExecutor(t)

	taskId := "123"
	task := taskMocks.NewMockTask()
	task.GetIdFunc.SetDefaultReturn(taskId)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)
	err := errors.New("init error")
	task.InitializeFunc.SetDefaultReturn(err)

	executor.handleTaskExecution(
		task,
		"",
		"123",
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.CalledOnce(t, task.InitializeFunc)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(err))
}

func Test_TaskExecutor_Close(t *testing.T) {
	executor, client, db, taskRespChan, _ := getTaskExecutor(t)

	taskId := "123"
	task := taskMocks.NewMockTask()
	task.GetIdFunc.SetDefaultReturn(taskId)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)
	task.InitializeFunc.SetDefaultReturn(nil)

	executor.close()
	executor.handleTaskExecution(
		task,
		"",
		"123",
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.CalledOnce(t, task.InitializeFunc)
	mockrequire.NotCalled(t, task.FinishFunc)
}

func Test_TaskExecutor_ShouldExecuteError(t *testing.T) {
	executor, client, db, taskRespChan, _ := getTaskExecutor(t)

	taskId := "123"
	task := taskMocks.NewMockTask()
	task.GetIdFunc.SetDefaultReturn(taskId)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)
	err := tasks.NewTaskErr("should execute error", false)
	task.InitializeFunc.SetDefaultReturn(nil)
	task.ShouldExecuteFunc.SetDefaultReturn(false, err)

	executor.handleTaskExecution(
		task,
		"",
		"123",
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.CalledOnce(t, task.InitializeFunc)
	mockrequire.CalledOnce(t, task.ShouldExecuteFunc)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(err))
}

func Test_TaskExecutor_ShouldExecuteFalse(t *testing.T) {
	executor, client, db, taskRespChan, _ := getTaskExecutor(t)

	taskId := "123"
	task := taskMocks.NewMockTask()
	task.GetIdFunc.SetDefaultReturn(taskId)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)
	task.InitializeFunc.SetDefaultReturn(nil)
	task.ShouldExecuteFunc.SetDefaultReturn(false, nil)

	executor.handleTaskExecution(
		task,
		"",
		"123",
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)
	mockrequire.CalledOnce(t, task.InitializeFunc)
	mockrequire.CalledOnce(t, task.ShouldExecuteFunc)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(nil))
}

func Test_TaskExecutor_ExecuteWithoutTxn(t *testing.T) {
	executor, client, db, taskRespChan, _ := getTaskExecutor(t)

	taskId := "123"
	task := taskMocks.NewMockTask()
	task.GetIdFunc.SetDefaultReturn(taskId)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)
	task.InitializeFunc.SetDefaultReturn(nil)
	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
	task.ExecuteFunc.SetDefaultReturn(nil, nil)

	executor.handleTaskExecution(
		task,
		"",
		"123",
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.CalledOnce(t, task.InitializeFunc)
	mockrequire.CalledOnce(t, task.ShouldExecuteFunc)
	mockrequire.CalledOnce(t, task.ExecuteFunc)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(nil))
}

func Test_TaskExecutor_HasToExecuteError(t *testing.T) {
	executor, client, db, taskRespChan, txWatcher := getTaskExecutor(t)

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(20),
	}

	receiptResponse := mocks.NewMockReceiptResponse()
	receiptResponse.IsReadyFunc.PushReturn(false)
	receiptResponse.IsReadyFunc.PushReturn(true)
	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)
	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)

	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)

	task := taskMocks.NewMockTask()
	task.PrepareFunc.SetDefaultReturn(nil)
	txn := types.NewTx(&types.LegacyTx{
		Nonce:    1,
		Value:    big.NewInt(1),
		Gas:      1,
		GasPrice: big.NewInt(1),
		Data:     []byte{52, 66, 175, 92},
	})
	task.ExecuteFunc.SetDefaultReturn(txn, nil)
	task.ShouldExecuteFunc.PushReturn(true, nil)
	err := tasks.NewTaskErr("should execute error", false)
	task.ShouldExecuteFunc.PushReturn(true, err)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)

	executor.handleTaskExecution(
		task,
		"",
		"123",
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.CalledOnce(t, task.PrepareFunc)
	mockrequire.CalledN(t, task.ShouldExecuteFunc, 2)
	mockrequire.CalledN(t, task.ExecuteFunc, 1)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(err))
}

func Test_TaskExecutor_ShouldExecuteRecoverableError(t *testing.T) {
	executor, client, db, taskRespChan, _ := getTaskExecutor(t)

	taskId := "123"
	task := taskMocks.NewMockTask()
	task.GetIdFunc.SetDefaultReturn(taskId)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)
	err := tasks.NewTaskErr("should execute error", true)
	task.InitializeFunc.SetDefaultReturn(nil)
	task.ShouldExecuteFunc.SetDefaultReturn(false, err)
	task.ExecuteFunc.SetDefaultReturn(nil, nil)

	executor.handleTaskExecution(
		task,
		"",
		taskId,
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.CalledOnce(t, task.InitializeFunc)
	mockrequire.CalledN(t, task.ShouldExecuteFunc, 10)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(nil))
}

func Test_TaskExecutor_KilledTask(t *testing.T) {
	executor, client, db, taskRespChan, _ := getTaskExecutor(t)

	taskId := "123"
	task := taskMocks.NewMockTask()
	task.GetIdFunc.SetDefaultReturn(taskId)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)
	task.InitializeFunc.SetDefaultReturn(nil)
	task.WasKilledFunc.SetDefaultReturn(true)

	executor.handleTaskExecution(
		task,
		"",
		taskId,
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)
	mockrequire.CalledOnce(t, task.InitializeFunc)
	mockrequire.NotCalled(t, task.PrepareFunc)
	mockrequire.NotCalled(t, task.ShouldExecuteFunc)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(tasks.ErrTaskKilled))
}

func Test_TaskExecutor_KilledTaskAfterRecoverableError(t *testing.T) {
	executor, client, db, taskRespChan, _ := getTaskExecutor(t)

	taskId := "123"
	task := taskMocks.NewMockTask()
	task.GetIdFunc.SetDefaultReturn(taskId)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)
	task.InitializeFunc.SetDefaultReturn(nil)
	err := tasks.NewTaskErr("prepare error", true)
	task.PrepareFunc.SetDefaultReturn(err)
	task.WasKilledFunc.SetDefaultReturn(false)
	killChan := make(chan struct{})
	close(killChan)
	task.KillChanFunc.SetDefaultReturn(killChan)

	executor.handleTaskExecution(
		task,
		"",
		taskId,
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)
	mockrequire.CalledOnce(t, task.InitializeFunc)
	mockrequire.CalledOnce(t, task.PrepareFunc)
	mockrequire.NotCalled(t, task.ShouldExecuteFunc)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(tasks.ErrTaskKilled))
}

func Test_TaskExecutor_CloseExecutorAfter1stExecution(t *testing.T) {
	executor, client, db, taskRespChan, _ := getTaskExecutor(t)

	taskId := "123"
	task := taskMocks.NewMockTask()
	task.GetIdFunc.SetDefaultReturn(taskId)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)
	task.InitializeFunc.SetDefaultReturn(nil)
	task.PrepareFunc.SetDefaultReturn(nil)
	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
	task.WasKilledFunc.SetDefaultReturn(false)
	err := tasks.NewTaskErr("execute error", true)
	task.ExecuteFunc.SetDefaultReturn(nil, err)

	go func() {
		delay := constants.MonitorRetryDelay - 1*time.Second
		time.Sleep(delay)
		executor.close()
	}()

	executor.handleTaskExecution(
		task,
		"",
		taskId,
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)
	mockrequire.CalledOnce(t, task.InitializeFunc)
	mockrequire.CalledOnce(t, task.PrepareFunc)
	mockrequire.CalledOnce(t, task.ShouldExecuteFunc)
	mockrequire.CalledOnce(t, task.ExecuteFunc)
	mockrequire.NotCalled(t, task.FinishFunc)
}

func Test_TaskExecutor_BackupTxnTaskKilled(t *testing.T) {
	executor, client, db, taskRespChan, txWatcher := getTaskExecutor(t)

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(20),
	}

	receiptResponse := mocks.NewMockReceiptResponse()
	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)

	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)

	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)

	task := taskMocks.NewMockTask()
	task.PrepareFunc.SetDefaultReturn(nil)
	txn := types.NewTx(&types.LegacyTx{
		Nonce:    1,
		Value:    big.NewInt(1),
		Gas:      1,
		GasPrice: big.NewInt(1),
		Data:     []byte{52, 66, 175, 92},
	})
	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)
	taskId := "123"
	task.GetIdFunc.SetDefaultReturn(taskId)
	killChan := make(chan struct{})
	close(killChan)
	task.KillChanFunc.SetDefaultReturn(killChan)
	task.WasKilledFunc.SetDefaultReturn(true)

	executor.TxsBackup[task.GetId()] = txn
	executor.handleTaskExecution(
		task,
		"",
		taskId,
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.CalledOnce(t, task.InitializeFunc)
	mockrequire.NotCalled(t, task.PrepareFunc)
	mockrequire.NotCalled(t, task.ExecuteFunc)
	assert.Emptyf(t, executor.TxsBackup, "Expected transactions to be empty")
}

func Test_TaskExecutor_BackupTxnExecutorClosed(t *testing.T) {
	executor, client, db, taskRespChan, txWatcher := getTaskExecutor(t)

	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(20),
	}

	receiptResponse := mocks.NewMockReceiptResponse()
	receiptResponse.IsReadyFunc.SetDefaultReturn(true)
	receiptResponse.GetReceiptBlockingFunc.SetDefaultReturn(receipt, nil)

	txWatcher.SubscribeFunc.SetDefaultReturn(receiptResponse, nil)

	client.GetTransactionReceiptFunc.SetDefaultReturn(receipt, nil)

	task := taskMocks.NewMockTask()
	task.PrepareFunc.SetDefaultReturn(nil)
	txn := types.NewTx(&types.LegacyTx{
		Nonce:    1,
		Value:    big.NewInt(1),
		Gas:      1,
		GasPrice: big.NewInt(1),
		Data:     []byte{52, 66, 175, 92},
	})
	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)
	taskId := "123"
	task.GetIdFunc.SetDefaultReturn(taskId)
	task.WasKilledFunc.SetDefaultReturn(false)

	executor.TxsBackup[task.GetId()] = txn
	executor.close()
	executor.handleTaskExecution(
		task,
		"",
		taskId,
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)

	mockrequire.CalledOnce(t, task.InitializeFunc)
	mockrequire.NotCalled(t, task.PrepareFunc)
	mockrequire.NotCalled(t, task.ExecuteFunc)
	mockrequire.NotCalled(t, task.FinishFunc)
}

func Test_TaskExecutor_ShouldExecuteKilledTask(t *testing.T) {
	executor, client, db, taskRespChan, _ := getTaskExecutor(t)

	taskId := "123"
	task := taskMocks.NewMockTask()
	task.GetIdFunc.SetDefaultReturn(taskId)
	task.GetLoggerFunc.SetDefaultReturn(executor.logger)
	task.InitializeFunc.SetDefaultReturn(nil)
	task.PrepareFunc.SetDefaultReturn(nil)
	task.ShouldExecuteFunc.SetDefaultReturn(true, nil)
	task.WasKilledFunc.PushReturn(false)
	task.WasKilledFunc.PushReturn(true)

	executor.handleTaskExecution(
		task,
		"",
		taskId,
		1,
		10,
		false,
		nil,
		db,
		executor.logger,
		client,
		mocks.NewMockAllSmartContracts(),
		taskRespChan,
	)
	mockrequire.CalledOnce(t, task.InitializeFunc)
	mockrequire.CalledOnce(t, task.PrepareFunc)
	mockrequire.NotCalled(t, task.ShouldExecuteFunc)
	mockrequire.NotCalled(t, task.ExecuteFunc)
	mockrequire.CalledOnceWith(t, task.FinishFunc, mockrequire.Values(tasks.ErrTaskKilled))
}

func Test_TaskExecutor_PersistError(t *testing.T) {
	executor, _, _, _, _ := getTaskExecutor(t)
	executor.database.DB().Close()
	err := executor.persistState()
	assert.NotNil(t, err)
}
