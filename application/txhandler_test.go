package application

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/dgraph-io/badger/v2"
)

func TestUTXOTrie(t *testing.T) {
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
	//signer := &crypto.Signer{}
	//verifier := &mockTxSigVerifier{}
	//hndlr := NewUTXOHandler(db, verifier)
	//hndlr.Init(1)
	//_ = constants.HashLen
}
