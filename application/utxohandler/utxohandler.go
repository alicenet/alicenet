package utxohandler

import (
	"bytes"
	"context"
	"fmt"

	"github.com/MadBase/MadNet/constants/dbprefix"
	"github.com/MadBase/MadNet/errorz"

	"github.com/MadBase/MadNet/application/db"
	"github.com/MadBase/MadNet/application/indexer"
	"github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/application/utxohandler/utxotrie"
	trie "github.com/MadBase/MadNet/badgerTrie"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

// TODO: cleanup expIndex after fastSync (do in finalization logic)
// TODO: cleanup valueIndex after fastSync

// A DS CAN ONLY BE WRITTEN IF
//    THE OWNER INDEX DOES NOT ALREADY EXIST OR IS CONSUMED DURING THE INPUTS
//    THE OWNER INDEX IS A UNIQUE OUTPUT IN THE BATCH OF TXS
//TODO SET UP PRUNING

// NewUTXOHandler constructs a new UTXOHandler
func NewUTXOHandler(dB *badger.DB) *UTXOHandler {
	return &UTXOHandler{
		logger:     logging.GetLogger(constants.LoggerApp),
		trie:       utxotrie.NewUTXOTrie(dB),
		expIndex:   indexer.NewExpSizeIndex(dbprefix.PrefixMinedUTXOEpcKey, dbprefix.PrefixMinedUTXOEpcRefKey),
		dataIndex:  indexer.NewDataIndex(dbprefix.PrefixMinedUTXODataKey, dbprefix.PrefixMinedUTXODataRefKey),
		valueIndex: indexer.NewValueIndex(dbprefix.PrefixMinedUTXOValueKey, dbprefix.PrefixMinedUTXOValueRefKey),
		db:         dB,
	}
}

// UTXOHandler is the object that indexes and stores UTXOs. This object
// also manages the UTXOTrie.
type UTXOHandler struct {
	logger     *logrus.Logger
	db         *badger.DB
	trie       *utxotrie.UTXOTrie
	expIndex   *indexer.ExpSizeIndex
	dataIndex  *indexer.DataIndex
	valueIndex *indexer.ValueIndex
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
///////////WRAPPERS FOR TRIE////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Init will initialize the UTXOHandler
func (ut *UTXOHandler) Init(height uint32) error {
	return ut.trie.Init(height)
}

// IsValid verifies the rules of batches across transactions as is generated in
// a block
func (ut *UTXOHandler) IsValid(txn *badger.Txn, txs objs.TxVec, currentHeight uint32, deposits objs.Vout) (objs.Vout, error) {
	depositMap := make(map[string]*objs.TXOut)
	for i := 0; i < len(deposits); i++ {
		utxoID, err := deposits[i].UTXOID()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		depositMap[string(utxoID)] = deposits[i]
	}
	// check that the list of transactions does not contain more than one
	// reference to the same output data store index
	outputIndexesInitial := make(map[string]bool)
	for i := 0; i < len(txs); i++ {
		tmp, err := txs[i].ValidateDataStoreIndexes(outputIndexesInitial)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		outputIndexesInitial = tmp
	}

	// check that for each transaction, if a transaction references a
	// datastore index as an output, it must either not already exist
	// OR it must be consumed in the same transaction
	var tx *objs.Tx
	var utxo *objs.TXOut
	var key []byte
	for i := 0; i < len(txs); i++ {
		tx = txs[i]
		knownIndexes := make(map[string]bool)
		outputIndexes := make(map[string]bool)
		inputIndexes := make(map[string]bool)
		consumedUTXOIDs, err := objs.TxVec([]*objs.Tx{tx}).ConsumedUTXOIDNoDeposits()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		utxos, missing, err := ut.Get(txn, consumedUTXOIDs)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		if len(missing) > 0 {
			return nil, errorz.ErrInvalid{}.New("missing consumed utxo")
		}
		var refUTXOs objs.Vout
		consumedUTXOIDsOnlyDeposits, err := objs.TxVec([]*objs.Tx{tx}).ConsumedUTXOIDOnlyDeposits()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		for j := 0; j < len(consumedUTXOIDsOnlyDeposits); j++ {
			curUTXOID := utils.CopySlice(consumedUTXOIDsOnlyDeposits[j])
			deposit, ok := depositMap[string(curUTXOID)]
			if !ok {
				return nil, errorz.ErrInvalid{}.New("missing consumed utxo (deposit)")
			}
			refUTXOs = append(refUTXOs, deposit)
		}
		for j := 0; j < len(utxos); j++ {
			refUTXOs = append(refUTXOs, utxos[j])
		}
		if err := tx.ValidateEqualVinVout(refUTXOs, currentHeight); err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		for j := 0; j < len(utxos); j++ {
			utxo = utxos[j]
			if utxo.HasDataStore() {
				owner, err := utxo.GenericOwner()
				if err != nil {
					utils.DebugTrace(ut.logger, err)
					return nil, err
				}
				ownerBytes, err := owner.MarshalBinary()
				if err != nil {
					utils.DebugTrace(ut.logger, err)
					return nil, err
				}
				ds, err := utxo.DataStore()
				if err != nil {
					utils.DebugTrace(ut.logger, err)
					return nil, err
				}
				index, err := ds.Index()
				if err != nil {
					utils.DebugTrace(ut.logger, err)
					return nil, err
				}
				key = []byte{}
				key = append(key, utils.CopySlice(ownerBytes)...)
				key = append(key, utils.CopySlice(index)...)
				inputIndexes[string(key)] = true
			}
		}
		for j := 0; j < len(tx.Vout); j++ {
			utxo = tx.Vout[j]
			if utxo.HasDataStore() {
				owner, err := utxo.GenericOwner()
				if err != nil {
					utils.DebugTrace(ut.logger, err)
					return nil, err
				}
				ownerBytes, err := owner.MarshalBinary()
				if err != nil {
					utils.DebugTrace(ut.logger, err)
					return nil, err
				}
				ds, err := utxo.DataStore()
				if err != nil {
					utils.DebugTrace(ut.logger, err)
					return nil, err
				}
				index, err := ds.Index()
				if err != nil {
					utils.DebugTrace(ut.logger, err)
					return nil, err
				}
				key = []byte{}
				key = append(key, utils.CopySlice(ownerBytes)...)
				key = append(key, utils.CopySlice(index)...)
				outputIndexes[string(key)] = true
				ok, err := ut.dataIndex.Contains(txn, owner, utils.CopySlice(index))
				if err != nil {
					utils.DebugTrace(ut.logger, err)
					return nil, err
				}
				if ok {
					knownIndexes[string(key)] = true
				}
			}
		}
		// if we are generating a utxo with a data index and that index already
		// exists, we must also be consuming the utxo with that index in the same
		// transaction.
		for k := range outputIndexes {
			if knownIndexes[k] {
				if !inputIndexes[k] {
					return nil, errorz.ErrInvalid{}.New("duplicate datastore index")
				}
			}
		}
	}

	// for consumed deposits the trie must not contain the deposit already
	// as this means it is spent
	consumedDepositUTXOIDs, err := txs.ConsumedUTXOIDOnlyDeposits()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	for j := 0; j < len(consumedDepositUTXOIDs); j++ {
		ok, err := ut.TrieContains(txn, consumedDepositUTXOIDs[j])
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		if ok {
			return nil, errorz.ErrInvalid{}.New("double spend of deposit found in trie")
		}
	}

	// the trie must not contain any of the new UTXOIDs being created already
	// as this would cause a conflict with those UTXOIDs
	generatedUTXOIDs, err := txs.GeneratedUTXOID()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	for j := 0; j < len(generatedUTXOIDs); j++ {
		ok, err := ut.TrieContains(txn, utils.CopySlice(generatedUTXOIDs[j]))
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		if ok {
			return nil, errorz.ErrInvalid{}.New("utxoID already in trie")
		}
	}

	// check that all consumed utxos are in the trie already
	// IE they are available to be spent
	consumedUTXOIDs, err := txs.ConsumedUTXOIDNoDeposits()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	for j := 0; j < len(consumedUTXOIDs); j++ {
		ok, err := ut.TrieContains(txn, utils.CopySlice(consumedUTXOIDs[j]))
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		if !ok {
			return nil, errorz.ErrInvalid{}.New("consumed utxoID not in trie")
		}
	}
	utxos, missing, err := ut.Get(txn, consumedUTXOIDs)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	if len(missing) > 0 {
		return nil, errorz.ErrInvalid{}.New("missing transactions")
	}
	return utxos, nil
}

// ApplyState will update the state trie with the given proposal data.
// Consumed UTXOs will be deleted from the trie.
// New UTXOs will be added to the trie.
// Consumed deposits will be added to the trie.
func (ut *UTXOHandler) ApplyState(txn *badger.Txn, txs objs.TxVec, height uint32) ([]byte, error) {
	if len(txs) == 0 {
		hsh, err := ut.trie.ApplyState(txn, txs, height)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		return hsh, nil
	}
	consumedUTXOIDs, err := txs.ConsumedUTXOIDNoDeposits()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	for j := 0; j < len(consumedUTXOIDs); j++ {
		// TODO: VALIDATE NO REGRESSION DUE TO ERROR RETURN
		utxoID := utils.CopySlice(consumedUTXOIDs[j])
		err := ut.dropFromIndexes(txn, utils.CopySlice(utxoID))
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
	}
	newUTXOs, err := txs.GeneratedUTXOs()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	for _, utxo := range newUTXOs {
		err = ut.addOne(txn, utxo)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
	}
	stateRoot, err := ut.trie.ApplyState(txn, txs, height)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	return stateRoot, nil
}

// GetStateRootForProposal allows a new stateRoot to be calculated for a
// proposal without committing the changes to the trie.
func (ut *UTXOHandler) GetStateRootForProposal(txn *badger.Txn, txs objs.TxVec) ([]byte, error) {
	sr, err := ut.trie.GetStateRootForProposal(txn, txs)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	return sr, nil
}

// TrieContains allows a utxoID to be referenced in the trie. If the trie contains
// the utxoID, the result will be true, nil.
func (ut *UTXOHandler) TrieContains(txn *badger.Txn, utxoID []byte) (bool, error) {
	missing, err := ut.trie.Contains(txn, [][]byte{utxoID})
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return false, err
	}
	if len(missing) > 0 {
		return false, nil
	}
	return true, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
///////////OPERATORS ON UTXO STORAGE////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Contains returns if a UTXO is stored in storage.
func (ut *UTXOHandler) Contains(txn *badger.Txn, utxoID []byte) (bool, error) {
	key := ut.makeUTXOKey(utxoID)
	_, err := utils.GetValue(txn, key)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			utils.DebugTrace(ut.logger, err)
			return false, err
		}
		return false, nil
	}
	return true, nil
}

