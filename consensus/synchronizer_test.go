package consensus

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/config"
	"github.com/MadBase/MadNet/consensus/admin"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/gossip"
	"github.com/MadBase/MadNet/constants"
	mncrypto "github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/utils"
	"github.com/stretchr/testify/assert"
)

const (
	timeToStop              = 3 * time.Second
	timeToFail              = timeToStop + 1*time.Second
	initialDelay            = 500 * time.Millisecond
	freq                    = 200 * time.Millisecond
	delayOnConditionFailure = 1 * time.Second
)

func TestSynchronizer_InitAndStart(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Shouldn't panic")
		}
	}()

	rawDb, err := utils.OpenBadger(context.Background().Done(), "", true)
	assert.Nil(t, err)

	database := &db.Database{}
	database.Init(rawDb)

	tdb, err := utils.OpenBadger(context.Background().Done(), "", true)
	assert.Nil(t, err)

	consAdminHandlers := &admin.Handlers{}
	consAdminHandlers.Init(1, database, mncrypto.Hasher([]byte(config.Configuration.Validator.SymmetricKey)), nil, make([]byte, constants.HashLen), objs.MakeMockStorageGetter())

	sync := &Synchronizer{}
	sync.Init(nil, nil, tdb, nil, &gossip.Handlers{}, nil, nil, nil, consAdminHandlers, nil, objs.MakeMockStorageGetter())
	go stopSync(sync)
	sync.Start()
	select {
	case <-sync.CloseChan():
	case <-time.After(timeToFail):
		t.Errorf("Shouldn't reach this line")
	}
}

func TestSynchronizer_loopWithFn(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Shouldn't panic")
		}
	}()

	rawDb, err := utils.OpenBadger(context.Background().Done(), "", true)
	assert.Nil(t, err)

	database := &db.Database{}
	database.Init(rawDb)

	consAdminHandlers := &admin.Handlers{}
	consAdminHandlers.Init(1, database, mncrypto.Hasher([]byte(config.Configuration.Validator.SymmetricKey)), nil, make([]byte, constants.HashLen), objs.MakeMockStorageGetter())

	sync := &Synchronizer{}
	sync.Init(nil, nil, nil, nil, &gossip.Handlers{}, nil, nil, nil, consAdminHandlers, nil, objs.MakeMockStorageGetter())

	loopFnOk := newLoopConfig().
		withName("loopFnOk").
		withInitialDelay(initialDelay).
		withFn(func() error { return nil }).
		withFreq(freq).
		withDelayOnConditionFailure(delayOnConditionFailure)
	sync.wg.Add(1)
	go sync.loop(loopFnOk)

	loopFnErr := newLoopConfig().
		withName("loopFnErr").
		withInitialDelay(initialDelay).
		withFn(func() error { return errors.New("fn error") }).
		withFreq(freq).
		withDelayOnConditionFailure(delayOnConditionFailure)
	sync.wg.Add(1)
	go sync.loop(loopFnErr)

	go stopSync(sync)

	select {
	case <-sync.CloseChan():
	case <-time.After(timeToFail):
		t.Errorf("Shouldn't reach this line")
	}
}

func TestSynchronizer_loopWithFn2(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Shouldn't panic")
		}
	}()

	rawDb, err := utils.OpenBadger(context.Background().Done(), "", true)
	assert.Nil(t, err)

	database := &db.Database{}
	database.Init(rawDb)

	consAdminHandlers := &admin.Handlers{}
	consAdminHandlers.Init(1, database, mncrypto.Hasher([]byte(config.Configuration.Validator.SymmetricKey)), nil, make([]byte, constants.HashLen), objs.MakeMockStorageGetter())

	sync := &Synchronizer{}
	sync.Init(nil, nil, nil, nil, &gossip.Handlers{}, nil, nil, nil, consAdminHandlers, nil, objs.MakeMockStorageGetter())

	loopFnOk := newLoopConfig().
		withName("loopFnOk").
		withInitialDelay(initialDelay).
		withFn2(func() (bool, error) { return true, nil }, func(bool) {}).
		withFreq(freq).
		withDelayOnConditionFailure(delayOnConditionFailure)
	sync.wg.Add(1)
	go sync.loop(loopFnOk)

	loopFnErr := newLoopConfig().
		withName("loopFnErr").
		withInitialDelay(initialDelay).
		withFn2(func() (bool, error) { return false, errors.New("fn2 error") }, func(bool) {}).
		withFreq(freq).
		withDelayOnConditionFailure(delayOnConditionFailure)
	sync.wg.Add(1)
	go sync.loop(loopFnErr)

	go stopSync(sync)

	select {
	case <-sync.CloseChan():
	case <-time.After(timeToFail):
		t.Errorf("Shouldn't reach this line")
	}
}

