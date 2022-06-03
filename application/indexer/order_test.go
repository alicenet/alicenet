package indexer

import (
	"bytes"
	"testing"

	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/internal/testing/environment"
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
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	index := makeInsertationOrderIndexer()
	txHash := crypto.Hasher([]byte("txHash"))

	err := db.Update(func(txn *badger.Txn) error {
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
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	index := makeInsertationOrderIndexer()
	txHash := crypto.Hasher([]byte("txHash"))

	err := db.Update(func(txn *badger.Txn) error {
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
	t.Parallel()

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
