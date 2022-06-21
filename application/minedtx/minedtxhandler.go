package minedtx

import (
	"github.com/alicenet/alicenet/constants/dbprefix"

	"github.com/alicenet/alicenet/application/db"
	"github.com/alicenet/alicenet/application/indexer"
	"github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

// NewMinedTxHandler creates a new MinedTxHandler object
func NewMinedTxHandler() *MinedTxHandler {
	return &MinedTxHandler{
		heightIdxIndex: indexer.NewHeightIdxIndex(dbprefix.PrefixMinedTxIndexKey, dbprefix.PrefixMinedTxIndexRefKey),
	}
}

// MinedTxHandler manages the storage of mined trasactions with indexing
type MinedTxHandler struct {
	heightIdxIndex *indexer.HeightIdxIndex
}

// Add adds txs at height to MinedTxHandler
func (mt *MinedTxHandler) Add(txn *badger.Txn, height uint32, txs []*objs.Tx) error {
	for j := 0; j < len(txs); j++ {
		tx := txs[j]
		txHash, err := tx.TxHash()
		if err != nil {
			return err
		}
		err = mt.addOneInternal(txn, tx, txHash, height)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete removes txs corresponding to txHashes from MinedTxHandler
func (mt *MinedTxHandler) Delete(txn *badger.Txn, txHashes [][]byte) error {
	for j := 0; j < len(txHashes); j++ {
		txHash := utils.CopySlice(txHashes[j])
		err := mt.heightIdxIndex.Delete(txn, utils.CopySlice(txHash))
		if err != nil {
			return err
		}
		key := mt.makeMinedTxKey(utils.CopySlice(txHash))
		if err := utils.DeleteValue(txn, key); err != nil {
			return err
		}
	}
	return nil
}

// Get retrieves txs as well as any txHashes which are missing
func (mt *MinedTxHandler) Get(txn *badger.Txn, txHashes [][]byte) ([]*objs.Tx, [][]byte, error) {
	var missing [][]byte
	var result []*objs.Tx
	for j := 0; j < len(txHashes); j++ {
		txHash := utils.CopySlice(txHashes[j])
		tx, err := mt.getOneInternal(txn, utils.CopySlice(txHash))
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return nil, nil, err
			}
			missing = append(missing, utils.CopySlice(txHash))
			continue
		}
		result = append(result, tx)
	}
	return result, missing, nil
}

// GetHeightForTx returns height for a given txHash
func (mt *MinedTxHandler) GetHeightForTx(txn *badger.Txn, txHash []byte) (uint32, error) {
	height, _, err := mt.heightIdxIndex.GetHeightIdx(txn, txHash)
	if err != nil {
		return 0, err
	}
	return height, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
/////////PRIVATE METHODS////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (mt *MinedTxHandler) getOneInternal(txn *badger.Txn, txHash []byte) (*objs.Tx, error) {
	key := mt.makeMinedTxKey(txHash)
	return db.GetTx(txn, key)
}

func (mt *MinedTxHandler) addOneInternal(txn *badger.Txn, tx *objs.Tx, txHash []byte, height uint32) error {
	if err := tx.ValidateIssuedAtForMining(height); err != nil {
		return err
	}
	key := mt.makeMinedTxKey(txHash)
	err := mt.heightIdxIndex.Add(txn, txHash, height, 0)
	if err != nil {
		return err
	}
	return db.SetTx(txn, key, tx)
}

func (mt *MinedTxHandler) makeMinedTxKey(txHash []byte) []byte {
	key := []byte{}
	key = append(key, dbprefix.PrefixMinedTx()...)
	key = append(key, txHash...)
	return key
}
