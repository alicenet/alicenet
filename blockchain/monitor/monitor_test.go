package monitor_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"sync"
	"testing"
	"time"

	aobjs "github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/etest"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/monitor"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func setupEthereum(t *testing.T, mineInterval time.Duration) interfaces.Ethereum {

	n := 4
	privKeys := etest.SetupPrivateKeys(n)
	eth, err := blockchain.NewEthereumSimulator(
		privKeys,
		1,
		time.Second*2,
		time.Second*5,
		0,
		big.NewInt(math.MaxInt64))
	assert.Nil(t, err, "Failed to build Ethereum endpoint...")
	assert.True(t, eth.IsEthereumAccessible(), "Web3 endpoint is not available.")
	defer eth.Close()

	c := eth.Contracts()

	go func() {
		for {
			time.Sleep(mineInterval)
			eth.Commit()
		}
	}()

	// Unlock deploy account and make sure it has a balance
	acct := eth.GetDefaultAccount()
	err = eth.UnlockAccount(acct)
	assert.Nil(t, err, "Failed to unlock deploy account")

	_, _, err = c.DeployContracts(context.TODO(), acct)
	assert.Nil(t, err, "Failed to deploy contracts...")

	return eth
}

//
//
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

//
// Mock implementation of interfaces.AdminHandler
//
type mockAdminHandler struct {
}

func (ah *mockAdminHandler) AddPrivateKey([]byte, constants.CurveSpec) error {
	return nil
}

func (ah *mockAdminHandler) AddSnapshot(*objs.BlockHeader, bool) error {
	return nil
}
func (ah *mockAdminHandler) AddValidatorSet(*objs.ValidatorSet) error {
	return nil
}

func (ah *mockAdminHandler) RegisterSnapshotCallback(func(*objs.BlockHeader) error) {

}

