package tasks

import (
	"context"
	"github.com/alicenet/alicenet/layer1/executor"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// Task the interface requirements of a task
type Task interface {
	Initialize(ctx context.Context, cancelFunc context.CancelFunc, database *db.Database, logger *logrus.Entry, eth layer1.Client, name string, id string, taskResponseChan InternalTaskResponseChan) error
	Prepare(ctx context.Context) *TaskErr
	Execute(ctx context.Context) (*types.Transaction, *TaskErr)
	ShouldExecute(ctx context.Context) (bool, *TaskErr)
	WasKilled() bool
	Finish(err error)
	Close()
	GetId() string
	GetStart() uint64
	GetEnd() uint64
	GetName() string
	GetAllowMultiExecution() bool
	GetSubscribeOptions() *transaction.SubscribeOptions
	GetCtx() context.Context
	GetClient() layer1.Client
	GetLogger() *logrus.Entry
}

type TaskHandler interface {
	ScheduleTask(ctx context.Context, task Task, id string) (*executor.HandlerResponse, error)
	KillTaskByType(ctx context.Context, task Task) (*executor.HandlerResponse, error)
	KillTaskById(ctx context.Context, id string) (*executor.HandlerResponse, error)
	Start()
	Close()
}

// TaskState the interface requirements of a task state
type TaskState interface {
	PersistState(txn *badger.Txn) error
	LoadState(txn *badger.Txn) error
}

// InternalTaskResponseChan the interface requirements of a task response chan
type InternalTaskResponseChan interface {
	Add(executor.ExecutorResponse)
}

type TaskResponse interface {
	IsReady() bool
	GetResponseBlocking(ctx context.Context) error
}
