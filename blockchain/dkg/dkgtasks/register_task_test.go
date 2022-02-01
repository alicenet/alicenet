package dkgtasks_test

import (
	"context"
	"github.com/MadBase/bridge/bindings"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/dkg/dtest"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/logging"
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
	valid, err := c.ValidatorPool().IsValidator(callOpts, acct.Address)
	assert.Nil(t, err, "Failed checking validator status")
	assert.True(t, valid)

	// Kick off EthDKG
	txnOpts, err := eth.GetTransactionOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)

	// mess, err := c.Ethdkg().InitialMessage(callOpts)
	// assert.Nil(t, err)
	// assert.NotNil(t, mess)

	// Reuse these
	var (
		txn  *types.Transaction
		rcpt *types.Receipt
	)

	// Shorten ethdkg phase for testing purposes
	txn, err = c.Ethdkg().SetPhaseLength(txnOpts, 4)
	assert.Nil(t, err)

	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	t.Logf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Cost())

	// Kick off ethdkg
	txn, err = c.ValidatorPool().InitializeETHDKG(txnOpts)
	assert.Nil(t, err)

	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)
	assert.NotNil(t, rcpt.Logs)

	t.Logf("Kicking off EthDKG used %v gas", rcpt.GasUsed)
	t.Logf("registration opens:%v", rcpt.BlockNumber)

	var openLog *types.Log
	for _, log := range rcpt.Logs {
		eventSelector := log.Topics[0].String()
		if eventSelector == "0xbda431b9b63510f1398bf33d700e013315bcba905507078a1780f13ea5b354b9" {
			openLog = log
		}
	}

	// Make sure we found the open event
	assert.NotNil(t, openLog)
	openEvent, err := c.Ethdkg().ParseRegistrationOpened(*openLog)
	assert.Nil(t, err)

	// Create a task to register and make sure it succeeds
	state := objects.NewDkgState(acct)
	state.Phase = objects.RegistrationOpen
	state.PhaseStart = openEvent.StartBlock.Uint64()
	state.PhaseLength = openEvent.PhaseLength.Uint64()
	var registrationEnd = state.PhaseStart + state.PhaseLength

	task := dkgtasks.NewRegisterTask(state, state.PhaseStart, registrationEnd)

	t.Logf("registration ends:%v", registrationEnd)

	log := logger.WithField("TaskID", "foo")

	err = task.Initialize(ctx, log, eth, state)
	assert.Nil(t, err)

	err = task.DoWork(ctx, log, eth)
	assert.Nil(t, err)
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
	txn, err := eth.Contracts().Ethdkg().SetPhaseLength(ownerOpts, 100)
	assert.Nil(t, err)
	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	txn, err = eth.Contracts().ValidatorPool().InitializeETHDKG(ownerOpts)
	assert.Nil(t, err)

	eth.Commit()
	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	var event *bindings.ETHDKGRegistrationOpened
	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == "0xbda431b9b63510f1398bf33d700e013315bcba905507078a1780f13ea5b354b9" {
			event, err = eth.Contracts().Ethdkg().ParseRegistrationOpened(*log)
			assert.Nil(t, err)
			break
			//
			//for _, dkgState := range dkgStates {
			//	dkgevents.PopulateSchedule(dkgState, event)
			//}
		}
	}
	assert.NotNil(t, event)

	// Do Register task
	tasks := make([]*dkgtasks.RegisterTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// Set Registration success to true
		dkgStates[idx].Phase = objects.RegistrationOpen
		dkgStates[idx].PhaseStart = event.StartBlock.Uint64()
		dkgStates[idx].PhaseLength = event.PhaseLength.Uint64()
		dkgStates[idx].ConfirmationLength = event.ConfirmationLength.Uint64()
		dkgStates[idx].Nonce = event.Nonce.Uint64()
		tasks[idx] = dkgtasks.NewRegisterTask(state, event.StartBlock.Uint64(), event.StartBlock.Uint64()+event.PhaseLength.Uint64())
		err = tasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, tasks[idx].Success)
	}

	// Advance to share distribution phase; afterward, we confirm all are valid
	//advanceTo(t, eth, dkgStates[0].ShareDistributionStart)

	// Check public keys are present and valid; last will be invalid
	for idx, acct := range accounts {
		callOpts := eth.GetCallOpts(context.Background(), acct)
		p, err := eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		// check points
		publicKey := dkgStates[idx].TransportPublicKey
		if (publicKey[0].Cmp(p.PublicKey[0]) != 0) || (publicKey[1].Cmp(p.PublicKey[1]) != 0) {
			t.Fatal("Invalid public key")
		}
		if p.Phase != uint8(objects.RegistrationOpen) {
			t.Fatal("Invalid participant phase")
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
	txn, err := eth.Contracts().Ethdkg().SetPhaseLength(ownerOpts, 100)
	assert.Nil(t, err)
	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	txn, err = eth.Contracts().ValidatorPool().InitializeETHDKG(ownerOpts)
	assert.Nil(t, err)

	eth.Commit()
	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	var event *bindings.ETHDKGRegistrationOpened
	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == "0xbda431b9b63510f1398bf33d700e013315bcba905507078a1780f13ea5b354b9" {
			event, err = eth.Contracts().Ethdkg().ParseRegistrationOpened(*log)
			assert.Nil(t, err)
			break
			//
			//for _, dkgState := range dkgStates {
			//	dkgevents.PopulateSchedule(dkgState, event)
			//}
		}
	}
	assert.NotNil(t, event)

	// Do Register task
	state := dkgStates[0]
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	task := &dkgtasks.RegisterTask{}
	task = dkgtasks.NewRegisterTask(state, event.StartBlock.Uint64(), event.StartBlock.Uint64()+event.PhaseLength.Uint64())
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
	txn, err := eth.Contracts().Ethdkg().SetPhaseLength(ownerOpts, 100)
	assert.Nil(t, err)
	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	txn, err = eth.Contracts().ValidatorPool().InitializeETHDKG(ownerOpts)
	assert.Nil(t, err)

	eth.Commit()
	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	var event *bindings.ETHDKGRegistrationOpened
	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == "0xbda431b9b63510f1398bf33d700e013315bcba905507078a1780f13ea5b354b9" {
			event, err = eth.Contracts().Ethdkg().ParseRegistrationOpened(*log)
			assert.Nil(t, err)
			break
			//for _, dkgState := range dkgStates {
			//	dkgevents.PopulateSchedule(dkgState, event)
			//}
		}
	}
	assert.NotNil(t, event)

	// Do Register task
	state := dkgStates[0]
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	task := &dkgtasks.RegisterTask{}
	task = dkgtasks.NewRegisterTask(state, event.StartBlock.Uint64(), event.StartBlock.Uint64()+event.PhaseLength.Uint64())
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
	task := dkgtasks.NewRegisterTask(state, 1, 5)
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

