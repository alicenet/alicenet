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
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//Here we test the happy path.
func TestShareDistributionGood(t *testing.T) {
	n := 5
	suite := StartFromRegistrationOpenPhase(t, n, 0, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()

	// Do Share Distribution task
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		shareDistributionTask := suite.shareDistTasks[idx]
		dkgData := tasks.NewTaskData(state)
		err := shareDistributionTask.Initialize(ctx, logger, suite.eth, dkgData)
		assert.Nil(t, err)
		err = shareDistributionTask.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, shareDistributionTask.Success)

	}
}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (invalid elliptic curve point).
// This should result in a failed submission.
func TestShareDistributionBad1(t *testing.T) {
	n := 5
	suite := StartFromRegistrationOpenPhase(t, n, 0, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()

	// Check public keys are present and valid
	for idx, acct := range accounts {
		callOpts := suite.eth.GetCallOpts(context.Background(), acct)
		p, err := suite.eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		// check points
		publicKey := suite.dkgStates[idx].TransportPublicKey
		if (publicKey[0].Cmp(p.PublicKey[0]) != 0) || (publicKey[1].Cmp(p.PublicKey[1]) != 0) {
			t.Fatal("Invalid public key")
		}
	}

	badIdx := n - 2
	//tasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		task := suite.shareDistTasks[idx]
		dkgData := tasks.NewTaskData(state)
		task.Initialize(ctx, logger, suite.eth, dkgData)

		com := state.Participants[accounts[idx].Address].Commitments
		// if we're on the last account, we just add 1 to the first commitment (y component)
		if idx == badIdx {
			// Mess up one of the commitments (first coefficient)
			com[0][1].Add(com[0][1], big.NewInt(1))
		}

		task.DoWork(ctx, logger, suite.eth)

		suite.eth.Commit()

		// The last task should have failed
		if idx == badIdx {
			assert.False(t, task.Success)
		} else {
			assert.True(t, task.Success)
		}
	}

	// Double check to Make sure all transactions were good
	rcpts, err := suite.eth.Queue().WaitGroupTransactions(ctx, 1)
	assert.Nil(t, err)

	for _, rcpt := range rcpts {
		assert.NotNil(t, rcpt)
		assert.Equal(t, uint64(1), rcpt.Status)
	}
}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (identity element).
// This should result in a failed submission.
func TestShareDistributionBad2(t *testing.T) {
	n := 4
	suite := StartFromRegistrationOpenPhase(t, n, 0, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()

	// Check public keys are present and valid
	for idx, acct := range accounts {
		callOpts := suite.eth.GetCallOpts(context.Background(), acct)
		p, err := suite.eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		// check points
		publicKey := suite.dkgStates[idx].TransportPublicKey
		if (publicKey[0].Cmp(p.PublicKey[0]) != 0) || (publicKey[1].Cmp(p.PublicKey[1]) != 0) {
			t.Fatalf("Invalid public key: internal: %v | network: %v", publicKey, p.PublicKey)
		}
	}

	badIdx := n - 1
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		task := suite.shareDistTasks[idx]
		dkgData := tasks.NewTaskData(state)
		task.Initialize(ctx, logger, suite.eth, dkgData)

		com := state.Participants[accounts[idx].Address].Commitments
		// if we're on the last account, change the one of the commitments to 0
		if idx == badIdx {
			// Mess up one of the commitments (first coefficient)
			com[0][0].Set(common.Big0)
			com[0][1].Set(common.Big0)
		}

		task.DoWork(ctx, logger, suite.eth)

		suite.eth.Commit()

		// The last task should have failed
		if idx == badIdx {
			assert.False(t, task.Success)
		} else {
			assert.True(t, task.Success)
		}
	}

	// Double check to Make sure all transactions were good
	rcpts, err := suite.eth.Queue().WaitGroupTransactions(ctx, 1)
	assert.Nil(t, err)

	for _, rcpt := range rcpts {
		assert.NotNil(t, rcpt)
		assert.Equal(t, uint64(1), rcpt.Status)
	}
}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (incorrect commitment length)
// This should result in a failed submission.
func TestShareDistributionBad4(t *testing.T) {
	n := 5
	suite := StartFromRegistrationOpenPhase(t, n, 0, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates

	// Check public keys are present and valid
	for idx, acct := range accounts {
		callOpts := eth.GetCallOpts(context.Background(), acct)
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

		task := suite.shareDistTasks[idx]
		dkgData := tasks.NewTaskData(state)
		task.Initialize(ctx, logger, suite.eth, dkgData)

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

	// Double check to Make sure all transactions were good
	rcpts, err := eth.Queue().WaitGroupTransactions(ctx, 1)
	assert.Nil(t, err)

	for _, rcpt := range rcpts {
		assert.NotNil(t, rcpt)
		assert.Equal(t, uint64(1), rcpt.Status)
	}
}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (incorrect encrypted shares length)
// This should result in a failed submission.
func TestShareDistributionBad5(t *testing.T) {
	n := 6
	suite := StartFromRegistrationOpenPhase(t, n, 0, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	eth := suite.eth
	dkgStates := suite.dkgStates

	badShareIdx := n - 2
	//tasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		task := suite.shareDistTasks[idx]
		dkgData := tasks.NewTaskData(state)
		task.Initialize(ctx, logger, suite.eth, dkgData)

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

	// Double check to Make sure all transactions were good
	rcpts, err := eth.Queue().WaitGroupTransactions(ctx, 1)
	assert.Nil(t, err)

	for _, rcpt := range rcpts {
		assert.NotNil(t, rcpt)
		assert.Equal(t, uint64(1), rcpt.Status)
	}
}

// We begin by submitting invalid information;
// we submit nil state information
func TestShareDistributionBad6(t *testing.T) {
	n := 5
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
	task := dkgtasks.NewShareDistributionTask(state, state.PhaseStart, state.PhaseStart+state.PhaseLength)
	log := logger.WithField("TaskID", "foo")

	dkgData := tasks.NewTaskData(state)
	err := task.Initialize(ctx, log, eth, dkgData)
	assert.NotNil(t, err)
}

// We test to ensure that everything behaves correctly.
// We submit invalid state information (again).
func TestShareDistributionBad7(t *testing.T) {
	n := 4
	ecdsaPrivateKeys, _ := dtest.InitializePrivateKeysAndAccounts(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Do bad Share Dispute task
	state := objects.NewDkgState(acct)
	log := logging.GetLogger("test").WithField("Validator", acct.Address.String())
	task := dkgtasks.NewShareDistributionTask(state, state.PhaseStart, state.PhaseStart+state.PhaseLength)
	dkgData := tasks.NewTaskData(state)
	err := task.Initialize(ctx, log, eth, dkgData)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestShareDistributionShouldRetryTrue(t *testing.T) {
	n := 5
	suite := StartFromRegistrationOpenPhase(t, n, 0, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()

	// Do Share Distribution task
	//shareDistributionTasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		task := suite.shareDistTasks[idx]
		dkgData := tasks.NewTaskData(state)
		err := task.Initialize(ctx, logger, suite.eth, dkgData)
		assert.Nil(t, err)

		shouldRetry := task.ShouldRetry(ctx, logger, suite.eth)
		assert.True(t, shouldRetry)
	}
}

func TestShareDistributionShouldRetryFalse(t *testing.T) {
	n := 5
	suite := StartFromRegistrationOpenPhase(t, n, 0, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()

	// Do Share Distribution task
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		task := suite.shareDistTasks[idx]
		dkgData := tasks.NewTaskData(state)
		err := task.Initialize(ctx, logger, suite.eth, dkgData)
		assert.Nil(t, err)
		err = task.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, task.Success)

		shouldRetry := task.ShouldRetry(ctx, logger, suite.eth)
		assert.False(t, shouldRetry)
	}
}