// Get returns a UTXO by utxoID.
func (ut *UTXOHandler) Get(txn *badger.Txn, utxoIDs [][]byte) ([]*objs.TXOut, [][]byte, error) {
	f := []*objs.TXOut{}
	m := [][]byte{}
	for i := 0; i < len(utxoIDs); i++ {
		utxoID := utils.CopySlice(utxoIDs[i])
		utxo, err := ut.getInternal(txn, utils.CopySlice(utxoID))
		if err != nil {
			if err != badger.ErrKeyNotFound {
				utils.DebugTrace(ut.logger, err)
				return nil, nil, err
			}
			m = append(m, utils.CopySlice(utxoID))
			continue
		}
		f = append(f, utxo)
	}
	return f, m, nil
}

// GetData returns the data stored in a utxo by owner and the data index.
func (ut *UTXOHandler) GetData(txn *badger.Txn, owner *objs.Owner, dataIdx []byte) ([]byte, error) {
	utxoID, err := ut.dataIndex.GetUTXOID(txn, owner, dataIdx)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	utxo, err := ut.getInternal(txn, utxoID)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	if utxo.HasDataStore() {
		ds, err := utxo.DataStore()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		rd, err := ds.RawData()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		return rd, nil
	}
	return nil, errorz.ErrInvalid{}.New("not a datastore")
}

