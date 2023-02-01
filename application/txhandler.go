package application

import (
	"context"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"

	"github.com/alicenet/alicenet/application/deposit"
	"github.com/alicenet/alicenet/application/minedtx"
	"github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/application/pendingtx"
	"github.com/alicenet/alicenet/application/utxohandler"
	"github.com/alicenet/alicenet/application/wrapper"
	trie "github.com/alicenet/alicenet/badgerTrie"
	consensusdb "github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
)

// TODO SET UP PRUNING

type txHandler struct {
	logger  *logrus.Logger
	db      *badger.DB
	cdb     *consensusdb.Database
	pTxHdlr *pendingtx.Handler
	mTxHdlr *minedtx.MinedTxHandler
	dHdlr   *deposit.Handler
	uHdlr   *utxohandler.UTXOHandler
	storage *wrapper.Storage
}

// GetTxsForGossip collects old, non-expired transactions
// in order to gossip them to other peers for inclusion.
func (tm *txHandler) GetTxsForGossip(txnState *badger.Txn, currentHeight uint32) ([]*objs.Tx, error) {
	ctx := context.Background()
	subCtx, cf := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cf()
	maxBytes, err := tm.storage.GetMaxBlockSize()
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	utxos, err := tm.pTxHdlr.GetTxsForGossip(txnState, subCtx, currentHeight, maxBytes)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	return utxos, nil
}

// IsValid returns the list of deposits and consumed UTXOs
// in addition to checking if the list of transactions
// constitute a valid state transition.
func (tm *txHandler) IsValid(txn *badger.Txn, tx []*objs.Tx, currentHeight uint32) (objs.Vout, error) {
	txs := objs.TxVec(tx)
	deposits, err := tm.dHdlr.IsValid(txn, txs)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	consumed, err := tm.uHdlr.IsValid(txn, txs, currentHeight, deposits)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	vout := objs.Vout{}
	vout = append(vout, deposits...)
	vout = append(vout, consumed...)
	return vout, nil
}

// ApplyState updates the state trie from the list of transactions
// and returns the resulting root hash of the state trie.
func (tm *txHandler) ApplyState(txn *badger.Txn, chainID, height uint32, tx []*objs.Tx) ([]byte, error) {
	if len(tx) == 0 {
		hsh, err := tm.uHdlr.ApplyState(txn, tx, height)
		if err != nil {
			utils.DebugTrace(tm.logger, err)
			return nil, err
		}
		return hsh, nil
	}
	txs := objs.TxVec(tx)
	if err := txs.PreValidateApplyState(chainID); err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	vout, err := tm.IsValid(txn, tx, height)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	if err := txs.Validate(height, vout, tm.storage); err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	rootHash, err := tm.uHdlr.ApplyState(txn, txs, height)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	if err := tm.mTxHdlr.Add(txn, height, txs); err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	txHashes, err := txs.TxHash()
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	if err := tm.pTxHdlr.DeleteMined(txn, height, txHashes); err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	return rootHash, nil
}

