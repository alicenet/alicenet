package events

import (
	"context"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/executor"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	dkgtasks "github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	monitorInterfaces "github.com/alicenet/alicenet/layer1/monitor/interfaces"
	"github.com/alicenet/alicenet/layer1/monitor/objects"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

func isValidator(acct accounts.Account, state *objects.MonitorState) bool {
	_, present := state.PotentialValidators[acct.Address]
	return present
}

func ProcessRegistrationOpened(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, log types.Log, monState *objects.MonitorState, monDB *db.Database, taskHandler executor.TaskHandler) error {
	logEntry := logger.WithField("eventProcessor", "ProcessRegistrationOpened")
	logEntry.Info("processing registration")
	event, err := contracts.EthereumContracts().Ethdkg().ParseRegistrationOpened(log)
	if err != nil {
		return err
	}

	// get potential validators
	var validatorAddresses []common.Address
	for address := range monState.PotentialValidators {
		validatorAddresses = append(validatorAddresses, address)
	}

	dkgState, registrationTask, disputeMissingRegistrationTask := UpdateStateOnRegistrationOpened(
		eth.GetDefaultAccount(),
		event.StartBlock.Uint64(),
		event.PhaseLength.Uint64(),
		event.ConfirmationLength.Uint64(),
		event.Nonce.Uint64(),
		isValidator(eth.GetDefaultAccount(), monState),
		validatorAddresses,
	)

	logEntry.WithFields(logrus.Fields{
		"StartBlock":         event.StartBlock,
		"NumberValidators":   event.NumberValidators,
		"Nonce":              event.Nonce,
		"PhaseLength":        event.PhaseLength,
		"ConfirmationLength": event.ConfirmationLength,
		"RegistrationEnd":    registrationTask.GetEnd(),
	}).Info("ETHDKG RegistrationOpened")

	if err = state.SaveDkgState(monDB, dkgState); err != nil {
		logEntry.Errorf("Failed to save dkgState on ProcessRegistrationOpened: %v", err)
		return err
	}

	if !dkgState.IsValidator {
		logEntry.Trace("not a validator, skipping task schedule")
		return nil
	}

	// schedule Registration
	logEntry.WithFields(logrus.Fields{
		"TaskStart": registrationTask.GetStart(),
		"TaskEnd":   registrationTask.GetEnd(),
	}).Info("Scheduling NewRegisterTask")

	ctx := context.Background()
	if _, err = taskHandler.ScheduleTask(ctx, registrationTask, ""); err != nil {
		return err
	}

	// schedule DisputeRegistration
	logEntry.WithFields(logrus.Fields{
		"TaskStart": disputeMissingRegistrationTask.GetStart(),
		"TaskEnd":   disputeMissingRegistrationTask.GetEnd(),
	}).Info("Scheduling NewDisputeRegistrationTask")

	if _, err = taskHandler.ScheduleTask(ctx, disputeMissingRegistrationTask, ""); err != nil {
		return err
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
	registrationTask := dkgtasks.NewRegisterTask(dkgState.PhaseStart, registrationEnds)
	disputeMissingRegistrationTask := dkgtasks.NewDisputeMissingRegistrationTask(registrationEnds, registrationEnds+dkgState.PhaseLength)

	return dkgState, registrationTask, disputeMissingRegistrationTask
}

func ProcessAddressRegistered(contracts layer1.AllSmartContracts, logger *logrus.Entry, log types.Log, monDB *db.Database) error {
	logEntry := logger.WithField("eventProcessor", "ProcessAddressRegistered")
	logEntry.Info("processing address registered")

	event, err := contracts.EthereumContracts().Ethdkg().ParseAddressRegistered(log)
	if err != nil {
		return err
	}

	dkgState, err := state.GetDkgState(monDB)
	if err != nil {
		logEntry.Errorf("Failed to load dkgState on ProcessAddressRegistered: %v", err)
		return err
	}

	logEntry.WithFields(logrus.Fields{
		"Account":       event.Account.Hex(),
		"Index":         event.Index,
		"numRegistered": event.Index,
		"Nonce":         event.Nonce,
		"PublicKey":     event.PublicKey,
		"#Participants": len(dkgState.Participants),
		"#Validators":   len(dkgState.ValidatorAddresses),
	}).Info("Address registered!")

	dkgState.OnAddressRegistered(event.Account, int(event.Index.Int64()), event.Nonce.Uint64(), event.PublicKey)

	if err = state.SaveDkgState(monDB, dkgState); err != nil {
		logEntry.Errorf("Failed to save dkgState on ProcessAddressRegistered: %v", err)
		return err
	}

	return nil
}

func ProcessRegistrationComplete(contracts layer1.AllSmartContracts, logger *logrus.Entry, log types.Log, monDB *db.Database, taskHandler executor.TaskHandler) error {
	logEntry := logger.WithField("eventProcessor", "ProcessRegistrationComplete")
	logEntry.Info("processing registration complete")

	shareDistributionTask := &dkgtasks.ShareDistributionTask{}
	disputeMissingShareDistributionTask := &dkgtasks.DisputeMissingShareDistributionTask{}

	dkgState, err := state.GetDkgState(monDB)
	if err != nil {
		logEntry.Errorf("Failed to load dkgState on ProcessRegistrationComplete: %v", err)
		return err
	}

	if !dkgState.IsValidator {
		logEntry.Trace("not a validator, skipping task schedule")
		return nil
	}

	event, err := contracts.EthereumContracts().Ethdkg().ParseRegistrationComplete(log)
	if err != nil {
		return err
	}

	logEntry.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
	}).Info("ETHDKG Registration Complete")

	shareDistributionTask, disputeMissingShareDistributionTask, disputeBadSharesTasks := UpdateStateOnRegistrationComplete(dkgState, event.BlockNumber.Uint64())

	if err = state.SaveDkgState(monDB, dkgState); err != nil {
		logEntry.Errorf("Failed to save dkgState on ProcessRegistrationComplete: %v", err)
		return err
	}

	//Killing previous tasks
	ctx := context.Background()
	if _, err = taskHandler.KillTaskByType(ctx, &dkg.RegisterTask{}); err != nil {
		return err
	}
	if _, err = taskHandler.KillTaskByType(ctx, &dkg.DisputeMissingRegistrationTask{}); err != nil {
		return err
	}

	// schedule ShareDistribution phase
	logEntry.WithFields(logrus.Fields{
		"TaskStart": shareDistributionTask.GetStart(),
		"TaskEnd":   shareDistributionTask.GetEnd(),
	}).Info("Scheduling NewShareDistributionTask")

	if _, err = taskHandler.ScheduleTask(ctx, shareDistributionTask, ""); err != nil {
		return err
	}

	// schedule DisputeParticipantDidNotDistributeSharesTask
	logEntry.WithFields(logrus.Fields{
		"TaskStart": disputeMissingShareDistributionTask.GetStart(),
		"TaskEnd":   disputeMissingShareDistributionTask.GetEnd(),
	}).Info("Scheduling NewDisputeParticipantDidNotDistributeSharesTask")

	if _, err = taskHandler.ScheduleTask(ctx, disputeMissingShareDistributionTask, ""); err != nil {
		return err
	}

	for _, disputeBadSharesTask := range disputeBadSharesTasks {
		// schedule DisputeDistributeSharesTask
		logEntry.WithFields(logrus.Fields{
			"TaskStart": disputeBadSharesTask.GetStart(),
			"TaskEnd":   disputeBadSharesTask.GetEnd(),
			"Address":   disputeBadSharesTask.Address,
		}).Info("Scheduling NewDisputeDistributeSharesTask")
		if _, err = taskHandler.ScheduleTask(ctx, disputeBadSharesTask, ""); err != nil {
			return err
		}
	}

	return nil
}

