package objs

import (
	"testing"

	"github.com/MadBase/MadNet/crypto"
)

func TestNextRound(t *testing.T) {
	bnVal := &crypto.BNGroupValidator{}
	secpVal := &crypto.Secp256k1Validator{}
	bclaimsList, txHashListList, err := generateChain(3)
	if err != nil {
		t.Fatal(err)
	}
	bclaims := bclaimsList[2]
	bhsh, err := bclaims.BlockHash()
	if err != nil {
		t.Fatal(err)
	}
	gk := &crypto.BNGroupSigner{}
	gk.SetPrivk(crypto.Hasher([]byte("secret")))
	sig, err := gk.Sign(bhsh)
	if err != nil {
		t.Fatal(err)
	}
	secpSigner := &crypto.Secp256k1Signer{}
	err = secpSigner.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	bh := &BlockHeader{
		BClaims:  bclaims,
		SigGroup: sig,
		TxHshLst: txHashListList[2],
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
	nr, err := rcert.NextRound(secpSigner, gk)
	if err != nil {
		t.Fatal(err)
	}
	err = nr.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		t.Fatal(err)
	}
	nrbytes, err := nr.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	nr2 := &NextRound{}
	err = nr2.UnmarshalBinary(nrbytes)
	if err != nil {
		t.Fatal(err)
	}
	err = nr2.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		t.Fatal(err)
	}
}
