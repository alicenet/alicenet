package pendingindex

import (
	"context"
	"testing"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/internal/testing/environment"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
)

func createDb(t *testing.T) *db.Database {
	t.Helper()
	rawEngineDb, err := utils.OpenBadger(context.Background().Done(), "", true)
	if err != nil {
		t.Fatal(err)
	}
	database := &db.Database{}
	database.Init(rawEngineDb)
	return database
}

func TestPendingTxIndexer_Add_shouldAdd(t *testing.T) {
	pendingtxIndexer := NewPendingTxIndexer()

	database := environment.SetupBadgerDatabase(t)

	err := database.Update(func(txn *badger.Txn) error {
		evicted, err := pendingtxIndexer.Add(txn, 0, []byte("txHash"), [][]byte{[]byte("utxoID")})

		assert.NoError(t, err)
		assert.Nil(t, evicted)
		return nil
	})
	assert.NoError(t, err)
}

func TestPendingTxIndexer_DeleteOne_shouldDeleteOne(t *testing.T) {
	pendingtxIndexer := NewPendingTxIndexer()

	database := environment.SetupBadgerDatabase(t)

	err := database.Update(func(txn *badger.Txn) error {
		err := pendingtxIndexer.DeleteOne(txn, []byte("txHash"))

		assert.NoError(t, err)
		return nil
	})
	assert.NoError(t, err)
}

func TestPendingTxIndexer_DeleteMined_shouldDeleteMined(t *testing.T) {
	pendingtxIndexer := NewPendingTxIndexer()

	database := environment.SetupBadgerDatabase(t)

	err := database.Update(func(txn *badger.Txn) error {

		txHashes, utxoIDs, err := pendingtxIndexer.DeleteMined(txn, []byte("txHash"))

		assert.NoError(t, err)
		assert.NotNil(t, txHashes)
		assert.NotNil(t, utxoIDs)
		return nil
	})
	assert.NoError(t, err)
}

func TestPendingTxIndexer_DropBefore_shouldDropBefore(t *testing.T) {
	pendingtxIndexer := NewPendingTxIndexer()

	database := environment.SetupBadgerDatabase(t)

	err := database.Update(func(txn *badger.Txn) error {
		epoch := uint32(1)
		txHashes, err := pendingtxIndexer.DropBefore(txn, epoch)

		assert.NoError(t, err)
		assert.NotNil(t, txHashes)
		return nil
	})
	assert.NoError(t, err)
}

func TestPendingTxIndexer_Add_shouldEvictOnThresholdReached(t *testing.T) {
	pendingtxIndexer := NewPendingTxIndexer()

	database := environment.SetupBadgerDatabase(t)

	err := database.Update(func(txn *badger.Txn) error {
		utxoIds := [][]byte{make([]byte, 164), make([]byte, 164), make([]byte, 164), make([]byte, 164)}

		evicted, err := pendingtxIndexer.Add(txn, 0, []byte("txHash"), utxoIds)

		assert.NoError(t, err)
		assert.NotNil(t, evicted)
		return nil
	})
	assert.NoError(t, err)
}
