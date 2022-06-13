/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
	"encoding/hex"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/internal/testing/environment"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"

	"fmt"
	"testing"
)

func prefixFn() []byte {
	return []byte("aa")
}

func TestSmtEmptyTrie(t *testing.T) {
	smt := NewSMT(nil, Hasher, prefixFn)
	if !bytes.Equal([]byte{}, smt.Root) {
		t.Fatal("empty trie root hash not correct")
	}
}

func testDb(t *testing.T, fn func(txn *badger.Txn) error) {
	t.Helper()
	db := environment.SetupBadgerDatabase(t)
	err := db.Update(fn)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSmtUpdateAndGet(t *testing.T) {
	t.Parallel()
	fn := func(txn *badger.Txn) error {
		smt := NewSMT(nil, Hasher, prefixFn)
		// Add data to empty trie
		keys := GetFreshData(10, 32)
		values := GetFreshData(10, 32)
		ch := make(chan mresult, 1)
		defer close(ch)
		smt.update(txn, smt.Root, keys, values, nil, 0, smt.TrieHeight, ch)
		res := <-ch
		root := res.update

		// Check all keys have been stored
		for i, key := range keys {
			value, _ := smt.get(txn, root, key, nil, 0, smt.TrieHeight)
			if !bytes.Equal(values[i], value) {
				t.Fatal("value not updated")
			}
		}

		// Append to the trie
		newKeys := GetFreshData(5, 32)
		newValues := GetFreshData(5, 32)
		ch = make(chan mresult, 1)
		defer close(ch)
		smt.update(txn, root, newKeys, newValues, nil, 0, smt.TrieHeight, ch)
		res = <-ch
		newRoot := res.update
		if bytes.Equal(root, newRoot) {
			t.Fatal("trie not updated")
		}
		for i, newKey := range newKeys {
			newValue, _ := smt.get(txn, newRoot, newKey, nil, 0, smt.TrieHeight)
			if !bytes.Equal(newValues[i], newValue) {
				t.Fatal("failed to get value")
			}
		}
		// Check old keys are still stored
		for i, key := range keys {
			value, _ := smt.get(txn, newRoot, key, nil, 0, smt.TrieHeight)
			if !bytes.Equal(values[i], value) {
				t.Fatal("failed to get value")
			}
		}
		return nil
	}
	testDb(t, fn)
}

func TestSmtPublicUpdateAndGet(t *testing.T) {
	t.Parallel()
	smt := NewSMT(nil, Hasher, prefixFn)
	fn := func(txn *badger.Txn) error {
		// Add data to empty trie
		keys := GetFreshData(5, 32)
		values := GetFreshData(5, 32)
		root, err := smt.Update(txn, keys, values)
		if err != nil {
			t.Fatal(err)
		}

		// Check all keys have been stored
		for i, key := range keys {
			value, _ := smt.Get(txn, key)
			if !bytes.Equal(values[i], value) {
				t.Fatal("trie not updated")
			}
		}
		if !bytes.Equal(root, smt.Root) {
			t.Fatal("Root not stored")
		}

		newValues := GetFreshData(5, 32)
		_, err = smt.Update(txn, keys, newValues)
		if err == nil {
			t.Fatal("multiple updates don't cause an error")
		}
		return nil
	}
	testDb(t, fn)
}

func TestSmtDelete(t *testing.T) {
	t.Parallel()
	fn := func(txn *badger.Txn) error {
		smt := NewSMT(nil, Hasher, prefixFn)
		// Add data to empty trie
		keys := GetFreshData(10, 32)
		values := GetFreshData(10, 32)
		ch := make(chan mresult, 1)
		defer close(ch)
		smt.update(txn, smt.Root, keys, values, nil, 0, smt.TrieHeight, ch)
		res := <-ch
		root := res.update
		value, _ := smt.get(txn, root, keys[0], nil, 0, smt.TrieHeight)
		if !bytes.Equal(values[0], value) {
			t.Fatal("trie not updated")
		}

		// Delete from trie
		// To delete a key, just set its value to Default leaf hash.
		ch = make(chan mresult, 1)
		defer close(ch)
		smt.update(txn, root, keys[0:1], [][]byte{DefaultLeaf}, nil, 0, smt.TrieHeight, ch)
		_, err := smt.Commit(txn, 1)
		if err != nil {
			t.Fatal(err)
		}
		res = <-ch
		newRoot := res.update
		newValue, _ := smt.get(txn, newRoot, keys[0], nil, 0, smt.TrieHeight)
		if len(newValue) != 0 {
			t.Fatal("Failed to delete from trie")
		}
		// Remove deleted key from keys and check root with a clean trie.
		smt2 := NewSMT(nil, Hasher, prefixFn)
		ch = make(chan mresult, 1)
		defer close(ch)
		smt2.update(txn, smt2.Root, keys[1:], values[1:], nil, 0, smt.TrieHeight, ch)
		_, err = smt2.Commit(txn, 1)
		if err != nil {
			t.Fatal(err)
		}
		res = <-ch
		cleanRoot := res.update
		if !bytes.Equal(newRoot, cleanRoot) {
			t.Fatal("roots mismatch")
		}

		//Empty the trie
		var newValues [][]byte
		for i := 0; i < 10; i++ {
			newValues = append(newValues, DefaultLeaf)
		}
		ch = make(chan mresult, 1)
		defer close(ch)
		smt.update(txn, root, keys, newValues, nil, 0, smt.TrieHeight, ch)
		res = <-ch
		root = res.update
		if len(root) != 0 {
			t.Fatal("empty trie root hash not correct")
		}
		// Test deleting an already empty key
		smt = NewSMT(nil, Hasher, prefixFn)
		keys = GetFreshData(2, 32)
		values = GetFreshData(2, 32)
		root, _ = smt.Update(txn, keys, values)
		key0 := make([]byte, 32)
		key1 := make([]byte, 32)
		_, err = smt.Update(txn, [][]byte{key0, key1}, [][]byte{DefaultLeaf, DefaultLeaf})
		if err == nil {
			t.Fatal("Should have raised an error")
		}
		if !bytes.Equal(root, smt.Root) {
			t.Fatal("deleting a default key shouldn't modify the tree")
		}
		return nil
	}
	testDb(t, fn)
}

// test updating and deleting at the same time
func TestTrieUpdateAndDelete(t *testing.T) {
	t.Parallel()
	fn := func(txn *badger.Txn) error {
		smt := NewSMT(nil, Hasher, prefixFn)
		keys := GetFreshData(2, 32)
		values := GetFreshData(2, 32)
		root, err := smt.Update(txn, keys, values)
		if err != nil {
			t.Fatal(err)
		}
		_, err = smt.Commit(txn, 1)
		if err != nil {
			t.Fatal(err)
		}
		cheight, err := smt.db.getCommitHeightDB(txn)
		if err != nil {
			t.Fatal(err)
		}
		if cheight != 1 {
			t.Fatal("commit height wrong after update with commit, cheight is", cheight)
		}
		croot, err := smt.db.getRootForHeightDB(txn, 1)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(croot, root) {
			t.Fatal("commit root wrong after update commit")
		}

		vv, err := smt.Get(txn, keys[0])
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(vv, values[0]) {
			t.Fatal("key not inserted after commit")
		}

		newvalues := [][]byte{}
		newvalues = append(newvalues, DefaultLeaf)
		newkeys := [][]byte{}
		newkeys = append(newkeys, keys[0])
		root2, err := smt.Update(txn, newkeys, newvalues)
		if err != nil {
			t.Fatal(err)
		}
		_, err = smt.Commit(txn, 2)
		if err != nil {
			t.Fatal(err)
		}

		vvv, err := smt.Get(txn, keys[0])
		if err != nil {
			t.Fatal(err)
		}
		if vvv != nil {
			t.Fatal("key not deleted")
		}

		var node Hash
		copy(node[:], root2)

		cheight, err = smt.db.getCommitHeightDB(txn)
		if err != nil {
			t.Fatal(err)
		}
		if cheight != 2 {
			t.Fatal("commit height wrong after update with commit, cheight is", cheight)
		}
		croot, err = smt.db.getRootForHeightDB(txn, 2)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(croot, root2) {
			t.Fatal("commit root wrong after update commit")
		}

		keys = GetFreshData(2, 32)
		values = GetFreshData(2, 32)
		root3, err := smt.Update(txn, keys, values)
		if err != nil {
			t.Fatal(err)
		}
		_, err = smt.Commit(txn, 3)
		if err != nil {
			t.Fatal(err)
		}

		cheight, err = smt.db.getCommitHeightDB(txn)
		if err != nil {
			t.Fatal(err)
		}
		if cheight != 3 {
			t.Fatal("commit height wrong after update with commit")
		}
		croot, err = smt.db.getRootForHeightDB(txn, 3)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(croot, root3) {
			t.Fatal("commit root wrong after update commit")
		}
		_, err = smt.db.getRootForHeightDB(txn, 2)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				t.Fatal("old commit root not removed after update commit", err)
			}
		}
		return nil
	}
	testDb(t, fn)
}

