package lstate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/consensus/request"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/dynamics"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/middleware"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const chanBuffering int = int(constants.EpochLength)
const maxNumber int = chanBuffering
const minWorkers = 4
const maxRetryCount = 6
const backOffAmount = 1
const backOffJitter = float64(.1)

type dlReq struct {
	snapShotHeight uint32
	key            []byte
	value          []byte
	layer          int
}

type nodeKey struct {
	key [constants.HashLen]byte
}

func newNodeKey(s []byte) (nodeKey, error) {
	if len(s) != constants.HashLen {
		return nodeKey{}, errorz.ErrInvalid{}.New("Error in newNodeKey: byte slice not constants.HashLen bytes")
	}
	nk := nodeKey{}
	nk.key = [constants.HashLen]byte{}
	copy(nk.key[:], s)
	return nk, nil
}

type nodeResponse struct {
	snapShotHeight uint32
	root           []byte
	layer          int
	batch          []byte
}

type stateResponse struct {
	snapShotHeight uint32
	key            []byte
	value          []byte
	data           []byte
}

type nodeCache struct {
	sync.RWMutex
	objs      map[uint32]map[nodeKey]*nodeResponse
	minHeight uint32
}

func (nc *nodeCache) Init() {
	nc.objs = make(map[uint32]map[nodeKey]*nodeResponse)
}

func (nc *nodeCache) getNodeKeys(height uint32, maxNumber int) []nodeKey {
	nc.RLock()
	defer nc.RUnlock()
	nodeKeys := []nodeKey{}
	for k := range nc.objs[height] {
		nodeKeys = append(nodeKeys, k)
		if len(nodeKeys) >= maxNumber {
			break
		}
	}
	return nodeKeys
}

func (nc *nodeCache) insert(height uint32, nr *nodeResponse) error {
	nc.Lock()
	defer nc.Unlock()
	if height < nc.minHeight {
		return nil
	}
	nk, err := newNodeKey(nr.root)
	if err != nil {
		return err
	}
	if nc.objs[height] == nil {
		nc.objs[height] = make(map[nodeKey]*nodeResponse)
	}
	nc.objs[height][nk] = nr
	return nil
}

func (nc *nodeCache) contains(height uint32, nk nodeKey) bool {
	nc.RLock()
	defer nc.RUnlock()
	if nc.objs[height] == nil {
		return false
	}
	if nc.objs[height][nk] == nil {
		return false
	}
	return true
}

func (nc *nodeCache) dropBefore(height uint32) {
	nc.Lock()
	defer nc.Unlock()
	nc.minHeight = height
	allHeights := []uint32{}
	for h := range nc.objs {
		if h < height {
			allHeights = append(allHeights, h)
		}
	}
	for _, h := range allHeights {
		delete(nc.objs, h)
	}
}

func (nc *nodeCache) pop(height uint32, hash []byte) (*nodeResponse, error) {
	nc.Lock()
	defer nc.Unlock()
	if nc.objs[height] == nil {
		// Not present, so ignore
		return nil, errorz.ErrInvalid{}.New("Error in nodeCache.pop: missing height in pop request")
	}
	nk, err := newNodeKey(hash)
	if err != nil {
		return nil, err
	}
	if nc.objs[height][nk] == nil {
		return nil, errorz.ErrInvalid{}.New("Error in nodeCache.pop: missing key in pop request")
	}
	result := nc.objs[height][nk]
	delete(nc.objs[height], nk)
	return result, nil
}

type stateCache struct {
	sync.RWMutex
	objs      map[uint32]map[nodeKey]*stateResponse
	minHeight uint32
}

func (sc *stateCache) Init() {
	sc.objs = make(map[uint32]map[nodeKey]*stateResponse)
}

func (sc *stateCache) getLeafKeys(height uint32, maxNumber int) []nodeKey {
	sc.RLock()
	defer sc.RUnlock()
	nodeKeys := []nodeKey{}
	for k := range sc.objs[height] {
		nodeKeys = append(nodeKeys, k)
		if len(nodeKeys) >= maxNumber {
			break
		}
	}
	return nodeKeys
}

func (sc *stateCache) contains(height uint32, nk nodeKey) bool {
	sc.RLock()
	defer sc.RUnlock()
	if sc.objs[height] == nil {
		return false
	}
	if sc.objs[height][nk] == nil {
		return false
	}
	return true
}

func (sc *stateCache) insert(height uint32, sr *stateResponse) error {
	sc.Lock()
	defer sc.Unlock()
	if height < sc.minHeight {
		return nil
	}
	nk, err := newNodeKey(sr.key)
	if err != nil {
		return err
	}
	if sc.objs[height] == nil {
		sc.objs[height] = make(map[nodeKey]*stateResponse)
	}
	sc.objs[height][nk] = sr
	return nil
}

func (sc *stateCache) dropBefore(height uint32) {
	sc.Lock()
	defer sc.Unlock()
	allHeights := []uint32{}
	sc.minHeight = height
	for h := range sc.objs {
		if h < height {
			allHeights = append(allHeights, h)
		}
	}
	for _, h := range allHeights {
		delete(sc.objs, h)
	}
}

func (sc *stateCache) pop(height uint32, key []byte) (*stateResponse, error) {
	sc.Lock()
	defer sc.Unlock()
	if sc.objs[height] == nil {
		// Not present, so ignore
		return nil, errorz.ErrInvalid{}.New("Error in stateCache.pop: missing height in pop request")
	}
	nk, err := newNodeKey(key)
	if err != nil {
		return nil, err
	}
	if sc.objs[height][nk] == nil {
		return nil, errorz.ErrInvalid{}.New("Error in stateCache.pop: missing key in pop request")
	}
	result := sc.objs[height][nk]
	delete(sc.objs[height], nk)
	return result, nil
}

