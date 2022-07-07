//go:build integration

package tests

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/tests/utils"
	"github.com/alicenet/alicenet/layer1/monitor/events"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestRegisterTask_Group_1_Task(t *testing.T) {
	n := 5
	fixture := setupEthereum(t, n)
	eth := fixture.Client
	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := ethereum.GetContracts()

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
	txn, rcpt, err = SetETHDKGPhaseLength(4, fixture, txnOpts, ctx)
	assert.Nil(t, err)

	fixture.Logger.Debugf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Gas())

	// Kick off ethdkg
	_, rcpt, err = InitializeETHDKG(fixture, txnOpts, ctx)
	assert.Nil(t, err)

	fixture.Logger.Debugf("Kicking off EthDKG used %v gas", rcpt.GasUsed)
	fixture.Logger.Debugf("registration opens:%v", rcpt.BlockNumber)

	openEvent, err := utils.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, openEvent)

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()

	var validatorAddresses []common.Address
	// all known addresses must be validators at this point
	for _, acc := range eth.GetKnownAccounts() {
		validatorAddresses = append(validatorAddresses, acc.Address)
	}

	// Create a task to register and make sure it succeeds
	dkgState, registrationTask, _ := events.UpdateStateOnRegistrationOpened(
		acct,
		openEvent.StartBlock.Uint64(),
		openEvent.PhaseLength.Uint64(),
		openEvent.ConfirmationLength.Uint64(),
		openEvent.Nonce.Uint64(),
		true,
		validatorAddresses,
	)
	dkgDb := GetDKGDb(t)
	err = state.SaveDkgState(dkgDb, dkgState)
	assert.Nil(t, err)
	err = registrationTask.Initialize(ctx, nil, dkgDb, fixture.Logger, eth, "RegistrationTask", "task-id", nil)
	assert.Nil(t, err)
	err = registrationTask.Prepare(ctx)
	assert.Nil(t, err)

	_, err = registrationTask.Execute(ctx)
	assert.Nil(t, err)
}

