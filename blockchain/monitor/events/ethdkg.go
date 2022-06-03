package events

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/blockchain/executor/constants"
	executorInterfaces "github.com/MadBase/MadNet/blockchain/executor/interfaces"
	dkgtasks "github.com/MadBase/MadNet/blockchain/executor/tasks/dkg"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
	monitorInterfaces "github.com/MadBase/MadNet/blockchain/monitor/interfaces"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

var (
	ErrNotAValidator = errors.New("nothing schedule for time")
)

// todo: improve this
func isValidator(eth ethereum.Network, logger *logrus.Entry, acct accounts.Account) (bool, error) {
	ctx := context.Background()
	callOpts, err := eth.GetCallOpts(ctx, acct)
	if err != nil {
		return false, errors.New(fmt.Sprintf("cannot check if I'm a validator, failed getting call options: %v", err))
	}
	isValidator, err := eth.Contracts().ValidatorPool().IsValidator(callOpts, acct.Address)
	if err != nil {
		return false, errors.New(fmt.Sprintf("cannot check if I'm a validator :%v", err))
	}
	if !isValidator {
		logger.Info("cannot take part in ETHDKG because I'm not a validator")
		return false, nil
	}

	return true, nil
}

func ProcessRegistrationOpened(eth ethereum.Network, logger *logrus.Entry, log types.Log, cdb *db.Database, taskRequestChan chan<- executorInterfaces.ITask) error {
	logger.Info("ProcessRegistrationOpened() ...")
	event, err := eth.Contracts().Ethdkg().ParseRegistrationOpened(log)
	if err != nil {
		return err
	}

	amIaValidator, err := isValidator(eth, logger, eth.GetDefaultAccount())
	if err != nil {
		return utils.LogReturnErrorf(logger, "I'm not a validator: %v", err)
	}

	// get validators from ValidatorPool
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()

	callOpts, err := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	if err != nil {
		return utils.LogReturnErrorf(logger, "failed to get call options to retrieve validators address from pool: %v", err)
	}
	validatorAddresses, err := utils.GetValidatorAddressesFromPool(callOpts, eth, logger)
	if err != nil {
		return utils.LogReturnErrorf(logger, "failed to retrieve validator state from validator pool: %v", err)
	}

	dkgState, registrationTask, disputeMissingRegistrationTask := UpdateStateOnRegistrationOpened(
		eth.GetDefaultAccount(),
		event.StartBlock.Uint64(),
		event.PhaseLength.Uint64(),
		event.ConfirmationLength.Uint64(),
		event.Nonce.Uint64(),
		amIaValidator,
		validatorAddresses,
	)

	logger.WithFields(logrus.Fields{
		"StartBlock":         event.StartBlock,
		"NumberValidators":   event.NumberValidators,
		"Nonce":              event.Nonce,
		"PhaseLength":        event.PhaseLength,
		"ConfirmationLength": event.ConfirmationLength,
		"RegistrationEnd":    registrationTask.GetExecutionData().GetEnd(),
	}).Info("ETHDKG RegistrationOpened")

	if !dkgState.IsValidator {
		return nil
	}

	// schedule Registration
	logger.WithFields(logrus.Fields{
		"TaskStart": registrationTask.GetStart(),
		"TaskEnd":   registrationTask.GetEnd(),
	}).Info("Scheduling NewRegisterTask")

	taskRequestChan <- registrationTask

	// schedule DisputeRegistration
	logger.WithFields(logrus.Fields{
		"TaskStart": disputeMissingRegistrationTask.GetStart(),
		"TaskEnd":   disputeMissingRegistrationTask.GetEnd(),
	}).Info("Scheduling NewDisputeRegistrationTask")

	taskRequestChan <- disputeMissingRegistrationTask

	err = cdb.Update(func(txn *badger.Txn) error {
		err := state.PersistEthDkgState(txn, logger, dkgState)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return utils.LogReturnErrorf(logger, "Failed to save dkgState on ProcessRegistrationOpened: %v", err)
	}

	if err = cdb.Sync(); err != nil {
		return utils.LogReturnErrorf(logger, "Failed to set sync on ProcessRegistrationOpened: %v", err)
	}

	return nil
}

func UpdateStateOnRegistrationOpened(account accounts.Account, startBlock, phaseLength, confirmationLength, nonce uint64, amIValidator bool, validatorAddresses []common.Address) (*state.DkgState, *dkgtasks.RegisterTask, *dkgtasks.DisputeMissingRegistrationTask) {
	dkgState := state.NewDkgState(account)
	dkgState.OnRegistrationOpened(
		startBlock,
		phaseLength,
		confirmationLength,
		nonce,
	)

	dkgState.IsValidator = amIValidator
	dkgState.ValidatorAddresses = validatorAddresses
	dkgState.NumberOfValidators = len(validatorAddresses)

	registrationEnds := dkgState.PhaseStart + dkgState.PhaseLength
	registrationTask := dkgtasks.NewRegisterTask(dkgState, dkgState.PhaseStart, registrationEnds)
	disputeMissingRegistrationTask := dkgtasks.NewDisputeMissingRegistrationTask(dkgState, registrationEnds, registrationEnds+dkgState.PhaseLength)

	return dkgState, registrationTask, disputeMissingRegistrationTask
}

func ProcessAddressRegistered(eth ethereum.Network, logger *logrus.Entry, log types.Log, cdb *db.Database) error {

	logger.Info("ProcessAddressRegistered() ...")

	event, err := eth.Contracts().Ethdkg().ParseAddressRegistered(log)
	if err != nil {
		return err
	}

	err = cdb.Update(func(txn *badger.Txn) error {
		dkgState, err := state.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		logger.WithFields(logrus.Fields{
			"Account":       event.Account.Hex(),
			"Index":         event.Index,
			"numRegistered": event.Index,
			"Nonce":         event.Nonce,
			"PublicKey":     event.PublicKey,
			"#Participants": len(dkgState.Participants),
			"#Validators":   len(dkgState.ValidatorAddresses),
		}).Info("Address registered!")

		dkgState.OnAddressRegistered(event.Account, int(event.Index.Int64()), event.Nonce.Uint64(), event.PublicKey)

		err = state.PersistEthDkgState(txn, logger, dkgState)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return utils.LogReturnErrorf(logger, "Failed to save dkgState on ProcessAddressRegistered: %v", err)
	}

	if err = cdb.Sync(); err != nil {
		return utils.LogReturnErrorf(logger, "Failed to set sync on ProcessAddressRegistered: %v", err)
	}

	return nil
}

func ProcessRegistrationComplete(eth ethereum.Network, logger *logrus.Entry, log types.Log, cdb *db.Database, taskRequestChan chan<- executorInterfaces.ITask, taskKillChan chan<- string) error {

	logger.Info("ProcessRegistrationComplete() ...")
	shareDistributionTask := &dkgtasks.ShareDistributionTask{}
	disputeMissingShareDistributionTask := &dkgtasks.DisputeMissingShareDistributionTask{}
	disputeBadSharesTask := &dkgtasks.DisputeShareDistributionTask{}

	err := cdb.Update(func(txn *badger.Txn) error {
		dkgState, err := state.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		if !dkgState.IsValidator {
			return ErrNotAValidator
		}

		event, err := eth.Contracts().Ethdkg().ParseRegistrationComplete(log)
		if err != nil {
			return err
		}

		logger.WithFields(logrus.Fields{
			"BlockNumber": event.BlockNumber,
		}).Info("ETHDKG Registration Complete")

		shareDistributionTask, disputeMissingShareDistributionTask, disputeBadSharesTask = UpdateStateOnRegistrationComplete(dkgState, event.BlockNumber.Uint64())

		err = state.PersistEthDkgState(txn, logger, dkgState)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		//If Im not a Validator we just return nil
		if errors.Is(err, ErrNotAValidator) {
			return nil
		}
		return utils.LogReturnErrorf(logger, "Failed to save dkgState on ProcessRegistrationComplete: %v", err)
	}

	if err = cdb.Sync(); err != nil {
		return utils.LogReturnErrorf(logger, "Failed to set sync on ProcessRegistrationComplete: %v", err)
	}

	//Killing previous tasks
	taskKillChan <- constants.RegisterTaskName
	taskKillChan <- constants.DisputeMissingRegistrationTaskName

	// schedule ShareDistribution phase
	logger.WithFields(logrus.Fields{
		"TaskStart": shareDistributionTask.GetStart(),
		"TaskEnd":   shareDistributionTask.GetEnd(),
	}).Info("Scheduling NewShareDistributionTask")

	taskRequestChan <- shareDistributionTask

	// schedule DisputeParticipantDidNotDistributeSharesTask
	logger.WithFields(logrus.Fields{
		"TaskStart": disputeMissingShareDistributionTask.GetStart(),
		"TaskEnd":   disputeMissingShareDistributionTask.GetEnd(),
	}).Info("Scheduling NewDisputeParticipantDidNotDistributeSharesTask")

	taskRequestChan <- disputeMissingShareDistributionTask

	// schedule DisputeDistributeSharesTask
	logger.WithFields(logrus.Fields{
		"TaskStart": disputeBadSharesTask.GetStart(),
		"TaskEnd":   disputeBadSharesTask.GetEnd(),
	}).Info("Scheduling NewDisputeDistributeSharesTask")

	taskRequestChan <- disputeBadSharesTask

	return nil
}

func UpdateStateOnRegistrationComplete(dkgState *state.DkgState, shareDistributionStartBlockNumber uint64) (*dkgtasks.ShareDistributionTask, *dkgtasks.DisputeMissingShareDistributionTask, *dkgtasks.DisputeShareDistributionTask) {
	dkgState.OnRegistrationComplete(shareDistributionStartBlockNumber)

	shareDistStartBlock := dkgState.PhaseStart
	shareDistEndBlock := shareDistStartBlock + dkgState.PhaseLength
	shareDistributionTask := dkgtasks.NewShareDistributionTask(dkgState, shareDistStartBlock, shareDistEndBlock)

	var dispShareStartBlock = shareDistEndBlock
	var dispShareEndBlock = dispShareStartBlock + dkgState.PhaseLength
	disputeMissingShareDistributionTask := dkgtasks.NewDisputeMissingShareDistributionTask(dkgState, dispShareStartBlock, dispShareEndBlock)
	disputeBadSharesTask := dkgtasks.NewDisputeShareDistributionTask(dkgState, dispShareStartBlock, dispShareEndBlock)

	return shareDistributionTask, disputeMissingShareDistributionTask, disputeBadSharesTask
}

func ProcessShareDistribution(eth ethereum.Network, logger *logrus.Entry, log types.Log, cdb *db.Database) error {

	logger.Info("ProcessShareDistribution() ...")

	event, err := eth.Contracts().Ethdkg().ParseSharesDistributed(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"Issuer":          event.Account.Hex(),
		"Index":           event.Index,
		"EncryptedShares": event.EncryptedShares,
		"Commitments":     event.Commitments,
	}).Info("Received share distribution")

	err = cdb.Update(func(txn *badger.Txn) error {
		dkgState, err := state.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		err = dkgState.OnSharesDistributed(logger, event.Account, event.EncryptedShares, event.Commitments)
		if err != nil {
			return err
		}

		err = state.PersistEthDkgState(txn, logger, dkgState)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return utils.LogReturnErrorf(logger, "Failed to save dkgState on ProcessShareDistribution: %v", err)
	}

	if err = cdb.Sync(); err != nil {
		return utils.LogReturnErrorf(logger, "Failed to set sync on ProcessShareDistribution: %v", err)
	}

	return nil
}

func ProcessShareDistributionComplete(eth ethereum.Network, logger *logrus.Entry, log types.Log, cdb *db.Database, taskRequestChan chan<- executorInterfaces.ITask, taskKillChan chan<- string) error {
	logger.Info("ProcessShareDistributionComplete() ...")
	disputeShareDistributionTask := &dkgtasks.DisputeShareDistributionTask{}
	keyShareSubmissionTask := &dkgtasks.KeyShareSubmissionTask{}
	disputeMissingKeySharesTask := &dkgtasks.DisputeMissingKeySharesTask{}

	err := cdb.Update(func(txn *badger.Txn) error {
		dkgState, err := state.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		if !dkgState.IsValidator {
			return ErrNotAValidator
		}

		event, err := eth.Contracts().Ethdkg().ParseShareDistributionComplete(log)
		if err != nil {
			return err
		}

		logger.WithFields(logrus.Fields{
			"BlockNumber": event.BlockNumber,
		}).Info("Received share distribution complete")

		disputeShareDistributionTask, keyShareSubmissionTask, disputeMissingKeySharesTask = UpdateStateOnShareDistributionComplete(dkgState, event.BlockNumber.Uint64())
		err = state.PersistEthDkgState(txn, logger, dkgState)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		//If Im not a Validator we just return nil
		if errors.Is(err, ErrNotAValidator) {
			return nil
		}
		return utils.LogReturnErrorf(logger, "Failed to save dkgState on ProcessShareDistributionComplete: %v", err)
	}

	if err = cdb.Sync(); err != nil {
		return utils.LogReturnErrorf(logger, "Failed to set sync on ProcessShareDistributionComplete: %v", err)
	}

	//Killing previous tasks
	taskKillChan <- constants.ShareDistributionTaskName
	taskKillChan <- constants.DisputeMissingShareDistributionTaskName
	taskKillChan <- constants.DisputeShareDistributionTaskName

	// schedule DisputeShareDistributionTask
	logger.WithFields(logrus.Fields{
		"TaskStart": disputeShareDistributionTask.GetStart(),
		"TaskEnd":   disputeShareDistributionTask.GetEnd(),
	}).Info("Scheduling NewDisputeShareDistributionTask")
	taskRequestChan <- disputeShareDistributionTask

	// schedule SubmitKeySharesPhase
	logger.WithFields(logrus.Fields{
		"TaskStart": keyShareSubmissionTask.GetStart(),
		"TaskEnd":   keyShareSubmissionTask.GetEnd(),
	}).Info("Scheduling NewKeyShareSubmissionTask")
	taskRequestChan <- keyShareSubmissionTask

	// schedule DisputeMissingKeySharesPhase
	logger.WithFields(logrus.Fields{
		"TaskStart": disputeMissingKeySharesTask.GetStart(),
		"TaskEnd":   disputeMissingKeySharesTask.GetEnd(),
	}).Info("Scheduling NewDisputeMissingKeySharesTask")
	taskRequestChan <- disputeMissingKeySharesTask

	return nil
}

func UpdateStateOnShareDistributionComplete(dkgState *state.DkgState, disputeShareDistributionStartBlock uint64) (*dkgtasks.DisputeShareDistributionTask, *dkgtasks.KeyShareSubmissionTask, *dkgtasks.DisputeMissingKeySharesTask) {
	dkgState.OnShareDistributionComplete(disputeShareDistributionStartBlock)

	phaseEnd := dkgState.PhaseStart + dkgState.PhaseLength
	disputeShareDistributionTask := dkgtasks.NewDisputeShareDistributionTask(dkgState, dkgState.PhaseStart, phaseEnd)

	// schedule SubmitKeySharesPhase
	submitKeySharesPhaseStart := phaseEnd
	submitKeySharesPhaseEnd := submitKeySharesPhaseStart + dkgState.PhaseLength
	keyshareSubmissionTask := dkgtasks.NewKeyShareSubmissionTask(dkgState, submitKeySharesPhaseStart, submitKeySharesPhaseEnd)

	// schedule DisputeMissingKeySharesPhase
	missingKeySharesDisputeStart := submitKeySharesPhaseEnd
	missingKeySharesDisputeEnd := missingKeySharesDisputeStart + dkgState.PhaseLength
	disputeMissingKeySharesTask := dkgtasks.NewDisputeMissingKeySharesTask(dkgState, missingKeySharesDisputeStart, missingKeySharesDisputeEnd)

	return disputeShareDistributionTask, keyshareSubmissionTask, disputeMissingKeySharesTask
}

func ProcessKeyShareSubmitted(eth ethereum.Network, logger *logrus.Entry, log types.Log, cdb *db.Database) error {

	logger.Info("ProcessKeyShareSubmitted() ...")

	event, err := eth.Contracts().Ethdkg().ParseKeyShareSubmitted(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"Issuer":                     event.Account.Hex(),
		"KeyShareG1":                 event.KeyShareG1,
		"KeyShareG1CorrectnessProof": event.KeyShareG1CorrectnessProof,
		"KeyShareG2":                 event.KeyShareG2,
	}).Info("Received key shares")

	err = cdb.Update(func(txn *badger.Txn) error {
		dkgState, err := state.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		dkgState.OnKeyShareSubmitted(event.Account, event.KeyShareG1, event.KeyShareG1CorrectnessProof, event.KeyShareG2)
		err = state.PersistEthDkgState(txn, logger, dkgState)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return utils.LogReturnErrorf(logger, "Failed to save dkgState on ProcessKeyShareSubmitted: %v", err)
	}

	if err = cdb.Sync(); err != nil {
		return utils.LogReturnErrorf(logger, "Failed to set sync on ProcessKeyShareSubmitted: %v", err)
	}

	return nil
}

func ProcessKeyShareSubmissionComplete(eth ethereum.Network, logger *logrus.Entry, log types.Log, cdb *db.Database, taskRequestChan chan<- executorInterfaces.ITask, taskKillChan chan<- string) error {
	event, err := eth.Contracts().Ethdkg().ParseKeyShareSubmissionComplete(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
	}).Info("ProcessKeyShareSubmissionComplete() ...")

	mpkSubmissionTask := &dkgtasks.MPKSubmissionTask{}
	err = cdb.Update(func(txn *badger.Txn) error {
		dkgState, err := state.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		if dkgState.IsValidator {
			return ErrNotAValidator
		}

		// schedule MPK submission
		mpkSubmissionTask = UpdateStateOnKeyShareSubmissionComplete(dkgState, event.BlockNumber.Uint64())
		err = state.PersistEthDkgState(txn, logger, dkgState)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		//If Im not a Validator we just return nil
		if errors.Is(err, ErrNotAValidator) {
			return nil
		}
		return utils.LogReturnErrorf(logger, "Failed to save dkgState on ProcessKeyShareSubmissionComplete: %v", err)
	}

	if err = cdb.Sync(); err != nil {
		return utils.LogReturnErrorf(logger, "Failed to set sync on ProcessKeyShareSubmissionComplete: %v", err)
	}

	//Killing previous tasks
	taskKillChan <- constants.KeyShareSubmissionTaskName
	taskKillChan <- constants.DisputeMissingKeySharesTaskName

	// schedule MPKSubmissionTask
	taskRequestChan <- mpkSubmissionTask

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"TaskStart":   mpkSubmissionTask.GetStart(),
		"TaskEnd":     mpkSubmissionTask.GetEnd(),
	}).Info("Scheduling MPKSubmissionTask")

	return nil
}

func UpdateStateOnKeyShareSubmissionComplete(dkgState *state.DkgState, mpkSubmissionStartBlock uint64) *dkgtasks.MPKSubmissionTask {
	dkgState.OnKeyShareSubmissionComplete(mpkSubmissionStartBlock)

	phaseEnd := dkgState.PhaseStart + dkgState.PhaseLength
	mpkSubmissionTask := dkgtasks.NewMPKSubmissionTask(dkgState, dkgState.PhaseStart, phaseEnd)

	return mpkSubmissionTask
}

func ProcessMPKSet(eth ethereum.Network, logger *logrus.Entry, log types.Log, adminHandler monitorInterfaces.IAdminHandler, cdb *db.Database, taskRequestChan chan<- executorInterfaces.ITask, taskKillChan chan<- string) error {

	event, err := eth.Contracts().Ethdkg().ParseMPKSet(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"Nonce":       event.Nonce,
		"MPK":         event.Mpk,
	}).Info("ProcessMPKSet() ...")

	gpkjSubmissionTask := &dkgtasks.GPKjSubmissionTask{}
	disputeMissingGPKjTask := &dkgtasks.DisputeMissingGPKjTask{}
	disputeGPKjTask := &dkgtasks.DisputeGPKjTask{}
	err = cdb.Update(func(txn *badger.Txn) error {
		dkgState, err := state.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		if dkgState.IsValidator {
			return ErrNotAValidator
		}

		gpkjSubmissionTask, disputeMissingGPKjTask, disputeGPKjTask = UpdateStateOnMPKSet(dkgState, event.BlockNumber.Uint64(), adminHandler)
		err = state.PersistEthDkgState(txn, logger, dkgState)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		//If Im not a Validator we just return nil
		if errors.Is(err, ErrNotAValidator) {
			return nil
		}
		return utils.LogReturnErrorf(logger, "Failed to save dkgState on ProcessMPKSet: %v", err)
	}

	if err = cdb.Sync(); err != nil {
		return utils.LogReturnErrorf(logger, "Failed to set sync on ProcessMPKSet: %v", err)
	}

	//Killing previous tasks
	taskKillChan <- constants.MPKSubmissionTaskName

	// schedule GPKJSubmissionTask
	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"TaskStart":   gpkjSubmissionTask.GetStart(),
		"TaskEnd":     gpkjSubmissionTask.GetEnd(),
	}).Info("Scheduling GPKJSubmissionTask")

	taskRequestChan <- gpkjSubmissionTask

	// schedule DisputeMissingGPKjTask
	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"TaskStart":   gpkjSubmissionTask.GetStart(),
		"TaskEnd":     gpkjSubmissionTask.GetEnd(),
	}).Info("Scheduling DisputeMissingGPKjTask")

	taskRequestChan <- disputeMissingGPKjTask

	// schedule DisputeGPKjTask
	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"TaskStart":   gpkjSubmissionTask.GetStart(),
		"TaskEnd":     gpkjSubmissionTask.GetEnd(),
	}).Info("Scheduling DisputeGPKjTask")

	taskRequestChan <- disputeGPKjTask

	return nil
}

