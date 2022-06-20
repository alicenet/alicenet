//go:build integration

package dkg_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/testutils"
	"github.com/MadBase/MadNet/layer1/executor/tasks/dkg"
	"github.com/MadBase/MadNet/layer1/executor/tasks/dkg/state"
	dkgTestUtils "github.com/MadBase/MadNet/layer1/executor/tasks/dkg/testutils"
	dkgUtils "github.com/MadBase/MadNet/layer1/executor/tasks/dkg/utils"
	"github.com/MadBase/MadNet/layer1/monitor/events"

	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRegisterTask_Group_1_Task(t *testing.T) {
	n := 5
	ecdsaPrivateKeys, _ := testutils.InitializePrivateKeysAndAccounts(n)

	//This shouldn't be needed
	//tr := &objects.TypeRegistry{}
	//tr.RegisterInstanceType(&RegisterTask{})

	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 500*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := eth.Contracts()

	// Check status
	callOpts, err := eth.GetCallOpts(ctx, acct)
	assert.Nil(t, err)
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
	txn, rcpt, err = dkgTestUtils.SetETHDKGPhaseLength(4, eth, txnOpts, ctx)
	assert.Nil(t, err)

	t.Logf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Cost())

	// Kick off ethdkg
	_, rcpt, err = dkgTestUtils.InitializeETHDKG(eth, txnOpts, ctx)
	assert.Nil(t, err)

	t.Logf("Kicking off EthDKG used %v gas", rcpt.GasUsed)
	t.Logf("registration opens:%v", rcpt.BlockNumber)

	openEvent, err := dkgTestUtils.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, openEvent)

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()

	//logger := logging.GetLogger("test").WithField("action", "GetValidatorAddressesFromPool")
	// callOpts := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	validatorAddresses, err := dkgUtils.GetValidatorAddressesFromPool(callOpts, eth, logger.WithField("action", "GetValidatorAddressesFromPool"))
	assert.Nil(t, err)

	// Create a task to register and make sure it succeeds
	_, registrationTask, _ := events.UpdateStateOnRegistrationOpened(
		acct,
		openEvent.StartBlock.Uint64(),
		openEvent.PhaseLength.Uint64(),
		openEvent.ConfirmationLength.Uint64(),
		openEvent.Nonce.Uint64(),
		true,
		validatorAddresses,
	)

	log := logger.WithField("TaskID", "foo")

	err = registrationTask.Initialize(ctx, log, eth)
	assert.Nil(t, err)

	err = registrationTask.DoWork(ctx, log, eth)
	assert.Nil(t, err)
}

// We attempt valid registration. Everything should succeed.
// This test calls Initialize and Execute.
func TestRegisterTask_Group_1_Good2(t *testing.T) {
	n := 6
	ecdsaPrivateKeys, accounts := testutils.InitializePrivateKeysAndAccounts(n)

	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)
	defer eth.Close()

	ctx := context.Background()

	owner := accounts[0]

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	txn, rcpt, err := dkgTestUtils.SetETHDKGPhaseLength(100, eth, ownerOpts, ctx)
	assert.Nil(t, err)

	t.Logf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Cost())

	// Kick off ethdkg
	txn, rcpt, err = dkgTestUtils.InitializeETHDKG(eth, ownerOpts, ctx)
	assert.Nil(t, err)

	t.Logf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Cost())

	event, err := dkgTestUtils.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, event)

	// get validator addresses
	//ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cf()
	logger := logging.GetLogger("test").WithField("action", "GetValidatorAddressesFromPool")
	callOpts, err := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)
	validatorAddresses, err := dkgUtils.GetValidatorAddressesFromPool(callOpts, eth, logger.WithField("action", "GetValidatorAddressesFromPool"))
	assert.Nil(t, err)

	// Do Register task
	tasksVec := make([]*dkg.RegisterTask, n)
	dkgStates := make([]*state.DkgState, n)
	for idx := 0; idx < n; idx++ {
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())
		state, registrationTask, _ := events.UpdateStateOnRegistrationOpened(
			accounts[idx],
			event.StartBlock.Uint64(),
			event.PhaseLength.Uint64(),
			event.ConfirmationLength.Uint64(),
			event.Nonce.Uint64(),
			true,
			validatorAddresses,
		)

		dkgStates[idx] = state
		tasksVec[idx] = registrationTask

		err = tasksVec[idx].Initialize(ctx, logger, eth)
		assert.Nil(t, err)
		err = tasksVec[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, tasksVec[idx].Success)
	}

	// Check public keys are present and valid; last will be invalid
	for idx, acct := range accounts {
		callOpts, err := eth.GetCallOpts(context.Background(), acct)
		assert.Nil(t, err)
		p, err := eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		// check points
		publicKey := dkgStates[idx].TransportPublicKey
		if (publicKey[0].Cmp(p.PublicKey[0]) != 0) || (publicKey[1].Cmp(p.PublicKey[1]) != 0) {
			t.Fatal("Invalid public key")
		}
		if p.Phase != uint8(state.RegistrationOpen) {
			t.Fatal("Invalid participant phase")
		}

	}
}

