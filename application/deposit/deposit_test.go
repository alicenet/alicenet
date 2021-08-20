package deposit

import (
	"bytes"
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	"github.com/MadBase/MadNet/constants/dbprefix"

	"github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
)

const (
	testingChainID uint32 = 100
)

func newDepositHandler() *Handler {
	dh := &Handler{}
	dh.Init()
	return dh
}

func testingOwner() *objs.Owner {
	signer := &crypto.BNSigner{}
	signer.SetPrivk([]byte("secret"))
	pubk, _ := signer.Pubkey()
	acct := crypto.GetAccount(pubk)
	owner := &objs.Owner{}
	curveSpec := constants.CurveSecp256k1
	err := owner.New(acct, curveSpec)
	if err != nil {
		panic(err)
	}
	return owner
}

type mockSpender struct {
	spent map[[constants.HashLen]byte]bool
}

func (ms *mockSpender) isSpent(txn *badger.Txn, utxoID []byte) (bool, error) {
	var hsh [constants.HashLen]byte
	copy(hsh[:], utxoID)
	return ms.spent[hsh], nil
}

func (ms *mockSpender) spend(utxoID []byte) {
	var hsh [constants.HashLen]byte
	copy(hsh[:], utxoID)
	ms.spent[hsh] = true
}

func TestDeposit(t *testing.T) {
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
	mis := &mockSpender{make(map[[constants.HashLen]byte]bool)}
	hndlr := newDepositHandler()
	hndlr.IsSpent = mis.isSpent
	one := new(big.Int).SetInt64(1)
	two := new(big.Int).SetInt64(2)
	three := new(big.Int).SetInt64(3)
	err = db.Update(func(txn *badger.Txn) error {
		err := hndlr.Add(txn, testingChainID, utils.ForceSliceToLength(one.Bytes(), constants.HashLen), one, testingOwner())
		//err := hndlr.Add(txn, testingChainID, utils.ForceSliceToLength(one.Bytes(), constants.HashLen), uint256.One(), testingOwner())
		if err != nil {
			t.Fatal(err)
		}
		err = hndlr.Add(txn, testingChainID, utils.ForceSliceToLength(two.Bytes(), constants.HashLen), one, testingOwner())
		//err = hndlr.Add(txn, testingChainID, utils.ForceSliceToLength(two.Bytes(), constants.HashLen), uint256.One(), testingOwner())
		if err != nil {
			t.Fatal(err)
		}
		found1, missing1, spent1, err := hndlr.Get(txn, [][]byte{utils.ForceSliceToLength(one.Bytes(), constants.HashLen)})
		if err != nil {
			t.Fatal(err)
		}
		if len(missing1) > 0 || len(spent1) > 0 {
			t.Fatal("not okay")
		}
		vs1, err := found1[0].ValueStore()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(vs1.TxHash, utils.ForceSliceToLength(one.Bytes(), constants.HashLen)) {
			t.Fatal("nonce mismatch")
		}
		found2, missing, spent, err := hndlr.Get(txn, [][]byte{utils.ForceSliceToLength(two.Bytes(), constants.HashLen)})
		if err != nil {
			t.Fatal(err)
		}
		if len(missing) > 0 || len(spent) > 0 {
			t.Fatal("not okay")
		}
		vs2, _ := found2[0].ValueStore()
		if !bytes.Equal(vs2.TxHash, utils.ForceSliceToLength(two.Bytes(), constants.HashLen)) {
			t.Fatal("nonce mismatch")
		}
		v, err := vs2.Value()
		if err != nil {
			t.Fatal(err)
		}
		//if v != 1 {
		if !v.Eq(uint256.One()) {
			t.Fatal("not 1", v)
		}
		err = hndlr.Add(txn, testingChainID, utils.ForceSliceToLength(three.Bytes(), constants.HashLen), one, testingOwner())
		//err = hndlr.Add(txn, testingChainID, utils.ForceSliceToLength(three.Bytes(), constants.HashLen), uint256.One(), testingOwner())
		if err != nil {
			t.Fatal(err)
		}
		//utxoIDs, retVal, err := hndlr.GetValueForOwner(txn, testingOwner(), 2)
		utxoIDs, retVal, _, err := hndlr.GetValueForOwner(txn, testingOwner(), uint256.Two(), 256, nil)
		if err != nil {
			t.Fatal(err)
		}
		//if retVal != 2 {
		if !retVal.Eq(uint256.Two()) {
			t.Fatal("bad value", retVal)
		}
		if len(utxoIDs) != 2 {
			t.Fatal("bad len", len(utxoIDs))
		}
		if !bytes.Equal(utxoIDs[0], utils.ForceSliceToLength(one.Bytes(), constants.HashLen)) {
			t.Fatal("bad value")
		}
		if !bytes.Equal(utxoIDs[1], utils.ForceSliceToLength(two.Bytes(), constants.HashLen)) {
			t.Fatal("bad value")
		}
		err = hndlr.Add(txn, testingChainID, utils.ForceSliceToLength(three.Bytes(), constants.HashLen), one, testingOwner())
		//err = hndlr.Add(txn, testingChainID, utils.ForceSliceToLength(three.Bytes(), constants.HashLen), uint256.One(), testingOwner())
		if err == nil {
			t.Fatal("did not fail")
		}
		mis.spend(utils.ForceSliceToLength(two.Bytes(), constants.HashLen))
		//utxoIDs, retVal, err = hndlr.GetValueForOwner(txn, testingOwner(), 2)
		utxoIDs, retVal, _, err = hndlr.GetValueForOwner(txn, testingOwner(), uint256.Two(), 256, nil)
		if err != nil {
			t.Fatal(err)
		}
		//if retVal != 2 {
		if !retVal.Eq(uint256.Two()) {
			t.Fatal("bad value", retVal)
		}
		if len(utxoIDs) != 2 {
			t.Fatal("bad len", len(utxoIDs))
		}
		if !bytes.Equal(utxoIDs[0], utils.ForceSliceToLength(one.Bytes(), constants.HashLen)) {
			t.Fatal("bad value")
		}
		if !bytes.Equal(utxoIDs[1], utils.ForceSliceToLength(three.Bytes(), constants.HashLen)) {
			t.Fatal("bad value")
		}
		err = hndlr.Add(txn, testingChainID, utils.ForceSliceToLength(two.Bytes(), constants.HashLen), one, testingOwner())
		//err = hndlr.Add(txn, testingChainID, utils.ForceSliceToLength(two.Bytes(), constants.HashLen), uint256.One(), testingOwner())
		if err == nil {
			t.Fatal("did not fail")
		}
		err = hndlr.Remove(txn, utils.ForceSliceToLength(one.Bytes(), constants.HashLen))
		if err != nil {
			t.Fatal(err)
		}
		err = hndlr.Remove(txn, utils.ForceSliceToLength(three.Bytes(), constants.HashLen))
		if err != nil {
			t.Fatal(err)
		}
		//utxoIDs, retVal, err = hndlr.GetValueForOwner(txn, testingOwner(), 2)
		utxoIDs, retVal, _, err = hndlr.GetValueForOwner(txn, testingOwner(), uint256.Two(), 256, nil)
		if err != nil {
			t.Fatal(err)
		}
		//if retVal != 0 {
		if !retVal.Eq(uint256.Zero()) {
			t.Fatal("bad value", retVal)
		}
		if len(utxoIDs) != 0 {
			t.Fatal("bad len", len(utxoIDs))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDepositMakeKey(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	////////////////////////////////////////
	hndlr := newDepositHandler()

	utxoID := make([]byte, constants.HashLen)
	trueKey := append(dbprefix.PrefixDeposit(), utxoID...)
	key := hndlr.makeKey(utxoID)
	if !bytes.Equal(key, trueKey) {
		t.Fatal("Key does not match expected")
	}
}

func TestDepositGetInternal(t *testing.T) {
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
	mis := &mockSpender{make(map[[constants.HashLen]byte]bool)}
	hndlr := newDepositHandler()
	hndlr.IsSpent = mis.isSpent
	one := new(big.Int).SetInt64(1)
	two := new(big.Int).SetInt64(2)
	three := new(big.Int).SetInt64(3)
	_ = one
	_ = two
	_ = three
	err = db.Update(func(txn *badger.Txn) error {
		utxoID := utils.ForceSliceToLength(one.Bytes(), constants.HashLen)
		// Check if utxoID present; should fail
		found, missing, spent, err := hndlr.getInternal(txn, utxoID)
		if err != nil {
			t.Fatal(err)
		}
		if found != nil {
			t.Fatal("Should not have found anything (1)")
		}
		if spent != nil {
			t.Fatal("Should not have spent anything (1)")
		}
		if !bytes.Equal(missing, utxoID) {
			t.Fatal("missing should match utxoID")
		}

		// Add utxoID to database; check if present
		err = hndlr.Add(txn, testingChainID, utxoID, one, testingOwner())
		//err = hndlr.Add(txn, testingChainID, utxoID, uint256.One(), testingOwner())
		if err != nil {
			t.Fatal(err)
		}
		found, missing, spent, err = hndlr.getInternal(txn, utxoID)
		if err != nil {
			t.Fatal(err)
		}
		if found == nil {
			t.Fatal("Should have found something (2)")
		}
		if spent != nil {
			t.Fatal("Should not have spent anything (2)")
		}
		if missing != nil {
			t.Fatal("Should not be missing anything (2)")
		}
		vs, err := found.ValueStore()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(vs.TxHash, utxoID) {
			t.Fatal("nonce mismatch (2)")
		}

		// Spend utxo and check that it is spent
		mis.spend(utxoID)
		found, missing, spent, err = hndlr.getInternal(txn, utxoID)
		if err != nil {
			t.Fatal(err)
		}
		if found != nil {
			t.Fatal("Should not have found anything (3)")
		}
		if spent == nil {
			t.Fatal("Should have spent something (3)")
		}
		if missing != nil {
			t.Fatal("Should not be missing anything (3)")
		}
		vs, err = spent.ValueStore()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(vs.TxHash, utxoID) {
			t.Fatal("nonce mismatch (3)")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDepositGet(t *testing.T) {
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
	mis := &mockSpender{make(map[[constants.HashLen]byte]bool)}
	hndlr := newDepositHandler()
	hndlr.IsSpent = mis.isSpent
	one := new(big.Int).SetInt64(1)
	two := new(big.Int).SetInt64(2)
	three := new(big.Int).SetInt64(3)
	_ = one
	_ = two
	_ = three
	err = db.Update(func(txn *badger.Txn) error {
		utxoID := utils.ForceSliceToLength(one.Bytes(), constants.HashLen)
		found, missing, spent, err := hndlr.Get(txn, [][]byte{utxoID})
		if err != nil {
			t.Fatal(err)
		}
		if len(found) != 0 {
			t.Fatal("Should not have found anything (1)")
		}
		if len(spent) != 0 {
			t.Fatal("Should not have spent anything (1)")
		}
		if len(missing) != 1 {
			t.Fatal("Should have 1 missing (1)")
		}
		if !bytes.Equal(missing[0], utxoID) {
			t.Fatal("missing should match utxoID1 (1)")
		}

		err = hndlr.Add(txn, testingChainID, utxoID, one, testingOwner())
		//err = hndlr.Add(txn, testingChainID, utxoID, uint256.One(), testingOwner())
		if err != nil {
			t.Fatal(err)
		}
		found, missing, spent, err = hndlr.Get(txn, [][]byte{utxoID})
		if err != nil {
			t.Fatal(err)
		}
		if len(found) != 1 {
			t.Fatal("Should not have found anything (2)")
		}
		if len(spent) != 0 {
			t.Fatal("Should not have spent anything (2)")
		}
		if len(missing) != 0 {
			t.Fatal("Should have none missing (2)")
		}

		mis.spend(utxoID)
		found, missing, spent, err = hndlr.Get(txn, [][]byte{utxoID})
		if err != nil {
			t.Fatal(err)
		}
		if len(found) != 0 {
			t.Fatal("Should not have found anything (3)")
		}
		if len(spent) != 1 {
			t.Fatal("Should have spent one (3)")
		}
		if len(missing) != 0 {
			t.Fatal("Should have none missing (3)")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDepositAdd(t *testing.T) {
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
	mis := &mockSpender{make(map[[constants.HashLen]byte]bool)}
	hndlr := newDepositHandler()
	hndlr.IsSpent = mis.isSpent
	one := new(big.Int).SetInt64(1)
	two := new(big.Int).SetInt64(2)
	three := new(big.Int).SetInt64(3)
	_ = one
	_ = two
	_ = three
	err = db.Update(func(txn *badger.Txn) error {
		utxoID := utils.ForceSliceToLength(one.Bytes(), constants.HashLen)

		// Raise error for invalid owner
		err = hndlr.Add(txn, testingChainID, utxoID, one, nil)
		//err = hndlr.Add(txn, testingChainID, utxoID, uint256.One(), nil)
		if err == nil {
			t.Fatal("Should have raised error for invalid owner")
		}

		// Add valid UTXO and then confirm it is present
		err = hndlr.Add(txn, testingChainID, utxoID, one, testingOwner())
		//err = hndlr.Add(txn, testingChainID, utxoID, uint256.One(), testingOwner())
		if err != nil {
			t.Fatal(err)
		}
		found, missing, spent, err := hndlr.Get(txn, [][]byte{utxoID})
		if err != nil {
			t.Fatal(err)
		}
		if len(found) != 1 {
			t.Fatal("Should not have found anything (2)")
		}
		if len(spent) != 0 {
			t.Fatal("Should not have spent anything (2)")
		}
		if len(missing) != 0 {
			t.Fatal("Should have none missing (2)")
		}

		// Re-add without changing anything
		err = hndlr.Add(txn, testingChainID, utxoID, one, testingOwner())
		//err = hndlr.Add(txn, testingChainID, utxoID, uint256.One(), testingOwner())
		if err == nil {
			t.Fatal("Should have raised error for being stale")
		}

		// Spend the utxo
		mis.spend(utxoID)

		// Re-add. Should raise an error
		err = hndlr.Add(txn, testingChainID, utxoID, one, testingOwner())
		//err = hndlr.Add(txn, testingChainID, utxoID, uint256.One(), testingOwner())
		if err == nil {
			t.Fatal("Should have raised error as UTXO already spent")
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDepositRemove(t *testing.T) {
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
	mis := &mockSpender{make(map[[constants.HashLen]byte]bool)}
	hndlr := newDepositHandler()
	hndlr.IsSpent = mis.isSpent
	one := new(big.Int).SetInt64(1)
	two := new(big.Int).SetInt64(2)
	three := new(big.Int).SetInt64(3)
	_ = one
	_ = two
	_ = three
	err = db.Update(func(txn *badger.Txn) error {
		utxoID := utils.ForceSliceToLength(one.Bytes(), constants.HashLen)

		// Swallows error if attempt to remove utxoID which is not present
		err = hndlr.Remove(txn, utxoID)
		if err != nil {
			t.Fatal("Should have swallowed error")
		}

		// Add and then Remove utxo. Should not raise an error
		err = hndlr.Add(txn, testingChainID, utxoID, one, testingOwner())
		//err = hndlr.Add(txn, testingChainID, utxoID, uint256.One(), testingOwner())
		if err != nil {
			t.Fatal(err)
		}
		err = hndlr.Remove(txn, utxoID)
		if err != nil {
			t.Fatal()
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDepositGetValueForOwner(t *testing.T) {
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
	mis := &mockSpender{make(map[[constants.HashLen]byte]bool)}
	hndlr := newDepositHandler()
	hndlr.IsSpent = mis.isSpent
	one := new(big.Int).SetInt64(1)
	two := new(big.Int).SetInt64(2)
	three := new(big.Int).SetInt64(3)
	_ = one
	_ = two
	_ = three
	//minValue := uint32(32)
	minValue, err := new(uint256.Uint256).FromUint64(32)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		_, _, _, err := hndlr.GetValueForOwner(txn, nil, minValue, 256, nil)
		if err == nil {
			t.Fatal("Should have raised error for invalid owner")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
