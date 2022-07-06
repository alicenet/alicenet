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

func TestDisputeMissingRegistrationTask_Group_1_DoTaskSuccessOneParticipantAccused(t *testing.T) {
	n := 4
	d := 1
	fixture := setupEthereum(t, n)
	suite := StartFromRegistrationOpenPhase(t, fixture, d, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	var receiptResponses []transaction.ReceiptResponse
	for idx := 0; idx < n; idx++ {
		err := suite.DispMissingRegTasks[idx].Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "DisputeMissingRegistrationTask", "task-id", nil)
		assert.Nil(t, err)

		err = suite.DispMissingRegTasks[idx].Prepare(ctx)
		assert.Nil(t, err)

		shouldExecute, err := suite.DispMissingRegTasks[idx].ShouldExecute(ctx)
		assert.Nil(t, err)
		assert.True(t, shouldExecute)

		txn, taskErr := suite.DispMissingRegTasks[idx].Execute(ctx)
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
	assert.Equal(t, int64(d), badParticipants.Int64())
}

func TestDisputeMissingRegistrationTask_Group_1_DoTaskSuccessThreeParticipantAccused(t *testing.T) {
	n := 5
	d := 3
	fixture := setupEthereum(t, n)
	suite := StartFromRegistrationOpenPhase(t, fixture, d, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	var receiptResponses []transaction.ReceiptResponse
	for idx := 0; idx < n; idx++ {
		err := suite.DispMissingRegTasks[idx].Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "DisputeMissingRegistrationTask", "task-id", nil)
		assert.Nil(t, err)

		err = suite.DispMissingRegTasks[idx].Prepare(ctx)
		assert.Nil(t, err)

		shouldExecute, err := suite.DispMissingRegTasks[idx].ShouldExecute(ctx)
		assert.Nil(t, err)
		assert.True(t, shouldExecute)

		txn, taskErr := suite.DispMissingRegTasks[idx].Execute(ctx)
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
	assert.Equal(t, int64(d), badParticipants.Int64())
}

func TestDisputeMissingRegistrationTask_Group_1_DoTaskSuccessAllParticipantsAreBad(t *testing.T) {
	n := 5
	d := 5
	fixture := setupEthereum(t, n)
	suite := StartFromRegistrationOpenPhase(t, fixture, d, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	var receiptResponses []transaction.ReceiptResponse
	for idx := 0; idx < n; idx++ {
		err := suite.DispMissingRegTasks[idx].Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "DisputeMissingRegistrationTask", "task-id", nil)
		assert.Nil(t, err)

		err = suite.DispMissingRegTasks[idx].Prepare(ctx)
		assert.Nil(t, err)

		shouldExecute, err := suite.DispMissingRegTasks[idx].ShouldExecute(ctx)
		assert.Nil(t, err)
		assert.True(t, shouldExecute)

		txn, taskErr := suite.DispMissingRegTasks[idx].Execute(ctx)
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
	assert.Equal(t, int64(d), badParticipants.Int64())
}
