package tasks

import (
	"context"
	"errors"
	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
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

		dkgLogger := logger.WithField("Method", "waitFinalityDelay")
		err = waitFinalityDelay(ctx, dkgLogger, eth, task)
		if err != nil {
			dkgLogger.Error("Failed to execute waitFinalityDelay ", err)
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
func executeTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task, retryCount int, retryDelay time.Duration) error {
	err := task.DoWork(ctx, logger, eth)
	if err != nil {
		retryLogger := logger.WithField("Method", "DoRetry")
		err = retryTask(ctx, retryLogger, eth, task, retryCount, retryDelay)
	}

	return err
}

// retryTask retry the Task
func retryTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task, retryCount int, retryDelay time.Duration) error {
	var count int
	var err error
	for err != nil && count < retryCount && task.ShouldRetry(ctx, logger, eth) {
		if errors.Is(err, objects.ErrCanNotContinue) {
			return err
		}

		err = sleepWithContext(ctx, retryDelay)
		if err != nil {
			return err
		}

		err = task.DoRetry(ctx, logger, eth)
		count++
	}

	return err
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

// waitFinalityDelay responsibilities:
// if the transaction was mined wait for FinalityDelay to ensure there was no rollback on Ethereum
// if the transaction wasn't mined during the txTimeoutForReplacement we increase the Fee
// to make sure the tx has priority for the next mined blocks
func waitFinalityDelay(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task) error {
	var dkgTask *dkgtasks.DkgTask
	var dkgTaskImpl dkgtasks.DkgTaskIfase
	dkgTaskIfase := reflect.TypeOf((*dkgtasks.DkgTaskIfase)(nil)).Elem()
	isDkgTask := reflect.TypeOf(task).Implements(dkgTaskIfase)
	if !isDkgTask {
		panic(ErrUnknownTaskType.Error())
	}

	dkgTaskImpl = task.(dkgtasks.DkgTaskIfase)
	dkgTask = dkgTaskImpl.GetDkgTask()

	var emptyHash [32]byte
	if dkgTask.TxReplOpts == nil || dkgTask.TxReplOpts.TxHash == emptyHash {
		return dkg.LogReturnErrorf(logger, "failed to get tx hash and replacement options")
	}

	currentBlock, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "failed to get current height ", err)
	}

	retryCount := eth.RetryCount()
	retryDelay := eth.RetryDelay()
	txCheckFrequency := eth.GetTxCheckFrequency()
	nextTxCheck := time.After(txCheckFrequency)
	txTimeoutForReplacement := eth.GetTxTimeoutForReplacement()
	txReplacement := time.After(txTimeoutForReplacement)
	dkgTaskEnd := dkgTask.End

	logger.WithFields(logrus.Fields{
		"retryCount":              retryCount,
		"retryDelay":              retryDelay,
		"txCheckFrequency":        txCheckFrequency,
		"txTimeoutForReplacement": txTimeoutForReplacement,
		"dkgTaskEnd":              dkgTaskEnd,
		"TxHash":                  dkgTask.TxReplOpts.TxHash.Hex(),
	}).Info("waitFinalityDelay()...")

	var txMinedOnBlock uint64
	var txPending = true
	for currentBlock < dkgTaskEnd || txPending {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-nextTxCheck:
			//check tx is pending
			_, txPending, err = eth.GetGethClient().TransactionByHash(ctx, dkgTask.TxReplOpts.TxHash)
			if err != nil {
				return dkg.LogReturnErrorf(logger, "failed to get tx with hash %s wit error %v", dkgTask.TxReplOpts.TxHash.Hex(), err)
			}

			logger.Infof("the tx %s isPending %v", dkgTask.TxReplOpts.TxHash.Hex(), txPending)

			//if tx was mined we can check the receipt
			if !txPending {

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
					err = executeTask(ctx, logger, eth, task, retryCount, retryDelay)
					if err != nil {
						return dkg.LogReturnErrorf(logger, "failed to retry task with error %v", err)
					}
				} else {
					// the transaction was successful and mined
					txMinedOnBlock = receipt.BlockNumber.Uint64()
					logger.Infof("the tx %s was mined on height %d", dkgTask.TxReplOpts.TxHash.Hex(), txMinedOnBlock)
					break
				}
			}

			//TODO: what to do if is not pending anymore? wait here?? See what happen with txReplacement

			logger.Infof("the tx %s is pending", dkgTask.TxReplOpts.TxHash.Hex())

			//if tx still pending we update nextTxCheck
			nextTxCheck = time.After(txCheckFrequency)

		case <-txReplacement:
			// if tx is pending we replace it with higher fee
			_, txPending, err := eth.GetGethClient().TransactionByHash(ctx, dkgTask.TxReplOpts.TxHash)
			if err != nil {
				return dkg.LogReturnErrorf(logger, "failed to get tx with hash %s wit error %v", dkgTask.TxReplOpts.TxHash.Hex(), err)
			}

			if txPending {
				//replace the tx with the same nonce but higher fee
				logger.Infof("tx %s retry with fee replacement started", dkgTask.TxReplOpts.TxHash.Hex())

				//TODO: REFACTOR THIS
				count = 0
				err = retryTaskWithFeeReplacement(ctx, logger, eth, task, dkgTask, count)
				if err != nil {
					return dkg.LogReturnErrorf(logger, "failed to replace tx with hash %s and error %v", dkgTask.TxReplOpts.TxHash.Hex(), err)
				}
			}

			//update the  txReplacement
			txReplacement = time.After(txTimeoutForReplacement)
		}

		//update the currentBlock
		currentBlock, err = eth.GetCurrentHeight(ctx)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "failed to get current height ", err)
		}
	}

	//TODO: wait for finality delay and check if the tx still successful
}

func retryTaskWithFeeReplacement(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task, dkgTask *dkgtasks.DkgTask, count int) error {
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

	//TODO: refactor this
	return retryTask(ctx, logger, eth, task, count)
}

//func retryTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task, count int) error {
//	var err error
//	retryCount := eth.RetryCount()
//	retryDelay := eth.RetryDelay()
//
//	retryLogger := logger.WithField("Method", "replaceTx_DoRetry")
//	for count < retryCount && task.ShouldRetry(ctx, retryLogger.WithField("Method", "replaceTx__ShouldRetry"), eth) {
//		err = task.DoRetry(ctx, retryLogger.WithField("replaceTx_RetryCount", count), eth)
//		if err == nil {
//			break
//		}
//
//		retryLogger.Error("failed to execute retry ", err)
//		time.Sleep(retryDelay)
//		count++
//	}
//
//	return err
//}
