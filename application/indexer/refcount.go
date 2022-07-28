package indexer

import (
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

// NewRefCounter makes a new RefCounter struct
func NewRefCounter(p prefixFunc) *RefCounter {
	return &RefCounter{p}
}

// RefCounter enables the ability to count multiple references
// to specific utxos.
type RefCounter struct {
	prefix prefixFunc
}

// RefCounterKey is a key for the RefCounter
type RefCounterKey struct {
	key []byte
}

// MarshalBinary returns the byte slice for the key object
func (rck *RefCounterKey) MarshalBinary() []byte {
	return utils.CopySlice(rck.key)
}

// UnmarshalBinary takes in a byte slice to set the key object
func (rck *RefCounterKey) UnmarshalBinary(data []byte) {
	rck.key = utils.CopySlice(data)
}

// Increment increases the number of references to a specific utxoID.
func (rc *RefCounter) Increment(txn *badger.Txn, utxoID []byte) (int64, error) {
	rcKey := rc.makeKey(utxoID)
	key := rcKey.MarshalBinary()
	v, err := utils.GetInt64(txn, key)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return 1, utils.SetInt64(txn, key, 1)
		}
		return 0, err
	}
	v++
	return v, utils.SetInt64(txn, key, v)
}

// Decrement decreases the number of references to a specific utxoID.
func (rc *RefCounter) Decrement(txn *badger.Txn, utxoID []byte) (int64, error) {
	rcKey := rc.makeKey(utxoID)
	key := rcKey.MarshalBinary()
	v, err := utils.GetInt64(txn, key)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return 0, nil
		}
		return 0, err
	}
	v--
	if v > 0 {
		return v, utils.SetInt64(txn, key, v)
	}
	return 0, utils.DeleteValue(txn, key)
}

// makeKey makes a RefCounterKey for a utxo.
func (rc *RefCounter) makeKey(utxoID []byte) *RefCounterKey {
	key := []byte{}
	key = append(key, rc.prefix()...)
	key = append(key, utils.CopySlice(utxoID)...)
	rck := &RefCounterKey{}
	rck.UnmarshalBinary(key)
	return rck
}
