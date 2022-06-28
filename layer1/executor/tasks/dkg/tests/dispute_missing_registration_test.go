//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/testutils"
	"github.com/alicenet/alicenet/logging"
	"github.com/stretchr/testify/assert"
)

func TestDisputeMissingRegistrationTask_Group_1_DoTaskSuccessOneParticipantAccused(t *testing.T) {
	n := 4
	d := 1
	suite := testutils.StartFromRegistrationOpenPhase(t, n, d, 100)
	defer suite.Eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.Eth.GetKnownAccounts()
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	for idx := 0; idx < len(suite.DKGStates); idx++ {
		err := suite.DispMissingRegTasks[idx].Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		err = suite.DispMissingRegTasks[idx].DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, suite.DispMissingRegTasks[idx].Success)
	}

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := suite.Eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(d), badParticipants.Int64())
}

func TestDisputeMissingRegistrationTask_Group_1_DoTaskSuccessThreeParticipantAccused(t *testing.T) {
	n := 5
	d := 3
	suite := testutils.StartFromRegistrationOpenPhase(t, n, d, 100)
	defer suite.Eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.Eth.GetKnownAccounts()
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	for idx := 0; idx < len(suite.DKGStates); idx++ {
		err := suite.DispMissingRegTasks[idx].Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		err = suite.DispMissingRegTasks[idx].DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, suite.DispMissingRegTasks[idx].Success)
	}

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := suite.Eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(d), badParticipants.Int64())
}

func TestDisputeMissingRegistrationTask_Group_1_DoTaskSuccessAllParticipantsAreBad(t *testing.T) {
	n := 5
	d := 5
	suite := testutils.StartFromRegistrationOpenPhase(t, n, d, 100)
	defer suite.Eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.Eth.GetKnownAccounts()
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	for idx := 0; idx < len(suite.DKGStates); idx++ {
		err := suite.DispMissingRegTasks[idx].Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		err = suite.DispMissingRegTasks[idx].DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, suite.DispMissingRegTasks[idx].Success)
	}

	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)
	badParticipants, err := suite.Eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, int64(d), badParticipants.Int64())
}

func TestDisputeMissingRegistrationTask_Group_2_ShouldRetryTrue(t *testing.T) {
	suite := testutils.StartFromRegistrationOpenPhase(t, 5, 1, 100)
	defer suite.Eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.Eth.GetKnownAccounts()
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())

	for idx := 0; idx < len(suite.DKGStates); idx++ {
		err := suite.DispMissingRegTasks[idx].Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		shouldRetry := suite.DispMissingRegTasks[idx].ShouldRetry(ctx, logger, suite.Eth)
		assert.True(t, shouldRetry)
	}
}

func TestDisputeMissingRegistrationTask_Group_2_ShouldNotRetryAfterSuccessfullyAccusingAllMissingParticipants(t *testing.T) {
	suite := testutils.StartFromRegistrationOpenPhase(t, 5, 0, 100)
	defer suite.Eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.Eth.GetKnownAccounts()
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())
	for idx := 0; idx < len(suite.DKGStates); idx++ {
		err := suite.DispMissingRegTasks[idx].Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		err = suite.DispMissingRegTasks[idx].DoWork(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		suite.Eth.Commit()
		assert.True(t, suite.DispMissingRegTasks[idx].Success)
	}

	for idx := 0; idx < len(suite.DKGStates); idx++ {
		shouldRetry := suite.DispMissingRegTasks[idx].ShouldRetry(ctx, logger, suite.Eth)
		assert.False(t, shouldRetry)
	}
}