type bhCache struct {
	sync.RWMutex
	objs map[nodeKey]*stateResponse
}

func (bhc *bhCache) Init() {
	bhc.objs = make(map[nodeKey]*stateResponse)
}

func (bhc *bhCache) getLeafKeys(maxNumber int) []nodeKey {
	bhc.RLock()
	defer bhc.RUnlock()
	nodeKeys := []nodeKey{}
	for k := range bhc.objs {
		nodeKeys = append(nodeKeys, k)
		if len(nodeKeys) >= maxNumber {
			break
		}
	}
	return nodeKeys
}

func (bhc *bhCache) contains(nk nodeKey) bool {
	bhc.RLock()
	defer bhc.RUnlock()
	return bhc.objs[nk] != nil
}

func (bhc *bhCache) insert(sr *stateResponse) error {
	bhc.Lock()
	defer bhc.Unlock()
	nk, err := newNodeKey(sr.key)
	if err != nil {
		return err
	}
	bhc.objs[nk] = sr
	return nil
}

func (bhc *bhCache) pop(key []byte) (*stateResponse, error) {
	bhc.Lock()
	defer bhc.Unlock()
	nk, err := newNodeKey(key)
	if err != nil {
		return nil, err
	}
	if bhc.objs[nk] == nil {
		return nil, errorz.ErrInvalid{}.New("Error in bhCache.pop: missing key in pop request")
	}
	result := bhc.objs[nk]
	delete(bhc.objs, nk)
	return result, nil
}

type bhNodeCache struct {
	sync.RWMutex
	objs map[nodeKey]*nodeResponse
}

func (bhnc *bhNodeCache) Init() {
	bhnc.objs = make(map[nodeKey]*nodeResponse)
}

func (bhnc *bhNodeCache) getNodeKeys(maxNumber int) []nodeKey {
	bhnc.RLock()
	defer bhnc.RUnlock()
	nodeKeys := []nodeKey{}
	for k := range bhnc.objs {
		nodeKeys = append(nodeKeys, k)
		if len(nodeKeys) >= maxNumber {
			break
		}
	}
	return nodeKeys
}

func (bhnc *bhNodeCache) contains(nk nodeKey) bool {
	bhnc.RLock()
	defer bhnc.RUnlock()
	return bhnc.objs[nk] != nil
}

func (bhnc *bhNodeCache) insert(sr *nodeResponse) error {
	bhnc.Lock()
	defer bhnc.Unlock()
	nk, err := newNodeKey(sr.root)
	if err != nil {
		return err
	}
	bhnc.objs[nk] = sr
	return nil
}

func (bhnc *bhNodeCache) pop(key []byte) (*nodeResponse, error) {
	bhnc.Lock()
	defer bhnc.Unlock()
	nk, err := newNodeKey(key)
	if err != nil {
		return nil, err
	}
	if bhnc.objs[nk] == nil {
		return nil, errorz.ErrInvalid{}.New("Error in bhNodeCache.pop: missing key in pop request")
	}
	result := bhnc.objs[nk]
	delete(bhnc.objs, nk)
	return result, nil
}

type downloadTracker struct {
	sync.RWMutex
	currentDLs map[nodeKey]bool
}

func (dt *downloadTracker) Push(k nodeKey) bool {
	dt.Lock()
	defer dt.Unlock()
	if dt.currentDLs[k] {
		return false
	}
	dt.currentDLs[k] = true
	return true
}

func (dt *downloadTracker) Pop(k nodeKey) {
	dt.Lock()
	defer dt.Unlock()
	delete(dt.currentDLs, k)
}

func (dt *downloadTracker) Contains(k nodeKey) bool {
	dt.RLock()
	defer dt.RUnlock()
	return dt.currentDLs[k]
}

func (dt *downloadTracker) Size() int {
	dt.RLock()
	defer dt.RUnlock()
	return len(dt.currentDLs)
}

type atomicU32 struct {
	sync.RWMutex
	value uint32
}

func (a *atomicU32) Set(v uint32) {
	a.Lock()
	defer a.Unlock()
	a.value = v
}

func (a *atomicU32) Get() uint32 {
	a.RLock()
	defer a.RUnlock()
	return a.value
}

type workFunc func()

type SnapShotManager struct {
	sync.Mutex

	appHandler     interfaces.Application
	requestBus     *request.Client
	database       *db.Database
	logger         *logrus.Logger
	snapShotHeight *atomicU32
	storage        dynamics.StorageGetter

	hdrNodeCache   *bhNodeCache
	hdrLeafCache   *bhCache
	stateNodeCache *nodeCache
	stateLeafCache *stateCache

	hdrNodeDLs   *downloadTracker
	hdrLeafDLs   *downloadTracker
	stateNodeDLs *downloadTracker
	stateLeafDLs *downloadTracker

	hdrLeafDlChan   chan *dlReq
	hdrNodeDlChan   chan *dlReq
	stateNodeDlChan chan *dlReq
	stateLeafDlChan chan *dlReq

	workChan chan workFunc

	statusChan chan string
	closeChan  chan struct{}
	closeOnce  sync.Once

	finalizeFastSyncChan chan struct{}
	finalizeOnce         sync.Once

	tailSyncHeight uint32

	numWorkers int
}

