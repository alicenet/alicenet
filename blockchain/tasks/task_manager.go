package tasks

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

var (
	ErrUnknownTaskName = errors.New("unknown task name")
	ErrUnknownTaskType = errors.New("unknown task type")
)

const NonceToLowError = "nonce too low"

func StartTask(logger *logrus.Entry, wg *sync.WaitGroup, eth interfaces.Ethereum, task interfaces.Task, state interface{}, onFinishCB *func()) error {

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer task.DoDone(logger.WithField("Method", "DoDone"))
		if onFinishCB != nil {
			defer (*onFinishCB)()
		}

		retryCount := eth.RetryCount()
		retryDelay := eth.RetryDelay()
		logger.WithFields(logrus.Fields{
			"RetryCount": retryCount,
			"RetryDelay": retryDelay,
		}).Info("StartTask()...")

		// Setup
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		initializationLogger := logger.WithField("Method", "Initialize")
		err := initializeTask(ctx, logger, eth, task, state, retryCount, retryDelay)
		if err != nil {
			initializationLogger.Errorf("Failed to initialize task: %v", err)
			return
		}

		workLogger := logger.WithField("Method", "DoWork")
		err = executeTask(ctx, workLogger, eth, task, retryCount, retryDelay)
		if err != nil {
			workLogger.Error("Failed to execute task ", err)
			return
		}

		dkgLogger := logger.WithField("Method", "handleExecutedTask")
		err = handleExecutedTask(ctx, dkgLogger, eth, task)
		if err != nil {
			dkgLogger.Error("Failed to execute handleExecutedTask ", err)
		}
	}()

	return nil
}

// initializeTask initialize the Task and retry if needed
func initializeTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task, state interface{}, retryCount int, retryDelay time.Duration) error {
	var count int
	var err error

	err = task.Initialize(ctx, logger, eth, state)
	for err != nil && count < retryCount {
		if errors.Is(err, objects.ErrCanNotContinue) {
			return err
		}

		err = sleepWithContext(ctx, retryDelay)
		if err != nil {
			return err
		}

		err = task.Initialize(ctx, logger, eth, state)
		count++
	}

	return err
}

// executeTask execute the Task and retry if needed
func executeTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task, retryCount int, retryDelay time.Duration) error {
	// Clearing TxOpts used for tx gas and nonce replacement
	clearTxOpts(task)
	err := task.DoWork(ctx, logger, eth)
	if err != nil {
		retryLogger := logger.WithField("Method", "DoRetry")
		err = retryTask(ctx, retryLogger, eth, task, retryCount, retryDelay)
	}

	return err
}

func retryTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task, retryCount int, retryDelay time.Duration) error {
	var count int
	var err error
	for count < retryCount && task.ShouldRetry(ctx, logger, eth) {
		err = sleepWithContext(ctx, retryDelay)
		if err != nil {
			return err
		}

		err = task.DoRetry(ctx, logger, eth)
		if err == nil {
			break
		} else if errors.Is(err, objects.ErrCanNotContinue) {
			return err
		} else if err.Error() == NonceToLowError {
			// if we receive "nonce too low" it means that the tx was already mined
			// as a success or a fail. If is a fail we should restart with a new nonce
			clearTxOpts(task)
		}
		count++
	}

	return err
}

