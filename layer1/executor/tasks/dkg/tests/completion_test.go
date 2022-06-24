//go:build integration

package dkg_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicenet/alicenet/blockchain/testutils"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	dkgTestUtils "github.com/alicenet/alicenet/layer1/executor/tasks/dkg/testutils"
	"github.com/alicenet/alicenet/layer1/monitor/events"

	"github.com/alicenet/alicenet/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// We complete everything correctly, happy path
func TestCompletion_Group_1_AllGood(t *testing.T) {
	n := 4

	err := testutils.InitializeValidatorFiles(5)
	assert.Nil(t, err)

	suite := dkgTestUtils.StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do GPKj Submission task
	for idx := 0; idx < n; idx++ {

		err := suite.GpkjSubmissionTasks[idx].Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		err = suite.GpkjSubmissionTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, suite.GpkjSubmissionTasks[idx].Success)
	}

	height, err := suite.Eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	disputeGPKjTasks := make([]*dkg.DisputeGPKjTask, n)
	completionTasks := make([]*dkg.CompletionTask, n)
	var completionStart uint64
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		disputeGPKjTask, completionTask := events.UpdateStateOnGPKJSubmissionComplete(state, height)
		disputeGPKjTasks[idx] = disputeGPKjTask
		completionTasks[idx] = completionTask
		completionStart = completionTask.GetStart()
	}

	// Advance to Completion phase
	testutils.AdvanceTo(t, eth, completionStart)

	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		err := completionTasks[idx].Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		amILeading := completionTasks[idx].AmILeading(ctx, eth, logger, state)
		err = completionTasks[idx].DoWork(ctx, logger, eth)
		if amILeading {
			assert.Nil(t, err)
			assert.True(t, completionTasks[idx].Success)
		} else {
			if completionTasks[idx].ShouldExecute(ctx, logger, eth) {
				assert.NotNil(t, err)
				assert.False(t, completionTasks[idx].Success)
			} else {
				assert.Nil(t, err)
				assert.True(t, completionTasks[idx].Success)
			}

		}
	}
}

func TestCompletion_Group_1_StartFromCompletion(t *testing.T) {
	n := 4
	suite := dkgTestUtils.StartFromCompletion(t, n, 100)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "k")

	// Do Completion task
	var hasLeader bool
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		task := suite.CompletionTasks[idx]

		err := task.Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		amILeading := task.AmILeading(ctx, eth, logger, state)

		if amILeading {
			hasLeader = true
			err = task.DoWork(ctx, logger, eth)
			eth.Commit()
			assert.Nil(t, err)
			assert.False(t, task.ShouldRetry(ctx, logger, eth))
			assert.True(t, task.Success)
		}
	}

	assert.True(t, hasLeader)
	assert.False(t, suite.CompletionTasks[0].ShouldRetry(ctx, logger, eth))

	callOpts, err := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)

	phase, err := suite.Eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint8(state.Completion), phase)

	// event
	for j := 0; j < n; j++ {
		// simulate receiving ValidatorSetCompleted event for all participants
		suite.DKGStates[j].OnCompletion()
		assert.Equal(t, suite.DKGStates[j].Phase, state.Completion)
	}
}

// We begin by submitting invalid information.
// This test is meant to raise an error resulting from an invalid argument
// for the Ethereum interface.
func TestCompletion_Group_2_Bad1(t *testing.T) {
	n := 4
	ecdsaPrivateKeys, _ := testutils.InitializePrivateKeysAndAccounts(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a task to share distribution and make sure it succeeds
	state := state.NewDkgState(acct)
	task := dkg.NewCompletionTask(state, 1, 100)
	log := logger.WithField("TaskID", "foo")

	err := task.Initialize(ctx, log, eth)
	assert.NotNil(t, err)
}

// We test to ensure that everything behaves correctly.
func TestCompletion_Group_2_Bad2(t *testing.T) {
	n := 4
	ecdsaPrivateKeys, _ := testutils.InitializePrivateKeysAndAccounts(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Do bad Completion task
	state := state.NewDkgState(acct)
	log := logger.WithField("TaskID", "foo")
	task := dkg.NewCompletionTask(state, 1, 100)

	err := task.Initialize(ctx, log, eth)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// We complete everything correctly, but we do not complete in time
func TestCompletion_Group_2_Bad3(t *testing.T) {
	n := 4
	suite := dkgTestUtils.StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do GPKj Submission task
	tasksVec := suite.GpkjSubmissionTasks
	for idx := 0; idx < n; idx++ {

		err := tasksVec[idx].Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		err = tasksVec[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, tasksVec[idx].Success)
	}

	height, err := suite.Eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	completionTasks := make([]*dkg.CompletionTask, n)
	var completionStart, completionEnd uint64
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		_, completionTask := events.UpdateStateOnGPKJSubmissionComplete(state, height)
		completionTasks[idx] = completionTask
		completionStart = completionTask.GetStart()
		completionEnd = completionTask.GetEnd()
	}

	// Advance to Completion phase
	testutils.AdvanceTo(t, eth, completionStart)

	// Advance to end of Completion phase
	testutils.AdvanceTo(t, eth, completionEnd)
	eth.Commit()

	err = completionTasks[0].Initialize(ctx, logger, eth)
	if err != nil {
		t.Fatal(err)
	}
	err = completionTasks[0].DoWork(ctx, logger, eth)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestCompletion_Group_3_ShouldRetry_returnsFalse(t *testing.T) {
	n := 4
	suite := dkgTestUtils.StartFromCompletion(t, n, 40)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do Completion task
	tasksVec := suite.CompletionTasks
	var hadLeaders bool
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		err := tasksVec[idx].Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		amILeading := tasksVec[idx].AmILeading(ctx, eth, logger, state)

		if amILeading {
			hadLeaders = true
			// only perform ETHDKG completion if validator is leading
			assert.True(t, tasksVec[idx].ShouldRetry(ctx, logger, eth))
			err = tasksVec[idx].DoWork(ctx, logger, eth)
			assert.Nil(t, err)
			assert.False(t, tasksVec[idx].ShouldRetry(ctx, logger, eth))
		}
	}

	assert.True(t, hadLeaders)

	// any task is able to tell if ETHDKG still needs completion
	// if for any reason no validator lead the process,
	// then all tasks will have ShouldExecute() returning true
	assert.False(t, tasksVec[0].ShouldRetry(ctx, logger, eth))
}

func TestCompletion_Group_3_ShouldRetry_returnsTrue(t *testing.T) {
	n := 4
	suite := dkgTestUtils.StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do GPKj Submission task
	for idx := 0; idx < n; idx++ {
		err := suite.GpkjSubmissionTasks[idx].Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		err = suite.GpkjSubmissionTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, suite.GpkjSubmissionTasks[idx].Success)
	}

	height, err := suite.Eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	disputeGPKjTasks := make([]*dkg.DisputeGPKjTask, n)
	completionTasks := make([]*dkg.CompletionTask, n)
	var completionStart uint64
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		disputeGPKjTask, completionTask := events.UpdateStateOnGPKJSubmissionComplete(state, height)
		disputeGPKjTasks[idx] = disputeGPKjTask
		completionTasks[idx] = completionTask
		completionStart = completionTask.GetStart()
	}

	// Advance to Completion phase
	testutils.AdvanceTo(t, eth, completionStart)
	eth.Commit()

	err = completionTasks[0].Initialize(ctx, logger, eth)
	assert.Nil(t, err)

	shouldRetry := completionTasks[0].ShouldExecute(ctx, logger, eth)
	assert.True(t, shouldRetry)
}
