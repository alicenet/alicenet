package minedtx

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/MadBase/MadNet/constants/dbprefix"

	"github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/dgraph-io/badger/v2"
)

func testingOwner() objs.Signer {
	signer := &crypto.Secp256k1Signer{}
	err := signer.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		panic(err)
	}
	return signer
}

func accountFromSigner(s objs.Signer) []byte {
	pubk, err := s.Pubkey()
	if err != nil {
		panic(err)
	}
	return crypto.GetAccount(pubk)
}

func makeVS(ownerSigner objs.Signer) *objs.TXOut {
	cid := uint32(2)
	//val := uint32(1)
	val := uint256.One()

	ownerAcct := accountFromSigner(ownerSigner)
	owner := &objs.ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	vsp := &objs.VSPreImage{
		ChainID: cid,
		Value:   val,
		Owner:   owner,
		Fee:     uint256.Zero(),
	}
	vs := &objs.ValueStore{
		VSPreImage: vsp,
		TxHash:     make([]byte, constants.HashLen),
	}
	utxInputs := &objs.TXOut{}
	err := utxInputs.NewValueStore(vs)
	if err != nil {
		panic(err)
	}
	return utxInputs
}

func makeVSTXIn(ownerSigner objs.Signer, txHash []byte) (*objs.TXOut, *objs.TXIn) {
	vs := makeVS(ownerSigner)
	vss, err := vs.ValueStore()
	if err != nil {
		panic(err)
	}
	if txHash == nil {
		txHash = make([]byte, constants.HashLen)
		rand.Read(txHash)
	}
	vss.TxHash = txHash

	txIn, err := vss.MakeTxIn()
	if err != nil {
		panic(err)
	}
	return vs, txIn
}

func makeTxInitial(ownerSigner objs.Signer) (objs.Vout, *objs.Tx) {
	consumedUTXOs := objs.Vout{}
	txInputs := []*objs.TXIn{}
	for i := 0; i < 2; i++ {
		utxo, txin := makeVSTXIn(ownerSigner, nil)
		consumedUTXOs = append(consumedUTXOs, utxo)
		txInputs = append(txInputs, txin)
	}
	generatedUTXOs := objs.Vout{}
	for i := 0; i < 2; i++ {
		generatedUTXOs = append(generatedUTXOs, makeVS(ownerSigner))
	}
	err := generatedUTXOs.SetTxOutIdx()
	if err != nil {
		panic(err)
	}
	txfee := uint256.Zero()
	tx := &objs.Tx{
		Vin:  txInputs,
		Vout: generatedUTXOs,
		Fee:  txfee,
	}
	err = tx.SetTxHash()
	if err != nil {
		panic(err)
	}
	for i := 0; i < 2; i++ {
		vs, err := consumedUTXOs[i].ValueStore()
		if err != nil {
			panic(err)
		}
		err = vs.Sign(tx.Vin[i], ownerSigner)
		if err != nil {
			panic(err)
		}
	}
	return consumedUTXOs, tx
}

func makeTxConsuming(ownerSigner objs.Signer, consumedUTXOs objs.Vout) *objs.Tx {
	txInputs := []*objs.TXIn{}
	for i := 0; i < 2; i++ {
		txin, err := consumedUTXOs[i].MakeTxIn()
		if err != nil {
			panic(err)
		}
		txInputs = append(txInputs, txin)
	}
	generatedUTXOs := objs.Vout{}
	for i := 0; i < 2; i++ {
		generatedUTXOs = append(generatedUTXOs, makeVS(ownerSigner))
	}
	err := generatedUTXOs.SetTxOutIdx()
	if err != nil {
		panic(err)
	}
	txfee := uint256.Zero()
	tx := &objs.Tx{
		Vin:  txInputs,
		Vout: generatedUTXOs,
		Fee:  txfee,
	}
	err = tx.SetTxHash()
	if err != nil {
		panic(err)
	}
	for i := 0; i < 2; i++ {
		vs, err := consumedUTXOs[i].ValueStore()
		if err != nil {
			panic(err)
		}
		err = vs.Sign(tx.Vin[i], ownerSigner)
		if err != nil {
			panic(err)
		}
	}
	return tx
}

