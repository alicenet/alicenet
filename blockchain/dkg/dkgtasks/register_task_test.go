package dkgtasks_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgevents"
	"github.com/MadBase/MadNet/blockchain/monitor"

	"github.com/MadBase/bridge/bindings"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/dkg/dtest"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRegisterTask(t *testing.T) {
	n := 5
	ecdsaPrivateKeys, _ := dtest.InitializePrivateKeysAndAccounts(n)

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

	eventMap := monitor.GetETHDKGEvents()
	eventInfo, ok := eventMap["RegistrationOpened"]
	if !ok {
		t.Fatal("event not found: RegistrationOpened")
	}

	var openLog *types.Log
	for _, log := range rcpt.Logs {
		eventSelector := log.Topics[0].String()
		if eventSelector == eventInfo.ID.String() {
			openLog = log
		}
	}

	// Make sure we found the open event
	assert.NotNil(t, openLog)
	openEvent, err := c.Ethdkg().ParseRegistrationOpened(*openLog)
	assert.Nil(t, err)

	// Create a task to register and make sure it succeeds
	state, registrationEnds, registrationTask, _ := dkgevents.UpdateStateOnRegistrationOpened(
		acct,
		openEvent.StartBlock.Uint64(),
		openEvent.PhaseLength.Uint64(),
		openEvent.ConfirmationLength.Uint64(),
		openEvent.Nonce.Uint64())

	t.Logf("registration ends:%v", registrationEnds)

	log := logger.WithField("TaskID", "foo")

	err = registrationTask.Initialize(ctx, log, eth, state)
	assert.Nil(t, err)

	err = registrationTask.DoWork(ctx, log, eth)
	assert.Nil(t, err)
}

// We attempt valid registration. Everything should succeed.
// This test calls Initialize and DoWork.
func TestRegistrationGood2(t *testing.T) {
	n := 6
	ecdsaPrivateKeys, accounts := dtest.InitializePrivateKeysAndAccounts(n)

	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)
	defer eth.Close()

	ctx := context.Background()

	owner := accounts[0]

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	txn, err := eth.Contracts().Ethdkg().SetPhaseLength(ownerOpts, 100)
	assert.Nil(t, err)
	_, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	txn, err = eth.Contracts().ValidatorPool().InitializeETHDKG(ownerOpts)
	assert.Nil(t, err)

	eth.Commit()
	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	eventMap := monitor.GetETHDKGEvents()
	eventInfo, ok := eventMap["RegistrationOpened"]
	if !ok {
		t.Fatal("event not found: RegistrationOpened")
	}

	var event *bindings.ETHDKGRegistrationOpened
	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == eventInfo.ID.String() {
			event, err = eth.Contracts().Ethdkg().ParseRegistrationOpened(*log)
			assert.Nil(t, err)
			break
		}
	}
	assert.NotNil(t, event)

	// Do Register task
	tasks := make([]*dkgtasks.RegisterTask, n)
	dkgStates := make([]*objects.DkgState, n)
	for idx := 0; idx < n; idx++ {
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())
		state, _, registrationTask, _ := dkgevents.UpdateStateOnRegistrationOpened(
			accounts[idx],
			event.StartBlock.Uint64(),
			event.PhaseLength.Uint64(),
			event.ConfirmationLength.Uint64(),
			event.Nonce.Uint64())

		dkgStates[idx] = state
		tasks[idx] = registrationTask
		err = tasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, tasks[idx].Success)
	}

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
	ecdsaPrivateKeys, accounts := dtest.InitializePrivateKeysAndAccounts(n)

	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)
	defer eth.Close()

	ctx := context.Background()

	owner := accounts[0]

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	txn, err := eth.Contracts().Ethdkg().SetPhaseLength(ownerOpts, 100)
	assert.Nil(t, err)
	_, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	txn, err = eth.Contracts().ValidatorPool().InitializeETHDKG(ownerOpts)
	assert.Nil(t, err)

	eth.Commit()
	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	eventMap := monitor.GetETHDKGEvents()
	eventInfo, ok := eventMap["RegistrationOpened"]
	if !ok {
		t.Fatal("event not found: RegistrationOpened")
	}

	var event *bindings.ETHDKGRegistrationOpened
	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == eventInfo.ID.String() {
			event, err = eth.Contracts().Ethdkg().ParseRegistrationOpened(*log)
			assert.Nil(t, err)
			break
		}
	}
	assert.NotNil(t, event)

	// Do Register task
	state, _, registrationTask, _ := dkgevents.UpdateStateOnRegistrationOpened(
		accounts[0],
		event.StartBlock.Uint64(),
		event.PhaseLength.Uint64(),
		event.ConfirmationLength.Uint64(),
		event.Nonce.Uint64())

	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	err = registrationTask.Initialize(ctx, logger, eth, state)
	assert.Nil(t, err)
	// Mess up private key
	state.TransportPrivateKey = big.NewInt(0)
	// Mess up public key; this should fail because it is invalid
	state.TransportPublicKey = [2]*big.Int{big.NewInt(0), big.NewInt(1)}
	err = registrationTask.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
}

