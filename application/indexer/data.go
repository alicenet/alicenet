package indexer

import (
	"bytes"

	"github.com/alicenet/alicenet/errorz"

	"github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

// NewDataIndex creates a new dataIndex object
func NewDataIndex(p, pp prefixFunc) *DataIndex {
	return &DataIndex{p, pp}
}

// DataIndex creates an index that allows datastores to be indexed by the
// specified owner and index of the datastore
type DataIndex struct {
	prefix    prefixFunc
	refPrefix prefixFunc
}

type DataIndexKey struct {
	prefix []byte
	owner  []byte
	index  []byte
}

// MarshalBinary returns the byte slice for the key object
func (dik *DataIndexKey) MarshalBinary() []byte {
	key := []byte{}
	key = append(key, dik.prefix...)
	key = append(key, dik.owner...)
	key = append(key, dik.index...)
	return utils.CopySlice(key)
}

// UnmarshalBinary takes in a byte slice to set the key object
func (dik *DataIndexKey) UnmarshalBinary(data []byte) {
	dik.prefix = utils.CopySlice(data[0:2])
	dik.owner = utils.CopySlice(data[2 : 2+constants.OwnerLen+1])
	dik.index = utils.CopySlice(data[2+constants.OwnerLen+1:])
}

type DataIndexRefKey struct {
	refkey []byte
}

// MarshalBinary returns the byte slice for the key object
func (dirk *DataIndexRefKey) MarshalBinary() []byte {
	return utils.CopySlice(dirk.refkey)
}

// UnmarshalBinary takes in a byte slice to set the key object
func (dirk *DataIndexRefKey) UnmarshalBinary(data []byte) {
	dirk.refkey = utils.CopySlice(data)
}

// Add adds an item to the index
func (di *DataIndex) Add(txn *badger.Txn, utxoID []byte, owner *objs.Owner, dataIndex []byte) error {
	return di.addInternal(txn, utxoID, owner, dataIndex, false)
}

// AddFastSync adds an item to the index and overwrites previous data
// if it is present
func (di *DataIndex) AddFastSync(txn *badger.Txn, utxoID []byte, owner *objs.Owner, dataIndex []byte) error {
	return di.addInternal(txn, utxoID, owner, dataIndex, true)
}

func (di *DataIndex) addInternal(txn *badger.Txn, utxoID []byte, owner *objs.Owner, dataIndex []byte, allowOverwrites bool) error {
	utxoIDCopy := utils.CopySlice(utxoID)
	dataIndexCopy := utils.CopySlice(dataIndex)
	diKey, err := di.makeKey(owner, dataIndexCopy)
	if err != nil {
		return err
	}
	key := diKey.MarshalBinary()
	if allowOverwrites {
		return di.addInternalOverwrite(txn, key, utxoIDCopy)
	}
	return di.addInternalNoOverwrite(txn, key, utxoIDCopy)
}

func (di *DataIndex) addInternalNoOverwrite(txn *badger.Txn, dik []byte, utxoID []byte) error {
	_, err := utils.GetValue(txn, dik)
	if err == nil {
		return errorz.ErrInvalid{}.New("dataIndex.addInternalNoOverwrite; index conflict")
	}
	if err == badger.ErrKeyNotFound {
		diRefKey := di.makeRefKey(utils.CopySlice(utxoID))
		refKey := diRefKey.MarshalBinary()
		err = utils.SetValue(txn, utils.CopySlice(refKey), utils.CopySlice(dik))
		if err != nil {
			return err
		}
		return utils.SetValue(txn, utils.CopySlice(dik), utils.CopySlice(utxoID))
	}
	return err
}

func (di *DataIndex) addInternalOverwrite(txn *badger.Txn, dik []byte, utxoID []byte) error {
	oldUtxoID, err := utils.GetValue(txn, dik)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			return err
		}
	}
	if err == nil {
		// need to delete existing refkey and overwrite
		oldRefKey := di.makeRefKey(oldUtxoID)
		oldRefKeyBytes := oldRefKey.MarshalBinary()
		err := utils.DeleteValue(txn, oldRefKeyBytes)
		if err != nil {
			return err
		}
	}
	refKey := di.makeRefKey(utils.CopySlice(utxoID))
	refKeyBytes := refKey.MarshalBinary()
	err = utils.SetValue(txn, refKeyBytes, utils.CopySlice(dik))
	if err != nil {
		return err
	}
	return utils.SetValue(txn, utils.CopySlice(dik), utils.CopySlice(utxoID))
}

