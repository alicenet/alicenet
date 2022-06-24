package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
)

func TestASOwnerSig(t *testing.T) {
	cid := uint32(1)
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	priOwner := &crypto.Secp256k1Signer{}
	if err := priOwner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	priPubk, err := priOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	priOwnerAcct := crypto.GetAccount(priPubk)

	altOwner := &crypto.Secp256k1Signer{}
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
	exp := uint32(5)
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

	if !bytes.Equal(asp2.Owner.PrimaryOwner.Account, priOwnerAcct) {
		t.Fatal("priAccount")
	}
	asp2OnrPA, err := asp2.Owner.PrimaryAccount()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(asp2OnrPA, priOwnerAcct) {
		t.Fatal("priAccount")
	}
	asp2OnrAA, err := asp2.Owner.AlternateAccount()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(asp2OnrAA, altOwnerAcct) {
		t.Fatal("altAccount")
	}

	aspoBytes, err := asp.Owner.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	aspoBytesBad := []byte{}
	aspoBytesBad = append(aspoBytesBad, aspoBytes[0:len(aspoBytes)-1]...)
	aspo := &AtomicSwapOwner{}
	err = aspo.UnmarshalBinary(aspoBytesBad)
	if err == nil {
		t.Fatal("Should have failed on account length")
	}
	_, err = asp.PreHash()
	if err != nil {
		t.Fatal(err)
	}

	aspo = &AtomicSwapOwner{}
	err = aspo.UnmarshalBinary(aspoBytes)
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("bar")
	sig, err := aspo.SignAsPrimary(msg, priOwner, hashKey)
	if err != nil {
		t.Fatal(err)
	}
	err = aspo.ValidateSignature(msg, sig, true)
	if err != nil {
		t.Fatal(err)
	}
	err = aspo.ValidateSignature(msg, sig, false)
	if err == nil {
		t.Fatal("Should have failed")
	}

	sig, err = aspo.SignAsPrimary(msg, priOwner, hashKey)
	if err != nil {
		t.Fatal(err)
	}
	err = aspo.ValidateSignature(msg, sig, true)
	if err != nil {
		t.Fatal(err)
	}
	sigBytes, err := sig.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	sig2 := &AtomicSwapSignature{}
	err = sig2.UnmarshalBinary(sigBytes)
	if err != nil {
		t.Fatal(err)
	}
	err = aspo.ValidateSignature(msg, sig2, true)
	if err != nil {
		t.Fatal(err)
	}

	sig, err = aspo.SignAsAlternate(msg, altOwner, hashKey)
	if err != nil {
		t.Fatal(err)
	}
	err = aspo.ValidateSignature(msg, sig, false)
	if err != nil {
		t.Fatal(err)
	}
	err = aspo.ValidateSignature(msg, sig, true)
	if err == nil {
		t.Fatal("Should have failed")
	}

	sig, err = aspo.SignAsAlternate(msg, priOwner, hashKey)
	if err != nil {
		t.Fatal(err)
	}
	err = aspo.ValidateSignature(msg, sig, false)
	if err == nil {
		t.Fatal("Should have failed")
	}
	err = aspo.ValidateSignature(msg, sig, true)
	if err == nil {
		t.Fatal("Should have failed")
	}

	sig, err = aspo.SignAsPrimary(msg, altOwner, hashKey)
	if err != nil {
		t.Fatal(err)
	}
	err = aspo.ValidateSignature(msg, sig, false)
	if err == nil {
		t.Fatal("Should have failed")
	}
	err = aspo.ValidateSignature(msg, sig, true)
	if err == nil {
		t.Fatal("Should have failed")
	}

	hashKeyBad := crypto.Hasher([]byte("bad"))
	_, err = aspo.SignAsAlternate(msg, altOwner, hashKeyBad)
	if err == nil {
		t.Fatal("should have failed")
	}
	_, err = aspo.SignAsPrimary(msg, priOwner, hashKeyBad)
	if err == nil {
		t.Fatal("should have failed")
	}

	// while testing I noticed capn allows trailing zero truncation
	// this test ensures that even if those zeros are truncated by some-one
	// the zeros will be replaced before hashing for sig verification.
	// this does mean that all hashes of capn objects must use the capn
	// encoder to be safe!
	aspb, err := asp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	aspb2 := utils.CopySlice(aspb)[0 : len(aspb)-5]

	asp00 := &ASPreImage{}
	err = asp00.UnmarshalBinary(aspb)
	if err != nil {
		t.Fatal(err)
	}
	aspb00, err := asp00.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	asp01 := &ASPreImage{}
	err = asp01.UnmarshalBinary(aspb2)
	if err != nil {
		t.Fatal(err)
	}
	aspb01, err := asp01.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(aspb00, aspb01) {
		t.Fatal("mutable")
	}

}

