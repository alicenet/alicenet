package dkgtasks

import (
	"context"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// DisputeMissingRegistrationTask contains required state for accusing missing registrations
type DisputeMissingRegistrationTask struct {
	State   *objects.DkgState
	Success bool
}

// NewDisputeMissingRegistrationTask creates a background task to accuse missing registrations during ETHDKG
func NewDisputeMissingRegistrationTask(state *objects.DkgState) *DisputeMissingRegistrationTask {
	return &DisputeMissingRegistrationTask{
		State: state,
	}
}

// Initialize begins the setup phase for Register.
// We construct our TransportPrivateKey and TransportPublicKey
// which will be used in the ShareDistribution phase for secure communication.
// These keys are *not* used otherwise.
func (t *DisputeMissingRegistrationTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {

	logger.Info("DisputeMissingShareDistributionTask Initializing...")

	return nil
}

// DoWork is the first attempt at registering with ethdkg
func (t *DisputeMissingRegistrationTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at registering with ethdkg
func (t *DisputeMissingRegistrationTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *DisputeMissingRegistrationTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("DisputeMissingRegistrationTask doTask()")

	var missingParticipants = make(map[common.Address]*objects.Participant)

	// get validators data
	callOpts := eth.GetCallOpts(ctx, t.State.Account)
	validators, _, err := dkg.RetrieveParticipants(callOpts, eth, logger)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "DisputeMissingRegistrationTask doTask() error getting validators data: %v", err)
	}

	// add all validators to missing
	for _, v := range validators {
		if v != nil {
			missingParticipants[v.Address] = v
		}
	}

	// filter out validators who registered
	for _, participant := range t.State.Participants {
		// remove from missing
		delete(missingParticipants, participant.Address)
	}

	// check for validator.PublicKey == 0
	var accusableParticipants []common.Address
	for address, v := range missingParticipants {
		if v.Nonce != t.State.Nonce ||
			v.Phase != uint8(objects.RegistrationOpen) ||
			v.PublicKey[0].Cmp(big.NewInt(0)) == 0 ||
			v.PublicKey[1].Cmp(big.NewInt(0)) == 0 {
			// did not register
			accusableParticipants = append(accusableParticipants, address)
		}
	}

	// accuse missing validators
	if len(accusableParticipants) > 0 {
		logger.Warnf("Accusing missing registrations: %v", accusableParticipants)

		txOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingRegistrationTask doTask() error getting txOpts: %v", err)
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantNotRegistered(txOpts, accusableParticipants)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingRegistrationTask doTask() error accusing missing key shares: %v", err)
		}

		// Waiting for receipt
		receipt, err := eth.Queue().QueueAndWait(ctx, txn)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingRegistrationTask doTask() error waiting for receipt failed: %v", err)
		}
		if receipt == nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingRegistrationTask doTask() error missing share dispute receipt")
		}

		// Check receipt to confirm we were successful
		if receipt.Status != uint64(1) {
			return dkg.LogReturnErrorf(logger, "missing registration dispute status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
		}
	} else {
		logger.Info("No accusations for missing registrations")
	}

	t.Success = true
	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
// -- the registration open hasn't moved, i.e. ETHDKG has not restarted
func (t *DisputeMissingRegistrationTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("DisputeMissingRegistrationTask ShouldRetry()")

	//c := eth.Contracts()
	callOpts := eth.GetCallOpts(ctx, t.State.Account)

	currentBlock, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return true
	}
	logger = logger.WithField("CurrentHeight", currentBlock)

	var phaseEnd = t.State.PhaseStart + t.State.PhaseLength

	// Definitely past quitting time
	if currentBlock >= phaseEnd {
		logger.Info("aborting registration due to scheduled end")
		return false
	}

	// Check to see if we are already registered
	ethdkg := eth.Contracts().Ethdkg()
	status, err := CheckRegistration(ctx, ethdkg, logger, callOpts, t.State.Account.Address, t.State.TransportPublicKey)
	if err != nil {
		logger.Warnf("could not check if we're registered: %v", err)
		return true
	}

	if status == Registered || status == BadRegistration {
		return false
	}

	return true
}

// DoDone just creates a log entry saying task is complete
func (t *DisputeMissingRegistrationTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("DisputeMissingRegistrationTask done")
}
