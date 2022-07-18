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

// DisputeMissingRegistrationTask contains required state for accusing missing registrations
type DisputeMissingRegistrationTask struct {
	*tasks.BaseTask
}

// asserting that DisputeMissingRegistrationTask struct implements interface tasks.Task
var _ tasks.Task = &DisputeMissingRegistrationTask{}

// NewDisputeMissingRegistrationTask creates a background task to accuse missing registrations during ETHDKG
func NewDisputeMissingRegistrationTask(start uint64, end uint64) *DisputeMissingRegistrationTask {
	return &DisputeMissingRegistrationTask{
		BaseTask: tasks.NewBaseTask(start, end, false, nil),
	}
}

// Prepare prepares for work to be done in the DisputeMissingRegistrationTask
func (t *DisputeMissingRegistrationTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debug("preparing task")
	return nil
}

// Execute executes the task business logic
func (t *DisputeMissingRegistrationTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()
	accusableParticipants, err := t.getAccusableParticipants(dkgState)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorGettingAccusableParticipants, err), true)
	}

	if len(accusableParticipants) <= 0 {
		logger.Debug("No accusations for missing registrations")
		return nil, nil
	}

	// accuse missing validators
	logger.Warnf("Accusing missing registrations: %v", accusableParticipants)

	txnOpts, err := client.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingTxnOpts, err), true)
	}

	txn, err := ethereum.GetContracts().Ethdkg().AccuseParticipantNotRegistered(txnOpts, accusableParticipants)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("error accusing missing registration: %v", err), true)
	}

	return txn, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *DisputeMissingRegistrationTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	if dkgState.Phase != state.RegistrationOpen {
		logger.Debugf("phase %v different from RegistrationOpen", dkgState.Phase)
		return false, nil
	}

	accusableParticipants, err := t.getAccusableParticipants(dkgState)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorGettingAccusableParticipants, err), true)
	}

	if len(accusableParticipants) == 0 {
		logger.Debug(tasks.NobodyToAccuse)
		return false, nil
	}

	return true, nil
}

func (t *DisputeMissingRegistrationTask) getAccusableParticipants(dkgState *state.DkgState) ([]common.Address, error) {
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

	// find participants who did not register
	for _, addr := range dkgState.ValidatorAddresses {

		participant, ok := dkgState.Participants[addr]
		_, isValidator := validatorsMap[addr]

		if isValidator && (!ok ||
			participant.Nonce != dkgState.Nonce ||
			participant.Phase != state.RegistrationOpen ||
			(participant.PublicKey[0].Cmp(big.NewInt(0)) == 0 &&
				participant.PublicKey[1].Cmp(big.NewInt(0)) == 0)) {

			// did not register
			accusableParticipants = append(accusableParticipants, addr)
		}
	}

	return accusableParticipants, nil
}
