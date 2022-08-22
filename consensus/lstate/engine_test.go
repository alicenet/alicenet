package lstate

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appObjs "github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/consensus/admin"
	"github.com/alicenet/alicenet/consensus/appmock"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/dman"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/consensus/request"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/dynamics"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/proto"
	"github.com/alicenet/alicenet/utils"
)

func TestEngine_Status_Ok(t *testing.T) {
	st := make(map[string]interface{})
	engine := initEngine(t, nil)

	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.sstore.database.SetOwnState(txn, os)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.sstore.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	st, err := engine.Status(st)
	assert.Nil(t, err)
}

func TestEngine_Status_Error(t *testing.T) {
	st := make(map[string]interface{})
	engine := initEngine(t, nil)

	_, err := engine.Status(st)
	assert.NotNil(t, err)
}

// ce.ethAcct != ownState.VAddr.
func TestEngine_UpdateLocalState1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	isSync, err := engine.UpdateLocalState()
	assert.Nil(t, err)
	assert.True(t, isSync)
}

// ce.ethAcct == ownState.VAddr
// os val GetPrivK not found.
func TestEngine_UpdateLocalState2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	engine.ethAcct = os.VAddr

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	isSync, err := engine.UpdateLocalState()
	assert.Nil(t, err)
	assert.True(t, isSync)
}

// ce.ethAcct == ownState.VAddr
// os val GetPrivK found but pubk mismatch.
func TestEngine_UpdateLocalState3(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Should have raised panic: pubkey mismatch!")
		}
	}()

	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	engine.ethAcct = os.VAddr

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		osVal := vs.Validators[len(vs.Validators)-1]
		es := &objs.EncryptedStore{
			Name: osVal.GroupShare,
		}

		err = es.Encrypt(engine.AdminBus)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.database.SetEncryptedStore(txn, es)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	_, _ = engine.UpdateLocalState()
}

// ce.ethAcct == ownState.VAddr
// os val GetPrivK not found
// new validators set.
func TestEngine_UpdateLocalState4(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	engine.ethAcct = os.VAddr

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		groupSigner := &crypto.BNGroupSigner{}
		err = groupSigner.SetPrivk(crypto.Hasher([]byte("secret123")))
		if err != nil {
			t.Fatal(err)
		}
		groupKey, _ := groupSigner.PubkeyShare()

		vs.GroupKey = groupKey
		vs.NotBefore = 2
		err = engine.database.SetValidatorSetPostApplication(txn, vs, 1)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	isSync, err := engine.UpdateLocalState()
	assert.Nil(t, err)
	assert.True(t, isSync)
}

// updateLoadedObjects = OK
// updateLocalStateInternal = OK.
func TestEngine_UpdateLocalState5(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	engine.ethAcct = os.VAddr
	os.GroupKey = vs.GroupKey

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		osVal := vs.Validators[len(vs.Validators)-1]
		es := &objs.EncryptedStore{
			Name: osVal.GroupShare,
		}

		err = es.Encrypt(engine.AdminBus)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.database.SetEncryptedStore(txn, es)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		signer := &crypto.BNGroupSigner{}
		pk := utils.CopySlice(es.ClearText)
		err = signer.SetPrivk(pk)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		engine.bnSigner = signer

		return nil
	})

	isSync, err := engine.UpdateLocalState()
	assert.Nil(t, err)
	assert.True(t, isSync)
}

// updateLoadedObjects = OK
// updateLocalStateInternal = OK
// bHeight = 1024 and not safe to proceed.
func TestEngine_UpdateLocalState6(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1024)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	engine.ethAcct = os.VAddr
	os.GroupKey = vs.GroupKey

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		osVal := vs.Validators[len(vs.Validators)-1]
		es := &objs.EncryptedStore{
			Name: osVal.GroupShare,
		}

		err = es.Encrypt(engine.AdminBus)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.database.SetEncryptedStore(txn, es)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		signer := &crypto.BNGroupSigner{}
		pk := utils.CopySlice(es.ClearText)
		err = signer.SetPrivk(pk)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		engine.bnSigner = signer

		return nil
	})

	isSync, err := engine.UpdateLocalState()
	assert.Nil(t, err)
	assert.True(t, isSync)
}

