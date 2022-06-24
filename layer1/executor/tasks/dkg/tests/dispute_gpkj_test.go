//go:build integration

package dkg_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/alicenet/alicenet/blockchain/testutils"
	dkgState "github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	dkgTestUtils "github.com/alicenet/alicenet/layer1/executor/tasks/dkg/testutils"
	"github.com/alicenet/alicenet/layer1/monitor/events"

	"github.com/alicenet/alicenet/crypto/bn256"
	"github.com/alicenet/alicenet/crypto/bn256/cloudflare"
	"github.com/alicenet/alicenet/logging"
	"github.com/stretchr/testify/assert"
)

// We test to ensure that everything behaves correctly.
func TestGPKjDispute_NoBadGPKj(t *testing.T) {
	n := 5
	unsubmittedGPKj := 0
	suite := dkgTestUtils.StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do gpkj submission task
	for idx := 0; idx < n-unsubmittedGPKj; idx++ {
		state := dkgStates[idx]
		gpkjSubmissionTask := suite.GpkjSubmissionTasks[idx]

		err := gpkjSubmissionTask.Initialize(ctx, logger, eth)
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

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	phase, err := suite.Eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint8(dkgState.DisputeGPKJSubmission), phase)

	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	disputePhaseAt := currentHeight + suite.DKGStates[0].ConfirmationLength

	testutils.AdvanceTo(t, eth, disputePhaseAt)

	// Do dispute bad gpkj task
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		disputeBadGPKjTask, _ := events.UpdateStateOnGPKJSubmissionComplete(state, disputePhaseAt)

		err := disputeBadGPKjTask.Initialize(ctx, logger, eth)
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
func TestGPKjDispute_1Invalid(t *testing.T) {
	n := 5
	unsubmittedGPKj := 0
	suite := dkgTestUtils.StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do gpkj submission task
	for idx := 0; idx < n-unsubmittedGPKj; idx++ {
		state := dkgStates[idx]
		gpkjSubmissionTask := suite.GpkjSubmissionTasks[idx]

		err := gpkjSubmissionTask.Initialize(ctx, logger, eth)
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

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	phase, err := suite.Eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint8(dkgState.DisputeGPKJSubmission), phase)

	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	disputePhaseAt := currentHeight + suite.DKGStates[0].ConfirmationLength

	testutils.AdvanceTo(t, eth, disputePhaseAt)

	// Do dispute bad gpkj task
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		disputeBadGPKjTask, _ := events.UpdateStateOnGPKJSubmissionComplete(state, disputePhaseAt)

		err := disputeBadGPKjTask.Initialize(ctx, logger, eth)
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
func TestGPKjDispute_GoodMaliciousAccusation(t *testing.T) {
	n := 5
	unsubmittedGPKj := 0
	suite := dkgTestUtils.StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do gpkj submission task
	for idx := 0; idx < n-unsubmittedGPKj; idx++ {
		state := dkgStates[idx]
		gpkjSubmissionTask := suite.GpkjSubmissionTasks[idx]

		err := gpkjSubmissionTask.Initialize(ctx, logger, eth)
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

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	phase, err := suite.Eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint8(dkgState.DisputeGPKJSubmission), phase)

	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	disputePhaseAt := currentHeight + suite.DKGStates[0].ConfirmationLength

	testutils.AdvanceTo(t, eth, disputePhaseAt)

	// Do dispute bad gpkj task
	badAccuserIdx := 0
	accusedIdx := 1
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		disputeBadGPKjTask, _ := events.UpdateStateOnGPKJSubmissionComplete(state, disputePhaseAt)

		err := disputeBadGPKjTask.Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		if idx == badAccuserIdx {
			state.DishonestValidators = dkgState.ParticipantList{state.GetSortedParticipants()[accusedIdx].Copy()}
		}
		err = disputeBadGPKjTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, disputeBadGPKjTask.Success)
	}

	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), badParticipants.Int64())

	nValidators, err := suite.Eth.Contracts().ValidatorPool().GetValidatorsCount(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint64(4), nValidators.Uint64())

	isValidator, err := eth.Contracts().ValidatorPool().IsValidator(callOpts, dkgStates[badAccuserIdx].Account.Address)
	assert.Nil(t, err)
	assert.Equal(t, false, isValidator)
}
