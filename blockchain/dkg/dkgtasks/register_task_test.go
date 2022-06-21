//go:build integration

package dkgtasks_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/alicenet/alicenet/blockchain/dkg"
	"github.com/alicenet/alicenet/blockchain/dkg/dkgevents"

	"github.com/alicenet/alicenet/blockchain/dkg/dkgtasks"
	"github.com/alicenet/alicenet/blockchain/dkg/dtest"
	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/alicenet/alicenet/logging"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRegisterTask_Group_1_Task(t *testing.T) {
	n := 5
	ecdsaPrivateKeys, _ := dtest.InitializePrivateKeysAndAccounts(n)
	tr := &objects.TypeRegistry{}
	tr.RegisterInstanceType(&dkgtasks.RegisterTask{})

	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 500*time.Millisecond)
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
	txn, rcpt, err = dkg.SetETHDKGPhaseLength(4, eth, txnOpts, ctx)
	assert.Nil(t, err)

	t.Logf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Cost())

	// Kick off ethdkg
	_, rcpt, err = dkg.InitializeETHDKG(eth, txnOpts, ctx)
	assert.Nil(t, err)

	t.Logf("Kicking off EthDKG used %v gas", rcpt.GasUsed)
	t.Logf("registration opens:%v", rcpt.BlockNumber)

	openEvent, err := dtest.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, openEvent)

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()

	//logger := logging.GetLogger("test").WithField("action", "GetValidatorAddressesFromPool")
	// callOpts := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	validatorAddresses, err := dkg.GetValidatorAddressesFromPool(callOpts, eth, logger.WithField("action", "GetValidatorAddressesFromPool"))
	assert.Nil(t, err)

	// Create a task to register and make sure it succeeds
	state, registrationEnds, registrationTask, _ := dkgevents.UpdateStateOnRegistrationOpened(
		acct,
		openEvent.StartBlock.Uint64(),
		openEvent.PhaseLength.Uint64(),
		openEvent.ConfirmationLength.Uint64(),
		openEvent.Nonce.Uint64(),
		true,
		validatorAddresses,
	)

	t.Logf("registration ends:%v", registrationEnds)

	log := logger.WithField("TaskID", "foo")

	dkgData := objects.NewETHDKGTaskData(state)
	err = registrationTask.Initialize(ctx, log, eth, dkgData)
	assert.Nil(t, err)

	err = registrationTask.DoWork(ctx, log, eth)
	assert.Nil(t, err)
}

