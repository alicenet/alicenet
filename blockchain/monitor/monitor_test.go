//go:build integration

package monitor_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/dkg/dtest"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"

	aobjs "github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/monitor"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/test/mocks"

	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

//
func createSharedKey(addr common.Address) [4]*big.Int {

	b := addr.Bytes()

	return [4]*big.Int{
		(&big.Int{}).SetBytes(b),
		(&big.Int{}).SetBytes(b),
		(&big.Int{}).SetBytes(b),
		(&big.Int{}).SetBytes(b)}
}

func createValidator(addrHex string, idx uint8) objects.Validator {
	addr := common.HexToAddress(addrHex)
	return objects.Validator{
		Account:   addr,
		Index:     idx,
		SharedKey: createSharedKey(addr),
	}
}

func populateMonitor(state *objects.MonitorState, addr0 common.Address, EPOCH uint32) {
	state.EthDKG.Account = accounts.Account{
		Address: addr0,
		URL: accounts.URL{
			Scheme: "keystore",
			Path:   ""}}
	state.EthDKG.Index = 1
	state.EthDKG.SecretValue = big.NewInt(512)
	meAsAParticipant := &objects.Participant{
		Address: state.EthDKG.Account.Address,
		Index:   state.EthDKG.Index,
	}
	state.EthDKG.Participants[addr0] = meAsAParticipant
	state.EthDKG.Participants[addr0].GPKj = [4]*big.Int{
		big.NewInt(44), big.NewInt(33), big.NewInt(22), big.NewInt(11)}
	state.EthDKG.Participants[addr0].Commitments = make([][2]*big.Int, 3)
	state.EthDKG.Participants[addr0].Commitments[0][0] = big.NewInt(5)
	state.EthDKG.Participants[addr0].Commitments[0][1] = big.NewInt(2)

	state.ValidatorSets[EPOCH] = objects.ValidatorSet{
		ValidatorCount:        4,
		NotBeforeMadNetHeight: 321,
		GroupKey:              [4]*big.Int{big.NewInt(3), big.NewInt(2), big.NewInt(1), big.NewInt(5)}}

	state.Validators[EPOCH] = []objects.Validator{
		createValidator("0x546F99F244b7B58B855330AE0E2BC1b30b41302F", 1),
		createValidator("0x9AC1c9afBAec85278679fF75Ef109217f26b1417", 2),
		createValidator("0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac", 3),
		createValidator("0x615695C4a4D6a60830e5fca4901FbA099DF26271", 4)}

}

