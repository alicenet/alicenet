package dkgevents

import (
	"context"
	"errors"
	"time"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// todo: improve this
func isValidator(eth interfaces.Ethereum, logger *logrus.Entry, acct accounts.Account) (bool, error) {
	ctx := context.Background()
	callOpts := eth.GetCallOpts(ctx, acct)
	isValidator, err := eth.Contracts().ValidatorPool().IsValidator(callOpts, acct.Address)
	if err != nil {
		return false, errors.New("cannot check if I'm a validator")
	}

	if !isValidator {
		logger.Info("cannot take part in ETHDKG because I'm not a validator")
		return false, nil
	}

	return true, nil
}

func ProcessRegistrationOpened(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	logger.Info("ProcessRegistrationOpened() ...")
	event, err := eth.Contracts().Ethdkg().ParseRegistrationOpened(log)
	if err != nil {
		return err
	}

	state.Lock()
	defer state.Unlock()

	amIaValidator, err := isValidator(eth, logger, state.EthDKG.Account)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "I'm not a validator: %v", err)
	}

	// get validators from ValidatorPool
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()

	callOpts := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	validatorAddresses, err := dkg.GetValidatorAddressesFromPool(callOpts, eth, logger)
	if err != nil {
		return err
		// return dkg.LogReturnErrorf(logger, "ProcessRegistrationOpened(): Unable to get validatorAddresses from ValidatorPool: %v", err)
	}

	dkgState, registrationEnds, registrationTask, disputeMissingRegistrationTask := UpdateStateOnRegistrationOpened(
		state.EthDKG.Account,
		event.StartBlock.Uint64(),
		event.PhaseLength.Uint64(),
		event.ConfirmationLength.Uint64(),
		event.Nonce.Uint64(),
		amIaValidator,
		validatorAddresses,
	)
	state.EthDKG = dkgState

	logger.WithFields(logrus.Fields{
		"StartBlock":         event.StartBlock,
		"NumberValidators":   event.NumberValidators,
		"Nonce":              event.Nonce,
		"PhaseLength":        event.PhaseLength,
		"ConfirmationLength": event.ConfirmationLength,
		"RegistrationEnd":    registrationEnds,
	}).Info("ETHDKG RegistrationOpened")

	logger.WithFields(logrus.Fields{
		"Phase": dkgState.Phase,
	}).Infof("Purging Schedule")
	state.Schedule.Purge()

	if !state.EthDKG.IsValidator {
		return nil
	}

	// schedule Registration
	logger.WithFields(logrus.Fields{
		"PhaseStart": state.EthDKG.PhaseStart,
		"PhaseEnd":   registrationEnds,
	}).Info("Scheduling NewRegisterTask")

	_, err = state.Schedule.Schedule(dkgState.PhaseStart, registrationEnds, registrationTask)
	if err != nil {
		return err
	}

	// schedule DisputeRegistration
	logger.WithFields(logrus.Fields{
		"PhaseStart": registrationEnds,
		"PhaseEnd":   registrationEnds + dkgState.PhaseLength,
	}).Info("Scheduling NewDisputeRegistrationTask")

	_, err = state.Schedule.Schedule(registrationEnds, registrationEnds+dkgState.PhaseLength, disputeMissingRegistrationTask)
	if err != nil {
		return err
	}

	state.Schedule.Status(logger)

	return nil
}

func UpdateStateOnRegistrationOpened(account accounts.Account, startBlock, phaseLength, confirmationLength, nonce uint64, amIValidator bool, validatorAddresses []common.Address) (*objects.DkgState, uint64, *dkgtasks.RegisterTask, *dkgtasks.DisputeMissingRegistrationTask) {
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

	return dkgState, registrationEnds, registrationTask, disputeMissingRegistrationTask
}

