package blockchain_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/dkg/dtest"
	"github.com/MadBase/MadNet/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestTransferFunds(t *testing.T) {
	n := 4
	ecdsaPrivateKeys, _ := dtest.InitializePrivateKeysAndAccounts(n)
	eth := dtest.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 100*time.Millisecond)
	defer eth.Close()

	logger := logging.GetLogger("ethereum")
	logger.SetLevel(logrus.DebugLevel)

	finalityDelay := uint64(6)
	eth.SetFinalityDelay(finalityDelay)
	eth.TransactionWatcher().SetNumOfConfirmationBlocks(finalityDelay)

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

	txWatcher := eth.TransactionWatcher()

	receipt, err := txWatcher.SubscribeAndWait(ctx, txn)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	assert.Equal(t, txn.Hash(), receipt.TxHash)
	logger.Infof("Receipt: %v", receipt)
}
