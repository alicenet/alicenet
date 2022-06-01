package objects

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Response of the monitoring system
type ReceiptResponse struct {
	TxnHash common.Hash    // Hash of the txs which this response belongs
	Err     error          // response error that happened during processing
	Receipt *types.Receipt // tx receipt after txConfirmationBlocks of a tx that was not queued in txGroup
}
