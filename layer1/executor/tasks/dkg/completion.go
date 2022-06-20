package dkg

import (
	"context"
	"fmt"
	"math/big"

	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/MadBase/MadNet/layer1/ethereum"
	"github.com/MadBase/MadNet/layer1/executor/constants"
	dkgConstants "github.com/MadBase/MadNet/layer1/executor/constants"
	"github.com/MadBase/MadNet/layer1/executor/tasks"
	"github.com/MadBase/MadNet/layer1/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/layer1/executor/tasks/dkg/utils"
	"github.com/MadBase/MadNet/layer1/transaction"
)

// CompletionTask contains required state for safely complete ETHDKG
type CompletionTask struct {
	*tasks.BaseTask
	// variables that are unique only for this task
	StartBlockHash common.Hash `json:"startBlockHash"`
}

// asserting that CompletionTask struct implements interface tasks.Task
var _ tasks.Task = &CompletionTask{}

// NewCompletionTask creates a background task that attempts to call Complete on ethdkg
func NewCompletionTask(start uint64, end uint64) *CompletionTask {
	return &CompletionTask{
		BaseTask: tasks.NewBaseTask(dkgConstants.CompletionTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the CompletionTask
func (t *CompletionTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debug("preparing task")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf(dkgConstants.ErrorLoadingDkgState, err), false)
	}

	if dkgState.Phase != state.DisputeGPKJSubmission {
		return tasks.NewTaskErr("it's not in DisputeGPKJSubmission phase", false)
	}

	// setup leader election
	block, err := t.GetClient().GetBlockByNumber(ctx, big.NewInt(int64(t.GetStart())))
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf("CompletionTask.Prepare(): error getting block by number: %v", err), true)
	}

	t.StartBlockHash = block.Hash()

	return nil
}

// Execute executes the task business logic
func (t *CompletionTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(dkgConstants.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()
	// submit if I'm a leader for this task
	if !utils.AmILeading(client, ctx, logger, int(t.GetStart()), t.StartBlockHash.Bytes(), dkgState.NumberOfValidators, dkgState.Index) {
		return nil, tasks.NewTaskErr(fmt.Sprintf("not leading Completion yet"), true)
	}

	c := ethereum.GetContracts()
	txnOpts, err := client.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(dkgConstants.FailedGettingTxnOpts, err), true)
	}

	// Complete ETHDKG
	logger.Info("Trying to complete ETHDKG")
	txn, err := c.Ethdkg().Complete(txnOpts)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("completion failed: %v", err), true)
	}

	return txn, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *CompletionTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	eth := t.GetClient()
	c := ethereum.GetContracts()

	callOpts, err := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(dkgConstants.FailedGettingCallOpts, err), true)
	}
	phase, err := c.Ethdkg().GetETHDKGPhase(callOpts)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf("error getting ethdkg phases in completion BaseTask: %v", err), true)
	}

	if phase == uint8(state.Completion) {
		logger.Debugf("completion already ocurred: %v", phase)
		return false, nil
	}

	return true, nil
}
