package utxohandler

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

func makeDeposit(t *testing.T, s objs.Signer, chainID uint32, i int, value *uint256.Uint256) *objs.ValueStore {
	pubkey, err := s.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	vs := &objs.ValueStore{
		VSPreImage: &objs.VSPreImage{
			TXOutIdx: constants.MaxUint32,
			Value:    value,
			ChainID:  chainID,
			Owner:    &objs.ValueStoreOwner{SVA: objs.ValueStoreSVA, CurveSpec: constants.CurveSecp256k1, Account: crypto.GetAccount(pubkey)},
		},
		TxHash: utils.ForceSliceToLength([]byte(strconv.Itoa(i)), constants.HashLen),
	}
	return vs
}

func makeTxs(t *testing.T, s objs.Signer, v *objs.ValueStore) *objs.Tx {
	txIn, err := v.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	value, err := v.Value()
	if err != nil {
		t.Fatal(err)
	}
	chainID, err := txIn.ChainID()
	if err != nil {
		t.Fatal(err)
	}
	pubkey, err := s.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	tx := &objs.Tx{}
	tx.Vin = []*objs.TXIn{txIn}
	newValueStore := &objs.ValueStore{
		VSPreImage: &objs.VSPreImage{
			ChainID:  chainID,
			Value:    value,
			Owner:    &objs.ValueStoreOwner{SVA: objs.ValueStoreSVA, CurveSpec: constants.CurveSecp256k1, Account: crypto.GetAccount(pubkey)},
			TXOutIdx: 0,
			Fee:      new(uint256.Uint256).SetZero(),
		},
		TxHash: make([]byte, constants.HashLen),
	}
	newUTXO := &objs.TXOut{}
	err = newUTXO.NewValueStore(newValueStore)
	if err != nil {
		t.Fatal(err)
	}
	tx.Vout = append(tx.Vout, newUTXO)
	tx.Fee = uint256.Zero()
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	err = v.Sign(tx.Vin[0], s)
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

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
	signer := &crypto.Secp256k1Signer{}
	err = signer.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	hndlr := NewUTXOHandler(db)
	err = hndlr.Init(1)
	if err != nil {
		t.Fatal(err)
	}
	//d := makeDeposit(t, signer, 1, 1, 1)
	d := makeDeposit(t, signer, 1, 1, uint256.One())
	utxoDep := &objs.TXOut{}
	err = utxoDep.NewValueStore(d)
	if err != nil {
		t.Fatal(err)
	}
	tx := makeTxs(t, signer, d)
	err = db.Update(func(txn *badger.Txn) error {
		_, err := hndlr.IsValid(txn, []*objs.Tx{tx}, 1, objs.Vout{utxoDep})
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		_, err := hndlr.ApplyState(txn, []*objs.Tx{tx}, 2)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		_, err := hndlr.ApplyState(txn, []*objs.Tx{tx}, 2)
		if err == nil {
			t.Fatal("should fail")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	utxoIDs, err := tx.GeneratedUTXOID()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		_, missing, err := hndlr.Get(txn, utxoIDs)
		if err != nil {
			t.Fatal(err)
		}
		if len(missing) != 0 {
			t.Fatal("missing utxoID")
		}
		_, missing, err = hndlr.Get(txn, [][]byte{crypto.Hasher([]byte("nil"))})
		if err != nil {
			t.Fatal(err)
		}
		if len(missing) != 1 {
			t.Fatal("not missing utxoID")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
