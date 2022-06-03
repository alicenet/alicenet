package interfaces

import (
	"context"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ITask the interface requirements of a task
type ITask interface {
	DoDone(*logrus.Entry)
	DoRetry(context.Context, *logrus.Entry, ethereum.Network) error
	DoWork(context.Context, *logrus.Entry, ethereum.Network) error
	Initialize(context.Context, *logrus.Entry, ethereum.Network) error
	ShouldRetry(context.Context, *logrus.Entry, ethereum.Network) bool
	GetExecutionData() ITaskExecutionData
}

type ITaskState interface {
	Lock()
	Unlock()
}

type ITaskExecutionData interface {
	Lock()
	Unlock()
	ClearTxData()
	GetStart() uint64
	GetEnd() uint64
	GetName() string
	SetId(string)
	SetContext(ctx context.Context, cancel context.CancelFunc)
	Close()
}

// IScheduler simple interface to a block based schedule
type IScheduler interface {
	Schedule(start uint64, end uint64, thing ITask) (uuid.UUID, error)
	Purge()
	PurgePrior(now uint64)
	Find(now uint64) (uuid.UUID, error)
	Retrieve(taskId uuid.UUID) (ITask, error)
	Length() int
	Remove(taskId uuid.UUID) error
	Status(logger *logrus.Entry)
}
