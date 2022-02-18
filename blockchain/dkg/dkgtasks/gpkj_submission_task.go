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

// GPKjSubmissionTask contains required state for gpk submission
type GPKjSubmissionTask struct {
	Start        uint64
	End          uint64
	State        *objects.DkgState
	Success      bool
	adminHandler interfaces.AdminHandler
}

// asserting that GPKjSubmissionTask struct implements interface interfaces.Task
var _ interfaces.Task = &GPKjSubmissionTask{}

// NewGPKjSubmissionTask creates a background task that attempts to submit the gpkj in ETHDKG
func NewGPKjSubmissionTask(state *objects.DkgState, start uint64, end uint64, adminHandler interfaces.AdminHandler) *GPKjSubmissionTask {
	return &GPKjSubmissionTask{
		Start:        start,
		End:          end,
		State:        state,
		Success:      false,
		adminHandler: adminHandler,
	}
}

// Initialize prepares for work to be done in GPKjSubmission phase.
// Here, we construct our gpkj and associated signature.
// We will submit them in DoWork.
func (t *GPKjSubmissionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("GPKSubmissionTask Initialize()...")

	// Collecting all the participants encrypted shares to be used for the GPKj
	var participantsList = t.State.GetSortedParticipants()
	encryptedShares := make([][]*big.Int, 0, t.State.NumberOfValidators)
	for _, participant := range participantsList {
		logger.Debugf("Collecting encrypted shares... Participant %v %v", participant.Index, participant.Address.Hex())
		encryptedShares = append(encryptedShares, participant.EncryptedShares)
	}

	// Generate the GPKj
	groupPrivateKey, groupPublicKey, err := math.GenerateGroupKeys(
		t.State.TransportPrivateKey, t.State.PrivateCoefficients,
		encryptedShares, t.State.Index, participantsList)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"t.State.Index": t.State.Index,
		}).Errorf("Could not generate group keys: %v", err)
		return dkg.LogReturnErrorf(logger, "Could not generate group keys: %v", err)
	}

	t.State.GroupPrivateKey = groupPrivateKey
	t.State.Participants[t.State.Account.Address].GPKj = groupPublicKey

	// Pass private key on to consensus
	logger.Infof("Adding private bn256eth key... using %p", t.adminHandler)
	err = t.adminHandler.AddPrivateKey(groupPrivateKey.Bytes(), constants.CurveBN256Eth)
	if err != nil {
		return fmt.Errorf("%w because error adding private key: %v", objects.ErrCanNotContinue, err)
	}

	return nil
}

// DoWork is the first attempt at submitting gpkj and signature.
func (t *GPKjSubmissionTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at submitting gpkj and signature.
func (t *GPKjSubmissionTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *GPKjSubmissionTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Infof("GPKSubmissionTask doTask(): %v", t.State.Account.Address)

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// Do it
	txn, err := eth.Contracts().Ethdkg().SubmitGPKJ(txnOpts, t.State.Participants[t.State.Account.Address].GPKj)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "submitting master public key failed: %v", err)
	}

	//TODO: add retry logic, add timeout

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
func (t *GPKjSubmissionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	t.State.Lock()
	defer t.State.Unlock()
	logger.Info("GPKSubmissionTask ShouldRetry()")

	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	if t.State.Phase != objects.GPKJSubmission {
		return false
	}

	//Check if my GPKj is submitted, if not should retry
	me := t.State.Account
	callOpts := eth.GetCallOpts(ctx, me)
	participantState, err := eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, me.Address)
	if err == nil && participantState.Gpkj[0].Cmp(t.State.Participants[me.Address].GPKj[0]) == 0 &&
		participantState.Gpkj[1].Cmp(t.State.Participants[me.Address].GPKj[1]) == 0 &&
		participantState.Gpkj[2].Cmp(t.State.Participants[me.Address].GPKj[2]) == 0 &&
		participantState.Gpkj[3].Cmp(t.State.Participants[me.Address].GPKj[3]) == 0 {
		return false
	}

	return true
}

// DoDone creates a log entry saying task is complete
func (t *GPKjSubmissionTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("GPKSubmissionTask done")
}

// SetAdminHandler sets the task adminHandler
func (t *GPKjSubmissionTask) SetAdminHandler(adminHandler interfaces.AdminHandler) {
	t.adminHandler = adminHandler
}
