package indexer

import (
	"bytes"
	"testing"

	"github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/internal/testing/environment"
	"github.com/dgraph-io/badger/v2"
)

func makeValueIndex() *ValueIndex {
	prefix1 := func() []byte {
		return []byte("ya")
	}
	prefix2 := func() []byte {
		return []byte("yb")
	}
	index := NewValueIndex(prefix1, prefix2)
	return index
}

func TestValueIndexAdd(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	index := makeValueIndex()
	owner := &objs.Owner{}
	utxoID := crypto.Hasher([]byte("utxoID"))
	value, err := new(uint256.Uint256).FromUint64(25519)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, utxoID, owner, value)
		if err == nil {
			// Invalid Owner
			t.Fatal("Should have raised error (1)")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	owner = makeOwner()
	err = db.Update(func(txn *badger.Txn) error {
		err = index.Add(txn, utxoID, owner, value)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestValueIndexDrop(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	index := makeValueIndex()
	owner := makeOwner()
	utxoID := crypto.Hasher([]byte("utxoID"))
	value, err := new(uint256.Uint256).FromUint64(25519)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
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
	if err != nil {
		t.Fatal(err)
	}
}

func TestValueIndexMakeKey(t *testing.T) {
	t.Parallel()

	index := makeValueIndex()
	owner := &objs.Owner{}
	utxoID := crypto.Hasher([]byte("utxoID"))
	value, err := new(uint256.Uint256).FromUint64(25519)
	if err != nil {
		t.Fatal(err)
	}

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
	valueBytes, err := value.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	trueKey = append(trueKey, valueBytes...)
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

func TestValueIndexMakeRefKey(t *testing.T) {
	t.Parallel()

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
