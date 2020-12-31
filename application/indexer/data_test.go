package indexer

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/MadBase/MadNet/application/objs"
	trie "github.com/MadBase/MadNet/badgerTrie"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/dgraph-io/badger/v2"
)

func makeOwner() *objs.Owner {
	owner := &objs.Owner{}
	acct := make([]byte, constants.OwnerLen)
	acct[0] = 255
	acct[constants.OwnerLen-1] = 255
	curveSpec := constants.CurveSecp256k1
	err := owner.New(acct, curveSpec)
	if err != nil {
		panic(err)
	}
	return owner
}

func makeDataIndex() *DataIndex {
	prefix1 := func() []byte {
		return []byte("za")
	}
	prefix2 := func() []byte {
		return []byte("zb")
	}
	index := NewDataIndex(prefix1, prefix2)
	return index
}

func TestDataIndexAdd(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	index := makeDataIndex()
	owner := &objs.Owner{}
	utxoID := crypto.Hasher([]byte("utxoID"))
	dataIndex := trie.Hasher([]byte("dataIndex"))

	err = db.Update(func(txn *badger.Txn) error {
		// Add value
		err := index.Add(txn, utxoID, owner, dataIndex)
		if err == nil {
			t.Fatal("Should raise an error (1)")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	owner = makeOwner()
	err = db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, utxoID, owner, dataIndex)
		if err != nil {
			t.Fatal(err)
		}
		val, err := index.Contains(txn, owner, dataIndex)
		if err != nil {
			t.Fatal(err)
		}
		if !val {
			// Value should be present
			t.Fatal("Should be present")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, utxoID, owner, dataIndex)
		if err == nil {
			t.Fatal("Should raise an error (2)")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDataIndexContains(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	index := makeDataIndex()
	owner := &objs.Owner{}
	utxoID := crypto.Hasher([]byte("utxoID"))
	dataIndex := trie.Hasher([]byte("dataIndex"))

	err = db.Update(func(txn *badger.Txn) error {
		_, err := index.Contains(txn, owner, utxoID)
		if err == nil {
			t.Fatal("Should have raised error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	owner = makeOwner()
	err = db.Update(func(txn *badger.Txn) error {
		present, err := index.Contains(txn, owner, utxoID)
		if err != nil {
			t.Fatal(err)
		}
		if present {
			// Value should not be present
			t.Fatal("Should not be present (1)")
		}
		// Add value
		err = index.Add(txn, utxoID, owner, dataIndex)
		if err != nil {
			t.Fatal(err)
		}
		present, err = index.Contains(txn, owner, dataIndex)
		if err != nil {
			t.Fatal(err)
		}
		if !present {
			// Value should be present
			t.Fatal("Should be present")
		}
		// Drop Value
		err = index.Drop(txn, utxoID)
		if err != nil {
			t.Fatal(err)
		}
		present, err = index.Contains(txn, owner, dataIndex)
		if err != nil {
			t.Fatal(err)
		}
		if present {
			// Value should not be present
			t.Fatal("Should not be present (2)")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDataIndexDrop(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	index := makeDataIndex()
	owner := makeOwner()
	utxoID := crypto.Hasher([]byte("utxoID"))
	dataIndex := trie.Hasher([]byte("dataIndex"))

	err = db.Update(func(txn *badger.Txn) error {
		err := index.Drop(txn, utxoID)
		if err == nil {
			// Value not present
			t.Fatal("Should have raised error")
		}
		// Add value
		err = index.Add(txn, utxoID, owner, dataIndex)
		if err != nil {
			t.Fatal(err)
		}
		// Drop value
		err = index.Drop(txn, utxoID)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDataIndexGetUTXOID(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	index := makeDataIndex()
	owner := &objs.Owner{}
	dataIndex := make([]byte, constants.HashLen)
	dataIndex[0] = 1
	dataIndex[constants.HashLen-1] = 1
	err = db.Update(func(txn *badger.Txn) error {
		_, err := index.GetUTXOID(txn, owner, dataIndex)
		if err == nil {
			// Invalid owner
			t.Fatal("Should have raised error (1)")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	owner = makeOwner()
	err = db.Update(func(txn *badger.Txn) error {
		_, err := index.GetUTXOID(txn, owner, dataIndex)
		if err == nil {
			// Invalid value not present
			t.Fatal("Should have raised error (2)")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDataIndexMakeIterKey(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	index := makeDataIndex()
	owner := &objs.Owner{}
	_, err = index.makeIterKey(owner)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	owner = makeOwner()
	ownerBytes, err := owner.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	trueIterKey := make([]byte, 0)
	trueIterKey = append(trueIterKey, index.prefix()...)
	trueIterKey = append(trueIterKey, ownerBytes...)
	iterKey, err := index.makeIterKey(owner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(iterKey, trueIterKey) {
		t.Fatal("iterKey does not match!")
	}
}

func TestDataIndexMakeKey(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	index := makeDataIndex()
	owner := &objs.Owner{}
	dataIndex := make([]byte, constants.HashLen)
	dataIndex[0] = 1
	dataIndex[constants.HashLen-1] = 1
	_, err = index.makeKey(owner, dataIndex)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	owner = makeOwner()
	ownerBytes, err := owner.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	trueKey := make([]byte, 0)
	trueKey = append(trueKey, index.prefix()...)
	trueKey = append(trueKey, ownerBytes...)
	trueKey = append(trueKey, dataIndex...)
	diKey, err := index.makeKey(owner, dataIndex)
	if err != nil {
		t.Fatal(err)
	}
	key := diKey.MarshalBinary()
	if !bytes.Equal(key, trueKey) {
		t.Fatal("key does not match!")
	}
}

func TestDataIndexMakeRefKey(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	index := makeDataIndex()
	utxoID := make([]byte, constants.HashLen)
	utxoID[0] = 255
	utxoID[constants.HashLen-1] = 255

	trueRefKey := make([]byte, 0)
	trueRefKey = append(trueRefKey, index.refPrefix()...)
	trueRefKey = append(trueRefKey, utxoID...)

	diRefKey := index.makeRefKey(utxoID)
	refKey := diRefKey.MarshalBinary()
	if !bytes.Equal(refKey, trueRefKey) {
		t.Fatal("refKey does not match!")
	}
}

func TestDataIndexKey(t *testing.T) {
	/* TODO FIX
	  diKey := &DataIndexKey{}
		keyTrue := crypto.Hasher([]byte("key"))
		diKey.key = keyTrue
		key := diKey.MarshalBinary()
		if !bytes.Equal(key, keyTrue) {
			t.Fatal("keys do not match (1)")
		}
		diKey2 := &DataIndexKey{}
		diKey2.UnmarshalBinary(key)
		if !bytes.Equal(diKey.key, diKey2.key) {
			t.Fatal("keys do not match (2)")
		}

		diRefkey := &DataIndexRefKey{}
		refkeyTrue := crypto.Hasher([]byte("refkey"))
		diRefkey.refkey = refkeyTrue
		refkey := diRefkey.MarshalBinary()
		if !bytes.Equal(refkey, refkeyTrue) {
			t.Fatal("refkeys do not match (1)")
		}
		diRefkey2 := &DataIndexRefKey{}
		diRefkey2.UnmarshalBinary(refkey)
		if !bytes.Equal(diRefkey.refkey, diRefkey2.refkey) {
			t.Fatal("refkeys do not match (2)")
		}
	*/
}
