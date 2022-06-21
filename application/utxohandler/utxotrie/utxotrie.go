package utxotrie

import (
	"bytes"

	"github.com/alicenet/alicenet/application/objs"
	trie "github.com/alicenet/alicenet/badgerTrie"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

func makeheightKey(height uint32) []byte {
	key := []byte{}
	prefix := dbprefix.PrefixTrieRootForHeight()
	key = append(key, prefix...)
	hbytes := utils.MarshalUint32(height)
	key = append(key, hbytes...)
	return key
}

func setRootForHeight(txn *badger.Txn, height uint32, root []byte) error {
	key := makeheightKey(height)
	return utils.SetValue(txn, key, root)
}

//nolint:unused,deadcode
func getRootForHeight(txn *badger.Txn, height uint32) ([]byte, error) {
	key := makeheightKey(height)
	return utils.GetValue(txn, key)
}

// NewUTXOTrie ...
func NewUTXOTrie(db *badger.DB) *UTXOTrie {
	return &UTXOTrie{
		DB:     db,
		logger: logging.GetLogger(constants.LoggerApp),
	}
}

// always keep two snapshots back
// if a revert is called, prune back to last snapshot and
// rebuild from snapshot

//  store three trie roots in database at all times
//  never prune any tries
//  in init:
//    if height == 1, then create three tries with nil as root
//

// UTXOTrie stores the active UTXO set.
// There are two types of entries in the utxoTrie.
// The first type is a standard UTXO. These entries are
// stored at the location of the UTXOID with the value
// equal to the preHash of the UTXO.
// The second type is the deposit entry. These are stored
// at the location of the Nonce of the deposit with the
// value equal to the PreHash of the Deposit.
// In order to spend a UTXO it must be present in the trie.
// In order to spend a deposit it MUST NOT be in the trie.
type UTXOTrie struct {
	DB     *badger.DB
	logger *logrus.Logger
}

func (ut *UTXOTrie) Init(_ uint32) error {
	return nil
}

func (ut *UTXOTrie) GetCanonicalTrie(txn *badger.Txn) (*trie.SMT, error) {
	root, err := GetCurrentStateRoot(txn)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
	}
	if bytes.Equal(root, make([]byte, constants.HashLen)) {
		root = nil
	}
	t := trie.NewSMT(root, trie.Hasher, func() []byte { return getTriePrefix() })
	return t, nil
}

func (ut *UTXOTrie) GetPendingTrie(txn *badger.Txn) (*trie.SMT, error) {
	root, err := GetPendingStateRoot(txn)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
	}
	if bytes.Equal(root, make([]byte, constants.HashLen)) {
		root = nil
	}
	t := trie.NewSMT(root, trie.Hasher, func() []byte { return getTriePrefix() })
	return t, nil
}

func (ut *UTXOTrie) GetCurrentTrie(txn *badger.Txn) (*trie.SMT, error) {
	root, err := GetCurrentStateRoot(txn)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
	}
	if bytes.Equal(root, make([]byte, constants.HashLen)) {
		root = nil
	}
	t := trie.NewSMT(root, trie.Hasher, func() []byte { return getTriePrefix() })
	return t, nil
}

func (ut *UTXOTrie) Get(txn *badger.Txn, utxoIDs [][]byte) ([][]byte, [][]byte, error) {
	utxoHashes := [][]byte{}
	missing := [][]byte{}
	current, err := ut.GetCurrentTrie(txn)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, nil, err
	}
	for j := 0; j < len(utxoIDs); j++ {
		utxoID := utils.CopySlice(utxoIDs[j])
		utxoHsh, err := current.Get(txn, utils.CopySlice(utxoID))
		if err != nil {
			if err != badger.ErrKeyNotFound {
				utils.DebugTrace(ut.logger, err)
				return nil, nil, err
			}
		}
		if len(utxoHsh) == 0 {
			missing = append(missing, utils.CopySlice(utxoID))
			continue
		}
		utxoHashes = append(utxoHashes, utxoHsh)
	}
	return utxoHashes, missing, nil
}

func (ut *UTXOTrie) Contains(txn *badger.Txn, utxoIDs [][]byte) ([][]byte, error) {
	current, err := ut.GetCurrentTrie(txn)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	missing := [][]byte{}
	for j := 0; j < len(utxoIDs); j++ {
		utxoID := utils.CopySlice(utxoIDs[j])
		txHash, err := current.Get(txn, utils.CopySlice(utxoID))
		if err != nil {
			if err != badger.ErrKeyNotFound {
				utils.DebugTrace(ut.logger, err)
				return nil, err
			}
		}
		if len(txHash) == 0 {
			missing = append(missing, utils.CopySlice(utxoID))
		}
	}
	return missing, nil
}

func (ut *UTXOTrie) ApplyState(txn *badger.Txn, txs objs.TxVec, height uint32) ([]byte, error) {
	current, fn, err := ut.session(txn)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	var stateRoot []byte
	if len(txs) > 0 {
		StateRoot, err := ut.add(txn, txs, current, fn)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			current.Discard()
			return nil, err
		}
		stateRoot = StateRoot
	} else {
		StateRoot, err := GetCurrentStateRoot(txn)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			current.Discard()
			return nil, err
		}
		stateRoot = StateRoot
	}
	if err := ut.updateRoots(txn, height, stateRoot, current); err != nil {
		utils.DebugTrace(ut.logger, err)
		current.Discard()
		return nil, err
	}
	return stateRoot, nil
}

