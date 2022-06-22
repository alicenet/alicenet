package dkgtasks

import (
	"context"
	"fmt"
	"math/big"

	"github.com/alicenet/alicenet/blockchain/dkg"
	"github.com/alicenet/alicenet/blockchain/dkg/math"
	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/crypto/bn256"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// DisputeGPKjTask contains required state for performing a group accusation
type DisputeGPKjTask struct {
	*ExecutionData
}

// asserting that DisputeGPKjTask struct implements interface interfaces.Task
var _ interfaces.Task = &DisputeGPKjTask{}

// NewDisputeGPKjTask creates a background task that attempts perform a group accusation if necessary
func NewDisputeGPKjTask(state *objects.DkgState, start uint64, end uint64) *DisputeGPKjTask {
	return &DisputeGPKjTask{
		ExecutionData: NewExecutionData(state, start, end),
	}
}

// Initialize prepares for work to be done in the GPKjDispute phase.
// Here, we determine if anyone submitted an invalid gpkj.
func (t *DisputeGPKjTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {

	logger.Info("GPKJDisputeTask Initialize()...")

	dkgData, ok := state.(objects.ETHDKGTaskData)
	if !ok {
		return objects.ErrCanNotContinue
	}

	unlock := dkgData.LockState()
	defer unlock()
	if dkgData.State != t.State {
		t.State = dkgData.State
	}

	if t.State.Phase != objects.DisputeGPKJSubmission && t.State.Phase != objects.GPKJSubmission {
		return fmt.Errorf("%w because it's not DisputeGPKJSubmission phase", objects.ErrCanNotContinue)
	}

	var (
		groupPublicKeys  [][4]*big.Int
		groupCommitments [][][2]*big.Int
	)

	var participantList = t.State.GetSortedParticipants()

	for _, participant := range participantList {
		// Build array
		groupPublicKeys = append(groupPublicKeys, participant.GPKj)
		groupCommitments = append(groupCommitments, participant.Commitments)
	}

	honest, dishonest, missing, err := math.CategorizeGroupSigners(groupPublicKeys, participantList, groupCommitments)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "Failed to determine honest vs dishonest validators: %v", err)
	}

	inverse, err := math.InverseArrayForUserCount(t.State.NumberOfValidators)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "Failed to calculate inversion: %v", err)
	}

	logger.Debugf("   Honest indices: %v", honest.ExtractIndices())
	logger.Debugf("Dishonest indices: %v", dishonest.ExtractIndices())
	logger.Debugf("  Missing indices: %v", missing.ExtractIndices())

	t.State.DishonestValidators = dishonest
	t.State.HonestValidators = honest
	t.State.Inverse = inverse

	return nil
}

// DoWork is the first attempt at submitting an invalid gpkj accusation
func (t *DisputeGPKjTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at submitting an invalid gpkj accusation
func (t *DisputeGPKjTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *DisputeGPKjTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("GPKJDisputeTask doTask()")

	// Perform group accusation
	logger.Infof("   Honest indices: %v", t.State.HonestValidators.ExtractIndices())
	logger.Infof("Dishonest indices: %v", t.State.DishonestValidators.ExtractIndices())

	var groupEncryptedSharesHash [][32]byte
	var groupCommitments [][][2]*big.Int
	var validatorAddresses []common.Address
	var participantList = t.State.GetSortedParticipants()

	for _, participant := range participantList {
		// Get group encrypted shares
		es := participant.EncryptedShares
		encryptedSharesBin, err := bn256.MarshalBigIntSlice(es)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "group accusation failed: %v", err)
		}
		hashSlice := crypto.Hasher(encryptedSharesBin)
		var hashSlice32 [32]byte
		copy(hashSlice32[:], hashSlice)
		groupEncryptedSharesHash = append(groupEncryptedSharesHash, hashSlice32)
		// Get group commitments
		com := participant.Commitments
		groupCommitments = append(groupCommitments, com)
		validatorAddresses = append(validatorAddresses, participant.Address)
	}

	callOpts := eth.GetCallOpts(ctx, t.State.Account)

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

	// Loop through dishonest participants and perform accusation
	for _, dishonestParticipant := range t.State.DishonestValidators {

		isValidator, err := eth.Contracts().ValidatorPool().IsValidator(callOpts, dishonestParticipant.Address)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "getting isValidator failed: %v", err)
		}

		if !isValidator {
			continue
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantSubmittedBadGPKJ(txnOpts, validatorAddresses, groupEncryptedSharesHash, groupCommitments, dishonestParticipant.Address)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "group accusation failed: %v", err)
		}
		t.TxOpts.TxHashes = append(t.TxOpts.TxHashes, txn.Hash())
		t.TxOpts.GasFeeCap = txn.GasFeeCap()
		t.TxOpts.GasTipCap = txn.GasTipCap()
		t.TxOpts.Nonce = big.NewInt(int64(txn.Nonce()))

		logger.WithFields(logrus.Fields{
			"GasFeeCap": t.TxOpts.GasFeeCap,
			"GasTipCap": t.TxOpts.GasTipCap,
			"Nonce":     t.TxOpts.Nonce,
		}).Info("bad gpkj dispute fees")

		// Queue transaction
		eth.Queue().QueueTransaction(ctx, txn)
	}

	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// if the DKG process is in the right phase and blocks
// range and there still someone to accuse, the retry
// is executed
func (t *DisputeGPKjTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("GPKJDisputeTask ShouldRetry()")

	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	if t.State.Phase != objects.DisputeGPKJSubmission {
		return false
	}

	callOpts := eth.GetCallOpts(ctx, t.State.Account)
	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	if err != nil {
		logger.Error("could not get BadParticipants")
		return true
	}

	logger.WithFields(logrus.Fields{
		"state.BadShares":     len(t.State.BadShares),
		"eth.badParticipants": badParticipants,
	}).Debug("DisputeGPKjTask ShouldRetry2()")

	return len(t.State.DishonestValidators) != int(badParticipants.Int64())
}

// DoDone creates a log entry saying task is complete
func (t *DisputeGPKjTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("GPKJDisputeTask done")
}

func (t *DisputeGPKjTask) GetExecutionData() interface{} {
	return t.ExecutionData
}