func UpdateStateOnMPKSet(dkgState *state.DkgState, gpkjSubmissionStartBlock uint64, adminHandler monitorInterfaces.IAdminHandler) (*dkgtasks.GPKjSubmissionTask, *dkgtasks.DisputeMissingGPKjTask, *dkgtasks.DisputeGPKjTask) {
	dkgState.OnMPKSet(gpkjSubmissionStartBlock)
	gpkjSubmissionEnd := dkgState.PhaseStart + dkgState.PhaseLength
	gpkjSubmissionTask := dkgtasks.NewGPKjSubmissionTask(dkgState, dkgState.PhaseStart, gpkjSubmissionEnd, adminHandler)

	disputeMissingGPKjStart := gpkjSubmissionEnd
	disputeMissingGPKjEnd := disputeMissingGPKjStart + dkgState.PhaseLength
	disputeMissingGPKjTask := dkgtasks.NewDisputeMissingGPKjTask(dkgState, disputeMissingGPKjStart, disputeMissingGPKjEnd)
	disputeGPKjTask := dkgtasks.NewDisputeGPKjTask(dkgState, disputeMissingGPKjStart, disputeMissingGPKjEnd)

	return gpkjSubmissionTask, disputeMissingGPKjTask, disputeGPKjTask
}

func ProcessGPKJSubmissionComplete(eth ethereum.Network, logger *logrus.Entry, log types.Log, cdb *db.Database, taskRequestChan chan<- executorInterfaces.ITask, taskKillChan chan<- string) error {

	event, err := eth.Contracts().Ethdkg().ParseGPKJSubmissionComplete(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
	}).Info("ProcessGPKJSubmissionComplete() ...")

	disputeGPKjTask := &dkgtasks.DisputeGPKjTask{}
	completionTask := &dkgtasks.CompletionTask{}
	err = cdb.Update(func(txn *badger.Txn) error {
		dkgState, err := state.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		if !dkgState.IsValidator {
			return ErrNotAValidator
		}

		disputeGPKjTask, completionTask = UpdateStateOnGPKJSubmissionComplete(dkgState, event.BlockNumber.Uint64())
		err = state.PersistEthDkgState(txn, logger, dkgState)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		//If Im not a Validator we just return nil
		if errors.Is(err, ErrNotAValidator) {
			return nil
		}
		return utils.LogReturnErrorf(logger, "Failed to save dkgState on ProcessGPKJSubmissionComplete: %v", err)
	}

	if err = cdb.Sync(); err != nil {
		return utils.LogReturnErrorf(logger, "Failed to set sync on ProcessGPKJSubmissionComplete: %v", err)
	}

	//Killing previous tasks
	taskKillChan <- constants.GPKjSubmissionTaskName
	taskKillChan <- constants.DisputeMissingGPKjTaskName
	taskKillChan <- constants.DisputeGPKjTaskName

	// schedule DisputeGPKJSubmissionTask
	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"TaskStart":   disputeGPKjTask.GetStart(),
		"TaskEnd":     disputeGPKjTask.GetEnd(),
	}).Info("Scheduling NewGPKJDisputeTask")

	taskRequestChan <- disputeGPKjTask

	// schedule Completion
	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"TaskStart":   completionTask.GetStart(),
		"TaskEnd":     completionTask.GetEnd(),
	}).Info("Scheduling NewCompletionTask")

	taskRequestChan <- completionTask

	return nil
}

func UpdateStateOnGPKJSubmissionComplete(dkgState *state.DkgState, disputeGPKjStartBlock uint64) (*dkgtasks.DisputeGPKjTask, *dkgtasks.CompletionTask) {
	dkgState.OnGPKJSubmissionComplete(disputeGPKjStartBlock)

	disputeGPKjPhaseEnd := dkgState.PhaseStart + dkgState.PhaseLength
	disputeGPKjTask := dkgtasks.NewDisputeGPKjTask(dkgState, dkgState.PhaseStart, disputeGPKjPhaseEnd)

	completionStart := disputeGPKjPhaseEnd
	completionEnd := completionStart + dkgState.PhaseLength
	completionTask := dkgtasks.NewCompletionTask(dkgState, completionStart, completionEnd)

	return disputeGPKjTask, completionTask
}