func UpdateStateOnRegistrationComplete(dkgState *state.DkgState, shareDistributionStartBlockNumber uint64) (*dkgtasks.ShareDistributionTask, *dkgtasks.DisputeMissingShareDistributionTask, []*dkgtasks.DisputeShareDistributionTask) {
	dkgState.OnRegistrationComplete(shareDistributionStartBlockNumber)

	shareDistStartBlock := dkgState.PhaseStart
	shareDistEndBlock := shareDistStartBlock + dkgState.PhaseLength
	shareDistributionTask := dkgtasks.NewShareDistributionTask(shareDistStartBlock, shareDistEndBlock)

	dispShareStartBlock := shareDistEndBlock
	dispShareEndBlock := dispShareStartBlock + dkgState.PhaseLength
	disputeMissingShareDistributionTask := dkgtasks.NewDisputeMissingShareDistributionTask(dispShareStartBlock, dispShareEndBlock)
	disputeBadSharesTasks := GetDisputeShareDistributionTasks(dkgState, dispShareStartBlock, dispShareEndBlock)

	return shareDistributionTask, disputeMissingShareDistributionTask, disputeBadSharesTasks
}

func ProcessShareDistribution(contracts layer1.AllSmartContracts, logger *logrus.Entry, log types.Log, monDB *db.Database) error {
	logEntry := logger.WithField("eventProcessor", "ProcessShareDistribution")
	logEntry.Info("processing share distribution")

	event, err := contracts.EthereumContracts().Ethdkg().ParseSharesDistributed(log)
	if err != nil {
		return err
	}

	logEntry.WithFields(logrus.Fields{
		"Issuer":          event.Account.Hex(),
		"Index":           event.Index,
		"EncryptedShares": event.EncryptedShares,
		"Commitments":     event.Commitments,
	}).Info("Received share distribution")

	dkgState, err := state.GetDkgState(monDB)
	if err != nil {
		logEntry.Errorf("Failed to save dkgState on ProcessShareDistribution: %v", err)
		return err
	}

	if err = dkgState.OnSharesDistributed(logEntry, event.Account, event.EncryptedShares, event.Commitments); err != nil {
		return err
	}

	if err = state.SaveDkgState(monDB, dkgState); err != nil {
		logEntry.Errorf("Failed to save dkgState on ProcessShareDistribution: %v", err)
		return err
	}

	return nil
}

