package dkg

import (
	"fmt"
	"math/big"

	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// DisputeGPKjTask contains required state for performing a group accusation
type DisputeGPKjTask struct {
	*objects.Task
}

// asserting that DisputeGPKjTask struct implements interface interfaces.Task
var _ interfaces.ITask = &DisputeGPKjTask{}

// NewDisputeGPKjTask creates a background task that attempts perform a group accusation if necessary
func NewDisputeGPKjTask(start uint64, end uint64) *DisputeGPKjTask {
	return &DisputeGPKjTask{
		Task: objects.NewTask(constants.DisputeGPKjTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the DisputeGPKjTask.
// Here, we determine if anyone submitted an invalid gpkj.
func (t *DisputeGPKjTask) Prepare() *interfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Tracef("preparing task")

	dkgState := &state.DkgState{}
	var isRecoverable bool
	err := t.GetDB().Update(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		if err != nil {
			isRecoverable = false
			return err
		}

		if dkgState.Phase != state.DisputeGPKJSubmission && dkgState.Phase != state.GPKJSubmission {
			isRecoverable = false
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
			isRecoverable = true
			return fmt.Errorf("failed to determine honest vs dishonest validators: %v", err)
		}

		inverse, err := state.InverseArrayForUserCount(dkgState.NumberOfValidators)
		if err != nil {
			isRecoverable = true
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
			isRecoverable = false
			return err
		}

		return nil
	})

	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf(constants.ErrorDuringPreparation, err), isRecoverable)
	}

	return nil
}

// Execute executes the task business logic
func (t *DisputeGPKjTask) Execute() ([]*types.Transaction, *interfaces.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Trace("initiate execution")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	// Perform group accusation
	logger.Infof("   Honest indices: %v", dkgState.HonestValidators.ExtractIndices())
	logger.Infof("Dishonest indices: %v", dkgState.DishonestValidators.ExtractIndices())

	var groupEncryptedSharesHash [][32]byte
	var groupCommitments [][][2]*big.Int
	var validatorAddresses []common.Address
	var participantList = dkgState.GetSortedParticipants()

	for _, participant := range participantList {
		// Get group encrypted shares
		es := participant.EncryptedShares
		encryptedSharesBin, err := bn256.MarshalBigIntSlice(es)
		if err != nil {
			return nil, interfaces.NewTaskErr(fmt.Sprintf("group accusation failed: %v", err), true)
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

	eth := t.GetClient()
	ctx := t.GetCtx()
	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf(constants.FailedGettingCallOpts, err), true)
	}

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf(constants.FailedGettingTxnOpts, err), true)
	}

	txns := make([]*types.Transaction, 0)
	// Loop through dishonest participants and perform accusation
	for _, dishonestParticipant := range dkgState.DishonestValidators {

		isValidator, err := eth.Contracts().ValidatorPool().IsValidator(callOpts, dishonestParticipant.Address)
		if err != nil {
			return nil, interfaces.NewTaskErr(fmt.Sprintf(constants.FailedGettingIsValidator, err), true)
		}

		if !isValidator {
			continue
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantSubmittedBadGPKJ(txnOpts, validatorAddresses, groupEncryptedSharesHash, groupCommitments, dishonestParticipant.Address)
		if err != nil {
			return nil, interfaces.NewTaskErr(fmt.Sprintf("group accusation failed: %v", err), true)
		}
		txns = append(txns, txn)
	}

	return txns, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *DisputeGPKjTask) ShouldExecute() *interfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Trace("should execute task")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	eth := t.GetClient()
	ctx := t.GetCtx()
	if dkgState.Phase != state.DisputeGPKJSubmission {
		return interfaces.NewTaskErr(fmt.Sprintf("phase %v different from DisputeGPKJSubmission", dkgState.Phase), false)
	}

	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf(constants.FailedGettingCallOpts, err), true)
	}
	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf("could not get BadParticipants: %v", err), true)
	}

	logger.WithFields(logrus.Fields{
		"state.BadShares":     len(dkgState.BadShares),
		"eth.badParticipants": badParticipants,
	}).Debug("DisputeGPKjTask ShouldExecute()")

	if len(dkgState.DishonestValidators) == int(badParticipants.Int64()) {
		return interfaces.NewTaskErr(fmt.Sprintf("all bad participants already accused"), false)
	}

	return nil
}
