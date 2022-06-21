package lstate

import (
	"context"
	appObjs "github.com/alicenet/alicenet/application/objs"
	trie "github.com/alicenet/alicenet/badgerTrie"
	"github.com/alicenet/alicenet/consensus/appmock"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/request"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/dynamics"
	"github.com/alicenet/alicenet/proto"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestSnapShotManager_startFastSync_Ok(t *testing.T) {
	ssm := initSnapShotManager(t, false, nil)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		err := ssm.startFastSync(txn, bh)
		assert.Nil(t, err)

		return nil
	})
}

func TestSnapShotManager_startFastSync_Error1(t *testing.T) {
	ssm := initSnapShotManager(t, false, nil)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	bh.BClaims.Height = 0

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		err := ssm.startFastSync(txn, bh)
		assert.NotNil(t, err)

		return nil
	})
}

func TestSnapShotManager_startFastSync_Error2(t *testing.T) {
	ssm := initSnapShotManager(t, true, nil)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		err := ssm.startFastSync(txn, bh)
		assert.NotNil(t, err)

		return nil
	})
}

func TestSnapShotManager_Update_Error1(t *testing.T) {
	ssm := initSnapShotManager(t, true, nil)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		ok, err := ssm.Update(txn, bh)
		assert.NotNil(t, err)
		assert.False(t, ok)

		return nil
	})
}

func TestSnapShotManager_Update_Error2(t *testing.T) {
	ssm := initSnapShotManager(t, false, nil)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	ssm.snapShotHeight.Set(2)

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		nk := nodeKey{}
		nk.key = [constants.HashLen]byte{}
		ssm.hdrNodeCache.objs[nk] = nil

		nk1 := nodeKey{}
		nk1.key = [constants.HashLen]byte{123}
		ssm.hdrNodeCache.objs[nk1] = &nodeResponse{
			root:  make([]byte, 32),
			batch: nil,
			layer: 0,
		}

		ok, err := ssm.Update(txn, bh)
		assert.NotNil(t, err)
		assert.False(t, ok)

		return nil
	})
}

func TestSnapShotManager_finalizeSync_Error1(t *testing.T) {
	ssm := initSnapShotManager(t, false, nil)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		err := ssm.startFastSync(txn, bh)
		assert.Nil(t, err)

		err = ssm.finalizeSync(txn, bh)
		assert.NotNil(t, err)

		return nil
	})
}

func TestSnapShotManager_updateDls_Ok(t *testing.T) {
	ssm := initSnapShotManager(t, false, nil)

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		err := ssm.updateDls(txn, 1, 2, 0)
		assert.Nil(t, err)

		return nil
	})
}

func TestSnapShotManager_findTailSyncHeight_Error1(t *testing.T) {
	ssm := initSnapShotManager(t, false, nil)

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		bh, err := ssm.findTailSyncHeight(txn, 1024, 1023)
		assert.NotNil(t, err)
		assert.Nil(t, bh)

		return nil
	})
}

func TestSnapShotManager_findTailSyncHeight_Ok(t *testing.T) {
	ssm := initSnapShotManager(t, false, nil)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(1024)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		err := ssm.database.SetSnapshotBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		bh2, err := ssm.findTailSyncHeight(txn, 1024, 1023)
		assert.Nil(t, err)
		assert.NotNil(t, bh2)
		assert.Equal(t, bh.BClaims.Height, bh2.BClaims.Height)

		return nil
	})
}

func TestSnapShotManager_syncTailingBlockHeaders_Ok1(t *testing.T) {
	ssm := initSnapShotManager(t, false, nil)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(3)
	round := uint32(3)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	ssm.tailSyncHeight = 5

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		bh.BClaims.Height = 1
		err := ssm.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		bh.BClaims.Height++
		err = ssm.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		bh.BClaims.Height++
		err = ssm.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = ssm.syncTailingBlockHeaders(txn, 3)
		assert.Nil(t, err)

		return nil
	})
}

func TestSnapShotManager_syncTailingBlockHeaders_Ok2(t *testing.T) {
	ssm := initSnapShotManager(t, false, nil)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	h := 1050
	height := uint32(h)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	ssm.tailSyncHeight = height

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		for i := 1; i <= 2050; i++ {
			bh.BClaims.Height = uint32(i)
			err := ssm.database.SetCommittedBlockHeader(txn, bh)
			if err != nil {
				t.Fatalf("Shouldn't have raised error: %v", err)
			}
		}

		for i := 1; i < h; i++ {
			err := ssm.database.DeleteCommittedBlockHeader(txn, uint32(i))
			if err != nil {
				t.Fatalf("Shouldn't have raised error: %v", err)
			}
		}

		err := ssm.syncTailingBlockHeaders(txn, 1050)
		assert.Nil(t, err)

		return nil
	})
}

