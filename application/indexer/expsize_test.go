package indexer

import (
	"bytes"
	"testing"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/internal/testing/environment"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
)

func makeExpSizeIndex() *ExpSizeIndex {
	prefix1 := func() []byte {
		return []byte("zc")
	}
	prefix2 := func() []byte {
		return []byte("zd")
	}
	index := NewExpSizeIndex(prefix1, prefix2)
	return index
}

func TestExpSizeIndexAdd(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	index := makeExpSizeIndex()
	epoch := uint32(1)
	utxoID := crypto.Hasher([]byte("utxoID"))
	size := uint32(25519)
	err := db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, epoch, utxoID, size)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestExpSizeIndexDrop(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	index := makeExpSizeIndex()
	epoch := uint32(1)
	utxoID := crypto.Hasher([]byte("utxoID"))
	size := uint32(25519)

	err := db.Update(func(txn *badger.Txn) error {
		err := index.Drop(txn, utxoID)
		if err == nil {
			t.Fatal("Should have raised error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, epoch, utxoID, size)
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

func TestExpSizeIndexMakeKey(t *testing.T) {
	t.Parallel()

	index := makeExpSizeIndex()
	epoch := uint32(1)
	utxoID := crypto.Hasher([]byte("utxoID"))
	size := uint32(25519)
	trueKey := []byte{}
	trueKey = append(trueKey, index.prefix()...)
	trueKey = append(trueKey, utils.MarshalUint32(epoch)...)
	trueKey = append(trueKey, utils.MarshalUint32(constants.MaxUint32-size)...)
	trueKey = append(trueKey, utxoID...)
	esiKey := index.makeKey(epoch, size, utxoID)
	key := esiKey.MarshalBinary()
	if !bytes.Equal(key, trueKey) {
		t.Fatal("keys do not match")
	}
}

func TestExpSizeIndexMakeRefKey(t *testing.T) {
	t.Parallel()

	index := makeExpSizeIndex()
	utxoID := crypto.Hasher([]byte("utxoID"))
	trueKey := []byte{}
	trueKey = append(trueKey, index.refPrefix()...)
	trueKey = append(trueKey, utxoID...)
	esiKey := index.makeRefKey(utxoID)
	key := esiKey.MarshalBinary()
	if !bytes.Equal(key, trueKey) {
		t.Fatal("keys do not match")
	}
}

func TestExpSizeIndexKey(t *testing.T) {
	esiKey := &ExpSizeIndexKey{}
	keyTrue := crypto.Hasher([]byte("key"))
	esiKey.key = keyTrue
	key := esiKey.MarshalBinary()
	if !bytes.Equal(key, keyTrue) {
		t.Fatal("keys do not match (1)")
	}
	esiKey2 := &ExpSizeIndexKey{}
	esiKey2.UnmarshalBinary(key)
	if !bytes.Equal(esiKey.key, esiKey2.key) {
		t.Fatal("keys do not match (2)")
	}

	esiRefkey := &ExpSizeIndexRefKey{}
	refkeyTrue := crypto.Hasher([]byte("refkey"))
	esiRefkey.refkey = refkeyTrue
	refkey := esiRefkey.MarshalBinary()
	if !bytes.Equal(refkey, refkeyTrue) {
		t.Fatal("refkeys do not match (1)")
	}
	esiRefkey2 := &ExpSizeIndexRefKey{}
	esiRefkey2.UnmarshalBinary(refkey)
	if !bytes.Equal(esiRefkey.refkey, esiRefkey2.refkey) {
		t.Fatal("refkeys do not match (2)")
	}
}
