package dkgtasks_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/tasks/dkgtasks"
	"github.com/MadBase/MadNet/logging"
	"github.com/stretchr/testify/assert"
)

func TestDisputeTask(t *testing.T) {
	logger := logging.GetLogger("dispute_task")
	eth := connectSimulatorEndpoint(t)
	_, pub, err := dkg.GenerateKeys()
	assert.Nil(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	task := dkgtasks.NewDisputeTask(logger, eth, eth.GetDefaultAccount(), pub, 40, 50)

	assert.True(t, task.DoWork(ctx))

	raw, err := json.Marshal(task)
	assert.Nil(t, err)

	assert.True(t, len(raw) > 0)

	newTask := &dkgtasks.DisputeTask{}

	err = json.Unmarshal(raw, newTask)
	assert.Nil(t, err)

	assert.True(t, newTask.DoWork(ctx))
}
