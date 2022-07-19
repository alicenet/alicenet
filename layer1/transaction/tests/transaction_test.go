package tests

import (
	"context"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"testing"
	"time"

	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/stretchr/testify/assert"
)

func Setup(t *testing.T, accounts int, pollingTime time.Duration) (*tests.ClientFixture, *transaction.FrontWatcher) {
	fixture := setupEthereum(t, accounts)
	db := mocks.NewTestDB()
	watcher := transaction.WatcherFromNetwork(fixture.Client, db, false, pollingTime)

	return fixture, watcher
}

func TestSubscribeAndWaitForValidTx(t *testing.T) {
	numAccounts := 2
	fixture, watcher := Setup(t, numAccounts, 1*time.Second)
	eth := fixture.Client
	accounts := eth.GetKnownAccounts()
	assert.Equal(t, numAccounts, len(accounts))

	owner := accounts[0]
	user := accounts[1]

	ctx, cf := context.WithTimeout(context.Background(), 300*time.Second)
	defer cf()

	amount := big.NewInt(12_345)
	txn, err := ethereum.TransferEther(eth, fixture.Logger, user.Address, owner.Address, amount)
	assert.Nil(t, err)

	receipt, err := watcher.SubscribeAndWait(ctx, txn, nil)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	assert.Equal(t, txn.Hash(), receipt.TxHash)

	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, currentHeight, receipt.BlockNumber.Uint64()+eth.GetFinalityDelay())

	_, isPending, err := eth.GetTransactionByHash(ctx, txn.Hash())
	assert.Nil(t, err)
	assert.False(t, isPending)

	mintTxnOpts, err := eth.GetTransactionOpts(ctx, user)
	assert.Nil(t, err)
	mintTxnOpts.NoSend = false
	//mintTxnOpts.GasFeeCap = big.NewInt(1_000_000_000)
	mintTxnOpts.Value = amount

	mintTxn, err := ethereum.GetContracts().BToken().MintTo(mintTxnOpts, owner.Address, big.NewInt(1))
	assert.Nil(t, err)
	assert.NotNil(t, mintTxn)

	txnRough := &types.DynamicFeeTx{}
	txnRough.ChainID = txn.ChainId()
	txnRough.To = txn.To()
	txnRough.GasFeeCap = new(big.Int).Mul(new(big.Int).SetInt64(2), txn.GasFeeCap())
	txnRough.GasTipCap = new(big.Int).Mul(new(big.Int).SetInt64(2), txn.GasTipCap())
	txnRough.Gas = txn.Gas()
	txnRough.Nonce = txn.Nonce() + 1
	txnRough.Value = txn.Value()
	txnRough.Data = txn.Data()

	<-time.After(2 * time.Second)
	fixture.Logger.Infof("New Gasfee: %v", txnRough.GasFeeCap.String())

	signer := types.NewLondonSigner(txnRough.ChainID)

	_, adminPk := tests.GetAdminAccount()
	signedTx, err := types.SignNewTx(adminPk, signer, txnRough)
	if err != nil {
		fixture.Logger.Errorf("signing error:%v", err)
	}
	err = eth.SendTransaction(ctx, signedTx)
	if err != nil {
		fixture.Logger.Errorf("sending error:%v", err)
	}

}

func TestSubscribeAndWaitForInvalidTxNotSigned(t *testing.T) {
	numAccounts := 2
	fixture, watcher := Setup(t, numAccounts, 1*time.Second)
	eth := fixture.Client

	accounts := eth.GetKnownAccounts()
	assert.Equal(t, numAccounts, len(accounts))

	owner := accounts[0]
	user := accounts[1]

	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()

	amount := big.NewInt(12_345)

	txOpts, err := eth.GetTransactionOpts(ctx, user)
	txOpts.NoSend = true
	txOpts.Value = amount
	assert.Nil(t, err)

	//Creating tx but not sending it
	txn, err := ethereum.GetContracts().BToken().MintTo(txOpts, owner.Address, big.NewInt(1))
	assert.Nil(t, err)

	txnRough := &types.DynamicFeeTx{}
	txnRough.ChainID = txn.ChainId()
	txnRough.To = txn.To()
	txnRough.GasFeeCap = new(big.Int).Mul(new(big.Int).SetInt64(2), txn.GasFeeCap())
	txnRough.GasTipCap = new(big.Int).Mul(new(big.Int).SetInt64(2), txn.GasTipCap())
	txnRough.Gas = txn.Gas()
	txnRough.Nonce = txn.Nonce() + 1
	txnRough.Value = txn.Value()
	txnRough.Data = txn.Data()

	txnNotSigned := types.NewTx(txnRough)

	receipt, err := watcher.SubscribeAndWait(ctx, txnNotSigned, nil)
	assert.NotNil(t, err)
	assert.Nil(t, receipt)
	_, ok := err.(*transaction.ErrInvalidMonitorRequest)
	assert.True(t, ok)
}

