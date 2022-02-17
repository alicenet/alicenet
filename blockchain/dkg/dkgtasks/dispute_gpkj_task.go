package dkgtasks

import (
	"context"
	"fmt"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// DisputeGPKjTask contains required state for performing a group accusation
type DisputeGPKjTask struct {
	Start   uint64
	End     uint64
	State   *objects.DkgState
	Success bool
}

// asserting that DisputeGPKjTask struct implements interface interfaces.Task
var _ interfaces.Task = &DisputeGPKjTask{}

// NewDisputeGPKjTask creates a background task that attempts perform a group accusation if necessary
func NewDisputeGPKjTask(state *objects.DkgState, start uint64, end uint64) *DisputeGPKjTask {
	return &DisputeGPKjTask{
		Start:   start,
		End:     end,
		State:   state,
		Success: false,
	}
}

// Initialize prepares for work to be done in the GPKjDispute phase.
// Here, we determine if anyone submitted an invalid gpkj.
func (t *DisputeGPKjTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("StateLocation", fmt.Sprintf("%p", t.State)).Info("GPKJDisputeTask Initialize()...")

	if t.State.Phase != objects.DisputeGPKJSubmission && t.State.Phase != objects.GPKJSubmission {
		return fmt.Errorf("%w because it's not DisputeGPKJSubmission phase", objects.ErrCanNotContinue)
	}

	var (
		groupPublicKeys  [][4]*big.Int
		groupCommitments [][][2]*big.Int
	)

	var participantList = t.State.GetSortedParticipants()

	for _, participant := range participantList {
		// Build array
		groupPublicKeys = append(groupPublicKeys, participant.GPKj)
		groupCommitments = append(groupCommitments, participant.Commitments)
	}

	honest, dishonest, missing, err := math.CategorizeGroupSigners(groupPublicKeys, participantList, groupCommitments)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "Failed to determine honest vs dishonest validators: %v", err)
	}

	inverse, err := math.InverseArrayForUserCount(t.State.NumberOfValidators)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "Failed to calculate inversion: %v", err)
	}

	logger.Debugf("   Honest indices: %v", honest.ExtractIndices())
	logger.Debugf("Dishonest indices: %v", dishonest.ExtractIndices())
	logger.Debugf("  Missing indices: %v", missing.ExtractIndices())

	t.State.DishonestValidators = dishonest
	t.State.HonestValidators = honest
	t.State.Inverse = inverse

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

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// Perform group accusation
	logger.Infof("   Honest indices: %v", t.State.HonestValidators.ExtractIndices())
	logger.Infof("Dishonest indices: %v", t.State.DishonestValidators.ExtractIndices())

	var groupEncryptedSharesHash [][32]byte
	var groupCommitments [][][2]*big.Int
	var validatorAddresses []common.Address
	var participantList = t.State.GetSortedParticipants()

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

	callOpts := eth.GetCallOpts(ctx, t.State.Account)
	// Loop through dishonest participants and perform accusation
	for _, dishonestParticipant := range t.State.DishonestValidators {

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

		//TODO: add retry logic, add timeout

		// Waiting for receipt
		receipt, err := eth.Queue().QueueAndWait(ctx, txn)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "waiting for receipt failed: %v", err)
		}
		if receipt == nil {
			return dkg.LogReturnErrorf(logger, "missing receipt")
		}

		// Check receipt to confirm we were successful
		if receipt.Status != uint64(1) {
			return dkg.LogReturnErrorf(logger, "bad gpkj error (%v) indicates failure: %v", receipt.Status, receipt.Logs)
		}
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

	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	if t.State.Phase != objects.DisputeGPKJSubmission {
		return false
	}

	callOpts := eth.GetCallOpts(ctx, t.State.Account)
	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	if err != nil {
		logger.Error("could not get BadParticipants")
		return true
	}

	logger.WithFields(logrus.Fields{
		"state.BadShares":     len(t.State.BadShares),
		"eth.badParticipants": badParticipants,
	}).Debug("DisputeGPKjTask ShouldRetry2()")

	return len(t.State.DishonestValidators) != int(badParticipants.Int64())
}

// DoDone creates a log entry saying task is complete
func (t *DisputeGPKjTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("GPKJDisputeTask done")
}