func TestASOwnerValidateBad1(t *testing.T) {
	aso := &AtomicSwapOwner{}
	err := aso.Validate()
	// Should raise error for invalid SVA
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	aso.SVA = HashedTimelockSVA
	err = aso.Validate()
	// Should raise error for invalid Hashlock
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	aso.HashLock = crypto.Hasher([]byte{})
	err = aso.Validate()
	// Should raise error for nil AlternateOwner
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	aso.AlternateOwner = &AtomicSwapSubOwner{}
	err = aso.Validate()
	// Should raise error for invalid AlternateOwner
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	aso.AlternateOwner = &AtomicSwapSubOwner{
		CurveSpec: constants.CurveSecp256k1,
		Account:   make([]byte, 20),
	}
	err = aso.Validate()
	// Should raise error for nil PrimaryOwner
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	aso.PrimaryOwner = &AtomicSwapSubOwner{}
	err = aso.Validate()
	// Should raise error for invalid PrimaryOwner
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}

	aso.PrimaryOwner = &AtomicSwapSubOwner{
		CurveSpec: constants.CurveSecp256k1,
		Account:   make([]byte, 20),
	}
	err = aso.Validate()
	// Should pass
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestASOwnerMarshalBinary(t *testing.T) {
	asp := &ASPreImage{}
	_, err := asp.Owner.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	aso := &AtomicSwapOwner{}
	_, err = aso.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	hashkey := make([]byte, 32)
	aso.SVA = HashedTimelockSVA
	aso.HashLock = crypto.Hasher(hashkey)
	priOwnerAcct := make([]byte, constants.OwnerLen)
	priOwn := &AtomicSwapSubOwner{
		CurveSpec: constants.CurveSecp256k1,
		Account:   priOwnerAcct,
	}
	altOwnerAcct := make([]byte, constants.OwnerLen)
	altOwn := &AtomicSwapSubOwner{
		CurveSpec: constants.CurveSecp256k1,
		Account:   altOwnerAcct,
	}
	aso.PrimaryOwner = priOwn
	aso.AlternateOwner = altOwn
	_, err = aso.MarshalBinary()
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestASOwnerUnmarshalBinary(t *testing.T) {
	aso := &AtomicSwapOwner{}
	data := make([]byte, 0)
	err := aso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	data = make([]byte, 1)
	err = aso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	data = make([]byte, 1)
	data[0] = uint8(HashedTimelockSVA)
	err = aso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	data = make([]byte, 1+32)
	data[0] = uint8(HashedTimelockSVA)
	err = aso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	data = make([]byte, 1+32+1+constants.OwnerLen)
	data[0] = uint8(HashedTimelockSVA)
	data[33] = uint8(constants.CurveSecp256k1)
	err = aso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}

	data = make([]byte, 1+32+1+constants.OwnerLen+1+constants.OwnerLen+1)
	data[0] = uint8(HashedTimelockSVA)
	data[33] = uint8(constants.CurveSecp256k1)
	data[54] = uint8(constants.CurveSecp256k1)
	err = aso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error (6)")
	}

	data = make([]byte, 1+32+1+constants.OwnerLen+1+constants.OwnerLen)
	data[0] = uint8(HashedTimelockSVA)
	data[33] = uint8(constants.CurveSecp256k1)
	data[54] = uint8(constants.CurveSecp256k1)
	err = aso.UnmarshalBinary(data)
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestASOwnerAccounts(t *testing.T) {
	aso := &AtomicSwapOwner{}
	_, err := aso.PrimaryAccount()
	if err == nil {
		t.Fatal("Should raise error (1)")
	}

	_, err = aso.AlternateAccount()
	if err == nil {
		t.Fatal("Should raise error (2)")
	}
}

