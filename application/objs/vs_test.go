package objs

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/stretchr/testify/assert"
)

func TestValueStoreGood(t *testing.T) {
	cid := uint32(2)
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	vsp := &VSPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	txHash := make([]byte, constants.HashLen)
	vs := &ValueStore{
		VSPreImage: vsp,
		TxHash:     txHash,
	}
	vs2 := &ValueStore{}
	vsBytes, err := vs.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = vs2.UnmarshalBinary(vsBytes)
	if err != nil {
		t.Fatal(err)
	}
	vsEqual(t, vs, vs2)
}

func vsEqual(t *testing.T, vs1, vs2 *ValueStore) {
	vspi1 := vs1.VSPreImage
	vspi2 := vs2.VSPreImage
	vspiEqual(t, vspi1, vspi2)
	if !bytes.Equal(vs1.TxHash, vs2.TxHash) {
		t.Fatal("Do not agree on TxHash!")
	}
}

func TestValueStoreBad1(t *testing.T) {
	cid := uint32(0) // Invalid ChainID
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	vsp := &VSPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	txHash := make([]byte, constants.HashLen)
	vs := &ValueStore{
		VSPreImage: vsp,
		TxHash:     txHash,
	}
	vs2 := &ValueStore{}
	vsBytes, err := vs.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = vs2.UnmarshalBinary(vsBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid VSPreImage!")
	}
}

func TestValueStoreBad2(t *testing.T) {
	cid := uint32(2)
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	vsp := &VSPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	txHash := make([]byte, constants.HashLen+1) // Invalid TxHash
	vs := &ValueStore{
		VSPreImage: vsp,
		TxHash:     txHash,
	}
	vs2 := &ValueStore{}
	vsBytes, err := vs.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = vs2.UnmarshalBinary(vsBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid TxHash: incorrect byte length!")
	}
}

func TestValueStoreNew(t *testing.T) {
	utxo := &TXOut{}
	chainID := uint32(0)
	value, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	fee, err := new(uint256.Uint256).FromUint64(0)
	if err != nil {
		t.Fatal(err)
	}
	acct := make([]byte, 0)
	curveSpec := constants.CurveSecp256k1
	txHash := make([]byte, 0)
	err = utxo.valueStore.New(chainID, value, fee, acct, curveSpec, txHash)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	vs := &ValueStore{}
	err = vs.New(chainID, value, fee, acct, curveSpec, txHash)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	acct = make([]byte, constants.OwnerLen)
	err = vs.New(chainID, value, fee, acct, curveSpec, txHash)
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	chainID = 1
	err = vs.New(chainID, value, fee, acct, curveSpec, txHash)
	if err == nil {
		t.Fatal("Should raise an error (4)")
	}

	txHash = make([]byte, constants.HashLen)
	err = vs.New(chainID, value, fee, acct, curveSpec, txHash)
	if err != nil {
		t.Fatal(err)
	}

	err = vs.New(chainID, nil, fee, acct, curveSpec, txHash)
	if err == nil {
		t.Fatal("Should raise an error (5)")
	}

	err = vs.New(chainID, uint256.Zero(), fee, acct, curveSpec, txHash)
	if err == nil {
		t.Fatal("Should raise an error (6)")
	}

	err = vs.New(chainID, value, nil, acct, curveSpec, txHash)
	if err == nil {
		t.Fatal("Should raise an error (7)")
	}
}

func TestValueStoreNewFromDeposit(t *testing.T) {
	utxo := &TXOut{}
	chainID := uint32(0)
	value, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	acct := make([]byte, 0)
	nonce := make([]byte, 0)
	err = utxo.valueStore.NewFromDeposit(chainID, value, acct, nonce)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	acct = make([]byte, constants.OwnerLen)
	vs := &ValueStore{}
	err = vs.NewFromDeposit(chainID, value, acct, nonce)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	chainID = 1
	err = vs.NewFromDeposit(chainID, value, acct, nonce)
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	nonce = make([]byte, constants.HashLen)
	err = vs.NewFromDeposit(chainID, value, acct, nonce)
	if err != nil {
		t.Fatal(err)
	}
}

