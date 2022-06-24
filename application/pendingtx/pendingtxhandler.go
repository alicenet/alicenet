package pendingtx

import (
	"context"
	"time"

	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/errorz"

	"github.com/alicenet/alicenet/application/db"
	"github.com/alicenet/alicenet/application/objs"
	index "github.com/alicenet/alicenet/application/pendingtx/pendingindex"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

type utxoHandler interface {
	TrieContains(txn *badger.Txn, utxoID []byte) (bool, error)
	IsValid(txn *badger.Txn, txs objs.TxVec, currentHeight uint32, deposits objs.Vout) (objs.Vout, error)
}

type depositHandler interface {
	Get(txn *badger.Txn, utxoIDs [][]byte) ([]*objs.TXOut, [][]byte, []*objs.TXOut, error)
}

// NewPendingTxHandler creates a new Handler object
func NewPendingTxHandler(db *badger.DB) *Handler {
	return &Handler{
		indexer: index.NewPendingTxIndexer(),
		db:      db,
		logger:  logging.GetLogger(constants.LoggerApp),
	}
}

// Handler is the object that acts as the pending tx pool
type Handler struct {
	indexer        *index.PendingTxIndexer
	db             *badger.DB
	UTXOHandler    utxoHandler
	logger         *logrus.Logger
	DepositHandler depositHandler
}

// Add stores a tx in the tx pool and possibly evicts other txs if the ref
// counting of utxo consumers requires it.
func (pt *Handler) Add(txnState *badger.Txn, txs []*objs.Tx, currentHeight uint32) error {
	if err := pt.checkIsValid(txnState, txs, currentHeight); err != nil {
		utils.DebugTrace(pt.logger, err)
		return err
	}
	return pt.db.Update(func(txn *badger.Txn) error {
		for i := 0; i < len(txs); i++ {
			tx := txs[i]
			utxoIds, err := tx.ConsumedUTXOID()
			if err != nil {
				utils.DebugTrace(pt.logger, err)
				return err
			}
			txHash, err := tx.TxHash()
			if err != nil {
				utils.DebugTrace(pt.logger, err)
				return err
			}
			eoe, err := tx.EpochOfExpirationForMining()
			if err != nil {
				utils.DebugTrace(pt.logger, err)
				return err
			}
			cooldownKey := pt.makePendingTxCooldownKey(txHash)
			_, err = utils.GetValue(txn, cooldownKey)
			if err != nil {
				if err == badger.ErrKeyNotFound {
					err := pt.addOneInternal(txn, tx, eoe, txHash, utxoIds)
					if err != nil {
						utils.DebugTrace(pt.logger, err)
						return err
					}
					continue
				}
				utils.DebugTrace(pt.logger, err)
				return nil
			}
			return errorz.ErrInvalid{}.New("ptHandler.Add; already mined")
		}
		return nil
	})
}

// Delete removes a list of txHashes from the tx pool
func (pt *Handler) Delete(txnState *badger.Txn, txHashes [][]byte) error {
	var txHash []byte
	return pt.db.Update(func(txn *badger.Txn) error {
		for i := 0; i < len(txHashes); i++ {
			txHash = utils.CopySlice(txHashes[i])
			if err := pt.deleteOneInternal(txn, utils.CopySlice(txHash), false); err != nil {
				utils.DebugTrace(pt.logger, err)
				return err
			}
		}
		return nil
	})
}

// Get returns a list of txs based on txHashes and a list of all txHashes that
// could not be found
func (pt *Handler) Get(txnState *badger.Txn, currentHeight uint32, txHashes [][]byte) ([]*objs.Tx, [][]byte, error) {
	var txs []*objs.Tx
	var missing [][]byte
	var txHash []byte
	err := pt.db.View(func(txn *badger.Txn) error {
		for i := 0; i < len(txHashes); i++ {
			txHash = utils.CopySlice(txHashes[i])
			epoch := utils.Epoch(currentHeight)
			tx, err := pt.getOneInternal(txn, epoch, utils.CopySlice(txHash))
			if err != nil {
				if err != errorz.ErrMissingTransactions {
					utils.DebugTrace(pt.logger, err)
					return err
				}
				missing = append(missing, utils.CopySlice(txHash))
				continue
			}
			txs = append(txs, tx)
		}
		return nil
	})
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return nil, nil, err
	}
	return txs, missing, nil
}

