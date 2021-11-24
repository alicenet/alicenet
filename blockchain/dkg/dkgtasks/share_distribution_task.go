package dkgtasks

import (
	"context"
	"errors"
	"fmt"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

// ShareDistributionTask stores the data required safely distribute shares
type ShareDistributionTask struct {
	OriginalRegistrationEnd uint64
	State                   *objects.DkgState
	Success                 bool
}

// NewShareDistributionTask creates a new task
func NewShareDistributionTask(state *objects.DkgState) *ShareDistributionTask {
	return &ShareDistributionTask{
		OriginalRegistrationEnd: state.RegistrationEnd, // If these quit being equal, this task should be abandoned
		State:                   state,
	}
}

// Initialize begins the setup phase for ShareDistribution.
// We construct our commitments and encrypted shares before
// submitting them to the associated smart contract.
func (t *ShareDistributionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	dkgState, validState := state.(*objects.DkgState)
	if !validState {
		panic(fmt.Errorf("%w invalid state type", objects.ErrCanNotContinue))
	}

	t.State = dkgState

	t.State.Lock()
	defer t.State.Unlock()

	me := t.State.Account
	callOpts := eth.GetCallOpts(ctx, me)

	if !t.State.Registration {
		return fmt.Errorf("%w because registration not successful", objects.ErrCanNotContinue)
	}

	// Retrieve information about other participants from smart contracts
	participants, index, err := dkg.RetrieveParticipants(callOpts, eth)
	if err != nil {
		logger.Errorf("Failed to retrieve other participants: %v", err)
		return err
	}

	//
	if logger.Logger.IsLevelEnabled(logrus.DebugLevel) {
		for idx, participant := range participants {
			logger.Debugf("Index:%v Participant Index:%v PublicKey:%v", idx, participant.Index, FormatPublicKey(participant.PublicKey))
		}
	}

	numParticipants := len(participants)
	threshold := math.ThresholdForUserCount(numParticipants)

	// Generate shares
	encryptedShares, privateCoefficients, commitments, err := math.GenerateShares(
		t.State.TransportPrivateKey, participants)
	if err != nil {
		logger.Errorf("Failed to generate shares: %v", err)
		return err
	}

	// Store calculated values
	t.State.Commitments[me.Address] = commitments
	t.State.EncryptedShares[me.Address] = encryptedShares
	t.State.Index = index
	t.State.NumberOfValidators = numParticipants
	t.State.Participants = participants
	t.State.PrivateCoefficients = privateCoefficients
	t.State.SecretValue = privateCoefficients[0]
	t.State.ValidatorThreshold = threshold

	return nil
}

// DoWork is the first attempt at distributing shares via ethdkg
func (t *ShareDistributionTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is subsequent attempts at distributing shares via ethdkg
func (t *ShareDistributionTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *ShareDistributionTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	// Setup
	t.State.Lock()
	defer t.State.Unlock()

	c := eth.Contracts()
	me := t.State.Account.Address
	logger.Debugf("me:%v", me.Hex())
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		logger.Errorf("getting txn opts failed: %v", err)
		return err
	}

	// Distribute shares
	logger.Infof("# shares:%d # commitments:%d", len(t.State.EncryptedShares), len(t.State.Commitments))
	txn, err := c.Ethdkg().DistributeShares(txnOpts, t.State.EncryptedShares[me], t.State.Commitments[me])
	if err != nil {
		logger.Errorf("distributing shares failed: %v", err)
		return err
	}
	// Waiting for receipt
	receipt, err := eth.Queue().QueueAndWait(ctx, txn)
	if err != nil {
		logger.Errorf("waiting for receipt failed: %v", err)
		return err
	}
	if receipt == nil {
		message := "missing distribute shares receipt"
		logger.Error(message)
		return errors.New(message)
	}

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		message := fmt.Sprintf("receipt status indicates failure: %v", receipt.Status)
		logger.Errorf(message)
		return errors.New(message)
	}
	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
func (t *ShareDistributionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	t.State.Lock()
	defer t.State.Unlock()

	// This wraps the retry logic for the general case
	generalRetry := GeneralTaskShouldRetry(ctx, t.State.Account, logger, eth,
		t.State.TransportPublicKey, t.OriginalRegistrationEnd, t.State.ShareDistributionEnd)

	// If it's generally good to retry, let's try to be more specific
	if generalRetry {
		callOpts := eth.GetCallOpts(ctx, t.State.Account)
		distributionHash, err := eth.Contracts().Ethdkg().ShareDistributionHashes(callOpts, t.State.Account.Address)
		if err != nil {
			return true
		}

		// TODO can I prove this is the correct share distribution hash?
		logger.Infof("DistributionHash: %x", distributionHash)
	}

	return false
}

// DoDone creates a log entry saying task is complete
func (t *ShareDistributionTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("done")

	t.State.ShareDistribution = t.Success
}
