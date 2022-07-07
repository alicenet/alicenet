//go:build integration

package tests

import (
	"context"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/utils"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/test/mocks"
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

	dkgState, err := state.GetDkgState(suite.DKGStatesDbs[0])
	assert.Nil(t, err)
	task := suite.MpkSubmissionTasks[0]
	err = task.Initialize(ctx, nil, suite.DKGStatesDbs[0], fixture.Logger, suite.Eth, "MPKSubmissionTask", "task-id", nil)
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

// We force an error.
// This is caused by submitting invalid state information (state is nil).
func TestMPKSubmission_Group_1_Bad2(t *testing.T) {
	task := dkg.NewMPKSubmissionTask(1, 100)
	db := mocks.NewTestDB()
	log := logging.GetLogger("test").WithField("test", "test")

	err := task.Initialize(context.Background(), nil, db, log, nil, "", "", nil)
	assert.Nil(t, err)

	taskErr := task.Prepare(context.Background())
	assert.NotNil(t, taskErr)
	assert.False(t, taskErr.IsRecoverable())

}

// We force an error.
// This is caused by submitting invalid state information by not successfully
// completing KeyShareSubmission phase.
func TestMPKSubmission_Group_2_Bad4(t *testing.T) {
	n := 4
	fixture := setupEthereum(t, n)
	suite := StartFromKeyShareSubmissionPhase(t, fixture, 0, 100)
	ctx := context.Background()

	// Do MPK Submission task
	dkgState := state.NewDkgState(suite.Eth.GetDefaultAccount())
	err := state.SaveDkgState(suite.DKGStatesDbs[0], dkgState)
	assert.Nil(t, err)

	task := suite.MpkSubmissionTasks[0]
	err = task.Initialize(ctx, nil, suite.DKGStatesDbs[0], fixture.Logger, suite.Eth, "MPKSubmissionTask", "task-id", nil)
	assert.Nil(t, err)

	taskErr := task.Prepare(ctx)
	assert.NotNil(t, taskErr)
	assert.False(t, taskErr.IsRecoverable())
}

func TestMPKSubmission_Group_2_LeaderElection(t *testing.T) {
	n := 4
	fixture := setupEthereum(t, n)
	suite := StartFromKeyShareSubmissionPhase(t, fixture, 0, 100)
	ctx := context.Background()
	leaders := 0

	for idx := 0; idx < n; idx++ {
		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)

		task := suite.MpkSubmissionTasks[idx]
		err = task.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "MPKSubmissionTask", "task-id", nil)
		assert.Nil(t, err)

		err = task.Prepare(ctx)
		assert.Nil(t, err)

		amILeading := utils.AmILeading(suite.Eth, ctx, fixture.Logger, int(task.GetStart()), task.StartBlockHash[:], n, dkgState.Index)
		if amILeading {
			leaders++
		}
	}

	assert.Equal(t, leaders, 1)
}
