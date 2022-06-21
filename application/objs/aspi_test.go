package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/crypto"
)

func TestASPreImageGood(t *testing.T) {
	cid := uint32(2)
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	priOwner := crypto.Secp256k1Signer{}
	if err := priOwner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	priPubk, err := priOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	priOwnerAcct := crypto.GetAccount(priPubk)

	altOwner := crypto.Secp256k1Signer{}
	if err := altOwner.SetPrivk(crypto.Hasher([]byte("b"))); err != nil {
		t.Fatal(err)
	}
	altPubk, err := altOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	altOwnerAcct := crypto.GetAccount(altPubk)

	hashKey := crypto.Hasher([]byte("foo"))
	owner := &AtomicSwapOwner{}
	err = owner.New(priOwnerAcct, altOwnerAcct, hashKey)
	if err != nil {
		t.Fatal(err)
	}

	iat := uint32(1)
	exp := uint32(1234)
	asp := &ASPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		IssuedAt: iat,
		Exp:      exp,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	asp2 := &ASPreImage{}
	aspBytes, err := asp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = asp2.UnmarshalBinary(aspBytes)
	if err != nil {
		t.Fatal(err)
	}
	aspiEqual(t, asp, asp2)
}

func aspiEqual(t *testing.T, aspi1, aspi2 *ASPreImage) {
	if aspi1.ChainID != aspi2.ChainID {
		t.Fatal("Do not agree on ChainID!")
	}
	if !aspi1.Value.Eq(aspi2.Value) {
		t.Fatal("Do not agree on Value!")
	}
	if aspi1.TXOutIdx != aspi2.TXOutIdx {
		t.Fatal("Do not agree on TXOutIdx!")
	}
	a1o, err := aspi1.Owner.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	a2o, err := aspi1.Owner.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a1o, a2o) {
		t.Fatal("Do not agree on Index!")
	}
	if aspi1.IssuedAt != aspi2.IssuedAt {
		t.Fatal("Do not agree on IssuedAt!")
	}
	if aspi1.Exp != aspi2.Exp {
		t.Fatal("Do not agree on Exp!")
	}
}

func TestASPreImageBad1(t *testing.T) {
	cid := uint32(0) // Invalid ChainID
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)
	priOwner := crypto.Secp256k1Signer{}
	if err := priOwner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	priPubk, err := priOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	priOwnerAcct := crypto.GetAccount(priPubk)

	altOwner := crypto.Secp256k1Signer{}
	if err := altOwner.SetPrivk(crypto.Hasher([]byte("b"))); err != nil {
		t.Fatal(err)
	}
	altPubk, err := altOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	altOwnerAcct := crypto.GetAccount(altPubk)

	hashKey := crypto.Hasher([]byte("foo"))
	owner := &AtomicSwapOwner{}
	err = owner.New(priOwnerAcct, altOwnerAcct, hashKey)
	if err != nil {
		t.Fatal(err)
	}

	iat := uint32(1)
	exp := uint32(1234)
	asp := &ASPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		IssuedAt: iat,
		Exp:      exp,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	asp2 := &ASPreImage{}
	aspBytes, err := asp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = asp2.UnmarshalBinary(aspBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid ChainID!")
	}
}

func TestASPreImageBad2(t *testing.T) {
	cid := uint32(1)
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	priOwner := crypto.Secp256k1Signer{}
	if err := priOwner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	priPubk, err := priOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	priOwnerAcct := crypto.GetAccount(priPubk)

	altOwner := crypto.Secp256k1Signer{}
	if err := altOwner.SetPrivk(crypto.Hasher([]byte("b"))); err != nil {
		t.Fatal(err)
	}
	altPubk, err := altOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	altOwnerAcct := crypto.Hasher(altPubk[1:])[13:] // Invalid acct

	hashKey := crypto.Hasher([]byte("foo"))
	owner := &AtomicSwapOwner{}
	err = owner.New(priOwnerAcct, altOwnerAcct, hashKey)
	if err == nil {
		t.Fatal("Should raise an error")
	}

	iat := uint32(1)
	exp := uint32(1234)
	asp := &ASPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		IssuedAt: iat,
		Exp:      exp,
	}
	_, err = asp.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise error for invalid Owner!")
	}
}