// We attempt valid registration. Everything should succeed.
// This test calls Initialize and DoWork.
func TestRegisterTask_Group_1_Good2(t *testing.T) {
	n := 6
	ecdsaPrivateKeys, accounts := dtest.InitializePrivateKeysAndAccounts(n)

	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)
	defer eth.Close()

	ctx := context.Background()

	owner := accounts[0]

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	txn, rcpt, err := dkg.SetETHDKGPhaseLength(100, eth, ownerOpts, ctx)
	assert.Nil(t, err)

	t.Logf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Cost())

	// Kick off ethdkg
	txn, rcpt, err = dkg.InitializeETHDKG(eth, ownerOpts, ctx)
	assert.Nil(t, err)

	t.Logf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Cost())

	event, err := dtest.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, event)

	// get validator addresses
	//ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cf()
	logger := logging.GetLogger("test").WithField("action", "GetValidatorAddressesFromPool")
	callOpts := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	validatorAddresses, err := dkg.GetValidatorAddressesFromPool(callOpts, eth, logger.WithField("action", "GetValidatorAddressesFromPool"))
	assert.Nil(t, err)

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
			event.Nonce.Uint64(),
			true,
			validatorAddresses,
		)

		dkgStates[idx] = state
		tasks[idx] = registrationTask
		dkgData := objects.NewETHDKGTaskData(state)
		err = tasks[idx].Initialize(ctx, logger, eth, dkgData)
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
func TestRegisterTask_Group_1_Bad1(t *testing.T) {
	n := 5
	ecdsaPrivateKeys, accounts := dtest.InitializePrivateKeysAndAccounts(n)

	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)
	defer eth.Close()

	ctx := context.Background()

	owner := accounts[0]
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	_, _, err = dkg.SetETHDKGPhaseLength(100, eth, ownerOpts, ctx)
	assert.Nil(t, err)

	// Start EthDKG
	_, rcpt, err := dkg.InitializeETHDKG(eth, ownerOpts, ctx)
	assert.Nil(t, err)

	event, err := dtest.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, event)

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	logger := logging.GetLogger("test").WithField("action", "GetValidatorAddressesFromPool")
	callOpts := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	validatorAddresses, err := dkg.GetValidatorAddressesFromPool(callOpts, eth, logger.WithField("action", "GetValidatorAddressesFromPool"))
	assert.Nil(t, err)

	// Do Register task
	state, _, registrationTask, _ := dkgevents.UpdateStateOnRegistrationOpened(
		accounts[0],
		event.StartBlock.Uint64(),
		event.PhaseLength.Uint64(),
		event.ConfirmationLength.Uint64(),
		event.Nonce.Uint64(),
		true,
		validatorAddresses,
	)

	logger = logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	dkgData := objects.NewETHDKGTaskData(state)
	err = registrationTask.Initialize(ctx, logger, eth, dkgData)
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
	ecdsaPrivateKeys, accounts := dtest.InitializePrivateKeysAndAccounts(n)

	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)
	defer eth.Close()

	ctx := context.Background()
	owner := accounts[0]
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	_, _, err = dkg.SetETHDKGPhaseLength(100, eth, ownerOpts, ctx)
	assert.Nil(t, err)

	// Start EthDKG
	_, rcpt, err := dkg.InitializeETHDKG(eth, ownerOpts, ctx)
	assert.Nil(t, err)

	event, err := dtest.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, event)

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	logger := logging.GetLogger("test").WithField("action", "GetValidatorAddressesFromPool")
	callOpts := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	validatorAddresses, err := dkg.GetValidatorAddressesFromPool(callOpts, eth, logger.WithField("action", "GetValidatorAddressesFromPool"))
	assert.Nil(t, err)

	// Do Register task
	state, _, registrationTask, _ := dkgevents.UpdateStateOnRegistrationOpened(
		accounts[0],
		event.StartBlock.Uint64(),
		event.PhaseLength.Uint64(),
		event.ConfirmationLength.Uint64(),
		event.Nonce.Uint64(),
		true,
		validatorAddresses,
	)
	logger = logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	dkgData := objects.NewETHDKGTaskData(state)
	err = registrationTask.Initialize(ctx, logger, eth, dkgData)
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
	ecdsaPrivateKeys, _ := dtest.InitializePrivateKeysAndAccounts(n)

	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
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
	_, _, err = dkg.SetETHDKGPhaseLength(100, eth, ownerOpts, ctx)
	assert.Nil(t, err)

	// Start EthDKG
	_, _, err = dkg.InitializeETHDKG(eth, ownerOpts, ctx)
	assert.NotNil(t, err)
}

// We attempt invalid registration.
// Here, we try to register after registration has closed.
// This should raise an error.
func TestRegisterTask_Group_2_Bad5(t *testing.T) {
	n := 5
	ecdsaPrivateKeys, accounts := dtest.InitializePrivateKeysAndAccounts(n)

	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)
	defer eth.Close()

	ctx := context.Background()
	owner := accounts[0]
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	_, _, err = dkg.SetETHDKGPhaseLength(100, eth, ownerOpts, ctx)
	assert.Nil(t, err)

	// Start EthDKG
	_, rcpt, err := dkg.InitializeETHDKG(eth, ownerOpts, ctx)
	assert.Nil(t, err)

	event, err := dtest.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, event)

	// Do share distribution; afterward, we confirm who is valid and who is not
	advanceTo(t, eth, event.StartBlock.Uint64()+event.PhaseLength.Uint64())

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	logger := logging.GetLogger("test").WithField("action", "GetValidatorAddressesFromPool")
	callOpts := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	validatorAddresses, err := dkg.GetValidatorAddressesFromPool(callOpts, eth, logger.WithField("action", "GetValidatorAddressesFromPool"))
	assert.Nil(t, err)

	// Do Register task
	state, _, registrationTask, _ := dkgevents.UpdateStateOnRegistrationOpened(
		accounts[0],
		event.StartBlock.Uint64(),
		event.PhaseLength.Uint64(),
		event.ConfirmationLength.Uint64(),
		event.Nonce.Uint64(),
		true,
		validatorAddresses,
	)
	logger = logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	dkgData := objects.NewETHDKGTaskData(state)
	err = registrationTask.Initialize(ctx, logger, eth, dkgData)
	assert.Nil(t, err)
	err = registrationTask.DoWork(ctx, logger, eth)
	assert.NotNil(t, err)
}

