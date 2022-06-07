package dkg

import (
	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	dkgUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
	exUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/utils"
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
		Task: objects.NewTask(constants.DisputeMissingShareDistributionTaskName, start, end),
	}
}

// Prepare prepares for work to be done in the DisputeMissingShareDistributionTask.
func (t *DisputeMissingShareDistributionTask) Prepare() error {
	t.GetLogger().Info("DisputeMissingShareDistributionTask Prepare...")
	return nil
}

// Execute executes the task business logic
func (t *DisputeMissingShareDistributionTask) Execute() ([]*types.Transaction, error) {
	logger := t.GetLogger()
	logger.Info("DisputeMissingShareDistributionTask Execute()")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, t.GetLogger())
		return err
	})
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingShareDistributionTask.Execute(): error loading dkgState: %v", err)
	}

	ctx := t.GetCtx()
	eth := t.GetEth()
	accusableParticipants, err := t.getAccusableParticipants(dkgState)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingShareDistributionTask Execute() error getting accusableParticipants: %v", err)
	}

	// accuse missing validators
	txns := make([]*types.Transaction, 0)
	if len(accusableParticipants) > 0 {
		logger.Warnf("Accusing missing distributed shares: %v", accusableParticipants)

		txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
		if err != nil {
			return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingShareDistributionTask Execute() error getting txnOpts: %v", err)
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantDidNotDistributeShares(txnOpts, accusableParticipants)
		if err != nil {
			return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingShareDistributionTask Execute() error accusing missing key shares: %v", err)
		}
		txns = append(txns, txn)
	} else {
		logger.Info("No accusations for missing distributed shares")
	}

	return txns, nil
}

// ShouldRetry checks if it makes sense to try again
func (t *DisputeMissingShareDistributionTask) ShouldExecute() bool {
	logger := t.GetLogger()
	logger.Info("DisputeMissingShareDistributionTask ShouldExecute()")

	ctx := t.GetCtx()
	eth := t.GetEth()
	generalRetry := exUtils.GeneralTaskShouldRetry(ctx, logger, eth, t.GetStart(), t.GetEnd())
	if !generalRetry {
		return false
	}

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, t.GetLogger())
		return err
	})
	if err != nil {
		logger.Errorf("DisputeMissingShareDistributionTask.ShouldExecute(): error loading dkgState: %v", err)
		return true
	}

	if dkgState.Phase != state.ShareDistribution {
		return false
	}

	accusableParticipants, err := t.getAccusableParticipants(dkgState)
	if err != nil {
		logger.Errorf("DisputeMissingShareDistributionTask ShouldExecute() error getting accusable participants: %v", err)
		return true
	}

	return len(accusableParticipants) > 0
}

func (t *DisputeMissingShareDistributionTask) getAccusableParticipants(dkgState *state.DkgState) ([]common.Address, error) {
	logger := t.GetLogger()
	ctx := t.GetCtx()
	eth := t.GetEth()

	var accusableParticipants []common.Address
	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingShareDistributionTask failed getting call options: %v", err)
	}

	validators, err := dkgUtils.GetValidatorAddressesFromPool(callOpts, eth, logger)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingShareDistributionTask getAccusableParticipants() error getting validators: %v", err)
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
