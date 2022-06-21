package objs

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"

	"github.com/alicenet/alicenet/crypto"
	bn256 "github.com/alicenet/alicenet/crypto/bn256/cloudflare"
)

const (
	tChainID uint32 = 2
)

func TestState(t *testing.T) {
	groupk, bnSigners, bnShares, secpSigners, secpPubks := makeSigners2(t)
	_ = secpPubks
	_ = bnShares
	vs := makeValidatorSet(1, groupk, secpSigners, bnSigners)
	_ = vs
	height := uint32(2)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pr1, _, _, _, _, nrl, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	cert, err := nrl.MakeRoundCert(bnSigners[0], bnShares)
	if err != nil {
		t.Fatal(err)
	}

	err = cert.ValidateSignature(&crypto.BNGroupValidator{})
	if err != nil {
		t.Fatal(err)
	}

	ovs := OwnValidatingState{
		ValidValue:  pr1[0],
		LockedValue: pr1[0],
	}

	ovs.SetRoundStarted()
	ovs.SetPreCommitStepStarted()
	ovs.SetPreVoteStepStarted()

	ptoExpired := ovs.PTOExpired(100 * time.Second)
	assert.False(t, ptoExpired)
	pvtoExpired := ovs.PVTOExpired(100 * time.Second)
	assert.False(t, pvtoExpired)
	pctoExpired := ovs.PCTOExpired(100 * time.Second)
	assert.True(t, pctoExpired)
	dbrnrExpired := ovs.DBRNRExpired(100 * time.Second)
	assert.True(t, dbrnrExpired)

	bn, err := ovs.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	ovs2 := &OwnValidatingState{}
	err = ovs2.UnmarshalBinary(bn)
	if err != nil {
		t.Fatal(err)
	}
}

func makeSecpSigner(seed []byte) (*crypto.Secp256k1Signer, []byte) {
	secpSigner := &crypto.Secp256k1Signer{}
	err := secpSigner.SetPrivk(crypto.Hasher(seed))
	if err != nil {
		panic(err)
	}
	secpKey, _ := secpSigner.Pubkey()
	return secpSigner, secpKey
}

