package application

import (
	"context"
	"time"

	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"

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
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
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

func (tm *txHandler) GetTxsForGossip(txnState *badger.Txn, currentHeight uint32) ([]*objs.Tx, error) {
	ctx := context.Background()
	subCtx, cf := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cf()
	utxos, err := tm.pTxHdlr.GetTxsForGossip(txnState, subCtx, currentHeight, tm.storage.GetMaxBytes())
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, err
	}
	return utxos, nil
}

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

func (tm *txHandler) ApplyState(txn *badger.Txn, chainID uint32, height uint32, tx []*objs.Tx) ([]byte, error) {
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

func (tm *txHandler) GetTxsForProposal(txn *badger.Txn, chainID uint32, height uint32, curveSpec constants.CurveSpec, signer objs.Signer, maxBytes uint32) (objs.TxVec, []byte, error) {
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
			utils.DebugTrace(tm.logger, err)
			return nil, nil, err
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

func (tm *txHandler) GetStateRootForProposal(txn *badger.Txn, tx []*objs.Tx) ([]byte, error) {
	return tm.uHdlr.GetStateRootForProposal(txn, tx)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
//Data Getters/Setters/RPC methods//////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (tm *txHandler) PendingTxAdd(txn *badger.Txn, chainID uint32, height uint32, tx []*objs.Tx) error {
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

func (tm *txHandler) PendingTxGet(txn *badger.Txn, height uint32, txHash [][]byte) ([]*objs.Tx, [][]byte, error) {
	return tm.pTxHdlr.Get(txn, height, txHash)
}

func (tm *txHandler) PendingTxContains(txn *badger.Txn, height uint32, txHash [][]byte) ([][]byte, error) {
	return tm.pTxHdlr.Contains(txn, height, txHash)
}

func (tm *txHandler) UTXOContains(txn *badger.Txn, utxoID []byte) (bool, error) {
	return tm.uHdlr.Contains(txn, utxoID)
}

func (tm *txHandler) UTXOGetData(txn *badger.Txn, owner *objs.Owner, dataIdx []byte) ([]byte, error) {
	return tm.uHdlr.GetData(txn, owner, dataIdx)
}

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

func (tm *txHandler) GetHeightForTx(txn *badger.Txn, txHash []byte) (uint32, error) {
	return tm.mTxHdlr.GetHeightForTx(txn, txHash)
}

func (tm *txHandler) StoreSnapShotNode(txn *badger.Txn, batch []byte, root []byte, layer int) ([][]byte, int, []trie.LeafNode, error) {
	return tm.uHdlr.StoreSnapShotNode(txn, batch, root, layer)
}

func (tm *txHandler) FinalizeSnapShotRoot(txn *badger.Txn, root []byte, height uint32) error {
	return tm.uHdlr.FinalizeSnapShotRoot(txn, root, height)
}

func (tm *txHandler) GetSnapShotNode(txn *badger.Txn, height uint32, key []byte) ([]byte, error) {
	return tm.uHdlr.GetSnapShotNode(txn, height, key)
}

func (tm *txHandler) StoreSnapShotStateData(txn *badger.Txn, key []byte, value []byte, data []byte) error {
	return tm.uHdlr.StoreSnapShotStateData(txn, key, value, data)
}

func (tm *txHandler) FinalizeSync(txn *badger.Txn) error {
	if err := tm.pTxHdlr.Drop(); err != nil {
		return err
	}
	return nil
}

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
