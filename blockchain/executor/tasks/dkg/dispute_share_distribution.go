package dkg

import (
	"crypto/rand"
	"fmt"

	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	executorInterfaces "github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// DisputeShareDistributionTask stores the data required to dispute shares
type DisputeShareDistributionTask struct {
	*objects.Task
}

// asserting that DisputeShareDistributionTask struct implements interface interfaces.Task
var _ interfaces.ITask = &DisputeShareDistributionTask{}

// NewDisputeShareDistributionTask creates a new task
func NewDisputeShareDistributionTask(start uint64, end uint64) *DisputeShareDistributionTask {
	return &DisputeShareDistributionTask{
		Task: objects.NewTask(constants.DisputeShareDistributionTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the DisputeShareDistributionTask.
// It determines if the shares previously distributed are valid.
// If any are invalid, disputes will be issued.
func (t *DisputeShareDistributionTask) Prepare() *executorInterfaces.TaskErr {
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

		if dkgState.Phase != state.DisputeShareDistribution && dkgState.Phase != state.ShareDistribution {
			isRecoverable = false
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

			logger.Infof("participant idx: %v:%v:%v\n", idx, participant.Index, dkgState.Index)
			valid, present, err := state.VerifyDistributedShares(dkgState, participant)
			if err != nil {
				// A major error occurred; we cannot continue
				logger.Errorf("VerifyDistributedShares broke; Participant Address: %v", participant.Address.Hex())
				isRecoverable = false
				return fmt.Errorf("VerifyDistributedShares broke: %v", err.Error())
			}
			if !present {
				logger.Warningf("No share from %v", participant.Address.Hex())
				continue
			}
			if !valid {
				logger.Warningf("Invalid share from %v", participant.Address.Hex())
				dkgState.BadShares[participant.Address] = participant
			}
		}

		err = dkgState.PersistState(txn)
		if err != nil {
			isRecoverable = false
			return err
		}

		return nil
	})

	if err != nil {
		return executorInterfaces.NewTaskErr(fmt.Sprintf(constants.ErrorDuringPreparation, err), isRecoverable)
	}

	return nil
}

// Execute executes the task business logic
func (t *DisputeShareDistributionTask) Execute() ([]*types.Transaction, *executorInterfaces.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Trace("initiate execution")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, executorInterfaces.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	ctx := t.GetCtx()
	eth := t.GetClient()
	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, executorInterfaces.NewTaskErr(fmt.Sprintf(constants.FailedGettingCallOpts, err), true)
	}

	txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, executorInterfaces.NewTaskErr(fmt.Sprintf(constants.FailedGettingTxnOpts, err), true)
	}

	txns := make([]*types.Transaction, 0)
	for _, participant := range dkgState.BadShares {
		isValidator, err := eth.Contracts().ValidatorPool().IsValidator(callOpts, participant.Address)
		if err != nil {
			return nil, executorInterfaces.NewTaskErr(fmt.Sprintf(constants.FailedGettingIsValidator, err), true)
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
			return nil, executorInterfaces.NewTaskErr(fmt.Sprintf("failed generating disputePublicKeyG1: %v", err), true)
		}
		sharedKeyG1 := cloudflare.GenerateSharedSecretG1(dkgState.TransportPrivateKey, disputePublicKeyG1)
		sharedKey, err := bn256.G1ToBigIntArray(sharedKeyG1)
		if err != nil {
			return nil, executorInterfaces.NewTaskErr(fmt.Sprintf("failed generating sharedKeyG1: %v", err), true)
		}

		// Construct shared key proof
		g1Base := new(cloudflare.G1).ScalarBaseMult(common.Big1)
		transportPublicKeyG1 := new(cloudflare.G1).ScalarBaseMult(dkgState.TransportPrivateKey)
		sharedKeyProof, err := cloudflare.GenerateDLEQProofG1(
			g1Base, transportPublicKeyG1, disputePublicKeyG1, sharedKeyG1, dkgState.TransportPrivateKey, rand.Reader)
		if err != nil {
			return nil, executorInterfaces.NewTaskErr(fmt.Sprintf("failed generating sharedKeyProof: %v", err), true)
		}

		// Accuse participant
		txn, err := eth.Contracts().Ethdkg().AccuseParticipantDistributedBadShares(txnOpts, dishonestAddress, encryptedShares, commitments, sharedKey, sharedKeyProof)
		if err != nil {
			return nil, executorInterfaces.NewTaskErr(fmt.Sprintf("submit share dispute failed: %v", err), true)
		}
		txns = append(txns, txn)
	}

	return txns, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *DisputeShareDistributionTask) ShouldExecute() *executorInterfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Trace("should execute task")

	ctx := t.GetCtx()
	eth := t.GetClient()
	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return executorInterfaces.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	if dkgState.Phase != state.DisputeShareDistribution {
		return executorInterfaces.NewTaskErr(fmt.Sprintf("phase %v different from DisputeShareDistribution", dkgState.Phase), false)
	}

	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return executorInterfaces.NewTaskErr(fmt.Sprintf(constants.FailedGettingCallOpts, err), true)
	}
	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	if err != nil {
		return executorInterfaces.NewTaskErr(fmt.Sprintf("could not get BadParticipants: %v", err), true)
	}

	// if there is someone that wasn't accused we need to retry
	if len(dkgState.BadShares) == int(badParticipants.Int64()) {
		return executorInterfaces.NewTaskErr(fmt.Sprintf("all bad participants already accused"), false)
	}

	return nil
}
