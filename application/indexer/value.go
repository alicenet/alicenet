package indexer

import (
	"github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
)

/*
<prefix>|<owner>|<value>
  <utxoID>
<prefix>|<utxoID>
  <owner>|<value>

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

func (vi *ValueIndex) GetValueForOwner(txn *badger.Txn, owner *objs.Owner, minValue *uint256.Uint256, exclude [][]byte) ([][]byte, *uint256.Uint256, error) {
	exclusionSet := make(map[[constants.HashLen]byte]bool)
	var hsh [constants.HashLen]byte
	for j := 0; j < len(exclude); j++ {
		ID := utils.CopySlice(exclude[j])
		copy(hsh[:], utils.CopySlice(ID))
		exclusionSet[hsh] = true
	}
	result := [][]byte{}
	valueCount := uint256.Zero()
	opts := badger.DefaultIteratorOptions
	prefix := vi.prefix()
	ownerBytes, err := owner.MarshalBinary()
	if err != nil {
		return nil, nil, err
	}
	prefix = append(prefix, ownerBytes...)
	prefixLen := len(prefix)
	opts.Prefix = prefix
	iter := txn.NewIterator(opts)
	defer iter.Close()
	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
		itm := iter.Item()
		key := itm.KeyCopy(nil)
		valueBytes := key[prefixLen : len(key)-constants.HashLen]
		value := &uint256.Uint256{}
		err := value.UnmarshalBinary(utils.CopySlice(valueBytes))
		if err != nil {
			return nil, nil, err
		}
		utxoID, err := itm.ValueCopy(nil)
		if err != nil {
			return nil, nil, err
		}
		copy(hsh[:], utxoID)
		if !exclusionSet[hsh] {
			valueCount, err = valueCount.Clone().Add(valueCount.Clone(), value.Clone())
			if err != nil {
				return nil, nil, err
			}
			result = append(result, utxoID)
		}
		if valueCount.Gte(minValue) {
			break
		}
	}
	return result, valueCount.Clone(), nil
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
