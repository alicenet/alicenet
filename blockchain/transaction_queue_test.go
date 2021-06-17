package blockchain_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestFoo(t *testing.T) {
	eth := connectSimulatorEndpoint(t)
	defer eth.Close()

	who := common.HexToAddress("0x9AC1c9afBAec85278679fF75Ef109217f26b1417")
	c := eth.Contracts()

	txnCount := 20
	txnGroupCount := 4

	queue := eth.Queue()

	ctx := context.Background()

	txnOpts, err := eth.GetTransactionOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)

	toctx, cf := context.WithTimeout(ctx, 100*time.Second)
	defer cf()

	amount := int64(1_999_999)
	bAmount := big.NewInt(amount)

	txn, err := c.StakingToken().Transfer(txnOpts, who, bAmount)
	assert.Nil(t, err)
	queue.QueueGroupTransaction(toctx, 511, txn)

	txn, err = eth.TransferEther(who, c.StakingTokenAddress(), bAmount)
	assert.Nil(t, err)
	queue.QueueGroupTransaction(toctx, 511, txn)

	rcpts, err := queue.WaitGroupTransactions(toctx, 511)
	assert.Nil(t, err)

	for idx := 0; idx < txnCount; idx++ {
		txn, err = c.StakingToken().Transfer(txnOpts, who, big.NewInt(amount))
		queue.QueueGroupTransaction(toctx, 5+idx%txnGroupCount, txn)
	}

	for idx := 0; idx < txnGroupCount; idx++ {
		rcpts, err = queue.WaitGroupTransactions(toctx, 5+idx)
		assert.Nil(t, err)
		assert.Equal(t, txnCount/txnGroupCount, len(rcpts))
	}

}
