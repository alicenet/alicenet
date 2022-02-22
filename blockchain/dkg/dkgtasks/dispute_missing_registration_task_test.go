package dkgtasks_test

import (
	"context"
	"testing"

	"github.com/MadBase/MadNet/logging"
	"github.com/stretchr/testify/assert"
)

func TestDoTaskSuccessOneParticipantAccused(t *testing.T) {
	n := 5
	d := 1
	suite := StartFromRegistrationOpenPhase(t, n, d, 100)
	defer suite.eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.eth.GetKnownAccounts()
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	for idx := 0; idx < len(suite.dkgStates); idx++ {
		err := suite.dispMissingRegTasks[idx].Initialize(ctx, logger, suite.eth, suite.dkgStates[idx])
		assert.Nil(t, err)

		err = suite.dispMissingRegTasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, suite.dispMissingRegTasks[idx].Success)
	}

	callOpts := suite.eth.GetCallOpts(ctx, accounts[0])
	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(d), badParticipants.Int64())
}

func TestDoTaskSuccessThreeParticipantAccused(t *testing.T) {
	n := 5
	d := 3
	suite := StartFromRegistrationOpenPhase(t, n, d, 100)
	defer suite.eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.eth.GetKnownAccounts()
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	for idx := 0; idx < len(suite.dkgStates); idx++ {
		err := suite.dispMissingRegTasks[idx].Initialize(ctx, logger, suite.eth, suite.dkgStates[idx])
		assert.Nil(t, err)

		err = suite.dispMissingRegTasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, suite.dispMissingRegTasks[idx].Success)
	}

	callOpts := suite.eth.GetCallOpts(ctx, accounts[0])
	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(d), badParticipants.Int64())
}

func TestDoTaskSuccessAllParticipantsAreBad(t *testing.T) {
	n := 5
	d := 5
	suite := StartFromRegistrationOpenPhase(t, n, d, 100)
	defer suite.eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.eth.GetKnownAccounts()
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	for idx := 0; idx < len(suite.dkgStates); idx++ {
		err := suite.dispMissingRegTasks[idx].Initialize(ctx, logger, suite.eth, suite.dkgStates[idx])
		assert.Nil(t, err)

		err = suite.dispMissingRegTasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, suite.dispMissingRegTasks[idx].Success)
	}

	callOpts := suite.eth.GetCallOpts(ctx, accounts[0])
	badParticipants, err := suite.eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(d), badParticipants.Int64())
}

func TestShouldRetryTrue(t *testing.T) {
	suite := StartFromRegistrationOpenPhase(t, 5, 1, 100)
	defer suite.eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.eth.GetKnownAccounts()
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	for idx := 0; idx < len(suite.dkgStates); idx++ {
		err := suite.dispMissingRegTasks[idx].Initialize(ctx, logger, suite.eth, suite.dkgStates[idx])
		assert.Nil(t, err)

		shouldRetry := suite.dispMissingRegTasks[idx].ShouldRetry(ctx, logger, suite.eth)
		assert.True(t, shouldRetry)
	}
}

func TestShouldNotRetryAfterSuccessfullyAccusingAllMissingParticipants(t *testing.T) {
	suite := StartFromRegistrationOpenPhase(t, 5, 0, 100)
	defer suite.eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.eth.GetKnownAccounts()
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())
	for idx := 0; idx < len(suite.dkgStates); idx++ {
		err := suite.dispMissingRegTasks[idx].Initialize(ctx, logger, suite.eth, suite.dkgStates[idx])
		assert.Nil(t, err)

		err = suite.dispMissingRegTasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, suite.dispMissingRegTasks[idx].Success)
	}

	for idx := 0; idx < len(suite.dkgStates); idx++ {
		shouldRetry := suite.dispMissingRegTasks[idx].ShouldRetry(ctx, logger, suite.eth)
		assert.False(t, shouldRetry)
	}
}