func ProcessShareDistributionComplete(contracts layer1.AllSmartContracts, logger *logrus.Entry, log types.Log, monDB *db.Database, taskHandler executor.TaskHandler) error {
	logEntry := logger.WithField("eventProcessor", "ProcessShareDistributionComplete")
	logEntry.Info("processing share distribution complete")

	keyShareSubmissionTask := &dkgtasks.KeyShareSubmissionTask{}
	disputeMissingKeySharesTask := &dkgtasks.DisputeMissingKeySharesTask{}

	dkgState, err := state.GetDkgState(monDB)
	if err != nil {
		logEntry.Errorf("Failed to load dkgState on ProcessShareDistributionCompleted: %v", err)
		return err
	}

	if !dkgState.IsValidator {
		logEntry.Trace("not a validator, skipping task schedule")
		return nil
	}

	event, err := contracts.EthereumContracts().Ethdkg().ParseShareDistributionComplete(log)
	if err != nil {
		return err
	}

	logEntry.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
	}).Info("Received share distribution complete")

	disputeShareDistributionTasks, keyShareSubmissionTask, disputeMissingKeySharesTask := UpdateStateOnShareDistributionComplete(dkgState, event.BlockNumber.Uint64())
	if err = state.SaveDkgState(monDB, dkgState); err != nil {
		logEntry.Errorf("Failed to save dkgState on ProcessShareDistributionComplete: %v", err)
		return err
	}

	//Killing previous tasks
	ctx := context.Background()
	if _, err = taskHandler.KillTaskByType(ctx, &dkg.ShareDistributionTask{}); err != nil {
		return err
	}
	if _, err = taskHandler.KillTaskByType(ctx, &dkg.DisputeMissingShareDistributionTask{}); err != nil {
		return err
	}
	if _, err = taskHandler.KillTaskByType(ctx, &dkg.DisputeShareDistributionTask{}); err != nil {
		return err
	}

	for _, disputeShareDistributionTask := range disputeShareDistributionTasks {
		// schedule DisputeShareDistributionTask
		logEntry.WithFields(logrus.Fields{
			"TaskStart": disputeShareDistributionTask.GetStart(),
			"TaskEnd":   disputeShareDistributionTask.GetEnd(),
			"Address":   disputeShareDistributionTask.Address,
		}).Info("Scheduling NewDisputeShareDistributionTask")
		if _, err = taskHandler.ScheduleTask(ctx, disputeShareDistributionTask, ""); err != nil {
			return err
		}
	}

	// schedule SubmitKeySharesPhase
	logEntry.WithFields(logrus.Fields{
		"TaskStart": keyShareSubmissionTask.GetStart(),
		"TaskEnd":   keyShareSubmissionTask.GetEnd(),
	}).Info("Scheduling NewKeyShareSubmissionTask")
	if _, err = taskHandler.ScheduleTask(ctx, keyShareSubmissionTask, ""); err != nil {
		return err
	}

	// schedule DisputeMissingKeySharesPhase
	logEntry.WithFields(logrus.Fields{
		"TaskStart": disputeMissingKeySharesTask.GetStart(),
		"TaskEnd":   disputeMissingKeySharesTask.GetEnd(),
	}).Info("Scheduling NewDisputeMissingKeySharesTask")
	if _, err = taskHandler.ScheduleTask(ctx, disputeMissingKeySharesTask, ""); err != nil {
		return err
	}

	return nil
}

