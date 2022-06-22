package indexer

/*
Given txHash get height index
  <prefix>|<txHash>
      <height>|<idx>

given height index get txHash
  <prefix>|<height>|<idx>
      <txHash>
*/

import (
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

func NewHeightIdxIndex(p, pp prefixFunc) *HeightIdxIndex {
	return &HeightIdxIndex{p, pp}
}

type HeightIdxIndex struct {
	prefix    prefixFunc
	prefixRef prefixFunc
}

type HeightIdxIndexKey struct {
	key []byte
}

// MarshalBinary returns the byte slice for the key object
func (hiik *HeightIdxIndexKey) MarshalBinary() []byte {
	return utils.CopySlice(hiik.key)
}

// UnmarshalBinary takes in a byte slice to set the key object
func (hiik *HeightIdxIndexKey) UnmarshalBinary(data []byte) {
	hiik.key = utils.CopySlice(data)
}

type HeightIdxIndexRefKey struct {
	refkey []byte
}

// MarshalBinary returns the byte slice for the key object
func (hiirk *HeightIdxIndexRefKey) MarshalBinary() []byte {
	return utils.CopySlice(hiirk.refkey)
}

// UnmarshalBinary takes in a byte slice to set the key object
func (hiirk *HeightIdxIndexRefKey) UnmarshalBinary(data []byte) {
	hiirk.refkey = utils.CopySlice(data)
}

func (hii *HeightIdxIndex) Add(txn *badger.Txn, txHash []byte, height uint32, idx uint32) error {
	hiiKey := hii.makeKey(txHash)
	key := hiiKey.MarshalBinary()
	hiiRefKey := hii.makeRefKey(height, idx)
	refKey := hiiRefKey.MarshalBinary()
	heightIdx := hii.makeHeightIdx(height, idx)
	err := utils.SetValue(txn, refKey, utils.CopySlice(txHash))
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, heightIdx)
}

func (hii *HeightIdxIndex) Delete(txn *badger.Txn, txHash []byte) error {
	hiiKey := hii.makeKey(txHash)
	key := hiiKey.MarshalBinary()
	heightIdx, err := utils.GetValue(txn, key)
	if err != nil {
		return err
	}
	height, idx, err := hii.getHeightIdx(heightIdx)
	if err != nil {
		return err
	}
	hiiRefKey := hii.makeRefKey(height, idx)
	refKey := hiiRefKey.MarshalBinary()
	err = utils.DeleteValue(txn, refKey)
	if err != nil {
		return err
	}
	return utils.DeleteValue(txn, key)
}

func (hii *HeightIdxIndex) GetHeightIdx(txn *badger.Txn, txHash []byte) (uint32, uint32, error) {
	hiiKey := hii.makeKey(txHash)
	key := hiiKey.MarshalBinary()
	heightIdx, err := utils.GetValue(txn, key)
	if err != nil {
		return 0, 0, err
	}
	return hii.getHeightIdx(heightIdx)
}

func (hii *HeightIdxIndex) GetTxHashFromHeightIdx(txn *badger.Txn, height uint32, idx uint32) ([]byte, error) {
	hiiRefKey := hii.makeRefKey(height, idx)
	refKey := hiiRefKey.MarshalBinary()
	return utils.GetValue(txn, refKey)
}

func (hii *HeightIdxIndex) makeKey(txHash []byte) *HeightIdxIndexKey {
	key := []byte{}
	key = append(key, hii.prefix()...)
	key = append(key, utils.CopySlice(txHash)...)
	hiiKey := &HeightIdxIndexKey{}
	hiiKey.UnmarshalBinary(key)
	return hiiKey
}

func (hii *HeightIdxIndex) makeRefKey(height uint32, idx uint32) *HeightIdxIndexRefKey {
	heightIdx := hii.makeHeightIdx(height, idx)
	refKey := make([]byte, 0, len(hii.prefixRef())+8)
	refKey = append(refKey, hii.prefixRef()...)
	refKey = append(refKey, heightIdx...)
	hiiRefKey := &HeightIdxIndexRefKey{}
	hiiRefKey.UnmarshalBinary(refKey)
	return hiiRefKey
}

func (hii *HeightIdxIndex) makeHeightIdx(height uint32, idx uint32) []byte {
	heightIdx := make([]byte, 0, 8)
	heightBytes := utils.MarshalUint32(height)
	heightIdx = append(heightIdx, heightBytes...)
	idxBytes := utils.MarshalUint32(idx)
	heightIdx = append(heightIdx, idxBytes...)
	return heightIdx
}

func (hii *HeightIdxIndex) getHeightIdx(heightIdx []byte) (uint32, uint32, error) {
	if len(heightIdx) != 8 {
		return 0, 0, errorz.ErrInvalid{}.New("heightIdxIndex.getHeightIdx: invalid byte length from key; should be 8")
	}
	heightBytes := heightIdx[:4]
	idxBytes := heightIdx[4:]
	// No errors are checked because both slices have length 4
	height, _ := utils.UnmarshalUint32(heightBytes)
	idx, _ := utils.UnmarshalUint32(idxBytes)
	return height, idx, nil
}