// We attempt to submit an invalid transport public key (a point not on the curve).
// This should raise an error and not allow that participant to proceed.
func TestRegisterTask_Group_1_Bad1(t *testing.T) {
	n := 5
	ecdsaPrivateKeys, accounts := testutils.InitializePrivateKeysAndAccounts(n)

	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)
	defer eth.Close()

	ctx := context.Background()

	owner := accounts[0]
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	_, _, err = dkgTestUtils.SetETHDKGPhaseLength(100, eth, ownerOpts, ctx)
	assert.Nil(t, err)

	// Start EthDKG
	_, rcpt, err := dkgTestUtils.InitializeETHDKG(eth, ownerOpts, ctx)
	assert.Nil(t, err)

	event, err := dkgTestUtils.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, event)

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	logger := logging.GetLogger("test").WithField("action", "GetValidatorAddressesFromPool")
	callOpts, err := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)
	validatorAddresses, err := dkgUtils.GetValidatorAddressesFromPool(callOpts, eth, logger.WithField("action", "GetValidatorAddressesFromPool"))
	assert.Nil(t, err)

	// Do Register task
	state, registrationTask, _ := events.UpdateStateOnRegistrationOpened(
		accounts[0],
		event.StartBlock.Uint64(),
		event.PhaseLength.Uint64(),
		event.ConfirmationLength.Uint64(),
		event.Nonce.Uint64(),
		true,
		validatorAddresses,
	)

	logger = logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	err = registrationTask.Initialize(ctx, logger, eth)
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
func TestRegisterTask_Group_2_Bad2(t *testing.T) {
	n := 7
	ecdsaPrivateKeys, accounts := testutils.InitializePrivateKeysAndAccounts(n)

	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)
	defer eth.Close()

	ctx := context.Background()
	owner := accounts[0]
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	_, _, err = dkgTestUtils.SetETHDKGPhaseLength(100, eth, ownerOpts, ctx)
	assert.Nil(t, err)

	// Start EthDKG
	_, rcpt, err := dkgTestUtils.InitializeETHDKG(eth, ownerOpts, ctx)
	assert.Nil(t, err)

	event, err := dkgTestUtils.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, event)

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	logger := logging.GetLogger("test").WithField("action", "GetValidatorAddressesFromPool")
	callOpts, err := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)
	validatorAddresses, err := dkgUtils.GetValidatorAddressesFromPool(callOpts, eth, logger.WithField("action", "GetValidatorAddressesFromPool"))
	assert.Nil(t, err)

	// Do Register task
	state, registrationTask, _ := events.UpdateStateOnRegistrationOpened(
		accounts[0],
		event.StartBlock.Uint64(),
		event.PhaseLength.Uint64(),
		event.ConfirmationLength.Uint64(),
		event.Nonce.Uint64(),
		true,
		validatorAddresses,
	)
	logger = logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	err = registrationTask.Initialize(ctx, logger, eth)
	assert.Nil(t, err)
	// Mess up private key
	state.TransportPrivateKey = big.NewInt(0)
	// Mess up public key; this should fail because it is invalid (the identity element)
	state.TransportPublicKey = [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	err = registrationTask.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
}

