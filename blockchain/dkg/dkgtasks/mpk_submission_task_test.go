package dkgtasks_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/dkg/dtest"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//We test to ensure that everything behaves correctly.
func TestMPKSubmissionGoodAllValid(t *testing.T) {
	n := 4
	suite := StartFromKeyShareSubmissionPhase(t, n, 0, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do MPK Submission task
	tasksVec := suite.mpkSubmissionTasks
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		dkgData := tasks.NewTaskData(state)
		err := tasksVec[idx].Initialize(ctx, logger, eth, dkgData)
		assert.Nil(t, err)
		amILeading := tasksVec[idx].AmILeading(ctx, eth, logger, state)
		err = tasksVec[idx].DoWork(ctx, logger, eth)
		if amILeading {
			assert.Nil(t, err)
			assert.True(t, tasksVec[idx].Success)
		} else {
			if tasksVec[idx].ShouldRetry(ctx, logger, eth) {
				assert.NotNil(t, err)
				assert.False(t, tasksVec[idx].Success)
			} else {
				assert.Nil(t, err)
				assert.True(t, tasksVec[idx].Success)
			}

		}
	}

	// Validate MPK
	for idx, acct := range accounts {
		callOpts := eth.GetCallOpts(context.Background(), acct)
		isMPKSet, err := eth.Contracts().Ethdkg().IsMasterPublicKeySet(callOpts)
		assert.Nil(t, err)
		assert.True(t, isMPKSet)

		// check mpk
		if dkgStates[idx].MasterPublicKey[0].Cmp(big.NewInt(0)) == 0 ||
			dkgStates[idx].MasterPublicKey[1].Cmp(big.NewInt(0)) == 0 ||
			dkgStates[idx].MasterPublicKey[2].Cmp(big.NewInt(0)) == 0 ||
			dkgStates[idx].MasterPublicKey[3].Cmp(big.NewInt(0)) == 0 {
			t.Fatal("Invalid master public key")
		}
	}
}

// Here we test for invalid mpk submission.
// In this test, *no* validator should submit an mpk.
// After ending the MPK submission phase, validators should attempt
// to submit the mpk; this should raise an error.
func TestMPKSubmissionBad1(t *testing.T) {
	// Perform correct registration setup.

	// Perform correct share submission

	// After shares have been submitted, quickly proceed through the mpk
	// submission phase.

	// After the completion of the mpk submission phase, cause a validator
	// to attempt to submit the mpk.
	// This should result in an error.
	// EthDKG restart should be required.
	n := 6
	suite := StartFromKeyShareSubmissionPhase(t, n, 0, 100)
	defer suite.eth.Close()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	task := suite.mpkSubmissionTasks[0]
	dkgData := tasks.NewTaskData(dkgStates[0])
	err := task.Initialize(ctx, logger, eth, dkgData)
	assert.Nil(t, err)
	eth.Commit()

	// Advance to gpkj submission phase; note we did *not* submit MPK
	advanceTo(t, eth, task.Start+dkgStates[0].PhaseLength)

	// Do MPK Submission task

	err = task.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
}

// We force an error.
// This is caused by submitting invalid state information (state is nil).
func TestMPKSubmissionBad2(t *testing.T) {
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
	task := dkgtasks.NewMPKSubmissionTask(state, 1, 100)
	log := logger.WithField("TaskID", "foo")

	dkgData := tasks.NewTaskData(state)
	err := task.Initialize(ctx, log, eth, dkgData)
	assert.NotNil(t, err)
}

// We force an error.
// This is caused by submitting invalid state information by not successfully
// completing KeyShareSubmission phase.
func TestMPKSubmissionBad4(t *testing.T) {
	n := 4
	ecdsaPrivateKeys, _ := dtest.InitializePrivateKeysAndAccounts(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Do MPK Submission task
	state := objects.NewDkgState(acct)
	log := logger.WithField("TaskID", "foo")
	task := dkgtasks.NewMPKSubmissionTask(state, 1, 100)
	dkgData := tasks.NewTaskData(state)
	err := task.Initialize(ctx, log, eth, dkgData)
	assert.NotNil(t, err)
}

func TestMPKSubmission_ShouldRetry_returnsFalse(t *testing.T) {
	n := 4
	suite := StartFromKeyShareSubmissionPhase(t, n, 0, 100)
	defer suite.eth.Close()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do MPK Submission task
	tasksVec := suite.mpkSubmissionTasks
	var hadLeaders bool
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		dkgData := tasks.NewTaskData(state)
		err := tasksVec[idx].Initialize(ctx, logger, eth, dkgData)
		assert.Nil(t, err)
		amILeading := tasksVec[idx].AmILeading(ctx, eth, logger, state)

		if amILeading {
			hadLeaders = true
			// only perform MPK submission if validator is leading
			assert.True(t, tasksVec[idx].ShouldRetry(ctx, logger, eth))
			err = tasksVec[idx].DoWork(ctx, logger, eth)
			assert.Nil(t, err)
			assert.False(t, tasksVec[idx].ShouldRetry(ctx, logger, eth))
		}
	}

	// make sure there were elected leaders
	assert.True(t, hadLeaders)

	// any task is able to tell if MPK still needs submission.
	// if for any reason no validator lead the submission,
	// then all tasks will have ShouldRetry() returning true
	assert.False(t, tasksVec[0].ShouldRetry(ctx, logger, eth))
}

func TestMPKSubmission_ShouldRetry_returnsTrue(t *testing.T) {
	n := 4
	suite := StartFromKeyShareSubmissionPhase(t, n, 0, 100)
	defer suite.eth.Close()
	ctx := context.Background()
	eth := suite.eth
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do MPK Submission task
	tasks := suite.mpkSubmissionTasks
	for idx := 0; idx < n; idx++ {
		taskState := tasks[idx].State.(*objects.DkgState)
		taskState.MasterPublicKey[0] = big.NewInt(1)

		shouldRetry := tasks[idx].ShouldRetry(ctx, logger, eth)
		assert.True(t, shouldRetry)
	}
}

func TestMPKSubmission_LeaderElection(t *testing.T) {
	n := 4
	suite := StartFromKeyShareSubmissionPhase(t, n, 0, 100)
	defer suite.eth.Close()
	ctx := context.Background()
	eth := suite.eth
	logger := logging.GetLogger("test").WithField("Validator", "")
	leaders := 0
	// Do MPK Submission task
	tasksVec := suite.mpkSubmissionTasks
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		dkgData := tasks.NewTaskData(state)
		tasksVec[idx].Initialize(ctx, logger, eth, dkgData)
		//tasks[idx].State.MasterPublicKey[0] = big.NewInt(1)

		if tasksVec[idx].AmILeading(ctx, eth, logger, state) {
			leaders++
		}
	}

	assert.Greater(t, leaders, 0)
}