// We attempt valid registration. Everything should succeed.
// This test calls Initialize and Execute.
func TestRegisterTask_Group_1_Good2(t *testing.T) {
	n := 6
	fixture := setupEthereum(t, n)
	eth := fixture.Client
	ctx := context.Background()
	accounts := eth.GetKnownAccounts()
	owner := accounts[0]

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	txn, rcpt, err := SetETHDKGPhaseLength(100, fixture, ownerOpts, ctx)
	assert.Nil(t, err)

	fixture.Logger.Debugf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Gas())

	// Kick off ethdkg
	txn, rcpt, err = InitializeETHDKG(fixture, ownerOpts, ctx)
	assert.Nil(t, err)

	fixture.Logger.Debugf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Gas())

	event, err := utils.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, event)

	// get validator addresses
	assert.Nil(t, err)
	var validatorAddresses []common.Address
	// all known addresses must be validators at this point
	for _, acc := range eth.GetKnownAccounts() {
		validatorAddresses = append(validatorAddresses, acc.Address)
	}

	// Do Register task
	tasksVec := make([]*dkg.RegisterTask, n)
	dkgStatesDbs := make([]*db.Database, n)
	for idx := 0; idx < n; idx++ {
		dkgState, registrationTask, _ := events.UpdateStateOnRegistrationOpened(
			accounts[idx],
			event.StartBlock.Uint64(),
			event.PhaseLength.Uint64(),
			event.ConfirmationLength.Uint64(),
			event.Nonce.Uint64(),
			true,
			validatorAddresses,
		)

		dkgDb := GetDKGDb(t)
		err = state.SaveDkgState(dkgDb, dkgState)
		assert.Nil(t, err)
		dkgStatesDbs[idx] = dkgDb
		tasksVec[idx] = registrationTask

		err = tasksVec[idx].Initialize(ctx, nil, dkgDb, fixture.Logger, eth, "RegistrationTask", "task-id", nil)
		assert.Nil(t, err)
		err = tasksVec[idx].Prepare(ctx)
		assert.Nil(t, err)
		_, err := tasksVec[idx].Execute(ctx)
		assert.Nil(t, err)
	}

	tests.MineFinalityDelayBlocks(eth)
	// Check public keys are present and valid; last will be invalid
	for idx, acct := range accounts {
		callOpts, err := eth.GetCallOpts(context.Background(), acct)
		assert.Nil(t, err)
		p, err := ethereum.GetContracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(dkgStatesDbs[idx])
		assert.Nil(t, err)
		// check points
		publicKey := dkgState.TransportPublicKey
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
	fixture := setupEthereum(t, n)
	eth := fixture.Client
	ctx := context.Background()
	accounts := eth.GetKnownAccounts()
	owner := accounts[0]
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	_, _, err = SetETHDKGPhaseLength(100, fixture, ownerOpts, ctx)
	assert.Nil(t, err)

	// Start EthDKG
	_, rcpt, err := InitializeETHDKG(fixture, ownerOpts, ctx)
	assert.Nil(t, err)

	event, err := utils.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, event)

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	var validatorAddresses []common.Address
	// all known addresses must be validators at this point
	for _, acc := range eth.GetKnownAccounts() {
		validatorAddresses = append(validatorAddresses, acc.Address)
	}

	// Do Register task
	dkgState, registrationTask, _ := events.UpdateStateOnRegistrationOpened(
		accounts[0],
		event.StartBlock.Uint64(),
		event.PhaseLength.Uint64(),
		event.ConfirmationLength.Uint64(),
		event.Nonce.Uint64(),
		true,
		validatorAddresses,
	)

	dkgDb := GetDKGDb(t)
	err = state.SaveDkgState(dkgDb, dkgState)
	assert.Nil(t, err)
	err = registrationTask.Initialize(ctx, nil, dkgDb, fixture.Logger, eth, "RegistrationTask", "task-id", nil)
	assert.Nil(t, err)
	err = registrationTask.Prepare(ctx)
	assert.Nil(t, err)
	// Mess up private key
	dkgState.TransportPrivateKey = big.NewInt(0)
	// Mess up public key; this should fail because it is invalid
	dkgState.TransportPublicKey = [2]*big.Int{big.NewInt(0), big.NewInt(1)}
	state.SaveDkgState(dkgDb, dkgState)
	_, err = registrationTask.Execute(ctx)
	assert.NotNil(t, err)
}

// We attempt to submit an invalid transport public key (submit identity element).
// This should raise an error and not allow that participant to proceed.
func TestRegisterTask_Group_2_Bad2(t *testing.T) {
	n := 7
	fixture := setupEthereum(t, n)
	eth := fixture.Client
	ctx := context.Background()
	accounts := eth.GetKnownAccounts()
	owner := accounts[0]
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	_, _, err = SetETHDKGPhaseLength(100, fixture, ownerOpts, ctx)
	assert.Nil(t, err)

	// Start EthDKG
	_, rcpt, err := InitializeETHDKG(fixture, ownerOpts, ctx)
	assert.Nil(t, err)

	event, err := utils.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, event)

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	var validatorAddresses []common.Address
	// all known addresses must be validators at this point
	for _, acc := range eth.GetKnownAccounts() {
		validatorAddresses = append(validatorAddresses, acc.Address)
	}

	// Do Register task
	dkgState, registrationTask, _ := events.UpdateStateOnRegistrationOpened(
		accounts[0],
		event.StartBlock.Uint64(),
		event.PhaseLength.Uint64(),
		event.ConfirmationLength.Uint64(),
		event.Nonce.Uint64(),
		true,
		validatorAddresses,
	)

	dkgDb := GetDKGDb(t)
	err = state.SaveDkgState(dkgDb, dkgState)
	assert.Nil(t, err)
	err = registrationTask.Initialize(ctx, nil, dkgDb, fixture.Logger, eth, "RegistrationTask", "task-id", nil)
	assert.Nil(t, err)
	err = registrationTask.Prepare(ctx)
	assert.Nil(t, err)
	// Mess up private key
	dkgState.TransportPrivateKey = big.NewInt(0)
	// Mess up public key; this should fail because it is invalid (the identity element)
	dkgState.TransportPublicKey = [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	state.SaveDkgState(dkgDb, dkgState)
	_, err = registrationTask.Execute(ctx)
	assert.NotNil(t, err)
}

