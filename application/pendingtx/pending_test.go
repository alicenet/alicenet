package pendingtx

import (
	"bytes"
	"context"
	"crypto/rand"
	"io/ioutil"
	"os"
	"testing"

	"github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/dgraph-io/badger/v2"
)

type mockTrie struct {
	m map[string]bool
}

func (mt *mockTrie) IsValid(txn *badger.Txn, txs objs.TxVec, currentHeight uint32, deposits objs.Vout) (objs.Vout, error) {
	return nil, nil
}

func (mt *mockTrie) TrieContains(txn *badger.Txn, utxo []byte) (bool, error) {
	return mt.m[string(utxo)], nil
}

func (mt *mockTrie) Add(utxo []byte) {
	mt.m[string(utxo)] = true
}

func (mt *mockTrie) Remove(utxo []byte) {
	delete(mt.m, string(utxo))
}

func (mt *mockTrie) Get(txn *badger.Txn, utxoIDs [][]byte) ([]*objs.TXOut, [][]byte, []*objs.TXOut, error) {
	return nil, nil, nil, nil
}

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

func makeVS(ownerSigner objs.Signer, fee *uint256.Uint256) *objs.TXOut {
	cid := uint32(2)
	val := uint256.One()

	ownerAcct := accountFromSigner(ownerSigner)
	owner := &objs.ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	vsp := &objs.VSPreImage{
		ChainID: cid,
		Value:   val,
		Owner:   owner,
		Fee:     fee,
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
	vs := makeVS(ownerSigner, uint256.One())
	vss, err := vs.ValueStore()
	if err != nil {
		panic(err)
	}
	if txHash == nil {
		txHash = make([]byte, constants.HashLen)
		_, err := rand.Read(txHash)
		if err != nil {
			panic(err)
		}
	}
	vss.TxHash = txHash
	txin, err := vss.MakeTxIn()
	if err != nil {
		panic(err)
	}
	return vs, txin
}

func makeVSwithFees(ownerSigner objs.Signer, value, vsfee *uint256.Uint256) *objs.TXOut {
	cid := uint32(2)

	ownerAcct := accountFromSigner(ownerSigner)
	owner := &objs.ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	vsp := &objs.VSPreImage{
		ChainID: cid,
		Value:   value.Clone(),
		Owner:   owner,
		Fee:     vsfee.Clone(),
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

func makeVSTXInwithFees(ownerSigner objs.Signer, txHash []byte, value, vsfee *uint256.Uint256) (*objs.TXOut, *objs.TXIn) {
	vs := makeVSwithFees(ownerSigner, value, vsfee)
	vss, err := vs.ValueStore()
	if err != nil {
		panic(err)
	}
	if txHash == nil {
		txHash = make([]byte, constants.HashLen)
		_, err := rand.Read(txHash)
		if err != nil {
			panic(err)
		}
	}
	vss.TxHash = txHash
	txin, err := vss.MakeTxIn()
	if err != nil {
		panic(err)
	}
	return vs, txin
}

func makeTxInitialwithFees(value, txfee, vsfee *uint256.Uint256) (objs.Vout, *objs.Tx) {
	ownerSigner := testingOwner()
	consumedUTXOs := objs.Vout{}
	txInputs := []*objs.TXIn{}
	utxo, txin := makeVSTXInwithFees(ownerSigner, nil, value, vsfee)
	consumedUTXOs = append(consumedUTXOs, utxo)
	txInputs = append(txInputs, txin)

	generatedUTXOs := objs.Vout{}
	valueMinusFees := new(uint256.Uint256)
	err := valueMinusFees.Set(value)
	if err != nil {
		panic(err)
	}
	_, err = valueMinusFees.Sub(valueMinusFees, vsfee)
	if err != nil {
		panic(err)
	}
	_, err = valueMinusFees.Sub(valueMinusFees, txfee)
	if err != nil {
		panic(err)
	}
	generatedUTXOs = append(generatedUTXOs, makeVSwithFees(ownerSigner, valueMinusFees, vsfee))
	err = generatedUTXOs.SetTxOutIdx()
	if err != nil {
		panic(err)
	}
	tx := &objs.Tx{
		Vin:  txInputs,
		Vout: generatedUTXOs,
		Fee:  txfee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		panic(err)
	}
	for i := 0; i < 1; i++ {
		vs, err := consumedUTXOs[i].ValueStore()
		if err != nil {
			panic(err)
		}
		err = vs.Sign(tx.Vin[i], ownerSigner)
		if err != nil {
			panic(err)
		}
	}
	err = tx.ValidateEqualVinVout(1, consumedUTXOs)
	if err != nil {
		panic(err)
	}
	return consumedUTXOs, tx
}

func makeTxInitial() (objs.Vout, *objs.Tx) {
	ownerSigner := testingOwner()
	consumedUTXOs := objs.Vout{}
	txInputs := []*objs.TXIn{}
	for i := 0; i < 2; i++ {
		utxo, txin := makeVSTXIn(ownerSigner, nil)
		consumedUTXOs = append(consumedUTXOs, utxo)
		txInputs = append(txInputs, txin)
	}
	generatedUTXOs := objs.Vout{}
	for i := 0; i < 2; i++ {
		generatedUTXOs = append(generatedUTXOs, makeVS(ownerSigner, uint256.One()))
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

func makeTxConsuming(consumedUTXOs objs.Vout) *objs.Tx {
	ownerSigner := testingOwner()
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
		generatedUTXOs = append(generatedUTXOs, makeVS(ownerSigner, uint256.One()))
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

func mustAddTx(t *testing.T, hndlr *Handler, tx *objs.Tx, currentHeight uint32) {
	err := hndlr.Add(nil, []*objs.Tx{tx}, currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	mustContain(t, hndlr, tx)
}

func mustNotAdd(t *testing.T, hndlr *Handler, tx *objs.Tx, currentHeight uint32) {
	hndlr.Add(nil, []*objs.Tx{tx}, currentHeight)
	mustNotContain(t, hndlr, tx)
}

func mustContain(t *testing.T, hndlr *Handler, tx *objs.Tx) {
	txHash, err := tx.TxHash()
	if err != nil {
		t.Fatal(err)
	}
	getTx1, missing, err := hndlr.Get(nil, 1, [][]byte{txHash})
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 0 {
		t.Fatalf("missing %x", txHash)
	}
	getTxHash1, err := getTx1[0].TxHash()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(getTxHash1, txHash) {
		t.Fatalf("txHash mismatch:\noriginalHash:%x\nreturnedHash:%x\n", txHash, getTxHash1)
	}
}

func mustNotContain(t *testing.T, hndlr *Handler, tx *objs.Tx) {
	txHash, err := tx.TxHash()
	if err != nil {
		t.Fatal(err)
	}
	_, missing, err := hndlr.Get(nil, 1, [][]byte{txHash})
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 1 {
		t.Fatalf("delete failure: %x", txHash)
	}
	missing, err = hndlr.Contains(nil, 1, [][]byte{txHash})
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 1 {
		t.Fatal("contains")
	}
}

func mustDelTx(t *testing.T, hndlr *Handler, tx *objs.Tx) {
	txHash, err := tx.TxHash()
	if err != nil {
		t.Fatal(err)
	}
	err = hndlr.Delete(nil, [][]byte{txHash})
	if err != nil {
		t.Fatal(err)
	}
	_, missing, err := hndlr.Get(nil, 1, [][]byte{txHash})
	if len(missing) != 1 {
		t.Fatalf("delete failure: %v", err)
	}
}

func setup(t *testing.T) (*Handler, *mockTrie, func()) {
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	fn := func() {
		defer os.RemoveAll(dir)
		defer db.Close()
	}
	////////////////////////////////////////
	mt := &mockTrie{}
	mt.m = make(map[string]bool)
	queueSize := 1024
	hndlr, err := NewPendingTxHandler(db, queueSize)
	if err != nil {
		t.Fatal(err)
	}
	hndlr.UTXOHandler = mt
	hndlr.DepositHandler = mt
	return hndlr, mt, fn
}

func TestAdd(t *testing.T) {
	hndlr, _, cleanup := setup(t)
	defer cleanup()
	_, tx := makeTxInitial()
	mustAddTx(t, hndlr, tx, 1)
}

func TestAddErrors(t *testing.T) {
	hndlr, _, cleanup := setup(t)
	defer cleanup()
	_, tx := makeTxInitial()
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	// Attempt to add empty tx
	txBad0 := &objs.Tx{}
	err = hndlr.Add(nil, []*objs.Tx{txBad0}, 1)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	// Attempt to add tx with tx.Vout == nil
	txBad1 := &objs.Tx{}
	err = txBad1.UnmarshalBinary(txBytes)
	if err != nil {
		t.Fatal(err)
	}
	txBad1.Vout = nil
	err = hndlr.Add(nil, []*objs.Tx{txBad1}, 1)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestDel(t *testing.T) {
	hndlr, _, cleanup := setup(t)
	defer cleanup()
	_, tx := makeTxInitial()
	mustAddTx(t, hndlr, tx, 1)
	mustDelTx(t, hndlr, tx)
}

func TestDeleteMined(t *testing.T) {
	hndlr, _, cleanup := setup(t)
	defer cleanup()
	vout, tx := makeTxInitial()
	mustAddTx(t, hndlr, tx, 1)
	tx2 := makeTxConsuming(vout)
	mustAddTx(t, hndlr, tx2, 1)
	txHash, err := tx.TxHash()
	if err != nil {
		t.Fatal(err)
	}
	consumedUTXOIDs, err := tx.ConsumedUTXOID()
	if err != nil {
		t.Fatal(err)
	}
	err = hndlr.DeleteMined(nil, 1, [][]byte{txHash}, consumedUTXOIDs)
	if err != nil {
		t.Fatal(err)
	}
	mustNotContain(t, hndlr, tx2)
	mustNotAdd(t, hndlr, tx, 1)
}

func TestMissing(t *testing.T) {
	hndlr, _, cleanup := setup(t)
	defer cleanup()
	_, tx := makeTxInitial()
	mustAddTx(t, hndlr, tx, 1)
	_, tx2 := makeTxInitial()
	mustNotContain(t, hndlr, tx2)
}

func TestGetProposal(t *testing.T) {
	hndlr, trie, cleanup := setup(t)
	defer cleanup()
	c1, tx1 := makeTxInitial()
	mustAddTx(t, hndlr, tx1, 1)
	c2, tx2 := makeTxInitial()
	mustAddTx(t, hndlr, tx2, 1)
	tx3 := makeTxConsuming(c1)
	mustAddTx(t, hndlr, tx3, 1)
	tx4 := makeTxConsuming(c2)
	mustAddTx(t, hndlr, tx4, 1)
	maxBytes := constants.MaxUint32
	hndlr.txqueue.ClearTxQueue()
	txs, _, err := hndlr.GetTxsForProposal(hndlr.db.NewTransaction(false), context.TODO(), 1, maxBytes, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Length: %v\n", len(txs))
	txHashes, err := objs.TxVec(txs).TxHash()
	if err != nil {
		t.Fatal(err)
	}
	// trie must contain utxos getting spent but must not contain
	// utxos being generated
	utxoIDs, err := objs.TxVec{tx1, tx2, tx3, tx4}.ConsumedUTXOID()
	if err != nil {
		t.Fatal(err)
	}
	for _, ut := range utxoIDs {
		trie.Add(ut)
	}
	utxoIDs, err = objs.TxVec{tx1, tx2, tx3, tx4}.GeneratedUTXOID()
	if err != nil {
		t.Fatal(err)
	}
	for _, ut := range utxoIDs {
		trie.Remove(ut)
	}
	txs, err = hndlr.GetTxsForGossip(nil, context.Background(), 1, constants.MaxUint32)
	if err != nil {
		t.Fatal(err)
	}
	if len(txs) != 2 {
		t.Fatalf("conflict: %x", txHashes)
	}
}

func TestGetTxsForProposal1(t *testing.T) {
	hndlr, trie, cleanup := setup(t)
	defer cleanup()
	value1, err := new(uint256.Uint256).FromUint64(123)
	if err != nil {
		t.Fatal(err)
	}
	txfee1, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	vsfee, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	_, tx1 := makeTxInitialwithFees(value1, txfee1, vsfee)
	txhash1, err := tx1.TxHash()
	if err != nil {
		t.Fatal(err)
	}
	mustAddTx(t, hndlr, tx1, 1)
	value2, err := new(uint256.Uint256).FromUint64(1234567)
	if err != nil {
		t.Fatal(err)
	}
	txfee2, err := new(uint256.Uint256).FromUint64(100)
	if err != nil {
		t.Fatal(err)
	}
	_, tx2 := makeTxInitialwithFees(value2, txfee2, vsfee)
	txhash2, err := tx2.TxHash()
	if err != nil {
		t.Fatal(err)
	}
	mustAddTx(t, hndlr, tx2, 1)
	value3, err := new(uint256.Uint256).FromUint64(1234)
	if err != nil {
		t.Fatal(err)
	}
	txfee3, err := new(uint256.Uint256).FromUint64(1000)
	if err != nil {
		t.Fatal(err)
	}
	_, tx3 := makeTxInitialwithFees(value3, txfee3, vsfee)
	txhash3, err := tx3.TxHash()
	if err != nil {
		t.Fatal(err)
	}
	mustAddTx(t, hndlr, tx3, 1)
	value4, err := new(uint256.Uint256).FromUint64(12341235235232)
	if err != nil {
		t.Fatal(err)
	}
	txfee4, err := new(uint256.Uint256).FromUint64(10)
	if err != nil {
		t.Fatal(err)
	}
	_, tx4 := makeTxInitialwithFees(value4, txfee4, vsfee)
	txhash4, err := tx4.TxHash()
	if err != nil {
		t.Fatal(err)
	}
	mustAddTx(t, hndlr, tx4, 1)
	// trie must contain utxos getting spent but must not contain
	// utxos being generated
	utxoIDs, err := objs.TxVec{tx1, tx2, tx3, tx4}.ConsumedUTXOID()
	if err != nil {
		t.Fatal(err)
	}
	for _, ut := range utxoIDs {
		trie.Add(ut)
	}
	maxBytes := constants.MaxUint32
	txs, _, err := hndlr.GetTxsForProposal(hndlr.db.NewTransaction(false), context.TODO(), 1+3*constants.EpochLength, maxBytes, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Length: %v\n", len(txs))
	if len(txs) != 4 {
		t.Fatal("invalid number of txs")
	}
	txHashes, err := objs.TxVec(txs).TxHash()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(txHashes[0], txhash3) {
		t.Fatal("invalid first hash")
	}
	if !bytes.Equal(txHashes[1], txhash2) {
		t.Fatal("invalid second hash")
	}
	if !bytes.Equal(txHashes[2], txhash4) {
		t.Fatal("invalid third hash")
	}
	if !bytes.Equal(txHashes[3], txhash1) {
		t.Fatal("invalid fourth hash")
	}
}

func TestGetTxsForProposal2(t *testing.T) {
	hndlr, trie, cleanup := setup(t)
	defer cleanup()
	value1, err := new(uint256.Uint256).FromUint64(123)
	if err != nil {
		t.Fatal(err)
	}
	txfee1, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	vsfee, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	_, tx1 := makeTxInitialwithFees(value1, txfee1, vsfee)
	mustAddTx(t, hndlr, tx1, 1)
	value2, err := new(uint256.Uint256).FromUint64(1234567)
	if err != nil {
		t.Fatal(err)
	}
	txfee2, err := new(uint256.Uint256).FromUint64(100)
	if err != nil {
		t.Fatal(err)
	}
	_, tx2 := makeTxInitialwithFees(value2, txfee2, vsfee)
	mustAddTx(t, hndlr, tx2, 1)
	value3, err := new(uint256.Uint256).FromUint64(1234)
	if err != nil {
		t.Fatal(err)
	}
	txfee3, err := new(uint256.Uint256).FromUint64(1000)
	if err != nil {
		t.Fatal(err)
	}
	_, tx3 := makeTxInitialwithFees(value3, txfee3, vsfee)
	mustAddTx(t, hndlr, tx3, 1)
	value4, err := new(uint256.Uint256).FromUint64(12341235235232)
	if err != nil {
		t.Fatal(err)
	}
	txfee4, err := new(uint256.Uint256).FromUint64(10)
	if err != nil {
		t.Fatal(err)
	}
	_, tx4 := makeTxInitialwithFees(value4, txfee4, vsfee)
	mustAddTx(t, hndlr, tx4, 1)
	// trie must contain utxos getting spent but must not contain
	// utxos being generated
	utxoIDs, err := objs.TxVec{tx1, tx2, tx3, tx4}.ConsumedUTXOID()
	if err != nil {
		t.Fatal(err)
	}
	for _, ut := range utxoIDs {
		trie.Add(ut)
	}
	maxBytes := constants.MaxUint32

	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	txs, retMaxBytes, err := hndlr.GetTxsForProposal(hndlr.db.NewTransaction(false), ctx, 1, maxBytes, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(txs) != 0 {
		t.Fatal("invalid number of txs")
	}
	if retMaxBytes != maxBytes {
		t.Fatal("invalid retMaxBytes")
	}
}

func TestAddTxsToQueueFullQueue(t *testing.T) {
	hndlr, _, cleanup := setup(t)
	defer cleanup()

	_, tx := makeTxInitial()
	mustAddTx(t, hndlr, tx, 1)

	_, tx2 := makeTxInitial()
	mustAddTx(t, hndlr, tx2, 1)

	hndlr.SetQueueSize(1)
	err := hndlr.AddTxsToQueue(nil, context.TODO(), 1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddTxsToQueueStatusChecks(t *testing.T) {
	hndlr, _, cleanup := setup(t)
	defer cleanup()

	_, tx := makeTxInitial()
	mustAddTx(t, hndlr, tx, 1)

	_, tx2 := makeTxInitial()
	mustAddTx(t, hndlr, tx2, 1)

	hndlr.SetQueueSize(1)
	if hndlr.TxQueueAddStatus() {
		t.Fatal("Status should be false")
	}
	hndlr.TxQueueAddStart()
	if !hndlr.TxQueueAddStatus() {
		t.Fatal("Status should be true")
	}
	hndlr.TxQueueAddStart()
	if !hndlr.TxQueueAddStatus() {
		t.Fatal("Status should be true")
	}
	hndlr.TxQueueAddStop()
	if hndlr.TxQueueAddStatus() {
		t.Fatal("Status should be false")
	}

	if hndlr.TxQueueAddFinished() {
		t.Fatal("StatusFinished should be false")
	}
	err := hndlr.AddTxsToQueue(hndlr.db.NewTransaction(false), context.TODO(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if !hndlr.TxQueueAddFinished() {
		t.Fatal("StatusFinished should be true")
	}
	err = hndlr.AddTxsToQueue(hndlr.db.NewTransaction(false), context.TODO(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if !hndlr.TxQueueAddFinished() {
		t.Fatal("StatusFinished should still be true")
	}
}

func TestAddTxsToQueueStartStop(t *testing.T) {
	hndlr, trie, cleanup := setup(t)
	defer cleanup()

	value1, err := new(uint256.Uint256).FromUint64(123)
	if err != nil {
		t.Fatal(err)
	}
	txfee1, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	vsfee, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	_, tx1 := makeTxInitialwithFees(value1, txfee1, vsfee)
	txhash1, err := tx1.TxHash()
	if err != nil {
		t.Fatal(err)
	}
	mustAddTx(t, hndlr, tx1, 1)
	value2, err := new(uint256.Uint256).FromUint64(1234567)
	if err != nil {
		t.Fatal(err)
	}
	txfee2, err := new(uint256.Uint256).FromUint64(100)
	if err != nil {
		t.Fatal(err)
	}
	_, tx2 := makeTxInitialwithFees(value2, txfee2, vsfee)
	txhash2, err := tx2.TxHash()
	if err != nil {
		t.Fatal(err)
	}
	mustAddTx(t, hndlr, tx2, 1)
	// trie must contain utxos getting spent but must not contain
	// utxos being generated
	utxoIDs, err := objs.TxVec{tx1, tx2}.ConsumedUTXOID()
	if err != nil {
		t.Fatal(err)
	}
	for _, ut := range utxoIDs {
		trie.Add(ut)
	}
	hndlr.txqueue.ClearTxQueue()
	hndlr.txqueue.SetQueueSize(1)

	// Attempt to add but force a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	err = hndlr.AddTxsToQueue(hndlr.db.NewTransaction(false), ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if hndlr.TxQueueAddFinished() {
		t.Fatal("StatusFinished should be false")
	}
	if hndlr.iterInfo.currentKey == nil {
		t.Fatal("currentKey should not be nil")
	}
	// Attempt to add more txs
	err = hndlr.AddTxsToQueue(hndlr.db.NewTransaction(false), context.TODO(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if !hndlr.TxQueueAddFinished() {
		t.Fatal("StatusFinished should still be true")
	}
	if hndlr.txqueue.Contains(txhash1) {
		t.Fatal("TxQueue should not contain tx1")
	}
	if !hndlr.txqueue.Contains(txhash2) {
		t.Fatal("TxQueue should contain tx2")
	}
}