// Contains returns a list of missing transactions when a list of tx hashes is
// passed in
func (pt *Handler) Contains(txnState *badger.Txn, currentHeight uint32, txHashes [][]byte) ([][]byte, error) {
	_, missing, err := pt.Get(txnState, currentHeight, txHashes)
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return nil, err
	}
	return missing, nil
}

// DeleteMined removes all specified transactions from the pool as well as any
// other transactions that reference a consumed UTXO from the set of passed in
// transactions
func (pt *Handler) DeleteMined(txnState *badger.Txn, currentHeight uint32, txHashes [][]byte) error {
	var txHash []byte
	return pt.db.Update(func(txn *badger.Txn) error {
		for i := 0; i < len(txHashes); i++ {
			txHash = utils.CopySlice(txHashes[i])
			cooldownKey := pt.makePendingTxCooldownKey(utils.CopySlice(txHash))
			e := badger.NewEntry(cooldownKey, []byte{uint8(1)}).WithTTL(time.Second * 20)
			err := txn.SetEntry(e)
			if err != nil {
				utils.DebugTrace(pt.logger, err)
			}
		}
		for i := 0; i < len(txHashes); i++ {
			txHash = utils.CopySlice(txHashes[i])
			err := pt.deleteOneInternal(txn, utils.CopySlice(txHash), true)
			if err != nil {
				utils.DebugTrace(pt.logger, err)
			}
		}
		dropHashes, err := pt.indexer.DropBefore(txn, utils.Epoch(currentHeight))
		if err != nil {
			utils.DebugTrace(pt.logger, err)
			return err
		}
		for i := 0; i < len(dropHashes); i++ {
			txHash = utils.CopySlice(dropHashes[i])
			cooldownKey := pt.makePendingTxCooldownKey(utils.CopySlice(txHash))
			e := badger.NewEntry(cooldownKey, []byte{uint8(1)}).WithTTL(time.Second * 20)
			err := txn.SetEntry(e)
			if err != nil {
				utils.DebugTrace(pt.logger, err)
				return err // New returned error (Hunter and Chris)
			}
			err = pt.deleteOneInternal(txn, utils.CopySlice(txHash), true)
			if err != nil {
				utils.DebugTrace(pt.logger, err)
				return err // New returned error (Hunter and Chris)
			}
		}
		return nil
	})
}

// GetTxsForProposal returns an set of txs that are mutually exclusive with
// respect to the consumed UTXOs. This is used to genrete new proposals.
func (pt *Handler) GetTxsForProposal(txnState *badger.Txn, ctx context.Context, currentHeight uint32, maxBytes uint32, tx *objs.Tx) (objs.TxVec, uint32, error) {
	var utxos objs.TxVec
	var err error
	utxos, maxBytes, err = pt.getTxsInternal(txnState, ctx, currentHeight, maxBytes, tx, false)
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return nil, 0, err
	}
	return utxos, maxBytes, nil
}

// GetTxsForGossip returns the oldest non-expired and non-consumed txs from the
// tx pool. These txs may be conflicting in terms of consumed UTXOS.
func (pt *Handler) GetTxsForGossip(txnState *badger.Txn, ctx context.Context, currentHeight uint32, maxBytes uint32) ([]*objs.Tx, error) {
	utxos, _, err := pt.getTxsInternal(txnState, ctx, currentHeight, maxBytes, nil, true)
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return nil, err
	}
	return utxos, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