func TestSynchronizer_loopWithLockedCondition(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Shouldn't panic")
		}
	}()

	rawDb, err := utils.OpenBadger(context.Background().Done(), "", true)
	assert.Nil(t, err)

	database := &db.Database{}
	database.Init(rawDb)

	consAdminHandlers := &admin.Handlers{}
	consAdminHandlers.Init(1, database, mncrypto.Hasher([]byte(config.Configuration.Validator.SymmetricKey)), nil, make([]byte, constants.HashLen), objs.MakeMockStorageGetter())

	sync := &Synchronizer{}
	sync.Init(nil, nil, nil, nil, &gossip.Handlers{}, nil, nil, nil, consAdminHandlers, nil, objs.MakeMockStorageGetter())

	loopLC := newLoopConfig().
		withName("loopLC").
		withInitialDelay(initialDelay).
		withFreq(freq).
		withDelayOnConditionFailure(delayOnConditionFailure).
		withLock().
		withLockedCondition(sync.isClosing)
	sync.wg.Add(1)
	go sync.loop(loopLC)

	go stopSync(sync)

	sync.initialized = newSetOnceVar(func() bool { return true })
	sync.ethSyncDone = newRemoteVar(func() bool { return true })
	sync.peerMinThresh = newRemoteVar(func() bool { return true })
	sync.madSyncDone = &resetVar{condition: true}
	assert.True(t, sync.Safe())

	select {
	case <-sync.CloseChan():
	case <-time.After(timeToFail):
		t.Errorf("Shouldn't reach this line")
	}
}

func TestSynchronizer_SafeOk(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Shouldn't panic")
		}
	}()

	rawDb, err := utils.OpenBadger(context.Background().Done(), "", true)
	assert.Nil(t, err)

	database := &db.Database{}
	database.Init(rawDb)

	consAdminHandlers := &admin.Handlers{}
	consAdminHandlers.Init(1, database, mncrypto.Hasher([]byte(config.Configuration.Validator.SymmetricKey)), nil, make([]byte, constants.HashLen), objs.MakeMockStorageGetter())

	sync := &Synchronizer{}
	sync.Init(nil, nil, nil, nil, &gossip.Handlers{}, nil, nil, nil, consAdminHandlers, nil, objs.MakeMockStorageGetter())

	go stopSync(sync)
	assert.False(t, sync.Safe())

	sync.initialized = newSetOnceVar(func() bool { return true })
	assert.False(t, sync.Safe())

	sync.ethSyncDone = newRemoteVar(func() bool { return true })
	assert.False(t, sync.Safe())

	madSyncDone := newResetVar()
	madSyncDone.set(true)
	sync.madSyncDone = madSyncDone
	sync.peerMinThresh = newRemoteVar(func() bool { return false })
	assert.False(t, sync.Safe())

	sync.peerMinThresh = newRemoteVar(func() bool { return true })
	assert.True(t, sync.Safe())

	select {
	case <-sync.CloseChan():
	case <-time.After(timeToFail):
		t.Errorf("Shouldn't reach this line")
	}
}

func stopSync(sync *Synchronizer) {
	<-time.After(timeToStop)
	sync.Stop()
}
