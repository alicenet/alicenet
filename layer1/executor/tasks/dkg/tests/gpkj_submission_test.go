//go:build integration

package dkg_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/alicenet/alicenet/blockchain/testutils"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	dkgState "github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	dkgTestUtils "github.com/alicenet/alicenet/layer1/executor/tasks/dkg/testutils"
	"github.com/alicenet/alicenet/layer1/monitor/interfaces"
	"github.com/alicenet/alicenet/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//We test to ensure that everything behaves correctly.
func TestGPKjSubmission_Group_1_GoodAllValid(t *testing.T) {
	n := 4
	suite := dkgTestUtils.StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
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

	// Check gpkjs and signatures are present and valid
	for idx, acct := range accounts {
		callOpts, err := eth.GetCallOpts(context.Background(), acct)
		assert.Nil(t, err)
		p, err := eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		// check gpkj
		stateGPKJ := dkgStates[idx].Participants[acct.Address].GPKj
		if (p.Gpkj[0].Cmp(stateGPKJ[0]) != 0) || (p.Gpkj[1].Cmp(stateGPKJ[1]) != 0) || (p.Gpkj[2].Cmp(stateGPKJ[2]) != 0) || (p.Gpkj[3].Cmp(stateGPKJ[3]) != 0) {
			t.Fatal("Invalid gpkj")
		}
	}
}

// We begin by submitting invalid information.
// Here, we submit nil for the state interface;
// this should raise an error.
func TestGPKjSubmission_Group_1_Bad1(t *testing.T) {
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
	adminHandler := new(interfaces.MockAdminHandler)
	task := dkg.NewGPKjSubmissionTask(state, 1, 100, adminHandler)
	log := logger.WithField("TaskID", "foo")

	err := task.Initialize(ctx, log, eth)
	assert.NotNil(t, err)
}

// We test to ensure that everything behaves correctly.
// Here, we should raise an error because we did not successfully complete
// the key share submission phase.
func TestGPKjSubmission_Group_1_Bad2(t *testing.T) {
	n := 4
	ecdsaPrivateKeys, _ := testutils.InitializePrivateKeysAndAccounts(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Do bad Share Dispute task
	state := dkgState.NewDkgState(acct)
	log := logger.WithField("TaskID", "foo")
	adminHandler := new(interfaces.MockAdminHandler)
	task := dkg.NewGPKjSubmissionTask(state, 1, 100, adminHandler)

	err := task.Initialize(ctx, log, eth)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// Here we test for an invalid gpkj submission.
// One or more validators should submit invalid gpkj information;
// that is, the gpkj public key and signature should not verify.
// This should result in no submission.
func TestGPKjSubmission_Group_2_Bad3(t *testing.T) {
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
	suite := dkgTestUtils.StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Initialize GPKj Submission task
	tasksVec := suite.GpkjSubmissionTasks
	for idx := 0; idx < n; idx++ {
		err := tasksVec[idx].Initialize(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
	}

	// Do GPKj Submission task; this will fail because invalid submission;
	// it does not pass the PairingCheck.
	task := tasksVec[0]

	taskState, ok := task.State.(*dkgState.DkgState)
	assert.True(t, ok)

	// Mess up GPKj; this will cause Execute to fail
	taskState.Participants[taskState.Account.Address].GPKj = [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
	err := task.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
}

func TestGPKjSubmission_Group_2_ShouldRetry_returnsFalse(t *testing.T) {
	n := 4
	suite := dkgTestUtils.StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
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

		shouldRetry := tasksVec[idx].ShouldRetry(ctx, logger, eth)
		assert.False(t, shouldRetry)
	}
}

func TestGPKjSubmission_Group_2_ShouldRetry_returnsTrue(t *testing.T) {
	n := 4
	suite := dkgTestUtils.StartFromMPKSubmissionPhase(t, n, 100)
	defer suite.Eth.Close()
	ctx := context.Background()
	eth := suite.Eth
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do GPKj Submission task
	tasksVec := suite.GpkjSubmissionTasks
	for idx := 0; idx < n; idx++ {
		err := tasksVec[idx].Initialize(ctx, logger, eth)
		assert.Nil(t, err)

		shouldRetry := tasksVec[idx].ShouldRetry(ctx, logger, eth)
		assert.True(t, shouldRetry)
	}
}
