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
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestTransferFunds(t *testing.T) {
	logger := logging.GetLogger("transaction")
	logger.SetLevel(logrus.TraceLevel)
	n := 2
	ecdsaPrivateKeys, _ := testutils.InitializePrivateKeysAndAccounts(n)
	eth, err := ethereum.NewSimulator(
		ecdsaPrivateKeys,
		6,
		10*time.Second,
		30*time.Second,
		0,
		big.NewInt(math.MaxInt64),
		50,
		math.MaxInt64)
	defer eth.Close()

	finalityDelay := uint64(6)
	eth.SetFinalityDelay(finalityDelay)
	knownSelectors := transaction.NewKnownSelectors()
	transaction := transaction.NewWatcher(eth.GetClient(), knownSelectors, 5)
	transaction.SetNumOfConfirmationBlocks(finalityDelay)

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

	receipt, err := transaction.SubscribeAndWait(ctx, txn)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	assert.Equal(t, txn.Hash(), receipt.TxHash)
	logger.Infof("Receipt: %v", receipt)
}
