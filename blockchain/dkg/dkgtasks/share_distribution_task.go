package dkgtasks

import (
	"context"
	"math/big"
	"sync"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// ShareDistributionTask stores the data required safely distribute shares
type ShareDistributionTask struct {
	sync.Mutex
	OriginalRegistrationEnd uint64
	State                   *objects.DkgState
}

// NewShareDistributionTask creates a new task
func NewShareDistributionTask(state *objects.DkgState) *ShareDistributionTask {
	return &ShareDistributionTask{
		OriginalRegistrationEnd: state.RegistrationEnd, // If these quit being equal, this task should be abandoned
		State:                   state,
	}
}

func (t *ShareDistributionTask) init(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) bool {

	state := t.State

	if state.Participants == nil {

		me := state.Account
		callOpts := eth.GetCallOpts(ctx, me)

		// Retrieve information about other participants from smart contracts
		participants, index, err := dkg.RetrieveParticipants(eth, callOpts)
		if err != nil {
			logger.Errorf("Failed to retrieve other participants: %v", err)
			return false
		}

		numParticipants := len(participants)
		threshold, _ := math.ThresholdForUserCount(numParticipants)

		// Generate shares
		encryptedShares, privateCoefficients, commitments, err := math.GenerateShares(
			state.TransportPrivateKey, state.TransportPublicKey,
			participants, threshold)
		if err != nil {
			logger.Errorf("Failed to generate shares: %v", err)
			return false
		}

		// Store calculated values

		state.Commitments = make(map[common.Address][][2]*big.Int)
		state.Commitments[me.Address] = commitments
		state.EncryptedShares = make(map[common.Address][]*big.Int)
		state.EncryptedShares[me.Address] = encryptedShares
		state.Index = index
		state.NumberOfValidators = numParticipants
		state.Participants = participants
		state.PrivateCoefficients = privateCoefficients
		state.SecretValue = privateCoefficients[0]
		state.ValidatorThreshold = threshold
	}

	// Making it here means initialization has been completed
	return true
}

// DoWork is the first attempt at distributing shares via ethdkg
func (t *ShareDistributionTask) DoWork(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) bool {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is subsequent attempts at distributing shares via ethdkg
func (t *ShareDistributionTask) DoRetry(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) bool {
	return t.doTask(ctx, logger, eth)
}

func (t *ShareDistributionTask) doTask(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) bool {

	t.Lock()
	defer t.Unlock()

	// Is there any point in running? Make sure we're both initialized and within block range
	if !t.init(ctx, logger, eth) {
		return false
	}

	// Setup
	c := eth.Contracts()
	state := t.State
	me := state.Account.Address
	logger.Debugf("me:%v", me.Hex())
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		logger.Errorf("getting txn opts failed: %v", err)
		return false
	}

	// Distribute shares
	logger.Infof("# shares:%d # commitments:%d", len(state.EncryptedShares), len(state.Commitments))
	txn, err := c.Ethdkg().DistributeShares(txnOpts, state.EncryptedShares[me], state.Commitments[me])
	if err != nil {
		logger.Errorf("distributing shares failed: %v", err)
		return false
	}
	// Waiting for receipt
	receipt, err := eth.Queue().QueueAndWait(ctx, txn)
	if err != nil {
		logger.Errorf("waiting for receipt failed: %v", err)
		return false
	}
	if receipt == nil {
		logger.Error("missing distribute shares receipt")
		return false
	}

	// Check receipt to confirm we were successful
	if receipt.Status != uint64(1) {
		logger.Errorf("receipt status indicates failure: %v", receipt.Status)
		return false
	}

	return true
}

// ShouldRetry checks if it makes sense to try again
func (t *ShareDistributionTask) ShouldRetry(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) bool {

	t.Lock()
	defer t.Unlock()

	state := t.State

	// This wraps the retry logic for the general case
	generalRetry := GeneralTaskShouldRetry(ctx, state.Account, logger, eth, state.TransportPublicKey, t.OriginalRegistrationEnd, state.ShareDistributionEnd)

	// If it's generally good to retry, let's try to be more specific
	if generalRetry {
		callOpts := eth.GetCallOpts(ctx, state.Account)
		distributionHash, err := eth.Contracts().Ethdkg().ShareDistributionHashes(callOpts, state.Account.Address)
		if err != nil {
			return true
		}

		// TODO can I prove this is the correct share distribution hash?
		logger.Infof("DistributionHash: %x", distributionHash)
	}

	// return generalRetry
	return false
}

// DoDone creates a log entry saying task is complete
func (t *ShareDistributionTask) DoDone(logger *logrus.Logger) {
	logger.Infof("done")
}
