package indexer

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/application/objs/uint256"
	trie "github.com/MadBase/MadNet/badgerTrie"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/internal/testing/environment"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
)

func TestRefCounter(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	hash := trie.Hasher([]byte("foo"))
	rc := &RefCounter{func() []byte { return []byte("aa") }}
	err := db.Update(func(txn *badger.Txn) error {
		for i := 1; i < 100; i++ {
			v, err := rc.Increment(txn, hash)
			if err != nil {
				t.Error(err)
			}
			if v != int64(i) {
				t.Error("bad count after increment", v)
			}
		}
		for i := 1; i < 99; i++ {
			v, err := rc.Decrement(txn, hash)
			if err != nil {
				t.Error(err)
			}
			if v != int64(99-i) {
				t.Error("bad count after increment", v)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		v, err := rc.Decrement(txn, hash)
		if err != nil {
			t.Error(err)
		}
		if v != 0 {
			t.Error("bad count after increment", v)
		}
		_, err = utils.GetValue(txn, append([]byte("aa"), hash...))
		if err != nil {
			if err != badger.ErrKeyNotFound {
				t.Error(err)
			}
		}
		if err == nil {
			t.Error("ref counter not cleaned up after decrement to zero")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEPC(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	epc := &EpochConstrainedList{func() []byte { return []byte("aa") }, func() []byte { return []byte("ab") }}
	txHash1 := trie.Hasher([]byte("foo"))
	epoch1 := uint32(1)
	txHash2 := trie.Hasher([]byte("bar"))
	epoch2 := uint32(2)
	err := db.Update(func(txn *badger.Txn) error {
		eclTx1Key := epc.makeKey(epoch1, txHash1)
		tx1Key := eclTx1Key.MarshalBinary()
		err := epc.Append(txn, epoch1, txHash1)
		if err != nil {
			t.Error(err)
		}
		err = epc.Drop(txn, txHash1)
		if err != nil {
			t.Error(err)
		}
		_, err = utils.GetValue(txn, tx1Key)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				t.Error(err)
			}
		}
		if err == nil {
			t.Error("epc not dropped")
		}
		err = epc.Append(txn, epoch1, txHash1)
		if err != nil {
			t.Error(err)
		}
		err = epc.Append(txn, epoch2, txHash2)
		if err != nil {
			t.Error(err)
		}
		dropkeys, err := epc.DropBefore(txn, epoch2)
		if err != nil {
			t.Error(err)
		}
		if len(dropkeys) != 1 {
			t.Error("dropkey length wrong")
		}
		if !bytes.Equal(dropkeys[0], txHash1) {
			t.Error("wrong dropkey")
		}
		_, err = utils.GetValue(txn, tx1Key)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				t.Error(err)
			}
		}
		if err == nil {
			t.Error("epc not dropped")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIOI(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	ioi := &InsertionOrderIndexer{func() []byte { return []byte("aa") }, func() []byte { return []byte("ab") }}
	txHash1 := trie.Hasher([]byte("foo"))
	txHash2 := trie.Hasher([]byte("bar"))
	txHash3 := trie.Hasher([]byte("baz"))
	err := db.Update(func(txn *badger.Txn) error {
		ioiIdxKey, ioiRevIdxKey, err := ioi.makeIndexKeys(txHash1)
		if err != nil {
			t.Fatal(err)
		}
		idxKey := ioiIdxKey.MarshalBinary()
		revIdxKey := ioiRevIdxKey.MarshalBinary()
		err = ioi.Add(txn, txHash1)
		if err != nil {
			t.Fatal(err)
		}
		err = ioi.Delete(txn, txHash1)
		if err != nil {
			t.Fatal(err)
		}
		_, err = utils.GetValue(txn, idxKey)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				t.Error(err)
			}
		}
		if err == nil {
			t.Error("idxKey not dropped")
		}
		_, err = utils.GetValue(txn, revIdxKey)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				t.Error(err)
			}
		}
		if err == nil {
			t.Error("revIdxKey not dropped")
		}
		err = ioi.Add(txn, txHash1)
		if err != nil {
			t.Fatal(err)
		}
		err = ioi.Add(txn, txHash2)
		if err != nil {
			t.Fatal(err)
		}
		err = ioi.Add(txn, txHash3)
		if err != nil {
			t.Fatal(err)
		}
		i := 0
		it, prefix := ioi.NewIter(txn)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			itm := it.Item()
			vBytes, err := itm.ValueCopy(nil)
			if err != nil {
				if i != 3 {
					t.Fatal(err)
				}
				if err != ErrIterClose {
					t.Fatal(err)
				}
				break
			}
			txHash := vBytes[len(prefix):]
			if i == 0 {
				if !bytes.Equal(txHash, txHash1) {
					t.Errorf("wrong insert order txHash1 should be %x is %x\n", txHash1, txHash)
				}
			}
			if i == 1 {
				if !bytes.Equal(txHash, txHash2) {
					t.Errorf("wrong insert order txHash2 should be %x is %x\n", txHash2, txHash)
				}
			}
			if i == 2 {
				if !bytes.Equal(txHash, txHash3) {
					t.Errorf("wrong insert order txHash2 should be %x is %x\n", txHash3, txHash)
				}
			}
			if i == 3 {
				t.Fatal("iter did not stop")
			}
			i++
		}
		if i != 3 {
			t.Fatalf("Stopped at index: %v", i)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRefLinker(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	prefix1 := func() []byte {
		return []byte("za")
	}

	prefix2 := func() []byte {
		return []byte("zb")
	}

	prefix3 := func() []byte {
		return []byte("zk")
	}

	// object to delete
	txHash1 := trie.Hasher([]byte("foo"))
	utxoIDs1 := [][]byte{}
	utxoIDs1 = append(utxoIDs1, trie.Hasher([]byte("a")))
	utxoIDs1 = append(utxoIDs1, trie.Hasher([]byte("b")))
	utxoIDs1 = append(utxoIDs1, trie.Hasher([]byte("c")))

	// object to be deleted due to overlap of utxoID
	txHash2 := trie.Hasher([]byte("bar"))
	utxoIDs2 := [][]byte{}
	utxoIDs2 = append(utxoIDs2, trie.Hasher([]byte("a")))
	utxoIDs2 = append(utxoIDs2, trie.Hasher([]byte("f")))
	utxoIDs2 = append(utxoIDs2, trie.Hasher([]byte("g")))

	// object that should remain
	txHash3 := trie.Hasher([]byte("baz"))
	utxoIDs3 := [][]byte{}
	utxoIDs3 = append(utxoIDs3, trie.Hasher([]byte("h")))
	utxoIDs3 = append(utxoIDs3, trie.Hasher([]byte("i")))
	utxoIDs3 = append(utxoIDs3, trie.Hasher([]byte("j")))

	// object that is deleted as single element
	txHash4 := trie.Hasher([]byte("boo"))
	utxoIDs4 := [][]byte{}
	utxoIDs4 = append(utxoIDs4, trie.Hasher([]byte("a")))
	utxoIDs4 = append(utxoIDs4, trie.Hasher([]byte("f")))
	utxoIDs4 = append(utxoIDs4, trie.Hasher([]byte("i")))

	rl := NewRefLinkerIndex(prefix1, prefix2, prefix3)
	err := db.Update(func(txn *badger.Txn) error {
		_, _, err := rl.Add(txn, txHash1, utxoIDs1)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = rl.Add(txn, txHash2, utxoIDs2)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = rl.Add(txn, txHash3, utxoIDs3)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = rl.Add(txn, txHash4, utxoIDs4)
		if err != nil {
			t.Fatal(err)
		}
		err = rl.Delete(txn, txHash4)
		if err != nil {
			t.Fatal(err)
		}
		txHashes, _, err := rl.DeleteMined(txn, txHash1)
		if err != nil {
			t.Fatal(err)
		}
		if len(txHashes) != 2 {
			t.Fatal("wrong txHashes length", len(txHashes))
		}
		if !bytes.Equal(txHashes[0], txHash1) {
			t.Fatal("txHash1 not at index zero")
		}
		if !bytes.Equal(txHashes[1], txHash2) {
			t.Fatal("txHash2 not at index one")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRefLinkerEvict(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	prefix1 := func() []byte {
		return []byte("za")
	}

	prefix2 := func() []byte {
		return []byte("zb")
	}

	prefix3 := func() []byte {
		return []byte("zk")
	}

	// object to delete
	txHash1 := trie.Hasher([]byte("foo"))
	utxoIDs1 := [][]byte{}
	utxoIDs1 = append(utxoIDs1, trie.Hasher([]byte("a")))
	utxoIDs1 = append(utxoIDs1, trie.Hasher([]byte("b")))
	utxoIDs1 = append(utxoIDs1, trie.Hasher([]byte("c")))

	// object to be deleted due to overlap of utxoID
	txHash2 := trie.Hasher([]byte("bar"))
	utxoIDs2 := [][]byte{}
	utxoIDs2 = append(utxoIDs2, trie.Hasher([]byte("a")))
	utxoIDs2 = append(utxoIDs2, trie.Hasher([]byte("b")))
	utxoIDs2 = append(utxoIDs2, trie.Hasher([]byte("c")))

	// object that should remain
	txHash3 := trie.Hasher([]byte("baz"))
	utxoIDs3 := [][]byte{}
	utxoIDs3 = append(utxoIDs3, trie.Hasher([]byte("a")))
	utxoIDs3 = append(utxoIDs3, trie.Hasher([]byte("b")))
	utxoIDs3 = append(utxoIDs3, trie.Hasher([]byte("c")))

	// object that is deleted as single element
	txHash4 := trie.Hasher([]byte("boo"))
	utxoIDs4 := [][]byte{}
	utxoIDs4 = append(utxoIDs4, trie.Hasher([]byte("a")))
	utxoIDs4 = append(utxoIDs4, trie.Hasher([]byte("b")))
	utxoIDs4 = append(utxoIDs4, trie.Hasher([]byte("c")))

	rl := NewRefLinkerIndex(prefix1, prefix2, prefix3)
	err := db.Update(func(txn *badger.Txn) error {
		_, _, err := rl.Add(txn, txHash1, utxoIDs1)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = rl.Add(txn, txHash2, utxoIDs2)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = rl.Add(txn, txHash3, utxoIDs3)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = rl.Add(txn, txHash4, utxoIDs4)
		if err != nil {
			t.Fatal(err)
		}
		txHashes, _, err := rl.DeleteMined(txn, txHash4)
		if err != nil {
			t.Fatal(err)
		}
		if len(txHashes) != 3 {
			t.Fatal("wrong txHashes length", len(txHashes))
		}
		if !bytes.Equal(txHashes[0], txHash2) {
			t.Fatal("txHash2 not at index zero")
		}
		if !bytes.Equal(txHashes[1], txHash4) {
			t.Fatal("txHash4 not at index one")
		}
		if !bytes.Equal(txHashes[2], txHash3) {
			t.Fatal("txHash3 not at index two")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestHeightIdx(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	prefix1 := func() []byte {
		return []byte("za")
	}

	prefix2 := func() []byte {
		return []byte("zb")
	}
	index := NewHeightIdxIndex(prefix1, prefix2)

	// object to delete
	txHash1 := trie.Hasher([]byte("foo"))
	height1 := uint32(1)
	idx1 := uint32(1)

	txHash2 := trie.Hasher([]byte("bar"))
	height2 := uint32(1)
	idx2 := uint32(2)

	txHash3 := trie.Hasher([]byte("baz"))
	height3 := uint32(2)
	idx3 := uint32(1)

	txHash4 := trie.Hasher([]byte("boo"))
	height4 := uint32(2)
	idx4 := uint32(2)

	err := db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, txHash1, height1, idx1)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Add(txn, txHash2, height2, idx2)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Add(txn, txHash3, height3, idx3)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Add(txn, txHash4, height4, idx4)
		if err != nil {
			t.Fatal(err)
		}
		txHash, err := index.GetTxHashFromHeightIdx(txn, height4, idx4)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(txHash, txHash4) {
			t.Fatal("txhash4 wrong")
		}
		height, idx, err := index.GetHeightIdx(txn, txHash4)
		if err != nil {
			t.Fatal(err)
		}
		if height != height4 {
			t.Fatal("txhash4 height wrong")
		}
		if idx != idx4 {
			t.Fatal("txhash4 height wrong")
		}
		err = index.Delete(txn, txHash4)
		if err != nil {
			t.Fatal(err)
		}
		_, err = index.GetTxHashFromHeightIdx(txn, height4, idx4)
		if err != badger.ErrKeyNotFound {
			t.Fatal(err)
		}
		_, _, err = index.GetHeightIdx(txn, txHash4)
		if err != badger.ErrKeyNotFound {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestExpSizeIndex(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	prefix1 := func() []byte {
		return []byte("za")
	}

	prefix2 := func() []byte {
		return []byte("zb")
	}
	index := NewExpSizeIndex(prefix1, prefix2)

	// object to delete
	txHash1 := trie.Hasher([]byte("foo"))
	epoch1 := uint32(1)
	size1 := uint32(1)

	txHash2 := trie.Hasher([]byte("bar"))
	epoch2 := uint32(1)
	size2 := uint32(2)

	txHash3 := trie.Hasher([]byte("baz"))
	epoch3 := uint32(2)
	size3 := uint32(1)

	txHash4 := trie.Hasher([]byte("boo"))
	epoch4 := uint32(2)
	size4 := uint32(2)

	err := db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, epoch1, txHash1, size1)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Add(txn, epoch2, txHash2, size2)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Add(txn, epoch3, txHash3, size3)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Add(txn, epoch4, txHash4, size4)
		if err != nil {
			t.Fatal(err)
		}
		maxBytes := uint32(99)
		expObjs, remainingBytes := index.GetExpiredObjects(txn, epoch1, maxBytes, constants.MaxTxVectorLength)
		if len(expObjs) != 2 {
			t.Fatal("expObjs len wrong", len(expObjs))
		}
		if !bytes.Equal(expObjs[0], txHash2) {
			t.Fatal("exp0 wrong")
		}
		if !bytes.Equal(expObjs[1], txHash1) {
			t.Fatal("exp1 wrong")
		}
		if maxBytes-remainingBytes != 2*constants.HashLen {
			t.Fatal("wrong size for remainingBytes")
		}
		err = index.Drop(txn, txHash1)
		if err != nil {
			t.Fatal(err)
		}
		maxBytes = 99
		expObjs, remainingBytes = index.GetExpiredObjects(txn, epoch1, maxBytes, constants.MaxTxVectorLength)
		if len(expObjs) != 1 {
			t.Fatal("expObjs len wrong", len(expObjs))
		}
		if !bytes.Equal(expObjs[0], txHash2) {
			t.Fatal("exp0 wrong")
		}
		if maxBytes-remainingBytes != constants.HashLen {
			t.Fatal("wrong size for remainingBytes")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDataIndex(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	prefix1 := func() []byte {
		return []byte("za")
	}

	prefix2 := func() []byte {
		return []byte("zb")
	}
	index := NewDataIndex(prefix1, prefix2)

	// object to delete
	utxoID := trie.Hasher([]byte("foo"))
	acct := trie.Hasher([]byte("bar"))[12:]
	owner := &objs.Owner{}
	err := owner.New(acct, constants.CurveSecp256k1)
	if err != nil {
		t.Fatal(err)
	}
	dataIndex := trie.Hasher([]byte("baz"))

	err = db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, utxoID, owner, dataIndex)
		if err != nil {
			t.Fatal(err)
		}
		utxoID2, err := index.GetUTXOID(txn, owner, dataIndex)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(utxoID2, utxoID) {
			t.Fatal("wrong utxoID")
		}
		err = index.Drop(txn, utxoID)
		if err != nil {
			t.Fatal(err)
		}
		utxoID2, err = index.GetUTXOID(txn, owner, dataIndex)
		if err != badger.ErrKeyNotFound {
			t.Fatalf("data index not deleted: %x", utxoID2)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestValueIndex(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	prefix1 := func() []byte {
		return []byte("za")
	}

	prefix2 := func() []byte {
		return []byte("zb")
	}
	index := NewValueIndex(prefix1, prefix2)
	value1, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	value5, err := new(uint256.Uint256).FromUint64(5)
	if err != nil {
		t.Fatal(err)
	}

	// object to delete
	acct1 := trie.Hasher([]byte("owner1"))[12:]
	owner1 := &objs.Owner{}
	err = owner1.New(acct1, constants.CurveSecp256k1)
	if err != nil {
		t.Fatal(err)
	}
	utxoido11 := trie.Hasher([]byte("utxo1_1"))
	utxoido15 := trie.Hasher([]byte("utxo1_5"))
	utxoidOld := trie.Hasher([]byte("utxo1_Old"))

	acct2 := trie.Hasher([]byte("owner2"))[12:]
	owner2 := &objs.Owner{}
	err = owner2.New(acct2, constants.CurveSecp256k1)
	if err != nil {
		t.Fatal(err)
	}
	utxoido21 := trie.Hasher([]byte("utxo2_1"))
	utxoido25 := trie.Hasher([]byte("utxo2_5"))

	err = db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, utxoido11, owner1, value1)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Add(txn, utxoido15, owner1, value5)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Add(txn, utxoido21, owner2, value1)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Add(txn, utxoido25, owner2, value5)
		if err != nil {
			t.Fatal(err)
		}

		utxoIDs, value, _, err := index.GetValueForOwner(txn, owner1, uint256.One(), nil, 256, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(utxoIDs) != 1 {
			t.Fatal("wrong utxo len")
		}
		if !value.Eq(uint256.One()) {
			t.Fatal("wrong utxo value")
		}
		if !bytes.Equal(utxoIDs[0], utxoido11) {
			t.Fatal("wrong utxo")
		}

		six, err := new(uint256.Uint256).FromUint64(6)
		if err != nil {
			t.Fatal(err)
		}
		utxoIDs, valueOut, _, err := index.GetValueForOwner(txn, owner1, six, nil, 256, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(utxoIDs) != 2 {
			t.Fatal("wrong utxo len")
		}
		if !valueOut.Eq(six) {
			t.Fatal("wrong utxo value")
		}
		if !bytes.Equal(utxoIDs[0], utxoido11) {
			t.Fatal("wrong utxo")
		}
		if !bytes.Equal(utxoIDs[1], utxoido15) {
			t.Fatal("wrong utxo")
		}

		big, err := new(uint256.Uint256).FromUint64(1e6)
		if err != nil {
			t.Fatal(err)
		}
		utxoIDs, value, lk, err := index.GetValueForOwner(txn, owner1, big, nil, 1, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(utxoIDs) != 1 {
			t.Fatal("wrong utxo len", len(utxoIDs))
		}
		if !value.Eq(value1) {
			t.Fatal("wrong utxo value")
		}
		if !bytes.Equal(utxoIDs[0], utxoido11) {
			t.Fatal("wrong utxo")
		}
		utxoIDs, value, _, err = index.GetValueForOwner(txn, owner1, big, nil, 2, lk)
		if err != nil {
			t.Fatal(err)
		}
		if len(utxoIDs) != 1 {
			t.Fatal("wrong utxo len", len(utxoIDs))
		}
		if !value.Eq(value5) {
			t.Fatal("wrong utxo value")
		}
		if !bytes.Equal(utxoIDs[0], utxoido15) {
			t.Fatal("wrong utxo")
		}

		err = index.Drop(txn, utxoido11)
		if err != nil {
			t.Fatal(err)
		}
		utxoIDs, value, _, err = index.GetValueForOwner(txn, owner1, uint256.One(), nil, 256, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(utxoIDs) != 1 {
			t.Fatal("wrong utxo len")
		}
		if !value.Eq(value5) {
			t.Fatal("wrong utxo value")
		}
		if !bytes.Equal(utxoIDs[0], utxoido15) {
			t.Fatal("wrong utxo")
		}

		err = index.Add(txn, utxoidOld, owner1, uint256.One())
		if err != nil {
			t.Fatal(err)
		}
		utxoIDs, value, _, err = index.GetValueForOwner(txn, owner1, uint256.BaseDatasizeConst(), nil, 256, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(utxoIDs) != 2 {
			t.Fatal("wrong utxo len")
		}
		if !value.Eq(six) {
			fmt.Println(value)
			fmt.Println(six)
			t.Fatal("wrong utxo value")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
