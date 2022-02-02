package dkgtasks_test

import (
	"context"
	"crypto/ecdsa"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/dkg/dtest"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/bridge/bindings"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type TestSuite struct {
	eth              interfaces.Ethereum
	dkgStates        []*objects.DkgState
	ecdsaPrivateKeys []*ecdsa.PrivateKey
}

func TestDoTaskSuccessOneParticipantAccused(t *testing.T) {
	suite := NewTestSuite(t, 5, 1, 100)
	defer suite.eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.eth.GetKnownAccounts()
	tasks := make([]*dkgtasks.DisputeMissingRegistrationTask, len(suite.dkgStates))
	for idx := 0; idx < len(suite.dkgStates); idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		phaseStart := suite.dkgStates[idx].PhaseStart + suite.dkgStates[idx].PhaseLength
		phaseEnd := phaseStart + suite.dkgStates[idx].PhaseLength
		tasks[idx] = dkgtasks.NewDisputeMissingRegistrationTask(state, phaseStart, phaseEnd)
		err := tasks[idx].Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, tasks[idx].Success)
	}

	callOpts := suite.eth.GetCallOpts(ctx, accounts[0])
	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), badParticipants.Int64())
}

func TestDoTaskSuccessThreeParticipantAccused(t *testing.T) {
	suite := NewTestSuite(t, 5, 3, 100)
	defer suite.eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.eth.GetKnownAccounts()
	tasks := make([]*dkgtasks.DisputeMissingRegistrationTask, len(suite.dkgStates))
	for idx := 0; idx < len(suite.dkgStates); idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		phaseStart := suite.dkgStates[idx].PhaseStart + suite.dkgStates[idx].PhaseLength
		phaseEnd := phaseStart + suite.dkgStates[idx].PhaseLength
		tasks[idx] = dkgtasks.NewDisputeMissingRegistrationTask(state, phaseStart, phaseEnd)
		err := tasks[idx].Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, tasks[idx].Success)
	}

	callOpts := suite.eth.GetCallOpts(ctx, accounts[0])
	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(3), badParticipants.Int64())
}

func TestDoTaskSuccessAllParticipantsAreBad(t *testing.T) {
	suite := NewTestSuite(t, 5, 5, 100)
	defer suite.eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.eth.GetKnownAccounts()
	tasks := make([]*dkgtasks.DisputeMissingRegistrationTask, len(suite.dkgStates))
	for idx := 0; idx < len(suite.dkgStates); idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		phaseStart := suite.dkgStates[idx].PhaseStart + suite.dkgStates[idx].PhaseLength
		phaseEnd := phaseStart + suite.dkgStates[idx].PhaseLength
		tasks[idx] = dkgtasks.NewDisputeMissingRegistrationTask(state, phaseStart, phaseEnd)
		err := tasks[idx].Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, tasks[idx].Success)
	}

	callOpts := suite.eth.GetCallOpts(ctx, accounts[0])
	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(5), badParticipants.Int64())
}

func TestShouldRetryTrue(t *testing.T) {
	suite := NewTestSuite(t, 5, 0, 100)
	defer suite.eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.eth.GetKnownAccounts()
	tasks := make([]*dkgtasks.DisputeMissingRegistrationTask, len(suite.dkgStates))
	for idx := 0; idx < len(suite.dkgStates); idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		phaseStart := suite.dkgStates[idx].PhaseStart + suite.dkgStates[idx].PhaseLength
		phaseEnd := phaseStart + suite.dkgStates[idx].PhaseLength
		tasks[idx] = dkgtasks.NewDisputeMissingRegistrationTask(state, phaseStart, phaseEnd)
		err := tasks[idx].Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, tasks[idx].Success)
	}

	for idx := 0; idx < len(suite.dkgStates); idx++ {
		suite.dkgStates[idx].Nonce++
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		shouldRetry := tasks[idx].ShouldRetry(ctx, logger, suite.eth)
		assert.True(t, shouldRetry)
	}
}

func TestShouldRetryFalse(t *testing.T) {
	suite := NewTestSuite(t, 5, 0, 100)
	defer suite.eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.eth.GetKnownAccounts()
	tasks := make([]*dkgtasks.DisputeMissingRegistrationTask, len(suite.dkgStates))
	for idx := 0; idx < len(suite.dkgStates); idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		phaseStart := suite.dkgStates[idx].PhaseStart + suite.dkgStates[idx].PhaseLength
		phaseEnd := phaseStart + suite.dkgStates[idx].PhaseLength
		tasks[idx] = dkgtasks.NewDisputeMissingRegistrationTask(state, phaseStart, phaseEnd)
		err := tasks[idx].Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, tasks[idx].Success)
	}

	for idx := 0; idx < len(suite.dkgStates); idx++ {
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		shouldRetry := tasks[idx].ShouldRetry(ctx, logger, suite.eth)
		assert.False(t, shouldRetry)
	}
}

func NewTestSuite(t *testing.T, n int, unregisteredValidators int, phaseLength uint16) *TestSuite {
	dkgStates, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(n)

	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)

	ctx := context.Background()

	accounts := eth.GetKnownAccounts()
	owner := accounts[0]
	err := eth.UnlockAccount(owner)
	assert.Nil(t, err)

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	txn, err := eth.Contracts().Ethdkg().SetPhaseLength(ownerOpts, phaseLength)
	assert.Nil(t, err)
	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	txn, err = eth.Contracts().ValidatorPool().InitializeETHDKG(ownerOpts)
	assert.Nil(t, err)

	eth.Commit()
	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	var event *bindings.ETHDKGRegistrationOpened
	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == "0xbda431b9b63510f1398bf33d700e013315bcba905507078a1780f13ea5b354b9" {
			event, err = eth.Contracts().Ethdkg().ParseRegistrationOpened(*log)
			assert.Nil(t, err)
			break
		}
	}
	assert.NotNil(t, event)

	// Do Register task
	tasks := make([]*dkgtasks.RegisterTask, n)
	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// Set Registration success to true
		state.Phase = objects.RegistrationOpen
		state.PhaseStart = event.StartBlock.Uint64()
		state.PhaseLength = event.PhaseLength.Uint64()
		state.ConfirmationLength = event.ConfirmationLength.Uint64()
		state.Nonce = event.Nonce.Uint64()
		tasks[idx] = dkgtasks.NewRegisterTask(state, event.StartBlock.Uint64(), event.StartBlock.Uint64()+event.PhaseLength.Uint64())
		err = tasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)

		if idx >= n-unregisteredValidators {
			continue
		}

		err = tasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, tasks[idx].Success)

		for i := 0; i < n; i++ {
			dkgStates[i].Participants[state.Account.Address] = &objects.Participant{
				Address:   state.Account.Address,
				Index:     idx + 1,
				PublicKey: state.TransportPublicKey,
				Phase:     uint8(objects.RegistrationOpen),
				Nonce:     state.Nonce,
			}
		}
	}

	advanceTo(t, eth, event.StartBlock.Uint64()+event.PhaseLength.Uint64())

	return &TestSuite{
		eth:              eth,
		dkgStates:        dkgStates,
		ecdsaPrivateKeys: ecdsaPrivateKeys,
	}
}