// We attempt to submit an invalid transport public key (submit identity element).
// This should raise an error and not allow that participant to proceed.
func TestRegistrationBad2(t *testing.T) {
	n := 7
	ecdsaPrivateKeys, accounts := dtest.InitializePrivateKeysAndAccounts(n)

	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)
	defer eth.Close()

	ctx := context.Background()
	owner := accounts[0]

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	txn, err := eth.Contracts().Ethdkg().SetPhaseLength(ownerOpts, 100)
	assert.Nil(t, err)
	_, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	txn, err = eth.Contracts().ValidatorPool().InitializeETHDKG(ownerOpts)
	assert.Nil(t, err)

	eth.Commit()
	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	eventMap := monitor.GetETHDKGEvents()
	eventInfo, ok := eventMap["RegistrationOpened"]
	if !ok {
		t.Fatal("event not found: RegistrationOpened")
	}

	var event *bindings.ETHDKGRegistrationOpened
	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == eventInfo.ID.String() {
			event, err = eth.Contracts().Ethdkg().ParseRegistrationOpened(*log)
			assert.Nil(t, err)
			break
		}
	}
	assert.NotNil(t, event)

	// Do Register task
	state, _, registrationTask, _ := dkgevents.UpdateStateOnRegistrationOpened(
		accounts[0],
		event.StartBlock.Uint64(),
		event.PhaseLength.Uint64(),
		event.ConfirmationLength.Uint64(),
		event.Nonce.Uint64())
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	err = registrationTask.Initialize(ctx, logger, eth, state)
	assert.Nil(t, err)
	// Mess up private key
	state.TransportPrivateKey = big.NewInt(0)
	// Mess up public key; this should fail because it is invalid (the identity element)
	state.TransportPublicKey = [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	err = registrationTask.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
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

	_, err = eth.Contracts().ValidatorPool().InitializeETHDKG(ownerOpts)
	assert.NotNil(t, err)
}

// We attempt invalid registration.
// Here, we try to register after registration has closed.
// This should raise an error.
func TestRegistrationBad5(t *testing.T) {
	n := 5
	ecdsaPrivateKeys, accounts := dtest.InitializePrivateKeysAndAccounts(n)

	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)
	defer eth.Close()

	ctx := context.Background()

	owner := accounts[0]

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	txn, err := eth.Contracts().Ethdkg().SetPhaseLength(ownerOpts, 100)
	assert.Nil(t, err)
	_, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	txn, err = eth.Contracts().ValidatorPool().InitializeETHDKG(ownerOpts)
	assert.Nil(t, err)

	eth.Commit()
	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	eventMap := monitor.GetETHDKGEvents()
	eventInfo, ok := eventMap["RegistrationOpened"]
	if !ok {
		t.Fatal("event not found: RegistrationOpened")
	}

	var event *bindings.ETHDKGRegistrationOpened
	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == eventInfo.ID.String() {
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
	state, _, registrationTask, _ := dkgevents.UpdateStateOnRegistrationOpened(
		accounts[0],
		event.StartBlock.Uint64(),
		event.PhaseLength.Uint64(),
		event.ConfirmationLength.Uint64(),
		event.Nonce.Uint64())
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	err = registrationTask.Initialize(ctx, logger, eth, state)
	assert.Nil(t, err)
	err = registrationTask.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
}