func TestSnapShotManager_syncTailingBlockHeaders_Ok3(t *testing.T) {
	ssm := initSnapShotManager(t, false, nil)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	h := 1050
	height := uint32(h)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	ssm.tailSyncHeight = height

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		for i := 1; i <= 2050; i++ {
			bh.BClaims.Height = uint32(i)
			err := ssm.database.SetCommittedBlockHeader(txn, bh)
			if err != nil {
				t.Fatalf("Shouldn't have raised error: %v", err)
			}
		}

		for i := 1; i < h; i++ {
			err := ssm.database.DeleteCommittedBlockHeader(txn, uint32(i))
			if err != nil {
				t.Fatalf("Shouldn't have raised error: %v", err)
			}
		}

		key := ssm.database.MakeHeaderTrieKeyFromHeight(uint32(h - 1))
		err := ssm.database.SetPendingHdrLeafKey(txn, key, make([]byte, constants.HashLen))
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = ssm.syncTailingBlockHeaders(txn, 1050)
		assert.Nil(t, err)

		return nil
	})
}

func TestSnapShotManager_syncTailingBlockHeaders_Error1(t *testing.T) {
	ssm := initSnapShotManager(t, false, nil)

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		err := ssm.syncTailingBlockHeaders(txn, 1)
		assert.NotNil(t, err)

		return nil
	})
}

func TestSnapShotManager_syncStateNodes_Ok1(t *testing.T) {
	ssm := initSnapShotManager(t, true, nil)

	nr := &nodeResponse{
		root:  make([]byte, 32),
		batch: make([]byte, 32),
		layer: 0,
	}

	err := ssm.stateNodeCache.insert(3, nr)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		err = ssm.syncStateNodes(txn, 3)
		assert.Nil(t, err)

		return nil
	})
}

func TestSnapShotManager_syncStateNodes_Error2(t *testing.T) {
	leafs := []trie.LeafNode{trie.LeafNode{Key: make([]byte, constants.HashLen), Value: make([]byte, constants.HashLen)}}
	ssm := initSnapShotManager(t, false, leafs)

	nr := &nodeResponse{
		root:  make([]byte, 32),
		batch: make([]byte, 32),
		layer: 0,
	}

	err := ssm.stateNodeCache.insert(3, nr)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		err = ssm.syncStateNodes(txn, 3)
		assert.NotNil(t, err)

		return nil
	})
}

func TestSnapShotManager_syncStateNodes_Error3(t *testing.T) {
	leafs := []trie.LeafNode{trie.LeafNode{Key: []byte{123}, Value: make([]byte, constants.HashLen)}}
	ssm := initSnapShotManager(t, false, leafs)

	nr := &nodeResponse{
		root:  make([]byte, 32),
		batch: make([]byte, 32),
		layer: 0,
	}

	err := ssm.stateNodeCache.insert(3, nr)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		err = ssm.syncStateNodes(txn, 3)
		assert.NotNil(t, err)

		return nil
	})
}

func TestSnapShotManager_syncStateNodes_Error4(t *testing.T) {
	leafs := []trie.LeafNode{trie.LeafNode{Key: []byte{31}, Value: make([]byte, constants.HashLen)}}
	ssm := initSnapShotManager(t, false, leafs)

	nr := &nodeResponse{
		root:  make([]byte, 32),
		batch: make([]byte, 32),
		layer: 0,
	}

	err := ssm.stateNodeCache.insert(3, nr)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		err = ssm.syncStateNodes(txn, 3)
		assert.NotNil(t, err)

		return nil
	})
}

func TestSnapShotManager_syncStateLeaves_Ok1(t *testing.T) {
	ssm := initSnapShotManager(t, true, nil)

	nr := &stateResponse{
		snapShotHeight: 3,
		key:            make([]byte, 32),
		value:          make([]byte, 32),
		data:           make([]byte, 32),
	}

	err := ssm.stateLeafCache.insert(3, nr)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		err := ssm.syncStateLeaves(txn, 3)
		assert.Nil(t, err)

		return nil
	})
}

func TestSnapShotManager_syncHdrLeaves_Ok1(t *testing.T) {
	ssm := initSnapShotManager(t, true, nil)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(3)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	bhBytes, err := bh.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}

	nr := &stateResponse{
		snapShotHeight: 3,
		key:            make([]byte, 32),
		value:          make([]byte, 32),
		data:           bhBytes,
	}

	err = ssm.hdrLeafCache.insert(nr)
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}

	_ = ssm.database.Update(func(txn *badger.Txn) error {
		err = ssm.syncHdrLeaves(txn, 3)
		assert.Nil(t, err)

		return nil
	})
}

func TestSnapShotManager_downloadWithRetryStateNodeClosure_Ok1(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Shouldn't have raised error: %v", r)
		}
	}()

	ssm := initSnapShotManager(t, true, nil)

	p2pClientMock := &request.P2PClientMock{}
	p2pClientMock.On("GetSnapShotNode", mock.Anything, mock.Anything, mock.Anything).Return(&proto.GetSnapShotNodeResponse{}, nil)
	client := &request.Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	ssm.requestBus = client

	dl := &dlReq{
		snapShotHeight: uint32(1),
		key:            make([]byte, 32),
		value:          make([]byte, 32),
		layer:          0,
	}

	fnc := ssm.downloadWithRetryStateNodeClosure(dl)
	fnc()
}

