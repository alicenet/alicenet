package dkg

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/layer1/ethereum"
	exConstants "github.com/MadBase/MadNet/layer1/executor/constants"
	"github.com/MadBase/MadNet/layer1/executor/tasks"
	"github.com/MadBase/MadNet/layer1/executor/tasks/dkg/state"
	monInterfaces "github.com/MadBase/MadNet/layer1/monitor/interfaces"
	"github.com/MadBase/MadNet/layer1/transaction"
)

// GPKjSubmissionTask contains required state for gpk submission
type GPKjSubmissionTask struct {
	*tasks.BaseTask
	adminHandler monInterfaces.AdminHandler
}

// asserting that GPKjSubmissionTask struct implements interface Task
var _ tasks.Task = &GPKjSubmissionTask{}

// NewGPKjSubmissionTask creates a background task that attempts to submit the gpkj in ETHDKG
func NewGPKjSubmissionTask(start uint64, end uint64, adminHandler monInterfaces.AdminHandler) *GPKjSubmissionTask {
	return &GPKjSubmissionTask{
		BaseTask:     tasks.NewBaseTask(exConstants.GPKjSubmissionTaskName, start, end, false, transaction.NewSubscribeOptions(true, exConstants.ETHDKGMaxStaleBlocks)),
		adminHandler: adminHandler,
	}
}

// Prepare prepares for work to be done in the GPKjSubmissionTask
func (t *GPKjSubmissionTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debug("preparing task")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf(exConstants.ErrorDuringPreparation, err), false)
	}

	if dkgState.GroupPrivateKey == nil ||
		dkgState.GroupPrivateKey.Cmp(big.NewInt(0)) == 0 {

		// Collecting all the participants encrypted shares to be used for the GPKj
		var participantsList = dkgState.GetSortedParticipants()
		encryptedShares := make([][]*big.Int, 0, dkgState.NumberOfValidators)
		for _, participant := range participantsList {
			logger.Tracef(
				"Collecting encrypted shares... Participant %v %v",
				participant.Index,
				participant.Address.Hex(),
			)
			encryptedShares = append(encryptedShares, participant.EncryptedShares)
		}

		// Generate the GPKj
		groupPrivateKey, groupPublicKey, err := state.GenerateGroupKeys(
			dkgState.TransportPrivateKey, dkgState.PrivateCoefficients,
			encryptedShares, dkgState.Index, participantsList)
		if err != nil {
			return tasks.NewTaskErr(
				fmt.Sprintf("Could not generate group keys: %v for index: %v", err, dkgState.Index), false,
			)
		}

		dkgState.GroupPrivateKey = groupPrivateKey
		dkgState.Participants[dkgState.Account.Address].GPKj = groupPublicKey

		// Pass private key on to consensus
		logger.Debugf("Adding private bn256eth key... using %p", t.adminHandler)
		err = t.adminHandler.AddPrivateKey(groupPrivateKey.Bytes(), constants.CurveBN256Eth)
		if err != nil {
			return tasks.NewTaskErr(fmt.Sprintf("error adding private key: %v", err), true)
		}

		err = state.SaveDkgState(t.GetDB(), dkgState)
		if err != nil {
			return tasks.NewTaskErr(fmt.Sprintf(exConstants.ErrorDuringPreparation, err), false)
		}
	} else {
		logger.Debugf("group private-public key already defined")
	}

	return nil
}

// Execute executes the task business logic
func (t *GPKjSubmissionTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(exConstants.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()

	txnOpts, err := client.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(exConstants.FailedGettingTxnOpts, err), true)
	}

	logger.Infof("submitting gpkj: %v", dkgState.Participants[dkgState.Account.Address].GPKj)
	txn, err := ethereum.GetContracts().Ethdkg().SubmitGPKJ(txnOpts, dkgState.Participants[dkgState.Account.Address].GPKj)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("submitting gpkj failed: %v", err), true)
	}

	return txn, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *GPKjSubmissionTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(exConstants.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()
	if dkgState.Phase != state.GPKJSubmission {
		logger.Debugf("phase %v different from GPKJSubmission", dkgState.Phase)
		return false, nil
	}

	//Check if my GPKj is submitted, if not should retry
	defaultAddr := dkgState.Account
	callOpts, err := client.GetCallOpts(ctx, defaultAddr)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(exConstants.FailedGettingCallOpts, err), true)
	}
	participantState, err := ethereum.GetContracts().Ethdkg().GetParticipantInternalState(callOpts, defaultAddr.Address)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf("failed getting participants state: %v", err), true)
	}
	if participantState.Gpkj[0].Cmp(dkgState.Participants[defaultAddr.Address].GPKj[0]) == 0 &&
		participantState.Gpkj[1].Cmp(dkgState.Participants[defaultAddr.Address].GPKj[1]) == 0 &&
		participantState.Gpkj[2].Cmp(dkgState.Participants[defaultAddr.Address].GPKj[2]) == 0 &&
		participantState.Gpkj[3].Cmp(dkgState.Participants[defaultAddr.Address].GPKj[3]) == 0 {
		logger.Debug("GPKj already set")
		return false, nil
	}

	return true, nil
}

// SetAdminHandler sets the task adminHandler
func (t *GPKjSubmissionTask) SetAdminHandler(adminHandler monInterfaces.AdminHandler) {
	t.adminHandler = adminHandler
}
