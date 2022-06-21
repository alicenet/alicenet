package dkg

import (
	"bytes"
	"context"
	"fmt"

	"github.com/MadBase/MadNet/layer1/ethereum"
	"github.com/MadBase/MadNet/layer1/executor/constants"
	"github.com/MadBase/MadNet/layer1/executor/tasks"
	"github.com/MadBase/MadNet/layer1/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/layer1/transaction"
	"github.com/ethereum/go-ethereum/core/types"
)

// ShareDistributionTask stores the state required safely distribute shares
type ShareDistributionTask struct {
	*tasks.BaseTask
}

// asserting that ShareDistributionTask struct implements interface tasks.Task
var _ tasks.Task = &ShareDistributionTask{}

// NewShareDistributionTask creates a new task
func NewShareDistributionTask(start uint64, end uint64) *ShareDistributionTask {
	return &ShareDistributionTask{
		BaseTask: tasks.NewBaseTask(constants.ShareDistributionTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the ShareDistributionTask.
// We construct our commitments and encrypted shares before
// submitting them to the associated smart contract.
func (t *ShareDistributionTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debug("preparing task")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf("error during the preparation: %v", err), false)
	}

	if dkgState.Phase != state.ShareDistribution {
		return tasks.NewTaskErr("not in ShareDistribution phase", false)
	}

	if dkgState.SecretValue == nil {
		participants := dkgState.GetSortedParticipants()
		numParticipants := len(participants)
		threshold := state.ThresholdForUserCount(numParticipants)

		// Generate shares
		encryptedShares, privateCoefficients, commitments, err := state.GenerateShares(
			dkgState.TransportPrivateKey, participants)
		if err != nil {
			return tasks.NewTaskErr(fmt.Sprintf("Failed to generate shares: %v %#v", err, participants), true)
		}

		// Store calculated values
		dkgState.Participants[dkgState.Account.Address].Commitments = commitments
		dkgState.Participants[dkgState.Account.Address].EncryptedShares = encryptedShares

		dkgState.PrivateCoefficients = privateCoefficients
		dkgState.SecretValue = privateCoefficients[0]
		dkgState.ValidatorThreshold = threshold

		err = state.SaveDkgState(t.GetDB(), dkgState)
		if err != nil {
			return tasks.NewTaskErr(fmt.Sprintf(constants.ErrorDuringPreparation, err), false)
		}
	} else {
		logger.Debug("encrypted shares already defined")
	}

	return nil
}

// Execute executes the task business logic
func (t *ShareDistributionTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("error loading dkgState: %v", err), false)
	}

	client := t.GetClient()
	contracts := ethereum.GetContracts()
	accountAddr := dkgState.Account.Address

	// Setup
	txnOpts, err := client.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("getting txn opts failed: %v", err), true)
	}

	logger.Info("distributing shares")
	// Distribute shares
	txn, err := contracts.Ethdkg().DistributeShares(
		txnOpts,
		dkgState.Participants[accountAddr].EncryptedShares,
		dkgState.Participants[accountAddr].Commitments,
	)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("distributing shares failed: %v", err), true)
	}

	return txn, nil
}

// ShouldRetry checks if it makes sense to try again
func (t *ShareDistributionTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf("could not get dkgState with error %v", err), false)
	}

	eth := t.GetClient()
	if dkgState.Phase != state.ShareDistribution {
		logger.Debugf("phase %v different from ShareDistribution", dkgState.Phase)
		return false, nil
	}

	// If it's generally good to retry, let's try to be more specific
	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf("failed getting call options: %v", err), true)
	}
	participantState, err := ethereum.GetContracts().Ethdkg().GetParticipantInternalState(callOpts, dkgState.Account.Address)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf("unable to GetParticipantInternalState(): %v", err), true)
	}

	logger.Debugf("DistributionHash: %x", participantState.DistributedSharesHash)
	var emptySharesHash [32]byte
	if !bytes.Equal(participantState.DistributedSharesHash[:], emptySharesHash[:]) {
		logger.Debug("did distribute shares after all. needs no retry")
		return false, nil
	}

	logger.Debugf("Did not distribute shares after all. needs retry")
	return true, nil
}