func TestMined(t *testing.T) {
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
	////////////////////////////////////////
	hndlr := NewMinedTxHandler()

	signer := &crypto.BNSigner{}
	err = signer.SetPrivk([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}

	msg := makeMockStorageGetter()
	storage := makeStorage(msg)

	////////////////////////////////////////
	ownerSigner := testingOwner()
	consumedUTXOs, tx := makeTxInitial(ownerSigner)

	tx2 := makeTxConsuming(ownerSigner, consumedUTXOs)

	_, err = tx.Validate(nil, 1, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx2.Validate(nil, 1, tx.Vout, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		txHash, err := tx.TxHash()
		if err != nil {
			t.Fatal(err)
		}
		err = hndlr.Add(txn, 1, []*objs.Tx{tx})
		if err != nil {
			t.Fatal(err)
		}
		getTx1, _, err := hndlr.Get(txn, [][]byte{txHash})
		if err != nil {
			t.Fatal(err)
		}
		getTxHash1, err := getTx1[0].TxHash()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(getTxHash1, txHash) {
			t.Fatalf("txHash mismatch:\noriginalHash:%x\nreturnedHash:%x\n", txHash, getTxHash1)
		}
		err = hndlr.Delete(txn, [][]byte{txHash})
		if err != nil {
			t.Fatal(err)
		}
		_, missing, err := hndlr.Get(txn, [][]byte{txHash})
		if len(missing) != 1 {
			t.Error(err)
			t.Fatal("delete failure")
		}
		tx2Hash, err := tx2.TxHash()
		if err != nil {
			t.Fatal(err)
		}
		err = hndlr.Add(txn, 1, []*objs.Tx{tx, tx2})
		if err != nil {
			t.Fatal(err)
		}
		getTx1, _, err = hndlr.Get(txn, [][]byte{txHash})
		if err != nil {
			t.Fatal(err)
		}
		getTxHash1, err = getTx1[0].TxHash()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(getTxHash1, txHash) {
			t.Fatalf("txHash mismatch:\noriginalHash:%x\nreturnedHash:%x\n", txHash, getTxHash1)
		}
		getTx2, _, err := hndlr.Get(txn, [][]byte{tx2Hash})
		if err != nil {
			t.Fatal(err)
		}
		getTxHash2, err := getTx2[0].TxHash()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(getTxHash2, tx2Hash) {
			t.Fatalf("txHash mismatch:\noriginalHash:%x\nreturnedHash:%x\n", tx2Hash, getTxHash2)
		}
		atHeight, err := hndlr.GetHeightForTx(txn, getTxHash2)
		if err != nil {
			t.Fatal(err)
		}
		if atHeight != 1 {
			t.Fatalf("not at height 1: %v", atHeight)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMinedDelete(t *testing.T) {
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
	////////////////////////////////////////
	hndlr := NewMinedTxHandler()

	ownerSigner := testingOwner()
	consumedUTXOs, tx := makeTxInitial(ownerSigner)

	msg := makeMockStorageGetter()
	storage := makeStorage(msg)

	_, err = tx.Validate(nil, 1, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txHash, err := tx.TxHash()
	if err != nil {
		t.Fatal(err)
	}

	height := uint32(1)

	err = db.Update(func(txn *badger.Txn) error {
		err := hndlr.Delete(txn, [][]byte{txHash})
		if err == nil {
			t.Fatal("Should have raised error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		err := hndlr.Add(txn, height, []*objs.Tx{tx})
		if err != nil {
			t.Fatal(err)
		}
		err = hndlr.Delete(txn, [][]byte{txHash})
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMinedGet(t *testing.T) {
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
	////////////////////////////////////////
	hndlr := NewMinedTxHandler()

	ownerSigner := testingOwner()
	consumedUTXOs, tx := makeTxInitial(ownerSigner)

	msg := makeMockStorageGetter()
	storage := makeStorage(msg)

	_, err = tx.Validate(nil, 1, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txHash, err := tx.TxHash()
	if err != nil {
		t.Fatal(err)
	}

	height := uint32(1)

	err = db.Update(func(txn *badger.Txn) error {
		txs, missing, err := hndlr.Get(txn, [][]byte{txHash})
		if err != nil {
			t.Fatal(err)
		}
		if len(missing) != 1 {
			t.Fatal("Should not return any missing txhashes")
		}
		if len(txs) != 0 {
			t.Fatal("Returned the incorrect txs; should not return any")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		err := hndlr.Add(txn, height, []*objs.Tx{tx})
		if err != nil {
			t.Fatal(err)
		}
		txs, missing, err := hndlr.Get(txn, [][]byte{txHash})
		if err != nil {
			t.Fatal(err)
		}
		if len(missing) != 0 {
			t.Fatal("Should not return any missing txhashes")
		}
		if len(txs) != 1 {
			t.Fatal("Returned the incorrect txs; should return 1")
		}
		retTx := txs[0]
		retTxHash, err := retTx.TxHash()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(retTxHash, txHash) {
			t.Fatal("TxHashes do not agree")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMinedGetHeightForTx(t *testing.T) {
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
	////////////////////////////////////////
	hndlr := NewMinedTxHandler()

	ownerSigner := testingOwner()
	consumedUTXOs, tx := makeTxInitial(ownerSigner)

	msg := makeMockStorageGetter()
	storage := makeStorage(msg)

	_, err = tx.Validate(nil, 1, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}

	txHashBad := make([]byte, constants.HashLen)
	height := uint32(1)

	err = db.Update(func(txn *badger.Txn) error {
		_, err := hndlr.GetHeightForTx(txn, txHashBad)
		if err == nil {
			t.Fatal("Should have raised error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		txHash, err := tx.TxHash()
		if err != nil {
			t.Fatal(err)
		}
		err = hndlr.Add(txn, height, []*objs.Tx{tx})
		if err != nil {
			t.Fatal(err)
		}
		retHeight, err := hndlr.GetHeightForTx(txn, txHash)
		if err != nil {
			t.Fatal(err)
		}
		if height != retHeight {
			t.Fatal("heights do not agree")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMinedGetOneInternal(t *testing.T) {
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
	////////////////////////////////////////
	hndlr := NewMinedTxHandler()

	ownerSigner := testingOwner()
	consumedUTXOs, tx := makeTxInitial(ownerSigner)

	msg := makeMockStorageGetter()
	storage := makeStorage(msg)

	_, err = tx.Validate(nil, 1, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}

	txHashBad := make([]byte, constants.HashLen)
	height := uint32(1)

	err = db.Update(func(txn *badger.Txn) error {
		_, err := hndlr.getOneInternal(txn, txHashBad)
		if err == nil {
			t.Fatal("Should have raised error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		txHash, err := tx.TxHash()
		if err != nil {
			t.Fatal(err)
		}
		err = hndlr.addOneInternal(txn, tx, txHash, height)
		if err != nil {
			t.Fatal(err)
		}
		retTx, err := hndlr.getOneInternal(txn, txHash)
		if err != nil {
			t.Fatal(err)
		}
		retTxBytes, err := retTx.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		txBytes, err := tx.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(txBytes, retTxBytes) {
			t.Fatal("txs do not agree")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMinedAddOneInternal(t *testing.T) {
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
	////////////////////////////////////////
	hndlr := NewMinedTxHandler()

	ownerSigner := testingOwner()
	consumedUTXOs, tx := makeTxInitial(ownerSigner)

	msg := makeMockStorageGetter()
	storage := makeStorage(msg)

	_, err = tx.Validate(nil, 1, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}

	txBad := &objs.Tx{}
	txHashBad := make([]byte, constants.HashLen)
	height := uint32(1)

	err = db.Update(func(txn *badger.Txn) error {
		err := hndlr.addOneInternal(txn, txBad, txHashBad, height)
		if err == nil {
			t.Fatal("Should have raised error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		txHash, err := tx.TxHash()
		if err != nil {
			t.Fatal(err)
		}
		err = hndlr.addOneInternal(txn, tx, txHash, height)
		if err != nil {
			t.Fatal(err)
		}
		retTx, err := hndlr.getOneInternal(txn, txHash)
		if err != nil {
			t.Fatal(err)
		}
		retTxBytes, err := retTx.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		txBytes, err := tx.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(txBytes, retTxBytes) {
			t.Fatal("txs do not agree")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMinedMakeMinedTxKey(t *testing.T) {
	txHash := crypto.Hasher([]byte("txHash"))
	trueKey := []byte{}
	trueKey = append(trueKey, dbprefix.PrefixMinedTx()...)
	trueKey = append(trueKey, txHash...)

	hndlr := NewMinedTxHandler()
	key := hndlr.makeMinedTxKey(txHash)
	if !bytes.Equal(trueKey, key) {
		t.Fatal("keys do not agree")
	}
}
