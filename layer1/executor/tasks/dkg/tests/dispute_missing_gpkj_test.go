//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/alicenet/alicenet/blockchain/testutils"
	dkgTestUtils "github.com/alicenet/alicenet/layer1/executor/tasks/dkg/testutils"

	"github.com/alicenet/alicenet/logging"
	"github.com/stretchr/testify/assert"
)

func TestDisputeMissingGPKjTask_Group_1_FourUnsubmittedGPKj_DoWork_Success(t *testing.T) {
	n := 10
	unsubmittedGPKj := 4
	suite := dkgTestUtils.StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")
	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	disputeGPKjStartBlock := currentHeight + suite.DKGStates[0].PhaseLength

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

	testutils.AdvanceTo(t, eth, disputeGPKjStartBlock)

	// Do dispute missing gpkj task
	for idx := 0; idx < n; idx++ {
		disputeMissingGPKjTask := suite.DisputeMissingGPKjTasks[idx]

		err := disputeMissingGPKjTask.Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		err = disputeMissingGPKjTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, disputeMissingGPKjTask.Success)
	}

	callOpts, err := eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(unsubmittedGPKj), badParticipants.Int64())
}

func TestDisputeMissingGPKjTask_Group_1_ShouldRetry_False(t *testing.T) {
	n := 5
	unsubmittedKeyShares := 1
	suite := dkgTestUtils.StartFromMPKSubmissionPhase(t, n, 300)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")
	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	// Do gpkj submission task
	for idx := 0; idx < n-unsubmittedKeyShares; idx++ {
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

	nextPhaseAt := currentHeight + dkgStates[0].PhaseLength
	testutils.AdvanceTo(t, eth, nextPhaseAt)

	// Do dispute missing gpkj task
	for idx := 0; idx < n; idx++ {
		disputeMissingGPKjTask := suite.DisputeMissingGPKjTasks[idx]

		err := disputeMissingGPKjTask.Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		err = disputeMissingGPKjTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, disputeMissingGPKjTask.Success)

		shouldRetry := disputeMissingGPKjTask.ShouldRetry(ctx, logger, eth)
		assert.False(t, shouldRetry)
	}
}

func TestDisputeMissingGPKjTask_Group_1_ShouldRetry_True(t *testing.T) {
	n := 5
	unsubmittedKeyShares := 1
	suite := dkgTestUtils.StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")
	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	// Do gpkj submission task
	for idx := 0; idx < n-unsubmittedKeyShares; idx++ {
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

	nextPhaseAt := currentHeight + dkgStates[0].PhaseLength
	testutils.AdvanceTo(t, eth, nextPhaseAt)

	// Do dispute missing gpkj task
	for idx := 0; idx < n; idx++ {
		disputeMissingGPKjTask := suite.DisputeMissingGPKjTasks[idx]

		err := disputeMissingGPKjTask.Initialize(ctx, logger, eth)
		assert.Nil(t, err)

		shouldRetry := disputeMissingGPKjTask.ShouldRetry(ctx, logger, eth)
		assert.True(t, shouldRetry)
	}
}

func TestDisputeMissingGPKjTask_Group_2_ShouldAccuseOneValidatorWhoDidNotDistributeGPKjAndAnotherSubmittedBadGPKj(t *testing.T) {
	n := 5
	suite := dkgTestUtils.StartFromGPKjPhase(t, n, []int{4}, []int{3}, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	// Do GPKj Dispute task
	for idx := range accounts {
		// disputeMissingGPKj
		disputeMissingGPKjTask := suite.DisputeMissingGPKjTasks[idx]

		err := disputeMissingGPKjTask.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		err = disputeMissingGPKjTask.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, disputeMissingGPKjTask.Success)

		// disputeGPKj
		disputeGPKjTask := suite.DisputeGPKjTasks[idx]

		err = disputeGPKjTask.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		err = disputeGPKjTask.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, disputeGPKjTask.Success)
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

func TestDisputeMissingGPKjTask_Group_2_ShouldAccuseTwoValidatorWhoDidNotDistributeGPKjAndAnotherTwoSubmittedBadGPKj(t *testing.T) {
	n := 5
	suite := dkgTestUtils.StartFromGPKjPhase(t, n, []int{3, 4}, []int{1, 2}, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Test", "Test1")

	// Do GPKj Dispute task
	for idx := range accounts {
		// disputeMissingGPKj
		disputeMissingGPKjTask := suite.DisputeMissingGPKjTasks[idx]

		err := disputeMissingGPKjTask.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		err = disputeMissingGPKjTask.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, disputeMissingGPKjTask.Success)

		// disputeGPKj
		disputeGPKjTask := suite.DisputeGPKjTasks[idx]

		err = disputeGPKjTask.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		err = disputeGPKjTask.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, disputeGPKjTask.Success)
	}

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)

	badParticipants, err := suite.Eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint64(4), badParticipants.Uint64())

	//assert bad participants are not validators anymore, i.e, they were fined and evicted
	isValidator, err := suite.Eth.Contracts().ValidatorPool().IsValidator(callOpts, suite.DKGStates[1].Account.Address)
	assert.Nil(t, err)
	assert.False(t, isValidator)

	isValidator, err = suite.Eth.Contracts().ValidatorPool().IsValidator(callOpts, suite.DKGStates[2].Account.Address)
	assert.Nil(t, err)
	assert.False(t, isValidator)

	isValidator, err = suite.Eth.Contracts().ValidatorPool().IsValidator(callOpts, suite.DKGStates[3].Account.Address)
	assert.Nil(t, err)
	assert.False(t, isValidator)

	isValidator, err = suite.Eth.Contracts().ValidatorPool().IsValidator(callOpts, suite.DKGStates[4].Account.Address)
	assert.Nil(t, err)
	assert.False(t, isValidator)
}
