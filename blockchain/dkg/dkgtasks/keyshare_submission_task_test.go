package dkgtasks_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/logging"
	"github.com/stretchr/testify/assert"
)

// We test to ensure that everything behaves correctly.
func TestKeyShareSubmissionGoodAllValid(t *testing.T) {
	n := 5
	suite := StartFromKeyShareSubmissionPhase(t, n, 0, 100)
	defer suite.eth.Close()
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	// currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	// assert.Nil(t, err)

	// todo: review from here

	callOpts := suite.eth.GetCallOpts(ctx, accounts[0])
	// validators, err := suite.eth.Contracts().ValidatorPool().GetValidatorAddresses(callOpts)
	// assert.Nil(t, err)

	// Check key shares are present and valid
	for idx, acct := range accounts {
		//callOpts := suite.eth.GetCallOpts(context.Background(), acct)
		p, err := suite.eth.Contracts().Ethdkg().GetParticipantInternalState(callOpts, acct.Address)
		assert.Nil(t, err)

		// check points
		keyShareG1 := suite.dkgStates[idx].Participants[acct.Address].KeyShareG1s
		if (keyShareG1[0].Cmp(p.KeyShares[0]) != 0) || (keyShareG1[1].Cmp(p.KeyShares[1]) != 0) {
			t.Fatal("Invalid key share")
		}
	}

	// assert that ETHDKG is at MPKSubmission phase
	phase, err := suite.eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	assert.Nil(t, err)

	assert.Equal(t, uint8(objects.MPKSubmission), phase)
}

// We raise an error with invalid inputs.
// This comes from invalid SecretValue in state.
// In practice, this should never arise, though.
func TestKeyShareSubmissionBad3(t *testing.T) {
	n := 5
	var phaseLength uint16 = 100
	suite := StartFromShareDistributionPhase(t, n, []int{}, []int{}, phaseLength)
	defer suite.eth.Close()
	//accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do key share submission task
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		// Mess up SecretValue
		state.SecretValue = nil

		keyshareSubmissionTask := suite.keyshareSubmissionTasks[idx]

		err := keyshareSubmissionTask.Initialize(ctx, logger, suite.eth, state)
		assert.NotNil(t, err)
	}
}

// We raise an error with invalid inputs.
// Here, we mess up KeyShare information before submission
// so that we raise an error on submission.
func TestKeyShareSubmissionBad4(t *testing.T) {
	n := 5
	var phaseLength uint16 = 100
	suite := StartFromShareDistributionPhase(t, n, []int{}, []int{}, phaseLength)
	defer suite.eth.Close()
	//accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do key share submission task
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		keyshareSubmissionTask := suite.keyshareSubmissionTasks[idx]

		err := keyshareSubmissionTask.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)

		// mess uo here
		state.Participants[state.Account.Address].KeyShareG1s = [2]*big.Int{big.NewInt(0), big.NewInt(1)}

		err = keyshareSubmissionTask.DoWork(ctx, logger, suite.eth)
		assert.NotNil(t, err)

	}

}
