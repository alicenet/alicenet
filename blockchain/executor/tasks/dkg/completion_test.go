//go:build integration

package dkg

import (
	"context"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/tasks/dkg/dkgevents"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/dtest"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/objects"

	"github.com/MadBase/MadNet/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// We complete everything correctly, happy path
func TestCompletion_Group_1_AllGood(t *testing.T) {
	n := 4

	err := dtest.InitializeValidatorFiles(5)
	assert.Nil(t, err)

	suite := StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.eth.Close()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do GPKj Submission task
	for idx := 0; idx < n; idx++ {

		err := suite.gpkjSubmissionTasks[idx].Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		err = suite.gpkjSubmissionTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, suite.gpkjSubmissionTasks[idx].Success)
	}

	height, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	disputeGPKjTasks := make([]*DisputeGPKjTask, n)
	completionTasks := make([]*CompletionTask, n)
	var completionStart uint64
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		disputeGPKjTask, completionTask := dkgevents.UpdateStateOnGPKJSubmissionComplete(state, height)
		disputeGPKjTasks[idx] = disputeGPKjTask
		completionTasks[idx] = completionTask
		completionStart = completionTask.GetStart()
	}

	// Advance to Completion phase
	advanceTo(t, eth, completionStart)

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
			if completionTasks[idx].ShouldRetry(ctx, logger, eth) {
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
	suite := StartFromCompletion(t, n, 100)
	defer suite.eth.Close()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "k")

	// Do Completion task
	var hasLeader bool
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		task := suite.completionTasks[idx]

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
	assert.False(t, suite.completionTasks[0].ShouldRetry(ctx, logger, eth))

	callOpts, err := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)

	phase, err := suite.eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint8(objects.Completion), phase)

	// event
	for j := 0; j < n; j++ {
		// simulate receiving ValidatorSetCompleted event for all participants
		suite.dkgStates[j].OnCompletion()
		assert.Equal(t, suite.dkgStates[j].Phase, objects.Completion)
	}
}

// We begin by submitting invalid information.
// This test is meant to raise an error resulting from an invalid argument
// for the Ethereum interface.
func TestCompletion_Group_2_Bad1(t *testing.T) {
	n := 4
	ecdsaPrivateKeys, _ := dtest.InitializePrivateKeysAndAccounts(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a task to share distribution and make sure it succeeds
	state := objects.NewDkgState(acct)
	task := NewCompletionTask(state, 1, 100)
	log := logger.WithField("TaskID", "foo")

	err := task.Initialize(ctx, log, eth)
	assert.NotNil(t, err)
}

// We test to ensure that everything behaves correctly.
func TestCompletion_Group_2_Bad2(t *testing.T) {
	n := 4
	ecdsaPrivateKeys, _ := dtest.InitializePrivateKeysAndAccounts(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Do bad Completion task
	state := objects.NewDkgState(acct)
	log := logger.WithField("TaskID", "foo")
	task := NewCompletionTask(state, 1, 100)

	err := task.Initialize(ctx, log, eth)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// We complete everything correctly, but we do not complete in time
func TestCompletion_Group_2_Bad3(t *testing.T) {
	n := 4
	suite := StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.eth.Close()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do GPKj Submission task
	tasksVec := suite.gpkjSubmissionTasks
	for idx := 0; idx < n; idx++ {

		err := tasksVec[idx].Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		err = tasksVec[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, tasksVec[idx].Success)
	}

	height, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	completionTasks := make([]*CompletionTask, n)
	var completionStart, completionEnd uint64
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		_, completionTask := dkgevents.UpdateStateOnGPKJSubmissionComplete(state, height)
		completionTasks[idx] = completionTask
		completionStart = completionTask.GetStart()
		completionEnd = completionTask.GetEnd()
	}

	// Advance to Completion phase
	advanceTo(t, eth, completionStart)

	// Advance to end of Completion phase
	advanceTo(t, eth, completionEnd)
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
	suite := StartFromCompletion(t, n, 40)
	defer suite.eth.Close()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do Completion task
	tasksVec := suite.completionTasks
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
	// then all tasks will have ShouldRetry() returning true
	assert.False(t, tasksVec[0].ShouldRetry(ctx, logger, eth))
}

func TestCompletion_Group_3_ShouldRetry_returnsTrue(t *testing.T) {
	n := 4
	suite := StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.eth.Close()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do GPKj Submission task
	for idx := 0; idx < n; idx++ {
		err := suite.gpkjSubmissionTasks[idx].Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		err = suite.gpkjSubmissionTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, suite.gpkjSubmissionTasks[idx].Success)
	}

	height, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	disputeGPKjTasks := make([]*DisputeGPKjTask, n)
	completionTasks := make([]*CompletionTask, n)
	var completionStart uint64
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		disputeGPKjTask, completionTask := dkgevents.UpdateStateOnGPKJSubmissionComplete(state, height)
		disputeGPKjTasks[idx] = disputeGPKjTask
		completionTasks[idx] = completionTask
		completionStart = completionTask.GetStart()
	}

	// Advance to Completion phase
	advanceTo(t, eth, completionStart)
	eth.Commit()

	err = completionTasks[0].Initialize(ctx, logger, eth)
	assert.Nil(t, err)

	shouldRetry := completionTasks[0].ShouldRetry(ctx, logger, eth)
	assert.True(t, shouldRetry)
}
