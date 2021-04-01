package dkgtasks_test

import (
	"context"
	"testing"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/blockchain/tasks/dkgtasks"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/common"
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

	// Check status
	callOpts := eth.GetCallOpts(ctx, acct)
	valid, err := eth.Contracts().Participants.IsValidator(callOpts, acct.Address)
	assert.Nil(t, err, "Failed checking validator status")
	assert.True(t, valid)

	// Kick off EthDKG
	txnOpts, err := eth.GetTransactionOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)

	eth.Contracts().Ethdkg.InitializeState(txnOpts)

	// Create a task to register and make sure it succeeds
	task := dkgtasks.NewRegisterTask(acct, pub, 50)

	success := task.DoWork(ctx, logger, eth)
	assert.True(t, success)

	raw, err := tasks.MarshalTask(task)
	assert.Nil(t, err)
	assert.True(t, len(raw) > 0)

	t.Logf("raw:%v", string(raw))

	newTask, err := tasks.UnmarshalTask(raw)
	assert.Nil(t, err)

	t.Logf("newTask:%v", newTask)

	eth.Contracts().Ethdkg.InitializeState(txnOpts)

	success = newTask.DoWork(ctx, logger, eth)
	assert.True(t, success)
}
