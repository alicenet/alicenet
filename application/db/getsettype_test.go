package db

import (
	"strconv"
	"testing"

	"github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/constants/dbprefix"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/internal/testing/environment"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
)

func TestSetAndGetUTXO_Success(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)
	chainID := uint32(1)
	value, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	fee, err := new(uint256.Uint256).FromUint64(0)
	if err != nil {
		t.Fatal(err)
	}
	acct := make([]byte, constants.OwnerLen)
	curveSpec := constants.CurveSecp256k1
	txHash := make([]byte, constants.HashLen)
	utxo := &objs.TXOut{}

	err = utxo.CreateValueStore(chainID, value, fee, acct, curveSpec, txHash)
	if err != nil {
		t.Fatal(err)
	}

	key := dbprefix.PrefixDeposit()
	utxoId, err := utxo.UTXOID()
	if err != nil {
		t.Fatal(err)
	}

	key = append(key, utils.CopySlice(utxoId)...)

	err = db.Update(func(txn *badger.Txn) error {
		err := SetUTXO(txn, key, utxo)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	err = db.View(func(txn *badger.Txn) error {
		result, err := GetUTXO(txn, key)
		if err != nil {
			t.Fatal(err)
		}

		if result == nil {
			t.Fatal("Invalid result")
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestSetAndGetUTXO_Error(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	utxo := &objs.TXOut{}
	key := make([]byte, 0)
	err := db.Update(func(txn *badger.Txn) error {
		err := SetUTXO(txn, key, utxo)
		if err == nil {
			t.Fatal("Should have raised error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.View(func(txn *badger.Txn) error {
		_, err := GetUTXO(txn, key)
		if err == nil {
			t.Fatal("Should have raised error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetAndGetTx_Success(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	consumedUTXOs := objs.Vout{}
	consumedUTXO, vs := makeVS(t, ownerSigner, 1)
	consumedUTXOs = append(consumedUTXOs, consumedUTXO)

	txsIn, err := consumedUTXOs.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	generatedUTXOs := objs.Vout{}
	generatedUTXO, _ := makeVS(t, ownerSigner, 0)
	generatedUTXOs = append(generatedUTXOs, generatedUTXO)

	err = generatedUTXOs.SetTxOutIdx()
	if err != nil {
		t.Fatal(err)
	}

	tx := &objs.Tx{
		Vin:  txsIn,
		Vout: generatedUTXOs,
		Fee:  uint256.Zero(),
	}

	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}

	err = vs.Sign(tx.Vin[0], ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	key := dbprefix.PrefixDeposit()
	utxoId, err := consumedUTXO.UTXOID()
	if err != nil {
		t.Fatal(err)
	}

	key = append(key, utils.CopySlice(utxoId)...)

	err = db.Update(func(txn *badger.Txn) error {
		err := SetTx(txn, key, tx)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	err = db.View(func(txn *badger.Txn) error {
		result, err := GetTx(txn, key)
		if err != nil {
			t.Fatal(err)
		}

		if result == nil {
			t.Fatal("Invalid result")
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestGetTx_Error(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	key := make([]byte, 0)
	err := db.View(func(txn *badger.Txn) error {
		_, err := GetUTXO(txn, key)
		if err == nil {
			t.Fatal("Should have raised error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func makeVS(t *testing.T, ownerSigner objs.Signer, i int) (*objs.TXOut, *objs.ValueStore) {
	t.Helper()
	cid := uint32(2)
	val := uint256.One()

	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &objs.ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	fee := new(uint256.Uint256)
	vsp := &objs.VSPreImage{
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
	vs := &objs.ValueStore{
		VSPreImage: vsp,
		TxHash:     txHash,
	}
	vs2 := &objs.ValueStore{}
	vsBytes, err := vs.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = vs2.UnmarshalBinary(vsBytes)
	if err != nil {
		t.Fatal(err)
	}
	utxInputs := &objs.TXOut{}
	err = utxInputs.NewValueStore(vs)
	if err != nil {
		t.Fatal(err)
	}
	return utxInputs, vs
}
