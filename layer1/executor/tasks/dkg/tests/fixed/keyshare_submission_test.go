//go:build integration

package fixed

import (
	"context"
	"math/big"
	"testing"

	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"

	"github.com/stretchr/testify/assert"
)

// We test to ensure that everything behaves correctly.
func TestKeyShareSubmission_GoodAllValid(t *testing.T) {
	n := 5
	fixture := setupEthereum(t, n)
	suite := StartFromKeyShareSubmissionPhase(t, fixture, 0, 100)
	accounts := suite.Eth.GetKnownAccounts()
	ctx := context.Background()
	callOpts, err := suite.Eth.GetCallOpts(ctx, accounts[0])
	assert.Nil(t, err)

	// Check key shares are present and valid
	for idx, acct := range accounts {

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)

		p, err := ethereum.GetContracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		// check points
		keyShareG1 := dkgState.Participants[acct.Address].KeyShareG1s
		if (keyShareG1[0].Cmp(p.KeyShares[0]) != 0) || (keyShareG1[1].Cmp(p.KeyShares[1]) != 0) {
			t.Fatal("Invalid key share")
		}
	}

	// assert that ETHDKG is at MPKSubmission phase
	phase, err := ethereum.GetContracts().Ethdkg().GetETHDKGPhase(callOpts)
	assert.Nil(t, err)

	assert.Equal(t, uint8(state.MPKSubmission), phase)
}

// We raise an error with invalid inputs.
// This comes from invalid SecretValue in state.
// In practice, this should never arise, though.
func TestKeyShareSubmission_Bad3(t *testing.T) {
	n := 5
	var phaseLength uint16 = 100
	fixture := setupEthereum(t, n)
	suite := StartFromShareDistributionPhase(t, fixture, []int{}, []int{}, phaseLength)
	ctx := context.Background()

	// Do key share submission task
	for idx := 0; idx < n; idx++ {
		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)
		// Mess up SecretValue
		dkgState.SecretValue = nil
		state.SaveDkgState(suite.DKGStatesDbs[idx], dkgState)
		keyshareSubmissionTask := suite.KeyshareSubmissionTasks[idx]

		keyshareSubmissionTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "KeyShareSubmissionTask", "3333", nil)
		err = keyshareSubmissionTask.Prepare(ctx)
		assert.NotNil(t, err)
	}
}

// We raise an error with invalid inputs.
// Here, we mess up KeyShare information before submission
// so that we raise an error on submission.
func TestKeyShareSubmission_Bad4(t *testing.T) {
	n := 5
	var phaseLength uint16 = 100
	fixture := setupEthereum(t, n)
	suite := StartFromShareDistributionPhase(t, fixture, []int{}, []int{}, phaseLength)
	defer suite.Eth.Close()
	ctx := context.Background()

	// Do key share submission task
	for idx := 0; idx < n; idx++ {
		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)
		keyshareSubmissionTask := suite.KeyshareSubmissionTasks[idx]

		keyshareSubmissionTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "KeyShareSubmissionTask", "3333", nil)
		err = keyshareSubmissionTask.Prepare(ctx)
		assert.Nil(t, err)

		// mess uo here
		dkgState.Participants[dkgState.Account.Address].KeyShareG1s = [2]*big.Int{big.NewInt(0), big.NewInt(1)}
		state.SaveDkgState(suite.DKGStatesDbs[idx], dkgState)

		_, err = keyshareSubmissionTask.Execute(ctx)
		assert.NotNil(t, err)

	}

}
