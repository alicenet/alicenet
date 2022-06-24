package lstate

import (
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/dgraph-io/badger/v2"
	"testing"
)

//Not IsCurrentValidator
func TestStateCastLocal_castProposalFromValue_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(3)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	p := pl[0]

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		rss.OwnState.VAddr = make([]byte, 32)
		err := engine.castProposalFromValue(txn, rss, p)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateCastLocal_castProposalFromValue_Ok2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(3)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	p := pl[0]

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castProposalFromValue(txn, rss, p)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateCastLocal_castProposalFromValue_Error(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(3)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	rss.PeerStateMap[string(rss.OwnState.VAddr)].RCert.RClaims.Round = constants.DEADBLOCKROUND
	p := pl[0]
	p.PClaims.RCert.RClaims.Round = constants.DEADBLOCKROUND
	p.TxHshLst = make([][]byte, 1)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castProposalFromValue(txn, rss, p)
		if err == nil {
			t.Fatal("Should have raised error")
		}

		return nil
	})
}

//Not IsCurrentValidator
func TestStateCastLocal_castPreVoteWithLock_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(3)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	p := pl[0]

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		rss.OwnState.VAddr = make([]byte, 32)
		err := engine.castPreVoteWithLock(txn, rss, p, p)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

//Same BClaims
func TestStateCastLocal_castPreVoteWithLock_Ok2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(3)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	p := pl[0]

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castPreVoteWithLock(txn, rss, p, p)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

//Diff BClaims
func TestStateCastLocal_castPreVoteWithLock_Ok3(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(3)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	p1 := pl[0]
	p2 := pl[1]
	p2.PClaims.BClaims.Height++

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castPreVoteWithLock(txn, rss, p1, p2)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

//Not IsCurrentValidator
func TestStateCastLocal_castPreCommit_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(3)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		rss.OwnState.VAddr = make([]byte, 32)
		err := engine.castPreCommit(txn, rss, pvl)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateCastLocal_castPreCommit_Ok2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(3)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castPreCommit(txn, rss, pvl)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateCastLocal_castPreCommit_Error(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(3)
	round := uint32(4)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	pvl[0].Proposal = nil

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castPreCommit(txn, rss, pvl)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

//Not IsCurrentValidator
func TestStateCastLocal_castPreCommitNil_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		rss.OwnState.VAddr = make([]byte, 32)
		err := engine.castPreCommitNil(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateCastLocal_castPreCommitNil_Ok2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castPreCommitNil(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateCastLocal_castPreCommitNil_Error(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnRoundState().RCert = nil

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castPreCommitNil(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

//Not IsCurrentValidator
func TestStateCastLocal_castNextRound_Ok1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		rss.OwnState.VAddr = make([]byte, 32)
		err := engine.castNextRound(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateCastLocal_castNextRound_Ok2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	bnSigner := &crypto.BNGroupSigner{}
	err := bnSigner.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	engine.bnSigner = bnSigner

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err = engine.castNextRound(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateCastLocal_castNextRound_Error1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	rss.OwnRoundState().RCert = nil

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castNextRound(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateCastLocal_castNextRound_Error2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castNextRound(txn, rss)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

//Not IsCurrentValidator
func TestStateCastLocal_castNextRoundRCert_Ok1(t *testing.T) {
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
	_, _, _, _, _, _, nrl, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		rss.OwnState.VAddr = make([]byte, 32)
		err := engine.castNextRoundRCert(txn, rss, nrl)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateCastLocal_castNextRoundRCert_Ok2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}
	vs.Validators = vs.Validators[0:len(bnShares)]
	for i, bnShare := range bnShares {
		vs.Validators[i].GroupShare = bnShare
	}

	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, nrl, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	engine.bnSigner = bnSigners[0]

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castNextRoundRCert(txn, rss, nrl)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateCastLocal_castNextRoundRCert_Error1(t *testing.T) {
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
	_, _, _, _, _, _, nrl, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castNextRoundRCert(txn, rss, nrl)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

//Not IsCurrentValidator
func TestStateCastLocal_castNextHeightFromNextHeight_Ok1(t *testing.T) {
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

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		rss.OwnState.VAddr = make([]byte, 32)
		err := engine.castNextHeightFromNextHeight(txn, rss, nhl[0])
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateCastLocal_castNextHeightFromNextHeight_Ok2(t *testing.T) {
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
	engine.bnSigner = bnSigners[0]

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castNextHeightFromNextHeight(txn, rss, nhl[0])
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateCastLocal_castNextHeightFromNextHeight_Error1(t *testing.T) {
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

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castNextHeightFromNextHeight(txn, rss, nhl[0])
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

//Not IsCurrentValidator
func TestStateCastLocal_castNextHeightFromPreCommits_Ok1(t *testing.T) {
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

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		rss.OwnState.VAddr = make([]byte, 32)
		err := engine.castNextHeightFromPreCommits(txn, rss, pcl)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateCastLocal_castNextHeightFromPreCommits_Ok2(t *testing.T) {
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
	engine.bnSigner = bnSigners[0]

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castNextHeightFromPreCommits(txn, rss, pcl)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateCastLocal_castNextHeightFromPreCommits_Error1(t *testing.T) {
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

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castNextHeightFromPreCommits(txn, rss, pcl)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateCastLocal_castNewCommittedBlockHeader_Error1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}
	vs.Validators = vs.Validators[0:len(bnShares)]
	for i, bnShare := range bnShares {
		vs.Validators[i].GroupShare = bnShare
	}

	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, nhl, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	engine.bnSigner = bnSigners[0]

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castNewCommittedBlockHeader(txn, rss, nhl)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}

func TestStateCastLocal_castNewCommittedBlockFromProposalAndRCert_Ok(t *testing.T) {
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
	p.PClaims.BClaims.Height = 1

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castNewCommittedBlockFromProposalAndRCert(txn, rss, p, p.PClaims.RCert)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStateCastLocal_castNewCommittedBlockFromProposalAndRCert_Error(t *testing.T) {
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

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.castNewCommittedBlockFromProposalAndRCert(txn, rss, pl[0], pl[0].PClaims.RCert)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		return nil
	})
}
