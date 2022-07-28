package txqueue

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/crypto"
)

func TestTxQueueInitGood(t *testing.T) {
	queueSize := 128
	tq := &TxQueue{}
	err := tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTxQueueInitBad(t *testing.T) {
	queueSize := 0
	tq := &TxQueue{}
	err := tq.Init(queueSize)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxQueueQueueSize(t *testing.T) {
	tq := &TxQueue{}
	qs := tq.QueueSize()
	if qs != 0 {
		t.Fatal("invalid QueueSize (1)")
	}
	queueSize := 0
	err := tq.SetQueueSize(queueSize)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	queueSize = 128
	err = tq.SetQueueSize(queueSize)
	if err != nil {
		t.Fatal(err)
	}
	qs = tq.QueueSize()
	if qs != queueSize {
		t.Fatal("invalid QueueSize (2)")
	}
}

func TestTxQueueEmptyFull(t *testing.T) {
	tq := &TxQueue{}
	if !tq.IsEmpty() {
		t.Fatal("Should be empty (1)")
	}
	if !tq.IsFull() {
		t.Fatal("Should be full")
	}
	queueSize := 128
	err := tq.SetQueueSize(queueSize)
	if err != nil {
		t.Fatal(err)
	}
	if !tq.IsEmpty() {
		t.Fatal("Should be empty (2)")
	}
	if tq.IsFull() {
		t.Fatal("Should not be full (1)")
	}

	err = tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}

	txhash := crypto.Hasher([]byte("TxHash1"))
	utxoID11 := crypto.Hasher([]byte("utxoID11"))
	utxoID12 := crypto.Hasher([]byte("utxoID12"))
	utxoIDs := [][]byte{utxoID11, utxoID12}
	value, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	isCleanup := false

	ok, err := tq.Add(txhash, value, utxoIDs, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}

	if tq.IsEmpty() {
		t.Fatal("Should not be empty")
	}
	if tq.IsFull() {
		t.Fatal("Should not be full (2)")
	}
}

func TestTxQueueDropAll(t *testing.T) {
	tq := &TxQueue{}
	err := tq.Init(128)
	if err != nil {
		t.Fatal(err)
	}

	txhash := crypto.Hasher([]byte("TxHash1"))
	utxoID1 := crypto.Hasher([]byte("utxoID1"))
	utxoID2 := crypto.Hasher([]byte("utxoID2"))
	utxoIDs := [][]byte{utxoID1, utxoID2}
	value, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	isCleanup := false
	ok, err := tq.Add(txhash, value, utxoIDs, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if len(tq.txheap) != 1 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 1 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 2 {
		t.Fatal("invalid length of utxoIDs")
	}

	if tq.IsEmpty() {
		t.Fatal("Should not be empty")
	}
}

func TestTxQueueAddBad1(t *testing.T) {
	tq := &TxQueue{}
	_, err := tq.Add(nil, nil, nil, false)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	queueSize := 128
	err = tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tq.Add(nil, nil, nil, false)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	_, err = tq.Add(nil, uint256.Zero(), nil, false)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}

func TestTxQueueAddBad2(t *testing.T) {
	tq := &TxQueue{}
	txhash := crypto.Hasher([]byte("TxHash1"))
	utxoID1 := crypto.Hasher([]byte("utxoID1"))
	utxoID2 := crypto.Hasher([]byte("utxoID2"))
	utxoIDs := [][]byte{utxoID1, utxoID2}
	_, err := tq.Add(txhash, uint256.Zero(), utxoIDs, false)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxQueueAddGood(t *testing.T) {
	tq := &TxQueue{}
	queueSize := 128
	err := tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}

	// Make and add tx
	txhash := crypto.Hasher([]byte("TxHash1"))
	utxoID1 := crypto.Hasher([]byte("utxoID1"))
	utxoID2 := crypto.Hasher([]byte("utxoID2"))
	utxoIDs := [][]byte{utxoID1, utxoID2}
	value, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	isCleanup := false
	ok, err := tq.Add(txhash, value, utxoIDs, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if len(tq.txheap) != 1 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 1 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 2 {
		t.Fatal("invalid length of utxoIDs")
	}

	if !tq.Contains(txhash) {
		t.Fatal("Queue should contain txhash")
	}
	if _, ok := tq.utxoIDs[string(utxoID1)]; !ok {
		t.Fatal("utxoID1 not present")
	}
	if _, ok := tq.utxoIDs[string(utxoID2)]; !ok {
		t.Fatal("utxoID2 not present")
	}

	tq.Drop(txhash)
	if len(tq.txheap) != 0 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 0 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 0 {
		t.Fatal("invalid length of utxoIDs")
	}
	tq.Drop(txhash)
}

// In depth test adding items and dropping them.
// Add item1, item2, item3 (all valid additions);
// Remove item2;
// Add item2 again;
// Make item4 which contains utxoIDs from item1 and item2;
// item4 is mined, so remove all information conflicting with it.
func TestTxQueueAddGood2(t *testing.T) {
	tq := &TxQueue{}
	queueSize := 128
	err := tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}

	// Make and add 3 txs
	txhash1 := crypto.Hasher([]byte("TxHash1"))
	utxoID11 := crypto.Hasher([]byte("utxoID11"))
	utxoID12 := crypto.Hasher([]byte("utxoID12"))
	value1, err := new(uint256.Uint256).FromUint64(19)
	if err != nil {
		t.Fatal(err)
	}
	item1 := &Item{
		txhash:    txhash1,
		value:     value1,
		utxoIDs:   [][]byte{utxoID11, utxoID12},
		isCleanup: false,
	}
	if _, conflict := tq.ConflictingUTXOIDs(item1.utxoIDs); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err := tq.Add(item1.txhash, item1.value, item1.utxoIDs, item1.isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if len(tq.txheap) != 1 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 1 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 2 {
		t.Fatal("invalid length of utxoIDs")
	}

	txhash2 := crypto.Hasher([]byte("TxHash2"))
	utxoID21 := crypto.Hasher([]byte("utxoID21"))
	utxoID22 := crypto.Hasher([]byte("utxoID22"))
	value2, err := new(uint256.Uint256).FromUint64(15)
	if err != nil {
		t.Fatal(err)
	}
	item2 := &Item{
		txhash:    txhash2,
		value:     value2,
		utxoIDs:   [][]byte{utxoID21, utxoID22},
		isCleanup: false,
	}
	if _, conflict := tq.ConflictingUTXOIDs(item2.utxoIDs); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err = tq.Add(item2.txhash, item2.value, item2.utxoIDs, item2.isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if len(tq.txheap) != 2 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 2 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 4 {
		t.Fatal("invalid length of utxoIDs")
	}

	txhash3 := crypto.Hasher([]byte("TxHash3"))
	utxoID31 := crypto.Hasher([]byte("utxoID31"))
	utxoID32 := crypto.Hasher([]byte("utxoID32"))
	value3, err := new(uint256.Uint256).FromUint64(3)
	if err != nil {
		t.Fatal(err)
	}
	item3 := &Item{
		txhash:    txhash3,
		value:     value3,
		utxoIDs:   [][]byte{utxoID31, utxoID32},
		isCleanup: false,
	}
	if _, conflict := tq.ConflictingUTXOIDs(item3.utxoIDs); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err = tq.Add(item3.txhash, item3.value, item3.utxoIDs, item3.isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if len(tq.txheap) != 3 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 3 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 6 {
		t.Fatal("invalid length of utxoIDs")
	}

	// Drop item2; confirm missing
	tq.Drop(item2.txhash)
	if len(tq.txheap) != 2 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 2 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 4 {
		t.Fatal("invalid length of utxoIDs")
	}
	if !tq.Contains(item1.txhash) {
		t.Fatal("does not contain item1")
	}
	if tq.Contains(item2.txhash) {
		t.Fatal("contains item2")
	}
	if !tq.Contains(item3.txhash) {
		t.Fatal("does not contain item3")
	}

	// Add item2 again
	if _, conflict := tq.ConflictingUTXOIDs(item2.utxoIDs); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err = tq.Add(item2.txhash, item2.value, item2.utxoIDs, item2.isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if len(tq.txheap) != 3 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 3 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 6 {
		t.Fatal("invalid length of utxoIDs")
	}
	if !tq.Contains(item1.txhash) {
		t.Fatal("does not contain item1")
	}
	if !tq.Contains(item2.txhash) {
		t.Fatal("does not contain item2")
	}
	if !tq.Contains(item3.txhash) {
		t.Fatal("does not contain item3")
	}

	// These are part of a "new tx" which has been mined
	txhash4 := crypto.Hasher([]byte("TxHash4"))
	value4, err := new(uint256.Uint256).FromUint64(25519)
	if err != nil {
		t.Fatal(err)
	}
	newUtxoIDs := [][]byte{utxoID11, utxoID22}
	item4 := &Item{
		txhash:  txhash4,
		value:   value4,
		utxoIDs: newUtxoIDs,
	}
	// We should not be able to add item4 because of conflicts
	if _, conflict := tq.ConflictingUTXOIDs(item4.utxoIDs); !conflict {
		t.Fatal("Should have conflict")
	}

	// Drop item4; confirm missing.
	// This should kick out item1 and item2.
	tq.DeleteMined(item4.utxoIDs)
	if len(tq.txheap) != 1 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 1 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 2 {
		t.Fatal("invalid length of utxoIDs")
	}
	if tq.Contains(item1.txhash) {
		t.Fatal("contains item1")
	}
	if tq.Contains(item2.txhash) {
		t.Fatal("contains item2")
	}
	if !tq.Contains(item3.txhash) {
		t.Fatal("does not contain item3")
	}
}

func TestTxQueueAddGood3(t *testing.T) {
	tq := &TxQueue{}
	queueSize := 1
	err := tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}

	// Make and add 3 txs
	txhash1 := crypto.Hasher([]byte("TxHash1"))
	utxoID11 := crypto.Hasher([]byte("utxoID11"))
	utxoID12 := crypto.Hasher([]byte("utxoID12"))
	value1, err := new(uint256.Uint256).FromUint64(19)
	if err != nil {
		t.Fatal(err)
	}
	item1 := &Item{
		txhash:    txhash1,
		value:     value1,
		utxoIDs:   [][]byte{utxoID11, utxoID12},
		isCleanup: false,
	}
	ok, err := tq.Add(item1.txhash, item1.value, item1.utxoIDs, item1.isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if len(tq.txheap) != 1 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 1 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 2 {
		t.Fatal("invalid length of utxoIDs")
	}

	// Attempt to add 2; will not be able to
	txhash2 := crypto.Hasher([]byte("TxHash2"))
	utxoID21 := crypto.Hasher([]byte("utxoID21"))
	utxoID22 := crypto.Hasher([]byte("utxoID22"))
	value2, err := new(uint256.Uint256).FromUint64(15)
	if err != nil {
		t.Fatal(err)
	}
	item2 := &Item{
		txhash:    txhash2,
		value:     value2,
		utxoIDs:   [][]byte{utxoID21, utxoID22},
		isCleanup: false,
	}
	ok, err = tq.Add(item2.txhash, item2.value, item2.utxoIDs, item2.isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("should not add")
	}
	if !tq.Contains(item1.txhash) {
		t.Fatal("Should contain txhash1")
	}
	if tq.Contains(item2.txhash) {
		t.Fatal("Should not contain txhash2")
	}

	txhash3 := crypto.Hasher([]byte("TxHash3"))
	utxoID31 := crypto.Hasher([]byte("utxoID31"))
	utxoID32 := crypto.Hasher([]byte("utxoID32"))
	value3, err := new(uint256.Uint256).FromUint64(21)
	if err != nil {
		t.Fatal(err)
	}
	item3 := &Item{
		txhash:    txhash3,
		value:     value3,
		utxoIDs:   [][]byte{utxoID31, utxoID32},
		isCleanup: false,
	}
	ok, err = tq.Add(item3.txhash, item3.value, item3.utxoIDs, item3.isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if !tq.Contains(item3.txhash) {
		t.Fatal("Should contain txhash3")
	}
	if tq.Contains(item1.txhash) {
		t.Fatal("Should not contain txhash1")
	}
	if tq.Contains(item2.txhash) {
		t.Fatal("Should not contain txhash2")
	}

	// Attempt to add tx3 again; should not add again
	ok, err = tq.Add(item3.txhash, item3.value, item3.utxoIDs, item3.isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("did not add again")
	}
}

func TestTxQueueValidAdd1(t *testing.T) {
	tq := &TxQueue{}
	queueSize := 128
	err := tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}

	// Make and add tx
	txhash := crypto.Hasher([]byte("TxHash1"))
	utxoID1 := crypto.Hasher([]byte("utxoID11"))
	utxoID2 := crypto.Hasher([]byte("utxoID12"))
	utxoIDs := [][]byte{utxoID1, utxoID2}
	value, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	isCleanup := false
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err := tq.Add(txhash, value, utxoIDs, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs); !conflict {
		t.Fatal("Should have conflict")
	}
}

func TestTxQueueValidAdd2(t *testing.T) {
	tq := &TxQueue{}
	queueSize := 128
	err := tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}

	// Make and add tx
	txhash := crypto.Hasher([]byte("TxHash1"))
	utxoID1 := crypto.Hasher([]byte("utxoID11"))
	utxoID2 := crypto.Hasher([]byte("utxoID12"))
	utxoIDs := [][]byte{utxoID1, utxoID2}
	value, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	isCleanup := false
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err := tq.Add(txhash, value, utxoIDs, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs); !conflict {
		t.Fatal("Should have conflict")
	}
	if len(tq.txheap) != 1 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 1 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 2 {
		t.Fatal("invalid length of utxoIDs")
	}
	if _, ok := tq.refmap[string(txhash)]; !ok {
		t.Fatal("txhash1 should be present")
	}
	if _, ok := tq.utxoIDs[string(utxoID1)]; !ok {
		t.Fatal("utxoID1 should be present")
	}
	if _, ok := tq.utxoIDs[string(utxoID2)]; !ok {
		t.Fatal("utxoID2 should be present")
	}

	// Add another tx with higher value, kicking out the first tx
	txhash2 := crypto.Hasher([]byte("TxHash2"))
	utxoIDs2 := [][]byte{utxoID1}
	value2, err := new(uint256.Uint256).FromUint64(2)
	if err != nil {
		t.Fatal(err)
	}
	isCleanup2 := false
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs2); !conflict {
		t.Fatal("Should have conflict")
	}
	ok, err = tq.Add(txhash2, value2, utxoIDs2, isCleanup2)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if len(tq.txheap) != 1 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 1 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 1 {
		t.Fatal("invalid length of utxoIDs")
	}
	if _, ok := tq.refmap[string(txhash)]; ok {
		t.Fatal("txhash1 should not be present")
	}
	if _, ok := tq.refmap[string(txhash2)]; !ok {
		t.Fatal("txhash2 should be present")
	}
	if _, ok := tq.utxoIDs[string(utxoID1)]; !ok {
		t.Fatal("utxoID1 should be present")
	}
	if _, ok := tq.utxoIDs[string(utxoID2)]; ok {
		t.Fatal("utxoID2 should not be present")
	}

	// Attempt to add another tx with lower value
	txhash3 := crypto.Hasher([]byte("TxHash3"))
	utxoIDs3 := [][]byte{utxoID1}
	value3, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	isCleanup3 := false
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs3); !conflict {
		t.Fatal("Should have conflict")
	}
	ok, err = tq.Add(txhash3, value3, utxoIDs3, isCleanup3)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("Should not have added")
	}
}

func TestTxQueueValidAdd3(t *testing.T) {
	tq := &TxQueue{}
	queueSize := 128
	err := tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}

	// Make and add tx
	txhash := crypto.Hasher([]byte("TxHash1"))
	utxoID1 := crypto.Hasher([]byte("utxoID11"))
	utxoID2 := crypto.Hasher([]byte("utxoID12"))
	utxoIDs := [][]byte{utxoID1, utxoID2}
	value, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	isCleanup := true
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err := tq.Add(txhash, value, utxoIDs, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs); !conflict {
		t.Fatal("Should have conflict")
	}
	if len(tq.txheap) != 1 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 1 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 2 {
		t.Fatal("invalid length of utxoIDs")
	}
	if _, ok := tq.refmap[string(txhash)]; !ok {
		t.Fatal("txhash1 should be present")
	}
	if _, ok := tq.utxoIDs[string(utxoID1)]; !ok {
		t.Fatal("utxoID1 should be present")
	}
	if _, ok := tq.utxoIDs[string(utxoID2)]; !ok {
		t.Fatal("utxoID2 should be present")
	}

	_, err = tq.PopMax()
	if err != nil {
		t.Fatal(err)
	}
	if len(tq.txheap) != 0 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 0 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 0 {
		t.Fatal("invalid length of utxoIDs")
	}
}

func TestTxQueueValidAdd4(t *testing.T) {
	// In-depth test about adding/popping multiple txs.
	tq := &TxQueue{}
	queueSize := 128
	err := tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}

	// Make and add tx
	txhash1 := crypto.Hasher([]byte("TxHash1"))
	utxoID11 := crypto.Hasher([]byte("utxoID11"))
	utxoID12 := crypto.Hasher([]byte("utxoID12"))
	utxoIDs1 := [][]byte{utxoID11, utxoID12}
	value1, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	isCleanup := false
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs1); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err := tq.Add(txhash1, value1, utxoIDs1, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}

	txhash2 := crypto.Hasher([]byte("TxHash2"))
	utxoID21 := crypto.Hasher([]byte("utxoID21"))
	utxoID22 := crypto.Hasher([]byte("utxoID22"))
	utxoIDs2 := [][]byte{utxoID21, utxoID22}
	value2, err := new(uint256.Uint256).FromUint64(2)
	if err != nil {
		t.Fatal(err)
	}
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs2); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err = tq.Add(txhash2, value2, utxoIDs2, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}

	txhash3 := crypto.Hasher([]byte("TxHash3"))
	utxoID31 := crypto.Hasher([]byte("utxoID31"))
	utxoID32 := crypto.Hasher([]byte("utxoID32"))
	utxoIDs3 := [][]byte{utxoID31, utxoID32}
	value3, err := new(uint256.Uint256).FromUint64(3)
	if err != nil {
		t.Fatal(err)
	}
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs3); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err = tq.Add(txhash3, value3, utxoIDs3, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}

	txhash4 := crypto.Hasher([]byte("TxHash4"))
	utxoID41 := crypto.Hasher([]byte("utxoID41"))
	utxoID42 := crypto.Hasher([]byte("utxoID42"))
	utxoIDs4 := [][]byte{utxoID41, utxoID42}
	value4, err := new(uint256.Uint256).FromUint64(4)
	if err != nil {
		t.Fatal(err)
	}
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs4); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err = tq.Add(txhash4, value4, utxoIDs4, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}

	txhash5 := crypto.Hasher([]byte("TxHash5"))
	utxoID51 := crypto.Hasher([]byte("utxoID51"))
	utxoID52 := crypto.Hasher([]byte("utxoID52"))
	utxoIDs5 := [][]byte{utxoID51, utxoID52}
	value5, err := new(uint256.Uint256).FromUint64(5)
	if err != nil {
		t.Fatal(err)
	}
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs5); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err = tq.Add(txhash5, value5, utxoIDs5, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}

	// Now to remove items from heap.
	retItem, err := tq.PopMax()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retItem.txhash, txhash5) {
		t.Fatal("invalid PopMax item")
	}
	retItem, err = tq.PopMin()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retItem.txhash, txhash1) {
		t.Fatal("invalid PopMin item")
	}
	retItem, err = tq.PopMax()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retItem.txhash, txhash4) {
		t.Fatal("invalid PopMax item")
	}
	retItem, err = tq.PopMin()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retItem.txhash, txhash2) {
		t.Fatal("invalid PopMin item")
	}
	retItem, err = tq.PopMax()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retItem.txhash, txhash3) {
		t.Fatal("invalid PopMax item")
	}
	if !tq.IsEmpty() {
		t.Fatal("Should be empty")
	}
}

