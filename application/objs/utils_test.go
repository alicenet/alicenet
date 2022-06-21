package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
)

func TestMakeUTXOID(t *testing.T) {
	txhash1 := make([]byte, constants.HashLen)
	txhash1[0] = 1
	idxMax := constants.MaxUint32
	ret := MakeUTXOID(txhash1, idxMax)
	if !bytes.Equal(ret, txhash1) {
		t.Fatal("Incorrect return value (1)")
	}

	txhash2 := make([]byte, constants.HashLen)
	txhash2[0] = 2
	idx := uint32(17)
	idxBytes := utils.MarshalUint32(idx)
	utxoID := crypto.Hasher(txhash2, idxBytes)
	ret = MakeUTXOID(txhash2, idx)
	if !bytes.Equal(ret, utxoID) {
		t.Fatal("Incorrect return value (2)")
	}
}

func TestExtractSignature(t *testing.T) {
	secpSigBad := make([]byte, constants.CurveSecp256k1SigLen-1)
	_, _, err := extractSignature(secpSigBad, constants.CurveSecp256k1)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	secpSigGood := make([]byte, constants.CurveSecp256k1SigLen)
	secpSigGood[0] = 1
	secpSigGood[constants.CurveSecp256k1SigLen-1] = 1
	sig, null, err := extractSignature(secpSigGood, constants.CurveSecp256k1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(secpSigGood, sig) {
		t.Fatal("sigs do not match (1)")
	}
	if len(null) != 0 {
		t.Fatal("Should be null (1)")
	}

	bn256SigBad := make([]byte, constants.CurveBN256EthSigLen-1)
	_, _, err = extractSignature(bn256SigBad, constants.CurveBN256Eth)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	bn256SigGood := make([]byte, constants.CurveBN256EthSigLen)
	bn256SigGood[0] = 1
	bn256SigGood[constants.CurveBN256EthSigLen-1] = 1
	bnsig, null, err := extractSignature(bn256SigGood, constants.CurveBN256Eth)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bn256SigGood, bnsig) {
		t.Fatal("sigs do not match (2)")
	}
	if len(null) != 0 {
		t.Fatal("Should be null (2)")
	}

	_, _, err = extractSignature(nil, constants.CurveSpec(0))
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}

func TestValidateSignature(t *testing.T) {
	secpSigBad := make([]byte, constants.CurveSecp256k1SigLen-1)
	err := validateSignatureLen(secpSigBad, constants.CurveSecp256k1)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	secpSigGood := make([]byte, constants.CurveSecp256k1SigLen)
	secpSigGood[0] = 1
	secpSigGood[constants.CurveSecp256k1SigLen-1] = 1
	err = validateSignatureLen(secpSigGood, constants.CurveSecp256k1)
	if err != nil {
		t.Fatal(err)
	}

	bn256SigBad := make([]byte, constants.CurveBN256EthSigLen-1)
	err = validateSignatureLen(bn256SigBad, constants.CurveBN256Eth)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	bn256SigGood := make([]byte, constants.CurveBN256EthSigLen)
	bn256SigGood[0] = 1
	bn256SigGood[constants.CurveBN256EthSigLen-1] = 1
	err = validateSignatureLen(bn256SigGood, constants.CurveBN256Eth)
	if err != nil {
		t.Fatal(err)
	}

	err = validateSignatureLen(nil, constants.CurveSpec(0))
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}

func TestExtractCurveSpec(t *testing.T) {
	owner := make([]byte, 0)
	_, _, err := extractCurveSpec(owner)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	owner2 := make([]byte, 2)
	curveSpec := constants.CurveSecp256k1
	owner2[0] = uint8(curveSpec)
	signal := uint8(255)
	owner2[1] = signal
	cs, val, err := extractCurveSpec(owner2)
	if err != nil {
		t.Fatal(err)
	}
	if cs != curveSpec {
		t.Fatal("CurveSpecs do not match")
	}
	if len(val) != 1 {
		t.Fatal("Incorrect return length")
	}
	if uint8(val[0]) != signal {
		t.Fatal("Returned values do not match")
	}
}

func TestExtractSignerRole(t *testing.T) {
	owner := make([]byte, 0)
	_, _, err := extractSignerRole(owner)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	owner2 := make([]byte, 2)
	signerRole := PrimarySignerRole
	owner2[0] = uint8(signerRole)
	signal := uint8(255)
	owner2[1] = signal
	sr, val, err := extractSignerRole(owner2)
	if err != nil {
		t.Fatal(err)
	}
	if sr != signerRole {
		t.Fatal("SignerRoles do not match")
	}
	if len(val) != 1 {
		t.Fatal("Incorrect return length")
	}
	if uint8(val[0]) != signal {
		t.Fatal("Returned values do not match")
	}
}

func TestExtractAccount(t *testing.T) {
	owner := make([]byte, 0)
	_, _, err := extractAccount(owner)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	acct := make([]byte, constants.OwnerLen)
	acct[0] = 1
	acct[constants.OwnerLen-1] = 1
	nextStuff := make([]byte, 10)
	nextStuff[0] = 255
	owner2 := append(acct, nextStuff...)
	retAcct, val, err := extractAccount(owner2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(acct, retAcct) {
		t.Fatal("Accounts do not match")
	}
	if !bytes.Equal(val, nextStuff) {
		t.Fatal("Trailing bytes do not match")
	}
}

func TestExtractSVA(t *testing.T) {
	owner := make([]byte, 0)
	_, _, err := extractSVA(owner)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	sva := ValueStoreSVA
	nextStuff := make([]byte, 10)
	nextStuff[0] = 255
	owner2 := []byte{uint8(sva)}
	owner2 = append(owner2, nextStuff...)
	retSVA, val, err := extractSVA(owner2)
	if err != nil {
		t.Fatal(err)
	}
	if retSVA != sva {
		t.Fatal("SVAs do not match")
	}
	if !bytes.Equal(val, nextStuff) {
		t.Fatal("Trailing bytes do not match")
	}
}

func TestExtractHash(t *testing.T) {
	owner := make([]byte, 0)
	_, _, err := extractHash(owner)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	hash := make([]byte, constants.HashLen)
	hash[0] = 255
	hash[constants.HashLen-1] = 255
	nextStuff := make([]byte, 10)
	nextStuff[0] = 255
	owner2 := append(hash, nextStuff...)
	retHash, val, err := extractHash(owner2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retHash, hash) {
		t.Fatal("Hashes do not match")
	}
	if !bytes.Equal(val, nextStuff) {
		t.Fatal("Trailing bytes do not match")
	}
}

func TestExtractZero(t *testing.T) {
	owner := make([]byte, 1)
	err := extractZero(owner)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	owner2 := make([]byte, 0)
	err = extractZero(owner2)
	if err != nil {
		t.Fatal(err)
	}
}
