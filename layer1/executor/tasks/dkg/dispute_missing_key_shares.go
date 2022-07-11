package dkg

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/utils"
	"github.com/ethereum/go-ethereum/common"
)

// DisputeMissingKeySharesTask stores the data required to dispute shares
type DisputeMissingKeySharesTask struct {
	*tasks.BaseTask
}

// asserting that DisputeMissingKeySharesTask struct implements interface tasks.Task
var _ tasks.Task = &DisputeMissingKeySharesTask{}

// NewDisputeMissingKeySharesTask creates a new task
func NewDisputeMissingKeySharesTask(start uint64, end uint64) *DisputeMissingKeySharesTask {
	return &DisputeMissingKeySharesTask{
		BaseTask: tasks.NewBaseTask(start, end, false, nil),
	}
}

// Prepare prepares for work to be done in the DisputeMissingKeySharesTask.
func (t *DisputeMissingKeySharesTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debug("preparing task")
	return nil
}

// Execute executes the task business logic
func (t *DisputeMissingKeySharesTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
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
		logger.Debug("No accusations for missing key shares")
	}

	// accuse missing validators
	logger.Warnf("Accusing missing key shares: %v", accusableParticipants)

	txnOpts, err := client.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingTxnOpts, err), true)
	}

	txn, err := ethereum.GetContracts().Ethdkg().AccuseParticipantDidNotSubmitKeyShares(txnOpts, accusableParticipants)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("error accusing missing key shares: %v", err), true)
	}

	return txn, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *DisputeMissingKeySharesTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	if dkgState.Phase != state.KeyShareSubmission {
		logger.Debugf("phase %v different from KeyShareSubmission", dkgState.Phase)
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

func (t *DisputeMissingKeySharesTask) getAccusableParticipants(ctx context.Context, dkgState *state.DkgState) ([]common.Address, error) {
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

	// find participants who did not submit they key shares
	for _, p := range dkgState.Participants {
		_, isValidator := validatorsMap[p.Address]
		if isValidator && (p.Nonce != dkgState.Nonce ||
			p.Phase != state.KeyShareSubmission ||
			(p.KeyShareG1s[0].Cmp(big.NewInt(0)) == 0 &&
				p.KeyShareG1s[1].Cmp(big.NewInt(0)) == 0) ||
			(p.KeyShareG1CorrectnessProofs[0].Cmp(big.NewInt(0)) == 0 &&
				p.KeyShareG1CorrectnessProofs[1].Cmp(big.NewInt(0)) == 0) ||
			(p.KeyShareG2s[0].Cmp(big.NewInt(0)) == 0 &&
				p.KeyShareG2s[1].Cmp(big.NewInt(0)) == 0 &&
				p.KeyShareG2s[2].Cmp(big.NewInt(0)) == 0 &&
				p.KeyShareG2s[3].Cmp(big.NewInt(0)) == 0)) {
			// did not submit
			accusableParticipants = append(accusableParticipants, p.Address)
		}
	}

	return accusableParticipants, nil
}
