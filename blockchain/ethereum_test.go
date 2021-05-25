package blockchain_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var accountAddresses []string = []string{
	"0x546F99F244b7B58B855330AE0E2BC1b30b41302F", "0x9AC1c9afBAec85278679fF75Ef109217f26b1417",
	"0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac", "0x615695C4a4D6a60830e5fca4901FbA099DF26271",
	"0x63a6627b79813A7A43829490C4cE409254f64177"}

func connectSimulatorEndpoint(t *testing.T) blockchain.Ethereum {
	eth, err := blockchain.NewEthereumSimulator(
		"../assets/test/keys",
		"../assets/test/passcodes.txt",
		6,
		1*time.Second,
		5*time.Second,
		0,
		big.NewInt(9223372036854775807),
		accountAddresses...)

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

	txnCount := 0
	txns := make([]*types.Transaction, 100)
	for idx := 1; idx < len(accountAddresses); idx++ {
		acct, err := eth.GetAccount(common.HexToAddress(accountAddresses[idx]))
		assert.Nil(t, err)
		eth.UnlockAccount(acct)

		txns[txnCount], err = c.StakingToken.Transfer(txnOpts, acct.Address, big.NewInt(10_000_000))
		assert.Nilf(t, err, "Failed on transfer %v", txnCount)
		txnCount++

		o, err := eth.GetTransactionOpts(ctx, acct)
		assert.Nil(t, err)

		txns[txnCount], err = c.StakingToken.Approve(o, c.ValidatorsAddress, big.NewInt(10_000_000))
		assert.Nilf(t, err, "Failed on approval %v", txnCount)
		txnCount++

		txns[txnCount], err = c.Staking.LockStake(o, big.NewInt(1_000_000))
		assert.Nilf(t, err, "Failed on lock %v", txnCount)
		txnCount++

		var validatorId [2]*big.Int

		validatorId[0] = big.NewInt(int64(idx))
		validatorId[1] = big.NewInt(int64(idx * 2))

		txns[txnCount], err = c.Validators.AddValidator(o, acct.Address, validatorId)
		assert.Nilf(t, err, "Failed on register %v", txnCount)
		txnCount++
	}

	for idx := 0; idx < txnCount; idx++ {
		rcpt, err := eth.WaitForReceipt(ctx, txns[idx])
		assert.Nil(t, err)

		t.Logf("rcpt for txn %v ... status %v", txns[idx].Hash().Hex(), rcpt.Status)
	}

	return eth
}

func setupEthereum(t *testing.T) (blockchain.Ethereum, error) {
	wei, ok := new(big.Int).SetString("9000000000000000000000", 10)
	assert.True(t, ok)

	logging.GetLogger("ethsim").SetLevel(logrus.InfoLevel)

	eth, err := blockchain.NewEthereumSimulator(
		"../assets/test/keys",
		"../assets/test/passcodes.txt",
		1,
		time.Second*2,
		time.Second*5,
		0,
		wei,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f", "9ac1c9afbaec85278679ff75ef109217f26b1417")
	assert.Nil(t, err)

	acct := eth.GetDefaultAccount()
	assert.Nil(t, eth.UnlockAccount(acct))

	c := eth.Contracts()
	_, _, err = c.DeployContracts(context.TODO(), acct)
	assert.Nil(t, err)

	go func() {
		for {
			t.Log(".")
			time.Sleep(2 * time.Second)
			eth.Commit()
		}
	}()

	return eth, err
}

func TestAccountsFound(t *testing.T) {
	eth, err := setupEthereum(t)
	assert.Nil(t, err)

	addressStrings := []string{
		"0x546F99F244b7B58B855330AE0E2BC1b30b41302F",
		"0x9AC1c9afBAec85278679fF75Ef109217f26b1417",
		"0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac"}

	for _, addressString := range addressStrings {
		address := common.HexToAddress(addressString)
		acct, err := eth.GetAccount(address)
		assert.Nilf(t, err, "Not able to find account: %v", address)

		err = eth.UnlockAccount(acct)
		assert.Nilf(t, err, "Not able to unlock account: %v", address)

		_, err = eth.GetAccountKeys(acct.Address)
		assert.Nilf(t, err, "Not able to get keys for account: %v", address)
	}

}

func TestValues(t *testing.T) {
	eth, err := setupEthereum(t)
	assert.Nil(t, err)

	c := eth.Contracts()

	txnOpts, err := eth.GetTransactionOpts(context.Background(), eth.GetDefaultAccount())
	assert.Nil(t, err)

	amount := big.NewInt(987654321)
	t.Logf("amount:%v", amount.Text(10))

	txn, err := c.Staking.SetMinimumStake(txnOpts, amount)
	assert.Nil(t, err)

	eth.WaitForReceipt(context.Background(), txn)

	ms, err := c.Staking.MinimumStake(eth.GetCallOpts(context.Background(), eth.GetDefaultAccount()))
	assert.Nil(t, err)
	t.Logf("minimum stake:%v", ms.Text(10))

	assert.Equal(t, 0, amount.Cmp(ms))
}

func TestCreateSelector(t *testing.T) {
	signature := "removeFacet(bytes4)"

	selector := blockchain.CalculateSelector(signature)

	assert.Equal(t, "ca5a0fae", fmt.Sprintf("%x", selector))
}
