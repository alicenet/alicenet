package dkgtasks

import (
	"context"
	"errors"
	"fmt"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/objects"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/utils"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/constants"
	"github.com/sirupsen/logrus"
)

// CompletionTask contains required state for safely complete ETHDKG
type CompletionTask struct {
	*objects.Task
}

// asserting that CompletionTask struct implements interface interfaces.Task
var _ interfaces.ITask = &CompletionTask{}

// NewCompletionTask creates a background task that attempts to call Complete on ethdkg
func NewCompletionTask(state *objects.DkgState, start uint64, end uint64) *CompletionTask {
	return &CompletionTask{
		Task: objects.NewTask(state, CompletionTaskName, start, end),
	}
}

// Initialize prepares for work to be done in the Completion phase
func (t *CompletionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("CompletionTask Initialize()...")

	taskState, ok := t.State.(*objects.DkgState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	if taskState.Phase != objects.DisputeGPKJSubmission {
		return fmt.Errorf("%w because it's not in DisputeGPKJSubmission phase", objects.ErrCanNotContinue)
	}

	// setup leader election
	block, err := eth.GetGethClient().BlockByNumber(ctx, big.NewInt(int64(t.Start)))
	if err != nil {
		return fmt.Errorf("CompletionTask.Initialize(): error getting block by number: %v", err)
	}

	logger.Infof("block hash: %v\n", block.Hash())
	t.StartBlockHash.SetBytes(block.Hash().Bytes())

	return nil
}

// DoWork is the first attempt
func (t *CompletionTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts
func (t *CompletionTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *CompletionTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {

	t.State.Lock()
	defer t.State.Unlock()

	taskState, ok := t.State.(*objects.DkgState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	logger.Info("CompletionTask doTask()")

	if t.isTaskCompleted(ctx, eth, logger, taskState) {
		t.Success = true
		return nil
	}

	// submit if I'm a leader for this task
	if !t.AmILeading(ctx, eth, logger, taskState) {
		return errors.New("not leading Completion yet")
	}

	// Setup
	c := eth.Contracts()
	txnOpts, err := eth.GetTransactionOpts(ctx, taskState.Account)
	if err != nil {
		return utils.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// If the TxOpts exists, meaning the Tx replacement timeout was reached,
	// we increase the Gas to have priority for the next blocks
	if t.TxOpts != nil && t.TxOpts.Nonce != nil {
		logger.Info("txnOpts Replaced")
		txnOpts.Nonce = t.TxOpts.Nonce
		txnOpts.GasFeeCap = t.TxOpts.GasFeeCap
		txnOpts.GasTipCap = t.TxOpts.GasTipCap
	}

	// Register
	txn, err := c.Ethdkg().Complete(txnOpts)
	if err != nil {
		return utils.LogReturnErrorf(logger, "completion failed: %v", err)
	}

	t.TxOpts.TxHashes = append(t.TxOpts.TxHashes, txn.Hash())
	t.TxOpts.GasFeeCap = txn.GasFeeCap()
	t.TxOpts.GasTipCap = txn.GasTipCap()
	t.TxOpts.Nonce = big.NewInt(int64(txn.Nonce()))

	logger.WithFields(logrus.Fields{
		"GasFeeCap": t.TxOpts.GasFeeCap,
		"GasTipCap": t.TxOpts.GasTipCap,
		"Nonce":     t.TxOpts.Nonce,
	}).Info("complete fees")

	logger.Info("CompletionTask sent completed call")

	// Queue transaction
	eth.TransactionWatcher().SubscribeTransaction(ctx, txn)

	logger.Info("CompletionTask complete!")
	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
// -- the registration open hasn't moved, i.e. ETHDKG has not restarted
func (t *CompletionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("CompletionTask ShouldRetry()")

	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	taskState, ok := t.State.(*objects.DkgState)
	if !ok {
		logger.Errorf("Invalid convertion of taskState object")
		return false
	}

	if t.isTaskCompleted(ctx, eth, logger, taskState) {
		logger.WithFields(logrus.Fields{
			"t.State.Phase":      taskState.Phase,
			"t.State.PhaseStart": taskState.PhaseStart,
		}).Info("CompletionTask ShouldRetry - will not retry because it's done")
		return false
	}

	logger.Info("CompletionTask ShouldRetry() will retry")

	return true
}

// DoDone creates a log entry saying task is complete
func (t *CompletionTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("CompletionTask done")
}

func (t *CompletionTask) GetExecutionData() interfaces.ITaskExecutionData {
	return t.Task
}

func (t *CompletionTask) isTaskCompleted(ctx context.Context, eth interfaces.Ethereum, logger *logrus.Entry, taskState *objects.DkgState) bool {
	c := eth.Contracts()

	callOpts, err := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	if err != nil {
		logger.Debugf("error getting call opts in completion task: %v", err)
		return false
	}
	phase, err := c.Ethdkg().GetETHDKGPhase(callOpts)
	if err != nil {
		logger.Debugf("error getting ethdkg phases in completion task: %v", err)
		return false
	}

	return phase == uint8(objects.Completion)
}

func (t *CompletionTask) AmILeading(ctx context.Context, eth interfaces.Ethereum, logger *logrus.Entry, taskState *objects.DkgState) bool {
	// check if I'm a leader for this task
	currentHeight, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return false
	}

	blocksSinceDesperation := int(currentHeight) - int(t.Start) - constants.ETHDKGDesperationDelay
	amILeading := utils.AmILeading(taskState.NumberOfValidators, taskState.Index-1, blocksSinceDesperation, t.StartBlockHash.Bytes(), logger)

	logger.WithFields(logrus.Fields{
		"currentHeight":                    currentHeight,
		"t.Start":                          t.Start,
		"constants.ETHDKGDesperationDelay": constants.ETHDKGDesperationDelay,
		"blocksSinceDesperation":           blocksSinceDesperation,
		"amILeading":                       amILeading,
	}).Infof("dkg.AmILeading")

	return amILeading
}
