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

func ProcessOpenRegistration(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log,
	adminHandler interfaces.AdminHandler) error {

	logger.Info("ProcessOpenRegistration() ...")
	event, err := eth.Contracts().Ethdkg().ParseRegistrationOpen(log)
	if err != nil {
		return err
	}

	// If we're already finalized before registration ends, there's no point in continuing
	finalizedBig := (&big.Int{}).SetUint64(state.HighestBlockFinalized)
	if event.RegistrationEnds.Cmp(finalizedBig) < 1 {
		logger.WithField("RegistrationEnd", event.RegistrationEnds.String()).Info("Too late to participate in EthDKG")
		return nil
	}

	if state.EthDKG != nil && state.EthDKG.RegistrationEnd >= event.RegistrationEnds.Uint64() {
		logger.WithField("RegistrationEnd", event.RegistrationEnds.String()).Info("EthDKG already in progress")
		return nil
	}

	logger.WithField("RegistrationEnd", event.RegistrationEnds.String()).Info("Participating in EthDKG")

	state.Lock()
	defer state.Unlock()

	dkgState := state.EthDKG

	state.Schedule.Purge()

	PopulateSchedule(state.EthDKG, event)

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

	logger.WithFields(logrus.Fields{
		"Issuer":     event.Issuer.Hex(),
		"KeyShareG1": event.KeyShareG1,
		"KeyShareG2": event.KeyShareG2,
	}).Info("Received key shares")

	state.Lock()
	state.EthDKG.KeyShareG1s[event.Issuer] = event.KeyShareG1
	state.EthDKG.KeyShareG1CorrectnessProofs[event.Issuer] = event.KeyShareG1CorrectnessProof
	state.EthDKG.KeyShareG2s[event.Issuer] = event.KeyShareG2
	state.Unlock()

	return nil
}

func ProcessShareDistribution(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	logger.Info("ProcessShareDistribution() ...")

	event, err := eth.Contracts().Ethdkg().ParseShareDistribution(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"Issuer":          event.Issuer.Hex(),
		"Index":           event.Index,
		"EncryptedShares": event.EncryptedShares,
	}).Info("Received share distribution")

	state.Lock()
	state.EthDKG.Commitments[event.Issuer] = event.Commitments
	state.EthDKG.EncryptedShares[event.Issuer] = event.EncryptedShares
	state.Unlock()

	return nil
}
func PopulateSchedule(state *objects.DkgState, event *bindings.ETHDKGRegistrationOpen) {

	state.RegistrationStart = event.DkgStarts.Uint64()
	state.RegistrationEnd = event.RegistrationEnds.Uint64()

	state.ShareDistributionStart = state.RegistrationEnd + 1
	state.ShareDistributionEnd = event.ShareDistributionEnds.Uint64()

	state.DisputeStart = state.ShareDistributionEnd + 1
	state.DisputeEnd = event.DisputeEnds.Uint64()

	state.KeyShareSubmissionStart = state.DisputeEnd + 1
	state.KeyShareSubmissionEnd = event.KeyShareSubmissionEnds.Uint64()

	state.MPKSubmissionStart = state.KeyShareSubmissionEnd + 1
	state.MPKSubmissionEnd = event.MpkSubmissionEnds.Uint64()

	state.GPKJSubmissionStart = state.MPKSubmissionEnd + 1
	state.GPKJSubmissionEnd = event.GpkjSubmissionEnds.Uint64()

	state.GPKJGroupAccusationStart = state.GPKJSubmissionEnd + 1
	state.GPKJGroupAccusationEnd = event.GpkjDisputeEnds.Uint64()

	state.CompleteStart = state.GPKJGroupAccusationEnd + 1
	state.CompleteEnd = event.DkgComplete.Uint64()
}