// updateLoadedObjects = OK
// updateLocalStateInternal = OK
// bHeight = 1024 and safe to proceed.
func TestEngine_UpdateLocalState7(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1024)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	engine.ethAcct = os.VAddr
	os.GroupKey = vs.GroupKey

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		osVal := vs.Validators[len(vs.Validators)-1]
		es := &objs.EncryptedStore{
			Name: osVal.GroupShare,
		}

		err = es.Encrypt(engine.AdminBus)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.database.SetEncryptedStore(txn, es)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		signer := &crypto.BNGroupSigner{}
		pk := utils.CopySlice(es.ClearText)
		err = signer.SetPrivk(pk)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		engine.bnSigner = signer

		err = engine.database.SetSafeToProceed(txn, 1025, true)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	isSync, err := engine.UpdateLocalState()
	assert.Nil(t, err)
	assert.True(t, isSync)
}

// MaxBHSeen and SyncToBH same height.
func TestEngine_Sync1(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	engine.ethAcct = os.VAddr
	os.GroupKey = vs.GroupKey

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	isSync, err := engine.Sync()
	assert.Nil(t, err)
	assert.True(t, isSync)
}

// MaxBHSeen and SyncToBH diff height
// fastSync not done.
func TestEngine_Sync2(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 5800)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	engine.ethAcct = os.VAddr
	os.GroupKey = vs.GroupKey
	bhBytes, err := os.SyncToBH.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newBh := &objs.BlockHeader{}
	err = newBh.UnmarshalBinary(bhBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newBh.BClaims.Height = 4096
	os.SyncToBH = newBh

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err = engine.database.SetSnapshotBlockHeader(txn, newBh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		err = engine.fastSync.startFastSync(txn, newBh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		os.SyncToBH.BClaims.Height = 1024

		err = engine.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	isSync, err := engine.Sync()
	assert.Nil(t, err)
	assert.False(t, isSync)
}

// MaxBHSeen and SyncToBH diff height.
func TestEngine_Sync3(t *testing.T) {
	engine := initEngine(t, nil)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
	engine.ethAcct = os.VAddr
	os.GroupKey = vs.GroupKey
	bhBytes, err := os.SyncToBH.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newBh := &objs.BlockHeader{}
	err = newBh.UnmarshalBinary(bhBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newBh.BClaims.Height++
	os.SyncToBH = newBh

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		err := engine.sstore.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = engine.sstore.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})

	isSync, err := engine.Sync()
	assert.Nil(t, err)
	assert.False(t, isSync)
}

// Not current validator
// No future heights.
func TestEngine_updateLocalStateInternal1(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.Nil(t, err)
		assert.True(t, ok)

		return nil
	})
}

// round jump
// dead block round.
func TestEngine_updateLocalStateInternal2(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	rsBytes, err := rss.PeerStateMap[string(vs.Validators[0].VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS := &objs.RoundState{}
	err = newRS.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS.RCert.RClaims.Round = 5
	rss.PeerStateMap[string(vs.Validators[0].VAddr)] = newRS

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.Nil(t, err)
		assert.True(t, ok)

		return nil
	})
}