// The initialization should fail because we dont allow less than 4 validators
func TestRegisterTask_Group_2_Bad4(t *testing.T) {
	n := 3
	ecdsaPrivateKeys, _ := testutils.InitializePrivateKeysAndAccounts(n)

	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
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
	_, _, err = dkgTestUtils.SetETHDKGPhaseLength(100, eth, ownerOpts, ctx)
	assert.Nil(t, err)

	// Start EthDKG
	_, _, err = dkgTestUtils.InitializeETHDKG(eth, ownerOpts, ctx)
	assert.NotNil(t, err)
}

// We attempt invalid registration.
// Here, we try to register after registration has closed.
// This should raise an error.
func TestRegisterTask_Group_2_Bad5(t *testing.T) {
	n := 5
	ecdsaPrivateKeys, accounts := testutils.InitializePrivateKeysAndAccounts(n)

	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)
	defer eth.Close()

	ctx := context.Background()
	owner := accounts[0]
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	_, _, err = dkgTestUtils.SetETHDKGPhaseLength(100, eth, ownerOpts, ctx)
	assert.Nil(t, err)

	// Start EthDKG
	_, rcpt, err := dkgTestUtils.InitializeETHDKG(eth, ownerOpts, ctx)
	assert.Nil(t, err)

	event, err := dkgTestUtils.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, event)

	// Do share distribution; afterward, we confirm who is valid and who is not
	testutils.AdvanceTo(t, eth, event.StartBlock.Uint64()+event.PhaseLength.Uint64())

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	logger := logging.GetLogger("test").WithField("action", "GetValidatorAddressesFromPool")
	callOpts, err := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)
	validatorAddresses, err := dkgUtils.GetValidatorAddressesFromPool(callOpts, eth, logger.WithField("action", "GetValidatorAddressesFromPool"))
	assert.Nil(t, err)

	// Do Register task
	_, registrationTask, _ := events.UpdateStateOnRegistrationOpened(
		accounts[0],
		event.StartBlock.Uint64(),
		event.PhaseLength.Uint64(),
		event.ConfirmationLength.Uint64(),
		event.Nonce.Uint64(),
		true,
		validatorAddresses,
	)
	logger = logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	err = registrationTask.Initialize(ctx, logger, eth)
	assert.Nil(t, err)
	err = registrationTask.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
}

