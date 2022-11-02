package executor

import (
	"context"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	monitorInterfaces "github.com/alicenet/alicenet/layer1/monitor/interfaces"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/logging"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var _ TaskHandler = &Handler{}

type Handler struct {
	manager          *TaskManager
	logger           *logrus.Entry
	closeMainContext context.CancelFunc
	requestChannel   chan managerRequest
}

// NewTaskHandler creates a new Handler instance.
func NewTaskHandler(monDB, consDB *db.Database, eth layer1.Client, contracts layer1.AllSmartContracts, adminHandler monitorInterfaces.AdminHandler, txWatcher transaction.Watcher) (TaskHandler, error) {
	// main context that will cancel all workers and go routines
	mainCtx, cf := context.WithCancel(context.Background())

	// Setup tasks scheduler
	requestChan := make(chan managerRequest, tasks.ManagerBufferSize)
	logger := logging.GetLogger("tasks")

	taskManager, err := newTaskManager(mainCtx, eth, contracts, monDB, consDB, logger.WithField("Component", "TaskManager"), adminHandler, requestChan, txWatcher)
	if err != nil {
		cf()
		return nil, err
	}

	handler := &Handler{
		manager:          taskManager,
		logger:           logger.WithField("Component", "TaskHandler"),
		closeMainContext: cf,
		requestChannel:   requestChan,
	}

	// register gob types

	return handler, nil
}

// Start the Handler and the subsequent pieces such as TaskManager.
func (i *Handler) Start() {
	i.logger.Info("Starting task handler")
	i.manager.start()
}

// Close the Handler and the subsequent pieces such as TaskManager.
func (i *Handler) Close() {
	i.logger.Warn("Closing task handler")
	i.manager.close()
	i.closeMainContext()
	close(i.requestChannel)
}

// ScheduleTask sends the Schedule Task request to the TaskManager.
func (i *Handler) ScheduleTask(ctx context.Context, task tasks.Task, id string) (*HandlerResponse, error) {
	// In case the id field is not specified, create it
	if id == "" {
		id = uuid.New().String()
	}
	req := managerRequest{task: task, id: id, action: Schedule, response: NewManagerResponseChannel()}
	err := i.waitForRequestProcessing(ctx, req)
	if err != nil {
		return nil, err
	}
	return req.response.listen(ctx)
}

// KillTaskByType sends the KillByType Task request to the TaskManager.
func (i *Handler) KillTaskByType(ctx context.Context, taskType tasks.Task) (*HandlerResponse, error) {
	req := managerRequest{task: taskType, action: KillByType, response: NewManagerResponseChannel()}
	err := i.waitForRequestProcessing(ctx, req)
	if err != nil {
		return nil, err
	}
	return req.response.listen(ctx)
}

// KillTaskById sends the KillById Task request to the TaskManager.
func (i *Handler) KillTaskById(ctx context.Context, id string) (*HandlerResponse, error) {
	req := managerRequest{id: id, action: KillById, response: NewManagerResponseChannel()}
	err := i.waitForRequestProcessing(ctx, req)
	if err != nil {
		return nil, err
	}
	return req.response.listen(ctx)
}

// waitForRequestProcessing or context deadline.
func (i *Handler) waitForRequestProcessing(ctx context.Context, req managerRequest) error {
	// wait for request to be accepted
	select {
	case i.requestChannel <- req:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
