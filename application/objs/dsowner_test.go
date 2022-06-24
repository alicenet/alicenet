package objs

import (
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
)

func TestDSOwnerMarshalBinary(t *testing.T) {
	dsp := &DSPreImage{}
	_, err := dsp.Owner.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	dso := &DataStoreOwner{}
	_, err = dso.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	dso.SVA = DataStoreSVA
	_, err = dso.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	dso.CurveSpec = constants.CurveSecp256k1
	_, err = dso.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (4)")
	}

	dso.Account = make([]byte, constants.OwnerLen)
	_, err = dso.MarshalBinary()
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestDSOwnerUnmarshalBinary(t *testing.T) {
	data := make([]byte, 0)
	dsp := &DSPreImage{}
	err := dsp.Owner.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (0)")
	}

	dso := &DataStoreOwner{}
	err = dso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	data = make([]byte, 1)
	data[0] = uint8(DataStoreSVA)
	err = dso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	data = make([]byte, 2)
	data[0] = uint8(DataStoreSVA)
	data[1] = uint8(constants.CurveSecp256k1)
	err = dso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	data = make([]byte, 2+constants.OwnerLen+1)
	data[0] = uint8(DataStoreSVA)
	data[1] = uint8(constants.CurveSecp256k1)
	err = dso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (4)")
	}

	data = make([]byte, 2+constants.OwnerLen)
	data[1] = uint8(constants.CurveSecp256k1)
	err = dso.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (5)")
	}

	data = make([]byte, 2+constants.OwnerLen)
	data[0] = uint8(DataStoreSVA)
	data[1] = uint8(constants.CurveSecp256k1)
	err = dso.UnmarshalBinary(data)
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestDSOwnerNewFromOwner(t *testing.T) {
	o := &Owner{}
	dsp := &DSPreImage{}
	err := dsp.Owner.NewFromOwner(o)
	if err == nil {
		t.Fatal("Should raise an error (0)")
	}

	dso := &DataStoreOwner{}
	err = dso.NewFromOwner(o)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	curveSpec := constants.CurveSpec(255) // bad curveSpec
	acct := make([]byte, constants.OwnerLen)
	err = o.New(acct, curveSpec)
	if err != nil {
		t.Fatal(err)
	}
	err = dso.NewFromOwner(o)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	curveSpec = constants.CurveSecp256k1
	err = o.New(acct, curveSpec)
	if err != nil {
		t.Fatal(err)
	}
	err = dso.NewFromOwner(o)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDSOwnerValidate(t *testing.T) {
	dsp := &DSPreImage{}
	err := dsp.Owner.Validate()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	dso := &DataStoreOwner{}
	err = dso.Validate()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	dso.SVA = DataStoreSVA
	err = dso.Validate()
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	dso.CurveSpec = constants.CurveSecp256k1
	err = dso.Validate()
	if err == nil {
		t.Fatal("Should raise an error (4)")
	}

	dso.Account = make([]byte, constants.OwnerLen)
	err = dso.Validate()
	if err != nil {
		t.Fatal("Should pass")
	}

	err = dsp.Owner.validateSVA()
	if err == nil {
		t.Fatal("Should raise an error (5)")
	}
	err = dsp.Owner.validateAccount()
	if err == nil {
		t.Fatal("Should raise an error (6)")
	}
	err = dsp.Owner.validateCurveSpec()
	if err == nil {
		t.Fatal("Should raise an error (7)")
	}
}

func TestDSOwnerValidateSignatureSecp(t *testing.T) {
	dsp := &DSPreImage{}
	msg := make([]byte, 0)
	dss := &DataStoreSignature{}
	err := dsp.Owner.ValidateSignature(msg, dss, false)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	dso := &DataStoreOwner{}
	err = dso.ValidateSignature(msg, dss, false)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	// Create valid DataStoreOwner for BN256
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
	acctBad := make([]byte, constants.OwnerLen)
	dso.New(acctBad, constants.CurveSecp256k1)
	err = dso.ValidateSignature(msg, dss, false)
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	dssBN := &DataStoreSignature{}
	dssBN.SVA = DataStoreSVA
	dssBN.CurveSpec = constants.CurveBN256Eth
	dssBN.Signature = make([]byte, constants.CurveBN256EthSigLen)
	err = dso.ValidateSignature(msg, dssBN, false)
	if err == nil {
		t.Fatal("Should raise an error (5)")
	}

	dss.CurveSpec = constants.CurveSecp256k1
	dss.Signature = make([]byte, constants.CurveSecp256k1SigLen)
	err = dso.ValidateSignature(msg, dss, false)
	if err == nil {
		t.Fatal("Should raise an error (6)")
	}

	dss, err = dso.Sign(msg, signer)
	if err != nil {
		t.Fatal(err)
	}
	err = dso.ValidateSignature(msg, dss, false)
	if err == nil {
		t.Fatal("Should raise an error (7)")
	}

	dss, err = dso.Sign(msg, signer)
	if err != nil {
		t.Fatal(err)
	}
	dso.New(acct, constants.CurveSecp256k1)
	err = dso.ValidateSignature(msg, dss, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDSOwnerValidateSignatureBN(t *testing.T) {
	dsp := &DSPreImage{}
	msg := make([]byte, 0)
	dss := &DataStoreSignature{}
	err := dsp.Owner.ValidateSignature(msg, dss, false)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	dso := &DataStoreOwner{}
	err = dso.ValidateSignature(msg, dss, false)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	// Create valid DataStoreOwner for BN256Eth
	signer := &crypto.BNSigner{}
	privk := make([]byte, 32)
	privk[0] = 1
	privk[31] = 1
	err = signer.SetPrivk(privk)
	if err != nil {
		t.Fatal(err)
	}
	pk, err := signer.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	acct := crypto.GetAccount(pk)

	acctBad := make([]byte, constants.OwnerLen)
	dso.New(acctBad, constants.CurveBN256Eth)
	err = dso.ValidateSignature(msg, dss, false)
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	dssBN := &DataStoreSignature{}
	dssBN.SVA = DataStoreSVA
	dssBN.CurveSpec = constants.CurveSecp256k1
	dssBN.Signature = make([]byte, constants.CurveSecp256k1SigLen)
	err = dso.ValidateSignature(msg, dssBN, false)
	if err == nil {
		t.Fatal("Should raise an error (5)")
	}

	dss.CurveSpec = constants.CurveBN256Eth
	dss.Signature = make([]byte, constants.CurveBN256EthSigLen)
	err = dso.ValidateSignature(msg, dss, false)
	if err == nil {
		t.Fatal("Should raise an error (6)")
	}

	msgBad := make([]byte, len(msg)+1)
	dss, err = dso.Sign(msgBad, signer)
	if err != nil {
		t.Fatal(err)
	}
	err = dso.ValidateSignature(msg, dss, false)
	if err == nil {
		t.Fatal("Should raise an error (7)")
	}

	dss, err = dso.Sign(msg, signer)
	if err != nil {
		t.Fatal(err)
	}
	err = dso.ValidateSignature(msg, dss, false)
	if err == nil {
		t.Fatal("Should raise an error (8)")
	}

	dss, err = dso.Sign(msg, signer)
	if err != nil {
		t.Fatal(err)
	}
	dso.New(acct, constants.CurveBN256Eth)
	err = dso.ValidateSignature(msg, dss, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDSSignatureMarshalBinary(t *testing.T) {
	dsSig := &DataStoreSignature{}
	_, err := dsSig.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	dsSig.SVA = DataStoreSVA
	_, err = dsSig.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	dsSig.CurveSpec = constants.CurveSecp256k1
	_, err = dsSig.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	dsSig.Signature = make([]byte, constants.CurveSecp256k1SigLen)
	_, err = dsSig.MarshalBinary()
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestDSSignatureUnmarshalBinary(t *testing.T) {
	signature := make([]byte, 0)
	ds := &DataStore{}
	err := ds.Signature.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should raise an error (0)")
	}

	dsSig := &DataStoreSignature{}
	err = dsSig.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	signature = make([]byte, 1)
	signature[0] = uint8(DataStoreSVA)
	err = dsSig.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	signature = make([]byte, 2)
	signature[0] = uint8(DataStoreSVA)
	signature[1] = uint8(constants.CurveSecp256k1)
	err = dsSig.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	signature = make([]byte, 2+constants.CurveSecp256k1SigLen+1)
	signature[0] = uint8(DataStoreSVA)
	signature[1] = uint8(constants.CurveSecp256k1)
	err = dsSig.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	signature = make([]byte, 2+constants.CurveSecp256k1SigLen)
	signature[0] = uint8(DataStoreSVA)
	signature[1] = uint8(constants.CurveSecp256k1)
	err = dsSig.UnmarshalBinary(signature)
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestDSSignatureValidate(t *testing.T) {
	ds := &DataStore{}
	err := ds.Signature.Validate()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	dsSig := &DataStoreSignature{}
	err = dsSig.Validate()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	dsSig.SVA = DataStoreSVA
	err = dsSig.Validate()
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	dsSig.CurveSpec = constants.CurveSecp256k1
	err = dsSig.Validate()
	if err == nil {
		t.Fatal("Should raise an error (4)")
	}

	dsSig.Signature = make([]byte, constants.CurveSecp256k1SigLen)
	err = dsSig.Validate()
	if err != nil {
		t.Fatal("Should pass")
	}

	err = ds.Signature.validateSVA()
	if err == nil {
		t.Fatal("Should raise an error (5)")
	}
	err = ds.Signature.validateCurveSpec()
	if err == nil {
		t.Fatal("Should raise an error (6)")
	}
}