// peer with future height.
func TestEngine_updateLocalStateInternal3(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	rsBytes, err := rss.PeerStateMap[string(vs.Validators[0].VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS := &objs.RoundState{}
	err = newRS.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS.RCert.RClaims.Height++
	rss.PeerStateMap[string(vs.Validators[0].VAddr)] = newRS

	bhBytes, err := os.SyncToBH.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newBH := &objs.BlockHeader{}
	err = newBH.UnmarshalBinary(bhBytes)
	newBH.BClaims.Height = newRS.RCert.RClaims.Height

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		newBH.BClaims.Height = 1
		err = engine.database.SetCommittedBlockHeader(txn, newBH)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		newBH.BClaims.Height = newRS.RCert.RClaims.Height
		err = engine.database.SetCommittedBlockHeader(txn, newBH)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		rssOS := rss.OwnRoundState()
		ownRcert := rssOS.RCert
		txs := [][]byte{}
		TxRoot, err := objs.MakeTxRoot(txs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		bclaims := rss.OwnState.SyncToBH.BClaims
		PrevBlock := utils.CopySlice(ownRcert.RClaims.PrevBlock)
		headerRoot, err := engine.database.GetHeaderTrieRoot(txn, rss.OwnState.SyncToBH.BClaims.Height)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		StateRoot := utils.CopySlice(bclaims.StateRoot)
		bh := &objs.BlockHeader{
			BClaims: &objs.BClaims{
				PrevBlock:  PrevBlock,
				HeaderRoot: headerRoot,
				StateRoot:  StateRoot,
				TxRoot:     TxRoot,
				ChainID:    ownRcert.RClaims.ChainID,
				Height:     ownRcert.RClaims.Height,
			},
			TxHshLst: txs,
			SigGroup: utils.CopySlice(newRS.RCert.SigGroup),
		}

		blkHash, err := bh.BlockHash()
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		newRS.RCert.RClaims.PrevBlock = blkHash

		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.Nil(t, err)
		assert.True(t, ok)

		return nil
	})
}

// next round in round preceding the dead block round
// invalid proposal.
func TestEngine_updateLocalStateInternal4(t *testing.T) {
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
	_, _, _, _, pcl, _, nrl, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	ownRS := rss.PeerStateMap[string(rss.OwnState.VAddr)]
	ownRS.NextRound = nrl[0]
	ownRS.PreCommit = pcl[0]
	ownRS.RCert.RClaims.Height = 2
	ownRS.RCert.RClaims.Round = 4

	for _, v := range vs.Validators {
		rss.PeerStateMap[string(v.VAddr)] = ownRS
	}

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.NotNil(t, err)
		assert.False(t, ok)

		return nil
	})
}

// next height exist
// NHCurrent
// invalid validators shares.
func TestEngine_updateLocalStateInternal5(t *testing.T) {
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

	ownRS := rss.PeerStateMap[string(rss.OwnState.VAddr)]
	ownRS.RCert.RClaims.Height = 2
	ownRS.RCert.RClaims.Round = 4
	for _, v := range vs.Validators {
		rss.PeerStateMap[string(v.VAddr)].NextHeight = nhl[0]
	}
	engine.bnSigner = bnSigners[0]

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.NotNil(t, err)
		assert.False(t, ok)

		return nil
	})
}

// round jump.
func TestEngine_updateLocalStateInternal6(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	rsBytes, err := rss.PeerStateMap[string(vs.Validators[0].VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS := &objs.RoundState{}
	err = newRS.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS.RCert.RClaims.Round = 3
	rss.PeerStateMap[string(vs.Validators[0].VAddr)] = newRS

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.Nil(t, err)
		assert.True(t, ok)

		return nil
	})
}

// NRCurrent.
func TestEngine_updateLocalStateInternal7(t *testing.T) {
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
	_, _, _, _, _, _, nrl, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	rsBytes, err := rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS := &objs.RoundState{}
	err = newRS.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS.NextRound = nrl[0]
	newRS.RCert.RClaims.Height = 2
	newRS.RCert.RClaims.Round = 3
	rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)] = newRS

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.Nil(t, err)
		assert.True(t, ok)

		return nil
	})
}

// PCCurrent
// PCTOExpired.
func TestEngine_updateLocalStateInternal8(t *testing.T) {
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
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	rsBytes, err := rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS := &objs.RoundState{}
	err = newRS.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS.PreCommit = pcl[0]
	newRS.RCert.RClaims.Height = 2
	newRS.RCert.RClaims.Round = 3
	rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)] = newRS

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.Nil(t, err)
		assert.True(t, ok)

		return nil
	})
}

