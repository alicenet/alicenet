//go:build integration

package fixed

import (
	"context"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/utils"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/logging"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

//We test to ensure that everything behaves correctly.
func TestMPKSubmission_Group_1_GoodAllValid(t *testing.T) {
	n := 4
	fixture := setupEthereum(t, n)
	suite := StartFromKeyShareSubmissionPhase(t, fixture, 0, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.Eth
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do MPK Submission task
	for idx := 0; idx < n; idx++ {

		mpkSubmissionTask := suite.MpkSubmissionTasks[idx]
		err := mpkSubmissionTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "MpkSubmissionTasks", "tak-id", nil)
		assert.Nil(t, err)
		err = mpkSubmissionTask.Prepare(ctx)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)

		shouldExecute, err := mpkSubmissionTask.ShouldExecute(ctx)
		assert.Nil(t, err)
		if shouldExecute {
			txn, taskError := mpkSubmissionTask.Execute(ctx)
			amILeading := utils.AmILeading(eth, ctx, logger, int(mpkSubmissionTask.GetStart()), mpkSubmissionTask.StartBlockHash[:], n, dkgState.Index)
			if amILeading {
				assert.Nil(t, taskError)
				rcptResponse, err := fixture.Watcher.Subscribe(ctx, txn, nil)
				assert.Nil(t, err)
				tests.WaitGroupReceipts(t, suite.Eth, []transaction.ReceiptResponse{rcptResponse})
			} else {
				assert.Nil(t, txn)
				assert.NotNil(t, taskError)
				assert.True(t, taskError.IsRecoverable())
			}
		}
	}

	// Validate MPK
	for idx, acct := range accounts {
		callOpts, err := eth.GetCallOpts(context.Background(), acct)
		assert.Nil(t, err)
		isMPKSet, err := ethereum.GetContracts().Ethdkg().IsMasterPublicKeySet(callOpts)
		assert.Nil(t, err)
		assert.True(t, isMPKSet)

		// check mpk
		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		if dkgState.MasterPublicKey[0].Cmp(big.NewInt(0)) == 0 ||
			dkgState.MasterPublicKey[1].Cmp(big.NewInt(0)) == 0 ||
			dkgState.MasterPublicKey[2].Cmp(big.NewInt(0)) == 0 ||
			dkgState.MasterPublicKey[3].Cmp(big.NewInt(0)) == 0 {
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
	fixture := setupEthereum(t, n)
	suite := StartFromKeyShareSubmissionPhase(t, fixture, 0, 100)
	ctx := context.Background()

	//dkgState := suite.DKGStatesDbs[0]
	dkgState, err := state.GetDkgState(suite.DKGStatesDbs[0])
	assert.Nil(t, err)
	task := suite.MpkSubmissionTasks[0]
	err = task.Initialize(ctx, nil, suite.DKGStatesDbs[0], fixture.Logger, suite.Eth, "ShareDistributionTask", "task-id", nil)
	assert.Nil(t, err)

	err = task.Prepare(ctx)
	assert.Nil(t, err)

	// Advance to gpkj submission phase; note we did *not* submit MPK
	tests.AdvanceTo(suite.Eth, task.Start+dkgState.PhaseLength)

	// Do MPK Submission task
	txn, err := task.Execute(ctx)
	assert.NotNil(t, err)
	assert.Nil(t, txn)
}

//// We force an error.
//// This is caused by submitting invalid state information (state is nil).
//func TestMPKSubmission_Group_1_Bad2(t *testing.T) {
//	n := 4
//	ecdsaPrivateKeys, _ := testutils.InitializePrivateKeysAndAccounts(n)
//	logger := logging.GetLogger("ethereum")
//	logger.SetLevel(logrus.DebugLevel)
//	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
//	defer eth.Close()
//
//	acct := eth.GetKnownAccounts()[0]
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	// Create a task to share distribution and make sure it succeeds
//	state := dkgState.NewDkgState(acct)
//	task := dkg.NewMPKSubmissionTask(state, 1, 100)
//	log := logger.WithField("TaskID", "foo")
//
//	err := task.Initialize(ctx, log, eth)
//	assert.NotNil(t, err)
//}

//
//// We force an error.
//// This is caused by submitting invalid state information by not successfully
//// completing KeyShareSubmission phase.
//func TestMPKSubmission_Group_2_Bad4(t *testing.T) {
//	n := 4
//	ecdsaPrivateKeys, _ := testutils.InitializePrivateKeysAndAccounts(n)
//	logger := logging.GetLogger("ethereum")
//	logger.SetLevel(logrus.DebugLevel)
//	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
//	defer eth.Close()
//
//	acct := eth.GetKnownAccounts()[0]
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	// Do MPK Submission task
//	state := state.NewDkgState(acct)
//	log := logger.WithField("TaskID", "foo")
//	task := dkg.NewMPKSubmissionTask(state, 1, 100)
//
//	err := task.Initialize(ctx, log, eth)
//	assert.NotNil(t, err)
//}
//
//func TestMPKSubmission_Group_2_ShouldRetry_returnsFalse(t *testing.T) {
//	n := 4
//	suite := dkgTestUtils.StartFromKeyShareSubmissionPhase(t, n, 0, 40)
//	defer suite.Eth.Close()
//	ctx := context.Background()
//	eth := suite.Eth
//	dkgStates := suite.DKGStates
//	logger := logging.GetLogger("test").WithField("Validator", "")
//
//	// Do MPK Submission task
//	tasksVec := suite.MpkSubmissionTasks
//	var hadLeaders bool
//	for idx := 0; idx < n; idx++ {
//		state := dkgStates[idx]
//
//		err := mpkSubmissionTask.Initialize(ctx, logger, eth)
//		assert.Nil(t, err)
//		amILeading := mpkSubmissionTask.AmILeading(ctx, eth, logger, state)
//
//		if amILeading {
//			hadLeaders = true
//			// only perform MPK submission if validator is leading
//			assert.True(t, mpkSubmissionTask.ShouldRetry(ctx, logger, eth))
//			err = mpkSubmissionTask.DoWork(ctx, logger, eth)
//			assert.Nil(t, err)
//			assert.False(t, mpkSubmissionTask.ShouldRetry(ctx, logger, eth))
//		}
//	}
//
//	// make sure there were elected leaders
//	assert.True(t, hadLeaders)
//
//	// any task is able to tell if MPK still needs submission.
//	// if for any reason no validator lead the submission,
//	// then all tasks will have ShouldExecute() returning true
//	assert.False(t, tasksVec[0].ShouldRetry(ctx, logger, eth))
//}
//
//func TestMPKSubmission_Group_2_ShouldRetry_returnsTrue(t *testing.T) {
//	n := 4
//	suite := dkgTestUtils.StartFromKeyShareSubmissionPhase(t, n, 0, 100)
//	defer suite.Eth.Close()
//	ctx := context.Background()
//	eth := suite.Eth
//	logger := logging.GetLogger("test").WithField("Validator", "")
//
//	// Do MPK Submission task
//	tasks := suite.MpkSubmissionTasks
//	for idx := 0; idx < n; idx++ {
//		taskState := tasks[idx].State.(*state.DkgState)
//		taskState.MasterPublicKey[0] = big.NewInt(1)
//
//		shouldRetry := tasks[idx].ShouldRetry(ctx, logger, eth)
//		assert.True(t, shouldRetry)
//	}
//}
//
//func TestMPKSubmission_Group_2_LeaderElection(t *testing.T) {
//	n := 4
//	suite := dkgTestUtils.StartFromKeyShareSubmissionPhase(t, n, 0, 100)
//	defer suite.Eth.Close()
//	ctx := context.Background()
//	eth := suite.Eth
//	logger := logging.GetLogger("test").WithField("Validator", "")
//	leaders := 0
//	// Do MPK Submission task
//	tasksVec := suite.MpkSubmissionTasks
//	for idx := 0; idx < n; idx++ {
//		state := suite.DKGStates[idx]
//
//		err := mpkSubmissionTask.Initialize(ctx, logger, eth)
//		assert.Nil(t, err)
//		//tasks[idx].State.MasterPublicKey[0] = big.NewInt(1)
//
//		if mpkSubmissionTask.AmILeading(ctx, eth, logger, state) {
//			leaders++
//		}
//	}
//
//	assert.Greater(t, leaders, 0)
//}
