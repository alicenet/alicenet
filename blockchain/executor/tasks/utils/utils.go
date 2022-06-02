package utils

import (
	"context"

	ethereumInterfaces "github.com/MadBase/MadNet/blockchain/ethereum/interfaces"
	"github.com/sirupsen/logrus"
)

// GeneralTaskShouldRetry is the general logic used to determine if a task should try again
// -- Process is
func GeneralTaskShouldRetry(ctx context.Context, logger *logrus.Entry,
	eth ethereumInterfaces.IEthereum, expectedFirstBlock uint64, expectedLastBlock uint64) bool {

	result := internalGeneralTaskShouldRetry(ctx, logger, eth, expectedFirstBlock, expectedLastBlock)

	logger.Infof("GeneralTaskRetry(expectedFirstBlock: %v expectedLastBlock:%v): %v",
		expectedFirstBlock,
		expectedLastBlock,
		result)

	return result
}

func internalGeneralTaskShouldRetry(ctx context.Context, logger *logrus.Entry,
	eth ethereumInterfaces.IEthereum, expectedFirstBlock uint64, expectedLastBlock uint64) bool {

	// Make sure we're in the right block range to continue
	currentBlock, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		// This probably means an endpoint issue, so we have to try again
		logger.Warnf("could not check current height of chain: %v", err)
		return true
	}

	if currentBlock >= expectedLastBlock {
		return false
	}

	// TODO Any other general cases where we know retry won't work?

	// We won't loop forever because eventually currentBlock > expectedLastBlock
	return true
}