//
// Mock implementation of interfaces.Task
//
type mockTask struct {
	DoneCalled bool
	State      *objects.DkgState
	DkgTask    *dkgtasks.ExecutionData
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

func (mt *mockTask) Initialize(context.Context, *logrus.Entry, interfaces.Ethereum, interface{}) error {
	return nil
}

func (mt *mockTask) ShouldRetry(context.Context, *logrus.Entry, interfaces.Ethereum) bool {
	return false
}

func (mt *mockTask) GetExecutionData() interface{} {
	return mt.DkgTask
}

//
// Mock implementation of interfaces.DepositHandler
//
type mockDepositHandler struct {
}

func (dh *mockDepositHandler) Add(*badger.Txn, uint32, []byte, *big.Int, *aobjs.Owner) error {
	return nil
}

//
// Actual tests
//
func TestMonitorPersist(t *testing.T) {
	rawDb, err := utils.OpenBadger(context.Background().Done(), "", true)
	assert.Nil(t, err)

	database := &db.Database{}
	database.Init(rawDb)

	eth := mocks.NewMockBaseEthereum()
	mon, err := monitor.NewMonitor(database, database, mocks.NewMockAdminHandler(), &mockDepositHandler{}, eth, 1*time.Second, time.Minute, 1)
	assert.Nil(t, err)

	addr0 := common.HexToAddress("0x546F99F244b7B58B855330AE0E2BC1b30b41302F")
	EPOCH := uint32(1)
	populateMonitor(mon.State, addr0, EPOCH)
	raw, err := json.Marshal(mon)
	assert.Nil(t, err)
	t.Logf("Raw: %v", string(raw))

	err = mon.PersistState()
	assert.Nil(t, err)

	//
	newMon, err := monitor.NewMonitor(database, database, mocks.NewMockAdminHandler(), &mockDepositHandler{}, eth, 1*time.Second, time.Minute, 1)
	assert.Nil(t, err)

	err = newMon.LoadState()
	assert.Nil(t, err)

	newRaw, err := json.Marshal(mon)
	assert.Nil(t, err)
	t.Logf("NewRaw: %v", string(newRaw))
}

func TestBidirectionalMarshaling(t *testing.T) {

	// setup
	adminHandler := mocks.NewMockAdminHandler()
	depositHandler := &mockDepositHandler{}

	ecdsaPrivateKeys, _ := dtest.InitializePrivateKeysAndAccounts(5)
	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()
	logger := logging.GetLogger("test")

	addr0 := common.HexToAddress("0x546F99F244b7B58B855330AE0E2BC1b30b41302F")

	EPOCH := uint32(1)

	// Setup monitor state
	mon, err := monitor.NewMonitor(&db.Database{}, &db.Database{}, adminHandler, depositHandler, eth, 2*time.Second, time.Minute, 1)
	assert.Nil(t, err)
	populateMonitor(mon.State, addr0, EPOCH)

	acct := eth.GetKnownAccounts()[0]
	state := objects.NewDkgState(acct)
	mockTsk := &mockTask{
		DkgTask: dkgtasks.NewExecutionData(state, 1, 40),
	}
	// Schedule some tasks
	mon.State.Schedule.Schedule(1, 2, mockTsk)
	mon.State.Schedule.Schedule(3, 4, mockTsk)
	mon.State.Schedule.Schedule(5, 6, mockTsk)
	mon.State.Schedule.Schedule(7, 8, mockTsk)

	// Marshal
	mon.TypeRegistry.RegisterInstanceType(mockTsk)
	raw, err := json.Marshal(mon)
	assert.Nil(t, err)
	t.Logf("RawData:%v", string(raw))

	// Unmarshal
	newMon, err := monitor.NewMonitor(&db.Database{}, &db.Database{}, adminHandler, depositHandler, eth, 2*time.Second, time.Minute, 1)
	assert.Nil(t, err)

	newMon.TypeRegistry.RegisterInstanceType(mockTsk)
	err = json.Unmarshal(raw, newMon)
	assert.Nil(t, err)

	// Compare raw data for mon and newMon
	newRaw, err := json.Marshal(newMon)
	assert.Nil(t, err)
	assert.Equal(t, len(raw), len(newRaw))
	t.Logf("Len(RawData): %v", len(raw))

	// Do comparisons
	validator0 := createValidator("0x546F99F244b7B58B855330AE0E2BC1b30b41302F", 1)

	assert.Equal(t, 0, validator0.SharedKey[0].Cmp(newMon.State.Validators[EPOCH][0].SharedKey[0]))
	assert.Equal(t, 0, big.NewInt(44).Cmp(newMon.State.EthDKG.Participants[addr0].GPKj[0]))

	// Compare the schedules
	_, err = newMon.State.Schedule.Find(9)
	assert.Equal(t, objects.ErrNothingScheduled, err)

	//
	taskID, err := newMon.State.Schedule.Find(1)
	assert.Nil(t, err)

	task, err := newMon.State.Schedule.Retrieve(taskID)
	assert.Nil(t, err)

	//
	taskID2, err := newMon.State.Schedule.Find(3)
	assert.Nil(t, err)

	task2, err := newMon.State.Schedule.Retrieve(taskID2)
	assert.Nil(t, err)

	taskStruct := task.(*mockTask)
	assert.False(t, taskStruct.DoneCalled)

	taskStruct2 := task2.(*mockTask)
	assert.False(t, taskStruct2.DoneCalled)

	t.Logf("State:%p State2:%p", taskStruct.State, taskStruct2.State)
	assert.Equal(t, taskStruct.State, taskStruct2.State)

	wg := &sync.WaitGroup{}
	err = tasks.StartTask(logger.WithField("Task", "Mocked"), wg, eth, task, nil, nil)
	assert.Nil(t, err)
	wg.Wait()

	assert.True(t, taskStruct.DoneCalled)
}

func TestWrapDoNotContinue(t *testing.T) {
	genErr := objects.ErrCanNotContinue
	specErr := errors.New("neutrinos")

	niceErr := fmt.Errorf("%w because %v", genErr, specErr)
	assert.True(t, errors.Is(niceErr, genErr))

	t.Logf("%v", niceErr)
}