func TestTxQueueMinValue(t *testing.T) {
	tq := &TxQueue{}
	queueSize := 128
	err := tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}

	// Make and add tx
	txhash1 := crypto.Hasher([]byte("TxHash1"))
	utxoID11 := crypto.Hasher([]byte("utxoID11"))
	utxoID12 := crypto.Hasher([]byte("utxoID12"))
	utxoIDs1 := [][]byte{utxoID11, utxoID12}
	value1, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	isCleanup := false
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs1); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err := tq.Add(txhash1, value1, utxoIDs1, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}

	txhash2 := crypto.Hasher([]byte("TxHash2"))
	utxoID21 := crypto.Hasher([]byte("utxoID21"))
	utxoID22 := crypto.Hasher([]byte("utxoID22"))
	utxoIDs2 := [][]byte{utxoID21, utxoID22}
	value2, err := new(uint256.Uint256).FromUint64(2)
	if err != nil {
		t.Fatal(err)
	}
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs2); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err = tq.Add(txhash2, value2, utxoIDs2, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}

	txhash3 := crypto.Hasher([]byte("TxHash3"))
	utxoID31 := crypto.Hasher([]byte("utxoID31"))
	utxoID32 := crypto.Hasher([]byte("utxoID32"))
	utxoIDs3 := [][]byte{utxoID31, utxoID32}
	value3, err := new(uint256.Uint256).FromUint64(3)
	if err != nil {
		t.Fatal(err)
	}
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs3); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err = tq.Add(txhash3, value3, utxoIDs3, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}

	txhash4 := crypto.Hasher([]byte("TxHash4"))
	utxoID41 := crypto.Hasher([]byte("utxoID41"))
	utxoID42 := crypto.Hasher([]byte("utxoID42"))
	utxoIDs4 := [][]byte{utxoID41, utxoID42}
	value4, err := new(uint256.Uint256).FromUint64(4)
	if err != nil {
		t.Fatal(err)
	}
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs4); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err = tq.Add(txhash4, value4, utxoIDs4, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}

	txhash5 := crypto.Hasher([]byte("TxHash5"))
	utxoID51 := crypto.Hasher([]byte("utxoID51"))
	utxoID52 := crypto.Hasher([]byte("utxoID52"))
	utxoIDs5 := [][]byte{utxoID51, utxoID52}
	value5, err := new(uint256.Uint256).FromUint64(5)
	if err != nil {
		t.Fatal(err)
	}
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs5); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err = tq.Add(txhash5, value5, utxoIDs5, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}

	trueMinValue := new(uint256.Uint256)
	trueMinValue.Set(value1)
	minValue, err := tq.MinValue()
	if err != nil {
		t.Fatal(err)
	}
	if !minValue.Eq(trueMinValue) {
		t.Fatal("Did not return correct min value")
	}
}

