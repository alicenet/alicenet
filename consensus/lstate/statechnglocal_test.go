package lstate

import (
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

//Not IsCurrentValidator
func TestStateChngLocal_doPendingProposalStep_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnState.VAddr = make([]byte, 32)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingProposalStep(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

//Deadblock round
func TestStateChngLocal_doPendingProposalStep_Ok2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnRoundState().RCert.RClaims.Round = constants.DEADBLOCKROUND

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingProposalStep(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateChngLocal_doPendingProposalStep_Ok3(t *testing.T) {
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
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	p := pl[0]
	rss.OwnRoundState().RCert.RClaims.Height = p.PClaims.RCert.RClaims.Height
	rss.OwnValidatingState.ValidValue = p

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingProposalStep(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateChngLocal_doPendingProposalStep_Ok4(t *testing.T) {
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
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	p := pl[0]
	rss.OwnRoundState().RCert.RClaims.Height = p.PClaims.RCert.RClaims.Height
	rss.OwnValidatingState.ValidValue = p
	rss.OwnValidatingState.LockedValue = p

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingProposalStep(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateChngLocal_doPendingProposalStep_Error1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingProposalStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

//Not IsCurrentValidator
func TestStateChngLocal_doPendingPreVoteStep_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnState.VAddr = make([]byte, 32)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingPreVoteStep(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

//Deadblock round
func TestStateChngLocal_doPendingPreVoteStep_Error1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnRoundState().RCert.RClaims.Round = constants.DEADBLOCKROUND

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingPreVoteStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

//Deadblock round
func TestStateChngLocal_doPendingPreVoteStep_Error2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnRoundState().RCert.RClaims.Round = constants.DEADBLOCKROUND

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.database.SetCommittedBlockHeader(txn, rss.OwnState.SyncToBH)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.doPendingPreVoteStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

//invalid proposal
func TestStateChngLocal_doPendingPreVoteStep_Error3(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, vv := createProposal(t)
	vv.PClaims.RCert.RClaims.Height = 1

	rss.PeerStateMap[string(rss.ValidatorSet.Validators[1].VAddr)].Proposal = vv

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingPreVoteStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doPendingPreVoteStep_Error4(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, vv := createProposal(t)
	vv.PClaims.RCert.RClaims.Height = 1

	rss.PeerStateMap[string(rss.ValidatorSet.Validators[1].VAddr)].Proposal = vv
	rss.OwnRoundState().RCert.RClaims.Height = vv.PClaims.RCert.RClaims.Height
	rss.OwnValidatingState.ValidValue = vv

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingPreVoteStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doPendingPreVoteStep_Error5(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, vv := createProposal(t)
	vv.PClaims.RCert.RClaims.Height = 1

	rss.PeerStateMap[string(rss.ValidatorSet.Validators[1].VAddr)].Proposal = vv
	rss.OwnRoundState().RCert.RClaims.Height = vv.PClaims.RCert.RClaims.Height
	rss.OwnValidatingState.ValidValue = vv
	rss.OwnValidatingState.LockedValue = vv

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingPreVoteStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

//Not IsCurrentValidator
func TestStateChngLocal_doPreVoteStep_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnState.VAddr = make([]byte, 32)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPreVoteStep(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateChngLocal_doPreVoteStep_Error1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pv := pvl[0]
	pv.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pv.Proposal.PClaims.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreVote = pv
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPreVoteStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

//Not IsCurrentValidator
func TestStateChngLocal_doPreVoteNilStep_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnState.VAddr = make([]byte, 32)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPreVoteNilStep(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateChngLocal_doPreVoteNilStep_Error1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pv := pvl[0]
	pv.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pv.Proposal.PClaims.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreVote = pv
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPreVoteNilStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doPreVoteNilStep_Error2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pv := pvl[0]
	pv.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pv.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pv.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreVote = pv
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPreVoteNilStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doPreVoteNilStep_Error3(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, pvnl, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pv := pvl[0]
	pvn := pvnl[0]
	pv.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pv.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pv.Proposal.PClaims.BClaims.Height = rss.height
	pvn.RCert.RClaims.Height = rss.height
	pvn.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreVoteNil = pvn
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPreVoteNilStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

//Not IsCurrentValidator
func TestStateChngLocal_doPendingPreCommit_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnState.VAddr = make([]byte, 32)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingPreCommit(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateChngLocal_doPendingPreCommit_Error1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pv := pvl[0]
	pv.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pv.Proposal.PClaims.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreVote = pv
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingPreCommit(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doPendingPreCommit_Error2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pv := pvl[0]
	pv.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pv.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pv.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreVote = pv
	}

	rsBytes, err := rss.PeerStateMap[string(rss.OwnState.VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	rsNew := &objs.RoundState{}
	err = rsNew.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	pv1 := pvl[1]
	rsNew.PreVote = pv1
	rss.PeerStateMap[string(rss.OwnState.VAddr)] = rsNew

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err = engine.doPendingPreCommit(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doPendingPreCommit_Error3(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, pvnl, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pvn := pvnl[0]
	pvn.RCert.RClaims.Height = rss.height
	pvn.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreVoteNil = pvn
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingPreCommit(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

//Not IsCurrentValidator
func TestStateChngLocal_doPreCommitStep_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnState.VAddr = make([]byte, 32)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPreCommitStep(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateChngLocal_doPreCommitStep_Error1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPreCommitStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doPreCommitStep_Error2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pc.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPreCommitStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doPreCommitStep_Error3(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pc.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		bh.BClaims.Height = 1
		err := engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		bh.BClaims.Height++
		err = engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		goodHeaderRoot, err := engine.database.GetHeaderTrieRoot(txn, 1)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		pc.Proposal.PClaims.BClaims.HeaderRoot = goodHeaderRoot
		err = engine.doPreCommitStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

//Not IsCurrentValidator
func TestStateChngLocal_doPreCommitNilStep_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnState.VAddr = make([]byte, 32)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPreCommitNilStep(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateChngLocal_doPreCommitNilStep_Error1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPreCommitNilStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doPreCommitNilStep_Error2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pc.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPreCommitNilStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doPreCommitNilStep_Error3(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pc.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		bh.BClaims.Height = 1
		err := engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		bh.BClaims.Height++
		err = engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		goodHeaderRoot, err := engine.database.GetHeaderTrieRoot(txn, 1)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		pc.Proposal.PClaims.BClaims.HeaderRoot = goodHeaderRoot
		err = engine.doPreCommitNilStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doPreCommitNilStep_Error4(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, pcnl, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pcn := pcnl[0]
	pcn.RCert.RClaims.Height = rss.height
	pcn.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommitNil = pcn
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPreCommitNilStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

//Not IsCurrentValidator
func TestStateChngLocal_doPendingNext_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnState.VAddr = make([]byte, 32)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingNext(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateChngLocal_doPendingNext_Error1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingNext(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doPendingNext_Error2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pc.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingNext(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doPendingNext_Error3(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pc.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		bh.BClaims.Height = 1
		err := engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		bh.BClaims.Height++
		err = engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		goodHeaderRoot, err := engine.database.GetHeaderTrieRoot(txn, 1)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		pc.Proposal.PClaims.BClaims.HeaderRoot = goodHeaderRoot
		err = engine.doPendingNext(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doPendingNext_Error4(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, pcnl, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pcn := pcnl[0]
	pcn.RCert.RClaims.Height = rss.height
	pcn.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommitNil = pcn
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingNext(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doPendingNext_Error5(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnRoundState().RCert.RClaims.Round = constants.DEADBLOCKROUNDNR
	rss.round = constants.DEADBLOCKROUNDNR

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, pcnl, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pcn := pcnl[0]
	pcn.RCert.RClaims.Height = rss.height
	pcn.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommitNil = pcn
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doPendingNext(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

//Not IsCurrentValidator
func TestStateChngLocal_doNextRoundStep_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnState.VAddr = make([]byte, 32)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doNextRoundStep(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateChngLocal_doNextRoundStep_Error1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doNextRoundStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doNextRoundStep_Error2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pc.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		bh.BClaims.Height = 1
		err := engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		bh.BClaims.Height++
		err = engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		goodHeaderRoot, err := engine.database.GetHeaderTrieRoot(txn, 1)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		pc.Proposal.PClaims.BClaims.HeaderRoot = goodHeaderRoot
		err = engine.doNextRoundStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doNextRoundStep_Error3(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, nrl, nhl, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	nr := nrl[0]
	nr.NRClaims.RCert.RClaims.Height = rss.height
	nr.NRClaims.RCert.RClaims.Round = rss.round
	nr.NRClaims.RClaims.Height = rss.height
	nr.NRClaims.RClaims.Round = rss.round

	nh := nhl[0]
	nh.NHClaims.Proposal.PClaims.RCert.RClaims.Height = rss.height
	nh.NHClaims.Proposal.PClaims.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].NextRound = nr
		rss.PeerStateMap[k].NextHeight = nh
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doNextRoundStep(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doRoundJump_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pc.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		bh.BClaims.Height = 1
		err := engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		bh.BClaims.Height++
		err = engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		goodHeaderRoot, err := engine.database.GetHeaderTrieRoot(txn, 1)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		pc.Proposal.PClaims.BClaims.HeaderRoot = goodHeaderRoot
		err = engine.doRoundJump(txn, rss, pc.Proposal.PClaims.RCert)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateChngLocal_doRoundJump_Error1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doRoundJump(txn, rss, pc.Proposal.PClaims.RCert)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doRoundJump_Error2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pc.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doRoundJump(txn, rss, pc.Proposal.PClaims.RCert)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doCheckValidValue_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pc.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		bh.BClaims.Height = 1
		err := engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		bh.BClaims.Height++
		err = engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		goodHeaderRoot, err := engine.database.GetHeaderTrieRoot(txn, 1)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		pc.Proposal.PClaims.BClaims.HeaderRoot = goodHeaderRoot
		err = engine.doCheckValidValue(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateChngLocal_doCheckValidValue_Ok2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pv := pvl[0]
	pv.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pv.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pv.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreVote = pv
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		bh.BClaims.Height = 1
		err := engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		bh.BClaims.Height++
		err = engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		goodHeaderRoot, err := engine.database.GetHeaderTrieRoot(txn, 1)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		pv.Proposal.PClaims.BClaims.HeaderRoot = goodHeaderRoot
		err = engine.doCheckValidValue(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateChngLocal_doCheckValidValue_Error1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, nhl, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	nh := nhl[0]
	nh.NHClaims.Proposal.PClaims.RCert.RClaims.Height = rss.height
	nh.NHClaims.Proposal.PClaims.RCert.RClaims.Round = rss.round
	nh.NHClaims.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].NextHeight = nh
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doCheckValidValue(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doCheckValidValue_Error2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, nhl, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	nh := nhl[0]
	nh.NHClaims.Proposal.PClaims.RCert.RClaims.Height = rss.height
	nh.NHClaims.Proposal.PClaims.RCert.RClaims.Round = rss.round
	nh.NHClaims.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].NextHeight = nh
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		bh.BClaims.Height = 1
		err := engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		bh.BClaims.Height++
		err = engine.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		goodHeaderRoot, err := engine.database.GetHeaderTrieRoot(txn, 1)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		nh.NHClaims.Proposal.PClaims.BClaims.HeaderRoot = goodHeaderRoot
		err = engine.doCheckValidValue(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doCheckValidValue_Error3(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doCheckValidValue(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doCheckValidValue_Error4(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pc := pcl[0]
	pc.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pc.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pc.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreCommit = pc
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doCheckValidValue(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doCheckValidValue_Error5(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pv := pvl[0]
	pv.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pv.Proposal.PClaims.RCert.RClaims.Round = rss.round

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreVote = pv
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doCheckValidValue(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doCheckValidValue_Error6(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pv := pvl[0]
	pv.Proposal.PClaims.RCert.RClaims.Height = rss.height
	pv.Proposal.PClaims.RCert.RClaims.Round = rss.round
	pv.Proposal.PClaims.BClaims.Height = rss.height

	for k := range rss.PeerStateMap {
		rss.PeerStateMap[k].PreVote = pv
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.doCheckValidValue(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateChngLocal_doHeightJumpStep_False1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	p := pl[0]
	p.PClaims.RCert.RClaims.Height = rss.height
	p.PClaims.RCert.RClaims.Round = rss.round
	p.PClaims.BClaims.Height = rss.height

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.doHeightJumpStep(txn, rss, p.PClaims.RCert)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		assert.False(t, ok)

		return nil
	})
}

func TestStateChngLocal_doHeightJumpStep_False2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	p := pl[0]
	p.PClaims.RCert.RClaims.Height = 2
	rss.OwnRoundState().RCert.RClaims.Height = 1

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.doHeightJumpStep(txn, rss, p.PClaims.RCert)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		assert.False(t, ok)

		return nil
	})
}

func TestStateChngLocal_doHeightJumpStep_True3(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	p1 := pl[0]

	p2 := pl[1]
	p2.PClaims.RCert.RClaims.Height = rss.height
	p2.PClaims.RCert.RClaims.Round = rss.round
	p2.PClaims.BClaims.Height = rss.height
	rss.OwnValidatingState.ValidValue = p2

	bhsh, err := rss.ValidValue().PClaims.BClaims.BlockHash()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	p1.PClaims.RCert.RClaims.PrevBlock = bhsh

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.doHeightJumpStep(txn, rss, p1.PClaims.RCert)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		assert.True(t, ok)

		return nil
	})
}
