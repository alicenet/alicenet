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

func TestDoTaskSuccessNoAccusableParticipants(t *testing.T) {
	suite := NewTestSuite(t, 5, 100)
	defer suite.eth.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accounts := suite.eth.GetKnownAccounts()
	tasks := make([]*dkgtasks.DisputeMissingRegistrationTask, len(suite.dkgStates))
	for idx := 0; idx < len(suite.dkgStates); idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// Set Registration success to true
		//suite.dkgStates[idx].Phase = objects.RegistrationOpen
		suite.dkgStates[idx].PhaseStart = 20
		//suite.dkgStates[idx].PhaseLength = event.PhaseLength.Uint64()
		//suite.dkgStates[idx].ConfirmationLength = event.ConfirmationLength.Uint64()
		//suite.dkgStates[idx].Nonce = event.Nonce.Uint64()
		tasks[idx] = dkgtasks.NewDisputeMissingRegistrationTask(state, suite.dkgStates[idx].PhaseStart, suite.dkgStates[idx].PhaseStart+suite.dkgStates[idx].PhaseLength)
		err := tasks[idx].Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, tasks[idx].Success)
	}
}

func NewTestSuite(t *testing.T, n int, phaseLength uint16) *TestSuite {
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
		dkgStates[idx].Phase = objects.RegistrationOpen
		dkgStates[idx].PhaseStart = event.StartBlock.Uint64()
		dkgStates[idx].PhaseLength = event.PhaseLength.Uint64()
		dkgStates[idx].ConfirmationLength = event.ConfirmationLength.Uint64()
		dkgStates[idx].Nonce = event.Nonce.Uint64()
		tasks[idx] = dkgtasks.NewRegisterTask(state, event.StartBlock.Uint64(), event.StartBlock.Uint64()+event.PhaseLength.Uint64())
		err = tasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)
		err = tasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, tasks[idx].Success)
	}

	return &TestSuite{
		eth:              eth,
		dkgStates:        dkgStates,
		ecdsaPrivateKeys: ecdsaPrivateKeys,
	}
}
