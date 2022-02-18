package dkgtasks_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/dkg/dtest"
	"github.com/MadBase/MadNet/blockchain/objects"
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
	tasks := suite.mpkSubmissionTasks
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		err := tasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()

		//This is because the MPK submission task should be
		//executed only by 1 validator
		if idx == 0 {
			assert.True(t, tasks[idx].Success)
		} else {
			assert.False(t, tasks[idx].Success)
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
	err := task.Initialize(ctx, logger, eth, dkgStates[0])
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
	_, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a task to share distribution and make sure it succeeds
	state := objects.NewDkgState(acct)
	task := dkgtasks.NewMPKSubmissionTask(state, 1, 100)
	log := logger.WithField("TaskID", "foo")

	err := task.Initialize(ctx, log, eth, nil)
	assert.NotNil(t, err)
}

// We force an error.
// This is caused by submitting invalid state information by not successfully
// completing KeyShareSubmission phase.
func TestMPKSubmissionBad4(t *testing.T) {
	n := 4
	_, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Do MPK Submission task
	state := objects.NewDkgState(acct)
	log := logger.WithField("TaskID", "foo")
	task := dkgtasks.NewMPKSubmissionTask(state, 1, 100)
	err := task.Initialize(ctx, log, eth, state)
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
	tasks := suite.mpkSubmissionTasks
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]

		err := tasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()

		//This is because the MPK submission task should be
		//executed only by 1 validator
		if idx == 0 {
			assert.True(t, tasks[idx].Success)
		} else {
			assert.False(t, tasks[idx].Success)
		}

		shouldRetry := tasks[idx].ShouldRetry(ctx, logger, eth)
		assert.False(t, shouldRetry)
	}
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
		tasks[idx].State.MasterPublicKey[0] = big.NewInt(1)

		shouldRetry := tasks[idx].ShouldRetry(ctx, logger, eth)
		assert.True(t, shouldRetry)
	}
}
