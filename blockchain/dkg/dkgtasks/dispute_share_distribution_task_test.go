package dkgtasks_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgevents"
	"github.com/MadBase/MadNet/logging"
	"github.com/stretchr/testify/assert"
)

// We test to ensure that everything behaves correctly.
func TestShareDisputeGoodAllValid(t *testing.T) {
	n := 5
	suite := StartFromShareDistributionPhase(t, n, 0, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	// currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	// assert.Nil(t, err)
	//logger := logging.GetLogger("test").WithField("Test", "Test1")

	// Confirm number of BadShares is zero
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		if len(state.BadShares) != 0 {
			t.Fatalf("Idx %v has incorrect number of BadShares", idx)
		}
	}

	// Do Share Dispute task
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		task := suite.disputeShareDistTasks[idx]
		err := task.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = task.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, task.Success)
	}

	// Check number of BadShares and confirm correct values
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		if len(state.BadShares) != 0 {
			t.Fatalf("Idx %v has incorrect number of BadShares", idx)
		}
	}

	// assert no bad participants on the ETHDKG contract
	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(suite.eth.GetCallOpts(ctx, accounts[0]))
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), badParticipants.Uint64())
}

// In this test, we have one validator submit invalid information.
// This causes another validator to submit a dispute against him,
// causing a stake-slashing event.
func TestShareDisputeGoodMaliciousShare(t *testing.T) {
	n := 5
	suite := StartFromRegistrationOpenPhase(t, n, 0, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()

	badShares := 1
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do Share Distribution task, with one of the validators distributing bad shares
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]

		task := suite.shareDistTasks[idx]
		err := task.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)

		if idx >= n-badShares {
			// inject bad shares
			for _, s := range state.Participants[state.Account.Address].EncryptedShares {
				s.Set(big.NewInt(0))
			}
		}

		err = task.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, task.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			err = dkgevents.UpdateStateOnSharesDistributed(
				suite.dkgStates[j],
				logger,
				state.Account.Address,
				state.Participants[state.Account.Address].EncryptedShares,
				state.Participants[state.Account.Address].Commitments,
			)
			assert.Nil(t, err)
		}
	}

	currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	nextPhaseAt := currentHeight + suite.dkgStates[0].ConfirmationLength
	advanceTo(t, suite.eth, nextPhaseAt)

	// Do Share Dispute task
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// state.Phase = objects.ShareDistribution
		disputeShareDistributionTask, _, _, _, _, _, _, _, _ := dkgevents.UpdateStateOnShareDistributionComplete(state, logger, nextPhaseAt)

		err := disputeShareDistributionTask.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = disputeShareDistributionTask.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, disputeShareDistributionTask.Success)

		// Set Dispute success to true
		// suite.dkgStates[idx].Dispute = true
	}

	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(suite.eth.GetCallOpts(ctx, accounts[0]))
	assert.Nil(t, err)
	assert.Equal(t, uint64(badShares), badParticipants.Uint64())

	callOpts := suite.eth.GetCallOpts(ctx, accounts[0])

	// assert bad participants are not validators anymore, i.e, they were fined and evicted
	for i := 0; i < badShares; i++ {
		idx := n - i - 1
		isValidator, err := suite.eth.Contracts().ValidatorPool().IsValidator(callOpts, suite.dkgStates[idx].Account.Address)
		assert.Nil(t, err)

		assert.False(t, isValidator)
	}
}