// Init initializes the SnapShotManager
func (ssm *SnapShotManager) Init(database *db.Database, storage dynamics.StorageGetter) {
	ssm.storage = storage
	ssm.snapShotHeight = new(atomicU32)
	ssm.logger = logging.GetLogger(constants.LoggerConsensus)
	ssm.database = database
	ssm.hdrNodeCache = &bhNodeCache{}
	ssm.hdrNodeCache.Init()
	ssm.hdrLeafCache = &bhCache{}
	ssm.hdrLeafCache.Init()
	ssm.stateNodeCache = &nodeCache{}
	ssm.stateNodeCache.Init()
	ssm.stateLeafCache = &stateCache{}
	ssm.stateLeafCache.Init()
	ssm.hdrNodeDLs = &downloadTracker{
		sync.RWMutex{},
		make(map[nodeKey]bool),
	}
	ssm.hdrLeafDLs = &downloadTracker{
		sync.RWMutex{},
		make(map[nodeKey]bool),
	}
	ssm.stateLeafDLs = &downloadTracker{
		sync.RWMutex{},
		make(map[nodeKey]bool),
	}
	ssm.stateNodeDLs = &downloadTracker{
		sync.RWMutex{},
		make(map[nodeKey]bool),
	}
	ssm.statusChan = make(chan string)
	ssm.hdrLeafDlChan = make(chan *dlReq, chanBuffering)
	ssm.hdrNodeDlChan = make(chan *dlReq, chanBuffering)
	ssm.stateNodeDlChan = make(chan *dlReq, chanBuffering)
	ssm.stateLeafDlChan = make(chan *dlReq, chanBuffering)
	ssm.workChan = make(chan workFunc, chanBuffering*2)
	ssm.closeChan = make(chan struct{})
	ssm.closeOnce = sync.Once{}

	go ssm.downloadWithRetryHdrLeafWorker()
	go ssm.downloadWithRetryHdrNodeWorker()
	go ssm.downloadWithRetryStateLeafWorker()
	go ssm.downloadWithRetryStateNodeWorker()
	go ssm.loggingDelayer()
}

func (ssm *SnapShotManager) startFastSync(txn *badger.Txn, snapShotBlockHeader *objs.BlockHeader) error {
	if ssm.finalizeFastSyncChan == nil {
		ssm.finalizeOnce = sync.Once{}
		ssm.finalizeFastSyncChan = make(chan struct{})
		ssm.Lock()
		for i := 0; i < minWorkers; i++ {
			ssm.numWorkers++
			go ssm.worker(ssm.finalizeFastSyncChan)
		}
		ssm.Unlock()
	}
	ssm.tailSyncHeight = snapShotBlockHeader.BClaims.Height
	if err := ssm.database.SetCommittedBlockHeaderFastSync(txn, snapShotBlockHeader); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}

	ssm.snapShotHeight.Set(snapShotBlockHeader.BClaims.Height)

	// call dropBefore on the caches
	ssm.stateNodeCache.dropBefore(snapShotBlockHeader.BClaims.Height)
	ssm.stateLeafCache.dropBefore(snapShotBlockHeader.BClaims.Height)

	// cleanup the db of any previous state
	if err := ssm.cleanupDatabase(txn); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}

	// insert the request for the root node into the database
	if !bytes.Equal(utils.CopySlice(snapShotBlockHeader.BClaims.StateRoot), make([]byte, constants.HashLen)) {
		// Do NOT request all-zero byte slice stateRoot
		if err := ssm.database.SetPendingNodeKey(txn, utils.CopySlice(snapShotBlockHeader.BClaims.StateRoot), 0); err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
	}
	if err := ssm.database.SetPendingHdrNodeKey(txn, utils.CopySlice(snapShotBlockHeader.BClaims.HeaderRoot), 0); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}
	canonicalBHTrieKey := ssm.database.MakeHeaderTrieKeyFromHeight(snapShotBlockHeader.BClaims.Height)
	bHash, err := snapShotBlockHeader.BlockHash()
	if err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}
	if err := ssm.database.SetPendingHdrLeafKey(txn, canonicalBHTrieKey, bHash); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}
	return nil
}

func (ssm *SnapShotManager) cleanupDatabase(txn *badger.Txn) error {
	if err := ssm.database.DropPendingLeafKeys(txn); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}
	if err := ssm.database.DropPendingNodeKeys(txn); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}
	if err := ssm.appHandler.BeginSnapShotSync(txn); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}
	return nil
}