func UpdateStateOnShareDistributionComplete(dkgState *state.DkgState, disputeShareDistributionStartBlock uint64) ([]*dkgtasks.DisputeShareDistributionTask, *dkgtasks.KeyShareSubmissionTask, *dkgtasks.DisputeMissingKeySharesTask) {
	dkgState.OnShareDistributionComplete(disputeShareDistributionStartBlock)

	phaseEnd := dkgState.PhaseStart + dkgState.PhaseLength

	disputeShareDistributionTasks := GetDisputeShareDistributionTasks(dkgState, dkgState.PhaseStart, phaseEnd)
	// schedule SubmitKeySharesPhase
	submitKeySharesPhaseStart := phaseEnd
	submitKeySharesPhaseEnd := submitKeySharesPhaseStart + dkgState.PhaseLength
	keyshareSubmissionTask := dkgtasks.NewKeyShareSubmissionTask(submitKeySharesPhaseStart, submitKeySharesPhaseEnd)

	// schedule DisputeMissingKeySharesPhase
	missingKeySharesDisputeStart := submitKeySharesPhaseEnd
	missingKeySharesDisputeEnd := missingKeySharesDisputeStart + dkgState.PhaseLength
	disputeMissingKeySharesTask := dkgtasks.NewDisputeMissingKeySharesTask(missingKeySharesDisputeStart, missingKeySharesDisputeEnd)

	return disputeShareDistributionTasks, keyshareSubmissionTask, disputeMissingKeySharesTask
}

func ProcessKeyShareSubmitted(contracts layer1.AllSmartContracts, logger *logrus.Entry, log types.Log, monDB *db.Database) error {
	logEntry := logger.WithField("eventProcessor", "ProcessKeyShareSubmitted")
	logEntry.Info("processing key share submission")

	event, err := contracts.EthereumContracts().Ethdkg().ParseKeyShareSubmitted(log)
	if err != nil {
		return err
	}

	logEntry.WithFields(logrus.Fields{
		"Issuer":                     event.Account.Hex(),
		"KeyShareG1":                 event.KeyShareG1,
		"KeyShareG1CorrectnessProof": event.KeyShareG1CorrectnessProof,
		"KeyShareG2":                 event.KeyShareG2,
	}).Info("Received key shares")

	dkgState, err := state.GetDkgState(monDB)
	if err != nil {
		logEntry.Errorf("Failed to load dkgState on ProcessKeyShareSubmitted: %v", err)
		return err
	}

	dkgState.OnKeyShareSubmitted(event.Account, event.KeyShareG1, event.KeyShareG1CorrectnessProof, event.KeyShareG2)
	if err = state.SaveDkgState(monDB, dkgState); err != nil {
		logEntry.Errorf("Failed to save dkgState on ProcessKeyShareSubmitted: %v", err)
		return err
	}

	return nil
}

