package objs

import (
	"bytes"
	"errors"
	"strconv"
	"testing"

	"github.com/alicenet/alicenet/errorz"

	"github.com/alicenet/alicenet/crypto"
)

func rsEqual(t *testing.T, a, b *RoundState) {
	if !bytes.Equal(a.VAddr, b.VAddr) {
		t.Fatal("fail")
	}
	if !bytes.Equal(a.GroupKey, b.GroupKey) {
		t.Fatal("fail")
	}
	if !bytes.Equal(a.GroupShare, b.GroupShare) {
		t.Fatal("fail")
	}
	if a.GroupIdx != b.GroupIdx {
		t.Fatal("fail")
	}
	if a.RCert == nil || b.RCert == nil {
		t.Fatal("fail")
	}
	if a.RCert != nil {
		if !PrevBlockEqual(a.RCert, b.RCert) {
			t.Fatal("fail")
		}
	}
	if a.ConflictingRCert != nil || b.ConflictingRCert != nil {
		if !PrevBlockEqual(a.ConflictingRCert, b.ConflictingRCert) {
			t.Fatal("fail")
		}
	}
	if a.Proposal != nil || b.Proposal != nil {
		if !PrevBlockEqual(a.Proposal, b.Proposal) {
			t.Fatal("fail")
		}
	}
	if a.ConflictingProposal != nil || b.ConflictingProposal != nil {
		if !PrevBlockEqual(a.ConflictingProposal, b.ConflictingProposal) {
			t.Fatal("fail")
		}
	}
	if a.PreVote != nil || b.PreVote != nil {
		if !PrevBlockEqual(a.PreVote, b.PreVote) {
			t.Fatal("fail")
		}
	}
	if a.ConflictingPreVote != nil || b.ConflictingPreVote != nil {
		if !PrevBlockEqual(a.ConflictingPreVote, b.ConflictingPreVote) {
			t.Fatal("fail")
		}
	}
	if a.PreVoteNil != nil || b.PreVoteNil != nil {
		if !PrevBlockEqual(a.PreVoteNil, b.PreVoteNil) {
			t.Fatal("fail")
		}
	}
	if a.ImplicitPVN != b.ImplicitPVN {
		t.Fatal("fail")
	}
	if a.PreCommit != nil || b.PreCommit != nil {
		if !PrevBlockEqual(a.PreCommit, b.PreCommit) {
			t.Fatal("fail")
		}
	}
	if a.ConflictingPreCommit != nil || b.ConflictingPreCommit != nil {
		if !PrevBlockEqual(a.ConflictingPreCommit, b.ConflictingPreCommit) {
			t.Fatal("fail")
		}
	}
	if a.PreCommitNil != nil || b.PreCommitNil != nil {
		if !PrevBlockEqual(a.PreCommitNil, b.PreCommitNil) {
			t.Fatal("fail")
		}
	}
	if a.ImplicitPCN != b.ImplicitPCN {
		t.Fatal("fail")
	}

	if a.NextRound != nil || b.NextRound != nil {
		if !PrevBlockEqual(a.NextRound, b.NextRound) {
			t.Fatal("fail")
		}
	}
	if a.NextHeight != nil || b.NextHeight != nil {
		if !PrevBlockEqual(a.NextHeight, b.NextHeight) {
			t.Fatal("fail")
		}
	}
	if a.ConflictingNextHeight != nil || b.ConflictingNextHeight != nil {
		if !PrevBlockEqual(a.ConflictingNextHeight, b.ConflictingNextHeight) {
			t.Fatal("fail")
		}
	}
}

