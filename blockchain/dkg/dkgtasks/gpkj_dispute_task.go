package dkgtasks

import (
	"context"
	"fmt"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

// GPKJDisputeTask contains required state for performing a group accusation
type GPKJDisputeTask struct {
	OriginalRegistrationEnd uint64
	State                   *objects.DkgState
	Success                 bool
}

// NewGPKJDisputeTask creates a background task that attempts perform a group accusation if necessary
func NewGPKJDisputeTask(state *objects.DkgState) *GPKJDisputeTask {
	return &GPKJDisputeTask{
		OriginalRegistrationEnd: state.RegistrationEnd, // If these quit being equal, this task should be abandoned
		State:                   state,
	}
}

// Initialize prepares for work to be done in the GPKjDispute phase.
// Here, we determine if anyone submitted an invalid gpkj.
func (t *GPKJDisputeTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	dkgState, validState := state.(*objects.DkgState)
	if !validState {
		panic(fmt.Errorf("%w invalid state type", objects.ErrCanNotContinue))
	}

	t.State = dkgState

	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("StateLocation", fmt.Sprintf("%p", t.State)).Info("Initialize()...")

	if !t.State.GPKJSubmission {
		return fmt.Errorf("%w because gpk submission phase not successful", objects.ErrCanNotContinue)
	}

	var (
		groupPublicKeys  [][4]*big.Int
		groupSignatures  [][2]*big.Int
		groupCommitments [][][2]*big.Int
	)

	callOpts := eth.GetCallOpts(ctx, t.State.Account)
	for _, participant := range t.State.Participants {
		// Retrieve values all group keys and signatures from contract
		groupPublicKey, err := dkg.RetrieveGroupPublicKey(callOpts, eth, participant.Address)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "Failed to retrieve group public key for %v", participant.Address.Hex())
		}

		groupSignature, err := dkg.RetrieveSignature(callOpts, eth, participant.Address)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "Failed to retrieve group signature for %v", participant.Address.Hex())
		}

		// Save the values
		t.State.GroupPublicKeys[participant.Address] = groupPublicKey
		t.State.GroupSignatures[participant.Address] = groupSignature

		// Build array
		groupPublicKeys = append(groupPublicKeys, groupPublicKey)
		groupSignatures = append(groupSignatures, groupSignature)
		groupCommitments = append(groupCommitments, t.State.Commitments[participant.Address])
	}

	//
	honest, dishonest, missing, err := math.CategorizeGroupSigners(groupPublicKeys, t.State.Participants, groupCommitments)
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
func (t *GPKJDisputeTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	logger.Info("DoWork() ...")

	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at submitting an invalid gpkj accusation
func (t *GPKJDisputeTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	logger.Info("DoRetry() ...")

	return t.doTask(ctx, logger, eth)
}

func (t *GPKJDisputeTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// Perform group accusation
	logger.Infof("   Honest indices: %v", t.State.HonestValidators.ExtractIndices())
	logger.Infof("Dishonest indices: %v", t.State.DishonestValidators.ExtractIndices())

	var groupEncryptedShares [][]*big.Int
	var groupCommitments [][][2]*big.Int

	for _, participant := range t.State.Participants {
		// Get group encrypted shares
		es := t.State.EncryptedShares[participant.Address]
		groupEncryptedShares = append(groupEncryptedShares, es)
		// Get group commitments
		com := t.State.Commitments[participant.Address]
		groupCommitments = append(groupCommitments, com)
	}

	// Loop through dishonest participants and perform accusation
	for _, participant := range t.State.DishonestValidators {
		// We convert the participant index to the "participant list index";
		// that is, we convert from base 1 to base 0.
		dishonestListIdxBig := new(big.Int).Sub(big.NewInt(int64(participant.Index)), big.NewInt(1))

		txn, err := eth.Contracts().Ethdkg().GroupAccusationGPKjComp(txnOpts, groupEncryptedShares, groupCommitments, dishonestListIdxBig, participant.Address)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "group accusation failed: %v", err)
		}

		// Waiting for receipt
		receipt, err := eth.Queue().QueueAndWait(ctx, txn)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "waiting for receipt failed: %v", err)
		}
		if receipt == nil {
			return dkg.LogReturnErrorf(logger, "missing registration receipt")
		}

		// Check receipt to confirm we were successful
		if receipt.Status != uint64(1) {
			return dkg.LogReturnErrorf(logger, "registration status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
		}
	}

	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
// -- the registration open hasn't moved, i.e. ETHDKG has not restarted
func (t *GPKJDisputeTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.State.Lock()
	defer t.State.Unlock()

	state := t.State

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, state.Account, logger, eth,
		state.TransportPublicKey, t.OriginalRegistrationEnd, state.GPKJGroupAccusationEnd)
}

// DoDone creates a log entry saying task is complete
func (t *GPKJDisputeTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("done")

	t.State.GPKJGroupAccusation = t.Success
}