// The initialization should fail because we dont allow less than 4 validators
func TestRegisterTask_Group_2_Bad4(t *testing.T) {
	n := 3
	fixture := setupEthereum(t, n)
	eth := fixture.Client
	ctx := context.Background()
	accounts := eth.GetKnownAccounts()
	owner := accounts[0]

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	_, _, err = SetETHDKGPhaseLength(100, fixture, ownerOpts, ctx)
	assert.Nil(t, err)

	// Start EthDKG
	_, _, err = InitializeETHDKG(fixture, ownerOpts, ctx)
	assert.NotNil(t, err)
}

// We attempt invalid registration.
// Here, we try to register after registration has closed.
// This should raise an error.
func TestRegisterTask_Group_2_Bad5(t *testing.T) {
	n := 5
	fixture := setupEthereum(t, n)
	eth := fixture.Client
	ctx := context.Background()
	accounts := eth.GetKnownAccounts()
	owner := accounts[0]

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	_, _, err = SetETHDKGPhaseLength(100, fixture, ownerOpts, ctx)
	assert.Nil(t, err)

	// Start EthDKG
	_, rcpt, err := InitializeETHDKG(fixture, ownerOpts, ctx)
	assert.Nil(t, err)

	event, err := utils.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, event)

	// Do share distribution; afterward, we confirm who is valid and who is not
	tests.AdvanceTo(eth, event.StartBlock.Uint64()+event.PhaseLength.Uint64())

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	var validatorAddresses []common.Address
	// all known addresses must be validators at this point
	for _, acc := range eth.GetKnownAccounts() {
		validatorAddresses = append(validatorAddresses, acc.Address)
	}

	// Do Register task
	dkgState, registrationTask, _ := events.UpdateStateOnRegistrationOpened(
		accounts[0],
		event.StartBlock.Uint64(),
		event.PhaseLength.Uint64(),
		event.ConfirmationLength.Uint64(),
		event.Nonce.Uint64(),
		true,
		validatorAddresses,
	)

	dkgDb := GetDKGDb(t)
	err = state.SaveDkgState(dkgDb, dkgState)
	assert.Nil(t, err)
	err = registrationTask.Initialize(ctx, nil, dkgDb, fixture.Logger, eth, "RegistrationTask", "task-id", nil)
	assert.Nil(t, err)
	err = registrationTask.Prepare(ctx)
	assert.Nil(t, err)
	_, err = registrationTask.Execute(ctx)
	assert.NotNil(t, err)
}

