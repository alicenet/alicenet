//go:build integration

package dkg_test

import (
	"context"
	"testing"

	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/testutils"

	"github.com/alicenet/alicenet/logging"
	"github.com/stretchr/testify/assert"
)

func TestDisputeMissingShareDistributionTask_Group_1_ShouldAccuseOneValidatorWhoDidNotDistributeShares(t *testing.T) {
	n := 5
	suite := testutils.StartFromShareDistributionPhase(t, n, []int{4}, []int{}, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	for idx := range accounts {
		task := suite.DisputeMissingShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		err = task.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, task.Success)
	}

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := suite.Eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), badParticipants.Uint64())
}

func TestDisputeMissingShareDistributionTask_Group_1_ShouldAccuseAllValidatorsWhoDidNotDistributeShares(t *testing.T) {
	n := 5
	suite := testutils.StartFromShareDistributionPhase(t, n, []int{0, 1, 2, 3, 4}, []int{}, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	for idx := range accounts {
		task := suite.DisputeMissingShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		err = task.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, task.Success)
	}

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := suite.Eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint64(n), badParticipants.Uint64())
}

func TestDisputeMissingShareDistributionTask_Group_1_ShouldNotAccuseValidatorsWhoDidDistributeShares(t *testing.T) {
	n := 5
	suite := testutils.StartFromShareDistributionPhase(t, n, []int{}, []int{}, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	for idx := range accounts {
		state := suite.DKGStates[idx]
		task := suite.DisputeMissingShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		if idx == n-1 {
			// injecting bad state into this validator
			var emptySharesHash [32]byte
			state.Participants[accounts[0].Address].DistributedSharesHash = emptySharesHash
		}

		err = task.DoWork(ctx, logger, suite.Eth)
		if idx == n-1 {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}

		suite.Eth.Commit()
		if idx == n-1 {
			assert.False(t, task.Success)
		} else {
			assert.True(t, task.Success)
		}
	}

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := suite.Eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), badParticipants.Uint64())
}

func TestDisputeMissingShareDistributionTask_Group_2_DisputeMissingShareDistributionTask_ShouldRetryTrue(t *testing.T) {
	n := 5
	suite := testutils.StartFromShareDistributionPhase(t, n, []int{0}, []int{}, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	for idx := range accounts {
		task := suite.DisputeMissingShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		shouldRetry := task.ShouldRetry(ctx, logger, suite.Eth)
		assert.True(t, shouldRetry)
	}
}

func TestDisputeMissingShareDistributionTask_Group_2_DisputeMissingShareDistributionTask_ShouldRetryFalse(t *testing.T) {
	n := 5
	suite := testutils.StartFromShareDistributionPhase(t, n, []int{}, []int{}, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	for idx := range accounts {
		task := suite.DisputeMissingShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		err = task.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, task.Success)
	}

	for idx := 0; idx < len(suite.DKGStates); idx++ {
		task := suite.DisputeMissingShareDistTasks[idx]
		shouldRetry := task.ShouldRetry(ctx, logger, suite.Eth)
		assert.False(t, shouldRetry)
	}
}

func TestDisputeMissingShareDistributionTask_Group_2_ShouldAccuseOneValidatorWhoDidNotDistributeSharesAndAnotherSubmittedBadShares(t *testing.T) {
	n := 5
	suite := testutils.StartFromShareDistributionPhase(t, n, []int{4}, []int{3}, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	// Do Share Dispute task
	for idx := range accounts {
		// disputeMissingShareDist
		disputeMissingShareDistTask := suite.DisputeMissingShareDistTasks[idx]

		err := disputeMissingShareDistTask.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		err = disputeMissingShareDistTask.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, disputeMissingShareDistTask.Success)

		// disputeShareDist
		disputeShareDistTask := suite.DisputeShareDistTasks[idx]

		err = disputeShareDistTask.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		err = disputeShareDistTask.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, disputeShareDistTask.Success)
	}

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := suite.Eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint64(2), badParticipants.Uint64())

	//assert bad participants are not validators anymore, i.e, they were fined and evicted
	isValidator, err := suite.Eth.Contracts().ValidatorPool().IsValidator(callOpts, suite.DKGStates[3].Account.Address)
	assert.Nil(t, err)
	assert.False(t, isValidator)

	isValidator, err = suite.Eth.Contracts().ValidatorPool().IsValidator(callOpts, suite.DKGStates[4].Account.Address)
	assert.Nil(t, err)
	assert.False(t, isValidator)
}
