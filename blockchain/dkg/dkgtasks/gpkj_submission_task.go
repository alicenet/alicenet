package dkgtasks

import (
	"context"
	"math/big"
	"sync"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

// GPKSubmissionTask contains required state for gpk submission
type GPKSubmissionTask struct {
	sync.Mutex              // TODO Do I need this? It might be sufficient to only use a RWMutex on `State`
	OriginalRegistrationEnd uint64
	State                   *objects.DkgState
	Success                 bool
}

// NewGPKSubmissionTask creates a background task that attempts to register with ETHDKG
func NewGPKSubmissionTask(state *objects.DkgState) *GPKSubmissionTask {
	return &GPKSubmissionTask{
		OriginalRegistrationEnd: state.RegistrationEnd, // If these quit being equal, this task should be abandoned
		State:                   state,
	}
}

func (t *GPKSubmissionTask) Initialize(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {

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

	return nil
}

// DoWork is the first attempt at registering with ethdkg
func (t *GPKSubmissionTask) DoWork(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at registering with ethdkg
func (t *GPKSubmissionTask) DoRetry(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *GPKSubmissionTask) doTask(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) error {
	t.Lock()
	defer t.Unlock()

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
func (t *GPKSubmissionTask) ShouldRetry(ctx context.Context, logger *logrus.Logger, eth interfaces.Ethereum) bool {
	t.Lock()
	defer t.Unlock()

	state := t.State

	// This wraps the retry logic for the general case
	return GeneralTaskShouldRetry(ctx, state.Account, logger, eth,
		state.TransportPublicKey, t.OriginalRegistrationEnd, state.KeyShareSubmissionEnd)
}

// DoDone creates a log entry saying task is complete
func (t *GPKSubmissionTask) DoDone(logger *logrus.Logger) {
	logger.Infof("done")
}
