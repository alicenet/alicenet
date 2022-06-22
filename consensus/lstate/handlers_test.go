package lstate

import (
	"context"
	"github.com/alicenet/alicenet/consensus/appmock"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/dman"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestHandlers_AddProposal_Ok(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	p := pl[0]
	btsP, err := p.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	p2 := &objs.Proposal{}
	err = p2.UnmarshalBinary(btsP)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	err = p2.ValidateSignatures(hdlr.secpVal, hdlr.bnVal)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	vs.ValidatorVAddrMap[string(p2.Proposer)] = len(vs.Validators)
	vs.ValidatorVAddrSet[string(p2.Proposer)] = true
	vs.ValidatorGroupShareMap[string(p2.Proposer)] = len(vs.Validators)
	vs.ValidatorGroupShareSet[string(p2.Proposer)] = true
	gShare := crypto.Hasher([]byte("g0"))
	pVal := &objs.Validator{
		VAddr:      p2.Proposer,
		GroupShare: gShare,
	}
	vs.Validators[4] = pVal

	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	p.Proposer = pVal.VAddr

	err = hdlr.PreValidate(p)
	assert.Nil(t, err)
	err = hdlr.AddProposal(p)
	assert.Nil(t, err)
}

func TestHandlers_AddProposal_Error(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	p := pl[0]

	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	p.Proposer = os.VAddr

	err := hdlr.PreValidate(p)
	assert.NotNil(t, err)
}

func TestHandlers_AddPreVote_Ok(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	p := pl[0]
	btsP, err := p.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	p2 := &objs.Proposal{}
	err = p2.UnmarshalBinary(btsP)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	err = p2.ValidateSignatures(hdlr.secpVal, hdlr.bnVal)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	vs.ValidatorVAddrMap[string(p2.Proposer)] = len(vs.Validators)
	vs.ValidatorVAddrSet[string(p2.Proposer)] = true
	vs.ValidatorGroupShareMap[string(p2.Proposer)] = len(vs.Validators)
	vs.ValidatorGroupShareSet[string(p2.Proposer)] = true
	gShare := crypto.Hasher([]byte("g0"))
	pVal := &objs.Validator{
		VAddr:      p2.Proposer,
		GroupShare: gShare,
	}
	vs.Validators[4] = pVal

	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	p.Proposer = pVal.VAddr
	pv := pvl[0]
	pv.Proposal = p
	pv.Voter = os.VAddr

	err = hdlr.PreValidate(pv)
	assert.Nil(t, err)
	err = hdlr.AddPreVote(pv)
	assert.Nil(t, err)
}

func TestHandlers_AddPreVote_Error(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	p := pl[0]

	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	p.Proposer = os.VAddr
	pv := pvl[0]
	pv.Proposal = p
	pv.Voter = os.VAddr

	err := hdlr.PreValidate(pv)
	assert.NotNil(t, err)
}

func TestHandlers_AddPreVoteNil(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, pvnl, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	pvn := pvnl[0]
	btsPvn, err := pvn.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	pvn2 := &objs.PreVoteNil{}
	err = pvn2.UnmarshalBinary(btsPvn)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	err = pvn2.ValidateSignatures(hdlr.secpVal, hdlr.bnVal)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	vs.ValidatorVAddrMap[string(pvn2.Voter)] = len(vs.Validators)
	vs.ValidatorVAddrSet[string(pvn2.Voter)] = true
	vs.ValidatorGroupShareMap[string(pvn2.Voter)] = len(vs.Validators)
	vs.ValidatorGroupShareSet[string(pvn2.Voter)] = true
	gShare := crypto.Hasher([]byte("g0"))
	pVal := &objs.Validator{
		VAddr:      pvn2.Voter,
		GroupShare: gShare,
	}
	vs.Validators[4] = pVal

	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	pvn.Voter = pVal.VAddr

	err = hdlr.PreValidate(pvn)
	assert.Nil(t, err)
	err = hdlr.AddPreVoteNil(pvn)
	assert.Nil(t, err)
}