func TestTxQueueMinValueBad(t *testing.T) {
	tq := &TxQueue{}
	queueSize := 128
	err := tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tq.MinValue()
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxQueueAddAboveThreshold(t *testing.T) {
	tq := &TxQueue{}
	queueSize := 2
	err := tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}

	// Make and add tx
	txhash := crypto.Hasher([]byte("TxHash1"))
	utxoID1 := crypto.Hasher([]byte("utxoID11"))
	utxoID2 := crypto.Hasher([]byte("utxoID12"))
	utxoIDs := [][]byte{utxoID1, utxoID2}
	value, err := new(uint256.Uint256).FromUint64(14)
	if err != nil {
		t.Fatal(err)
	}
	isCleanup := false
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs); conflict {
		t.Fatal("Should not have conflict")
	}
	ok, err := tq.Add(txhash, value, utxoIDs, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs); !conflict {
		t.Fatal("Should have conflict")
	}
	if len(tq.txheap) != 1 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 1 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 2 {
		t.Fatal("invalid length of utxoIDs")
	}
	if _, ok := tq.refmap[string(txhash)]; !ok {
		t.Fatal("txhash1 should be present")
	}
	if _, ok := tq.utxoIDs[string(utxoID1)]; !ok {
		t.Fatal("utxoID1 should be present")
	}
	if _, ok := tq.utxoIDs[string(utxoID2)]; !ok {
		t.Fatal("utxoID2 should be present")
	}

	// Add another tx with higher value, kicking out the first tx
	txhash2 := crypto.Hasher([]byte("TxHash2"))
	utxoIDs2 := [][]byte{utxoID1}
	value2, err := new(uint256.Uint256).FromUint64(2)
	if err != nil {
		t.Fatal(err)
	}
	isCleanup2 := false
	if _, conflict := tq.ConflictingUTXOIDs(utxoIDs2); !conflict {
		t.Fatal("Should have conflict")
	}
	ok, err = tq.Add(txhash2, value2, utxoIDs2, isCleanup2)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("should not add")
	}
	if len(tq.txheap) != 1 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 1 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 2 {
		t.Fatal("invalid length of utxoIDs")
	}
	if _, ok := tq.refmap[string(txhash)]; !ok {
		t.Fatal("txhash1 should be present")
	}
	if _, ok := tq.utxoIDs[string(utxoID1)]; !ok {
		t.Fatal("utxoID1 should be present")
	}
	if _, ok := tq.utxoIDs[string(utxoID2)]; !ok {
		t.Fatal("utxoID2 should be present")
	}
}

