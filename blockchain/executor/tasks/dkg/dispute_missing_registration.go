package dkg

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	dkgUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
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
		Task: objects.NewTask(constants.DisputeMissingRegistrationTaskName, start, end, false),
	}
}

// Prepare prepares for work to be done in the DisputeMissingRegistrationTask
func (t *DisputeMissingRegistrationTask) Prepare() error {
	t.GetLogger().Info("DisputeMissingRegistrationTask Prepare()...")
	return nil
}

// Execute executes the task business logic
func (t *DisputeMissingRegistrationTask) Execute() ([]*types.Transaction, error) {
	logger := t.GetLogger()
	logger.Info("DisputeMissingRegistrationTask Execute()")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, t.GetLogger())
		return err
	})
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingRegistrationTask.Execute(): error loading dkgState: %v", err)
	}

	ctx := t.GetCtx()
	eth := t.GetEth()
	accusableParticipants, err := t.getAccusableParticipants(dkgState)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingRegistrationTask Execute() error getting accusable participants: %v", err)
	}

	// accuse missing validators
	txns := make([]*types.Transaction, 0)
	if len(accusableParticipants) > 0 {
		logger.Warnf("Accusing missing registrations: %v", accusableParticipants)

		txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
		if err != nil {
			return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingRegistrationTask Execute() error getting txnOpts: %v", err)
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantNotRegistered(txnOpts, accusableParticipants)
		if err != nil {
			return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingRegistrationTask Execute() error accusing missing registration: %v", err)
		}
		txns = append(txns, txn)
	} else {
		logger.Info("No accusations for missing registrations")
	}

	return txns, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *DisputeMissingRegistrationTask) ShouldExecute() bool {
	logger := t.GetLogger()
	logger.Info("DisputeMissingRegistrationTask ShouldExecute()")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, t.GetLogger())
		return err
	})
	if err != nil {
		logger.Errorf("DisputeMissingRegistrationTask.ShouldExecute(): error loading dkgState: %v", err)
		return true
	}

	if dkgState.Phase != state.RegistrationOpen {
		return false
	}

	accusableParticipants, err := t.getAccusableParticipants(dkgState)
	if err != nil {
		logger.Errorf("DisputeMissingRegistrationTask ShouldExecute() error getting accusable participants: %v", err)
		return true
	}

	return len(accusableParticipants) > 0
}

func (t *DisputeMissingRegistrationTask) getAccusableParticipants(dkgState *state.DkgState) ([]common.Address, error) {
	logger := t.GetLogger()
	ctx := t.GetCtx()
	eth := t.GetEth()

	var accusableParticipants []common.Address
	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingRegistrationTask failed getting call options: %v", err)
	}

	validators, err := dkgUtils.GetValidatorAddressesFromPool(callOpts, eth, logger)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingRegistrationTask getAccusableParticipants() error getting validators: %v", err)
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