func (ah *mockAdminHandler) SetSynchronized(v bool) {

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
// Mock implementation of interfaces.Ethereum
//
type mockEthereum struct {
}

func (eth *mockEthereum) ChainID() *big.Int {
	return nil
}

func (eth *mockEthereum) GetFinalityDelay() uint64 {
	return 12
}

func (eth *mockEthereum) Close() error {
	return nil
}

func (eth *mockEthereum) Commit() {

}

func (eth *mockEthereum) IsEthereumAccessible() bool {
	return false
}

func (eth *mockEthereum) GetCallOpts(context.Context, accounts.Account) *bind.CallOpts {
	return nil
}

func (eth *mockEthereum) GetTransactionOpts(context.Context, accounts.Account) (*bind.TransactOpts, error) {
	return nil, nil
}

func (eth *mockEthereum) LoadAccounts(string) {}

func (eth *mockEthereum) LoadPasscodes(string) error {
	return nil
}

func (eth *mockEthereum) UnlockAccount(accounts.Account) error {
	return nil
}

func (eth *mockEthereum) UnlockAccountWithPasscode(accounts.Account, string) error {
	return nil
}

func (eth *mockEthereum) TransferEther(common.Address, common.Address, *big.Int) (*types.Transaction, error) {
	return nil, nil
}

func (eth *mockEthereum) GetAccount(addr common.Address) (accounts.Account, error) {
	return accounts.Account{Address: addr}, nil
}
func (eth *mockEthereum) GetAccountKeys(addr common.Address) (*keystore.Key, error) {
	return nil, nil
}
func (eth *mockEthereum) GetBalance(common.Address) (*big.Int, error) {
	return nil, nil
}
func (eth *mockEthereum) GetGethClient() interfaces.GethClient {
	return nil
}

func (eth *mockEthereum) GetCoinbaseAddress() common.Address {
	return eth.GetDefaultAccount().Address
}

func (eth *mockEthereum) GetCurrentHeight(context.Context) (uint64, error) {
	return 0, nil
}

func (eth *mockEthereum) GetDefaultAccount() accounts.Account {
	return accounts.Account{}
}
func (eth *mockEthereum) GetEndpoint() string {
	return "na"
}
func (eth *mockEthereum) GetEvents(ctx context.Context, firstBlock uint64, lastBlock uint64, addresses []common.Address) ([]types.Log, error) {
	return nil, nil
}
func (eth *mockEthereum) GetFinalizedHeight(context.Context) (uint64, error) {
	return 0, nil
}
func (eth *mockEthereum) GetPeerCount(context.Context) (uint64, error) {
	return 0, nil
}
func (eth *mockEthereum) GetSnapshot() ([]byte, error) {
	return nil, nil
}
func (eth *mockEthereum) GetSyncProgress() (bool, *ethereum.SyncProgress, error) {
	return false, nil, nil
}
func (eth *mockEthereum) GetTimeoutContext() (context.Context, context.CancelFunc) {
	return nil, nil
}
func (eth *mockEthereum) GetValidators(context.Context) ([]common.Address, error) {
	return nil, nil
}

func (eth *mockEthereum) GetKnownAccounts() []accounts.Account {
	return []accounts.Account{}
}

func (eth *mockEthereum) KnownSelectors() interfaces.SelectorMap {
	return nil
}

func (eth *mockEthereum) Queue() interfaces.TxnQueue {
	return nil
}

func (eth *mockEthereum) RetryCount() int {
	return 0
}
func (eth *mockEthereum) RetryDelay() time.Duration {
	return time.Second
}

func (eth *mockEthereum) Timeout() time.Duration {
	return time.Second
}

func (eth *mockEthereum) Contracts() interfaces.Contracts {
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

	mon, err := monitor.NewMonitor(database, database, &mockAdminHandler{}, &mockDepositHandler{}, &mockEthereum{}, 1*time.Second, time.Minute, 1)
	assert.Nil(t, err)

	addr0 := common.HexToAddress("0x546F99F244b7B58B855330AE0E2BC1b30b41302F")
	EPOCH := uint32(1)
	populateMonitor(mon.State, addr0, EPOCH)
	raw, err := json.Marshal(mon)
	assert.Nil(t, err)
	t.Logf("Raw: %v", string(raw))

	mon.PersistState()

	//
	newMon, err := monitor.NewMonitor(database, database, &mockAdminHandler{}, &mockDepositHandler{}, &mockEthereum{}, 1*time.Second, time.Minute, 1)
	assert.Nil(t, err)

	newMon.LoadState()

	newRaw, err := json.Marshal(mon)
	assert.Nil(t, err)
	t.Logf("NewRaw: %v", string(newRaw))
}

func TestBidirectionalMarshaling(t *testing.T) {

	// setup
	adminHandler := &mockAdminHandler{}
	depositHandler := &mockDepositHandler{}
	eth := &mockEthereum{}
	logger := logging.GetLogger("test")

	addr0 := common.HexToAddress("0x546F99F244b7B58B855330AE0E2BC1b30b41302F")

	EPOCH := uint32(1)

	// Setup monitor state
	mon, err := monitor.NewMonitor(&db.Database{}, &db.Database{}, adminHandler, depositHandler, eth, 2*time.Second, time.Minute, 1)
	assert.Nil(t, err)
	populateMonitor(mon.State, addr0, EPOCH)

	// Schedule some tasks
	_, err = mon.State.Schedule.Schedule(1, 2, &mockTask{})
	assert.Nil(t, err)

	_, err = mon.State.Schedule.Schedule(3, 4, &mockTask{})
	assert.Nil(t, err)

	_, err = mon.State.Schedule.Schedule(5, 6, &mockTask{})
	assert.Nil(t, err)

	_, err = mon.State.Schedule.Schedule(7, 8, &mockTask{})
	assert.Nil(t, err)

	// Marshal
	mon.TypeRegistry.RegisterInstanceType(&mockTask{})
	raw, err := json.Marshal(mon)
	assert.Nil(t, err)
	t.Logf("RawData:%v", string(raw))

	// Unmarshal
	newMon, err := monitor.NewMonitor(&db.Database{}, &db.Database{}, adminHandler, depositHandler, eth, 2*time.Second, time.Minute, 1)
	assert.Nil(t, err)

	newMon.TypeRegistry.RegisterInstanceType(&mockTask{})
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
	tasks.StartTask(logger.WithField("Task", "Mocked"), wg, eth, task, nil)
	wg.Wait()

	assert.True(t, taskStruct.DoneCalled)
}

func TestWrapDoNotContinue(t *testing.T) {
	genErr := objects.ErrCanNotContinue
	specErr := errors.New("neutrinos")

	niceErr := errors.Wrapf(genErr, "Caused by %v", specErr)
	assert.True(t, errors.Is(niceErr, genErr))

	t.Logf("NiceErr: %v", niceErr)

	nice2Err := fmt.Errorf("%w because %v", genErr, specErr)
	assert.True(t, errors.Is(nice2Err, genErr))

	t.Logf("Nice2Err: %v", nice2Err)
}
