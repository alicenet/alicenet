package indexer

import (
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

/*

provides reference linking for pending transactions

a refcounter is maintained

for each referenced mined utxo there is a custom prefix - utxoID
k1: <prefix1>|<utxoID>|<current txHash>
  <utxoID>

k2: <prefix2>|<current txHash>|<utxoID >
  <utxoID>|<current txHash>

onAdd:
      create k1
      create k2

onDeleteOne:
      iterate k2, do for each: create k1 from k2, delete k1, delete k2

on mined tx:
      iterate k2, do for each:
          get list of mined utxoID
          make k1/k2 list for this tx
          for each utxoID:
            iterate <prefix1>|<mined utxoID>, do for each:
                get <current txHash> from each item
                invoke deleteOne on item

*/

func NewRefLinkerIndex(p, pp, ppp prefixFunc) *RefLinker {
	refCounter := NewRefCounter(ppp)
	return &RefLinker{3, p, pp, refCounter}
}

type RefLinker struct {
	threshold    int64
	prefixRef    prefixFunc
	prefixRevRef prefixFunc
	refCounter   *RefCounter
}

type RefLinkerRefKey struct {
	refkey []byte
}

// MarshalBinary returns the byte slice for the key object
func (rlrk *RefLinkerRefKey) MarshalBinary() []byte {
	return utils.CopySlice(rlrk.refkey)
}

// UnmarshalBinary takes in a byte slice to set the key object
func (rlrk *RefLinkerRefKey) UnmarshalBinary(data []byte) {
	rlrk.refkey = utils.CopySlice(data)
}

type RefLinkerRevRefKey struct {
	revrefkey []byte
}

// MarshalBinary returns the byte slice for the key object
func (rlrrk *RefLinkerRevRefKey) MarshalBinary() []byte {
	return utils.CopySlice(rlrrk.revrefkey)
}

// UnmarshalBinary takes in a byte slice to set the key object
func (rlrrk *RefLinkerRevRefKey) UnmarshalBinary(data []byte) {
	rlrrk.revrefkey = utils.CopySlice(data)
}

func (rlrrk *RefLinkerRevRefKey) XXXIsKey() {}

func (rl *RefLinker) evictOne(txn *badger.Txn, utxoID []byte) ([]byte, error) {
	utxoIDCopy := utils.CopySlice(utxoID)
	var evictedHash []byte
	opts := badger.DefaultIteratorOptions
	prefix := append(rl.prefixRevRef(), utxoIDCopy...)
	opts.Prefix = prefix
	iter := txn.NewIterator(opts)
	iter.Seek(prefix)
	itm := iter.Item()
	refKey, err := itm.ValueCopy(nil)
	iter.Close()
	if err != nil {
		return nil, err
	}
	evictedHash = refKey[len(refKey)-64 : len(refKey)-32]
	err = rl.Delete(txn, evictedHash)
	if err != nil {
		return nil, err
	}
	return evictedHash, nil
}

func (rl *RefLinker) Add(txn *badger.Txn, txHash []byte, utxoIDs [][]byte) (bool, [][]byte, error) {
	evictions := [][]byte{}
	for i := 0; i < len(utxoIDs); i++ {
		utxoID := utils.CopySlice(utxoIDs[i])
		count, err := rl.refCounter.Increment(txn, utxoID)
		if err != nil {
			return false, nil, err
		}
		if count > rl.threshold {
			txHashEvicted, err := rl.evictOne(txn, utxoID)
			if err != nil {
				return false, nil, err
			}
			evictions = append(evictions, txHashEvicted)
		}
		rlRefKey := rl.makeRefKey(txHash, utxoID)
		refKey := rlRefKey.MarshalBinary()
		rlRevRefKey := rl.makeRevRefKey(txHash, utxoID)
		revRefKey := rlRevRefKey.MarshalBinary()
		err = utils.SetValue(txn, refKey, utxoID)
		if err != nil {
			return false, nil, err
		}
		err = utils.SetValue(txn, revRefKey, refKey)
		if err != nil {
			return false, nil, err
		}
	}
	if len(evictions) == 0 {
		return false, nil, nil
	}
	return true, evictions, nil
}

func (rl *RefLinker) DeleteMined(txn *badger.Txn, txHash []byte) ([][]byte, [][]byte, error) {
	txHashCopy := utils.CopySlice(txHash)
	utxoIDs := [][]byte{}
	txHashes := [][]byte{}
	txMap := make(map[string]bool)
	fn1 := func() error {
		opts := badger.DefaultIteratorOptions
		prefix := append(rl.prefixRef(), txHashCopy...)
		opts.Prefix = prefix
		iter := txn.NewIterator(opts)
		defer iter.Close()
		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			itm := iter.Item()
			utxoID, err := itm.ValueCopy(nil)
			if err != nil {
				return err
			}
			utxoIDs = append(utxoIDs, utxoID)
		}
		return nil
	}
	fn2 := func(utxoID []byte) error {
		utxoIDCopy := utils.CopySlice(utxoID)
		_, err := rl.refCounter.Decrement(txn, utxoIDCopy)
		if err != nil {
			return err
		}
		opts := badger.DefaultIteratorOptions
		prefix := append(rl.prefixRevRef(), utxoIDCopy...)
		opts.Prefix = prefix
		iter := txn.NewIterator(opts)
		defer iter.Close()
		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			itm := iter.Item()
			revRefKey := itm.KeyCopy(nil)
			refKey, err := itm.ValueCopy(nil)
			if err != nil {
				return err
			}
			txHash := refKey[len(refKey)-64 : len(refKey)-32]
			if !txMap[string(txHash)] {
				txHashes = append(txHashes, txHash[:])
				txMap[string(txHash)] = true
			}
			err = utils.DeleteValue(txn, refKey)
			if err != nil {
				return err
			}
			err = utils.DeleteValue(txn, revRefKey)
			if err != nil {
				return err
			}
		}
		return nil
	}
	err := fn1()
	if err != nil {
		return nil, nil, err
	}
	for _, utxoID := range utxoIDs {
		err := fn2(utils.CopySlice(utxoID))
		if err != nil {
			return nil, nil, err
		}
	}
	return txHashes, utxoIDs, nil
}

