//go:build integration

package dkg_test

import (
	"context"
	"testing"

	"github.com/alicenet/alicenet/blockchain/testutils"
	dkgTestUtils "github.com/alicenet/alicenet/layer1/executor/tasks/dkg/testutils"
	"github.com/alicenet/alicenet/logging"
	"github.com/stretchr/testify/assert"
)

func TestDisputeMissingKeySharesTask_FourUnsubmittedKeyShare_DoWork_Success(t *testing.T) {
	n := 5
	unsubmittedKeyShares := 4
	suite := dkgTestUtils.StartFromShareDistributionPhase(t, n, []int{}, []int{}, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// skip DisputeShareDistribution and move to KeyShareSubmission phase
	testutils.AdvanceTo(t, eth, suite.KeyshareSubmissionTasks[0].Start)

	// Do key share submission task
	for idx := 0; idx < n-unsubmittedKeyShares; idx++ {
		state := dkgStates[idx]
		keyshareSubmissionTask := suite.KeyshareSubmissionTasks[idx]

		err := keyshareSubmissionTask.Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		err = keyshareSubmissionTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, keyshareSubmissionTask.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			dkgStates[j].OnKeyShareSubmitted(
				state.Account.Address,
				state.Participants[state.Account.Address].KeyShareG1s,
				state.Participants[state.Account.Address].KeyShareG1CorrectnessProofs,
				state.Participants[state.Account.Address].KeyShareG2s,
			)
		}
	}

	// advance into the end of KeyShareSubmission phase,
	// which is the start of DisputeMissingKeyShares phase
	testutils.AdvanceTo(t, eth, suite.KeyshareSubmissionTasks[0].End)

	// Do dispute missing key share task
	for idx := 0; idx < n; idx++ {
		disputeMissingKeyshareTask := suite.DisputeMissingKeyshareTasks[idx]

		err := disputeMissingKeyshareTask.Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		err = disputeMissingKeyshareTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, disputeMissingKeyshareTask.Success)
	}

	callOpts, err := eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(unsubmittedKeyShares), badParticipants.Int64())
}

func TestDisputeMissingKeySharesTask_ShouldRetry_False(t *testing.T) {
	n := 5
	unsubmittedKeyShares := 1
	suite := dkgTestUtils.StartFromShareDistributionPhase(t, n, []int{}, []int{}, 40)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// skip DisputeShareDistribution and move to KeyShareSubmission phase
	testutils.AdvanceTo(t, eth, suite.KeyshareSubmissionTasks[0].Start)

	// Do key share submission task
	for idx := 0; idx < n-unsubmittedKeyShares; idx++ {
		state := dkgStates[idx]
		keyshareSubmissionTask := suite.KeyshareSubmissionTasks[idx]

		err := keyshareSubmissionTask.Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		err = keyshareSubmissionTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, keyshareSubmissionTask.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			dkgStates[j].OnKeyShareSubmitted(
				state.Account.Address,
				state.Participants[state.Account.Address].KeyShareG1s,
				state.Participants[state.Account.Address].KeyShareG1CorrectnessProofs,
				state.Participants[state.Account.Address].KeyShareG2s,
			)
		}
	}

	// advance into the end of KeyShareSubmission phase,
	// which is the start of DisputeMissingKeyShares phase
	testutils.AdvanceTo(t, eth, suite.KeyshareSubmissionTasks[0].End)

	// Do dispute missing key share task
	for idx := 0; idx < n; idx++ {
		disputeMissingKeyshareTask := suite.DisputeMissingKeyshareTasks[idx]

		err := disputeMissingKeyshareTask.Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		err = disputeMissingKeyshareTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		assert.True(t, disputeMissingKeyshareTask.Success)

		eth.Commit()
		shouldRetry := disputeMissingKeyshareTask.ShouldRetry(ctx, logger, eth)
		assert.False(t, shouldRetry)
	}
}

func TestDisputeMissingKeySharesTask_ShouldRetry_True(t *testing.T) {
	n := 5
	unsubmittedKeyShares := 1
	suite := dkgTestUtils.StartFromShareDistributionPhase(t, n, []int{}, []int{}, 100)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// skip DisputeShareDistribution and move to KeyShareSubmission phase
	testutils.AdvanceTo(t, eth, suite.KeyshareSubmissionTasks[0].Start)

	// Do key share submission task
	for idx := 0; idx < n-unsubmittedKeyShares; idx++ {
		state := dkgStates[idx]
		keyshareSubmissionTask := suite.KeyshareSubmissionTasks[idx]

		err := keyshareSubmissionTask.Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		err = keyshareSubmissionTask.DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, keyshareSubmissionTask.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			dkgStates[j].OnKeyShareSubmitted(
				state.Account.Address,
				state.Participants[state.Account.Address].KeyShareG1s,
				state.Participants[state.Account.Address].KeyShareG1CorrectnessProofs,
				state.Participants[state.Account.Address].KeyShareG2s,
			)
		}
	}

	// advance into the end of KeyShareSubmission phase,
	// which is the start of DisputeMissingKeyShares phase
	testutils.AdvanceTo(t, eth, suite.KeyshareSubmissionTasks[0].End)

	// Do dispute missing key share task
	for idx := 0; idx < n; idx++ {
		disputeMissingKeyshareTask := suite.DisputeMissingKeyshareTasks[idx]

		err := disputeMissingKeyshareTask.Initialize(ctx, logger, eth)
		assert.Nil(t, err)

		shouldRetry := disputeMissingKeyshareTask.ShouldRetry(ctx, logger, eth)
		assert.True(t, shouldRetry)
	}
}
