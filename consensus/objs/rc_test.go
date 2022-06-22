package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
)

func TestRCert(t *testing.T) {
	bnVal := &crypto.BNGroupValidator{}
	//secpVal := &crypto.Secp256k1Validator{}
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
	rc := &RCert{
		RClaims:  nr.NRClaims.RClaims,
		SigGroup: nr.NRClaims.SigShare,
	}
	err = rc.ValidateSignature(bnVal)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRCertMarshal(t *testing.T) {
	pvn := &PreVoteNil{}
	_, err := pvn.RCert.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	rc := &RCert{}
	_, err = rc.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
}

func TestRCertValidateSignature(t *testing.T) {
	bnVal := &crypto.BNGroupValidator{}
	pvn := &PreVoteNil{}
	err := pvn.RCert.ValidateSignature(bnVal)
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	// Invalid RClaims object
	rc := &RCert{}
	err = rc.ValidateSignature(bnVal)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	// Everything is good
	rc.RClaims = &RClaims{}
	rc.RClaims.ChainID = 1
	rc.RClaims.Height = 1
	rc.RClaims.Round = 1
	err = rc.ValidateSignature(bnVal)
	if err != nil {
		t.Fatal(err)
	}
	groupKey := make([]byte, constants.CurveBN256EthPubkeyLen)
	if !bytes.Equal(rc.GroupKey, groupKey) {
		t.Fatal("Invalid GroupKey")
	}

	// Invalid Height/Round combination; not possible to be Height 1, Round 2
	rc.GroupKey = nil
	rc.RClaims.Height = 1
	rc.RClaims.Round = 2
	err = rc.ValidateSignature(bnVal)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	// No error is raised
	rc.RClaims.Height = 2
	rc.RClaims.Round = 1
	err = rc.ValidateSignature(bnVal)
	if err != nil {
		t.Fatal(err)
	}

	// Should raise an error for invalid RClaims object
	rc.RClaims.Height = 2
	rc.RClaims.Round = constants.DEADBLOCKROUND + 1
	err = rc.ValidateSignature(bnVal)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	// Should raise an error for invalid PrevBlock
	rc.RClaims.Height = 2
	rc.RClaims.Round = 2
	rc.RClaims.PrevBlock = make([]byte, constants.HashLen+1)
	err = rc.ValidateSignature(bnVal)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	// Should raise an error for invalid PrevBlock
	rc.RClaims.Height = 2
	rc.RClaims.Round = 2
	rc.RClaims.PrevBlock = make([]byte, constants.HashLen)
	err = rc.ValidateSignature(bnVal)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}

	// Should raise an error for invalid PrevBlock
	rc.RClaims.Height = 3
	rc.RClaims.Round = 1
	rc.RClaims.PrevBlock = make([]byte, constants.HashLen)
	err = rc.ValidateSignature(bnVal)
	if err == nil {
		t.Fatal("Should have raised error (6)")
	}
}

func TestRCertPreVoteNil(t *testing.T) {
	rc := &RCert{}
	_, err := rc.PreVoteNil(nil)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	rc.RClaims = &RClaims{
		ChainID:   1,
		Height:    1,
		Round:     1,
		PrevBlock: make([]byte, constants.HashLen),
	}
	_, err = rc.PreVoteNil(nil)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestRCertPreCommitNil(t *testing.T) {
	rc := &RCert{}
	_, err := rc.PreCommitNil(nil)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	rc.RClaims = &RClaims{
		ChainID:   1,
		Height:    1,
		Round:     1,
		PrevBlock: make([]byte, constants.HashLen),
	}
	_, err = rc.PreCommitNil(nil)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestRCertNextRound(t *testing.T) {
	nrc := &NRClaims{}
	_, err := nrc.RCert.NextRound(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	rc := &RCert{}
	_, err = rc.NextRound(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	rc.RClaims = &RClaims{
		ChainID:   1,
		Height:    1,
		Round:     1,
		PrevBlock: make([]byte, constants.HashLen),
	}
	_, err = rc.NextRound(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}