func TestHandlers_AddPreCommit_Ok(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	pc := pcl[0]
	btsPc, err := pc.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	pc2 := &objs.PreCommit{}
	err = pc2.UnmarshalBinary(btsPc)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	err = pc2.ValidateSignatures(hdlr.secpVal, hdlr.bnVal)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	vs.ValidatorVAddrMap[string(pc2.Voter)] = len(vs.Validators)
	vs.ValidatorVAddrSet[string(pc2.Voter)] = true
	vs.ValidatorGroupShareMap[string(pc2.Voter)] = len(vs.Validators)
	vs.ValidatorGroupShareSet[string(pc2.Voter)] = true
	gShare := crypto.Hasher([]byte("g0"))
	pVal := &objs.Validator{
		VAddr:      pc2.Voter,
		GroupShare: gShare,
	}
	vs.Validators[4] = pVal

	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	p := pl[0]
	p.Proposer = pVal.VAddr
	pc.Proposal = p
	pc.Proposer = pVal.VAddr
	pc.Voter = pVal.VAddr

	pvl, err := pc.MakeImplPreVotes()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	for _, pv := range pvl {
		rss.PeerStateMap[string(pv.Voter)] = rss.PeerStateMap[string(os.VAddr)]
		share := crypto.Hasher([]byte("g4"))
		validator := &objs.Validator{
			VAddr:      pv.Voter,
			GroupShare: share,
		}
		rss.ValidatorSet.Validators = append(rss.ValidatorSet.Validators, validator)
	}

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	err = hdlr.PreValidate(pc)
	assert.Nil(t, err)
	err = hdlr.AddPreCommit(pc)
	assert.Nil(t, err)
}

func TestHandlers_AddPreCommit_Error(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	pc := pcl[0]

	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	p := pl[0]
	p.Proposer = os.VAddr
	pc.Proposal = p
	pc.Proposer = os.VAddr
	pc.Voter = os.VAddr

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	err := hdlr.PreValidate(pc)
	assert.NotNil(t, err)
}

func TestHandlers_AddPreCommitNil_Ok(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, pcnl, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	pcn := pcnl[0]
	btsPcn, err := pcn.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	pcn2 := &objs.PreCommitNil{}
	err = pcn2.UnmarshalBinary(btsPcn)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	err = pcn2.ValidateSignatures(hdlr.secpVal, hdlr.bnVal)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	vs.ValidatorVAddrMap[string(pcn2.Voter)] = len(vs.Validators)
	vs.ValidatorVAddrSet[string(pcn2.Voter)] = true
	vs.ValidatorGroupShareMap[string(pcn2.Voter)] = len(vs.Validators)
	vs.ValidatorGroupShareSet[string(pcn2.Voter)] = true
	gShare := crypto.Hasher([]byte("g0"))
	pVal := &objs.Validator{
		VAddr:      pcn2.Voter,
		GroupShare: gShare,
	}
	vs.Validators[4] = pVal
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	pcn.Voter = pVal.VAddr

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	err = hdlr.PreValidate(pcn)
	assert.Nil(t, err)
	err = hdlr.AddPreCommitNil(pcn)
	assert.Nil(t, err)
}

func TestHandlers_AddPreCommitNil_Error(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, pcnl, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	pcn := pcnl[0]
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	pcn.Voter = os.VAddr

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	err := hdlr.PreValidate(pcn)
	assert.NotNil(t, err)
}

func TestHandlers_AddNextRound_Ok(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, nrl, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	nr := nrl[0]
	btsNr, err := nr.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	nr2 := &objs.NextRound{}
	err = nr2.UnmarshalBinary(btsNr)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	err = nr2.ValidateSignatures(hdlr.secpVal, hdlr.bnVal)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	vs.ValidatorVAddrMap[string(nr2.Voter)] = len(vs.Validators)
	vs.ValidatorVAddrSet[string(nr2.Voter)] = true
	vs.ValidatorGroupShareMap[string(nr2.GroupShare)] = len(vs.Validators)
	vs.ValidatorGroupShareSet[string(nr2.GroupShare)] = true
	pVal := &objs.Validator{
		VAddr:      nr2.Voter,
		GroupShare: nr2.GroupShare,
	}
	vs.Validators[4] = pVal

	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	nr.Voter = pVal.VAddr

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	err = hdlr.PreValidate(nr)
	assert.Nil(t, err)
	err = hdlr.AddNextRound(nr)
	assert.Nil(t, err)
}

