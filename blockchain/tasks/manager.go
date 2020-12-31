package tasks

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ========================================================
// Manager interfaces and structs
// ========================================================

// Manager describtes the basic functionality of a task Manager
type Manager interface {
	NewTaskHandler(timeout time.Duration, retryDelay time.Duration, t Task) TaskHandler
	WaitForTasks()
}

// ManagerDetails contains information required for implmentation of task Manager
type ManagerDetails struct {
	wg     sync.WaitGroup
	logger *logrus.Logger
}

// ========================================================
// Manager implementation
// ========================================================

// NewTaskHandler creates a new task handler, where each phase can take upto 'timeout'
// duration and there is a delay of 'retryDelay' before a retry.
func (md *ManagerDetails) NewTaskHandler(
	timeout time.Duration, retryDelay time.Duration, task Task) TaskHandler {

	taskID := rand.Intn(0x10000)
	taskLabel := fmt.Sprintf("0x%04x", taskID)

	md.logger.Infof("Creating task %v with timeout of %v and retryDelay of %v", taskID, timeout, retryDelay)

	return &TaskHandlerDetails{
		ID:          taskID,
		doWork:      task.DoWork,
		doRetry:     task.DoRetry,
		shouldRetry: task.ShouldRetry,
		doDone:      task.DoDone,
		logger:      md.logger.WithField("TaskID", taskLabel),
		wg:          &md.wg,
		timeout:     timeout,
		retryDelay:  retryDelay,
	}
}

// WaitForTasks blocks until all tasks associated withis Manager have completed
func (md *ManagerDetails) WaitForTasks() {
	md.wg.Wait()
}

// ========================================================
// Support functionality
// ========================================================

// NewManager creates a new Manager
func NewManager(logger *logrus.Logger) Manager {
	return &ManagerDetails{
		logger: logger,
		wg:     sync.WaitGroup{},
	}
}
