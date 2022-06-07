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
	exUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/utils"

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
		Task: objects.NewTask(constants.DisputeMissingGPKjTaskName, start, end),
	}
}

// Prepare prepares for work to be done in the DisputeMissingGPKjTask.
func (t *DisputeMissingGPKjTask) Prepare() error {
	t.GetLogger().Info("DisputeMissingGPKjTask Prepare()...")
	return nil
}

// Execute executes the task business logic
func (t *DisputeMissingGPKjTask) Execute() ([]*types.Transaction, error) {
	logger := t.GetLogger()
	logger.Info("DisputeMissingGPKjTask Execute()")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, t.GetLogger())
		return err
	})
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingGPKjTask.Execute(): error loading dkgState: %v", err)
	}

	ctx := t.GetCtx()
	eth := t.GetEth()
	accusableParticipants, err := t.getAccusableParticipants(dkgState)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingGPKjTask Execute() error getting accusableParticipants: %v", err)
	}

	// accuse missing validators
	txns := make([]*types.Transaction, 0)
	if len(accusableParticipants) > 0 {
		logger.Warnf("Accusing missing gpkj: %v", accusableParticipants)

		txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
		if err != nil {
			return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingGPKjTask Execute() error getting txnOpts: %v", err)
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantDidNotSubmitGPKJ(txnOpts, accusableParticipants)
		if err != nil {
			return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingGPKjTask Execute() error accusing missing gpkj: %v", err)
		}
		txns = append(txns, txn)
	} else {
		logger.Info("No accusations for missing gpkj")
	}

	return txns, nil
}

// ShouldRetry checks if it makes sense to try again
func (t *DisputeMissingGPKjTask) ShouldExecute() bool {
	logger := t.GetLogger()
	logger.Info("DisputeMissingGPKjTask ShouldExecute()")

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
		logger.Errorf("DisputeMissingGPKjTask.ShouldExecute(): error loading dkgState: %v", err)
		return true
	}

	if dkgState.Phase != state.GPKJSubmission {
		return false
	}

	accusableParticipants, err := t.getAccusableParticipants(dkgState)
	if err != nil {
		logger.Errorf("DisputeMissingGPKjTask ShouldExecute() error getting accusable participants: %v", err)
		return true
	}

	return len(accusableParticipants) > 0
}

func (t *DisputeMissingGPKjTask) getAccusableParticipants(dkgState *state.DkgState) ([]common.Address, error) {
	logger := t.GetLogger()
	ctx := t.GetCtx()
	eth := t.GetEth()

	var accusableParticipants []common.Address
	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingGPKjTask failed getting call options: %v", err)
	}

	validators, err := dkgUtils.GetValidatorAddressesFromPool(callOpts, eth, logger)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingGPKjTask getAccusableParticipants() error getting validators: %v", err)
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
