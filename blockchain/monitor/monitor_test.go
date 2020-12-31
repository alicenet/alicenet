package monitor_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/ethereum/go-ethereum/common"

	"github.com/stretchr/testify/assert"
)

func setupEthereum(t *testing.T) (blockchain.Ethereum, func()) {

	// eth, err := blockchain.NewEthereumEndpoint(
	// 	"http://localhost:8545",
	// 	"../../assets/test/keys",
	// 	"../../assets/test/passcodes.txt",
	// 	"0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac",
	// 	3*time.Second,
	// 	3,
	// 	5*time.Second,
	// 	12)
	eth, commit, err := blockchain.NewEthereumSimulator(
		"../../assets/test/keys",
		"../../assets/test/passcodes.txt",
		3,
		5*time.Second,
		0,
		big.NewInt(9223372036854775807),
		"0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac",
		"0x546F99F244b7B58B855330AE0E2BC1b30b41302F")

	assert.Nil(t, err, "Failed to build Ethereum endpoint...")
	assert.True(t, eth.IsEthereumAccessible(), "Web3 endpoint is not available.")

	c := eth.Contracts()

	// Unlock deploy account and make sure it has a balance
	acct := eth.GetDefaultAccount()
	err = eth.UnlockAccount(acct)
	assert.Nil(t, err, "Failed to unlock deploy account")

	deployAcct, _ := eth.GetAccount(common.HexToAddress("0x546F99F244b7B58B855330AE0E2BC1b30b41302F"))
	err = eth.UnlockAccount(deployAcct)
	assert.Nil(t, err, "Failed to unlock deploy account")

	bal, err := eth.GetBalance(acct.Address)
	assert.Nil(t, err, "Can't check balance for %v.", deployAcct.Address.Hex())
	t.Logf(" deploy account (%v) balance: %v", deployAcct.Address.Hex(), bal)

	// Unlock testing account and make sure it has a balance
	err = eth.UnlockAccount(acct)
	assert.Nil(t, err, "Failed to unlock default account")

	bal, err = eth.GetBalance(acct.Address)
	assert.Nil(t, err, "Can't check balance for %v.", acct.Address.Hex())
	t.Logf("default account (%v) balance: %v", acct.Address.Hex(), bal)

	// Transfer some eth
	testingEth := big.NewInt(200000)
	t.Logf("Transfering %v from %v to %v", testingEth, deployAcct.Address.Hex(), acct.Address.Hex())
	err = eth.TransferEther(deployAcct.Address, acct.Address, testingEth)
	assert.Nil(t, err, "Failed to transfer ether to default account")

	_, _, err = c.DeployContracts(context.TODO(), acct)
	assert.Nil(t, err, "Failed to deploy contracts...")

	return eth, commit
}

func TestMonitor(t *testing.T) {
	eth, commit := setupEthereum(t)
	c := eth.Contracts()

	txnOpts, err := eth.GetTransactionOpts(context.TODO(), eth.GetDefaultAccount())
	assert.Nil(t, err, "Failed to build txnOpts endpoint... %v", err)

	_, err = c.Ethdkg.InitializeState(txnOpts)
	assert.Nil(t, err, "Failed to Initialize state... %v", err)

	commit()
}