func TestSubscribeAndWaitForTxNotFound(t *testing.T) {
	numAccounts := 2
	pollingTime := 100 * time.Millisecond
	fixture, watcher := Setup(t, numAccounts, pollingTime)
	eth := fixture.Client

	accounts := eth.GetKnownAccounts()
	assert.Equal(t, numAccounts, len(accounts))

	owner := accounts[0]
	user := accounts[1]

	ctx, cf := context.WithTimeout(context.Background(), 300*time.Second)
	defer cf()

	amount := big.NewInt(12_345)

	txOpts, err := eth.GetTransactionOpts(ctx, user)
	txOpts.NoSend = true
	txOpts.Value = amount
	assert.Nil(t, err)

	//Creating tx but not sending it
	txn, err := ethereum.GetContracts().BToken().MintTo(txOpts, owner.Address, big.NewInt(1))
	assert.Nil(t, err)

	txnRough := &types.DynamicFeeTx{}
	txnRough.ChainID = txn.ChainId()
	txnRough.To = txn.To()
	txnRough.GasFeeCap = new(big.Int).Mul(new(big.Int).SetInt64(2), txn.GasFeeCap())
	txnRough.GasTipCap = new(big.Int).Mul(new(big.Int).SetInt64(2), txn.GasTipCap())
	txnRough.Gas = txn.Gas()
	txnRough.Nonce = txn.Nonce() + 1
	txnRough.Value = txn.Value()
	txnRough.Data = txn.Data()

	signer := types.NewLondonSigner(txnRough.ChainID)

	_, adminPk := tests.GetAdminAccount()
	signedTx, err := types.SignNewTx(adminPk, signer, txnRough)
	if err != nil {
		fixture.Logger.Errorf("signing error:%v", err)
	}

	hardhatEndpoint := "http://127.0.0.1:8545"
	tests.SetBlockInterval(hardhatEndpoint, 500)

	receipt, err := watcher.SubscribeAndWait(ctx, signedTx, nil)
	assert.NotNil(t, err)
	assert.Nil(t, receipt)

	_, ok := err.(*transaction.ErrTxNotFound)
	assert.True(t, ok)
}

func TestSubscribeAndWaitForStaleTx(t *testing.T) {
	numAccounts := 2
	fixture, watcher := Setup(t, numAccounts, 1*time.Second)
	eth := fixture.Client

	hardhatEndpoint := "http://127.0.0.1:8545"
	tests.SetBlockInterval(hardhatEndpoint, 500)
	// setting base fee to 10k GWei
	tests.SetNextBlockBaseFee(hardhatEndpoint, 10_000_000_000_000)

	accounts := eth.GetKnownAccounts()
	assert.Equal(t, numAccounts, len(accounts))

	owner := accounts[0]
	user := accounts[1]

	ctx, cf := context.WithTimeout(context.Background(), 300*time.Second)
	defer cf()

	amount := big.NewInt(12_345)

	txOpts, err := eth.GetTransactionOpts(ctx, user)
	txOpts.GasFeeCap = big.NewInt(1_000_000_000)
	txOpts.Value = amount
	assert.Nil(t, err)
	txn, err := ethereum.GetContracts().BToken().MintTo(txOpts, owner.Address, big.NewInt(1))
	assert.Nil(t, err)

	subscribeOpts := transaction.NewSubscribeOptions(false, 3)
	receipt, err := watcher.SubscribeAndWait(ctx, txn, subscribeOpts)

	assert.NotNil(t, err)
	assert.Nil(t, receipt)
	_, ok := err.(*transaction.ErrTransactionStale)
	assert.True(t, ok)
}

func TestSubscribeAndWaitForStaleTxWithAutoRetry(t *testing.T) {
	numAccounts := 2
	fixture, watcher := Setup(t, numAccounts, 1*time.Second)
	eth := fixture.Client

	hardhatEndpoint := "http://127.0.0.1:8545"
	tests.SetBlockInterval(hardhatEndpoint, 500)
	// setting base fee to 10k GWei
	tests.SetNextBlockBaseFee(hardhatEndpoint, 10_000_000_000_000)

	accounts := eth.GetKnownAccounts()
	assert.Equal(t, numAccounts, len(accounts))

	owner := accounts[0]
	user := accounts[1]

	ctx, cf := context.WithTimeout(context.Background(), 300*time.Second)
	defer cf()

	amount := big.NewInt(12_345)

	txOpts, err := eth.GetTransactionOpts(ctx, owner)
	txOpts.GasFeeCap = big.NewInt(1_000_000_000)
	txOpts.Value = amount
	assert.Nil(t, err)
	txn, err := ethereum.GetContracts().BToken().MintTo(txOpts, user.Address, big.NewInt(1))
	assert.Nil(t, err)

	subscribeOpts := transaction.NewSubscribeOptions(true, 3)
	receipt, err := watcher.SubscribeAndWait(ctx, txn, subscribeOpts)

	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
}
