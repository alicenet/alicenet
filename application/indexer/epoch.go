package indexer

import (
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

func NewEpochConstrainedIndex(p, pp prefixFunc) *EpochConstrainedList {
	return &EpochConstrainedList{p, pp}
}

// EpochConstrainedList creates an index that allows objects to be dropped
// by epoch
type EpochConstrainedList struct {
	prefix    prefixFunc
	refPrefix prefixFunc
}

type EpochConstrainedListKey struct {
	key []byte
}

// MarshalBinary returns the byte slice for the key object
func (eclk *EpochConstrainedListKey) MarshalBinary() []byte {
	return utils.CopySlice(eclk.key)
}

// UnmarshalBinary takes in a byte slice to set the key object
func (eclk *EpochConstrainedListKey) UnmarshalBinary(data []byte) {
	eclk.key = utils.CopySlice(data)
}

type EpochConstrainedListRefKey struct {
	refkey []byte
}

// MarshalBinary returns the byte slice for the key object
func (eclrk *EpochConstrainedListRefKey) MarshalBinary() []byte {
	return utils.CopySlice(eclrk.refkey)
}

// UnmarshalBinary takes in a byte slice to set the key object
func (eclrk *EpochConstrainedListRefKey) UnmarshalBinary(data []byte) {
	eclrk.refkey = utils.CopySlice(data)
}

// Append adds an item to the list
func (ecl *EpochConstrainedList) Append(txn *badger.Txn, epoch uint32, txHash []byte) error {
	txHashCopy := utils.CopySlice(txHash)
	eclKey := ecl.makeKey(epoch, txHashCopy)
	key := eclKey.MarshalBinary()
	eclRefKey := ecl.makeRefKey(txHashCopy)
	refKey := eclRefKey.MarshalBinary()
	epochBytes := utils.MarshalUint32(epoch)
	err := utils.SetValue(txn, refKey, epochBytes)
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, []byte{})
}

// DropBefore returns a list of all txHashes that should be dropped
func (ecl *EpochConstrainedList) DropBefore(txn *badger.Txn, epoch uint32) ([][]byte, error) {
	dropKeys := [][]byte{}
	dropHashes := [][]byte{}
	prefix := ecl.prefix()
	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix
	it := txn.NewIterator(opts)
	defer it.Close()
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		k := item.KeyCopy(nil)
		keyNoPrefix := k[len(ecl.prefix()):]
		keyEpochBytes := keyNoPrefix[0:4]
		txHash := keyNoPrefix[4:]
		keyEpoch, err := utils.UnmarshalUint32(keyEpochBytes)
		if err != nil {
			return nil, err
		}
		if keyEpoch < epoch {
			dropKeys = append(dropKeys, k)
			dropHashes = append(dropHashes, txHash)
		}
	}
	for j := 0; j < len(dropKeys); j++ {
		k := utils.CopySlice(dropKeys[j])
		err := utils.DeleteValue(txn, utils.CopySlice(k))
		if err != nil {
			return nil, err
		}
	}
	return dropHashes, nil
}

func (ecl *EpochConstrainedList) Drop(txn *badger.Txn, txHash []byte) error {
	txHashCopy := utils.CopySlice(txHash)
	eclRefKey := ecl.makeRefKey(txHashCopy)
	refKey := eclRefKey.MarshalBinary()
	epochBytes, err := utils.GetValue(txn, refKey)
	if err != nil {
		return err
	}
	epoch, err := utils.UnmarshalUint32(epochBytes)
	if err != nil {
		return err
	}
	eclKey := ecl.makeKey(epoch, txHashCopy)
	key := eclKey.MarshalBinary()
	return utils.DeleteValue(txn, key)
}

func (ecl *EpochConstrainedList) GetEpoch(txn *badger.Txn, txHash []byte) (uint32, error) {
	txHashCopy := utils.CopySlice(txHash)
	eclRefKey := ecl.makeRefKey(txHashCopy)
	refKey := eclRefKey.MarshalBinary()
	epochBytes, err := utils.GetValue(txn, refKey)
	if err != nil {
		return 0, err
	}
	epoch, err := utils.UnmarshalUint32(epochBytes)
	if err != nil {
		return 0, err
	}
	return epoch, nil
}

func (ecl *EpochConstrainedList) makeKey(epoch uint32, txHash []byte) *EpochConstrainedListKey {
	txHashCopy := utils.CopySlice(txHash)
	key := []byte{}
	key = append(key, ecl.prefix()...)
	epochBytes := utils.MarshalUint32(epoch)
	key = append(key, epochBytes...)
	key = append(key, txHashCopy...)
	eclKey := &EpochConstrainedListKey{}
	eclKey.UnmarshalBinary(key)
	return eclKey
}

func (ecl *EpochConstrainedList) makeRefKey(txHash []byte) *EpochConstrainedListRefKey {
	txHashCopy := utils.CopySlice(txHash)
	key := []byte{}
	key = append(key, ecl.refPrefix()...)
	key = append(key, txHashCopy...)
	eclRefKey := &EpochConstrainedListRefKey{}
	eclRefKey.UnmarshalBinary(key)
	return eclRefKey
}
