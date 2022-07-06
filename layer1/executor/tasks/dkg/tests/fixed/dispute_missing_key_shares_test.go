//go:build integration

package fixed

import (
	"context"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	"github.com/alicenet/alicenet/layer1/tests"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisputeMissingKeySharesTask_FourUnSubmittedKeyShare(t *testing.T) {
	n := 5
	fixture := setupEthereum(t, n)
	unSubmittedKeyShare := 4
	suite := StartFromShareDistributionPhase(t, fixture, []int{}, []int{}, 100)
	ctx := context.Background()
	accounts := suite.Eth.GetKnownAccounts()

	// skip DisputeShareDistribution and move to KeyShareSubmission phase
	tests.AdvanceTo(suite.Eth, suite.KeyshareSubmissionTasks[0].Start)

	// Do key share submission task
	for idx := 0; idx < n-unSubmittedKeyShare; idx++ {
		keyshareSubmissionTask := suite.KeyshareSubmissionTasks[idx]

		err := keyshareSubmissionTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "KeyShareSubmissionTask", "task-id", nil)
		assert.Nil(t, err)
		err = keyshareSubmissionTask.Prepare(ctx)
		assert.Nil(t, err)
		_, err = keyshareSubmissionTask.Execute(ctx)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			participantDkgState, err := state.GetDkgState(suite.DKGStatesDbs[j])
			assert.Nil(t, err)
			participantDkgState.OnKeyShareSubmitted(
				dkgState.Account.Address,
				dkgState.Participants[dkgState.Account.Address].KeyShareG1s,
				dkgState.Participants[dkgState.Account.Address].KeyShareG1CorrectnessProofs,
				dkgState.Participants[dkgState.Account.Address].KeyShareG2s,
			)
			err = state.SaveDkgState(suite.DKGStatesDbs[j], participantDkgState)
			assert.Nil(t, err)
		}
	}

	// advance into the end of KeyShareSubmission phase,
	// which is the start of DisputeMissingKeyShares phase
	tests.AdvanceTo(suite.Eth, suite.KeyshareSubmissionTasks[0].End)
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	// Do dispute missing key share task
	for idx := 0; idx < n; idx++ {
		disputeMissingKeyShareTask := suite.DisputeMissingKeyshareTasks[idx]
		err := disputeMissingKeyShareTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "disputeMissingKeyShareTask", "task-id", nil)
		assert.Nil(t, err)
		err = disputeMissingKeyShareTask.Prepare(ctx)
		assert.Nil(t, err)
		txn, taskErr := disputeMissingKeyShareTask.Execute(ctx)
		if idx == 0 {
			assert.Nil(t, taskErr)
			assert.NotNil(t, txn)
		} else {
			assert.NotNil(t, taskErr)
			assert.True(t, taskErr.IsRecoverable())
			assert.Nil(t, txn)
		}
	}

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := ethereum.GetContracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(unSubmittedKeyShare), badParticipants.Int64())
}

func TestDisputeMissingKeySharesTask_NoUnSubmittedKeyShare(t *testing.T) {
	n := 5
	fixture := setupEthereum(t, n)
	unSubmittedKeyShare := 0
	suite := StartFromShareDistributionPhase(t, fixture, []int{}, []int{}, 100)
	ctx := context.Background()
	accounts := suite.Eth.GetKnownAccounts()

	// skip DisputeShareDistribution and move to KeyShareSubmission phase
	tests.AdvanceTo(suite.Eth, suite.KeyshareSubmissionTasks[0].Start)

	// Do key share submission task
	for idx := 0; idx < n-unSubmittedKeyShare; idx++ {
		keyshareSubmissionTask := suite.KeyshareSubmissionTasks[idx]

		err := keyshareSubmissionTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "KeyShareSubmissionTask", "task-id", nil)
		assert.Nil(t, err)
		err = keyshareSubmissionTask.Prepare(ctx)
		assert.Nil(t, err)
		_, err = keyshareSubmissionTask.Execute(ctx)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			participantDkgState, err := state.GetDkgState(suite.DKGStatesDbs[j])
			assert.Nil(t, err)
			participantDkgState.OnKeyShareSubmitted(
				dkgState.Account.Address,
				dkgState.Participants[dkgState.Account.Address].KeyShareG1s,
				dkgState.Participants[dkgState.Account.Address].KeyShareG1CorrectnessProofs,
				dkgState.Participants[dkgState.Account.Address].KeyShareG2s,
			)
			err = state.SaveDkgState(suite.DKGStatesDbs[j], participantDkgState)
			assert.Nil(t, err)
		}
	}

	// advance into the end of KeyShareSubmission phase,
	// which is the start of DisputeMissingKeyShares phase
	tests.AdvanceTo(suite.Eth, suite.KeyshareSubmissionTasks[0].End)
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	// Do dispute missing key share task
	for idx := 0; idx < n; idx++ {
		disputeMissingKeyshareTask := suite.DisputeMissingKeyshareTasks[idx]
		err := disputeMissingKeyshareTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "disputeMissingKeyshareTask", "task-id", nil)
		assert.Nil(t, err)
		err = disputeMissingKeyshareTask.Prepare(ctx)
		assert.Nil(t, err)
		txn, taskErr := disputeMissingKeyshareTask.Execute(ctx)
		// Will return error cause phase is in DisputeShareDistribution
		assert.NotNil(t, taskErr)
		assert.Nil(t, txn)
	}

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := ethereum.GetContracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(unSubmittedKeyShare), badParticipants.Int64())
}
