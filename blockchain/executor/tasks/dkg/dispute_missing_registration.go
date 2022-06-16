package dkg

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/ethereum/go-ethereum/common"
)

// DisputeMissingRegistrationTask contains required state for accusing missing registrations
type DisputeMissingRegistrationTask struct {
	*objects.Task
}

// asserting that DisputeMissingRegistrationTask struct implements interface interfaces.Task
var _ interfaces.ITask = &DisputeMissingRegistrationTask{}

// NewDisputeMissingRegistrationTask creates a background task to accuse missing registrations during ETHDKG
func NewDisputeMissingRegistrationTask(start uint64, end uint64) *DisputeMissingRegistrationTask {
	return &DisputeMissingRegistrationTask{
		Task: objects.NewTask(constants.DisputeMissingRegistrationTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the DisputeMissingRegistrationTask
func (t *DisputeMissingRegistrationTask) Prepare() *interfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debug("preparing task")
	return nil
}

// Execute executes the task business logic
func (t *DisputeMissingRegistrationTask) Execute() ([]*types.Transaction, *interfaces.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	ctx := t.GetCtx()
	eth := t.GetClient()
	accusableParticipants, err := t.getAccusableParticipants(dkgState)
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf(constants.ErrorGettingAccusableParticipants, err), true)
	}

	// accuse missing validators
	txns := make([]*types.Transaction, 0)
	if len(accusableParticipants) > 0 {
		logger.Warnf("Accusing missing registrations: %v", accusableParticipants)

		txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
		if err != nil {
			return nil, interfaces.NewTaskErr(fmt.Sprintf(constants.FailedGettingTxnOpts, err), true)
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantNotRegistered(txnOpts, accusableParticipants)
		if err != nil {
			return nil, interfaces.NewTaskErr(fmt.Sprintf("error accusing missing registration: %v", err), true)
		}
		txns = append(txns, txn)
	} else {
		logger.Debug("No accusations for missing registrations")
	}

	return txns, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *DisputeMissingRegistrationTask) ShouldExecute() *interfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	if dkgState.Phase != state.RegistrationOpen {
		return interfaces.NewTaskErr(fmt.Sprintf("phase %v different from RegistrationOpen", dkgState.Phase), false)
	}

	accusableParticipants, err := t.getAccusableParticipants(dkgState)
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf(constants.ErrorGettingAccusableParticipants, err), true)
	}

	if len(accusableParticipants) == 0 {
		return interfaces.NewTaskErr(fmt.Sprintf(constants.NobodyToAccuse), false)
	}

	return nil
}

func (t *DisputeMissingRegistrationTask) getAccusableParticipants(dkgState *state.DkgState) ([]common.Address, error) {
	logger := t.GetLogger()
	ctx := t.GetCtx()
	eth := t.GetClient()

	var accusableParticipants []common.Address
	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(constants.FailedGettingCallOpts, err))
	}

	validators, err := utils.GetValidatorAddressesFromPool(callOpts, eth, logger)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(constants.ErrorGettingValidators, err))
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
