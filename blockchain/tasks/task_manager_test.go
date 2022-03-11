package tasks_test

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/stretchr/testify/assert"
)

func TestIsAdminClient(t *testing.T) {
	adminInterface := reflect.TypeOf((*interfaces.AdminClient)(nil)).Elem()

	task := &dkgtasks.GPKjSubmissionTask{}
	isAdminClient := reflect.TypeOf(task).Implements(adminInterface)

	assert.True(t, isAdminClient)
}

func TestStartTask_initializeTask_Error(t *testing.T) {
	logger := logging.GetLogger("test")

	state := objects.NewDkgState(accounts.Account{})
	dkgTask := dkgtasks.NewDkgTaskMock(state, 1, 100)

	dkgTask.On("Initialize", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("initialize error"))

	wg := sync.WaitGroup{}

	ethMock := &interfaces.EthereumMock{}
	ethMock.On("RetryCount").Return(3)
	ethMock.On("RetryDelay").Return(10 * time.Millisecond)

	tasks.StartTask(logger.WithField("Task", 0), &wg, ethMock, dkgTask, state)

	wg.Wait()

	assert.False(t, dkgTask.Success)
}

func TestStartTask_executeTask_NonceTooLowError(t *testing.T) {
	logger := logging.GetLogger("test")

	state := objects.NewDkgState(accounts.Account{})
	dkgTask := dkgtasks.NewDkgTaskMock(state, 1, 100)
	dkgTask.TxReplOpts = &dkgtasks.TxReplOpts{
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

	tasks.StartTask(logger.WithField("Task", 0), &wg, ethMock, dkgTask, state)

	wg.Wait()

	assert.False(t, dkgTask.Success)
	assert.Nil(t, dkgTask.TxReplOpts.Nonce)
}

// Happy path with mined tx present after finality delay
func TestStartTask_handleExecutedTask_TxMined1(t *testing.T) {
	logger := logging.GetLogger("test")

	state := objects.NewDkgState(accounts.Account{})
	dkgTaskMock := dkgtasks.NewDkgTaskMock(state, 1, 100)
	dkgTaskMock.TxReplOpts.TxHash = common.BigToHash(big.NewInt(123871239))

	dkgTaskMock.On("Initialize", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	dkgTaskMock.On("DoWork", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	wg := sync.WaitGroup{}

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

	tasks.StartTask(logger.WithField("Task", 0), &wg, ethMock, dkgTaskMock, state)

	wg.Wait()

	assert.False(t, dkgTaskMock.Success)
	var emptyHash [32]byte
	assert.NotEqual(t, emptyHash, dkgTaskMock.TxReplOpts.TxHash)
	assert.Equal(t, uint64(1), dkgTaskMock.TxReplOpts.MinedInBlock)
}

// Tx was mined, but it's not present after finality delay
func TestStartTask_handleExecutedTask_TxMined2(t *testing.T) {
	logger := logging.GetLogger("test")

	state := objects.NewDkgState(accounts.Account{})
	dkgTaskMock := dkgtasks.NewDkgTaskMock(state, 1, 100)
	//dkgTaskMock.TxReplOpts.TxHash = common.BigToHash(big.NewInt(123871239))

	dkgTaskMock.On("Initialize", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	dkgTaskMock.On("DoWork", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	wg := sync.WaitGroup{}

	gethClientMock := &interfaces.GethClientMock{}
	gethClientMock.On("TransactionByHash", mock.Anything, mock.Anything).Return(&types.Transaction{}, false, nil)
	receiptOk1 := &types.Receipt{
		Status:      uint64(1),
		BlockNumber: big.NewInt(2),
	}
	receiptOk2 := &types.Receipt{
		Status:      uint64(1),
		BlockNumber: big.NewInt(9),
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

	tasks.StartTask(logger.WithField("Task", 0), &wg, ethMock, dkgTaskMock, state)

	wg.Wait()

	assert.False(t, dkgTaskMock.Success)
	var emptyHash [32]byte
	assert.NotEqual(t, emptyHash, dkgTaskMock.TxReplOpts.TxHash)
	assert.Equal(t, uint64(9), dkgTaskMock.TxReplOpts.MinedInBlock)
}
