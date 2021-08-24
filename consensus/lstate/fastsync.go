package lstate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/MadBase/MadNet/consensus/appmock"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/consensus/request"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
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
const minWorkers = chanBuffering
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
	key [32]byte
}

func newNodeKey(s []byte) (nodeKey, error) {
	if len(s) != 32 {
		return nodeKey{}, errorz.ErrInvalid{}.New("Error in newNodeKey: byte slice not 32 bytes")
	}
	nk := nodeKey{}
	nk.key = [32]byte{}
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

func (nc *nodeCache) Init() error {
	nc.objs = make(map[uint32]map[nodeKey]*nodeResponse)
	return nil
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

func (sc *stateCache) Init() error {
	sc.objs = make(map[uint32]map[nodeKey]*stateResponse)
	return nil
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
	objs      map[nodeKey]*stateResponse
	minHeight uint32
}

func (sc *bhCache) Init() error {
	sc.objs = make(map[nodeKey]*stateResponse)
	return nil
}

func (sc *bhCache) getLeafKeys(maxNumber int) []nodeKey {
	sc.RLock()
	defer sc.RUnlock()
	nodeKeys := []nodeKey{}
	for k := range sc.objs {
		nodeKeys = append(nodeKeys, k)
		if len(nodeKeys) >= maxNumber {
			break
		}
	}
	return nodeKeys
}

func (sc *bhCache) contains(nk nodeKey) bool {
	sc.RLock()
	defer sc.RUnlock()
	if sc.objs[nk] == nil {
		return false
	}
	return true
}

func (sc *bhCache) insert(sr *stateResponse) error {
	sc.Lock()
	defer sc.Unlock()
	nk, err := newNodeKey(sr.key)
	if err != nil {
		return err
	}
	sc.objs[nk] = sr
	return nil
}

func (sc *bhCache) pop(key []byte) (*stateResponse, error) {
	sc.Lock()
	defer sc.Unlock()
	nk, err := newNodeKey(key)
	if err != nil {
		return nil, err
	}
	if sc.objs[nk] == nil {
		return nil, errorz.ErrInvalid{}.New("Error in bhCache.pop: missing key in pop request")
	}
	result := sc.objs[nk]
	delete(sc.objs, nk)
	return result, nil
}

type bhNodeCache struct {
	sync.RWMutex
	objs      map[nodeKey]*nodeResponse
	minHeight uint32
}

func (sc *bhNodeCache) Init() error {
	sc.objs = make(map[nodeKey]*nodeResponse)
	return nil
}

func (sc *bhNodeCache) getNodeKeys(maxNumber int) []nodeKey {
	sc.RLock()
	defer sc.RUnlock()
	nodeKeys := []nodeKey{}
	for k := range sc.objs {
		nodeKeys = append(nodeKeys, k)
		if len(nodeKeys) >= maxNumber {
			break
		}
	}
	return nodeKeys
}

func (sc *bhNodeCache) contains(nk nodeKey) bool {
	sc.RLock()
	defer sc.RUnlock()
	if sc.objs[nk] == nil {
		return false
	}
	return true
}

func (sc *bhNodeCache) insert(sr *nodeResponse) error {
	sc.Lock()
	defer sc.Unlock()
	nk, err := newNodeKey(sr.root)
	if err != nil {
		return err
	}
	sc.objs[nk] = sr
	return nil
}

func (sc *bhNodeCache) pop(key []byte) (*nodeResponse, error) {
	sc.Lock()
	defer sc.Unlock()
	nk, err := newNodeKey(key)
	if err != nil {
		return nil, err
	}
	if sc.objs[nk] == nil {
		return nil, errorz.ErrInvalid{}.New("Error in bhNodeCache.pop: missing key in pop request")
	}
	result := sc.objs[nk]
	delete(sc.objs, nk)
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
	appHandler appmock.Application
	requestBus *request.Client

	database       *db.Database
	logger         *logrus.Logger
	snapShotHeight *atomicU32

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
	reqCount int
	reqAvg   float64

	statusChan chan string
	closeChan  chan struct{}
	closeOnce  sync.Once

	finalizeFastSyncChan chan struct{}
	finalizeOnce         sync.Once

	tailSyncHeight uint32
}

// Init initializes the SnapShotManager
func (ndm *SnapShotManager) Init(database *db.Database) error {
	ndm.snapShotHeight = new(atomicU32)
	ndm.logger = logging.GetLogger(constants.LoggerConsensus)
	ndm.database = database
	ndm.hdrNodeCache = &bhNodeCache{}
	if err := ndm.hdrNodeCache.Init(); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	ndm.hdrLeafCache = &bhCache{}
	if err := ndm.hdrLeafCache.Init(); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	ndm.stateNodeCache = &nodeCache{}
	if err := ndm.stateNodeCache.Init(); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	ndm.stateLeafCache = &stateCache{}
	if err := ndm.stateLeafCache.Init(); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	ndm.hdrNodeDLs = &downloadTracker{
		sync.RWMutex{},
		make(map[nodeKey]bool),
	}
	ndm.hdrLeafDLs = &downloadTracker{
		sync.RWMutex{},
		make(map[nodeKey]bool),
	}
	ndm.stateLeafDLs = &downloadTracker{
		sync.RWMutex{},
		make(map[nodeKey]bool),
	}
	ndm.stateNodeDLs = &downloadTracker{
		sync.RWMutex{},
		make(map[nodeKey]bool),
	}
	ndm.statusChan = make(chan string)
	ndm.hdrLeafDlChan = make(chan *dlReq, chanBuffering)
	ndm.hdrNodeDlChan = make(chan *dlReq, chanBuffering)
	ndm.stateNodeDlChan = make(chan *dlReq, chanBuffering)
	ndm.stateLeafDlChan = make(chan *dlReq, chanBuffering)
	ndm.workChan = make(chan workFunc, chanBuffering*2)
	ndm.closeChan = make(chan struct{})
	ndm.closeOnce = sync.Once{}

	go ndm.downloadWithRetryHdrLeafWorker()
	go ndm.downloadWithRetryHdrNodeWorker()
	go ndm.downloadWithRetryStateLeafWorker()
	go ndm.downloadWithRetryStateNodeWorker()
	go ndm.loggingDelayer()

	return nil
}

func (ndm *SnapShotManager) close() {
	ndm.closeOnce.Do(func() {
		close(ndm.closeChan)
	})
}

func (ndm *SnapShotManager) startFastSync(txn *badger.Txn, snapShotBlockHeader *objs.BlockHeader) error {
	if ndm.finalizeFastSyncChan == nil {
		ndm.finalizeOnce = sync.Once{}
		ndm.finalizeFastSyncChan = make(chan struct{})
		for i := 0; i < minWorkers; i++ {
			go ndm.worker(ndm.finalizeFastSyncChan)
		}
	}
	ndm.tailSyncHeight = snapShotBlockHeader.BClaims.Height
	if err := ndm.database.SetCommittedBlockHeaderFastSync(txn, snapShotBlockHeader); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}

	ndm.snapShotHeight.Set(snapShotBlockHeader.BClaims.Height)

	// call dropBefore on the caches
	ndm.stateNodeCache.dropBefore(snapShotBlockHeader.BClaims.Height)
	ndm.stateLeafCache.dropBefore(snapShotBlockHeader.BClaims.Height)

	// cleanup the db of any previous data
	if err := ndm.cleanupDatabase(txn); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}

	// insert the request for the root node into the database
	if !bytes.Equal(utils.CopySlice(snapShotBlockHeader.BClaims.StateRoot), make([]byte, constants.HashLen)) {
		// Do NOT request all-zero byte slice stateRoot
		if err := ndm.database.SetPendingNodeKey(txn, utils.CopySlice(snapShotBlockHeader.BClaims.StateRoot), 0); err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
	}
	if err := ndm.database.SetPendingHdrNodeKey(txn, utils.CopySlice(snapShotBlockHeader.BClaims.HeaderRoot), 0); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	canonicalBHTrieKey := ndm.database.MakeHeaderTrieKeyFromHeight(snapShotBlockHeader.BClaims.Height)
	bHash, err := snapShotBlockHeader.BlockHash()
	if err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.database.SetPendingHdrLeafKey(txn, canonicalBHTrieKey, bHash); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	return nil
}

