package interfaces

import (
	"context"

	"github.com/MadBase/MadNet/blockchain/txwatcher/objects"
	"github.com/ethereum/go-ethereum/core/types"
)

type ITransactionWatcher interface {
	StartLoop()
	Close()
	SetNumOfConfirmationBlocks(numBlocks uint64)
	SubscribeTransaction(ctx context.Context, txn *types.Transaction) (<-chan *objects.ReceiptResponse, error)
	WaitTransaction(ctx context.Context, receiptResponseChannel <-chan *objects.ReceiptResponse) (*types.Receipt, error)
	SubscribeAndWait(ctx context.Context, txn *types.Transaction) (*types.Receipt, error)
	Status(ctx context.Context) error
}

type FuncSelector [4]byte

type SelectorMap interface {
	Selector(signature string) FuncSelector
	Signature(selector FuncSelector) string
}
