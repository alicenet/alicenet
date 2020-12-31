package indexer

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/MadBase/MadNet/crypto"
	"github.com/dgraph-io/badger/v2"
)

func makeInsertationOrderIndexer() *InsertionOrderIndexer {
	prefix1 := func() []byte {
		return []byte("zg")
	}
	prefix2 := func() []byte {
		return []byte("zh")
	}
	index := NewInsertionOrderIndex(prefix1, prefix2)
	return index
}

func TestInsertationOrderIndexerAdd(t *testing.T) {
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

	index := makeInsertationOrderIndexer()
	txHash := crypto.Hasher([]byte("txHash"))

	err = db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, txHash)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestInsertationOrderIndexerDelete(t *testing.T) {
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

	index := makeInsertationOrderIndexer()
	txHash := crypto.Hasher([]byte("txHash"))

	err = db.Update(func(txn *badger.Txn) error {
		err := index.Delete(txn, txHash)
		if err == nil {
			t.Fatal("Should have raised error")
		}
		err = index.Add(txn, txHash)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Delete(txn, txHash)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestInsertationOrderIndexerMakeIndexKeys(t *testing.T) {
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

	index := makeInsertationOrderIndexer()
	txHash := crypto.Hasher([]byte("txHash"))
	irkTrue := make([]byte, 0)
	irkTrue = append(irkTrue, index.revPrefix()...)
	irkTrue = append(irkTrue, txHash...)

	_, ioiRevIdxKey, err := index.makeIndexKeys(txHash)
	if err != nil {
		t.Fatal(err)
	}
	revIdxKey := ioiRevIdxKey.MarshalBinary()
	if !bytes.Equal(irkTrue, revIdxKey) {
		t.Fatal("revIdxKeys do not match!")
	}
}
