package minedtx

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/MadBase/MadNet/constants/dbprefix"
	"github.com/MadBase/MadNet/internal/testing/environment"

	"github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/dgraph-io/badger/v2"
)

func testingOwner(t *testing.T) objs.Signer {
	t.Helper()
	signer := &crypto.Secp256k1Signer{}
	err := signer.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

func accountFromSigner(t *testing.T, s objs.Signer) []byte {
	t.Helper()
	pubk, err := s.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	return crypto.GetAccount(pubk)
}

func makeVS(t *testing.T, ownerSigner objs.Signer) *objs.TXOut {
	t.Helper()
	cid := uint32(2)
	val := uint256.One()

	ownerAcct := accountFromSigner(t, ownerSigner)
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
		t.Fatal(err)
	}
	return utxInputs
}

func makeVSTXIn(t *testing.T, ownerSigner objs.Signer, txHash []byte) (*objs.TXOut, *objs.TXIn) {
	t.Helper()
	vs := makeVS(t, ownerSigner)
	vss, err := vs.ValueStore()
	if err != nil {
		t.Fatal(err)
	}
	if txHash == nil {
		txHash = make([]byte, constants.HashLen)
		rand.Read(txHash)
	}
	vss.TxHash = txHash

	txIn, err := vss.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	return vs, txIn
}

func makeTxInitial(t *testing.T, ownerSigner objs.Signer) (objs.Vout, *objs.Tx) {
	t.Helper()
	consumedUTXOs := objs.Vout{}
	txInputs := []*objs.TXIn{}
	for i := 0; i < 2; i++ {
		utxo, txin := makeVSTXIn(t, ownerSigner, nil)
		consumedUTXOs = append(consumedUTXOs, utxo)
		txInputs = append(txInputs, txin)
	}
	generatedUTXOs := objs.Vout{}
	for i := 0; i < 2; i++ {
		generatedUTXOs = append(generatedUTXOs, makeVS(t, ownerSigner))
	}
	err := generatedUTXOs.SetTxOutIdx()
	if err != nil {
		t.Fatal(err)
	}
	txfee := uint256.Zero()
	tx := &objs.Tx{
		Vin:  txInputs,
		Vout: generatedUTXOs,
		Fee:  txfee,
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		vs, err := consumedUTXOs[i].ValueStore()
		if err != nil {
			t.Fatal(err)
		}
		err = vs.Sign(tx.Vin[i], ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}
	return consumedUTXOs, tx
}

func makeTxConsuming(t *testing.T, ownerSigner objs.Signer, consumedUTXOs objs.Vout) *objs.Tx {
	txInputs := []*objs.TXIn{}
	for i := 0; i < 2; i++ {
		txin, err := consumedUTXOs[i].MakeTxIn()
		if err != nil {
			t.Fatal(err)
		}
		txInputs = append(txInputs, txin)
	}
	generatedUTXOs := objs.Vout{}
	for i := 0; i < 2; i++ {
		generatedUTXOs = append(generatedUTXOs, makeVS(t, ownerSigner))
	}
	err := generatedUTXOs.SetTxOutIdx()
	if err != nil {
		t.Fatal(err)
	}
	txfee := uint256.Zero()
	tx := &objs.Tx{
		Vin:  txInputs,
		Vout: generatedUTXOs,
		Fee:  txfee,
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		vs, err := consumedUTXOs[i].ValueStore()
		if err != nil {
			t.Fatal(err)
		}
		err = vs.Sign(tx.Vin[i], ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}
	return tx
}

func TestMined(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)
	hndlr := NewMinedTxHandler()

	signer := &crypto.BNSigner{}
	err := signer.SetPrivk([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}

	msg := makeMockStorageGetter()
	storage := makeStorage(msg)

	ownerSigner := testingOwner(t)
	consumedUTXOs, tx := makeTxInitial(t, ownerSigner)

	tx2 := makeTxConsuming(t, ownerSigner, consumedUTXOs)

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
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)
	hndlr := NewMinedTxHandler()

	ownerSigner := testingOwner(t)
	consumedUTXOs, tx := makeTxInitial(t, ownerSigner)

	msg := makeMockStorageGetter()
	storage := makeStorage(msg)

	_, err := tx.Validate(nil, 1, consumedUTXOs, storage)
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
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)
	hndlr := NewMinedTxHandler()

	ownerSigner := testingOwner(t)
	consumedUTXOs, tx := makeTxInitial(t, ownerSigner)

	msg := makeMockStorageGetter()
	storage := makeStorage(msg)

	_, err := tx.Validate(nil, 1, consumedUTXOs, storage)
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
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)
	hndlr := NewMinedTxHandler()

	ownerSigner := testingOwner(t)
	consumedUTXOs, tx := makeTxInitial(t, ownerSigner)

	msg := makeMockStorageGetter()
	storage := makeStorage(msg)

	_, err := tx.Validate(nil, 1, consumedUTXOs, storage)
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
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)
	hndlr := NewMinedTxHandler()

	ownerSigner := testingOwner(t)
	consumedUTXOs, tx := makeTxInitial(t, ownerSigner)

	msg := makeMockStorageGetter()
	storage := makeStorage(msg)

	_, err := tx.Validate(nil, 1, consumedUTXOs, storage)
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
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)
	hndlr := NewMinedTxHandler()

	ownerSigner := testingOwner(t)
	consumedUTXOs, tx := makeTxInitial(t, ownerSigner)

	msg := makeMockStorageGetter()
	storage := makeStorage(msg)

	_, err := tx.Validate(nil, 1, consumedUTXOs, storage)
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
	t.Parallel()
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