func ProcessAddressRegistered(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	logger.Info("ProcessAddressRegistered() ...")

	event, err := eth.Contracts().Ethdkg().ParseAddressRegistered(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"Account":       event.Account.Hex(),
		"Index":         event.Index,
		"numRegistered": event.Index,
		"Nonce":         event.Nonce,
		"PublicKey":     event.PublicKey,
		"#Participants": len(state.EthDKG.Participants),
		"#Validators":   len(state.EthDKG.ValidatorAddresses),
	}).Info("Address registered!")

	state.Lock()
	defer state.Unlock()

	state.EthDKG.OnAddressRegistered(event.Account, int(event.Index.Int64()), event.Nonce.Uint64(), event.PublicKey)

	return nil
}

func ProcessRegistrationComplete(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	logger.Info("ProcessRegistrationComplete() ...")

	if !state.EthDKG.IsValidator {
		return nil
	}

	event, err := eth.Contracts().Ethdkg().ParseRegistrationComplete(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
	}).Info("ETHDKG Registration Complete")

	state.Lock()
	defer state.Unlock()

	shareDistributionTask, shareDistributionStart, shareDistributionEnd, disputeMissingShareDistributionTask, disputeBadSharesTask, disputeStart, disputeEnd := UpdateStateOnRegistrationComplete(state.EthDKG, event.BlockNumber.Uint64())

	logger.WithFields(logrus.Fields{
		"Phase": state.EthDKG.Phase,
	}).Info("Purging schedule")
	state.Schedule.Purge()

	// schedule ShareDistribution phase
	logger.WithFields(logrus.Fields{
		"PhaseStart": shareDistributionStart,
		"PhaseEnd":   shareDistributionEnd,
	}).Info("Scheduling NewShareDistributionTask")

	_, err = state.Schedule.Schedule(shareDistributionStart, shareDistributionEnd, shareDistributionTask)
	if err != nil {
		return err
	}

	// schedule DisputeParticipantDidNotDistributeSharesTask
	logger.WithFields(logrus.Fields{
		"PhaseStart": disputeStart,
		"PhaseEnd":   disputeEnd,
	}).Info("Scheduling NewDisputeParticipantDidNotDistributeSharesTask")

	_, err = state.Schedule.Schedule(disputeStart, disputeEnd, disputeMissingShareDistributionTask)
	if err != nil {
		return err
	}

	// schedule DisputeDistributeSharesTask
	logger.WithFields(logrus.Fields{
		"PhaseStart": disputeStart,
		"PhaseEnd":   disputeEnd,
	}).Info("Scheduling NewDisputeDistributeSharesTask")

	_, err = state.Schedule.Schedule(disputeStart, disputeEnd, disputeBadSharesTask)
	if err != nil {
		return err
	}

	return nil
}

func UpdateStateOnRegistrationComplete(state *objects.DkgState, shareDistributionStartBlockNumber uint64) (*dkgtasks.ShareDistributionTask, uint64, uint64, *dkgtasks.DisputeMissingShareDistributionTask, *dkgtasks.DisputeShareDistributionTask, uint64, uint64) {
	state.OnRegistrationComplete(shareDistributionStartBlockNumber)

	shareDistStartBlock := state.PhaseStart
	shareDistEndBlock := shareDistStartBlock + state.PhaseLength

	// schedule ShareDistribution phase
	shareDistributionTask := dkgtasks.NewShareDistributionTask(state, shareDistStartBlock, shareDistEndBlock)

	// schedule DisputeParticipantDidNotDistributeSharesTask
	var dispShareStartBlock = shareDistEndBlock
	var dispShareEndBlock = dispShareStartBlock + state.PhaseLength

	disputeMissingShareDistributionTask := dkgtasks.NewDisputeMissingShareDistributionTask(state, dispShareStartBlock, dispShareEndBlock)

	// schedule DisputeShareDistributionTask
	disputeBadSharesTask := dkgtasks.NewDisputeShareDistributionTask(state, dispShareStartBlock, dispShareEndBlock)

	return shareDistributionTask, shareDistStartBlock, shareDistEndBlock, disputeMissingShareDistributionTask, disputeBadSharesTask, dispShareStartBlock, dispShareEndBlock
}

func ProcessShareDistribution(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

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

	state.Lock()
	defer state.Unlock()

	return state.EthDKG.OnSharesDistributed(logger, event.Account, event.EncryptedShares, event.Commitments)
}