func TestValueStoreMarshalBinary(t *testing.T) {
	utxo := &TXOut{}
	_, err := utxo.valueStore.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	_, err = utxo.valueStore.MarshalCapn(nil)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
	vs := &ValueStore{}
	_, err = vs.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}
	_, err = vs.MarshalCapn(nil)
	if err == nil {
		t.Fatal("Should raise an error (4)")
	}

	cid := uint32(2)
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	vsp := &VSPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	txHash := make([]byte, constants.HashLen)
	vs = &ValueStore{
		VSPreImage: vsp,
		TxHash:     txHash,
	}
	vsBytes, err := vs.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	vs2 := &ValueStore{}
	err = vs2.UnmarshalBinary(vsBytes)
	if err != nil {
		t.Fatal(err)
	}
	vsEqual(t, vs, vs2)
}

func TestValueStoreUnmarshalBinary(t *testing.T) {
	data := make([]byte, 0)
	utxo := &TXOut{}
	err := utxo.valueStore.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	vs := &ValueStore{}
	err = vs.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	cid := uint32(2)
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	vsp := &VSPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	txHash := make([]byte, constants.HashLen)
	vs = &ValueStore{
		VSPreImage: vsp,
		TxHash:     txHash,
	}
	vsBytes, err := vs.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	vs2 := &ValueStore{}
	err = vs2.UnmarshalBinary(vsBytes)
	if err != nil {
		t.Fatal(err)
	}
	vsEqual(t, vs, vs2)
}

func TestValueStorePreHash(t *testing.T) {
	utxo := &TXOut{}
	_, err := utxo.valueStore.PreHash()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	vs := &ValueStore{}
	_, err = vs.PreHash()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestValueStoreUTXOID(t *testing.T) {
	utxo := &TXOut{}
	_, err := utxo.valueStore.UTXOID()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	vs := &ValueStore{}
	_, err = vs.UTXOID()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	cid := uint32(2)
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	vsp := &VSPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
	}
	vs = &ValueStore{
		VSPreImage: vsp,
		TxHash:     nil,
	}
	_, err = vs.UTXOID()
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	txHash := make([]byte, constants.HashLen)
	vs.TxHash = txHash
	utxoID, err := vs.UTXOID()
	if err != nil {
		t.Fatal(err)
	}
	out, err := vs.UTXOID()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, utxoID) {
		t.Fatal("utxoIDs do not match")
	}
}

func TestValueStoreTxOutIdx(t *testing.T) {
	utxo := &TXOut{}
	_, err := utxo.valueStore.TxOutIdx()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	vs := &ValueStore{}
	_, err = vs.TxOutIdx()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
	vs.VSPreImage = &VSPreImage{}
	txOutIdx := uint32(17)
	vs.VSPreImage.TXOutIdx = txOutIdx
	out, err := vs.TxOutIdx()
	if err != nil {
		t.Fatal(err)
	}
	if out != txOutIdx {
		t.Fatal("TxOutIndices do not match")
	}
}

func TestValueStoreSetTxOutIdx(t *testing.T) {
	idx := uint32(0)
	utxo := &TXOut{}
	err := utxo.valueStore.SetTxOutIdx(idx)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	vs := &ValueStore{}
	err = vs.SetTxOutIdx(idx)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
	vs.VSPreImage = &VSPreImage{}
	err = vs.SetTxOutIdx(idx)
	if err != nil {
		t.Fatal(err)
	}
	out, err := vs.TxOutIdx()
	if err != nil {
		t.Fatal(err)
	}
	if out != idx {
		t.Fatal("TXOutIdxes do not match")
	}
}

