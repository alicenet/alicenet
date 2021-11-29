package blockchain_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {

	eth, err := setupEthereum(t, 4)
	assert.Nil(t, err)
	defer eth.Close()

	c := eth.Contracts()
	ctx := context.TODO()

	txnOpts, err := eth.GetTransactionOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err, "Failed to set txn options.")

	_, err = c.Registry().Register(txnOpts, "myself", c.RegistryAddress())
	assert.Nil(t, err, "Failed to create registry entry")
	eth.Commit()

	callOpts := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	addr, err := c.Registry().Lookup(callOpts, "myself")
	assert.Nil(t, err, "Failed to lookup of name from registry")

	assert.Equal(t, c.RegistryAddress(), addr)
}
