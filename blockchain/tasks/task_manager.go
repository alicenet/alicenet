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
		timeout := eth.Timeout()
		logger.WithFields(logrus.Fields{
			"Timeout":    timeout,
			"RetryCount": retryCount,
			"RetryDelay": retryDelay,
		}).Info("StartTask()...")

		// Setup
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var count int
		var err error

		initializationLogger := logger.WithField("Method", "Initialize")
		err = task.Initialize(ctx, initializationLogger, eth, state)
		for err != nil && count < retryCount {
			if errors.Is(err, objects.ErrCanNotContinue) {
				initializationLogger.Error(err)
				return
			}
			time.Sleep(retryDelay)
			err = task.Initialize(ctx, initializationLogger, eth, state)
			count++
		}
		if err != nil {
			initializationLogger.Errorf("Failed to initialize task: %v", err)
			return
		}

		count = 0

		workLogger := logger.WithField("Method", "DoWork")
		err = task.DoWork(ctx, workLogger, eth)

		retryLogger := logger.WithField("Method", "DoRetry")
		for err != nil && count < retryCount && task.ShouldRetry(ctx, logger.WithField("Method", "ShouldRetry"), eth) {
			if errors.Is(err, objects.ErrCanNotContinue) {
				initializationLogger.Error(err)
				return
			}
			time.Sleep(retryDelay)
			count++
			err = task.DoRetry(ctx, retryLogger.WithField("RetryCount", count), eth)
		}

		if err != nil {
			logger.Error("Failed to execute task ", err)
			return
		}

		dkgLogger := logger.WithField("Method", "waitFinalityDelay")
		go waitFinalityDelay(context.Background(), dkgLogger, eth, task, count)
	}()

	return nil
}

// waitFinalityDelay responsibilities:
// if the transaction was mined wait for FinalityDelay to ensure there was no rollback on Ethereum
// if the transaction wasn't mined during the txTimeoutForReplacement we increase the Fee
// to make sure the tx has priority for the next mined blocks
func waitFinalityDelay(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task, count int) {
	var dkgTask *dkgtasks.DkgTask
	var dkgTaskImpl dkgtasks.DkgTaskIfase
	dkgTaskIfase := reflect.TypeOf((*dkgtasks.DkgTaskIfase)(nil)).Elem()
	isDkgTask := reflect.TypeOf(task).Implements(dkgTaskIfase)
	if !isDkgTask {
		panic(ErrUnknownTaskType.Error())
	}

	dkgTaskImpl = task.(dkgtasks.DkgTaskIfase)
	dkgTask = dkgTaskImpl.GetDkgTask()

	logger.Infof("the dkgTask %#v", dkgTask)
	var emptyHash [32]byte
	if dkgTask.TxReplacementOpts == nil || dkgTask.TxReplacementOpts.TxHash == emptyHash {
		return
	}

	currentBlock, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		logger.Error("failed to get current height ", err)
		return
	}

	txCheckFrequency := eth.GetTxCheckFrequency()
	nextTxCheck := time.After(txCheckFrequency)
	txTimeoutForReplacement := eth.GetTxTimeoutForReplacement()
	txReplacement := time.After(txTimeoutForReplacement)
	dkgTaskEnd := dkgTask.End

	logger.WithFields(logrus.Fields{
		"txCheckFrequency":        txCheckFrequency,
		"txTimeoutForReplacement": txTimeoutForReplacement,
		"dkgTaskEnd":              dkgTaskEnd,
		"TxHash":                  dkgTask.TxReplacementOpts.TxHash.Hex(),
	}).Info("waitFinalityDelay()...")

	var txMinedOnBlock uint64
	isPending := true
	for currentBlock < dkgTaskEnd && txMinedOnBlock == 0 {
		select {
		case <-nextTxCheck:
			//check if the tx is pending or was mined
			_, isPending, err = eth.GetGethClient().TransactionByHash(ctx, dkgTask.TxReplacementOpts.TxHash)
			if err != nil {
				logger.Errorf("failed to get tx with hash %s wit error %v", dkgTask.TxReplacementOpts.TxHash.Hex(), err)
				return
			}

			logger.Infof("the tx %s isPending %v", dkgTask.TxReplacementOpts.TxHash.Hex(), isPending)

			//if tx was mined we can check the receipt
			if !isPending {

				logger.Infof("the tx %s is not Pending", dkgTask.TxReplacementOpts.TxHash.Hex())
				receipt, err := eth.GetGethClient().TransactionReceipt(ctx, dkgTask.TxReplacementOpts.TxHash)
				if err != nil {
					logger.Errorf("failed to get receipt for tx %s with error %s ", dkgTask.TxReplacementOpts.TxHash.Hex(), err.Error())
					return
				}
				if receipt == nil {
					logger.Errorf("missing receipt for tx %s", dkgTask.TxReplacementOpts.TxHash.Hex())
					return
				}

				logger.Infof("the tx %s receipt not nil", dkgTask.TxReplacementOpts.TxHash.Hex())

				// Check receipt to confirm we were successful
				if receipt.Status != uint64(1) {
					logger.Errorf("receipt status indicates failure: %v for tx %s", receipt.Status, dkgTask.TxReplacementOpts.TxHash.Hex())

					// Clearing TxReplacementOpts used for tx replacement
					dkgTask.TxReplacementOpts.Clean()
					// Retry until reach the retryCount
					err = retryTask(ctx, logger, eth, task, count)
					if err != nil {
						logger.Errorf("failed to retry task with error %v", err)
						return
					}
				} else {
					// the transaction was successful and mined
					txMinedOnBlock = receipt.BlockNumber.Uint64()
					logger.Infof("the tx %s was mined on height %d", dkgTask.TxReplacementOpts.TxHash.Hex(), txMinedOnBlock)
					break
				}
			}

			logger.Info("the tx is pending")

			//if tx still pending we update the currentBlock and nextTxCheck
			currentBlock, err = eth.GetCurrentHeight(ctx)
			if err != nil {
				logger.Error("failed to get current height ", err)
				return
			}
			nextTxCheck = time.After(txCheckFrequency)
		case <-txReplacement:
			//replace the tx with the same nonce but higher fee
			logger.Infof("tx %s retry with fee replacement started", dkgTask.TxReplacementOpts.TxHash.Hex())
			count = 0
			err = retryTaskWithFeeReplacement(ctx, logger, eth, task, dkgTask, count)
			if err != nil {
				logger.Errorf("failed to replace tx with hash %s and error %v", dkgTask.TxReplacementOpts.TxHash.Hex(), err)
				return
			}

			//update the currentBlock and txReplacement
			currentBlock, err = eth.GetCurrentHeight(ctx)
			if err != nil {
				logger.Error("failed to get current height ", err)
				return
			}
			txReplacement = time.After(txTimeoutForReplacement)
		}
	}

	//TODO: wait for finality delay and check if the tx still successful
}

