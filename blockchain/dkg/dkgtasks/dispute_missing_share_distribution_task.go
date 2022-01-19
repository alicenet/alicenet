package dkgtasks

import (
	"context"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// DisputeMissingShareDistributionTask stores the data required to dispute shares
type DisputeMissingShareDistributionTask struct {
	State   *objects.DkgState
	Success bool
}

// NewDisputeMissingShareDistributionTask creates a new task
func NewDisputeMissingShareDistributionTask(state *objects.DkgState) *DisputeMissingShareDistributionTask {
	return &DisputeMissingShareDistributionTask{
		State: state,
	}
}

// Initialize begins the setup phase for DisputeMissingShareDistributionTask.
// It determines if the shares previously distributed are valid.
// If any are invalid, disputes will be issued.
func (t *DisputeMissingShareDistributionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {

	logger.Info("DisputeMissingShareDistributionTask Initializing...")

	return nil
}

// DoWork is the first attempt at disputing distributed shares
func (t *DisputeMissingShareDistributionTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is subsequent attempts at disputing distributed shares
func (t *DisputeMissingShareDistributionTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *DisputeMissingShareDistributionTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("DisputeMissingShareDistributionTask doTask()")

	var missingParticipants = make(map[common.Address]*objects.Participant)

	// get validators data
	callOpts := eth.GetCallOpts(ctx, t.State.Account)
	validators, _, err := dkg.RetrieveParticipants(callOpts, eth, logger)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "DisputeMissingShareDistributionTask doTask() error getting validators data: %v", err)
	}

	// add all validators to missing
	for _, v := range validators {
		if v != nil {
			missingParticipants[v.Address] = v
		}
	}

	// filter out validators who submitted key shares
	for participant := range t.State.KeyShareG1s {
		// remove from missing
		delete(missingParticipants, participant)
	}

	// check for validator.DistributedShares == 0
	var accusableParticipants []common.Address
	var emptySharesHash [32]byte
	for address, v := range missingParticipants {
		if v.DistributedSharesHash == emptySharesHash {
			// did not submit
			accusableParticipants = append(accusableParticipants, address)
		}
	}

	// accuse missing validators
	if len(accusableParticipants) > 0 {
		txOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingShareDistributionTask doTask() error getting txOpts: %v", err)
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantDidNotDistributeShares(txOpts, accusableParticipants)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingShareDistributionTask doTask() error accusing missing key shares: %v", err)
		}

		// Waiting for receipt
		receipt, err := eth.Queue().QueueAndWait(ctx, txn)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingShareDistributionTask doTask() error waiting for receipt failed: %v", err)
		}
		if receipt == nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingShareDistributionTask doTask() error missing share dispute receipt")
		}

		// Check receipt to confirm we were successful
		if receipt.Status != uint64(1) {
			return dkg.LogReturnErrorf(logger, "missing share distribution dispute status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
		}
	}

	t.Success = true
	return nil
}

// ShouldRetry checks if it makes sense to try again
func (t *DisputeMissingShareDistributionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("DisputeMissingShareDistributionTask ShouldRetry()")

	var phaseStart = t.State.PhaseStart + t.State.PhaseLength
	var phaseEnd = phaseStart + t.State.PhaseLength

	currentBlock, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return true
	}
	logger = logger.WithField("CurrentHeight", currentBlock)

	if t.State.Phase == objects.ShareDistribution &&
		phaseStart <= currentBlock &&
		currentBlock < phaseEnd {
		return true
	}

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, t.State.Account, logger, eth,
		t.State.TransportPublicKey, phaseStart, phaseEnd)
}

// DoDone creates a log entry saying task is complete
func (t *DisputeMissingShareDistributionTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Info("DisputeMissingShareDistributionTask done")
}
