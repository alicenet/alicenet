//go:build integration

package tests

import (
	"context"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

//We test to ensure that everything behaves correctly.
func TestGPKjSubmission_Group_1_GoodAllValid(t *testing.T) {
	n := 4
	fixture := setupEthereum(t, n)
	suite := StartFromMPKSubmissionPhase(t, fixture, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.Eth

	// Do GPKj Submission task
	var receiptResponses []transaction.ReceiptResponse
	for idx := 0; idx < n; idx++ {
		gpkjSubmissionTask := suite.GpkjSubmissionTasks[idx]
		err := gpkjSubmissionTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "GpkjSubmissionTask", "tak-id", nil)
		assert.Nil(t, err)
		err = gpkjSubmissionTask.Prepare(ctx)
		assert.Nil(t, err)

		shouldExecute, err := gpkjSubmissionTask.ShouldExecute(ctx)
		assert.Nil(t, err)
		assert.True(t, shouldExecute)

		txn, err := gpkjSubmissionTask.Execute(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, txn)

		rcptResponse, subsErr := fixture.Watcher.Subscribe(ctx, txn, nil)
		assert.Nil(t, subsErr)
		receiptResponses = append(receiptResponses, rcptResponse)
	}

	tests.WaitGroupReceipts(t, suite.Eth, receiptResponses)

	// Check gpkjs and signatures are present and valid
	for idx, acct := range accounts {
		callOpts, err := eth.GetCallOpts(context.Background(), acct)
		assert.Nil(t, err)
		p, err := ethereum.GetContracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)

		// check gpkj
		stateGPKJ := dkgState.Participants[acct.Address].GPKj
		if (p.Gpkj[0].Cmp(stateGPKJ[0]) != 0) || (p.Gpkj[1].Cmp(stateGPKJ[1]) != 0) || (p.Gpkj[2].Cmp(stateGPKJ[2]) != 0) || (p.Gpkj[3].Cmp(stateGPKJ[3]) != 0) {
			t.Fatal("Invalid gpkj")
		}
	}
}

func TestGPKjSubmission_Group_1_Bad1(t *testing.T) {
	n := 6
	fixture := setupEthereum(t, n)
	suite := StartFromMPKSubmissionPhase(t, fixture, 100)
	ctx := context.Background()

	dkgState, err := state.GetDkgState(suite.DKGStatesDbs[0])
	assert.Nil(t, err)
	task := suite.GpkjSubmissionTasks[0]
	err = task.Initialize(ctx, nil, suite.DKGStatesDbs[0], fixture.Logger, suite.Eth, "GpkjSubmissionTask", "task-id", nil)
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

// We begin by submitting invalid information.
// Here, we submit nil for the state interface;
// this should raise an error.
func TestGPKjSubmission_Group_1_Bad2(t *testing.T) {
	task := dkg.NewGPKjSubmissionTask(1, 100, nil)
	db := mocks.NewTestDB()
	log := logging.GetLogger("test").WithField("test", "test")

	err := task.Initialize(context.Background(), nil, db, log, nil, "", "", nil)
	assert.Nil(t, err)

	taskErr := task.Prepare(context.Background())
	assert.NotNil(t, taskErr)
	assert.False(t, taskErr.IsRecoverable())
}

// We test to ensure that everything behaves correctly.
// Here, we should raise an error because we did not successfully complete
// the key share submission phase.
func TestGPKjSubmission_Group_1_Bad3(t *testing.T) {
	n := 4
	fixture := setupEthereum(t, n)
	suite := StartFromMPKSubmissionPhase(t, fixture, 100)
	ctx := context.Background()

	// Do MPK Submission task
	dkgState := state.NewDkgState(suite.Eth.GetDefaultAccount())
	err := state.SaveDkgState(suite.DKGStatesDbs[0], dkgState)
	assert.Nil(t, err)

	task := suite.GpkjSubmissionTasks[0]
	err = task.Initialize(ctx, nil, suite.DKGStatesDbs[0], fixture.Logger, suite.Eth, "GpkjSubmissionTask", "task-id", nil)
	assert.Nil(t, err)

	taskErr := task.Prepare(ctx)
	assert.NotNil(t, taskErr)
	assert.False(t, taskErr.IsRecoverable())
}

// Here we test for an invalid gpkj submission.
// One or more validators should submit invalid gpkj information;
// that is, the gpkj public key and signature should not verify.
// This should result in no submission.
func TestGPKjSubmission_Group_2_Bad4(t *testing.T) {
	// Perform correct registration setup.

	// Perform correct share submission

	// Correctly submit the mpk

	// After correctly submitting the mpk,
	// one or more validators should submit invalid gpkj information.
	// This will consist of a signature and public key which are not valid;
	// that is, attempting to verify initialMessage with the signature
	// and public key should fail verification.
	// This should raise an error, as this is not allowed by the protocol.
	n := 4
	fixture := setupEthereum(t, n)
	suite := StartFromMPKSubmissionPhase(t, fixture, 100)
	ctx := context.Background()

	// Initialize GPKj Submission task
	tasksVec := suite.GpkjSubmissionTasks
	for idx := 0; idx < n; idx++ {
		gpkjSubmissionTask := suite.GpkjSubmissionTasks[idx]
		err := gpkjSubmissionTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "GpkjSubmissionTask", "tak-id", nil)
		assert.Nil(t, err)
		err = gpkjSubmissionTask.Prepare(ctx)
		assert.Nil(t, err)
	}

	// Do GPKj Submission task; this will fail because invalid submission;
	// it does not pass the PairingCheck.
	badIdx := 0
	task := tasksVec[badIdx]
	dkgState, err := state.GetDkgState(suite.DKGStatesDbs[badIdx])
	assert.Nil(t, err)

	// Mess up GPKj; this will cause Execute to fail
	dkgState.Participants[dkgState.Account.Address].GPKj = [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
	err = state.SaveDkgState(suite.DKGStatesDbs[badIdx], dkgState)
	assert.Nil(t, err)
	_, taskErr := task.Execute(ctx)
	assert.NotNil(t, taskErr)
	assert.True(t, taskErr.IsRecoverable())
}
