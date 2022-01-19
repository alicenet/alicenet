package dkgtasks

import (
	"context"
	"fmt"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// DisputeMissingKeySharesTask stores the data required to dispute shares
type DisputeMissingKeySharesTask struct {
	State   *objects.DkgState
	Success bool
}

// NewDisputeMissingKeySharesTask creates a new task
func NewDisputeMissingKeySharesTask(state *objects.DkgState) *DisputeMissingKeySharesTask {
	return &DisputeMissingKeySharesTask{
		State: state,
	}
}

// Initialize begins the setup phase for DisputeMissingKeySharesTask.
// It determines if the shares previously distributed are valid.
// If any are invalid, disputes will be issued.
func (t *DisputeMissingKeySharesTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	dkgState, validState := state.(*objects.DkgState)
	if !validState {
		panic(fmt.Errorf("%w invalid state type", objects.ErrCanNotContinue))
	}

	t.State = dkgState

	logger.Info("Initializing DisputeMissingKeySharesTask...")

	return nil
}

// DoWork is the first attempt at disputing distributed shares
func (t *DisputeMissingKeySharesTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is subsequent attempts at disputing distributed shares
func (t *DisputeMissingKeySharesTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *DisputeMissingKeySharesTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("DisputeMissingKeySharesTask doTask()")

	var missingParticipants = make(map[common.Address]*objects.Participant)

	// get validators data
	callOpts := eth.GetCallOpts(ctx, t.State.Account)
	validators, _, err := dkg.RetrieveParticipants(callOpts, eth, logger)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "DisputeMissingKeySharesTask doTask() error getting validators data: %v", err)
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

	// check for validator.KeyShares == 0
	var accusableParticipants []common.Address
	for address, v := range missingParticipants {
		if v.KeyShares[0].Cmp(big.NewInt(0)) == 0 && v.KeyShares[1].Cmp(big.NewInt(0)) == 0 {
			// did not submit
			accusableParticipants = append(accusableParticipants, address)
		}
	}

	// accuse missing validators
	if len(accusableParticipants) > 0 {
		txOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingKeySharesTask doTask() error getting txOpts: %v", err)
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantDidNotSubmitKeyShares(txOpts, accusableParticipants)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingKeySharesTask doTask() error accusing missing key shares: %v", err)
		}

		// Waiting for receipt
		receipt, err := eth.Queue().QueueAndWait(ctx, txn)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingKeySharesTask doTask() error waiting for receipt failed: %v", err)
		}
		if receipt == nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingKeySharesTask doTask() error missing share dispute receipt")
		}

		// Check receipt to confirm we were successful
		if receipt.Status != uint64(1) {
			return dkg.LogReturnErrorf(logger, "missing key share dispute status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
		}
	}

	t.Success = true
	return nil
}

// ShouldRetry checks if it makes sense to try again
func (t *DisputeMissingKeySharesTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("DisputeMissingKeySharesTask ShouldRetry()")

	var phaseStart = t.State.PhaseStart + t.State.PhaseLength
	var phaseEnd = phaseStart + t.State.PhaseLength

	currentBlock, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return true
	}
	logger = logger.WithField("CurrentHeight", currentBlock)

	if t.State.Phase == objects.KeyShareSubmission &&
		phaseStart <= currentBlock &&
		currentBlock < phaseEnd {
		return true
	}

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, t.State.Account, logger, eth,
		t.State.TransportPublicKey, phaseStart, phaseEnd)
}

// DoDone creates a log entry saying task is complete
func (t *DisputeMissingKeySharesTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Info("DisputeMissingKeySharesTask done")
}
