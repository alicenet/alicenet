package dkgtasks_test

import (
	"context"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgevents"
	"github.com/MadBase/MadNet/logging"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDisputeMissingKeySharesTaskFourUnsubmittedKeyShare_DoWork_Success(t *testing.T) {
	n := 10
	unsubmittedKeyShares := 4
	suite := StartFromShareDistributionPhase(t, n, 0, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")
	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	// Do key share submission task
	for idx := 0; idx < n-unsubmittedKeyShares; idx++ {
		state := dkgStates[idx]
		keyshareSubmissionTask := suite.keyshareSubmissionTasks[idx]

		err := keyshareSubmissionTask.Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = keyshareSubmissionTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, keyshareSubmissionTask.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			dkgevents.UpdateStateOnKeyShareSubmitted(
				dkgStates[j],
				state.Account.Address,
				state.Participants[state.Account.Address].KeyShareG1s,
				state.Participants[state.Account.Address].KeyShareG1CorrectnessProofs,
				state.Participants[state.Account.Address].KeyShareG2s,
			)
		}
	}

	nextPhaseAt := currentHeight + dkgStates[0].PhaseLength
	advanceTo(t, eth, nextPhaseAt)

	// Do dispute missing key share task
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		disputeMissingKeyshareTask := suite.disputeMissingKeyshareTasks[idx]

		err := disputeMissingKeyshareTask.Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = disputeMissingKeyshareTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, disputeMissingKeyshareTask.Success)
	}

	callOpts := eth.GetCallOpts(ctx, accounts[0])
	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(unsubmittedKeyShares), badParticipants.Int64())
}

func TestDisputeMissingKeySharesTask_ShouldRetry_False(t *testing.T) {
	n := 5
	unsubmittedKeyShares := 1
	suite := StartFromShareDistributionPhase(t, n, 0, 300)
	defer suite.eth.Close()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")
	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	// Do key share submission task
	for idx := 0; idx < n-unsubmittedKeyShares; idx++ {
		state := dkgStates[idx]
		keyshareSubmissionTask := suite.keyshareSubmissionTasks[idx]

		err := keyshareSubmissionTask.Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = keyshareSubmissionTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, keyshareSubmissionTask.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			dkgevents.UpdateStateOnKeyShareSubmitted(
				dkgStates[j],
				state.Account.Address,
				state.Participants[state.Account.Address].KeyShareG1s,
				state.Participants[state.Account.Address].KeyShareG1CorrectnessProofs,
				state.Participants[state.Account.Address].KeyShareG2s,
			)
		}
	}

	nextPhaseAt := currentHeight + dkgStates[0].PhaseLength
	advanceTo(t, eth, nextPhaseAt)

	// Do dispute missing key share task
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		disputeMissingKeyshareTask := suite.disputeMissingKeyshareTasks[idx]

		err := disputeMissingKeyshareTask.Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = disputeMissingKeyshareTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, disputeMissingKeyshareTask.Success)

		shouldRetry := disputeMissingKeyshareTask.ShouldRetry(ctx, logger, eth)
		assert.False(t, shouldRetry)
	}
}

func TestDisputeMissingKeySharesTask_ShouldRetry_True(t *testing.T) {
	n := 5
	unsubmittedKeyShares := 1
	suite := StartFromShareDistributionPhase(t, n, 0, 100)
	defer suite.eth.Close()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")
	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	// Do key share submission task
	for idx := 0; idx < n-unsubmittedKeyShares; idx++ {
		state := dkgStates[idx]
		keyshareSubmissionTask := suite.keyshareSubmissionTasks[idx]

		err := keyshareSubmissionTask.Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = keyshareSubmissionTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, keyshareSubmissionTask.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			dkgevents.UpdateStateOnKeyShareSubmitted(
				dkgStates[j],
				state.Account.Address,
				state.Participants[state.Account.Address].KeyShareG1s,
				state.Participants[state.Account.Address].KeyShareG1CorrectnessProofs,
				state.Participants[state.Account.Address].KeyShareG2s,
			)
		}
	}

	nextPhaseAt := currentHeight + dkgStates[0].PhaseLength
	advanceTo(t, eth, nextPhaseAt)

	// Do dispute missing key share task
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		disputeMissingKeyshareTask := suite.disputeMissingKeyshareTasks[idx]

		err := disputeMissingKeyshareTask.Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)

		shouldRetry := disputeMissingKeyshareTask.ShouldRetry(ctx, logger, eth)
		assert.True(t, shouldRetry)
	}
}
