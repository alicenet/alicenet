package indexer

import (
	"github.com/dgraph-io/badger/v2"

	"github.com/alicenet/alicenet/utils"
)

// NewRefCounter returns a new RefCounter
func NewRefCounter(p prefixFunc) *RefCounter {
	return &RefCounter{p}
}

// RefCounter allows for counting the number of references to an object
type RefCounter struct {
	prefix prefixFunc
}

type RefCounterKey struct {
	key []byte
}

// MarshalBinary returns the byte slice for the key object.
func (rck *RefCounterKey) MarshalBinary() []byte {
	return utils.CopySlice(rck.key)
}

// UnmarshalBinary takes in a byte slice to set the key object.
func (rck *RefCounterKey) UnmarshalBinary(data []byte) {
	rck.key = utils.CopySlice(data)
}

// Increment increases the reference count to the txhash
func (rc *RefCounter) Increment(txn *badger.Txn, txHash []byte) (int64, error) {
	rcKey := rc.makeKey(txHash)
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

// Decrement decreases the reference count to the txhash
func (rc *RefCounter) Decrement(txn *badger.Txn, txHash []byte) (int64, error) {
	rcKey := rc.makeKey(txHash)
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

func (rc *RefCounter) makeKey(txHash []byte) *RefCounterKey {
	key := []byte{}
	key = append(key, rc.prefix()...)
	key = append(key, utils.CopySlice(txHash)...)
	rck := &RefCounterKey{}
	rck.UnmarshalBinary(key)
	return rck
}
