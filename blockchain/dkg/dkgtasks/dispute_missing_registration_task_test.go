package dkgtasks_test

import (
	"context"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/logging"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoTaskSuccessOneParticipantAccused(t *testing.T) {
	suite := StartAtDistributeSharesPhase(t, 5, 1, 100)
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
	suite := StartAtDistributeSharesPhase(t, 5, 3, 100)
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
	suite := StartAtDistributeSharesPhase(t, 5, 5, 100)
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
	suite := StartAtDistributeSharesPhase(t, 5, 0, 100)
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

func TestShouldNotRetryAfterSuccessfulyAccusingAllMissingParticipants(t *testing.T) {
	suite := StartAtDistributeSharesPhase(t, 5, 0, 100)
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
