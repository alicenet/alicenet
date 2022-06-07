package dkg

import (
	"fmt"
	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	dkgUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
	exUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
)

// ShareDistributionTask stores the state required safely distribute shares
type ShareDistributionTask struct {
	*objects.Task
}

// asserting that ShareDistributionTask struct implements interface interfaces.Task
var _ interfaces.ITask = &ShareDistributionTask{}

// NewShareDistributionTask creates a new task
func NewShareDistributionTask(start uint64, end uint64) *ShareDistributionTask {
	return &ShareDistributionTask{
		Task: objects.NewTask(constants.ShareDistributionTaskName, start, end),
	}
}

// Prepare prepares for work to be done in the ShareDistributionTask.
// We construct our commitments and encrypted shares before
// submitting them to the associated smart contract.
func (t *ShareDistributionTask) Prepare() error {
	logger := t.GetLogger()
	logger.Infof("ShareDistributionTask Prepare()")

	dkgState := &state.DkgState{}
	err := t.GetDB().Update(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, logger)
		if err != nil {
			return err
		}

		if dkgState.Phase != state.ShareDistribution {
			return fmt.Errorf("%w because it's not in ShareDistribution phase", objects.ErrCanNotContinue)
		}

		if dkgState.SecretValue == nil {

			participants := dkgState.GetSortedParticipants()
			numParticipants := len(participants)
			threshold := state.ThresholdForUserCount(numParticipants)

			// Generate shares
			encryptedShares, privateCoefficients, commitments, err := state.GenerateShares(
				dkgState.TransportPrivateKey, participants)
			if err != nil {
				logger.Errorf("Failed to generate shares: %v %#v", err, participants)
				return err
			}

			// Store calculated values
			dkgState.Participants[dkgState.Account.Address].Commitments = commitments
			dkgState.Participants[dkgState.Account.Address].EncryptedShares = encryptedShares

			dkgState.PrivateCoefficients = privateCoefficients
			dkgState.SecretValue = privateCoefficients[0]
			dkgState.ValidatorThreshold = threshold

			err = dkgState.PersistState(txn, logger)
			if err != nil {
				return err
			}
		} else {
			logger.Infof("ShareDistributionTask Prepare(): encrypted shares already defined")
		}

		return nil
	})

	if err != nil {
		return dkgUtils.LogReturnErrorf(logger, "ShareDistributionTask.Prepare(): error during the preparation: %v", err)
	}

	return nil
}

// Execute executes the task business logic
func (t *ShareDistributionTask) Execute() ([]*types.Transaction, error) {
	logger := t.GetLogger()
	logger.Info("ShareDistributionTask doTask()")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, logger)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("RegisterTask.Execute(): error loading dkgState: %v", err)
	}

	eth := t.GetEth()
	ctx := t.GetCtx()
	c := eth.Contracts()
	me := dkgState.Account.Address
	logger.Debugf("me:%v", me.Hex())

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// Distribute shares
	txn, err := c.Ethdkg().DistributeShares(txnOpts, dkgState.Participants[me].EncryptedShares, dkgState.Participants[me].Commitments)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "distributing shares failed: %v", err)
	}

	return []*types.Transaction{txn}, nil
}

// ShouldRetry checks if it makes sense to try again
func (t *ShareDistributionTask) ShouldExecute() bool {
	logger := t.GetLogger()
	logger.Info("ShareDistributionTask ShouldExecute()")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, t.GetLogger())
		return err
	})
	if err != nil {
		logger.Errorf("could not get dkgState with error %v", err)
		return true
	}

	eth := t.GetEth()
	ctx := t.GetCtx()
	generalRetry := exUtils.GeneralTaskShouldRetry(ctx, logger, eth, t.GetStart(), t.GetEnd())
	if !generalRetry {
		return false
	}

	if dkgState.Phase != state.ShareDistribution {
		return false
	}

	// If it's generally good to retry, let's try to be more specific
	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		logger.Errorf("ShareDistributionTask.ShoudRetry() failed getting call options: %v", err)
		return true
	}
	participantState, err := eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, dkgState.Account.Address)
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
