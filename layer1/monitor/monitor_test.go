////go:build integration

package monitor

import (
	"encoding/json"
	"errors"
	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/layer1/monitor/events"
	"github.com/dgraph-io/badger/v2"
	"math/big"
	"testing"
	"time"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/layer1/executor"
	ethdkgState "github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	snapshotState "github.com/alicenet/alicenet/layer1/executor/tasks/snapshots/state"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/core/types"

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

func getMonitor(t *testing.T) (*monitor, executor.TaskHandler, *mocks.MockClient, accounts.Account) {
	monDB := mocks.NewTestDB()
	adminHandler := mocks.NewMockAdminHandler()
	depositHandler := mocks.NewMockDepositHandler()
	eth := mocks.NewMockClient()
	contracts := mocks.NewMockAllSmartContracts()
	txWatcher := transaction.NewWatcher(eth, 12, monDB, false, constants.TxPollingTime)

	account := accounts.Account{
		Address: common.HexToAddress("546F99F244b7B58B855330AE0E2BC1b30b41302F"),
		URL: accounts.URL{
			Scheme: "http",
			Path:   "",
		},
	}
	eth.GetDefaultAccountFunc.SetDefaultReturn(account)

	ethDkgMock := mocks.NewMockIETHDKG()
	event := &bindings.ETHDKGRegistrationOpened{
		StartBlock:         big.NewInt(10),
		NumberValidators:   big.NewInt(4),
		Nonce:              big.NewInt(1),
		PhaseLength:        big.NewInt(40),
		ConfirmationLength: big.NewInt(10),
		Raw:                types.Log{},
	}
	ethDkgMock.ParseRegistrationOpenedFunc.SetDefaultReturn(event, nil)

	ethereumContracts := mocks.NewMockEthereumContracts()
	ethereumContracts.EthdkgFunc.SetDefaultReturn(ethDkgMock)
	ethereumContracts.GetAllAddressesFunc.SetDefaultReturn([]common.Address{})
	contracts.EthereumContractsFunc.SetDefaultReturn(ethereumContracts)

	tasksHandler, err := executor.NewTaskHandler(monDB, eth, contracts, adminHandler, txWatcher)

	mon, err := NewMonitor(monDB, monDB, adminHandler, depositHandler, eth, contracts, contracts.EthereumContracts().GetAllAddresses(), 2*time.Second, 100, tasksHandler)
	assert.Nil(t, err)
	EPOCH := uint32(1)
	populateMonitor(mon.State, EPOCH)

	t.Cleanup(func() {
		mon.Close()
		tasksHandler.Close()
	})

	return mon, tasksHandler, eth, account
}

//
// Actual tests
//
func TestMonitorPersist(t *testing.T) {
	mon, taskHandler, eth, _ := getMonitor(t)
	raw, err := json.Marshal(mon)
	assert.Nil(t, err)
	t.Logf("Raw: %v", string(raw))

	err = mon.State.PersistState(mon.db)
	assert.Nil(t, err)

	newMon, err := NewMonitor(mon.db, mon.db, mocks.NewMockAdminHandler(), mocks.NewMockDepositHandler(), eth, mon.contracts, mon.contracts.EthereumContracts().GetAllAddresses(), 10*time.Millisecond, 100, taskHandler)
	assert.Nil(t, err)

	err = newMon.State.LoadState(mon.db)
	assert.Nil(t, err)

	newRaw, err := json.Marshal(mon)
	assert.Nil(t, err)
	t.Logf("NewRaw: %v", string(newRaw))
}

func TestProcessEvents(t *testing.T) {
	mon, taskHandler, eth, defaultAcc := getMonitor(t)

	taskHandler.Start()

	mon.State.PotentialValidators[defaultAcc.Address] = objects.PotentialValidator{
		Account: defaultAcc.Address,
	}
	eth.EndpointInSyncFunc.SetDefaultReturn(true, 4, nil)
	eth.EndpointInSyncFunc.PushReturn(false, 0, errors.New("network failure"))
	eth.GetFinalizedHeightFunc.SetDefaultReturn(1, nil)

	ethDKGEvents := events.GetETHDKGEvents()
	logs := []types.Log{
		{Topics: []common.Hash{ethDKGEvents["RegistrationOpened"].ID}},
	}

	eth.GetEventsFunc.SetDefaultReturn(nil, nil)
	eth.GetEventsFunc.PushReturn(logs, nil)

	err := mon.Start()
	assert.Nil(t, err)

	for {
		select {
		case <-time.After(4500 * time.Millisecond):
			t.Fatal("didn't update dkg state in time")
		default:
		}
		dkgState, err := ethdkgState.GetDkgState(mon.db)
		if err != nil {
			assert.Equal(t, badger.ErrKeyNotFound, err)
		}

		if dkgState != nil {
			assert.Equal(t, ethdkgState.RegistrationOpen, dkgState.Phase)
			break
		}

		<-time.After(100 * time.Millisecond)
	}
}

func TestPersistSnapshot(t *testing.T) {
	mon, taskHandler, eth, _ := getMonitor(t)
	eth.GetFinalizedHeightFunc.SetDefaultReturn(1, nil)

	taskHandler.Start()

	height := 10
	bh := &objs.BlockHeader{
		BClaims: &objs.BClaims{
			ChainID:    1,
			Height:     uint32(height),
			TxCount:    0,
			PrevBlock:  make([]byte, constants.HashLen),
			TxRoot:     crypto.Hasher([]byte{}),
			StateRoot:  make([]byte, constants.HashLen),
			HeaderRoot: make([]byte, constants.HashLen),
		},
		TxHshLst: [][]byte{},
		SigGroup: make([]byte, 192),
	}
	err := PersistSnapshot(eth, bh, taskHandler, mon.db)

	state, err := snapshotState.GetSnapshotState(mon.db)
	assert.Nil(t, err)
	assert.Equal(t, uint32(height), state.BlockHeader.BClaims.Height)
}
