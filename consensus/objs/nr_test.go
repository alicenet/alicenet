package objs

import (
	"github.com/alicenet/alicenet/constants"
	"testing"

	"github.com/alicenet/alicenet/crypto"
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

	nrcBytes, err := nr2.NRClaims.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	nrcl := &NRClaims{}
	err = nrcl.UnmarshalBinary(nrcBytes)
	if err != nil {
		t.Fatal(err)
	}

	nrcl.RClaims.Height++
	nrclBytes, err := nrcl.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	nrcErr := &NRClaims{}
	err = nrcErr.UnmarshalBinary(nrclBytes)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	nrcl.RClaims.Height--
	nrcl.RClaims.Round++
	nrclBytes, err = nrcl.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	nrcErr = &NRClaims{}
	err = nrcErr.UnmarshalBinary(nrclBytes)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	nrcl.RClaims.Round--
	nrcl.RClaims.ChainID++
	nrclBytes, err = nrcl.MarshalBinary()
	if err != nil {
		t.Fatal("Should have raised error (3)")
	}
	nrcErr = &NRClaims{}
	err = nrcErr.UnmarshalBinary(nrclBytes)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	nrcl.RClaims.ChainID--
	nrcl.RClaims.PrevBlock = make([]byte, constants.HashLen)
	nrclBytes, err = nrcl.MarshalBinary()
	if err != nil {
		t.Fatal("Should have raised error (5)")
	}
	nrcErr = &NRClaims{}
	err = nrcErr.UnmarshalBinary(nrclBytes)
	if err == nil {
		t.Fatal("Should have raised error (6)")
	}

	nrcl.RClaims.PrevBlock = []byte{1, 2, 3}
	nrclBytes, err = nrcl.MarshalBinary()
	if err != nil {
		t.Fatal("Should have raised error (7)")
	}
	nrcErr = &NRClaims{}
	err = nrcErr.UnmarshalBinary(nrclBytes)
	if err == nil {
		t.Fatal("Should have raised error (8)")
	}
}
