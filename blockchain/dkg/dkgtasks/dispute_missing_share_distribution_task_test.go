package dkgtasks_test

import (
	"context"
	"testing"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/logging"
	"github.com/stretchr/testify/assert"
)

func TestShouldAccuseOneValidatorWhoDidNotDistributeShares(t *testing.T) {
	n := 5
	suite := StartFromShareDistributionPhase(t, n, 1, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	// currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	// assert.Nil(t, err)
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	tasks := make([]*dkgtasks.DisputeMissingShareDistributionTask, len(suite.dkgStates))

	for idx := range accounts {
		state := suite.dkgStates[idx]
		phaseStart := suite.dkgStates[idx].PhaseStart + suite.dkgStates[idx].PhaseLength
		phaseEnd := phaseStart + suite.dkgStates[idx].PhaseLength
		tasks[idx] = dkgtasks.NewDisputeMissingShareDistributionTask(state, phaseStart, phaseEnd)
		err := tasks[idx].Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, tasks[idx].Success)
	}

	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(suite.eth.GetCallOpts(ctx, accounts[0]))
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), badParticipants.Uint64())
}

func TestShouldAccuseAllValidatorsWhoDidNotDistributeShares(t *testing.T) {
	n := 5
	suite := StartFromShareDistributionPhase(t, n, n, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	// currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	// assert.Nil(t, err)
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	tasks := make([]*dkgtasks.DisputeMissingShareDistributionTask, len(suite.dkgStates))

	for idx := range accounts {
		state := suite.dkgStates[idx]
		phaseStart := suite.dkgStates[idx].PhaseStart + suite.dkgStates[idx].PhaseLength
		phaseEnd := phaseStart + suite.dkgStates[idx].PhaseLength
		tasks[idx] = dkgtasks.NewDisputeMissingShareDistributionTask(state, phaseStart, phaseEnd)
		err := tasks[idx].Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, tasks[idx].Success)
	}

	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(suite.eth.GetCallOpts(ctx, accounts[0]))
	assert.Nil(t, err)
	assert.Equal(t, uint64(n), badParticipants.Uint64())
}

func TestShouldNotAccuseValidatorsWhoDidDistributeShares(t *testing.T) {
	n := 5
	suite := StartFromShareDistributionPhase(t, n, 0, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	// currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	// assert.Nil(t, err)
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	tasks := make([]*dkgtasks.DisputeMissingShareDistributionTask, len(suite.dkgStates))

	for idx := range accounts {
		state := suite.dkgStates[idx]
		phaseStart := suite.dkgStates[idx].PhaseStart + suite.dkgStates[idx].PhaseLength
		phaseEnd := phaseStart + suite.dkgStates[idx].PhaseLength
		tasks[idx] = dkgtasks.NewDisputeMissingShareDistributionTask(state, phaseStart, phaseEnd)
		err := tasks[idx].Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)

		if idx == n-1 {
			// injecting bad state into this validator
			var emptySharesHash [32]byte
			state.Participants[accounts[0].Address].DistributedSharesHash = emptySharesHash
		}

		err = tasks[idx].DoWork(ctx, logger, suite.eth)
		if idx == n-1 {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}

		suite.eth.Commit()
		if idx == n-1 {
			assert.False(t, tasks[idx].Success)
		} else {
			assert.True(t, tasks[idx].Success)
		}
	}

	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(suite.eth.GetCallOpts(ctx, accounts[0]))
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), badParticipants.Uint64())
}
