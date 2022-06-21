package dkgtasks

import (
	"context"
	"math/big"

	"github.com/alicenet/alicenet/blockchain/dkg"
	"github.com/alicenet/alicenet/blockchain/dkg/math"
	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

// KeyshareSubmissionTask is the task for submitting Keyshare information
type KeyshareSubmissionTask struct {
	*ExecutionData
}

// asserting that KeyshareSubmissionTask struct implements interface interfaces.Task
var _ interfaces.Task = &KeyshareSubmissionTask{}

// NewKeyshareSubmissionTask creates a new task
func NewKeyshareSubmissionTask(state *objects.DkgState, start uint64, end uint64) *KeyshareSubmissionTask {
	return &KeyshareSubmissionTask{
		ExecutionData: NewExecutionData(state, start, end),
	}
}

// Initialize prepares for work to be done in KeyShareSubmission phase.
// Here, the G1 key share, G1 proof, and G2 key share are constructed
// and stored for submission.
func (t *KeyshareSubmissionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	logger.Info("KeyshareSubmissionTask Initialize()")

	dkgData, ok := state.(objects.ETHDKGTaskData)
	if !ok {
		return objects.ErrCanNotContinue
	}

	unlock := dkgData.LockState()
	defer unlock()
	if dkgData.State != t.State {
		t.State = dkgData.State
	}

	me := t.State.Account.Address

	// check if task already defined key shares
	if t.State.Participants[me].KeyShareG1s[0] == nil ||
		t.State.Participants[me].KeyShareG1s[1] == nil ||
		(t.State.Participants[me].KeyShareG1s[0].Cmp(big.NewInt(0)) == 0 &&
			t.State.Participants[me].KeyShareG1s[1].Cmp(big.NewInt(0)) == 0) {

		// Generate the key shares
		g1KeyShare, g1Proof, g2KeyShare, err := math.GenerateKeyShare(t.State.SecretValue)
		if err != nil {
			return err
		}

		t.State.Participants[me].KeyShareG1s = g1KeyShare
		t.State.Participants[me].KeyShareG1CorrectnessProofs = g1Proof
		t.State.Participants[me].KeyShareG2s = g2KeyShare

		unlock()
		dkgData.PersistStateCB()
	} else {
		logger.Infof("KeyshareSubmissionTask Initialize(): key shares already defined")
	}

	return nil
}

// DoWork is the first attempt at the performing the KeyShareSubmission phase
func (t *KeyshareSubmissionTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	logger.Info("DoWork() ...")
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at the performing the KeyShareSubmission phase
func (t *KeyshareSubmissionTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	logger.Info("DoRetry() ...")
	return t.doTask(ctx, logger, eth)
}

func (t *KeyshareSubmissionTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("KeyshareSubmissionTask doTask()")

	// Setup
	me := t.State.Account

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

	// Submit Keyshares
	logger.Infof("submitting key shares: %v %v %v %v",
		me.Address,
		t.State.Participants[me.Address].KeyShareG1s,
		t.State.Participants[me.Address].KeyShareG1CorrectnessProofs,
		t.State.Participants[me.Address].KeyShareG2s)
	txn, err := eth.Contracts().Ethdkg().SubmitKeyShare(txnOpts,
		t.State.Participants[me.Address].KeyShareG1s,
		t.State.Participants[me.Address].KeyShareG1CorrectnessProofs,
		t.State.Participants[me.Address].KeyShareG2s)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "submitting keyshare failed: %v", err)
	}
	t.TxOpts.TxHashes = append(t.TxOpts.TxHashes, txn.Hash())
	t.TxOpts.GasFeeCap = txn.GasFeeCap()
	t.TxOpts.GasTipCap = txn.GasTipCap()
	t.TxOpts.Nonce = big.NewInt(int64(txn.Nonce()))

	logger.WithFields(logrus.Fields{
		"GasFeeCap": t.TxOpts.GasFeeCap,
		"GasTipCap": t.TxOpts.GasTipCap,
		"Nonce":     t.TxOpts.Nonce,
	}).Info("key share submission fees")

	// Queue transaction
	eth.Queue().QueueTransaction(ctx, txn)
	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
func (t *KeyshareSubmissionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("KeyshareSubmissionTask ShouldRetry()")

	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	state := t.State

	me := state.Account
	callOpts := eth.GetCallOpts(ctx, me)

	phase, err := eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	if err != nil {
		logger.Infof("KeyshareSubmissionTask ShouldRetry GetETHDKGPhase error: %v", err)
		return true
	}

	// DisputeShareDistribution || KeyShareSubmission
	if phase != uint8(objects.DisputeShareDistribution) && phase != uint8(objects.KeyShareSubmission) {
		return false
	}

	// Check the key share submission status
	status, err := CheckKeyShare(ctx, eth.Contracts().Ethdkg(), logger, callOpts, me.Address, state.Participants[me.Address].KeyShareG1s)
	if err != nil {
		logger.Errorf("KeyshareSubmissionTask ShouldRetry CheckKeyShare error: %v", err)
		return true
	}

	if status == KeyShared || status == BadKeyShared {
		return false
	}

	return true
}

// DoDone creates a log entry saying task is complete
func (t *KeyshareSubmissionTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("KeyshareSubmissionTask done")
}

func (t *KeyshareSubmissionTask) GetExecutionData() interface{} {
	return t.ExecutionData
}
