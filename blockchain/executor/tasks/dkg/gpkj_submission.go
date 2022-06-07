package dkg

import (
	"fmt"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"

	exConstants "github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	dkgUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
	exUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/utils"
	monInterfaces "github.com/MadBase/MadNet/blockchain/monitor/interfaces"
	"github.com/MadBase/MadNet/constants"
	"github.com/sirupsen/logrus"
)

// GPKjSubmissionTask contains required state for gpk submission
type GPKjSubmissionTask struct {
	*objects.Task
	adminHandler monInterfaces.IAdminHandler
}

// asserting that GPKjSubmissionTask struct implements interface exInterfacesinterfaces.Task
var _ interfaces.ITask = &GPKjSubmissionTask{}

// NewGPKjSubmissionTask creates a background task that attempts to submit the gpkj in ETHDKG
func NewGPKjSubmissionTask(start uint64, end uint64, adminHandler monInterfaces.IAdminHandler) *GPKjSubmissionTask {
	return &GPKjSubmissionTask{
		Task:         objects.NewTask(exConstants.GPKjSubmissionTaskName, start, end),
		adminHandler: adminHandler,
	}
}

// Prepare prepares for work to be done in the GPKjSubmissionTask
func (t *GPKjSubmissionTask) Prepare() error {
	logger := t.GetLogger()
	logger.Info("GPKSubmissionTask Prepare()...")

	dkgState := &state.DkgState{}
	err := t.GetDB().Update(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, logger)
		if err != nil {
			return err
		}

		if dkgState.GroupPrivateKey == nil ||
			dkgState.GroupPrivateKey.Cmp(big.NewInt(0)) == 0 {

			// Collecting all the participants encrypted shares to be used for the GPKj
			var participantsList = dkgState.GetSortedParticipants()
			encryptedShares := make([][]*big.Int, 0, dkgState.NumberOfValidators)
			for _, participant := range participantsList {
				logger.Debugf("Collecting encrypted shares... Participant %v %v", participant.Index, participant.Address.Hex())
				encryptedShares = append(encryptedShares, participant.EncryptedShares)
			}

			// Generate the GPKj
			groupPrivateKey, groupPublicKey, err := state.GenerateGroupKeys(
				dkgState.TransportPrivateKey, dkgState.PrivateCoefficients,
				encryptedShares, dkgState.Index, participantsList)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"t.State.Index": dkgState.Index,
				}).Errorf("Could not generate group keys: %v", err)
				return dkgUtils.LogReturnErrorf(logger, "Could not generate group keys: %v", err)
			}

			dkgState.GroupPrivateKey = groupPrivateKey
			dkgState.Participants[dkgState.Account.Address].GPKj = groupPublicKey

			// Pass private key on to consensus
			logger.Infof("Adding private bn256eth key... using %p", t.adminHandler)
			err = t.adminHandler.AddPrivateKey(groupPrivateKey.Bytes(), constants.CurveBN256Eth)
			if err != nil {
				return fmt.Errorf("%w because error adding private key: %v", objects.ErrCanNotContinue, err)
			}

			err = dkgState.PersistState(txn, logger)
			if err != nil {
				return err
			}
		} else {
			logger.Infof("GPKSubmissionTask Initialize(): group private-public key already defined")
		}

		return nil
	})

	if err != nil {
		return dkgUtils.LogReturnErrorf(logger, "GPKjSubmissionTask.Prepare(): error during the preparation: %v", err)
	}

	return nil
}

// Execute executes the task business logic
func (t *GPKjSubmissionTask) Execute() ([]*types.Transaction, error) {
	logger := t.GetLogger()

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn, t.GetLogger())
		return err
	})
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "GPKjSubmissionTask.Execute(): error loading dkgState: %v", err)
	}

	eth := t.GetEth()
	ctx := t.GetCtx()
	logger.Infof("GPKSubmissionTask Execute(): %v", dkgState.Account.Address)

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// Do it
	txn, err := eth.Contracts().Ethdkg().SubmitGPKJ(txnOpts, dkgState.Participants[dkgState.Account.Address].GPKj)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "submitting master public key failed: %v", err)
	}

	return []*types.Transaction{txn}, nil
}

// ShouldRetry checks if it makes sense to try again
func (t *GPKjSubmissionTask) ShouldExecute() bool {
	logger := t.GetLogger()
	logger.Info("GPKjSubmissionTask ShouldExecute()")

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

	if dkgState.Phase != state.GPKJSubmission {
		return false
	}

	//Check if my GPKj is submitted, if not should retry
	me := dkgState.Account
	callOpts, err := eth.GetCallOpts(ctx, me)
	if err != nil {
		logger.Debug("GPKjSubmissionTask ShouldExecute() failed getting call opts")
		return true
	}
	participantState, err := eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, me.Address)
	if err == nil && participantState.Gpkj[0].Cmp(dkgState.Participants[me.Address].GPKj[0]) == 0 &&
		participantState.Gpkj[1].Cmp(dkgState.Participants[me.Address].GPKj[1]) == 0 &&
		participantState.Gpkj[2].Cmp(dkgState.Participants[me.Address].GPKj[2]) == 0 &&
		participantState.Gpkj[3].Cmp(dkgState.Participants[me.Address].GPKj[3]) == 0 {
		return false
	}

	return true
}

// SetAdminHandler sets the task adminHandler
func (t *GPKjSubmissionTask) SetAdminHandler(adminHandler monInterfaces.IAdminHandler) {
	t.adminHandler = adminHandler
}
