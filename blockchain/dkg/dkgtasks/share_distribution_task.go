package dkgtasks

import (
	"context"
	"errors"
	"fmt"

	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

// ShareDistributionTask stores the data required safely distribute shares
type ShareDistributionTask struct {
	Start   uint64
	End     uint64
	State   *objects.DkgState
	Success bool
}

// asserting that ShareDistributionTask struct implements interface interfaces.Task
var _ interfaces.Task = &ShareDistributionTask{}

// NewShareDistributionTask creates a new task
func NewShareDistributionTask(state *objects.DkgState, start uint64, end uint64) *ShareDistributionTask {
	return &ShareDistributionTask{
		Start:   start,
		End:     end,
		State:   state,
		Success: false,
	}
}

// Initialize begins the setup phase for ShareDistribution.
// We construct our commitments and encrypted shares before
// submitting them to the associated smart contract.
func (t *ShareDistributionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	t.State.Lock()
	defer t.State.Unlock()

	if t.State.Phase != objects.ShareDistribution {
		return fmt.Errorf("%w because it's not in ShareDistribution phase", objects.ErrCanNotContinue)
	}

	participants := t.State.GetSortedParticipants()
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
	t.State.Participants[t.State.Account.Address].Commitments = commitments
	t.State.Participants[t.State.Account.Address].EncryptedShares = encryptedShares

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

	logger.Info("ShareDistributionTask doTask()")

	c := eth.Contracts()
	me := t.State.Account.Address
	logger.Debugf("me:%v", me.Hex())
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		logger.Errorf("getting txn opts failed: %v", err)
		return err
	}

	//TODO: add retry logic and timeout

	// Distribute shares
	txn, err := c.Ethdkg().DistributeShares(txnOpts, t.State.Participants[me].EncryptedShares, t.State.Participants[me].Commitments)
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
// if the DKG process is in the right phase and blocks
// range and the distributed share hash is empty, we retry
func (t *ShareDistributionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("ShareDistributionTask ShouldRetry()")

	// This wraps the retry logic for the general case
	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	if t.State.Phase != objects.ShareDistribution {
		return false
	}

	// If it's generally good to retry, let's try to be more specific
	callOpts := eth.GetCallOpts(ctx, t.State.Account)
	participantState, err := eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, t.State.Account.Address)
	if err != nil {
		logger.Errorf("ShareDistributionTask.ShoudRetry() unable to GetParticipantInternalState(): %v", err)
		return true
	}

	logger.Infof("DistributionHash: %x", participantState.DistributedSharesHash)
	var emptySharesHash [32]byte
	if participantState.DistributedSharesHash == emptySharesHash {
		logger.Warn("Did not distribute shares after all. needs retry")
		return true
	}

	logger.Info("Did distribute shares after all. needs no retry")

	return false
}

// DoDone creates a log entry saying task is complete
func (t *ShareDistributionTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("ShareDistributionTask done")
}