// GetExpiredForProposal returns a list of UTXOs, the IDs of those UTXOs, and
// the total byte count of the returned UTXOs. This is used to collect expired
// dataStores for deletion.
func (ut *UTXOHandler) GetExpiredForProposal(txn *badger.Txn, ctx context.Context, chainID, height uint32, curveSpec constants.CurveSpec, signer objs.Signer, maxBytes uint32) (*objs.Tx, error) {
	utxoIDs, _ := ut.expIndex.GetExpiredObjects(txn, utils.Epoch(height), maxBytes)
	utxos := []*objs.TXOut{}
	var utxoID []byte
	for i := 0; i < len(utxoIDs); i++ {
		utxoID = utils.CopySlice(utxoIDs[i])
		select {
		case <-ctx.Done():
			return nil, nil
		default:
			// this prevents a node from losing a turn to propose due to taking too long
		}
		missing, err := ut.trie.Contains(txn, [][]byte{utils.CopySlice(utxoID)})
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		if len(missing) > 0 {
			continue
		}
		utxo, err := ut.getInternal(txn, utils.CopySlice(utxoID))
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		utxos = append(utxos, utxo)
		if len(utxos) == constants.MaxTxVectorLength {
			break
		}
	}
	if len(utxos) > 0 {
		utxos := objs.Vout(utxos)
		txIns, err := utxos.MakeTxIn()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		value, err := utxos.RemainingValue(height)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		pubk, err := signer.Pubkey()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		account := crypto.GetAccount(pubk)
		vsf := &objs.ValueStore{}
		err = vsf.New(chainID, value, account, curveSpec, make([]byte, constants.HashLen))
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		utxo := &objs.TXOut{}
		err = utxo.NewValueStore(vsf)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		tx := &objs.Tx{
			Vin:  txIns,
			Vout: objs.Vout{utxo},
		}
		err = tx.SetTxHash()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		for i := 0; i < len(utxos); i++ {
			ds, err := utxos[i].DataStore()
			if err != nil {
				utils.DebugTrace(ut.logger, err)
				return nil, err
			}
			err = ds.Sign(txIns[i], signer)
			if err != nil {
				utils.DebugTrace(ut.logger, err)
				return nil, err
			}
		}
		return tx, nil
	}
	return nil, nil
}

