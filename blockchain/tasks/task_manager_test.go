package tasks_test

import (
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/alicenet/alicenet/blockchain/tasks"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/stretchr/testify/assert"
)

func TestStartTask_initializeTask_HappyPath(t *testing.T) {
	eth := mocks.NewMockEthereum()
	task := mocks.NewMockTask()

	wg := sync.WaitGroup{}
	err := tasks.StartTask(mocks.NewMockLogger().WithField("", nil), &wg, eth, task, nil, nil)
	assert.NoError(t, err)
	wg.Wait()

	mockrequire.Called(t, task.DoWorkFunc)
	mockrequire.NotCalled(t, task.DoRetryFunc)
	mockrequire.Called(t, task.DoDoneFunc)
}

func TestStartTask_initializeTask_Error(t *testing.T) {
	eth := mocks.NewMockEthereum()
	task := mocks.NewMockTask()
	task.InitializeFunc.SetDefaultReturn(errors.New("initialize error"))

	wg := sync.WaitGroup{}
	err := tasks.StartTask(mocks.NewMockLogger().WithField("", nil), &wg, eth, task, nil, nil)
	assert.NoError(t, err)
	wg.Wait()

	mockrequire.NotCalled(t, task.DoWorkFunc)
	mockrequire.NotCalled(t, task.DoRetryFunc)
	mockrequire.Called(t, task.DoDoneFunc)

}

func TestStartTask_executeTask_ErrorRetry(t *testing.T) {
	eth := mocks.NewMockEthereum()
	eth.RetryCountFunc.SetDefaultReturn(10)

	task := mocks.NewMockTask()
	task.ShouldRetryFunc.SetDefaultReturn(true)
	task.DoWorkFunc.SetDefaultReturn(errors.New("DoWork_error"))
	task.DoRetryFunc.SetDefaultReturn(errors.New(tasks.NonceToLowError))

	wg := sync.WaitGroup{}
	err := tasks.StartTask(mocks.NewMockLogger().WithField("Task", 0), &wg, eth, task, nil, nil)
	assert.NoError(t, err)
	wg.Wait()

	mockrequire.Called(t, task.DoWorkFunc)
	mockrequire.CalledN(t, task.DoRetryFunc, 10)
	mockrequire.Called(t, task.DoDoneFunc)
}

// Happy path with mined tx present after finality delay
func TestStartTask_handleExecutedTask_FinalityDelay1(t *testing.T) {
	task := mocks.NewMockTaskWithExecutionData(1, 100)
	task.ExecutionData.TxOpts.TxHashes = append(task.ExecutionData.TxOpts.TxHashes, common.BigToHash(big.NewInt(123871239)))

	eth := mocks.NewMockEthereum()
	eth.GethClientMock.TransactionByHashFunc.SetDefaultReturn(&types.Transaction{}, false, nil)
	eth.GethClientMock.TransactionReceiptFunc.SetDefaultReturn(&types.Receipt{Status: uint64(1), BlockNumber: big.NewInt(1)}, nil)

	wg := sync.WaitGroup{}
	err := tasks.StartTask(mocks.NewMockLogger().WithField("Task", 0), &wg, eth, task, objects.NewDkgState(accounts.Account{}), nil)
	assert.NoError(t, err)
	wg.Wait()

	mockrequire.Called(t, task.DoWorkFunc)
	mockrequire.Called(t, task.DoDoneFunc)
	assert.Len(t, task.ExecutionData.TxOpts.TxHashes, 1)
	assert.Equal(t, uint64(1), task.ExecutionData.TxOpts.MinedInBlock)
}

// Tx was mined, but it's not present after finality delay
func TestStartTask_handleExecutedTask_FinalityDelay2(t *testing.T) {
	minedInBlock := 9

	task := mocks.NewMockTaskWithExecutionData(1, 100)
	task.ExecutionData.TxOpts.TxHashes = append(task.ExecutionData.TxOpts.TxHashes, common.BigToHash(big.NewInt(123871239)))

	eth := mocks.NewMockEthereum()
	eth.GethClientMock.TransactionByHashFunc.SetDefaultReturn(&types.Transaction{}, false, nil)
	eth.GethClientMock.TransactionReceiptFunc.PushReturn(&types.Receipt{Status: uint64(1), BlockNumber: big.NewInt(2)}, nil)
	eth.GethClientMock.TransactionReceiptFunc.PushReturn(&types.Receipt{}, errors.New("error getting receipt"))
	eth.GethClientMock.TransactionReceiptFunc.PushReturn(&types.Receipt{Status: uint64(1), BlockNumber: big.NewInt(int64(minedInBlock))}, nil)
	eth.GethClientMock.TransactionReceiptFunc.PushReturn(&types.Receipt{Status: uint64(1), BlockNumber: big.NewInt(int64(minedInBlock))}, nil)

	wg := sync.WaitGroup{}
	err := tasks.StartTask(mocks.NewMockLogger().WithField("Task", 0), &wg, eth, task, objects.NewDkgState(accounts.Account{}), nil)
	assert.NoError(t, err)
	wg.Wait()

	mockrequire.Called(t, task.DoWorkFunc)
	mockrequire.Called(t, task.DoDoneFunc)
	assert.Len(t, task.ExecutionData.TxOpts.TxHashes, 1)
	assert.Equal(t, task.ExecutionData.TxOpts.MinedInBlock, uint64(minedInBlock))
}