/////////PRIVATE METHODS////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (pt *Handler) getTxsInternal(txnState *badger.Txn, ctx context.Context, currentHeight uint32, maxBytes uint32, tx *objs.Tx, allowConflict bool) ([]*objs.Tx, uint32, error) {
	txs := objs.TxVec{}
	if tx != nil {
		txs = append(txs, tx)
	}
	byteCount := uint32(0)
	if len(txs) > 0 {
		byteCount += constants.HashLen
	}
	dropKeys := [][]byte{}
	err := pt.db.View(func(txn *badger.Txn) error {
		it, prefix := pt.indexer.GetOrderedIter(txn)
		err := func() error {
			defer it.Close()
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				itm := it.Item()
				vBytes, err := itm.ValueCopy(nil)
				if err != nil {
					utils.DebugTrace(pt.logger, err)
					return err
				}
				txHash := vBytes[len(prefix):]
				select {
				case <-ctx.Done():
					break
				default:
				}
				tx, err := pt.getOneInternal(txn, utils.Epoch(currentHeight), txHash)
				if err != nil {
					utils.DebugTrace(pt.logger, err)
					continue
				}
				if tx == nil {
					continue
				}
				if ok := pt.checkSize(maxBytes, byteCount); !ok {
					break
				}
				ok, err := pt.checkTx(txnState, tx, currentHeight)
				if err != nil {
					utils.DebugTrace(pt.logger, err)
					if len(dropKeys) < 1000 {
						dropKeys = append(dropKeys, utils.CopySlice(txHash))
					}
					continue
				}
				if !ok {
					continue
				}
				err = tx.ValidateIssuedAtForMining(currentHeight)
				if err != nil {
					continue
				}
				txs = append(txs, tx)
				if !allowConflict && len(txs) > 1 {
					if _, err := txs.ValidateUnique(nil); err != nil {
						txs = txs[0 : len(txs)-1]
						continue
					}
					if _, err := txs.ValidateDataStoreIndexes(nil); err != nil {
						txs = txs[0 : len(txs)-1]
						continue
					}
					if err := pt.checkIsValid(txnState, txs, currentHeight); err != nil {
						txs = txs[0 : len(txs)-1]
						continue
					}

				}
				byteCount += constants.HashLen
			}
			return nil
		}()
		if err != nil {
			utils.DebugTrace(pt.logger, err)
			return err
		}
		return nil
	})
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return nil, 0, err
	}

	err = pt.db.Update(func(txn *badger.Txn) error {
		for i := 0; i < len(dropKeys); i++ {
			err := pt.deleteOneInternal(txn, dropKeys[i], false)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return nil, 0, err
	}

	out := []*objs.Tx{}
	for i := 0; i < len(txs); i++ {
		if txs[i] != nil {
			out = append(out, txs[i])
		}
	}
	remainingBytes := maxBytes - byteCount
	return out, remainingBytes, err
}

func (pt *Handler) checkSize(maxBytes uint32, byteCount uint32) bool {
	return byteCount+constants.HashLen <= maxBytes
}

func (pt *Handler) checkTx(txnState *badger.Txn, tx *objs.Tx, currentHeight uint32) (bool, error) {
	ok, err := pt.checkGenerated(txnState, tx)
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return false, err
	}
	if !ok {
		return false, nil
	}
	ok, err = pt.checkConsumedUTXOs(txnState, tx)
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return false, err
	}
	if !ok {
		return false, nil
	}
	ok, err = pt.checkConsumedDeposits(txnState, tx)
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return false, err
	}
	if !ok {
		return false, nil
	}
	err = pt.checkIsValid(txnState, []*objs.Tx{tx}, currentHeight)
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return false, nil
	}
	return true, nil
}

func (pt *Handler) checkGenerated(txnState *badger.Txn, tx *objs.Tx) (bool, error) {
	generatedUTXOIDs, err := tx.GeneratedUTXOID()
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return false, err
	}
	for k := 0; k < len(generatedUTXOIDs); k++ {
		trieContains, err := pt.UTXOHandler.TrieContains(txnState, generatedUTXOIDs[k])
		if err != nil {
			utils.DebugTrace(pt.logger, err)
			return false, err
		}
		if trieContains {
			utils.DebugTrace(pt.logger, err)
			return false, nil
		}
	}
	return true, nil
}

func (pt *Handler) checkConsumedUTXOs(txnState *badger.Txn, tx *objs.Tx) (bool, error) {
	txv := objs.TxVec{tx}
	consumedUTXOIDs, err := txv.ConsumedUTXOIDNoDeposits()
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return false, err
	}
	for k := 0; k < len(consumedUTXOIDs); k++ {
		trieContains, err := pt.UTXOHandler.TrieContains(txnState, consumedUTXOIDs[k])
		if err != nil {
			utils.DebugTrace(pt.logger, err)
			return false, err
		}
		if !trieContains {
			utils.DebugTrace(pt.logger, err)
			return false, nil
		}
	}
	return true, nil
}

func (pt *Handler) checkConsumedDeposits(txnState *badger.Txn, tx *objs.Tx) (bool, error) {
	consumedUTXOIDs, err := objs.TxVec([]*objs.Tx{tx}).ConsumedUTXOIDOnlyDeposits()
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return false, err
	}
	for k := 0; k < len(consumedUTXOIDs); k++ {
		trieContainsSpent, err := pt.UTXOHandler.TrieContains(txnState, consumedUTXOIDs[k])
		if err != nil {
			utils.DebugTrace(pt.logger, err)
			return false, err
		}
		if trieContainsSpent {
			return false, nil
		}
	}
	return true, nil
}

