package indexer

import (
	"errors"
	"time"

	"github.com/dgraph-io/badger/v2"

	"github.com/alicenet/alicenet/utils"
)

var ErrIterClose = errors.New("iter closed")

// NewInsertionOrderIndex returns a new InsertionOrderIndex
func NewInsertionOrderIndex(p, pp prefixFunc) *InsertionOrderIndexer {
	return &InsertionOrderIndexer{p, pp}
}

// InsertionOrderIndexer is an index which allows for ordering txs for insertion
type InsertionOrderIndexer struct {
	prefix    prefixFunc
	revPrefix prefixFunc
}

type InsertionOrderIndexerKey struct {
	key []byte
}

// MarshalBinary returns the byte slice for the key object.
func (ioik *InsertionOrderIndexerKey) MarshalBinary() []byte {
	return utils.CopySlice(ioik.key)
}

// UnmarshalBinary takes in a byte slice to set the key object.
func (ioik *InsertionOrderIndexerKey) UnmarshalBinary(data []byte) {
	ioik.key = utils.CopySlice(data)
}

type InsertionOrderIndexerRevKey struct {
	revkey []byte
}

// MarshalBinary returns the byte slice for the key object.
func (ioirk *InsertionOrderIndexerRevKey) MarshalBinary() []byte {
	return utils.CopySlice(ioirk.revkey)
}

// UnmarshalBinary takes in a byte slice to set the key object.
func (ioirk *InsertionOrderIndexerRevKey) UnmarshalBinary(data []byte) {
	ioirk.revkey = utils.CopySlice(data)
}

// Add adds a txhash to the indexer
func (ioi *InsertionOrderIndexer) Add(txn *badger.Txn, txHash []byte) error {
	txHashCopy := utils.CopySlice(txHash)
	ioiIdxKey, ioiRevIdxKey, err := ioi.makeIndexKeys(txHashCopy)
	if err != nil {
		return err
	}
	idxKey := ioiIdxKey.MarshalBinary()
	revIdxKey := ioiRevIdxKey.MarshalBinary()
	err = utils.SetValue(txn, idxKey, revIdxKey)
	if err != nil {
		return err
	}
	err = utils.SetValue(txn, revIdxKey, idxKey)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a txhash to the indexer
func (ioi *InsertionOrderIndexer) Delete(txn *badger.Txn, txHash []byte) error {
	txHashCopy := utils.CopySlice(txHash)
	_, ioiRevIdxKey, err := ioi.makeIndexKeys(txHashCopy)
	if err != nil {
		return err
	}
	revIdxKey := ioiRevIdxKey.MarshalBinary()
	idxKey, err := utils.GetValue(txn, revIdxKey)
	if err != nil {
		return err
	}
	err = utils.DeleteValue(txn, idxKey)
	if err != nil {
		return err
	}
	err = utils.DeleteValue(txn, revIdxKey)
	if err != nil {
		return err
	}
	return nil
}

func (ioi *InsertionOrderIndexer) makeIndexKeys(txHash []byte) (*InsertionOrderIndexerKey, *InsertionOrderIndexerRevKey, error) {
	txHashCopy := utils.CopySlice(txHash)
	ts := time.Now()
	tsBytes, err := ts.MarshalBinary()
	if err != nil {
		return nil, nil, err
	}
	idxKey := []byte{}
	idxKey = append(idxKey, ioi.prefix()...)
	idxKey = append(idxKey, tsBytes...)
	idxKey = append(idxKey, txHashCopy...)
	ioiKey := &InsertionOrderIndexerKey{}
	ioiKey.UnmarshalBinary(idxKey)
	revIdxKey := []byte{}
	revIdxKey = append(revIdxKey, ioi.revPrefix()...)
	revIdxKey = append(revIdxKey, txHashCopy...)
	ioiRevKey := &InsertionOrderIndexerRevKey{}
	ioiRevKey.UnmarshalBinary(revIdxKey)
	return ioiKey, ioiRevKey, nil
}

// NewIter returns an iterator for iterating through the indexer
func (ioi *InsertionOrderIndexer) NewIter(txn *badger.Txn) (*badger.Iterator, []byte) {
	prefix := ioi.prefix()
	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix
	return txn.NewIterator(opts), prefix
}
