//go:build integration

package dkg_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/alicenet/alicenet/blockchain/testutils"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	dkgState "github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	dkgTestUtils "github.com/alicenet/alicenet/layer1/executor/tasks/dkg/testutils"
	"github.com/alicenet/alicenet/layer1/monitor/events"

	"github.com/alicenet/alicenet/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// We test to ensure that everything behaves correctly.
func TestDisputeShareDistributionTask_Group_1_GoodAllValid(t *testing.T) {
	n := 5
	suite := dkgTestUtils.StartFromShareDistributionPhase(t, n, []int{}, []int{}, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()

	// Confirm number of BadShares is zero
	for idx := 0; idx < n; idx++ {
		state := suite.DKGStates[idx]
		if len(state.BadShares) != 0 {
			t.Fatalf("Idx %v has incorrect number of BadShares", idx)
		}
	}

	// Do Share Dispute task
	for idx := 0; idx < n; idx++ {
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		task := suite.DisputeShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		err = task.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, task.Success)
	}

	// Check number of BadShares and confirm correct values
	for idx := 0; idx < n; idx++ {
		state := suite.DKGStates[idx]
		if len(state.BadShares) != 0 {
			t.Fatalf("Idx %v has incorrect number of BadShares", idx)
		}
	}

	callOptions, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	// assert no bad participants on the ETHDKG contract
	badParticipants, err := suite.Eth.Contracts().Ethdkg().GetBadParticipants(callOptions)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), badParticipants.Uint64())
}

// In this test, we have one validator submit invalid information.
// This causes another validator to submit a dispute against him,
// causing a stake-slashing event.
func TestDisputeShareDistributionTask_Group_1_GoodMaliciousShare(t *testing.T) {
	n := 5
	suite := dkgTestUtils.StartFromRegistrationOpenPhase(t, n, 0, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()

	badShares := 1
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do Share Distribution task, with one of the validators distributing bad shares
	for idx := 0; idx < n; idx++ {
		state := suite.DKGStates[idx]

		task := suite.ShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		if idx >= n-badShares {
			// inject bad shares
			for _, s := range state.Participants[state.Account.Address].EncryptedShares {
				s.Set(big.NewInt(0))
			}
		}

		err = task.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, task.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			err = suite.DKGStates[j].OnSharesDistributed(
				logger,
				state.Account.Address,
				state.Participants[state.Account.Address].EncryptedShares,
				state.Participants[state.Account.Address].Commitments,
			)
			assert.Nil(t, err)
		}
	}

	currentHeight, err := suite.Eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	nextPhaseAt := currentHeight + suite.DKGStates[0].ConfirmationLength
	testutils.AdvanceTo(t, suite.Eth, nextPhaseAt)

	// Do Share Dispute task
	for idx := 0; idx < n; idx++ {
		state := suite.DKGStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		disputeShareDistributionTask, _, _ := events.UpdateStateOnShareDistributionComplete(state, nextPhaseAt)

		err := disputeShareDistributionTask.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		err = disputeShareDistributionTask.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, disputeShareDistributionTask.Success)
	}

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := suite.Eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint64(badShares), badParticipants.Uint64())

	// assert bad participants are not validators anymore, i.e, they were fined and evicted
	for i := 0; i < badShares; i++ {
		idx := n - i - 1
		isValidator, err := suite.Eth.Contracts().ValidatorPool().IsValidator(callOpts, suite.DKGStates[idx].Account.Address)
		assert.Nil(t, err)

		assert.False(t, isValidator)
	}
}

