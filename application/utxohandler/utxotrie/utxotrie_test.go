package utxotrie

/*
import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/alicenet/alicenet/crypto"
	"github.com/dgraph-io/badger/v2"
)

func TestUTXOTrie(t *testing.T) {
	height := uint32(1)
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
	//////////////////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////
	hndlr := NewUTXOTrie(db)

	hndlr.Init(1)
	ok, err := hndlr.Contains(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("did not fail")
	}
	db.Update(func(txn *badger.Txn) error {
		utxoID1 := crypto.Hasher([]byte("utxoID1"))
		utxoHash1 := crypto.Hasher([]byte("utxoHash1"))

		utxoID2 := crypto.Hasher([]byte("utxoID2"))
		utxoHash2 := crypto.Hasher([]byte("utxoHash2"))

		utxoID3 := crypto.Hasher([]byte("utxoID3"))
		utxoHash3 := crypto.Hasher([]byte("utxoHash3"))

		utxoID4 := crypto.Hasher([]byte("utxoID4"))
		utxoHash4 := crypto.Hasher([]byte("utxoHash4"))

		utxoID5 := crypto.Hasher([]byte("utxoID5"))
		utxoHash5 := crypto.Hasher([]byte("utxoHash5"))

		utxoID6 := crypto.Hasher([]byte("utxoID6"))
		utxoHash6 := crypto.Hasher([]byte("utxoHash6"))

		newUTXOIDs := [][]byte{}
		newUTXOIDs = append(newUTXOIDs, utxoID1)
		newUTXOIDs = append(newUTXOIDs, utxoID2)
		newUTXOIDs = append(newUTXOIDs, utxoID3)
		newUTXOHashes := [][]byte{}
		newUTXOHashes = append(newUTXOHashes, utxoHash1)
		newUTXOHashes = append(newUTXOHashes, utxoHash2)
		newUTXOHashes = append(newUTXOHashes, utxoHash3)
		stateRootProposal, err := hndlr.GetStateRootForProposal(txn, newUTXOIDs, newUTXOHashes, [][]byte{})
		if err != nil {
			t.Fatal(err)
		}
		stateRoot, err := hndlr.ApplyState(txn, newUTXOIDs, newUTXOHashes, [][]byte{}, height)
		if err != nil {
			t.Fatal(err)
		}
		height++
		if !bytes.Equal(stateRoot, stateRootProposal) {
			t.Fatal("roots not equal")
		}
		ok, err = hndlr.Contains(nil, utxoID1)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("missing utxo")
		}
		ok, err = hndlr.Contains(nil, utxoID2)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("missing utxo")
		}
		ok, err = hndlr.Contains(nil, utxoID3)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("missing utxo")
		}
		newUTXOIDs = [][]byte{}
		newUTXOIDs = append(newUTXOIDs, utxoID4)
		newUTXOIDs = append(newUTXOIDs, utxoID5)
		newUTXOIDs = append(newUTXOIDs, utxoID6)
		newUTXOHashes = [][]byte{}
		newUTXOHashes = append(newUTXOHashes, utxoHash4)
		newUTXOHashes = append(newUTXOHashes, utxoHash5)
		newUTXOHashes = append(newUTXOHashes, utxoHash6)
		consumedUTXOIDs := [][]byte{}
		consumedUTXOIDs = append(consumedUTXOIDs, utxoID1)
		consumedUTXOIDs = append(consumedUTXOIDs, utxoID2)
		consumedUTXOIDs = append(consumedUTXOIDs, utxoID3)
		stateRootProposal, err = hndlr.GetStateRootForProposal(txn, newUTXOIDs, newUTXOHashes, consumedUTXOIDs)
		if err != nil {
			t.Fatal(err)
		}
		stateRoot, err = hndlr.ApplyState(txn, newUTXOIDs, newUTXOHashes, consumedUTXOIDs, height)
		if err != nil {
			t.Fatal(err)
		}
		height++
		if !bytes.Equal(stateRoot, stateRootProposal) {
			t.Fatal("roots not equal")
		}
		ok, err = hndlr.Contains(nil, utxoID1)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Fatalf("did not fail")
		}
		ok, err = hndlr.Contains(nil, utxoID2)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Fatalf("did not fail")
		}
		ok, err = hndlr.Contains(nil, utxoID3)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Fatalf("did not fail")
		}
		ok, err = hndlr.Contains(nil, utxoID4)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("missing utxo")
		}
		ok, err = hndlr.Contains(nil, utxoID5)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("missing utxo")
		}
		ok, err = hndlr.Contains(nil, utxoID6)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("missing utxo")
		}
		return nil
	})

}
*/
