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

func TestDisputeMissingShareDistributionTask_Group_1_ShouldAccuseOneValidatorWhoDidNotDistributeShares(t *testing.T) {
	n := 5
	u := []int{4}
	fixture := setupEthereum(t, n)
	suite := StartFromShareDistributionPhase(t, fixture, u, []int{}, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	var receiptResponses []transaction.ReceiptResponse
	for idx := range accounts {
		task := suite.DisputeMissingShareDistTasks[idx]

		err := task.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "DisputeMissingShareDistributionTask", "task-id", nil)
		assert.Nil(t, err)

		err = task.Prepare(ctx)
		assert.Nil(t, err)

		shouldExecute, err := task.ShouldExecute(ctx)
		assert.Nil(t, err)
		assert.True(t, shouldExecute)

		txn, taskErr := task.Execute(ctx)
		//after the first accusation the ethereum contracts will return that
		//the validator is already accused
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
	assert.Equal(t, len(u), int(badParticipants.Uint64()))

	CheckBadValidators(t, u, suite)
}

func TestDisputeMissingShareDistributionTask_Group_1_ShouldAccuseAllValidatorsWhoDidNotDistributeShares(t *testing.T) {
	n := 5
	u := []int{0, 1, 2, 3, 4}
	fixture := setupEthereum(t, n)
	suite := StartFromShareDistributionPhase(t, fixture, u, []int{}, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	var receiptResponses []transaction.ReceiptResponse
	for idx := range accounts {
		task := suite.DisputeMissingShareDistTasks[idx]

		err := task.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "DisputeMissingShareDistributionTask", "task-id", nil)
		assert.Nil(t, err)

		err = task.Prepare(ctx)
		assert.Nil(t, err)

		shouldExecute, err := task.ShouldExecute(ctx)
		assert.Nil(t, err)
		assert.True(t, shouldExecute)

		txn, taskErr := task.Execute(ctx)
		//after the first accusation the ethereum contracts will return that
		//the validator is already accused
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
	assert.Equal(t, len(u), int(badParticipants.Uint64()))

	CheckBadValidators(t, u, suite)
}

func TestDisputeMissingShareDistributionTask_Group_1_ShouldNotAccuseValidatorsWhoDidDistributeShares(t *testing.T) {
	n := 5
	u := []int{}
	fixture := setupEthereum(t, n)
	suite := StartFromShareDistributionPhase(t, fixture, u, []int{}, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	for idx := range accounts {
		task := suite.DisputeMissingShareDistTasks[idx]

		err := task.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "DisputeMissingShareDistributionTask", "task-id", nil)
		assert.Nil(t, err)

		err = task.Prepare(ctx)
		assert.Nil(t, err)

		shouldExecute, err := task.ShouldExecute(ctx)
		assert.Nil(t, err)
		assert.False(t, shouldExecute)
	}

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := ethereum.GetContracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, len(u), int(badParticipants.Uint64()))

	CheckBadValidators(t, u, suite)
}