// ShouldExecute() return false because the registration was successful
func TestRegisterTask_Group_3_ShouldRetryFalse(t *testing.T) {
	n := 5
	fixture := setupEthereum(t, n)
	eth := fixture.Client
	ctx := context.Background()
	accounts := eth.GetKnownAccounts()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := ethereum.GetContracts()
	tests.MineFinalityDelayBlocks(eth)
	// Check status
	callOpts, err := eth.GetCallOpts(ctx, acct)
	assert.Nil(t, err)
	valid, err := c.ValidatorPool().IsValidator(callOpts, acct.Address)
	assert.Nil(t, err, "Failed checking validator status")
	assert.True(t, valid)

	// Kick off EthDKG
	txnOpts, err := eth.GetTransactionOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	txn, rcpt, err := SetETHDKGPhaseLength(4, fixture, txnOpts, ctx)
	assert.Nil(t, err)

	fixture.Logger.Debugf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Gas())

	// Start EthDKG
	_, rcpt, err = InitializeETHDKG(fixture, txnOpts, ctx)
	assert.Nil(t, err)
	assert.NotNil(t, rcpt)

	fixture.Logger.Debugf("Kicking off EthDKG used %v gas", rcpt.GasUsed)
	fixture.Logger.Debugf("registration opens:%v", rcpt.BlockNumber)

	openEvent, err := utils.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, openEvent)

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	var validatorAddresses []common.Address
	// all known addresses must be validators at this point
	for _, acc := range eth.GetKnownAccounts() {
		validatorAddresses = append(validatorAddresses, acc.Address)
	}

	// Do Register task
	dkgState, registrationTask, _ := events.UpdateStateOnRegistrationOpened(
		accounts[0],
		openEvent.StartBlock.Uint64(),
		openEvent.PhaseLength.Uint64(),
		openEvent.ConfirmationLength.Uint64(),
		openEvent.Nonce.Uint64(),
		true,
		validatorAddresses,
	)

	dkgDb := GetDKGDb(t)
	err = state.SaveDkgState(dkgDb, dkgState)
	assert.Nil(t, err)
	err = registrationTask.Initialize(ctx, nil, dkgDb, fixture.Logger, eth, "RegistrationTask", "task-id", nil)
	assert.Nil(t, err)
	err = registrationTask.Prepare(ctx)
	assert.Nil(t, err)

	_, err = registrationTask.Execute(ctx)
	assert.Nil(t, err)

	tests.MineFinalityDelayBlocks(eth)

	retry, err := registrationTask.ShouldExecute(ctx)
	assert.Nil(t, err)
	assert.False(t, retry)
}

// ShouldExecute() return true because we din't wait for receipt or the registration failed
func TestRegisterTask_Group_3_ShouldRetryTrue(t *testing.T) {
	n := 5
	fixture := setupEthereum(t, n)
	eth := fixture.Client
	ctx := context.Background()
	accounts := eth.GetKnownAccounts()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := ethereum.GetContracts()

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
	txn, rcpt, err = SetETHDKGPhaseLength(4, fixture, txnOpts, ctx)
	assert.Nil(t, err)

	fixture.Logger.Debugf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Gas())

	// Start EthDKG
	_, rcpt, err = InitializeETHDKG(fixture, txnOpts, ctx)
	assert.Nil(t, err)

	fixture.Logger.Debugf("Kicking off EthDKG used %v gas", rcpt.GasUsed)
	fixture.Logger.Debugf("registration opens:%v", rcpt.BlockNumber)

	openEvent, err := utils.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, openEvent)

	// get validator addresses
	ctx, cf := context.WithTimeout(context.Background(), 30*time.Second)
	defer cf()
	var validatorAddresses []common.Address
	// all known addresses must be validators at this point
	for _, acc := range eth.GetKnownAccounts() {
		validatorAddresses = append(validatorAddresses, acc.Address)
	}

	// Do Register task
	dkgState, registrationTask, _ := events.UpdateStateOnRegistrationOpened(
		accounts[0],
		openEvent.StartBlock.Uint64(),
		openEvent.PhaseLength.Uint64(),
		openEvent.ConfirmationLength.Uint64(),
		openEvent.Nonce.Uint64(),
		true,
		validatorAddresses,
	)

	dkgDb := GetDKGDb(t)
	err = state.SaveDkgState(dkgDb, dkgState)
	assert.Nil(t, err)
	err = registrationTask.Initialize(ctx, nil, dkgDb, fixture.Logger, eth, "RegistrationTask", "task-id", nil)
	assert.Nil(t, err)
	err = registrationTask.Prepare(ctx)

	expiredCtx, cf := context.WithCancel(context.Background())
	cf()
	// send an expired ctx and never do the tx
	_, err = registrationTask.Execute(expiredCtx)
	assert.NotNil(t, err)

	retry, err := registrationTask.ShouldExecute(ctx)
	assert.Nil(t, err)
	assert.True(t, retry)

	tests.MineFinalityDelayBlocks(eth)
	retry, err = registrationTask.ShouldExecute(ctx)
	assert.Nil(t, err)
	assert.True(t, retry)
}
