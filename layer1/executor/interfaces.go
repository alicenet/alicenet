package executor

import (
	"context"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
)

// TaskHandler to be implemented by the Handler that receives requests to schedule or
// kill a Task.
type TaskHandler interface {
	ScheduleTask(ctx context.Context, task tasks.Task, id string) (*HandlerResponse, error)
	KillTaskByType(ctx context.Context, task tasks.Task) (*HandlerResponse, error)
	KillTaskById(ctx context.Context, id string) (*HandlerResponse, error)
	Start()
	Close()
}

// TaskResponse to be implemented by a response structure that will be returned to the
// TaskHandler client.
type TaskResponse interface {
	IsReady() bool
	GetResponseBlocking(ctx context.Context) error
}
