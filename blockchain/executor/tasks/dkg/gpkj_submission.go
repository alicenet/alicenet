package dkg

import (
	"fmt"
	"math/big"

	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"

	exConstants "github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	monInterfaces "github.com/MadBase/MadNet/blockchain/monitor/interfaces"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/MadBase/MadNet/constants"
)

// GPKjSubmissionTask contains required state for gpk submission
type GPKjSubmissionTask struct {
	*objects.Task
	adminHandler monInterfaces.IAdminHandler
}

// asserting that GPKjSubmissionTask struct implements interface Task
var _ interfaces.ITask = &GPKjSubmissionTask{}

// NewGPKjSubmissionTask creates a background task that attempts to submit the gpkj in ETHDKG
func NewGPKjSubmissionTask(start uint64, end uint64, adminHandler monInterfaces.IAdminHandler) *GPKjSubmissionTask {
	return &GPKjSubmissionTask{
		Task:         objects.NewTask(exConstants.GPKjSubmissionTaskName, start, end, false, transaction.NewSubscribeOptions(true, exConstants.ETHDKGMaxStaleBlocks)),
		adminHandler: adminHandler,
	}
}

// Prepare prepares for work to be done in the GPKjSubmissionTask
func (t *GPKjSubmissionTask) Prepare() *interfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debug("preparing task")

	dkgState := &state.DkgState{}
	var isRecoverable bool
	err := t.GetDB().Update(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		if err != nil {
			isRecoverable = false
			return err
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
				isRecoverable = false
				return fmt.Errorf(
					"Could not generate group keys: %v for index: %v",
					err,
					dkgState.Index,
				)
			}

			dkgState.GroupPrivateKey = groupPrivateKey
			dkgState.Participants[dkgState.Account.Address].GPKj = groupPublicKey

			// Pass private key on to consensus
			logger.Debugf("Adding private bn256eth key... using %p", t.adminHandler)
			err = t.adminHandler.AddPrivateKey(groupPrivateKey.Bytes(), constants.CurveBN256Eth)
			if err != nil {
				isRecoverable = true
				return fmt.Errorf("error adding private key: %v", err)
			}

			err = dkgState.PersistState(txn)
			if err != nil {
				isRecoverable = false
				return err
			}
		} else {
			logger.Debugf("group private-public key already defined")
		}

		return nil
	})

	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf(exConstants.ErrorDuringPreparation, err), isRecoverable)
	}

	return nil
}

// Execute executes the task business logic
func (t *GPKjSubmissionTask) Execute() ([]*types.Transaction, *interfaces.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf(exConstants.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()
	ctx := t.GetCtx()

	txnOpts, err := client.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf(exConstants.FailedGettingTxnOpts, err), true)
	}

	logger.Infof("submitting gpkj: %v", dkgState.Participants[dkgState.Account.Address].GPKj)
	txn, err := client.Contracts().Ethdkg().SubmitGPKJ(txnOpts, dkgState.Participants[dkgState.Account.Address].GPKj)
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf("submitting gpkj failed: %v", err), true)
	}

	return []*types.Transaction{txn}, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *GPKjSubmissionTask) ShouldExecute() *interfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf(exConstants.ErrorLoadingDkgState, err), false)
	}

	client := t.GetClient()
	ctx := t.GetCtx()
	if dkgState.Phase != state.GPKJSubmission {
		return interfaces.NewTaskErr(fmt.Sprintf("phase %v different from GPKJSubmission", dkgState.Phase), false)
	}

	//Check if my GPKj is submitted, if not should retry
	defaultAddr := dkgState.Account
	callOpts, err := client.GetCallOpts(ctx, defaultAddr)
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf(exConstants.FailedGettingCallOpts, err), true)
	}
	participantState, err := client.Contracts().Ethdkg().GetParticipantInternalState(callOpts, defaultAddr.Address)
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf("failed getting participants state: %v", err), true)
	}
	if participantState.Gpkj[0].Cmp(dkgState.Participants[defaultAddr.Address].GPKj[0]) == 0 &&
		participantState.Gpkj[1].Cmp(dkgState.Participants[defaultAddr.Address].GPKj[1]) == 0 &&
		participantState.Gpkj[2].Cmp(dkgState.Participants[defaultAddr.Address].GPKj[2]) == 0 &&
		participantState.Gpkj[3].Cmp(dkgState.Participants[defaultAddr.Address].GPKj[3]) == 0 {
		return interfaces.NewTaskErr(fmt.Sprint("GPKj already set"), false)
	}

	return nil
}

// SetAdminHandler sets the task adminHandler
func (t *GPKjSubmissionTask) SetAdminHandler(adminHandler monInterfaces.IAdminHandler) {
	t.adminHandler = adminHandler
}
