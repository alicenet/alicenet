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

// DisputeMissingKeySharesTask stores the data required to dispute shares
type DisputeMissingKeySharesTask struct {
	*objects.Task
}

// asserting that DisputeMissingKeySharesTask struct implements interface interfaces.Task
var _ interfaces.ITask = &DisputeMissingKeySharesTask{}

// NewDisputeMissingKeySharesTask creates a new task
func NewDisputeMissingKeySharesTask(start uint64, end uint64) *DisputeMissingKeySharesTask {
	return &DisputeMissingKeySharesTask{
		Task: objects.NewTask(constants.DisputeMissingKeySharesTaskName, start, end),
	}
}

// Prepare prepares for work to be done in the DisputeMissingKeySharesTask.
func (t *DisputeMissingKeySharesTask) Prepare() error {
	t.GetLogger().Info("DisputeMissingKeySharesTask Prepare()...")
	return nil
}

// Execute executes the task business logic
func (t *DisputeMissingKeySharesTask) Execute() ([]*types.Transaction, error) {
	logger := t.GetLogger()
	logger.Info("DisputeMissingKeySharesTask Execute()")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, t.GetLogger())
		return err
	})
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingKeySharesTask.Execute(): error loading dkgState: %v", err)
	}

	ctx := t.GetCtx()
	eth := t.GetEth()
	accusableParticipants, err := t.getAccusableParticipants(dkgState)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingKeySharesTask Execute() error getting accusableParticipants: %v", err)
	}

	// accuse missing validators
	txns := make([]*types.Transaction, 0)
	if len(accusableParticipants) > 0 {
		logger.Warnf("Accusing missing key shares: %v", accusableParticipants)

		txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
		if err != nil {
			return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingKeySharesTask Execute() error getting txnOpts: %v", err)
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantDidNotSubmitKeyShares(txnOpts, accusableParticipants)
		if err != nil {
			return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingKeySharesTask Execute() error accusing missing key shares: %v", err)
		}
		txns = append(txns, txn)
	} else {
		logger.Info("No accusations for missing key shares")
	}

	return txns, nil
}

// ShouldRetry checks if it makes sense to try again
func (t *DisputeMissingKeySharesTask) ShouldExecute() bool {
	logger := t.GetLogger()
	logger.Info("DisputeMissingKeySharesTask ShouldExecute()")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, t.GetLogger())
		return err
	})
	if err != nil {
		logger.Errorf("DisputeMissingKeySharesTask.ShouldExecute(): error loading dkgState: %v", err)
		return true
	}

	ctx := t.GetCtx()
	eth := t.GetEth()
	generalRetry := exUtils.GeneralTaskShouldRetry(ctx, logger, eth, t.GetStart(), t.GetEnd())
	if !generalRetry {
		return false
	}

	if dkgState.Phase != state.KeyShareSubmission {
		return false
	}

	accusableParticipants, err := t.getAccusableParticipants(dkgState)
	if err != nil {
		logger.Error("could not get accusableParticipants")
		return true
	}

	return len(accusableParticipants) > 0
}

func (t *DisputeMissingKeySharesTask) getAccusableParticipants(dkgState *state.DkgState) ([]common.Address, error) {
	logger := t.GetLogger()
	ctx := t.GetCtx()
	eth := t.GetEth()

	var accusableParticipants []common.Address
	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingKeySharesTask failed getting call options: %v", err)
	}

	validators, err := dkgUtils.GetValidatorAddressesFromPool(callOpts, eth, logger)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingKeySharesTask getAccusableParticipants() error getting validators: %v", err)
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
