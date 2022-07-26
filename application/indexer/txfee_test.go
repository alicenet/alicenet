package indexer

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

func makeTxFeeIndex() *TxFeeIndex {
	prefix1 := func() []byte {
		return []byte("ze")
	}
	prefix2 := func() []byte {
		return []byte("zf")
	}
	index := NewTxFeeIndex(prefix1, prefix2)
	return index
}

func TestTxFeeIndexAddGood(t *testing.T) {
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

	index := makeTxFeeIndex()
	feeCostRatio := uint256.One()
	txHash := crypto.Hasher([]byte("txHash"))
	err = db.Update(func(txn *badger.Txn) error {
		err = index.Add(txn, feeCostRatio, txHash)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTxFeeIndexAddBad(t *testing.T) {
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

	index := makeTxFeeIndex()
	txHash := crypto.Hasher([]byte("txHash"))
	err = db.Update(func(txn *badger.Txn) error {
		err = index.Add(txn, nil, txHash)
		if err == nil {
			t.Fatal("Should have raised error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTxFeeIndexDrop(t *testing.T) {
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

	index := makeTxFeeIndex()
	feeCostRatio := uint256.One()
	txHash := crypto.Hasher([]byte("txHash"))

	err = db.Update(func(txn *badger.Txn) error {
		err := index.Drop(txn, txHash)
		if err == nil {
			t.Fatal("Should have raised error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		err := index.Add(txn, feeCostRatio, txHash)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Drop(txn, txHash)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTxFeeIndexMakeKey(t *testing.T) {
	index := makeTxFeeIndex()
	txHash := crypto.Hasher([]byte("txHash"))
	feeRatioBytes := make([]byte, 32)
	for k := 0; k < len(feeRatioBytes); k++ {
		feeRatioBytes[k] = byte(k)
	}
	costBytes := make([]byte, 32)
	for k := 0; k < len(costBytes); k++ {
		costBytes[k] = byte(k + 128)
	}

	trueKey := []byte{}
	trueKey = append(trueKey, index.prefix()...)
	trueKey = append(trueKey, feeRatioBytes...)
	trueKey = append(trueKey, txHash...)
	tfiKey := index.makeKey(feeRatioBytes, txHash)
	key := tfiKey.MarshalBinary()
	if !bytes.Equal(key, trueKey) {
		t.Fatal("keys do not match")
	}
}

func TestTxFeeIndexMakeRefKey(t *testing.T) {
	index := makeTxFeeIndex()
	txHash := crypto.Hasher([]byte("txHash"))

	trueKey := []byte{}
	trueKey = append(trueKey, index.refPrefix()...)
	trueKey = append(trueKey, txHash...)
	tfiKey := index.makeRefKey(txHash)
	key := tfiKey.MarshalBinary()
	if !bytes.Equal(key, trueKey) {
		t.Fatal("keys do not match")
	}
}

func TestTxFeeIndexNewIter(t *testing.T) {
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

	index := makeTxFeeIndex()
	_, prefix := index.NewIter(db.NewTransaction(false))
	if !bytes.Equal(prefix, index.prefix()) {
		t.Fatal("Invalid returned prefix")
	}
}

func TestTxFeeIndexIteration1(t *testing.T) {
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

	index := makeTxFeeIndex()
	// Begin adding txs to index
	fcr1 := uint256.One()
	txhash1 := crypto.Hasher([]byte("txHash1"))
	fcr2 := uint256.Two()
	txhash2 := crypto.Hasher([]byte("txHash2"))
	fcr3, _ := new(uint256.Uint256).FromUint64(3)
	txhash3 := crypto.Hasher([]byte("txHash3"))
	fcr4, _ := new(uint256.Uint256).FromUint64(4)
	txhash4 := crypto.Hasher([]byte("txHash4"))

	err = db.Update(func(txn *badger.Txn) error {
		err = index.Add(txn, fcr1, txhash1)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Add(txn, fcr3, txhash3)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Add(txn, fcr2, txhash2)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Add(txn, fcr4, txhash4)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	txhashList := [][]byte{}
	err = db.View(func(txn *badger.Txn) error {
		it, prefix := index.NewIter(txn)
		err = func() error {
			defer it.Close()
			startKey := append(utils.CopySlice(prefix), []byte{255, 255, 255, 255, 255}...)
			for it.Seek(startKey); it.ValidForPrefix(prefix); it.Next() {
				itm := it.Item()
				vBytes, err := itm.ValueCopy(nil)
				if err != nil {
					return err
				}
				txhash := vBytes[len(prefix):]
				txhashList = append(txhashList, utils.CopySlice(txhash))
			}
			return nil
		}()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(txhashList) != 4 {
		t.Logf("Length: %v\n", len(txhashList))
		t.Fatal("invalid txhash list")
	}
	if !bytes.Equal(txhashList[0], txhash4) {
		t.Fatal("Did not return txhash4 as highest")
	}
	if !bytes.Equal(txhashList[1], txhash3) {
		t.Fatal("Did not return txhash3 as 2nd highest")
	}
	if !bytes.Equal(txhashList[2], txhash2) {
		t.Fatal("Did not return txhash2 as 3rd highest")
	}
	if !bytes.Equal(txhashList[3], txhash1) {
		t.Fatal("Did not return txhash1 as lowest")
	}
}

func TestTxFeeIndexIteration2(t *testing.T) {
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

	index := makeTxFeeIndex()
	// Begin adding txs to index
	fcr1 := uint256.One()
	txhash1 := crypto.Hasher([]byte("txHash1"))
	fcr2 := uint256.Two()
	txhash2 := crypto.Hasher([]byte("txHash2"))
	fcr3, _ := new(uint256.Uint256).FromUint64(3)
	txhash3 := crypto.Hasher([]byte("txHash3"))
	fcr4, _ := new(uint256.Uint256).FromUint64(4)
	txhash4 := crypto.Hasher([]byte("txHash4"))

	err = db.Update(func(txn *badger.Txn) error {
		err = index.Add(txn, fcr1, txhash1)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Add(txn, fcr3, txhash3)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Add(txn, fcr2, txhash2)
		if err != nil {
			t.Fatal(err)
		}
		err = index.Add(txn, fcr4, txhash4)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// We iterate one key and store a new start key before continuing
	newStartKey := []byte{}
	err = db.View(func(txn *badger.Txn) error {
		it, prefix := index.NewIter(txn)
		err = func() error {
			defer it.Close()
			startKey := append(utils.CopySlice(prefix), []byte{255, 255, 255, 255, 255}...)
			// Go to first valid key
			it.Seek(startKey)
			// Skip to next one
			it.Next()
			itm := it.Item()
			// Save key to begin iterating next time
			newStartKey = itm.KeyCopy(nil)
			return nil
		}()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	txhashList := [][]byte{}
	err = db.View(func(txn *badger.Txn) error {
		it, prefix := index.NewIter(txn)
		err = func() error {
			defer it.Close()
			// We continue iterating where we left off
			for it.Seek(newStartKey); it.ValidForPrefix(prefix); it.Next() {
				itm := it.Item()
				vBytes, err := itm.ValueCopy(nil)
				if err != nil {
					return err
				}
				txhash := vBytes[len(prefix):]
				txhashList = append(txhashList, utils.CopySlice(txhash))
			}
			return nil
		}()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Because we skipped the first key, there should only be 3
	if len(txhashList) != 3 {
		t.Logf("Length: %v\n", len(txhashList))
		t.Fatal("invalid txhash list")
	}
	if !bytes.Equal(txhashList[0], txhash3) {
		t.Fatal("Did not return txhash3 as highest")
	}
	if !bytes.Equal(txhashList[1], txhash2) {
		t.Fatal("Did not return txhash2 as 2rd highest")
	}
	if !bytes.Equal(txhashList[2], txhash1) {
		t.Fatal("Did not return txhash1 as lowest")
	}
}