// ShouldRetry() return false because the registration was successful
func TestRegisterTask_Group_3_ShouldRetryFalse(t *testing.T) {
	n := 5
	ecdsaPrivateKeys, accounts := dtest.InitializePrivateKeysAndAccounts(n)

	t.Logf("ecdsaPrivateKeys:%v", ecdsaPrivateKeys)

	tr := &objects.TypeRegistry{}

	tr.RegisterInstanceType(&dkgtasks.RegisterTask{})

	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 500*time.Millisecond)
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

	// var (
	// 	txn  *types.Transaction
	// 	rcpt *types.Receipt
	// )

	// Shorten ethdkg phase for testing purposes
	txn, rcpt, err := dkg.SetETHDKGPhaseLength(4, eth, txnOpts, ctx)
	assert.Nil(t, err)

	t.Logf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Cost())

	// Start EthDKG
	_, rcpt, err = dkg.InitializeETHDKG(eth, txnOpts, ctx)
	assert.Nil(t, err)
	assert.NotNil(t, rcpt)

	t.Logf("Kicking off EthDKG used %v gas", rcpt.GasUsed)
	t.Logf("registration opens:%v", rcpt.BlockNumber)

	openEvent, err := dtest.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, openEvent)

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	validatorAddresses, err := dkg.GetValidatorAddressesFromPool(callOpts, eth, logger.WithField("action", "GetValidatorAddressesFromPool"))
	assert.Nil(t, err)

	// Do Register task
	state, registrationEnds, registrationTask, _ := dkgevents.UpdateStateOnRegistrationOpened(
		accounts[0],
		openEvent.StartBlock.Uint64(),
		openEvent.PhaseLength.Uint64(),
		openEvent.ConfirmationLength.Uint64(),
		openEvent.Nonce.Uint64(),
		true,
		validatorAddresses,
	)

	t.Logf("registration ends:%v", registrationEnds)

	log := logger.WithField("TaskID", "foo")

	dkgData := objects.NewETHDKGTaskData(state)
	err = registrationTask.Initialize(ctx, log, eth, dkgData)
	assert.Nil(t, err)

	err = registrationTask.DoWork(ctx, log, eth)
	assert.Nil(t, err)

	eth.Commit()
	retry := registrationTask.ShouldRetry(ctx, log, eth)
	assert.False(t, retry)
}

// ShouldRetry() return true because the registration was unsuccessful
func TestRegisterTask_Group_3_ShouldRetryTrue(t *testing.T) {
	n := 5
	ecdsaPrivateKeys, accounts := dtest.InitializePrivateKeysAndAccounts(n)

	t.Logf("ecdsaPrivateKeys:%v", ecdsaPrivateKeys)

	tr := &objects.TypeRegistry{}

	tr.RegisterInstanceType(&dkgtasks.RegisterTask{})

	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 500*time.Millisecond)
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
	txn, rcpt, err = dkg.SetETHDKGPhaseLength(4, eth, txnOpts, ctx)
	assert.Nil(t, err)

	t.Logf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Cost())

	// Start EthDKG
	_, rcpt, err = dkg.InitializeETHDKG(eth, txnOpts, ctx)
	assert.Nil(t, err)

	t.Logf("Kicking off EthDKG used %v gas", rcpt.GasUsed)
	t.Logf("registration opens:%v", rcpt.BlockNumber)

	openEvent, err := dtest.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, openEvent)

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 30*time.Second)
	defer cf()
	//logger = logging.GetLogger("test").WithField("action", "GetValidatorAddressesFromPool")
	//callOpts := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	validatorAddresses, err := dkg.GetValidatorAddressesFromPool(callOpts, eth, logger.WithField("action", "GetValidatorAddressesFromPool"))
	assert.Nil(t, err)

	// Do Register task
	state, registrationEnds, registrationTask, _ := dkgevents.UpdateStateOnRegistrationOpened(
		accounts[0],
		openEvent.StartBlock.Uint64(),
		openEvent.PhaseLength.Uint64(),
		openEvent.ConfirmationLength.Uint64(),
		openEvent.Nonce.Uint64(),
		true,
		validatorAddresses,
	)

	t.Logf("registration ends:%v", registrationEnds)

	log := logger.WithField("TaskID", "foo")

	dkgData := objects.NewETHDKGTaskData(state)
	err = registrationTask.Initialize(ctx, log, eth, dkgData)
	assert.Nil(t, err)

	state.TransportPublicKey[0] = big.NewInt(0)
	state.TransportPublicKey[0] = big.NewInt(0)
	err = registrationTask.DoWork(ctx, log, eth)
	assert.NotNil(t, err)

	retry := registrationTask.ShouldRetry(ctx, log, eth)
	assert.True(t, retry)
}
