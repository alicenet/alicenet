package dkgtasks

import (
	"context"
	"fmt"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/math"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/objects"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/constants"
	"github.com/sirupsen/logrus"
)

// GPKjSubmissionTask contains required state for gpk submission
type GPKjSubmissionTask struct {
	*tasks.Task
	adminHandler interfaces.AdminHandler
}

// asserting that GPKjSubmissionTask struct implements interface interfaces.Task
var _ interfaces.ITask = &GPKjSubmissionTask{}

// NewGPKjSubmissionTask creates a background task that attempts to submit the gpkj in ETHDKG
func NewGPKjSubmissionTask(state *objects.DkgState, start uint64, end uint64, adminHandler interfaces.AdminHandler) *GPKjSubmissionTask {
	return &GPKjSubmissionTask{
		Task:         tasks.NewTask(state, start, end),
		adminHandler: adminHandler,
	}
}

// Initialize prepares for work to be done in GPKjSubmission phase.
// Here, we construct our gpkj and associated signature.
// We will submit them in DoWork.
func (t *GPKjSubmissionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {

	t.State.Lock()
	defer t.State.Unlock()
	logger.Info("GPKSubmissionTask Initialize()...")

	taskState, ok := t.State.(*objects.DkgState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	if taskState.GroupPrivateKey == nil ||
		taskState.GroupPrivateKey.Cmp(big.NewInt(0)) == 0 {

		// Collecting all the participants encrypted shares to be used for the GPKj
		var participantsList = taskState.GetSortedParticipants()
		encryptedShares := make([][]*big.Int, 0, taskState.NumberOfValidators)
		for _, participant := range participantsList {
			logger.Debugf("Collecting encrypted shares... Participant %v %v", participant.Index, participant.Address.Hex())
			encryptedShares = append(encryptedShares, participant.EncryptedShares)
		}

		// Generate the GPKj
		groupPrivateKey, groupPublicKey, err := math.GenerateGroupKeys(
			taskState.TransportPrivateKey, taskState.PrivateCoefficients,
			encryptedShares, taskState.Index, participantsList)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"t.State.Index": taskState.Index,
			}).Errorf("Could not generate group keys: %v", err)
			return tasks.LogReturnErrorf(logger, "Could not generate group keys: %v", err)
		}

		taskState.GroupPrivateKey = groupPrivateKey
		taskState.Participants[taskState.Account.Address].GPKj = groupPublicKey

		// Pass private key on to consensus
		logger.Infof("Adding private bn256eth key... using %p", t.adminHandler)
		err = t.adminHandler.AddPrivateKey(groupPrivateKey.Bytes(), constants.CurveBN256Eth)
		if err != nil {
			return fmt.Errorf("%w because error adding private key: %v", objects.ErrCanNotContinue, err)
		}
	} else {
		logger.Infof("GPKSubmissionTask Initialize(): group private-public key already defined")
	}

	return nil
}

// DoWork is the first attempt at submitting gpkj and signature.
func (t *GPKjSubmissionTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at submitting gpkj and signature.
func (t *GPKjSubmissionTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *GPKjSubmissionTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	taskState, ok := t.State.(*objects.DkgState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	logger.Infof("GPKSubmissionTask doTask(): %v", taskState.Account.Address)

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, taskState.Account)
	if err != nil {
		return tasks.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// If the TxOpts exists, meaning the Tx replacement timeout was reached,
	// we increase the Gas to have priority for the next blocks
	if t.TxOpts != nil && t.TxOpts.Nonce != nil {
		logger.Info("txnOpts Replaced")
		txnOpts.Nonce = t.TxOpts.Nonce
		txnOpts.GasFeeCap = t.TxOpts.GasFeeCap
		txnOpts.GasTipCap = t.TxOpts.GasTipCap
	}

	// Do it
	txn, err := eth.Contracts().Ethdkg().SubmitGPKJ(txnOpts, taskState.Participants[taskState.Account.Address].GPKj)
	if err != nil {
		return tasks.LogReturnErrorf(logger, "submitting master public key failed: %v", err)
	}
	t.TxOpts.TxHashes = append(t.TxOpts.TxHashes, txn.Hash())
	t.TxOpts.GasFeeCap = txn.GasFeeCap()
	t.TxOpts.GasTipCap = txn.GasTipCap()
	t.TxOpts.Nonce = big.NewInt(int64(txn.Nonce()))

	logger.WithFields(logrus.Fields{
		"GasFeeCap": t.TxOpts.GasFeeCap,
		"GasTipCap": t.TxOpts.GasTipCap,
		"Nonce":     t.TxOpts.Nonce,
	}).Info("GPKj submission fees")

	// Queue transaction
	eth.Queue().QueueTransaction(ctx, txn)
	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
func (t *GPKjSubmissionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	t.State.Lock()
	defer t.State.Unlock()
	logger.Info("GPKSubmissionTask ShouldRetry()")

	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	taskState, ok := t.State.(*objects.DkgState)
	if !ok {
		logger.Error("Invalid convertion of taskState object")
		return false
	}

	if taskState.Phase != objects.GPKJSubmission {
		return false
	}

	//Check if my GPKj is submitted, if not should retry
	me := taskState.Account
	callOpts, err := eth.GetCallOpts(ctx, me)
	if err != nil {
		logger.Debug("PKSubmissionTask ShouldRetry() failed getting call opts")
		return true
	}
	participantState, err := eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, me.Address)
	if err == nil && participantState.Gpkj[0].Cmp(taskState.Participants[me.Address].GPKj[0]) == 0 &&
		participantState.Gpkj[1].Cmp(taskState.Participants[me.Address].GPKj[1]) == 0 &&
		participantState.Gpkj[2].Cmp(taskState.Participants[me.Address].GPKj[2]) == 0 &&
		participantState.Gpkj[3].Cmp(taskState.Participants[me.Address].GPKj[3]) == 0 {
		return false
	}

	return true
}

// DoDone creates a log entry saying task is complete
func (t *GPKjSubmissionTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("GPKSubmissionTask done")
}

// SetAdminHandler sets the task adminHandler
func (t *GPKjSubmissionTask) SetAdminHandler(adminHandler interfaces.AdminHandler) {
	t.adminHandler = adminHandler
}

func (t *GPKjSubmissionTask) GetExecutionData() interfaces.ITaskExecutionData {
	return t.Task
}