func TestASOwnerNew(t *testing.T) {
	asp := &ASPreImage{}
	priAcct := make([]byte, 0)
	altAcct := make([]byte, 0)
	hashKey := make([]byte, 0)
	err := asp.Owner.New(priAcct, altAcct, hashKey)
	if err == nil {
		t.Fatal("Should raise error (1)")
	}

	aso := &AtomicSwapOwner{}
	err = aso.New(priAcct, altAcct, hashKey)
	if err == nil {
		t.Fatal("Should raise error (2)")
	}

	hashKey = make([]byte, constants.HashLen)
	err = aso.New(priAcct, altAcct, hashKey)
	if err == nil {
		t.Fatal("Should raise error (3)")
	}

	priAcct = make([]byte, constants.OwnerLen)
	priAcct[0] = 1
	err = aso.New(priAcct, altAcct, hashKey)
	if err == nil {
		t.Fatal("Should raise error (4)")
	}

	altAcct = make([]byte, constants.OwnerLen)
	altAcct[0] = 2
	err = aso.New(priAcct, altAcct, hashKey)
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestASOwnerNewFromOwner(t *testing.T) {
	asp := &ASPreImage{}
	priOwner := &Owner{}
	altOwner := &Owner{}
	hashKey := make([]byte, 0)
	err := asp.Owner.NewFromOwner(priOwner, altOwner, hashKey)
	if err == nil {
		t.Fatal("Should raise error (1)")
	}

	aso := &AtomicSwapOwner{}
	err = aso.NewFromOwner(priOwner, altOwner, hashKey)
	if err == nil {
		t.Fatal("Should raise error (2)")
	}

	hashKey = make([]byte, constants.HashLen)
	err = aso.NewFromOwner(priOwner, altOwner, hashKey)
	if err == nil {
		t.Fatal("Should raise error (3)")
	}

	priAcct := make([]byte, constants.OwnerLen)
	priAcct[0] = 1
	err = priOwner.New(priAcct, constants.CurveSecp256k1)
	if err != nil {
		t.Fatal(err)
	}
	err = aso.NewFromOwner(priOwner, altOwner, hashKey)
	if err == nil {
		t.Fatal("Should raise error (4)")
	}

	altAcct := make([]byte, constants.OwnerLen)
	altAcct[0] = 2
	err = altOwner.New(altAcct, constants.CurveSecp256k1)
	if err != nil {
		t.Fatal(err)
	}
	err = aso.NewFromOwner(priOwner, altOwner, hashKey)
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestASSubOwnerMarshalBinary(t *testing.T) {
	aso := &AtomicSwapOwner{}
	_, err := aso.PrimaryOwner.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	asso := &AtomicSwapSubOwner{}
	_, err = asso.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	asso.CurveSpec = constants.CurveSecp256k1
	_, err = asso.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	asso.Account = make([]byte, constants.OwnerLen)
	_, err = asso.MarshalBinary()
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestASSubOwnerUnmarshalBinary(t *testing.T) {
	asso := &AtomicSwapSubOwner{}

	data := make([]byte, 0)
	_, err := asso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	data = make([]byte, 1)
	data[0] = uint8(constants.CurveSecp256k1)
	_, err = asso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	data = make([]byte, 1+constants.OwnerLen)
	data[0] = uint8(constants.CurveSecp256k1)
	_, err = asso.UnmarshalBinary(data)
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestASSubOwnerNewFromOwner(t *testing.T) {
	o := &Owner{}
	asso := &AtomicSwapSubOwner{}
	err := asso.NewFromOwner(o)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	curveSpec := constants.CurveBN256Eth
	acct := make([]byte, constants.OwnerLen)
	err = o.New(acct, curveSpec)
	if err != nil {
		t.Fatal(err)
	}
	err = asso.NewFromOwner(o)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	curveSpec = constants.CurveSecp256k1
	err = o.New(acct, curveSpec)
	if err != nil {
		t.Fatal(err)
	}
	err = asso.NewFromOwner(o)
	if err != nil {
		t.Fatal(err)
	}
}
