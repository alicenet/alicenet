package dkg

import (
	"fmt"
	"math/big"

	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/MadBase/MadNet/blockchain/executor/constants"
	executorInterfaces "github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
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
var _ executorInterfaces.ITask = &DisputeGPKjTask{}

// NewDisputeGPKjTask creates a background task that attempts perform a group accusation if necessary
func NewDisputeGPKjTask(start uint64, end uint64) *DisputeGPKjTask {
	return &DisputeGPKjTask{
		Task: objects.NewTask(constants.DisputeGPKjTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the DisputeGPKjTask.
// Here, we determine if anyone submitted an invalid gpkj.
func (t *DisputeGPKjTask) Prepare() *executorInterfaces.TaskErr {
	logger := t.GetLogger()
	logger.Info("DisputeGPKjTask Prepare()...")

	dkgState := &state.DkgState{}
	err := t.GetDB().Update(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		if err != nil {
			return err
		}

		if dkgState.Phase != state.DisputeGPKJSubmission && dkgState.Phase != state.GPKJSubmission {
			return fmt.Errorf("%w because it's not DisputeGPKJSubmission phase", objects.ErrCanNotContinue)
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
		return utils.LogReturnErrorf(logger, "DisputeGPKjTask.Prepare(): error during the preparation: %v", err)
	}

	return nil
}

// Execute executes the task business logic
func (t *DisputeGPKjTask) Execute() ([]*types.Transaction, *executorInterfaces.TaskErr) {
	logger := t.GetLogger()
	logger.Info("DisputeGPKjTask Execute()")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, utils.LogReturnErrorf(logger, "DisputeGPKjTask.Execute(): error loading dkgState: %v", err)
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
			return nil, utils.LogReturnErrorf(logger, "group accusation failed: %v", err)
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
		return nil, utils.LogReturnErrorf(logger, "getting call opts failed: %v", err)
	}

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, utils.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	txns := make([]*types.Transaction, 0)
	// Loop through dishonest participants and perform accusation
	for _, dishonestParticipant := range dkgState.DishonestValidators {

		isValidator, err := eth.Contracts().ValidatorPool().IsValidator(callOpts, dishonestParticipant.Address)
		if err != nil {
			return nil, utils.LogReturnErrorf(logger, "getting isValidator failed: %v", err)
		}

		if !isValidator {
			continue
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantSubmittedBadGPKJ(txnOpts, validatorAddresses, groupEncryptedSharesHash, groupCommitments, dishonestParticipant.Address)
		if err != nil {
			return nil, utils.LogReturnErrorf(logger, "group accusation failed: %v", err)
		}
		txns = append(txns, txn)
	}

	return txns, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *DisputeGPKjTask) ShouldExecute() (bool, *executorInterfaces.TaskErr) {
	logger := t.GetLogger()
	logger.Info("DisputeGPKjTask ShouldExecute()")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		logger.Errorf("could not get dkgState with error %v", err)
		return true
	}

	eth := t.GetClient()
	ctx := t.GetCtx()
	if dkgState.Phase != state.DisputeGPKJSubmission {
		return false
	}

	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		logger.Error("could not get call opts disputeDPKj")
		return true
	}
	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	if err != nil {
		logger.Error("could not get BadParticipants")
		return true
	}

	logger.WithFields(logrus.Fields{
		"state.BadShares":     len(dkgState.BadShares),
		"eth.badParticipants": badParticipants,
	}).Debug("DisputeGPKjTask ShouldExecute()")

	return len(dkgState.DishonestValidators) != int(badParticipants.Int64())
}
