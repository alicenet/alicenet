package dkgtasks_test

import (
	"context"
	"testing"

	"github.com/MadBase/MadNet/logging"
	"github.com/stretchr/testify/assert"
)

func TestDisputeMissingGPKjTaskFourUnsubmittedGPKj_DoWork_Success(t *testing.T) {
	n := 10
	unsubmittedGPKj := 4
	suite := StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")
	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	disputeGPKjStartBlock := currentHeight + suite.dkgStates[0].PhaseLength

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

	advanceTo(t, eth, disputeGPKjStartBlock)

	// Do dispute missing gpkj task
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		disputeMissingGPKjTask := suite.disputeMissingGPKjTasks[idx]

		err := disputeMissingGPKjTask.Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = disputeMissingGPKjTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, disputeMissingGPKjTask.Success)
	}

	callOpts := eth.GetCallOpts(ctx, accounts[0])
	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(unsubmittedGPKj), badParticipants.Int64())
}

func TestDisputeMissingGPKjTask_ShouldRetry_False(t *testing.T) {
	n := 5
	unsubmittedKeyShares := 1
	suite := StartFromMPKSubmissionPhase(t, n, 300)
	defer suite.eth.Close()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")
	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	// Do gpkj submission task
	for idx := 0; idx < n-unsubmittedKeyShares; idx++ {
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

	nextPhaseAt := currentHeight + dkgStates[0].PhaseLength
	advanceTo(t, eth, nextPhaseAt)

	// Do dispute missing gpkj task
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		disputeMissingGPKjTask := suite.disputeMissingGPKjTasks[idx]

		err := disputeMissingGPKjTask.Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = disputeMissingGPKjTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, disputeMissingGPKjTask.Success)

		shouldRetry := disputeMissingGPKjTask.ShouldRetry(ctx, logger, eth)
		assert.False(t, shouldRetry)
	}
}

func TestDisputeMissingGPKjTask_ShouldRetry_True(t *testing.T) {
	n := 5
	unsubmittedKeyShares := 1
	suite := StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.eth.Close()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")
	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	// Do gpkj submission task
	for idx := 0; idx < n-unsubmittedKeyShares; idx++ {
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

	nextPhaseAt := currentHeight + dkgStates[0].PhaseLength
	advanceTo(t, eth, nextPhaseAt)

	// Do dispute missing gpkj task
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		disputeMissingGPKjTask := suite.disputeMissingGPKjTasks[idx]

		err := disputeMissingGPKjTask.Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)

		shouldRetry := disputeMissingGPKjTask.ShouldRetry(ctx, logger, eth)
		assert.True(t, shouldRetry)
	}
}