func ProcessShareDistributionComplete(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	logger.Info("ProcessShareDistributionComplete() ...")

	if !state.EthDKG.IsValidator {
		return nil
	}

	event, err := eth.Contracts().Ethdkg().ParseShareDistributionComplete(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
	}).Info("Received share distribution complete")

	state.Lock()
	defer state.Unlock()

	disputeShareDistributionTask, disputeShareDistributionStart, disputeShareDistributionEnd, keyshareSubmissionTask, submitKeySharesPhaseStart, submitKeySharesPhaseEnd, disputeMissingKeySharesTask, missingKeySharesDisputeStart, missingKeySharesDisputeEnd := UpdateStateOnShareDistributionComplete(state.EthDKG, logger, event.BlockNumber.Uint64())

	logger.WithFields(logrus.Fields{
		"Phase": state.EthDKG.Phase,
	}).Infof("Purging schedule")
	state.Schedule.Purge()

	// schedule DisputeShareDistributionTask
	_, err = state.Schedule.Schedule(disputeShareDistributionStart, disputeShareDistributionEnd, disputeShareDistributionTask)
	if err != nil {
		return err
	}

	// schedule SubmitKeySharesPhase
	_, err = state.Schedule.Schedule(submitKeySharesPhaseStart, submitKeySharesPhaseEnd, keyshareSubmissionTask)
	if err != nil {
		return err
	}
	// schedule DisputeMissingKeySharesPhase
	_, err = state.Schedule.Schedule(missingKeySharesDisputeStart, missingKeySharesDisputeEnd, disputeMissingKeySharesTask)
	if err != nil {
		return err
	}

	return nil
}

func UpdateStateOnShareDistributionComplete(state *objects.DkgState, logger *logrus.Entry, disputeShareDistributionStartBlock uint64) (*dkgtasks.DisputeShareDistributionTask, uint64, uint64, *dkgtasks.KeyshareSubmissionTask, uint64, uint64, *dkgtasks.DisputeMissingKeySharesTask, uint64, uint64) {

	state.OnShareDistributionComplete(disputeShareDistributionStartBlock)
	var phaseEnd uint64 = state.PhaseStart + state.PhaseLength

	//state.Schedule.Schedule(state.PhaseStart, phaseEnd, )
	disputeShareDistributionTask := dkgtasks.NewDisputeShareDistributionTask(state, state.PhaseStart, phaseEnd)

	logger.WithFields(logrus.Fields{
		"BlockNumber": disputeShareDistributionStartBlock,
		"PhaseStart":  state.PhaseStart,
		"phaseEnd":    phaseEnd,
	}).Info("Scheduling DisputeShareDistributionTask")

	// schedule SubmitKeySharesPhase
	var submitKeySharesPhaseStart uint64 = phaseEnd
	var submitKeySharesPhaseEnd = submitKeySharesPhaseStart + state.PhaseLength

	keyshareSubmissionTask := dkgtasks.NewKeyshareSubmissionTask(state, submitKeySharesPhaseStart, submitKeySharesPhaseEnd)

	logger.WithFields(logrus.Fields{
		"BlockNumber": disputeShareDistributionStartBlock,
		"PhaseStart":  submitKeySharesPhaseStart,
		"phaseEnd":    submitKeySharesPhaseEnd,
	}).Info("Scheduling KeyshareSubmissionTask")

	// schedule DisputeMissingKeySharesPhase
	var missingKeySharesDisputeStart uint64 = submitKeySharesPhaseEnd
	var missingKeySharesDisputeEnd = missingKeySharesDisputeStart + state.PhaseLength

	//state.Schedule.Schedule(missingKeySharesDisputeStart, phaseEnd, )
	disputeMissingKeySharesTask := dkgtasks.NewDisputeMissingKeySharesTask(state, missingKeySharesDisputeStart, missingKeySharesDisputeEnd)

	logger.WithFields(logrus.Fields{
		"BlockNumber": disputeShareDistributionStartBlock,
		"PhaseStart":  missingKeySharesDisputeStart,
		"phaseEnd":    missingKeySharesDisputeEnd,
	}).Info("Scheduling DisputeMissingKeySharesTask")

	return disputeShareDistributionTask, state.PhaseStart, phaseEnd, keyshareSubmissionTask, submitKeySharesPhaseStart, submitKeySharesPhaseEnd, disputeMissingKeySharesTask, missingKeySharesDisputeStart, missingKeySharesDisputeEnd
}

