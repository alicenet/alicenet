package dkgtasks

import (
	"context"
	"fmt"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/constants"
	"github.com/sirupsen/logrus"
)

// GPKSubmissionTask contains required state for gpk submission
type GPKSubmissionTask struct {
	OriginalRegistrationEnd uint64
	State                   *objects.DkgState
	Success                 bool
	adminHandler            interfaces.AdminHandler
}

// NewGPKSubmissionTask creates a background task that attempts to register with ETHDKG
func NewGPKSubmissionTask(state *objects.DkgState, adminHandler interfaces.AdminHandler) *GPKSubmissionTask {
	return &GPKSubmissionTask{
		OriginalRegistrationEnd: state.RegistrationEnd, // If these quit being equal, this task should be abandoned
		State:                   state,
		adminHandler:            adminHandler,
	}
}

func (t *GPKSubmissionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {

	dkgState, validState := state.(*objects.DkgState)
	if !validState {
		panic(fmt.Errorf("%w invalid state type", objects.ErrCanNotContinue))
	}

	t.State = dkgState

	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("StateLocation", fmt.Sprintf("%p", t.State)).Info("Initialize()...")

	if !t.State.MPKSubmission {
		return fmt.Errorf("%w because mpk submission not successful", objects.ErrCanNotContinue)
	}

	// TODO Guard
	callOpts := eth.GetCallOpts(ctx, t.State.Account)
	initialMessage, err := eth.Contracts().Ethdkg().InitialMessage(callOpts)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "Could not retrieve initial message: %v", err)
	}

	encryptedShares := make([][]*big.Int, t.State.NumberOfValidators)
	for idx, participant := range t.State.Participants {
		logger.Debugf("Collecting encrypted shares... Participant %v %v", participant.Index, participant.Address.Hex())
		pes, present := t.State.EncryptedShares[participant.Address]
		if present && idx >= 0 && idx < t.State.NumberOfValidators {
			encryptedShares[idx] = pes
		} else {
			logger.Errorf("Encrypted share state broken for %v", idx)
		}
	}

	groupPrivateKey, groupPublicKey, groupSignature, err := math.GenerateGroupKeys(initialMessage,
		t.State.TransportPrivateKey, t.State.TransportPublicKey, t.State.PrivateCoefficients,
		encryptedShares, t.State.Index, t.State.Participants, t.State.ValidatorThreshold)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "Could not generate group keys: %v", err)
	}

	t.State.InitialMessage = initialMessage
	t.State.GroupPrivateKey = groupPrivateKey
	t.State.GroupPublicKey = groupPublicKey
	t.State.GroupSignature = groupSignature

	// Pass private key on to consensus
	logger.Infof("Adding private bn256eth key... using %p", t.adminHandler)
	err = t.adminHandler.AddPrivateKey(groupPrivateKey.Bytes(), constants.CurveBN256Eth)
	if err != nil {
		return fmt.Errorf("%w because error adding private key: %v", objects.ErrCanNotContinue, err) // TODO this is seriously bad, any better actions possible?
	}

	return nil
}

// DoWork is the first attempt at registering with ethdkg
func (t *GPKSubmissionTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at registering with ethdkg
func (t *GPKSubmissionTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *GPKSubmissionTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// Do it
	txn, err := eth.Contracts().Ethdkg().SubmitGPKj(txnOpts, t.State.GroupPublicKey, t.State.GroupSignature)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "submitting master public key failed: %v", err)
	}

	// Waiting for receipt
	receipt, err := eth.Queue().QueueAndWait(ctx, txn)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "waiting for receipt failed: %v", err)
	}
	if receipt == nil {
		return dkg.LogReturnErrorf(logger, "missing registration receipt")
	}
	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
// -- the registration open hasn't moved, i.e. ETHDKG has not restarted
func (t *GPKSubmissionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	t.State.Lock()
	defer t.State.Unlock()

	state := t.State

	// This wraps the retry logic for the general case
	return GeneralTaskShouldRetry(ctx, state.Account, logger, eth,
		state.TransportPublicKey, t.OriginalRegistrationEnd, state.GPKJSubmissionEnd)
}

// DoDone creates a log entry saying task is complete
func (t *GPKSubmissionTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("done")

	t.State.GPKJSubmission = t.Success
}

func (t *GPKSubmissionTask) SetAdminHandler(adminHandler interfaces.AdminHandler) {
	t.adminHandler = adminHandler
}
