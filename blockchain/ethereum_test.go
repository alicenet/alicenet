package blockchain_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func setupEthereum(t *testing.T) (blockchain.Ethereum, error) {
	wei, ok := new(big.Int).SetString("9000000000000000000000", 10)
	assert.True(t, ok)

	eth, err := blockchain.NewEthereumSimulator(
		"../assets/test/keys",
		"../assets/test/passcodes.txt",
		1,
		time.Second*2,
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