func buildRound(t *testing.T, bnSigners []*crypto.BNGroupSigner, groupShares [][]byte, secpSigners []*crypto.Secp256k1Signer, height uint32, round uint32, prevBlock []byte) (*BlockHeader, []*Proposal, PreVoteList, []*PreVoteNil, PreCommitList, []*PreCommitNil, NextRoundList, NextHeightList, *BlockHeader) {
	pl := []*Proposal{}
	pvl := PreVoteList{}
	pvnl := []*PreVoteNil{}
	pcl := PreCommitList{}
	pcnl := []*PreCommitNil{}
	nrl := NextRoundList{}
	nhl := NextHeightList{}
	var bh *BlockHeader
	for i := 0; i < len(secpSigners); i++ {
		rcc := &RClaims{
			Height:    height,
			Round:     round,
			PrevBlock: prevBlock,
			ChainID:   tChainID,
		}
		var rc *RCert
		if round > 0 {
			sigs := [][]byte{}
			rcBytes, err := rcc.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			for _, signer := range bnSigners {
				sig, err := signer.Sign(rcBytes)
				if err != nil {
					t.Fatal(err)
				}
				sigs = append(sigs, sig)
			}
			sigGroup, err := bnSigners[i].Aggregate(sigs, groupShares)
			if err != nil {
				t.Fatal(err)
			}
			rcert := &RCert{
				SigGroup: sigGroup,
				RClaims:  rcc,
			}
			rc = rcert
		} else {
			sigs := [][]byte{}
			rcBytes := prevBlock
			for _, signer := range bnSigners {
				sig, err := signer.Sign(rcBytes)
				if err != nil {
					t.Fatal(err)
				}
				sigs = append(sigs, sig)
			}
			sigGroup, err := bnSigners[i].Aggregate(sigs, groupShares)
			if err != nil {
				t.Fatal(err)
			}
			rcert := &RCert{
				SigGroup: sigGroup,
				RClaims:  rcc,
			}
			rc = rcert
		}
		bc := &BClaims{
			ChainID:    tChainID,
			Height:     height,
			PrevBlock:  prevBlock,
			StateRoot:  prevBlock,
			HeaderRoot: crypto.Hasher([]byte{}),
			TxRoot:     crypto.Hasher([]byte{}),
		}
		p := &Proposal{
			TxHshLst: [][]byte{},
			PClaims: &PClaims{
				BClaims: bc,
				RCert:   rc,
			},
		}
		err := p.Sign(secpSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		pl = append(pl, p)
		pv, err := p.PreVote(secpSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		pvl = append(pvl, pv)
		pvn, err := rc.PreVoteNil(secpSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		pvnl = append(pvnl, pvn)
		pcn, err := rc.PreCommitNil(secpSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		pcnl = append(pcnl, pcn)
		nr, err := rc.NextRound(secpSigners[i], bnSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		nrl = append(nrl, nr)
		bheader := &BlockHeader{
			TxHshLst: [][]byte{},
			BClaims:  bc,
		}
		ppsl, err := bheader.MakeDeadBlockRoundProposal(rc, crypto.Hasher([]byte{}))
		if err != nil {
			t.Fatal(err)
		}
		if ppsl == nil {
			t.Fatal("invalid Dead Block Round proposal")
		}
		bhsh, err := bheader.BlockHash()
		if err != nil {
			t.Fatal(err)
		}
		sigs := [][]byte{}
		for _, signer := range bnSigners {
			sig, err := signer.Sign(bhsh)
			if err != nil {
				t.Fatal(err)
			}
			sigs = append(sigs, sig)
		}
		sigGroup, err := bnSigners[i].Aggregate(sigs, groupShares)
		if err != nil {
			t.Fatal(err)
		}
		bheader.SigGroup = sigGroup
		bh = bheader
	}
	for i := 0; i < len(secpSigners); i++ {
		pcc, err := pvl.MakePreCommit(secpSigners[i])
		if err != nil {
			t.Fatal(err)
		}

		ppsl, err := pvl.GetProposal()
		if err != nil {
			t.Fatal(err)
		}
		if ppsl == nil {
			t.Fatal("invalid proposal")
		}
		pcl = append(pcl, pcc)
	}
	for i := 0; i < len(secpSigners); i++ {
		nh, err := pcl.MakeNextHeight(secpSigners[i], bnSigners[i])
		if err != nil {
			t.Fatal(err)
		}

		ppsl, err := pcl.GetProposal()
		if err != nil {
			t.Fatal(err)
		}
		if ppsl == nil {
			t.Fatal("invalid proposal")
		}
		nhl = append(nhl, nh)
	}
	newBh, _, err := nhl.MakeBlockHeader(bnSigners[0], groupShares)
	if err != nil {
		t.Fatal(err)
	}
	return bh, pl, pvl, pvnl, pcl, pcnl, nrl, nhl, newBh
}

func makeValidatorSet(nbbh uint32, groupk []byte, secpSigners []*crypto.Secp256k1Signer, bnSigners []*crypto.BNGroupSigner) *ValidatorSet {
	valList := []*Validator{}
	for i, ss := range secpSigners {
		groupKey, _ := bnSigners[i].PubkeyShare()
		secpKey, _ := ss.Pubkey()
		val := &Validator{
			VAddr:      secpKey, // change
			GroupShare: groupKey,
		}
		valList = append(valList, val)
	}
	vs := &ValidatorSet{
		Validators: valList,
		GroupKey:   groupk,
		NotBefore:  nbbh,
	}
	return vs
}

func makeSigners2(t *testing.T) ([]byte, []*crypto.BNGroupSigner, [][]byte, []*crypto.Secp256k1Signer, [][]byte) {
	s := new(crypto.BNGroupSigner)
	msg := []byte("A message to sign")

	secret1 := big.NewInt(100)
	secret2 := big.NewInt(101)
	secret3 := big.NewInt(102)
	secret4 := big.NewInt(103)

	msk := big.NewInt(0)
	msk.Add(msk, secret1)
	msk.Add(msk, secret2)
	msk.Add(msk, secret3)
	msk.Add(msk, secret4)
	msk.Mod(msk, bn256.Order)
	mpk := new(bn256.G2).ScalarBaseMult(msk)

	big1 := big.NewInt(1)
	big2 := big.NewInt(2)

	privCoefs1 := []*big.Int{secret1, big1, big2}
	privCoefs2 := []*big.Int{secret2, big1, big2}
	privCoefs3 := []*big.Int{secret3, big1, big2}
	privCoefs4 := []*big.Int{secret4, big1, big2}

	share1to1 := bn256.PrivatePolyEval(privCoefs1, 1)
	share1to2 := bn256.PrivatePolyEval(privCoefs1, 2)
	share1to3 := bn256.PrivatePolyEval(privCoefs1, 3)
	share1to4 := bn256.PrivatePolyEval(privCoefs1, 4)
	share2to1 := bn256.PrivatePolyEval(privCoefs2, 1)
	share2to2 := bn256.PrivatePolyEval(privCoefs2, 2)
	share2to3 := bn256.PrivatePolyEval(privCoefs2, 3)
	share2to4 := bn256.PrivatePolyEval(privCoefs2, 4)
	share3to1 := bn256.PrivatePolyEval(privCoefs3, 1)
	share3to2 := bn256.PrivatePolyEval(privCoefs3, 2)
	share3to3 := bn256.PrivatePolyEval(privCoefs3, 3)
	share3to4 := bn256.PrivatePolyEval(privCoefs3, 4)
	share4to1 := bn256.PrivatePolyEval(privCoefs4, 1)
	share4to2 := bn256.PrivatePolyEval(privCoefs4, 2)
	share4to3 := bn256.PrivatePolyEval(privCoefs4, 3)
	share4to4 := bn256.PrivatePolyEval(privCoefs4, 4)

	groupShares := make([][]byte, 4)
	for k := 0; k < len(groupShares); k++ {
		groupShares[k] = make([]byte, len(mpk.Marshal()))
	}

	listOfSS1 := []*big.Int{share1to1, share2to1, share3to1, share4to1}
	gsk1 := bn256.GenerateGroupSecretKeyPortion(listOfSS1)
	gpk1 := new(bn256.G2).ScalarBaseMult(gsk1)
	groupShares[0] = gpk1.Marshal()
	s1 := new(crypto.BNGroupSigner)
	err := s1.SetPrivk(gsk1.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	sig1, err := s1.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}

	listOfSS2 := []*big.Int{share1to2, share2to2, share3to2, share4to2}
	gsk2 := bn256.GenerateGroupSecretKeyPortion(listOfSS2)
	gpk2 := new(bn256.G2).ScalarBaseMult(gsk2)
	groupShares[1] = gpk2.Marshal()
	s2 := new(crypto.BNGroupSigner)
	err = s2.SetPrivk(gsk2.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	sig2, err := s2.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}

	listOfSS3 := []*big.Int{share1to3, share2to3, share3to3, share4to3}
	gsk3 := bn256.GenerateGroupSecretKeyPortion(listOfSS3)
	gpk3 := new(bn256.G2).ScalarBaseMult(gsk3)
	groupShares[2] = gpk3.Marshal()
	s3 := new(crypto.BNGroupSigner)
	err = s3.SetPrivk(gsk3.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	sig3, err := s3.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}

	listOfSS4 := []*big.Int{share1to4, share2to4, share3to4, share4to4}
	gsk4 := bn256.GenerateGroupSecretKeyPortion(listOfSS4)
	gpk4 := new(bn256.G2).ScalarBaseMult(gsk4)
	groupShares[3] = gpk4.Marshal()
	s4 := new(crypto.BNGroupSigner)
	err = s4.SetPrivk(gsk4.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	sig4, err := s4.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}

	sigs := make([][]byte, 4)
	for k := 0; k < len(sigs); k++ {
		sigs[k] = make([]byte, 192)
	}
	sigs[0] = sig1
	sigs[1] = sig2
	sigs[2] = sig3
	sigs[3] = sig4

	err = s.SetGroupPubk(mpk.Marshal())
	if err != nil {
		t.Fatal(err)
	}

	// Make bad sigs array
	sigsBad := make([][]byte, 2)
	for k := 0; k < len(sigsBad); k++ {
		sigsBad[k] = make([]byte, 192)
	}
	sigsBad[0] = sig1
	sigsBad[1] = sig2
	_, err = s.Aggregate(sigsBad, groupShares)
	if err == nil {
		t.Fatal("Should have raised an error for too few signatures!")
	}

	// Finally submit signature
	grpsig, err := s.Aggregate(sigs, groupShares)
	if err != nil {
		t.Fatal(err)
	}

	bnVal := &crypto.BNGroupValidator{}
	groupk, err := bnVal.PubkeyFromSig(grpsig)
	if err != nil {
		t.Fatal(err)
	}

	bnSigners := []*crypto.BNGroupSigner{}
	bnSigners = append(bnSigners, s1)
	bnSigners = append(bnSigners, s2)
	bnSigners = append(bnSigners, s3)
	bnSigners = append(bnSigners, s4)

	secpSigners := []*crypto.Secp256k1Signer{}
	secpPubks := [][]byte{}
	for _, share := range groupShares {
		signer, pubk := makeSecpSigner(share)
		secpPubks = append(secpPubks, pubk)
		secpSigners = append(secpSigners, signer)
	}

	for _, signer := range bnSigners {
		err := signer.SetGroupPubk(groupk)
		if err != nil {
			t.Fatal(err)
		}
	}

	return groupk, bnSigners, groupShares, secpSigners, secpPubks
}
