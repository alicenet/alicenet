package interfaces

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"math/big"
	"time"
)

//
// Mock implementation of interfaces.Ethereum
//
type EthereumMock struct {
	mock.Mock
}

func (eth *EthereumMock) ChainID() *big.Int {
	args := eth.Called()
	return big.NewInt(int64(args.Int(0)))
}

func (eth *EthereumMock) GetFinalityDelay() uint64 {
	args := eth.Called()
	return uint64(args.Int(0))
}

func (eth *EthereumMock) Close() error {
	args := eth.Called()
	return args.Error(0)
}

func (eth *EthereumMock) Commit() {

}

func (eth *EthereumMock) IsEthereumAccessible() bool {
	args := eth.Called()
	return args.Bool(0)
}

func (eth *EthereumMock) GetCallOpts(ctx context.Context, acc accounts.Account) *bind.CallOpts {
	args := eth.Called(ctx, acc)
	return args.Get(0).(*bind.CallOpts)
}

func (eth *EthereumMock) GetTransactionOpts(ctx context.Context, acc accounts.Account) (*bind.TransactOpts, error) {
	args := eth.Called(ctx, acc)
	return args.Get(0).(*bind.TransactOpts), args.Error(1)
}

func (eth *EthereumMock) LoadAccounts(string) {}

func (eth *EthereumMock) LoadPasscodes(pc string) error {
	args := eth.Called(pc)
	return args.Error(1)
}

func (eth *EthereumMock) UnlockAccount(acc accounts.Account) error {
	args := eth.Called(acc)
	return args.Error(1)
}

func (eth *EthereumMock) UnlockAccountWithPasscode(acc accounts.Account, pc string) error {
	args := eth.Called(acc, pc)
	return args.Error(1)
}

func (eth *EthereumMock) TransferEther(addr1 common.Address, addr2 common.Address, ether *big.Int) (*types.Transaction, error) {
	args := eth.Called(addr1, addr2, ether)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (eth *EthereumMock) GetAccount(addr common.Address) (accounts.Account, error) {
	args := eth.Called(addr)
	return args.Get(0).(accounts.Account), args.Error(1)
}
func (eth *EthereumMock) GetAccountKeys(addr common.Address) (*keystore.Key, error) {
	args := eth.Called(addr)
	return args.Get(0).(*keystore.Key), args.Error(1)
}
func (eth *EthereumMock) GetBalance(addr common.Address) (*big.Int, error) {
	args := eth.Called(addr)
	return big.NewInt(int64(args.Int(0))), args.Error(1)
}
func (eth *EthereumMock) GetGethClient() GethClient {
	args := eth.Called()
	return args.Get(0).(GethClient)
}

func (eth *EthereumMock) GetCoinbaseAddress() common.Address {
	args := eth.Called()
	return args.Get(0).(common.Address)
}

func (eth *EthereumMock) GetCurrentHeight(ctx context.Context) (uint64, error) {
	args := eth.Called(ctx)
	return uint64(args.Int(0)), args.Error(1)
}

func (eth *EthereumMock) GetDefaultAccount() accounts.Account {
	args := eth.Called()
	return args.Get(0).(accounts.Account)
}
func (eth *EthereumMock) GetEndpoint() string {
	args := eth.Called()
	return args.String(0)
}
func (eth *EthereumMock) GetEvents(ctx context.Context, firstBlock uint64, lastBlock uint64, addresses []common.Address) ([]types.Log, error) {
	args := eth.Called(ctx, firstBlock, lastBlock, addresses)
	return args.Get(0).([]types.Log), args.Error(1)
}
func (eth *EthereumMock) GetFinalizedHeight(ctx context.Context) (uint64, error) {
	args := eth.Called(ctx)
	return uint64(args.Int(0)), args.Error(1)
}
func (eth *EthereumMock) GetPeerCount(ctx context.Context) (uint64, error) {
	args := eth.Called(ctx)
	return uint64(args.Int(0)), args.Error(1)
}
func (eth *EthereumMock) GetSnapshot() ([]byte, error) {
	args := eth.Called()
	return args.Get(0).([]byte), args.Error(1)
}
func (eth *EthereumMock) GetSyncProgress() (bool, *ethereum.SyncProgress, error) {
	args := eth.Called()
	return args.Bool(0), args.Get(1).(*ethereum.SyncProgress), args.Error(2)
}
func (eth *EthereumMock) GetTimeoutContext() (context.Context, context.CancelFunc) {
	args := eth.Called()
	return args.Get(0).(context.Context), args.Get(1).(context.CancelFunc)
}
func (eth *EthereumMock) GetValidators(ctx context.Context) ([]common.Address, error) {
	args := eth.Called(ctx)
	return args.Get(0).([]common.Address), args.Error(1)
}

func (eth *EthereumMock) GetKnownAccounts() []accounts.Account {
	args := eth.Called()
	return args.Get(0).([]accounts.Account)
}

func (eth *EthereumMock) KnownSelectors() SelectorMap {
	args := eth.Called()
	return args.Get(0).(SelectorMap)
}

func (eth *EthereumMock) Queue() TxnQueue {
	args := eth.Called()
	return args.Get(0).(TxnQueue)
}

func (eth *EthereumMock) RetryCount() int {
	args := eth.Called()
	return args.Int(0)
}
func (eth *EthereumMock) RetryDelay() time.Duration {
	args := eth.Called()
	return args.Get(0).(time.Duration)
}

func (eth *EthereumMock) Timeout() time.Duration {
	args := eth.Called()
	return args.Get(0).(time.Duration)
}

func (eth *EthereumMock) Contracts() Contracts {
	args := eth.Called()
	return args.Get(0).(Contracts)
}

func (eth *EthereumMock) GetTxFeePercentageToIncrease() int {
	args := eth.Called()
	return args.Int(0)
}

func (eth *EthereumMock) GetTxMaxFeeThresholdInGwei() uint64 {
	args := eth.Called()
	return uint64(args.Int(0))
}

func (eth *EthereumMock) GetTxCheckFrequency() time.Duration {
	args := eth.Called()
	return args.Get(0).(time.Duration)
}

func (eth *EthereumMock) GetTxTimeoutForReplacement() time.Duration {
	args := eth.Called()
	return args.Get(0).(time.Duration)
}

type GethClientMock struct {
	mock.Mock
}

func (e *GethClientMock) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	args := e.Called(ctx, hash)
	return args.Get(0).(*types.Block), args.Error(1)
}
func (e *GethClientMock) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	args := e.Called(ctx, number)
	return args.Get(0).(*types.Block), args.Error(1)
}
func (e *GethClientMock) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	args := e.Called(ctx, hash)
	return args.Get(0).(*types.Header), args.Error(1)
}
func (e *GethClientMock) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	args := e.Called(ctx, number)
	return args.Get(0).(*types.Header), args.Error(1)
}
func (e *GethClientMock) TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error) {
	args := e.Called(ctx, blockHash)
	return uint(args.Int(0)), args.Error(1)
}
func (e *GethClientMock) TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*types.Transaction, error) {
	args := e.Called(ctx, blockHash, index)
	return args.Get(0).(*types.Transaction), args.Error(1)
}
func (e *GethClientMock) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	args := e.Called(ctx, ch)
	return args.Get(0).(ethereum.Subscription), args.Error(1)
}