func TestHandlers_AddNextRound_Error(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, nrl, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	nr := nrl[0]
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	nr.Voter = os.VAddr

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	err := hdlr.PreValidate(nr)
	assert.NotNil(t, err)
}

func TestHandlers_AddNextHeight_Ok(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, nhl, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)

	nh := nhl[0]
	btsNh, err := nh.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	nh2 := &objs.NextHeight{}
	err = nh2.UnmarshalBinary(btsNh)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	err = nh2.ValidateSignatures(hdlr.secpVal, hdlr.bnVal)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}

	vs.ValidatorVAddrMap[string(nh2.Voter)] = len(vs.Validators)
	vs.ValidatorVAddrSet[string(nh2.Voter)] = true
	vs.ValidatorGroupShareMap[string(nh2.GroupShare)] = len(vs.Validators)
	vs.ValidatorGroupShareSet[string(nh2.GroupShare)] = true
	pVal := &objs.Validator{
		VAddr:      nh2.Voter,
		GroupShare: nh2.GroupShare,
	}
	vs.Validators[4] = pVal

	for i, cs := range nh2.Signers {
		if vs.ValidatorVAddrSet[string(cs)] {
			continue
		}
		gShare := crypto.Hasher([]byte("g" + strconv.Itoa(i)))
		vs.ValidatorVAddrSet[string(cs)] = true
		vs.ValidatorVAddrMap[string(cs)] = len(vs.Validators)
		vs.ValidatorGroupShareMap[string(gShare)] = len(vs.Validators)
		vs.ValidatorGroupShareSet[string(gShare)] = true
		val := &objs.Validator{
			VAddr:      cs,
			GroupShare: gShare,
		}
		vs.Validators = append(vs.Validators, val)
	}

	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	nh.Voter = pVal.VAddr

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	err = hdlr.PreValidate(nh)
	assert.Nil(t, err)
	err = hdlr.AddNextHeight(nh)
	assert.Nil(t, err)
}

func TestHandlers_AddNextHeight_Error(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, nhl, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)

	nh := nhl[0]
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	nh.Voter = os.VAddr

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	err := hdlr.PreValidate(nh)
	assert.NotNil(t, err)
}

func TestHandlers_AddBlockHeader_Error1(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	err := hdlr.AddBlockHeader(bh)
	assert.NotNil(t, err)
}

func TestHandlers_AddBlockHeader_Error2(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(5)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	err := hdlr.PreValidate(bh)
	assert.NotNil(t, err)
}

func TestHandlers_AddBlockHeader_Ok(t *testing.T) {
	hdlr := initHandlers(t)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(5)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	os := createOwnState(t, 3)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	btsBh, err := bh.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	bh2 := &objs.BlockHeader{}
	err = bh2.UnmarshalBinary(btsBh)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	err = bh2.ValidateSignatures(hdlr.bnVal)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	vs.GroupKey = bh2.GroupKey

	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = hdlr.sstore.database.Update(func(txn *badger.Txn) error {
		err := hdlr.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = hdlr.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	err = hdlr.PreValidate(bh)
	assert.Nil(t, err)
	err = hdlr.AddBlockHeader(bh)
	assert.Nil(t, err)
}

func initHandlers(t *testing.T) *Handlers {
	rawHdlrsDb, err := utils.OpenBadger(context.Background().Done(), "", true)
	if err != nil {
		t.Fatal(err)
	}
	hdlrDb := &db.Database{}
	hdlrDb.Init(rawHdlrsDb)

	app := appmock.New()
	_, vv := createProposal(t)
	app.SetNextValidValue(vv)

	reqBusViewMock := &dman.ReqBusViewMock{}
	dMan := &dman.DMan{}
	dMan.Init(hdlrDb, app, reqBusViewMock)

	hdlr := &Handlers{}
	hdlr.Init(hdlrDb, dMan)

	return hdlr
}
