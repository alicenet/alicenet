package blockchain_test

import (
	"context"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/etest"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/stretchr/testify/assert"
)

func ConnectSimulator(t *testing.T, numberAccounts int, mineInterval time.Duration) interfaces.Ethereum {
	privKeys := etest.SetupPrivateKeys(numberAccounts)
	eth, err := blockchain.NewEthereumSimulator(
		privKeys,
		1,
		time.Second*2,
		time.Second*5,
		0,
		big.NewInt(math.MaxInt64))
	assert.Nil(t, err, "Failed to build Ethereum endpoint...")
	assert.True(t, eth.IsEthereumAccessible(), "Web3 endpoint is not available.")

	go func() {
		for {
			time.Sleep(mineInterval)
			eth.Commit()
		}
	}()

	return eth
}

func TestFoo(t *testing.T) {

	n := 2

	eth := ConnectSimulator(t, n, 100*time.Millisecond)
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
