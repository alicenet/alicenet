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

// Here we test the happy path.
func TestShareDistributionGood(t *testing.T) {
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
	tasks := make([]*dkgtasks.RegisterTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		tasks[idx] = dkgtasks.NewRegisterTask(state)
		err = tasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, tasks[idx].Success)

		// Set Registration success to true
		dkgStates[idx].Registration = true
	}

	// Advance to share distribution phase; afterward, we confirm all are valid
	advanceTo(t, eth, dkgStates[0].ShareDistributionStart)

	// Check public keys are present and valid; last will be invalid
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
		if !dkgStates[idx].Registration {
			t.Fatal("Registration failed")
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
}

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (invalid elliptic curve point).
// This should result in a failed submission.
func TestShareDistributionBad1(t *testing.T) {
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
	dtest.GenerateEncryptedSharesAndCommitments(dkgStates)

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

	badIdx := n - 2
	tasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		com := state.Commitments[accounts[idx].Address]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// if we're on the last account, we just add 1 to the first commitment (y component)
		if idx == badIdx {
			// Mess up one of the commitments (first coefficient)
			com[0][1].Add(com[0][1], big.NewInt(1))
		}

		tasks[idx] = dkgtasks.NewShareDistributionTask(state)
		tasks[idx].DoWork(ctx, logger, eth)

		eth.Commit()

		// The last task should have failed
		if idx == badIdx {
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
// One validator attempts to submit invalid commitments (identity element).
// This should result in a failed submission.
func TestShareDistributionBad2(t *testing.T) {
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
	}

	// Advance to share distribution phase
	advanceTo(t, eth, dkgStates[0].ShareDistributionStart)
	dtest.GenerateEncryptedSharesAndCommitments(dkgStates)

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

	badIdx := n - 1
	tasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		com := state.Commitments[accounts[idx].Address]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// if we're on the last account, change the one of the commitments to 0
		if idx == badIdx {
			// Mess up one of the commitments (first coefficient)
			com[0][0].Set(common.Big0)
			com[0][1].Set(common.Big0)
		}

		tasks[idx] = dkgtasks.NewShareDistributionTask(state)
		tasks[idx].DoWork(ctx, logger, eth)

		eth.Commit()

		// The last task should have failed
		if idx == badIdx {
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

// Here we test for no share distribution.
// That is, in this test someone distributes *no* share distribution
// even though there was a valid registration.
// In this case, an EthDKG restart should occur.
func TestShareDistributionBad3(t *testing.T) {
	// Perform correct registration setup.

	// Before submitting share information, ensure one participant does not
	// submit a share.
	// This could occur by killing the process.
	// At this point, continue with the protocol.

	// After receiving no share, an EthDKG restart should be required.
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

	// Advance to share distribution phase; afterward, we confirm all are valid
	advanceTo(t, eth, dkgStates[0].ShareDistributionStart)

	// Check Stake
	stakeBalances := make([]*big.Int, n)
	callOpts := eth.GetCallOpts(ctx, owner)
	for idx, acct := range accounts {
		stakeBalances[idx], err = eth.Contracts().Staking().BalanceStakeFor(callOpts, acct.Address)
		assert.Nil(t, err)

		t.Logf("stakeBalance:%v", stakeBalances[idx].String())
	}

	// Check public keys are present and valid; last will be invalid
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
		if !dkgStates[idx].Registration {
			t.Fatal("Registration failed")
		}
	}

	noShareIdx := n - 1
	// Do Share Distribution task
	shareDistributionTasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		if idx == noShareIdx {
			// Skip the last participant; he does not submit shares
			continue
		}
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
	// Now go through and delete the value for bad index...
	for idx := 0; idx < n; idx++ {
		delete(dkgStates[idx].Commitments, dkgStates[noShareIdx].Account.Address)
		delete(dkgStates[idx].EncryptedShares, dkgStates[noShareIdx].Account.Address)
	}

	// Advance to share dispute phase
	advanceTo(t, eth, dkgStates[0].DisputeStart)

	// Do Share Dispute task
	tasks := make([]*dkgtasks.DisputeTask, n)
	for idx := 0; idx < n; idx++ {
		if idx == noShareIdx {
			continue
		}
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		tasks[idx] = dkgtasks.NewDisputeTask(state)
		err := tasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, tasks[idx].Success)

		dkgStates[idx].Dispute = true
	}

	// Double check to Make sure all transactions were good
	rcpts, err := eth.Queue().WaitGroupTransactions(ctx, 1)
	assert.Nil(t, err)

	for _, rcpt := range rcpts {
		assert.NotNil(t, rcpt)
		assert.Equal(t, uint64(1), rcpt.Status)
	}

	// Advance to KeyShare submission phase
	advanceTo(t, eth, dkgStates[0].KeyShareSubmissionStart)

	// Do Key Share Submission task; this will "fail"
	state0 := dkgStates[0]
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	keyShareTasks0 := dkgtasks.NewKeyshareSubmissionTask(state0)
	err = keyShareTasks0.Initialize(ctx, logger, eth, state0)
	assert.Nil(t, err)
	err = keyShareTasks0.DoWork(ctx, logger, eth)
	assert.Nil(t, err)

	eth.Commit()
	assert.True(t, keyShareTasks0.Success)

	// Do Key Share Submission task; this will fail because everything
	// has reverted and we are starting a new EthDKG round.
	state1 := dkgStates[1]
	logger = logging.GetLogger("test").WithField("Validator", accounts[1].Address.String())

	keyShareTasks1 := dkgtasks.NewKeyshareSubmissionTask(state1)
	err = keyShareTasks1.Initialize(ctx, logger, eth, state1)
	assert.Nil(t, err)
	err = keyShareTasks1.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)

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

// Here we test for invalid share distribution.
// One validator attempts to submit invalid commitments (incorrect commitment length)
// This should result in a failed submission.
func TestShareDistributionBad4(t *testing.T) {
	n := 7
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
	}

	// Advance to share distribution phase
	advanceTo(t, eth, dkgStates[0].ShareDistributionStart)
	dtest.GenerateEncryptedSharesAndCommitments(dkgStates)

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

	badCommitmentIdx := n - 3
	tasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		com := state.Commitments[accounts[idx].Address]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// if we're on the last account, we just add 1 to the first commitment (y component)
		if idx == badCommitmentIdx {
			// Mess up one of the commitments (incorrect length)
			com = com[:len(com)-1]
			state.Commitments[accounts[idx].Address] = com
		}

		tasks[idx] = dkgtasks.NewShareDistributionTask(state)
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
	}

	// Advance to share distribution phase
	advanceTo(t, eth, dkgStates[0].ShareDistributionStart)
	dtest.GenerateEncryptedSharesAndCommitments(dkgStates)

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

	badShareIdx := n - 2
	tasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		encryptedShares := state.EncryptedShares[accounts[idx].Address]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		if idx == badShareIdx {
			// Mess up one of the encryptedShares (incorrect length)
			encryptedShares = encryptedShares[:len(encryptedShares)-1]
			state.EncryptedShares[accounts[idx].Address] = encryptedShares
		}

		tasks[idx] = dkgtasks.NewShareDistributionTask(state)
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
	task := dkgtasks.NewShareDistributionTask(state)
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
	task := dkgtasks.NewShareDistributionTask(state)
	err := task.Initialize(ctx, log, eth, state)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}
