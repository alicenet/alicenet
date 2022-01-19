package dkgevents

import (
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

func ProcessRegistrationOpened(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	logger.Info("ProcessRegistrationOpened() ...")
	event, err := eth.Contracts().Ethdkg().ParseRegistrationOpened(log)
	if err != nil {
		return err
	}

	state.Lock()
	defer state.Unlock()

	dkgState := state.EthDKG
	dkgState.Phase = objects.RegistrationOpen
	dkgState.PhaseStart = event.StartBlock.Uint64()
	dkgState.PhaseLength = event.PhaseLength.Uint64()
	dkgState.ConfirmationLength = event.ConfirmationLength.Uint64()
	dkgState.NumberOfValidators = event.NumberValidators.Uint64()
	dkgState.Nonce = event.Nonce.Uint64()
	dkgState.Participants = nil

	registrationEnds := dkgState.PhaseStart + dkgState.PhaseLength

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

	// schedule Registration
	logger.WithFields(logrus.Fields{
		"PhaseStart": state.EthDKG.PhaseStart,
		"PhaseEnd":   registrationEnds,
	}).Info("Scheduling NewRegisterTask")

	state.Schedule.Schedule(dkgState.PhaseStart, registrationEnds, dkgtasks.NewRegisterTask(dkgState))

	// schedule DisputeRegistration
	logger.WithFields(logrus.Fields{
		"PhaseStart": registrationEnds,
		"PhaseEnd":   registrationEnds + dkgState.PhaseLength,
	}).Info("Scheduling NewDisputeRegistrationTask")

	state.Schedule.Schedule(registrationEnds, registrationEnds+dkgState.PhaseLength, dkgtasks.NewDisputeMissingRegistrationTask(dkgState))

	state.Schedule.Status(logger)

	return nil
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
		"#Participants": state.EthDKG.Participants.Len(),
		"#Validators":   state.EthDKG.NumberOfValidators,
	}).Info("Address registered!")

	state.Lock()
	defer state.Unlock()

	state.EthDKG.Participants = append(state.EthDKG.Participants, &objects.Participant{
		Address:   event.Account,
		Index:     int(event.Index.Uint64()),
		PublicKey: event.PublicKey,
	})

	// update state.Index with my index, if this event was mine
	if event.Account.String() == state.EthDKG.Account.Address.String() {
		logger.WithFields(logrus.Fields{
			"event.Account": event.Account,
			"state.Account": state.EthDKG.Account.Address,
			"index":         event.Index,
		}).Info("Added my registration")
		state.EthDKG.Index = int(event.Index.Int64())
	}

	return nil
}

func ProcessRegistrationComplete(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	logger.Info("ProcessRegistrationComplete() ...")

	event, err := eth.Contracts().Ethdkg().ParseRegistrationComplete(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
	}).Info("ETHDKG Registration Complete")

	state.Lock()
	defer state.Unlock()

	// set phase
	state.EthDKG.Phase = objects.ShareDistribution
	state.EthDKG.PhaseStart = event.BlockNumber.Uint64() + state.EthDKG.ConfirmationLength
	var phaseEnd = state.EthDKG.PhaseStart + state.EthDKG.PhaseLength

	logger.WithFields(logrus.Fields{
		"Phase": state.EthDKG.Phase,
	}).Infof("Purging schedule")
	state.Schedule.Purge()

	// schedule ShareDistribution phase
	logger.WithFields(logrus.Fields{
		"PhaseStart": state.EthDKG.PhaseStart,
		"PhaseEnd":   phaseEnd,
	}).Info("Scheduling NewShareDistributionTask")

	state.Schedule.Schedule(state.EthDKG.PhaseStart, phaseEnd, dkgtasks.NewShareDistributionTask(state.EthDKG))

	// schedule DisputeParticipantDidNotDistributeSharesTask
	var phaseStart = phaseEnd
	phaseEnd = phaseEnd + state.EthDKG.PhaseLength
	logger.WithFields(logrus.Fields{
		"PhaseStart": phaseStart,
		"PhaseEnd":   phaseEnd,
	}).Info("Scheduling NewDisputeParticipantDidNotDistributeSharesTask")

	state.Schedule.Schedule(phaseStart, phaseEnd, dkgtasks.NewDisputeMissingShareDistributionTask(state.EthDKG))

	return nil
}

func ProcessShareDistribution(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	logger.Info("ProcessShareDistribution() ...")

	event, err := eth.Contracts().Ethdkg().ParseSharesDistributed(log)
	if err != nil {
		return err
	}

	l := logger.WithFields(logrus.Fields{
		"Issuer":          event.Account.Hex(),
		"Index":           event.Index,
		"EncryptedShares": event.EncryptedShares,
	})

	state.Lock()
	defer state.Unlock()

	state.EthDKG.Commitments[event.Account] = event.Commitments
	state.EthDKG.EncryptedShares[event.Account] = event.EncryptedShares

	l.WithFields(logrus.Fields{
		"TotalCommitments":     len(state.EthDKG.Commitments[event.Account]),
		"TotalEncryptedShares": len(state.EthDKG.EncryptedShares[event.Account]),
	}).Info("Received share distribution")

	return nil
}

