package dkgtasks

import (
	"context"
	"fmt"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// DisputeGPKjTask contains required state for performing a group accusation
type DisputeGPKjTask struct {
	*tasks.Task
}

// asserting that DisputeGPKjTask struct implements interface interfaces.Task
var _ interfaces.ITask = &DisputeGPKjTask{}

// NewDisputeGPKjTask creates a background task that attempts perform a group accusation if necessary
func NewDisputeGPKjTask(state *objects.DkgState, start uint64, end uint64) *DisputeGPKjTask {
	return &DisputeGPKjTask{
		Task: tasks.NewTask(state, start, end),
	}
}

// Initialize prepares for work to be done in the GPKjDispute phase.
// Here, we determine if anyone submitted an invalid gpkj.
func (t *DisputeGPKjTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("GPKJDisputeTask Initialize()...")

	taskState, ok := t.State.(*objects.DkgState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	if taskState.Phase != objects.DisputeGPKJSubmission && taskState.Phase != objects.GPKJSubmission {
		return fmt.Errorf("%w because it's not DisputeGPKJSubmission phase", objects.ErrCanNotContinue)
	}

	var (
		groupPublicKeys  [][4]*big.Int
		groupCommitments [][][2]*big.Int
	)

	var participantList = taskState.GetSortedParticipants()

	for _, participant := range participantList {
		// Build array
		groupPublicKeys = append(groupPublicKeys, participant.GPKj)
		groupCommitments = append(groupCommitments, participant.Commitments)
	}

	honest, dishonest, missing, err := math.CategorizeGroupSigners(groupPublicKeys, participantList, groupCommitments)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "Failed to determine honest vs dishonest validators: %v", err)
	}

	inverse, err := math.InverseArrayForUserCount(taskState.NumberOfValidators)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "Failed to calculate inversion: %v", err)
	}

	logger.Debugf("   Honest indices: %v", honest.ExtractIndices())
	logger.Debugf("Dishonest indices: %v", dishonest.ExtractIndices())
	logger.Debugf("  Missing indices: %v", missing.ExtractIndices())

	taskState.DishonestValidators = dishonest
	taskState.HonestValidators = honest
	taskState.Inverse = inverse

	return nil
}

// DoWork is the first attempt at submitting an invalid gpkj accusation
func (t *DisputeGPKjTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at submitting an invalid gpkj accusation
func (t *DisputeGPKjTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *DisputeGPKjTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("GPKJDisputeTask doTask()")

	taskState, ok := t.State.(*objects.DkgState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	// Perform group accusation
	logger.Infof("   Honest indices: %v", taskState.HonestValidators.ExtractIndices())
	logger.Infof("Dishonest indices: %v", taskState.DishonestValidators.ExtractIndices())

	var groupEncryptedSharesHash [][32]byte
	var groupCommitments [][][2]*big.Int
	var validatorAddresses []common.Address
	var participantList = taskState.GetSortedParticipants()

	for _, participant := range participantList {
		// Get group encrypted shares
		es := participant.EncryptedShares
		encryptedSharesBin, err := bn256.MarshalBigIntSlice(es)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "group accusation failed: %v", err)
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

	callOpts, err := eth.GetCallOpts(ctx, taskState.Account)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "getting call opts failed: %v", err)
	}

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, taskState.Account)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// If the TxOpts exists, meaning the Tx replacement timeout was reached,
	// we increase the Gas to have priority for the next blocks
	if t.TxOpts != nil && t.TxOpts.Nonce != nil {
		logger.Info("txnOpts Replaced")
		txnOpts.Nonce = t.TxOpts.Nonce
		txnOpts.GasFeeCap = t.TxOpts.GasFeeCap
		txnOpts.GasTipCap = t.TxOpts.GasTipCap
	}

	// Loop through dishonest participants and perform accusation
	for _, dishonestParticipant := range taskState.DishonestValidators {

		isValidator, err := eth.Contracts().ValidatorPool().IsValidator(callOpts, dishonestParticipant.Address)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "getting isValidator failed: %v", err)
		}

		if !isValidator {
			continue
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantSubmittedBadGPKJ(txnOpts, validatorAddresses, groupEncryptedSharesHash, groupCommitments, dishonestParticipant.Address)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "group accusation failed: %v", err)
		}
		t.TxOpts.TxHashes = append(t.TxOpts.TxHashes, txn.Hash())
		t.TxOpts.GasFeeCap = txn.GasFeeCap()
		t.TxOpts.GasTipCap = txn.GasTipCap()
		t.TxOpts.Nonce = big.NewInt(int64(txn.Nonce()))

		logger.WithFields(logrus.Fields{
			"GasFeeCap": t.TxOpts.GasFeeCap,
			"GasTipCap": t.TxOpts.GasTipCap,
			"Nonce":     t.TxOpts.Nonce,
		}).Info("bad gpkj dispute fees")

		// Queue transaction
		eth.TransactionWatcher().SubscribeTransaction(ctx, txn)
	}

	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// if the DKG process is in the right phase and blocks
// range and there still someone to accuse, the retry
// is executed
func (t *DisputeGPKjTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("GPKJDisputeTask ShouldRetry()")

	taskState, ok := t.State.(*objects.DkgState)
	if !ok {
		logger.Error("Invalid convertion of taskState object")
		return false
	}

	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	if taskState.Phase != objects.DisputeGPKJSubmission {
		return false
	}

	callOpts, err := eth.GetCallOpts(ctx, taskState.Account)
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
		"state.BadShares":     len(taskState.BadShares),
		"eth.badParticipants": badParticipants,
	}).Debug("DisputeGPKjTask ShouldRetry2()")

	return len(taskState.DishonestValidators) != int(badParticipants.Int64())
}

// DoDone creates a log entry saying task is complete
func (t *DisputeGPKjTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("GPKJDisputeTask done")
}

func (t *DisputeGPKjTask) GetExecutionData() interfaces.ITaskExecutionData {
	return t.Task
}