func (ssm *SnapShotManager) Update(txn *badger.Txn, snapShotBlockHeader *objs.BlockHeader) (bool, error) {
	// a difference in height implies the target has changed for the canonical
	// state, thus re-init the object and drop all stale state
	// return after the drop so the next iteration sees the dropped state is in the
	// db transaction
	if ssm.snapShotHeight.Get() != snapShotBlockHeader.BClaims.Height {
		err := ssm.startFastSync(txn, snapShotBlockHeader)
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return false, err
		}
		return false, nil
	}
	if err := ssm.updateSync(txn, snapShotBlockHeader.BClaims.Height); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return false, err
	}
	hnCount, snCount, slCount, hlCount, bhCount, err := ssm.getKeyCounts(txn)
	if err != nil {
		utils.DebugTrace(ssm.logger, err)
		return false, err
	}
	logMsg := fmt.Sprintf("FastSyncing@%v |HN:%v HL:%v CBH:%v |SN:%v SL:%v |Prct:%v", snapShotBlockHeader.BClaims.Height, hnCount, hlCount, bhCount, snCount, slCount, (bhCount*100)/int(snapShotBlockHeader.BClaims.Height))
	ssm.status(logMsg)
	if err := ssm.updateDls(txn, snapShotBlockHeader.BClaims.Height, bhCount, hlCount); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return false, err
	}
	pCount := (int(snapShotBlockHeader.BClaims.Height) - bhCount)
	if pCount < 0 {
		pCount = 0
	}
	pCount += ssm.hdrLeafDLs.Size() + ssm.hdrNodeDLs.Size()
	pCount += ssm.stateNodeDLs.Size() + ssm.stateLeafDLs.Size()
	pCount += snCount + slCount + hnCount + hlCount
	if pCount == 0 {
		if err := ssm.finalizeSync(txn, snapShotBlockHeader); err != nil {
			utils.DebugTrace(ssm.logger, err)
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (ssm *SnapShotManager) status(msg string) {
	select {
	case ssm.statusChan <- msg:
		return
	default:
		return
	}
}

func (ssm *SnapShotManager) loggingDelayer() {
	for {
		select {
		case msg := <-ssm.statusChan:
			ssm.logger.Info(msg)
		case <-ssm.closeChan:
			return
		}
		time.Sleep(5 * time.Second)
	}
}

func (ssm *SnapShotManager) getKeyCounts(txn *badger.Txn) (int, int, int, int, int, error) {
	hnCount, err := ssm.database.CountPendingHdrNodeKeys(txn)
	if err != nil {
		utils.DebugTrace(ssm.logger, err)
		return 0, 0, 0, 0, 0, err
	}
	snCount, err := ssm.database.CountPendingNodeKeys(txn)
	if err != nil {
		utils.DebugTrace(ssm.logger, err)
		return 0, 0, 0, 0, 0, err
	}
	slCount, err := ssm.database.CountPendingLeafKeys(txn)
	if err != nil {
		utils.DebugTrace(ssm.logger, err)
		return 0, 0, 0, 0, 0, err
	}
	hlCount, err := ssm.database.CountPendingHdrLeafKeys(txn)
	if err != nil {
		utils.DebugTrace(ssm.logger, err)
		return 0, 0, 0, 0, 0, err
	}
	bhCount, err := ssm.database.CountCommittedBlockHeaders(txn)
	if err != nil {
		utils.DebugTrace(ssm.logger, err)
		return 0, 0, 0, 0, 0, err
	}
	return hnCount, snCount, slCount, hlCount, bhCount, nil
}

func (ssm *SnapShotManager) finalizeSync(txn *badger.Txn, snapShotBlockHeader *objs.BlockHeader) error {
	ssm.finalizeOnce.Do(func() { close(ssm.finalizeFastSyncChan) })
	if err := ssm.database.UpdateHeaderTrieRootFastSync(txn, snapShotBlockHeader); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}
	if err := ssm.appHandler.FinalizeSnapShotRoot(txn, snapShotBlockHeader.BClaims.StateRoot, snapShotBlockHeader.BClaims.Height); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}
	return nil
}

func (ssm *SnapShotManager) updateDls(txn *badger.Txn, snapShotHeight uint32, bhCount int, hlCount int) error {
	if err := ssm.dlHdrNodes(txn, snapShotHeight); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}
	if err := ssm.dlHdrLeaves(txn, snapShotHeight); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}
	if hlCount == 0 && bhCount >= int(snapShotHeight)-int(constants.EpochLength) {
		if err := ssm.dlStateNodes(txn, snapShotHeight); err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
		if err := ssm.dlStateLeaves(txn, snapShotHeight); err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
	}
	return nil
}

func (ssm *SnapShotManager) updateSync(txn *badger.Txn, snapShotHeight uint32) error {
	if err := ssm.syncHdrNodes(txn, snapShotHeight); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}
	if err := ssm.syncHdrLeaves(txn, snapShotHeight); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}
	if err := ssm.syncTailingBlockHeaders(txn, snapShotHeight); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}
	if err := ssm.syncStateNodes(txn, snapShotHeight); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}
	if err := ssm.syncStateLeaves(txn, snapShotHeight); err != nil {
		utils.DebugTrace(ssm.logger, err)
		return err
	}
	return nil
}

func (ssm *SnapShotManager) syncHdrNodes(txn *badger.Txn, snapShotHeight uint32) error {
	// get a set of node header keys and sync the header trie based on those
	// node keys
	nodeHdrKeys := ssm.hdrNodeCache.getNodeKeys(maxNumber)
	for i := 0; i < len(nodeHdrKeys); i++ {
		resp, err := ssm.hdrNodeCache.pop(utils.CopySlice(nodeHdrKeys[i].key[:]))
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			continue
		}
		// remove the keys from the pending set in the database
		err = ssm.database.DeletePendingHdrNodeKey(txn, utils.CopySlice(nodeHdrKeys[i].key[:]))
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
		}
		// store all of those nodes into the database and get new pending keys
		pendingBatch, newLayer, lvs, err := ssm.database.SetSnapShotHdrNode(txn, resp.batch, resp.root, resp.layer)
		if err != nil {
			// should not return if err invalid
			utils.DebugTrace(ssm.logger, err)
			continue
		}
		// store new pending keys to db
		for j := 0; j < len(pendingBatch); j++ {
			nk, err := newNodeKey(utils.CopySlice(pendingBatch[j]))
			if err != nil {
				utils.DebugTrace(ssm.logger, err)
				return err
			}
			if ssm.hdrNodeCache.contains(nk) {
				continue
			}
			if ssm.hdrNodeDLs.Contains(nk) {
				continue
			}
			ok, err := ssm.database.ContainsSnapShotHdrNode(txn, utils.CopySlice(pendingBatch[j]))
			if err != nil {
				return err
			}
			if ok {
				continue
			}
			if err := ssm.database.SetPendingHdrNodeKey(txn, utils.CopySlice(pendingBatch[j]), newLayer); err != nil {
				utils.DebugTrace(ssm.logger, err)
				return err
			}
		}
		for j := 0; j < len(lvs); j++ {
			// TODO: Handle this error
			nk, _ := newNodeKey(lvs[j].Key)
			if ssm.hdrLeafDLs.Contains(nk) {
				continue
			}
			if ssm.hdrLeafCache.contains(nk) {
				continue
			}
			exists := true
			_, err = ssm.database.GetCommittedBlockHeaderByHash(txn, utils.CopySlice(lvs[j].Value))
			if err != nil {
				if err != badger.ErrKeyNotFound {
					utils.DebugTrace(ssm.logger, err)
					return err
				}
				exists = false
			}
			if exists {
				continue
			}
			if err := ssm.database.SetPendingHdrLeafKey(txn, utils.CopySlice(lvs[j].Key), utils.CopySlice(lvs[j].Value)); err != nil {
				utils.DebugTrace(ssm.logger, err)
				return err
			}
		}
	}
	return nil
}