func retryTaskWithFeeReplacement(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task, execData *dkgtasks.ExecutionData, retryCount int, retryDelay time.Duration) error {
	logger.WithFields(logrus.Fields{
		"GasFeeCap": execData.TxOpts.GasFeeCap,
		"GasTipCap": execData.TxOpts.GasTipCap,
		"Nonce":     execData.TxOpts.Nonce,
	}).Info("retryTaskWithFeeReplacementFrom")

	// increase gas and tip cap
	gasFeeCap, gasTipCap := dkg.IncreaseFeeAndTipCap(
		execData.TxOpts.GasFeeCap,
		execData.TxOpts.GasTipCap,
		eth.GetTxFeePercentageToIncrease(),
		eth.GetTxMaxFeeThresholdInGwei())
	execData.TxOpts.GasFeeCap = gasFeeCap
	execData.TxOpts.GasTipCap = gasTipCap

	logger.WithFields(logrus.Fields{
		"GasFeeCap": execData.TxOpts.GasFeeCap,
		"GasTipCap": execData.TxOpts.GasTipCap,
		"Nonce":     execData.TxOpts.Nonce,
	}).Info("retryTaskWithFeeReplacementTo")

	return retryTask(ctx, logger, eth, task, retryCount, retryDelay)
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

// handleExecutedTask responsibilities:
// if the Tx was mined wait for FinalityDelay to confirm the Tx
// if the Tx wasn't mined during the txTimeoutForReplacement we increase the Fee
// to make sure the Tx will have priority for the next mined blocks
func handleExecutedTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task) error {
	// TxOpts or TxHash are empty means that no tx was queued, this could happen
	// if there's nobody to accuse during the dispute
	execData, ok := task.GetExecutionData().(*dkgtasks.ExecutionData)
	if !ok || execData.TxOpts == nil || execData.TxOpts.TxHashes == nil || len(execData.TxOpts.TxHashes) == 0 {
		return nil
	}

	currentBlock, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "failed to get current height %v", err)
	}

	retryCount := eth.RetryCount()
	retryDelay := eth.RetryDelay()
	txCheckFrequency := eth.GetTxCheckFrequency()
	txTimeoutForReplacement := eth.GetTxTimeoutForReplacement()
	txReplacement := getTxReplacementTime(txTimeoutForReplacement)
	taskEnd := execData.End

	logger.WithFields(logrus.Fields{
		"currentBlock":            currentBlock,
		"retryCount":              retryCount,
		"retryDelay":              retryDelay,
		"txCheckFrequency":        txCheckFrequency,
		"txTimeoutForReplacement": txTimeoutForReplacement,
		"txReplacement":           txReplacement,
		"taskEnd":                 taskEnd,
	}).Info("handleExecutedTask()...")

	var isTxConfirmed bool
	for currentBlock < taskEnd && len(execData.TxOpts.TxHashes) != 0 && !isTxConfirmed {
		err = sleepWithContext(ctx, txCheckFrequency)
		if err != nil {
			return err
		}

		isTxMined, receipt := isTxMined(ctx, logger, eth, execData.TxOpts.TxHashes)
		if isTxMined {
			// the transaction was successful and mined
			execData.TxOpts.MinedInBlock = receipt.BlockNumber.Uint64()
			logger.Infof("the tx %s was mined on height %d", execData.TxOpts.GetHexTxsHashes(), execData.TxOpts.MinedInBlock)

			isTxConfirmed, err = waitForFinalityDelay(ctx, logger, eth, execData.TxOpts.TxHashes, eth.GetFinalityDelay(), execData.TxOpts.MinedInBlock, txCheckFrequency)
			if err != nil {
				return dkg.LogReturnErrorf(logger, "failed to retry task with error %v", err)
			}
			if isTxConfirmed {
				logger.Infof("the tx %s is confirmed %t on height %d", execData.TxOpts.GetHexTxsHashes(), isTxMined, currentBlock)
				break
			}

			// if Tx wasn't confirmed after being mined we execute task again
			err = executeTask(ctx, logger, eth, task, retryCount, retryDelay)
			if err != nil {
				return dkg.LogReturnErrorf(logger, "failed to retry task with error %v", err)
			}
			// set new Tx replacement time
			txReplacement = getTxReplacementTime(txTimeoutForReplacement)
		} else {
			//if tx wasn't mined, check if we should replace
			logger.Infof("the tx %s was not mined", execData.TxOpts.GetHexTxsHashes())

			if time.Now().After(txReplacement) {
				logger.Infof("tx timed out: replacing tx %s with higher fee", execData.TxOpts.GetHexTxsHashes())

				err = retryTaskWithFeeReplacement(ctx, logger, eth, task, execData, retryCount, retryDelay)
				if err != nil {
					return dkg.LogReturnErrorf(logger, "failed to replace tx with hash %s and error %v", execData.TxOpts.GetHexTxsHashes(), err)
				}
				// set new Tx replacement time
				txReplacement = getTxReplacementTime(txTimeoutForReplacement)
			} else if receipt != nil && receipt.Status != uint64(1) {
				logger.Infof("tx timed out: recreating tx %s", execData.TxOpts.GetHexTxsHashes())

				// if the receipt indicates tx failed, then retry creating new tx
				clearTxOpts(task)
				err = retryTask(ctx, logger, eth, task, retryCount, retryDelay)
				if err != nil {
					return dkg.LogReturnErrorf(logger, "failed to replace tx with hash %s and error %v", execData.TxOpts.GetHexTxsHashes(), err)
				}
				// set new Tx replacement time
				txReplacement = getTxReplacementTime(txTimeoutForReplacement)
			}
		}

		//update the currentBlock
		currentBlock, err = eth.GetCurrentHeight(ctx)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "failed to get current height %v", err)
		}
	}

	logger.Infof("the tx %s was confirmed %t", execData.TxOpts.GetHexTxsHashes(), isTxConfirmed)

	return nil
}

