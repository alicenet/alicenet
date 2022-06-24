package indexer

import (
	"github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

/*

== BADGER KEYS ==

lookup:
key: <prefix>|<owner>|<value>|<utxoID>
  value: <utxoID>

reverse lookup:
key: <prefix>|<utxoID>
  value: <owner>|<value>|<utxoID>

*/

func NewValueIndex(p, pp prefixFunc) *ValueIndex {
	return &ValueIndex{p, pp}
}

// ValueIndex creates an index that allows objects to be dropped
// by epoch
type ValueIndex struct {
	prefix    prefixFunc
	refPrefix prefixFunc
}

type ValueIndexKey struct {
	key []byte
}

func (vik *ValueIndexKey) MarshalBinary() []byte {
	return utils.CopySlice(vik.key)
}

func (vik *ValueIndexKey) UnmarshalBinary(data []byte) {
	vik.key = utils.CopySlice(data)
}

type ValueIndexRefKey struct {
	refkey []byte
}

// MarshalBinary returns the byte slice for the key object
func (virk *ValueIndexRefKey) MarshalBinary() []byte {
	return utils.CopySlice(virk.refkey)
}

// UnmarshalBinary takes in a byte slice to set the key object
func (virk *ValueIndexRefKey) UnmarshalBinary(data []byte) {
	virk.refkey = utils.CopySlice(data)
}

// Add adds an item to the list
func (vi *ValueIndex) Add(txn *badger.Txn, utxoID []byte, owner *objs.Owner, valueOrig *uint256.Uint256) error {
	valueClone := valueOrig.Clone()
	viKey, err := vi.makeKey(owner, valueClone.Clone(), utxoID)
	if err != nil {
		return err
	}
	key := viKey.MarshalBinary()
	viRefKey := vi.makeRefKey(utxoID)
	refKey := viRefKey.MarshalBinary()
	valueIndex, err := vi.makeValueIndex(owner, valueClone.Clone(), utxoID)
	if err != nil {
		return err
	}
	err = utils.SetValue(txn, refKey, valueIndex)
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, utils.CopySlice(utxoID))
}

// Drop returns a list of all txHashes that should be dropped
func (vi *ValueIndex) Drop(txn *badger.Txn, utxoID []byte) error {
	utxoIDCopy := utils.CopySlice(utxoID)
	viRefKey := vi.makeRefKey(utxoIDCopy)
	refKey := viRefKey.MarshalBinary()
	valueIndex, err := utils.GetValue(txn, refKey)
	if err != nil {
		return err
	}
	key := []byte{}
	key = append(key, vi.prefix()...)
	key = append(key, valueIndex...)
	err = utils.DeleteValue(txn, refKey)
	if err != nil {
		return err
	}
	return utils.DeleteValue(txn, key)
}

func (vi *ValueIndex) GetValueForOwner(txn *badger.Txn, owner *objs.Owner, minValue *uint256.Uint256, excludeFn func([]byte) (bool, error), maxCount int, lastKey []byte) ([][]byte, *uint256.Uint256, []byte, error) {
	ownerBytes, err := owner.MarshalBinary()
	if err != nil {
		return nil, nil, nil, err
	}

	prefix := vi.prefix()
	prefix = append(prefix, ownerBytes...)
	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix
	iter := txn.NewIterator(opts)
	defer iter.Close()

	result := [][]byte{}
	totalValue := uint256.Zero()
	prefixLen := len(prefix)

	if lastKey != nil {
		iter.Seek(lastKey)
		if !iter.ValidForPrefix(prefix) {
			return result, totalValue, nil, nil
		}
		iter.Next()
	} else {
		iter.Seek(prefix)
	}

	for ; iter.ValidForPrefix(prefix); iter.Next() {
		itm := iter.Item()
		key := itm.KeyCopy(nil)

		valueBytes := key[prefixLen : len(key)-constants.HashLen]
		value := &uint256.Uint256{}
		err := value.UnmarshalBinary(valueBytes)
		if err != nil {
			return nil, nil, nil, err
		}

		utxoID, err := itm.ValueCopy(nil)
		if err != nil {
			return nil, nil, nil, err
		}

		if excludeFn != nil {
			shouldExclude, err := excludeFn(utxoID)
			if err != nil {
				return nil, nil, nil, err
			}
			if shouldExclude {
				continue
			}
		}

		totalValue, err = totalValue.Add(totalValue, value)
		if err != nil {
			return nil, nil, nil, err
		}

		result = append(result, utxoID)

		if totalValue.Gte(minValue) {
			break
		}
		if len(result) >= maxCount {
			return result, totalValue, key, nil
		}
	}
	return result, totalValue, nil, nil
}

func (vi *ValueIndex) makeKey(owner *objs.Owner, valueOrig *uint256.Uint256, utxoID []byte) (*ValueIndexKey, error) {
	valueClone := valueOrig.Clone()
	valueIndex, err := vi.makeValueIndex(owner, valueClone.Clone(), utxoID)
	if err != nil {
		return nil, err
	}
	key := []byte{}
	key = append(key, vi.prefix()...)
	key = append(key, valueIndex...)
	viKey := &ValueIndexKey{}
	viKey.UnmarshalBinary(key)
	return viKey, nil
}

func (vi *ValueIndex) makeRefKey(utxoID []byte) *ValueIndexRefKey {
	refKey := []byte{}
	refKey = append(refKey, vi.refPrefix()...)
	refKey = append(refKey, utils.CopySlice(utxoID)...)
	viRefKey := &ValueIndexRefKey{}
	viRefKey.UnmarshalBinary(refKey)
	return viRefKey
}

func (vi *ValueIndex) makeValueIndex(owner *objs.Owner, valueOrig *uint256.Uint256, utxoID []byte) ([]byte, error) {
	valueBytes, err := valueOrig.MarshalBinary()
	if err != nil {
		return nil, err
	}
	ownerBytes, err := owner.MarshalBinary()
	if err != nil {
		return nil, err
	}
	valueIndex := []byte{}
	valueIndex = append(valueIndex, ownerBytes...)
	valueIndex = append(valueIndex, valueBytes...)
	valueIndex = append(valueIndex, utils.CopySlice(utxoID)...)
	return valueIndex, nil
}
