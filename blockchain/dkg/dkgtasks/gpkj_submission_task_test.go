package dkgtasks_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgevents"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/dkg/dtest"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// We test to ensure that everything behaves correctly.
func TestGPKjSubmissionGoodAllValid(t *testing.T) {
	n := 4
	dkgStates, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(n)

	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)
	defer eth.Close()

	ctx := context.Background()

	accounts := eth.GetKnownAccounts()

	owner := accounts[0]
	err := eth.UnlockAccount(owner)
	assert.Nil(t, err)

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Check Stake
	callOpts := eth.GetCallOpts(ctx, owner)
	for _, acct := range accounts {
		stakeBalance, err := eth.Contracts().Staking().BalanceStakeFor(callOpts, acct.Address)
		assert.Nil(t, err)

		t.Logf("stakeBalance:%v", stakeBalance.String())
	}

	// Shorten ethdkg phase for testing purposes
	txn, err := eth.Contracts().Ethdkg().UpdatePhaseLength(ownerOpts, big.NewInt(100))
	assert.Nil(t, err)
	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	txn, err = eth.Contracts().Ethdkg().InitializeState(ownerOpts)
	assert.Nil(t, err)

	eth.Commit()
	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == "0x9c6f8368fe7e77e8cb9438744581403bcb3f53298e517f04c1b8475487402e97" {
			event, err := eth.Contracts().Ethdkg().ParseRegistrationOpen(*log)
			assert.Nil(t, err)

			for _, dkgState := range dkgStates {
				dkgevents.PopulateSchedule(dkgState, event)
			}
		}
	}

	txnOpts := make([]*bind.TransactOpts, len(accounts))
	for idx, acct := range accounts {
		txnOpts[idx], err = eth.GetTransactionOpts(ctx, acct)
		assert.Nil(t, err)

		// Register for EthDKG
		publicKey := dkgStates[idx].TransportPublicKey
		txn, err := eth.Contracts().Ethdkg().Register(txnOpts[idx], publicKey)
		assert.Nil(t, err)
		eth.Queue().QueueGroupTransaction(ctx, 1, txn)
		eth.Commit()

		// Set Registration success
		dkgStates[idx].Registration = true
	}

	// Advance to share distribution phase
	advanceTo(t, eth, dkgStates[0].ShareDistributionStart)

	// Check public keys are present and valid
	for idx, acct := range accounts {
		callOpts := eth.GetCallOpts(context.Background(), acct)
		k0, err := eth.Contracts().Ethdkg().PublicKeys(callOpts, acct.Address, common.Big0)
		assert.Nil(t, err)
		k1, err := eth.Contracts().Ethdkg().PublicKeys(callOpts, acct.Address, common.Big1)
		assert.Nil(t, err)

		// check points
		publicKey := dkgStates[idx].TransportPublicKey
		if (publicKey[0].Cmp(k0) != 0) || (publicKey[1].Cmp(k1) != 0) {
			t.Fatal("Invalid public key")
		}
	}

	// Do Share Distribution task
	shareDistributionTasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		shareDistributionTasks[idx] = dkgtasks.NewShareDistributionTask(state)
		err = shareDistributionTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = shareDistributionTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, shareDistributionTasks[idx].Success)

		// Set ShareDistribution success to true
		dkgStates[idx].ShareDistribution = true
	}
	// Ensure all participants have valid share information
	dtest.PopulateEncryptedSharesAndCommitments(dkgStates)

	// Advance to share dispute phase
	advanceTo(t, eth, dkgStates[0].DisputeStart)

	// Do Dispute task
	disputeTasks := make([]*dkgtasks.DisputeTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		disputeTasks[idx] = dkgtasks.NewDisputeTask(state)
		err = disputeTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = disputeTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, disputeTasks[idx].Success)

		// Set Dispute success to true
		dkgStates[idx].Dispute = true
	}

	// Confirm number of BadShares is zero
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		if len(state.BadShares) != 0 {
			t.Fatalf("Idx %v has incorrect number of BadShares", idx)
		}
	}

	// Advance to key share distribution phase
	advanceTo(t, eth, dkgStates[0].KeyShareSubmissionStart)
	dtest.GenerateKeyShares(dkgStates)

	// Do Key Share Submission task
	keyShareSubmissionTasks := make([]*dkgtasks.KeyshareSubmissionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		keyShareSubmissionTasks[idx] = dkgtasks.NewKeyshareSubmissionTask(state)
		err = keyShareSubmissionTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = keyShareSubmissionTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, keyShareSubmissionTasks[idx].Success)

		// Set Dispute success to true
		dkgStates[idx].KeyShareSubmission = true
	}

	// Double check to Make sure all transactions were good
	rcpts, err := eth.Queue().WaitGroupTransactions(ctx, 1)
	assert.Nil(t, err)

	for _, rcpt := range rcpts {
		assert.NotNil(t, rcpt)
		assert.Equal(t, uint64(1), rcpt.Status)
	}

	eth.Commit()

	// Check key shares are present and valid
	for idx, acct := range accounts {
		callOpts := eth.GetCallOpts(context.Background(), acct)
		k0, err := eth.Contracts().Ethdkg().KeyShares(callOpts, acct.Address, common.Big0)
		assert.Nil(t, err)
		k1, err := eth.Contracts().Ethdkg().KeyShares(callOpts, acct.Address, common.Big1)
		assert.Nil(t, err)

		// check points
		keyShareG1 := dkgStates[idx].KeyShareG1s[acct.Address]
		if (keyShareG1[0].Cmp(k0) != 0) || (keyShareG1[1].Cmp(k1) != 0) {
			t.Fatal("Invalid key share")
		}
	}

	// Advance to mpk submission phase
	advanceTo(t, eth, dkgStates[0].MPKSubmissionStart)
	dtest.GenerateMasterPublicKey(dkgStates)

	// Do MPK Submission task
	mpkSubmissionTasks := make([]*dkgtasks.MPKSubmissionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		mpkSubmissionTasks[idx] = dkgtasks.NewMPKSubmissionTask(state)
		err = mpkSubmissionTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = mpkSubmissionTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, mpkSubmissionTasks[idx].Success)

		// Set Dispute success to true
		dkgStates[idx].MPKSubmission = true
	}

	// Double check to Make sure all transactions were good
	rcpts, err = eth.Queue().WaitGroupTransactions(ctx, 1)
	assert.Nil(t, err)

	for _, rcpt := range rcpts {
		assert.NotNil(t, rcpt)
		assert.Equal(t, uint64(1), rcpt.Status)
	}

	eth.Commit()

	// Validate MPK
	for idx, acct := range accounts {
		callOpts := eth.GetCallOpts(context.Background(), acct)
		k0, err := eth.Contracts().Ethdkg().MasterPublicKey(callOpts, common.Big0)
		assert.Nil(t, err)
		k1, err := eth.Contracts().Ethdkg().MasterPublicKey(callOpts, common.Big1)
		assert.Nil(t, err)
		k2, err := eth.Contracts().Ethdkg().MasterPublicKey(callOpts, common.Big2)
		assert.Nil(t, err)
		k3, err := eth.Contracts().Ethdkg().MasterPublicKey(callOpts, common.Big3)
		assert.Nil(t, err)

		// check mpk
		mpk := dkgStates[idx].MasterPublicKey
		if (mpk[0].Cmp(k0) != 0) || (mpk[1].Cmp(k1) != 0) || (mpk[2].Cmp(k2) != 0) || (mpk[3].Cmp(k3) != 0) {
			t.Fatal("Invalid master public key")
		}
	}

	// Advance to gpkj submission phase
	advanceTo(t, eth, dkgStates[0].GPKJSubmissionStart)
	dtest.GenerateGPKJ(dkgStates)

	// Do GPKj Submission task
	tasks := make([]*dkgtasks.GPKSubmissionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		adminHandler := new(adminHandlerMock)
		tasks[idx] = dkgtasks.NewGPKSubmissionTask(state, adminHandler)
		err = tasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, tasks[idx].Success)

		// Set Dispute success to true
		dkgStates[idx].GPKJSubmission = true
	}

	// Double check to Make sure all transactions were good
	rcpts, err = eth.Queue().WaitGroupTransactions(ctx, 1)
	assert.Nil(t, err)

	for _, rcpt := range rcpts {
		assert.NotNil(t, rcpt)
		assert.Equal(t, uint64(1), rcpt.Status)
	}

	eth.Commit()

	// Check gpkjs and signatures are present and valid
	for idx, acct := range accounts {
		callOpts := eth.GetCallOpts(context.Background(), acct)
		s0, err := eth.Contracts().Ethdkg().InitialSignatures(callOpts, acct.Address, common.Big0)
		assert.Nil(t, err)
		s1, err := eth.Contracts().Ethdkg().InitialSignatures(callOpts, acct.Address, common.Big1)
		assert.Nil(t, err)

		// check signature
		signature := dkgStates[idx].GroupSignature
		if (signature[0].Cmp(s0) != 0) || (signature[1].Cmp(s1) != 0) {
			t.Fatal("Invalid signature")
		}

		k0, err := eth.Contracts().Ethdkg().GpkjSubmissions(callOpts, acct.Address, common.Big0)
		assert.Nil(t, err)
		k1, err := eth.Contracts().Ethdkg().GpkjSubmissions(callOpts, acct.Address, common.Big1)
		assert.Nil(t, err)
		k2, err := eth.Contracts().Ethdkg().GpkjSubmissions(callOpts, acct.Address, common.Big2)
		assert.Nil(t, err)
		k3, err := eth.Contracts().Ethdkg().GpkjSubmissions(callOpts, acct.Address, common.Big3)
		assert.Nil(t, err)

		// check gpkj
		gpkj := dkgStates[idx].GroupPublicKey
		if (gpkj[0].Cmp(k0) != 0) || (gpkj[1].Cmp(k1) != 0) || (gpkj[2].Cmp(k2) != 0) || (gpkj[3].Cmp(k3) != 0) {
			t.Fatal("Invalid gpkj")
		}
	}
}