func (ssm *SnapShotManager) findTailSyncHeight(txn *badger.Txn, thisHeight int, lastHeight int) (*objs.BlockHeader, error) {
	var lastKnown *objs.BlockHeader
	for i := thisHeight; lastHeight < i; i-- {
		if i <= 2 {
			break
		}
		if lastKnown == nil && i%int(constants.EpochLength) == 0 {
			bh, err := ssm.database.GetCommittedBlockHeader(txn, uint32(i))
			if err != nil {
				if err != badger.ErrKeyNotFound {
					return nil, err
				}
			}
			if bh != nil {
				lastKnown = bh
			}
			if lastKnown == nil {
				ssbh, err := ssm.database.GetSnapshotBlockHeader(txn, uint32(i))
				if err != nil {
					return nil, err
				}
				if ssbh != nil {
					if err := ssm.database.SetCommittedBlockHeaderFastSync(txn, ssbh); err != nil {
						utils.DebugTrace(ssm.logger, err)
						return nil, err
					}
					lastKnown = ssbh
				}
			}
		}
		bh, err := ssm.database.GetCommittedBlockHeader(txn, uint32(i))
		if err != nil {
			if err != badger.ErrKeyNotFound {
				utils.DebugTrace(ssm.logger, err)
				return nil, err
			}
			continue
		}
		lastKnown = bh
	}
	if lastKnown == nil {
		return nil, badger.ErrKeyNotFound
	}
	return lastKnown, nil
}

func (ssm *SnapShotManager) syncTailingBlockHeaders(txn *badger.Txn, snapShotHeight uint32) error {
	count := 0
	{
		if ssm.tailSyncHeight == 2 {
			return nil
		}
		lastKnown, err := ssm.findTailSyncHeight(txn, int(ssm.tailSyncHeight), 1)
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
		if lastKnown.BClaims.Height != ssm.tailSyncHeight {
			ssm.tailSyncHeight = lastKnown.BClaims.Height
			key := ssm.database.MakeHeaderTrieKeyFromHeight(lastKnown.BClaims.Height - 1)
			exists := true
			_, err = ssm.database.GetPendingHdrLeafKey(txn, utils.CopySlice(key))
			if err != nil {
				if err != badger.ErrKeyNotFound {
					utils.DebugTrace(ssm.logger, err)
					return err
				}
				exists = false
			}
			nk, err := newNodeKey(utils.CopySlice(key))
			if err != nil {
				utils.DebugTrace(ssm.logger, err)
				return err
			}
			if ssm.hdrLeafDLs.Contains(nk) {
				exists = true
			}
			if ssm.hdrLeafCache.contains(nk) {
				exists = true
			}
			if !exists {
				value := utils.CopySlice(lastKnown.BClaims.PrevBlock)
				if err := ssm.database.SetPendingHdrLeafKey(txn, utils.CopySlice(key), utils.CopySlice(value)); err != nil {
					utils.DebugTrace(ssm.logger, err)
					return err
				}
				count++
			}
		}
	}
	if ssm.tailSyncHeight <= constants.EpochLength {
		return nil
	}
	start := (int(ssm.tailSyncHeight) - int(constants.EpochLength)) % int(constants.EpochLength)
	start = int(start) * int(constants.EpochLength)
	if start <= int(constants.EpochLength) {
		return nil
	}
	for i := start; int(constants.EpochLength)+1 < i; i -= int(constants.EpochLength) {
		if count >= 32 {
			return nil
		}
		count++
		known, err := ssm.findTailSyncHeight(txn, i, i-int(constants.EpochLength))
		if err != nil {
			if err != badger.ErrKeyNotFound {
				utils.DebugTrace(ssm.logger, err)
				return err
			}
			continue
		}
		if known.BClaims.Height == 1 {
			return nil
		}
		key := ssm.database.MakeHeaderTrieKeyFromHeight(known.BClaims.Height - 1)
		nk, err := newNodeKey(utils.CopySlice(key))
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
		if ssm.hdrLeafDLs.Contains(nk) {
			continue
		}
		if ssm.hdrLeafCache.contains(nk) {
			continue
		}
		_, err = ssm.database.GetPendingHdrLeafKey(txn, utils.CopySlice(key))
		if err != nil {
			if err != badger.ErrKeyNotFound {
				utils.DebugTrace(ssm.logger, err)
				return err
			}
			continue
		}
		value := utils.CopySlice(known.BClaims.PrevBlock)
		if err := ssm.database.SetPendingHdrLeafKey(txn, utils.CopySlice(key), utils.CopySlice(value)); err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
	}
	return nil
}

