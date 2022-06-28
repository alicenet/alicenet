//go:build integration

package monitor

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	aobjs "github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/monitor/objects"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"
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

func populateMonitor(monitorState *objects.MonitorState, EPOCH uint32) {

	monitorState.ValidatorSets[EPOCH] = objects.ValidatorSet{
		ValidatorCount:          4,
		NotBeforeAliceNetHeight: 321,
		GroupKey:                [4]*big.Int{big.NewInt(3), big.NewInt(2), big.NewInt(1), big.NewInt(5)},
	}

	monitorState.Validators[EPOCH] = []objects.Validator{
		createValidator("0x546F99F244b7B58B855330AE0E2BC1b30b41302F", 1),
		createValidator("0x9AC1c9afBAec85278679fF75Ef109217f26b1417", 2),
		createValidator("0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac", 3),
		createValidator("0x615695C4a4D6a60830e5fca4901FbA099DF26271", 4)}

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
	database := mocks.NewTestDB()

	eth := mocks.NewMockClient()
	mon, err := NewMonitor(database, database, mocks.NewMockAdminHandler(), &mockDepositHandler{}, eth, 1*time.Second, 1, make(chan tasks.TaskRequest, 10))
	assert.Nil(t, err)

	EPOCH := uint32(1)
	populateMonitor(mon.State, EPOCH)
	raw, err := json.Marshal(mon)
	assert.Nil(t, err)
	t.Logf("Raw: %v", string(raw))

	err = mon.State.PersistState(mon.db)
	assert.Nil(t, err)

	//
	newMon, err := NewMonitor(database, database, mocks.NewMockAdminHandler(), &mockDepositHandler{}, eth, 1*time.Second, 1, make(chan tasks.TaskRequest, 10))
	assert.Nil(t, err)

	err = newMon.State.LoadState(mon.db)
	assert.Nil(t, err)

	newRaw, err := json.Marshal(mon)
	assert.Nil(t, err)
	t.Logf("NewRaw: %v", string(newRaw))
}

// TODO - fix this tes with new scheduler
//
//func TestBidirectionalMarshaling(t *testing.T) {
//
//	// setup
//	adminHandler := mocks.NewMockAdminHandler()
//	depositHandler := &mockDepositHandler{}
//
//	ecdsaPrivateKeys, _ := testutils.InitializePrivateKeysAndAccounts(5)
//	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
//	defer eth.Close()
//	logger := logging.GetLogger("test")
//
//	addr0 := common.HexToAddress("0x546F99F244b7B58B855330AE0E2BC1b30b41302F")
//
//	EPOCH := uint32(1)
//
//	// Setup monitor state
//	mon, err := monitor.NewMonitor(&db.Database{}, &db.Database{}, adminHandler, depositHandler, eth, 2*time.Second, time.Minute, 1, make(chan interfaces.ITask, 10), make(chan string, 10))
//	assert.Nil(t, err)
//	populateMonitor(mon.State, EPOCH)
//
//	acct := eth.GetKnownAccounts()[0]
//	dkgState := state.NewDkgState(acct)
//	mockTsk := &mockTask{
//		DkgBaseTask: objects2.NewTask(nil, "name", 1, 40),
//	}
//	// Schedule some tasks
//	mon.State.Schedule.Schedule(1, 2, mockTsk)
//	mon.State.Schedule.Schedule(3, 4, mockTsk)
//	mon.State.Schedule.Schedule(5, 6, mockTsk)
//	mon.State.Schedule.Schedule(7, 8, mockTsk)
//
//	// Marshal
//	mon.TypeRegistry.RegisterInstanceType(mockTsk)
//	raw, err := json.Marshal(mon)
//	assert.Nil(t, err)
//	t.Logf("RawData:%v", string(raw))
//
//	// Unmarshal
//	newMon, err := monitor.NewMonitor(&db.Database{}, &db.Database{}, adminHandler, depositHandler, eth, 2*time.Second, time.Minute, 1, make(chan interfaces.ITask, 10), make(chan string, 10))
//	assert.Nil(t, err)
//
//	tr := &marshaller.TypeRegistry{}
//	tr.RegisterInstanceType(mockTsk)
//	err = json.Unmarshal(raw, newMon)
//	assert.Nil(t, err)
//
//	// Compare raw data for mon and newMon
//	newRaw, err := json.Marshal(newMon)
//	assert.Nil(t, err)
//	assert.Equal(t, len(raw), len(newRaw))
//	t.Logf("Len(RawData): %v", len(raw))
//
//	// Do comparisons
//	validator0 := createValidator("0x546F99F244b7B58B855330AE0E2BC1b30b41302F", 1)
//
//	assert.Equal(t, 0, validator0.SharedKey[0].Cmp(newMon.State.Validators[EPOCH][0].SharedKey[0]))
//	assert.Equal(t, 0, big.NewInt(44).Cmp(newMon.State.EthDKG.Participants[addr0].GPKj[0]))
//
//	// Compare the schedules
//	_, err = newMon.State.Schedule.Find(9)
//	assert.Equal(t, tasks.ErrNothingScheduled, err)
//
//	//
//	taskID, err := newMon.State.Schedule.Find(1)
//	assert.Nil(t, err)
//
//	task, err := newMon.State.Schedule.Retrieve(taskID)
//	assert.Nil(t, err)
//
//	//
//	taskID2, err := newMon.State.Schedule.Find(3)
//	assert.Nil(t, err)
//
//	task2, err := newMon.State.Schedule.Retrieve(taskID2)
//	assert.Nil(t, err)
//
//	taskStruct := task.(*mockTask)
//	assert.False(t, taskStruct.DoneCalled)
//
//	taskStruct2 := task2.(*mockTask)
//	assert.False(t, taskStruct2.DoneCalled)
//
//	t.Logf("State:%p State2:%p", taskStruct.State, taskStruct2.State)
//	assert.Equal(t, taskStruct.State, taskStruct2.State)
//
//	wg := &sync.WaitGroup{}
//	err = tasks.StartTask(logger.WithField("Task", "Mocked"), wg, eth, task, nil, nil)
//	assert.Nil(t, err)
//	wg.Wait()
//
//	assert.True(t, taskStruct.DoneCalled)
//}

