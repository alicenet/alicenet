package dkgtasks_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgevents"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/MadBase/MadNet/logging"
	"github.com/stretchr/testify/assert"
)

// We test to ensure that everything behaves correctly.
func TestGPKjDisputeNoBadGPKj(t *testing.T) {
	n := 5
	unsubmittedGPKj := 0
	suite := StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")
	// currentHeight, err := eth.GetCurrentHeight(ctx)
	// assert.Nil(t, err)
	// disputeGPKjStartBlock := currentHeight + suite.dkgStates[0].PhaseLength

	// Do gpkj submission task
	for idx := 0; idx < n-unsubmittedGPKj; idx++ {
		state := dkgStates[idx]
		gpkjSubmissionTask := suite.gpkjSubmissionTasks[idx]

		err := gpkjSubmissionTask.Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)

		err = gpkjSubmissionTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, gpkjSubmissionTask.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			dkgStates[j].OnGPKjSubmitted(state.Account.Address, state.Participants[state.Account.Address].GPKj)
		}
	}

	callOpts := suite.eth.GetCallOpts(ctx, accounts[0])
	phase, err := suite.eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint8(objects.DisputeGPKJSubmission), phase)

	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	disputePhaseAt := currentHeight + suite.dkgStates[0].ConfirmationLength

	advanceTo(t, eth, disputePhaseAt)

	// Do dispute bad gpkj task
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		disputeBadGPKjTask, _, _, _, _, _ := dkgevents.UpdateStateOnGPKJSubmissionComplete(state, logger, disputePhaseAt)

		err := disputeBadGPKjTask.Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = disputeBadGPKjTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, disputeBadGPKjTask.Success)
	}

	//callOpts = eth.GetCallOpts(ctx, accounts[0])
	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), badParticipants.Int64())
}

// Here, we have a malicious gpkj submission.
func TestGPKjDispute1Invalid(t *testing.T) {
	n := 5
	unsubmittedGPKj := 0
	suite := StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")
	// currentHeight, err := eth.GetCurrentHeight(ctx)
	// assert.Nil(t, err)
	// disputeGPKjStartBlock := currentHeight + suite.dkgStates[0].PhaseLength

	// Do gpkj submission task
	for idx := 0; idx < n-unsubmittedGPKj; idx++ {
		state := dkgStates[idx]
		gpkjSubmissionTask := suite.gpkjSubmissionTasks[idx]

		err := gpkjSubmissionTask.Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)

		// inject bad GPKj data
		if state.Index == 1 {
			// mess up with group private key (gskj)
			gskjBad := new(big.Int).Add(state.GroupPrivateKey, big.NewInt(1))
			// here's the group public key
			gpkj := new(cloudflare.G2).ScalarBaseMult(gskjBad)
			gpkjBad, err := bn256.G2ToBigIntArray(gpkj)
			assert.Nil(t, err)

			state.GroupPrivateKey = gskjBad
			state.Participants[state.Account.Address].GPKj = gpkjBad
		}

		err = gpkjSubmissionTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, gpkjSubmissionTask.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			dkgStates[j].OnGPKjSubmitted(state.Account.Address, state.Participants[state.Account.Address].GPKj)
		}
	}

	callOpts := suite.eth.GetCallOpts(ctx, accounts[0])
	phase, err := suite.eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint8(objects.DisputeGPKJSubmission), phase)

	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	disputePhaseAt := currentHeight + suite.dkgStates[0].ConfirmationLength

	advanceTo(t, eth, disputePhaseAt)

	// Do dispute bad gpkj task
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		disputeBadGPKjTask, _, _, _, _, _ := dkgevents.UpdateStateOnGPKJSubmissionComplete(state, logger, disputePhaseAt)

		err := disputeBadGPKjTask.Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = disputeBadGPKjTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, disputeBadGPKjTask.Success)
	}

	//callOpts = eth.GetCallOpts(ctx, accounts[0])
	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), badParticipants.Int64())
}

// We test to ensure that everything behaves correctly.
// Here, we have a malicious accusation.
func TestGPKjDisputeGoodMaliciousAccusation(t *testing.T) {
	n := 5
	unsubmittedGPKj := 0
	suite := StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")
	// currentHeight, err := eth.GetCurrentHeight(ctx)
	// assert.Nil(t, err)
	// disputeGPKjStartBlock := currentHeight + suite.dkgStates[0].PhaseLength

	// Do gpkj submission task
	for idx := 0; idx < n-unsubmittedGPKj; idx++ {
		state := dkgStates[idx]
		gpkjSubmissionTask := suite.gpkjSubmissionTasks[idx]

		err := gpkjSubmissionTask.Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)

		err = gpkjSubmissionTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, gpkjSubmissionTask.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			dkgStates[j].OnGPKjSubmitted(state.Account.Address, state.Participants[state.Account.Address].GPKj)
		}
	}

	callOpts := suite.eth.GetCallOpts(ctx, accounts[0])
	phase, err := suite.eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint8(objects.DisputeGPKJSubmission), phase)

	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	disputePhaseAt := currentHeight + suite.dkgStates[0].ConfirmationLength

	advanceTo(t, eth, disputePhaseAt)

	// Do dispute bad gpkj task
	badAccuserIdx := 0
	accusedIdx := 1
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		disputeBadGPKjTask, _, _, _, _, _ := dkgevents.UpdateStateOnGPKJSubmissionComplete(state, logger, disputePhaseAt)

		err := disputeBadGPKjTask.Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		if idx == badAccuserIdx {
			state.DishonestValidators = objects.ParticipantList{state.GetSortedParticipants()[accusedIdx].Copy()}
		}
		err = disputeBadGPKjTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, disputeBadGPKjTask.Success)
	}

	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), badParticipants.Int64())

	nValidators, err := suite.eth.Contracts().ValidatorPool().GetValidatorsCount(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint64(4), nValidators.Uint64())

	isValidator, err := eth.Contracts().ValidatorPool().IsValidator(callOpts, dkgStates[badAccuserIdx].Account.Address)
	assert.Nil(t, err)
	assert.Equal(t, false, isValidator)
}
