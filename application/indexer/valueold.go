package indexer

import (
	"github.com/MadBase/MadNet/application/objs"
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

// TODO HANDLE UINT32 OVERFLOW

func newValueIndexOld(p, pp prefixFunc) *valueIndexOld {
	return &valueIndexOld{p, pp}
}

// valueIndexOld creates an index that allows objects to be dropped
// by epoch
type valueIndexOld struct {
	prefix    prefixFunc
	refPrefix prefixFunc
}

type valueIndexOldKey struct {
	key []byte
}

func (vik *valueIndexOldKey) MarshalBinary() []byte {
	return utils.CopySlice(vik.key)
}

func (vik *valueIndexOldKey) UnmarshalBinary(data []byte) {
	vik.key = utils.CopySlice(data)
}

type valueIndexOldRefKey struct {
	refkey []byte
}

// MarshalBinary returns the byte slice for the key object
func (virk *valueIndexOldRefKey) MarshalBinary() []byte {
	return utils.CopySlice(virk.refkey)
}

// UnmarshalBinary takes in a byte slice to set the key object
func (virk *valueIndexOldRefKey) UnmarshalBinary(data []byte) {
	virk.refkey = utils.CopySlice(data)
}

// Add adds an item to the list
func (vi *valueIndexOld) Add(txn *badger.Txn, utxoID []byte, owner *objs.Owner, value uint32) error {
	viKey, err := vi.makeKey(owner, value, utxoID)
	if err != nil {
		return err
	}
	key := viKey.MarshalBinary()
	viRefKey := vi.makeRefKey(utxoID)
	refKey := viRefKey.MarshalBinary()
	valueIndexOld, err := vi.makevalueIndexOld(owner, value, utxoID)
	if err != nil {
		return err
	}
	err = utils.SetValue(txn, refKey, valueIndexOld)
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, utils.CopySlice(utxoID))
}

// Drop returns a list of all txHashes that should be dropped
func (vi *valueIndexOld) Drop(txn *badger.Txn, utxoID []byte) error {
	utxoIDCopy := utils.CopySlice(utxoID)
	viRefKey := vi.makeRefKey(utxoIDCopy)
	refKey := viRefKey.MarshalBinary()
	valueIndexOld, err := utils.GetValue(txn, refKey)
	if err != nil {
		return err
	}
	key := []byte{}
	key = append(key, vi.prefix()...)
	key = append(key, valueIndexOld...)
	err = utils.DeleteValue(txn, refKey)
	if err != nil {
		return err
	}
	return utils.DeleteValue(txn, key)
}

func (vi *valueIndexOld) GetValueForOwner(txn *badger.Txn, owner *objs.Owner, minValue uint32, exclude [][]byte) ([][]byte, uint32, error) {
	exclusionSet := make(map[[constants.HashLen]byte]bool)
	var hsh [constants.HashLen]byte
	for _, ID := range exclude {
		copy(hsh[:], utils.CopySlice(ID))
		exclusionSet[hsh] = true
	}
	result := [][]byte{}
	valueCount := uint32(0)
	opts := badger.DefaultIteratorOptions
	prefix := vi.prefix()
	ownerBytes, err := owner.MarshalBinary()
	if err != nil {
		return nil, 0, err
	}
	prefix = append(prefix, ownerBytes...)
	opts.Prefix = prefix
	iter := txn.NewIterator(opts)
	defer iter.Close()
	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
		itm := iter.Item()
		key := itm.KeyCopy(nil)
		valueBytes := key[len(prefix) : len(prefix)+4]
		value, _ := utils.UnmarshalUint32(valueBytes)
		utxoID, err := itm.ValueCopy(nil)
		if err != nil {
			return nil, 0, err
		}
		copy(hsh[:], utxoID)
		if !exclusionSet[hsh] {
			valueCount += value
			result = append(result, utxoID)
		}
		if valueCount >= minValue {
			break
		}
	}
	return result, valueCount, nil
}

func (vi *valueIndexOld) makeKey(owner *objs.Owner, value uint32, utxoID []byte) (*valueIndexOldKey, error) {
	valueIndexOld, err := vi.makevalueIndexOld(owner, value, utxoID)
	if err != nil {
		return nil, err
	}
	key := []byte{}
	key = append(key, vi.prefix()...)
	key = append(key, valueIndexOld...)
	viKey := &valueIndexOldKey{}
	viKey.UnmarshalBinary(key)
	return viKey, nil
}

func (vi *valueIndexOld) makeRefKey(utxoID []byte) *valueIndexOldRefKey {
	refKey := []byte{}
	refKey = append(refKey, vi.refPrefix()...)
	refKey = append(refKey, utils.CopySlice(utxoID)...)
	viRefKey := &valueIndexOldRefKey{}
	viRefKey.UnmarshalBinary(refKey)
	return viRefKey
}

func (vi *valueIndexOld) makevalueIndexOld(owner *objs.Owner, value uint32, utxoID []byte) ([]byte, error) {
	valueBytes := utils.MarshalUint32(value)
	ownerBytes, err := owner.MarshalBinary()
	if err != nil {
		return nil, err
	}
	valueIndexOld := []byte{}
	valueIndexOld = append(valueIndexOld, ownerBytes...)
	valueIndexOld = append(valueIndexOld, valueBytes...)
	valueIndexOld = append(valueIndexOld, utils.CopySlice(utxoID)...)
	return valueIndexOld, nil
}
