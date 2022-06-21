package application

import (
	"bytes"
	"errors"

	"github.com/alicenet/alicenet/dynamics"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
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
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/logging"
	"github.com/dgraph-io/badger/v2"
)

var _ interfaces.Application = (*Application)(nil)

// Application ...
type Application struct {
	logger           *logrus.Logger
	txHandler        *txHandler
	defaultSigner    objs.Signer
	defaultCurveSpec constants.CurveSpec
	defaultAccount   []byte
}

// Init initializes Application ...
func (a *Application) Init(conDB *consensusdb.Database, memDB *badger.DB, dph *deposit.Handler, storageInterface dynamics.StorageGetter) error {
	a.logger = logging.GetLogger(constants.LoggerApp)
	storage := wrapper.NewStorage(storageInterface)
	uHdlr := utxohandler.NewUTXOHandler(conDB.DB())
	pHdlr := pendingtx.NewPendingTxHandler(memDB)
	pHdlr.UTXOHandler = uHdlr
	pHdlr.DepositHandler = dph
	a.txHandler = &txHandler{
		db:      conDB.DB(),
		logger:  logging.GetLogger(constants.LoggerApp),
		pTxHdlr: pHdlr,
		mTxHdlr: minedtx.NewMinedTxHandler(),
		dHdlr:   dph,
		uHdlr:   uHdlr,
		cdb:     conDB,
		storage: storage,
	}
	a.txHandler.dHdlr.IsSpent = a.txHandler.uHdlr.TrieContains
	// initialize the application with a random key.
	// this will be over-written before first use in
	// state modifying logic, but is created here to ensure
	// that a race does not cause spurious failures for read
	// operations that occur before being set
	rkey, err := utils.RandomBytes(32)
	if err != nil {
		return err
	}
	err = a.SetMiningKey(rkey, 1)
	if err != nil {
		return err
	}
	return nil
}

var _ interfaces.Transaction = (*objs.Tx)(nil)

// UnmarshalTx allows a transaction to be unmarshalled into a transaction
// interface for use by the consensus algorithm
func (a *Application) UnmarshalTx(txb []byte) (interfaces.Transaction, error) {
	tx := &objs.Tx{}
	err := tx.UnmarshalBinary(txb)
	if err != nil {
		utils.DebugTrace(a.logger, err)
		return nil, err
	}
	return tx, nil
}

func (a *Application) convertTxToIface(txs []*objs.Tx) []interfaces.Transaction {
	out := make([]interfaces.Transaction, len(txs))
	for i := 0; i < len(txs); i++ {
		out[i] = txs[i]
	}
	return out
}

func (a *Application) convertIfaceToTx(txs []interfaces.Transaction) ([]*objs.Tx, bool) {
	out := make([]*objs.Tx, len(txs))
	for i := 0; i < len(txs); i++ {
		tt, ok := txs[i].(*objs.Tx)
		if !ok {
			return nil, false
		}
		out[i] = tt
	}
	return out, true
}

// GetTxsForGossip returns a list of transactions that should be gossipped
func (a *Application) GetTxsForGossip(txnState *badger.Txn, currentHeight uint32) ([]interfaces.Transaction, error) {
	r, err := a.txHandler.GetTxsForGossip(txnState, currentHeight)
	if err != nil {
		utils.DebugTrace(a.logger, err)
		return nil, err
	}
	return a.convertTxToIface(r), nil
}

// IsValid returns true if the list of transactions is a valid transformation
// and false if the list is not valid. If an error is returned, it indicates
// a low level failure that should stop the main loop.
func (a *Application) IsValid(txn *badger.Txn, chainID uint32, height uint32, stateHash []byte, txi []interfaces.Transaction) (bool, error) {
	tx, ok := a.convertIfaceToTx(txi)
	if !ok {
		return false, errorz.ErrCorrupt
	}
	if len(tx) == 0 {
		stateRoot, err := a.txHandler.GetStateRootForProposal(txn, tx)
		if err != nil {
			utils.DebugTrace(a.logger, err)
			return false, err
		}
		if !bytes.Equal(stateHash, stateRoot) {
			utils.DebugTrace(a.logger, err)
			return false, nil
		}
		return true, nil
	}
	txs := objs.TxVec(tx)
	if err := txs.PreValidateApplyState(chainID); err != nil {
		utils.DebugTrace(a.logger, err)
		return false, err
	}
	vout, err := a.txHandler.IsValid(txn, tx, height)
	if err != nil {
		utils.DebugTrace(a.logger, err)
		return false, err
	}
	if err := txs.Validate(height, vout, a.txHandler.storage); err != nil {
		utils.DebugTrace(a.logger, err)
		return false, err
	}
	stateRoot, err := a.txHandler.GetStateRootForProposal(txn, tx)
	if err != nil {
		e := errorz.ErrInvalid{}.New("")
		if errors.As(err, &e) {
			utils.DebugTrace(a.logger, err)
			return false, nil
		}
		utils.DebugTrace(a.logger, err)
		return false, err
	}
	if stateHash == nil {
		stateHash = make([]byte, constants.HashLen)
	}
	if !bytes.Equal(stateHash, stateRoot) {
		utils.DebugTrace(a.logger, err)
		return false, nil
	}
	return true, nil
}

