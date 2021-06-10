package interfaces

import (
	"context"
	"math/big"
	"time"

	"github.com/MadBase/bridge/bindings"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

//Ethereum contains state information about a connection to Ethereum
type Ethereum interface {

	// Extensions for use with simulator
	Close() error
	Commit()

	IsEthereumAccessible() bool

	GetCallOpts(context.Context, accounts.Account) *bind.CallOpts
	GetTransactionOpts(context.Context, accounts.Account) (*bind.TransactOpts, error)

	LoadAccounts(string)
	LoadPasscodes(string) error

	UnlockAccount(accounts.Account) error

	TransferEther(common.Address, common.Address, *big.Int) (*types.Transaction, error)

	GetAccount(common.Address) (accounts.Account, error)
	GetAccountKeys(addr common.Address) (*keystore.Key, error)
	GetBalance(common.Address) (*big.Int, error)
	GetGethClient() GethClient
	GetCoinbaseAddress() common.Address
	GetCurrentHeight(context.Context) (uint64, error)
	GetDefaultAccount() accounts.Account
	GetEndpoint() string
	GetEvents(ctx context.Context, firstBlock uint64, lastBlock uint64, addresses []common.Address) ([]types.Log, error)
	GetFinalizedHeight(context.Context) (uint64, error)
	GetPeerCount(context.Context) (uint64, error)
	GetSnapshot() ([]byte, error)
	GetSyncProgress() (bool, *ethereum.SyncProgress, error)
	GetTimeoutContext() (context.Context, context.CancelFunc)
	GetValidators(context.Context) ([]common.Address, error)

	KnownSelectors() SelectorMap
	Queue() TxnQueue

	RetryCount() int
	RetryDelay() time.Duration

	Timeout() time.Duration

	Contracts() Contracts
}

type GethClient interface {

	// geth.ChainReader
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error)
	TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*types.Transaction, error)
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)

	// geth.TransactionReader
	TransactionByHash(ctx context.Context, txHash common.Hash) (tx *types.Transaction, isPending bool, err error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)

	// geth.ChainStateReader
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)

	// bind.ContractBackend
	// -- bind.ContractCaller
	// CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error)
	CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)

	// -- bind.ContractTransactor
	PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error

	// -- bind.ContractFilterer
	FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error)
	SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error)
}

type TxnQueue interface {
	Close()
	QueueTransaction(ctx context.Context, txn *types.Transaction)
	QueueGroupTransaction(ctx context.Context, grp int, txn *types.Transaction)
	QueueAndWait(ctx context.Context, txn *types.Transaction) (*types.Receipt, error)
	StartLoop()
	WaitTransaction(ctx context.Context, txn *types.Transaction) (*types.Receipt, error)
	WaitGroupTransactions(ctx context.Context, grp int) ([]*types.Receipt, error)
}

type FuncSelector [4]byte

type SelectorMap interface {
	Selector(signature string) FuncSelector
	Signature(selector FuncSelector) string
}

// Contracts contains bindings to smart contract system
type Contracts interface {
	LookupContracts(ctx context.Context, registryAddress common.Address) error
	DeployContracts(ctx context.Context, account accounts.Account) (*bindings.Registry, common.Address, error)

	Crypto() *bindings.Crypto
	CryptoAddress() common.Address
	Deposit() *bindings.Deposit
	DepositAddress() common.Address
	Ethdkg() *bindings.ETHDKG
	EthdkgAddress() common.Address
	Participants() *bindings.Participants
	Registry() *bindings.Registry
	RegistryAddress() common.Address
	Snapshots() *bindings.Snapshots
	Staking() *bindings.Staking
	StakingToken() *bindings.Token
	StakingTokenAddress() common.Address
	UtilityToken() *bindings.Token
	UtilityTokenAddress() common.Address
	Validators() *bindings.Validators
	ValidatorsAddress() common.Address
}
