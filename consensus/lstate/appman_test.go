package lstate

import (
	"errors"
	appObjs "github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestAppMan_AddPendingTx_Error1(t *testing.T) {
	engine := initEngine(t, nil)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.AddPendingTx(txn, []interfaces.Transaction{})
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestAppMan_AddPendingTx_Error2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	os.SyncToBH.BClaims.ChainID = 7777
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.AddPendingTx(txn, []interfaces.Transaction{})
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestAppMan_AddPendingTx_Ok(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.AddPendingTx(txn, []interfaces.Transaction{})
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestAppMan_getValidValue_Error1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnState.SyncToBH.BClaims.ChainID = 7777

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		_, _, _, _, err := engine.getValidValue(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestAppMan_getValidValue_Error2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnState.SyncToBH.BClaims.ChainID = 8888

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		_, _, _, _, err := engine.getValidValue(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestAppMan_getValidValue_Ok(t *testing.T) {
	txs := []*appObjs.Tx{makeTx(t)}

	engine := initEngine(t, txs)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		_, _, _, _, err := engine.getValidValue(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestAppMan_isValid_False1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		bh.BClaims.Height = 1
		err := engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		isValid, err := engine.isValid(txn, rss, 1, make([]byte, constants.HashLen), make([]byte, constants.HashLen), []interfaces.Transaction{})
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		assert.False(t, isValid)

		return nil
	})
}

func TestAppMan_isValid_False2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		bh.BClaims.Height = 1
		err := engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		goodHeaderRoot, err := engine.database.GetHeaderTrieRoot(txn, bh.BClaims.Height)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		tx := &appObjs.Tx{
			Fee: nil,
		}
		txs := []interfaces.Transaction{tx}
		isValid, err := engine.isValid(txn, rss, 1, make([]byte, constants.HashLen), goodHeaderRoot, txs)
		if err == nil {
			t.Fatalf("Should have raised error")
		}
		assert.False(t, isValid)

		return nil
	})
}

func TestAppMan_isValid_False3(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		bh.BClaims.Height = 1
		err := engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		goodHeaderRoot, err := engine.database.GetHeaderTrieRoot(txn, bh.BClaims.Height)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		tx := makeTx(t)
		txs := []interfaces.Transaction{tx}
		isValid, err := engine.isValid(txn, rss, 7777, make([]byte, constants.HashLen), goodHeaderRoot, txs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		assert.False(t, isValid)

		return nil
	})
}

func TestAppMan_isValid_False4(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		bh.BClaims.Height = 1
		err := engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		goodHeaderRoot, err := engine.database.GetHeaderTrieRoot(txn, bh.BClaims.Height)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		tx := makeTx(t)
		txs := []interfaces.Transaction{tx}
		isValid, err := engine.isValid(txn, rss, 8888, make([]byte, constants.HashLen), goodHeaderRoot, txs)
		if err == nil {
			t.Fatalf("Should have raised error")
		}
		assert.False(t, isValid)

		return nil
	})
}

func TestAppMan_isValid_False5(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		bh.BClaims.Height = 1
		err := engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		goodHeaderRoot, err := engine.database.GetHeaderTrieRoot(txn, bh.BClaims.Height)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		tx := makeTx(t)
		txs := []interfaces.Transaction{tx}
		isValid, err := engine.isValid(txn, rss, 9999, make([]byte, constants.HashLen), goodHeaderRoot, txs)
		if err == nil {
			t.Fatalf("Should have raised error")
		}
		assert.False(t, isValid)

		return nil
	})
}

func TestAppMan_isValid_True(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		bh.BClaims.Height = 1
		err := engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		goodHeaderRoot, err := engine.database.GetHeaderTrieRoot(txn, bh.BClaims.Height)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		tx := makeTx(t)
		txs := []interfaces.Transaction{tx}
		isValid, err := engine.isValid(txn, rss, 1, make([]byte, constants.HashLen), goodHeaderRoot, txs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		assert.True(t, isValid)

		return nil
	})
}

func TestAppMan_applyState_Error1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.applyState(txn, rss, 1, make([][]byte, 1))
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestAppMan_applyState_Error2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		txHash := make([]byte, constants.HashLen)
		txsHashes := [][]byte{txHash}
		err := engine.applyState(txn, rss, 1, txsHashes)
		if err == nil {
			t.Fatalf("Should have raised error")
		}
		if !errors.As(err, &errorz.ErrMissingTransactions) {
			t.Fatalf("Should have raised errorz.ErrMissingTransactions error")
		}

		return nil
	})
}

func makeTx(t *testing.T) *appObjs.Tx {
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	consumedUTXOs := appObjs.Vout{}
	vss := make([]*appObjs.ValueStore, 0)
	for i := 1; i < 5; i++ {
		consUTXO, vs := makeVS(t, ownerSigner, i)
		consumedUTXOs = append(consumedUTXOs, consUTXO)
		vss = append(vss, vs)
	}
	err := consumedUTXOs.SetTxOutIdx()
	if err != nil {
		t.Fatal(err)
	}

	txInputs := []*appObjs.TXIn{}
	for i := 0; i < 4; i++ {
		txIn, err := consumedUTXOs[i].MakeTxIn()
		if err != nil {
			t.Fatal(err)
		}
		txInputs = append(txInputs, txIn)
	}
	generatedUTXOs := appObjs.Vout{}
	for i := 1; i < 5; i++ {
		genUTXO, _ := makeVS(t, ownerSigner, 0)
		generatedUTXOs = append(generatedUTXOs, genUTXO)
	}
	err = generatedUTXOs.SetTxOutIdx()
	if err != nil {
		t.Fatal(err)
	}
	tx := &appObjs.Tx{
		Vin:  txInputs,
		Vout: generatedUTXOs,
		Fee:  uint256.Zero(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4; i++ {
		err = vss[i].Sign(tx.Vin[i], ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}
	return tx
}

func makeVS(t *testing.T, ownerSigner appObjs.Signer, i int) (*appObjs.TXOut, *appObjs.ValueStore) {
	cid := uint32(2)
	val := uint256.One()

	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &appObjs.ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	fee := new(uint256.Uint256)
	vsp := &appObjs.VSPreImage{
		ChainID: cid,
		Value:   val,
		Owner:   owner,
		Fee:     fee.Clone(),
	}
	var txHash []byte
	if i == 0 {
		txHash = make([]byte, constants.HashLen)
	} else {
		txHash = crypto.Hasher([]byte(strconv.Itoa(i)))
	}
	vs := &appObjs.ValueStore{
		VSPreImage: vsp,
		TxHash:     txHash,
	}
	vs2 := &appObjs.ValueStore{}
	vsBytes, err := vs.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = vs2.UnmarshalBinary(vsBytes)
	if err != nil {
		t.Fatal(err)
	}
	utxInputs := &appObjs.TXOut{}
	err = utxInputs.NewValueStore(vs)
	if err != nil {
		t.Fatal(err)
	}
	return utxInputs, vs
}
