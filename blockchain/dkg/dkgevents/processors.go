package dkgevents

import (
	"math/big"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/bridge/bindings"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

func ProcessRegistrationOpened(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	logger.Info("ProcessRegistrationOpened() ...")
	event, err := eth.Contracts().Ethdkg().ParseRegistrationOpened(log)
	if err != nil {
		return err
	}

	registrationEnds := (&big.Int{}).Add(event.StartBlock, event.PhaseLength)

	logger.WithFields(logrus.Fields{
		"StartBlock":         event.StartBlock,
		"NumberValidators":   event.NumberValidators,
		"Nonce":              event.Nonce,
		"PhaseLength":        event.PhaseLength,
		"ConfirmationLength": event.ConfirmationLength,
	}).Info("ETHDKG RegistrationOpened")

	// If we're already finalized before registration ends, there's no point in continuing
	finalizedBig := (&big.Int{}).SetUint64(state.HighestBlockFinalized)
	if registrationEnds.Cmp(finalizedBig) < 1 {
		logger.WithField("RegistrationEnd", registrationEnds.String()).Info("Too late to participate in EthDKG")
		return nil
	}

	// todo: L&R This is bad! Pay attention to this! This will prevent us from restarting ethdkg
	// ETHDKG contract should be one to tell you if there's an ETHDKG going on or not
	if state.EthDKG != nil && state.EthDKG.RegistrationEnd >= registrationEnds.Uint64() {
		logger.WithField("RegistrationEnd", registrationEnds.String()).Info("EthDKG already in progress")
		return nil
	}

	logger.WithField("RegistrationEnd", registrationEnds.String()).Info("Participating in EthDKG")

	state.Lock()
	defer state.Unlock()

	dkgState := state.EthDKG
	dkgState.Phase = objects.Registration
	dkgState.PhaseLength = event.PhaseLength.Uint64()
	dkgState.ConfirmationLength = event.ConfirmationLength.Uint64()
	dkgState.NumberOfValidators = event.NumberValidators.Uint64()
	dkgState.Participants = nil

	logger.WithFields(logrus.Fields{
		"Phase": dkgState.Phase,
	}).Warn("Purging Schedule")
	state.Schedule.Purge()

	//PopulateSchedule(state.EthDKG, event)
	dkgState.RegistrationStart = event.StartBlock.Uint64() + 1
	dkgState.RegistrationEnd = event.StartBlock.Uint64() + event.PhaseLength.Uint64()

	state.Schedule.Schedule(dkgState.RegistrationStart, dkgState.RegistrationEnd, dkgtasks.NewRegisterTask(dkgState)) // Registration
	state.Schedule.Schedule(dkgState.RegistrationEnd, dkgState.RegistrationEnd+dkgState.PhaseLength, dkgtasks.NewDisputeRegistrationTask(dkgState))

	// state.Schedule.Schedule(dkgState.ShareDistributionStart, dkgState.ShareDistributionEnd, dkgtasks.NewShareDistributionTask(dkgState))     // ShareDistribution
	// state.Schedule.Schedule(dkgState.DisputeStart, dkgState.DisputeEnd, dkgtasks.NewDisputeTask(dkgState))                                   // DisputeShares
	// state.Schedule.Schedule(dkgState.KeyShareSubmissionStart, dkgState.KeyShareSubmissionEnd, dkgtasks.NewKeyshareSubmissionTask(dkgState))  // KeyShareSubmission
	// state.Schedule.Schedule(dkgState.MPKSubmissionStart, dkgState.MPKSubmissionEnd, dkgtasks.NewMPKSubmissionTask(dkgState))                 // MasterPublicKeySubmission
	// state.Schedule.Schedule(dkgState.GPKJSubmissionStart, dkgState.GPKJSubmissionEnd, dkgtasks.NewGPKSubmissionTask(dkgState, adminHandler)) // GroupPublicKeySubmission
	// state.Schedule.Schedule(dkgState.GPKJGroupAccusationStart, dkgState.GPKJGroupAccusationEnd, dkgtasks.NewGPKJDisputeTask(dkgState))       // DisputeGroupPublicKey
	// state.Schedule.Schedule(dkgState.CompleteStart, dkgState.CompleteEnd, dkgtasks.NewCompletionTask(dkgState))                              // Complete

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

	logger.WithFields(logrus.Fields{
		"Phase": state.EthDKG.Phase,
	}).Infof("Purging schedule")
	state.Schedule.Purge()

	state.EthDKG.ShareDistributionStart = event.BlockNumber.Uint64() + 1 // + state.EthDKG.ConfirmationLength
	state.EthDKG.ShareDistributionEnd = state.EthDKG.ShareDistributionStart + state.EthDKG.PhaseLength

	state.Schedule.Schedule(state.EthDKG.ShareDistributionStart, state.EthDKG.ShareDistributionEnd, dkgtasks.NewShareDistributionTask(state.EthDKG)) // ShareDistribution

	// schedule bad path where validators did not distribute shares
	// or distributed invalid shares
	state.EthDKG.DisputeShareDistributionStart = state.EthDKG.ShareDistributionEnd //  + state.EthDKG.ConfirmationLength
	state.EthDKG.DisputeShareDistributionEnd = state.EthDKG.DisputeShareDistributionStart + state.EthDKG.PhaseLength

	return nil
}

func ProcessKeyShareSubmission(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	logger.Info("ProcessKeyShareSubmission() ...")

	event, err := eth.Contracts().Ethdkg().ParseKeyShareSubmitted(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"Issuer":     event.Account.Hex(),
		"KeyShareG1": event.KeyShareG1,
		"KeyShareG2": event.KeyShareG2,
	}).Info("Received key shares")

	state.Lock()
	state.EthDKG.KeyShareG1s[event.Account] = event.KeyShareG1
	state.EthDKG.KeyShareG1CorrectnessProofs[event.Account] = event.KeyShareG1CorrectnessProof
	state.EthDKG.KeyShareG2s[event.Account] = event.KeyShareG2

	logger.WithFields(logrus.Fields{
		"#KeyShareG1s":                 len(state.EthDKG.KeyShareG1s[event.Account]),
		"#KeyShareG1CorrectnessProofs": len(state.EthDKG.KeyShareG1CorrectnessProofs[event.Account]),
		"#KeyShareG2s":                 len(state.EthDKG.KeyShareG2s[event.Account]),
	}).Info("Received key shares2")
	state.Unlock()

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
	state.EthDKG.Commitments[event.Account] = event.Commitments
	state.EthDKG.EncryptedShares[event.Account] = event.EncryptedShares
	state.Unlock()

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

	//state.Lock()
	//state.EthDKG.Commitments[event.Account] = event.Commitments
	//state.EthDKG.EncryptedShares[event.Account] = event.EncryptedShares
	//state.Unlock()

	return nil
}

