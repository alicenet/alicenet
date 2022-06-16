package dkg

import (
	"fmt"
	"math/big"

	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/MadBase/MadNet/blockchain/executor/constants"
	dkgConstants "github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/blockchain/transaction"
)

// CompletionTask contains required state for safely complete ETHDKG
type CompletionTask struct {
	*objects.Task
}

// asserting that CompletionTask struct implements interface interfaces.Task
var _ interfaces.ITask = &CompletionTask{}

// NewCompletionTask creates a background task that attempts to call Complete on ethdkg
func NewCompletionTask(start uint64, end uint64) *CompletionTask {
	return &CompletionTask{
		Task: objects.NewTask(dkgConstants.CompletionTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the CompletionTask
func (t *CompletionTask) Prepare() *interfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debug("preparing task")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf(dkgConstants.ErrorLoadingDkgState, err), false)
	}

	if dkgState.Phase != state.DisputeGPKJSubmission {
		return interfaces.NewTaskErr("it's not in DisputeGPKJSubmission phase", false)
	}

	// setup leader election
	block, err := t.GetClient().GetBlockByNumber(t.GetCtx(), big.NewInt(int64(t.GetStart())))
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf("CompletionTask.Prepare(): error getting block by number: %v", err), true)
	}

	t.SetStartBlockHash(block.Hash().Bytes())

	return nil
}

// Execute executes the task business logic
func (t *CompletionTask) Execute() ([]*types.Transaction, *interfaces.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf(dkgConstants.ErrorLoadingDkgState, err), false)
	}

	// submit if I'm a leader for this task
	if !t.AmILeading(dkgState) {
		return nil, interfaces.NewTaskErr(fmt.Sprintf("not leading Completion yet"), true)
	}

	// Setup
	eth := t.GetClient()
	c := eth.Contracts()
	txnOpts, err := eth.GetTransactionOpts(t.GetCtx(), dkgState.Account)
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf(dkgConstants.FailedGettingTxnOpts, err), true)
	}

	// Complete ETHDKG
	logger.Info("Trying to complete ETHDKG")
	txn, err := c.Ethdkg().Complete(txnOpts)
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf("completion failed: %v", err), true)
	}

	return []*types.Transaction{txn}, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *CompletionTask) ShouldExecute() *interfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	eth := t.GetClient()
	c := eth.Contracts()

	callOpts, err := eth.GetCallOpts(t.GetCtx(), eth.GetDefaultAccount())
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf(dkgConstants.FailedGettingCallOpts, err), true)
	}
	phase, err := c.Ethdkg().GetETHDKGPhase(callOpts)
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf("error getting ethdkg phases in completion task: %v", err), true)
	}

	if phase == uint8(state.Completion) {
		return interfaces.NewTaskErr(fmt.Sprintf("completion already ocurred: %v", phase), false)
	}

	return nil
}