func ProcessKeyShareSubmitted(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

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

	state.Lock()
	defer state.Unlock()

	state.EthDKG.OnKeyShareSubmitted(event.Account, event.KeyShareG1, event.KeyShareG1CorrectnessProof, event.KeyShareG2)

	return nil
}

func ProcessKeyShareSubmissionComplete(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	event, err := eth.Contracts().Ethdkg().ParseKeyShareSubmissionComplete(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
	}).Info("ProcessKeyShareSubmissionComplete() ...")

	if !state.EthDKG.IsValidator {
		return nil
	}

	// schedule MPK submission
	state.Lock()
	defer state.Unlock()

	mpkSubmissionTask, phaseStart, phaseEnd := UpdateStateOnKeyShareSubmissionComplete(state.EthDKG, logger, event.BlockNumber.Uint64())

	logger.WithFields(logrus.Fields{
		"Phase": state.EthDKG.Phase,
	}).Infof("Purging schedule")
	state.Schedule.Purge()

	// schedule MPKSubmissionTask
	_,err = state.Schedule.Schedule(phaseStart, phaseEnd, mpkSubmissionTask)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"phaseStart":  phaseStart,
		"phaseEnd":    phaseEnd,
	}).Info("Scheduling MPKSubmissionTask")

	return nil
}

func UpdateStateOnKeyShareSubmissionComplete(state *objects.DkgState, logger *logrus.Entry, mpkSubmissionStartBlock uint64) (*dkgtasks.MPKSubmissionTask, uint64, uint64) {
	state.OnKeyShareSubmissionComplete(mpkSubmissionStartBlock)
	var phaseEnd uint64 = state.PhaseStart + state.PhaseLength

	mpkSubmissionTask := dkgtasks.NewMPKSubmissionTask(state, state.PhaseStart, phaseEnd)

	return mpkSubmissionTask, state.PhaseStart, phaseEnd
}

func ProcessMPKSet(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log, adminHandler interfaces.AdminHandler) error {

	event, err := eth.Contracts().Ethdkg().ParseMPKSet(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"Nonce":       event.Nonce,
		"MPK":         event.Mpk,
	}).Info("ProcessMPKSet() ...")

	if !state.EthDKG.IsValidator {
		return nil
	}

	state.Lock()
	defer state.Unlock()

	gpkjSubmissionTask, gpkjSubmissionStart, gpkjSubmissionEnd, disputeMissingGPKjTask, disputeGPKjTask, disputeMissingGPKjStart, disputeMissingGPKjEnd := UpdateStateOnMPKSet(state.EthDKG, logger, event.BlockNumber.Uint64(), adminHandler)

	logger.WithFields(logrus.Fields{
		"Phase": state.EthDKG.Phase,
	}).Infof("Purging schedule")
	state.Schedule.Purge()

	// schedule GPKJSubmissionTask
	_, err = state.Schedule.Schedule(gpkjSubmissionStart, gpkjSubmissionEnd, gpkjSubmissionTask)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"phaseStart":  gpkjSubmissionStart,
		"phaseEnd":    gpkjSubmissionEnd,
	}).Info("Scheduling GPKJSubmissionTask")

	// schedule DisputeMissingGPKjTask
	_, err = state.Schedule.Schedule(disputeMissingGPKjStart, disputeMissingGPKjEnd, disputeMissingGPKjTask)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"phaseStart":  disputeMissingGPKjStart,
		"phaseEnd":    disputeMissingGPKjEnd,
	}).Info("Scheduling DisputeMissingGPKjTask")

	// schedule DisputeGPKjTask
	_, err = state.Schedule.Schedule(disputeMissingGPKjStart, disputeMissingGPKjEnd, disputeGPKjTask)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"phaseStart":  disputeMissingGPKjStart,
		"phaseEnd":    disputeMissingGPKjEnd,
	}).Info("Scheduling DisputeGPKjTask")

	return nil
}