// GetValueForOwner allows a list of utxoIDs to be returned that are equal or
// greater than the value passed as minValue, and are owned by owner.
func (ut *UTXOHandler) GetValueForOwner(txn *badger.Txn, owner *objs.Owner, minValue *uint256.Uint256, maxCount int, startKey []byte) ([][]byte, *uint256.Uint256, []byte, error) {
	// This function operates under the assumption that the valueIndex and the trie are always in sync
	// If you make any change breaking this assumption, the results must be checked against the trie
	return ut.valueIndex.GetValueForOwner(txn, owner, minValue, nil, maxCount, startKey)
}

// PaginateDataByOwner ...
func (ut *UTXOHandler) PaginateDataByOwner(txn *badger.Txn, owner *objs.Owner, currentHeight uint32, numItems int, startIndex []byte) ([]*objs.PaginationResponse, error) {
	exclude := make(map[string]bool)
	resp := []*objs.PaginationResponse{}
	for j := 0; ; j++ {
		if len(resp) == numItems {
			return resp, nil
		}
		pageResp, err := ut.dataIndex.PaginateDataStores(txn, owner, int(numItems), startIndex, exclude)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		if len(pageResp) == 0 {
			return resp, nil
		}
		for i := 0; i < len(pageResp); i++ {
			item := pageResp[i]
			exclude[string(item.UTXOID)] = true
			utxo, missing, err := ut.Get(txn, [][]byte{item.UTXOID})
			if err != nil {
				return nil, err
			}
			if len(missing) > 0 {
				continue
			}
			if !utxo[0].HasDataStore() {
				continue
			}
			ds, _ := utxo[0].DataStore()
			if ok, _ := ds.DSLinker.DSPreImage.IsExpired(currentHeight + 1); ok {
				continue
			}
			ok, err := ut.TrieContains(txn, item.UTXOID)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			resp = append(resp, pageResp[i])
			if len(resp) >= int(numItems) {
				return resp, nil
			}
		}
		if len(resp) > 0 {
			startIndex = resp[len(resp)-1].Index
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
///////////PRIVATE METHODS//////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ut *UTXOHandler) addOne(txn *badger.Txn, utxo *objs.TXOut) error {
	utxoID, err := utxo.UTXOID()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	missing, err := ut.trie.Contains(txn, [][]byte{utxoID})
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	if len(missing) == 0 {
		utils.DebugTrace(ut.logger, err)
		return errorz.ErrInvalid{}.New("utxoID conflict")
	}
	owner, err := utxo.GenericOwner()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	if utxo.HasDataStore() {
		ds, err := utxo.DataStore()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		expEpoch, err := ds.EpochOfExpiration()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		size, err := utils.GetObjSize(utxo)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		err = ut.expIndex.Add(txn, expEpoch, utxoID, uint32(size))
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		dataIndex, err := ds.Index()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		err = ut.dataIndex.Add(txn, utxoID, owner, dataIndex)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
	} else {
		value, err := utxo.Value()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		err = ut.valueIndex.Add(txn, utxoID, owner, value)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
	}
	key := ut.makeUTXOKey(utxoID)
	if err := db.SetUTXO(txn, key, utxo); err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	return nil
}

func (ut *UTXOHandler) getInternal(txn *badger.Txn, utxoID []byte) (*objs.TXOut, error) {
	key := ut.makeUTXOKey(utxoID)
	utxo, err := db.GetUTXO(txn, key)
	if err != nil {
		return nil, err
	}
	return utxo, nil
}

// dropFromIndexes takes in a utxoID and drops it from the appropriate indexers
func (ut *UTXOHandler) dropFromIndexes(txn *badger.Txn, utxoID []byte) error {
	utxo, err := ut.getInternal(txn, utxoID)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return errorz.ErrInvalid{}.New("missing utxo for utxoID")
		}
		utils.DebugTrace(ut.logger, err)
		return err
	}
	if utxo.HasDataStore() {
		err = ut.expIndex.Drop(txn, utxoID)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		err = ut.dataIndex.Drop(txn, utxoID)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
	} else {
		err = ut.valueIndex.Drop(txn, utxoID)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
	}
	return nil
}