// func TestTaskPersistance(t *testing.T) {
// 	rawDb, err := utils.OpenBadger(context.Background().Done(), "", true)
// 	assert.Nil(t, err)

// 	database := &db.Database{}
// 	database.Init(rawDb)

// 	eth := mocks.NewMockBaseEthereum()
// 	mon, err := monitor.NewMonitor(database, database, mocks.NewMockAdminHandler(), &mockDepositHandler{}, eth, 1*time.Second, time.Minute, 1)
// 	assert.Nil(t, err)

// 	addr0 := common.HexToAddress("0x546F99F244b7B58B855330AE0E2BC1b30b41302F")
// 	EPOCH := uint32(1)
// 	populateMonitor(mon.State, addr0, EPOCH)

// 	snapshotTask := &tasks.SnapshotTask{
// 		BaseTask: tasks.NewTask(&tasks.SnapshotState{
// 			account:     addr0,
// 			blockHeader: bh,
// 			consensusDb: db,
// 		}, start, end),
// 	}
// 	// Schedule some tasks
// 	_, err = mon.State.Schedule.Schedule(1, 2, mockTsk)
// 	assert.Nil(t, err)

// 	_, err = mon.State.Schedule.Schedule(3, 4, mockTsk)
// 	assert.Nil(t, err)

// 	_, err = mon.State.Schedule.Schedule(5, 6, mockTsk)
// 	assert.Nil(t, err)

// 	_, err = mon.State.Schedule.Schedule(7, 8, mockTsk)
// 	assert.Nil(t, err)

// 	raw, err := json.Marshal(mon)
// 	assert.Nil(t, err)
// 	t.Logf("Raw: %v", string(raw))

// 	mon.PersistState()

// 	//
// 	newMon, err := monitor.NewMonitor(database, database, mocks.NewMockAdminHandler(), &mockDepositHandler{}, eth, 1*time.Second, time.Minute, 1)
// 	assert.Nil(t, err)

// 	newMon.LoadState()

// 	newRaw, err := json.Marshal(mon)
// 	assert.Nil(t, err)
// 	t.Logf("NewRaw: %v", string(newRaw))
// }