//// In this test, we cause one validator to submit an invalid accusation.
//// This causes this validator to submit an accusation which is invalid;
//// that is, the submitted information is actually correct,
//// causing a stake-slashing event against the false accusor.
//func TestShareDisputeGoodMaliciousAccusation(t *testing.T) {
//	n := 4
//	dkgStates, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(n)
//
//	// We have a bad index; this validator will submit a malicious accusation
//	// against the good index.
//	badIdx := n - 2
//	// We have a good index
//	goodIdx := n - 1
//	if (badIdx == goodIdx) || badIdx < 0 || badIdx >= n || goodIdx < 0 || goodIdx >= n {
//		t.Fatal("invalid good and bad indices")
//	}
//
//	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
//	assert.NotNil(t, eth)
//	defer eth.Close()
//
//	ctx := context.Background()
//
//	accounts := eth.GetKnownAccounts()
//
//	owner := accounts[0]
//	err := eth.UnlockAccount(owner)
//	assert.Nil(t, err)
//
//	// Start EthDKG
//	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
//	assert.Nil(t, err)
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
//	// Shorten ethdkg phase for testing purposes
//	txn, err := eth.Contracts().Ethdkg().UpdatePhaseLength(ownerOpts, big.NewInt(100))
//	assert.Nil(t, err)
//	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
//	assert.Nil(t, err)
//
//	txn, err = eth.Contracts().Ethdkg().InitializeState(ownerOpts)
//	assert.Nil(t, err)
//
//	eth.Commit()
//	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
//	assert.Nil(t, err)
//
//	var ethdkgStarted bool
//	for _, log := range rcpt.Logs {
//		if log.Topics[0].String() == "0x9c6f8368fe7e77e8cb9438744581403bcb3f53298e517f04c1b8475487402e97" {
//			event, err := eth.Contracts().Ethdkg().ParseRegistrationOpen(*log)
//			assert.Nil(t, err)
//
//			ethdkgStarted = true
//
//			for _, dkgState := range dkgStates {
//				dkgevents.PopulateSchedule(dkgState, event)
//			}
//		}
//	}
//	assert.True(t, ethdkgStarted, "Didn't see RegistrationOpen event")
//
//	// Do Register task
//	regTasks := make([]*dkgtasks.RegisterTask, n)
//	for idx := 0; idx < n; idx++ {
//		state := dkgStates[idx]
//		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())
//
//		regTasks[idx] = dkgtasks.NewRegisterTask(state)
//		err = regTasks[idx].Initialize(ctx, logger, eth, state)
//		assert.Nil(t, err)
//		err = regTasks[idx].DoWork(ctx, logger, eth)
//		assert.Nil(t, err)
//
//		eth.Commit()
//		assert.True(t, regTasks[idx].Success)
//
//		// Set Registration success to true
//		dkgStates[idx].Registration = true
//	}
//
//	// Advance to share distribution phase
//	advanceTo(t, eth, dkgStates[0].ShareDistributionStart)
//	dtest.GenerateEncryptedSharesAndCommitments(dkgStates)
//
//	// Do Share Distribution task
//	shareDistributionTasks := make([]*dkgtasks.ShareDistributionTask, n)
//	for idx := 0; idx < n; idx++ {
//		state := dkgStates[idx]
//		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())
//
//		shareDistributionTasks[idx] = dkgtasks.NewShareDistributionTask(state)
//		err = shareDistributionTasks[idx].Initialize(ctx, logger, eth, state)
//		assert.Nil(t, err)
//		err = shareDistributionTasks[idx].DoWork(ctx, logger, eth)
//		assert.Nil(t, err)
//
//		eth.Commit()
//		assert.True(t, shareDistributionTasks[idx].Success)
//
//		// Set ShareDistribution success to true
//		dkgStates[idx].ShareDistribution = true
//	}
//	// Ensure all participants have valid share information
//	dtest.PopulateEncryptedSharesAndCommitments(dkgStates)
//
//	// Advance to share dispute phase
//	advanceTo(t, eth, dkgStates[0].DisputeStart)
//
//	// Do Share Dispute task
//	tasks := make([]*dkgtasks.DisputeShareDistributionTask, n)
//	for idx := 0; idx < n; idx++ {
//		state := dkgStates[idx]
//		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())
//
//		tasks[idx] = dkgtasks.NewDisputeShareDistributionTask(state)
//		err := tasks[idx].Initialize(ctx, logger, eth, state)
//		assert.Nil(t, err)
//		// We ensure that there is no participant who submitted invalid shares.
//		assert.Equal(t, 0, len(state.BadShares))
//
//		if idx == badIdx {
//			// We force an error by *adding* a participant to BadShares;
//			// in practice, this would not happen.
//			state.BadShares[accounts[goodIdx].Address] = state.Participants[goodIdx]
//			// We ensure good participant was added to list
//			assert.Equal(t, 1, len(state.BadShares))
//			// We now perform the accusatiomn logic;
//			// this will result in a false accusation against good index.
//		}
//
//		err = tasks[idx].DoWork(ctx, logger, eth)
//		assert.Nil(t, err)
//
//		eth.Commit()
//		assert.True(t, tasks[idx].Success)
//
//		// Set Dispute success to true
//		dkgStates[idx].Dispute = true
//	}
//
//	// Advance to key share phase
//	advanceTo(t, eth, dkgStates[0].KeyShareSubmissionStart)
//
//	// Do one keyshare
//	dtest.GenerateKeyShares(dkgStates)
//
//	// Do Key Share Submission task
//	keyShareTask := &dkgtasks.KeyshareSubmissionTask{}
//	state := dkgStates[0]
//	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())
//
//	keyShareTask = dkgtasks.NewKeyshareSubmissionTask(state)
//	err = keyShareTask.Initialize(ctx, logger, eth, state)
//	assert.Nil(t, err)
//	err = keyShareTask.DoWork(ctx, logger, eth)
//	assert.Nil(t, err)
//
//	eth.Commit()
//	assert.True(t, keyShareTask.Success)
//
//	// Check Stake; badIdx should be fined for performing a false accusation.
//	for idx, acct := range accounts {
//		stakeBalance, err := eth.Contracts().Staking().BalanceStakeFor(callOpts, acct.Address)
//		assert.Nil(t, err)
//		assert.NotNil(t, stakeBalance)
//
//		if idx == badIdx {
//			assert.Equal(t, -1, stakeBalance.Cmp(stakeBalances[idx]))
//		} else {
//			assert.Equal(t, 0, stakeBalance.Cmp(stakeBalances[idx]))
//		}
//		t.Logf("New stake balance:%v", stakeBalance)
//	}
//}
//
//// We begin by submitting invalid information.
//// This test is meant to raise an error resulting from an invalid argument
//// for the Ethereum interface.
//func TestShareDisputeBad1(t *testing.T) {
//	n := 4
//	_, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(n)
//	logger := logging.GetLogger("ethereum")
//	logger.SetLevel(logrus.DebugLevel)
//	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
//	defer eth.Close()
//
//	acct := eth.GetKnownAccounts()[0]
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	// Create a task to share distribution and make sure it succeeds
//	state := objects.NewDkgState(acct)
//	task := dkgtasks.NewDisputeShareDistributionTask(state)
//	log := logger.WithField("TaskID", "foo")
//
//	defer func() {
//		// If we didn't get here by recovering from a panic() we failed
//		if reason := recover(); reason == nil {
//			t.Log("No panic in sight")
//			t.Fatal("Should have panicked")
//		} else {
//			t.Logf("Good panic because: %v", reason)
//		}
//	}()
//	task.Initialize(ctx, log, eth, nil)
//}
//
//// We test to ensure that everything behaves correctly.
//// This test is meant to raise an error resulting from an invalid argument
//// for the Ethereum interface;
//// this should raise an error resulting from not successfully completing
//// ShareDistribution phase.
//func TestShareDisputeBad2(t *testing.T) {
//	n := 4
//	_, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(n)
//	logger := logging.GetLogger("ethereum")
//	logger.SetLevel(logrus.DebugLevel)
//	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
//	defer eth.Close()
//
//	acct := eth.GetKnownAccounts()[0]
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	// Do bad Share Dispute task
//	// Mess up participant information
//	state := objects.NewDkgState(acct)
//	log := logging.GetLogger("test").WithField("Validator", acct.Address.String())
//	task := dkgtasks.NewShareDistributionTask(state)
//	for k := 0; k < len(state.Participants); k++ {
//		state.Participants[k] = &objects.Participant{}
//	}
//	err := task.Initialize(ctx, log, eth, state)
//	if err == nil {
//		t.Fatal("Should have raised error")
//	}
//}
