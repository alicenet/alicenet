package constants

import "time"

// Monitor constants.
const (
	MonitorInterval time.Duration = 7 * time.Second
	// Number of attempts that we are going to retry a certain logic in the monitoring service.
	MonitorRetryCount uint64 = 10
	// How much time we are going to wait for retrying a certain logic in the monitoring service.
	MonitorRetryDelay time.Duration = 5 * time.Second
	// Monitor timeout for retrying a certain logic in the monitoring service.
	MonitorTimeout time.Duration = 1 * time.Minute
)

// Transaction Watcher constants.
const (
	// How many blocks we should wait for removing a receipt from the cache.
	TxReceiptCacheMaxBlocks uint64 = 100
	// time which we should poll the layer1 node to check for new blocks.
	TxPollingTime time.Duration = 7 * time.Second
	// time which we should display the gas metrics.
	TxStatusTime time.Duration = 60 * time.Second
	// max timeout for all rpc call requests during an iteration.
	TxNetworkTimeout time.Duration = 2 * time.Second
	// Timeout for the monitor Tx workers.
	TxWorkerTimeout time.Duration = 4 * time.Second
	// how many times the worker tries to recover from a recoverable error when getting the receipt.
	TxWorkerMaxWorkRetries uint64 = 3
	// this constant set how many times we will be multiplying the maxStaleBlock
	// amount before trying to replace a transaction that was already replaced more
	// than once.
	TxBackOffDelayStaleTxMultiplier = 2
)

// ethereum client const.
const (
	// Percentage in which we will be increasing the miner tip cap during a transaction retry.
	EthereumTipCapPercentageBump int64 = 50
	// How many times are are going to multiply a block baseFee to compute the gas
	// fee cap. 2x should make a tx valid for the next 8 full blocks before we are
	// priced out.
	EthereumBaseFeeMultiplier int64 = 2
	// How many blocks we should wait for removing a tx in case we don't find it in the layer1 chain.
	EthereumTxNotFoundMaxBlocks uint64 = 50
	// Number of blocks to wait for a tx in the memory pool w/o returning to the caller asking for retry.
	EthereumTxMaxStaleBlocks uint64 = 6
	// Minimum value that we accept for a txMaxGasFeeAllowedInGwei config parameter.
	EthereumMinGasFeeAllowedInGwei uint64 = 300
	// Default finality delay value. Mostly used as fallback in case we cannot
	// retrieve the current finality delay used via smart contracts.
	EthereumFinalityDelay uint64 = 12
	// the maximum number of of times that we allow the bumped gas tip to be greater
	// than the suggested gas tip for a block.
	EthereumMaxGasTipMultiplier int64 = 10
)
