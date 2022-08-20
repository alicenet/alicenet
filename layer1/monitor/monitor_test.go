//go:build integration

package monitor

import (
	"encoding/json"
	"errors"
	"github.com/alicenet/alicenet/utils"
	"math/big"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/layer1/executor"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	ethdkgState "github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	snapshotState "github.com/alicenet/alicenet/layer1/executor/tasks/snapshots/state"
	"github.com/alicenet/alicenet/layer1/monitor/events"
	"github.com/alicenet/alicenet/layer1/monitor/objects"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/test/mocks"
)

func createSharedKey(addr common.Address) [4]*big.Int {
	b := addr.Bytes()

	return [4]*big.Int{
		(&big.Int{}).SetBytes(b),
		(&big.Int{}).SetBytes(b),
		(&big.Int{}).SetBytes(b),
		(&big.Int{}).SetBytes(b),
	}
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
		createValidator("0x615695C4a4D6a60830e5fca4901FbA099DF26271", 4),
	}
}

func getMonitor(t *testing.T) (*monitor, *executor.TasksScheduler, chan tasks.TaskRequest, *mocks.MockClient, *mocks.MockEthereumContracts, accounts.Account) {
	monDB := mocks.NewTestDB()
	consDB := mocks.NewTestDB()
	adminHandler := mocks.NewMockAdminHandler()
	depositHandler := mocks.NewMockDepositHandler()
	eth := mocks.NewMockClient()
	tasksReqChan := make(chan tasks.TaskRequest, 10)
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

	contracts := mocks.NewMockAllSmartContracts()
	contracts.EthereumContractsFunc.SetDefaultReturn(ethereumContracts)

	tasksScheduler, err := executor.NewTasksScheduler(monDB, eth, contracts, adminHandler, tasksReqChan, txWatcher)
	mon, err := NewMonitor(consDB, monDB, adminHandler, depositHandler, eth, contracts, []common.Address{}, 2*time.Second, 100, tasksReqChan)
	assert.Nil(t, err)
	EPOCH := uint32(1)
	populateMonitor(mon.State, EPOCH)

	t.Cleanup(func() {
		mon.Close()
		tasksScheduler.Close()

		<-time.After(15 * time.Millisecond)
		close(tasksReqChan)
	})

	return mon, tasksScheduler, tasksReqChan, eth, ethereumContracts, account
}

// Actual tests.
func TestMonitorPersist(t *testing.T) {
	mon, _, _, eth, _, _ := getMonitor(t)
	raw, err := json.Marshal(mon)
	assert.Nil(t, err)
	t.Logf("Raw: %v", string(raw))

	err = mon.State.PersistState(mon.db)
	assert.Nil(t, err)

	newMon, err := NewMonitor(mon.db, mon.db, mocks.NewMockAdminHandler(), mocks.NewMockDepositHandler(), eth, mocks.NewMockAllSmartContracts(), []common.Address{}, 10*time.Millisecond, 100, make(chan tasks.TaskRequest, 10))
	assert.Nil(t, err)

	err = newMon.State.LoadState(mon.db)
	assert.Nil(t, err)

	newRaw, err := json.Marshal(mon)
	assert.Nil(t, err)
	t.Logf("NewRaw: %v", string(newRaw))
}

