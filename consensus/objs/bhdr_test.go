package objs

import (
	"testing"

	"github.com/alicenet/alicenet/crypto"
)

func TestBlockHeader(t *testing.T) {
	bnVal := &crypto.BNGroupValidator{}
	bclaimsList, txHashListList, err := generateChain(1)
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
	bhdrBytes, err := bh.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	bh2 := &BlockHeader{}
	err = bh2.UnmarshalBinary(bhdrBytes)
	if err != nil {
		t.Fatal(err)
	}
	bclaimsEqual(t, bh.BClaims, bh2.BClaims)
	err = bh2.ValidateSignatures(bnVal)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBlockHeaderBad(t *testing.T) {
	bh := &BlockHeader{}
	_, err := bh.BlockHash()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	bnVal := &crypto.BNGroupValidator{}
	err = bh.ValidateSignatures(bnVal)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	_, err = bh.GetRCert()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	bclaimsList, txHashListList, err := generateChain(1)
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
	bh = &BlockHeader{
		BClaims:  bclaims,
		SigGroup: sig,
		TxHshLst: nil,
	}
	err = bh.ValidateSignatures(bnVal)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	// Mess up TxCount for testing unmarshal
	bclaims.TxCount = 0
	bh = &BlockHeader{
		BClaims:  bclaims,
		SigGroup: sig,
		TxHshLst: txHashListList[0],
	}
	bhdrBytes, err := bh.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	bh2 := &BlockHeader{}
	err = bh2.UnmarshalBinary(bhdrBytes)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}
}
