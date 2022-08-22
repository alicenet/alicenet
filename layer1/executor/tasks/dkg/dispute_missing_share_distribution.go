package dkg

import (
	"context"
	"fmt"

	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// DisputeMissingShareDistributionTask stores the data required to dispute shares.
type DisputeMissingShareDistributionTask struct {
	*tasks.BaseTask
}

// asserting that DisputeMissingShareDistributionTask struct implements interface tasks.Task.
var _ tasks.Task = &DisputeMissingShareDistributionTask{}

// NewDisputeMissingShareDistributionTask creates a new task.
func NewDisputeMissingShareDistributionTask(start, end uint64) *DisputeMissingShareDistributionTask {
	return &DisputeMissingShareDistributionTask{
		BaseTask: tasks.NewBaseTask(start, end, false, nil),
	}
}

// Prepare prepares for work to be done in the DisputeMissingShareDistributionTask.
func (t *DisputeMissingShareDistributionTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debug("preparing task")
	return nil
}

// Execute executes the task business logic.
func (t *DisputeMissingShareDistributionTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()
	accusableParticipants, err := t.getAccusableParticipants(ctx, dkgState)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorGettingAccusableParticipants, err), true)
	}

	if len(accusableParticipants) <= 0 {
		logger.Debug("No accusations for missing distributed shares")
		return nil, nil
	}

	// accuse missing validators
	txnOpts, err := client.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingTxnOpts, err), true)
	}

	logger.Warnf("accusing participants: %v of not distributing shares", accusableParticipants)
	txn, err := t.GetContractsHandler().EthereumContracts().Ethdkg().AccuseParticipantDidNotDistributeShares(txnOpts, accusableParticipants)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("error accusing missing key shares: %v", err), true)
	}
	return txn, nil
}

// ShouldExecute checks if it makes sense to execute the task.
func (t *DisputeMissingShareDistributionTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	if dkgState.Phase != state.ShareDistribution {
		logger.Debugf("phase %v different from ShareDistribution", dkgState.Phase)
		return false, nil
	}

	accusableParticipants, err := t.getAccusableParticipants(ctx, dkgState)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorGettingAccusableParticipants, err), true)
	}

	if len(accusableParticipants) == 0 {
		logger.Debug(tasks.NobodyToAccuse)
		return false, nil
	}

	return true, nil
}

func (t *DisputeMissingShareDistributionTask) getAccusableParticipants(ctx context.Context, dkgState *state.DkgState) ([]common.Address, error) {
	logger := t.GetLogger()
	var accusableParticipants []common.Address

	validators, err := utils.GetValidatorAddresses(t.GetDB(), logger)
	if err != nil {
		return nil, fmt.Errorf(tasks.ErrorGettingValidators, err)
	}

	validatorsMap := make(map[common.Address]bool)
	for _, validator := range validators {
		validatorsMap[validator] = true
	}

	// find participants who did not submit their shares
	var emptySharesHash [32]byte
	for _, p := range dkgState.Participants {
		_, isValidator := validatorsMap[p.Address]
		if isValidator && (p.Nonce != dkgState.Nonce ||
			p.Phase != state.ShareDistribution ||
			p.DistributedSharesHash == emptySharesHash) {
			// did not distribute shares
			accusableParticipants = append(accusableParticipants, p.Address)
		}
	}

	return accusableParticipants, nil
}
