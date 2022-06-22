package trie

import (
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

func prefixNode() []byte {
	return []byte("T")
}

func prefixCommitHeight() []byte {
	return []byte("H")
}

func prefixRootHash() []byte {
	return []byte("R")
}

func (db *cacheDB) getNodeDB(txn *badger.Txn, key []byte) ([]byte, error) {
	key = convNilToBytes(key)
	var node Hash
	copy(node[:], key)
	v, err := GetNodeDB(txn, db.prefixFunc(), key)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func getNodeDB(txn *badger.Txn, prefix []byte, key []byte) ([]byte, error) {
	nodekey := []byte{}
	nodekey = append(nodekey, prefix...)
	nodekey = append(nodekey, prefixNode()...)
	nodekey = append(nodekey, key...)
	value, err := utils.GetValue(txn, nodekey)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, nil
		}
		return nil, err
	}
	return value, nil
}

func (db *cacheDB) setNodeDB(txn *badger.Txn, key, value []byte) error {
	var node Hash
	copy(node[:], key)
	err := setNodeDB(txn, db.prefixFunc(), key, value)
	if err != nil {
		return err
	}
	return nil
}

func setNodeDB(txn *badger.Txn, prefix []byte, key, value []byte) error {
	key = convNilToBytes(key)
	value = convNilToBytes(value)
	nodekey := []byte{}
	nodekey = append(nodekey, prefix...)
	nodekey = append(nodekey, prefixNode()...)
	nodekey = append(nodekey, key...)
	return utils.SetValue(txn, nodekey, value)
}

func (db *cacheDB) getCommitHeightDB(txn *badger.Txn) (uint32, error) {
	return getCommitHeightDB(txn, db.prefixFunc())
}

func getCommitHeightDB(txn *badger.Txn, prefix []byte) (uint32, error) {
	nodekey := []byte{}
	nodekey = append(nodekey, prefix...)
	nodekey = append(nodekey, prefixCommitHeight()...)
	v, err := utils.GetValue(txn, nodekey)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return 0, nil
		}
		return 0, err
	}
	return utils.UnmarshalUint32(v)
}

func (db *cacheDB) setCommitHeightDB(txn *badger.Txn, value uint32) error {
	return setCommitHeightDB(txn, db.prefixFunc(), value)
}

func setCommitHeightDB(txn *badger.Txn, prefix []byte, value uint32) error {
	nodekey := []byte{}
	nodekey = append(nodekey, prefix...)
	nodekey = append(nodekey, prefixCommitHeight()...)
	v := utils.MarshalUint32(value)
	return utils.SetValue(txn, nodekey, v)
}

func (db *cacheDB) getRootForHeightDB(txn *badger.Txn, height uint32) ([]byte, error) {
	nodekey := []byte{}
	nodekey = append(nodekey, db.prefixFunc()...)
	nodekey = append(nodekey, prefixRootHash()...)
	nodekey = append(nodekey, utils.MarshalUint32(height)...)
	return utils.GetValue(txn, nodekey)
}

func (db *cacheDB) setRootForHeightDB(txn *badger.Txn, height uint32, root []byte) error {
	return setRootForHeightDB(txn, db.prefixFunc(), height, root)
}

func setRootForHeightDB(txn *badger.Txn, prefix []byte, height uint32, root []byte) error {
	nodekey := []byte{}
	nodekey = append(nodekey, prefix...)
	nodekey = append(nodekey, prefixRootHash()...)
	nodekey = append(nodekey, utils.MarshalUint32(height)...)
	return utils.SetValue(txn, nodekey, root)
}
