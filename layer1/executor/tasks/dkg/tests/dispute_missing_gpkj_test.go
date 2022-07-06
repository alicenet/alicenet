//go:build integration

package tests

import (
	"context"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/layer1/transaction"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisputeMissingGPKjTask_Group_1_FourUnsubmittedGPKj_DoWork_Success(t *testing.T) {
	n := 10
	m := []int{3, 4, 7, 8}
	fixture := setupEthereum(t, n)
	suite := StartFromGPKjPhase(t, fixture, m, []int{}, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	// Do dispute missing gpkj task
	var receiptResponses []transaction.ReceiptResponse
	for idx := 0; idx < n; idx++ {
		disputeMissingGPKjTask := suite.DisputeMissingGPKjTasks[idx]

		err := disputeMissingGPKjTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "DisputeMissingGPKjTask", "task-id", nil)
		assert.Nil(t, err)
		err = disputeMissingGPKjTask.Prepare(ctx)
		assert.Nil(t, err)

		shouldExecute, err := disputeMissingGPKjTask.ShouldExecute(ctx)
		assert.Nil(t, err)
		assert.True(t, shouldExecute)

		txn, taskErr := disputeMissingGPKjTask.Execute(ctx)
		if idx == 0 {
			assert.Nil(t, taskErr)
			assert.NotNil(t, txn)
			rcptResponse, subsErr := fixture.Watcher.Subscribe(ctx, txn, nil)
			assert.Nil(t, subsErr)
			receiptResponses = append(receiptResponses, rcptResponse)
		} else {
			assert.NotNil(t, taskErr)
			assert.True(t, taskErr.IsRecoverable())
			assert.Nil(t, txn)
		}
	}

	tests.WaitGroupReceipts(t, suite.Eth, receiptResponses)

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := ethereum.GetContracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, len(m), int(badParticipants.Int64()))

	CheckBadValidators(t, m, suite)
}
