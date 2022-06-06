package transaction_test

import (
	"context"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/blockchain/testutils"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Setup(finalityDelay uint64, numAccounts int, registryAddress common.Address) (ethereum.Network, *logrus.Logger, error) {
	logger := logging.GetLogger("test")
	logger.SetLevel(logrus.TraceLevel)
	ecdsaPrivateKeys, _ := testutils.InitializePrivateKeysAndAccounts(numAccounts)
	eth, err := ethereum.NewSimulator(
		ecdsaPrivateKeys,
		6,
		10*time.Second,
		30*time.Second,
		0,
		big.NewInt(math.MaxInt64),
		50,
		math.MaxInt64)
	if err != nil {
		return nil, logger, err
	}

	eth.SetFinalityDelay(finalityDelay)
	knownSelectors := transaction.NewKnownSelectors()
	transaction := transaction.NewWatcher(eth.GetClient(), knownSelectors, 5)
	transaction.SetNumOfConfirmationBlocks(finalityDelay)

	//todo: redeploy and get the registryAddress here
	err = eth.Contracts().LookupContracts(context.Background(), registryAddress)
	if err != nil {
		return nil, logger, err
	}
	return eth, logger, nil
}

func TestSubscribeAndWaitForValidTx(t *testing.T) {
	finalityDelay := uint64(6)
	numAccounts := 2
	eth, _, err := Setup(finalityDelay, numAccounts, common.HexToAddress("0x0b1F9c2b7bED6Db83295c7B5158E3806d67eC5bc"))
	assert.Nil(t, err)
	defer eth.Close()

	testutils.SetBlockInterval(t, eth, 500)

	accounts := eth.GetKnownAccounts()
	assert.Equal(t, numAccounts, len(accounts))

	for _, acct := range accounts {
		err := eth.UnlockAccount(acct)
		assert.Nil(t, err)
	}

	owner := accounts[0]
	user := accounts[1]

	ctx, cf := context.WithTimeout(context.Background(), 300*time.Second)
	defer cf()

	amount := big.NewInt(12_345)

	txn, err := eth.TransferEther(owner.Address, user.Address, amount)
	assert.Nil(t, err)

	watcher := transaction.WatcherFromNetwork(eth)

	receipt, err := watcher.SubscribeAndWait(ctx, txn)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	assert.Equal(t, txn.Hash(), receipt.TxHash)

	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, currentHeight, receipt.BlockNumber.Uint64()+finalityDelay)

	// _, isPending, err := eth.GetEthereumClient().TransactionByHash(ctx, common.HexToHash("0xbf5a6c45305b0a1abf27d3a29f3463cba269d66c57fe06d1962e227baaf47f10"))
	// logger.Infof("Pending: %v", isPending)

	// txOpts, err := eth.GetTransactionOpts(ctx, owner)
	// txOpts.NoSend = false
	// txOpts.GasFeeCap = big.NewInt(1_000_000_000)
	// txOpts.Value = amount
	// assert.Nil(t, err)
	// txn, err := eth.Contracts().BToken().MintTo(txOpts, user.Address, big.NewInt(1))
	// assert.Nil(t, err)

	// txRough := &types.DynamicFeeTx{}
	// txRough.ChainID = txn.ChainId()
	// txRough.To = txn.To()
	// txRough.GasFeeCap = new(big.Int).Mul(new(big.Int).SetInt64(2), txn.GasFeeCap())
	// txRough.GasTipCap = new(big.Int).Mul(new(big.Int).SetInt64(2), txn.GasTipCap())
	// txRough.Gas = txn.Gas()
	// txRough.Nonce = txn.Nonce() + 1
	// txRough.Value = txn.Value()
	// txRough.Data = txn.Data()

	// <-time.After(2 * time.Second)
	// logger.Infof("New Gasfee: %v", txRough.GasFeeCap.String())

	// signer := types.NewLondonSigner(txRough.ChainID)

	// signedTx, err := types.SignNewTx(ecdsaPrivateKeys[0], signer, txRough)
	// if err != nil {
	// 	logger.Errorf("signing error:%v", err)
	// }
	// err = eth.GetEthereumClient().SendTransaction(ctx, signedTx)
	// if err != nil {
	// 	logger.Errorf("sending error:%v", err)
	// }

}

func TestSubscribeAndWaitForInvalidTx(t *testing.T) {
	finalityDelay := uint64(6)
	numAccounts := 2
	eth, _, err := Setup(finalityDelay, numAccounts, common.HexToAddress("0x0b1F9c2b7bED6Db83295c7B5158E3806d67eC5bc"))
	assert.Nil(t, err)
	defer eth.Close()

	testutils.SetBlockInterval(t, eth, 200)

	accounts := eth.GetKnownAccounts()
	assert.Equal(t, numAccounts, len(accounts))

	for _, acct := range accounts {
		err := eth.UnlockAccount(acct)
		assert.Nil(t, err)
	}

	owner := accounts[0]
	user := accounts[1]

	ctx, cf := context.WithTimeout(context.Background(), 300*time.Second)
	defer cf()

	amount := big.NewInt(12_345)

	txOpts, err := eth.GetTransactionOpts(ctx, owner)
	txOpts.NoSend = true
	txOpts.Value = amount
	assert.Nil(t, err)
	//Creating tx but not sending it
	txn, err := eth.Contracts().BToken().MintTo(txOpts, user.Address, big.NewInt(1))
	assert.Nil(t, err)

	watcher := transaction.WatcherFromNetwork(eth)
	receipt, err := watcher.SubscribeAndWait(ctx, txn)
	assert.NotNil(t, err)
	assert.Nil(t, receipt)
	_, ok := err.(*transaction.ErrNonRecoverable)
	assert.True(t, ok)
}

func TestSubscribeAndWaitForStaleTx(t *testing.T) {
	finalityDelay := uint64(6)
	numAccounts := 2
	eth, _, err := Setup(finalityDelay, numAccounts, common.HexToAddress("0x0b1F9c2b7bED6Db83295c7B5158E3806d67eC5bc"))
	assert.Nil(t, err)
	defer eth.Close()

	testutils.SetBlockInterval(t, eth, 500)
	// setting base fee to 10k GWei
	testutils.SetNextBlockBaseFee(t, eth, 10_000_000_000_000)

	accounts := eth.GetKnownAccounts()
	assert.Equal(t, numAccounts, len(accounts))

	for _, acct := range accounts {
		err := eth.UnlockAccount(acct)
		assert.Nil(t, err)
	}

	owner := accounts[0]
	user := accounts[1]

	ctx, cf := context.WithTimeout(context.Background(), 300*time.Second)
	defer cf()

	amount := big.NewInt(12_345)

	txOpts, err := eth.GetTransactionOpts(ctx, owner)
	txOpts.GasFeeCap = big.NewInt(1_000_000_000)
	txOpts.Value = amount
	assert.Nil(t, err)
	txn, err := eth.Contracts().BToken().MintTo(txOpts, user.Address, big.NewInt(1))
	assert.Nil(t, err)

	watcher := transaction.WatcherFromNetwork(eth)
	receipt, err := watcher.SubscribeAndWait(ctx, txn)

	assert.NotNil(t, err)
	assert.Nil(t, receipt)
	_, ok := err.(*transaction.ErrTransactionStale)
	assert.True(t, ok)
}
