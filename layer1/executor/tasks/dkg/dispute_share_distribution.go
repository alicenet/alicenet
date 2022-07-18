package dkg

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"

	"github.com/alicenet/alicenet/crypto/bn256"
	"github.com/alicenet/alicenet/crypto/bn256/cloudflare"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// DisputeShareDistributionTask stores the data required to dispute shares
type DisputeShareDistributionTask struct {
	*tasks.BaseTask
	// additional fields that are not part of the default task
	Address common.Address
}

// asserting that DisputeShareDistributionTask struct implements interface tasks.Task
var _ tasks.Task = &DisputeShareDistributionTask{}

// NewDisputeShareDistributionTask creates a new task
func NewDisputeShareDistributionTask(start uint64, end uint64, address common.Address) *DisputeShareDistributionTask {
	return &DisputeShareDistributionTask{
		BaseTask: tasks.NewBaseTask(start, end, true, nil),
		Address:  address,
	}
}

// Prepare prepares for work to be done in the DisputeShareDistributionTask. It
// determines if the shares previously distributed are valid. If any are
// invalid, disputes will be issued.
func (t *DisputeShareDistributionTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()").WithField("address", t.Address)
	logger.Debug("preparing task")
	return nil
}

// Execute executes the task business logic
func (t *DisputeShareDistributionTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()").WithField("address", t.Address)
	logger.Debug("initiate execution")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorDuringPreparation, err), false)
	}

	if dkgState.Phase != state.DisputeShareDistribution && dkgState.Phase != state.ShareDistribution {
		return nil, tasks.NewTaskErr("it's not DisputeShareDistribution or ShareDistribution phase", false)
	}

	isValidator, err := utils.IsValidator(t.GetDB(), logger, t.Address)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingIsValidator, err), false)
	}

	if !isValidator {
		logger.Debugf("%v is not a validator anymore", t.Address.Hex())
		return nil, nil
	}

	var participantsList = dkgState.GetSortedParticipants()
	var participantState *state.Participant
	// Loop through all participants and check to see if shares are valid
	for idx := 0; idx < dkgState.NumberOfValidators; idx++ {
		participant := participantsList[idx]
		if bytes.Equal(participant.Address.Bytes(), t.Address.Bytes()) {
			participantState = participant
		}
	}

	if participantState == nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("couldn't find %v in the dkgState ParticipantList", t.Address.Hex()), false)
	}

	valid, present, err := state.VerifyDistributedShares(dkgState, participantState)
	if err != nil {
		// A major error occurred; we cannot continue
		return nil, tasks.NewTaskErr(
			fmt.Sprintf("VerifyDistributedShares broke: %v Participant Address: %v", err.Error(), participantState.Address.Hex()), false,
		)
	}
	// another task will accuse the guy of not participating
	if !present {
		logger.Debugf("No share from %v", participantState.Address.Hex())
		return nil, nil
	}
	if valid {
		logger.Infof("honest participant %v", participantState.Address.Hex())
		return nil, nil
	}

	dishonestAddress := t.Address
	encryptedShares := dkgState.Participants[t.Address].EncryptedShares
	commitments := dkgState.Participants[t.Address].Commitments

	// Construct shared key
	disputePublicKeyG1, err := bn256.BigIntArrayToG1(participantState.PublicKey)
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
	client := t.GetClient()
	txnOpts, err := client.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingTxnOpts, err), true)
	}
	// Accuse participant
	txn, err := ethereum.GetContracts().Ethdkg().AccuseParticipantDistributedBadShares(txnOpts, dishonestAddress, encryptedShares, commitments, sharedKey, sharedKeyProof)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("submit share dispute failed: %v", err), true)
	}

	return txn, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *DisputeShareDistributionTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()").WithField("address", t.Address)
	logger.Debug("should execute task")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	if dkgState.Phase != state.DisputeShareDistribution {
		logger.Debugf("phase %v different from DisputeShareDistribution", dkgState.Phase)
		return false, nil
	}

	isValidator, err := utils.IsValidator(t.GetDB(), logger, t.Address)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingIsValidator, err), false)
	}
	logger.WithFields(logrus.Fields{"eth.badParticipant": t.Address.Hex()}).Debug("participant was not accused yet")

	return isValidator, nil
}