// ShouldExecute() return false because the registration was successful
func TestRegisterTask_Group_3_ShouldRetryFalse(t *testing.T) {
	n := 5
	ecdsaPrivateKeys, accounts := testutils.InitializePrivateKeysAndAccounts(n)

	t.Logf("ecdsaPrivateKeys:%v", ecdsaPrivateKeys)

	//This shouldn't be needed
	//tr := &objects.TypeRegistry{}
	//
	//tr.RegisterInstanceType(&RegisterTask{})

	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 500*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := eth.Contracts()

	// Check status
	callOpts, err := eth.GetCallOpts(ctx, acct)
	assert.Nil(t, err)
	valid, err := c.ValidatorPool().IsValidator(callOpts, acct.Address)
	assert.Nil(t, err, "Failed checking validator status")
	assert.True(t, valid)

	// Kick off EthDKG
	txnOpts, err := eth.GetTransactionOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)

	// var (
	// 	txn  *types.Transaction
	// 	rcpt *types.Receipt
	// )

	// Shorten ethdkg phase for testing purposes
	txn, rcpt, err := dkgTestUtils.SetETHDKGPhaseLength(4, eth, txnOpts, ctx)
	assert.Nil(t, err)

	t.Logf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Cost())

	// Start EthDKG
	_, rcpt, err = dkgTestUtils.InitializeETHDKG(eth, txnOpts, ctx)
	assert.Nil(t, err)
	assert.NotNil(t, rcpt)

	t.Logf("Kicking off EthDKG used %v gas", rcpt.GasUsed)
	t.Logf("registration opens:%v", rcpt.BlockNumber)

	openEvent, err := dkgTestUtils.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, openEvent)

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	validatorAddresses, err := dkgUtils.GetValidatorAddressesFromPool(callOpts, eth, logger.WithField("action", "GetValidatorAddressesFromPool"))
	assert.Nil(t, err)

	// Do Register task
	_, registrationTask, _ := events.UpdateStateOnRegistrationOpened(
		accounts[0],
		openEvent.StartBlock.Uint64(),
		openEvent.PhaseLength.Uint64(),
		openEvent.ConfirmationLength.Uint64(),
		openEvent.Nonce.Uint64(),
		true,
		validatorAddresses,
	)

	log := logger.WithField("TaskID", "foo")

	err = registrationTask.Initialize(ctx, log, eth)
	assert.Nil(t, err)

	err = registrationTask.DoWork(ctx, log, eth)
	assert.Nil(t, err)

	eth.Commit()
	retry := registrationTask.ShouldExecute(ctx, log, eth)
	assert.False(t, retry)
}

// ShouldExecute() return true because the registration was unsuccessful
func TestRegisterTask_Group_3_ShouldRetryTrue(t *testing.T) {
	n := 5
	ecdsaPrivateKeys, accounts := testutils.InitializePrivateKeysAndAccounts(n)

	t.Logf("ecdsaPrivateKeys:%v", ecdsaPrivateKeys)

	//This shouldn't be needed
	//tr := &objects.TypeRegistry{}
	//
	//tr.RegisterInstanceType(&RegisterTask{})

	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 500*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := eth.Contracts()

	// Check status
	callOpts, err := eth.GetCallOpts(ctx, acct)
	assert.Nil(t, err)
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
	txn, rcpt, err = dkgTestUtils.SetETHDKGPhaseLength(4, eth, txnOpts, ctx)
	assert.Nil(t, err)

	t.Logf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Cost())

	// Start EthDKG
	_, rcpt, err = dkgTestUtils.InitializeETHDKG(eth, txnOpts, ctx)
	assert.Nil(t, err)

	t.Logf("Kicking off EthDKG used %v gas", rcpt.GasUsed)
	t.Logf("registration opens:%v", rcpt.BlockNumber)

	openEvent, err := dkgTestUtils.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, openEvent)

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 30*time.Second)
	defer cf()
	//logger = logging.GetLogger("test").WithField("action", "GetValidatorAddressesFromPool")
	//callOpts := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	validatorAddresses, err := dkgUtils.GetValidatorAddressesFromPool(callOpts, eth, logger.WithField("action", "GetValidatorAddressesFromPool"))
	assert.Nil(t, err)

	// Do Register task
	state, registrationTask, _ := events.UpdateStateOnRegistrationOpened(
		accounts[0],
		openEvent.StartBlock.Uint64(),
		openEvent.PhaseLength.Uint64(),
		openEvent.ConfirmationLength.Uint64(),
		openEvent.Nonce.Uint64(),
		true,
		validatorAddresses,
	)

	log := logger.WithField("TaskID", "foo")

	err = registrationTask.Initialize(ctx, log, eth)
	assert.Nil(t, err)

	state.TransportPublicKey[0] = big.NewInt(0)
	state.TransportPublicKey[0] = big.NewInt(0)
	err = registrationTask.DoWork(ctx, log, eth)
	assert.NotNil(t, err)

	retry := registrationTask.ShouldExecute(ctx, log, eth)
	assert.True(t, retry)
}
