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

	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/test/mocks"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/stretchr/testify/assert"
)

func TestStartTask_initializeTask_HappyPath(t *testing.T) {
	eth := mocks.NewMockEthereum()
	task := mocks.NewMockTask()

	wg := sync.WaitGroup{}
	tasks.StartTask(mocks.NewMockLogger().WithField("", nil), &wg, eth, task, nil, nil)
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
<<<<<<< HEAD

	ethMock := &interfaces.EthereumMock{}
	ethMock.On("RetryCount").Return(3)
	ethMock.On("RetryDelay").Return(10 * time.Millisecond)

	tasks.StartTask(logger.WithField("Task", 0), &wg, ethMock, dkgTask, nil, nil)

=======
	tasks.StartTask(mocks.NewMockLogger().WithField("", nil), &wg, eth, task, nil, nil)
>>>>>>> upstream/candidate
	wg.Wait()

	mockrequire.NotCalled(t, task.DoWorkFunc)
	mockrequire.NotCalled(t, task.DoRetryFunc)
	mockrequire.Called(t, task.DoDoneFunc)

}

func TestStartTask_executeTask_ErrorRetry(t *testing.T) {
	eth := mocks.NewMockEthereum()
	eth.RetryCountFunc.SetDefaultReturn(10)

<<<<<<< HEAD
	state := objects.NewDkgState(accounts.Account{})
	dkgTask := dkgtasks.NewDkgTaskMock(state, 1, 100)
	dkgTask.TxOpts = &tasks.TxOpts{
		Nonce: big.NewInt(1),
	}

	dkgTask.On("Initialize", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	dkgTask.On("DoWork", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("DoWork_error"))
	dkgTask.On("ShouldRetry", mock.Anything, mock.Anything, mock.Anything).Return(true)
	dkgTask.On("DoRetry", mock.Anything, mock.Anything, mock.Anything).Return(errors.New(tasks.NonceToLowError))

	wg := sync.WaitGroup{}

	ethMock := &interfaces.EthereumMock{}
	ethMock.On("RetryCount").Return(3)
	ethMock.On("RetryDelay").Return(10 * time.Millisecond)

	tasks.StartTask(logger.WithField("Task", 0), &wg, ethMock, dkgTask, nil, nil)

=======
	task := mocks.NewMockTask()
	task.ShouldRetryFunc.SetDefaultReturn(true)
	task.DoWorkFunc.SetDefaultReturn(errors.New("DoWork_error"))
	task.DoRetryFunc.SetDefaultReturn(errors.New(tasks.NonceToLowError))

	wg := sync.WaitGroup{}
	tasks.StartTask(mocks.NewMockLogger().WithField("Task", 0), &wg, eth, task, nil, nil)
>>>>>>> upstream/candidate
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
<<<<<<< HEAD

	gethClientMock := &interfaces.GethClientMock{}
	gethClientMock.On("TransactionByHash", mock.Anything, mock.Anything).Return(&types.Transaction{}, false, nil)
	receipt := &types.Receipt{
		Status:      uint64(1),
		BlockNumber: big.NewInt(1),
	}
	gethClientMock.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receipt, nil)

	ethMock := &interfaces.EthereumMock{}
	ethMock.On("GetGethClient").Return(gethClientMock)
	ethMock.On("GetFinalityDelay").Return(2)
	ethMock.On("RetryCount").Return(3)
	ethMock.On("RetryDelay").Return(10 * time.Millisecond)
	ethMock.On("GetTxCheckFrequency").Return(3 * time.Second)
	ethMock.On("GetTxTimeoutForReplacement").Return(30 * time.Second)
	ethMock.On("GetCurrentHeight", mock.Anything).Return(1, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(2, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(3, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(4, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(5, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(6, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(7, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(8, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(9, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(10, nil).Once()

	tasks.StartTask(logger.WithField("Task", 0), &wg, ethMock, dkgTaskMock, nil, nil)

=======
	tasks.StartTask(mocks.NewMockLogger().WithField("Task", 0), &wg, eth, task, objects.NewDkgState(accounts.Account{}), nil)
>>>>>>> upstream/candidate
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
<<<<<<< HEAD

	gethClientMock := &interfaces.GethClientMock{}
	gethClientMock.On("TransactionByHash", mock.Anything, mock.Anything).Return(&types.Transaction{}, false, nil)
	receiptOk1 := &types.Receipt{
		Status:      uint64(1),
		BlockNumber: big.NewInt(2),
	}
	receiptOk2 := &types.Receipt{
		Status:      uint64(1),
		BlockNumber: big.NewInt(int64(minedInBlock)),
	}
	gethClientMock.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receiptOk1, nil).Once()
	gethClientMock.On("TransactionReceipt", mock.Anything, mock.Anything).Return(&types.Receipt{}, errors.New("error getting receipt")).Once()
	gethClientMock.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receiptOk2, nil).Once()
	gethClientMock.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receiptOk2, nil).Once()

	ethMock := &interfaces.EthereumMock{}
	ethMock.On("GetGethClient").Return(gethClientMock)
	ethMock.On("GetFinalityDelay").Return(2)
	ethMock.On("RetryCount").Return(3)
	ethMock.On("RetryDelay").Return(10 * time.Millisecond)
	ethMock.On("GetTxCheckFrequency").Return(3 * time.Second)
	ethMock.On("GetTxTimeoutForReplacement").Return(30 * time.Second)
	ethMock.On("GetCurrentHeight", mock.Anything).Return(1, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(2, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(3, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(4, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(5, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(6, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(7, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(8, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(9, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(10, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(11, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(12, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(13, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(14, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(15, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(16, nil).Once()

	tasks.StartTask(logger.WithField("Task", 0), &wg, ethMock, dkgTaskMock, nil, nil)

=======
	tasks.StartTask(mocks.NewMockLogger().WithField("Task", 0), &wg, eth, task, objects.NewDkgState(accounts.Account{}), nil)
>>>>>>> upstream/candidate
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
<<<<<<< HEAD

	gethClientMock := &interfaces.GethClientMock{}
	gethClientMock.On("TransactionByHash", mock.Anything, mock.Anything).Return(&types.Transaction{}, false, nil)
	receiptFailedStatus := &types.Receipt{
		Status: uint64(0),
	}
	receiptOk := &types.Receipt{
		Status:      uint64(1),
		BlockNumber: big.NewInt(int64(minedInBlock)),
	}
	gethClientMock.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receiptFailedStatus, nil).Once()
	gethClientMock.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receiptOk, nil).Once()
	gethClientMock.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receiptOk, nil).Once()

	ethMock := &interfaces.EthereumMock{}
	ethMock.On("GetGethClient").Return(gethClientMock)
	ethMock.On("GetFinalityDelay").Return(2)
	ethMock.On("RetryCount").Return(3)
	ethMock.On("RetryDelay").Return(10 * time.Millisecond)
	ethMock.On("GetTxCheckFrequency").Return(3 * time.Second)
	ethMock.On("GetTxTimeoutForReplacement").Return(30 * time.Second)
	ethMock.On("GetCurrentHeight", mock.Anything).Return(1, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(2, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(3, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(4, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(5, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(6, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(7, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(8, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(9, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(10, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(11, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(12, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(13, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(14, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(15, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(16, nil).Once()

	tasks.StartTask(logger.WithField("Task", 0), &wg, ethMock, dkgTaskMock, nil, nil)

=======
	tasks.StartTask(mocks.NewMockLogger().WithField("Task", 0), &wg, eth, task, objects.NewDkgState(accounts.Account{}), nil)
>>>>>>> upstream/candidate
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
<<<<<<< HEAD

	gethClientMock := &interfaces.GethClientMock{}
	gethClientMock.On("TransactionByHash", mock.Anything, mock.Anything).Return(&types.Transaction{}, true, nil).Once()
	gethClientMock.On("TransactionByHash", mock.Anything, mock.Anything).Return(&types.Transaction{}, true, nil).Once()
	gethClientMock.On("TransactionByHash", mock.Anything, mock.Anything).Return(&types.Transaction{}, false, nil)
	receiptOk := &types.Receipt{
		Status:      uint64(1),
		BlockNumber: big.NewInt(int64(minedInBlock)),
	}
	gethClientMock.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receiptOk, nil)

	ethMock := &interfaces.EthereumMock{}
	ethMock.On("GetGethClient").Return(gethClientMock)
	ethMock.On("GetFinalityDelay").Return(2)
	ethMock.On("RetryCount").Return(3)
	ethMock.On("RetryDelay").Return(1 * time.Millisecond)
	ethMock.On("GetTxCheckFrequency").Return(3 * time.Second)
	ethMock.On("GetTxTimeoutForReplacement").Return(6 * time.Second)
	ethMock.On("GetTxFeePercentageToIncrease").Return(43)
	ethMock.On("GetTxMaxFeeThresholdInGwei").Return(uint64(1000000))

	ethMock.On("GetCurrentHeight", mock.Anything).Return(1, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(2, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(3, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(4, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(5, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(6, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(7, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(8, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(9, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(10, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(11, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(12, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(13, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(14, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(15, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(16, nil).Once()

	tasks.StartTask(logger.WithField("Task", 0), &wg, ethMock, dkgTaskMock, nil, nil)

	expectedGasFeeCap := big.NewInt(203569)
	expectedGasTipCap := big.NewInt(52)

=======
	tasks.StartTask(mocks.NewMockLogger().WithField("Task", 0), &wg, eth, task, objects.NewDkgState(accounts.Account{}), nil)
>>>>>>> upstream/candidate
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
<<<<<<< HEAD

	gethClientMock := &interfaces.GethClientMock{}
	gethClientMock.On("TransactionByHash", mock.Anything, mock.Anything).Return(&types.Transaction{}, true, nil).Once()
	gethClientMock.On("TransactionByHash", mock.Anything, mock.Anything).Return(&types.Transaction{}, true, nil).Once()
	gethClientMock.On("TransactionByHash", mock.Anything, mock.Anything).Return(&types.Transaction{}, false, nil)
	receiptOk := &types.Receipt{
		Status:      uint64(1),
		BlockNumber: big.NewInt(int64(minedInBlock)),
	}
	gethClientMock.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receiptOk, nil)

	ethMock := &interfaces.EthereumMock{}
	ethMock.On("GetGethClient").Return(gethClientMock)
	ethMock.On("GetFinalityDelay").Return(2)
	ethMock.On("RetryCount").Return(3)
	ethMock.On("RetryDelay").Return(1 * time.Millisecond)
	ethMock.On("GetTxCheckFrequency").Return(3 * time.Second)
	ethMock.On("GetTxTimeoutForReplacement").Return(6 * time.Second)
	ethMock.On("GetTxFeePercentageToIncrease").Return(143)
	ethMock.On("GetTxMaxFeeThresholdInGwei").Return(uint64(200000))

	ethMock.On("GetCurrentHeight", mock.Anything).Return(1, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(2, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(3, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(4, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(5, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(6, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(7, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(8, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(9, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(10, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(11, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(12, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(13, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(14, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(15, nil).Once()
	ethMock.On("GetCurrentHeight", mock.Anything).Return(16, nil).Once()

	tasks.StartTask(logger.WithField("Task", 0), &wg, ethMock, dkgTaskMock, nil, nil)

	expectedGasFeeCap := big.NewInt(200000)
	expectedGasTipCap := big.NewInt(89)

=======
	tasks.StartTask(mocks.NewMockLogger().WithField("Task", 0), &wg, eth, task, objects.NewDkgState(accounts.Account{}), nil)
>>>>>>> upstream/candidate
	wg.Wait()

	expectedGasFeeCap := big.NewInt(1000000)

	mockrequire.Called(t, task.DoWorkFunc)
	mockrequire.Called(t, task.DoDoneFunc)
	assert.Equal(t, expectedGasFeeCap, task.ExecutionData.TxOpts.GasFeeCap)
}
