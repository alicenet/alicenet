package objects

import (
	"context"
	"errors"
	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/constants"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"math/big"
	"strings"
)

// ErrCanNotContinue standard error if we must drop out of ETHDKG
var (
	ErrCanNotContinue = errors.New("can not continue distributed key generation")
)

type Task struct {
	id               string
	name             string
	start            uint64
	end              uint64
	ctx              context.Context
	cancelFunc       context.CancelFunc
	database         *db.Database
	logger           *logrus.Entry
	eth              ethereum.Network
	taskResponseChan interfaces.ITaskResponseChan
	startBlockHash   common.Hash
	txOpts           *TxOpts
}

type TxOpts struct {
	TxHashes     []common.Hash
	Nonce        *big.Int
	GasFeeCap    *big.Int
	GasTipCap    *big.Int
	MinedInBlock uint64
}

func (to *TxOpts) GetHexTxsHashes() string {
	var hashes strings.Builder
	for _, txHash := range to.TxHashes {
		hashes.WriteString(txHash.Hex())
		hashes.WriteString(" ")
	}
	return hashes.String()
}

func NewTask(name string, start uint64, end uint64) *Task {
	ctx, cf := context.WithCancel(context.Background())

	return &Task{
		name:       name,
		start:      start,
		end:        end,
		ctx:        ctx,
		cancelFunc: cf,
		txOpts:     &TxOpts{TxHashes: make([]common.Hash, 0)},
	}
}

// Initialize default implementation for the ITask interface
func (t *Task) Initialize(ctx context.Context, cancelFunc context.CancelFunc, database *db.Database, logger *logrus.Entry, eth ethereum.Network, id string, taskResponseChan interfaces.ITaskResponseChan) {
	t.id = id
	t.ctx = ctx
	t.cancelFunc = cancelFunc
	t.database = database
	t.logger = logger
	t.eth = eth
	t.taskResponseChan = taskResponseChan
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

	t.taskResponseChan.Add(TaskResponse{Id: t.id, Err: err})
}

func (t *Task) GetLogger() *logrus.Entry {
	return t.logger
}

func (t *Task) GetDB() *db.Database {
	return t.database
}

func (t *Task) GetEth() ethereum.Network {
	return t.eth
}

func (t *Task) GetCtx() context.Context {
	return t.ctx
}

func (t *Task) SetStartBlockHash(startBlockHash []byte) {
	t.startBlockHash.SetBytes(startBlockHash)
}

func (t *Task) ClearTxData() {
	t.txOpts = &TxOpts{
		TxHashes: make([]common.Hash, 0),
	}
}

func (t *Task) AmILeading(dkgState *state.DkgState) bool {
	// check if I'm a leader for this task
	currentHeight, err := t.eth.GetCurrentHeight(t.ctx)
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

type TaskResponse struct {
	Id  string
	Err error
}
