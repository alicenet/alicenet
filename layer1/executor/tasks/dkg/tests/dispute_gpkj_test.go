//go:build integration

package tests

import (
	"context"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/stretchr/testify/assert"
	"testing"
)

// We test to ensure that everything behaves correctly.
func TestGPKjDispute_NoBadGPKj(t *testing.T) {
	n := 6
	b := []int{}
	fixture := setupEthereum(t, n)
	suite := StartFromGPKjPhase(t, fixture, []int{}, b, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	// Do dispute bad gpkj task
	var receiptResponses []transaction.ReceiptResponse
	for idx := 0; idx < n; idx++ {
		for j := 0; j < n; j++ {
			disputeBadGPKjTask := suite.DisputeGPKjTasks[idx][j]

			err := disputeBadGPKjTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "DisputeBadGPKjTask", "task-id", nil)
			assert.Nil(t, err)
			err = disputeBadGPKjTask.Prepare(ctx)
			assert.Nil(t, err)

			shouldExecute, err := disputeBadGPKjTask.ShouldExecute(ctx)
			assert.Nil(t, err)
			assert.True(t, shouldExecute)
			txn, taskErr := disputeBadGPKjTask.Execute(ctx)
			if suite.BadAddresses[disputeBadGPKjTask.Address] {
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
			} else {
				assert.Nil(t, taskErr)
				assert.Nil(t, txn)
			}
		}
	}

	tests.WaitGroupReceipts(t, suite.Eth, receiptResponses)

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := ethereum.GetContracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, len(b), int(badParticipants.Int64()))
}

// Here, we have two malicious gpkj submission.
func TestGPKjDispute_TwoInvalid(t *testing.T) {
	n := 6
	b := []int{3, 4}
	fixture := setupEthereum(t, n)
	suite := StartFromGPKjPhase(t, fixture, []int{}, b, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	// Do dispute bad gpkj task
	var receiptResponses []transaction.ReceiptResponse
	for idx := 0; idx < n; idx++ {
		for j := 0; j < n; j++ {
			disputeBadGPKjTask := suite.DisputeGPKjTasks[idx][j]

			err := disputeBadGPKjTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "DisputeBadGPKjTask", "task-id", nil)
			assert.Nil(t, err)
			err = disputeBadGPKjTask.Prepare(ctx)
			assert.Nil(t, err)

			shouldExecute, err := disputeBadGPKjTask.ShouldExecute(ctx)
			assert.Nil(t, err)
			assert.True(t, shouldExecute)
			txn, taskErr := disputeBadGPKjTask.Execute(ctx)
			if suite.BadAddresses[disputeBadGPKjTask.Address] {
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
			} else {
				assert.Nil(t, taskErr)
				assert.Nil(t, txn)
			}
		}
	}

	tests.WaitGroupReceipts(t, suite.Eth, receiptResponses)

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := ethereum.GetContracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, len(b), int(badParticipants.Int64()))

	CheckBadValidators(t, b, suite)
}

// Here, we have two malicious gpkj submission.
func TestGPKjDispute_FiveInvalid(t *testing.T) {
	n := 6
	b := []int{1, 2, 3, 4, 5}
	fixture := setupEthereum(t, n)
	suite := StartFromGPKjPhase(t, fixture, []int{}, b, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	// Do dispute bad gpkj task
	var receiptResponses []transaction.ReceiptResponse
	for idx := 0; idx < n; idx++ {
		for j := 0; j < n; j++ {
			disputeBadGPKjTask := suite.DisputeGPKjTasks[idx][j]

			err := disputeBadGPKjTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "DisputeBadGPKjTask", "task-id", nil)
			assert.Nil(t, err)
			err = disputeBadGPKjTask.Prepare(ctx)
			assert.Nil(t, err)

			shouldExecute, err := disputeBadGPKjTask.ShouldExecute(ctx)
			assert.Nil(t, err)
			assert.True(t, shouldExecute)
			txn, taskErr := disputeBadGPKjTask.Execute(ctx)
			if suite.BadAddresses[disputeBadGPKjTask.Address] {
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
			} else {
				assert.Nil(t, taskErr)
				assert.Nil(t, txn)
			}
		}
	}

	tests.WaitGroupReceipts(t, suite.Eth, receiptResponses)

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := ethereum.GetContracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, len(b), int(badParticipants.Int64()))

	CheckBadValidators(t, b, suite)
}
