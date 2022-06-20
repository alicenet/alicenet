package dkg

import (
	"context"
	"fmt"
	"math/big"

	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/layer1/ethereum"
	"github.com/MadBase/MadNet/layer1/executor/constants"
	"github.com/MadBase/MadNet/layer1/executor/tasks"
	"github.com/MadBase/MadNet/layer1/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/layer1/transaction"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// DisputeGPKjTask contains required state for performing a group accusation
type DisputeGPKjTask struct {
	*tasks.BaseTask
}

// asserting that DisputeGPKjTask struct implements interface tasks.Task
var _ tasks.Task = &DisputeGPKjTask{}

// NewDisputeGPKjTask creates a background task that attempts perform a group accusation if necessary
func NewDisputeGPKjTask(start uint64, end uint64) *DisputeGPKjTask {
	return &DisputeGPKjTask{
		BaseTask: tasks.NewBaseTask(constants.DisputeGPKjTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the DisputeGPKjTask.
// Here, we determine if anyone submitted an invalid gpkj.
func (t *DisputeGPKjTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debug("preparing task")

	dkgState := &state.DkgState{}
	err := t.GetDB().Update(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		if err != nil {
			return err
		}

		if dkgState.Phase != state.DisputeGPKJSubmission && dkgState.Phase != state.GPKJSubmission {
			return fmt.Errorf("it's not DisputeGPKJSubmission or GPKJSubmission phase")
		}

		var (
			groupPublicKeys  [][4]*big.Int
			groupCommitments [][][2]*big.Int
		)

		var participantList = dkgState.GetSortedParticipants()

		for _, participant := range participantList {
			// Build array
			groupPublicKeys = append(groupPublicKeys, participant.GPKj)
			groupCommitments = append(groupCommitments, participant.Commitments)
		}

		honest, dishonest, missing, err := state.CategorizeGroupSigners(groupPublicKeys, participantList, groupCommitments)
		if err != nil {
			return fmt.Errorf("failed to determine honest vs dishonest validators: %v", err)
		}

		inverse, err := state.InverseArrayForUserCount(dkgState.NumberOfValidators)
		if err != nil {
			return fmt.Errorf("failed to calculate inversion: %v", err)
		}

		logger.Debugf("   Honest indices: %v", honest.ExtractIndices())
		logger.Debugf("Dishonest indices: %v", dishonest.ExtractIndices())
		logger.Debugf("  Missing indices: %v", missing.ExtractIndices())

		dkgState.DishonestValidators = dishonest
		dkgState.HonestValidators = honest
		dkgState.Inverse = inverse

		err = dkgState.PersistState(txn)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// all errors are non recoverable
		return tasks.NewTaskErr(fmt.Sprintf(constants.ErrorDuringPreparation, err), false)
	}

	return nil
}

// Execute executes the task business logic
func (t *DisputeGPKjTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
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

	// Perform group accusation
	logger.Debugf("   Honest indices: %v", dkgState.HonestValidators.ExtractIndices())
	logger.Debugf("Dishonest indices: %v", dkgState.DishonestValidators.ExtractIndices())

	var groupEncryptedSharesHash [][32]byte
	var groupCommitments [][][2]*big.Int
	var validatorAddresses []common.Address
	var participantList = dkgState.GetSortedParticipants()

	for _, participant := range participantList {
		// Get group encrypted shares
		es := participant.EncryptedShares
		encryptedSharesBin, err := bn256.MarshalBigIntSlice(es)
		if err != nil {
			return nil, tasks.NewTaskErr(fmt.Sprintf("group accusation failed: %v", err), true)
		}
		hashSlice := crypto.Hasher(encryptedSharesBin)
		var hashSlice32 [32]byte
		copy(hashSlice32[:], hashSlice)
		groupEncryptedSharesHash = append(groupEncryptedSharesHash, hashSlice32)
		// Get group commitments
		com := participant.Commitments
		groupCommitments = append(groupCommitments, com)
		validatorAddresses = append(validatorAddresses, participant.Address)
	}

	client := t.GetClient()
	callOpts, err := client.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(constants.FailedGettingCallOpts, err), true)
	}

	// Setup
	txnOpts, err := client.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(constants.FailedGettingTxnOpts, err), true)
	}

	txns := make([]*types.Transaction, 0)
	// Loop through dishonest participants and perform accusation
	for _, dishonestParticipant := range dkgState.DishonestValidators {

		isValidator, err := ethereum.GetContracts().ValidatorPool().IsValidator(callOpts, dishonestParticipant.Address)
		if err != nil {
			return nil, tasks.NewTaskErr(fmt.Sprintf(constants.FailedGettingIsValidator, err), true)
		}

		if !isValidator {
			continue
		}

		logger.Warnf("accusing participant: %v of distributing bad dpkj", dishonestParticipant.Address.Hex())
		txn, err := ethereum.GetContracts().Ethdkg().AccuseParticipantSubmittedBadGPKJ(txnOpts, validatorAddresses, groupEncryptedSharesHash, groupCommitments, dishonestParticipant.Address)
		if err != nil {
			return nil, tasks.NewTaskErr(fmt.Sprintf("group accusation failed: %v", err), true)
		}
		txns = append(txns, txn)
	}
	//todo: fix this, split this task in multiple tasks
	return txns[0], nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *DisputeGPKjTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
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

	if dkgState.Phase != state.DisputeGPKJSubmission {
		logger.Debug("phase %v different from DisputeGPKJSubmission", dkgState.Phase)
		return false, nil
	}

	client := t.GetClient()

	callOpts, err := client.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(constants.FailedGettingCallOpts, err), true)
	}
	badParticipants, err := ethereum.GetContracts().Ethdkg().GetBadParticipants(callOpts)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf("could not get BadParticipants: %v", err), true)
	}

	if len(dkgState.DishonestValidators) == int(badParticipants.Int64()) {
		logger.WithFields(logrus.Fields{
			"state.BadShares":     len(dkgState.BadShares),
			"eth.badParticipants": badParticipants,
		}).Debug("all bad participants already accused")
		return false, nil
	}

	return true, nil
}
