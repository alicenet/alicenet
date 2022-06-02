package interfaces

import (
	"context"

	"github.com/MadBase/MadNet/blockchain/transaction/objects"
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

type ISelectorMap interface {
	Selector(signature string) objects.FuncSelector
	Signature(selector objects.FuncSelector) string
}
