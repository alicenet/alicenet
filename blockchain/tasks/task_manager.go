package tasks

import (
	"context"
	"errors"
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
	}()

	return nil
}
