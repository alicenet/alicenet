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
		groupPublicKeys [][4]*big.Int
		groupSignatures [][2]*big.Int
	)

	// dkg.RetrieveGroupPublicKey(callOpts *bind.CallOpts, eth interfaces.Ethereum, addr common.Address)

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
	}

	//
	honest, dishonest, err := math.CategorizeGroupSigners(t.State.InitialMessage, t.State.MasterPublicKey, groupPublicKeys, groupSignatures, t.State.Participants, t.State.ValidatorThreshold)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "Failed to determine honest vs dishonest validators: %v", err)
	}

	inverse, err := math.InverseArrayForUserCount(t.State.NumberOfValidators)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "Failed to calculate inversion: %v", err)
	}

	// The indices returned are into the participants array _not_ the Index of the participant in smart contract
	for idx := range honest {
		honest[idx] = t.State.Participants[idx].Index + 1
	}
	for idx := range dishonest {
		dishonest[idx] = t.State.Participants[idx].Index + 1
	}

	logger.Debugf("   Honest indices: %v", honest)
	logger.Debugf("Dishonest indices: %v", dishonest)

	t.State.DishonestValidatorsIndicies = dkg.IntsToBigInts(dishonest)
	t.State.HonestValidatorsIndicies = dkg.IntsToBigInts(honest)
	t.State.Inverse = inverse

	return nil
}

// DoWork is the first attempt at registering with ethdkg
func (t *GPKJDisputeTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	logger.Info("DoWork() ...")

	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at registering with ethdkg
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
	logger.Infof("   Honest: %v", t.State.HonestValidatorsIndicies)
	logger.Infof("Dishonest: %v", t.State.DishonestValidatorsIndicies)

	// TODO How about I just store these already with the + 1?
	honest := make([]*big.Int, len(t.State.HonestValidatorsIndicies))
	for i := range t.State.HonestValidatorsIndicies {
		// honest[i] = big.NewInt(1)
		honest[i] = big.NewInt(0)
		honest[i].Add(honest[i], t.State.HonestValidatorsIndicies[i])
	}

	txn, err := eth.Contracts().Ethdkg().GroupAccusationGPKj(txnOpts, t.State.Inverse, honest, t.State.DishonestValidatorsIndicies)
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