// Tx was mined after a retry because of a failed receipt
func TestStartTask_handleExecutedTask_RetrySameFee(t *testing.T) {
	minedInBlock := 7

	task := mocks.NewMockTaskWithExecutionData(1, 100)
	task.ExecutionData.TxOpts.TxHashes = append(task.ExecutionData.TxOpts.TxHashes, common.BigToHash(big.NewInt(123871239)))
	task.ShouldRetryFunc.SetDefaultReturn(true)

	eth := mocks.NewMockEthereum()
	eth.GethClientMock.TransactionByHashFunc.SetDefaultReturn(&types.Transaction{}, false, nil)
	eth.GethClientMock.TransactionReceiptFunc.PushReturn(&types.Receipt{Status: 0}, nil)
	eth.GethClientMock.TransactionReceiptFunc.PushReturn(&types.Receipt{Status: uint64(1), BlockNumber: big.NewInt(int64(minedInBlock))}, nil)
	eth.GethClientMock.TransactionReceiptFunc.PushReturn(&types.Receipt{Status: uint64(1), BlockNumber: big.NewInt(int64(minedInBlock))}, nil)

	wg := sync.WaitGroup{}
	err := tasks.StartTask(mocks.NewMockLogger().WithField("Task", 0), &wg, eth, task, objects.NewDkgState(accounts.Account{}), nil)
	assert.NoError(t, err)
	wg.Wait()

	mockrequire.Called(t, task.DoWorkFunc)
	mockrequire.Called(t, task.DoRetryFunc)
	mockrequire.Called(t, task.DoDoneFunc)
	assert.Len(t, task.ExecutionData.TxOpts.TxHashes, 1)
	assert.Equal(t, task.ExecutionData.TxOpts.MinedInBlock, uint64(minedInBlock))
}

// Tx reached replacement timeout, tx mined after retry with replacement
func TestStartTask_handleExecutedTask_RetryReplacingFee(t *testing.T) {
	minedInBlock := 10

	task := mocks.NewMockTaskWithExecutionData(1, 100)
	task.ExecutionData.TxOpts.TxHashes = append(task.ExecutionData.TxOpts.TxHashes, common.BigToHash(big.NewInt(123871239)))

	eth := mocks.NewMockEthereum()
	eth.GetTxCheckFrequencyFunc.SetDefaultReturn(5 * time.Millisecond)
	eth.GethClientMock.TransactionByHashFunc.PushReturn(&types.Transaction{}, true, nil)
	eth.GethClientMock.TransactionByHashFunc.PushReturn(&types.Transaction{}, true, nil)
	eth.GethClientMock.TransactionByHashFunc.PushReturn(&types.Transaction{}, false, nil)
	eth.GethClientMock.TransactionReceiptFunc.SetDefaultReturn(&types.Receipt{Status: uint64(1), BlockNumber: big.NewInt(int64(minedInBlock))}, nil)

	wg := sync.WaitGroup{}
	err := tasks.StartTask(mocks.NewMockLogger().WithField("Task", 0), &wg, eth, task, objects.NewDkgState(accounts.Account{}), nil)
	assert.NoError(t, err)
	wg.Wait()

	expectedGasFeeCap := big.NewInt(213534)
	expectedGasTipCap := big.NewInt(55)

	mockrequire.Called(t, task.DoWorkFunc)
	mockrequire.Called(t, task.DoDoneFunc)
	assert.Len(t, task.ExecutionData.TxOpts.TxHashes, 1)
	assert.Equal(t, task.ExecutionData.TxOpts.MinedInBlock, uint64(minedInBlock))
	assert.Equal(t, expectedGasFeeCap, task.ExecutionData.TxOpts.GasFeeCap)
	assert.Equal(t, expectedGasTipCap, task.ExecutionData.TxOpts.GasTipCap)
}

// Tx reached replacement timeout, tx mined after retry with replacement
func TestStartTask_handleExecutedTask_RetryReplacingFeeExceedingThreshold(t *testing.T) {
	task := mocks.NewMockTaskWithExecutionData(1, 100)

	eth := mocks.NewMockEthereum()
	eth.GetTxCheckFrequencyFunc.SetDefaultReturn(5 * time.Millisecond)
	for i := 0; i < 20; i++ {
		eth.GethClientMock.TransactionByHashFunc.PushReturn(&types.Transaction{}, true, nil)
	}
	eth.GethClientMock.TransactionByHashFunc.PushReturn(&types.Transaction{}, false, nil)
	eth.GethClientMock.TransactionReceiptFunc.SetDefaultReturn(&types.Receipt{Status: uint64(1), BlockNumber: big.NewInt(int64(10))}, nil)

	wg := sync.WaitGroup{}
	err := tasks.StartTask(mocks.NewMockLogger().WithField("Task", 0), &wg, eth, task, objects.NewDkgState(accounts.Account{}), nil)
	assert.NoError(t, err)
	wg.Wait()

	expectedGasFeeCap := big.NewInt(1000000)

	mockrequire.Called(t, task.DoWorkFunc)
	mockrequire.Called(t, task.DoDoneFunc)
	assert.Equal(t, expectedGasFeeCap, task.ExecutionData.TxOpts.GasFeeCap)
}