func generateRSChain(length int, seed []byte) ([]*BClaims, [][][]byte, error) {
	chain := []*BClaims{}
	txHashes := [][][]byte{}
	txhash := crypto.Hasher(bytes.Join([][]byte{[]byte(strconv.Itoa(1)), {}}, []byte{}))
	txHshLst := [][]byte{txhash}
	txRoot, err := MakeTxRoot(txHshLst)
	if err != nil {
		return nil, nil, err
	}
	txHashes = append(txHashes, txHshLst)
	bclaims := &BClaims{
		ChainID:    1,
		Height:     1,
		TxCount:    1,
		PrevBlock:  crypto.Hasher([]byte("foo")),
		TxRoot:     txRoot,
		StateRoot:  crypto.Hasher([]byte("")),
		HeaderRoot: crypto.Hasher([]byte("")),
	}
	chain = append(chain, bclaims)
	for i := 1; i < length; i++ {
		bhsh, err := chain[i-1].BlockHash()
		if err != nil {
			return nil, nil, err
		}
		txhash := crypto.Hasher(bytes.Join([][]byte{[]byte(strconv.Itoa(i)), seed}, []byte{}))
		txHshLst := [][]byte{txhash}
		txRoot, err := MakeTxRoot(txHshLst)
		if err != nil {
			return nil, nil, err
		}
		txHashes = append(txHashes, txHshLst)
		bclaims := &BClaims{
			ChainID:    1,
			Height:     uint32(len(chain) + 1),
			TxCount:    1,
			PrevBlock:  bhsh,
			TxRoot:     txRoot,
			StateRoot:  chain[i-1].StateRoot,
			HeaderRoot: chain[i-1].HeaderRoot,
		}
		chain = append(chain, bclaims)
	}
	return chain, txHashes, nil
}

func makeSigners(t *testing.T, num int) ([]*crypto.Secp256k1Signer, []*crypto.BNGroupSigner) {
	bnSigners := []*crypto.BNGroupSigner{}
	secpSigners := []*crypto.Secp256k1Signer{}
	for i := 0; i < num; i++ {
		secpSigner := &crypto.Secp256k1Signer{}
		err := secpSigner.SetPrivk(crypto.Hasher([]byte("secret" + strconv.Itoa(i))))
		if err != nil {
			t.Fatal(err)
		}
		secpSigners = append(secpSigners, secpSigner)
		bnSigner := &crypto.BNGroupSigner{}
		err = bnSigner.SetPrivk(crypto.Hasher([]byte("secret" + strconv.Itoa(i))))
		if err != nil {
			t.Fatal(err)
		}
		bnSigners = append(bnSigners, bnSigner)
	}
	return secpSigners, bnSigners
}

func mkBH(t *testing.T, bnSigner *crypto.BNGroupSigner, bclaims *BClaims, txHashList [][]byte) *BlockHeader {
	bhsh, err := bclaims.BlockHash()
	if err != nil {
		t.Fatal(err)
	}
	sig, err := bnSigner.Sign(bhsh)
	if err != nil {
		t.Fatal(err)
	}
	bh := &BlockHeader{
		BClaims:  bclaims,
		SigGroup: sig,
		TxHshLst: txHashList,
	}
	return bh
}

func mkP(t *testing.T, secpSigner *crypto.Secp256k1Signer, prevBH *BlockHeader, bh *BlockHeader) *Proposal {
	rcert, err := prevBH.GetRCert()
	if err != nil {
		t.Fatal(err)
	}
	pclms := &PClaims{
		BClaims: bh.BClaims,
		RCert:   rcert,
	}
	prop := &Proposal{
		PClaims:  pclms,
		TxHshLst: bh.TxHshLst,
	}
	err = prop.Sign(secpSigner)
	if err != nil {
		t.Fatal(err)
	}
	return prop
}

func mkPVL(t *testing.T, secpSigners []*crypto.Secp256k1Signer, prop *Proposal) PreVoteList {
	pvl := PreVoteList{}
	for _, signer := range secpSigners {
		pv, err := prop.PreVote(signer)
		if err != nil {
			t.Fatal(err)
		}
		pvl = append(pvl, pv)
	}
	return pvl
}

func mkPCL(t *testing.T, secpSigners []*crypto.Secp256k1Signer, pvl PreVoteList) PreCommitList {
	pcl := PreCommitList{}
	for _, signer := range secpSigners {
		pc, err := pvl.MakePreCommit(signer)
		if err != nil {
			t.Fatal(err)
		}
		pcl = append(pcl, pc)
	}
	return pcl
}

