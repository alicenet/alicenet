package dkgtasks

import (
	"context"
	"fmt"
	"math/big"

	"github.com/alicenet/alicenet/blockchain/dkg"
	"github.com/alicenet/alicenet/blockchain/dkg/math"
	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

// ShareDistributionTask stores the data required safely distribute shares
type ShareDistributionTask struct {
	*ExecutionData
}

// asserting that ShareDistributionTask struct implements interface interfaces.Task
var _ interfaces.Task = &ShareDistributionTask{}

// NewShareDistributionTask creates a new task
func NewShareDistributionTask(state *objects.DkgState, start uint64, end uint64) *ShareDistributionTask {
	return &ShareDistributionTask{
		ExecutionData: NewExecutionData(state, start, end),
	}
}

// Initialize begins the setup phase for ShareDistribution.
// We construct our commitments and encrypted shares before
// submitting them to the associated smart contract.
func (t *ShareDistributionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {

	logger.Infof("ShareDistributionTask Initialize()")

	dkgData, ok := state.(objects.ETHDKGTaskData)
	if !ok {
		return objects.ErrCanNotContinue
	}

	unlock := dkgData.LockState()
	defer unlock()
	if dkgData.State != t.State {
		t.State = dkgData.State
	}

	if t.State.Phase != objects.ShareDistribution {
		return fmt.Errorf("%w because it's not in ShareDistribution phase", objects.ErrCanNotContinue)
	}

	if t.State.SecretValue == nil {

		participants := t.State.GetSortedParticipants()
		numParticipants := len(participants)
		threshold := math.ThresholdForUserCount(numParticipants)

		// Generate shares
		encryptedShares, privateCoefficients, commitments, err := math.GenerateShares(
			t.State.TransportPrivateKey, participants)
		if err != nil {
			logger.Errorf("Failed to generate shares: %v %#v", err, participants)
			return err
		}

		// Store calculated values
		t.State.Participants[t.State.Account.Address].Commitments = commitments
		t.State.Participants[t.State.Account.Address].EncryptedShares = encryptedShares

		t.State.PrivateCoefficients = privateCoefficients
		t.State.SecretValue = privateCoefficients[0]
		t.State.ValidatorThreshold = threshold

		unlock()
		dkgData.PersistStateCB()
	} else {
		logger.Infof("ShareDistributionTask Initialize(): encrypted shares already defined")
	}

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

	// If the TxOpts exists, meaning the Tx replacement timeout was reached,
	// we increase the Gas to have priority for the next blocks
	if t.TxOpts != nil && t.TxOpts.Nonce != nil {
		logger.Info("txnOpts Replaced")
		txnOpts.Nonce = t.TxOpts.Nonce
		txnOpts.GasFeeCap = t.TxOpts.GasFeeCap
		txnOpts.GasTipCap = t.TxOpts.GasTipCap
	}

	// Distribute shares
	txn, err := c.Ethdkg().DistributeShares(txnOpts, t.State.Participants[me].EncryptedShares, t.State.Participants[me].Commitments)
	if err != nil {
		logger.Errorf("distributing shares failed: %v", err)
		return err
	}
	t.TxOpts.TxHashes = append(t.TxOpts.TxHashes, txn.Hash())
	t.TxOpts.GasFeeCap = txn.GasFeeCap()
	t.TxOpts.GasTipCap = txn.GasTipCap()
	t.TxOpts.Nonce = big.NewInt(int64(txn.Nonce()))

	logger.WithFields(logrus.Fields{
		"GasFeeCap": t.TxOpts.GasFeeCap,
		"GasTipCap": t.TxOpts.GasTipCap,
		"Nonce":     t.TxOpts.Nonce,
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

func (t *ShareDistributionTask) GetExecutionData() interface{} {
	return t.ExecutionData
}
