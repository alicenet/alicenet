package consensus

import (
	"fmt"
	"sync"
	"time"

	"github.com/alicenet/alicenet/application"
	"github.com/alicenet/alicenet/consensus/accusation"
	"github.com/alicenet/alicenet/consensus/admin"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/evidence"
	"github.com/alicenet/alicenet/consensus/gossip"
	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/dynamics"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/peering"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

type remoteVar struct {
	condition func() bool
}

func (s *remoteVar) isSet() bool {
	return s.condition()
}

func (s *remoteVar) isNotSet() bool {
	return !s.isSet()
}

func newRemoteVar(fn func() bool) *remoteVar {
	return &remoteVar{
		condition: fn,
	}
}

type resetVar struct {
	sync.Mutex
	condition bool
}

func (s *resetVar) isSet() bool {
	s.Lock()
	state := s.condition
	s.Unlock()
	return state
}

func (s *resetVar) isNotSet() bool {
	return !s.isSet()
}

func (s *resetVar) set(state bool) {
	s.Lock()
	s.condition = state
	s.Unlock()
}

func newResetVar() *resetVar {
	return &resetVar{}
}

type setOnceVar struct {
	condition func() bool
	isSetChan chan struct{}
	setOnce   sync.Once
}

func (s *setOnceVar) isSet() bool {
	select {
	case <-s.isSetChan:
		return true
	default:
		if s.condition() {
			s.setOnce.Do(func() {
				close(s.isSetChan)
			})
			return true
		}
		return false
	}
}

func (s *setOnceVar) isNotSet() bool {
	return !s.isSet()
}

func newSetOnceVar(condition func() bool) *setOnceVar {
	return &setOnceVar{
		setOnce:   sync.Once{},
		isSetChan: make(chan struct{}),
		condition: condition,
	}
}

type conditionFn func() bool
type singleRetFn func() error
type twoRetFn func() (bool, error)

func newLoopConfig() *loopConfig {
	return &loopConfig{}
}

type loopConfig struct {
	name          string
	lock          bool
	hasDelay      bool
	delay         time.Duration
	freq          time.Duration
	initialDelay  time.Duration
	singleRetFunc bool
	fn            singleRetFn
	twoRetFunc    bool
	fn2           twoRetFn
	lfConditions  []conditionFn
	lsConditions  []conditionFn
	varSetter     func(bool)
}

func (lc *loopConfig) withName(name string) *loopConfig {
	lc.hasDelay = true
	lc.name = name
	return lc
}

func (lc *loopConfig) withDelayOnConditionFailure(delay time.Duration) *loopConfig {
	lc.delay = delay
	return lc
}

func (lc *loopConfig) withLockFreeCondition(fn conditionFn) *loopConfig {
	if lc.lfConditions == nil {
		lc.lfConditions = []conditionFn{}
	}
	lc.lfConditions = append(lc.lfConditions, fn)
	return lc
}

func (lc *loopConfig) withLockedCondition(fn conditionFn) *loopConfig {
	if !lc.lock {
		panic("lock not set")
	}
	if lc.lsConditions == nil {
		lc.lsConditions = []conditionFn{}
	}
	lc.lsConditions = append(lc.lsConditions, fn)
	return lc
}

func (lc *loopConfig) withFreq(freq time.Duration) *loopConfig {
	lc.freq = freq
	return lc
}

func (lc *loopConfig) withFn2(fn twoRetFn, varSetter func(bool)) *loopConfig {
	if lc.singleRetFunc {
		panic("conflicting function types")
	}
	lc.twoRetFunc = true
	lc.varSetter = varSetter
	lc.fn2 = fn
	return lc
}

func (lc *loopConfig) withFn(fn singleRetFn) *loopConfig {
	if lc.twoRetFunc {
		panic("conflicting function types")
	}
	lc.singleRetFunc = true
	lc.fn = fn
	return lc
}

func (lc *loopConfig) withLock() *loopConfig {
	lc.lock = true
	return lc
}

func (lc *loopConfig) withInitialDelay(idt time.Duration) *loopConfig {
	lc.initialDelay = idt
	return lc
}

// Synchronizer controls logic hand off between services
// This system coordinates what services may run under what conditions
// The system operates as a scheduler as well as a reactor to external
// events
type Synchronizer struct {
	sync.Mutex
	wg        sync.WaitGroup
	startOnce sync.Once
	logger    *logrus.Logger
	closeChan chan struct{}
	closeOnce sync.Once

	cdb             *db.Database
	mdb             *badger.DB
	tdb             *badger.DB
	gossipClient    *gossip.Client
	gossipHandler   *gossip.Handlers
	evidenceHandler *evidence.Pool
	stateHandler    *lstate.Engine
	appHandler      *application.Application
	adminHandler    *admin.Handlers
	peerMan         *peering.PeerManager
	accusationMan   *accusation.Manager

	initialized   *setOnceVar
	peerMinThresh *remoteVar
	ethSyncDone   *remoteVar
	madSyncDone   *resetVar

	storage dynamics.StorageGetter
}

// Init initializes the struct
func (s *Synchronizer) Init(cdb *db.Database, mdb *badger.DB, tdb *badger.DB, gc *gossip.Client, gh *gossip.Handlers, ep *evidence.Pool, eng *lstate.Engine, app *application.Application, ah *admin.Handlers, pman *peering.PeerManager, aman *accusation.Manager, storage dynamics.StorageGetter) {
	s.logger = logging.GetLogger(constants.LoggerConsensus)
	s.cdb = cdb
	s.mdb = mdb
	s.tdb = tdb
	s.gossipClient = gc
	s.gossipHandler = gh
	s.evidenceHandler = ep
	s.stateHandler = eng
	s.appHandler = app
	s.adminHandler = ah
	s.peerMan = pman
	s.accusationMan = aman
	s.wg = sync.WaitGroup{}
	s.closeChan = make(chan struct{})
	s.closeOnce = sync.Once{}
	s.startOnce = sync.Once{}
	s.initialized = newSetOnceVar(s.adminHandler.IsInitialized)
	s.ethSyncDone = newRemoteVar(s.adminHandler.IsSynchronized)
	s.peerMinThresh = newRemoteVar(s.peerMan.PeeringComplete)
	s.madSyncDone = newResetVar()
	s.storage = storage
}

func (s *Synchronizer) CloseChan() <-chan struct{} {
	return s.closeChan
}

// Start will start the Synchronizer
func (s *Synchronizer) Start() {
	s.startOnce.Do(func() {
		s.logger.Debugf("Started Syncronizer")
		s.wg.Add(1)
		go s.adminInteruptLoop()
		s.wg.Add(1)
		go s.gossipInteruptLoop()
		s.setupLoops()
		go s.adminHandler.InitializationMonitor(s.closeChan)
	})
}

// Stop terminates the Synchronizer and all managed services
func (s *Synchronizer) Stop() {
	s.closeOnce.Do(func() {
		s.logger.Warning("Graceful stop of Synchronizer started")
		close(s.closeChan)
	})
	s.wg.Wait()
	s.logger.Warning("Synchronizer loops stopped.")
}

func (s *Synchronizer) onError(location string, err error) {
	s.closeOnce.Do(func() {
		s.logger.Errorf("Triggering shutdown of Synchronizer due to error at location %s: %v", location, err)
		close(s.closeChan)
	})
	if err != errorz.ErrClosing {
		s.logger.Error("Follow up error in Synchronizer:", err)
	}
}

func (s *Synchronizer) isClosing() bool {
	select {
	case <-s.closeChan:
		return true
	default:
		return false
	}
}

func (s *Synchronizer) isNotClosing() bool {
	return !s.isClosing()
}

func (s *Synchronizer) loop(lc *loopConfig) {
	defer s.wg.Done()
	defer func() { s.logger.Warnf("Stopping %s loop", lc.name) }()
	s.logger.Warnf("Starting %s loop", lc.name)
	var initialDelayDone bool
	if lc.initialDelay == 0 {
		initialDelayDone = true
	} else {
		initialDelayDone = false
	}
	for {
		if s.isClosing() {
			return
		}
		select {
		case <-s.closeChan:
			return
		case <-time.After(lc.freq):
			func() {
				for i := 0; i < len(lc.lfConditions); i++ {
					if !lc.lfConditions[i]() {
						if lc.hasDelay {
							select {
							case <-s.closeChan:
								return
							case <-time.After(lc.delay):
								return
							}
						}
						return
					}
				}
				if !initialDelayDone {
					select {
					case <-time.After(lc.initialDelay):
						initialDelayDone = true
						return
					case <-s.closeChan:
						initialDelayDone = true
						return
					}
				}
				if lc.lock {
					s.Lock()
					defer s.Unlock()
				}
				for i := 0; i < len(lc.lsConditions); i++ {
					if !lc.lsConditions[i]() {
						if lc.hasDelay {
							select {
							case <-s.closeChan:
								return
							case <-time.After(lc.delay):
								return
							}
						}
						return
					}
				}
				if lc.singleRetFunc {
					if err := lc.fn(); err != nil {
						s.onError(fmt.Sprintf("Error in synchronizer at %s:", lc.name), err)
					}
					return
				}
				if lc.twoRetFunc {
					ok, err := lc.fn2()
					if err != nil {
						s.onError(fmt.Sprintf("Error in synchronizer at %s:", lc.name), err)
					}
					lc.varSetter(ok)
					return
				}
			}()
		}
	}
}

func (s *Synchronizer) Safe() bool {
	if s.initialized.isNotSet() {
		return false
	}
	if s.ethSyncDone.isNotSet() {
		return false
	}
	if s.madSyncDone.isNotSet() {
		return false
	}
	if s.peerMinThresh.isNotSet() {
		return false
	}
	if s.isClosing() {
		return false
	}
	return true
}

func (s *Synchronizer) gossipInteruptLoop() {
	defer s.wg.Done()
	defer func() { s.logger.Warn("Stopping Gossip loop") }()
	s.logger.Warn("Starting Gossip loop")
	for {
		select {
		case s.gossipHandler.ReceiveLock <- s:
			continue
		case <-s.closeChan:
			return
		}
	}
}

func (s *Synchronizer) adminInteruptLoop() {
	defer s.wg.Done()
	defer func() { s.logger.Warn("Stopping AdminInterupt loop") }()
	s.logger.Warn("Starting AdminInterupt loop")
	for {
		select {
		case s.adminHandler.ReceiveLock <- s:
			continue
		case <-s.closeChan:
			return
		}
	}
}

func (s *Synchronizer) setupLoops() {
	stateLoopInSyncConfig := newLoopConfig().
		withName("StateLoop-InSync").
		withInitialDelay(9*constants.MsgTimeout).
		withFn2(s.stateHandler.UpdateLocalState, s.madSyncDone.set).
		withFreq(200 * time.Millisecond).
		withDelayOnConditionFailure(200 * time.Millisecond).
		withLockFreeCondition(s.isNotClosing).
		withLockFreeCondition(s.initialized.isSet).
		withLockFreeCondition(s.ethSyncDone.isSet).
		withLockFreeCondition(s.madSyncDone.isSet).
		withLockFreeCondition(s.peerMinThresh.isSet).
		withLock().
		withLockedCondition(s.isNotClosing).
		withLockedCondition(s.madSyncDone.isSet)
	s.wg.Add(1)
	go s.loop(stateLoopInSyncConfig)

	stateLoopNoSyncConfig := newLoopConfig().
		withName("StateLoop-NoSync").
		withFn2(s.stateHandler.Sync, s.madSyncDone.set).
		withFreq(100 * time.Millisecond).
		withDelayOnConditionFailure(1 * time.Second).
		withLockFreeCondition(s.isNotClosing).
		withLockFreeCondition(s.initialized.isSet).
		withLockFreeCondition(s.ethSyncDone.isSet).
		withLockFreeCondition(s.madSyncDone.isNotSet).
		withLockFreeCondition(s.peerMinThresh.isSet).
		withLock().
		withLockedCondition(s.isNotClosing).
		withLockedCondition(s.madSyncDone.isNotSet)
	s.wg.Add(1)
	go s.loop(stateLoopNoSyncConfig)

	appLoopConfig := newLoopConfig().
		withName("AppLoop").
		withFn(s.appHandler.Cleanup).
		withFreq(181 * time.Second).
		withDelayOnConditionFailure(91 * time.Second).
		withLockFreeCondition(s.isNotClosing).
		withLockFreeCondition(s.initialized.isSet).
		withLockFreeCondition(s.ethSyncDone.isSet).
		withLockFreeCondition(s.peerMinThresh.isSet).
		withLock().
		withLockedCondition(s.isNotClosing)
	s.wg.Add(1)
	go s.loop(appLoopConfig)

	reGossipLoopConfig := newLoopConfig().
		withName("ReGossipLoop").
		withInitialDelay(9 * constants.MsgTimeout).
		withFn(s.gossipClient.ReGossip).
		withFreq(9 * s.storage.GetMsgTimeout()).
		withDelayOnConditionFailure(s.storage.GetMsgTimeout()).
		withLockFreeCondition(s.isNotClosing).
		withLockFreeCondition(s.initialized.isSet).
		withLockFreeCondition(s.ethSyncDone.isSet).
		withLockFreeCondition(s.madSyncDone.isSet).
		withLockFreeCondition(s.peerMinThresh.isSet)
	s.wg.Add(1)
	go s.loop(reGossipLoopConfig)

	evidenceLoopConfig := newLoopConfig().
		withName("EvidenceLoop").
		withFn(s.evidenceHandler.Cleanup).
		withFreq(179 * time.Second).
		withDelayOnConditionFailure(121 * time.Second).
		withLockFreeCondition(s.isNotClosing).
		withLockFreeCondition(s.initialized.isSet).
		withLockFreeCondition(s.ethSyncDone.isSet).
		withLock()
	s.wg.Add(1)
	go s.loop(evidenceLoopConfig)

	cdbgcLoopConfig := newLoopConfig().
		withName("CDB-GCLoop").
		withFn(s.cdb.GarbageCollect).
		withFreq(30 * time.Second).
		withDelayOnConditionFailure(17 * time.Second).
		withLockFreeCondition(s.isNotClosing).
		withLockFreeCondition(s.initialized.isSet).
		withLock().
		withLockedCondition(s.isNotClosing)
	s.wg.Add(1)
	go s.loop(cdbgcLoopConfig)

	if s.tdb != nil { // prevent cleanup on in memory by setting to nil
		tdbgcLoopConfig := newLoopConfig().
			withName("TDB-GCLoop").
			withFn(
				func() error {
					s.tdb.RunValueLogGC(constants.BadgerDiscardRatio)
					s.tdb.RunValueLogGC(constants.BadgerDiscardRatio)
					return nil
				}).
			withFreq(600 * time.Second).
			withDelayOnConditionFailure(600 * time.Second).
			withLockFreeCondition(s.isNotClosing).
			withLock().
			withLockedCondition(s.isNotClosing)
		s.wg.Add(1)
		go s.loop(tdbgcLoopConfig)
	}

	accusationManagerLoopConfig := newLoopConfig().
		withName("AccusationManagerLoop").
		//withInitialDelay(9*constants.MsgTimeout).
		withFn(s.accusationMan.Poll).
		withFreq(100 * time.Millisecond).
		withDelayOnConditionFailure(100 * time.Millisecond).
		withLockFreeCondition(s.isNotClosing).
		withLockFreeCondition(s.initialized.isSet).
		withLockFreeCondition(s.ethSyncDone.isSet).
		withLockFreeCondition(s.madSyncDone.isSet).
		withLockFreeCondition(s.peerMinThresh.isSet).
		withLock().
		withLockedCondition(s.isNotClosing).
		withLockedCondition(s.madSyncDone.isSet)
	s.wg.Add(1)
	go s.loop(accusationManagerLoopConfig)
}