// We begin by submitting invalid information.
// Here, we submit nil for the state interface;
// this should raise an error.
func TestGPKjSubmissionBad1(t *testing.T) {
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
	adminHandler := new(adminHandlerMock)
	task := dkgtasks.NewGPKSubmissionTask(state, adminHandler)
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
// Here, we should raise an error because we did not successfully complete
// the key share submission phase.
func TestGPKjSubmissionBad2(t *testing.T) {
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
	log := logger.WithField("TaskID", "foo")
	adminHandler := new(adminHandlerMock)
	task := dkgtasks.NewGPKSubmissionTask(state, adminHandler)
	err := task.Initialize(ctx, log, eth, state)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// Here we test for an invalid gpkj submission.
// One or more validators should not submit gpkj information.
// After the phase is completed, an EthDKG restart should be required.
func TestGPKjSubmissionBad3(t *testing.T) {
	// Perform correct registration setup.

	// Perform correct share submission

	// Correctly submit the mpk

	// After correctly submitting the mpk,
	// one or more validators should fail to submit gpkj information.
	// This should result in each validator receiving a minor fine
	// and the EthDKG protocol will need to be restarted.
	n := 4
	dkgStates, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(n)

	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)
	defer eth.Close()

	ctx := context.Background()

	accounts := eth.GetKnownAccounts()

	owner := accounts[0]
	err := eth.UnlockAccount(owner)
	assert.Nil(t, err)

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Check Stake
	callOpts := eth.GetCallOpts(ctx, owner)
	for _, acct := range accounts {
		stakeBalance, err := eth.Contracts().Staking().BalanceStakeFor(callOpts, acct.Address)
		assert.Nil(t, err)

		t.Logf("stakeBalance:%v", stakeBalance.String())
	}

	// Shorten ethdkg phase for testing purposes
	txn, err := eth.Contracts().Ethdkg().UpdatePhaseLength(ownerOpts, big.NewInt(100))
	assert.Nil(t, err)
	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	txn, err = eth.Contracts().Ethdkg().InitializeState(ownerOpts)
	assert.Nil(t, err)

	eth.Commit()
	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == "0x9c6f8368fe7e77e8cb9438744581403bcb3f53298e517f04c1b8475487402e97" {
			event, err := eth.Contracts().Ethdkg().ParseRegistrationOpen(*log)
			assert.Nil(t, err)

			for _, dkgState := range dkgStates {
				dkgevents.PopulateSchedule(dkgState, event)
			}
		}
	}

	// Do Register task
	registerTasks := make([]*dkgtasks.RegisterTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		registerTasks[idx] = dkgtasks.NewRegisterTask(state)
		err = registerTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = registerTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, registerTasks[idx].Success)

		// Set Registration success to true
		dkgStates[idx].Registration = true
	}

	// Advance to share distribution phase
	advanceTo(t, eth, dkgStates[0].ShareDistributionStart)

	// Check Stake
	stakeBalances := make([]*big.Int, n)
	for idx, acct := range accounts {
		stakeBalances[idx], err = eth.Contracts().Staking().BalanceStakeFor(callOpts, acct.Address)
		assert.Nil(t, err)

		t.Logf("stakeBalance:%v", stakeBalances[idx].String())
	}

	// Check public keys are present and valid
	for idx, acct := range accounts {
		callOpts := eth.GetCallOpts(context.Background(), acct)
		k0, err := eth.Contracts().Ethdkg().PublicKeys(callOpts, acct.Address, common.Big0)
		assert.Nil(t, err)
		k1, err := eth.Contracts().Ethdkg().PublicKeys(callOpts, acct.Address, common.Big1)
		assert.Nil(t, err)

		// check points
		publicKey := dkgStates[idx].TransportPublicKey
		if (publicKey[0].Cmp(k0) != 0) || (publicKey[1].Cmp(k1) != 0) {
			t.Fatal("Invalid public key")
		}
	}

	// Do Share Distribution task
	shareDistributionTasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		shareDistributionTasks[idx] = dkgtasks.NewShareDistributionTask(state)
		err = shareDistributionTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = shareDistributionTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, shareDistributionTasks[idx].Success)

		// Set ShareDistribution success to true
		dkgStates[idx].ShareDistribution = true
	}
	// Ensure all participants have valid share information
	dtest.PopulateEncryptedSharesAndCommitments(dkgStates)

	// Advance to share dispute phase
	advanceTo(t, eth, dkgStates[0].DisputeStart)

	// Do Dispute task
	disputeTasks := make([]*dkgtasks.DisputeTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		disputeTasks[idx] = dkgtasks.NewDisputeTask(state)
		err = disputeTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = disputeTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, disputeTasks[idx].Success)

		// Set Dispute success to true
		dkgStates[idx].Dispute = true
	}

	// Confirm number of BadShares is zero
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		if len(state.BadShares) != 0 {
			t.Fatalf("Idx %v has incorrect number of BadShares", idx)
		}
	}
	// Advance to key share distribution phase
	advanceTo(t, eth, dkgStates[0].KeyShareSubmissionStart)
	dtest.GenerateKeyShares(dkgStates)

	// Do Key Share Submission task
	keyShareTasks := make([]*dkgtasks.KeyshareSubmissionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		keyShareTasks[idx] = dkgtasks.NewKeyshareSubmissionTask(state)
		err = keyShareTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = keyShareTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, keyShareTasks[idx].Success)

		// Set Dispute success to true
		dkgStates[idx].KeyShareSubmission = true
	}

	// Double check to Make sure all transactions were good
	rcpts, err := eth.Queue().WaitGroupTransactions(ctx, 1)
	assert.Nil(t, err)

	for _, rcpt := range rcpts {
		assert.NotNil(t, rcpt)
		assert.Equal(t, uint64(1), rcpt.Status)
	}

	eth.Commit()

	// Check key shares are present and valid
	for idx, acct := range accounts {
		callOpts := eth.GetCallOpts(context.Background(), acct)
		k0, err := eth.Contracts().Ethdkg().KeyShares(callOpts, acct.Address, common.Big0)
		assert.Nil(t, err)
		k1, err := eth.Contracts().Ethdkg().KeyShares(callOpts, acct.Address, common.Big1)
		assert.Nil(t, err)

		// check points
		keyShareG1 := dkgStates[idx].KeyShareG1s[acct.Address]
		if (keyShareG1[0].Cmp(k0) != 0) || (keyShareG1[1].Cmp(k1) != 0) {
			t.Fatal("Invalid key share")
		}
	}

	// Advance to mpk submission phase
	advanceTo(t, eth, dkgStates[0].MPKSubmissionStart)
	dtest.GenerateMasterPublicKey(dkgStates)

	// Do MPK Submission task
	mpkTasks := make([]*dkgtasks.MPKSubmissionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		mpkTasks[idx] = dkgtasks.NewMPKSubmissionTask(state)
		err = mpkTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = mpkTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, mpkTasks[idx].Success)

		// Set Dispute success to true
		dkgStates[idx].MPKSubmission = true
	}

	// Double check to Make sure all transactions were good
	rcpts, err = eth.Queue().WaitGroupTransactions(ctx, 1)
	assert.Nil(t, err)

	for _, rcpt := range rcpts {
		assert.NotNil(t, rcpt)
		assert.Equal(t, uint64(1), rcpt.Status)
	}

	eth.Commit()

	// Validate MPK
	for idx, acct := range accounts {
		callOpts := eth.GetCallOpts(context.Background(), acct)
		k0, err := eth.Contracts().Ethdkg().MasterPublicKey(callOpts, common.Big0)
		assert.Nil(t, err)
		k1, err := eth.Contracts().Ethdkg().MasterPublicKey(callOpts, common.Big1)
		assert.Nil(t, err)
		k2, err := eth.Contracts().Ethdkg().MasterPublicKey(callOpts, common.Big2)
		assert.Nil(t, err)
		k3, err := eth.Contracts().Ethdkg().MasterPublicKey(callOpts, common.Big3)
		assert.Nil(t, err)

		// check mpk
		mpk := dkgStates[idx].MasterPublicKey
		if (mpk[0].Cmp(k0) != 0) || (mpk[1].Cmp(k1) != 0) || (mpk[2].Cmp(k2) != 0) || (mpk[3].Cmp(k3) != 0) {
			t.Fatal("Invalid master public key")
		}
	}

	// Advance to gpkj submission phase
	advanceTo(t, eth, dkgStates[0].GPKJSubmissionStart)
	dtest.GenerateGPKJ(dkgStates)

	// Do GPKj Submission task
	noShareIdx := n - 2
	tasks := make([]*dkgtasks.GPKSubmissionTask, n)
	for idx := 0; idx < n; idx++ {
		if idx == noShareIdx {
			continue
		}
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		adminHandler := new(adminHandlerMock)
		tasks[idx] = dkgtasks.NewGPKSubmissionTask(state, adminHandler)
		err = tasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, tasks[idx].Success)

		// Set Dispute success to true
		dkgStates[idx].GPKJSubmission = true
	}

	eth.Commit()

	// Check gpkjs and signatures are present and valid
	for idx, acct := range accounts {
		if idx == noShareIdx {
			continue
		}
		callOpts := eth.GetCallOpts(context.Background(), acct)
		s0, err := eth.Contracts().Ethdkg().InitialSignatures(callOpts, acct.Address, common.Big0)
		assert.Nil(t, err)
		s1, err := eth.Contracts().Ethdkg().InitialSignatures(callOpts, acct.Address, common.Big1)
		assert.Nil(t, err)

		// check signature
		signature := dkgStates[idx].GroupSignature
		if (signature[0].Cmp(s0) != 0) || (signature[1].Cmp(s1) != 0) {
			t.Fatal("Invalid signature")
		}

		k0, err := eth.Contracts().Ethdkg().GpkjSubmissions(callOpts, acct.Address, common.Big0)
		assert.Nil(t, err)
		k1, err := eth.Contracts().Ethdkg().GpkjSubmissions(callOpts, acct.Address, common.Big1)
		assert.Nil(t, err)
		k2, err := eth.Contracts().Ethdkg().GpkjSubmissions(callOpts, acct.Address, common.Big2)
		assert.Nil(t, err)
		k3, err := eth.Contracts().Ethdkg().GpkjSubmissions(callOpts, acct.Address, common.Big3)
		assert.Nil(t, err)

		// check gpkj
		gpkj := dkgStates[idx].GroupPublicKey
		if (gpkj[0].Cmp(k0) != 0) || (gpkj[1].Cmp(k1) != 0) || (gpkj[2].Cmp(k2) != 0) || (gpkj[3].Cmp(k3) != 0) {
			t.Fatal("Invalid gpkj")
		}
	}

	// Advance to gpkj dispute phase
	advanceTo(t, eth, dkgStates[0].GPKJGroupAccusationStart)

	// Do GPKjDispute task
	gpkjDisputeTasks := make([]*dkgtasks.GPKJDisputeTask, n)
	for idx := 0; idx < n; idx++ {
		if idx == noShareIdx {
			continue
		}
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		gpkjDisputeTasks[idx] = dkgtasks.NewGPKJDisputeTask(state)
		logger.Errorf("Idx: %v\n", idx)
		err = gpkjDisputeTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = gpkjDisputeTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, gpkjDisputeTasks[idx].Success)

		// Set Dispute success to true
		dkgStates[idx].GPKJGroupAccusation = true
	}

	// Double check to Make sure all transactions were good
	rcpts, err = eth.Queue().WaitGroupTransactions(ctx, 1)
	assert.Nil(t, err)

	for _, rcpt := range rcpts {
		assert.NotNil(t, rcpt)
		assert.Equal(t, uint64(1), rcpt.Status)
	}

	// Advance to completion phase
	advanceTo(t, eth, dkgStates[0].CompleteStart)

	// Do Completion task
	completionTask := &dkgtasks.CompletionTask{}
	state := dkgStates[0]
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	completionTask = dkgtasks.NewCompletionTask(state)
	err = completionTask.Initialize(ctx, logger, eth, state)
	assert.Nil(t, err)
	err = completionTask.DoWork(ctx, logger, eth)
	assert.Nil(t, err)
	eth.Commit()

	// Check Stake
	for idx, acct := range accounts {
		stakeBalance, err := eth.Contracts().Staking().BalanceStakeFor(callOpts, acct.Address)
		assert.Nil(t, err)
		assert.NotNil(t, stakeBalance)

		if idx == noShareIdx {
			// Stake value should be less as the result of a fine.
			assert.Equal(t, -1, stakeBalance.Cmp(stakeBalances[idx]))
		} else {
			// Everyone else should be the same.
			assert.Equal(t, 0, stakeBalance.Cmp(stakeBalances[idx]))
		}
		t.Logf("New stake balance:%v", stakeBalance)
	}
}

