package application

import (
	"context"
	"time"

	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"

	"github.com/MadBase/MadNet/application/deposit"
	"github.com/MadBase/MadNet/application/minedtx"
	"github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/application/pendingtx"
	"github.com/MadBase/MadNet/application/utxohandler"
	trie "github.com/MadBase/MadNet/badgerTrie"
	consensusdb "github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/constants/dbprefix"
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
}

func (tm *txHandler) GetTxsForGossip(txnState *badger.Txn, currentHeight uint32) ([]*objs.Tx, error) {
	ctx := context.Background()
	subCtx, cf := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cf()
	return tm.pTxHdlr.GetTxsForGossip(txnState, subCtx, currentHeight, constants.MaxBytes)
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
	if err := txs.Validate(height, vout); err != nil {
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
	tx, err := tm.uHdlr.GetExpiredForProposal(txn, subCtx, chainID, height, curveSpec, signer, maxBytes)
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
		if err := tx.PostValidatePending(height, consumedUTXOs); err != nil {
			utils.DebugTrace(tm.logger, err)
			return nil, nil, err
		}
		if err := tm.PendingTxAdd(txn, chainID, height, []*objs.Tx{tx}); err != nil {
			utils.DebugTrace(tm.logger, err)
			return nil, nil, err
		}
	}
	txs, err := tm.pTxHdlr.GetTxsForProposal(txn, subCtx, height, maxBytes, tx)
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
		return nil, nil, errorz.ErrInvalid{}.New("missing transactions")
	}
	if len(spent) > 0 {
		return nil, nil, errorz.ErrInvalid{}.New("spent transactions")
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
		return errorz.ErrInvalid{}.New("duplicate")
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
	if err := txs.PostValidatePending(height, consumedUTXOs); err != nil {
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

func (tm *txHandler) GetValueForOwner(txn *badger.Txn, owner *objs.Owner, minValue *uint256.Uint256) ([][]byte, *uint256.Uint256, error) {
	var u [][]byte
	v := uint256.Zero()
	utxoIDs, value, err := tm.uHdlr.GetValueForOwner(txn, owner, minValue)
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, nil, err
	}
	v, err = v.Clone().Add(v.Clone(), value.Clone())
	if err != nil {
		utils.DebugTrace(tm.logger, err)
		return nil, nil, err
	}
	u = append(u, utxoIDs...)
	if v.Lt(minValue) {
		remainder, err := new(uint256.Uint256).Sub(minValue.Clone(), v.Clone())
		if err != nil {
			utils.DebugTrace(tm.logger, err)
			return nil, nil, err
		}
		utxoIDs, value, err := tm.dHdlr.GetValueForOwner(txn, owner, remainder)
		if err != nil {
			utils.DebugTrace(tm.logger, err)
			return nil, nil, err
		}
		v, err = v.Clone().Add(v.Clone(), value.Clone())
		if err != nil {
			utils.DebugTrace(tm.logger, err)
			return nil, nil, err
		}
		u = append(u, utxoIDs...)
	}
	return u, v, nil
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