// The initialization should fail because we dont allow less than 4 validators
func TestRegistrationBad4(t *testing.T) {
	n := 3
	_, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(n)

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
	txn, err := eth.Contracts().Ethdkg().SetPhaseLength(ownerOpts, 100)
	assert.Nil(t, err)
	_, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	txn, err = eth.Contracts().ValidatorPool().InitializeETHDKG(ownerOpts)
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
	txn, err := eth.Contracts().Ethdkg().SetPhaseLength(ownerOpts, 100)
	assert.Nil(t, err)
	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	txn, err = eth.Contracts().ValidatorPool().InitializeETHDKG(ownerOpts)
	assert.Nil(t, err)

	eth.Commit()
	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	var event *bindings.ETHDKGRegistrationOpened

	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == "0xbda431b9b63510f1398bf33d700e013315bcba905507078a1780f13ea5b354b9" {
			event, err = eth.Contracts().Ethdkg().ParseRegistrationOpened(*log)
			assert.Nil(t, err)
			break
			//for _, dkgState := range dkgStates {
			//	dkgevents.PopulateSchedule(dkgState, event)
			//}
		}
	}
	assert.NotNil(t, event)

	// Do share distribution; afterward, we confirm who is valid and who is not
	advanceTo(t, eth, event.StartBlock.Uint64()+event.PhaseLength.Uint64())

	// Do Register task
	state := dkgStates[0]
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	task := &dkgtasks.RegisterTask{}
	task = dkgtasks.NewRegisterTask(state, event.StartBlock.Uint64(), event.StartBlock.Uint64()+event.PhaseLength.Uint64())
	err = task.Initialize(ctx, logger, eth, state)
	assert.Nil(t, err)
	err = task.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
}

