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

// ========================================================
// Task interfaces and structs
// ========================================================

// TaskWrapper is used when marshalling and unmarshalling tasks

// TaskHandlerDetails contains all the data required to execute a task
// type TaskHandlerDetails struct {
// 	complete   bool
// 	count      int
// 	eth        interfaces.Ethereum
// 	ID         int
// 	label      string
// 	logger     *logrus.Logger
// 	successful bool
// 	task       interfaces.Task
// 	taskCancel context.CancelFunc
// 	wg         *sync.WaitGroup
// }

// ========================================================
// Here are the standard functions that orchestrate tasks
// ========================================================

// Start begins orchestrating the task in a goroutine
// func (td *TaskHandlerDetails) Start() {

// 	td.Logger().Infof("Starting...")
// 	td.wg.Add(1)

// 	var taskContext context.Context
// 	taskContext, td.taskCancel = context.WithCancel(context.Background())

// 	retryDelay := td.eth.RetryDelay()

// 	go func(ctx context.Context) {
// 		if td.wg != nil {
// 			defer td.wg.Done()
// 		}
// 		defer td.wrappedDoDone()
// 		defer td.taskCancel()

// 		td.task.Initialize(ctx, td.logger, td.eth)

// 		td.successful = td.wrappedDoWork(ctx)

// 		for !td.successful && td.wrappedShouldRetry(ctx) {

// 			err := blockchain.SleepWithContext(ctx, retryDelay)
// 			if err != nil {
// 				td.Logger().Warnf("Task interupted: %v", ctx.Err())
// 				return
// 			}

// 			td.successful = td.wrappedDoRetry(ctx)
// 		}

// 	}(taskContext)
// }

// // Cancel stops the task's goroutine
// func (td *TaskHandlerDetails) Cancel() {
// 	td.Logger().Infof("Canceling ...")
// 	if td.taskCancel != nil {
// 		td.taskCancel()
// 	}
// }

// // Successful returns indication of whether the task was eventually successful
// func (td *TaskHandlerDetails) Successful() bool {
// 	return td.successful
// }

// // Complete returns indication of whether has completed
// func (td *TaskHandlerDetails) Complete() bool {
// 	return td.complete
// }

// func (td *TaskHandlerDetails) Logger() *logrus.Entry {
// 	return td.logger.WithField("TaskID", td.label)
// }

// ========================================================
// Here are the methods that invoke task specific functionality
// ========================================================

// func (td *TaskHandlerDetails) wrappedDoWork(ctx context.Context) bool {

// 	td.count++

// 	timeout := td.eth.Timeout()

// 	ctx, cancelFunc := context.WithTimeout(ctx, timeout)
// 	defer cancelFunc()

// 	resp = td.task.DoWork(ctx, td.logger, td.eth)

// 	td.Logger().Infof("wrappedDoWork(...) try %d: %v", td.count, resp)

// 	return resp
// }

// func (td *TaskHandlerDetails) wrappedDoRetry(ctx context.Context) bool {

// 	td.count++

// 	var resp bool

// 	timeout := td.eth.Timeout()

// 	ctx, cancelFunc := context.WithTimeout(ctx, timeout)
// 	defer cancelFunc()
// 	resp = td.task.DoRetry(ctx, td.logger, td.eth)

// 	td.Logger().Infof("wrappedDoRetry(...) try %d: %v", td.count, resp)

// 	return resp
// }

// func (td *TaskHandlerDetails) wrappedShouldRetry(ctx context.Context) bool {

// 	var resp bool

// 	timeout := td.eth.Timeout()

// 	ctx, cancelFunc := context.WithTimeout(ctx, timeout)
// 	defer cancelFunc()
// 	resp = td.task.ShouldRetry(ctx, td.logger, td.eth)

// 	td.Logger().Infof("wrappedShouldRetry(...): %v", resp)

// 	return resp
// }

// func (td *TaskHandlerDetails) wrappedDoDone() {
// 	td.Logger().Infof("wrappedDoDone(...) tries %d", td.count)
// 	td.task.DoDone(td.logger)
// 	td.complete = true
// }

// ========================================================
// Manager interfaces and structs
// ========================================================

// Manager describtes the basic functionality of a task Manager
// type Manager interface {
// 	// NewTaskHandler(logger *logrus.Logger, eth interfaces.Ethereum, t interfaces.Task) interfaces.TaskHandler
// 	StartTask(logger *logrus.Entry, eth interfaces.Ethereum, t interfaces.Task) interfaces.TaskHandler
// 	WaitForTasks()
// }

// // ManagerDetails contains information required for implmentation of task Manager
// type ManagerDetails struct {
// 	wg sync.WaitGroup
// }

// ========================================================
// Manager implementation
// ========================================================

// NewTaskHandler creates a new task handler, where each phase can take upto 'timeout'
// duration and there is a delay of 'retryDelay' before a retry.
// func (md *ManagerDetails) NewTaskHandler(logger *logrus.Logger, eth interfaces.Ethereum, task interfaces.Task) interfaces.TaskHandler {

// 	taskID := rand.Intn(0x10000)
// 	taskLabel := fmt.Sprintf("0x%04x", taskID)

// 	logger.Infof("Creating task %v with timeout of %v and retryDelay of %v", taskID, eth.Timeout(), eth.RetryDelay())

// 	return &TaskHandlerDetails{
// 		ID:     taskID,
// 		eth:    eth,
// 		task:   task,
// 		label:  taskLabel,
// 		logger: logger,
// 		wg:     &md.wg}
// }

func StartTask(logger *logrus.Entry, wg *sync.WaitGroup, eth interfaces.Ethereum, task interfaces.Task) interfaces.TaskHandler {

	go func() {
		defer task.DoDone(logger)
		defer wg.Done()

		retryCount := eth.RetryCount()
		retryDelay := eth.RetryDelay()
		timeout := eth.Timeout()
		logger.WithFields(logrus.Fields{
			"Timeout":    timeout,
			"RetryCount": retryCount,
			"RetryDelay": retryDelay,
		}).Debug("StartTask()...")

		// Setup
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var count int
		var err error

		err = task.Initialize(ctx, logger, eth)
		logger.Debugf("Initialize ... error %v", err)
		for err != nil && count < retryCount {
			if err == objects.ErrCanNotContinue {
				logger.Error("can not continue:", err)
				return
			}
			time.Sleep(retryDelay)
			err = task.Initialize(ctx, logger, eth)
			count++
		}
		if err != nil {
			logger.Errorf("Failed to initialize task: %v", err)
			return
		}

		count = 0
		err = task.DoWork(ctx, logger, eth)
		for err != nil && count < retryCount && task.ShouldRetry(ctx, logger, eth) {
			if err == objects.ErrCanNotContinue {
				logger.Error("can not continue", err)
				return
			}
			time.Sleep(retryDelay)
			err = task.DoRetry(ctx, logger, eth)
			count++
		}
		if err != nil {
			logger.Errorf("Failed to execute task: %v", err)
			return
		}
	}()

	return nil
}

// ========================================================
// Custom Marshal/Unmarshal for tasks
// ========================================================