func PopulateSchedule(state *objects.DkgState, event *bindings.ETHDKGRegistrationOpened) {

	state.RegistrationStart = event.StartBlock.Uint64()
	state.RegistrationEnd = event.StartBlock.Uint64() + event.PhaseLength.Uint64()

	// state.ShareDistributionStart = state.RegistrationEnd + event.ConfirmationLength.Uint64()
	// state.ShareDistributionEnd = state.ShareDistributionStart + event.PhaseLength.Uint64()

	// state.DisputeStart = state.ShareDistributionEnd + event.ConfirmationLength.Uint64()
	// state.DisputeEnd = state.DisputeStart + event.PhaseLength.Uint64()

	// state.KeyShareSubmissionStart = state.DisputeEnd + event.ConfirmationLength.Uint64()
	// state.KeyShareSubmissionEnd = state.KeyShareSubmissionStart + event.PhaseLength.Uint64()

	// state.MPKSubmissionStart = state.KeyShareSubmissionEnd + event.ConfirmationLength.Uint64()
	// state.MPKSubmissionEnd = state.MPKSubmissionStart + event.PhaseLength.Uint64()

	// state.GPKJSubmissionStart = state.MPKSubmissionEnd + event.ConfirmationLength.Uint64()
	// state.GPKJSubmissionEnd = state.GPKJSubmissionStart + event.PhaseLength.Uint64()

	// state.GPKJGroupAccusationStart = state.GPKJSubmissionEnd + event.ConfirmationLength.Uint64()
	// state.GPKJGroupAccusationEnd = state.GPKJGroupAccusationStart + event.PhaseLength.Uint64()

	// state.CompleteStart = state.GPKJGroupAccusationEnd + event.ConfirmationLength.Uint64()
	// state.CompleteEnd = state.CompleteStart + event.PhaseLength.Uint64()
}
