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
	Start   uint64
	End     uint64
	State   *objects.DkgState
	Success bool
}

// asserting that DisputeMissingKeySharesTask struct implements interface interfaces.Task
var _ interfaces.Task = &DisputeMissingKeySharesTask{}

// NewDisputeMissingKeySharesTask creates a new task
func NewDisputeMissingKeySharesTask(state *objects.DkgState, start uint64, end uint64) *DisputeMissingKeySharesTask {
	return &DisputeMissingKeySharesTask{
		Start:   start,
		End:     end,
		State:   state,
		Success: false,
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

	var accusableParticipants []common.Address

	// find participants who did not submit key shares
	for _, p := range t.State.Participants {
		if p.Nonce != t.State.Nonce ||
			p.Phase != uint8(objects.KeyShareSubmission) ||
			(p.KeyShares[0].Cmp(big.NewInt(0)) == 0 &&
				p.KeyShares[1].Cmp(big.NewInt(0)) == 0) {
			// did not submit
			accusableParticipants = append(accusableParticipants, p.Address)
		}
	}

	// accuse missing validators
	if len(accusableParticipants) > 0 {
		logger.Warnf("Accusing missing key shares: %v", accusableParticipants)

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
	} else {
		logger.Info("No accusations for missing key shares")
	}

	t.Success = true
	return nil
}

// ShouldRetry checks if it makes sense to try again
func (t *DisputeMissingKeySharesTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("DisputeMissingKeySharesTask ShouldRetry()")

	currentBlock, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return true
	}
	logger = logger.WithField("CurrentHeight", currentBlock)

	if t.State.Phase == objects.KeyShareSubmission &&
		t.Start <= currentBlock &&
		currentBlock < t.End {
		return true
	}

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, t.State.Account, logger, eth,
		t.State.TransportPublicKey, t.Start, t.End)
}

// DoDone creates a log entry saying task is complete
func (t *DisputeMissingKeySharesTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Info("DisputeMissingKeySharesTask done")
}
