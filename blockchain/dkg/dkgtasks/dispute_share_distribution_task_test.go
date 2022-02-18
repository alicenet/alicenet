package dkgtasks_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgevents"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/dkg/dtest"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// We test to ensure that everything behaves correctly.
func TestShareDisputeGoodAllValid(t *testing.T) {
	n := 5
	suite := StartFromShareDistributionPhase(t, n, []int{}, []int{}, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()

	// Confirm number of BadShares is zero
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		if len(state.BadShares) != 0 {
			t.Fatalf("Idx %v has incorrect number of BadShares", idx)
		}
	}

	// Do Share Dispute task
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		task := suite.disputeShareDistTasks[idx]
		err := task.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = task.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, task.Success)
	}

	// Check number of BadShares and confirm correct values
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		if len(state.BadShares) != 0 {
			t.Fatalf("Idx %v has incorrect number of BadShares", idx)
		}
	}

	// assert no bad participants on the ETHDKG contract
	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(suite.eth.GetCallOpts(ctx, accounts[0]))
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), badParticipants.Uint64())
}

// In this test, we have one validator submit invalid information.
// This causes another validator to submit a dispute against him,
// causing a stake-slashing event.
func TestShareDisputeGoodMaliciousShare(t *testing.T) {
	n := 5
	suite := StartFromRegistrationOpenPhase(t, n, 0, 100)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()

	badShares := 1
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do Share Distribution task, with one of the validators distributing bad shares
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]

		task := suite.shareDistTasks[idx]
		err := task.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)

		if idx >= n-badShares {
			// inject bad shares
			for _, s := range state.Participants[state.Account.Address].EncryptedShares {
				s.Set(big.NewInt(0))
			}
		}

		err = task.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, task.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			err = suite.dkgStates[j].OnSharesDistributed(
				logger,
				state.Account.Address,
				state.Participants[state.Account.Address].EncryptedShares,
				state.Participants[state.Account.Address].Commitments,
			)
			assert.Nil(t, err)
		}
	}

	currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	nextPhaseAt := currentHeight + suite.dkgStates[0].ConfirmationLength
	advanceTo(t, suite.eth, nextPhaseAt)

	// Do Share Dispute task
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		disputeShareDistributionTask, _, _, _, _, _, _, _, _ := dkgevents.UpdateStateOnShareDistributionComplete(state, logger, nextPhaseAt)

		err := disputeShareDistributionTask.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = disputeShareDistributionTask.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, disputeShareDistributionTask.Success)
	}

	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(suite.eth.GetCallOpts(ctx, accounts[0]))
	assert.Nil(t, err)
	assert.Equal(t, uint64(badShares), badParticipants.Uint64())

	callOpts := suite.eth.GetCallOpts(ctx, accounts[0])

	// assert bad participants are not validators anymore, i.e, they were fined and evicted
	for i := 0; i < badShares; i++ {
		idx := n - i - 1
		isValidator, err := suite.eth.Contracts().ValidatorPool().IsValidator(callOpts, suite.dkgStates[idx].Account.Address)
		assert.Nil(t, err)

		assert.False(t, isValidator)
	}
}

// We begin by submitting invalid information.
// This test is meant to raise an error resulting from an invalid argument
// for the Ethereum interface.
func TestShareDisputeBad1(t *testing.T) {
	n := 4
	_, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	acct := eth.GetKnownAccounts()[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a task to share distribution and make sure it succeeds
	state := objects.NewDkgState(acct)
	task := dkgtasks.NewDisputeShareDistributionTask(state, state.PhaseStart, state.PhaseStart+state.PhaseLength)
	log := logger.WithField("TaskID", "foo")

	err := task.Initialize(ctx, log, eth, nil)
	assert.NotNil(t, err)
}

// We test to ensure that everything behaves correctly.
// This test is meant to raise an error resulting from an invalid argument
// for the Ethereum interface;
// this should raise an error resulting from not successfully completing
// ShareDistribution phase.
func TestShareDisputeBad2(t *testing.T) {
	n := 4
	_, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(n)
	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)
	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	defer eth.Close()

	accts := eth.GetKnownAccounts()
	acct := accts[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Do bad Share Dispute task
	// Mess up participant information
	state := objects.NewDkgState(acct)
	log := logging.GetLogger("test").WithField("Validator", acct.Address.String())
	task := dkgtasks.NewDisputeShareDistributionTask(state, state.PhaseStart, state.PhaseStart+state.PhaseLength)
	for k := 0; k < len(state.Participants); k++ {
		state.Participants[accts[k].Address] = &objects.Participant{}
	}
	err := task.Initialize(ctx, log, eth, state)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestDisputeShareDistributionTask_DoRetry_returnsFalse(t *testing.T) {
	n := 5
	suite := StartFromShareDistributionPhase(t, n, []int{}, []int{}, 100)
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Confirm number of BadShares is zero
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		if len(state.BadShares) != 0 {
			t.Fatalf("Idx %v has incorrect number of BadShares", idx)
		}
	}

	// Do Share Dispute task
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]

		task := suite.disputeShareDistTasks[idx]
		err := task.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = task.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, task.Success)
	}

	// Check number of BadShares and confirm correct values
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		if len(state.BadShares) != 0 {
			t.Fatalf("Idx %v has incorrect number of BadShares", idx)
		}
	}

	// Do Should Retry
	for idx := 0; idx < n; idx++ {
		task := suite.disputeShareDistTasks[idx]
		shouldRetry := task.ShouldRetry(ctx, logger, suite.eth)
		assert.False(t, shouldRetry)
	}
}

func TestDisputeShareDistributionTask_DoRetry_returnsTrue(t *testing.T) {
	n := 5
	suite := StartFromShareDistributionPhase(t, n, []int{}, []int{}, 100)
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")
	accounts := suite.eth.GetKnownAccounts()

	// Confirm number of BadShares is zero
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		if len(state.BadShares) != 0 {
			t.Fatalf("Idx %v has incorrect number of BadShares", idx)
		}
	}

	// Do Share Dispute task
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]

		task := suite.disputeShareDistTasks[idx]
		err := task.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = task.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, task.Success)
	}

	suite.eth.Commit()
	suite.eth.Commit()
	suite.eth.Commit()

	// Do Should Retry
	for idx := 0; idx < n; idx++ {
		task := suite.disputeShareDistTasks[idx]
		task.State.BadShares = make(map[common.Address]*objects.Participant)
		task.State.BadShares[accounts[idx].Address] = &objects.Participant{}
		shouldRetry := task.ShouldRetry(ctx, logger, suite.eth)
		assert.True(t, shouldRetry)
	}
}