// GetValidProposal is a function that returns a list of transactions
// that will cause a valid state transition function for the local node's
// current state. This is the function used to create a new proposal.
// comes from application logic
func (a *Application) GetValidProposal(txn *badger.Txn, chainID, height, maxBytes uint32) ([]interfaces.Transaction, []byte, error) {
	r, h, err := a.txHandler.GetTxsForProposal(txn, chainID, height, a.defaultCurveSpec, a.defaultSigner, maxBytes)
	if err != nil {
		utils.DebugTrace(a.logger, err)
		return nil, nil, err
	}
	return a.convertTxToIface(r), h, nil
}

// ApplyState is a function that returns a list of transactions
// that will cause a valid state transition function for the local node's
// current state. This is the function used to create a new proposal.
// comes from application logic
func (a *Application) ApplyState(txn *badger.Txn, chainID uint32, height uint32, txs []interfaces.Transaction) (stateHash []byte, err error) {
	tx, ok := a.convertIfaceToTx(txs)
	if !ok {
		return nil, errorz.ErrMissingTransactions
	}
	return a.txHandler.ApplyState(txn, chainID, height, tx)
}

// PendingTxAdd adds a transaction to the txPool and cleans up any stale
// tx as a result.
func (a *Application) PendingTxAdd(txn *badger.Txn, chainID uint32, height uint32, txs []interfaces.Transaction) error {
	tx, ok := a.convertIfaceToTx(txs)
	if !ok {
		return errorz.ErrMissingTransactions
	}
	return a.txHandler.PendingTxAdd(txn, chainID, height, tx)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
//Data Getters/Setters/RPC methods//////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// SetMiningKey updates the mining key. This key is used for collecting
// block mining rewards/fees.
func (a *Application) SetMiningKey(privKey []byte, curveSpec constants.CurveSpec) error {
	switch curveSpec {
	case constants.CurveBN256Eth:
		signer := &crypto.BNSigner{}
		err := signer.SetPrivk(privKey)
		if err != nil {
			utils.DebugTrace(a.logger, err)
			return err
		}
		pubk, err := signer.Pubkey()
		if err != nil {
			utils.DebugTrace(a.logger, err)
			return err
		}
		ac := crypto.GetAccount(pubk)
		a.defaultSigner = signer
		a.defaultAccount = ac
		a.defaultCurveSpec = curveSpec
	case constants.CurveSecp256k1:
		signer := &crypto.Secp256k1Signer{}
		err := signer.SetPrivk(privKey)
		if err != nil {
			utils.DebugTrace(a.logger, err)
			return err
		}
		pubk, err := signer.Pubkey()
		if err != nil {
			utils.DebugTrace(a.logger, err)
			return err
		}
		ac := crypto.GetAccount(pubk)
		a.defaultSigner = signer
		a.defaultAccount = ac
		a.defaultCurveSpec = curveSpec
	default:
		return errorz.ErrInvalid{}.New("app.SetMiningKey; invalid curve spec")
	}
	return nil
}

// MinedTxGet returns a list of mined transactions and a list of missing
// transaction hashes for mined transactions
func (a *Application) MinedTxGet(txn *badger.Txn, txHash [][]byte) ([]interfaces.Transaction, [][]byte, error) {
	r, m, err := a.txHandler.MinedTxGet(txn, txHash)
	if err != nil {
		utils.DebugTrace(a.logger, err)
		return nil, nil, err
	}
	return a.convertTxToIface(r), m, nil
}

// PendingTxGet returns a list of transactions and a list of missing
// transaction hashes from the pending transaction pool
func (a *Application) PendingTxGet(txn *badger.Txn, height uint32, txHashes [][]byte) ([]interfaces.Transaction, [][]byte, error) {
	r, m, err := a.txHandler.PendingTxGet(txn, height, txHashes)
	if err != nil {
		utils.DebugTrace(a.logger, err)
		return nil, nil, err
	}
	return a.convertTxToIface(r), m, nil
}

// PendingTxContains returns a list of missing transaction hashes
// from the pending tx pool
func (a *Application) PendingTxContains(txn *badger.Txn, height uint32, txHashes [][]byte) ([][]byte, error) {
	return a.txHandler.PendingTxContains(txn, height, txHashes)
}

// UTXOContains returns true if the passed UTXOID is known and associated with
// a UTXO
func (a *Application) UTXOContains(txn *badger.Txn, utxoID []byte) (bool, error) {
	return a.txHandler.UTXOContains(txn, utxoID)
}

// UTXOGetData returns the data from a DataStore UTXO
func (a *Application) UTXOGetData(txn *badger.Txn, curveSpec constants.CurveSpec, account []byte, dataIdx []byte) ([]byte, error) {
	owner := &objs.Owner{}
	err := owner.New(account, curveSpec)
	if err != nil {
		utils.DebugTrace(a.logger, err)
		return nil, err
	}
	return a.txHandler.UTXOGetData(txn, owner, dataIdx)
}

func (a *Application) GetValueForOwner(txn *badger.Txn, curveSpec constants.CurveSpec, account []byte, minValue *uint256.Uint256, ptBytes []byte) ([][]byte, *uint256.Uint256, *objs.PaginationToken, error) {
	owner := &objs.Owner{}
	err := owner.New(account, curveSpec)
	if err != nil {
		utils.DebugTrace(a.logger, err)
		return nil, nil, nil, err
	}

	var pt *objs.PaginationToken
	if ptBytes != nil {
		pt = &objs.PaginationToken{}
		err := pt.UnmarshalBinary(ptBytes)
		if err != nil {
			utils.DebugTrace(a.logger, err)
			return nil, nil, nil, err
		}
	}

	return a.txHandler.GetValueForOwner(txn, owner, minValue, pt)
}

// UTXOGet returns a list of UTXO objects
func (a *Application) UTXOGet(txn *badger.Txn, utxoIDs [][]byte) ([]*objs.TXOut, error) {
	return a.txHandler.UTXOGet(txn, utxoIDs)
}

// PaginateDataByOwner returns a list of UTXOIDs and indexes from an account
// namespace
func (a *Application) PaginateDataByOwner(txn *badger.Txn, curveSpec constants.CurveSpec, account []byte, height uint32, numItems int, startIndex []byte) ([]*objs.PaginationResponse, error) {
	owner := &objs.Owner{}
	err := owner.New(account, curveSpec)
	if err != nil {
		utils.DebugTrace(a.logger, err)
		return nil, err
	}
	return a.txHandler.PaginateDataByOwner(txn, owner, height, numItems, startIndex)
}

// GetHeightForTx returns the height at which a tx was mined
func (a *Application) GetHeightForTx(txn *badger.Txn, txHash []byte) (uint32, error) {
	return a.txHandler.GetHeightForTx(txn, txHash)
}

// Cleanup does nothing at this time
func (a *Application) Cleanup() error {
	return nil
}

// StoreSnapShotNode will store a node of the state trie during fast sync
func (a *Application) StoreSnapShotNode(txn *badger.Txn, batch []byte, root []byte, layer int) ([][]byte, int, []trie.LeafNode, error) {
	return a.txHandler.StoreSnapShotNode(txn, batch, root, layer)
}

// GetSnapShotStateData will return a UTXO for snapshot fast sync
func (a *Application) GetSnapShotStateData(txn *badger.Txn, utxoID []byte) ([]byte, error) {
	utxo, err := a.txHandler.GetSnapShotStateData(txn, [][]byte{utxoID})
	if err != nil {
		utils.DebugTrace(a.logger, err)
		return nil, err
	}
	if len(utxo) != 1 {
		return nil, badger.ErrKeyNotFound
	}
	utxoBytes, err := utxo[0].MarshalBinary()
	if err != nil {
		utils.DebugTrace(a.logger, err)
		return nil, err
	}
	return utxoBytes, nil
}

// GetSnapShotNode will return a node from the state trie
func (a *Application) GetSnapShotNode(txn *badger.Txn, height uint32, key []byte) ([]byte, error) {
	return a.txHandler.GetSnapShotNode(txn, height, key)
}

// StoreSnapShotStateData stores fast sync state
func (a *Application) StoreSnapShotStateData(txn *badger.Txn, key []byte, value []byte, data []byte) error {
	return a.txHandler.StoreSnapShotStateData(txn, key, value, data)
}

// FinalizeSnapShotRoot will complete a snapshot fast sync by setting the trie
// root lookupkeys for the state trie
func (a *Application) FinalizeSnapShotRoot(txn *badger.Txn, root []byte, height uint32) error {
	return a.txHandler.FinalizeSnapShotRoot(txn, root, height)
}

// BeginSnapShotSync drops all pending txs and state data before fast sync
func (a *Application) BeginSnapShotSync(txn *badger.Txn) error {
	return a.txHandler.BeginSnapShotSync(txn)
}

// FinalizeSync will finalize state after a fast sync
func (a *Application) FinalizeSync(txn *badger.Txn) error {
	return a.txHandler.FinalizeSync(txn)
}
