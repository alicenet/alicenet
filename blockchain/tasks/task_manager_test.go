package tasks_test

import (
	"context"
	"encoding/json"
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
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

	tasks.StartTask(logger.WithField("Task", 0), &wg, &mockEthereum{}, task0, state)
	tasks.StartTask(logger.WithField("Task", 1), &wg, &mockEthereum{}, task1, state)

	wg.Wait()

	assert.Equal(t, 2, state.count)
}

func TestIsAdminClient(t *testing.T) {
	adminInterface := reflect.TypeOf((*interfaces.AdminClient)(nil)).Elem()

	task := &dkgtasks.GPKjSubmissionTask{}
	isAdminClient := reflect.TypeOf(task).Implements(adminInterface)

	assert.True(t, isAdminClient)
}