func (e *GethClientMock) TransactionByHash(ctx context.Context, txHash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	args := e.Called(ctx, txHash)
	return args.Get(0).(*types.Transaction), args.Bool(1), args.Error(2)
}
func (e *GethClientMock) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	args := e.Called(ctx, txHash)
	return args.Get(0).(*types.Receipt), args.Error(1)
}

func (e *GethClientMock) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	args := e.Called(ctx, account, blockNumber)
	return big.NewInt(int64(args.Int(0))), args.Error(1)
}
func (e *GethClientMock) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	args := e.Called(ctx, account, key, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}
func (e *GethClientMock) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	args := e.Called(ctx, account, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}
func (e *GethClientMock) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	args := e.Called(ctx, account, blockNumber)
	return uint64(args.Int(0)), args.Error(1)
}
func (e *GethClientMock) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	args := e.Called(ctx, call, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}
func (e *GethClientMock) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	args := e.Called(ctx, account)
	return args.Get(0).([]byte), args.Error(1)
}
func (e *GethClientMock) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	args := e.Called(ctx, account)
	return uint64(args.Int(0)), args.Error(1)
}
func (e *GethClientMock) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	args := e.Called(ctx)
	return big.NewInt(int64(args.Int(0))), args.Error(1)
}
func (e *GethClientMock) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	args := e.Called(ctx)
	return big.NewInt(int64(args.Int(0))), args.Error(1)
}
func (e *GethClientMock) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	args := e.Called(ctx, call)
	return uint64(args.Int(0)), args.Error(1)
}
func (e *GethClientMock) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	args := e.Called(ctx, tx)
	return args.Error(0)
}
func (e *GethClientMock) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	args := e.Called(ctx, query)
	return args.Get(0).([]types.Log), args.Error(1)
}

func (e *GethClientMock) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	args := e.Called(ctx, query, ch)
	return args.Get(0).(ethereum.Subscription), args.Error(1)
}