// ShouldRetry() return false because the registration was successful
func TestRegisterTaskShouldRetryFalse(t *testing.T) {
	n := 5
	ecdsaPrivateKeys, accounts := dtest.InitializePrivateKeysAndAccounts(n)

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

	eventMap := monitor.GetETHDKGEvents()
	eventInfo, ok := eventMap["RegistrationOpened"]
	if !ok {
		t.Fatal("event not found: RegistrationOpened")
	}

	var openLog *types.Log
	for _, log := range rcpt.Logs {
		eventSelector := log.Topics[0].String()
		if eventSelector == eventInfo.ID.String() {
			openLog = log
		}
	}

	// Make sure we found the open event
	assert.NotNil(t, openLog)
	openEvent, err := c.Ethdkg().ParseRegistrationOpened(*openLog)
	assert.Nil(t, err)

	// Do Register task
	state, registrationEnds, registrationTask, _ := dkgevents.UpdateStateOnRegistrationOpened(
		accounts[0],
		openEvent.StartBlock.Uint64(),
		openEvent.PhaseLength.Uint64(),
		openEvent.ConfirmationLength.Uint64(),
		openEvent.Nonce.Uint64())

	t.Logf("registration ends:%v", registrationEnds)

	log := logger.WithField("TaskID", "foo")

	err = registrationTask.Initialize(ctx, log, eth, state)
	assert.Nil(t, err)

	err = registrationTask.DoWork(ctx, log, eth)
	assert.Nil(t, err)

	retry := registrationTask.ShouldRetry(ctx, log, eth)
	assert.False(t, retry)
}

// ShouldRetry() return true because the registration was unsuccessful
func TestRegisterTaskShouldRetryTrue(t *testing.T) {
	n := 5
	ecdsaPrivateKeys, accounts := dtest.InitializePrivateKeysAndAccounts(n)

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

	eventMap := monitor.GetETHDKGEvents()
	eventInfo, ok := eventMap["RegistrationOpened"]
	if !ok {
		t.Fatal("event not found: RegistrationOpened")
	}

	var openLog *types.Log
	for _, log := range rcpt.Logs {
		eventSelector := log.Topics[0].String()
		if eventSelector == eventInfo.ID.String() {
			openLog = log
		}
	}

	// Make sure we found the open event
	assert.NotNil(t, openLog)
	openEvent, err := c.Ethdkg().ParseRegistrationOpened(*openLog)
	assert.Nil(t, err)

	// Do Register task
	state, registrationEnds, registrationTask, _ := dkgevents.UpdateStateOnRegistrationOpened(
		accounts[0],
		openEvent.StartBlock.Uint64(),
		openEvent.PhaseLength.Uint64(),
		openEvent.ConfirmationLength.Uint64(),
		openEvent.Nonce.Uint64())

	t.Logf("registration ends:%v", registrationEnds)

	log := logger.WithField("TaskID", "foo")

	err = registrationTask.Initialize(ctx, log, eth, state)
	assert.Nil(t, err)

	state.TransportPublicKey[0] = big.NewInt(0)
	state.TransportPublicKey[0] = big.NewInt(0)
	err = registrationTask.DoWork(ctx, log, eth)
	assert.NotNil(t, err)

	retry := registrationTask.ShouldRetry(ctx, log, eth)
	assert.True(t, retry)
}
