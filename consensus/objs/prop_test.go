package objs

import (
	"testing"

	"github.com/alicenet/alicenet/crypto"
)

func TestProposal(t *testing.T) {
	bnVal := &crypto.BNGroupValidator{}
	secpVal := &crypto.Secp256k1Validator{}
	bclaimsList, txHashListList, err := generateChain(2)
	if err != nil {
		t.Fatal(err)
	}
	bclaims := bclaimsList[0]
	bhsh, err := bclaims.BlockHash()
	if err != nil {
		t.Fatal(err)
	}
	gk := &crypto.BNGroupSigner{}
	err = gk.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
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
	prop := &Proposal{
		PClaims:  pclms,
		TxHshLst: txHashListList[1],
	}
	err = prop.Sign(secpSigner)
	if err != nil {
		t.Fatal(err)
	}
	err = prop.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		t.Fatal(err)
	}
	propBytes, err := prop.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	prop2 := &Proposal{}
	err = prop2.UnmarshalBinary(propBytes)
	if err != nil {
		t.Fatal(err)
	}
	bclaimsEqual(t, prop.PClaims.BClaims, prop2.PClaims.BClaims)
	err = prop2.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		t.Fatal(err)
	}
	prop3, err := prop.RePropose(secpSigner, prop.PClaims.RCert)
	if err != nil {
		t.Fatal(err)
	}
	err = prop3.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		t.Fatal(err)
	}
}
