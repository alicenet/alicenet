package events

import (
	"context"
	"errors"
	"fmt"
	"time"

	executorInterfaces "github.com/MadBase/MadNet/blockchain/executor/interfaces"
	dkgtasks "github.com/MadBase/MadNet/blockchain/executor/tasks/dkg"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/dgraph-io/badger/v2"

	ethereumInterfaces "github.com/MadBase/MadNet/blockchain/ethereum/interfaces"
	monitorInterfaces "github.com/MadBase/MadNet/blockchain/monitor/interfaces"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

var (
	ErrNotAValidator = errors.New("nothing schedule for time")
)

// todo: improve this
func isValidator(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, acct accounts.Account) (bool, error) {
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

func ProcessRegistrationOpened(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, log types.Log, cdb *db.Database, taskRequestChan chan<- executorInterfaces.ITask) error {
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
		return utils.LogReturnErrorf(logger, "failed to retrieve validator data from validator pool: %v", err)
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
		err := objects.PersistEthDkgState(txn, logger, dkgState)
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

func UpdateStateOnRegistrationOpened(account accounts.Account, startBlock, phaseLength, confirmationLength, nonce uint64, amIValidator bool, validatorAddresses []common.Address) (*dkgObjects.DkgState, *dkgtasks.RegisterTask, *dkgtasks.DisputeMissingRegistrationTask) {
	dkgState := objects.NewDkgState(account)
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

func ProcessAddressRegistered(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, log types.Log, cdb *db.Database) error {

	logger.Info("ProcessAddressRegistered() ...")

	event, err := eth.Contracts().Ethdkg().ParseAddressRegistered(log)
	if err != nil {
		return err
	}

	err = cdb.Update(func(txn *badger.Txn) error {
		dkgState, err := objects.LoadEthDkgState(txn, logger)
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

		err = objects.PersistEthDkgState(txn, logger, dkgState)
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

func ProcessRegistrationComplete(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, log types.Log, cdb *db.Database, taskRequestChan chan<- executorInterfaces.ITask, taskKillChan chan<- string) error {

	logger.Info("ProcessRegistrationComplete() ...")
	shareDistributionTask := &dkgtasks.ShareDistributionTask{}
	disputeMissingShareDistributionTask := &dkgtasks.DisputeMissingShareDistributionTask{}
	disputeBadSharesTask := &dkgtasks.DisputeShareDistributionTask{}

	err := cdb.Update(func(txn *badger.Txn) error {
		dkgState, err := objects.LoadEthDkgState(txn, logger)
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

		err = objects.PersistEthDkgState(txn, logger, dkgState)
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
	taskKillChan <- dkgtasks.RegisterTaskName
	taskKillChan <- dkgtasks.DisputeMissingRegistrationTaskName

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

func UpdateStateOnRegistrationComplete(dkgState *dkgObjects.DkgState, shareDistributionStartBlockNumber uint64) (*dkgtasks.ShareDistributionTask, *dkgtasks.DisputeMissingShareDistributionTask, *dkgtasks.DisputeShareDistributionTask) {
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

func ProcessShareDistribution(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, log types.Log, cdb *db.Database) error {

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
		dkgState, err := objects.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		err = dkgState.OnSharesDistributed(logger, event.Account, event.EncryptedShares, event.Commitments)
		if err != nil {
			return err
		}

		err = objects.PersistEthDkgState(txn, logger, dkgState)
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

func ProcessShareDistributionComplete(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, log types.Log, cdb *db.Database, taskRequestChan chan<- executorInterfaces.ITask, taskKillChan chan<- string) error {
	logger.Info("ProcessShareDistributionComplete() ...")
	disputeShareDistributionTask := &dkgtasks.DisputeShareDistributionTask{}
	keyShareSubmissionTask := &dkgtasks.KeyShareSubmissionTask{}
	disputeMissingKeySharesTask := &dkgtasks.DisputeMissingKeySharesTask{}

	err := cdb.Update(func(txn *badger.Txn) error {
		dkgState, err := objects.LoadEthDkgState(txn, logger)
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
		err = objects.PersistEthDkgState(txn, logger, dkgState)
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
	taskKillChan <- dkgtasks.ShareDistributionTaskName
	taskKillChan <- dkgtasks.DisputeMissingShareDistributionTaskName
	taskKillChan <- dkgtasks.DisputeShareDistributionTaskName

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

func UpdateStateOnShareDistributionComplete(state *dkgObjects.DkgState, disputeShareDistributionStartBlock uint64) (*dkgtasks.DisputeShareDistributionTask, *dkgtasks.KeyShareSubmissionTask, *dkgtasks.DisputeMissingKeySharesTask) {
	state.OnShareDistributionComplete(disputeShareDistributionStartBlock)

	phaseEnd := state.PhaseStart + state.PhaseLength
	disputeShareDistributionTask := dkgtasks.NewDisputeShareDistributionTask(state, state.PhaseStart, phaseEnd)

	// schedule SubmitKeySharesPhase
	submitKeySharesPhaseStart := phaseEnd
	submitKeySharesPhaseEnd := submitKeySharesPhaseStart + state.PhaseLength
	keyshareSubmissionTask := dkgtasks.NewKeyShareSubmissionTask(state, submitKeySharesPhaseStart, submitKeySharesPhaseEnd)

	// schedule DisputeMissingKeySharesPhase
	missingKeySharesDisputeStart := submitKeySharesPhaseEnd
	missingKeySharesDisputeEnd := missingKeySharesDisputeStart + state.PhaseLength
	disputeMissingKeySharesTask := dkgtasks.NewDisputeMissingKeySharesTask(state, missingKeySharesDisputeStart, missingKeySharesDisputeEnd)

	return disputeShareDistributionTask, keyshareSubmissionTask, disputeMissingKeySharesTask
}

func ProcessKeyShareSubmitted(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, log types.Log, cdb *db.Database) error {

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
		dkgState, err := objects.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		dkgState.OnKeyShareSubmitted(event.Account, event.KeyShareG1, event.KeyShareG1CorrectnessProof, event.KeyShareG2)
		err = objects.PersistEthDkgState(txn, logger, dkgState)
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

func ProcessKeyShareSubmissionComplete(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, log types.Log, cdb *db.Database, taskRequestChan chan<- executorInterfaces.ITask, taskKillChan chan<- string) error {
	event, err := eth.Contracts().Ethdkg().ParseKeyShareSubmissionComplete(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
	}).Info("ProcessKeyShareSubmissionComplete() ...")

	mpkSubmissionTask := &dkgtasks.MPKSubmissionTask{}
	err = cdb.Update(func(txn *badger.Txn) error {
		dkgState, err := objects.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		if dkgState.IsValidator {
			return ErrNotAValidator
		}

		// schedule MPK submission
		mpkSubmissionTask = UpdateStateOnKeyShareSubmissionComplete(dkgState, event.BlockNumber.Uint64())
		err = objects.PersistEthDkgState(txn, logger, dkgState)
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
	taskKillChan <- dkgtasks.KeyShareSubmissionTaskName
	taskKillChan <- dkgtasks.DisputeMissingKeySharesTaskName

	// schedule MPKSubmissionTask
	taskRequestChan <- mpkSubmissionTask

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"TaskStart":   mpkSubmissionTask.GetStart(),
		"TaskEnd":     mpkSubmissionTask.GetEnd(),
	}).Info("Scheduling MPKSubmissionTask")

	return nil
}

func UpdateStateOnKeyShareSubmissionComplete(state *dkgObjects.DkgState, mpkSubmissionStartBlock uint64) *dkgtasks.MPKSubmissionTask {
	state.OnKeyShareSubmissionComplete(mpkSubmissionStartBlock)

	phaseEnd := state.PhaseStart + state.PhaseLength
	mpkSubmissionTask := dkgtasks.NewMPKSubmissionTask(state, state.PhaseStart, phaseEnd)

	return mpkSubmissionTask
}

func ProcessMPKSet(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, log types.Log, adminHandler monitorInterfaces.AdminHandler, cdb *db.Database, taskRequestChan chan<- executorInterfaces.ITask, taskKillChan chan<- string) error {

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
		dkgState, err := objects.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		if dkgState.IsValidator {
			return ErrNotAValidator
		}

		gpkjSubmissionTask, disputeMissingGPKjTask, disputeGPKjTask = UpdateStateOnMPKSet(dkgState, event.BlockNumber.Uint64(), adminHandler)
		err = objects.PersistEthDkgState(txn, logger, dkgState)
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
	taskKillChan <- dkgtasks.MPKSubmissionTaskName

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

func UpdateStateOnMPKSet(state *dkgObjects.DkgState, gpkjSubmissionStartBlock uint64, adminHandler monitorInterfaces.AdminHandler) (*dkgtasks.GPKjSubmissionTask, *dkgtasks.DisputeMissingGPKjTask, *dkgtasks.DisputeGPKjTask) {
	state.OnMPKSet(gpkjSubmissionStartBlock)
	gpkjSubmissionEnd := state.PhaseStart + state.PhaseLength
	gpkjSubmissionTask := dkgtasks.NewGPKjSubmissionTask(state, state.PhaseStart, gpkjSubmissionEnd, adminHandler)

	disputeMissingGPKjStart := gpkjSubmissionEnd
	disputeMissingGPKjEnd := disputeMissingGPKjStart + state.PhaseLength
	disputeMissingGPKjTask := dkgtasks.NewDisputeMissingGPKjTask(state, disputeMissingGPKjStart, disputeMissingGPKjEnd)
	disputeGPKjTask := dkgtasks.NewDisputeGPKjTask(state, disputeMissingGPKjStart, disputeMissingGPKjEnd)

	return gpkjSubmissionTask, disputeMissingGPKjTask, disputeGPKjTask
}

func ProcessGPKJSubmissionComplete(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, log types.Log, cdb *db.Database, taskRequestChan chan<- executorInterfaces.ITask, taskKillChan chan<- string) error {

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
		dkgState, err := objects.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		if !dkgState.IsValidator {
			return ErrNotAValidator
		}

		disputeGPKjTask, completionTask = UpdateStateOnGPKJSubmissionComplete(dkgState, event.BlockNumber.Uint64())
		err = objects.PersistEthDkgState(txn, logger, dkgState)
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
	taskKillChan <- dkgtasks.GPKjSubmissionTaskName
	taskKillChan <- dkgtasks.DisputeMissingGPKjTaskName
	taskKillChan <- dkgtasks.DisputeGPKjTaskName

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

func UpdateStateOnGPKJSubmissionComplete(state *dkgObjects.DkgState, disputeGPKjStartBlock uint64) (*dkgtasks.DisputeGPKjTask, *dkgtasks.CompletionTask) {
	state.OnGPKJSubmissionComplete(disputeGPKjStartBlock)

	disputeGPKjPhaseEnd := state.PhaseStart + state.PhaseLength
	disputeGPKjTask := dkgtasks.NewDisputeGPKjTask(state, state.PhaseStart, disputeGPKjPhaseEnd)

	completionStart := disputeGPKjPhaseEnd
	completionEnd := completionStart + state.PhaseLength
	completionTask := dkgtasks.NewCompletionTask(state, completionStart, completionEnd)

	return disputeGPKjTask, completionTask
}