func ProcessKeyShareSubmissionComplete(contracts layer1.AllSmartContracts, logger *logrus.Entry, log types.Log, monDB *db.Database, taskHandler executor.TaskHandler) error {
	logEntry := logger.WithField("eventProcessor", "ProcessKeyShareSubmissionComplete")
	logEntry.Info("processing key share submission complete")

	event, err := contracts.EthereumContracts().Ethdkg().ParseKeyShareSubmissionComplete(log)
	if err != nil {
		return err
	}

	logEntry.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
	}).Info("ProcessKeyShareSubmissionComplete() ...")

	mpkSubmissionTask := &dkgtasks.MPKSubmissionTask{}
	dkgState, err := state.GetDkgState(monDB)
	if err != nil {
		logEntry.Errorf("Failed to load dkgState on ProcessKeyShareSubmissionComplete: %v", err)
		return err
	}

	if !dkgState.IsValidator {
		logEntry.Trace("not a validator, skipping task schedule")
		return nil
	}

	// schedule MPK submission
	mpkSubmissionTask = UpdateStateOnKeyShareSubmissionComplete(dkgState, event.BlockNumber.Uint64())

	if err = state.SaveDkgState(monDB, dkgState); err != nil {
		logEntry.Errorf("Failed to save dkgState on ProcessKeyShareSubmissionComplete: %v", err)
		return err
	}

	//Killing previous tasks
	ctx := context.Background()
	if _, err = taskHandler.KillTaskByType(ctx, &dkg.KeyShareSubmissionTask{}); err != nil {
		return err
	}
	if _, err = taskHandler.KillTaskByType(ctx, &dkg.DisputeMissingKeySharesTask{}); err != nil {
		return err
	}

	// schedule MPKSubmissionTask
	logEntry.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"TaskStart":   mpkSubmissionTask.GetStart(),
		"TaskEnd":     mpkSubmissionTask.GetEnd(),
	}).Info("Scheduling MPKSubmissionTask")
	if _, err = taskHandler.ScheduleTask(ctx, mpkSubmissionTask, ""); err != nil {
		return err
	}

	return nil
}

func UpdateStateOnKeyShareSubmissionComplete(dkgState *state.DkgState, mpkSubmissionStartBlock uint64) *dkgtasks.MPKSubmissionTask {
	dkgState.OnKeyShareSubmissionComplete(mpkSubmissionStartBlock)

	phaseEnd := dkgState.PhaseStart + dkgState.PhaseLength
	mpkSubmissionTask := dkgtasks.NewMPKSubmissionTask(dkgState.PhaseStart, phaseEnd)

	return mpkSubmissionTask
}

