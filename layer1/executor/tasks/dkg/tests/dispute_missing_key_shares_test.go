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

func TestDisputeMissingKeySharesTask_FourUnsubmittedKeyShare_DoWork_Success(t *testing.T) {
	n := 5
	m := 4
	fixture := setupEthereum(t, n)
	suite := StartFromKeyShareSubmissionPhase(t, fixture, m, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	// Do dispute missing key share task
	var receiptResponses []transaction.ReceiptResponse
	for idx := 0; idx < n; idx++ {
		disputeMissingKeyshareTask := suite.DisputeMissingKeyshareTasks[idx]

		err := disputeMissingKeyshareTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "DisputeMissingKeyshareTask", "task-id", nil)
		assert.Nil(t, err)

		err = disputeMissingKeyshareTask.Prepare(ctx)
		assert.Nil(t, err)

		shouldExecute, err := disputeMissingKeyshareTask.ShouldExecute(ctx)
		assert.Nil(t, err)
		assert.True(t, shouldExecute)

		txn, taskErr := disputeMissingKeyshareTask.Execute(ctx)
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
	assert.Equal(t, int64(m), badParticipants.Int64())
}

func TestDisputeMissingKeySharesTask_NoUnSubmittedKeyShare(t *testing.T) {
	n := 5
	m := 0
	fixture := setupEthereum(t, n)
	suite := StartFromKeyShareSubmissionPhase(t, fixture, m, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	// Do dispute missing key share task
	for idx := 0; idx < n; idx++ {
		disputeMissingKeyshareTask := suite.DisputeMissingKeyshareTasks[idx]
		err := disputeMissingKeyshareTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "disputeMissingKeyshareTask", "task-id", nil)
		assert.Nil(t, err)
		err = disputeMissingKeyshareTask.Prepare(ctx)
		assert.Nil(t, err)
		txn, taskErr := disputeMissingKeyshareTask.Execute(ctx)
		// Will return error cause phase is in DisputeShareDistribution
		assert.NotNil(t, taskErr)
		assert.Nil(t, txn)
	}

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := ethereum.GetContracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(m), badParticipants.Int64())
}
