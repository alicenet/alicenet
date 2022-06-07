package dkg

import (
	"fmt"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"

	dkgConstants "github.com/MadBase/MadNet/blockchain/executor/constants"
	executorInterfaces "github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
	taskUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/utils"
)

// CompletionTask contains required state for safely complete ETHDKG
type CompletionTask struct {
	*objects.Task
}

// asserting that CompletionTask struct implements interface interfaces.Task
var _ executorInterfaces.ITask = &CompletionTask{}

// NewCompletionTask creates a background task that attempts to call Complete on ethdkg
func NewCompletionTask(start uint64, end uint64) *CompletionTask {
	return &CompletionTask{
		Task: objects.NewTask(dkgConstants.CompletionTaskName, start, end),
	}
}

// Prepare prepares for work to be done in the CompletionTask
func (t *CompletionTask) Prepare() error {
	logger := t.GetLogger()
	logger.Info("CompletionTask Prepare()...")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, logger)
		return err
	})
	if err != nil {
		return fmt.Errorf("CompletionTask.Prepare(): error loading dkgState: %v", err)
	}

	if dkgState.Phase != state.DisputeGPKJSubmission {
		return fmt.Errorf("%w because it's not in DisputeGPKJSubmission phase", objects.ErrCanNotContinue)
	}

	// setup leader election
	block, err := t.GetEth().GetInternalClient().BlockByNumber(t.GetCtx(), big.NewInt(int64(t.GetStart())))
	if err != nil {
		return fmt.Errorf("CompletionTask.Prepare(): error getting block by number: %v", err)
	}

	logger.Infof("block hash: %v\n", block.Hash())
	t.SetStartBlockHash(block.Hash().Bytes())

	return nil
}

// Execute executes the task business logic
func (t *CompletionTask) Execute() ([]*types.Transaction, error) {
	logger := t.GetLogger()
	logger.Info("CompletionTask Execute()")

	if t.isTaskCompleted() {
		return nil, nil
	}

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, t.GetLogger())
		return err
	})
	if err != nil {
		return nil, utils.LogReturnErrorf(logger, "CompletionTask.Execute(): error loading dkgState: %v", err)
	}

	// submit if I'm a leader for this task
	if !t.AmILeading(dkgState) {
		return nil, utils.LogReturnErrorf(logger, "not leading Completion yet")
	}

	// Setup
	eth := t.GetEth()
	c := eth.Contracts()
	txnOpts, err := eth.GetTransactionOpts(t.GetCtx(), dkgState.Account)
	if err != nil {
		return nil, utils.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// Register
	txn, err := c.Ethdkg().Complete(txnOpts)
	if err != nil {
		return nil, utils.LogReturnErrorf(logger, "completion failed: %v", err)
	}

	return []*types.Transaction{txn}, nil
}

// ShouldRetry checks if it makes sense to try again
func (t *CompletionTask) ShouldExecute() bool {
	logger := t.GetLogger()
	logger.Info("CompletionTask ShouldExecute()")

	generalRetry := taskUtils.GeneralTaskShouldRetry(t.GetCtx(), logger, t.GetEth(), t.GetStart(), t.GetEnd())
	if !generalRetry {
		return false
	}

	if t.isTaskCompleted() {
		return false
	}

	logger.Info("CompletionTask ShouldExecute() will retry")

	return true
}

func (t *CompletionTask) isTaskCompleted() bool {
	logger := t.GetLogger()
	eth := t.GetEth()
	c := eth.Contracts()

	callOpts, err := eth.GetCallOpts(t.GetCtx(), eth.GetDefaultAccount())
	if err != nil {
		logger.Debugf("error getting call opts in completion task: %v", err)
		return false
	}
	phase, err := c.Ethdkg().GetETHDKGPhase(callOpts)
	if err != nil {
		logger.Debugf("error getting ethdkg phases in completion task: %v", err)
		return false
	}

	return phase == uint8(state.Completion)
}
