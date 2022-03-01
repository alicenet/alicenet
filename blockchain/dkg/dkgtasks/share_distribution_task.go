package dkgtasks

import (
	"context"
	"fmt"
	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
	"math/big"
)

// ShareDistributionTask stores the data required safely distribute shares
type ShareDistributionTask struct {
	*DkgTask
}

// asserting that ShareDistributionTask struct implements interface interfaces.Task
var _ interfaces.Task = &ShareDistributionTask{}

// asserting that ShareDistributionTask struct implements DkgTaskIfase
var _ DkgTaskIfase = &ShareDistributionTask{}

// NewShareDistributionTask creates a new task
func NewShareDistributionTask(state *objects.DkgState, start uint64, end uint64) *ShareDistributionTask {
	return &ShareDistributionTask{
		DkgTask: NewDkgTask(state, start, end),
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

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// If the TxReplOpts exists, meaning the Tx replacement timeout was reached,
	// we increase the Gas to have priority for the next blocks
	if t.TxReplOpts != nil && t.TxReplOpts.Nonce != nil {
		logger.Info("txnOpts Replaced")
		txnOpts.Nonce = t.TxReplOpts.Nonce
		txnOpts.GasFeeCap = t.TxReplOpts.GasFeeCap
		txnOpts.GasTipCap = t.TxReplOpts.GasTipCap
	}

	// Distribute shares
	txn, err := c.Ethdkg().DistributeShares(txnOpts, t.State.Participants[me].EncryptedShares, t.State.Participants[me].Commitments)
	if err != nil {
		logger.Errorf("distributing shares failed: %v", err)
		return err
	}
	t.TxReplOpts.TxHash = txn.Hash()
	t.TxReplOpts.GasFeeCap = txn.GasFeeCap()
	t.TxReplOpts.GasTipCap = txn.GasTipCap()
	t.TxReplOpts.Nonce = big.NewInt(int64(txn.Nonce()))

	logger.WithFields(logrus.Fields{
		"GasFeeCap": t.TxReplOpts.GasFeeCap,
		"GasTipCap": t.TxReplOpts.GasTipCap,
		"Nonce":     t.TxReplOpts.Nonce,
		"Hash":      t.TxReplOpts.TxHash.Hex(),
	}).Info("share distribution fees")

	// Queue transaction
	eth.Queue().QueueTransaction(ctx, txn)
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

func (t *ShareDistributionTask) GetDkgTask() *DkgTask {
	return t.DkgTask
}

func (t *ShareDistributionTask) SetDkgTask(dkgTask *DkgTask) {
	t.DkgTask = dkgTask
}