// Here we test for an invalid gpkj submission.
// One or more validators should submit invalid gpkj information;
// that is, the gpkj public key and signature should not verify.
// This should result in no submission.
func TestGPKjSubmissionBad4(t *testing.T) {
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
	dkgStates, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(n)

	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)
	defer eth.Close()

	ctx := context.Background()

	accounts := eth.GetKnownAccounts()

	owner := accounts[0]
	err := eth.UnlockAccount(owner)
	assert.Nil(t, err)

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Check Stake
	callOpts := eth.GetCallOpts(ctx, owner)
	for _, acct := range accounts {
		stakeBalance, err := eth.Contracts().Staking().BalanceStakeFor(callOpts, acct.Address)
		assert.Nil(t, err)

		t.Logf("stakeBalance:%v", stakeBalance.String())
	}

	// Shorten ethdkg phase for testing purposes
	txn, err := eth.Contracts().Ethdkg().UpdatePhaseLength(ownerOpts, big.NewInt(100))
	assert.Nil(t, err)
	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	txn, err = eth.Contracts().Ethdkg().InitializeState(ownerOpts)
	assert.Nil(t, err)

	eth.Commit()
	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == "0x9c6f8368fe7e77e8cb9438744581403bcb3f53298e517f04c1b8475487402e97" {
			event, err := eth.Contracts().Ethdkg().ParseRegistrationOpen(*log)
			assert.Nil(t, err)

			for _, dkgState := range dkgStates {
				dkgevents.PopulateSchedule(dkgState, event)
			}
		}
	}

	txnOpts := make([]*bind.TransactOpts, len(accounts))
	for idx, acct := range accounts {
		txnOpts[idx], err = eth.GetTransactionOpts(ctx, acct)
		assert.Nil(t, err)

		// Register for EthDKG
		publicKey := dkgStates[idx].TransportPublicKey
		txn, err := eth.Contracts().Ethdkg().Register(txnOpts[idx], publicKey)
		assert.Nil(t, err)
		eth.Queue().QueueGroupTransaction(ctx, 1, txn)
		eth.Commit()

		// Set Registration success
		dkgStates[idx].Registration = true
	}

	// Advance to share distribution phase
	advanceTo(t, eth, dkgStates[0].ShareDistributionStart)

	// Check public keys are present and valid
	for idx, acct := range accounts {
		callOpts := eth.GetCallOpts(context.Background(), acct)
		k0, err := eth.Contracts().Ethdkg().PublicKeys(callOpts, acct.Address, common.Big0)
		assert.Nil(t, err)
		k1, err := eth.Contracts().Ethdkg().PublicKeys(callOpts, acct.Address, common.Big1)
		assert.Nil(t, err)

		// check points
		publicKey := dkgStates[idx].TransportPublicKey
		if (publicKey[0].Cmp(k0) != 0) || (publicKey[1].Cmp(k1) != 0) {
			t.Fatal("Invalid public key")
		}
	}

	// Do Share Distribution task
	shareDistributionTasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		shareDistributionTasks[idx] = dkgtasks.NewShareDistributionTask(state)
		err = shareDistributionTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = shareDistributionTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, shareDistributionTasks[idx].Success)

		// Set ShareDistribution success to true
		dkgStates[idx].ShareDistribution = true
	}
	// Ensure all participants have valid share information
	dtest.PopulateEncryptedSharesAndCommitments(dkgStates)

	// Advance to share dispute phase
	advanceTo(t, eth, dkgStates[0].DisputeStart)

	// Do Dispute task
	disputeTasks := make([]*dkgtasks.DisputeTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		disputeTasks[idx] = dkgtasks.NewDisputeTask(state)
		err = disputeTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = disputeTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, disputeTasks[idx].Success)

		// Set Dispute success to true
		dkgStates[idx].Dispute = true
	}

	// Confirm number of BadShares is zero
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		if len(state.BadShares) != 0 {
			t.Fatalf("Idx %v has incorrect number of BadShares", idx)
		}
	}

	// Advance to key share distribution phase
	advanceTo(t, eth, dkgStates[0].KeyShareSubmissionStart)
	dtest.GenerateKeyShares(dkgStates)

	// Do Key Share Submission task
	keyShareSubmissionTasks := make([]*dkgtasks.KeyshareSubmissionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		keyShareSubmissionTasks[idx] = dkgtasks.NewKeyshareSubmissionTask(state)
		err = keyShareSubmissionTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = keyShareSubmissionTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, keyShareSubmissionTasks[idx].Success)

		// Set Dispute success to true
		dkgStates[idx].KeyShareSubmission = true
	}

	// Double check to Make sure all transactions were good
	rcpts, err := eth.Queue().WaitGroupTransactions(ctx, 1)
	assert.Nil(t, err)

	for _, rcpt := range rcpts {
		assert.NotNil(t, rcpt)
		assert.Equal(t, uint64(1), rcpt.Status)
	}

	eth.Commit()

	// Check key shares are present and valid
	for idx, acct := range accounts {
		callOpts := eth.GetCallOpts(context.Background(), acct)
		k0, err := eth.Contracts().Ethdkg().KeyShares(callOpts, acct.Address, common.Big0)
		assert.Nil(t, err)
		k1, err := eth.Contracts().Ethdkg().KeyShares(callOpts, acct.Address, common.Big1)
		assert.Nil(t, err)

		// check points
		keyShareG1 := dkgStates[idx].KeyShareG1s[acct.Address]
		if (keyShareG1[0].Cmp(k0) != 0) || (keyShareG1[1].Cmp(k1) != 0) {
			t.Fatal("Invalid key share")
		}
	}

	// Advance to mpk submission phase
	advanceTo(t, eth, dkgStates[0].MPKSubmissionStart)
	dtest.GenerateMasterPublicKey(dkgStates)

	// Do MPK Submission task
	mpkSubmissionTasks := make([]*dkgtasks.MPKSubmissionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		mpkSubmissionTasks[idx] = dkgtasks.NewMPKSubmissionTask(state)
		err = mpkSubmissionTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = mpkSubmissionTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, mpkSubmissionTasks[idx].Success)

		// Set Dispute success to true
		dkgStates[idx].MPKSubmission = true
	}

	// Double check to Make sure all transactions were good
	rcpts, err = eth.Queue().WaitGroupTransactions(ctx, 1)
	assert.Nil(t, err)

	for _, rcpt := range rcpts {
		assert.NotNil(t, rcpt)
		assert.Equal(t, uint64(1), rcpt.Status)
	}

	eth.Commit()

	// Validate MPK
	for idx, acct := range accounts {
		callOpts := eth.GetCallOpts(context.Background(), acct)
		k0, err := eth.Contracts().Ethdkg().MasterPublicKey(callOpts, common.Big0)
		assert.Nil(t, err)
		k1, err := eth.Contracts().Ethdkg().MasterPublicKey(callOpts, common.Big1)
		assert.Nil(t, err)
		k2, err := eth.Contracts().Ethdkg().MasterPublicKey(callOpts, common.Big2)
		assert.Nil(t, err)
		k3, err := eth.Contracts().Ethdkg().MasterPublicKey(callOpts, common.Big3)
		assert.Nil(t, err)

		// check mpk
		mpk := dkgStates[idx].MasterPublicKey
		if (mpk[0].Cmp(k0) != 0) || (mpk[1].Cmp(k1) != 0) || (mpk[2].Cmp(k2) != 0) || (mpk[3].Cmp(k3) != 0) {
			t.Fatal("Invalid master public key")
		}
	}

	// Advance to gpkj submission phase
	advanceTo(t, eth, dkgStates[0].GPKJSubmissionStart)
	dtest.GenerateGPKJ(dkgStates)

	// Do GPKj Submission task; this will fail because invalid submission;
	// it does not pass the PairingCheck.
	state := dkgStates[0]
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	adminHandler := new(adminHandlerMock)
	task := dkgtasks.NewGPKSubmissionTask(state, adminHandler)
	err = task.Initialize(ctx, logger, eth, state)
	assert.Nil(t, err)
	// Mess up signature; this will cause DoWork to fail
	state.GroupSignature = [2]*big.Int{big.NewInt(1), big.NewInt(2)}
	err = task.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
}