func ProcessMPKSet(contracts layer1.AllSmartContracts, logger *logrus.Entry, log types.Log, adminHandler monitorInterfaces.AdminHandler, monDB *db.Database, taskHandler executor.TaskHandler) error {
	logEntry := logger.WithField("eventProcessor", "ProcessMPKSet")
	logEntry.Info("processing master public key set")

	event, err := contracts.EthereumContracts().Ethdkg().ParseMPKSet(log)
	if err != nil {
		return err
	}

	logEntry.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"Nonce":       event.Nonce,
		"MPK":         event.Mpk,
	}).Info("ProcessMPKSet() ...")

	gpkjSubmissionTask := &dkgtasks.GPKjSubmissionTask{}
	disputeMissingGPKjTask := &dkgtasks.DisputeMissingGPKjTask{}

	dkgState, err := state.GetDkgState(monDB)
	if err != nil {
		logEntry.Errorf("Failed to load dkgState on ProcessMPKSet: %v", err)
		return err
	}

	if !dkgState.IsValidator {
		logEntry.Trace("not a validator, skipping task schedule")
		return nil
	}

	gpkjSubmissionTask, disputeMissingGPKjTask, disputeGPKjTasks := UpdateStateOnMPKSet(dkgState, event.BlockNumber.Uint64(), adminHandler)

	if err = state.SaveDkgState(monDB, dkgState); err != nil {
		logEntry.Errorf("Failed to save dkgState on ProcessMPKSet: %v", err)
		return err
	}

	//Killing previous tasks
	ctx := context.Background()
	if _, err = taskHandler.KillTaskByType(ctx, &dkg.MPKSubmissionTask{}); err != nil {
		return err
	}

	// schedule GPKJSubmissionTask
	logEntry.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"TaskStart":   gpkjSubmissionTask.GetStart(),
		"TaskEnd":     gpkjSubmissionTask.GetEnd(),
	}).Info("Scheduling GPKJSubmissionTask")
	if _, err = taskHandler.ScheduleTask(ctx, gpkjSubmissionTask, ""); err != nil {
		return err
	}

	// schedule DisputeMissingGPKjTask
	logEntry.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"TaskStart":   gpkjSubmissionTask.GetStart(),
		"TaskEnd":     gpkjSubmissionTask.GetEnd(),
	}).Info("Scheduling DisputeMissingGPKjTask")
	if _, err = taskHandler.ScheduleTask(ctx, disputeMissingGPKjTask, ""); err != nil {
		return err
	}

	// schedule DisputeGPKjTask
	for _, disputeGPKjTask := range disputeGPKjTasks {
		logEntry.WithFields(logrus.Fields{
			"BlockNumber": event.BlockNumber,
			"TaskStart":   disputeGPKjTask.GetStart(),
			"TaskEnd":     disputeGPKjTask.GetEnd(),
		}).Info("Scheduling DisputeGPKjTask")
		if _, err = taskHandler.ScheduleTask(ctx, disputeGPKjTask, ""); err != nil {
			return err
		}
	}

	return nil
}

func UpdateStateOnMPKSet(dkgState *state.DkgState, gpkjSubmissionStartBlock uint64, adminHandler monitorInterfaces.AdminHandler) (*dkgtasks.GPKjSubmissionTask, *dkgtasks.DisputeMissingGPKjTask, []*dkgtasks.DisputeGPKjTask) {
	dkgState.OnMPKSet(gpkjSubmissionStartBlock)
	gpkjSubmissionEnd := dkgState.PhaseStart + dkgState.PhaseLength
	gpkjSubmissionTask := dkgtasks.NewGPKjSubmissionTask(dkgState.PhaseStart, gpkjSubmissionEnd, adminHandler)

	disputeMissingGPKjStart := gpkjSubmissionEnd
	disputeMissingGPKjEnd := disputeMissingGPKjStart + dkgState.PhaseLength
	disputeMissingGPKjTask := dkgtasks.NewDisputeMissingGPKjTask(disputeMissingGPKjStart, disputeMissingGPKjEnd)
	disputeGPKjTasks := GetDisputeGPKjTasks(dkgState, disputeMissingGPKjStart, disputeMissingGPKjEnd)

	return gpkjSubmissionTask, disputeMissingGPKjTask, disputeGPKjTasks
}

