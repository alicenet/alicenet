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

// DisputeMissingGPKjTask stores the data required to dispute shares
type DisputeMissingGPKjTask struct {
	*objects.Task
}

// asserting that DisputeMissingGPKjTask struct implements interface interfaces.Task
var _ interfaces.ITask = &DisputeMissingGPKjTask{}

// NewDisputeMissingGPKjTask creates a new task
func NewDisputeMissingGPKjTask(start uint64, end uint64) *DisputeMissingGPKjTask {
	return &DisputeMissingGPKjTask{
		Task: objects.NewTask(constants.DisputeMissingGPKjTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the DisputeMissingGPKjTask.
func (t *DisputeMissingGPKjTask) Prepare() *interfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debug("preparing task")
	return nil
}

// Execute executes the task business logic
func (t *DisputeMissingGPKjTask) Execute() ([]*types.Transaction, *interfaces.TaskErr) {
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
		logger.Warnf("accusing missing gpkj: %v", accusableParticipants)

		txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
		if err != nil {
			return nil, interfaces.NewTaskErr(fmt.Sprintf(constants.FailedGettingTxnOpts, err), true)
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantDidNotSubmitGPKJ(txnOpts, accusableParticipants)
		if err != nil {
			return nil, interfaces.NewTaskErr(fmt.Sprintf("error accusing missing gpkj: %v", err), true)
		}
		txns = append(txns, txn)
	} else {
		logger.Debug("no accusations for missing gpkj")
	}

	return txns, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *DisputeMissingGPKjTask) ShouldExecute() *interfaces.TaskErr {
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

	if dkgState.Phase != state.GPKJSubmission {
		return interfaces.NewTaskErr(fmt.Sprintf("phase %v different from GPKJSubmission", dkgState.Phase), false)
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

func (t *DisputeMissingGPKjTask) getAccusableParticipants(dkgState *state.DkgState) ([]common.Address, error) {
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

	// find participants who did not submit GPKj
	for _, p := range dkgState.Participants {
		_, isValidator := validatorsMap[p.Address]
		if isValidator && (p.Nonce != dkgState.Nonce ||
			p.Phase != state.GPKJSubmission ||
			(p.GPKj[0].Cmp(big.NewInt(0)) == 0 &&
				p.GPKj[1].Cmp(big.NewInt(0)) == 0 &&
				p.GPKj[2].Cmp(big.NewInt(0)) == 0 &&
				p.GPKj[3].Cmp(big.NewInt(0)) == 0)) {
			// did not submit
			accusableParticipants = append(accusableParticipants, p.Address)
		}
	}

	return accusableParticipants, nil
}
