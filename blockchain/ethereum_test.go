package blockchain_test

import (
	"context"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/dkg/dtest"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func setupEthereum(t *testing.T, n int) interfaces.Ethereum {
	logging.GetLogger("ethereum").SetLevel(logrus.InfoLevel)

	ecdsaPrivateKeys, _ := dtest.InitializePrivateKeysAndAccounts(n)
	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 1000*time.Millisecond)
	assert.NotNil(t, eth)

	acct := eth.GetDefaultAccount()
	assert.Nil(t, eth.UnlockAccount(acct))

	return eth
}

func TestAccountsFound(t *testing.T) {
	eth := setupEthereum(t, 4)
	defer eth.Close()

	accountList := eth.GetKnownAccounts()

	for _, acct := range accountList {

		err := eth.UnlockAccount(acct)
		assert.Nilf(t, err, "Not able to unlock account: %v", acct.Address)

		_, err = eth.GetAccountKeys(acct.Address)
		assert.Nilf(t, err, "Not able to get keys for account: %v", acct.Address)
	}

}

func TestHardhatNode(t *testing.T) {
	privateKeys, _ := dtest.InitializePrivateKeysAndAccounts(4)

	eth, err := blockchain.NewEthereumSimulator(
		privateKeys,
		6,
		10*time.Second,
		30*time.Second,
		0,
		big.NewInt(math.MaxInt64),
		50,
		math.MaxInt64,
		5*time.Second,
		30*time.Second)
	defer func() {
		err := eth.Close()
		if err != nil {
			t.Fatalf("error closing eth: %v", err)
		}
	}()

	assert.Nil(t, err, "Failed to build Ethereum endpoint...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Unlock the default account and use it to deploy contracts
	deployAccount := eth.GetDefaultAccount()
	err = eth.UnlockAccount(deployAccount)
	assert.Nil(t, err, "Failed to unlock default account")

	t.Logf("deploy account: %v", deployAccount.Address.String())

	err = dtest.StartHardHatNode(eth)
	if err != nil {
		t.Fatalf("error starting hardhat node: %v", err)
	}

	t.Logf("waiting on hardhat node to start...")

	err = dtest.WaitForHardHatNode(ctx)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	t.Logf("done testing")
}
