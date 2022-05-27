package dkgtasks

import (
	"context"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/objects"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/utils"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// DisputeMissingRegistrationTask contains required state for accusing missing registrations
type DisputeMissingRegistrationTask struct {
	*objects.Task
}

// asserting that DisputeMissingRegistrationTask struct implements interface interfaces.Task
var _ interfaces.ITask = &DisputeMissingRegistrationTask{}

// NewDisputeMissingRegistrationTask creates a background task to accuse missing registrations during ETHDKG
func NewDisputeMissingRegistrationTask(state *objects.DkgState, start uint64, end uint64) *DisputeMissingRegistrationTask {
	return &DisputeMissingRegistrationTask{
		Task: objects.NewTask(state, DisputeMissingRegistrationTaskName, start, end),
	}
}

// Initialize begins the setup phase for Dispute Registration.
func (t *DisputeMissingRegistrationTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	logger.Info("DisputeMissingRegistrationTask Initializing...")
	return nil
}

// DoWork is the first attempt at Disputing Missing Registrations with ethdkg
func (t *DisputeMissingRegistrationTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at registering with ethdkg
func (t *DisputeMissingRegistrationTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *DisputeMissingRegistrationTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	taskState, ok := t.State.(*objects.DkgState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	logger.Info("DisputeMissingRegistrationTask doTask()")

	accusableParticipants, err := t.getAccusableParticipants(ctx, eth, logger)
	if err != nil {
		return utils.LogReturnErrorf(logger, "DisputeMissingRegistrationTask doTask() error getting accusable participants: %v", err)
	}

	// accuse missing validators
	if len(accusableParticipants) > 0 {
		logger.Warnf("Accusing missing registrations: %v", accusableParticipants)

		txnOpts, err := eth.GetTransactionOpts(ctx, taskState.Account)
		if err != nil {
			return utils.LogReturnErrorf(logger, "DisputeMissingRegistrationTask doTask() error getting txnOpts: %v", err)
		}

		// If the TxOpts exists, meaning the Tx replacement timeout was reached,
		// we increase the Gas to have priority for the next blocks
		if t.TxOpts != nil && t.TxOpts.Nonce != nil {
			logger.Info("txnOpts Replaced")
			txnOpts.Nonce = t.TxOpts.Nonce
			txnOpts.GasFeeCap = t.TxOpts.GasFeeCap
			txnOpts.GasTipCap = t.TxOpts.GasTipCap
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantNotRegistered(txnOpts, accusableParticipants)
		if err != nil {
			return utils.LogReturnErrorf(logger, "DisputeMissingRegistrationTask doTask() error accusing missing registration: %v", err)
		}
		t.TxOpts.TxHashes = append(t.TxOpts.TxHashes, txn.Hash())
		t.TxOpts.GasFeeCap = txn.GasFeeCap()
		t.TxOpts.GasTipCap = txn.GasTipCap()
		t.TxOpts.Nonce = big.NewInt(int64(txn.Nonce()))

		logger.WithFields(logrus.Fields{
			"GasFeeCap": t.TxOpts.GasFeeCap,
			"GasTipCap": t.TxOpts.GasTipCap,
			"Nonce":     t.TxOpts.Nonce,
		}).Info("missing registration dispute fees")

		// Queue transaction
		eth.Queue().QueueTransaction(ctx, txn)
	} else {
		logger.Info("No accusations for missing registrations")
	}

	t.Success = true
	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
// -- the registration open hasn't moved, i.e. ETHDKG has not restarted
// -- We have unregistered participants
func (t *DisputeMissingRegistrationTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("DisputeMissingRegistrationTask ShouldRetry()")

	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	taskState, ok := t.State.(*objects.DkgState)
	if !ok {
		logger.Error("Invalid convertion of taskState object")
		return false
	}

	if taskState.Phase != objects.RegistrationOpen {
		return false
	}

	accusableParticipants, err := t.getAccusableParticipants(ctx, eth, logger)
	if err != nil {
		logger.Errorf("DisputeMissingRegistrationTask ShouldRetry() error getting accusable participants: %v", err)
		return true
	}

	if len(accusableParticipants) > 0 {
		return true
	}

	return false
}

// DoDone just creates a log entry saying task is complete
func (t *DisputeMissingRegistrationTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("DisputeMissingRegistrationTask done")
}

func (t *DisputeMissingRegistrationTask) GetExecutionData() interfaces.ITaskExecutionData {
	return t.Task
}

func (t *DisputeMissingRegistrationTask) getAccusableParticipants(ctx context.Context, eth interfaces.Ethereum, logger *logrus.Entry) ([]common.Address, error) {

	taskState, ok := t.State.(*objects.DkgState)
	if !ok {
		return nil, objects.ErrCanNotContinue
	}

	var accusableParticipants []common.Address
	callOpts, err := eth.GetCallOpts(ctx, taskState.Account)
	if err != nil {
		return nil, utils.LogReturnErrorf(logger, "DisputeMissingRegistrationTask failed getting call options: %v", err)
	}

	validators, err := utils.GetValidatorAddressesFromPool(callOpts, eth, logger)
	if err != nil {
		return nil, utils.LogReturnErrorf(logger, "DisputeMissingRegistrationTask getAccusableParticipants() error getting validators: %v", err)
	}

	validatorsMap := make(map[common.Address]bool)
	for _, validator := range validators {
		validatorsMap[validator] = true
	}

	// find participants who did not register
	for _, addr := range taskState.ValidatorAddresses {

		participant, ok := taskState.Participants[addr]
		_, isValidator := validatorsMap[addr]

		if isValidator && (!ok ||
			participant.Nonce != taskState.Nonce ||
			participant.Phase != objects.RegistrationOpen ||
			(participant.PublicKey[0].Cmp(big.NewInt(0)) == 0 &&
				participant.PublicKey[1].Cmp(big.NewInt(0)) == 0)) {

			// did not register
			accusableParticipants = append(accusableParticipants, addr)
		}
	}

	return accusableParticipants, nil
}
