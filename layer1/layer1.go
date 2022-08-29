package layer1

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/alicenet/alicenet/bridge/bindings"
)

// Client contains state information about a connection to the Ethereum node.
type Client interface {
	Close()
	IsAccessible() bool
	EndpointInSync(ctx context.Context) (bool, uint32, error)
	GetPeerCount(ctx context.Context) (uint64, error)
	GetChainID() *big.Int
	GetTxNotFoundMaxBlocks() uint64
	GetTxMaxStaleBlocks() uint64
	GetTransactionByHash(ctx context.Context, txHash common.Hash) (tx *types.Transaction, isPending bool, err error)
	GetTransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	GetHeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	GetBlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	GetBlockBaseFeeAndSuggestedGasTip(ctx context.Context) (*big.Int, *big.Int, error)
	GetCallOpts(context.Context, accounts.Account) (*bind.CallOpts, error)
	GetCallOptsLatestBlock(ctx context.Context, account accounts.Account) *bind.CallOpts
	GetTransactionOpts(context.Context, accounts.Account) (*bind.TransactOpts, error)
	GetAccount(common.Address) (accounts.Account, error)
	GetBalance(common.Address) (*big.Int, error)
	GetCurrentHeight(context.Context) (uint64, error)
	GetFinalizedHeight(context.Context) (uint64, error)
	GetEndpoint() string
	GetDefaultAccount() accounts.Account
	GetKnownAccounts() []accounts.Account
	GetTimeoutContext() (context.Context, context.CancelFunc)
	GetEvents(ctx context.Context, firstBlock, lastBlock uint64, addresses []common.Address) ([]types.Log, error)
	GetFinalityDelay() uint64
	GetTxMaxGasFeeAllowed() *big.Int
	GetPendingNonce(ctx context.Context, account common.Address) (uint64, error)
	SignTransaction(tx types.TxData, signerAddress common.Address) (*types.Transaction, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	ExtractTransactionSender(tx *types.Transaction) (common.Address, error)
	RetryTransaction(ctx context.Context, tx *types.Transaction, baseFee, gasTipCap *big.Int) (*types.Transaction, error)
}

type BasicContracts interface {
	GetAllAddresses() []common.Address
}

type EthereumContracts interface {
	BasicContracts
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
	Dynamics() bindings.IDynamics
	DynamicsAddress() common.Address
}

type AllSmartContracts interface {
	EthereumContracts() EthereumContracts
}
