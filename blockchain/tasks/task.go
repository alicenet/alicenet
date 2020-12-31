package tasks

import (
	"context"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/sirupsen/logrus"
)

// ========================================================
// Task interfaces and structs
// ========================================================

// Task the interface requirements of a task
type Task interface {
	DoDone()
	DoRetry(context.Context) bool
	DoWork(context.Context) bool
	ShouldRetry(context.Context) bool
}

// TaskHandler required functionality of a task
type TaskHandler interface {
	Cancel()
	Start()
	Complete() bool
	Successful() bool
}

// TaskDoFunc is shorthand for func(context.Context) bool, which is what the Do*() task functions are
// -- Return value indicates if task work has completed succesfully
type TaskDoFunc func(context.Context) bool

// TaskShouldFunc is shorthand for func(context.Context) bool, which is the ShouldRetry() type
type TaskShouldFunc func(context.Context) bool

// TaskDoneFunc is shorthand for func() bool, which is the DoDone() type
// -- This is executed as a cleanup, so can't be canceled
type TaskDoneFunc func()

// TaskHandlerDetails contains all the data required to implment a task
type TaskHandlerDetails struct {
	doDone      TaskDoneFunc
	doRetry     TaskDoFunc
	doWork      TaskDoFunc
	shouldRetry TaskShouldFunc
	complete    bool
	successful  bool
	count       int
	logger      *logrus.Entry
	ID          int
	wg          *sync.WaitGroup
	timeout     time.Duration
	retryDelay  time.Duration
	taskCancel  context.CancelFunc
}

// ========================================================
// Here are the standard functions that orchestrate tasks
// ========================================================

// Start begins orchestrating the task in a goroutine
func (td *TaskHandlerDetails) Start() {

	td.logger.Infof("Starting...")
	td.wg.Add(1)

	var taskContext context.Context
	taskContext, td.taskCancel = context.WithCancel(context.Background())

	go func(ctx context.Context) {
		if td.wg != nil {
			defer td.wg.Done()
		}
		defer td.wrappedDoDone()
		defer td.taskCancel() // TODO double check

		td.successful = td.wrappedDoWork(ctx)

		for !td.successful && td.wrappedShouldRetry(ctx) {

			err := blockchain.SleepWithContext(ctx, td.retryDelay)
			if err != nil {
				td.logger.Warnf("Task interupted: %v", ctx.Err())
				return
			}

			td.successful = td.wrappedDoRetry(ctx)
		}

	}(taskContext)
}

// Cancel stops the task's goroutine
func (td *TaskHandlerDetails) Cancel() {
	td.logger.Infof("Canceling ...")
	if td.taskCancel != nil {
		td.taskCancel()
	}
}

// Successful returns indication of whether the task was eventually successful
func (td *TaskHandlerDetails) Successful() bool {
	return td.successful
}

// Complete returns indication of whether has completed
func (td *TaskHandlerDetails) Complete() bool {
	return td.complete
}

// ========================================================
// Here are the methods that invoke task specific functionality
// ========================================================

func (td *TaskHandlerDetails) wrappedDoWork(ctx context.Context) bool {

	td.count++

	var resp bool

	if td.doWork != nil {
		ctx, cancelFunc := context.WithTimeout(ctx, td.timeout)
		defer cancelFunc()
		resp = td.doWork(ctx)
	} else {
		resp = true
	}

	td.logger.Infof("wrappedDoWork(...) try %d: %v", td.count, resp)

	return resp
}

func (td *TaskHandlerDetails) wrappedDoRetry(ctx context.Context) bool {

	td.count++

	var resp bool

	if td.doRetry != nil {
		ctx, cancelFunc := context.WithTimeout(ctx, td.timeout)
		defer cancelFunc()
		resp = td.doRetry(ctx)
	} else {
		resp = true
	}

	td.logger.Infof("wrappedDoRetry(...) try %d: %v", td.count, resp)

	return resp
}

func (td *TaskHandlerDetails) wrappedShouldRetry(ctx context.Context) bool {

	var resp bool

	if td.shouldRetry != nil {
		ctx, cancelFunc := context.WithTimeout(ctx, td.timeout)
		defer cancelFunc()
		resp = td.shouldRetry(ctx)
	} else {
		resp = false
	}

	td.logger.Infof("wrappedShouldRetry(...): %v", resp)

	return resp
}

func (td *TaskHandlerDetails) wrappedDoDone() {
	td.logger.Infof("wrappedDoDone(...) tries %d", td.count)
	if td.doDone != nil {
		td.doDone()
	}
	td.complete = true
}
