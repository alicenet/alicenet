package dkgtasks_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/blockchain/tasks/dkgtasks"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/bridge/bindings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestRegisterTask(t *testing.T) {

	tasks.RegisterTask(&dkgtasks.RegisterTask{})

	logger := logging.GetLogger("register_task")
	eth := connectSimulatorEndpoint(t)
	_, pub, err := dkg.GenerateKeys()
	assert.Nil(t, err)

	acct, err := eth.GetAccount(common.HexToAddress("0x9AC1c9afBAec85278679fF75Ef109217f26b1417"))
	assert.Nil(t, err, "Could not find account")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := eth.Contracts()

	// Check status
	callOpts := eth.GetCallOpts(ctx, acct)
	valid, err := c.Participants.IsValidator(callOpts, acct.Address)
	assert.Nil(t, err, "Failed checking validator status")
	assert.True(t, valid)

	// Kick off EthDKG
	txnOpts, err := eth.GetTransactionOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)

	mess, err := c.Ethdkg.InitialMessage(callOpts)
	assert.Nil(t, err)
	assert.NotNil(t, mess)

	// Reuse these
	var (
		txn  *types.Transaction
		rcpt *types.Receipt
	)

	// Shorten ethdkg phase for testing purposes
	txn, err = c.Ethdkg.UpdatePhaseLength(txnOpts, big.NewInt(4))
	eth.Queue().QueueTransaction(ctx, txn)
	assert.Nil(t, err)

	// rcpt, err = eth.WaitForReceipt(ctx, txn)
	rcpt, err = eth.Queue().WaitTransaction(ctx, txn)
	assert.Nil(t, err)
	assert.NotNil(t, rcpt)

	t.Logf("Updating phase length used %v gas vs %v", rcpt.GasUsed, txn.Cost())

	// Kick off ethdkg
	txn, err = c.Ethdkg.InitializeState(txnOpts)
	assert.Nil(t, err)
	eth.Queue().QueueTransaction(ctx, txn)

	rcpt, err = eth.Queue().WaitTransaction(ctx, txn)
	assert.Nil(t, err)
	assert.NotNil(t, rcpt.Logs)

	t.Logf("Kicking off EthDKG used %v gas", rcpt.GasUsed)

	t.Logf("registration opens:%v", rcpt.BlockNumber)

	var openEvent *bindings.ETHDKGRegistrationOpen
	for _, log := range rcpt.Logs {
		eventSelector := log.Topics[0].String()
		if eventSelector == "0x9c6f8368fe7e77e8cb9438744581403bcb3f53298e517f04c1b8475487402e97" {
			openEvent, err = c.Ethdkg.ParseRegistrationOpen(*log)
			assert.Nil(t, err)
		}
	}
	assert.NotNil(t, openEvent)

	// Create a task to register and make sure it succeeds
	task := dkgtasks.NewRegisterTask(acct, pub, openEvent.RegistrationEnds.Uint64())

	t.Logf("registration ends:%v", openEvent.RegistrationEnds.Uint64())

	success := task.DoWork(ctx, logger, eth)
	assert.True(t, success)
}
