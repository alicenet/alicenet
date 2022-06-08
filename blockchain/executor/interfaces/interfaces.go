package interfaces

import (
	"context"
	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// ITask the interface requirements of a task
type ITask interface {
	Initialize(ctx context.Context, cancelFunc context.CancelFunc, database *db.Database, logger *logrus.Entry, eth ethereum.Network, id string, taskResponseChan ITaskResponseChan)
	Prepare() error
	Execute() ([]*types.Transaction, error)
	ShouldExecute() bool
	Finish(err error)
	Close()
	GetId() string
	GetStart() uint64
	GetEnd() uint64
	GetName() string
	GetAllowMultiExecution() bool
	GetCtx() context.Context
	GetEth() ethereum.Network
	GetLogger() *logrus.Entry
}

// ITaskState the interface requirements of a task state
type ITaskState interface {
	PersistState(txn *badger.Txn) error
	LoadState(txn *badger.Txn) error
}

// ITaskResponseChan the interface requirements of a task response chan
type ITaskResponseChan interface {
	Add(objects.TaskResponse)
}