// Contains determines whether or not the value at dataIndex is present
func (di *DataIndex) Contains(txn *badger.Txn, owner *objs.Owner, dataIndex []byte) (bool, error) {
	dataIndexCopy := utils.CopySlice(dataIndex)
	diKey, err := di.makeKey(owner, dataIndexCopy)
	if err != nil {
		return false, err
	}
	key := diKey.MarshalBinary()
	_, err = utils.GetValue(txn, key)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Drop removes a utxoID from the index
func (di *DataIndex) Drop(txn *badger.Txn, utxoID []byte) error {
	utxoIDCopy := utils.CopySlice(utxoID)
	diRefKey := di.makeRefKey(utxoIDCopy)
	refKey := diRefKey.MarshalBinary()
	totalDataIndex, err := utils.GetValue(txn, refKey)
	if err != nil {
		return err
	}
	if err := utils.DeleteValue(txn, refKey); err != nil {
		return err
	}
	return utils.DeleteValue(txn, totalDataIndex)
}

// GetUTXOID returns the UTXOID of a datastore based on owner and index
func (di *DataIndex) GetUTXOID(txn *badger.Txn, owner *objs.Owner, dataIndex []byte) ([]byte, error) {
	dataIndexCopy := utils.CopySlice(dataIndex)
	diKey, err := di.makeKey(owner, dataIndexCopy)
	if err != nil {
		return nil, err
	}
	key := diKey.MarshalBinary()
	return utils.GetValue(txn, key)
}

// PaginateDataStores returns utxoIDs for owner below specified num
func (di *DataIndex) PaginateDataStores(txn *badger.Txn, owner *objs.Owner, num int, startIndex []byte, exclude map[string]bool) ([]*objs.PaginationResponse, error) {
	if len(startIndex) == 0 {
		startIndex = make([]byte, constants.HashLen)
	}
	// PaginateDataStores is called by UTXOHandler.PaginateDataByOwner,
	// and PaginateDataByOwner numItems (num in this function)
	// is of type uint8. Thus, the largest value num can take is 255
	// and this statement should not be used.
	if num > 256 {
		num = 256
	}
	result := []*objs.PaginationResponse{}
	itemCount := 0
	opts := badger.DefaultIteratorOptions
	prefix, err := di.makeIterKey(owner)
	if err != nil {
		return nil, err
	}
	diSeekKey, err := di.makeKey(owner, startIndex)
	if err != nil {
		return nil, err
	}
	seekKey := diSeekKey.MarshalBinary()
	opts.Prefix = prefix
	iter := txn.NewIterator(opts)
	defer iter.Close()
	for iter.Seek(seekKey); iter.ValidForPrefix(prefix); iter.Next() {
		itm := iter.Item()
		key := itm.KeyCopy(nil)
		dk := &DataIndexKey{}
		dk.UnmarshalBinary(key)
		utxoID, err := itm.ValueCopy(nil)
		if err != nil {
			return nil, err
		}
		exclude[string(utxoID)] = true
		if itemCount == 0 && num != 1 {
			if bytes.Equal(dk.index, startIndex) {
				continue
			}
		}
		itemCount++
		result = append(result, &objs.PaginationResponse{
			UTXOID: utils.CopySlice(utxoID),
			Index:  utils.CopySlice(dk.index),
		})
		if itemCount >= num {
			break
		}
	}
	return result, nil
}

func (di *DataIndex) makeIterKey(owner *objs.Owner) ([]byte, error) {
	ownerBytes, err := owner.MarshalBinary()
	if err != nil {
		return nil, err
	}
	key := []byte{}
	key = append(key, di.prefix()...)
	key = append(key, ownerBytes...)
	return key, nil
}

func (di *DataIndex) makeKey(owner *objs.Owner, dataIndex []byte) (*DataIndexKey, error) {
	ownerBytes, err := owner.MarshalBinary()
	if err != nil {
		return nil, err
	}
	diKey := &DataIndexKey{
		prefix: utils.CopySlice(di.prefix()),
		owner:  ownerBytes,
		index:  dataIndex,
	}
	return diKey, nil
}

func (di *DataIndex) makeRefKey(utxoID []byte) *DataIndexRefKey {
	key := []byte{}
	key = append(key, di.refPrefix()...)
	key = append(key, utxoID...)
	diRefKey := &DataIndexRefKey{}
	diRefKey.UnmarshalBinary(key)
	return diRefKey
}
