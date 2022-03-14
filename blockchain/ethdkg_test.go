package blockchain_test

import (
	"testing"
)

func TestAccuse(t *testing.T) {
	// eth, commit, err := setupEthereum()

}

/* func TestRegistration(t *testing.T) {
	eth, err := setupEthereum(t, 4)
	if err != nil {
		t.Fatal(err)
	}
	c := eth.Contracts()
	ctx := context.TODO()

	callOpts := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	submitMPKAddress, err := c.ContractFactory().Lookup(callOpts, "ethdkgSubmitMPK/v1")
	assert.Nil(t, err)
	t.Logf("submitMPKAddress:%v", submitMPKAddress.Hex())
}
*/
