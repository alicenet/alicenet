package dkgevents

import (
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

func ProcessOpenRegistration(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log,
	adminHandler interfaces.AdminHandler) error {

	logger.Info("ProcessOpenRegistration() ...")

	event, err := eth.Contracts().Ethdkg().ParseRegistrationOpen(log)
	if err != nil {
		return err
	}

	dkgState := state.EthDKG

	state.Schedule.Purge()

	state.EthDKG.PopulateSchedule(event)

	state.Schedule.Schedule(dkgState.RegistrationStart, dkgState.RegistrationEnd, dkgtasks.NewRegisterTask(dkgState))                        // Registration
	state.Schedule.Schedule(dkgState.ShareDistributionStart, dkgState.ShareDistributionEnd, dkgtasks.NewShareDistributionTask(dkgState))     // ShareDistribution
	state.Schedule.Schedule(dkgState.DisputeStart, dkgState.DisputeEnd, dkgtasks.NewDisputeTask(dkgState))                                   // DisputeShares
	state.Schedule.Schedule(dkgState.KeyShareSubmissionStart, dkgState.KeyShareSubmissionEnd, dkgtasks.NewKeyshareSubmissionTask(dkgState))  // KeyShareSubmission
	state.Schedule.Schedule(dkgState.MPKSubmissionStart, dkgState.MPKSubmissionEnd, dkgtasks.NewMPKSubmissionTask(dkgState))                 // MasterPublicKeySubmission
	state.Schedule.Schedule(dkgState.GPKJSubmissionStart, dkgState.GPKJSubmissionEnd, dkgtasks.NewGPKSubmissionTask(dkgState, adminHandler)) // GroupPublicKeySubmission
	state.Schedule.Schedule(dkgState.GPKJGroupAccusationStart, dkgState.GPKJGroupAccusationEnd, dkgtasks.NewGPKJDisputeTask(dkgState))       // DisputeGroupPublicKey
	state.Schedule.Schedule(dkgState.CompleteStart, dkgState.CompleteEnd, dkgtasks.NewCompletionTask(dkgState))                              // Complete

	state.Schedule.Status(logger)

	return nil
}

func ProcessKeyShareSubmission(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	logger.Info("ProcessKeyShareSubmission() ...")

	event, err := eth.Contracts().Ethdkg().ParseKeyShareSubmission(log)
	if err != nil {
		return err
	}

	logger.Debugf("KeyShareSubmission: %+v", event)

	state.EthDKG.KeyShareG1s[event.Issuer] = event.KeyShareG1
	state.EthDKG.KeyShareG1CorrectnessProofs[event.Issuer] = event.KeyShareG1CorrectnessProof
	state.EthDKG.KeyShareG2s[event.Issuer] = event.KeyShareG2

	return nil
}

func ProcessShareDistribution(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	logger.Info("ProcessShareDistribution() ...")

	event, err := eth.Contracts().Ethdkg().ParseShareDistribution(log)
	if err != nil {
		return err
	}

	logger.Debugf("ShareDistribution: %+v", event)

	state.EthDKG.Commitments[event.Issuer] = event.Commitments
	state.EthDKG.EncryptedShares[event.Issuer] = event.EncryptedShares

	return nil
}
