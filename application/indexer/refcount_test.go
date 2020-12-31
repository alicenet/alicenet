package indexer

import (
	"io/ioutil"
	"os"
	"testing"

	trie "github.com/MadBase/MadNet/badgerTrie"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
)

func makeRefCounter() *RefCounter {
	prefix := func() []byte {
		return []byte("rr")
	}
	rc := NewRefCounter(prefix)
	return rc
}

func TestRefCounterIncrement(t *testing.T) {
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

	txHash := trie.Hasher([]byte("txHash"))
	rc := makeRefCounter()
	num := 100
	err = db.Update(func(txn *badger.Txn) error {
		for i := 1; i < num; i++ {
			v, err := rc.Increment(txn, txHash)
			if err != nil {
				t.Error(err)
			}
			if v != int64(i) {
				t.Error("bad count after increment", v)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRefCounterDecrement(t *testing.T) {
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

	txHash := trie.Hasher([]byte("txHash"))
	rc := makeRefCounter()
	err = db.Update(func(txn *badger.Txn) error {
		// Attempting to decrement a counter already set to zero;
		// no error is raised because no error is expected.
		_, err := rc.Decrement(txn, txHash)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	num := 100
	err = db.Update(func(txn *badger.Txn) error {
		for i := 1; i < num; i++ {
			v, err := rc.Increment(txn, txHash)
			if err != nil {
				t.Error(err)
			}
			if v != int64(i) {
				t.Error("bad count after increment", v)
			}
		}
		for i := 1; i < (num - 1); i++ {
			v, err := rc.Decrement(txn, txHash)
			if err != nil {
				t.Error(err)
			}
			if v != int64(num-1-i) {
				t.Error("bad count after increment", v)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		v, err := rc.Decrement(txn, txHash)
		if err != nil {
			t.Error(err)
		}
		if v != 0 {
			t.Error("bad count after increment", v)
		}
		rcKey := append(rc.prefix(), txHash...)
		_, err = utils.GetValue(txn, rcKey)
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
