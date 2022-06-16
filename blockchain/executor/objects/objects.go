package objects

import (
	"context"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/constants"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

type Task struct {
	id                  string
	name                string
	start               uint64
	end                 uint64
	allowMultiExecution bool
	ctx                 context.Context
	cancelFunc          context.CancelFunc
	database            *db.Database
	logger              *logrus.Entry
	client              ethereum.Network
	taskResponseChan    interfaces.ITaskResponseChan
	startBlockHash      common.Hash
	subscribedTxns      []*types.Transaction
	subscribeOptions    *transaction.SubscribeOptions
}

func NewTask(name string, start uint64, end uint64, allowMultiExecution bool, subscribeOptions *transaction.SubscribeOptions) *Task {
	return &Task{
		name:                name,
		start:               start,
		end:                 end,
		allowMultiExecution: allowMultiExecution,
		subscribeOptions:    subscribeOptions,
	}
}

// Initialize default implementation for the ITask interface
func (t *Task) Initialize(ctx context.Context, cancelFunc context.CancelFunc, database *db.Database, logger *logrus.Entry, eth ethereum.Network, id string, taskResponseChan interfaces.ITaskResponseChan) {
	t.id = id
	t.ctx = ctx
	t.cancelFunc = cancelFunc
	t.database = database
	t.logger = logger
	t.client = eth
	t.taskResponseChan = taskResponseChan
}

// GetId default implementation for the ITask interface
func (t *Task) GetId() string {
	return t.id
}

// GetStart default implementation for the ITask interface
func (t *Task) GetStart() uint64 {
	return t.start
}

// GetEnd default implementation for the ITask interface
func (t *Task) GetEnd() uint64 {
	return t.end
}

// GetName default implementation for the ITask interface
func (t *Task) GetName() string {
	return t.name
}

// GetAllowMultiExecution default implementation for the ITask interface
func (t *Task) GetAllowMultiExecution() bool {
	return t.allowMultiExecution
}

func (t *Task) GetSubscribedTxs() []*types.Transaction {
	return t.subscribedTxns
}

func (t *Task) GetSubscribeOptions() *transaction.SubscribeOptions {
	return t.subscribeOptions
}

// GetCtx default implementation for the ITask interface
func (t *Task) GetCtx() context.Context {
	return t.ctx
}

// GetEth default implementation for the ITask interface
func (t *Task) GetClient() ethereum.Network {
	return t.client
}

// GetLogger default implementation for the ITask interface
func (t *Task) GetLogger() *logrus.Entry {
	return t.logger
}

// Close default implementation for the ITask interface
func (t *Task) Close() {
	if t.cancelFunc != nil {
		t.cancelFunc()
	}
}

// Finish default implementation for the ITask interface
func (t *Task) Finish(err error) {
	if err != nil {
		t.logger.WithError(err).Errorf("Id: %s, name: %s task is done", t.id, t.name)
	} else {
		t.logger.Infof("Id: %s, name: %s task is done", t.id, t.name)
	}

	t.taskResponseChan.Add(interfaces.TaskResponse{Id: t.id, Err: err})
}

func (t *Task) GetDB() *db.Database {
	return t.database
}

func (t *Task) SetStartBlockHash(startBlockHash []byte) {
	t.startBlockHash.SetBytes(startBlockHash)
}

func (t *Task) AmILeading(dkgState *state.DkgState) bool {
	// check if I'm a leader for this task
	currentHeight, err := t.client.GetCurrentHeight(t.ctx)
	if err != nil {
		return false
	}

	blocksSinceDesperation := int(currentHeight) - int(t.start) - constants.ETHDKGDesperationDelay
	amILeading := utils.AmILeading(dkgState.NumberOfValidators, dkgState.Index-1, blocksSinceDesperation, t.startBlockHash.Bytes(), t.logger)

	t.logger.WithFields(logrus.Fields{
		"currentHeight":                    currentHeight,
		"t.Start":                          t.start,
		"constants.ETHDKGDesperationDelay": constants.ETHDKGDesperationDelay,
		"blocksSinceDesperation":           blocksSinceDesperation,
		"amILeading":                       amILeading,
	}).Infof("dkg.AmILeading")

	return amILeading
}
