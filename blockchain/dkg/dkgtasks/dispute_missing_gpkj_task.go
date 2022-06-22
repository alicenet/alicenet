package dkgtasks

import (
	"context"
	"math/big"

	"github.com/alicenet/alicenet/blockchain/dkg"
	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// DisputeMissingGPKjTask stores the data required to dispute shares
type DisputeMissingGPKjTask struct {
	*ExecutionData
}

// asserting that DisputeMissingGPKjTask struct implements interface interfaces.Task
var _ interfaces.Task = &DisputeMissingGPKjTask{}

// NewDisputeMissingGPKjTask creates a new task
func NewDisputeMissingGPKjTask(state *objects.DkgState, start uint64, end uint64) *DisputeMissingGPKjTask {
	return &DisputeMissingGPKjTask{
		ExecutionData: NewExecutionData(state, start, end),
	}
}

// Initialize begins the setup phase for DisputeMissingGPKjTask.
// It determines if the shares previously distributed are valid.
// If any are invalid, disputes will be issued.
func (t *DisputeMissingGPKjTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	logger.Info("Initializing DisputeMissingGPKjTask...")

	dkgData, ok := state.(objects.ETHDKGTaskData)
	if !ok {
		return objects.ErrCanNotContinue
	}

	unlock := dkgData.LockState()
	defer unlock()
	if dkgData.State != t.State {
		t.State = dkgData.State
	}

	return nil
}

// DoWork is the first attempt at disputing distributed shares
func (t *DisputeMissingGPKjTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is subsequent attempts at disputing distributed shares
func (t *DisputeMissingGPKjTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *DisputeMissingGPKjTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("DisputeMissingGPKjTask doTask()")

	accusableParticipants, err := t.getAccusableParticipants(ctx, eth, logger)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "DisputeMissingGPKjTask doTask() error getting accusableParticipants: %v", err)
	}

	// accuse missing validators
	if len(accusableParticipants) > 0 {
		logger.Warnf("Accusing missing gpkj: %v", accusableParticipants)

		txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingGPKjTask doTask() error getting txnOpts: %v", err)
		}

		// If the TxOpts exists, meaning the Tx replacement timeout was reached,
		// we increase the Gas to have priority for the next blocks
		if t.TxOpts != nil && t.TxOpts.Nonce != nil {
			logger.Info("txnOpts Replaced")
			txnOpts.Nonce = t.TxOpts.Nonce
			txnOpts.GasFeeCap = t.TxOpts.GasFeeCap
			txnOpts.GasTipCap = t.TxOpts.GasTipCap
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantDidNotSubmitGPKJ(txnOpts, accusableParticipants)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingGPKjTask doTask() error accusing missing gpkj: %v", err)
		}
		t.TxOpts.TxHashes = append(t.TxOpts.TxHashes, txn.Hash())
		t.TxOpts.GasFeeCap = txn.GasFeeCap()
		t.TxOpts.GasTipCap = txn.GasTipCap()
		t.TxOpts.Nonce = big.NewInt(int64(txn.Nonce()))

		logger.WithFields(logrus.Fields{
			"GasFeeCap": t.TxOpts.GasFeeCap,
			"GasTipCap": t.TxOpts.GasTipCap,
			"Nonce":     t.TxOpts.Nonce,
		}).Info("missing gpkj dispute fees")

		// Queue transaction
		eth.Queue().QueueTransaction(ctx, txn)
	} else {
		logger.Info("No accusations for missing gpkj")
	}

	t.Success = true
	return nil
}

// ShouldRetry checks if it makes sense to try again
// if the DKG process is in the right phase and blocks
// range and there still someone to accuse, the retry
// is executed
func (t *DisputeMissingGPKjTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("DisputeMissingGPKjTask ShouldRetry()")

	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	if t.State.Phase != objects.GPKJSubmission {
		return false
	}

	accusableParticipants, err := t.getAccusableParticipants(ctx, eth, logger)
	if err != nil {
		logger.Errorf("DisputeMissingGPKjTask ShouldRetry() error getting accusable participants: %v", err)
		return true
	}

	if len(accusableParticipants) > 0 {
		return true
	}

	return false
}

// DoDone creates a log entry saying task is complete
func (t *DisputeMissingGPKjTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Info("DisputeMissingGPKjTask done")
}

func (t *DisputeMissingGPKjTask) GetExecutionData() interface{} {
	return t.ExecutionData
}

func (t *DisputeMissingGPKjTask) getAccusableParticipants(ctx context.Context, eth interfaces.Ethereum, logger *logrus.Entry) ([]common.Address, error) {
	var accusableParticipants []common.Address
	callOpts := eth.GetCallOpts(ctx, t.State.Account)

	validators, err := dkg.GetValidatorAddressesFromPool(callOpts, eth, logger)
	if err != nil {
		return nil, dkg.LogReturnErrorf(logger, "DisputeMissingGPKjTask getAccusableParticipants() error getting validators: %v", err)
	}

	validatorsMap := make(map[common.Address]bool)
	for _, validator := range validators {
		validatorsMap[validator] = true
	}

	// find participants who did not submit GPKj
	for _, p := range t.State.Participants {
		_, isValidator := validatorsMap[p.Address]
		if isValidator && (p.Nonce != t.State.Nonce ||
			p.Phase != objects.GPKJSubmission ||
			(p.GPKj[0].Cmp(big.NewInt(0)) == 0 &&
				p.GPKj[1].Cmp(big.NewInt(0)) == 0 &&
				p.GPKj[2].Cmp(big.NewInt(0)) == 0 &&
				p.GPKj[3].Cmp(big.NewInt(0)) == 0)) {
			// did not submit
			accusableParticipants = append(accusableParticipants, p.Address)
		}
	}

	return accusableParticipants, nil
}
