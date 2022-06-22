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

// DisputeMissingKeySharesTask stores the data required to dispute shares
type DisputeMissingKeySharesTask struct {
	*ExecutionData
}

// asserting that DisputeMissingKeySharesTask struct implements interface interfaces.Task
var _ interfaces.Task = &DisputeMissingKeySharesTask{}

// NewDisputeMissingKeySharesTask creates a new task
func NewDisputeMissingKeySharesTask(state *objects.DkgState, start uint64, end uint64) *DisputeMissingKeySharesTask {
	return &DisputeMissingKeySharesTask{
		ExecutionData: NewExecutionData(state, start, end),
	}
}

// Initialize begins the setup phase for DisputeMissingKeySharesTask.
func (t *DisputeMissingKeySharesTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	logger.Info("Initializing DisputeMissingKeySharesTask...")

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

	accusableParticipants, err := t.getAccusableParticipants(ctx, eth, logger)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "DisputeMissingKeySharesTask doTask() error getting accusableParticipants: %v", err)
	}

	// accuse missing validators
	if len(accusableParticipants) > 0 {
		logger.Warnf("Accusing missing key shares: %v", accusableParticipants)

		txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingKeySharesTask doTask() error getting txnOpts: %v", err)
		}

		// If the TxOpts exists, meaning the Tx replacement timeout was reached,
		// we increase the Gas to have priority for the next blocks
		if t.TxOpts != nil && t.TxOpts.Nonce != nil {
			logger.Info("txnOpts Replaced")
			txnOpts.Nonce = t.TxOpts.Nonce
			txnOpts.GasFeeCap = t.TxOpts.GasFeeCap
			txnOpts.GasTipCap = t.TxOpts.GasTipCap
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantDidNotSubmitKeyShares(txnOpts, accusableParticipants)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "DisputeMissingKeySharesTask doTask() error accusing missing key shares: %v", err)
		}
		t.TxOpts.TxHashes = append(t.TxOpts.TxHashes, txn.Hash())
		t.TxOpts.GasFeeCap = txn.GasFeeCap()
		t.TxOpts.GasTipCap = txn.GasTipCap()
		t.TxOpts.Nonce = big.NewInt(int64(txn.Nonce()))

		logger.WithFields(logrus.Fields{
			"GasFeeCap": t.TxOpts.GasFeeCap,
			"GasTipCap": t.TxOpts.GasTipCap,
			"Nonce":     t.TxOpts.Nonce,
		}).Info("missing key shares dispute fees")

		// Queue transaction
		eth.Queue().QueueTransaction(ctx, txn)
	} else {
		logger.Info("No accusations for missing key shares")
	}

	t.Success = true
	return nil
}

// ShouldRetry checks if it makes sense to try again
// if the DKG process is in the right phase and blocks
// range and there still someone to accuse, the retry
// is executed
func (t *DisputeMissingKeySharesTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("DisputeMissingKeySharesTask ShouldRetry()")

	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	if t.State.Phase != objects.KeyShareSubmission {
		return false
	}

	accusableParticipants, err := t.getAccusableParticipants(ctx, eth, logger)
	if err != nil {
		logger.Error("could not get accusableParticipants")
		return true
	}

	if len(accusableParticipants) > 0 {
		return true
	}

	return false
}

// DoDone creates a log entry saying task is complete
func (t *DisputeMissingKeySharesTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Info("DisputeMissingKeySharesTask done")
}

func (t *DisputeMissingKeySharesTask) GetExecutionData() interface{} {
	return t.ExecutionData
}

func (t *DisputeMissingKeySharesTask) getAccusableParticipants(ctx context.Context, eth interfaces.Ethereum, logger *logrus.Entry) ([]common.Address, error) {
	var accusableParticipants []common.Address
	callOpts := eth.GetCallOpts(ctx, t.State.Account)

	validators, err := dkg.GetValidatorAddressesFromPool(callOpts, eth, logger)
	if err != nil {
		return nil, dkg.LogReturnErrorf(logger, "DisputeMissingShareDistributionTask getAccusableParticipants() error getting validators: %v", err)
	}

	validatorsMap := make(map[common.Address]bool)
	for _, validator := range validators {
		validatorsMap[validator] = true
	}

	// find participants who did not submit they key shares
	for _, p := range t.State.Participants {
		_, isValidator := validatorsMap[p.Address]
		if isValidator && (p.Nonce != t.State.Nonce ||
			p.Phase != objects.KeyShareSubmission ||
			(p.KeyShareG1s[0].Cmp(big.NewInt(0)) == 0 &&
				p.KeyShareG1s[1].Cmp(big.NewInt(0)) == 0) ||
			(p.KeyShareG1CorrectnessProofs[0].Cmp(big.NewInt(0)) == 0 &&
				p.KeyShareG1CorrectnessProofs[1].Cmp(big.NewInt(0)) == 0) ||
			(p.KeyShareG2s[0].Cmp(big.NewInt(0)) == 0 &&
				p.KeyShareG2s[1].Cmp(big.NewInt(0)) == 0 &&
				p.KeyShareG2s[2].Cmp(big.NewInt(0)) == 0 &&
				p.KeyShareG2s[3].Cmp(big.NewInt(0)) == 0)) {
			// did not submit
			accusableParticipants = append(accusableParticipants, p.Address)
		}
	}

	return accusableParticipants, nil
}