func mkNHL(t *testing.T, secpSigners []*crypto.Secp256k1Signer, bnSigners []*crypto.BNGroupSigner, pcl PreCommitList) NextHeightList {
	nhl := NextHeightList{}
	for idx, signer := range secpSigners {
		nh, err := pcl.MakeNextHeight(signer, bnSigners[idx])
		if err != nil {
			t.Fatal(err)
		}
		nhl = append(nhl, nh)
	}
	return nhl
}

func mkPVNL(t *testing.T, secpSigners []*crypto.Secp256k1Signer, prop *Proposal) []*PreVoteNil {
	pvnl := []*PreVoteNil{}
	for _, signer := range secpSigners {
		pvn, err := prop.PClaims.RCert.PreVoteNil(signer)
		if err != nil {
			t.Fatal(err)
		}
		pvnl = append(pvnl, pvn)
	}
	return pvnl
}

func mkPCN(t *testing.T, secpSigners []*crypto.Secp256k1Signer, prop *Proposal) []*PreCommitNil {
	pvnl := []*PreCommitNil{}
	for _, signer := range secpSigners {
		pvn, err := prop.PClaims.RCert.PreCommitNil(signer)
		if err != nil {
			t.Fatal(err)
		}
		pvnl = append(pvnl, pvn)
	}
	return pvnl
}

func mkNRL(t *testing.T, secpSigners []*crypto.Secp256k1Signer, bnSigners []*crypto.BNGroupSigner, prop *Proposal) NextRoundList {
	nrl := NextRoundList{}
	for idx, signer := range secpSigners {
		nr, err := prop.PClaims.RCert.NextRound(signer, bnSigners[idx])
		if err != nil {
			t.Fatal(err)
		}
		nrl = append(nrl, nr)
	}
	return nrl
}

func initRS(t *testing.T, idx uint8, secpSigner *crypto.Secp256k1Signer, bnSigner *crypto.BNGroupSigner, groupSigner *crypto.BNGroupSigner, rcert *RCert) *RoundState {
	secpPK, err := secpSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	groupShare, err := bnSigner.PubkeyShare()
	if err != nil {
		t.Fatal(err)
	}
	groupKey, err := groupSigner.PubkeyShare()
	if err != nil {
		t.Fatal(err)
	}
	rs := &RoundState{
		VAddr:      secpPK,
		GroupKey:   groupKey,
		GroupShare: groupShare,
		GroupIdx:   idx,
		RCert:      rcert,
	}
	return rs
}

func setup(t *testing.T) (*crypto.BNGroupSigner, []*crypto.Secp256k1Signer, []*crypto.BNGroupSigner, map[int][]*BlockHeader, map[int]*RoundState) {
	bhMap := make(map[int][]*BlockHeader)
	rsMap := make(map[int]*RoundState)
	_, groupSigners := makeSigners(t, 1)
	groupSigner := groupSigners[0]
	secpSigners, bnSigners := makeSigners(t, 4)
	bclaimsList1, txHashListList1, err := generateRSChain(10, []byte("s1"))
	if err != nil {
		t.Fatal(err)
	}
	for idx, bc := range bclaimsList1 {
		bh := mkBH(t, groupSigner, bc, txHashListList1[idx])
		bhMap[idx] = append(bhMap[idx], bh)
	}
	bclaimsList2, txHashListList2, err := generateRSChain(10, []byte("s2"))
	if err != nil {
		t.Fatal(err)
	}
	for idx, bc := range bclaimsList2 {
		bh := mkBH(t, groupSigner, bc, txHashListList2[idx])
		bhMap[idx] = append(bhMap[idx], bh)
	}
	for idx, bhl := range bhMap {
		if idx > 0 {
			rc, err := bhl[0].GetRCert()
			if err != nil {
				t.Fatal(err)
			}
			rs := initRS(t, 0, secpSigners[0], bnSigners[0], groupSigner, rc)
			rsMap[idx] = rs
		} else {
			rc := &RCert{
				SigGroup: bhl[0].SigGroup,
				RClaims: &RClaims{
					PrevBlock: bclaimsList1[0].PrevBlock,
					ChainID:   bclaimsList2[0].ChainID,
					Round:     1,
					Height:    bhl[0].BClaims.Height,
				},
			}
			rs := initRS(t, 0, secpSigners[0], bnSigners[0], groupSigner, rc)
			rsMap[idx] = rs
		}
	}
	for idx, bhList := range bhMap {
		ok, err := BClaimsEqual(bhList[0], bhList[1])
		if err != nil {
			t.Fatal(err)
		}
		if idx > 0 {
			if ok {
				t.Fatal("BlockHeaders are same - test setup logic broken for RoundClaims")
			}
		} else {
			if !ok {
				t.Fatal("BlockHeaders are not same for first block - test setup logic broken for RoundClaims")
			}
		}
	}
	return groupSigner, secpSigners, bnSigners, bhMap, rsMap
}

