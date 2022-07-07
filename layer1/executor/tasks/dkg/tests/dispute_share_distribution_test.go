//go:build integration

package tests

import (
	"bytes"
	"context"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/utils"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

// We test to ensure that everything behaves correctly.
func TestDisputeShareDistributionTask_Group_1_OneValidatorSubmittingInvalidCredentials(t *testing.T) {
	n := 5
	badId := 4
	b := []int{badId}
	fixture := setupEthereum(t, n)
	suite := StartFromShareDistributionPhase(t, fixture, []int{}, b, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	// Do Share Dispute task
	var receiptResponses []transaction.ReceiptResponse
	for idx := 0; idx < n; idx++ {
		for j := 0; j < n; j++ {
			task := suite.DisputeShareDistTasks[idx][j]

			err := task.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "DisputeShareDistributionTask", "task-id", nil)
			assert.Nil(t, err)

			err = task.Prepare(ctx)
			assert.Nil(t, err)

			shouldExecute, err := task.ShouldExecute(ctx)
			assert.Nil(t, err)
			assert.True(t, shouldExecute)

			dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
			assert.NotNil(t, dkgState)
			txn, taskErr := task.Execute(ctx)

			badAddress := false
			if suite.BadAddresses[task.Address] {
				isValidator, err := utils.IsValidator(suite.DKGStatesDbs[idx], task.GetLogger(), task.Address)
				assert.Nil(t, err)
				if idx == 0 {
					assert.Nil(t, taskErr)
					assert.NotNil(t, txn)
					rcptResponse, subsErr := fixture.Watcher.Subscribe(ctx, txn, nil)
					assert.Nil(t, subsErr)
					receiptResponses = append(receiptResponses, rcptResponse)
				} else {
					if isValidator {
						var participantsList = dkgState.GetSortedParticipants()
						if bytes.Equal(task.Address.Bytes(), participantsList[idx].Address.Bytes()) {
							assert.Nil(t, taskErr)
							assert.Nil(t, txn)
						} else {
							assert.NotNil(t, taskErr)
							assert.Nil(t, txn)
							assert.True(t, taskErr.IsRecoverable())
						}
					} else {
						assert.Nil(t, taskErr)
						assert.Nil(t, txn)
					}
				}
				badAddress = true
			}

			if !badAddress {
				assert.Nil(t, taskErr)
				assert.Nil(t, txn)
			}
		}
	}

	tests.WaitGroupReceipts(t, suite.Eth, receiptResponses)

	callOptions, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	// assert no bad participants on the ETHDKG contract
	badParticipants, err := ethereum.GetContracts().Ethdkg().GetBadParticipants(callOptions)
	assert.Nil(t, err)
	assert.Equal(t, len(b), int(badParticipants.Uint64()))

	CheckBadValidators(t, b, suite)
}

// We force an error.
// This is caused by submitting invalid state information (state is nil).
func TestDisputeShareDistributionTask_Group_1_Bad2(t *testing.T) {
	task := dkg.NewDisputeShareDistributionTask(1, 100, common.Address{})
	db := mocks.NewTestDB()
	log := logging.GetLogger("test").WithField("test", "test")

	err := task.Initialize(context.Background(), nil, db, log, nil, "", "", nil)
	assert.Nil(t, err)

	taskErr := task.Prepare(context.Background())
	assert.Nil(t, taskErr)
	txn, taskErr := task.Execute(context.Background())
	assert.Nil(t, txn)
	assert.NotNil(t, taskErr)
	assert.False(t, taskErr.IsRecoverable())
}

func TestDisputeShareDistributionTask_Group_1_TwoValidatorSubmittingInvalidCredentials(t *testing.T) {
	n := 5
	b := []int{2, 4}
	fixture := setupEthereum(t, n)
	suite := StartFromShareDistributionPhase(t, fixture, []int{}, b, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	// Do Share Dispute task
	var receiptResponses []transaction.ReceiptResponse
	for idx := 0; idx < n; idx++ {
		for j := 0; j < n; j++ {
			task := suite.DisputeShareDistTasks[idx][j]

			err := task.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "DisputeShareDistributionTask", "task-id", nil)
			assert.Nil(t, err)

			err = task.Prepare(ctx)
			assert.Nil(t, err)

			shouldExecute, err := task.ShouldExecute(ctx)
			assert.Nil(t, err)
			assert.True(t, shouldExecute)

			dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
			assert.NotNil(t, dkgState)
			txn, taskErr := task.Execute(ctx)

			badAddress := false
			if suite.BadAddresses[task.Address] {
				isValidator, err := utils.IsValidator(suite.DKGStatesDbs[idx], task.GetLogger(), task.Address)
				assert.Nil(t, err)
				if idx == 0 {
					assert.Nil(t, taskErr)
					assert.NotNil(t, txn)
					rcptResponse, subsErr := fixture.Watcher.Subscribe(ctx, txn, nil)
					assert.Nil(t, subsErr)
					receiptResponses = append(receiptResponses, rcptResponse)
				} else {
					if isValidator {
						var participantsList = dkgState.GetSortedParticipants()
						if bytes.Equal(task.Address.Bytes(), participantsList[idx].Address.Bytes()) {
							assert.Nil(t, taskErr)
							assert.Nil(t, txn)
						} else {
							assert.NotNil(t, taskErr)
							assert.Nil(t, txn)
							assert.True(t, taskErr.IsRecoverable())
						}
					} else {
						assert.Nil(t, taskErr)
						assert.Nil(t, txn)
					}
				}
				badAddress = true
			}

			if !badAddress {
				assert.Nil(t, taskErr)
				assert.Nil(t, txn)
			}
		}
	}

	tests.WaitGroupReceipts(t, suite.Eth, receiptResponses)

	callOptions, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	// assert no bad participants on the ETHDKG contract
	badParticipants, err := ethereum.GetContracts().Ethdkg().GetBadParticipants(callOptions)
	assert.Nil(t, err)
	assert.Equal(t, len(b), int(badParticipants.Uint64()))

	CheckBadValidators(t, b, suite)
}

func TestDisputeShareDistributionTask_Group_1_AllValidatorSubmittingInvalidCredentials(t *testing.T) {
	n := 5
	b := []int{0, 1, 2, 3, 4}
	fixture := setupEthereum(t, n)
	suite := StartFromShareDistributionPhase(t, fixture, []int{}, b, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	RegisterPotentialValidatorOnMonitor(t, suite, accounts)

	// Do Share Dispute task
	var receiptResponses []transaction.ReceiptResponse
	for idx := 0; idx < n; idx++ {
		for j := 0; j < n; j++ {
			task := suite.DisputeShareDistTasks[idx][j]

			err := task.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "DisputeShareDistributionTask", "task-id", nil)
			assert.Nil(t, err)

			err = task.Prepare(ctx)
			assert.Nil(t, err)

			shouldExecute, err := task.ShouldExecute(ctx)
			assert.Nil(t, err)
			assert.True(t, shouldExecute)

			dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
			assert.NotNil(t, dkgState)
			txn, taskErr := task.Execute(ctx)

			badAddress := false
			if suite.BadAddresses[task.Address] {
				var participantsList = dkgState.GetSortedParticipants()
				if bytes.Equal(task.Address.Bytes(), participantsList[idx].Address.Bytes()) {
					assert.Nil(t, taskErr)
					assert.Nil(t, txn)
				} else {
					// During the first run the validator will accuse the rest of the participants but not himself
					if idx == 0 {
						assert.Nil(t, taskErr)
						assert.NotNil(t, txn)
						rcptResponse, subsErr := fixture.Watcher.Subscribe(ctx, txn, nil)
						assert.Nil(t, subsErr)
						receiptResponses = append(receiptResponses, rcptResponse)
					} else {
						isValidator, err := utils.IsValidator(suite.DKGStatesDbs[idx], task.GetLogger(), task.Address)
						assert.Nil(t, err)
						if isValidator {
							assert.NotNil(t, taskErr)
							assert.Nil(t, txn)
							assert.True(t, taskErr.IsRecoverable())
						} else {
							assert.Nil(t, taskErr)
							assert.Nil(t, txn)
						}
					}
				}
				badAddress = true
			}

			if !badAddress {
				assert.Nil(t, taskErr)
				assert.Nil(t, txn)
			}
		}
	}

	tests.WaitGroupReceipts(t, suite.Eth, receiptResponses)

	callOptions, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	// assert no bad participants on the ETHDKG contract
	badParticipants, err := ethereum.GetContracts().Ethdkg().GetBadParticipants(callOptions)
	assert.Nil(t, err)
	// The -1 is because during the first run the validator will accuse all the participants but himself
	// the second the rest of participants are not longer validators so they cannot accuse the first one
	assert.Equal(t, len(b)-1, int(badParticipants.Uint64()))
}
