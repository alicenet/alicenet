package dkg

import (
	"crypto/rand"
	"fmt"
	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	dkgUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
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
		Task: objects.NewTask(constants.DisputeShareDistributionTaskName, start, end, false),
	}
}

// Prepare prepares for work to be done in the DisputeShareDistributionTask.
// It determines if the shares previously distributed are valid.
// If any are invalid, disputes will be issued.
func (t *DisputeShareDistributionTask) Prepare() error {
	logger := t.GetLogger()
	logger.Info("DisputeShareDistributionTask Prepare()...")

	dkgState := &state.DkgState{}
	err := t.GetDB().Update(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, logger)
		if err != nil {
			return err
		}

		if dkgState.Phase != state.DisputeShareDistribution && dkgState.Phase != state.ShareDistribution {
			return fmt.Errorf("%w because it's not DisputeShareDistribution phase", objects.ErrCanNotContinue)
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
				return fmt.Errorf("VerifyDistributedShares broke: %v; %v", err.Error(), objects.ErrCanNotContinue)
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

		err = dkgState.PersistState(txn, logger)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return dkgUtils.LogReturnErrorf(logger, "DisputeShareDistributionTask.Prepare(): error during the preparation: %v", err)
	}

	return nil
}

// Execute executes the task business logic
func (t *DisputeShareDistributionTask) Execute() ([]*types.Transaction, error) {
	logger := t.GetLogger()
	logger.Info("DisputeShareDistributionTask doTask()")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, t.GetLogger())
		return err
	})
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeMissingGPKjTask.Execute(): error loading dkgState: %v", err)
	}

	ctx := t.GetCtx()
	eth := t.GetEth()
	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeShareDistribution.Execute() failed getting call options: %v", err)
	}

	txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "DisputeShareDistribution.Execute() failed getting txn opts: %v", err)
	}

	txns := make([]*types.Transaction, 0)
	for _, participant := range dkgState.BadShares {
		isValidator, err := eth.Contracts().ValidatorPool().IsValidator(callOpts, participant.Address)
		if err != nil {
			return nil, dkgUtils.LogReturnErrorf(logger, "getting isValidator failed: %v", err)
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
			return nil, err
		}
		sharedKeyG1 := cloudflare.GenerateSharedSecretG1(dkgState.TransportPrivateKey, disputePublicKeyG1)
		sharedKey, err := bn256.G1ToBigIntArray(sharedKeyG1)
		if err != nil {
			return nil, err
		}

		// Construct shared key proof
		g1Base := new(cloudflare.G1).ScalarBaseMult(common.Big1)
		transportPublicKeyG1 := new(cloudflare.G1).ScalarBaseMult(dkgState.TransportPrivateKey)
		sharedKeyProof, err := cloudflare.GenerateDLEQProofG1(
			g1Base, transportPublicKeyG1, disputePublicKeyG1, sharedKeyG1, dkgState.TransportPrivateKey, rand.Reader)
		if err != nil {
			return nil, err
		}

		// Accuse participant
		txn, err := eth.Contracts().Ethdkg().AccuseParticipantDistributedBadShares(txnOpts, dishonestAddress, encryptedShares, commitments, sharedKey, sharedKeyProof)
		if err != nil {
			return nil, dkgUtils.LogReturnErrorf(logger, "submit share dispute failed: %v", err)
		}
		txns = append(txns, txn)
	}

	return txns, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *DisputeShareDistributionTask) ShouldExecute() bool {
	logger := t.GetLogger()
	logger.Info("DisputeShareDistributionTask ShouldExecute()")

	ctx := t.GetCtx()
	eth := t.GetEth()
	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, t.GetLogger())
		return err
	})
	if err != nil {
		logger.Errorf("DisputeShareDistributionTask.ShouldExecute(): error loading dkgState: %v", err)
		return true
	}

	if dkgState.Phase != state.DisputeShareDistribution {
		return false
	}

	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		logger.Error(fmt.Sprintf("DisputeShareDistribution.ShouldExecute() could not get call options: %v", err))
		return true
	}
	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	if err != nil {
		logger.Error(fmt.Sprintf("DisputeShareDistribution.ShouldExecute() could not get BadParticipants: %v", err))
		return true
	}

	// if there is someone that wasn't accused we need to retry
	return len(dkgState.BadShares) != int(badParticipants.Int64())
}
