package tasks_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgevents"
	"github.com/MadBase/MadNet/blockchain/dkg/dtest"
	"github.com/MadBase/MadNet/blockchain/monitor"
	"github.com/MadBase/bridge/bindings"
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
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//
// Mock implementation of interfaces.Task
//
type mockState struct {
	sync.Mutex
	count int
}

type mockTask struct {
	DoneCalled bool
	State      *mockState
}

func (mt *mockTask) DoDone(logger *logrus.Entry) {
	mt.DoneCalled = true
}

func (mt *mockTask) DoRetry(context.Context, *logrus.Entry, interfaces.Ethereum) error {
	return nil
}

func (mt *mockTask) DoWork(context.Context, *logrus.Entry, interfaces.Ethereum) error {
	return nil
}

func (mt *mockTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	dkgState := state.(*mockState)

	mt.State = dkgState

	mt.State.Lock()
	defer mt.State.Unlock()

	mt.State.count += 1

	return nil
}

func (mt *mockTask) ShouldRetry(context.Context, *logrus.Entry, interfaces.Ethereum) bool {
	return false
}

func TestFoo(t *testing.T) {
	var s map[string]string

	raw, err := json.Marshal(s)
	assert.Nil(t, err)

	t.Logf("Raw data:%v", string(raw))
}

func TestType(t *testing.T) {
	state := objects.NewDkgState(accounts.Account{})
	ct := dkgtasks.NewCompletionTask(state, 1, 10)

	var task interfaces.Task = ct
	raw, err := json.Marshal(task)
	assert.Nil(t, err)
	assert.Greater(t, len(raw), 1)

	tipe := reflect.TypeOf(task)
	t.Logf("type0:%v", tipe.String())

	if tipe.Kind() == reflect.Ptr {
		tipe = tipe.Elem()
	}

	typeName := tipe.String()
	t.Logf("type1:%v", typeName)

}

func TestSharedState(t *testing.T) {
	logger := logging.GetLogger("test")

	state := &mockState{}

	task0 := &mockTask{}
	task1 := &mockTask{}

	wg := sync.WaitGroup{}

	tasks.StartTask(logger.WithField("Task", 0), &wg, &interfaces.EthereumMock{}, task0, state)
	tasks.StartTask(logger.WithField("Task", 1), &wg, &interfaces.EthereumMock{}, task1, state)

	wg.Wait()

	assert.Equal(t, 2, state.count)
}

func TestIsAdminClient(t *testing.T) {
	adminInterface := reflect.TypeOf((*interfaces.AdminClient)(nil)).Elem()

	task := &dkgtasks.GPKjSubmissionTask{}
	isAdminClient := reflect.TypeOf(task).Implements(adminInterface)

	assert.True(t, isAdminClient)
}

func TestRegistrationOpenPhase(t *testing.T) {
	n := 5
	var phaseLength uint16 = 100
	ecdsaPrivateKeys, accounts := dtest.InitializePrivateKeysAndAccounts(n)

	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 1000*time.Millisecond)
	assert.NotNil(t, eth)

	ctx := context.Background()
	owner := accounts[0]

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	txn, err := eth.Contracts().Ethdkg().SetPhaseLength(ownerOpts, phaseLength)
	assert.Nil(t, err)
	_, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	txn, err = eth.Contracts().ValidatorPool().InitializeETHDKG(ownerOpts)
	assert.Nil(t, err)

	eth.Commit()
	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	eventMap := monitor.GetETHDKGEvents()
	eventInfo, ok := eventMap["RegistrationOpened"]
	if !ok {
		t.Fatal("event not found: RegistrationOpened")
	}
	var event *bindings.ETHDKGRegistrationOpened
	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == eventInfo.ID.String() {
			event, err = eth.Contracts().Ethdkg().ParseRegistrationOpened(*log)
			assert.Nil(t, err)
			break
		}
	}
	assert.NotNil(t, event)

	// Do Register task
	regTasks := make([]*dkgtasks.RegisterTask, n)
	dkgStates := make([]*objects.DkgState, n)
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())
	for idx := 0; idx < n; idx++ {
		// Set Registration success to true
		state, _, regTask, _ := dkgevents.UpdateStateOnRegistrationOpened(
			accounts[idx],
			event.StartBlock.Uint64(),
			event.PhaseLength.Uint64(),
			event.ConfirmationLength.Uint64(),
			event.Nonce.Uint64())

		dkgStates[idx] = state
		regTasks[idx] = regTask
	}

	err = tasks.StartTask(logger, &sync.WaitGroup{}, eth, regTasks[0], nil)
	assert.Nil(t, err)
}

func TestStartTask_initializeTask_Error(t *testing.T) {
	logger := logging.GetLogger("test")

	state := objects.NewDkgState(accounts.Account{})
	dkgTask := dkgtasks.NewDkgTaskMock(state, 1, 100)

	dkgTask.On("Initialize", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("initialize error"))

	wg := sync.WaitGroup{}

	tasks.StartTask(logger.WithField("Task", 0), &wg, &interfaces.EthereumMock{}, dkgTask, state)

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

	tasks.StartTask(logger.WithField("Task", 0), &wg, &interfaces.EthereumMock{}, dkgTask, state)

	wg.Wait()

	assert.False(t, dkgTask.Success)
	assert.Nil(t, dkgTask.TxReplOpts.Nonce)
}

func TestStartTask_handleExecutedTask_returnsNonceTooLowError(t *testing.T) {
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

	tasks.StartTask(logger.WithField("Task", 0), &wg, &interfaces.EthereumMock{}, dkgTask, state)

	wg.Wait()

	assert.False(t, dkgTask.Success)
	assert.Nil(t, dkgTask.TxReplOpts.Nonce)
}

func TestStartTask_handleExecutedTask_TxMined(t *testing.T) {
	logger := logging.GetLogger("test")

	state := objects.NewDkgState(accounts.Account{})
	dkgTaskMock := dkgtasks.NewDkgTaskMock(state, 1, 100)
	dkgTaskMock.TxReplOpts = &dkgtasks.TxReplOpts{
		Nonce: big.NewInt(1),
	}

	dkgTaskMock.On("Initialize", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	dkgTaskMock.On("DoWork", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("DoWork_error"))
	dkgTaskMock.On("ShouldRetry", mock.Anything, mock.Anything, mock.Anything).Return(true)
	dkgTaskMock.On("DoRetry", mock.Anything, mock.Anything, mock.Anything).Return(errors.New(tasks.NonceToLowError))

	wg := sync.WaitGroup{}

	gethClientMock := &interfaces.GethClientMock{}
	gethClientMock.On("TransactionByHash", mock.Anything, mock.Anything).Return(nil, false, nil)

	ethMock := &interfaces.EthereumMock{}
	ethMock.On("GetGethClient").Return(gethClientMock)

	tasks.StartTask(logger.WithField("Task", 0), &wg, ethMock, dkgTaskMock, state)

	wg.Wait()

	assert.False(t, dkgTaskMock.Success)
	assert.Nil(t, dkgTaskMock.TxReplOpts.Nonce)
}
