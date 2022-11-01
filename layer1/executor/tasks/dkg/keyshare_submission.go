package dkg

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
)

// KeyShareSubmissionTask is the task for submitting KeyShare information.
type KeyShareSubmissionTask struct {
	*tasks.BaseTask
}

// asserting that KeyShareSubmissionTask struct implements interface tasks.Task.
var _ tasks.Task = &KeyShareSubmissionTask{}

// NewKeyShareSubmissionTask creates a new task.
func NewKeyShareSubmissionTask(start, end uint64) *KeyShareSubmissionTask {
	return &KeyShareSubmissionTask{
		BaseTask: tasks.NewBaseTask(start, end, false, nil),
	}
}

// Prepare prepares for work to be done in the KeyShareSubmissionTask.
// Here, the G1 key share, G1 proof, and G2 key share are constructed
// and stored for submission.
func (t *KeyShareSubmissionTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debug("preparing task")

	dkgState, err := state.GetDkgState(t.GetMonDB())
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorDuringPreparation, err), false)
	}

	defaultAddr := dkgState.Account.Address

	isKeyShareNil := dkgState.Participants[defaultAddr].KeyShareG1s[0] == nil ||
		dkgState.Participants[defaultAddr].KeyShareG1s[1] == nil

	// check if task already defined key shares
	if isKeyShareNil || (dkgState.Participants[defaultAddr].KeyShareG1s[0].Cmp(big.NewInt(0)) == 0 &&
		dkgState.Participants[defaultAddr].KeyShareG1s[1].Cmp(big.NewInt(0)) == 0) {
		// Generate the key shares. If this function fails it means that we don't have
		// all the data or we have bad data stored in state. No way to recover
		g1KeyShare, g1Proof, g2KeyShare, err := state.GenerateKeyShare(dkgState.SecretValue)
		if err != nil {
			return tasks.NewTaskErr(fmt.Sprintf("failed to generate key share: %v", err), false)
		}

		dkgState.Participants[defaultAddr].KeyShareG1s = g1KeyShare
		dkgState.Participants[defaultAddr].KeyShareG1CorrectnessProofs = g1Proof
		dkgState.Participants[defaultAddr].KeyShareG2s = g2KeyShare

		err = state.SaveDkgState(t.GetMonDB(), dkgState)
		if err != nil {
			return tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorDuringPreparation, err), false)
		}
	} else {
		logger.Debugf("key shares already defined")
	}

	return nil
}

// Execute executes the task business logic.
func (t *KeyShareSubmissionTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	dkgState, err := state.GetDkgState(t.GetMonDB())
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	// Setup
	defaultAddr := dkgState.Account

	// Setup
	eth := t.GetClient()
	txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingTxnOpts, err), true)
	}

	// Submit KeyShares
	logger.Infof(
		"submitting key shares with account: %v g1s: %v correctnessProofs: %v keyShareG2s: %v",
		defaultAddr.Address,
		dkgState.Participants[defaultAddr.Address].KeyShareG1s,
		dkgState.Participants[defaultAddr.Address].KeyShareG1CorrectnessProofs,
		dkgState.Participants[defaultAddr.Address].KeyShareG2s,
	)
	txn, err := t.GetContractsHandler().EthereumContracts().Ethdkg().SubmitKeyShare(txnOpts,
		dkgState.Participants[defaultAddr.Address].KeyShareG1s,
		dkgState.Participants[defaultAddr.Address].KeyShareG1CorrectnessProofs,
		dkgState.Participants[defaultAddr.Address].KeyShareG2s)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("registering failed: %v", err), true)
	}

	return txn, nil
}

// ShouldExecute checks if it makes sense to execute the task.
func (t *KeyShareSubmissionTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	dkgState, err := state.GetDkgState(t.GetMonDB())
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()
	defaultAccount := dkgState.Account
	callOpts, err := client.GetCallOpts(ctx, defaultAccount)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingCallOpts, err), true)
	}

	phase, err := t.GetContractsHandler().EthereumContracts().Ethdkg().GetETHDKGPhase(callOpts)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf("error getting ETHDKGPhase: %v", err), true)
	}

	// DisputeShareDistribution || KeyShareSubmission
	if phase != uint8(state.DisputeShareDistribution) && phase != uint8(state.KeyShareSubmission) {
		logger.Debugf("on dispute ShareDistribution phase should not submit keyShare")
		return false, nil
	}

	// Check the key share submission status
	status, err := state.CheckKeyShare(ctx, t.GetContractsHandler().EthereumContracts().Ethdkg(), logger, callOpts, defaultAccount.Address, dkgState.Participants[defaultAccount.Address].KeyShareG1s)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf("error checkingKeyShare: %v", err), true)
	}

	if status == state.KeyShared || status == state.BadKeyShared {
		logger.Debug("already shared keyShare")
		return false, nil
	}

	return true, nil
}
