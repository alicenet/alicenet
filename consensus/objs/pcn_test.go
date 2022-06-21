package objs

import (
	"testing"

	"github.com/alicenet/alicenet/crypto"
)

func TestPreCommitNil(t *testing.T) {
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
	pvn, err := rcert.PreCommitNil(secpSigner)
	if err != nil {
		t.Fatal(err)
	}
	err = pvn.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		t.Fatal(err)
	}
	pvnbytes, err := pvn.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	pvn2 := &PreCommitNil{}
	err = pvn2.UnmarshalBinary(pvnbytes)
	if err != nil {
		t.Fatal(err)
	}
	err = pvn2.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		t.Fatal(err)
	}
}