// GetTxsForProposal returns a list of transactions which result
// in a valid state transition and the new StateRoot (root hash of the state trie).
func (tm *txHandler) GetTxsForProposal(txn *badger.Txn, chainID, height uint32, curveSpec constants.CurveSpec, signer objs.Signer, maxBytes uint32) (objs.TxVec, []byte, error) {
	ctx := context.Background()
	subCtx, cf := context.WithTimeout(ctx, 1*time.Second)
	defer cf()
	tx, maxBytes, err := tm.uHdlr.GetExpiredForProposal(txn, subCtx, chainID, height, curveSpec, signer, maxBytes, tm.storage)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, nil, errorz.ErrInvalid{}.New(err.Error())
	}
	// only perform checks if the tx is not nil
	// if it is nil, the call to tm.pTxHdlr.GetTxsForProposal
	// will filter out the nil tx
	if tx != nil {
		consumedUTXOs, err := tm.IsValid(txn, []*objs.Tx{tx}, height)
		if err != nil {
			utils.DebugTrace(tm.logger, err)
			return nil, nil, err
		}
		if err := tx.PreValidatePending(chainID); err != nil {
			utils.DebugTrace(tm.logger, err)
			return nil, nil, err
		}
		if err := tx.PostValidatePending(height, consumedUTXOs, tm.storage); err != nil {
			utils.DebugTrace(tm.logger, err)
			return nil, nil, err
		}
		if err := tm.PendingTxAdd(txn, chainID, height, []*objs.Tx{tx}); err != nil {
			if !strings.Contains(err.Error(), "duplicate") {
				utils.DebugTrace(tm.logger, err)
				return nil, nil, err
			}
		}
	}
	txs, _, err := tm.pTxHdlr.GetTxsForProposal(txn, subCtx, height, maxBytes, tx)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, nil, errorz.ErrInvalid{}.New(err.Error())
	}
	stateRoot, err := tm.uHdlr.GetStateRootForProposal(txn, txs)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, nil, err
	}
	consumedDeposits, err := txs.ConsumedUTXOIDOnlyDeposits()
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, nil, err
	}
	found, missing, spent, err := tm.dHdlr.Get(txn, consumedDeposits)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, nil, err
	}
	if len(missing) > 0 {
		return nil, nil, errorz.ErrInvalid{}.New("txhandler.GetTxsForProposal; missing transactions")
	}
	if len(spent) > 0 {
		return nil, nil, errorz.ErrInvalid{}.New("txhandler.GetTxsForProposal; spent transactions")
	}
	if _, err := tm.uHdlr.IsValid(txn, txs, height, found); err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, nil, err
	}
	for i := 0; i < len(txs); i++ {
		tx := txs[i]
		txb, err := tx.MarshalBinary()
		if err != nil {
			utils.DebugTrace(tm.logger, err)
			return nil, nil, err
		}
		if err := tm.cdb.SetBroadcastTransaction(txn, txb); err != nil {
			utils.DebugTrace(tm.logger, err)
			return nil, nil, err
		}
	}
	return txs, stateRoot, nil
}

// GetStateRootForProposal returns the resulting StateRoot
// (root hash of the state trie) after applying the transactions;
// the StateRoot is just computed, *not* applied.
func (tm *txHandler) GetStateRootForProposal(txn *badger.Txn, tx []*objs.Tx) ([]byte, error) {
	return tm.uHdlr.GetStateRootForProposal(txn, tx)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
//Data Getters/Setters/RPC methods//////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// PendingTxAdd adds a list of transactions to the pending transaction pool.
func (tm *txHandler) PendingTxAdd(txn *badger.Txn, chainID, height uint32, tx []*objs.Tx) error {
	txs := objs.TxVec(tx)
	if err := txs.PreValidatePending(chainID); err != nil {
		utils.DebugTrace(tm.logger, err)
		return err
	}
	txHashes, err := txs.TxHash()
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return err
	}
	missing, err := tm.pTxHdlr.Contains(txn, height, txHashes)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return err
	}
	if len(missing) == 0 {
		return errorz.ErrInvalid{}.New("txhandler.PendingTxAdd; duplicate")
	}
	missingMap := make(map[string]int)
	for i := 0; i < len(txHashes); i++ {
		missingMap[string(txHashes[i])] = i
	}
	consumedUTXOs, err := tm.IsValid(txn, tx, height)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return err
	}
	if err := txs.PostValidatePending(height, consumedUTXOs, tm.storage); err != nil {
		utils.DebugTrace(tm.logger, err)
		return err
	}
	for i := 0; i < len(missing); i++ {
		idx := missingMap[string(missing[i])]
		txb, err := txs[idx].MarshalBinary()
		if err != nil {
			utils.DebugTrace(tm.logger, err)
			return err
		}
		if err := tm.cdb.SetBroadcastTransaction(txn, txb); err != nil {
			utils.DebugTrace(tm.logger, err)
			return err
		}
	}
	if err := tm.pTxHdlr.Add(txn, txs, height); err != nil {
		utils.DebugTrace(tm.logger, err)
		return err
	}
	return nil
}

func (tm *txHandler) MinedTxGet(txn *badger.Txn, txHash [][]byte) ([]*objs.Tx, [][]byte, error) {
	return tm.mTxHdlr.Get(txn, txHash)
}

