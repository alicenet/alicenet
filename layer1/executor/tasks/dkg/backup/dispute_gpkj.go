package dkg

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/crypto/bn256"
	"github.com/alicenet/alicenet/layer1/ethereum"

	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/utils"
	"github.com/sirupsen/logrus"
)

// DisputeGPKjTask contains required state for performing a group accusation
type DisputeGPKjTask struct {
	*tasks.BaseTask
	// additional fields that are not part of the default task
	Address common.Address
}

// asserting that DisputeGPKjTask struct implements interface tasks.Task
var _ tasks.Task = &DisputeGPKjTask{}

// NewDisputeGPKjTask creates a background task that attempts perform a group accusation if necessary
func NewDisputeGPKjTask(start uint64, end uint64, address common.Address) *DisputeGPKjTask {
	return &DisputeGPKjTask{
		BaseTask: tasks.NewBaseTask(start, end, true, nil),
		Address:  address,
	}
}

// Prepare prepares for work to be done in the DisputeGPKjTask.
// Here, we determine if anyone submitted an invalid gpkj.
func (t *DisputeGPKjTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()").WithField("address", t.Address)
	logger.Debug("preparing task")
	return nil
}

// Execute executes the task business logic
func (t *DisputeGPKjTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()").WithField("address", t.Address)
	logger.Debug("initiate execution")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	if dkgState.Phase != state.DisputeGPKJSubmission && dkgState.Phase != state.GPKJSubmission {
		return nil, tasks.NewTaskErr("it's not DisputeGPKJSubmission or GPKJSubmission phase", false)
	}

	var (
		groupPublicKeys  [][4]*big.Int
		groupCommitments [][][2]*big.Int
	)

	var participantList = dkgState.GetSortedParticipants()

	for _, participant := range participantList {
		// Build array
		groupPublicKeys = append(groupPublicKeys, participant.GPKj)
		groupCommitments = append(groupCommitments, participant.Commitments)
	}

	honest, dishonest, missing, err := state.CategorizeGroupSigners(groupPublicKeys, participantList, groupCommitments)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("failed to determine honest vs dishonest validators: %v", err), false)
	}
	// we had a address that was not processed
	if len(honest)+len(missing)+len(dishonest) != len(participantList) {
		return nil, tasks.NewTaskErr(fmt.Sprintf("missing information when computing honest, dishonest, missing validators: %v", err), false)
	}

	for _, honestParticipant := range honest {
		if bytes.Equal(honestParticipant.Address.Bytes(), t.Address.Bytes()) {
			logger.Infof("finishing accusation task %v is honest", t.Address.Hex())
			return nil, nil
		}
	}

	for _, dishonestParticipant := range dishonest {
		if bytes.Equal(dishonestParticipant.Address.Bytes(), t.Address.Bytes()) {
			return t.accuseDishonestValidator(ctx, logger, dkgState, participantList, groupCommitments)
		}
	}

	logger.Tracef("finishing accusation task, %v is not dishonest ", t.Address.Hex())
	// address didn't shared it's gpkj, another task will be responsible for accusing him
	return nil, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *DisputeGPKjTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()").WithField("address", t.Address)
	logger.Debug("should execute task")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	if dkgState.Phase != state.DisputeGPKJSubmission {
		logger.Debugf("phase %v different from DisputeGPKJSubmission", dkgState.Phase)
		return false, nil
	}

	isValidator, err := utils.IsValidator(t.GetDB(), logger, t.Address)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingIsValidator, err), false)
	}
	logger.WithFields(logrus.Fields{"eth.badParticipant": t.Address.Hex()}).Debug("participant was not accused yet")

	return isValidator, nil
}

func (t *DisputeGPKjTask) accuseDishonestValidator(ctx context.Context, logger *logrus.Entry, dkgState *state.DkgState, participantList state.ParticipantList, groupCommitments [][][2]*big.Int) (*types.Transaction, *tasks.TaskErr) {
	var groupEncryptedSharesHash [][32]byte
	var validatorAddresses []common.Address

	for _, participant := range participantList {
		// Get group encrypted shares
		es := participant.EncryptedShares
		encryptedSharesBin, err := bn256.MarshalBigIntSlice(es)
		if err != nil {
			return nil, tasks.NewTaskErr(fmt.Sprintf("group accusation failed: %v", err), true)
		}
		hashSlice := crypto.Hasher(encryptedSharesBin)
		var hashSlice32 [32]byte
		copy(hashSlice32[:], hashSlice)
		groupEncryptedSharesHash = append(groupEncryptedSharesHash, hashSlice32)
		validatorAddresses = append(validatorAddresses, participant.Address)
	}

	client := t.GetClient()

	// Setup
	txnOpts, err := client.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingTxnOpts, err), true)
	}

	isValidator, err := utils.IsValidator(t.GetDB(), logger, t.Address)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingIsValidator, err), false)
	}

	// it means that the guys was already accused and evicted from the validatorPool
	if !isValidator {
		logger.Debugf("%v is not a validator anymore", t.Address.Hex())
		return nil, nil
	}

	logger.Warnf("accusing participant: %v of distributing bad dpkj", t.Address.Hex())
	txn, err := ethereum.GetContracts().Ethdkg().AccuseParticipantSubmittedBadGPKJ(txnOpts, validatorAddresses, groupEncryptedSharesHash, groupCommitments, t.Address)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("bad dpkj accusation failed: %v", err), true)
	}
	return txn, nil
}