func (ssm *SnapShotManager) dlHdrNodes(txn *badger.Txn, snapShotHeight uint32) error {
	count := 0
	// iterate the pending keys in database and start downloads for each pending until the limit is reached
	iter := ssm.database.GetPendingHdrNodeKeysIter(txn)
	defer iter.Close()
	for {
		if count >= maxNumber {
			return nil
		}
		dlroot, dllayer, isDone, err := iter.Next()
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
		if isDone {
			return nil
		}
		nk, err := newNodeKey(utils.CopySlice(dlroot))
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
		ok, err := ssm.database.ContainsSnapShotHdrNode(txn, utils.CopySlice(dlroot))
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
		if ok {
			if err := ssm.database.DeletePendingHdrNodeKey(txn, dlroot); err != nil {
				utils.DebugTrace(ssm.logger, err)
				return nil
			}
			continue
		}
		if ssm.hdrNodeDLs.Contains(nk) {
			continue
		}
		if ssm.hdrNodeCache.contains(nk) {
			continue
		}
		r := &dlReq{
			snapShotHeight: snapShotHeight,
			key:            utils.CopySlice(dlroot),
			layer:          dllayer,
		}
		ssm.hdrNodeDLs.Push(nk)
		select {
		case ssm.hdrNodeDlChan <- r:
			count++
		default:
			ssm.hdrNodeDLs.Pop(nk)
			return nil
		}
	}
}

func (ssm *SnapShotManager) syncStateNodes(txn *badger.Txn, snapShotHeight uint32) error {
	// get keys from node cache and use those keys to store elements from the
	// node cache into the database as well as to get the leaf keys and store
	// those into the database as well
	nodeKeys := ssm.stateNodeCache.getNodeKeys(snapShotHeight, maxNumber)
	//for each key
	for i := 0; i < len(nodeKeys); i++ {
		resp, err := ssm.stateNodeCache.pop(snapShotHeight, utils.CopySlice(nodeKeys[i].key[:]))
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			continue
		}
		// remove the keys from the pending set in the database
		err = ssm.database.DeletePendingNodeKey(txn, utils.CopySlice(nodeKeys[i].key[:]))
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
		// store all of those nodes into the database and get new pending keys
		pendingBatch, newLayer, lvs, err := ssm.appHandler.StoreSnapShotNode(txn, utils.CopySlice(resp.batch), utils.CopySlice(resp.root), resp.layer)
		if err != nil {
			// should not return if err invalid
			utils.DebugTrace(ssm.logger, err)
			continue
		}
		// store pending leaves ( blockheader height/hashes)
		for j := 0; j < len(lvs); j++ {
			exists := true
			if _, err := ssm.appHandler.GetSnapShotStateData(txn, utils.CopySlice(lvs[j].Key)); err != nil {
				if err != badger.ErrKeyNotFound {
					utils.DebugTrace(ssm.logger, err)
					return err
				}
				exists = false
			}
			if exists {
				continue
			}
			if err := ssm.database.SetPendingLeafKey(txn, utils.CopySlice(lvs[j].Key), utils.CopySlice(lvs[j].Value)); err != nil {
				utils.DebugTrace(ssm.logger, err)
				return err
			}
		}
		// store new pending keys to db
		for j := 0; j < len(pendingBatch); j++ {
			if err := ssm.database.SetPendingNodeKey(txn, utils.CopySlice(pendingBatch[j]), newLayer); err != nil {
				utils.DebugTrace(ssm.logger, err)
				return err
			}
		}
	}
	return nil
}

func (ssm *SnapShotManager) dlStateNodes(txn *badger.Txn, snapShotHeight uint32) error {
	// iterate the pending keys in database and start downloads for each pending until the limit is reached
	iter := ssm.database.GetPendingNodeKeysIter(txn)
	defer iter.Close()
	for {
		dlroot, dllayer, isDone, err := iter.Next()
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
		if isDone {
			return nil
		}
		nk, err := newNodeKey(dlroot)
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
		if ssm.stateNodeDLs.Contains(nk) {
			continue
		}
		if ssm.stateNodeCache.contains(snapShotHeight, nk) {
			continue
		}
		r := &dlReq{
			snapShotHeight: snapShotHeight,
			layer:          dllayer,
			key:            dlroot,
		}
		ssm.stateNodeDLs.Push(nk)
		select {
		case ssm.stateNodeDlChan <- r:
		default:
			ssm.stateNodeDLs.Pop(nk)
			return nil
		}
	}
}

func (ssm *SnapShotManager) syncStateLeaves(txn *badger.Txn, snapShotHeight uint32) error {
	// get keys from state cache use those to store the state into the db
	leafKeys := ssm.stateLeafCache.getLeafKeys(snapShotHeight, maxNumber)
	// loop through LeafNode keys and retrieve from stateCache;
	// store state before deleting from database.
	for i := 0; i < len(leafKeys); i++ {
		resp, err := ssm.stateLeafCache.pop(snapShotHeight, utils.CopySlice(leafKeys[i].key[:]))
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			continue
		}
		// remove the keys from the pending set in the database
		err = ssm.database.DeletePendingLeafKey(txn, utils.CopySlice(leafKeys[i].key[:]))
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
		}
		// store key-value state
		err = ssm.appHandler.StoreSnapShotStateData(txn, utils.CopySlice(resp.key), utils.CopySlice(resp.value), utils.CopySlice(resp.data))
		if err != nil {
			// should not return if err invalid
			utils.DebugTrace(ssm.logger, err)
			continue
		}
	}
	return nil
}