// PendingTxGet returns a list of transactions and a list of missing
// transaction hashes from the pending transaction pool.
func (tm *txHandler) PendingTxGet(txn *badger.Txn, height uint32, txHash [][]byte) ([]*objs.Tx, [][]byte, error) {
	return tm.pTxHdlr.Get(txn, height, txHash)
}

// PendingTxContains returns a list of missing transactions from the list of txhashes;
// returns a list of txhashes.
func (tm *txHandler) PendingTxContains(txn *badger.Txn, height uint32, txHash [][]byte) ([][]byte, error) {
	return tm.pTxHdlr.Contains(txn, height, txHash)
}

// UTXOContains returns if a UTXO is in the state trie.
func (tm *txHandler) UTXOContains(txn *badger.Txn, utxoID []byte) (bool, error) {
	return tm.uHdlr.Contains(txn, utxoID)
}

// UTXOGetData returns the data stored in a utxo by owner and the data index.
func (tm *txHandler) UTXOGetData(txn *badger.Txn, owner *objs.Owner, dataIdx []byte) ([]byte, error) {
	return tm.uHdlr.GetData(txn, owner, dataIdx)
}

// GetValueForOwner returns a list of utxoIDs (of ValueStores) and the total returned value.
// The purpose of this function is to return a list UTXOs of a certain value
// which may be consumed within a transaction.
func (tm *txHandler) GetValueForOwner(txn *badger.Txn, owner *objs.Owner, minValue *uint256.Uint256, pt *objs.PaginationToken) ([][]byte, *uint256.Uint256, *objs.PaginationToken, error) {
	const maxCount = 256
	allIds := [][]byte{}

	totalValue := uint256.Zero()
	if pt != nil {
		var err error
		totalValue, err = totalValue.Add(totalValue, pt.TotalValue)
		if err != nil {
			utils.DebugTrace(tm.logger, err)
			return nil, nil, nil, err
		}
	}

	txTypes := []struct {
		retrieve func(*badger.Txn, *objs.Owner, *uint256.Uint256, int, []byte) ([][]byte, *uint256.Uint256, []byte, error)
		lpType   objs.LastPaginatedType
	}{
		{tm.uHdlr.GetValueForOwner, objs.LastPaginatedUtxo},
		{tm.dHdlr.GetValueForOwner, objs.LastPaginatedDeposit},
	}

	started := pt == nil
	for _, v := range txTypes {
		if totalValue.Gte(minValue) {
			break
		}

		var lastKey []byte
		if !started {
			if pt.LastPaginatedType == v.lpType {
				started = true
				lastKey = pt.LastKey
			} else {
				break
			}
		}

		remainder, err := new(uint256.Uint256).Sub(minValue, totalValue)
		if err != nil {
			break // underflow -> value exceeded
		}

		utxoIDs, value, lk, err := v.retrieve(txn, owner, remainder, maxCount-len(allIds), lastKey)
		if err != nil {
			utils.DebugTrace(tm.logger, err)
			return nil, nil, nil, err
		}

		totalValue, err = totalValue.Add(totalValue, value)
		if err != nil {
			utils.DebugTrace(tm.logger, err)
			return nil, nil, nil, err
		}

		allIds = append(allIds, utxoIDs...)

		if len(allIds) >= maxCount {
			return allIds, totalValue, &objs.PaginationToken{LastPaginatedType: v.lpType, TotalValue: totalValue, LastKey: lk}, nil
		}
	}

	return allIds, totalValue, nil, nil
}

// UTXOGet returns a list of UTXOs from a list of utxoIDs.
// The returned UTXOs are those which are present;
// any missing UTXOs are not specified.
func (tm *txHandler) UTXOGet(txn *badger.Txn, utxoIDs [][]byte) ([]*objs.TXOut, error) {
	f := []*objs.TXOut{}
	found, _, _, err := tm.dHdlr.Get(txn, utxoIDs)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	if len(found) > 0 {
		f = append(f, found...)
	}
	found2, _, err := tm.uHdlr.Get(txn, utxoIDs)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	if len(found2) > 0 {
		f = append(f, found2...)
	}
	return f, nil
}