// ShouldRetry() return false because the registration was successful
func TestRegisterTaskShouldRetryFalse(t *testing.T) {
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
	valid, err := c.ValidatorPool().IsValidator(callOpts, acct.Address)
	assert.Nil(t, err, "Failed checking validator status")
	assert.True(t, valid)

	// Kick off EthDKG
	txnOpts, err := eth.GetTransactionOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)

	var (
		txn  *types.Transaction
		rcpt *types.Receipt
	)

	// Shorten ethdkg phase for testing purposes
	txn, err = c.Ethdkg().SetPhaseLength(txnOpts, 4)
	assert.Nil(t, err)

	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	t.Logf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Cost())

	// Kick off ethdkg
	txn, err = c.ValidatorPool().InitializeETHDKG(txnOpts)
	assert.Nil(t, err)

	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)
	assert.NotNil(t, rcpt.Logs)

	t.Logf("Kicking off EthDKG used %v gas", rcpt.GasUsed)
	t.Logf("registration opens:%v", rcpt.BlockNumber)

	var openLog *types.Log
	for _, log := range rcpt.Logs {
		eventSelector := log.Topics[0].String()
		if eventSelector == "0xbda431b9b63510f1398bf33d700e013315bcba905507078a1780f13ea5b354b9" {
			openLog = log
		}
	}

	// Make sure we found the open event
	assert.NotNil(t, openLog)
	openEvent, err := c.Ethdkg().ParseRegistrationOpened(*openLog)
	assert.Nil(t, err)

	// Create a task to register and make sure it succeeds
	state := objects.NewDkgState(acct)
	state.Phase = objects.RegistrationOpen
	state.PhaseStart = openEvent.StartBlock.Uint64()
	state.PhaseLength = openEvent.PhaseLength.Uint64()
	var registrationEnd = state.PhaseStart + state.PhaseLength

	task := dkgtasks.NewRegisterTask(state, state.PhaseStart, registrationEnd)

	t.Logf("registration ends:%v", registrationEnd)

	log := logger.WithField("TaskID", "foo")

	err = task.Initialize(ctx, log, eth, state)
	assert.Nil(t, err)

	err = task.DoWork(ctx, log, eth)
	assert.Nil(t, err)

	retry := task.ShouldRetry(ctx, log, eth)
	assert.False(t, retry)
}

// ShouldRetry() return true because the registration was unsuccessful
func TestRegisterTaskShouldRetryTrue(t *testing.T) {
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
	valid, err := c.ValidatorPool().IsValidator(callOpts, acct.Address)
	assert.Nil(t, err, "Failed checking validator status")
	assert.True(t, valid)

	// Kick off EthDKG
	txnOpts, err := eth.GetTransactionOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)

	var (
		txn  *types.Transaction
		rcpt *types.Receipt
	)

	// Shorten ethdkg phase for testing purposes
	txn, err = c.Ethdkg().SetPhaseLength(txnOpts, 4)
	assert.Nil(t, err)

	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	t.Logf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Cost())

	// Kick off ethdkg
	txn, err = c.ValidatorPool().InitializeETHDKG(txnOpts)
	assert.Nil(t, err)

	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)
	assert.NotNil(t, rcpt.Logs)

	t.Logf("Kicking off EthDKG used %v gas", rcpt.GasUsed)
	t.Logf("registration opens:%v", rcpt.BlockNumber)

	var openLog *types.Log
	for _, log := range rcpt.Logs {
		eventSelector := log.Topics[0].String()
		if eventSelector == "0xbda431b9b63510f1398bf33d700e013315bcba905507078a1780f13ea5b354b9" {
			openLog = log
		}
	}

	// Make sure we found the open event
	assert.NotNil(t, openLog)
	openEvent, err := c.Ethdkg().ParseRegistrationOpened(*openLog)
	assert.Nil(t, err)

	// Create a task to register and make sure it succeeds
	state := objects.NewDkgState(acct)
	state.Phase = objects.RegistrationOpen
	state.PhaseStart = openEvent.StartBlock.Uint64()
	state.PhaseLength = openEvent.PhaseLength.Uint64()
	var registrationEnd = state.PhaseStart + state.PhaseLength

	task := dkgtasks.NewRegisterTask(state, state.PhaseStart, registrationEnd)

	t.Logf("registration ends:%v", registrationEnd)

	log := logger.WithField("TaskID", "foo")

	err = task.Initialize(ctx, log, eth, state)
	assert.Nil(t, err)

	//k1 := state.TransportPublicKey[0]
	//k2 := state.TransportPublicKey[1]
	state.TransportPublicKey[0] = big.NewInt(0)
	state.TransportPublicKey[0] = big.NewInt(0)
	err = task.DoWork(ctx, log, eth)
	assert.NotNil(t, err)

	//state.TransportPublicKey[0] = k1
	//state.TransportPublicKey[0] = k2
	retry := task.ShouldRetry(ctx, log, eth)
	assert.True(t, retry)
}
