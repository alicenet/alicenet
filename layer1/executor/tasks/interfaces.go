package tasks

import (
	"context"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// Task the interface requirements of a task
type Task interface {
	Initialize(ctx context.Context, cancelFunc context.CancelFunc, database *db.Database, logger *logrus.Entry, eth layer1.Client, name string, id string, taskResponseChan TaskResponseChan) error
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

// TaskState the interface requirements of a task state
type TaskState interface {
	PersistState(txn *badger.Txn) error
	LoadState(txn *badger.Txn) error
}

// TaskResponseChan the interface requirements of a task response chan
type TaskResponseChan interface {
	Add(TaskResponse)
}
