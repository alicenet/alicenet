package dkg

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/MadBase/MadNet/layer1/ethereum"
	"github.com/MadBase/MadNet/layer1/executor/constants"
	"github.com/MadBase/MadNet/layer1/executor/tasks"
	"github.com/MadBase/MadNet/layer1/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/layer1/transaction"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// DisputeShareDistributionTask stores the data required to dispute shares
type DisputeShareDistributionTask struct {
	*tasks.BaseTask
}

// asserting that DisputeShareDistributionTask struct implements interface tasks.Task
var _ tasks.Task = &DisputeShareDistributionTask{}

// NewDisputeShareDistributionTask creates a new task
func NewDisputeShareDistributionTask(start uint64, end uint64) *DisputeShareDistributionTask {
	return &DisputeShareDistributionTask{
		BaseTask: tasks.NewBaseTask(constants.DisputeShareDistributionTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the DisputeShareDistributionTask.
// It determines if the shares previously distributed are valid.
// If any are invalid, disputes will be issued.
func (t *DisputeShareDistributionTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debug("preparing task")

	dkgState := &state.DkgState{}
	err := t.GetDB().Update(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		if err != nil {
			return err
		}

		if dkgState.Phase != state.DisputeShareDistribution && dkgState.Phase != state.ShareDistribution {
			return fmt.Errorf("it's not DisputeShareDistribution or ShareDistribution phase")
		}

		var participantsList = dkgState.GetSortedParticipants()
		// Loop through all participants and check to see if shares are valid
		for idx := 0; idx < dkgState.NumberOfValidators; idx++ {
			participant := participantsList[idx]

			var emptyHash [32]byte
			if participant.DistributedSharesHash == emptyHash {
				continue
			}

			logger.Debugf("participant idx: %v:%v:%v\n", idx, participant.Index, dkgState.Index)
			valid, present, err := state.VerifyDistributedShares(dkgState, participant)
			if err != nil {
				// A major error occurred; we cannot continue
				return fmt.Errorf("VerifyDistributedShares broke: %v Participant Address: %v", err.Error(), participant.Address.Hex())
			}
			if !present {
				logger.Warnf("No share from %v", participant.Address.Hex())
				continue
			}
			if !valid {
				logger.Warnf("Invalid share from %v", participant.Address.Hex())
				dkgState.BadShares[participant.Address] = participant
			}
		}

		err = dkgState.PersistState(txn)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// all errors are not recoverable
		return tasks.NewTaskErr(fmt.Sprintf(constants.ErrorDuringPreparation, err), false)
	}

	return nil
}

// Execute executes the task business logic
func (t *DisputeShareDistributionTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()
	callOpts, err := client.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(constants.FailedGettingCallOpts, err), true)
	}

	txnOpts, err := client.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(constants.FailedGettingTxnOpts, err), true)
	}

	txns := make([]*types.Transaction, 0)
	for _, participant := range dkgState.BadShares {
		isValidator, err := ethereum.GetContracts().ValidatorPool().IsValidator(callOpts, participant.Address)
		if err != nil {
			return nil, tasks.NewTaskErr(fmt.Sprintf(constants.FailedGettingIsValidator, err), true)
		}

		if !isValidator {
			continue
		}

		dishonestAddress := participant.Address
		encryptedShares := dkgState.Participants[participant.Address].EncryptedShares
		commitments := dkgState.Participants[participant.Address].Commitments

		// Construct shared key
		disputePublicKeyG1, err := bn256.BigIntArrayToG1(participant.PublicKey)
		if err != nil {
			return nil, tasks.NewTaskErr(fmt.Sprintf("failed generating disputePublicKeyG1: %v", err), false)
		}
		sharedKeyG1 := cloudflare.GenerateSharedSecretG1(dkgState.TransportPrivateKey, disputePublicKeyG1)
		sharedKey, err := bn256.G1ToBigIntArray(sharedKeyG1)
		if err != nil {
			return nil, tasks.NewTaskErr(fmt.Sprintf("failed generating sharedKeyG1: %v", err), false)
		}

		// Construct shared key proof
		g1Base := new(cloudflare.G1).ScalarBaseMult(common.Big1)
		transportPublicKeyG1 := new(cloudflare.G1).ScalarBaseMult(dkgState.TransportPrivateKey)
		sharedKeyProof, err := cloudflare.GenerateDLEQProofG1(
			g1Base, transportPublicKeyG1, disputePublicKeyG1, sharedKeyG1, dkgState.TransportPrivateKey, rand.Reader)
		if err != nil {
			return nil, tasks.NewTaskErr(fmt.Sprintf("failed generating sharedKeyProof: %v", err), false)
		}

		logger.Warnf("accusing participant: %v of distributing bad shares", dishonestAddress)
		// Accuse participant
		txn, err := ethereum.GetContracts().Ethdkg().AccuseParticipantDistributedBadShares(txnOpts, dishonestAddress, encryptedShares, commitments, sharedKey, sharedKeyProof)
		if err != nil {
			return nil, tasks.NewTaskErr(fmt.Sprintf("submit share dispute failed: %v", err), true)
		}
		txns = append(txns, txn)
	}
	//todo: fix this, split this task in multiple tasks
	return txns[0], nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *DisputeShareDistributionTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	client := t.GetClient()
	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	if dkgState.Phase != state.DisputeShareDistribution {
		logger.Debugf("phase %v different from DisputeShareDistribution", dkgState.Phase)
		return false, nil
	}

	callOpts, err := client.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(constants.FailedGettingCallOpts, err), true)
	}
	badParticipants, err := ethereum.GetContracts().Ethdkg().GetBadParticipants(callOpts)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf("could not get BadParticipants: %v", err), true)
	}

	if len(dkgState.BadShares) == int(badParticipants.Int64()) {
		logger.Debug("all bad participants already accused")
		return false, nil
	}

	return true, nil
}
