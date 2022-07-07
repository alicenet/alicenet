//go:build integration

package tests

import (
	"context"
	"math/big"
	"testing"

	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

//Here we test the happy path.
func TestShareDistribution_Group_1_Good(t *testing.T) {
	n := 5
	fixture := setupEthereum(t, n)
	suite := StartFromRegistrationOpenPhase(t, fixture, 0, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()

	// Do Share Distribution task
	for idx := 0; idx < n; idx++ {
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		shareDistributionTask := suite.ShareDistTasks[idx]

		err := shareDistributionTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], logger, suite.Eth, "ShareDistributionTask", "task-id", nil)
		assert.Nil(t, err)
		err = shareDistributionTask.Prepare(ctx)
		assert.Nil(t, err)
		_, err = shareDistributionTask.Execute(ctx)
		assert.Nil(t, err)
	}
}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (invalid elliptic curve point).
// This should result in a failed submission.
func TestShareDistribution_Group_1_Bad1(t *testing.T) {
	n := 5
	fixture := setupEthereum(t, n)
	suite := StartFromRegistrationOpenPhase(t, fixture, 0, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()

	// Check public keys are present and valid
	for idx, acct := range accounts {
		callOpts, err := suite.Eth.GetCallOpts(context.Background(), acct)
		assert.Nil(t, err)
		p, err := ethereum.GetContracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)
		// check points
		publicKey := dkgState.TransportPublicKey
		if (publicKey[0].Cmp(p.PublicKey[0]) != 0) || (publicKey[1].Cmp(p.PublicKey[1]) != 0) {
			t.Fatal("Invalid public key")
		}
	}

	badIdx := n - 2
	for idx := 0; idx < n; idx++ {
		task := suite.ShareDistTasks[idx]
		err := task.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "ShareDistributionTask", "task-id", nil)
		assert.Nil(t, err)
		err = task.Prepare(ctx)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)

		com := dkgState.Participants[accounts[idx].Address].Commitments
		// if we're on the last account, we just add 1 to the first commitment (y component)
		if idx == badIdx {
			// Mess up one of the commitments (first coefficient)
			com[0][1].Add(com[0][1], big.NewInt(1))
		}
		dkgState.Participants[accounts[idx].Address].Commitments = com
		err = state.SaveDkgState(suite.DKGStatesDbs[idx], dkgState)
		assert.Nil(t, err)
		_, err = task.Execute(ctx)
		if idx == badIdx {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (identity element).
// This should result in a failed submission.
func TestShareDistribution_Group_1_Bad2(t *testing.T) {
	n := 4
	fixture := setupEthereum(t, n)
	suite := StartFromRegistrationOpenPhase(t, fixture, 0, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()

	// Check public keys are present and valid
	for idx, acct := range accounts {
		callOpts, err := suite.Eth.GetCallOpts(context.Background(), acct)
		assert.Nil(t, err)
		p, err := ethereum.GetContracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)
		// check points
		publicKey := dkgState.TransportPublicKey
		if (publicKey[0].Cmp(p.PublicKey[0]) != 0) || (publicKey[1].Cmp(p.PublicKey[1]) != 0) {
			t.Fatalf("Invalid public key: internal: %v | network: %v", publicKey, p.PublicKey)
		}
	}

	badIdx := n - 1
	for idx := 0; idx < n; idx++ {
		task := suite.ShareDistTasks[idx]
		err := task.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "ShareDistributionTask", "task-id", nil)
		assert.Nil(t, err)
		err = task.Prepare(ctx)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)

		com := dkgState.Participants[accounts[idx].Address].Commitments
		// if we're on the last account, change the one of the commitments to 0
		if idx == badIdx {
			// Mess up one of the commitments (first coefficient)
			com[0][0].Set(common.Big0)
			com[0][1].Set(common.Big0)
		}
		dkgState.Participants[accounts[idx].Address].Commitments = com
		err = state.SaveDkgState(suite.DKGStatesDbs[idx], dkgState)
		assert.Nil(t, err)
		_, err = task.Execute(ctx)
		// The last task should have failed
		if idx == badIdx {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (incorrect commitment length)
// This should result in a failed submission.
func TestShareDistribution_Group_2_Bad4(t *testing.T) {
	n := 5
	fixture := setupEthereum(t, n)
	suite := StartFromRegistrationOpenPhase(t, fixture, 0, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.Eth

	// Check public keys are present and valid
	for idx, acct := range accounts {
		callOpts, err := eth.GetCallOpts(context.Background(), acct)
		assert.Nil(t, err)
		p, err := ethereum.GetContracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)
		// check points
		publicKey := dkgState.TransportPublicKey
		if (publicKey[0].Cmp(p.PublicKey[0]) != 0) || (publicKey[1].Cmp(p.PublicKey[1]) != 0) {
			t.Fatalf("Invalid public key: internal: %v | network: %v", publicKey, p.PublicKey)
		}
	}

	badCommitmentIdx := n - 3
	for idx := 0; idx < n; idx++ {
		task := suite.ShareDistTasks[idx]
		err := task.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "ShareDistributionTask", "task-id", nil)
		assert.Nil(t, err)
		err = task.Prepare(ctx)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)

		// if we're on the last account, we just add 1 to the first commitment (y component)
		com := dkgState.Participants[accounts[idx].Address].Commitments
		if idx == badCommitmentIdx {
			// Mess up one of the commitments (incorrect length)
			com = com[:len(com)-1]
			dkgState.Participants[accounts[idx].Address].Commitments = com
		}

		err = state.SaveDkgState(suite.DKGStatesDbs[idx], dkgState)
		assert.Nil(t, err)
		_, err = task.Execute(ctx)
		// The last task should have failed
		if idx == badCommitmentIdx {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (incorrect encrypted shares length)
// This should result in a failed submission.
func TestShareDistribution_Group_2_Bad5(t *testing.T) {
	n := 6
	fixture := setupEthereum(t, n)
	suite := StartFromRegistrationOpenPhase(t, fixture, 0, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()

	badShareIdx := n - 2
	for idx := 0; idx < n; idx++ {
		task := suite.ShareDistTasks[idx]
		err := task.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "ShareDistributionTask", "task-id", nil)
		assert.Nil(t, err)
		err = task.Prepare(ctx)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)

		encryptedShares := dkgState.Participants[accounts[idx].Address].EncryptedShares
		if idx == badShareIdx {
			// Mess up one of the encryptedShares (incorrect length)
			encryptedShares = encryptedShares[:len(encryptedShares)-1]
			dkgState.Participants[accounts[idx].Address].EncryptedShares = encryptedShares
		}

		err = state.SaveDkgState(suite.DKGStatesDbs[idx], dkgState)
		assert.Nil(t, err)
		_, err = task.Execute(ctx)
		// The last task should have failed
		if idx == badShareIdx {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

// We begin by submitting invalid information;
// we submit nil state information
func TestShareDistribution_Group_2_Bad6(t *testing.T) {
	n := 5
	fixture := setupEthereum(t, n)
	eth := fixture.Client
	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a task to share distribution and make sure it succeeds
	dkgState := state.NewDkgState(acct)
	dkgDb := GetDKGDb(t)
	err := state.SaveDkgState(dkgDb, dkgState)
	assert.Nil(t, err)
	task := dkg.NewShareDistributionTask(dkgState.PhaseStart, dkgState.PhaseStart+dkgState.PhaseLength)
	err = task.Initialize(ctx, nil, dkgDb, fixture.Logger, eth, "ShareDistributionTask", "task-id", nil)
	assert.Nil(t, err)
	err = task.Prepare(ctx)
	assert.NotNil(t, err)
}

func TestShareDistribution_Group_3_ShouldRetryTrue(t *testing.T) {
	n := 5
	fixture := setupEthereum(t, n)
	suite := StartFromRegistrationOpenPhase(t, fixture, 0, 100)
	ctx := context.Background()

	// Do Share Distribution task
	for idx := 0; idx < n; idx++ {
		task := suite.ShareDistTasks[idx]
		err := task.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "ShareDistributionTask", "task-id", nil)
		assert.Nil(t, err)
		err = task.Prepare(ctx)
		assert.Nil(t, err)

		shouldRetry, _ := task.ShouldExecute(ctx)
		assert.Nil(t, err)
		assert.True(t, shouldRetry)
	}
}

func TestShareDistribution_Group_3_ShouldRetryFalse(t *testing.T) {
	n := 5
	fixture := setupEthereum(t, n)
	suite := StartFromRegistrationOpenPhase(t, fixture, 0, 100)
	ctx := context.Background()

	// Do Share Distribution task
	for idx := 0; idx < n; idx++ {
		task := suite.ShareDistTasks[idx]
		err := task.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "ShareDistributionTask", "task-id", nil)
		assert.Nil(t, err)
		err = task.Prepare(ctx)
		assert.Nil(t, err)
		_, err = task.Execute(ctx)
		assert.Nil(t, err)
		tests.MineFinalityDelayBlocks(suite.Eth)
		shouldRetry, _ := task.ShouldExecute(ctx)
		assert.Nil(t, err)
		assert.False(t, shouldRetry)
	}
}
