package dkgtasks_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestRegisterTask(t *testing.T) {

	var accountAddresses []string = []string{
		"0x546F99F244b7B58B855330AE0E2BC1b30b41302F", "0x9AC1c9afBAec85278679fF75Ef109217f26b1417",
		"0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac", "0x615695C4a4D6a60830e5fca4901FbA099DF26271",
		"0x63a6627b79813A7A43829490C4cE409254f64177"}

	tasks.RegisterTask(&dkgtasks.RegisterTask{})

	logger := logging.GetLogger("register_task")
	eth := connectSimulatorEndpoint(t, accountAddresses)
	defer eth.Close()

	acct, err := eth.GetAccount(common.HexToAddress("0x9AC1c9afBAec85278679fF75Ef109217f26b1417"))
	assert.Nil(t, err, "Could not find account")

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

	eth.Queue().QueueAndWait(ctx, txn)

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

	success := task.DoWork(ctx, logger, eth)
	assert.True(t, success)
}
