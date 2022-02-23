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
	ErrUnknownTaskType = errors.New("unkonwn task type")
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
		go waitFinalityDelay(ctx, dkgLogger, eth, task)
	}()

	return nil
}

func waitFinalityDelay(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task) {
	currentBlock, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		logger.Error("Failed to get current height ", err)
		return
	}

	//TODO: delete this and use config from eth
	TxCheckFrequency := 5 * time.Second
	nextTxCheck := time.After(TxCheckFrequency)
	txReplacement := time.After(30 * time.Second)
	txHash := common.Hash{}
	dkgTaskEnd := uint64(40)

	var minedOnBlock uint64
	isPending := true
	for currentBlock < dkgTaskEnd && isPending {
		select {
		case <-nextTxCheck:
			_, isPending, err = eth.GetGethClient().TransactionByHash(ctx, txHash)
			if err != nil {
				logger.Errorf("Failed to get tx with hash %s wit error %v", txHash, err)
				return
			}

			currentBlock, err = eth.GetCurrentHeight(ctx)
			if err != nil {
				logger.Error("Failed to get current height ", err)
				return
			}

			if !isPending {
				logger.Infof("the tx is mined on height %d ", currentBlock)
				minedOnBlock = currentBlock
				break
			}

			logger.Infof("the tx is pending on height %d ", currentBlock)
			//TODO: delete this and use config from eth
			nextTxCheck = time.After(TxCheckFrequency)
		case <-txReplacement:
			err = replaceTx(ctx, logger, eth, task)
		}
	}

	logger.Infof("the tx was mined on block %d ", minedOnBlock)

	//TODO: complete this
	_, err = eth.GetGethClient().TransactionReceipt(ctx, txHash)
	if err != nil {
		logger.Errorf("Failed to get receipt for tx %s with error %s ", txHash, err.Error())
		return
	}
}

func replaceTx(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, task interfaces.Task) error {
	var dkgTask *dkgtasks.DkgTask
	var dkgTaskImpl dkgtasks.DkgTaskIfase
	dkgTaskIfase := reflect.TypeOf((*dkgtasks.DkgTaskIfase)(nil)).Elem()
	isDkgTask := reflect.TypeOf(task).Implements(dkgTaskIfase)
	if isDkgTask {
		dkgTaskImpl = task.(dkgtasks.DkgTaskIfase)
		dkgTask = dkgTaskImpl.GetDkgTask()
		panic("wrong task implementation")
	}

	gasFeeCap, gasTipCap := dkg.IncreaseFeeAndTipCap(
		dkgTask.CallOptions.TxOpts.GasFeeCap,
		dkgTask.CallOptions.TxOpts.GasTipCap,
		dkgTask.CallOptions.TxFeePercentageToIncrease)
	dkgTask.CallOptions.TxOpts.GasFeeCap = gasFeeCap
	dkgTask.CallOptions.TxOpts.GasTipCap = gasTipCap

	count := 0
	var err error
	retryCount := eth.RetryCount()
	retryDelay := eth.RetryDelay()

	retryLogger := logger.WithField("Method", "replaceTx_DoRetry")
	for count < retryCount && task.ShouldRetry(ctx, logger.WithField("Method", "ShouldRetry"), eth) {

		err = task.DoRetry(ctx, retryLogger.WithField("RetryCount", count), eth)
		if err == nil {
			break
		}

		retryLogger.Error("Failed to execute retry ", err)
		time.Sleep(retryDelay)
		count++
	}

	return err
}
