package lstate

import (
	"errors"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/errorz"
	"testing"
)

func TestLocalState_setMostRecentProposal_Error1(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, p := createProposal(t)
	p.PClaims.RCert.RClaims.Round = constants.DEADBLOCKROUND

	err := engine.setMostRecentProposal(rss, p)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestLocalState_setMostRecentProposal_Error2(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, p := createProposal(t)
	p.PClaims.RCert.RClaims.Height = 1

	err := engine.setMostRecentProposal(rss, p)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if !errors.Is(err, errorz.ErrCorrupt) {
		t.Fatal("Should have raised errorz.ErrCorrupt error")
	}
}

func TestLocalState_setMostRecentPreVote_Error1(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pv := pvl[0]
	pv.Proposal.PClaims.RCert.RClaims.Round = constants.DEADBLOCKROUND
	pv.Proposal.TxHshLst = make([][]byte, 1)

	err := engine.setMostRecentPreVote(rss, pv)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestLocalState_setMostRecentPreVote_Error2(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(1)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pv := pvl[0]

	err := engine.setMostRecentPreVote(rss, pv)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if !errors.Is(err, errorz.ErrCorrupt) {
		t.Fatal("Should have raised errorz.ErrCorrupt error")
	}
}

func TestLocalState_setMostRecentPreVote_Ok(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(4)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pv := pvl[0]

	err := engine.setMostRecentPreVote(rss, pv)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
}

func TestLocalState_setMostRecentPreVoteNil_Error1(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, pvnl, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pvn := pvnl[0]
	pvn.RCert.RClaims.Round = constants.DEADBLOCKROUND

	err := engine.setMostRecentPreVoteNil(rss, pvn)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestLocalState_setMostRecentPreVoteNil_Error2(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(1)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, pvnl, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pvn := pvnl[0]

	err := engine.setMostRecentPreVoteNil(rss, pvn)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if !errors.Is(err, errorz.ErrCorrupt) {
		t.Fatal("Should have raised errorz.ErrCorrupt error")
	}
}

func TestLocalState_setMostRecentPreCommit_Error1(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Round = constants.DEADBLOCKROUND
	pc.Proposal.TxHshLst = make([][]byte, 1)

	err := engine.setMostRecentPreCommit(rss, pc)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestLocalState_setMostRecentPreCommit_Error2(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(1)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]

	err := engine.setMostRecentPreCommit(rss, pc)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if !errors.Is(err, errorz.ErrCorrupt) {
		t.Fatal("Should have raised errorz.ErrCorrupt error")
	}
}

func TestLocalState_setMostRecentPreCommit_Ok(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(4)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]

	err := engine.setMostRecentPreCommit(rss, pc)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
}

func TestLocalState_setMostRecentPreCommitNil_Error1(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, pcnl, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pcn := pcnl[0]
	pcn.RCert.RClaims.Round = constants.DEADBLOCKROUND

	err := engine.setMostRecentPreCommitNil(rss, pcn)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestLocalState_setMostRecentPreCommitNil_Error2(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(1)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, pcnl, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pcn := pcnl[0]

	err := engine.setMostRecentPreCommitNil(rss, pcn)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if !errors.Is(err, errorz.ErrCorrupt) {
		t.Fatal("Should have raised errorz.ErrCorrupt error")
	}
}

func TestLocalState_setMostRecentPreCommitNil_Ok(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(4)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, pcnl, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pcn := pcnl[0]

	err := engine.setMostRecentPreCommitNil(rss, pcn)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
}

func TestLocalState_setMostRecentNextRound_Error1(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, nrl, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	nr := nrl[0]
	nr.NRClaims.RCert.RClaims.Round = constants.DEADBLOCKROUND

	err := engine.setMostRecentNextRound(rss, nr)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestLocalState_setMostRecentNextRound_Error2(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(1)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, nrl, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	nr := nrl[0]

	err := engine.setMostRecentNextRound(rss, nr)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if !errors.Is(err, errorz.ErrCorrupt) {
		t.Fatal("Should have raised errorz.ErrCorrupt error")
	}
}

func TestLocalState_setMostRecentNextRound_Ok(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(4)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, nrl, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	nr := nrl[0]

	err := engine.setMostRecentNextRound(rss, nr)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
}

func TestLocalState_setMostRecentNextHeight_Error1(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, nhl, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	nh := nhl[0]
	nh.NHClaims.Proposal.PClaims.RCert.RClaims.Round = constants.DEADBLOCKROUND
	nh.NHClaims.Proposal.TxHshLst = make([][]byte, 1)

	err := engine.setMostRecentNextHeight(rss, nh)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestLocalState_setMostRecentNextHeight_Error2(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(1)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, nhl, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	nh := nhl[0]

	err := engine.setMostRecentNextHeight(rss, nh)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if !errors.Is(err, errorz.ErrCorrupt) {
		t.Fatal("Should have raised errorz.ErrCorrupt error")
	}
}

func TestLocalState_setMostRecentNextHeight_Ok(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(4)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, nhl, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	nh := nhl[0]

	err := engine.setMostRecentNextHeight(rss, nh)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
}

func TestLocalState_setMostRecentBlockHeaderFastSync_Error1(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	bh = nil

	err := engine.setMostRecentBlockHeaderFastSync(nil, rss, bh)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestLocalState_setMostRecentBlockHeaderFastSync_Ok(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(1024)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	err := engine.setMostRecentBlockHeaderFastSync(nil, rss, bh)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
}
