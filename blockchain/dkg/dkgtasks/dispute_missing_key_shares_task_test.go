//go:build integration

package dkgtasks_test

import (
	"context"
	"testing"

	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/alicenet/alicenet/logging"
	"github.com/stretchr/testify/assert"
)

func TestDisputeMissingKeySharesTask_FourUnsubmittedKeyShare_DoWork_Success(t *testing.T) {
	n := 5
	unsubmittedKeyShares := 4
	suite := StartFromShareDistributionPhase(t, n, []int{}, []int{}, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// skip DisputeShareDistribution and move to KeyShareSubmission phase
	advanceTo(t, eth, suite.keyshareSubmissionTasks[0].Start)

	// Do key share submission task
	for idx := 0; idx < n-unsubmittedKeyShares; idx++ {
		state := dkgStates[idx]
		keyshareSubmissionTask := suite.keyshareSubmissionTasks[idx]

		dkgData := objects.NewETHDKGTaskData(state)
		err := keyshareSubmissionTask.Initialize(ctx, logger, eth, dkgData)
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
	advanceTo(t, eth, suite.keyshareSubmissionTasks[0].End)

	// Do dispute missing key share task
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		disputeMissingKeyshareTask := suite.disputeMissingKeyshareTasks[idx]
		dkgData := objects.NewETHDKGTaskData(state)
		err := disputeMissingKeyshareTask.Initialize(ctx, logger, eth, dkgData)
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
	suite := StartFromShareDistributionPhase(t, n, []int{}, []int{}, 40)
	defer suite.eth.Close()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// skip DisputeShareDistribution and move to KeyShareSubmission phase
	advanceTo(t, eth, suite.keyshareSubmissionTasks[0].Start)

	// Do key share submission task
	for idx := 0; idx < n-unsubmittedKeyShares; idx++ {
		state := dkgStates[idx]
		keyshareSubmissionTask := suite.keyshareSubmissionTasks[idx]

		dkgData := objects.NewETHDKGTaskData(state)
		err := keyshareSubmissionTask.Initialize(ctx, logger, eth, dkgData)
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
	advanceTo(t, eth, suite.keyshareSubmissionTasks[0].End)

	// Do dispute missing key share task
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		disputeMissingKeyshareTask := suite.disputeMissingKeyshareTasks[idx]

		dkgData := objects.NewETHDKGTaskData(state)
		err := disputeMissingKeyshareTask.Initialize(ctx, logger, eth, dkgData)
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
	suite := StartFromShareDistributionPhase(t, n, []int{}, []int{}, 100)
	defer suite.eth.Close()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// skip DisputeShareDistribution and move to KeyShareSubmission phase
	advanceTo(t, eth, suite.keyshareSubmissionTasks[0].Start)

	// Do key share submission task
	for idx := 0; idx < n-unsubmittedKeyShares; idx++ {
		state := dkgStates[idx]
		keyshareSubmissionTask := suite.keyshareSubmissionTasks[idx]

		dkgData := objects.NewETHDKGTaskData(state)
		err := keyshareSubmissionTask.Initialize(ctx, logger, eth, dkgData)
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
	advanceTo(t, eth, suite.keyshareSubmissionTasks[0].End)

	// Do dispute missing key share task
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		disputeMissingKeyshareTask := suite.disputeMissingKeyshareTasks[idx]

		dkgData := objects.NewETHDKGTaskData(state)
		err := disputeMissingKeyshareTask.Initialize(ctx, logger, eth, dkgData)
		assert.Nil(t, err)

		shouldRetry := disputeMissingKeyshareTask.ShouldRetry(ctx, logger, eth)
		assert.True(t, shouldRetry)
	}
}
