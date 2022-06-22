package indexer

import (
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

/*
Need double index
First index is when it expires
Second index is the size of the object

<prefix>|<epoch of exp>|<bitwise not of the object size>|<utxoID>
  <utxoID>

iterate in fwd direction
*/

func NewExpSizeIndex(p, pp prefixFunc) *ExpSizeIndex {
	return &ExpSizeIndex{p, pp}
}

// ExpSizeIndex creates an index that allows objects to be dropped
// by epoch
type ExpSizeIndex struct {
	prefix    prefixFunc
	refPrefix prefixFunc
}

type ExpSizeIndexKey struct {
	key []byte
}

// MarshalBinary returns the byte slice for the key object
func (esik *ExpSizeIndexKey) MarshalBinary() []byte {
	return utils.CopySlice(esik.key)
}

// UnmarshalBinary takes in a byte slice to set the key object
func (esik *ExpSizeIndexKey) UnmarshalBinary(data []byte) {
	esik.key = utils.CopySlice(data)
}

type ExpSizeIndexRefKey struct {
	refkey []byte
}

// MarshalBinary returns the byte slice for the key object
func (esirk *ExpSizeIndexRefKey) MarshalBinary() []byte {
	return utils.CopySlice(esirk.refkey)
}

// UnmarshalBinary takes in a byte slice to set the key object
func (esirk *ExpSizeIndexRefKey) UnmarshalBinary(data []byte) {
	esirk.refkey = utils.CopySlice(data)
}

func (esi *ExpSizeIndex) Add(txn *badger.Txn, epoch uint32, utxoID []byte, size uint32) error {
	esiKey := esi.makeKey(epoch, size, utxoID)
	key := esiKey.MarshalBinary()
	esiRefKey := esi.makeRefKey(utxoID)
	refKey := esiRefKey.MarshalBinary()
	refValue := esi.makeRefValue(epoch, size)
	err := utils.SetValue(txn, refKey, refValue)
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, []byte{})
}

func (esi *ExpSizeIndex) Drop(txn *badger.Txn, utxoID []byte) error {
	esiRefKey := esi.makeRefKey(utxoID)
	refKey := esiRefKey.MarshalBinary()
	epochSizeBytes, err := utils.GetValue(txn, refKey)
	if err != nil {
		return err
	}
	epochBytes := epochSizeBytes[:4]
	// slice is 4 bytes so no error will be raised
	epoch, _ := utils.UnmarshalUint32(epochBytes)
	sizeInvBytes := epochSizeBytes[4:]
	sizeInv, err := utils.UnmarshalUint32(sizeInvBytes)
	if err != nil {
		return err
	}
	size := constants.MaxUint32 - sizeInv
	err = utils.DeleteValue(txn, refKey)
	if err != nil {
		return err
	}
	esiKey := esi.makeKey(epoch, size, utxoID)
	key := esiKey.MarshalBinary()
	return utils.DeleteValue(txn, key)
}

func (esi *ExpSizeIndex) GetExpiredObjects(txn *badger.Txn, epoch uint32, maxBytes uint32, maxObjects int) ([][]byte, uint32) {
	result := [][]byte{}
	byteCount := uint32(0)
	objCount := 0
	opts := badger.DefaultIteratorOptions
	prefix := esi.prefix()
	opts.Prefix = prefix
	iter := txn.NewIterator(opts)
	defer iter.Close()
	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
		if objCount+1 > maxObjects {
			break
		}
		if byteCount+constants.HashLen > maxBytes {
			break
		}
		itm := iter.Item()
		key := itm.KeyCopy(nil)
		key = key[len(prefix):]
		epochBytes := key[:4]
		// slice is 4 bytes so no error will be raised
		epochObj, _ := utils.UnmarshalUint32(epochBytes)
		utxoID := key[8:]
		if epochObj > epoch {
			break
		}
		byteCount += constants.HashLen
		result = append(result, utxoID)
		objCount++
	}
	remainingBytes := maxBytes - byteCount
	return result, remainingBytes
}

func (esi *ExpSizeIndex) makeKey(epoch uint32, size uint32, utxoID []byte) *ExpSizeIndexKey {
	utxoIDCopy := utils.CopySlice(utxoID)
	key := []byte{}
	key = append(key, esi.prefix()...)
	epochBytes := utils.MarshalUint32(epoch)
	key = append(key, epochBytes...)
	sizeInv := constants.MaxUint32 - size
	sizeInvBytes := utils.MarshalUint32(sizeInv)
	key = append(key, sizeInvBytes...)
	key = append(key, utxoIDCopy...)
	esiKey := &ExpSizeIndexKey{}
	esiKey.UnmarshalBinary(key)
	return esiKey
}

func (esi *ExpSizeIndex) makeRefKey(utxoID []byte) *ExpSizeIndexRefKey {
	utxoIDCopy := utils.CopySlice(utxoID)
	key := []byte{}
	key = append(key, esi.refPrefix()...)
	key = append(key, utxoIDCopy...)
	esiRefKey := &ExpSizeIndexRefKey{}
	esiRefKey.UnmarshalBinary(key)
	return esiRefKey
}

func (esi *ExpSizeIndex) makeRefValue(epoch uint32, size uint32) []byte {
	epochBytes := utils.MarshalUint32(epoch)
	sizeInv := constants.MaxUint32 - size
	sizeInvBytes := utils.MarshalUint32(sizeInv)
	refValue := []byte{}
	refValue = append(refValue, epochBytes...)
	refValue = append(refValue, sizeInvBytes...)
	return refValue
}