// PCCurrent
// NOT PCTOExpired.
func TestEngine_updateLocalStateInternal9(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{PreCommitStepStarted: time.Now().Add(1 * time.Hour).Unix()})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, pcl, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	rsBytes, err := rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS := &objs.RoundState{}
	err = newRS.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS.PreCommit = pcl[0]
	newRS.RCert.RClaims.Height = 2
	newRS.RCert.RClaims.Round = 3
	rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)] = newRS

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.Nil(t, err)
		assert.True(t, ok)

		return nil
	})
}

// PCNCurrent
// PCTOExpired.
func TestEngine_updateLocalStateInternal10(t *testing.T) {
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
	_, _, _, _, _, pcnl, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	rsBytes, err := rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS := &objs.RoundState{}
	err = newRS.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS.PreCommitNil = pcnl[0]
	newRS.RCert.RClaims.Height = 2
	newRS.RCert.RClaims.Round = 3
	rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)] = newRS

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.Nil(t, err)
		assert.True(t, ok)

		return nil
	})
}

// PCNCurrent
// NOT PCTOExpired.
func TestEngine_updateLocalStateInternal11(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{PreCommitStepStarted: time.Now().Add(1 * time.Hour).Unix()})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, pcnl, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	rsBytes, err := rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS := &objs.RoundState{}
	err = newRS.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS.PreCommitNil = pcnl[0]
	newRS.RCert.RClaims.Height = 2
	newRS.RCert.RClaims.Round = 3
	rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)] = newRS

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.Nil(t, err)
		assert.True(t, ok)

		return nil
	})
}

// NOT PVCurrent
// PVTOExpired.
func TestEngine_updateLocalStateInternal12(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	rsBytes, err := rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS := &objs.RoundState{}
	err = newRS.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}

	rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)] = newRS

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.Nil(t, err)
		assert.True(t, ok)

		return nil
	})
}

// NOT PVCurrent
// PVTOExpired.
func TestEngine_updateLocalStateInternal12_2(t *testing.T) {
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
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	rsBytes, err := rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS := &objs.RoundState{}
	err = newRS.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS.PreVote = pvl[0]
	rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)] = newRS

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.NotNil(t, err)
		assert.False(t, ok)

		return nil
	})
}

// NOT PVCurrent
// PVTOExpired.
func TestEngine_updateLocalStateInternal12_3(t *testing.T) {
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
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	rsBytes, err := rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS := &objs.RoundState{}
	err = newRS.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS.PreVote = pvl[0]
	newRS.RCert.RClaims.Height = 2
	newRS.RCert.RClaims.Round = 3
	rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)] = newRS

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.NotNil(t, err)
		assert.False(t, ok)

		return nil
	})
}

// PVCurrent
// NOT PVTOExpired.
func TestEngine_updateLocalStateInternal13(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{PreVoteStepStarted: time.Now().Add(1 * time.Hour).Unix()})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, pvl, _, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	rsBytes, err := rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS := &objs.RoundState{}
	err = newRS.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS.PreVote = pvl[0]
	newRS.RCert.RClaims.Height = 2
	newRS.RCert.RClaims.Round = 3
	rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)] = newRS

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.Nil(t, err)
		assert.True(t, ok)

		return nil
	})
}

// NOT PVNCurrent
// PVTOExpired.
func TestEngine_updateLocalStateInternal14(t *testing.T) {
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
	_, _, _, pvnl, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	rsBytes, err := rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS := &objs.RoundState{}
	err = newRS.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS.PreVoteNil = pvnl[0]
	rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)] = newRS

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.NotNil(t, err)
		assert.False(t, ok)

		return nil
	})
}

// NOT PVNCurrent
// PVTOExpired.
func TestEngine_updateLocalStateInternal14_2(t *testing.T) {
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
	_, _, _, pvnl, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	rsBytes, err := rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS := &objs.RoundState{}
	err = newRS.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS.PreVoteNil = pvnl[0]
	newRS.RCert.RClaims.Height = 2
	newRS.RCert.RClaims.Round = 3
	rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)] = newRS

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.NotNil(t, err)
		assert.False(t, ok)

		return nil
	})
}

