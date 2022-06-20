package dkg

import (
	"context"
	"errors"
	"fmt"

	"github.com/MadBase/MadNet/layer1/ethereum"
	"github.com/MadBase/MadNet/layer1/executor/constants"
	"github.com/MadBase/MadNet/layer1/executor/tasks"
	"github.com/MadBase/MadNet/layer1/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/layer1/executor/tasks/dkg/utils"
	"github.com/MadBase/MadNet/layer1/transaction"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// DisputeMissingShareDistributionTask stores the data required to dispute shares
type DisputeMissingShareDistributionTask struct {
	*tasks.BaseTask
}

// asserting that DisputeMissingShareDistributionTask struct implements interface tasks.Task
var _ tasks.Task = &DisputeMissingShareDistributionTask{}

// NewDisputeMissingShareDistributionTask creates a new task
func NewDisputeMissingShareDistributionTask(start uint64, end uint64) *DisputeMissingShareDistributionTask {
	return &DisputeMissingShareDistributionTask{
		BaseTask: tasks.NewBaseTask(constants.DisputeMissingShareDistributionTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the DisputeMissingShareDistributionTask.
func (t *DisputeMissingShareDistributionTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debug("preparing task")
	return nil
}

// Execute executes the task business logic
func (t *DisputeMissingShareDistributionTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()
	accusableParticipants, err := t.getAccusableParticipants(ctx, dkgState)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(constants.ErrorGettingAccusableParticipants, err), true)
	}

	if len(accusableParticipants) <= 0 {
		logger.Debug("No accusations for missing distributed shares")
		return nil, nil
	}

	// accuse missing validators
	txnOpts, err := client.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(constants.FailedGettingTxnOpts, err), true)
	}

	logger.Warnf("accusing participants: %v of not distributing shares", accusableParticipants)
	txn, err := ethereum.GetContracts().Ethdkg().AccuseParticipantDidNotDistributeShares(txnOpts, accusableParticipants)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("error accusing missing key shares: %v", err), true)
	}
	return txn, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *DisputeMissingShareDistributionTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	if dkgState.Phase != state.ShareDistribution {
		logger.Debugf("phase %v different from ShareDistribution", dkgState.Phase)
		return false, nil
	}

	accusableParticipants, err := t.getAccusableParticipants(ctx, dkgState)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(constants.ErrorGettingAccusableParticipants, err), true)
	}

	if len(accusableParticipants) == 0 {
		logger.Debug(constants.NobodyToAccuse)
		return false, nil
	}

	return true, nil
}

func (t *DisputeMissingShareDistributionTask) getAccusableParticipants(ctx context.Context, dkgState *state.DkgState) ([]common.Address, error) {
	logger := t.GetLogger()
	client := t.GetClient()

	var accusableParticipants []common.Address
	callOpts, err := client.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(constants.FailedGettingCallOpts, err))
	}

	validators, err := utils.GetValidatorAddressesFromPool(callOpts, client, logger)
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
