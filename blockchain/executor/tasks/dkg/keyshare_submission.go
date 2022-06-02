package dkg

import (
	"context"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	dkgUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
	exUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/utils"

	ethereumInterfaces "github.com/MadBase/MadNet/blockchain/ethereum/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/sirupsen/logrus"
)

// KeyShareSubmissionTask is the task for submitting Keyshare information
type KeyShareSubmissionTask struct {
	*objects.Task
}

// asserting that KeyShareSubmissionTask struct implements interface interfaces.Task
var _ interfaces.ITask = &KeyShareSubmissionTask{}

// NewKeyShareSubmissionTask creates a new task
func NewKeyShareSubmissionTask(dkgState *state.DkgState, start uint64, end uint64) *KeyShareSubmissionTask {
	return &KeyShareSubmissionTask{
		Task: objects.NewTask(dkgState, constants.KeyShareSubmissionTaskName, start, end),
	}
}

// Initialize prepares for work to be done in KeyShareSubmission phase.
// Here, the G1 key share, G1 proof, and G2 key share are constructed
// and stored for submission.
func (t *KeyShareSubmissionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth ethereumInterfaces.IEthereum) error {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("KeyShareSubmissionTask Initialize()")

	taskState, ok := t.State.(*state.DkgState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	me := taskState.Account.Address

	// check if task already defined key shares
	if taskState.Participants[me].KeyShareG1s[0] == nil ||
		taskState.Participants[me].KeyShareG1s[1] == nil ||
		(taskState.Participants[me].KeyShareG1s[0].Cmp(big.NewInt(0)) == 0 &&
			taskState.Participants[me].KeyShareG1s[1].Cmp(big.NewInt(0)) == 0) {

		// Generate the key shares
		g1KeyShare, g1Proof, g2KeyShare, err := state.GenerateKeyShare(taskState.SecretValue)
		if err != nil {
			return err
		}

		taskState.Participants[me].KeyShareG1s = g1KeyShare
		taskState.Participants[me].KeyShareG1CorrectnessProofs = g1Proof
		taskState.Participants[me].KeyShareG2s = g2KeyShare
	} else {
		logger.Infof("KeyShareSubmissionTask Initialize(): key shares already defined")
	}

	return nil
}

// DoWork is the first attempt at the performing the KeyShareSubmission phase
func (t *KeyShareSubmissionTask) DoWork(ctx context.Context, logger *logrus.Entry, eth ethereumInterfaces.IEthereum) error {
	logger.Info("DoWork() ...")
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at the performing the KeyShareSubmission phase
func (t *KeyShareSubmissionTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth ethereumInterfaces.IEthereum) error {
	logger.Info("DoRetry() ...")
	return t.doTask(ctx, logger, eth)
}

func (t *KeyShareSubmissionTask) doTask(ctx context.Context, logger *logrus.Entry, eth ethereumInterfaces.IEthereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	taskState, ok := t.State.(*state.DkgState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	logger.Info("KeyShareSubmissionTask doTask()")

	// Setup
	me := taskState.Account

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, taskState.Account)
	if err != nil {
		return dkgUtils.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// If the TxOpts exists, meaning the Tx replacement timeout was reached,
	// we increase the Gas to have priority for the next blocks
	if t.TxOpts != nil && t.TxOpts.Nonce != nil {
		logger.Info("txnOpts Replaced")
		txnOpts.Nonce = t.TxOpts.Nonce
		txnOpts.GasFeeCap = t.TxOpts.GasFeeCap
		txnOpts.GasTipCap = t.TxOpts.GasTipCap
	}

	// Submit Keyshares
	logger.Infof("submitting key shares: %v %v %v %v",
		me.Address,
		taskState.Participants[me.Address].KeyShareG1s,
		taskState.Participants[me.Address].KeyShareG1CorrectnessProofs,
		taskState.Participants[me.Address].KeyShareG2s)
	txn, err := eth.Contracts().Ethdkg().SubmitKeyShare(txnOpts,
		taskState.Participants[me.Address].KeyShareG1s,
		taskState.Participants[me.Address].KeyShareG1CorrectnessProofs,
		taskState.Participants[me.Address].KeyShareG2s)
	if err != nil {
		return dkgUtils.LogReturnErrorf(logger, "submitting keyshare failed: %v", err)
	}
	t.TxOpts.TxHashes = append(t.TxOpts.TxHashes, txn.Hash())
	t.TxOpts.GasFeeCap = txn.GasFeeCap()
	t.TxOpts.GasTipCap = txn.GasTipCap()
	t.TxOpts.Nonce = big.NewInt(int64(txn.Nonce()))

	logger.WithFields(logrus.Fields{
		"GasFeeCap": t.TxOpts.GasFeeCap,
		"GasTipCap": t.TxOpts.GasTipCap,
		"Nonce":     t.TxOpts.Nonce,
	}).Info("key share submission fees")

	// Queue transaction
	eth.TransactionWatcher().Subscribe(ctx, txn)
	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
func (t *KeyShareSubmissionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth ethereumInterfaces.IEthereum) bool {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("KeyShareSubmissionTask ShouldRetry()")

	generalRetry := exUtils.GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	taskState, ok := t.State.(*state.DkgState)
	if !ok {
		logger.Error("Invalid convertion of taskState object")
		return false
	}

	me := taskState.Account
	callOpts, err := eth.GetCallOpts(ctx, me)
	if err != nil {
		logger.Debugf("KeyShareSubmissionTask ShouldRetry failed getting call options: %v", err)
		return true
	}

	phase, err := eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	if err != nil {
		logger.Infof("KeyShareSubmissionTask ShouldRetry GetETHDKGPhase error: %v", err)
		return true
	}

	// DisputeShareDistribution || KeyShareSubmission
	if phase != uint8(state.DisputeShareDistribution) && phase != uint8(state.KeyShareSubmission) {
		return false
	}

	// Check the key share submission status
	status, err := state.CheckKeyShare(ctx, eth.Contracts().Ethdkg(), logger, callOpts, me.Address, taskState.Participants[me.Address].KeyShareG1s)
	if err != nil {
		logger.Errorf("KeyShareSubmissionTask ShouldRetry CheckKeyShare error: %v", err)
		return true
	}

	if status == state.KeyShared || status == state.BadKeyShared {
		return false
	}

	return true
}

// DoDone creates a log entry saying task is complete
func (t *KeyShareSubmissionTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("KeyShareSubmissionTask done")
}

func (t *KeyShareSubmissionTask) GetExecutionData() interfaces.ITaskExecutionData {
	return t.Task
}
