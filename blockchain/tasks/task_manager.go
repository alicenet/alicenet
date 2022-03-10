package tasks

import (
	"context"
	"errors"
	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/ethereum/go-ethereum/common"
	"reflect"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

var (
	ErrUnknownTaskName = errors.New("unknown task name")
	ErrUnknownTaskType = errors.New("unknown task type")
)

const NonceToLowError = "nonce too low"

func StartTask(logger *logrus.Entry, wg *sync.WaitGroup, eth interfaces.Ethereum, task interfaces.Task, state interface{}) error {

	wg.Add(1)
	go func() {
		defer task.DoDone(logger.WithField("Method", "DoDone"))
		defer wg.Done()

		retryCount := eth.RetryCount()
		retryDelay := eth.RetryDelay()
		logger.WithFields(logrus.Fields{
			"RetryCount": retryCount,
			"RetryDelay": retryDelay,
		}).Info("StartTask()...")

		var dkgTask *dkgtasks.DkgTask
		var dkgTaskImpl dkgtasks.DkgTaskIfase
		dkgTaskIfase := reflect.TypeOf((*dkgtasks.DkgTaskIfase)(nil)).Elem()
		isDkgTask := reflect.TypeOf(task).Implements(dkgTaskIfase)
		if !isDkgTask {
			panic(ErrUnknownTaskType.Error())
		}

		dkgTaskImpl = task.(dkgtasks.DkgTaskIfase)
		dkgTask = dkgTaskImpl.GetDkgTask()

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
		err = executeTask(ctx, workLogger, eth, task, dkgTask, retryCount, retryDelay)
		if err != nil {
			workLogger.Error("Failed to execute task ", err)
			return
		}

		dkgLogger := logger.WithField("Method", "handleExecutedTask")
		err = handleExecutedTask(ctx, dkgLogger, eth, task, dkgTask)
		if err != nil {
			dkgLogger.Error("Failed to execute handleExecutedTask ", err)
		}
	}()

	return nil
}

// initializeTask initialize the Task and retry needed
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

// executeTask execute the Task and retry needed
func executeTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task, dkgTask *dkgtasks.DkgTask, retryCount int, retryDelay time.Duration) error {
	err := task.DoWork(ctx, logger, eth)
	if err != nil {
		retryLogger := logger.WithField("Method", "DoRetry")
		err = retryTask(ctx, retryLogger, eth, task, dkgTask, retryCount, retryDelay)
	}

	return err
}

// retryTask retry the Task
func retryTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task, dkgTask *dkgtasks.DkgTask, retryCount int, retryDelay time.Duration) error {
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
			// as a success or fail. If is a fail we should restart with a new nonce
			dkgTask.Clear()
		}
		count++
	}

	return err
}

