package executor

import (
	"context"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
)

// TaskHandler interface requirements
type TaskHandler interface {
	ScheduleTask(ctx context.Context, task tasks.Task, id string) (*HandlerResponse, error)
	KillTaskByType(ctx context.Context, task tasks.Task) (*HandlerResponse, error)
	KillTaskById(ctx context.Context, id string) (*HandlerResponse, error)
	Start()
	Close()
}

// TaskResponse interface requirements for the shared response
type TaskResponse interface {
	IsReady() bool
	GetResponseBlocking(ctx context.Context) error
}
