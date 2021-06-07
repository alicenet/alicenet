package dkgtasks_test

import (
	"testing"
)

func TestDisputeTask(t *testing.T) {

	// var accountAddresses []string = []string{
	// 	"0x546F99F244b7B58B855330AE0E2BC1b30b41302F", "0x9AC1c9afBAec85278679fF75Ef109217f26b1417",
	// 	"0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac", "0x615695C4a4D6a60830e5fca4901FbA099DF26271",
	// 	"0x63a6627b79813A7A43829490C4cE409254f64177"}

	// tasks.RegisterTask(&dkgtasks.DisputeTask{})

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// logger := logging.GetLogger("dispute_task")
	// eth := connectSimulatorEndpoint(t, accountAddresses)

	// _, pub, err := dkg.GenerateKeys()
	// assert.Nil(t, err)

	// task := dkgtasks.NewDisputeTask(eth.GetDefaultAccount(), pub, 40, 50)

	// assert.True(t, task.DoWork(ctx, logger, eth))

	// raw, err := tasks.MarshalTask(task)
	// assert.Nil(t, err)
	// assert.True(t, len(raw) > 0)

	// t.Logf("raw:%v", string(raw))

	// newTask, err := tasks.UnmarshalTask(raw)
	// assert.Nil(t, err)

	// t.Logf("newTask:%v", newTask)

	// assert.True(t, newTask.DoWork(ctx, logger, eth))
}
