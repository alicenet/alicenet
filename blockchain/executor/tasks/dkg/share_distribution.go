package dkg

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/blockchain/transaction"
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
		Task: objects.NewTask(constants.ShareDistributionTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the ShareDistributionTask.
// We construct our commitments and encrypted shares before
// submitting them to the associated smart contract.
func (t *ShareDistributionTask) Prepare() *interfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Tracef("preparing task")

	dkgState := &state.DkgState{}
	var isRecoverable bool
	err := t.GetDB().Update(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		if err != nil {
			isRecoverable = false
			return err
		}

		if dkgState.Phase != state.ShareDistribution {
			isRecoverable = false
			return errors.New("not in ShareDistribution phase")
		}

		if dkgState.SecretValue == nil {

			participants := dkgState.GetSortedParticipants()
			numParticipants := len(participants)
			threshold := state.ThresholdForUserCount(numParticipants)

			// Generate shares
			encryptedShares, privateCoefficients, commitments, err := state.GenerateShares(
				dkgState.TransportPrivateKey, participants)
			if err != nil {
				isRecoverable = true
				return fmt.Errorf("Failed to generate shares: %v %#v", err, participants)
			}

			// Store calculated values
			dkgState.Participants[dkgState.Account.Address].Commitments = commitments
			dkgState.Participants[dkgState.Account.Address].EncryptedShares = encryptedShares

			dkgState.PrivateCoefficients = privateCoefficients
			dkgState.SecretValue = privateCoefficients[0]
			dkgState.ValidatorThreshold = threshold

			err = dkgState.PersistState(txn)
			if err != nil {
				isRecoverable = false
				return err
			}
		} else {
			logger.Infof("ShareDistributionTask Prepare(): encrypted shares already defined")
		}

		return nil
	})

	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf("error during the preparation: %v", err), isRecoverable)
	}

	return nil
}

// Execute executes the task business logic
func (t *ShareDistributionTask) Execute() ([]*types.Transaction, *interfaces.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Trace("initiate execution")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf("error loading dkgState: %v", err), false)
	}

	client := t.GetClient()
	ctx := t.GetCtx()
	contracts := client.Contracts()
	accountAddr := dkgState.Account.Address

	// Setup
	txnOpts, err := client.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf("getting txn opts failed: %v", err), true)
	}

	// Distribute shares
	txn, err := contracts.Ethdkg().DistributeShares(
		txnOpts,
		dkgState.Participants[accountAddr].EncryptedShares,
		dkgState.Participants[accountAddr].Commitments,
	)
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf("distributing shares failed: %v", err), true)
	}

	return []*types.Transaction{txn}, nil
}

// ShouldRetry checks if it makes sense to try again
func (t *ShareDistributionTask) ShouldExecute() *interfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Trace("should execute task")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf("could not get dkgState with error %v", err), false)
	}

	eth := t.GetClient()
	ctx := t.GetCtx()
	if dkgState.Phase != state.ShareDistribution {
		return interfaces.NewTaskErr(fmt.Sprintf("phase %v different from ShareDistribution", dkgState.Phase), false)
	}

	// If it's generally good to retry, let's try to be more specific
	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf("failed getting call options: %v", err), true)
	}
	participantState, err := eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, dkgState.Account.Address)
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf("unable to GetParticipantInternalState(): %v", err), true)
	}

	logger.Infof("DistributionHash: %x", participantState.DistributedSharesHash)
	var emptySharesHash [32]byte
	if !bytes.Equal(participantState.DistributedSharesHash[:], emptySharesHash[:]) {
		return interfaces.NewTaskErr("did distribute shares after all. needs no retry", false)
	}

	logger.Debugf("Did not distribute shares after all. needs retry")
	return nil
}