func TestVerifySubtree(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	smt := NewSMT(nil, Hasher, prefixFn)
	for i := 0; i < 1024; i++ {
		keys := GetFreshData(1, 32)
		values := GetFreshData(1, 32)

		fn := func(txn *badger.Txn) error {
			// Add data to empty trie
			_, err := smt.Update(txn, keys, values)
			if err != nil {
				return err
			}
			_, err = smt.Commit(txn, 1)
			if err != nil {
				return err
			}
			return nil
		}

		err := db.Update(fn)
		if err != nil {
			t.Error(err)
		}

		var node Hash
		copy(node[:], smt.Root)

		var batch [][]byte
		err = db.View(func(txn *badger.Txn) error {
			var err error
			batch, err = smt.loadBatch(txn, smt.Root)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		//_, res := smt.verifyBatch(batch, 0, 4, 252, smt.Root, false)
		_, res := smt.verifyBatchEasy(batch, smt.Root, 0)
		//_, res := smt.verifyBatch(batch, 0, 1, 256, smt.Root, false)
		if res != true {
			t.Fatal("the sub tree verification did not succeed")
		}

		unfinished, _, ok := smt.getInteriorNodesNext(batch, 0, 4, 252, smt.Root)
		if !ok {
			t.Fatal("not ok")
		}

		for i := 0; i < len(unfinished); i++ {
			var subBatch [][]byte
			err = db.View(func(txn *badger.Txn) error {
				var err error
				subBatch, err = smt.loadBatch(txn, unfinished[i])
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}

			if len(unfinished[i]) > 0 && bytes.Equal(subBatch[0], []byte{0}) {
				_, res := smt.verifyBatchEasy(subBatch, unfinished[i], 1)
				if res != true {
					t.Fatal("the sub tree verification did not succeed")
				}

			} else {
				_, res := smt.verifyBatchEasy(subBatch, unfinished[i], 1)
				if res != true {
					t.Fatal("the sub tree verification did not succeed")
				}

				tmp, _, ok := smt.getInteriorNodesNext(subBatch, 0, 4, 248, unfinished[i])
				if !ok {
					t.Fatal("not ok")
				}
				if len(tmp) > 0 {
					t.Fatal("RETURNED INTERIORNODESNEXT FOR SHORTCUT NODE")
				}
			}
		}
	}
}

func TestSmtFastSync(t *testing.T) {
	t.Parallel()
	type pending struct {
		layer int
		hash  []byte
	}
	db := environment.SetupBadgerDatabase(t)
	db2 := environment.SetupBadgerDatabase(t)

	numKeys := 64
	depKeys := 32
	randKeys := numKeys - depKeys
	smt := NewSMT(nil, Hasher, prefixFn)
	smt2 := NewSMT(nil, Hasher, prefixFn)
	loopStart := 1
	keys := [][]byte{}
	for i := loopStart; i < loopStart+depKeys; i++ {
		iBytes := utils.MarshalUint32(uint32(i))
		keys = append(keys, utils.ForceSliceToLength(iBytes, 32))
	}
	keysNext := GetFreshDataUnsorted(randKeys, 32)
	keys = append(keys, keysNext...)
	values := GetFreshDataUnsorted(numKeys, 32)
	keysSorted, valuesSorted, err := utils.SortKVs(keys, values)
	if err != nil {
		t.Fatal(err)
	}
	keys2 := [][]byte{}
	values2 := [][]byte{}
	for i := 0; i < len(keysSorted); i++ {
		keyCopyI := utils.CopySlice(keysSorted[i])
		valueCopyI := utils.CopySlice(valuesSorted[i])
		keys2 = append(keys2, keyCopyI)
		values2 = append(values2, valueCopyI)
	}

	for i := 0; i < len(keysSorted); i++ {
		err = db.Update(func(txn *badger.Txn) error {
			// Add data to empty trie
			_, err := smt.Update(txn, [][]byte{keysSorted[i]}, [][]byte{valuesSorted[i]})
			if err != nil {
				return err
			}
			_, err = smt.Commit(txn, uint32(1+i))
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			t.Error(err)
		}
	}

	leaves := []LeafNode{}
	var subBatchStart []pending
	err = db2.Update(func(txn *badger.Txn) error {

		var batch []byte
		err := db.View(func(txn2 *badger.Txn) error {
			tmp, err := GetNodeDB(txn2, prefixFn(), smt.Root)
			if err != nil {
				t.Fatal(err)
			}
			batch = utils.CopySlice(tmp)
			_, err = smt.parseBatch(tmp)
			if err != nil {
				t.Fatal(err)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		tmp, layer, lvs, err := smt2.StoreSnapShotNode(txn, batch, smt.Root, 0)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < len(tmp); i++ {
			subBatchStart = append(subBatchStart, pending{layer, tmp[i]})
		}
		for i := 0; i < len(lvs); i++ {
			leaves = append(leaves, lvs[i])
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	recur := func(subBatch []pending) []pending {
		subBatchNew := []pending{}
		for i := 0; i < len(subBatch); i++ {
			err = db2.Update(func(txn *badger.Txn) error {
				var batch []byte
				err := db.View(func(txn2 *badger.Txn) error {
					tmp, err := GetNodeDB(txn2, prefixFn(), subBatch[i].hash)
					if err != nil {
						t.Fatal(err)
					}
					_, err = smt.parseBatch(tmp)
					if err != nil {
						t.Fatal(err)
					}
					batch = utils.CopySlice(tmp)
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
				tmp, layer, lvs, err := smt2.StoreSnapShotNode(txn, batch, subBatch[i].hash, subBatch[i].layer)
				if err != nil {
					t.Fatal(err)
				}
				l := uint16(layer)
				for j := 0; j < len(tmp); j++ {
					subBatchNew = append(subBatchNew, pending{int(l), tmp[j]})
				}
				for i := 0; i < len(lvs); i++ {
					leaves = append(leaves, lvs[i])
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		}
		return subBatchNew
	}

	subBatchNext := recur(subBatchStart)
	for {
		subBatchNext = recur(subBatchNext)
		if len(subBatchNext) == 0 {
			break
		}
	}

	err = db.Update(func(txn *badger.Txn) error {
		err := smt2.FinalizeSnapShotRoot(txn, smt.Root, uint32(len(keysSorted)))
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	smt3 := NewSMT(smt.Root, Hasher, prefixFn)

	err = db2.View(func(txn *badger.Txn) error {
		batch, err := smt3.loadBatch(txn, smt.Root)
		if err != nil {
			t.Fatal(err)
		}
		_, ok := smt3.verifyBatchEasy(batch, smt.Root, 0)
		if !ok {
			t.Fatal("not okay here")
		}
		for k := 0; k < len(valuesSorted); k++ {
			fmt.Printf("Idx: %v; key: %x; value: %x\n", k, keysSorted[k], valuesSorted[k])
		}
		for i, key := range keys2 {
			value, err := smt3.Get(txn, key)
			if err != nil {
				t.Fatalf("%v, err: %v\n", i, err)
			}
			if !bytes.Equal(value, valuesSorted[i]) {
				t.Fatalf("values do not match at i = %v; valid idx: %v\n    key: %x\n    valueTrue: %x\n    valueIs:   %x", i, -1, key, values2[i], value)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(leaves) != len(keys) {
		t.Fatalf("Cardinality is different: leaves:%v  keys:%v", len(leaves), len(keys))
	}

	setmap := make(map[[32]byte][]byte)
	for i := 0; i < len(keys); i++ {
		k := [32]byte{}
		copy(k[:], keys[i])
		setmap[k] = values[i]
	}

	leafmap := make(map[[32]byte][]byte)
	for i := 0; i < len(leaves); i++ {
		k := [32]byte{}
		copy(k[:], leaves[i].Key)
		leafmap[k] = leaves[i].Value
	}

	boolmap := make(map[[32]byte]bool)
	for k, v := range leafmap {
		boolmap[k] = true
		if !bytes.Equal(v[:], leafmap[k]) {
			t.Fatal("not equal")
		}
	}

	if len(boolmap) != len(keys) {
		t.Fatal("not equal")
	}
}

func TestTrieMerkleProof(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	smt := NewSMT(nil, Hasher, prefixFn)

	err := db.Update(func(txn *badger.Txn) error {
		// Add data to empty trie
		keys := GetFreshData(10, 32)
		values := GetFreshData(10, 32)
		_, err := smt.Update(txn, keys, values)
		if err != nil {
			t.Fatal(err)
		}
		_, err = smt.Commit(txn, 1)
		if err != nil {
			t.Fatal(err)
		}
		for i, key := range keys {
			ap, _, k, v, _ := smt.MerkleProof(txn, key)
			if !smt.VerifyInclusion(ap, key, values[i]) {
				t.Fatalf("failed to verify inclusion proof")
			}
			if !bytes.Equal(key, k) && !bytes.Equal(values[i], v) {
				t.Fatalf("merkle proof didnt return the correct key-value pair")
			}
		}
		emptyKey := Hasher([]byte("non-member"))
		ap, included, proofKey, proofValue, _ := smt.MerkleProof(txn, emptyKey)
		if included {
			t.Fatalf("failed to verify non inclusion proof")
		}
		if !smt.VerifyNonInclusion(ap, emptyKey, proofValue, proofKey) {
			t.Fatalf("failed to verify non inclusion proof")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTrieMerkleProofCompressed(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	smt := NewSMT(nil, Hasher, prefixFn)
	err := db.Update(func(txn *badger.Txn) error {
		// Add data to empty trie
		keys := GetFreshData(10, 32)
		values := GetFreshData(10, 32)
		_, err := smt.Update(txn, keys, values)
		if err != nil {
			t.Fatal(err)
		}
		_, err = smt.Commit(txn, 1)
		if err != nil {
			t.Fatal(err)
		}

		for i, key := range keys {
			bitmap, ap, length, _, k, v, _ := smt.MerkleProofCompressed(txn, key)
			if !smt.VerifyInclusionC(bitmap, key, values[i], ap, length) {
				t.Fatalf("failed to verify inclusion proof")
			}
			if !bytes.Equal(key, k) && !bytes.Equal(values[i], v) {
				t.Fatalf("merkle proof didnt return the correct key-value pair")
			}
		}
		emptyKey := Hasher([]byte("non-member"))
		bitmap, ap, length, included, proofKey, proofValue, _ := smt.MerkleProofCompressed(txn, emptyKey)
		if included {
			t.Fatalf("failed to verify non inclusion proof")
		}
		if !smt.VerifyNonInclusionC(ap, length, bitmap, emptyKey, proofValue, proofKey) {
			t.Fatalf("failed to verify non inclusion proof")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetFinalLeafNodes(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	smt := NewSMT(nil, Hasher, prefixFn)
	err := db.Update(func(txn *badger.Txn) error {
		// Add data to empty trie
		// keys := GetFreshData(13, 32)
		// values := GetFreshData(13, 32)

		keys := make([][]byte, 3)
		values := make([][]byte, 3)

		keys[0], _ = hex.DecodeString("053f904343f38a38a0241c3b0cfa0ece5261339ce0a274a862d767e4ac6a593a")
		keys[1], _ = hex.DecodeString("9b0213327fefa8e3a844d4394a39178bb2a9a744cac6259986cdb650d9dc3d9e")
		keys[2], _ = hex.DecodeString("fe7718c7f36546a26becb72e332d2920c5d7212a88c4b6e104895fdbf242a347")

		values[0], _ = hex.DecodeString("1f27af0d2dbc9751bc957e9d4d974ebfd922b2a27edbfcac9398b7efb4fd0f79")
		values[1], _ = hex.DecodeString("69bae8b9aa17bf8ae4ff4d9c18fcdaf6539d69beec116f0526e01527e505f800")
		values[2], _ = hex.DecodeString("d657ea5294a3028465353cba73b5b8fcc8f5e631de48240cc7d980fd43fdb612")

		_, err := smt.Update(txn, keys, values)
		if err != nil {
			t.Fatal(err)
		}
		_, err = smt.Commit(txn, 1)
		if err != nil {
			t.Fatal(err)
		}

		temp := smt.Root

		smt.loadDbMux.Lock()
		_, err = smt.db.getNodeDB(txn, temp[:constants.HashLen])
		smt.loadDbMux.Unlock()
		if err != nil {
			fmt.Println(err)
			return err
		}

		batch, err := smt.loadBatch(txn, temp)
		if err != nil {
			t.Fatal(err)
		}

		finalLeafNodes := smt.getFinalLeafNodes(batch, 0)

		if len(finalLeafNodes) != 3 {
			t.Fatal("returned an incorrect number of final leaf values")
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetFinalLeafNodesrRValues(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	smt := NewSMT(nil, Hasher, prefixFn)
	err := db.Update(func(txn *badger.Txn) error {
		// Add data to empty trie
		keys := GetFreshData(13, 32)
		values := GetFreshData(13, 32)

		_, err := smt.Update(txn, keys, values)
		if err != nil {
			t.Fatal(err)
		}
		_, err = smt.Commit(txn, 1)
		if err != nil {
			t.Fatal(err)
		}

		temp := smt.Root

		smt.loadDbMux.Lock()
		_, err = smt.db.getNodeDB(txn, temp[:constants.HashLen])
		smt.loadDbMux.Unlock()
		if err != nil {
			fmt.Println(err)
			return err
		}

		batch, err := smt.loadBatch(txn, temp)
		if err != nil {
			t.Fatal(err)
		}

		for i := 0; i < len(batch); i++ {
			fmt.Printf("%v %x\n", i, batch[i])
		}

		finalLeafNodes := smt.getFinalLeafNodes(batch, 0)

		fmt.Println("number of final leaves is", len(finalLeafNodes))

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// This test is only looking for panics
func TestParseBatchNoPanic(t *testing.T) {
	t.Parallel()
	smt := NewSMT(nil, Hasher, prefixFn)
	for i := 0; i < 100; i++ {
		for j := 0; j < 2000; j++ {
			temp := i % 32
			var data []byte
			if i > 32 {
				for k := 0; k < (i / 32); k++ {
					data = append(data, GetFreshData(1, 32)[0]...)
				}
				data = append(data, GetFreshData(1, temp)[0]...)
			} else {
				data = GetFreshData(1, i)[0]
			}
			smt.parseBatch(data) //nolint:errcheck
		}
	}
}

func TestGetInteriorNodesNext(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	smt := NewSMT(nil, Hasher, prefixFn)
	err := db.Update(func(txn *badger.Txn) error {
		// Add data to empty trie
		keys := GetFreshData(13, 32)
		values := GetFreshData(13, 32)

		_, err := smt.Update(txn, keys, values)
		if err != nil {
			t.Fatal(err)
		}
		_, err = smt.Commit(txn, 1)
		if err != nil {
			t.Fatal(err)
		}

		temp := smt.Root

		smt.loadDbMux.Lock()
		_, err = smt.db.getNodeDB(txn, temp[:constants.HashLen])
		smt.loadDbMux.Unlock()
		if err != nil {
			fmt.Println(err)
			return err
		}

		batch, err := smt.loadBatch(txn, temp)
		if err != nil {
			t.Fatal(err)
		}

		for i := 0; i < len(batch); i++ {
			fmt.Printf("%v %x\n", i, batch[i])
		}

		unfinishedNodes, _, _ := smt.getInteriorNodesNext(batch, 0, 4, 0, temp)

		fmt.Println("number of unfinished interior nodes is", len(unfinishedNodes))

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSmtCommit(t *testing.T) {
	t.Parallel()
	smt := NewSMT(nil, Hasher, prefixFn)
	keys := GetFreshData(32, 32)
	values := GetFreshData(32, 32)
	fn := func(txn *badger.Txn) error {
		_, err := smt.Update(txn, keys, values)
		if err != nil {
			t.Fatal(err)
		}
		_, err = smt.Commit(txn, 1)
		if err != nil {
			t.Fatal(err)
		}
		for i := range keys {
			value, _ := smt.Get(txn, keys[i])
			if !bytes.Equal(values[i], value) {
				t.Fatal("failed to get value in committed db")
			}
		}

		// test loading a shortcut batch
		smt = NewSMT(nil, Hasher, prefixFn)
		keys = GetFreshData(1, 32)
		values = GetFreshData(1, 32)
		_, err = smt.Update(txn, keys, values)
		if err != nil {
			t.Fatal(err)
		}
		_, err = smt.Commit(txn, 1)
		if err != nil {
			t.Fatal(err)
		}
		value, _ := smt.Get(txn, keys[0])
		if !bytes.Equal(values[0], value) {
			t.Fatal("failed to get value in committed db")
		}
		return nil
	}
	testDb(t, fn)
}

func TestDoubleUpdate(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)
	err := db.Update(func(outerTxn *badger.Txn) error {
		err := outerTxn.Set([]byte("b"), []byte("outer"))
		if err != nil {
			return err
		}
		err = outerTxn.Set([]byte("a"), []byte("outer"))
		if err != nil {
			return err
		}
		_, err = outerTxn.Get([]byte("b"))
		if err != nil {
			return err
		}
		return db.Update(func(innerTxn *badger.Txn) error {
			err := innerTxn.Set([]byte("c"), []byte("inner"))
			if err != nil {
				return err
			}
			return nil
		})
	})
	if err != nil {
		t.Error(err)
	}
	err = db.View(func(txn *badger.Txn) error {
		value, err := utils.GetValue(txn, []byte("a"))
		if err != nil {
			return err
		}
		if !bytes.Equal(value, []byte("outer")) {
			t.Error("a not equal to inner")
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	err = db.View(func(txn *badger.Txn) error {
		value, err := utils.GetValue(txn, []byte("c"))
		if err != nil {
			return err
		}
		if !bytes.Equal(value, []byte("inner")) {
			t.Error("a not equal to inner")
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

func TestSmtRaisesError(t *testing.T) {
	t.Parallel()
	smt := NewSMT(nil, Hasher, prefixFn)
	// Add data to empty trie
	keys := GetFreshData(10, 32)
	values := GetFreshData(10, 32)
	fn := func(txn *badger.Txn) error {
		_, err := smt.Update(txn, keys, values)
		if err != nil {
			t.Fatal(err)
		}
		smt.db.updatedNodes = make(map[Hash][][]byte)
		// Check errors are raised is a keys is not in cache nor db
		for _, key := range keys {
			_, err := smt.Get(txn, key)
			if err == nil {
				t.Fatal("Error not created if database doesnt have a node")
			}
		}
		_, err = smt.Update(txn, keys, values)
		if err == nil {
			t.Fatal("Error not created if database doesnt have a node")
		}
		return nil
	}
	testDb(t, fn)
}

func TestDiscard(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)
	var rootTest []byte
	smt := NewSMT(nil, Hasher, prefixFn)
	keys := GetFreshData(20, 32)
	fn1 := func(txn *badger.Txn) error {
		// Add data to empty trie
		values := GetFreshData(20, 32)
		root, _ := smt.Update(txn, keys, values)
		rootTest = root
		_, err := smt.Commit(txn, 1)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	}
	fn2 := func(txn *badger.Txn) error {
		keys = GetFreshData(20, 32)
		values := GetFreshData(20, 32)
		_, err := smt.Update(txn, keys, values)
		if err != nil {
			t.Fatal(err)
		}
		smt.Discard()
		return nil
	}
	fn3 := func(txn *badger.Txn) error {
		keys = GetFreshData(20, 32)
		values := GetFreshData(20, 32)
		root, err := smt.Update(txn, keys, values)
		if err != nil {
			t.Error(err)
		}
		rootTest = root
		_, err = smt.Commit(txn, 1)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	}
	fn4 := func(txn *badger.Txn) error {
		keys := GetFreshData(20, 32)
		values := GetFreshData(20, 32)
		root, err := smt.Update(txn, keys, values)
		if err != nil {
			t.Error(err)
		}
		rootTest = root
		smt.Discard()
		return nil
	}
	err := db.Update(fn1)
	if err != nil {
		t.Error(err)
	}
	err = db.Update(fn2)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(smt.Root, rootTest) {
		t.Fatal("Trie not rolled back")
	}
	if len(smt.db.updatedNodes) != 0 {
		t.Fatal("Trie not rolled back")
	}
	err = db.Update(fn3)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(smt.Root, rootTest) {
		t.Fatal("Trie not rolled back")
	}
	if len(smt.db.updatedNodes) != 0 {
		t.Fatal("Trie not rolled back")
	}
	err = db.Update(fn4)
	if err != nil {
		t.Error(err)
	}
}

func TestWalkers(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)
	smt := NewSMT(nil, Hasher, prefixFn)
	for i := 0; i < 30; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			// Add data to empty trie
			keysInitial := GetFreshData(10, 32)
			valuesInitial := GetFreshData(10, 32)
			_, err := smt.Update(txn, keysInitial, valuesInitial)
			if err != nil {
				return err
			}
			_, err = smt.Commit(txn, uint32(i+1))
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		nodeCounter := make(map[Hash]bool)
		iteratorCount := 0
		err = db.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchSize = 0
			prefix := []byte{}
			prefix = append(prefix, smt.db.prefixFunc()...)
			prefix = append(prefix, prefixNode()...)
			opts.Prefix = prefix
			opts.PrefetchValues = false
			opts.AllVersions = false
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				iteratorCount++
			}
			cb := func(d []byte, v []byte) error {
				var node Hash
				copy(node[:], d)
				if nodeCounter[node] {
					return nil
				}
				nodeCounter[node] = true
				return nil
			}
			for j := 0; j <= i+1; j++ {
				oldRoot, err := smt.db.getRootForHeightDB(txn, uint32(j))
				if err != nil {
					if err != badger.ErrKeyNotFound {
						return err
					}
				}
				if len(oldRoot) > 0 {
					err := smt.walkNodes(txn, oldRoot, cb)
					if err != nil {
						return err
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(nodeCounter) != iteratorCount {
			fmt.Println("node counts not equal", i, len(nodeCounter), iteratorCount)
			t.Error("node counts not equal", i, len(nodeCounter), iteratorCount)
		}
	}
}

func TestBigDelete(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	smt := NewSMT(nil, Hasher, prefixFn)

	for i := 0; i < 50; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			keys := GetFreshData(12, 32)
			values := GetFreshData(12, 32)
			_, err := smt.Update(txn, keys, values)
			if err != nil {
				t.Fatal(err)
			}
			_, err = smt.Commit(txn, 1)
			if err != nil {
				t.Fatal(err)
			}
			deletes := make([][]byte, 2)
			for i := 0; i < 2; i++ {
				deletes[i] = DefaultLeaf
			}
			_, err = smt.Update(txn, keys[:2], deletes)
			if err != nil {
				t.Fatal(err)
			}
			_, err = smt.Commit(txn, 1)
			if err != nil {
				t.Fatal(err)
			}
			for j := 0; j < 2; j++ {
				v, err := smt.Get(txn, keys[j])
				if err != nil {
					t.Fatal(err)
				}
				if len(v) > 0 {
					t.Fatal("deleted key still present")
				}
			}
			for j := 2; j < 12; j++ {
				v, err := smt.Get(txn, keys[j])
				if err != nil {
					t.Fatal(err)
				}
				if len(v) < 32 {
					t.Fatal("non-deleted key not present")
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestSnapShotDrop(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)
	smt := NewSMT(nil, Hasher, prefixFn)
	numKeys := 3000
	keys := GetFreshData(numKeys, 32)
	values := GetFreshData(numKeys, 32)
	err := db.Update(func(txn *badger.Txn) error {
		var err error
		_, err = smt.Update(txn, keys, values)
		if err != nil {
			t.Fatal(err)
		}
		_, err = smt.Commit(txn, 1)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	newPrefix := func() []byte {
		return []byte("!!")
	}
	var newSmt *SMT
	err = db.Update(func(txn *badger.Txn) error {
		var err error
		newSmt, err = smt.SnapShot(txn, newPrefix)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		for i, key := range keys {
			v, err := smt.Get(txn, key)
			if err != nil {
				t.Fatal(err)
			}
			if len(v) == 0 {
				t.Fatal("snapshot missing key")
			}
			if !bytes.Equal(v, values[i]) {
				t.Fatal("snapshot has bad value")
			}
		}
		for i, key := range keys {
			v, err := newSmt.Get(txn, key)
			if err != nil {
				t.Fatal(err)
			}
			if len(v) == 0 {
				t.Fatal("snapshot missing key")
			}
			if !bytes.Equal(v, values[i]) {
				t.Fatal("snapshot has bad value")
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = newSmt.Drop(db)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		for i, key := range keys {
			v, err := smt.Get(txn, key)
			if err != nil {
				t.Fatal(err)
			}
			if len(v) == 0 {
				t.Fatal("snapshot missing key")
			}
			if !bytes.Equal(v, values[i]) {
				t.Fatal("snapshot has bad value")
			}
		}
		for _, key := range keys {
			_, err := newSmt.Get(txn, key)
			if err == nil {
				t.Fatal(err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
