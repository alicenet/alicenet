package indexer

import (
	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

/*
Key:      <prefix>|<feeCostRatio>|<txHash>
Value:    RefKey

RefKey:   <prefix>|<txHash>
RefValue: <feeCostRatio>

iterate in reverse direction; this ensures we iterate from largest to smallest
in terms of fee-dense transactions.
*/

// NewTxFeeIndex makes a new TxFeeIndex object
func NewTxFeeIndex(p, pp prefixFunc) *TxFeeIndex {
	return &TxFeeIndex{p, pp}
}

// TxFeeIndex creates an index that allows objects to be indexed by
// txfee ratio; that is fee per byte
type TxFeeIndex struct {
	prefix    prefixFunc
	refPrefix prefixFunc
}

// TxFeeIndexKey is the key for the TxFeeIndex
type TxFeeIndexKey struct {
	key []byte
}

// MarshalBinary returns the byte slice for the key object
func (tfik *TxFeeIndexKey) MarshalBinary() []byte {
	return utils.CopySlice(tfik.key)
}

// UnmarshalBinary takes in a byte slice to set the key object
func (tfik *TxFeeIndexKey) UnmarshalBinary(data []byte) {
	tfik.key = utils.CopySlice(data)
}

// TxFeeIndexRefKey is the reference key for the TxFeeIndex
type TxFeeIndexRefKey struct {
	refkey []byte
}

// MarshalBinary returns the byte slice for the key object
func (tfirk *TxFeeIndexRefKey) MarshalBinary() []byte {
	return utils.CopySlice(tfirk.refkey)
}

// UnmarshalBinary takes in a byte slice to set the key object
func (tfirk *TxFeeIndexRefKey) UnmarshalBinary(data []byte) {
	tfirk.refkey = utils.CopySlice(data)
}

// Add adds a transaction to the TxFeeIndex
func (tfi *TxFeeIndex) Add(txn *badger.Txn, feeCostRatio *uint256.Uint256, txHash []byte) error {
	feeCostRatioBytes, err := feeCostRatio.MarshalBinary()
	if err != nil {
		return err
	}
	tfiKey := tfi.makeKey(feeCostRatioBytes, txHash)
	key := tfiKey.MarshalBinary()
	tfiRefKey := tfi.makeRefKey(txHash)
	refKey := tfiRefKey.MarshalBinary()
	refValue := tfi.makeRefValue(feeCostRatioBytes)
	err = utils.SetValue(txn, refKey, refValue)
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, refKey)
}

// Drop drops a transaction from the TxFeeIndex
func (tfi *TxFeeIndex) Drop(txn *badger.Txn, txHash []byte) error {
	tfiRefKey := tfi.makeRefKey(txHash)
	refKey := tfiRefKey.MarshalBinary()
	feeCostRatioBytes, err := utils.GetValue(txn, refKey)
	if err != nil {
		return err
	}
	err = utils.DeleteValue(txn, refKey)
	if err != nil {
		return err
	}
	tfiKey := tfi.makeKey(feeCostRatioBytes, txHash)
	key := tfiKey.MarshalBinary()
	return utils.DeleteValue(txn, key)
}

func (tfi *TxFeeIndex) makeKey(feeCostRatioBytes []byte, txHash []byte) *TxFeeIndexKey {
	key := []byte{}
	key = append(key, tfi.prefix()...)
	key = append(key, utils.CopySlice(feeCostRatioBytes)...)
	key = append(key, utils.CopySlice(txHash)...)
	tfiKey := &TxFeeIndexKey{}
	tfiKey.UnmarshalBinary(key)
	return tfiKey
}

func (tfi *TxFeeIndex) makeRefKey(txHash []byte) *TxFeeIndexRefKey {
	key := []byte{}
	key = append(key, tfi.refPrefix()...)
	key = append(key, utils.CopySlice(txHash)...)
	tfiRefKey := &TxFeeIndexRefKey{}
	tfiRefKey.UnmarshalBinary(key)
	return tfiRefKey
}

func (tfi *TxFeeIndex) makeRefValue(feeCostRatioBytes []byte) []byte {
	refValue := []byte{}
	refValue = append(refValue, utils.CopySlice(feeCostRatioBytes)...)
	return refValue
}

// NewIter makes a new iterator for iterating through TxFeeIndex
func (tfi *TxFeeIndex) NewIter(txn *badger.Txn) (*badger.Iterator, []byte) {
	prefix := tfi.prefix()
	opts := badger.DefaultIteratorOptions
	opts.Reverse = true
	opts.Prefix = prefix
	return txn.NewIterator(opts), prefix
}