func ProcessGPKJSubmissionComplete(contracts layer1.AllSmartContracts, logger *logrus.Entry, log types.Log, monDB *db.Database, taskHandler executor.TaskHandler) error {
	logEntry := logger.WithField("eventProcessor", "ProcessGPKJSubmissionComplete")
	logEntry.Info("processing gpkj submission complete")
	event, err := contracts.EthereumContracts().Ethdkg().ParseGPKJSubmissionComplete(log)
	if err != nil {
		return err
	}

	logEntry.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
	}).Info("ProcessGPKJSubmissionComplete() ...")

	completionTask := &dkgtasks.CompletionTask{}
	dkgState, err := state.GetDkgState(monDB)
	if err != nil {
		logEntry.Errorf("Failed to load dkgState on ProcessGPKJSubmissionComplete: %v", err)
		return err
	}

	if !dkgState.IsValidator {
		logEntry.Trace("not a validator, skipping task schedule")
		return nil
	}

	disputeGPKjTasks, completionTask := UpdateStateOnGPKJSubmissionComplete(dkgState, event.BlockNumber.Uint64())

	if err = state.SaveDkgState(monDB, dkgState); err != nil {
		logEntry.Errorf("Failed to save dkgState on ProcessGPKJSubmissionComplete: %v", err)
		return err
	}

	//Killing previous tasks
	ctx := context.Background()
	if _, err = taskHandler.KillTaskByType(ctx, &dkg.GPKjSubmissionTask{}); err != nil {
		return err
	}
	if _, err = taskHandler.KillTaskByType(ctx, &dkg.DisputeMissingGPKjTask{}); err != nil {
		return err
	}
	if _, err = taskHandler.KillTaskByType(ctx, &dkg.DisputeGPKjTask{}); err != nil {
		return err
	}

	for _, disputeGPKjTask := range disputeGPKjTasks {
		// schedule DisputeGPKJSubmissionTask
		logEntry.WithFields(logrus.Fields{
			"BlockNumber": event.BlockNumber,
			"TaskStart":   disputeGPKjTask.GetStart(),
			"TaskEnd":     disputeGPKjTask.GetEnd(),
			"Address":     disputeGPKjTask.Address,
		}).Info("Scheduling NewGPKJDisputeTask")
		if _, err = taskHandler.ScheduleTask(ctx, disputeGPKjTask, ""); err != nil {
			return err
		}
	}

	// schedule Completion
	logEntry.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"TaskStart":   completionTask.GetStart(),
		"TaskEnd":     completionTask.GetEnd(),
	}).Info("Scheduling NewCompletionTask")
	if _, err = taskHandler.ScheduleTask(ctx, completionTask, ""); err != nil {
		return err
	}

	return nil
}

func UpdateStateOnGPKJSubmissionComplete(dkgState *state.DkgState, disputeGPKjStartBlock uint64) ([]*dkgtasks.DisputeGPKjTask, *dkgtasks.CompletionTask) {
	dkgState.OnGPKJSubmissionComplete(disputeGPKjStartBlock)

	disputeGPKjPhaseEnd := dkgState.PhaseStart + dkgState.PhaseLength

	disputeGPKjTasks := GetDisputeGPKjTasks(dkgState, dkgState.PhaseStart, disputeGPKjPhaseEnd)
	completionStart := disputeGPKjPhaseEnd
	completionEnd := completionStart + dkgState.PhaseLength
	completionTask := dkgtasks.NewCompletionTask(completionStart, completionEnd)

	return disputeGPKjTasks, completionTask
}

func GetDisputeShareDistributionTasks(dkgState *state.DkgState, phaseStart, phaseEnd uint64) []*dkgtasks.DisputeShareDistributionTask {
	var disputeShareDistributionTasks []*dkgtasks.DisputeShareDistributionTask
	for address := range dkgState.Participants {
		disputeShareDistributionTasks = append(disputeShareDistributionTasks, dkgtasks.NewDisputeShareDistributionTask(phaseStart, phaseEnd, address))
	}
	return disputeShareDistributionTasks
}

func GetDisputeGPKjTasks(dkgState *state.DkgState, phaseStart, phaseEnd uint64) []*dkgtasks.DisputeGPKjTask {
	var disputeGPKjTasks []*dkgtasks.DisputeGPKjTask
	for address := range dkgState.Participants {
		disputeGPKjTasks = append(disputeGPKjTasks, dkgtasks.NewDisputeGPKjTask(phaseStart, phaseEnd, address))
	}
	return disputeGPKjTasks
}
