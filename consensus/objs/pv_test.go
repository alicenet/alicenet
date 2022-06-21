package objs

import (
	"testing"

	"github.com/alicenet/alicenet/crypto"
)

func TestPreVote(t *testing.T) {
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
	gk := crypto.BNGroupSigner{}
	err = gk.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	sig, err := gk.Sign(bhsh)
	if err != nil {
		t.Fatal(err)
	}
	secpSigner := crypto.Secp256k1Signer{}
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
	pclmsBytes, err := pclms.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	sig, err = secpSigner.Sign(pclmsBytes)
	if err != nil {
		t.Fatal(err)
	}
	prop := &Proposal{
		PClaims:   pclms,
		Signature: sig,
		TxHshLst:  txHashListList[1],
	}
	sig, err = secpSigner.Sign(pclmsBytes)
	if err != nil {
		t.Fatal(err)
	}
	pv := &PreVote{
		Proposal:  prop,
		Signature: sig,
	}
	err = pv.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		t.Fatal(err)
	}
	pvBytes, err := pv.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	pv2 := &PreVote{}
	err = pv2.UnmarshalBinary(pvBytes)
	if err != nil {
		t.Fatal(err)
	}
	bclaimsEqual(t, pv.Proposal.PClaims.BClaims, pv2.Proposal.PClaims.BClaims)
	err = pv2.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		t.Fatal(err)
	}
}
