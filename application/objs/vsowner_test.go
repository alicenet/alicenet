package objs

import (
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
)

func TestVSOwnerMarshalBinary(t *testing.T) {
	vsp := &VSPreImage{}
	_, err := vsp.Owner.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	vso := &ValueStoreOwner{}
	_, err = vso.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	vso.SVA = ValueStoreSVA
	_, err = vso.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	vso.CurveSpec = constants.CurveSecp256k1
	_, err = vso.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	vso.Account = make([]byte, constants.OwnerLen)
	_, err = vso.MarshalBinary()
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestVSOwnerUnmarshalBinary(t *testing.T) {
	data := make([]byte, 0)
	vsp := &VSPreImage{}
	err := vsp.Owner.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	vso := &ValueStoreOwner{}
	err = vso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	data = make([]byte, 1)
	data[0] = uint8(ValueStoreSVA)
	err = vso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	data = make([]byte, 2)
	data[0] = uint8(ValueStoreSVA)
	data[1] = uint8(constants.CurveSecp256k1)
	err = vso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	data = make([]byte, 23)
	data[0] = uint8(ValueStoreSVA)
	data[1] = uint8(constants.CurveSecp256k1)
	err = vso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (4)")
	}

	data = make([]byte, 2+constants.OwnerLen)
	data[1] = uint8(constants.CurveSecp256k1)
	err = vso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (5)")
	}

	data = make([]byte, 2+constants.OwnerLen)
	data[0] = uint8(ValueStoreSVA)
	data[1] = uint8(constants.CurveSecp256k1)
	err = vso.UnmarshalBinary(data)
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestVSOwnerNewFromOwner(t *testing.T) {
	o := &Owner{}
	vsp := &VSPreImage{}
	err := vsp.Owner.NewFromOwner(o)
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	vso := &ValueStoreOwner{}
	err = vso.NewFromOwner(o)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	curveSpec := constants.CurveSpec(255) // bad curveSpec
	acct := make([]byte, constants.OwnerLen)
	err = o.New(acct, curveSpec)
	if err != nil {
		t.Fatal(err)
	}
	err = vso.NewFromOwner(o)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	curveSpec = constants.CurveSecp256k1
	err = o.New(acct, curveSpec)
	if err != nil {
		t.Fatal(err)
	}
	err = vso.NewFromOwner(o)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVSOwnerValidate(t *testing.T) {
	vsp := &VSPreImage{}
	err := vsp.Owner.Validate()
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}
	err = vsp.Owner.validateSVA()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	err = vsp.Owner.validateAccount()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	err = vsp.Owner.validateCurveSpec()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	vso := &ValueStoreOwner{}
	err = vso.Validate()
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	vso.SVA = ValueStoreSVA
	err = vso.Validate()
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}

	vso.CurveSpec = constants.CurveSecp256k1
	err = vso.Validate()
	if err == nil {
		t.Fatal("Should have raised error (6)")
	}

	vso.Account = make([]byte, constants.OwnerLen)
	err = vso.Validate()
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestVSOwnerSign(t *testing.T) {
	vso := &ValueStoreOwner{}
	msg := make([]byte, 0)
	signerSecp := &crypto.Secp256k1Signer{}
	_, err := vso.Sign(msg, signerSecp)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	signerBN := &crypto.BNSigner{}
	_, err = vso.Sign(msg, signerBN)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestVSOwnerValidateSignatureSecp(t *testing.T) {
	msg := make([]byte, 32)
	vss := &ValueStoreSignature{}
	vsp := &VSPreImage{}
	err := vsp.Owner.ValidateSignature(msg, vss)
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	vso := &ValueStoreOwner{}
	err = vso.ValidateSignature(msg, vss)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	acct := make([]byte, constants.OwnerLen)
	vso.New(acct, constants.CurveSecp256k1)
	err = vso.ValidateSignature(msg, vss)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	secpSigner := &crypto.Secp256k1Signer{}
	privk := make([]byte, 32)
	privk[0] = 1
	privk[31] = 1
	if err := secpSigner.SetPrivk(privk); err != nil {
		t.Fatal(err)
	}
	bnSigner := &crypto.BNSigner{}
	err = bnSigner.SetPrivk(privk)
	if err != nil {
		t.Fatal(err)
	}
	vssBN, err := vso.Sign(msg, bnSigner)
	if err != nil {
		t.Fatal(err)
	}
	err = vso.ValidateSignature(msg, vssBN)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	msgBad := make([]byte, len(msg)+1)
	vssBad, err := vso.Sign(msgBad, secpSigner)
	if err != nil {
		t.Fatal(err)
	}
	err = vso.ValidateSignature(msg, vssBad)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	// Correct vso
	pk, err := secpSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	vso.Account = crypto.GetAccount(pk)
	vss, err = vso.Sign(msg, secpSigner)
	if err != nil {
		t.Fatal(err)
	}
	err = vso.ValidateSignature(msg, vss)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVSOwnerValidateSignatureBN(t *testing.T) {
	msg := make([]byte, 32)
	vss := &ValueStoreSignature{}
	vsp := &VSPreImage{}
	err := vsp.Owner.ValidateSignature(msg, vss)
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	vso := &ValueStoreOwner{}
	err = vso.ValidateSignature(msg, vss)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	acct := make([]byte, constants.OwnerLen)
	vso.New(acct, constants.CurveBN256Eth)
	err = vso.ValidateSignature(msg, vss)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	bnSigner := &crypto.BNSigner{}
	privk := make([]byte, 32)
	privk[0] = 1
	privk[31] = 1
	err = bnSigner.SetPrivk(privk)
	if err != nil {
		t.Fatal(err)
	}
	secpSigner := &crypto.Secp256k1Signer{}
	if err := secpSigner.SetPrivk(privk); err != nil {
		t.Fatal(err)
	}
	vssSecp, err := vso.Sign(msg, secpSigner)
	if err != nil {
		t.Fatal(err)
	}
	err = vso.ValidateSignature(msg, vssSecp)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	msgBad := make([]byte, len(msg)+1)
	vssBad, err := vso.Sign(msgBad, bnSigner)
	if err != nil {
		t.Fatal(err)
	}
	// Sig validation will fail
	err = vso.ValidateSignature(msg, vssBad)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
	// accounts will not match
	err = vso.ValidateSignature(msgBad, vssBad)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}

	// Correct vso
	pk, err := bnSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	vso.Account = crypto.GetAccount(pk)
	vss, err = vso.Sign(msg, bnSigner)
	if err != nil {
		t.Fatal(err)
	}
	err = vso.ValidateSignature(msg, vss)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVSSignatureMarshalBinary(t *testing.T) {
	vss := &ValueStoreSignature{}
	_, err := vss.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	vss.SVA = ValueStoreSVA
	_, err = vss.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	vss.CurveSpec = constants.CurveSecp256k1
	_, err = vss.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	vss.Signature = make([]byte, constants.CurveSecp256k1SigLen)
	_, err = vss.MarshalBinary()
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestVSSignatureUnmarshalBinary(t *testing.T) {
	vss := &ValueStoreSignature{}
	signature := make([]byte, 0)
	err := vss.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	signature = make([]byte, 1)
	signature[0] = uint8(ValueStoreSVA)
	err = vss.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	signature = make([]byte, 2)
	signature[0] = uint8(ValueStoreSVA)
	signature[1] = uint8(constants.CurveSecp256k1)
	err = vss.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	signature = make([]byte, 2)
	signature[0] = uint8(ValueStoreSVA)
	signature[1] = uint8(constants.CurveBN256Eth)
	err = vss.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	signature = make([]byte, 2)
	signature[0] = uint8(ValueStoreSVA)
	err = vss.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}

	signature = make([]byte, 2+constants.CurveSecp256k1SigLen+1)
	signature[0] = uint8(ValueStoreSVA)
	signature[1] = uint8(constants.CurveSecp256k1)
	err = vss.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised error (6)")
	}

	signature = make([]byte, 2+constants.CurveSecp256k1SigLen)
	signature[0] = uint8(ValueStoreSVA)
	signature[1] = uint8(constants.CurveSecp256k1)
	err = vss.UnmarshalBinary(signature)
	if err != nil {
		t.Fatal("Should pass")
	}
}
