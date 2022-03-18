package blockchain_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/dkg/dtest"
	"github.com/stretchr/testify/assert"
)

func TestFoo(t *testing.T) {

	n := 4
	ecdsaPrivateKeys, _ := dtest.InitializePrivateKeysAndAccounts(n)
	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 100*time.Millisecond)
	defer eth.Close()

	accounts := eth.GetKnownAccounts()
	assert.Equal(t, n, len(accounts))

	for _, acct := range accounts {
		eth.UnlockAccount(acct)
	}

	owner := accounts[0]
	user := accounts[1]

	ctx, cf := context.WithTimeout(context.Background(), 100*time.Second)
	defer cf()

	amount := big.NewInt(12_345)

	txn, err := eth.TransferEther(owner.Address, user.Address, amount)
	assert.Nil(t, err)

	queue := eth.Queue()

	queue.QueueAndWait(ctx, txn)
}