func (ssm *SnapShotManager) dlStateLeaves(txn *badger.Txn, snapShotHeight uint32) error {
	// iterate the pending keys in database and start downloads for each pending until the limit is reached
	iter := ssm.database.GetPendingLeafKeysIter(txn)
	defer iter.Close()
	for {
		dlkey, dlvalue, isDone, err := iter.Next()
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
		if isDone {
			return nil
		}
		nk, err := newNodeKey(dlkey)
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
		if ssm.stateLeafDLs.Contains(nk) {
			continue
		}
		if ssm.stateLeafCache.contains(snapShotHeight, nk) {
			continue
		}
		r := &dlReq{
			snapShotHeight: snapShotHeight,
			key:            dlkey,
			value:          dlvalue,
		}
		ssm.stateLeafDLs.Push(nk)
		select {
		case ssm.stateLeafDlChan <- r:
		default:
			ssm.stateLeafDLs.Pop(nk)
			return nil
		}
	}
}

func (ssm *SnapShotManager) syncHdrLeaves(txn *badger.Txn, snapShotHeight uint32) error {
	// get keys from state cache use those to store the state into the db
	leafKeys := ssm.hdrLeafCache.getLeafKeys(maxNumber)
	// loop through LeafNode keys and retrieve from stateCache;
	// store state before deleting from database.
	for i := 0; i < len(leafKeys); i++ {
		resp, err := ssm.hdrLeafCache.pop(leafKeys[i].key[:])
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			continue
		}
		err = ssm.database.DeletePendingHdrLeafKey(txn, leafKeys[i].key[:])
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
		bh := &objs.BlockHeader{}
		err = bh.UnmarshalBinary(resp.data)
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
		err = ssm.database.SetCommittedBlockHeaderFastSync(txn, bh)
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
	}
	return nil
}

func (ssm *SnapShotManager) dlHdrLeaves(txn *badger.Txn, snapShotHeight uint32) error {
	// iterate the pending keys in database and start downloads for each pending until the limit is reached
	iter := ssm.database.GetPendingHdrLeafKeysIter(txn)
	defer iter.Close()
	for {
		dlkey, dlvalue, isDone, err := iter.Next()
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
		if isDone {
			break
		}
		nk, err := newNodeKey(dlkey)
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return err
		}
		if ssm.hdrLeafDLs.Contains(nk) {
			continue
		}
		if ssm.hdrLeafCache.contains(nk) {
			continue
		}
		r := &dlReq{
			snapShotHeight: snapShotHeight,
			key:            dlkey,
			value:          dlvalue,
		}
		ssm.hdrLeafDLs.Push(nk)
		select {
		case ssm.hdrLeafDlChan <- r:
		default:
			ssm.hdrLeafDLs.Pop(nk)
			return nil
		}
	}
	return nil
}

func (ssm *SnapShotManager) worker(killChan <-chan struct{}) {
	defer func() {
		ssm.Lock()
		ssm.numWorkers--
		ssm.Unlock()
	}()
	for {
		select {
		case <-killChan:
			return
		case <-ssm.closeChan:
			return
		case <-time.After(1 * time.Second):
			ssm.Lock()
			if ssm.numWorkers > minWorkers {
				ssm.Unlock()
				return
			}
			ssm.Unlock()
		case w := <-ssm.workChan:
			w()
		}
	}
}

func (ssm *SnapShotManager) sendWork(w func()) {
	for {
		select {
		case <-ssm.closeChan:
			return
		case ssm.workChan <- w:
			return
		default:
			ssm.Lock()
			if ssm.numWorkers < maxNumber {
				ssm.numWorkers++
				go ssm.worker(ssm.finalizeFastSyncChan)
				ssm.Unlock()
			} else {
				ssm.Unlock()
				select {
				case <-ssm.closeChan:
					return
				case ssm.workChan <- w:
					return
				}
			}

		}
	}
}

func (ssm *SnapShotManager) downloadWithRetryStateNodeWorker() {
	for {
		select {
		case <-ssm.closeChan:
			return
		case dl := <-ssm.stateNodeDlChan:
			ssm.sendWork(ssm.downloadWithRetryStateNodeClosure(dl))
		}
	}
}

func (ssm *SnapShotManager) downloadWithRetryHdrLeafWorker() {
	for {
		select {
		case <-ssm.closeChan:
			return
		default:
			var stop <-chan time.Time
			cache := []*dlReq{}
			func() {
				for {
					select {
					case <-ssm.closeChan:
						return
					case <-stop:
						return
					case dl := <-ssm.hdrLeafDlChan:
						if stop == nil {
							stop = time.After(50 * time.Millisecond)
						}
						cache = append(cache, dl)
						if len(cache) >= int(constants.EpochLength/2) {
							return
						}
					}
				}
			}()
			ssm.sendWork(ssm.downloadWithRetryHdrLeafClosure(cache))
		}
	}
}

func (ssm *SnapShotManager) downloadWithRetryHdrNodeWorker() {
	for {
		select {
		case <-ssm.closeChan:
			return
		case dl := <-ssm.hdrNodeDlChan:
			ssm.sendWork(ssm.downloadWithRetryHdrNodeClosure(dl))
		}
	}
}

func (ssm *SnapShotManager) downloadWithRetryStateLeafWorker() {
	for {
		select {
		case <-ssm.closeChan:
			return
		case dl := <-ssm.stateLeafDlChan:
			ssm.sendWork(ssm.downloadWithRetryStateLeafClosure(dl))
		}
	}
}