func UpdateStateOnMPKSet(state *objects.DkgState, logger *logrus.Entry, gpkjSubmissionStartBlock uint64, adminHandler interfaces.AdminHandler) (*dkgtasks.GPKjSubmissionTask, uint64, uint64, *dkgtasks.DisputeMissingGPKjTask, *dkgtasks.DisputeGPKjTask, uint64, uint64) {
	state.OnMPKSet(gpkjSubmissionStartBlock)
	var gpkjSubmissionEnd uint64 = state.PhaseStart + state.PhaseLength

	gpkjSubmissionTask := dkgtasks.NewGPKjSubmissionTask(state, state.PhaseStart, gpkjSubmissionEnd, adminHandler)

	// schedule DisputeMissingGPKjTask
	disputeMissingGPKjStart := gpkjSubmissionEnd
	disputeMissingGPKjEnd := disputeMissingGPKjStart + state.PhaseLength

	disputeMissingGPKjTask := dkgtasks.NewDisputeMissingGPKjTask(state, disputeMissingGPKjStart, disputeMissingGPKjEnd)

	disputeGPKjTask := dkgtasks.NewDisputeGPKjTask(state, disputeMissingGPKjStart, disputeMissingGPKjEnd)

	return gpkjSubmissionTask, state.PhaseStart, gpkjSubmissionEnd, disputeMissingGPKjTask, disputeGPKjTask, disputeMissingGPKjStart, disputeMissingGPKjEnd
}

func ProcessGPKJSubmissionComplete(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	event, err := eth.Contracts().Ethdkg().ParseGPKJSubmissionComplete(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
	}).Info("ProcessGPKJSubmissionComplete() ...")

	if !state.EthDKG.IsValidator {
		return nil
	}

	state.Lock()
	defer state.Unlock()

	disputeGPKjTask, disputeGPKjPhaseStart, disputeGPKjPhaseEnd, completionTask, completionStart, completionEnd := UpdateStateOnGPKJSubmissionComplete(state.EthDKG, logger, event.BlockNumber.Uint64())

	logger.WithFields(logrus.Fields{
		"Phase": state.EthDKG.Phase,
	}).Infof("Purging schedule")
	state.Schedule.Purge()

	// schedule DisputeGPKJSubmissionTask
	_, err = state.Schedule.Schedule(disputeGPKjPhaseStart, disputeGPKjPhaseEnd, disputeGPKjTask)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"phaseStart":  disputeGPKjPhaseStart,
		"phaseEnd":    disputeGPKjPhaseEnd,
	}).Info("Scheduling NewGPKJDisputeTask")

	// schedule Completion
	_, err = state.Schedule.Schedule(completionStart, completionEnd, completionTask)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber":     event.BlockNumber,
		"completionStart": completionStart,
		"phaseEnd":        completionEnd,
	}).Info("Scheduling NewCompletionTask")

	return nil
}

func UpdateStateOnGPKJSubmissionComplete(state *objects.DkgState, logger *logrus.Entry, disputeGPKjStartBlock uint64) (*dkgtasks.DisputeGPKjTask, uint64, uint64, *dkgtasks.CompletionTask, uint64, uint64) {
	state.OnGPKJSubmissionComplete(disputeGPKjStartBlock)
	var disputeGPKjPhaseEnd uint64 = state.PhaseStart + state.PhaseLength

	disputeGPKjTask := dkgtasks.NewDisputeGPKjTask(state, state.PhaseStart, disputeGPKjPhaseEnd)

	// schedule Completion
	var completionStart = disputeGPKjPhaseEnd
	var completionEnd = completionStart + state.PhaseLength

	completionTask := dkgtasks.NewCompletionTask(state, completionStart, completionEnd)

	return disputeGPKjTask, state.PhaseStart, disputeGPKjPhaseEnd, completionTask, completionStart, completionEnd
}