func TestASPreImageBad3(t *testing.T) {
	cid := uint32(1)
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	priOwner := crypto.Secp256k1Signer{}
	if err := priOwner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	priPubk, err := priOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	priOwnerAcct := crypto.GetAccount(priPubk)

	altOwner := crypto.Secp256k1Signer{}
	if err := altOwner.SetPrivk(crypto.Hasher([]byte("b"))); err != nil {
		t.Fatal(err)
	}
	altPubk, err := altOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	altOwnerAcct := crypto.GetAccount(altPubk)

	hashKey := crypto.Hasher([]byte("foo"))
	owner := &AtomicSwapOwner{}
	err = owner.New(priOwnerAcct, altOwnerAcct, hashKey)
	if err != nil {
		t.Fatal(err)
	}

	iat := uint32(0) // Invalid IssuedAt
	exp := uint32(1234)
	asp := &ASPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		IssuedAt: iat,
		Exp:      exp,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	asp2 := &ASPreImage{}
	aspBytes, err := asp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = asp2.UnmarshalBinary(aspBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid IssuedAt!")
	}
}

func TestASPreImageBad4(t *testing.T) {
	cid := uint32(1)
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	priOwner := crypto.Secp256k1Signer{}
	if err := priOwner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	priPubk, err := priOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	priOwnerAcct := crypto.GetAccount(priPubk)

	altOwner := crypto.Secp256k1Signer{}
	if err := altOwner.SetPrivk(crypto.Hasher([]byte("b"))); err != nil {
		t.Fatal(err)
	}
	altPubk, err := altOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	altOwnerAcct := crypto.GetAccount(altPubk)

	hashKey := crypto.Hasher([]byte("foo"))
	owner := &AtomicSwapOwner{}
	err = owner.New(priOwnerAcct, altOwnerAcct, hashKey)
	if err != nil {
		t.Fatal(err)
	}

	iat := uint32(1)
	exp := uint32(0) // Invalid Exp
	asp := &ASPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		IssuedAt: iat,
		Exp:      exp,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	asp2 := &ASPreImage{}
	aspBytes, err := asp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = asp2.UnmarshalBinary(aspBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid Exp!")
	}
}

func TestASPreImageBad5(t *testing.T) {
	cid := uint32(1)
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	priOwner := crypto.Secp256k1Signer{}
	if err := priOwner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	priPubk, err := priOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	priOwnerAcct := crypto.GetAccount(priPubk)

	altOwner := crypto.Secp256k1Signer{}
	if err := altOwner.SetPrivk(crypto.Hasher([]byte("b"))); err != nil {
		t.Fatal(err)
	}
	altPubk, err := altOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	altOwnerAcct := crypto.GetAccount(altPubk)

	hashKey := crypto.Hasher([]byte("foo"))
	owner := &AtomicSwapOwner{}
	err = owner.New(priOwnerAcct, altOwnerAcct, hashKey)
	if err != nil {
		t.Fatal(err)
	}

	iat := uint32(1)
	exp := uint32(1) // Invalid IssuedAt and Exp
	asp := &ASPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		IssuedAt: iat,
		Exp:      exp,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	asp2 := &ASPreImage{}
	aspBytes, err := asp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = asp2.UnmarshalBinary(aspBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid IssuedAt and Exp!")
	}
}

func TestASPreImageBad6(t *testing.T) {
	as := &AtomicSwap{}
	_, err := as.ASPreImage.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	currentHeight := uint32(0)
	_, err = as.ASPreImage.IsExpired(currentHeight)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	asp := &ASPreImage{}
	_, err = asp.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	_, err = asp.PreHash()
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

}

func TestASPreImageSigning(t *testing.T) {
	as := &AtomicSwap{}
	msg := make([]byte, 0)
	sig := &AtomicSwapSignature{}
	currentHeight := uint32(0)
	err := as.ASPreImage.ValidateSignature(currentHeight, msg, sig)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	asp := &ASPreImage{}
	signer := &crypto.Secp256k1Signer{}
	hashKey := make([]byte, 0)
	_, err = as.ASPreImage.SignAsPrimary(msg, signer, hashKey)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	_, err = as.ASPreImage.SignAsAlternate(msg, signer, hashKey)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	_, err = asp.SignAsPrimary(msg, signer, hashKey)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	_, err = asp.SignAsAlternate(msg, signer, hashKey)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}
}
