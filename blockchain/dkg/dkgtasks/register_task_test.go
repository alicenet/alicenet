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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRegisterTask(t *testing.T) {
	dkgStates, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(5)

	t.Logf("dkgStates:%v", dkgStates)
	t.Logf("ecdsaPrivateKeys:%v", ecdsaPrivateKeys)

	tr := &objects.TypeRegistry{}

	tr.RegisterInstanceType(&dkgtasks.RegisterTask{})

	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 500*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := eth.Contracts()

	// Check status
	callOpts := eth.GetCallOpts(ctx, acct)
	valid, err := c.Participants().IsValidator(callOpts, acct.Address)
	assert.Nil(t, err, "Failed checking validator status")
	assert.True(t, valid)

	// Kick off EthDKG
	txnOpts, err := eth.GetTransactionOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)

	mess, err := c.Ethdkg().InitialMessage(callOpts)
	assert.Nil(t, err)
	assert.NotNil(t, mess)

	// Reuse these
	var (
		txn  *types.Transaction
		rcpt *types.Receipt
	)

	// Shorten ethdkg phase for testing purposes
	txn, err = c.Ethdkg().UpdatePhaseLength(txnOpts, big.NewInt(4))
	assert.Nil(t, err)

	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	t.Logf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Cost())

	// Kick off ethdkg
	txn, err = c.Ethdkg().InitializeState(txnOpts)
	assert.Nil(t, err)

	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)
	assert.NotNil(t, rcpt.Logs)

	t.Logf("Kicking off EthDKG used %v gas", rcpt.GasUsed)
	t.Logf("registration opens:%v", rcpt.BlockNumber)

	var openLog *types.Log
	for _, log := range rcpt.Logs {
		eventSelector := log.Topics[0].String()
		if eventSelector == "0x9c6f8368fe7e77e8cb9438744581403bcb3f53298e517f04c1b8475487402e97" {
			openLog = log
		}
	}

	// Make sure we found the open event
	assert.NotNil(t, openLog)
	openEvent, err := c.Ethdkg().ParseRegistrationOpen(*openLog)
	assert.Nil(t, err)

	// Create a task to register and make sure it succeeds
	state := objects.NewDkgState(acct)
	state.RegistrationStart = openLog.BlockNumber
	state.RegistrationEnd = openEvent.RegistrationEnds.Uint64()

	task := dkgtasks.NewRegisterTask(state)

	t.Logf("registration ends:%v", openEvent.RegistrationEnds.Uint64())

	log := logger.WithField("TaskID", "foo")

	err = task.Initialize(ctx, log, eth, state)
	assert.Nil(t, err)

	err = task.DoWork(ctx, log, eth)
	assert.Nil(t, err)
}

// We attempt valid registration. Everything should succeed.
func TestRegistrationGood(t *testing.T) {
	n := 9
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
		publicKey := [2]*big.Int{}
		publicKey[0] = dkgStates[idx].TransportPublicKey[0]
		publicKey[1] = dkgStates[idx].TransportPublicKey[1]

		txn, err := eth.Contracts().Ethdkg().Register(txnOpts[idx], publicKey)
		assert.Nil(t, err)
		eth.Queue().QueueGroupTransaction(ctx, 1, txn)
		eth.Commit()
	}

	// Do share distribution; afterward, we confirm who is valid and who is not
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
	}
}

// We attempt valid registration. Everything should succeed.
// This test calls Initialize and DoWork.
func TestRegistrationGood2(t *testing.T) {
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
}

// We attempt to submit an invalid transport public key (a point not on the curve).
// This should raise an error and not allow that participant to proceed.
func TestRegistrationBad1(t *testing.T) {
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

	// Do Register task
	state := dkgStates[0]
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	task := &dkgtasks.RegisterTask{}
	task = dkgtasks.NewRegisterTask(state)
	err = task.Initialize(ctx, logger, eth, state)
	assert.Nil(t, err)
	// Mess up private key
	state.TransportPrivateKey = big.NewInt(0)
	// Mess up public key; this should fail because it is invalid
	state.TransportPublicKey = [2]*big.Int{big.NewInt(0), big.NewInt(1)}
	err = task.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
}

// We attempt to submit an invalid transport public key (submit identity element).
// This should raise an error and not allow that participant to proceed.
func TestRegistrationBad2(t *testing.T) {
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

	// Do Register task
	state := dkgStates[0]
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	task := &dkgtasks.RegisterTask{}
	task = dkgtasks.NewRegisterTask(state)
	err = task.Initialize(ctx, logger, eth, state)
	assert.Nil(t, err)
	// Mess up private key
	state.TransportPrivateKey = big.NewInt(0)
	// Mess up public key; this should fail because it is invalid (the identity element)
	state.TransportPublicKey = [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	err = task.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
}

// Here we expect the test to fail because we provide an invalid
// value in the state interface.
func TestRegistrationBad3(t *testing.T) {
	_, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(5)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a task to register and make sure it succeeds
	state := objects.NewDkgState(acct)
	task := dkgtasks.NewRegisterTask(state)
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

// Everything should succeed but we have too few validators.
func TestRegistrationBad4(t *testing.T) {
	// Need to run registration with 3 validators.
	// After registration, we need to make sure that after we attempt
	// to distribute shares, we receive a required forced restart
	// of EthDKG.
	n := 3
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

	eth.Commit()
	eth.Commit()
	eth.Commit()
	eth.Commit()

	// Do Share Distribution task
	shareDistributionTask0 := &dkgtasks.ShareDistributionTask{}
	state0 := dkgStates[0]
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	shareDistributionTask0 = dkgtasks.NewShareDistributionTask(state0)
	err = shareDistributionTask0.Initialize(ctx, logger, eth, state0)
	assert.Nil(t, err)
	err = shareDistributionTask0.DoWork(ctx, logger, eth)
	assert.Nil(t, err)

	eth.Commit()

	// Do Share Distribution task again; this should fail because of a restart.
	// Specifically, the error arises because we attempt to retrieve
	// the total number of participants.
	// Because we are restarting, the participant numberis 0.
	shareDistributionTask1 := &dkgtasks.ShareDistributionTask{}
	state1 := dkgStates[1]
	logger = logging.GetLogger("test").WithField("Validator", accounts[1].Address.String())

	shareDistributionTask1 = dkgtasks.NewShareDistributionTask(state1)
	err = shareDistributionTask1.Initialize(ctx, logger, eth, state1)
	assert.NotNil(t, err)
}

// We attempt invalid registration.
// Here, we try to register after registration has closed.
// This should raise an error.
func TestRegistrationBad5(t *testing.T) {
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

	// Do share distribution; afterward, we confirm who is valid and who is not
	advanceTo(t, eth, dkgStates[0].ShareDistributionStart)

	// Do Register task
	state := dkgStates[0]
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	task := &dkgtasks.RegisterTask{}
	task = dkgtasks.NewRegisterTask(state)
	err = task.Initialize(ctx, logger, eth, state)
	assert.Nil(t, err)
	err = task.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
}