func retryTaskWithFeeReplacement(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task, dkgTask *dkgtasks.DkgTask, retryCount int, retryDelay time.Duration) error {
	logger.WithFields(logrus.Fields{
		"GasFeeCap": dkgTask.TxReplOpts.GasFeeCap,
		"GasTipCap": dkgTask.TxReplOpts.GasTipCap,
		"Nonce":     dkgTask.TxReplOpts.Nonce,
		"TxHash":    dkgTask.TxReplOpts.TxHash.Hex(),
	}).Info("retryTaskWithFeeReplacement")

	// increase gas and tip cap
	gasFeeCap, gasTipCap := dkg.IncreaseFeeAndTipCap(
		dkgTask.TxReplOpts.GasFeeCap,
		dkgTask.TxReplOpts.GasTipCap,
		eth.GetTxFeePercentageToIncrease())
	dkgTask.TxReplOpts.GasFeeCap = gasFeeCap
	dkgTask.TxReplOpts.GasTipCap = gasTipCap

	logger.WithFields(logrus.Fields{
		"GasFeeCap": dkgTask.TxReplOpts.GasFeeCap,
		"GasTipCap": dkgTask.TxReplOpts.GasTipCap,
		"Nonce":     dkgTask.TxReplOpts.Nonce,
		"TxHash":    dkgTask.TxReplOpts.TxHash.Hex(),
	}).Info("retryTaskWithFeeReplacement2")

	return retryTask(ctx, logger, eth, task, dkgTask, retryCount, retryDelay)
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
// to make sure the Tx has priority for the next mined blocks
func handleExecutedTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task, dkgTask *dkgtasks.DkgTask) error {
	// TxReplOpts or TxHash are empty means that no tx was queued, this could happen
	// if there's nobody to accuse during the dispute
	var emptyHash [32]byte
	if dkgTask.TxReplOpts == nil || dkgTask.TxReplOpts.TxHash == emptyHash {
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
	dkgTaskEnd := dkgTask.End

	logger.WithFields(logrus.Fields{
		"currentBlock":            currentBlock,
		"retryCount":              retryCount,
		"retryDelay":              retryDelay,
		"txCheckFrequency":        txCheckFrequency,
		"txTimeoutForReplacement": txTimeoutForReplacement,
		"txReplacement":           txReplacement,
		"dkgTaskEnd":              dkgTaskEnd,
		"TxHash":                  dkgTask.TxReplOpts.TxHash.Hex(),
	}).Info("handleExecutedTask()...")

	var txMinedInBlock uint64
	var txConfirmed bool
	for currentBlock < dkgTaskEnd && dkgTask.TxReplOpts.TxHash != emptyHash && !txConfirmed {
		err = sleepWithContext(ctx, txCheckFrequency)
		if err != nil {
			return err
		}

		//check tx is pending
		_, txPending, err := eth.GetGethClient().TransactionByHash(ctx, dkgTask.TxReplOpts.TxHash)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "failed to get tx with hash %s with error %v", dkgTask.TxReplOpts.TxHash.Hex(), err)
		}

		logger.Infof("the tx %s isPending %t", dkgTask.TxReplOpts.TxHash.Hex(), txPending)

		if txPending {
			//if tx is pending, check if we should replace
			logger.Infof("the tx %s is pending", dkgTask.TxReplOpts.TxHash.Hex())

			if time.Now().After(txReplacement) {
				logger.Infof("replacing tx %s with higher fee", dkgTask.TxReplOpts.TxHash.Hex())

				err = retryTaskWithFeeReplacement(ctx, logger, eth, task, dkgTask, retryCount, retryDelay)
				if err != nil {
					return dkg.LogReturnErrorf(logger, "failed to replace tx with hash %s and error %v", dkgTask.TxReplOpts.TxHash.Hex(), err)
				}
				// set new Tx replacement time
				txReplacement = getTxReplacementTime(txTimeoutForReplacement)
			}
		} else {
			//if tx was mined we can check the receipt
			logger.Infof("the tx %s is not Pending", dkgTask.TxReplOpts.TxHash.Hex())
			receipt, err := eth.GetGethClient().TransactionReceipt(ctx, dkgTask.TxReplOpts.TxHash)
			if err != nil {
				return dkg.LogReturnErrorf(logger, "failed to get receipt for tx %s with error %s ", dkgTask.TxReplOpts.TxHash.Hex(), err.Error())
			}
			if receipt == nil {
				return dkg.LogReturnErrorf(logger, "missing receipt for tx %s", dkgTask.TxReplOpts.TxHash.Hex())
			}

			logger.Infof("the tx %s receipt not nil", dkgTask.TxReplOpts.TxHash.Hex())

			// Check receipt to confirm we were successful
			if receipt.Status != uint64(1) {
				logger.Errorf("receipt status indicates failure: %v for tx %s", receipt.Status, dkgTask.TxReplOpts.TxHash.Hex())

				// Clearing TxReplOpts used for tx gas ana nonce replacement
				dkgTask.Clear()
				err = executeTask(ctx, logger, eth, task, dkgTask, retryCount, retryDelay)
				if err != nil {
					return dkg.LogReturnErrorf(logger, "failed to retry task with error %v", err)
				}
				// set new Tx replacement time
				txReplacement = getTxReplacementTime(txTimeoutForReplacement)
			} else {
				// the transaction was successful and mined
				txMinedInBlock = receipt.BlockNumber.Uint64()
				logger.Infof("the tx %s was mined on height %d", dkgTask.TxReplOpts.TxHash.Hex(), txMinedInBlock)

				txConfirmed, err = waitForFinalityDelay(ctx, logger, eth, dkgTask.TxReplOpts.TxHash, eth.GetFinalityDelay(), txMinedInBlock, txCheckFrequency)
				if err != nil {
					return dkg.LogReturnErrorf(logger, "failed to retry task with error %v", err)
				}
				if txConfirmed {
					break
				}
			}
		}

		//update the currentBlock
		currentBlock, err = eth.GetCurrentHeight(ctx)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "failed to get current height %v", err)
		}
	}

	logger.Infof("the tx %s was confirmed %t", dkgTask.TxReplOpts.TxHash.Hex(), txConfirmed)

	return nil
}

func waitForFinalityDelay(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, txHash common.Hash, finalityDelay, txMinedInBlock uint64, txCheckFrequency time.Duration) (bool, error) {
	var err error
	currentBlock := txMinedInBlock
	confirmationBlock := txMinedInBlock + finalityDelay

	logger.WithFields(logrus.Fields{
		"currentBlock":      currentBlock,
		"confirmationBlock": confirmationBlock,
		"txCheckFrequency":  txCheckFrequency,
		"TxHash":            txHash,
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

	stillMinedTx := isMinedTx(ctx, logger, eth, txHash)

	logger.Infof("the tx %s is confirmed %t on height %d", txHash.Hex(), stillMinedTx, currentBlock)

	return stillMinedTx, nil
}

func isMinedTx(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, txHash common.Hash) bool {
	//check tx is pending
	_, txPending, err := eth.GetGethClient().TransactionByHash(ctx, txHash)
	if err != nil {
		logger.Errorf("failed to get tx with hash %s wit error %v", txHash.Hex(), err)
	}

	if txPending {
		return false
	}

	receipt, err := eth.GetGethClient().TransactionReceipt(ctx, txHash)
	if err != nil {
		logger.Errorf("failed to get receipt for tx %s with error %s ", txHash.Hex(), err.Error())
		return false
	}

	if receipt == nil {
		logger.Errorf("missing receipt for tx %s", txHash.Hex())
		return false
	}

	logger.Infof("the tx %s receipt not nil", txHash.Hex())

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		logger.Errorf("receipt status indicates failure: %v for tx %s", receipt.Status, txHash.Hex())
		return false
	}

	return true
}

func getTxReplacementTime(timeoutForReplacement time.Duration) time.Time {
	return time.Now().Add(timeoutForReplacement)
}
