package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
)

func TestVSPreImageGood(t *testing.T) {
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
	vsp2 := &VSPreImage{}
	vspBytes, err := vsp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = vsp2.UnmarshalBinary(vspBytes)
	if err != nil {
		t.Fatal(err)
	}
	vspiEqual(t, vsp, vsp2)
}

func vspiEqual(t *testing.T, vspi1, vspi2 *VSPreImage) {
	if vspi1.ChainID != vspi2.ChainID {
		t.Fatal("Do not agree on ChainID!")
	}
	if !vspi1.Value.Eq(vspi2.Value) {
		t.Fatal("Do not agree on Next!")
	}
	if vspi1.TXOutIdx != vspi2.TXOutIdx {
		t.Fatal("Do not agree on TXOutIdx!")
	}
	if !bytes.Equal(vspi1.Owner.Account, vspi2.Owner.Account) {
		t.Fatal("Do not agree on Index!")
	}
}

func TestVSPreImageBad1(t *testing.T) {
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
	vsp2 := &VSPreImage{}
	vspBytes, err := vsp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = vsp2.UnmarshalBinary(vspBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid ChainID!")
	}
}

func TestVSPreImageMarshalBinary(t *testing.T) {
	vs := &ValueStore{}
	_, err := vs.VSPreImage.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	vsp := &VSPreImage{}
	_, err = vsp.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestVSPreImageUnmarshalBinary(t *testing.T) {
	data := make([]byte, 0)
	vs := &ValueStore{}
	err := vs.VSPreImage.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	vsp := &VSPreImage{}
	err = vsp.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestVSPreImagePreHash(t *testing.T) {
	vs := &ValueStore{}
	_, err := vs.VSPreImage.PreHash()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	vsp := &VSPreImage{}
	_, err = vsp.PreHash()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	// preHash is present; should not fail
	vspGoodPH := &VSPreImage{}
	vspGoodPH.preHash = make([]byte, 32)
	vspGoodPHOut, err := vspGoodPH.PreHash()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(vspGoodPHOut, vspGoodPH.preHash) {
		t.Fatal("PreHashes do not match (1)")
	}

	// Make new
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

	vsp = &VSPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	out, err := vsp.PreHash()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, vsp.preHash) {
		t.Fatal("PreHashes do not match (2)")
	}
}

func TestVSPreImageValidate(t *testing.T) {
	msg := make([]byte, 0)
	sig := &ValueStoreSignature{}
	vs := &ValueStore{}
	err := vs.VSPreImage.ValidateSignature(msg, sig)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	vsp := &VSPreImage{}
	err = vsp.ValidateSignature(msg, sig)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}
