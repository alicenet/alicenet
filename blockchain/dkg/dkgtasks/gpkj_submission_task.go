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

// Initialize prepares for work to be done in GPKSubmission phase.
// Here, we construct our gpkj and associated signature.
// We will submit them in DoWork.
func (t *GPKjSubmissionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	dkgState, validState := state.(*objects.DkgState)
	if !validState {
		panic(fmt.Errorf("%w invalid state type", objects.ErrCanNotContinue))
	}

	t.State = dkgState

	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("StateLocation", fmt.Sprintf("%p", t.State)).Info("GPKSubmissionTask Initialize()...")

	var participantsList = t.State.GetSortedParticipants()
	encryptedShares := make([][]*big.Int, 0, t.State.NumberOfValidators)
	for _, participant := range participantsList {
		logger.Debugf("Collecting encrypted shares... Participant %v %v", participant.Index, participant.Address.Hex())
		encryptedShares = append(encryptedShares, participant.EncryptedShares)
	}

	groupPrivateKey, groupPublicKey, err := math.GenerateGroupKeys(
		t.State.TransportPrivateKey, t.State.PrivateCoefficients,
		encryptedShares, t.State.Index, participantsList)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"t.State.Index": t.State.Index,
		}).Errorf("Could not generate group keys: %v", err)
		return dkg.LogReturnErrorf(logger, "Could not generate group keys: %v", err)
	}

	// todo: delete this
	// inject bad GPKj data
	// if t.State.Index == 5 {
	// 	// mess up with group provate key (gskj)
	// 	gskjBad := new(big.Int).Add(groupPrivateKey, big.NewInt(1))
	// 	// here's the group public key
	// 	gpkj := new(cloudflare.G2).ScalarBaseMult(gskjBad)
	// 	gpkjBig, err := bn256.G2ToBigIntArray(gpkj)
	// 	if err != nil {
	// 		return dkg.LogReturnErrorf(logger, "error generating invalid gpkj: %v", err)
	// 	}

	// 	groupPrivateKey = gskjBad
	// 	groupPublicKey = gpkjBig
	// }

	t.State.GroupPrivateKey = groupPrivateKey
	t.State.Participants[t.State.Account.Address].GPKj = groupPublicKey

	// Pass private key on to consensus
	logger.Infof("Adding private bn256eth key... using %p", t.adminHandler)
	err = t.adminHandler.AddPrivateKey(groupPrivateKey.Bytes(), constants.CurveBN256Eth)
	if err != nil {
		return fmt.Errorf("%w because error adding private key: %v", objects.ErrCanNotContinue, err) // TODO this is seriously bad, any better actions possible?
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

	// debug case
	// if t.State.Account.Address.String() == "0x565128Dd9Fe84629E1d3b3F2B2Fee43b801d5adE" {
	// 	// this is val5, just drop
	// 	// logger.Warn("GPKSubmissionTask won't doTask()")
	// 	// return nil
	// 	return dkg.LogReturnErrorf(logger, "GPKSubmissionTask won't doTask()")
	// }

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
func (t *GPKjSubmissionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	t.State.Lock()
	defer t.State.Unlock()
	logger.Info("GPKSubmissionTask ShouldRetry()")

	currentBlock, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return true
	}
	// logger = logger.WithField("CurrentHeight", currentBlock)

	if t.State.Phase == objects.GPKJSubmission &&
		t.Start <= currentBlock &&
		currentBlock < t.End {
		return true
	}

	// var shouldRetry bool = GeneralTaskShouldRetry(ctx, t.State.Account, logger, eth, t.State.TransportPublicKey, t.Start, t.End)

	// logger.WithFields(logrus.Fields{
	// 	"shouldRetry": shouldRetry,
	// }).Info("GPKSubmissionTask ShouldRetry2()")

	// This wraps the retry logic for the general case
	// todo: fix this
	return false
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
