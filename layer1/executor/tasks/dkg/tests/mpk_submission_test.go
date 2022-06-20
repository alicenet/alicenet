//go:build integration

package dkg_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/testutils"
	"github.com/MadBase/MadNet/layer1/executor/tasks/dkg"
	dkgState "github.com/MadBase/MadNet/layer1/executor/tasks/dkg/state"
	dkgTestUtils "github.com/MadBase/MadNet/layer1/executor/tasks/dkg/testutils"

	"github.com/MadBase/MadNet/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//We test to ensure that everything behaves correctly.
func TestMPKSubmission_Group_1_GoodAllValid(t *testing.T) {
	n := 4
	suite := dkgTestUtils.StartFromKeyShareSubmissionPhase(t, n, 0, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do MPK Submission task
	tasksVec := suite.MpkSubmissionTasks
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		err := tasksVec[idx].Initialize(ctx, logger, eth)
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
		callOpts, err := eth.GetCallOpts(context.Background(), acct)
		assert.Nil(t, err)
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
func TestMPKSubmission_Group_1_Bad1(t *testing.T) {
	// Perform correct registration setup.

	// Perform correct share submission

	// After shares have been submitted, quickly proceed through the mpk
	// submission phase.

	// After the completion of the mpk submission phase, cause a validator
	// to attempt to submit the mpk.
	// This should result in an error.
	// EthDKG restart should be required.
	n := 6
	suite := dkgTestUtils.StartFromKeyShareSubmissionPhase(t, n, 0, 100)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	task := suite.MpkSubmissionTasks[0]
	err := task.Initialize(ctx, logger, eth)
	assert.Nil(t, err)
	eth.Commit()

	// Advance to gpkj submission phase; note we did *not* submit MPK
	testutils.AdvanceTo(t, eth, task.Start+dkgStates[0].PhaseLength)

	// Do MPK Submission task

	err = task.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
}

// We force an error.
// This is caused by submitting invalid state information (state is nil).
func TestMPKSubmission_Group_1_Bad2(t *testing.T) {
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
	state := dkgState.NewDkgState(acct)
	task := dkg.NewMPKSubmissionTask(state, 1, 100)
	log := logger.WithField("TaskID", "foo")

	err := task.Initialize(ctx, log, eth)
	assert.NotNil(t, err)
}

// We force an error.
// This is caused by submitting invalid state information by not successfully
// completing KeyShareSubmission phase.
func TestMPKSubmission_Group_2_Bad4(t *testing.T) {
	n := 4
	ecdsaPrivateKeys, _ := testutils.InitializePrivateKeysAndAccounts(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Do MPK Submission task
	state := dkgState.NewDkgState(acct)
	log := logger.WithField("TaskID", "foo")
	task := dkg.NewMPKSubmissionTask(state, 1, 100)

	err := task.Initialize(ctx, log, eth)
	assert.NotNil(t, err)
}

func TestMPKSubmission_Group_2_ShouldRetry_returnsFalse(t *testing.T) {
	n := 4
	suite := dkgTestUtils.StartFromKeyShareSubmissionPhase(t, n, 0, 40)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do MPK Submission task
	tasksVec := suite.MpkSubmissionTasks
	var hadLeaders bool
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		err := tasksVec[idx].Initialize(ctx, logger, eth)
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
	// then all tasks will have ShouldExecute() returning true
	assert.False(t, tasksVec[0].ShouldRetry(ctx, logger, eth))
}

func TestMPKSubmission_Group_2_ShouldRetry_returnsTrue(t *testing.T) {
	n := 4
	suite := dkgTestUtils.StartFromKeyShareSubmissionPhase(t, n, 0, 100)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do MPK Submission task
	tasks := suite.MpkSubmissionTasks
	for idx := 0; idx < n; idx++ {
		taskState := tasks[idx].State.(*dkgState.DkgState)
		taskState.MasterPublicKey[0] = big.NewInt(1)

		shouldRetry := tasks[idx].ShouldRetry(ctx, logger, eth)
		assert.True(t, shouldRetry)
	}
}

func TestMPKSubmission_Group_2_LeaderElection(t *testing.T) {
	n := 4
	suite := dkgTestUtils.StartFromKeyShareSubmissionPhase(t, n, 0, 100)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
	logger := logging.GetLogger("test").WithField("Validator", "")
	leaders := 0
	// Do MPK Submission task
	tasksVec := suite.MpkSubmissionTasks
	for idx := 0; idx < n; idx++ {
		state := suite.DKGStates[idx]

		err := tasksVec[idx].Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		//tasks[idx].State.MasterPublicKey[0] = big.NewInt(1)

		if tasksVec[idx].AmILeading(ctx, eth, logger, state) {
			leaders++
		}
	}

	assert.Greater(t, leaders, 0)
}