func (ut *UTXOTrie) updateRoots(txn *badger.Txn, height uint32, stateRoot []byte, current *trie.SMT) error {
	if err := setRootForHeight(txn, height, stateRoot); err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	if err := SetCurrentStateRoot(txn, stateRoot); err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	switch {
	case height == 1:
		err := SetPendingStateRoot(txn, stateRoot)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		err = SetCanonicalStateRoot(txn, stateRoot)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
	case height%constants.EpochLength == 0:
		cproot, err := GetPendingStateRoot(txn)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		err = SetCanonicalStateRoot(txn, utils.CopySlice(cproot))
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
		err = SetPendingStateRoot(txn, stateRoot)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return err
		}
	}
	_, err := current.Commit(txn, height)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	return nil
}

func (ut *UTXOTrie) GetCurrentStateRoot(txn *badger.Txn) ([]byte, error) {
	rt, err := GetCurrentStateRoot(txn)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	return rt, nil
}

func (ut *UTXOTrie) GetStateRootForProposal(txn *badger.Txn, txs objs.TxVec) ([]byte, error) {
	if len(txs) == 0 {
		sr, err := GetCurrentStateRoot(txn)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		return sr, nil
	}
	current, fn, err := ut.session(txn)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	defer current.Discard()
	sr, err := ut.add(txn, txs, current, fn)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	return sr, nil
}

func (ut *UTXOTrie) add(txn *badger.Txn, txs objs.TxVec, current *trie.SMT, fn func(txn *badger.Txn, current *trie.SMT, newUTXOIDs [][]byte, newUTXOHashes [][]byte, consumedUTXOIDS [][]byte) ([]byte, error)) ([]byte, error) {
	addkeys := [][]byte{}
	addvalues := [][]byte{}
	delkeys := [][]byte{}
	aa, err := txs.ConsumedUTXOIDNoDeposits()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	if len(aa) > 0 {
		for i := 0; i < len(aa); i++ {
			delkeys = append(delkeys, aa[i])
		}
	}
	cc, err := txs.GeneratedUTXOID()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	dd, err := txs.GeneratedPreHash()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	if len(cc) > 0 {
		for i := 0; i < len(cc); i++ {
			addkeys = append(addkeys, cc[i])
			addvalues = append(addvalues, dd[i])
		}
	}
	ee, err := txs.ConsumedUTXOIDOnlyDeposits()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	ff, err := txs.ConsumedPreHashOnlyDeposits()
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	if len(ee) > 0 {
		for i := 0; i < len(ee); i++ {
			addkeys = append(addkeys, ee[i])
			addvalues = append(addvalues, ff[i])
		}
	}
	return fn(txn, current, addkeys, addvalues, delkeys)
}

func (ut *UTXOTrie) session(txn *badger.Txn) (*trie.SMT, func(txn *badger.Txn, current *trie.SMT, newUTXOIDs [][]byte, newUTXOHashes [][]byte, consumedUTXOIDS [][]byte) ([]byte, error), error) {
	fn := func(txn *badger.Txn, current *trie.SMT, newUTXOIDs [][]byte, newUTXOHashes [][]byte, consumedUTXOIDS [][]byte) ([]byte, error) {
		updateKeys := [][]byte{}
		updateValues := [][]byte{}
		for i := 0; i < len(newUTXOIDs); i++ {
			updateKeys = append(updateKeys, utils.CopySlice(newUTXOIDs[i]))
			updateValues = append(updateValues, utils.CopySlice(newUTXOHashes[i]))
		}
		for i := 0; i < len(consumedUTXOIDS); i++ {
			updateKeys = append(updateKeys, utils.CopySlice(consumedUTXOIDS[i]))
			updateValues = append(updateValues, utils.CopySlice(trie.DefaultLeaf))
		}
		updateKeysSorted, updateValuesSorted, err := utils.SortKVs(updateKeys, updateValues)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		stateRoot, err := current.Update(txn, updateKeysSorted, updateValuesSorted)
		if err != nil {
			utils.DebugTrace(ut.logger, err)
			return nil, err
		}
		return stateRoot, nil
	}
	current, err := ut.GetCurrentTrie(txn)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, fn, err
	}
	return current, fn, nil
}

func (ut *UTXOTrie) StoreSnapShotNode(txn *badger.Txn, batch []byte, root []byte, layer int) ([][]byte, int, []trie.LeafNode, error) {
	t := trie.NewSMT(root, trie.Hasher, func() []byte { return getTriePrefix() })
	return t.StoreSnapShotNode(txn, batch, root, layer)
}

func (ut *UTXOTrie) FinalizeSnapShotRoot(txn *badger.Txn, root []byte, height uint32) error {
	if err := setRootForHeight(txn, height, root); err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	if err := SetCurrentStateRoot(txn, root); err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	if err := SetPendingStateRoot(txn, root); err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	if err := SetCanonicalStateRoot(txn, root); err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	t := trie.NewSMT(root, trie.Hasher, func() []byte { return getTriePrefix() })
	if err := t.FinalizeSnapShotRoot(txn, root, height); err != nil {
		utils.DebugTrace(ut.logger, err)
		return err
	}
	return nil
}

func (ut *UTXOTrie) GetSnapShotNode(txn *badger.Txn, height uint32, key []byte) ([]byte, error) {
	snapShotNode, err := trie.GetNodeDB(txn, getTriePrefix(), key)
	if err != nil {
		utils.DebugTrace(ut.logger, err)
		return nil, err
	}
	return snapShotNode, nil
}
