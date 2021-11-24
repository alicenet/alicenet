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
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// We test to ensure that everything behaves correctly.
func TestMPKSubmissionGoodAllValid(t *testing.T) {
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
	tasks := make([]*dkgtasks.MPKSubmissionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		tasks[idx] = dkgtasks.NewMPKSubmissionTask(state)
		err = tasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, tasks[idx].Success)

		// Set Dispute success to true
		dkgStates[idx].MPKSubmission = true
	}

	// Double check to make sure all transactions were good
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
}

// Here we test for invalid mpk submission.
// In this test, *no* validator should submit an mpk.
// After ending the MPK submission phase, validators should attempt
// to submit the mpk; this should raise an error.
func TestMPKSubmissionBad1(t *testing.T) {
	// Perform correct registration setup.

	// Perform correct share submission

	// After shares have been submitted, quickly proceed through the mpk
	// submission phase.

	// After the completion of the mpk submission phase, cause a validator
	// to attempt to submit the mpk.
	// This should result in an error.
	// EthDKG restart should be required.
	n := 6
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

	// Advance to gpkj submission phase; note we did *not* submit MPK
	advanceTo(t, eth, dkgStates[0].GPKJSubmissionStart)
	dtest.GenerateGPKJ(dkgStates)

	// Do MPK Submission task
	task := &dkgtasks.MPKSubmissionTask{}
	state := dkgStates[0]
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	task = dkgtasks.NewMPKSubmissionTask(state)
	err = task.Initialize(ctx, logger, eth, state)
	assert.Nil(t, err)
	err = task.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
}

// Here we test for invalid mpk submission.
// In this test, one or more validators should submit an invalid mpk
// before a valid mpk has been submitted; this should raise an error.
func TestMPKSubmissionBad2(t *testing.T) {
	// Perform correct registration setup.

	// Perform correct share submission

	// After shares have been submitted and during the mpk submission phase,
	// a (dishonest) validator should attempt to submit an invalid mpk
	// before the valid mpk has been submitted.
	// This should result in an error but no fine.
	n := 5
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
	task := &dkgtasks.MPKSubmissionTask{}
	state := dkgStates[0]
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	task = dkgtasks.NewMPKSubmissionTask(state)
	err = task.Initialize(ctx, logger, eth, state)
	assert.Nil(t, err)
	// Enter invalid mpk
	state.MasterPublicKey = [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
	err = task.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
}

// We force an error.
// This is caused by submitting invalid state information (state is nil).
func TestMPKSubmissionBad3(t *testing.T) {
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
	task := dkgtasks.NewMPKSubmissionTask(state)
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

// We force an error.
// This is caused by submitting invalid state information by not successfully
// completing KeyShareSubmission phase.
func TestMPKSubmissionBad4(t *testing.T) {
	n := 4
	_, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Do MPK Submission task
	state := objects.NewDkgState(acct)
	log := logger.WithField("TaskID", "foo")
	task := dkgtasks.NewMPKSubmissionTask(state)
	err := task.Initialize(ctx, log, eth, state)
	assert.NotNil(t, err)
}

// Here we test what happens when a validator fails to submit
// key share information.
// We raise an error when attempting to generate the mpk.
// This should result in a fine and a required restart.
func TestMPKSubmissionBad5(t *testing.T) {
	n := 5
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
	noShareIdx := n - 1
	keyShareSubmissionTasks := make([]*dkgtasks.KeyshareSubmissionTask, n)
	for idx := 0; idx < n; idx++ {
		if idx == noShareIdx {
			continue
		}
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
	// Delete key shares for noShareIdx
	for idx := 0; idx < n; idx++ {
		delete(dkgStates[idx].KeyShareG1s, dkgStates[noShareIdx].Account.Address)
		delete(dkgStates[idx].KeyShareG1CorrectnessProofs, dkgStates[noShareIdx].Account.Address)
		delete(dkgStates[idx].KeyShareG2s, dkgStates[noShareIdx].Account.Address)
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
		if idx == noShareIdx {
			continue
		}
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
	task := &dkgtasks.MPKSubmissionTask{}
	state := dkgStates[0]
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	task = dkgtasks.NewMPKSubmissionTask(state)
	err = task.Initialize(ctx, logger, eth, state)
	assert.Nil(t, err)
	err = task.DoWork(ctx, logger, eth)
	assert.Nil(t, err)

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