// We begin by submitting invalid information.
// This test is meant to raise an error resulting from an invalid argument
// for the Ethereum interface.
func TestDisputeShareDistributionTask_Group_1_Bad1(t *testing.T) {
	n := 4
	ecdsaPrivateKeys, _ := testutils.InitializePrivateKeysAndAccounts(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a task to share distribution and make sure it succeeds
	state := dkgState.NewDkgState(acct)
	task := dkg.NewDisputeShareDistributionTask(state, state.PhaseStart, state.PhaseStart+state.PhaseLength)
	log := logger.WithField("TaskID", "foo")

	err := task.Initialize(ctx, log, eth)
	assert.NotNil(t, err)
}

// We test to ensure that everything behaves correctly.
// This test is meant to raise an error resulting from an invalid argument
// for the Ethereum interface;
// this should raise an error resulting from not successfully completing
// ShareDistribution phase.
func TestDisputeShareDistributionTask_Group_2_Bad2(t *testing.T) {
	n := 4
	ecdsaPrivateKeys, _ := testutils.InitializePrivateKeysAndAccounts(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := testutils.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	accts := eth.GetKnownAccounts()
	acct := accts[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Do bad Share Dispute task
	// Mess up participant information
	state := dkgState.NewDkgState(acct)
	log := logging.GetLogger("test").WithField("Validator", acct.Address.String())
	task := dkg.NewDisputeShareDistributionTask(state, state.PhaseStart, state.PhaseStart+state.PhaseLength)
	for k := 0; k < len(state.Participants); k++ {
		state.Participants[accts[k].Address] = &dkgState.Participant{}
	}

	err := task.Initialize(ctx, log, eth)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestDisputeShareDistributionTask_Group_2_DoRetry_returnsFalse(t *testing.T) {
	n := 5
	suite := dkgTestUtils.StartFromShareDistributionPhase(t, n, []int{}, []int{}, 100)
	defer suite.Eth.Close()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Confirm number of BadShares is zero
	for idx := 0; idx < n; idx++ {
		state := suite.DKGStates[idx]
		if len(state.BadShares) != 0 {
			t.Fatalf("Idx %v has incorrect number of BadShares", idx)
		}
	}

	// Do Share Dispute task
	for idx := 0; idx < n; idx++ {
		task := suite.DisputeShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		err = task.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, task.Success)
	}

	// Check number of BadShares and confirm correct values
	for idx := 0; idx < n; idx++ {
		state := suite.DKGStates[idx]
		if len(state.BadShares) != 0 {
			t.Fatalf("Idx %v has incorrect number of BadShares", idx)
		}
	}

	// Do Should Retry
	for idx := 0; idx < n; idx++ {
		task := suite.DisputeShareDistTasks[idx]
		shouldRetry := task.ShouldRetry(ctx, logger, suite.Eth)
		assert.False(t, shouldRetry)
	}
}

func TestDisputeShareDistributionTask_Group_2_DoRetry_returnsTrue(t *testing.T) {
	n := 5
	suite := dkgTestUtils.StartFromShareDistributionPhase(t, n, []int{}, []int{}, 100)
	defer suite.Eth.Close()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")
	accounts := suite.Eth.GetKnownAccounts()

	// Confirm number of BadShares is zero
	for idx := 0; idx < n; idx++ {
		state := suite.DKGStates[idx]
		if len(state.BadShares) != 0 {
			t.Fatalf("Idx %v has incorrect number of BadShares", idx)
		}
	}

	// Do Share Dispute task
	for idx := 0; idx < n; idx++ {
		task := suite.DisputeShareDistTasks[idx]

		err := task.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)
		err = task.DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, task.Success)
	}

	suite.Eth.Commit()
	suite.Eth.Commit()
	suite.Eth.Commit()

	// Do Should Retry
	for idx := 0; idx < n; idx++ {
		task := suite.DisputeShareDistTasks[idx]
		taskState, ok := task.State.(*dkgState.DkgState)
		assert.True(t, ok)
		taskState.BadShares = make(map[common.Address]*dkgState.Participant)
		taskState.BadShares[accounts[idx].Address] = &dkgState.Participant{}
		shouldRetry := task.ShouldRetry(ctx, logger, suite.Eth)
		assert.True(t, shouldRetry)
	}
}