func ProcessShareDistributionComplete(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	logger.Info("ProcessShareDistributionComplete() ...")

	event, err := eth.Contracts().Ethdkg().ParseShareDistributionComplete(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
	}).Info("Received share distribution complete")

	state.Lock()
	defer state.Unlock()

	state.EthDKG.Phase = objects.DisputeShareDistribution

	logger.WithFields(logrus.Fields{
		"Phase": state.EthDKG.Phase,
	}).Infof("Purging schedule")
	state.Schedule.Purge()

	// schedule DisputeShareDistributionTask
	state.EthDKG.PhaseStart = event.BlockNumber.Uint64() + state.EthDKG.ConfirmationLength
	var phaseEnd uint64 = state.EthDKG.PhaseStart + state.EthDKG.PhaseLength

	state.Schedule.Schedule(state.EthDKG.PhaseStart, phaseEnd, dkgtasks.NewDisputeShareDistributionTask(state.EthDKG))

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"PhaseStart":  state.EthDKG.PhaseStart,
		"phaseEnd":    phaseEnd,
	}).Info("Scheduling DisputeShareDistributionTask")

	// schedule SubmitKeySharesPhase
	var submitKeySharesPhaseStart uint64 = phaseEnd
	phaseEnd = submitKeySharesPhaseStart + state.EthDKG.PhaseLength

	state.Schedule.Schedule(submitKeySharesPhaseStart, phaseEnd, dkgtasks.NewKeyshareSubmissionTask(state.EthDKG))

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
		"PhaseStart":  submitKeySharesPhaseStart,
		"phaseEnd":    phaseEnd,
	}).Info("Scheduling KeyshareSubmissionTask")

	return nil
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

	state.EthDKG.Phase = objects.KeyShareSubmission

	state.EthDKG.KeyShareG1s[event.Account] = event.KeyShareG1
	state.EthDKG.KeyShareG1CorrectnessProofs[event.Account] = event.KeyShareG1CorrectnessProof
	state.EthDKG.KeyShareG2s[event.Account] = event.KeyShareG2

	logger.WithFields(logrus.Fields{
		"#KeyShareG1s":                 len(state.EthDKG.KeyShareG1s[event.Account]),
		"#KeyShareG1CorrectnessProofs": len(state.EthDKG.KeyShareG1CorrectnessProofs[event.Account]),
		"#KeyShareG2s":                 len(state.EthDKG.KeyShareG2s[event.Account]),
	}).Info("Received key shares2")

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

	// schedule MPK submission
	state.Lock()
	defer state.Unlock()

	state.EthDKG.Phase = objects.MPKSubmission

	logger.WithFields(logrus.Fields{
		"Phase": state.EthDKG.Phase,
	}).Infof("Purging schedule")
	state.Schedule.Purge()

	// schedule MPKSubmissionTask
	state.EthDKG.PhaseStart = event.BlockNumber.Uint64() + state.EthDKG.ConfirmationLength
	var phaseEnd uint64 = state.EthDKG.PhaseStart + state.EthDKG.PhaseLength

	state.Schedule.Schedule(state.EthDKG.PhaseStart, phaseEnd, dkgtasks.NewMPKSubmissionTask(state.EthDKG))

	logger.WithFields(logrus.Fields{
		"BlockNumber":             event.BlockNumber,
		"state.EthDKG.PhaseStart": state.EthDKG.PhaseStart,
		"phaseEnd":                phaseEnd,
	}).Info("Scheduling MPKSubmissionTask")

	return nil
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

	// schedule GPKJ Submission phase
	state.Lock()
	defer state.Unlock()

	state.EthDKG.Phase = objects.GPKJSubmission

	logger.WithFields(logrus.Fields{
		"Phase": state.EthDKG.Phase,
	}).Infof("Purging schedule")
	state.Schedule.Purge()

	// schedule GPKJSubmissionTask
	state.EthDKG.PhaseStart = event.BlockNumber.Uint64() // + state.EthDKG.ConfirmationLength + 1
	var phaseEnd uint64 = state.EthDKG.PhaseStart + state.EthDKG.PhaseLength

	state.Schedule.Schedule(state.EthDKG.PhaseStart, phaseEnd, dkgtasks.NewGPKSubmissionTask(state.EthDKG, adminHandler))

	logger.WithFields(logrus.Fields{
		"BlockNumber":             event.BlockNumber,
		"state.EthDKG.PhaseStart": state.EthDKG.PhaseStart,
		"phaseEnd":                phaseEnd,
	}).Info("Scheduling GPKJSubmissionTask")

	return nil
}

func ProcessGPKJSubmissionComplete(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	event, err := eth.Contracts().Ethdkg().ParseGPKJSubmissionComplete(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"BlockNumber": event.BlockNumber,
	}).Info("ProcessGPKJSubmissionComplete() ...")

	// schedule DisputeGPKJSubmission phase
	state.Lock()
	defer state.Unlock()

	state.EthDKG.Phase = objects.DisputeGPKJSubmission

	logger.WithFields(logrus.Fields{
		"Phase": state.EthDKG.Phase,
	}).Infof("Purging schedule")
	state.Schedule.Purge()

	// schedule DisputeGPKJSubmissionTask
	state.EthDKG.PhaseStart = event.BlockNumber.Uint64() + state.EthDKG.ConfirmationLength
	var phaseEnd uint64 = state.EthDKG.PhaseStart + state.EthDKG.PhaseLength

	state.Schedule.Schedule(state.EthDKG.PhaseStart, phaseEnd, dkgtasks.NewGPKJDisputeTask(state.EthDKG))

	logger.WithFields(logrus.Fields{
		"BlockNumber":             event.BlockNumber,
		"state.EthDKG.PhaseStart": state.EthDKG.PhaseStart,
		"phaseEnd":                phaseEnd,
	}).Info("Scheduling NewGPKJDisputeTask")

	// schedule Completion
	var completionStart = phaseEnd
	phaseEnd = completionStart + state.EthDKG.PhaseLength

	state.Schedule.Schedule(completionStart, phaseEnd, dkgtasks.NewCompletionTask(state.EthDKG))

	logger.WithFields(logrus.Fields{
		"BlockNumber":     event.BlockNumber,
		"completionStart": completionStart,
		"phaseEnd":        phaseEnd,
	}).Info("Scheduling NewCompletionTask")

	return nil
}
