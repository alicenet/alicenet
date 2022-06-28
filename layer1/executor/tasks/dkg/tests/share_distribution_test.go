//go:build integration

package tests

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/alicenet/alicenet/blockchain/testutils"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	dkgTestUtils "github.com/alicenet/alicenet/layer1/executor/tasks/dkg/testutils"
	"github.com/alicenet/alicenet/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//Here we test the happy path.
func TestShareDistribution_Group_1_Good(t *testing.T) {
	n := 5
	suite := dkgTestUtils.StartFromRegistrationOpenPhase(t, n, 0, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()

	// Do Share Distribution task
	for idx := 0; idx < n; idx++ {
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		shareDistributionTask := suite.ShareDistTasks[idx]

		err := shareDistributionTask.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		err = shareDistributionTask.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, shareDistributionTask.Success)

	}
}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (invalid elliptic curve point).
// This should result in a failed submission.
func TestShareDistribution_Group_1_Bad1(t *testing.T) {
	n := 5
	suite := dkgTestUtils.StartFromRegistrationOpenPhase(t, n, 0, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()

	// Check public keys are present and valid
	for idx, acct := range accounts {
		callOpts, err := suite.Eth.GetCallOpts(context.Background(), acct)
		assert.Nil(t, err)
		p, err := suite.Eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		// check points
		publicKey := suite.DKGStates[idx].TransportPublicKey
		if (publicKey[0].Cmp(p.PublicKey[0]) != 0) || (publicKey[1].Cmp(p.PublicKey[1]) != 0) {
			t.Fatal("Invalid public key")
		}
	}

	badIdx := n - 2
	//tasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := suite.DKGStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		task := suite.ShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		com := state.Participants[accounts[idx].Address].Commitments
		// if we're on the last account, we just add 1 to the first commitment (y component)
		if idx == badIdx {
			// Mess up one of the commitments (first coefficient)
			com[0][1].Add(com[0][1], big.NewInt(1))
		}

		task.DoWork(ctx, logger, suite.Eth)

		suite.Eth.Commit()

		// The last task should have failed
		if idx == badIdx {
			assert.False(t, task.Success)
		} else {
			assert.True(t, task.Success)
		}
	}

	// // Double check to Make sure all transactions were good
	// rcpts, err := suite.Eth.Queue().WaitGroupTransactions(ctx, 1)
	// assert.Nil(t, err)

	// for _, rcpt := range rcpts {
	// 	assert.NotNil(t, rcpt)
	// 	assert.Equal(t, uint64(1), rcpt.Status)
	// }
}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (identity element).
// This should result in a failed submission.
func TestShareDistribution_Group_1_Bad2(t *testing.T) {
	n := 4
	suite := dkgTestUtils.StartFromRegistrationOpenPhase(t, n, 0, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()

	// Check public keys are present and valid
	for idx, acct := range accounts {
		callOpts, err := suite.Eth.GetCallOpts(context.Background(), acct)
		assert.Nil(t, err)
		p, err := suite.Eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		// check points
		publicKey := suite.DKGStates[idx].TransportPublicKey
		if (publicKey[0].Cmp(p.PublicKey[0]) != 0) || (publicKey[1].Cmp(p.PublicKey[1]) != 0) {
			t.Fatalf("Invalid public key: internal: %v | network: %v", publicKey, p.PublicKey)
		}
	}

	badIdx := n - 1
	for idx := 0; idx < n; idx++ {
		state := suite.DKGStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		task := suite.ShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		com := state.Participants[accounts[idx].Address].Commitments
		// if we're on the last account, change the one of the commitments to 0
		if idx == badIdx {
			// Mess up one of the commitments (first coefficient)
			com[0][0].Set(common.Big0)
			com[0][1].Set(common.Big0)
		}

		task.DoWork(ctx, logger, suite.Eth)

		suite.Eth.Commit()

		// The last task should have failed
		if idx == badIdx {
			assert.False(t, task.Success)
		} else {
			assert.True(t, task.Success)
		}
	}

	// // Double check to Make sure all transactions were good
	// rcpts, err := suite.Eth.Queue().WaitGroupTransactions(ctx, 1)
	// assert.Nil(t, err)

	// for _, rcpt := range rcpts {
	// 	assert.NotNil(t, rcpt)
	// 	assert.Equal(t, uint64(1), rcpt.Status)
	// }
}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (incorrect commitment length)
// This should result in a failed submission.
func TestShareDistribution_Group_2_Bad4(t *testing.T) {
	n := 5
	suite := dkgTestUtils.StartFromRegistrationOpenPhase(t, n, 0, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates

	// Check public keys are present and valid
	for idx, acct := range accounts {
		callOpts, err := eth.GetCallOpts(context.Background(), acct)
		assert.Nil(t, err)
		p, err := eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		// check points
		publicKey := dkgStates[idx].TransportPublicKey
		if (publicKey[0].Cmp(p.PublicKey[0]) != 0) || (publicKey[1].Cmp(p.PublicKey[1]) != 0) {
			t.Fatalf("Invalid public key: internal: %v | network: %v", publicKey, p.PublicKey)
		}
	}

	badCommitmentIdx := n - 3
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		task := suite.ShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		// if we're on the last account, we just add 1 to the first commitment (y component)
		com := state.Participants[accounts[idx].Address].Commitments
		if idx == badCommitmentIdx {
			// Mess up one of the commitments (incorrect length)
			com = com[:len(com)-1]
			state.Participants[accounts[idx].Address].Commitments = com
		}

		task.DoWork(ctx, logger, eth)

		eth.Commit()

		// The last task should have failed
		if idx == badCommitmentIdx {
			assert.False(t, task.Success)
		} else {
			assert.True(t, task.Success)
		}
	}

	// // Double check to Make sure all transactions were good
	// rcpts, err := eth.Queue().WaitGroupTransactions(ctx, 1)
	// assert.Nil(t, err)

	// for _, rcpt := range rcpts {
	// 	assert.NotNil(t, rcpt)
	// 	assert.Equal(t, uint64(1), rcpt.Status)
	// }
}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (incorrect encrypted shares length)
// This should result in a failed submission.
func TestShareDistribution_Group_2_Bad5(t *testing.T) {
	n := 6
	suite := dkgTestUtils.StartFromRegistrationOpenPhase(t, n, 0, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.Eth
	dkgStates := suite.DKGStates

	badShareIdx := n - 2
	//tasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		task := suite.ShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		encryptedShares := state.Participants[accounts[idx].Address].EncryptedShares
		if idx == badShareIdx {
			// Mess up one of the encryptedShares (incorrect length)
			encryptedShares = encryptedShares[:len(encryptedShares)-1]
			state.Participants[accounts[idx].Address].EncryptedShares = encryptedShares
		}

		task.DoWork(ctx, logger, eth)

		eth.Commit()

		// The last task should have failed
		if idx == badShareIdx {
			assert.False(t, task.Success)
		} else {
			assert.True(t, task.Success)
		}
	}

	// // Double check to Make sure all transactions were good
	// rcpts, err := eth.Queue().WaitGroupTransactions(ctx, 1)
	// assert.Nil(t, err)

	// for _, rcpt := range rcpts {
	// 	assert.NotNil(t, rcpt)
	// 	assert.Equal(t, uint64(1), rcpt.Status)
	// }
}

// We begin by submitting invalid information;
// we submit nil state information
func TestShareDistribution_Group_2_Bad6(t *testing.T) {
	n := 5
	ecdsaPrivateKeys, _ := testutils.InitializePrivateKeysAndAccounts(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a task to share distribution and make sure it succeeds
	state := state.NewDkgState(acct)
	task := dkg.NewShareDistributionTask(state, state.PhaseStart, state.PhaseStart+state.PhaseLength)
	log := logger.WithField("TaskID", "foo")

	err := task.Initialize(ctx, log, eth)
	assert.NotNil(t, err)
}

// We test to ensure that everything behaves correctly.
// We submit invalid state information (again).
func TestShareDistribution_Group_3_Bad7(t *testing.T) {
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
	state := state.NewDkgState(acct)
	log := logging.GetLogger("test").WithField("Validator", acct.Address.String())
	task := dkg.NewShareDistributionTask(state, state.PhaseStart, state.PhaseStart+state.PhaseLength)

	err := task.Initialize(ctx, log, eth)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestShareDistribution_Group_3_ShouldRetryTrue(t *testing.T) {
	n := 5
	suite := dkgTestUtils.StartFromRegistrationOpenPhase(t, n, 0, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()

	// Do Share Distribution task
	//shareDistributionTasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		task := suite.ShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		shouldRetry := task.ShouldRetry(ctx, logger, suite.Eth)
		assert.True(t, shouldRetry)
	}
}

func TestShareDistribution_Group_3_ShouldRetryFalse(t *testing.T) {
	n := 5
	suite := dkgTestUtils.StartFromRegistrationOpenPhase(t, n, 0, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()

	// Do Share Distribution task
	for idx := 0; idx < n; idx++ {
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		task := suite.ShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		err = task.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, task.Success)

		shouldRetry := task.ShouldRetry(ctx, logger, suite.Eth)
		assert.False(t, shouldRetry)
	}
}
