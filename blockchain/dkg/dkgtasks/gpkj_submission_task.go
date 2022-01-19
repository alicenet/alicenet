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
	State        *objects.DkgState
	Success      bool
	adminHandler interfaces.AdminHandler
}

// NewGPKSubmissionTask creates a background task that attempts to submit the gpkj in ETHDKG
func NewGPKSubmissionTask(state *objects.DkgState, adminHandler interfaces.AdminHandler) *GPKSubmissionTask {
	return &GPKSubmissionTask{
		State:        state,
		adminHandler: adminHandler,
	}
}

// Initialize prepares for work to be done in GPKSubmission phase.
// Here, we construct our gpkj and associated signature.
// We will submit them in DoWork.
func (t *GPKSubmissionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	dkgState, validState := state.(*objects.DkgState)
	if !validState {
		panic(fmt.Errorf("%w invalid state type", objects.ErrCanNotContinue))
	}

	t.State = dkgState

	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("StateLocation", fmt.Sprintf("%p", t.State)).Info("GPKSubmissionTask Initialize()...")

	// todo: delete this bc State.MPKSubmission should not exist
	// if !t.State.MPKSubmission {
	// 	return fmt.Errorf("%w because mpk submission not successful", objects.ErrCanNotContinue)
	// }

	encryptedShares := make([][]*big.Int, 0, t.State.NumberOfValidators)
	for idx, participant := range t.State.Participants {
		logger.Debugf("Collecting encrypted shares... Participant %v %v", participant.Index, participant.Address.Hex())
		pes, present := t.State.EncryptedShares[participant.Address]
		if present && idx >= 0 && idx < int(t.State.NumberOfValidators) {
			encryptedShares = append(encryptedShares, pes)
		} else {
			logger.Errorf("Encrypted share state broken for %v", idx)
		}
	}

	// todo: get my index (done)

	groupPrivateKey, groupPublicKey, err := math.GenerateGroupKeys(
		t.State.TransportPrivateKey, t.State.PrivateCoefficients,
		encryptedShares, t.State.Index, t.State.Participants)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"t.State.Index": t.State.Index,
		}).Errorf("Could not generate group keys: %v", err)
		return dkg.LogReturnErrorf(logger, "Could not generate group keys: %v", err)
	}

	t.State.GroupPrivateKey = groupPrivateKey
	t.State.GroupPublicKey = groupPublicKey

	// Pass private key on to consensus
	logger.Infof("Adding private bn256eth key... using %p", t.adminHandler)
	err = t.adminHandler.AddPrivateKey(groupPrivateKey.Bytes(), constants.CurveBN256Eth)
	if err != nil {
		return fmt.Errorf("%w because error adding private key: %v", objects.ErrCanNotContinue, err) // TODO this is seriously bad, any better actions possible?
	}

	return nil
}

// DoWork is the first attempt at submitting gpkj and signature.
func (t *GPKSubmissionTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at submitting gpkj and signature.
func (t *GPKSubmissionTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *GPKSubmissionTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("GPKSubmissionTask doTask()")

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// Do it
	txn, err := eth.Contracts().Ethdkg().SubmitGPKJ(txnOpts, t.State.GroupPublicKey)
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

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		return dkg.LogReturnErrorf(logger, "submit gpkj status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
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
	logger.Info("GPKSubmissionTask ShouldRetry()")

	var phaseStart = t.State.PhaseStart
	var phaseEnd = phaseStart + t.State.PhaseLength

	currentBlock, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return true
	}
	logger = logger.WithField("CurrentHeight", currentBlock)

	if t.State.Phase == objects.GPKJSubmission &&
		phaseStart <= currentBlock &&
		currentBlock < phaseEnd {
		return true
	}

	var shouldRetry bool = GeneralTaskShouldRetry(ctx, t.State.Account, logger, eth, t.State.TransportPublicKey, phaseStart, phaseEnd)

	logger.WithFields(logrus.Fields{
		"shouldRetry": shouldRetry,
	}).Info("GPKSubmissionTask ShouldRetry2()")

	// This wraps the retry logic for the general case
	// todo: fix this
	return shouldRetry
}

// DoDone creates a log entry saying task is complete
func (t *GPKSubmissionTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("GPKSubmissionTask done")
}

// SetAdminHandler sets the task adminHandler
func (t *GPKSubmissionTask) SetAdminHandler(adminHandler interfaces.AdminHandler) {
	t.adminHandler = adminHandler
}