func (ndm *SnapShotManager) cleanupDatabase(txn *badger.Txn) error {
	if err := ndm.database.DropPendingLeafKeys(txn); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.database.DropPendingNodeKeys(txn); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.appHandler.BeginSnapShotSync(txn); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	return nil
}

func (ndm *SnapShotManager) Update(txn *badger.Txn, snapShotBlockHeader *objs.BlockHeader) (bool, error) {
	// a difference in height implies the target has changed for the canonical
	// state, thus re-init the object and drop all stale data
	// return after the drop so the next iteration sees the dropped data is in the
	// db transaction
	if ndm.snapShotHeight.Get() != snapShotBlockHeader.BClaims.Height {
		err := ndm.startFastSync(txn, snapShotBlockHeader)
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return false, err
		}
		return false, nil
	}
	if err := ndm.updateSync(txn, snapShotBlockHeader.BClaims.Height); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return false, err
	}
	hnCount, snCount, slCount, hlCount, bhCount, err := ndm.getKeyCounts(txn)
	if err != nil {
		utils.DebugTrace(ndm.logger, err)
		return false, err
	}
	logMsg := fmt.Sprintf("FastSyncing@%v |HN:%v HL:%v CBH:%v |SN:%v SL:%v |Prct:%v", snapShotBlockHeader.BClaims.Height, hnCount, hlCount, bhCount, snCount, slCount, (bhCount*100)/int(snapShotBlockHeader.BClaims.Height))
	ndm.status(logMsg)
	if err := ndm.updateDls(txn, snapShotBlockHeader.BClaims.Height, bhCount, hlCount); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return false, err
	}
	pCount := (int(snapShotBlockHeader.BClaims.Height) - bhCount)
	if pCount < 0 {
		pCount = 0
	}
	pCount += ndm.hdrLeafDLs.Size() + ndm.hdrNodeDLs.Size()
	pCount += ndm.stateNodeDLs.Size() + ndm.stateLeafDLs.Size()
	pCount += snCount + slCount + hnCount + hlCount
	if pCount == 0 {
		if err := ndm.finalizeSync(txn, snapShotBlockHeader); err != nil {
			utils.DebugTrace(ndm.logger, err)
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (ndm *SnapShotManager) status(msg string) {
	select {
	case ndm.statusChan <- msg:
		return
	default:
		return
	}
}

func (ndm *SnapShotManager) loggingDelayer() {
	for {
		select {
		case msg := <-ndm.statusChan:
			ndm.logger.Info(msg)
		case <-ndm.closeChan:
			return
		}
		time.Sleep(5 * time.Second)
	}
}

func (ndm *SnapShotManager) getKeyCounts(txn *badger.Txn) (int, int, int, int, int, error) {
	hnCount, err := ndm.database.CountPendingHdrNodeKeys(txn)
	if err != nil {
		utils.DebugTrace(ndm.logger, err)
		return 0, 0, 0, 0, 0, err
	}
	snCount, err := ndm.database.CountPendingNodeKeys(txn)
	if err != nil {
		utils.DebugTrace(ndm.logger, err)
		return 0, 0, 0, 0, 0, err
	}
	slCount, err := ndm.database.CountPendingLeafKeys(txn)
	if err != nil {
		utils.DebugTrace(ndm.logger, err)
		return 0, 0, 0, 0, 0, err
	}
	hlCount, err := ndm.database.CountPendingHdrLeafKeys(txn)
	if err != nil {
		utils.DebugTrace(ndm.logger, err)
		return 0, 0, 0, 0, 0, err
	}
	bhCount, err := ndm.database.CountCommittedBlockHeaders(txn)
	if err != nil {
		utils.DebugTrace(ndm.logger, err)
		return 0, 0, 0, 0, 0, err
	}
	return hnCount, snCount, slCount, hlCount, bhCount, nil
}

func (ndm *SnapShotManager) finalizeSync(txn *badger.Txn, snapShotBlockHeader *objs.BlockHeader) error {
	ndm.finalizeOnce.Do(func() { close(ndm.finalizeFastSyncChan) })
	if err := ndm.database.UpdateHeaderTrieRootFastSync(txn, snapShotBlockHeader); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.appHandler.FinalizeSnapShotRoot(txn, snapShotBlockHeader.BClaims.StateRoot, snapShotBlockHeader.BClaims.Height); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	return nil
}

func (ndm *SnapShotManager) updateDls(txn *badger.Txn, snapShotHeight uint32, bhCount int, hlCount int) error {
	if err := ndm.dlHdrNodes(txn, snapShotHeight); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.dlHdrLeaves(txn, snapShotHeight); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if hlCount == 0 && bhCount >= int(snapShotHeight)-int(constants.EpochLength) {
		if err := ndm.dlStateNodes(txn, snapShotHeight); err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		if err := ndm.dlStateLeaves(txn, snapShotHeight); err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
	}
	return nil
}

func (ndm *SnapShotManager) updateSync(txn *badger.Txn, snapShotHeight uint32) error {
	if err := ndm.syncHdrNodes(txn, snapShotHeight); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.syncHdrLeaves(txn, snapShotHeight); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.syncTailingBlockHeaders(txn, snapShotHeight); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.syncStateNodes(txn, snapShotHeight); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.syncStateLeaves(txn, snapShotHeight); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	return nil
}

func (ndm *SnapShotManager) syncHdrNodes(txn *badger.Txn, snapShotHeight uint32) error {
	// get a set of node header keys and sync the header trie based on those
	// node keys
	nodeHdrKeys := ndm.hdrNodeCache.getNodeKeys(maxNumber)
	for i := 0; i < len(nodeHdrKeys); i++ {
		resp, err := ndm.hdrNodeCache.pop(utils.CopySlice(nodeHdrKeys[i].key[:]))
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			continue
		}
		// remove the keys from the pending set in the database
		err = ndm.database.DeletePendingHdrNodeKey(txn, utils.CopySlice(nodeHdrKeys[i].key[:]))
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
		}
		// store all of those nodes into the database and get new pending keys
		pendingBatch, newLayer, lvs, err := ndm.database.SetSnapShotHdrNode(txn, resp.batch, resp.root, resp.layer)
		if err != nil {
			// should not return if err invalid
			utils.DebugTrace(ndm.logger, err)
			continue
		}
		// store new pending keys to db
		for j := 0; j < len(pendingBatch); j++ {
			ok, err := ndm.database.ContainsSnapShotHdrNode(txn, utils.CopySlice(pendingBatch[j]))
			if err != nil {
				return err
			}
			if ok {
				continue
			}
			if err := ndm.database.SetPendingHdrNodeKey(txn, utils.CopySlice(pendingBatch[j]), newLayer); err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
		}
		for j := 0; j < len(lvs); j++ {
			nk, _ := newNodeKey(lvs[j].Key)
			if ndm.hdrLeafDLs.Contains(nk) {
				continue
			}
			if ndm.hdrLeafCache.contains(nk) {
				continue
			}
			exists := true
			_, err = ndm.database.GetCommittedBlockHeaderByHash(txn, utils.CopySlice(lvs[j].Value))
			if err != nil {
				if err != badger.ErrKeyNotFound {
					utils.DebugTrace(ndm.logger, err)
					return err
				}
				exists = false
			}
			if exists {
				continue
			}
			if err := ndm.database.SetPendingHdrLeafKey(txn, utils.CopySlice(lvs[j].Key), utils.CopySlice(lvs[j].Value)); err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
		}
	}
	return nil
}

