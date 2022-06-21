package gossip

import (
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/crypto"
	bn256 "github.com/alicenet/alicenet/crypto/bn256/cloudflare"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

const (
	tChainID uint32 = 2
)

func TestState(t *testing.T) {
	groupk, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	_ = secpPubks
	_ = bnShares
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	DB, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer DB.Close()
	database := &db.Database{}
	database.Init(DB)
	if err != nil {
		t.Fatal(err)
	}
	height := uint32(2)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	pbh, pl, pvl, pvnl, pcl, pcnl, nrl, nhl, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	_ = pbh
	_ = pl
	_ = pvl
	_ = pvnl
	_ = pcl
	_ = pcnl
	_ = nrl
	_ = nhl
	_ = bh

	cert, err := nrl.MakeRoundCert(bnSigners[0], bnShares)
	if err != nil {
		t.Fatal(err)
	}

	err = cert.ValidateSignature(&crypto.BNGroupValidator{})
	if err != nil {
		t.Fatal(err)
	}

	err = DB.Update(func(txn *badger.Txn) error {
		vs := makeValidatorSet(1, groupk, secpSigners, bnSigners)
		err := database.SetValidatorSet(txn, vs)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = DB.Update(func(txn *badger.Txn) error {
		ownState := makeOwnState(secpPubks[0], pbh, pbh, pbh, pbh)
		err = database.SetOwnState(txn, ownState)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = DB.Update(func(txn *badger.Txn) error {
		roundStates := makeRoundStates(secpSigners, bnSigners, groupk, pl)
		for _, rs := range roundStates {
			err = database.SetCurrentRoundState(txn, rs)
			if err != nil {
				return err
			}
		}
		return nil
	})
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

func buildRound(t *testing.T, bnSigners []*crypto.BNGroupSigner, groupSharesOrig [][]byte, secpSigners []*crypto.Secp256k1Signer, height uint32, round uint32, prevBlockOrig []byte) (*objs.BlockHeader, []*objs.Proposal, objs.PreVoteList, []*objs.PreVoteNil, objs.PreCommitList, []*objs.PreCommitNil, objs.NextRoundList, objs.NextHeightList, *objs.BlockHeader) {
	groupShares := make([][]byte, len(groupSharesOrig))
	copy(groupShares, groupSharesOrig)
	prevBlock := utils.CopySlice(prevBlockOrig)
	bnVal := &crypto.BNGroupValidator{}
	secpVal := &crypto.Secp256k1Validator{}
	pl := []*objs.Proposal{}
	pvl := objs.PreVoteList{}
	pvnl := []*objs.PreVoteNil{}
	pcl := objs.PreCommitList{}
	pcnl := []*objs.PreCommitNil{}
	nrl := objs.NextRoundList{}
	nhl := objs.NextHeightList{}
	var bh *objs.BlockHeader
	for i := 0; i < len(secpSigners); i++ {
		rcc := &objs.RClaims{
			Height:    height,
			Round:     round,
			PrevBlock: prevBlock,
			ChainID:   tChainID,
		}
		var rc *objs.RCert
		if round > 1 {
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
			rcert := &objs.RCert{
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
			rcert := &objs.RCert{
				SigGroup: sigGroup,
				RClaims:  rcc,
			}
			rc = rcert
		}
		err := rc.ValidateSignature(bnVal)
		if err != nil {
			t.Fatal(err)
		}
		bc := &objs.BClaims{
			ChainID:    tChainID,
			Height:     height,
			PrevBlock:  prevBlock,
			StateRoot:  prevBlock,
			HeaderRoot: crypto.Hasher([]byte{}),
			TxRoot:     crypto.Hasher([]byte{}),
		}
		p := &objs.Proposal{
			TxHshLst: [][]byte{},
			PClaims: &objs.PClaims{
				BClaims: bc,
				RCert:   rc,
			},
		}
		err = p.Sign(secpSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		pBytes, err := p.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		p = &objs.Proposal{}
		err = p.UnmarshalBinary(pBytes)
		if err != nil {
			t.Fatal(err)
		}
		err = p.ValidateSignatures(secpVal, bnVal)
		if err != nil {
			t.Fatal(err)
		}
		pl = append(pl, p)
		pv, err := p.PreVote(secpSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		pvBytes, err := pv.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		pv = &objs.PreVote{}
		err = pv.UnmarshalBinary(pvBytes)
		if err != nil {
			t.Fatal(err)
		}
		err = pv.ValidateSignatures(secpVal, bnVal)
		if err != nil {
			t.Fatal(err)
		}
		pvl = append(pvl, pv)
		pvn, err := rc.PreVoteNil(secpSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		pvnBytes, err := pvn.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		pvn = &objs.PreVoteNil{}
		err = pvn.UnmarshalBinary(pvnBytes)
		if err != nil {
			t.Fatal(err)
		}
		err = pvn.ValidateSignatures(secpVal, bnVal)
		if err != nil {
			t.Fatal(err)
		}
		pvnl = append(pvnl, pvn)
		pcn, err := rc.PreCommitNil(secpSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		pcnBytes, err := pcn.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		pcn = &objs.PreCommitNil{}
		err = pcn.UnmarshalBinary(pcnBytes)
		if err != nil {
			t.Fatal(err)
		}
		err = pcn.ValidateSignatures(secpVal, bnVal)
		if err != nil {
			t.Fatal(err)
		}
		pcnl = append(pcnl, pcn)
		nr, err := rc.NextRound(secpSigners[i], bnSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		nrBytes, err := nr.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		nr = &objs.NextRound{}
		err = nr.UnmarshalBinary(nrBytes)
		if err != nil {
			t.Fatal(err)
		}
		err = nr.ValidateSignatures(secpVal, bnVal)
		if err != nil {
			t.Fatal(err)
		}
		nrl = append(nrl, nr)
		bheader := &objs.BlockHeader{
			TxHshLst: [][]byte{},
			BClaims:  bc,
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
		err = bh.ValidateSignatures(bnVal)
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < len(secpSigners); i++ {
		pcc, err := pvl.MakePreCommit(secpSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		pccBytes, err := pcc.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		pcc = &objs.PreCommit{}
		err = pcc.UnmarshalBinary(pccBytes)
		if err != nil {
			t.Fatal(err)
		}
		err = pcc.ValidateSignatures(secpVal, bnVal)
		if err != nil {
			t.Fatal(err)
		}
		pcl = append(pcl, pcc)
	}
	for i := 0; i < len(secpSigners); i++ {
		nh, err := pcl.MakeNextHeight(secpSigners[i], bnSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		nhBytes, err := nh.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		nh = &objs.NextHeight{}
		err = nh.UnmarshalBinary(nhBytes)
		if err != nil {
			t.Fatal(err)
		}
		err = nh.ValidateSignatures(secpVal, bnVal)
		if err != nil {
			t.Fatal(err)
		}
		nhl = append(nhl, nh)
	}
	newBh, _, err := nhl.MakeBlockHeader(bnSigners[0], groupShares)
	if err != nil {
		t.Fatal(err)
	}
	return bh, pl, pvl, pvnl, pcl, pcnl, nrl, nhl, newBh
}

func makeOwnState(vkey []byte, SyncToBH, MaxBHSeen, CanonicalSnapShot, PendingSnapShot *objs.BlockHeader) *objs.OwnState {
	return &objs.OwnState{
		VAddr:             vkey,
		SyncToBH:          SyncToBH,
		MaxBHSeen:         MaxBHSeen,
		CanonicalSnapShot: CanonicalSnapShot,
		PendingSnapShot:   PendingSnapShot,
	}
}

func makeValidatorSet(nbbh uint32, groupkOrig []byte, secpSigners []*crypto.Secp256k1Signer, bnSigners []*crypto.BNGroupSigner) *objs.ValidatorSet {
	valList := []*objs.Validator{}
	for i, ss := range secpSigners {
		groupShareOrig, _ := bnSigners[i].PubkeyShare()
		groupShare := utils.CopySlice(groupShareOrig)
		secpKeyOrig, _ := ss.Pubkey()
		secpKey := utils.CopySlice(secpKeyOrig)
		val := &objs.Validator{
			VAddr:      secpKey,
			GroupShare: groupShare,
		}
		valList = append(valList, val)
	}
	groupKey := utils.CopySlice(groupkOrig)
	vs := &objs.ValidatorSet{
		Validators: valList,
		GroupKey:   groupKey,
		NotBefore:  nbbh,
	}
	return vs
}

func makeRoundStates(secpSigners []*crypto.Secp256k1Signer, bnSigners []*crypto.BNGroupSigner, groupk []byte, pl []*objs.Proposal) []*objs.RoundState {
	rsl := []*objs.RoundState{}
	for i := 0; i < len(secpSigners); i++ {
		rcert := pl[i].PClaims.RCert
		rcBytes, err := rcert.MarshalBinary()
		if err != nil {
			panic(err)
		}
		rc := &objs.RCert{}
		err = rc.UnmarshalBinary(rcBytes)
		if err != nil {
			panic(err)
		}
		secpPubkOrig, _ := secpSigners[i].Pubkey()
		secpPubk := utils.CopySlice(secpPubkOrig)
		groupShareOrig, _ := bnSigners[i].PubkeyShare()
		groupShare := utils.CopySlice(groupShareOrig)
		rs := makeRoundState(secpPubk, groupShare, groupk, i, rc)
		rsl = append(rsl, rs)
	}
	return rsl
}

func makeRoundState(secpKeyOrig []byte, groupShareOrig []byte, groupkOrig []byte, idx int, rcertOrig *objs.RCert) *objs.RoundState {
	secpKey := utils.CopySlice(secpKeyOrig)
	groupKey := utils.CopySlice(groupkOrig)
	groupShare := make([]byte, len(groupShareOrig))
	rcertBytes, err := rcertOrig.MarshalBinary()
	if err != nil {
		panic(err)
	}
	rcert := &objs.RCert{}
	err = rcert.UnmarshalBinary(rcertBytes)
	if err != nil {
		panic(err)
	}
	return &objs.RoundState{
		VAddr:      secpKey,
		GroupKey:   groupKey,
		GroupShare: groupShare,
		GroupIdx:   uint8(idx),
		RCert:      rcert,
	}
}

func makeSigners(t *testing.T) ([]byte, []*crypto.BNGroupSigner, [][]byte, []*crypto.Secp256k1Signer, [][]byte) {
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

/*
	pc := &objs.PreCommit{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
		PreVotes:  [][]byte{makeOnes(make([]byte, 65))},
	}
	nh := &objs.NextHeight{
		NHClaims: &objs.NHClaims{
			Proposal: p,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature:  makeOnes(make([]byte, 65)),
		PreCommits: [][]byte{makeOnes(make([]byte, 65))},
	}


	val := &objs.Validator{
		VAddr:       makeOnes(make([]byte, 32)),
		GroupShare: makeOnes(make([]byte, 128)),
	}
	vs := &objs.ValidatorSet{
		Validators: []*objs.Validator{val},
		GroupKey:   makeOnes(make([]byte, 128)),
		NotBefore:  1,
	}
	rs := &objs.RoundState{
		VAddr:       makeOnes(make([]byte, 32)),
		GroupKey:   makeOnes(make([]byte, 128)),
		GroupShare: makeOnes(make([]byte, 128)),
		GroupIdx:   0,
		RCert:      rc,
	}
}
*/