func TestProcessRegistrationOpenedEvent(t *testing.T) {
	mon, tasksScheduler, _, eth, _, defaultAcc := getMonitor(t)

	err := tasksScheduler.Start()
	assert.Nil(t, err)

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

	err = mon.Start()
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

func TestProcessNewAliceNetNodeVersionAvailableEvent(t *testing.T) {
	mon, tasksScheduler, _, eth, contracts, defaultAcc := getMonitor(t)

	err := tasksScheduler.Start()
	assert.Nil(t, err)

	mon.State.PotentialValidators[defaultAcc.Address] = objects.PotentialValidator{
		Account: defaultAcc.Address,
	}
	eth.EndpointInSyncFunc.SetDefaultReturn(true, 4, nil)
	eth.GetFinalizedHeightFunc.SetDefaultReturn(1, nil)

	localVersion := utils.GetLocalVersion()
	localVersion.Major++
	version := &bindings.DynamicsNewAliceNetNodeVersionAvailable{
		Version: localVersion,
	}
	dynamics := mocks.NewMockIDynamics()
	dynamics.ParseNewAliceNetNodeVersionAvailableFunc.SetDefaultReturn(version, nil)
	contracts.DynamicsFunc.SetDefaultReturn(dynamics)

	dynamicsEvents := events.GetDynamicsEvents()
	logs := []types.Log{
		{Topics: []common.Hash{dynamicsEvents["NewAliceNetNodeVersionAvailable"].ID}},
	}

	eth.GetEventsFunc.SetDefaultReturn(nil, nil)
	eth.GetEventsFunc.PushReturn(logs, nil)

	err = mon.Start()
	assert.Nil(t, err)

	for {
		select {
		case <-time.After(4500 * time.Millisecond):
			t.Fatal("didn't update event in time")
		default:
		}

		monState := &objects.MonitorState{}
		err := monState.LoadState(mon.db)
		if err != nil {
			assert.Equal(t, badger.ErrKeyNotFound, err)
		}

		if monState.CanonicalVersion.Major != 0 {
			assert.Equal(t, version.Version.Major, monState.CanonicalVersion.Major)
			assert.Equal(t, version.Version.Minor, monState.CanonicalVersion.Minor)
			assert.Equal(t, version.Version.Patch, monState.CanonicalVersion.Patch)
			assert.Equal(t, version.Version.ExecutionEpoch, monState.CanonicalVersion.ExecutionEpoch)
			break
		}

		<-time.After(100 * time.Millisecond)
	}
}

func TestProcessSnapshotTakenEventWithOutdatedCanonicalVersion(t *testing.T) {
	mon, _, _, eth, contracts, defaultAcc := getMonitor(t)

	mon.State.PotentialValidators[defaultAcc.Address] = objects.PotentialValidator{
		Account: defaultAcc.Address,
	}
	eth.EndpointInSyncFunc.SetDefaultReturn(true, 4, nil)
	eth.GetFinalizedHeightFunc.SetDefaultReturn(1, nil)

	snapshotTakenEvent := &bindings.SnapshotsSnapshotTaken{
		Epoch:                    big.NewInt(10),
		Height:                   big.NewInt(10240),
		ChainId:                  big.NewInt(1337),
		Validator:                defaultAcc.Address,
		IsSafeToProceedConsensus: true,
		BClaims: bindings.BClaimsParserLibraryBClaims{
			ChainId:    1337,
			Height:     10240,
			TxCount:    0,
			PrevBlock:  [32]byte{},
			TxRoot:     [32]byte{},
			StateRoot:  [32]byte{},
			HeaderRoot: [32]byte{},
		},
	}
	snapshots := mocks.NewMockISnapshots()
	snapshots.ParseSnapshotTakenFunc.SetDefaultReturn(snapshotTakenEvent, nil)
	contracts.SnapshotsFunc.SetDefaultReturn(snapshots)

	localVersion := utils.GetLocalVersion()
	localVersion.Major++
	dynamics := mocks.NewMockIDynamics()
	dynamics.GetLatestAliceNetVersionFunc.SetDefaultReturn(localVersion, nil)
	contracts.DynamicsFunc.SetDefaultReturn(dynamics)

	snapshotEvents := events.GetSnapshotEvents()
	logs := []types.Log{
		{Topics: []common.Hash{snapshotEvents["SnapshotTaken"].ID}},
	}

	eth.GetEventsFunc.SetDefaultReturn(nil, nil)
	eth.GetEventsFunc.PushReturn(logs, nil)
	eth.GetCallOptsFunc.SetDefaultReturn(nil, nil)

	err := mon.Start()
	assert.Nil(t, err)

	select {
	case <-time.After(4500 * time.Millisecond):
		t.Fatal("didn't update event in time")
	case <-mon.CloseChan():
	}
}

func TestPersistSnapshot(t *testing.T) {
	mon, tasksScheduler, tasksReqChan, eth, _, _ := getMonitor(t)
	eth.GetFinalizedHeightFunc.SetDefaultReturn(1, nil)

	err := tasksScheduler.Start()
	assert.Nil(t, err)

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
	err = PersistSnapshot(eth, bh, 10, 1, tasksReqChan, mon.db)

	state, err := snapshotState.GetSnapshotState(mon.db)
	assert.Nil(t, err)
	assert.Equal(t, uint32(height), state.BlockHeader.BClaims.Height)
}
