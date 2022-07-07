//go:build integration

package monitor

import (
	"encoding/json"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/layer1/executor"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"testing"
	"time"

	"github.com/alicenet/alicenet/layer1/monitor/objects"

	"github.com/alicenet/alicenet/test/mocks"

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
		GroupKey:                [4]*big.Int{big.NewInt(3), big.NewInt(2), big.NewInt(1), big.NewInt(5)}}

	monitorState.Validators[EPOCH] = []objects.Validator{
		createValidator("0x546F99F244b7B58B855330AE0E2BC1b30b41302F", 1),
		createValidator("0x9AC1c9afBAec85278679fF75Ef109217f26b1417", 2),
		createValidator("0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac", 3),
		createValidator("0x615695C4a4D6a60830e5fca4901FbA099DF26271", 4)}

}

func getMonitor(t *testing.T) (*monitor, *executor.TasksScheduler, chan tasks.TaskRequest, *mocks.MockClient) {
	monDB := mocks.NewTestDB()
	adminHandler := mocks.NewMockAdminHandler()
	depositHandler := mocks.NewMockDepositHandler()
	eth := mocks.NewMockClient()
	tasksReqChan := make(chan tasks.TaskRequest, 10)
	txWatcher := transaction.NewWatcher(eth, 12, monDB, false, constants.TxPollingTime)
	tasksScheduler, err := executor.NewTasksScheduler(monDB, eth, adminHandler, tasksReqChan, txWatcher)

	mon, err := NewMonitor(monDB, monDB, adminHandler, depositHandler, eth, mocks.NewMockContracts(), 2*time.Second, 100, tasksReqChan)
	assert.Nil(t, err)
	EPOCH := uint32(1)
	populateMonitor(mon.State, EPOCH)

	return mon, tasksScheduler, tasksReqChan, eth
}

//
// Actual tests
//
func TestMonitorPersist(t *testing.T) {
	mon, _, tasksReqChan, eth := getMonitor(t)
	defer close(tasksReqChan)
	raw, err := json.Marshal(mon)
	assert.Nil(t, err)
	t.Logf("Raw: %v", string(raw))

	err = mon.State.PersistState(mon.db)
	assert.Nil(t, err)

	newMon, err := NewMonitor(mon.db, mon.db, mocks.NewMockAdminHandler(), mocks.NewMockDepositHandler(), eth, mocks.NewMockContracts(), 10*time.Millisecond, 100, make(chan tasks.TaskRequest, 10))
	assert.Nil(t, err)

	err = newMon.State.LoadState(mon.db)
	assert.Nil(t, err)

	newRaw, err := json.Marshal(mon)
	assert.Nil(t, err)
	t.Logf("NewRaw: %v", string(newRaw))
}

func TestProcessEvents(t *testing.T) {
	mon, tasksScheduler, tasksReqChan, eth := getMonitor(t)

	account := accounts.Account{Address: common.Address{1, 2, 3, 4}}
	mon.State.PotentialValidators[account.Address] = objects.PotentialValidator{
		Account: account.Address,
	}
	eth.EndpointInSyncFunc.SetDefaultReturn(true, 4, nil)
	eth.GetFinalizedHeightFunc.SetDefaultReturn(100, nil)
	eth.GetDefaultAccountFunc.SetDefaultReturn(account)

	logHash := common.BytesToHash([]byte("RegistrationOpened"))
	logs := []types.Log{
		{Topics: []common.Hash{logHash}},
	}

	eth.GetEventsFunc.SetDefaultReturn(logs, nil)

	defer close(tasksReqChan)
	err := tasksScheduler.Start()
	assert.Nil(t, err)
	defer tasksScheduler.Close()

	err = mon.Start()
	assert.Nil(t, err)
	defer mon.Close()

	<-time.After(100 * time.Millisecond)

	assert.Equal(t, 2, len(tasksScheduler.Schedule))

	dkgState, err := state.GetDkgState(mon.db)
	assert.Nil(t, err)
	assert.NotNil(t, dkgState)
}
