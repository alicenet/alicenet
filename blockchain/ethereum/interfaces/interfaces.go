package interfaces

import (
	"context"
	"math/big"
	"time"

	transactionInterfaces "github.com/MadBase/MadNet/blockchain/transaction/interfaces"
	"github.com/MadBase/MadNet/bridge/bindings"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

//IEthereum contains state information about a connection to the Ethereum node
type IEthereum interface {

	// Extensions for use with simulator
	ChainID() *big.Int
	Close() error
	Commit()

	IsEthereumAccessible() bool

	GetCallOpts(context.Context, accounts.Account) (*bind.CallOpts, error)
	GetCallOptsLatestBlock(ctx context.Context, account accounts.Account) *bind.CallOpts
	GetTransactionOpts(context.Context, accounts.Account) (*bind.TransactOpts, error)

	UnlockAccount(accounts.Account) error
	UnlockAccountWithPasscode(accounts.Account, string) error

	TransferEther(common.Address, common.Address, *big.Int) (*types.Transaction, error)

	GetAccount(common.Address) (accounts.Account, error)
	GetAccountKeys(addr common.Address) (*keystore.Key, error)
	GetBalance(common.Address) (*big.Int, error)
	GetEthereumClient() IEthereumClient
	GetCoinbaseAddress() common.Address
	GetCurrentHeight(context.Context) (uint64, error)
	GetDefaultAccount() accounts.Account
	GetEndpoint() string
	GetEvents(ctx context.Context, firstBlock uint64, lastBlock uint64, addresses []common.Address) ([]types.Log, error)
	GetFinalizedHeight(context.Context) (uint64, error)
	GetKnownAccounts() []accounts.Account
	GetPeerCount(context.Context) (uint64, error)
	GetSnapshot() ([]byte, error)
	GetSyncProgress() (bool, *ethereum.SyncProgress, error)
	GetTimeoutContext() (context.Context, context.CancelFunc)
	GetValidators(context.Context) ([]common.Address, error)
	SetFinalityDelay(uint64)
	GetFinalityDelay() uint64

	KnownSelectors() transactionInterfaces.ISelectorMap
	TransactionWatcher() transactionInterfaces.IWatcher

	RetryCount() int
	RetryDelay() time.Duration

	Timeout() time.Duration

	GetTxFeePercentageToIncrease() int
	GetTxMaxGasFeeAllowedInGwei() uint64

	Contracts() IContracts
}

type IEthereumClient interface {

	// geth.ChainReader
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	BlockNumber(ctx context.Context) (uint64, error)
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

	CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)

	// -- bind.ContractTransactor
	PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error

	// -- bind.ContractFilterer
	FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error)
	SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error)
}

// IContracts contains bindings to smart contract system
type IContracts interface {
	LookupContracts(ctx context.Context, registryAddress common.Address) error

	Ethdkg() bindings.IETHDKG
	EthdkgAddress() common.Address
	AToken() bindings.IAToken
	ATokenAddress() common.Address
	BToken() bindings.IBToken
	BTokenAddress() common.Address
	PublicStaking() bindings.IPublicStaking
	PublicStakingAddress() common.Address
	ValidatorStaking() bindings.IValidatorStaking
	ValidatorStakingAddress() common.Address
	ContractFactory() bindings.IAliceNetFactory
	ContractFactoryAddress() common.Address
	SnapshotsAddress() common.Address
	Snapshots() bindings.ISnapshots
	ValidatorPool() bindings.IValidatorPool
	ValidatorPoolAddress() common.Address
	Governance() bindings.IGovernance
	GovernanceAddress() common.Address
}