func TestConflictingProposal2(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop2 := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop2Bytes, err := prop2.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	prop2 = &Proposal{}
	err = prop2.UnmarshalBinary(prop2Bytes)
	if err != nil {
		t.Fatal(err)
	}
	prop2.PClaims.RCert.RClaims.PrevBlock = crypto.Hasher([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	prop2.PClaims.BClaims.PrevBlock = crypto.Hasher([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	ok, err := rsMap[0].SetProposal(prop)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Should be ok")
	}
	ok, err = rsMap[0].SetProposal(prop)
	if err == nil {
		t.Fatal("Should have raise error (1)")
	}
	if ok {
		t.Fatal("Should not be ok (1)")
	}
	ok, err = rsMap[0].SetProposal(prop2)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("Should not be ok (2)")
	}
	ok, err = rsMap[0].SetProposal(prop2)
	if err == nil {
		t.Fatal("Should have raise error (2)")
	}
	if ok {
		t.Fatal("Should not be ok (3)")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestConflictingProposal1(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop2 := mkP(t, secpSigners[0], bhMap[0][1], bhMap[1][1])
	ok, err := BClaimsEqual(prop, prop2)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("Should not be ok (1)")
	}
	ok, err = rsMap[0].SetProposal(prop)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Should be ok")
	}
	ok, err = rsMap[0].SetProposal(prop2)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("Should not be ok (2)")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestStaleProposal2(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[1][0], bhMap[2][0])
	prop2 := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop2.PClaims.BClaims.TxRoot = bhMap[1][1].BClaims.TxRoot
	prop2.TxHshLst = bhMap[1][1].TxHshLst
	ok, err := rsMap[0].SetProposal(prop)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Should be ok")
	}
	ok, err = rsMap[0].SetProposal(prop2)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if ok {
		t.Fatal("Should not be ok")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestStaleProposal3(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop.PClaims.RCert.RClaims.Round = 2
	prop2 := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	ok, err := rsMap[0].SetProposal(prop)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Should be ok")
	}
	ok, err = rsMap[0].SetProposal(prop2)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if ok {
		t.Fatal("Should not be ok")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestStaleProposal(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop2 := mkP(t, secpSigners[0], bhMap[0][1], bhMap[1][1])
	ok, err := rsMap[0].SetProposal(prop)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetProposal(prop2)
	if err != nil {
		if !errors.Is(err, &errorz.ErrStale{}) {
			t.Fatal("fail")
		}
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func TestConflictingPreVote(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop2 := mkP(t, secpSigners[0], bhMap[0][1], bhMap[1][1])
	pv, err := prop.PreVote(secpSigners[0])
	if err != nil {
		t.Fatal(err)
	}
	pv2, err := prop2.PreVote(secpSigners[0])
	if err != nil {
		t.Fatal(err)
	}
	ok, err := rsMap[0].SetPreVote(pv)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreVote(pv2)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestConflictingPreVote2(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop2 := mkP(t, secpSigners[0], bhMap[0][1], bhMap[1][1])
	prop2.PClaims.BClaims.TxRoot = bhMap[1][1].BClaims.TxRoot
	prop2.TxHshLst = bhMap[1][1].TxHshLst
	pv, err := prop.PreVote(secpSigners[0])
	if err != nil {
		t.Fatal(err)
	}
	pv2, err := prop2.PreVote(secpSigners[0])
	if err != nil {
		t.Fatal(err)
	}
	ok, err := rsMap[0].SetPreVote(pv)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreVote(pv2)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreVote(pv2)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestConflictingPreVote4(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[1][0], bhMap[2][0])
	prop2 := mkP(t, secpSigners[0], bhMap[1][1], bhMap[2][1])
	prop2.PClaims.BClaims.TxRoot = bhMap[1][1].BClaims.TxRoot
	prop2.TxHshLst = bhMap[1][1].TxHshLst
	pv, err := prop.PreVote(secpSigners[0])
	if err != nil {
		t.Fatal(err)
	}
	pv2, err := prop2.PreVote(secpSigners[0])
	if err != nil {
		t.Fatal(err)
	}
	ok, err := rsMap[0].SetPreVote(pv)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreVote(pv2)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreVote(pv2)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestStalePreVote2(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[1][0], bhMap[2][0])
	prop2 := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop2.PClaims.BClaims.TxRoot = bhMap[1][1].BClaims.TxRoot
	prop2.TxHshLst = bhMap[1][1].TxHshLst
	pv, err := prop.PreVote(secpSigners[0])
	if err != nil {
		t.Fatal(err)
	}
	pv2, err := prop2.PreVote(secpSigners[0])
	if err != nil {
		t.Fatal(err)
	}
	ok, err := rsMap[0].SetPreVote(pv)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreVote(pv2)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestStalePreVote3(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop.PClaims.RCert.RClaims.Round = 2
	prop2 := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	pv, err := prop.PreVote(secpSigners[0])
	if err != nil {
		t.Fatal(err)
	}
	pv2, err := prop2.PreVote(secpSigners[0])
	if err != nil {
		t.Fatal(err)
	}
	ok, err := rsMap[0].SetPreVote(pv)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreVote(pv2)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestStalePreVote(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop2 := mkP(t, secpSigners[0], bhMap[0][1], bhMap[1][1])
	pv, err := prop.PreVote(secpSigners[0])
	if err != nil {
		t.Fatal(err)
	}
	pv2, err := prop2.PreVote(secpSigners[0])
	if err != nil {
		t.Fatal(err)
	}
	ok, err := rsMap[0].SetPreVote(pv)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreVote(pv2)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func TestConflictingPreCommit(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop2 := mkP(t, secpSigners[0], bhMap[0][1], bhMap[1][1])

	pvl := mkPVL(t, secpSigners, prop)
	pvl2 := mkPVL(t, secpSigners, prop2)
	pcl := mkPCL(t, secpSigners, pvl)
	pcl2 := mkPCL(t, secpSigners, pvl2)
	ok, err := rsMap[0].SetPreCommit(pcl[0])
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreCommit(pcl2[0])
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestConflictingPreCommit2(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	pvnl := mkPVNL(t, secpSigners, prop)
	ok, err := rsMap[0].SetPreVoteNil(pvnl[0])
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	pvl := mkPVL(t, secpSigners, prop)
	pcl := mkPCL(t, secpSigners, pvl)
	ok, err = rsMap[0].SetPreCommit(pcl[0])
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestConflictingPreCommit3(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop2 := mkP(t, secpSigners[0], bhMap[0][1], bhMap[1][1])
	prop2.PClaims.BClaims.TxRoot = bhMap[1][1].BClaims.TxRoot
	prop2.TxHshLst = bhMap[1][1].TxHshLst

	pvl := mkPVL(t, secpSigners, prop)
	pvl2 := mkPVL(t, secpSigners, prop2)
	pcl := mkPCL(t, secpSigners, pvl)
	pcl2 := mkPCL(t, secpSigners, pvl2)
	ok, err := rsMap[0].SetPreCommit(pcl[0])
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreCommit(pcl2[0])
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreCommit(pcl2[0])
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestConflictingPreCommit4(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[1][0], bhMap[2][0])
	prop2 := mkP(t, secpSigners[0], bhMap[1][1], bhMap[2][1])
	prop2.PClaims.BClaims.TxRoot = bhMap[1][1].BClaims.TxRoot
	prop2.TxHshLst = bhMap[1][1].TxHshLst

	pvl := mkPVL(t, secpSigners, prop)
	pvl2 := mkPVL(t, secpSigners, prop2)
	pcl := mkPCL(t, secpSigners, pvl)
	pcl2 := mkPCL(t, secpSigners, pvl2)
	ok, err := rsMap[0].SetPreCommit(pcl[0])
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreCommit(pcl2[0])
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreCommit(pcl2[0])
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestStalePreCommit2(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[1][0], bhMap[2][0])
	prop2 := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop2.PClaims.BClaims.TxRoot = bhMap[1][1].BClaims.TxRoot
	prop2.TxHshLst = bhMap[1][1].TxHshLst

	pvl := mkPVL(t, secpSigners, prop)
	pvl2 := mkPVL(t, secpSigners, prop2)
	pcl := mkPCL(t, secpSigners, pvl)
	pcl2 := mkPCL(t, secpSigners, pvl2)
	ok, err := rsMap[0].SetPreCommit(pcl[0])
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreCommit(pcl2[0])
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestStalePreCommit3(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop.PClaims.RCert.RClaims.Round = 2
	prop2 := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])

	pvl := mkPVL(t, secpSigners, prop)
	pvl2 := mkPVL(t, secpSigners, prop2)
	pcl := mkPCL(t, secpSigners, pvl)
	pcl2 := mkPCL(t, secpSigners, pvl2)
	ok, err := rsMap[0].SetPreCommit(pcl[0])
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreCommit(pcl2[0])
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestStalePreCommit(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop2 := mkP(t, secpSigners[0], bhMap[0][1], bhMap[1][1])
	pvl := mkPVL(t, secpSigners, prop)
	pvl2 := mkPVL(t, secpSigners, prop2)
	pcl := mkPCL(t, secpSigners, pvl)
	pcl2 := mkPCL(t, secpSigners, pvl2)
	ok, err := rsMap[0].SetPreCommit(pcl[0])
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetPreCommit(pcl2[0])
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func TestConflictingNextHeight4(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[1][0], bhMap[2][0])
	prop2 := mkP(t, secpSigners[0], bhMap[1][1], bhMap[2][1])
	prop2.PClaims.BClaims.TxRoot = bhMap[1][1].BClaims.TxRoot
	prop2.TxHshLst = bhMap[1][1].TxHshLst
	pvl := mkPVL(t, secpSigners, prop)
	pvl2 := mkPVL(t, secpSigners, prop2)
	pcl := mkPCL(t, secpSigners, pvl)
	pcl2 := mkPCL(t, secpSigners, pvl2)
	nh := mkNHL(t, secpSigners, bnSigners, pcl)
	nh2 := mkNHL(t, secpSigners, bnSigners, pcl2)
	ok, err := rsMap[0].SetNextHeight(nh[0])
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetNextHeight(nh2[0])
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetNextHeight(nh2[0])
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestStaleNextHeight2(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[1][0], bhMap[2][0])
	prop2 := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop2.PClaims.BClaims.TxRoot = bhMap[1][1].BClaims.TxRoot
	prop2.TxHshLst = bhMap[1][1].TxHshLst

	pvl := mkPVL(t, secpSigners, prop)
	pvl2 := mkPVL(t, secpSigners, prop2)
	pcl := mkPCL(t, secpSigners, pvl)
	pcl2 := mkPCL(t, secpSigners, pvl2)
	nh := mkNHL(t, secpSigners, bnSigners, pcl)
	nh2 := mkNHL(t, secpSigners, bnSigners, pcl2)
	ok, err := rsMap[0].SetNextHeight(nh[0])
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetNextHeight(nh2[0])
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestStaleNextHeight3(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop.PClaims.RCert.RClaims.Round = 2
	prop2 := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])

	pvl := mkPVL(t, secpSigners, prop)
	pvl2 := mkPVL(t, secpSigners, prop2)
	pcl := mkPCL(t, secpSigners, pvl)
	pcl2 := mkPCL(t, secpSigners, pvl2)
	nh := mkNHL(t, secpSigners, bnSigners, pcl)
	nh2 := mkNHL(t, secpSigners, bnSigners, pcl2)
	ok, err := rsMap[0].SetNextHeight(nh[0])
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetNextHeight(nh2[0])
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

func TestStaleNextHeight(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	prop := mkP(t, secpSigners[0], bhMap[0][0], bhMap[1][0])
	prop2 := mkP(t, secpSigners[0], bhMap[0][1], bhMap[1][1])
	pvl := mkPVL(t, secpSigners, prop)
	pvl2 := mkPVL(t, secpSigners, prop2)
	pcl := mkPCL(t, secpSigners, pvl)
	pcl2 := mkPCL(t, secpSigners, pvl2)
	nh := mkNHL(t, secpSigners, bnSigners, pcl)
	nh2 := mkNHL(t, secpSigners, bnSigners, pcl2)
	ok, err := rsMap[0].SetNextHeight(nh[0])
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("fail")
	}
	ok, err = rsMap[0].SetNextHeight(nh2[0])
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("fail")
	}
	rsbytes, err := rsMap[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RoundState{}
	err = rs2.UnmarshalBinary(rsbytes)
	if err != nil {
		t.Fatal(err)
	}
	rsEqual(t, rsMap[0], rs2)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func TestNilProgress(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	for i := 0; i < len(bhMap)-1; i++ {
		prop := mkP(t, secpSigners[0], bhMap[i][0], bhMap[i+1][0])
		pvnl := mkPVNL(t, secpSigners, prop)
		pcnl := mkPCN(t, secpSigners, prop)
		nrl := mkNRL(t, secpSigners, bnSigners, prop)
		ok, err := rsMap[i].SetProposal(prop)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("fail")
		}
		ok, err = rsMap[i].SetPreVoteNil(pvnl[0])
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("fail")
		}
		ok, err = rsMap[i].SetPreCommitNil(pcnl[0])
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("fail")
		}
		ok, err = rsMap[i].SetNextRound(nrl[0])
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("fail")
		}
		rsbytes, err := rsMap[i].MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		rs2 := &RoundState{}
		err = rs2.UnmarshalBinary(rsbytes)
		if err != nil {
			t.Fatal(err)
		}
		rsEqual(t, rsMap[i], rs2)
	}
}

func TestNilProgressJump(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	for i := 1; i < len(bhMap)-3; i++ {
		prop := mkP(t, secpSigners[0], bhMap[i+1][0], bhMap[i+2][0])
		pvnl := mkPVNL(t, secpSigners, prop)
		pcnl := mkPCN(t, secpSigners, prop)
		nrl := mkNRL(t, secpSigners, bnSigners, prop)
		if i == 0 {
			ok, err := rsMap[i-1].SetProposal(prop)
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Fatal("fail")
			}
		}
		if i == 1 {
			ok, err := rsMap[i-1].SetPreVoteNil(pvnl[0])
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Fatal("fail")
			}
		}
		if i == 2 {
			ok, err := rsMap[i-1].SetPreCommitNil(pcnl[0])
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Fatal("fail")
			}
		}
		if i == 3 {
			nrl[0].NRClaims.RClaims.Round = 3
			nrl[0].NRClaims.RCert.RClaims.Round = 2
			ok, err := rsMap[i-1].SetNextRound(nrl[0])
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Fatal("fail")
			}
		}
		if i == 4 {
			ok, err := rsMap[i-1].SetNextRound(nrl[0])
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Fatal("fail")
			}
		}
	}
}

func TestProgress(t *testing.T) {
	groupSigner, secpSigners, bnSigners, bhMap, rsMap := setup(t)
	_ = bnSigners
	_ = groupSigner
	for i := 0; i < len(bhMap)-1; i++ {
		prop := mkP(t, secpSigners[0], bhMap[i][0], bhMap[i+1][0])
		pvl := mkPVL(t, secpSigners, prop)
		pcl := mkPCL(t, secpSigners, pvl)
		nhl := mkNHL(t, secpSigners, bnSigners, pcl)
		ok, err := rsMap[0].SetProposal(prop)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("fail")
		}
		ok, err = rsMap[0].SetPreVote(pvl[0])
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("fail")
		}
		ok, err = rsMap[0].SetPreCommit(pcl[0])
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("fail")
		}
		ok, err = rsMap[0].SetNextHeight(nhl[0])
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("fail")
		}
		rsbytes, err := rsMap[0].MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		rs2 := &RoundState{}
		err = rs2.UnmarshalBinary(rsbytes)
		if err != nil {
			t.Fatal(err)
		}
		rsEqual(t, rsMap[0], rs2)
	}
}