func TestValueStoreSetTxHash(t *testing.T) {
	vs := &ValueStore{}
	txHash := make([]byte, 0)
	err := vs.SetTxHash(txHash)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	vs.VSPreImage = &VSPreImage{}
	err = vs.SetTxHash(txHash)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	txHash = make([]byte, constants.HashLen)
	err = vs.SetTxHash(txHash)
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestValueStoreChainID(t *testing.T) {
	vs := &ValueStore{}
	_, err := vs.ChainID()
	if err == nil {
		t.Fatal("Should raise an error")
	}
	vs.VSPreImage = &VSPreImage{}
	cid := uint32(17)
	vs.VSPreImage.ChainID = cid
	chainID, err := vs.ChainID()
	if err != nil {
		t.Fatal(err)
	}
	if cid != chainID {
		t.Fatal("ChainIDs do not match")
	}
}

func TestValueStoreValue(t *testing.T) {
	vs := &ValueStore{}
	_, err := vs.Value()
	if err == nil {
		t.Fatal("Should raise an error")
	}
	vs.VSPreImage = &VSPreImage{}
	val, err := new(uint256.Uint256).FromUint64(1234567890)
	if err != nil {
		t.Fatal(err)
	}
	vs.VSPreImage.Value = val.Clone()
	value, err := vs.Value()
	if err != nil {
		t.Fatal(err)
	}
	if !value.Eq(val) {
		t.Fatal("Values do not match")
	}
}

func TestValueStoreValuePlusFeeGood(t *testing.T) {
	// Prepare ValueStore
	vs := &ValueStore{}
	vs.VSPreImage = &VSPreImage{}
	value32 := uint32(1234567890)
	value, err := new(uint256.Uint256).FromUint64(uint64(value32))
	if err != nil {
		t.Fatal(err)
	}
	vs.VSPreImage.Value = value.Clone()
	vs.VSPreImage.Fee = new(uint256.Uint256)

	// Check valuePlusFee
	valuePlusFee, err := vs.ValuePlusFee()
	if err != nil {
		t.Fatal(err)
	}
	if !value.Eq(valuePlusFee) {
		t.Fatal("value and valuePlusFee do not match")
	}

	// Prepare storage with nonzero fee
	valueStoreFee := uint32(1000)
	vsFee, err := new(uint256.Uint256).FromUint64(uint64(valueStoreFee))
	if err != nil {
		t.Fatal(err)
	}
	vs.VSPreImage.Fee = vsFee.Clone()
	trueVPF, err := new(uint256.Uint256).Add(vsFee, value)
	if err != nil {
		t.Fatal(err)
	}

	// Check valuePlusFee
	valuePlusFee, err = vs.ValuePlusFee()
	if err != nil {
		t.Fatal(err)
	}
	if !trueVPF.Eq(valuePlusFee) {
		t.Fatal("valuePlusFee is not correct")
	}
	if value.Eq(valuePlusFee) {
		t.Fatal("value and valuePlusFee should not be equal")
	}
}

func TestValueStoreValuePlusFeeBad1(t *testing.T) {
	vs := &ValueStore{}
	_, err := vs.ValuePlusFee()
	if err == nil {
		t.Fatal("Should raise an error")
	}
}

func TestValueStoreValuePlusFeeBad2(t *testing.T) {
	vs := &ValueStore{}
	vs.VSPreImage = &VSPreImage{}

	// Cause overflow when computing valuePlusFee
	bigString := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	bigBig, ok := new(big.Int).SetString(bigString, 16)
	if !ok {
		t.Fatal("Invalid conversion")
	}
	bigUint256, err := new(uint256.Uint256).FromBigInt(bigBig)
	if err != nil {
		t.Fatal(err)
	}

	// Prep for failure
	vs.VSPreImage.Value = bigUint256.Clone()
	_, err = vs.ValuePlusFee()
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestValueStoreIsDeposit(t *testing.T) {
	vs := &ValueStore{}
	val := vs.IsDeposit()
	if val {
		t.Fatal("Should be false (1)")
	}
	vs.VSPreImage = &VSPreImage{}
	txOutIdx := uint32(17)
	vs.VSPreImage.TXOutIdx = txOutIdx
	val = vs.IsDeposit()
	if val {
		t.Fatal("Should be false (2)")
	}
	vs.VSPreImage.TXOutIdx = constants.MaxUint32
	val = vs.IsDeposit()
	if !val {
		t.Fatal("Should be true")
	}
}

func TestValueStoreOwnerCall(t *testing.T) {
	utxo := &TXOut{}
	_, err := utxo.valueStore.Owner()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	vs := &ValueStore{}
	_, err = vs.Owner()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	vs.VSPreImage = &VSPreImage{}
	vs.VSPreImage.Owner = &ValueStoreOwner{}
	_, err = vs.Owner()
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	curveSpec := constants.CurveSecp256k1
	acct := make([]byte, constants.OwnerLen)
	owner := &Owner{}
	err = owner.New(acct, curveSpec)
	if err != nil {
		t.Fatal(err)
	}
	err = vs.VSPreImage.Owner.NewFromOwner(owner)
	if err != nil {
		t.Fatal(err)
	}
	_, err = vs.Owner()
	if err != nil {
		t.Fatal(err)
	}
}

func TestValueStoreGenericOwner(t *testing.T) {
	utxo := &TXOut{}
	_, err := utxo.valueStore.GenericOwner()
	if err == nil {
		t.Fatal("Should raise an error")
	}

	vs := &ValueStore{}
	vs.VSPreImage = &VSPreImage{}
	vs.VSPreImage.Owner = &ValueStoreOwner{}
	curveSpec := constants.CurveSecp256k1
	acct := make([]byte, constants.OwnerLen)
	owner := &Owner{}
	err = owner.New(acct, curveSpec)
	if err != nil {
		t.Fatal(err)
	}
	err = vs.VSPreImage.Owner.NewFromOwner(owner)
	if err != nil {
		t.Fatal(err)
	}
	_, err = vs.GenericOwner()
	if err != nil {
		t.Fatal(err)
	}
}

func TestValueStoreSign(t *testing.T) {
	txIn := &TXIn{}
	signer := &crypto.Secp256k1Signer{}
	utxo := &TXOut{}
	err := utxo.valueStore.Sign(txIn, signer)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	vs := &ValueStore{}
	err = vs.Sign(txIn, signer)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	vs.VSPreImage = &VSPreImage{}
	txOutIdx := uint32(1)
	vs.VSPreImage.TXOutIdx = txOutIdx
	cid := uint32(1)
	vs.VSPreImage.ChainID = cid
	vsTxHash := make([]byte, constants.HashLen)
	err = vs.SetTxHash(vsTxHash)
	if err != nil {
		t.Fatal(err)
	}
	txIn, err = vs.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	err = utxo.valueStore.Sign(txIn, signer)
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}
	err = vs.Sign(txIn, signer)
	if err == nil {
		t.Fatal("Should raise an error (4)")
	}

	acctBad := make([]byte, constants.OwnerLen)
	onrBad := &Owner{}
	err = onrBad.New(acctBad, constants.CurveSecp256k1)
	assert.Nil(t, err)
	if err := onrBad.Validate(); err != nil {
		t.Fatal(err)
	}
	vsoBad := &ValueStoreOwner{}
	if err := vsoBad.NewFromOwner(onrBad); err != nil {
		t.Fatal(err)
	}
	vs.VSPreImage.Owner = vsoBad
	err = vs.Sign(txIn, signer)
	if err == nil {
		t.Fatal("Should raise an error (5)")
	}

	privk := make([]byte, 32)
	privk[0] = 1
	privk[31] = 1
	if err := signer.SetPrivk(privk); err != nil {
		t.Fatal(err)
	}
	pk, err := signer.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	acct := crypto.GetAccount(pk)
	onr := &Owner{}
	err = onr.New(acct, constants.CurveSecp256k1)
	assert.Nil(t, err)
	if err := onr.Validate(); err != nil {
		t.Fatal(err)
	}
	vso := &ValueStoreOwner{}
	if err := vso.NewFromOwner(onr); err != nil {
		t.Fatal(err)
	}
	vs.VSPreImage.Owner = vso
	err = vs.Sign(txIn, signer)
	if err != nil {
		t.Fatal(err)
	}
}

func TestValueStoreValidateSignature(t *testing.T) {
	txIn := &TXIn{}
	utxo := &TXOut{}
	err := utxo.valueStore.ValidateSignature(txIn)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	vs := &ValueStore{}
	err = vs.ValidateSignature(txIn)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	vs.VSPreImage = &VSPreImage{}
	txOutIdx := uint32(1)
	vs.VSPreImage.TXOutIdx = txOutIdx
	cid := uint32(1)
	vs.VSPreImage.ChainID = cid
	vsTxHash := make([]byte, constants.HashLen)
	err = vs.SetTxHash(vsTxHash)
	if err != nil {
		t.Fatal(err)
	}
	txIn, err = vs.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	err = vs.ValidateSignature(txIn)
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	signer := &crypto.Secp256k1Signer{}
	privk := make([]byte, 32)
	privk[0] = 1
	privk[31] = 1
	if err := signer.SetPrivk(privk); err != nil {
		t.Fatal(err)
	}
	pk, err := signer.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	acct := crypto.GetAccount(pk)
	onr := &Owner{}
	err = onr.New(acct, constants.CurveSecp256k1)
	assert.Nil(t, err)
	if err := onr.Validate(); err != nil {
		t.Fatal(err)
	}
	vso := &ValueStoreOwner{}
	if err := vso.NewFromOwner(onr); err != nil {
		t.Fatal(err)
	}
	vs.VSPreImage.Owner = vso
	err = vs.Sign(txIn, signer)
	if err != nil {
		t.Fatal(err)
	}
	err = vs.ValidateSignature(txIn)
	if err != nil {
		t.Fatal(err)
	}
}

func TestValueStoreMakeTxIn(t *testing.T) {
	utxo := &TXOut{}
	_, err := utxo.valueStore.MakeTxIn()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	vs := &ValueStore{}
	_, err = vs.MakeTxIn()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
	vs.VSPreImage = &VSPreImage{}
	txOutIdx := uint32(1)
	vs.VSPreImage.TXOutIdx = txOutIdx
	_, err = vs.MakeTxIn()
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}
	cid := uint32(1)
	vs.VSPreImage.ChainID = cid
	_, err = vs.MakeTxIn()
	if err == nil {
		t.Fatal("Should raise an error (4)")
	}
	vsTxHash := make([]byte, constants.HashLen)
	err = vs.SetTxHash(vsTxHash)
	if err != nil {
		t.Fatal(err)
	}
	_, err = vs.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
}

