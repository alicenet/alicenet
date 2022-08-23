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
	"sync"
)

var _ TaskHandler = &Handler{}

type Handler struct {
	manager        *TaskManager
	logger         *logrus.Entry
	closeOnce      sync.Once
	closeChan      chan struct{}
	requestChannel chan managerRequest
}

// NewTaskHandler creates a new Handler instance.
func NewTaskHandler(database *db.Database, eth layer1.Client, contracts layer1.AllSmartContracts, adminHandler monitorInterfaces.AdminHandler, txWatcher transaction.Watcher) (TaskHandler, error) {
	// Setup tasks scheduler
	requestChan := make(chan managerRequest, tasks.ManagerBufferSize)
	logger := logging.GetLogger("tasks")

	taskManager, err := newTaskManager(eth, contracts, database, logger.WithField("Component", "TaskManager"), adminHandler, requestChan, txWatcher)
	if err != nil {
		return nil, err
	}

	handler := &Handler{
		manager:        taskManager,
		logger:         logger.WithField("Component", "TaskHandler"),
		closeChan:      make(chan struct{}),
		closeOnce:      sync.Once{},
		requestChannel: requestChan,
	}

	return handler, nil
}

// Start the Handler and the subsequent pieces such as TaskManager.
func (h *Handler) Start() {
	h.logger.Info("Starting task handler")
	h.manager.start()
}

// Close the Handler and the subsequent pieces such as TaskManager.
func (h *Handler) Close() {
	h.closeOnce.Do(func() {
		h.logger.Warn("Closing task handler")
		h.manager.close()
		close(h.requestChannel)
		close(h.closeChan)
	})
}

// CloseChan returns a channel that is closed when the Handler is
// shutting down.
func (h *Handler) CloseChan() <-chan struct{} {
	return h.closeChan
}

// ScheduleTask sends the Schedule Task request to the TaskManager.
func (h *Handler) ScheduleTask(ctx context.Context, task tasks.Task, id string) (*HandlerResponse, error) {
	// In case the id field is not specified, create it
	if id == "" {
		id = uuid.New().String()
	}
	req := managerRequest{task: task, id: id, action: Schedule, response: NewManagerResponseChannel()}
	err := h.waitForRequestProcessing(ctx, req)
	if err != nil {
		return nil, err
	}
	return req.response.listen(ctx)
}

// KillTaskByType sends the KillByType Task request to the TaskManager.
func (h *Handler) KillTaskByType(ctx context.Context, taskType tasks.Task) (*HandlerResponse, error) {
	req := managerRequest{task: taskType, action: KillByType, response: NewManagerResponseChannel()}
	err := h.waitForRequestProcessing(ctx, req)
	if err != nil {
		return nil, err
	}
	return req.response.listen(ctx)
}

// KillTaskById sends the KillById Task request to the TaskManager.
func (h *Handler) KillTaskById(ctx context.Context, id string) (*HandlerResponse, error) {
	req := managerRequest{id: id, action: KillById, response: NewManagerResponseChannel()}
	err := h.waitForRequestProcessing(ctx, req)
	if err != nil {
		return nil, err
	}
	return req.response.listen(ctx)
}

//waitForRequestProcessing or context deadline.
func (h *Handler) waitForRequestProcessing(ctx context.Context, req managerRequest) error {
	// wait for request to be accepted
	select {
	case h.requestChannel <- req:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