func retryTaskWithFeeReplacement(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task, dkgTask *dkgtasks.DkgTask, count int) error {
	logger.WithFields(logrus.Fields{
		"GasFeeCap": dkgTask.TxReplacementOpts.GasFeeCap,
		"GasTipCap": dkgTask.TxReplacementOpts.GasTipCap,
		"Nonce":     dkgTask.TxReplacementOpts.Nonce,
		"TxHash":    dkgTask.TxReplacementOpts.TxHash.Hex(),
	}).Info("retryTaskWithFeeReplacement")

	// increase gas and tip cap
	gasFeeCap, gasTipCap := dkg.IncreaseFeeAndTipCap(
		dkgTask.TxReplacementOpts.GasFeeCap,
		dkgTask.TxReplacementOpts.GasTipCap,
		eth.GetTxFeePercentageToIncrease())
	dkgTask.TxReplacementOpts.GasFeeCap = gasFeeCap
	dkgTask.TxReplacementOpts.GasTipCap = gasTipCap

	logger.WithFields(logrus.Fields{
		"GasFeeCap": dkgTask.TxReplacementOpts.GasFeeCap,
		"GasTipCap": dkgTask.TxReplacementOpts.GasTipCap,
		"Nonce":     dkgTask.TxReplacementOpts.Nonce,
		"TxHash":    dkgTask.TxReplacementOpts.TxHash.Hex(),
	}).Info("retryTaskWithFeeReplacement2")

	return retryTask(ctx, logger, eth, task, count)
}

func retryTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task, count int) error {
	var err error
	retryCount := eth.RetryCount()
	retryDelay := eth.RetryDelay()

	retryLogger := logger.WithField("Method", "replaceTx_DoRetry")
	for count < retryCount && task.ShouldRetry(ctx, retryLogger.WithField("Method", "replaceTx__ShouldRetry"), eth) {
		err = task.DoRetry(ctx, retryLogger.WithField("replaceTx_RetryCount", count), eth)
		if err == nil {
			break
		}

		retryLogger.Error("failed to execute retry ", err)
		time.Sleep(retryDelay)
		count++
	}

	return err
}