func TestSnapShotManager_downloadWithRetryStateNodeClosure_Ok2(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Shouldn't have raised error: %v", r)
		}
	}()

	ssm := initSnapShotManager(t, true, nil)

	p2pClientMock := &request.P2PClientMock{}
	p2pClientMock.On("GetSnapShotNode", mock.Anything, mock.Anything, mock.Anything).Return(&proto.GetSnapShotNodeResponse{Node: make([]byte, 32)}, nil)
	client := &request.Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	ssm.requestBus = client

	dl := &dlReq{
		snapShotHeight: uint32(1),
		key:            make([]byte, 32),
		value:          make([]byte, 32),
		layer:          0,
	}

	fnc := ssm.downloadWithRetryStateNodeClosure(dl)
	fnc()
}

func TestSnapShotManager_downloadWithRetryHdrLeafClosure_Ok1(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Shouldn't have raised error: %v", r)
		}
	}()

	ssm := initSnapShotManager(t, true, nil)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, _, _, _, _, _, _, _, bh := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)
	bhBytes1, err := bh.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}

	bh.BClaims.Height++
	bhBytes2, err := bh.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}

	bh.BClaims.Height = 7
	bhBytes3, err := bh.MarshalBinary()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}

	p2pClientMock := &request.P2PClientMock{}
	p2pClientMock.On("GetBlockHeaders", mock.Anything, mock.Anything, mock.Anything).Return(&proto.GetBlockHeadersResponse{BlockHeaders: [][]byte{bhBytes1, bhBytes2, bhBytes3}}, nil)
	client := &request.Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	ssm.requestBus = client

	dl1 := &dlReq{
		snapShotHeight: uint32(2),
		key:            utils.MarshalUint32(1),
		value:          make([]byte, 32),
		layer:          0,
	}

	dl2 := &dlReq{
		snapShotHeight: uint32(2),
		key:            utils.MarshalUint32(3),
		value:          make([]byte, 32),
		layer:          0,
	}

	dl3 := &dlReq{
		snapShotHeight: uint32(3),
		key:            utils.MarshalUint32(4),
		value:          make([]byte, 32),
		layer:          0,
	}

	bhHash, err := bh.BlockHash()
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
	}

	dl4 := &dlReq{
		snapShotHeight: uint32(4),
		key:            utils.MarshalUint32(7),
		value:          bhHash,
		layer:          0,
	}

	fnc := ssm.downloadWithRetryHdrLeafClosure([]*dlReq{dl1, dl2, dl3, dl4})
	fnc()
}

func TestSnapShotManager_downloadWithRetryStateLeafClosure_Ok1(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Shouldn't have raised error: %v", r)
		}
	}()

	ssm := initSnapShotManager(t, true, nil)

	p2pClientMock := &request.P2PClientMock{}
	p2pClientMock.On("GetSnapShotStateData", mock.Anything, mock.Anything, mock.Anything).Return(&proto.GetSnapShotStateDataResponse{Data: nil}, nil)
	client := &request.Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	ssm.requestBus = client

	dl := &dlReq{
		snapShotHeight: uint32(1),
		key:            make([]byte, 32),
		value:          make([]byte, 32),
		layer:          0,
	}

	fnc := ssm.downloadWithRetryStateLeafClosure(dl)
	fnc()
}

func TestSnapShotManager_downloadWithRetryStateLeafClosure_Ok2(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Shouldn't have raised error: %v", r)
		}
	}()

	ssm := initSnapShotManager(t, true, nil)

	p2pClientMock := &request.P2PClientMock{}
	p2pClientMock.On("GetSnapShotStateData", mock.Anything, mock.Anything, mock.Anything).Return(&proto.GetSnapShotStateDataResponse{Data: make([]byte, 32)}, nil)
	client := &request.Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	ssm.requestBus = client

	dl := &dlReq{
		snapShotHeight: uint32(1),
		key:            make([]byte, 32),
		value:          make([]byte, 32),
		layer:          0,
	}

	fnc := ssm.downloadWithRetryStateLeafClosure(dl)
	fnc()
}

func initSnapShotManager(t *testing.T, shouldFail bool, leafs []trie.LeafNode) *SnapShotManager {
	ctx := context.Background()

	rawSsmDb, err := utils.OpenBadger(ctx.Done(), "", true)
	if err != nil {
		t.Fatal(err)
	}
	ssmDb := &db.Database{}
	ssmDb.Init(rawSsmDb)

	storage := appObjs.MakeMockStorageGetter()

	ssm := &SnapShotManager{}
	ssm.Init(ssmDb, storage)

	app := appmock.New()
	_, vv := createProposal(t)
	app.SetNextValidValue(vv)
	app.SetShouldFail(shouldFail)
	app.SetLeafs(leafs)

	ssm.appHandler = app

	return ssm
}
