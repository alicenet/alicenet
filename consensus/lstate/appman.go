package lstate

import (
	"bytes"
	"errors"

	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

// AddPendingTx ...
func (ce *Engine) AddPendingTx(txn *badger.Txn, d []interfaces.Transaction) error {
	os, err := ce.database.GetOwnState(txn)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	err = ce.appHandler.PendingTxAdd(txn, os.SyncToBH.BClaims.ChainID, os.SyncToBH.BClaims.Height+1, d)
	if err != nil {
		e := errorz.ErrInvalid{}.New("already  mined")
		if errors.As(err, &e) {
			utils.DebugTrace(ce.logger, err)
			return err
		}
	}
	return nil
}

func (ce *Engine) getValidValue(txn *badger.Txn, rs *RoundStates) ([][]byte, []byte, []byte, []byte, error) {
	chainID := rs.OwnState.SyncToBH.BClaims.ChainID
	txs, stateRoot, err := ce.appHandler.GetValidProposal(txn, chainID, rs.OwnState.SyncToBH.BClaims.Height+1, ce.storage.GetMaxProposalSize())
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return nil, nil, nil, nil, err
	}
	if err := ce.dm.AddTxs(txn, rs.OwnState.SyncToBH.BClaims.Height+1, txs); err != nil {
		utils.DebugTrace(ce.logger, err)
		return nil, nil, nil, nil, err
	}
	txHashes := make([][]byte, len(txs))
	for i := 0; i < len(txs); i++ {
		tx := txs[i]
		txHash, err := tx.TxHash()
		if err != nil {
			return nil, nil, nil, nil, err
		}
		txHashes[i] = txHash
	}
	txRootHash, err := objs.MakeTxRoot(txHashes)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return nil, nil, nil, nil, err
	}
	headerRoot, err := ce.database.GetHeaderTrieRoot(txn, rs.OwnState.SyncToBH.BClaims.Height)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			utils.DebugTrace(ce.logger, err)
			return nil, nil, nil, nil, err
		}
		headerRoot = make([]byte, constants.HashLen)
	}
	return txHashes, txRootHash, stateRoot, headerRoot, nil
}

func (ce *Engine) isValid(txn *badger.Txn, rs *RoundStates, chainID uint32, stateHash []byte, headerRoot []byte, txs []interfaces.Transaction) (bool, error) {
	goodHeaderRoot, err := ce.database.GetHeaderTrieRoot(txn, rs.OwnState.SyncToBH.BClaims.Height)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	if !bytes.Equal(goodHeaderRoot, headerRoot) {
		utils.DebugTrace(ce.logger, err)
		return false, nil
	}
	ok, err := ce.appHandler.IsValid(txn, chainID, rs.OwnState.SyncToBH.BClaims.Height+1, stateHash, txs)
	if err != nil {
		e := errorz.ErrInvalid{}.New("")
		if errors.As(err, &e) {
			return false, nil
		}
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	if !ok {
		return false, errorz.ErrInvalid{}.New("is valid returned not ok")
	}
	if err := ce.dm.AddTxs(txn, rs.OwnState.SyncToBH.BClaims.Height+1, txs); err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	return true, nil
}

func (ce *Engine) applyState(txn *badger.Txn, rs *RoundStates, chainID uint32, txHashes [][]byte) error {
	txs, missing, err := ce.dm.GetTxs(txn, rs.OwnState.SyncToBH.BClaims.Height+1, rs.round, txHashes)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return errorz.ErrMissingTransactions
	}
	if err := ce.dm.AddTxs(txn, rs.OwnState.SyncToBH.BClaims.Height+1, txs); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	_, err = ce.appHandler.ApplyState(txn, chainID, rs.OwnState.SyncToBH.BClaims.Height+1, txs)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}