func (rl *RefLinker) Delete(txn *badger.Txn, txHash []byte) error {
	txHashCopy := utils.CopySlice(txHash)
	opts := badger.DefaultIteratorOptions
	prefix := append(rl.prefixRef(), txHashCopy...)
	opts.Prefix = prefix
	iter := txn.NewIterator(opts)
	defer iter.Close()
	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
		itm := iter.Item()
		refKey := itm.KeyCopy(nil)
		utxoID, err := itm.ValueCopy(nil)
		if err != nil {
			return err
		}
		_, err = rl.refCounter.Decrement(txn, utxoID)
		if err != nil {
			return err
		}
		rlRevRefKey := rl.makeRevRefKey(txHash, utxoID)
		revRefKey := rlRevRefKey.MarshalBinary()
		err = utils.DeleteValue(txn, refKey)
		if err != nil {
			return err
		}
		err = utils.DeleteValue(txn, revRefKey)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rl *RefLinker) makeRefKey(txHash []byte, utxoID []byte) *RefLinkerRefKey {
	refKey := []byte{}
	refKey = append(refKey, rl.prefixRef()...)
	refKey = append(refKey, utils.CopySlice(txHash)...)
	refKey = append(refKey, utils.CopySlice(utxoID)...)
	rlRefKey := &RefLinkerRefKey{}
	rlRefKey.UnmarshalBinary(refKey)
	return rlRefKey
}

func (rl *RefLinker) makeRevRefKey(txHash []byte, utxoID []byte) *RefLinkerRevRefKey {
	revRefKey := []byte{}
	revRefKey = append(revRefKey, rl.prefixRevRef()...)
	revRefKey = append(revRefKey, utils.CopySlice(utxoID)...)
	revRefKey = append(revRefKey, utils.CopySlice(txHash)...)
	rlRevRefKey := &RefLinkerRevRefKey{}
	rlRevRefKey.UnmarshalBinary(revRefKey)
	return rlRevRefKey
}