func (ut *UTXOHandler) makeUTXOKey(utxoID []byte) []byte {
	utxoIDCopy := utils.CopySlice(utxoID)
	key := dbprefix.PrefixMinedUTXO()
	key = append(key, utxoIDCopy...)
	return key
}

// addOneFastSync implements the logic we need for fast sync;
// it differs some from addOne because of different requriements
func (ut *UTXOHandler) addOneFastSync(txn *badger.Txn, utxo *objs.TXOut) error {
	utxoID, err := utxo.UTXOID()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	owner, err := utxo.GenericOwner()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	if utxo.HasDataStore() {
		ds, err := utxo.DataStore()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		expEpoch, err := ds.EpochOfExpiration()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		size, err := utils.GetObjSize(utxo)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		err = ut.expIndex.Add(txn, expEpoch, utxoID, uint32(size))
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		dataIndex, err := ds.Index()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		err = ut.dataIndex.AddFastSync(txn, utxoID, owner, dataIndex)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
	} else {
		value, err := utxo.Value()
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		err = ut.valueIndex.Add(txn, utxoID, owner, value)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
	}
	key := ut.makeUTXOKey(utxoID)
	if err := db.SetUTXO(txn, key, utxo); err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	return nil
}

func (ut *UTXOHandler) StoreSnapShotNode(txn *badger.Txn, batch []byte, root []byte, layer int) ([][]byte, int, []trie.LeafNode, error) {
	return ut.trie.StoreSnapShotNode(txn, batch, root, layer)
}

func (ut *UTXOHandler) FinalizeSnapShotRoot(txn *badger.Txn, root []byte, height uint32) error {
	if err := ut.trie.FinalizeSnapShotRoot(txn, root, height); err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	return nil
}

func (ut *UTXOHandler) GetSnapShotNode(txn *badger.Txn, height uint32, key []byte) ([]byte, error) {
	snapShotNode, err := ut.trie.GetSnapShotNode(txn, height, key)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	return snapShotNode, nil
}

func (ut *UTXOHandler) StoreSnapShotStateData(txn *badger.Txn, utxoID []byte, preHash []byte, utxoBytes []byte) error {
	utxo := &objs.TXOut{}
	err := utxo.UnmarshalBinary(utxoBytes)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	calcUtxoID, err := utxo.UTXOID()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	if !bytes.Equal(calcUtxoID, utxoID) {
		calcTxHash, err := utxo.TxHash()
		if err != nil {
			return err
		}
		utxoIdxOut, err := utxo.TXOutIdx()
		if err != nil {
			return err
		}
		return errorz.ErrInvalid{}.New(fmt.Sprintf("utxoID does not match calcUtxoID; utxoID: %x; calcUtxoID: %x calcTxHash: %x TxOutIdx: %v", utxoID, calcUtxoID, calcTxHash, utxoIdxOut))
	}
	calcPreHash, err := utxo.PreHash()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	if !bytes.Equal(calcPreHash, preHash) {
		utils.DebugTrace(ut.logger, err)
		return errorz.ErrInvalid{}.New(fmt.Sprintf("preHash does not match calcPreHash; preHash: %x; calcPreHash: %x; utxoID: %x", preHash, calcPreHash, utxoID))
	}
	if err := ut.addOneFastSync(txn, utxo); err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	return nil
}
