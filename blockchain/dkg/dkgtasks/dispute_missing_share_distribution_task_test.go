package dkgtasks_test

import (
	"context"
	"testing"

	"github.com/MadBase/MadNet/logging"
	"github.com/stretchr/testify/assert"
)

func TestShouldAccuseOneValidatorWhoDidNotDistributeShares(t *testing.T) {
	n := 5
	suite := StartFromShareDistributionPhase(t, n, []int{4}, []int{}, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	for idx := range accounts {
		state := suite.dkgStates[idx]
		task := suite.disputeMissingShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = task.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, task.Success)
	}

	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(suite.eth.GetCallOpts(ctx, accounts[0]))
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), badParticipants.Uint64())
}

func TestShouldAccuseAllValidatorsWhoDidNotDistributeShares(t *testing.T) {
	n := 5
	suite := StartFromShareDistributionPhase(t, n, []int{0, 1, 2, 3, 4}, []int{}, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	for idx := range accounts {
		state := suite.dkgStates[idx]
		task := suite.disputeMissingShareDistTasks[idx]
		err := task.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = task.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, task.Success)
	}

	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(suite.eth.GetCallOpts(ctx, accounts[0]))
	assert.Nil(t, err)
	assert.Equal(t, uint64(n), badParticipants.Uint64())
}

func TestShouldNotAccuseValidatorsWhoDidDistributeShares(t *testing.T) {
	n := 5
	suite := StartFromShareDistributionPhase(t, n, []int{}, []int{}, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	for idx := range accounts {
		state := suite.dkgStates[idx]
		task := suite.disputeMissingShareDistTasks[idx]
		err := task.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)

		if idx == n-1 {
			// injecting bad state into this validator
			var emptySharesHash [32]byte
			state.Participants[accounts[0].Address].DistributedSharesHash = emptySharesHash
		}

		err = task.DoWork(ctx, logger, suite.eth)
		if idx == n-1 {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}

		suite.eth.Commit()
		if idx == n-1 {
			assert.False(t, task.Success)
		} else {
			assert.True(t, task.Success)
		}
	}

	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(suite.eth.GetCallOpts(ctx, accounts[0]))
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), badParticipants.Uint64())
}

func TestDisputeMissingShareDistributionTask_ShouldRetryTrue(t *testing.T) {
	n := 5
	suite := StartFromShareDistributionPhase(t, n, []int{0}, []int{}, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	for idx := range accounts {
		state := suite.dkgStates[idx]
		task := suite.disputeMissingShareDistTasks[idx]
		err := task.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		shouldRetry := task.ShouldRetry(ctx, logger, suite.eth)
		assert.True(t, shouldRetry)
	}
}

func TestDisputeMissingShareDistributionTask_ShouldRetryFalse(t *testing.T) {
	n := 5
	suite := StartFromShareDistributionPhase(t, n, []int{}, []int{}, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	for idx := range accounts {
		state := suite.dkgStates[idx]
		task := suite.disputeMissingShareDistTasks[idx]
		err := task.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = task.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, task.Success)
	}

	for idx := 0; idx < len(suite.dkgStates); idx++ {
		task := suite.disputeMissingShareDistTasks[idx]
		shouldRetry := task.ShouldRetry(ctx, logger, suite.eth)
		assert.False(t, shouldRetry)
	}
}

func TestShouldAccuseOneValidatorWhoDidNotDistributeSharesAndAnotherSubmittedBadShares(t *testing.T) {
	n := 5
	suite := StartFromShareDistributionPhase(t, n, []int{4}, []int{3}, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	// Do Share Dispute task
	for idx := range accounts {
		state := suite.dkgStates[idx]

		// disputeMissingShareDist
		disputeMissingShareDistTask := suite.disputeMissingShareDistTasks[idx]

		err := disputeMissingShareDistTask.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = disputeMissingShareDistTask.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, disputeMissingShareDistTask.Success)

		// disputeShareDist
		disputeShareDistTask := suite.disputeShareDistTasks[idx]

		err = disputeShareDistTask.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = disputeShareDistTask.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, disputeShareDistTask.Success)
	}

	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(suite.eth.GetCallOpts(ctx, accounts[0]))
	assert.Nil(t, err)
	assert.Equal(t, uint64(2), badParticipants.Uint64())

	callOpts := suite.eth.GetCallOpts(ctx, accounts[0])

	//assert bad participants are not validators anymore, i.e, they were fined and evicted
	isValidator, err := suite.eth.Contracts().ValidatorPool().IsValidator(callOpts, suite.dkgStates[3].Account.Address)
	assert.Nil(t, err)
	assert.False(t, isValidator)

	isValidator, err = suite.eth.Contracts().ValidatorPool().IsValidator(callOpts, suite.dkgStates[4].Account.Address)
	assert.Nil(t, err)
	assert.False(t, isValidator)
}
