package objs

import (
	"testing"

	"github.com/MadBase/MadNet/crypto"
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
}
