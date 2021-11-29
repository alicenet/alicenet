package blockchain_test

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/etest"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const SETUP_GROUP int = 13

/*
var accountAddresses []string = []string{
	"0x546F99F244b7B58B855330AE0E2BC1b30b41302F", "0x9AC1c9afBAec85278679fF75Ef109217f26b1417",
	"0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac", "0x615695C4a4D6a60830e5fca4901FbA099DF26271",
	"0x63a6627b79813A7A43829490C4cE409254f64177"}

func connectSimulatorEndpoint(t *testing.T) interfaces.Ethereum {

	dkgStates, ecdsaPrivateKeys, _ := InitializeNewDkgStateInfo(5)

	eth, err := blockchain.NewEthereumSimulator(
		ecdsaPrivateKeys,
		6,
		1*time.Second,
		5*time.Second,
		0,
		big.NewInt(9223372036854775807))

	assert.Nil(t, err, "Failed to build Ethereum endpoint...")
	assert.True(t, eth.IsEthereumAccessible(), "Web3 endpoint is not available.")

	// Mine a block once a second
	go func() {
		for true {
			time.Sleep(1 * time.Second)
			eth.Commit()
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Unlock the default account and use it to deploy contracts
	deployAccount := eth.GetDefaultAccount()
	err = eth.UnlockAccount(deployAccount)
	assert.Nil(t, err, "Failed to unlock default account")

	txnOpts, err := eth.GetTransactionOpts(ctx, deployAccount)
	assert.Nil(t, err, "Failed to create txn opts")

	c := eth.Contracts()
	_, _, err = c.DeployContracts(ctx, deployAccount)
	assert.Nil(t, err, "Failed to deploy contracts...")

	var txn *types.Transaction
	for idx := 1; idx < len(accountAddresses); idx++ {
		acct, err := eth.GetAccount(common.HexToAddress(accountAddresses[idx]))
		assert.Nil(t, err)
		eth.UnlockAccount(acct)

		txn, err = c.StakingToken().Transfer(txnOpts, acct.Address, big.NewInt(10_000_000))
		assert.Nilf(t, err, "Failed on transfer %v", idx)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)

		o, err := eth.GetTransactionOpts(ctx, acct)
		assert.Nil(t, err)

		txn, err = c.StakingToken().Approve(o, c.ValidatorsAddress(), big.NewInt(10_000_000))
		assert.Nilf(t, err, "Failed on approval %v", idx)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)

		txn, err = c.Staking().LockStake(o, big.NewInt(1_000_000))
		assert.Nilf(t, err, "Failed on lock %v", idx)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)

		var validatorId [2]*big.Int

		validatorId[0] = big.NewInt(int64(idx))
		validatorId[1] = big.NewInt(int64(idx * 2))

		txn, err = c.Validators().AddValidator(o, acct.Address, validatorId)
		assert.Nilf(t, err, "Failed on register %v", idx)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)
	}

	rcpts, err := eth.Queue().WaitGroupTransactions(ctx, SETUP_GROUP)
	assert.Nil(t, err)

	for _, rcpt := range rcpts {
		assert.Equal(t, uint64(1), rcpt.Status)
	}

	return eth
}
*/

func setupEthereum(t *testing.T, n int) (interfaces.Ethereum, error) {
	logging.GetLogger("ethereum").SetLevel(logrus.InfoLevel)

	privKeys := etest.SetupPrivateKeys(n)
	eth, err := blockchain.NewEthereumSimulator(
		privKeys,
		1,
		time.Second*2,
		time.Second*5,
		0,
		big.NewInt(math.MaxInt64))
	assert.Nil(t, err)

	acct := eth.GetDefaultAccount()
	assert.Nil(t, eth.UnlockAccount(acct))

	c := eth.Contracts()
	_, _, err = c.DeployContracts(context.TODO(), acct)
	assert.Nil(t, err)

	go func() {
		for {
			t.Log(".")
			time.Sleep(100 * time.Second)
			eth.Commit()
		}
	}()

	return eth, err
}

func TestAccountsFound(t *testing.T) {
	eth, err := setupEthereum(t, 4)
	assert.Nil(t, err)
	defer eth.Close()

	accountList := eth.GetKnownAccounts()

	for _, acct := range accountList {

		err = eth.UnlockAccount(acct)
		assert.Nilf(t, err, "Not able to unlock account: %v", acct.Address)

		_, err = eth.GetAccountKeys(acct.Address)
		assert.Nilf(t, err, "Not able to get keys for account: %v", acct.Address)
	}

}

func TestValues(t *testing.T) {
	eth, err := setupEthereum(t, 4)
	assert.Nil(t, err)
	defer eth.Close()

	c := eth.Contracts()
	c.DeployContracts(context.Background(), eth.GetDefaultAccount())

	eth.Commit()

	txnOpts, err := eth.GetTransactionOpts(context.Background(), eth.GetDefaultAccount())
	assert.Nil(t, err)

	amount := big.NewInt(987654321)
	t.Logf("amount:%v", amount.Text(10))

	txn, err := c.Staking().SetMinimumStake(txnOpts, amount)
	assert.Nil(t, err)
	eth.Commit()
	eth.Queue().QueueAndWait(context.Background(), txn)

	eth.Commit()
	ms, err := c.Staking().MinimumStake(eth.GetCallOpts(context.Background(), eth.GetDefaultAccount()))

	eth.Commit()
	assert.Nil(t, err)
	t.Logf("minimum stake:%v", ms.Text(10))

	assert.Equal(t, 0, amount.Cmp(ms))
}

func TestCreateSelector(t *testing.T) {
	signature := "removeFacet(bytes4)"

	selector := blockchain.CalculateSelector(signature)

	assert.Equal(t, "ca5a0fae", fmt.Sprintf("%x", selector))
}
