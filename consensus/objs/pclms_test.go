package objs

import (
	"testing"

	"github.com/alicenet/alicenet/constants"

	"github.com/alicenet/alicenet/crypto"
)

func TestPClaims(t *testing.T) {
	bnVal := &crypto.BNGroupValidator{}
	bclaimsList, txHashListList, err := generateChain(2)
	if err != nil {
		t.Fatal(err)
	}
	bclaims := bclaimsList[0]
	bhsh, err := bclaims.BlockHash()
	if err != nil {
		t.Fatal(err)
	}
	gk := crypto.BNGroupSigner{}
	err = gk.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	sig, err := gk.Sign(bhsh)
	if err != nil {
		t.Fatal(err)
	}
	bh := &BlockHeader{
		BClaims:  bclaims,
		SigGroup: sig,
		TxHshLst: txHashListList[0],
	}
	err = bh.ValidateSignatures(bnVal)
	if err != nil {
		t.Fatal(err)
	}
	rcert, err := bh.GetRCert()
	if err != nil {
		t.Fatal(err)
	}
	err = rcert.ValidateSignature(bnVal)
	if err != nil {
		t.Fatal(err)
	}
	pclms := &PClaims{
		BClaims: bclaimsList[1],
		RCert:   rcert,
	}
	pclmsBytes, err := pclms.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	pclms2 := &PClaims{}
	err = pclms2.UnmarshalBinary(pclmsBytes)
	if err != nil {
		t.Fatal(err)
	}
	err = pclms.RCert.ValidateSignature(bnVal)
	if err != nil {
		t.Fatal(err)
	}
	err = pclms2.RCert.ValidateSignature(bnVal)
	if err != nil {
		t.Fatal(err)
	}
	bclaimsEqual(t, pclms.BClaims, pclms2.BClaims)

	pclms2.BClaims.ChainID++
	pclms2Bytes, err := pclms2.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	pclms2Err := &PClaims{}
	err = pclms2Err.UnmarshalBinary(pclms2Bytes)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	pclms2.BClaims.ChainID--
	pclms2.BClaims.Height++
	pclms2Bytes, err = pclms2.MarshalBinary()
	if err != nil {
		t.Fatal("Should have raised error (2)")
	}
	pclms2Err = &PClaims{}
	err = pclms2Err.UnmarshalBinary(pclms2Bytes)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	pclms2.BClaims.Height--
	pclms2.BClaims.PrevBlock = make([]byte, constants.HashLen)
	pclms2Bytes, err = pclms2.MarshalBinary()
	if err != nil {
		t.Fatal("Should have raised error (4)")
	}
	pclms2Err = &PClaims{}
	err = pclms2Err.UnmarshalBinary(pclms2Bytes)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}

	pclms2.BClaims.PrevBlock = []byte{1, 2, 3}
	pclms2Bytes, err = pclms2.MarshalBinary()
	if err != nil {
		t.Fatal("Should have raised error (6)")
	}
	pclms2Err = &PClaims{}
	err = pclms2Err.UnmarshalBinary(pclms2Bytes)
	if err == nil {
		t.Fatal("Should have raised error (7)")
	}
}
