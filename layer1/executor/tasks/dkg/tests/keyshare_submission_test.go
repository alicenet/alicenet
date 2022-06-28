//go:build integration

package tests

import (
	"context"
	"math/big"
	"testing"

	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/testutils"

	"github.com/alicenet/alicenet/logging"
	"github.com/stretchr/testify/assert"
)

// We test to ensure that everything behaves correctly.
func TestKeyShareSubmission_GoodAllValid(t *testing.T) {
	n := 5
	suite := testutils.StartFromKeyShareSubmissionPhase(t, n, 0, 100)
	defer suite.Eth.Close()
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)

	// Check key shares are present and valid
	for idx, acct := range accounts {
		//callOpts := suite.Eth.GetCallOpts(context.Background(), acct)
		p, err := suite.Eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		// check points
		keyShareG1 := suite.DKGStates[idx].Participants[acct.Address].KeyShareG1s
		if (keyShareG1[0].Cmp(p.KeyShares[0]) != 0) || (keyShareG1[1].Cmp(p.KeyShares[1]) != 0) {
			t.Fatal("Invalid key share")
		}
	}

	// assert that ETHDKG is at MPKSubmission phase
	phase, err := suite.Eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	assert.Nil(t, err)

	assert.Equal(t, uint8(state.MPKSubmission), phase)
}

// We raise an error with invalid inputs.
// This comes from invalid SecretValue in state.
// In practice, this should never arise, though.
func TestKeyShareSubmission_Bad3(t *testing.T) {
	n := 5
	var phaseLength uint16 = 100
	suite := testutils.StartFromShareDistributionPhase(t, n, []int{}, []int{}, phaseLength)
	defer suite.Eth.Close()
	//accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do key share submission task
	for idx := 0; idx < n; idx++ {
		state := suite.DKGStates[idx]
		// Mess up SecretValue
		state.SecretValue = nil

		keyshareSubmissionTask := suite.KeyshareSubmissionTasks[idx]

		err := keyshareSubmissionTask.Initialize(ctx, logger, suite.Eth)
		assert.NotNil(t, err)
	}
}

// We raise an error with invalid inputs.
// Here, we mess up KeyShare information before submission
// so that we raise an error on submission.
func TestKeyShareSubmission_Bad4(t *testing.T) {
	n := 5
	var phaseLength uint16 = 100
	suite := testutils.StartFromShareDistributionPhase(t, n, []int{}, []int{}, phaseLength)
	defer suite.Eth.Close()
	//accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do key share submission task
	for idx := 0; idx < n; idx++ {
		state := suite.DKGStates[idx]
		keyshareSubmissionTask := suite.KeyshareSubmissionTasks[idx]

		err := keyshareSubmissionTask.Initialize(ctx, logger, suite.Eth)
		assert.Nil(t, err)

		// mess uo here
		state.Participants[state.Account.Address].KeyShareG1s = [2]*big.Int{big.NewInt(0), big.NewInt(1)}

		err = keyshareSubmissionTask.DoWork(ctx, logger, suite.Eth)
		assert.NotNil(t, err)

	}

}