// PVNCurrent
// NOT PVTOExpired.
func TestEngine_updateLocalStateInternal15(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{PreVoteStepStarted: time.Now().Add(1 * time.Hour).Unix()})

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, pvnl, _, _, _, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	rsBytes, err := rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)].MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS := &objs.RoundState{}
	err = newRS.UnmarshalBinary(rsBytes)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newRS.PreVoteNil = pvnl[0]
	newRS.RCert.RClaims.Height = 2
	newRS.RCert.RClaims.Round = 3
	rss.PeerStateMap[string(vs.Validators[len(vs.Validators)-1].VAddr)] = newRS

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.Nil(t, err)
		assert.True(t, ok)

		return nil
	})
}

// PTOExpired
// ISProposer
// NOT PCurrent.
func TestEngine_updateLocalStateInternal16(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{RoundStarted: time.Now().Add(1 * time.Hour).Unix()})

	rss.height = 5
	rss.round = 5

	bhBytes, err := os.SyncToBH.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}
	newBH := &objs.BlockHeader{}
	err = newBH.UnmarshalBinary(bhBytes)

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		newBH.BClaims.Height = 1
		err = engine.database.SetCommittedBlockHeader(txn, newBH)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		newBH.BClaims.Height++
		err = engine.database.SetCommittedBlockHeader(txn, newBH)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.Nil(t, err)
		assert.True(t, ok)

		return nil
	})
}

// Nothing to update.
func TestEngine_updateLocalStateInternal17(t *testing.T) {
	engine := initEngine(t, nil)

	os := createOwnState(t, 2)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{RoundStarted: time.Now().Add(1 * time.Hour).Unix()})

	_ = engine.sstore.database.Update(func(txn *badger.Txn) error {
		ok, err := engine.updateLocalStateInternal(txn, rss)
		assert.Nil(t, err)
		assert.True(t, ok)

		return nil
	})
}

func initEngine(t *testing.T, txs []*appObjs.Tx) *Engine {
	ctx := context.Background()
	logger := logging.GetLogger("test")

	rawEngineDb, err := utils.OpenBadger(ctx.Done(), "", true)
	if err != nil {
		t.Fatal(err)
	}
	engineDb := &db.Database{}
	engineDb.Init(rawEngineDb)

	p2pClientMock := &request.P2PClientMock{}
	p2pClientMock.On("GetSnapShotHdrNode", mock.Anything, mock.Anything, mock.Anything).Return(&proto.GetSnapShotHdrNodeResponse{}, nil)
	p2pClientMock.On("GetBlockHeaders", mock.Anything, mock.Anything, mock.Anything).Return(&proto.GetBlockHeadersResponse{}, nil)
	client := &request.Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})

	app := appmock.New()
	_, vv := createProposal(t)
	app.SetNextValidValue(vv)
	app.SetTxs(txs)

	secpSigner := &crypto.Secp256k1Signer{}
	err = secpSigner.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}

	bnSigner := &crypto.BNGroupSigner{}
	err = bnSigner.SetPrivk(crypto.Hasher([]byte("secret2")))
	if err != nil {
		t.Fatal(err)
	}

	adminBus := initAdminBus(t, logger, engineDb)
	storage := appObjs.MakeMockStorageGetter()
	reqBusViewMock := &dman.ReqBusViewMock{}
	dMan := &dman.DMan{}
	dMan.Init(engineDb, app, reqBusViewMock)

	engine := &Engine{}
	engine.Init(engineDb, dMan, app, secpSigner, adminBus, make([]byte, constants.HashLen), client, storage)

	return engine
}

func initAdminBus(t *testing.T, logger *logrus.Logger, db *db.Database) *admin.Handlers {
	app := appmock.New()
	s := initStorage(t, logger)

	handler := &admin.Handlers{}
	handler.Init(1, db, crypto.Hasher([]byte(config.Configuration.Validator.SymmetricKey)), app, make([]byte, constants.HashLen), s)

	return handler
}

func initStorage(t *testing.T, logger *logrus.Logger) *dynamics.Storage {
	s := &dynamics.Storage{}
	err := s.Init(&dynamics.MockRawDB{}, logger)
	if err != nil {
		t.Fatal(err)
	}

	return s
}
