//go:build integration

package blockchain_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/alicenet/alicenet/blockchain/dkg/dtest"
	"github.com/stretchr/testify/assert"
)

func TestTransferFunds(t *testing.T) {
	n := 4
	ecdsaPrivateKeys, _ := dtest.InitializePrivateKeysAndAccounts(n)
	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 100*time.Millisecond)
	defer eth.Close()

	accounts := eth.GetKnownAccounts()
	assert.Equal(t, n, len(accounts))

	for _, acct := range accounts {
		err := eth.UnlockAccount(acct)
		assert.Nil(t, err)
	}

	owner := accounts[0]
	user := accounts[1]

	ctx, cf := context.WithTimeout(context.Background(), 100*time.Second)
	defer cf()

	amount := big.NewInt(12_345)

	txn, err := eth.TransferEther(owner.Address, user.Address, amount)
	assert.Nil(t, err)

	queue := eth.Queue()

	receipt, err := queue.QueueAndWait(ctx, txn)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	assert.Equal(t, txn.Hash(), receipt.TxHash)
}
