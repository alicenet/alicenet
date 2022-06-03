package dkg

import (
	"context"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	dkgUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
	exUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/utils"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// DisputeMissingKeySharesTask stores the data required to dispute shares
type DisputeMissingKeySharesTask struct {
	*objects.Task
}

// asserting that DisputeMissingKeySharesTask struct implements interface interfaces.Task
var _ interfaces.ITask = &DisputeMissingKeySharesTask{}

// NewDisputeMissingKeySharesTask creates a new task
func NewDisputeMissingKeySharesTask(dkgState *state.DkgState, start uint64, end uint64) *DisputeMissingKeySharesTask {
	return &DisputeMissingKeySharesTask{
		Task: objects.NewTask(dkgState, constants.DisputeMissingKeySharesTaskName, start, end),
	}
}

// Initialize begins the setup phase for DisputeMissingKeySharesTask.
func (t *DisputeMissingKeySharesTask) Initialize(ctx context.Context, logger *logrus.Entry, eth ethereum.Network) error {
	logger.Info("Initializing DisputeMissingKeySharesTask...")
	return nil
}

// DoWork is the first attempt at disputing distributed shares
func (t *DisputeMissingKeySharesTask) DoWork(ctx context.Context, logger *logrus.Entry, eth ethereum.Network) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is subsequent attempts at disputing distributed shares
func (t *DisputeMissingKeySharesTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth ethereum.Network) error {
	return t.doTask(ctx, logger, eth)
}

func (t *DisputeMissingKeySharesTask) doTask(ctx context.Context, logger *logrus.Entry, eth ethereum.Network) error {
	t.State.Lock()
	defer t.State.Unlock()

	taskState, ok := t.State.(*state.DkgState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	logger.Info("DisputeMissingKeySharesTask doTask()")

	accusableParticipants, err := t.getAccusableParticipants(ctx, eth, logger)
	if err != nil {
		return dkgUtils.LogReturnErrorf(logger, "DisputeMissingKeySharesTask doTask() error getting accusableParticipants: %v", err)
	}

	// accuse missing validators
	if len(accusableParticipants) > 0 {
		logger.Warnf("Accusing missing key shares: %v", accusableParticipants)

		txnOpts, err := eth.GetTransactionOpts(ctx, taskState.Account)
		if err != nil {
			return dkgUtils.LogReturnErrorf(logger, "DisputeMissingKeySharesTask doTask() error getting txnOpts: %v", err)
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
			return dkgUtils.LogReturnErrorf(logger, "DisputeMissingKeySharesTask doTask() error accusing missing key shares: %v", err)
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
		watcher := transaction.WatcherFromNetwork(eth)
		watcher.Subscribe(ctx, txn)
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
func (t *DisputeMissingKeySharesTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth ethereum.Network) bool {

	t.State.Lock()
	defer t.State.Unlock()

	taskState, ok := t.State.(*state.DkgState)
	if !ok {
		logger.Error("Invalid convertion of taskState object")
		return false
	}

	logger.Info("DisputeMissingKeySharesTask ShouldRetry()")

	generalRetry := exUtils.GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	if taskState.Phase != state.KeyShareSubmission {
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

func (t *DisputeMissingKeySharesTask) GetExecutionData() interfaces.ITaskExecutionData {
	return t.Task
}

func (t *DisputeMissingKeySharesTask) getAccusableParticipants(ctx context.Context, eth ethereum.Network, logger *logrus.Entry) ([]common.Address, error) {

	taskState, ok := t.State.(*state.DkgState)
	if !ok {
		return nil, objects.ErrCanNotContinue
	}

	var accusableParticipants []common.Address
	callOpts, err := eth.GetCallOpts(ctx, taskState.Account)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingShareDistributionTask failed getting call options: %v", err)
	}

	validators, err := dkgUtils.GetValidatorAddressesFromPool(callOpts, eth, logger)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingShareDistributionTask getAccusableParticipants() error getting validators: %v", err)
	}

	validatorsMap := make(map[common.Address]bool)
	for _, validator := range validators {
		validatorsMap[validator] = true
	}

	// find participants who did not submit they key shares
	for _, p := range taskState.Participants {
		_, isValidator := validatorsMap[p.Address]
		if isValidator && (p.Nonce != taskState.Nonce ||
			p.Phase != state.KeyShareSubmission ||
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
