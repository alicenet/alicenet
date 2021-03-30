package dkgtasks_test

import (
	"context"
	"testing"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/blockchain/tasks/dkgtasks"
	"github.com/MadBase/MadNet/logging"
	"github.com/stretchr/testify/assert"
)

func TestDisputeTask(t *testing.T) {

	tasks.RegisterTask(&dkgtasks.DisputeTask{})

	logger := logging.GetLogger("dispute_task")
	eth := connectSimulatorEndpoint(t)
	_, pub, err := dkg.GenerateKeys()
	assert.Nil(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	task := dkgtasks.NewDisputeTask(pub, 40, 50)

	assert.True(t, task.DoWork(ctx, logger, eth))

	raw, err := tasks.MarshalTask(task)
	assert.Nil(t, err)
	assert.True(t, len(raw) > 0)

	t.Logf("raw:%v", string(raw))

	newTask, err := tasks.UnmarshalTask(raw)
	assert.Nil(t, err)

	t.Logf("newTask:%v", newTask)

	assert.True(t, newTask.DoWork(ctx, logger, eth))
}