func (pt *Handler) checkIsValid(txn *badger.Txn, txs objs.TxVec, currentHeight uint32) error {
	utxoIDs, err := txs.ConsumedUTXOIDOnlyDeposits()
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return err
	}
	deposits, missing, spent, err := pt.DepositHandler.Get(txn, utxoIDs)
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return err
	}
	if len(missing) > 0 {
		utils.DebugTrace(pt.logger, err)
		return errorz.ErrMissingTransactions
	}
	if len(spent) > 0 {
		utils.DebugTrace(pt.logger, err)
		return errorz.ErrInvalid{}.New("ptHandler.checkIsValid; spent")
	}
	_, err = pt.UTXOHandler.IsValid(txn, txs, currentHeight, deposits)
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return err
	}
	return nil
}

func (pt *Handler) getOneInternal(txn *badger.Txn, epoch uint32, txHash []byte) (*objs.Tx, error) {
	expEpoch, err := pt.indexer.GetEpoch(txn, txHash)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			utils.DebugTrace(pt.logger, err)
			return nil, err
		}
		return nil, errorz.ErrMissingTransactions
	}
	if expEpoch < epoch {
		utils.DebugTrace(pt.logger, err)
		return nil, errorz.ErrInvalid{}.New("ptHandler.getOneInternal; expired")
	}
	key := pt.makePendingTxKey(txHash)
	tx, err := db.GetTx(txn, key)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			utils.DebugTrace(pt.logger, err)
			return nil, err
		}
		utils.DebugTrace(pt.logger, err)
		return nil, errorz.ErrMissingTransactions
	}
	return tx, nil
}

func (pt *Handler) addOneInternal(txn *badger.Txn, tx *objs.Tx, expEpoch uint32, txHash []byte, utxoIDs [][]byte) error {
	contains, err := pt.containsOneInternal(txn, expEpoch, txHash)
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return err
	}
	if contains {
		return nil
	}
	evicted, err := pt.indexer.Add(txn, expEpoch, txHash, utxoIDs)
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		return err
	}
	for _, evictedHash := range evicted {
		err := pt.deleteOneInternal(txn, utils.CopySlice(evictedHash), false)
		if err != nil {
			utils.DebugTrace(pt.logger, err)
			return err
		}
	}
	key := pt.makePendingTxKey(txHash)
	if err := db.SetTx(txn, key, tx); err != nil {
		utils.DebugTrace(pt.logger, err)
		return err
	}
	return nil
}

func (pt *Handler) deleteOneInternal(txn *badger.Txn, txHash []byte, minedDelete bool) error {
	if minedDelete {
		txHashes, _, err := pt.indexer.DeleteMined(txn, txHash)
		if err != nil {
			utils.DebugTrace(pt.logger, err)
			return err
		}
		for j := 0; j < len(txHashes); j++ {
			txH := utils.CopySlice(txHashes[j])
			key := pt.makePendingTxKey(utils.CopySlice(txH))
			err := utils.DeleteValue(txn, key)
			if err != nil {
				if err != badger.ErrKeyNotFound {
					utils.DebugTrace(pt.logger, err)
					return err
				}
			}
		}
	} else {
		err := pt.indexer.DeleteOne(txn, txHash)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				utils.DebugTrace(pt.logger, err)
				return err
			}
		}
	}
	key := pt.makePendingTxKey(txHash)
	err := utils.DeleteValue(txn, key)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			utils.DebugTrace(pt.logger, err)
			return err
		}
	}
	return nil
}

func (pt *Handler) containsOneInternal(txn *badger.Txn, epoch uint32, txHash []byte) (bool, error) {
	_, err := pt.getOneInternal(txn, epoch, txHash)
	if err != nil {
		if err != errorz.ErrMissingTransactions {
			utils.DebugTrace(pt.logger, err)
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func (pt *Handler) makePendingTxKey(txHash []byte) []byte {
	key := dbprefix.PrefixPendingTx()
	key = append(key, txHash...)
	return key
}

func (pt *Handler) makePendingTxCooldownKey(txHash []byte) []byte {
	txHashCopy := utils.CopySlice(txHash)
	key := dbprefix.PrefixPendingTxCooldownKey()
	key = append(key, txHashCopy...)
	return key
}

// Drop deletes all data from the pending tx pool
func (pt *Handler) Drop() error {
	return pt.db.DropAll()
}
