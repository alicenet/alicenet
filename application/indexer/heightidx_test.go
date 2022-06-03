package indexer

import (
	"bytes"
	"testing"

	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/internal/testing/environment"
	"github.com/dgraph-io/badger/v2"
)

func makeHeightIdxIndex() *HeightIdxIndex {
	prefix1 := func() []byte {
		return []byte("ze")
	}
	prefix2 := func() []byte {
		return []byte("zf")
	}
	index := NewHeightIdxIndex(prefix1, prefix2)
	return index
}

func TestHeightIdxIndexAdd(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	index := makeHeightIdxIndex()
	txHash := crypto.Hasher([]byte("utxoID"))
	height := uint32(1234)
	idx := uint32(25519)

	err := db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, txHash, height, idx)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestHeightIdxIndexDelete(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	index := makeHeightIdxIndex()
	txHash := crypto.Hasher([]byte("utxoID"))
	height := uint32(1234)
	idx := uint32(25519)

	err := db.Update(func(txn *badger.Txn) error {
		err := index.Delete(txn, txHash)
		if err == nil {
			t.Fatal("Should raise an error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, txHash, height, idx)
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

func TestHeightIdxIndexGetHeightIdx(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	index := makeHeightIdxIndex()
	txHash := crypto.Hasher([]byte("utxoID"))
	height := uint32(1234)
	idx := uint32(25519)

	err := db.Update(func(txn *badger.Txn) error {
		_, _, err := index.GetHeightIdx(txn, txHash)
		if err == nil {
			// txHash does not exist
			t.Fatal("Should have raised error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, txHash, height, idx)
		if err != nil {
			t.Fatal(err)
		}
		returnedHeight, returnedIdx, err := index.GetHeightIdx(txn, txHash)
		if err != nil {
			t.Fatal(err)
		}
		if height != returnedHeight {
			t.Fatal("heights to do not agree")
		}
		if idx != returnedIdx {
			t.Fatal("idxs to do not agree")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestHeightIdxIndexGetTxHashFromHeightIdx(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	index := makeHeightIdxIndex()
	txHash := crypto.Hasher([]byte("utxoID"))
	height := uint32(1234)
	idx := uint32(25519)

	err := db.Update(func(txn *badger.Txn) error {
		_, err := index.GetTxHashFromHeightIdx(txn, height, idx)
		if err == nil {
			// txHash does not exist
			t.Fatal("Should have raised error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, txHash, height, idx)
		if err != nil {
			t.Fatal(err)
		}
		returnedTxHash, err := index.GetTxHashFromHeightIdx(txn, height, idx)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(txHash, returnedTxHash) {
			t.Fatal("txHashs do not agree")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestHeightIdxIndexgetHeightIdx(t *testing.T) {
	t.Parallel()

	index := makeHeightIdxIndex()
	//txHash := crypto.Hasher([]byte("utxoID"))
	heightTrue := uint32(1234)
	idxTrue := uint32(25519)

	heightIdx := index.makeHeightIdx(heightTrue, idxTrue)
	height, idx, err := index.getHeightIdx(heightIdx)
	if err != nil {
		t.Fatal(err)
	}
	if height != heightTrue {
		t.Fatal("Error in height")
	}
	if idx != idxTrue {
		t.Fatal("Error in idx")
	}

	badHeightIdx1 := make([]byte, 7)
	_, _, err = index.getHeightIdx(badHeightIdx1)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	badHeightIdx2 := make([]byte, 9)
	_, _, err = index.getHeightIdx(badHeightIdx2)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}
