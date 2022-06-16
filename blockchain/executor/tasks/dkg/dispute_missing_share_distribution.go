package dkg

import (
	"errors"
	"fmt"

	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// DisputeMissingShareDistributionTask stores the data required to dispute shares
type DisputeMissingShareDistributionTask struct {
	*objects.Task
}

// asserting that DisputeMissingShareDistributionTask struct implements interface interfaces.Task
var _ interfaces.ITask = &DisputeMissingShareDistributionTask{}

// NewDisputeMissingShareDistributionTask creates a new task
func NewDisputeMissingShareDistributionTask(start uint64, end uint64) *DisputeMissingShareDistributionTask {
	return &DisputeMissingShareDistributionTask{
		Task: objects.NewTask(constants.DisputeMissingShareDistributionTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the DisputeMissingShareDistributionTask.
func (t *DisputeMissingShareDistributionTask) Prepare() *interfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debugf("preparing task")
	return nil
}

// Execute executes the task business logic
func (t *DisputeMissingShareDistributionTask) Execute() ([]*types.Transaction, *interfaces.TaskErr) {
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

		txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
		if err != nil {
			return nil, interfaces.NewTaskErr(fmt.Sprintf(constants.FailedGettingTxnOpts, err), true)
		}

		logger.Warnf("accusing participants: %v of not distributing shares", accusableParticipants)
		txn, err := eth.Contracts().Ethdkg().AccuseParticipantDidNotDistributeShares(txnOpts, accusableParticipants)
		if err != nil {
			return nil, interfaces.NewTaskErr(fmt.Sprintf("error accusing missing key shares: %v", err), true)
		}
		txns = append(txns, txn)
	} else {
		logger.Debug("No accusations for missing distributed shares")
	}

	return txns, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *DisputeMissingShareDistributionTask) ShouldExecute() *interfaces.TaskErr {
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

	if dkgState.Phase != state.ShareDistribution {
		return interfaces.NewTaskErr(fmt.Sprintf("phase %v different from ShareDistribution", dkgState.Phase), false)
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

func (t *DisputeMissingShareDistributionTask) getAccusableParticipants(dkgState *state.DkgState) ([]common.Address, error) {
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
