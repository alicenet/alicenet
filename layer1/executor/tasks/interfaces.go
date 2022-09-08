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

// Task to be implemented by every task to be used by TaskHandler.
type Task interface {
	Lock()
	Unlock()
	Initialize(ctx context.Context, cancelFunc context.CancelFunc, database *db.Database, logger *logrus.Entry, eth layer1.Client, contracts layer1.AllSmartContracts, name string, id string, start uint64, end uint64, allowMultiExecution bool, subscribeOptions *transaction.SubscribeOptions, taskResponseChan InternalTaskResponseChan) error
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
	GetContractsHandler() layer1.AllSmartContracts
	GetLogger() *logrus.Entry
}

// TaskState to be implemented by every task for persistence.
type TaskState interface {
	PersistState(txn *badger.Txn) error
	LoadState(txn *badger.Txn) error
}

// InternalTaskResponseChan to be implemented by a response channel used
// for communication between the TaskManager and TaskExecutor.
type InternalTaskResponseChan interface {
	Add(id string, err error)
}
