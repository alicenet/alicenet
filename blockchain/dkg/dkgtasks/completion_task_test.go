//go:build integration

package dkgtasks_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicenet/alicenet/blockchain/dkg/dkgevents"
	"github.com/alicenet/alicenet/blockchain/dkg/dkgtasks"
	"github.com/alicenet/alicenet/blockchain/dkg/dtest"
	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/alicenet/alicenet/logging"
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
		state := dkgStates[idx]

		dkgData := objects.NewETHDKGTaskData(state)
		err := suite.gpkjSubmissionTasks[idx].Initialize(ctx, logger, eth, dkgData)
		assert.Nil(t, err)
		err = suite.gpkjSubmissionTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, suite.gpkjSubmissionTasks[idx].Success)
	}

	height, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	disputeGPKjTasks := make([]*dkgtasks.DisputeGPKjTask, n)
	completionTasks := make([]*dkgtasks.CompletionTask, n)
	var completionStart uint64
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		disputeGPKjTask, _, _, completionTask, start, _ := dkgevents.UpdateStateOnGPKJSubmissionComplete(state, logger, height)
		disputeGPKjTasks[idx] = disputeGPKjTask
		completionTasks[idx] = completionTask
		completionStart = start
	}

	// Advance to Completion phase
	advanceTo(t, eth, completionStart)

	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		dkgData := objects.NewETHDKGTaskData(state)
		err := completionTasks[idx].Initialize(ctx, logger, eth, dkgData)
		assert.Nil(t, err)
		amILeading := completionTasks[idx].AmILeading(ctx, eth, logger)
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

		dkgData := objects.NewETHDKGTaskData(state)
		err := task.Initialize(ctx, logger, eth, dkgData)
		assert.Nil(t, err)
		amILeading := task.AmILeading(ctx, eth, logger)

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

	phase, err := suite.eth.Contracts().Ethdkg().GetETHDKGPhase(eth.GetCallOpts(ctx, suite.eth.GetDefaultAccount()))
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
	task := dkgtasks.NewCompletionTask(state, 1, 100)
	log := logger.WithField("TaskID", "foo")

	dkgData := objects.NewETHDKGTaskData(state)
	err := task.Initialize(ctx, log, eth, dkgData)
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
	task := dkgtasks.NewCompletionTask(state, 1, 100)
	dkgData := objects.NewETHDKGTaskData(state)
	err := task.Initialize(ctx, log, eth, dkgData)
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
	tasks := suite.gpkjSubmissionTasks
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		dkgData := objects.NewETHDKGTaskData(state)
		err := tasks[idx].Initialize(ctx, logger, eth, dkgData)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, tasks[idx].Success)
	}

	height, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	completionTasks := make([]*dkgtasks.CompletionTask, n)
	var completionStart, completionEnd uint64
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		_, _, _, completionTask, start, end := dkgevents.UpdateStateOnGPKJSubmissionComplete(state, logger, height)
		completionTasks[idx] = completionTask
		completionStart = start
		completionEnd = end
	}

	// Advance to Completion phase
	advanceTo(t, eth, completionStart)

	// Advance to end of Completion phase
	advanceTo(t, eth, completionEnd)
	eth.Commit()

	// Do bad Completion task; this should fail because we are past
	state := dkgStates[0]
	dkgData := objects.NewETHDKGTaskData(state)
	err = completionTasks[0].Initialize(ctx, logger, eth, dkgData)
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
	tasks := suite.completionTasks
	var hadLeaders bool
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		dkgData := objects.NewETHDKGTaskData(state)
		err := tasks[idx].Initialize(ctx, logger, eth, dkgData)
		assert.Nil(t, err)
		amILeading := tasks[idx].AmILeading(ctx, eth, logger)

		if amILeading {
			hadLeaders = true
			// only perform ETHDKG completion if validator is leading
			assert.True(t, tasks[idx].ShouldRetry(ctx, logger, eth))
			err = tasks[idx].DoWork(ctx, logger, eth)
			assert.Nil(t, err)
			assert.False(t, tasks[idx].ShouldRetry(ctx, logger, eth))
		}
	}

	assert.True(t, hadLeaders)

	// any task is able to tell if ETHDKG still needs completion
	// if for any reason no validator lead the process,
	// then all tasks will have ShouldRetry() returning true
	assert.False(t, tasks[0].ShouldRetry(ctx, logger, eth))
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
		state := dkgStates[idx]
		dkgData := objects.NewETHDKGTaskData(state)
		err := suite.gpkjSubmissionTasks[idx].Initialize(ctx, logger, eth, dkgData)
		assert.Nil(t, err)
		err = suite.gpkjSubmissionTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, suite.gpkjSubmissionTasks[idx].Success)
	}

	height, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	disputeGPKjTasks := make([]*dkgtasks.DisputeGPKjTask, n)
	completionTasks := make([]*dkgtasks.CompletionTask, n)
	var completionStart uint64
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		disputeGPKjTask, _, _, completionTask, start, _ := dkgevents.UpdateStateOnGPKJSubmissionComplete(state, logger, height)
		disputeGPKjTasks[idx] = disputeGPKjTask
		completionTasks[idx] = completionTask
		completionStart = start
	}

	// Advance to Completion phase
	advanceTo(t, eth, completionStart)
	eth.Commit()

	// Do bad Completion task; this should fail because we are past
	state := dkgStates[0]
	dkgData := objects.NewETHDKGTaskData(state)
	err = completionTasks[0].Initialize(ctx, logger, eth, dkgData)
	assert.Nil(t, err)

	shouldRetry := completionTasks[0].ShouldRetry(ctx, logger, eth)
	assert.True(t, shouldRetry)
}
