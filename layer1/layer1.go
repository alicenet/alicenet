package layer1

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Client contains state information about a connection to the Ethereum node
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
	GetEvents(ctx context.Context, firstBlock uint64, lastBlock uint64, addresses []common.Address) ([]types.Log, error)
	GetFinalityDelay() uint64
	GetTxMaxGasFeeAllowed() *big.Int
	GetPendingNonce(ctx context.Context, account common.Address) (uint64, error)
	SignTransaction(tx types.TxData, signerAddress common.Address) (*types.Transaction, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	ExtractTransactionSender(tx *types.Transaction) (common.Address, error)
	RetryTransaction(ctx context.Context, tx *types.Transaction, baseFee *big.Int, gasTipCap *big.Int) (*types.Transaction, error)
}

type Contracts interface {
	GetAllAddresses() []common.Address
}