func waitForFinalityDelay(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, txHashes []common.Hash, finalityDelay, txMinedInBlock uint64, txCheckFrequency time.Duration) (bool, error) {
	var err error
	currentBlock := txMinedInBlock
	confirmationBlock := txMinedInBlock + finalityDelay

	logger.WithFields(logrus.Fields{
		"currentBlock":      currentBlock,
		"confirmationBlock": confirmationBlock,
		"txCheckFrequency":  txCheckFrequency,
	}).Info("waitForFinalityDelay()...")

	// waiting for confirmation block
	for currentBlock < confirmationBlock {
		err = sleepWithContext(ctx, txCheckFrequency)
		if err != nil {
			return false, err
		}

		//update the currentBlock
		currentBlock, err = eth.GetCurrentHeight(ctx)
		if err != nil {
			return false, dkg.LogReturnErrorf(logger, "failed to get current height %v", err)
		}
	}

	isTxMined, _ := isTxMined(ctx, logger, eth, txHashes)

	return isTxMined, nil
}

func isTxMined(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, txHashes []common.Hash) (bool, *types.Receipt) {
	var receipt *types.Receipt
	var err error
	var isTxPending bool
	for _, txHash := range txHashes {
		//check tx is pending
		_, isTxPending, err = eth.GetGethClient().TransactionByHash(ctx, txHash)
		if err != nil {
			logger.Errorf("failed to get tx with hash %s wit error %v", txHash.Hex(), err)
			return false, nil
		}

		if isTxPending {
			return false, nil
		}

		//if tx was mined we can check the receipt
		receipt, err = eth.GetGethClient().TransactionReceipt(ctx, txHash)
		if err != nil {
			logger.Errorf("failed to get receipt for tx %s with error %s ", txHash.Hex(), err.Error())
			return false, nil
		}

		if receipt == nil {
			logger.Errorf("missing receipt for tx %s", txHash.Hex())
			return false, nil
		}

		// Check receipt to confirm we were successful
		if receipt.Status != uint64(1) {
			logger.Errorf("receipt status indicates failure: %v for tx %s", receipt.Status, txHash.Hex())
			return false, receipt
		}
	}

	return true, receipt
}

func getTxReplacementTime(timeoutForReplacement time.Duration) time.Time {
	return time.Now().Add(timeoutForReplacement)
}

func clearTxOpts(task interfaces.Task) {
	execData, ok := task.GetExecutionData().(*dkgtasks.ExecutionData)
	if ok && execData != nil {
		execData.Clear()
	}
}
