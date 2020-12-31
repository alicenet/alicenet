package indexer

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
)

func makeValueIndexOld() *valueIndexOld {
	prefix1 := func() []byte {
		return []byte("ya")
	}
	prefix2 := func() []byte {
		return []byte("yb")
	}
	index := newValueIndexOld(prefix1, prefix2)
	return index
}

func TestValueOldIndexAdd(t *testing.T) {
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

	index := makeValueIndexOld()
	owner := &objs.Owner{}
	utxoID := crypto.Hasher([]byte("utxoID"))
	value := uint32(25519)

	db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, utxoID, owner, value)
		if err == nil {
			// Invalid Owner
			t.Fatal("Should have raised error (1)")
		}
		return nil
	})
	owner = makeOwner()
	db.Update(func(txn *badger.Txn) error {
		err = index.Add(txn, utxoID, owner, value)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
}

func TestValueOldIndexDrop(t *testing.T) {
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

	index := makeValueIndexOld()
	owner := makeOwner()
	utxoID := crypto.Hasher([]byte("utxoID"))
	value := uint32(25519)

	db.Update(func(txn *badger.Txn) error {
		err := index.Drop(txn, utxoID)
		if err == nil {
			// Not present
			t.Fatal("Should have raised error")
		}
		err = index.Add(txn, utxoID, owner, value)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Drop(txn, utxoID)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
}

func TestValueOldIndexMakeKey(t *testing.T) {
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

	index := makeValueIndexOld()
	owner := &objs.Owner{}
	utxoID := crypto.Hasher([]byte("utxoID"))
	value := uint32(25519)
	_, err = index.makeKey(owner, value, utxoID)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	owner = makeOwner()
	ownerBytes, err := owner.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	trueKey := []byte{}
	trueKey = append(trueKey, index.prefix()...)
	trueKey = append(trueKey, ownerBytes...)
	trueKey = append(trueKey, utils.MarshalUint32(value)...)
	trueKey = append(trueKey, utxoID...)
	viKey, err := index.makeKey(owner, value, utxoID)
	if err != nil {
		t.Fatal(err)
	}
	key := viKey.MarshalBinary()
	if !bytes.Equal(key, trueKey) {
		t.Fatal("keys do not agree")
	}
}

func TestValueOldIndexMakeRefKey(t *testing.T) {
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

	index := makeValueIndex()
	utxoID := crypto.Hasher([]byte("utxoID"))
	trueKey := []byte{}
	trueKey = append(trueKey, index.refPrefix()...)
	trueKey = append(trueKey, utxoID...)
	viKey := index.makeRefKey(utxoID)
	key := viKey.MarshalBinary()
	if !bytes.Equal(key, trueKey) {
		t.Fatal("keys do not agree")
	}
}