func (ssm *SnapShotManager) downloadWithRetryStateNodeClosure(dl *dlReq) workFunc {
	snapShotHeight := dl.snapShotHeight
	root := dl.key
	layer := dl.layer
	nk, _ := newNodeKey(root)
	return func() {
		defer ssm.stateNodeDLs.Pop(nk)
		if snapShotHeight < ssm.snapShotHeight.Get() {
			return
		}
		opts := []grpc.CallOption{
			grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
			grpc_retry.WithMax(maxRetryCount),
		}
		resp, err := ssm.requestBus.RequestP2PGetSnapShotNode(context.Background(), snapShotHeight, root, opts...)
		if err != nil {
			utils.DebugTrace(ssm.logger, err)
			return
		}
		if len(resp) == 0 {
			return
		}
		nr := &nodeResponse{
			snapShotHeight: snapShotHeight,
			layer:          layer,
			root:           root,
			batch:          resp,
		}
		//    store to the cache
		ssm.stateNodeCache.insert(snapShotHeight, nr)
	}
}

func (ssm *SnapShotManager) downloadWithRetryHdrLeafClosure(dl []*dlReq) workFunc {
	return func() {
		heightList := make([]uint32, len(dl))
		hashMap := make(map[uint32][]byte)
		keyMap := make(map[uint32]nodeKey)
		for i := 0; i < len(dl); i++ {
			key := dl[i].key
			value := dl[i].value
			nk, _ := newNodeKey(key)
			blockHeight, _ := utils.UnmarshalUint32(key[0:4])
			keyMap[blockHeight] = nk
			heightList[i] = blockHeight
			hashMap[blockHeight] = utils.CopySlice(value)
			defer ssm.hdrLeafDLs.Pop(nk)
		}
		opts := []grpc.CallOption{
			grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
			grpc_retry.WithMax(maxRetryCount),
		}
		peerOpt := middleware.NewPeerInterceptor()
		newOpts := append(opts, peerOpt)
		resp, err := ssm.requestBus.RequestP2PGetBlockHeaders(context.Background(), heightList, newOpts...)
		if err != nil {
			return
		}
		peer := peerOpt.Peer()
		for i := 0; i < len(resp); i++ {
			if resp[i] == nil {
				peer.Feedback(-2)
				continue
			}
			height := resp[i].BClaims.Height
			bhash, ok := hashMap[height]
			if !ok {
				peer.Feedback(-2)
				continue
			}
			bhashResp, err := resp[i].BlockHash()
			if err != nil {
				peer.Feedback(-2)
				utils.DebugTrace(ssm.logger, err)
				continue
			}
			if !bytes.Equal(bhash, bhashResp) {
				peer.Feedback(-2)
				utils.DebugTrace(ssm.logger, errors.New("bad block hash"))
				continue
			}
			bhBytes, err := resp[i].MarshalBinary()
			if err != nil {
				peer.Feedback(-2)
				utils.DebugTrace(ssm.logger, err)
				return
			}
			nk := keyMap[height]
			key := nk.key[:]
			sr := &stateResponse{
				key:   utils.CopySlice(key),
				value: utils.CopySlice(bhash),
				data:  bhBytes,
			}
			//    store to the cache
			peer.Feedback(1)
			ssm.hdrLeafCache.insert(sr)
		}
	}
}

func (ssm *SnapShotManager) downloadWithRetryStateLeafClosure(dl *dlReq) workFunc {
	snapShotHeight := dl.snapShotHeight
	key := dl.key
	value := dl.value
	nk, _ := newNodeKey(key)
	return func() {
		defer ssm.stateLeafDLs.Pop(nk)
		if snapShotHeight < ssm.snapShotHeight.Get() {
			return
		}
		opts := []grpc.CallOption{
			grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
			grpc_retry.WithMax(maxRetryCount),
		}
		peerOpt := middleware.NewPeerInterceptor()
		newOpts := append(opts, peerOpt)
		resp, err := ssm.requestBus.RequestP2PGetSnapShotStateData(context.Background(), key, newOpts...)

		if err != nil {
			return
		}
		peer := peerOpt.Peer()
		if len(resp) == 0 {
			peer.Feedback(-2)
			return
		}
		sr := &stateResponse{
			snapShotHeight: snapShotHeight,
			key:            utils.CopySlice(key),
			value:          utils.CopySlice(value),
			data:           utils.CopySlice(resp),
		}
		//    store to the cache
		ssm.stateLeafCache.insert(snapShotHeight, sr)
	}
}

func (ssm *SnapShotManager) downloadWithRetryHdrNodeClosure(dl *dlReq) workFunc {
	snapShotHeight := dl.snapShotHeight
	root := dl.key
	layer := dl.layer
	nk, _ := newNodeKey(root)
	return func() {
		opts := []grpc.CallOption{
			grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
			grpc_retry.WithMax(maxRetryCount),
		}
		peerOpt := middleware.NewPeerInterceptor()
		newOpts := append(opts, peerOpt)
		defer ssm.hdrNodeDLs.Pop(nk)
		resp, err := ssm.requestBus.RequestP2PGetSnapShotHdrNode(context.Background(), root, newOpts...)
		if err != nil {
			return
		}
		peer := peerOpt.Peer()
		if len(resp) == 0 {
			peer.Feedback(-2)
			return
		}
		nr := &nodeResponse{
			snapShotHeight: snapShotHeight,
			layer:          layer,
			root:           root,
			batch:          resp,
		}
		//    store to the cache
		if err := ssm.hdrNodeCache.insert(nr); err != nil {
			utils.DebugTrace(ssm.logger, err)
		}
	}
}
