package dkgtasks_test

import (
	"context"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/dkg/dtest"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

//Here we test the happy path.
func TestShareDistributionGood(t *testing.T) {
	n := 5
	suite := StartAtDistributeSharesPhase(t, n, 0, 100)
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
	dtest.PopulateEncryptedSharesAndCommitments(suite.dkgStates)
}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (invalid elliptic curve point).
// This should result in a failed submission.
func TestShareDistributionBad1(t *testing.T) {
	n := 5
	suite := StartAtDistributeSharesPhase(t, n, 0, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	dtest.GenerateEncryptedSharesAndCommitments(suite.dkgStates)

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
		com := state.Participants[accounts[idx].Address].Commitments
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// if we're on the last account, we just add 1 to the first commitment (y component)
		if idx == badIdx {
			// Mess up one of the commitments (first coefficient)
			com[0][1].Add(com[0][1], big.NewInt(1))
		}

		// set phase
		state.Phase = objects.ShareDistribution
		state.PhaseStart = currentHeight + state.ConfirmationLength
		phaseEnd := state.PhaseStart + state.PhaseLength

		tasks[idx] = dkgtasks.NewShareDistributionTask(state, state.PhaseStart, phaseEnd)
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
	suite := StartAtDistributeSharesPhase(t, n, 0, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	dtest.GenerateEncryptedSharesAndCommitments(suite.dkgStates)

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
		com := state.Participants[accounts[idx].Address].Commitments
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// if we're on the last account, change the one of the commitments to 0
		if idx == badIdx {
			// Mess up one of the commitments (first coefficient)
			com[0][0].Set(common.Big0)
			com[0][1].Set(common.Big0)
		}

		// set phase
		state.Phase = objects.ShareDistribution
		state.PhaseStart = currentHeight + state.ConfirmationLength
		phaseEnd := state.PhaseStart + state.PhaseLength

		tasks[idx] = dkgtasks.NewShareDistributionTask(state, state.PhaseStart, phaseEnd)
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

//// Here we test for no share distribution.
//// That is, in this test someone distributes *no* share distribution
//// even though there was a valid registration.
//// In this case, an EthDKG restart should occur.
//func TestShareDistributionBad3(t *testing.T) {
//	// Perform correct registration setup.
//
//	// Before submitting share information, ensure one participant does not
//	// submit a share.
//	// This could occur by killing the process.
//	// At this point, continue with the protocol.
//
//	// After receiving no share, an EthDKG restart should be required.
//	n := 4
//	suite := StartAtDistributeSharesPhase(t, n, 0, 100)
//	accounts := suite.eth.GetKnownAccounts()
//	ctx := context.Background()
//	eth := suite.eth
//	owner := accounts[0]
//	dkgStates := suite.dkgStates
//
//	// Check Stake
//	stakeBalances := make([]*big.Int, n)
//	callOpts := eth.GetCallOpts(ctx, owner)
//	for idx, acct := range accounts {
//		stakeBalances[idx], err = eth.Contracts().Staking().BalanceStakeFor(callOpts, acct.Address)
//		assert.Nil(t, err)
//
//		t.Logf("stakeBalance:%v", stakeBalances[idx].String())
//	}
//
//	// Check public keys are present and valid; last will be invalid
//	for idx, acct := range accounts {
//		callOpts := eth.GetCallOpts(context.Background(), acct)
//		p, err := eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
//		assert.Nil(t, err)
//
//		// check points
//		publicKey := dkgStates[idx].TransportPublicKey
//		if (publicKey[0].Cmp(p.PublicKey[0]) != 0) || (publicKey[1].Cmp(p.PublicKey[1]) != 0) {
//			t.Fatal("Invalid public key")
//		}
//		if dkgStates[idx].Phase != objects.RegistrationOpen {
//			t.Fatal("Registration failed")
//		}
//	}
//
//	noShareIdx := n - 1
//	// Do Share Distribution task
//	currentHeight, err := suite.eth.GetCurrentHeight(ctx)
//	assert.Nil(t, err)
//	startPhase := currentHeight + dkgStates[0].ConfirmationLength
//	phaseEnd := startPhase + dkgStates[0].PhaseLength
//
//	shareDistributionTasks := make([]*dkgtasks.ShareDistributionTask, n)
//	for idx := 0; idx < n; idx++ {
//		if idx == noShareIdx {
//			// Skip the last participant; he does not submit shares
//			continue
//		}
//		state := dkgStates[idx]
//		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())
//
//		// set phase
//		state.Phase = objects.ShareDistribution
//		state.PhaseStart = startPhase
//
//		shareDistributionTasks[idx] = dkgtasks.NewShareDistributionTask(state, startPhase, phaseEnd)
//		err = shareDistributionTasks[idx].Initialize(ctx, logger, eth, state)
//		assert.Nil(t, err)
//		err = shareDistributionTasks[idx].DoWork(ctx, logger, eth)
//		assert.Nil(t, err)
//
//		eth.Commit()
//		assert.True(t, shareDistributionTasks[idx].Success)
//	}
//	// Ensure all participants have valid share information
//	dtest.PopulateEncryptedSharesAndCommitments(dkgStates)
//	// Now go through and delete the value for bad index...
//	for idx := 0; idx < n; idx++ {
//		p := dkgStates[idx].Participants[accounts[idx].Address]
//		p.Commitments = nil
//		p.EncryptedShares = nil
//	}
//
//	// Advance to share dispute phase
//	advanceTo(t, eth, phaseEnd)
//
//	// Do Share Dispute task
//	tasks := make([]*dkgtasks.DisputeShareDistributionTask, n)
//	for idx := 0; idx < n; idx++ {
//		if idx == noShareIdx {
//			continue
//		}
//		state := dkgStates[idx]
//		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())
//
//		tasks[idx] = dkgtasks.NewDisputeShareDistributionTask(state)
//		err := tasks[idx].Initialize(ctx, logger, eth, state)
//		assert.Nil(t, err)
//		err = tasks[idx].DoWork(ctx, logger, eth)
//		assert.Nil(t, err)
//
//		eth.Commit()
//		assert.True(t, tasks[idx].Success)
//
//		dkgStates[idx].Dispute = true
//	}
//
//	// Double check to Make sure all transactions were good
//	rcpts, err := eth.Queue().WaitGroupTransactions(ctx, 1)
//	assert.Nil(t, err)
//
//	for _, rcpt := range rcpts {
//		assert.NotNil(t, rcpt)
//		assert.Equal(t, uint64(1), rcpt.Status)
//	}
//
//	// Advance to KeyShare submission phase
//	advanceTo(t, eth, dkgStates[0].KeyShareSubmissionStart)
//
//	// Do Key Share Submission task; this will "fail"
//	state0 := dkgStates[0]
//	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())
//
//	keyShareTasks0 := dkgtasks.NewKeyshareSubmissionTask(state0)
//	err = keyShareTasks0.Initialize(ctx, logger, eth, state0)
//	assert.Nil(t, err)
//	err = keyShareTasks0.DoWork(ctx, logger, eth)
//	assert.Nil(t, err)
//
//	eth.Commit()
//	assert.True(t, keyShareTasks0.Success)
//
//	// Do Key Share Submission task; this will fail because everything
//	// has reverted and we are starting a new EthDKG round.
//	state1 := dkgStates[1]
//	logger = logging.GetLogger("test").WithField("Validator", accounts[1].Address.String())
//
//	keyShareTasks1 := dkgtasks.NewKeyshareSubmissionTask(state1)
//	err = keyShareTasks1.Initialize(ctx, logger, eth, state1)
//	assert.Nil(t, err)
//	err = keyShareTasks1.DoWork(ctx, logger, eth)
//	assert.NotNil(t, err)
//
//	// Check Stake
//	for idx, acct := range accounts {
//		stakeBalance, err := eth.Contracts().Staking().BalanceStakeFor(callOpts, acct.Address)
//		assert.Nil(t, err)
//		assert.NotNil(t, stakeBalance)
//
//		if idx == noShareIdx {
//			// Stake value should be less as the result of a fine.
//			assert.Equal(t, -1, stakeBalance.Cmp(stakeBalances[idx]))
//		} else {
//			// Everyone else should be the same.
//			assert.Equal(t, 0, stakeBalance.Cmp(stakeBalances[idx]))
//		}
//		t.Logf("New stake balance:%v", stakeBalance)
//	}
//}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (incorrect commitment length)
// This should result in a failed submission.
func TestShareDistributionBad4(t *testing.T) {
	n := 7
	suite := StartAtDistributeSharesPhase(t, n, 0, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	eth := suite.eth
	dkgStates := suite.dkgStates

	dtest.GenerateEncryptedSharesAndCommitments(dkgStates)

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
		com := state.Participants[accounts[idx].Address].Commitments
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// if we're on the last account, we just add 1 to the first commitment (y component)
		if idx == badCommitmentIdx {
			// Mess up one of the commitments (incorrect length)
			com = com[:len(com)-1]
			state.Participants[accounts[idx].Address].Commitments = com
		}

		// set phase
		state.Phase = objects.ShareDistribution

		tasks[idx] = dkgtasks.NewShareDistributionTask(state, startPhase, phaseEnd)
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
	suite := StartAtDistributeSharesPhase(t, n, 0, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	eth := suite.eth
	dkgStates := suite.dkgStates
	dtest.GenerateEncryptedSharesAndCommitments(dkgStates)

	badShareIdx := n - 2
	startPhase := currentHeight + dkgStates[0].ConfirmationLength
	endPhase := startPhase + dkgStates[0].PhaseLength
	tasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		encryptedShares := state.Participants[accounts[idx].Address].EncryptedShares
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		if idx == badShareIdx {
			// Mess up one of the encryptedShares (incorrect length)
			encryptedShares = encryptedShares[:len(encryptedShares)-1]
			state.Participants[accounts[idx].Address].EncryptedShares = encryptedShares
		}

		// set phase
		state.Phase = objects.ShareDistribution

		tasks[idx] = dkgtasks.NewShareDistributionTask(state, startPhase, endPhase)
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
