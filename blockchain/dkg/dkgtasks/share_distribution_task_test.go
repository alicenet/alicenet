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
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//Here we test the happy path.
func TestShareDistributionGood(t *testing.T) {
	n := 5
	suite := StartFromRegistrationOpenPhase(t, n, 0, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	// Do Share Distribution task
	shareDistributionTasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// set phase
		state.Phase = objects.ShareDistribution
		state.PhaseStart = currentHeight + state.ConfirmationLength
		phaseEnd := state.PhaseStart + state.PhaseLength

		shareDistributionTasks[idx] = dkgtasks.NewShareDistributionTask(state, state.PhaseStart, phaseEnd)
		err := shareDistributionTasks[idx].Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = shareDistributionTasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, shareDistributionTasks[idx].Success)

	}
	// Ensure all participants have valid share information
	dtest.PopulateEncryptedSharesAndCommitments(t, suite.dkgStates)
}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (invalid elliptic curve point).
// This should result in a failed submission.
func TestShareDistributionBad1(t *testing.T) {
	n := 5
	suite := StartFromRegistrationOpenPhase(t, n, 0, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

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
	tasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// set phase
		state.Phase = objects.ShareDistribution
		state.PhaseStart = currentHeight + state.ConfirmationLength
		phaseEnd := state.PhaseStart + state.PhaseLength

		tasks[idx] = dkgtasks.NewShareDistributionTask(state, state.PhaseStart, phaseEnd)
		tasks[idx].Initialize(ctx, logger, suite.eth, state)

		com := state.Participants[accounts[idx].Address].Commitments
		// if we're on the last account, we just add 1 to the first commitment (y component)
		if idx == badIdx {
			// Mess up one of the commitments (first coefficient)
			com[0][1].Add(com[0][1], big.NewInt(1))
		}

		tasks[idx].DoWork(ctx, logger, suite.eth)

		suite.eth.Commit()

		// The last task should have failed
		if idx == badIdx {
			assert.False(t, tasks[idx].Success)
		} else {
			assert.True(t, tasks[idx].Success)
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
	n := 5
	suite := StartFromRegistrationOpenPhase(t, n, 0, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

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

	badIdx := n - 1
	tasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// set phase
		state.Phase = objects.ShareDistribution
		state.PhaseStart = currentHeight + state.ConfirmationLength
		phaseEnd := state.PhaseStart + state.PhaseLength

		tasks[idx] = dkgtasks.NewShareDistributionTask(state, state.PhaseStart, phaseEnd)
		tasks[idx].Initialize(ctx, logger, suite.eth, state)

		com := state.Participants[accounts[idx].Address].Commitments
		// if we're on the last account, change the one of the commitments to 0
		if idx == badIdx {
			// Mess up one of the commitments (first coefficient)
			com[0][0].Set(common.Big0)
			com[0][1].Set(common.Big0)
		}

		tasks[idx].DoWork(ctx, logger, suite.eth)

		suite.eth.Commit()

		// The last task should have failed
		if idx == badIdx {
			assert.False(t, tasks[idx].Success)
		} else {
			assert.True(t, tasks[idx].Success)
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
	n := 7
	suite := StartFromRegistrationOpenPhase(t, n, 0, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
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
			t.Fatal("Invalid public key")
		}
	}

	badCommitmentIdx := n - 3
	startPhase := currentHeight + dkgStates[0].ConfirmationLength
	phaseEnd := startPhase + dkgStates[0].PhaseLength
	tasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// set phase
		state.Phase = objects.ShareDistribution

		tasks[idx] = dkgtasks.NewShareDistributionTask(state, startPhase, phaseEnd)
		tasks[idx].Initialize(ctx, logger, suite.eth, state)

		// if we're on the last account, we just add 1 to the first commitment (y component)
		com := state.Participants[accounts[idx].Address].Commitments
		if idx == badCommitmentIdx {
			// Mess up one of the commitments (incorrect length)
			com = com[:len(com)-1]
			state.Participants[accounts[idx].Address].Commitments = com
		}

		tasks[idx].DoWork(ctx, logger, eth)

		eth.Commit()

		// The last task should have failed
		if idx == badCommitmentIdx {
			assert.False(t, tasks[idx].Success)
		} else {
			assert.True(t, tasks[idx].Success)
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
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	eth := suite.eth
	dkgStates := suite.dkgStates

	badShareIdx := n - 2
	phaseStart := currentHeight + dkgStates[0].ConfirmationLength
	phaseEnd := phaseStart + dkgStates[0].PhaseLength
	tasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// set phase
		state.Phase = objects.ShareDistribution

		tasks[idx] = dkgtasks.NewShareDistributionTask(state, phaseStart, phaseEnd)
		tasks[idx].Initialize(ctx, logger, suite.eth, state)

		encryptedShares := state.Participants[accounts[idx].Address].EncryptedShares
		if idx == badShareIdx {
			// Mess up one of the encryptedShares (incorrect length)
			encryptedShares = encryptedShares[:len(encryptedShares)-1]
			state.Participants[accounts[idx].Address].EncryptedShares = encryptedShares
		}

		tasks[idx].DoWork(ctx, logger, eth)

		eth.Commit()

		// The last task should have failed
		if idx == badShareIdx {
			assert.False(t, tasks[idx].Success)
		} else {
			assert.True(t, tasks[idx].Success)
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
	task := dkgtasks.NewShareDistributionTask(state, state.PhaseStart, state.PhaseStart+state.PhaseLength)
	log := logger.WithField("TaskID", "foo")

	defer func() {
		// If we didn't get here by recovering from a panic() we failed
		if reason := recover(); reason == nil {
			t.Log("No panic in sight")
			t.Fatal("Should have panicked")
		} else {
			t.Logf("Good panic because: %v", reason)
		}
	}()
	task.Initialize(ctx, log, eth, nil)
}

// We test to ensure that everything behaves correctly.
// We submit invalid state information (again).
func TestShareDistributionBad7(t *testing.T) {
	n := 4
	_, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Do bad Share Dispute task
	state := objects.NewDkgState(acct)
	log := logging.GetLogger("test").WithField("Validator", acct.Address.String())
	task := dkgtasks.NewShareDistributionTask(state, state.PhaseStart, state.PhaseStart+state.PhaseLength)
	err := task.Initialize(ctx, log, eth, state)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestShareDistributionShouldRetryTrue(t *testing.T) {
	n := 5
	suite := StartFromRegistrationOpenPhase(t, n, 0, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	// Do Share Distribution task
	shareDistributionTasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// set phase
		state.Phase = objects.ShareDistribution
		state.PhaseStart = currentHeight + state.ConfirmationLength
		phaseEnd := state.PhaseStart + state.PhaseLength

		shareDistributionTasks[idx] = dkgtasks.NewShareDistributionTask(state, state.PhaseStart, phaseEnd)
		err := shareDistributionTasks[idx].Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = shareDistributionTasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, shareDistributionTasks[idx].Success)
	}

	suite.eth.Commit()
	suite.eth.Commit()
	suite.eth.Commit()
	suite.eth.Commit()

	for idx := 0; idx < n; idx++ {
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())
		shareDistributionTasks[idx].State.Account.Address = common.Address{}
		shouldRetry := shareDistributionTasks[idx].ShouldRetry(ctx, logger, suite.eth)
		assert.True(t, shouldRetry)
	}
}

func TestShareDistributionShouldRetryFalse(t *testing.T) {
	n := 5
	suite := StartFromRegistrationOpenPhase(t, n, 0, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	// Do Share Distribution task
	shareDistributionTasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// set phase
		state.Phase = objects.ShareDistribution
		state.PhaseStart = currentHeight + state.ConfirmationLength
		phaseEnd := state.PhaseStart + state.PhaseLength

		shareDistributionTasks[idx] = dkgtasks.NewShareDistributionTask(state, state.PhaseStart, phaseEnd)
		err := shareDistributionTasks[idx].Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = shareDistributionTasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, shareDistributionTasks[idx].Success)

		shouldRetry := shareDistributionTasks[idx].ShouldRetry(ctx, logger, suite.eth)
		assert.False(t, shouldRetry)
	}
}