// GetSnapShotStateData returns a list of found UTXOs (deposits and UTXOs) and spent deposits.
func (tm *txHandler) GetSnapShotStateData(txn *badger.Txn, utxoIDs [][]byte) ([]*objs.TXOut, error) {
	f := []*objs.TXOut{}
	found, _, spent, err := tm.dHdlr.Get(txn, utxoIDs)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	if len(found) > 0 {
		f = append(f, found...)
	}
	if len(spent) > 0 {
		f = append(f, spent...)
	}
	found2, _, err := tm.uHdlr.Get(txn, utxoIDs)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	if len(found2) > 0 {
		f = append(f, found2...)
	}
	return f, nil
}

func (tm *txHandler) PaginateDataByOwner(txn *badger.Txn, owner *objs.Owner, height uint32, numItems int, startIndex []byte) ([]*objs.PaginationResponse, error) {
	return tm.uHdlr.PaginateDataByOwner(txn, owner, height, numItems, startIndex)
}

// GetHeightForTx returns height for a mined transaction.
func (tm *txHandler) GetHeightForTx(txn *badger.Txn, txHash []byte) (uint32, error) {
	return tm.mTxHdlr.GetHeightForTx(txn, txHash)
}

func (tm *txHandler) StoreSnapShotNode(txn *badger.Txn, batch, root []byte, layer int) ([][]byte, int, []trie.LeafNode, error) {
	return tm.uHdlr.StoreSnapShotNode(txn, batch, root, layer)
}

// FinalizeSnapShotRoot stores the Current State Root, Pending State Root,
// and Canonical State Root of the State Trie.
func (tm *txHandler) FinalizeSnapShotRoot(txn *badger.Txn, root []byte, height uint32) error {
	return tm.uHdlr.FinalizeSnapShotRoot(txn, root, height)
}

func (tm *txHandler) GetSnapShotNode(txn *badger.Txn, height uint32, key []byte) ([]byte, error) {
	return tm.uHdlr.GetSnapShotNode(txn, height, key)
}

func (tm *txHandler) StoreSnapShotStateData(txn *badger.Txn, key, value, data []byte) error {
	return tm.uHdlr.StoreSnapShotStateData(txn, key, value, data)
}

// FinalizeSync drops all transactions from the pending transaction pool.
func (tm *txHandler) FinalizeSync(txn *badger.Txn) error {
	if err := tm.pTxHdlr.Drop(); err != nil {
		return err
	}
	return nil
}

// BeginSnapShotSync drops all pending txs and state data.
func (tm *txHandler) BeginSnapShotSync(txn *badger.Txn) error {
	if err := tm.pTxHdlr.Drop(); err != nil {
		return err
	}
	if err := tm.db.DropPrefix(dbprefix.PrefixMinedTx()); err != nil {
		return err
	}
	if err := tm.db.DropPrefix(dbprefix.PrefixMinedTxIndexRefKey()); err != nil {
		return err
	}
	if err := tm.db.DropPrefix(dbprefix.PrefixMinedTxIndexKey()); err != nil {
		return err
	}
	if err := tm.db.DropPrefix(dbprefix.PrefixMinedUTXO()); err != nil {
		return err
	}
	if err := tm.db.DropPrefix(dbprefix.PrefixMinedUTXOEpcKey()); err != nil {
		return err
	}
	if err := tm.db.DropPrefix(dbprefix.PrefixMinedUTXOEpcRefKey()); err != nil {
		return err
	}
	if err := tm.db.DropPrefix(dbprefix.PrefixMinedUTXODataKey()); err != nil {
		return err
	}
	if err := tm.db.DropPrefix(dbprefix.PrefixMinedUTXODataRefKey()); err != nil {
		return err
	}
	if err := tm.db.DropPrefix(dbprefix.PrefixMinedUTXOValueRefKey()); err != nil {
		return err
	}
	if err := tm.db.DropPrefix(dbprefix.PrefixMinedUTXOValueKey()); err != nil {
		return err
	}
	return nil
}