func TestValueStoreValidateFee(t *testing.T) {
	msg := MakeMockStorageGetter()
	storage := MakeStorage(msg)

	utxo := &TXOut{}
	err := utxo.valueStore.ValidateFee(storage)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	vs := &ValueStore{}
	err = vs.ValidateFee(storage)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	vs.VSPreImage = &VSPreImage{}
	vs.VSPreImage.Fee = new(uint256.Uint256).SetZero()
	err = vs.ValidateFee(storage)
	if err != nil {
		t.Fatal(err)
	}

	vsFee := big.NewInt(1)
	msg.SetValueStoreFee(vsFee)
	storage = MakeStorage(msg)
	err = vs.ValidateFee(storage)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	// Now tests for deposit
	vs.VSPreImage.TXOutIdx = constants.MaxUint32
	err = vs.ValidateFee(storage)
	if err != nil {
		t.Fatal(err)
	}

	vs.VSPreImage.Fee.SetOne()
	err = vs.ValidateFee(storage)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
}

func TestValueStoreOwnerSig(t *testing.T) {
	cid := uint32(2)
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	vsp := &VSPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	txHash := make([]byte, constants.HashLen)
	vs := &ValueStore{
		VSPreImage: vsp,
		TxHash:     txHash,
	}
	vs2 := &ValueStore{}
	vsBytes, err := vs.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = vs2.UnmarshalBinary(vsBytes)
	if err != nil {
		t.Fatal(err)
	}
	vsEqual(t, vs, vs2)

	{
		txIn, err := vs.MakeTxIn()
		if err != nil {
			t.Fatal(err)
		}
		txIn.TXInLinker.TxHash = crypto.Hasher([]byte("hsh"))
		err = vs.Sign(txIn, ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
		err = vs.ValidateSignature(txIn)
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		txIn, err := vs.MakeTxIn()
		if err != nil {
			t.Fatal(err)
		}
		txIn.TXInLinker.TxHash = crypto.Hasher([]byte("hsh"))
		err = vs.Sign(txIn, ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
		txIn.TXInLinker.TxHash = crypto.Hasher([]byte("hshs"))
		err = vs.ValidateSignature(txIn)
		if err == nil {
			t.Fatal("Should raise an error")
		}
	}
}

func TestValueStoreOwnerSig2(t *testing.T) {
	cid := uint32(2)
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	ownerSigner := &crypto.BNSigner{}
	err = ownerSigner.SetPrivk(crypto.Hasher([]byte("a")))
	if err != nil {
		t.Fatal(err)
	}
	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveBN256Eth)

	vsp := &VSPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	txHash := make([]byte, constants.HashLen)
	vs := &ValueStore{
		VSPreImage: vsp,
		TxHash:     txHash,
	}
	vs2 := &ValueStore{}
	vsBytes, err := vs.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = vs2.UnmarshalBinary(vsBytes)
	if err != nil {
		t.Fatal(err)
	}
	vsEqual(t, vs, vs2)

	{
		txIn, err := vs.MakeTxIn()
		if err != nil {
			t.Fatal(err)
		}
		txIn.TXInLinker.TxHash = crypto.Hasher([]byte("hsh"))
		err = vs.Sign(txIn, ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
		err = vs.ValidateSignature(txIn)
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		txIn, err := vs.MakeTxIn()
		if err != nil {
			t.Fatal(err)
		}
		txIn.TXInLinker.TxHash = crypto.Hasher([]byte("hsh"))
		err = vs.Sign(txIn, ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
		txIn.TXInLinker.TxHash = crypto.Hasher([]byte("hshs"))
		err = vs.ValidateSignature(txIn)
		if err == nil {
			t.Fatal("Should raise an error")
		}
	}
}