func TestTxQueuePopMaxBad(t *testing.T) {
	tq := &TxQueue{}
	_, err := tq.PopMax()
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxQueuePopMaxGood(t *testing.T) {
	tq := &TxQueue{}
	queueSize := 128
	err := tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}

	// Make and add tx
	txhash := crypto.Hasher([]byte("TxHash1"))
	utxoID1 := crypto.Hasher([]byte("utxoID11"))
	utxoID2 := crypto.Hasher([]byte("utxoID12"))
	utxoIDs := [][]byte{utxoID1, utxoID2}
	value, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	isCleanup := false
	ok, err := tq.Add(txhash, value, utxoIDs, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if len(tq.txheap) != 1 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 1 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 2 {
		t.Fatal("invalid length of utxoIDs")
	}

	retItem, err := tq.PopMax()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retItem.txhash, txhash) {
		t.Fatal("returned hash value does not match")
	}
}

func TestTxQueuePopMinBad(t *testing.T) {
	tq := &TxQueue{}
	_, err := tq.PopMin()
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxQueueDropMined1(t *testing.T) {
	tq := &TxQueue{}
	queueSize := 128
	err := tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}

	// Make and add tx
	txhash := crypto.Hasher([]byte("TxHash1"))
	utxoID1 := crypto.Hasher([]byte("utxoID1"))
	utxoIDs := [][]byte{utxoID1}
	value, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	isCleanup := false
	ok, err := tq.Add(txhash, value, utxoIDs, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if len(tq.txheap) != 1 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 1 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 1 {
		t.Fatal("invalid length of utxoIDs")
	}

	tq.DeleteMined(utxoIDs)
	if len(tq.txheap) != 0 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 0 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 0 {
		t.Fatal("invalid length of utxoIDs")
	}
}

func TestTxQueueDropMined2(t *testing.T) {
	tq := &TxQueue{}
	queueSize := 128
	err := tq.Init(queueSize)
	if err != nil {
		t.Fatal(err)
	}

	// Make and add tx
	txhash := crypto.Hasher([]byte("TxHash1"))
	utxoID1 := crypto.Hasher([]byte("utxoID11"))
	utxoID2 := crypto.Hasher([]byte("utxoID12"))
	utxoIDs := [][]byte{utxoID1, utxoID2}
	value, err := new(uint256.Uint256).FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	isCleanup := false
	ok, err := tq.Add(txhash, value, utxoIDs, isCleanup)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("did not add")
	}
	if len(tq.txheap) != 1 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 1 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 2 {
		t.Fatal("invalid length of utxoIDs")
	}

	tq.DeleteMined(utxoIDs)
	if len(tq.txheap) != 0 {
		t.Fatal("invalid length of txheap")
	}
	if len(tq.refmap) != 0 {
		t.Fatal("invalid length of refmap")
	}
	if len(tq.utxoIDs) != 0 {
		t.Fatal("invalid length of utxoIDs")
	}
}

func TestTxQueueDropRefernces(t *testing.T) {
	tq := &TxQueue{}
	tq.dropReferences(nil)
}