func (ndm *SnapShotManager) findTailSyncHeight(txn *badger.Txn, thisHeight int, lastHeight int) (*objs.BlockHeader, error) {
	var lastKnown *objs.BlockHeader
	for i := thisHeight; lastHeight < i; i-- {
		if i <= 2 {
			break
		}
		if lastKnown == nil && i%int(constants.EpochLength) == 0 {
			bh, err := ndm.database.GetCommittedBlockHeader(txn, uint32(i))
			if err != nil {
				if err != badger.ErrKeyNotFound {
					return nil, err
				}
			}
			if bh != nil {
				lastKnown = bh
			}
			if lastKnown == nil {
				ssbh, err := ndm.database.GetSnapshotBlockHeader(txn, uint32(i))
				if err != nil {
					return nil, err
				}
				if ssbh != nil {
					if err := ndm.database.SetCommittedBlockHeaderFastSync(txn, ssbh); err != nil {
						utils.DebugTrace(ndm.logger, err)
						return nil, err
					}
					lastKnown = ssbh
				}
			}
		}
		bh, err := ndm.database.GetCommittedBlockHeader(txn, uint32(i))
		if err != nil {
			if err != badger.ErrKeyNotFound {
				utils.DebugTrace(ndm.logger, err)
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

func (ndm *SnapShotManager) syncTailingBlockHeaders(txn *badger.Txn, snapShotHeight uint32) error {
	count := 0
	{
		if ndm.tailSyncHeight == 2 {
			return nil
		}
		lastKnown, err := ndm.findTailSyncHeight(txn, int(ndm.tailSyncHeight), 1)
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		if lastKnown.BClaims.Height != ndm.tailSyncHeight {
			ndm.tailSyncHeight = lastKnown.BClaims.Height
			key := ndm.database.MakeHeaderTrieKeyFromHeight(lastKnown.BClaims.Height - 1)
			exists := true
			_, err = ndm.database.GetPendingHdrLeafKey(txn, utils.CopySlice(key))
			if err != nil {
				if err != badger.ErrKeyNotFound {
					utils.DebugTrace(ndm.logger, err)
					return err
				}
				exists = false
			}
			nk, err := newNodeKey(utils.CopySlice(key))
			if err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
			if ndm.hdrLeafDLs.Contains(nk) {
				exists = true
			}
			if ndm.hdrLeafCache.contains(nk) {
				exists = true
			}
			if !exists {
				value := utils.CopySlice(lastKnown.BClaims.PrevBlock)
				if err := ndm.database.SetPendingHdrLeafKey(txn, utils.CopySlice(key), utils.CopySlice(value)); err != nil {
					utils.DebugTrace(ndm.logger, err)
					return err
				}
				count++
			}
		}
	}
	if ndm.tailSyncHeight <= constants.EpochLength {
		return nil
	}
	start := int(ndm.tailSyncHeight) - int(constants.EpochLength)%int(constants.EpochLength)
	start = int(start) * int(constants.EpochLength)
	if start <= int(constants.EpochLength) {
		return nil
	}
	for i := start; int(constants.EpochLength)+1 < i; i -= int(constants.EpochLength) {
		if count >= 32 {
			return nil
		}
		count++
		known, err := ndm.findTailSyncHeight(txn, i, i-int(constants.EpochLength))
		if err != nil {
			if err != badger.ErrKeyNotFound {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
			continue
		}
		if known.BClaims.Height == 1 {
			return nil
		}
		key := ndm.database.MakeHeaderTrieKeyFromHeight(known.BClaims.Height - 1)
		nk, err := newNodeKey(utils.CopySlice(key))
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		if ndm.hdrLeafDLs.Contains(nk) {
			continue
		}
		if ndm.hdrLeafCache.contains(nk) {
			continue
		}
		_, err = ndm.database.GetPendingHdrLeafKey(txn, utils.CopySlice(key))
		if err != nil {
			if err != badger.ErrKeyNotFound {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
			continue
		}
		value := utils.CopySlice(known.BClaims.PrevBlock)
		if err := ndm.database.SetPendingHdrLeafKey(txn, utils.CopySlice(key), utils.CopySlice(value)); err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
	}
	return nil
}

func (ndm *SnapShotManager) dlHdrNodes(txn *badger.Txn, snapShotHeight uint32) error {
	count := 0
	// iterate the pending keys in database and start downloads for each pending until the limit is reached
	iter := ndm.database.GetPendingHdrNodeKeysIter(txn)
	defer iter.Close()
	for {
		if count >= maxNumber {
			return nil
		}
		dlroot, dllayer, isDone, err := iter.Next()
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		if isDone {
			return nil
		}
		nk, err := newNodeKey(utils.CopySlice(dlroot))
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		ok, err := ndm.database.ContainsSnapShotHdrNode(txn, utils.CopySlice(dlroot))
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		if ok {
			if err := ndm.database.DeletePendingHdrNodeKey(txn, dlroot); err != nil {
				utils.DebugTrace(ndm.logger, err)
				return nil
			}
			continue
		}
		if ndm.hdrNodeDLs.Contains(nk) {
			continue
		}
		if ndm.hdrNodeCache.contains(nk) {
			continue
		}
		r := &dlReq{
			snapShotHeight: snapShotHeight,
			key:            utils.CopySlice(dlroot),
			layer:          dllayer,
		}
		ndm.hdrNodeDLs.Push(nk)
		select {
		case ndm.hdrNodeDlChan <- r:
			count++
		default:
			ndm.hdrNodeDLs.Pop(nk)
			return nil
		}
	}
}

func (ndm *SnapShotManager) syncStateNodes(txn *badger.Txn, snapShotHeight uint32) error {
	// get keys from node cache and use those keys to store elements from the
	// node cache into the database as well as to get the leaf keys and store
	// those into the database as well
	nodeKeys := ndm.stateNodeCache.getNodeKeys(snapShotHeight, maxNumber)
	//for each key
	for i := 0; i < len(nodeKeys); i++ {
		resp, err := ndm.stateNodeCache.pop(snapShotHeight, utils.CopySlice(nodeKeys[i].key[:]))
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			continue
		}
		// remove the keys from the pending set in the database
		err = ndm.database.DeletePendingNodeKey(txn, utils.CopySlice(nodeKeys[i].key[:]))
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		// store all of those nodes into the database and get new pending keys
		pendingBatch, newLayer, lvs, err := ndm.appHandler.StoreSnapShotNode(txn, utils.CopySlice(resp.batch), utils.CopySlice(resp.root), resp.layer)
		if err != nil {
			// should not return if err invalid
			utils.DebugTrace(ndm.logger, err)
			continue
		}
		// store pending leaves ( blockheader height/hashes)
		for j := 0; j < len(lvs); j++ {
			exists := true
			if _, err := ndm.appHandler.GetSnapShotStateData(txn, utils.CopySlice(lvs[j].Key)); err != nil {
				if err != badger.ErrKeyNotFound {
					utils.DebugTrace(ndm.logger, err)
					return err
				}
				exists = false
			}
			if exists {
				continue
			}
			if err := ndm.database.SetPendingLeafKey(txn, utils.CopySlice(lvs[j].Key), utils.CopySlice(lvs[j].Value)); err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
		}
		// store new pending keys to db
		for j := 0; j < len(pendingBatch); j++ {
			if err := ndm.database.SetPendingNodeKey(txn, utils.CopySlice(pendingBatch[j]), newLayer); err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
		}
	}
	return nil
}

func (ndm *SnapShotManager) dlStateNodes(txn *badger.Txn, snapShotHeight uint32) error {
	// iterate the pending keys in database and start downloads for each pending until the limit is reached
	iter := ndm.database.GetPendingNodeKeysIter(txn)
	defer iter.Close()
	for {
		dlroot, dllayer, isDone, err := iter.Next()
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		if isDone {
			return nil
		}
		nk, err := newNodeKey(dlroot)
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		if ndm.stateNodeDLs.Contains(nk) {
			continue
		}
		if ndm.stateNodeCache.contains(snapShotHeight, nk) {
			continue
		}
		r := &dlReq{
			snapShotHeight: snapShotHeight,
			layer:          dllayer,
			key:            dlroot,
		}
		ndm.stateNodeDLs.Push(nk)
		select {
		case ndm.stateNodeDlChan <- r:
		default:
			ndm.stateNodeDLs.Pop(nk)
			return nil
		}
	}
}

func (ndm *SnapShotManager) syncStateLeaves(txn *badger.Txn, snapShotHeight uint32) error {
	// get keys from state cache use those to store the state data into the db
	leafKeys := ndm.stateLeafCache.getLeafKeys(snapShotHeight, maxNumber)
	// loop through LeafNode keys and retrieve from stateCache;
	// store data before deleting from database.
	for i := 0; i < len(leafKeys); i++ {
		resp, err := ndm.stateLeafCache.pop(snapShotHeight, utils.CopySlice(leafKeys[i].key[:]))
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			continue
		}
		// remove the keys from the pending set in the database
		err = ndm.database.DeletePendingLeafKey(txn, utils.CopySlice(leafKeys[i].key[:]))
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
		}
		// store key-value state data
		err = ndm.appHandler.StoreSnapShotStateData(txn, utils.CopySlice(resp.key), utils.CopySlice(resp.value), utils.CopySlice(resp.data))
		if err != nil {
			// should not return if err invalid
			utils.DebugTrace(ndm.logger, err)
			continue
		}
	}
	return nil
}

func (ndm *SnapShotManager) dlStateLeaves(txn *badger.Txn, snapShotHeight uint32) error {
	// iterate the pending keys in database and start downloads for each pending until the limit is reached
	iter := ndm.database.GetPendingLeafKeysIter(txn)
	defer iter.Close()
	for {
		dlkey, dlvalue, isDone, err := iter.Next()
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		if isDone {
			return nil
		}
		nk, err := newNodeKey(dlkey)
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		if ndm.stateLeafDLs.Contains(nk) {
			continue
		}
		if ndm.stateLeafCache.contains(snapShotHeight, nk) {
			continue
		}
		r := &dlReq{
			snapShotHeight: snapShotHeight,
			key:            dlkey,
			value:          dlvalue,
		}
		ndm.stateLeafDLs.Push(nk)
		select {
		case ndm.stateLeafDlChan <- r:
		default:
			ndm.stateLeafDLs.Pop(nk)
			return nil
		}
	}
}

func (ndm *SnapShotManager) syncHdrLeaves(txn *badger.Txn, snapShotHeight uint32) error {
	// get keys from state cache use those to store the state data into the db
	leafKeys := ndm.hdrLeafCache.getLeafKeys(maxNumber)
	// loop through LeafNode keys and retrieve from stateCache;
	// store data before deleting from database.
	for i := 0; i < len(leafKeys); i++ {
		resp, err := ndm.hdrLeafCache.pop(leafKeys[i].key[:])
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			continue
		}
		err = ndm.database.DeletePendingHdrLeafKey(txn, leafKeys[i].key[:])
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		bh := &objs.BlockHeader{}
		err = bh.UnmarshalBinary(resp.data)
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		err = ndm.database.SetCommittedBlockHeaderFastSync(txn, bh)
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
	}
	return nil
}

func (ndm *SnapShotManager) dlHdrLeaves(txn *badger.Txn, snapShotHeight uint32) error {
	// iterate the pending keys in database and start downloads for each pending until the limit is reached
	iter := ndm.database.GetPendingHdrLeafKeysIter(txn)
	defer iter.Close()
	for {
		dlkey, dlvalue, isDone, err := iter.Next()
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		if isDone {
			break
		}
		nk, err := newNodeKey(dlkey)
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		if ndm.hdrLeafDLs.Contains(nk) {
			continue
		}
		if ndm.hdrLeafCache.contains(nk) {
			continue
		}
		r := &dlReq{
			snapShotHeight: snapShotHeight,
			key:            dlkey,
			value:          dlvalue,
		}
		ndm.hdrLeafDLs.Push(nk)
		select {
		case ndm.hdrLeafDlChan <- r:
		default:
			ndm.hdrLeafDLs.Pop(nk)
			break
		}
	}
	return nil
}

func (ndm *SnapShotManager) worker(killChan <-chan struct{}) {
	for {
		select {
		case <-killChan:
			return
		case <-ndm.closeChan:
			return
		case w := <-ndm.workChan:
			w()
		}
	}
}

func (ndm *SnapShotManager) sendWork(w func()) {
	select {
	case <-ndm.closeChan:
		return
	case ndm.workChan <- w:
		return
	}
}

func (ndm *SnapShotManager) downloadWithRetryStateNodeWorker() {
	for {
		select {
		case <-ndm.closeChan:
			return
		case dl := <-ndm.stateNodeDlChan:
			ndm.sendWork(ndm.downloadWithRetryStateNodeClosure(dl))
		}
	}
}

func (ndm *SnapShotManager) downloadWithRetryHdrLeafWorker() {
	for {
		select {
		case <-ndm.closeChan:
			return
		default:
			var stop <-chan time.Time
			cache := []*dlReq{}
			func() {
				for {
					select {
					case <-ndm.closeChan:
						return
					case <-stop:
						return
					case dl := <-ndm.hdrLeafDlChan:
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
			ndm.sendWork(ndm.downloadWithRetryHdrLeafClosure(cache))
		}
	}
}

func (ndm *SnapShotManager) downloadWithRetryHdrNodeWorker() {
	for {
		select {
		case <-ndm.closeChan:
			return
		case dl := <-ndm.hdrNodeDlChan:
			ndm.sendWork(ndm.downloadWithRetryHdrNodeClosure(dl))
		}
	}
}

func (ndm *SnapShotManager) downloadWithRetryStateLeafWorker() {
	for {
		select {
		case <-ndm.closeChan:
			return
		case dl := <-ndm.stateLeafDlChan:
			ndm.sendWork(ndm.downloadWithRetryStateLeafClosure(dl))
		}
	}
}

func (ndm *SnapShotManager) downloadWithRetryStateNodeClosure(dl *dlReq) workFunc {
	snapShotHeight := dl.snapShotHeight
	root := dl.key
	layer := dl.layer
	nk, _ := newNodeKey(root)
	return func() {
		defer ndm.stateNodeDLs.Pop(nk)
		if snapShotHeight < ndm.snapShotHeight.Get() {
			return
		}
		opts := []grpc.CallOption{
			grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
			grpc_retry.WithMax(maxRetryCount),
		}
		resp, err := ndm.requestBus.RequestP2PGetSnapShotNode(context.Background(), snapShotHeight, root, opts...)
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
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
		ndm.stateNodeCache.insert(snapShotHeight, nr)
	}
}

func (ndm *SnapShotManager) downloadWithRetryHdrLeafClosure(dl []*dlReq) workFunc {
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
			defer ndm.hdrLeafDLs.Pop(nk)
		}
		opts := []grpc.CallOption{
			grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
			grpc_retry.WithMax(maxRetryCount),
		}
		peerOpt := middleware.NewPeerInterceptor()
		newOpts := append(opts, peerOpt)
		resp, err := ndm.requestBus.RequestP2PGetBlockHeaders(context.Background(), heightList, newOpts...)
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
				utils.DebugTrace(ndm.logger, err)
				continue
			}
			if !bytes.Equal(bhash, bhashResp) {
				peer.Feedback(-2)
				utils.DebugTrace(ndm.logger, errors.New("Bad block hash"))
				continue
			}
			bhBytes, err := resp[i].MarshalBinary()
			if err != nil {
				peer.Feedback(-2)
				utils.DebugTrace(ndm.logger, err)
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
			ndm.hdrLeafCache.insert(sr)
		}
	}
}

func (ndm *SnapShotManager) downloadWithRetryStateLeafClosure(dl *dlReq) workFunc {
	snapShotHeight := dl.snapShotHeight
	key := dl.key
	value := dl.value
	nk, _ := newNodeKey(key)
	return func() {
		defer ndm.stateLeafDLs.Pop(nk)
		if snapShotHeight < ndm.snapShotHeight.Get() {
			return
		}
		opts := []grpc.CallOption{
			grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
			grpc_retry.WithMax(maxRetryCount),
		}
		peerOpt := middleware.NewPeerInterceptor()
		newOpts := append(opts, peerOpt)
		resp, err := ndm.requestBus.RequestP2PGetSnapShotStateData(context.Background(), key, newOpts...)

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
		ndm.stateLeafCache.insert(snapShotHeight, sr)
	}
}

func (ndm *SnapShotManager) downloadWithRetryHdrNodeClosure(dl *dlReq) workFunc {
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
		defer ndm.hdrNodeDLs.Pop(nk)
		resp, err := ndm.requestBus.RequestP2PGetSnapShotHdrNode(context.Background(), root, newOpts...)
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
		if err := ndm.hdrNodeCache.insert(nr); err != nil {
			utils.DebugTrace(ndm.logger, err)
		}
	}
}
