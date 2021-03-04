package lstate

import (
	"bytes"
	"context"
	"sync"

	"github.com/MadBase/MadNet/consensus/appmock"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/consensus/request"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

const maxNumber int = 100
const maxDLCount int = 16

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
	root  []byte
	layer int
	batch []byte
}

type stateResponse struct {
	key   []byte
	value []byte
	data  []byte
}

type nodeCache struct {
	sync.RWMutex
	objs map[uint32]map[nodeKey]*nodeResponse
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

func (nc *nodeCache) dropBefore(height uint32) {
	nc.Lock()
	defer nc.Unlock()
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
	objs map[uint32]map[nodeKey]*stateResponse
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

func (sc *stateCache) insert(height uint32, sr *stateResponse) error {
	sc.Lock()
	defer sc.Unlock()
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

type SnapShotManager struct {
	appHandler appmock.Application
	requestBus *request.Client

	currentCtx               context.Context
	currentCtxCancel         func()
	currentWg                *sync.WaitGroup
	database                 db.DatabaseIface
	hcache                   *nodeCache
	ncache                   *nodeCache
	nscache                  *stateCache
	hscache                  *stateCache
	currentDLs               map[nodeKey]bool
	currentHeight            uint32
	fullCanonicalBlockHeader *objs.BlockHeader
	logger                   *logrus.Logger
}

// Init initializes the SnapShotManager
func (ndm *SnapShotManager) Init(database db.DatabaseIface) error {
	ndm.logger = logging.GetLogger(constants.LoggerConsensus)
	ctx := context.Background()
	subCtx, cf := context.WithCancel(ctx)
	ndm.currentCtx = subCtx
	ndm.currentCtxCancel = cf
	ndm.currentWg = &sync.WaitGroup{}
	ndm.hcache = &nodeCache{}
	if err := ndm.hcache.Init(); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	ndm.hscache = &stateCache{}
	if err := ndm.hscache.Init(); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	ndm.ncache = &nodeCache{}
	if err := ndm.ncache.Init(); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	ndm.nscache = &stateCache{}
	if err := ndm.nscache.Init(); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	ndm.currentDLs = make(map[nodeKey]bool)
	ndm.database = database
	return nil
}

func (ndm *SnapShotManager) startFastSync(txn *badger.Txn, height uint32, stateRoot []byte, hdrRoot []byte, canonicalBlockHash []byte) error {
	// stop all previous downloads
	ndm.currentCtxCancel()
	ndm.currentWg.Wait()
	ctx := context.Background()
	subCtx, cf := context.WithCancel(ctx)
	ndm.currentCtx = subCtx
	ndm.currentCtxCancel = cf
	ndm.currentDLs = make(map[nodeKey]bool)
	// cleanup the db of any previous data
	if err := ndm.cleanupDatabase(txn); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	// call dropBefore on the caches
	ndm.hcache.dropBefore(height)
	ndm.ncache.dropBefore(height)
	ndm.nscache.dropBefore(height)
	ndm.hscache.dropBefore(height)
	// set the currentHeight
	ndm.currentHeight = height
	// insert the request for the root node into the database
	if !bytes.Equal(stateRoot, make([]byte, constants.HashLen)) {
		// Do NOT request all-zero byte slice stateRoot
		if err := ndm.database.SetPendingNodeKey(txn, stateRoot, 0); err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
	}
	if err := ndm.database.SetPendingHdrNodeKey(txn, hdrRoot, 0); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	canonicalBHTrieKey := ndm.database.MakeHeaderTrieKeyFromHeight(height)
	if err := ndm.database.SetPendingHdrLeafKey(txn, canonicalBHTrieKey, canonicalBlockHash); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	return nil
}

func (ndm *SnapShotManager) cleanupDatabase(txn *badger.Txn) error {
	ndm.appHandler.BeginSnapShotSync(txn)
	if err := ndm.database.DropPendingLeafKeys(); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.database.DropPendingHdrLeafKeys(); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.database.DropPendingHdrNodeKeys(); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.database.DropPendingNodeKeys(); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	return nil
}

func (ndm *SnapShotManager) Update(txn *badger.Txn, snapShotHeight uint32, syncToHeight uint32, stateRoot []byte, hdrRoot []byte, canonicalBlockHash []byte) (bool, error) {
	// a difference in height implies the target has changed for the canonical
	// state, thus re-init the object and drop all stale data
	// return after the drop so the next iteration sees the dropped data is in the
	// db transaction
	if ndm.currentHeight != snapShotHeight {
		err := ndm.startFastSync(txn, snapShotHeight, stateRoot, hdrRoot, canonicalBlockHash)
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return false, err
		}
		return false, nil
	}
	if err := ndm.updateSync(txn, snapShotHeight, syncToHeight, stateRoot, hdrRoot); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return false, err
	}
	nhcount, nkcount, nlcount, hlcount, sbhcount, err := ndm.getKeyCounts(txn)
	if err != nil {
		utils.DebugTrace(ndm.logger, err)
		return false, err
	}
	ndm.logger.Debugf("FastSync: %v %v %v %v %v", nhcount, hlcount, sbhcount, nkcount, nlcount)
	if err := ndm.updateDls(txn, snapShotHeight, stateRoot, hdrRoot, nhcount, nkcount, nlcount); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return false, err
	}
	if len(ndm.currentDLs) == 0 && nkcount == 0 && nlcount == 0 && nhcount == 0 && hlcount == 0 && sbhcount == 0 {
		if err := ndm.finalizeSync(txn, snapShotHeight, stateRoot, hdrRoot); err != nil {
			utils.DebugTrace(ndm.logger, err)
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (ndm *SnapShotManager) getKeyCounts(txn *badger.Txn) (int, int, int, int, int, error) {
	nhcount, err := ndm.database.CountPendingHdrNodeKeys(txn)
	if err != nil {
		utils.DebugTrace(ndm.logger, err)
		return 0, 0, 0, 0, 0, err
	}
	nkcount, err := ndm.database.CountPendingNodeKeys(txn)
	if err != nil {
		utils.DebugTrace(ndm.logger, err)
		return 0, 0, 0, 0, 0, err
	}
	nlcount, err := ndm.database.CountPendingLeafKeys(txn)
	if err != nil {
		utils.DebugTrace(ndm.logger, err)
		return 0, 0, 0, 0, 0, err
	}
	hlcount, err := ndm.database.CountPendingHdrLeafKeys(txn)
	if err != nil {
		utils.DebugTrace(ndm.logger, err)
		return 0, 0, 0, 0, 0, err
	}
	sbhcount, err := ndm.database.CountStagedBlockHeaderKeys(txn)
	if err != nil {
		utils.DebugTrace(ndm.logger, err)
		return 0, 0, 0, 0, 0, err
	}
	return nhcount, nkcount, nlcount, hlcount, sbhcount, nil
}

func (ndm *SnapShotManager) finalizeSync(txn *badger.Txn, height uint32, stateRoot []byte, hdrRoot []byte) error {
	if err := ndm.appHandler.FinalizeSnapShotRoot(txn, stateRoot, height); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	return nil
}

func (ndm *SnapShotManager) updateDls(txn *badger.Txn, height uint32, stateRoot []byte, hdrRoot []byte, nhcount int, nkcount int, nlcount int) error {
	if err := ndm.dlHdrNodes(txn, height, stateRoot, hdrRoot); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if nhcount == 0 {
		if err := ndm.dlHdrLeaves(txn, height, stateRoot, hdrRoot); err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
	}
	if err := ndm.dlStateNodes(txn, height, stateRoot, hdrRoot); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if nkcount == 0 {
		if err := ndm.dlStateLeaves(txn, height, stateRoot, hdrRoot); err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
	}
	return nil
}

func (ndm *SnapShotManager) updateSync(txn *badger.Txn, snapShotHeight uint32, syncToHeight uint32, stateRoot []byte, hdrRoot []byte) error {
	if err := ndm.syncHdrNodes(txn, snapShotHeight, syncToHeight, stateRoot, hdrRoot); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.syncHdrLeaves(txn, snapShotHeight, stateRoot, hdrRoot); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.syncStateNodes(txn, snapShotHeight, stateRoot, hdrRoot); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.syncStateLeaves(txn, snapShotHeight, stateRoot, hdrRoot); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	if err := ndm.syncStagedBlockHeaders(txn, snapShotHeight, syncToHeight, stateRoot, hdrRoot); err != nil {
		utils.DebugTrace(ndm.logger, err)
		return err
	}
	return nil
}

func (ndm *SnapShotManager) syncHdrNodes(txn *badger.Txn, snapShotHeight uint32, syncToHeight uint32, stateRoot []byte, hdrRoot []byte) error {
	// get a set of node header keys and sync the header trie based on those
	// node keys
	nodeHdrKeys := ndm.hcache.getNodeKeys(snapShotHeight, maxNumber)
	for _, k := range nodeHdrKeys {
		resp, err := ndm.hcache.pop(snapShotHeight, k.key[:])
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		delete(ndm.currentDLs, k)
		// store all of those nodes into the database and get new pending keys
		pendingBatch, newLayer, lvs, err := ndm.database.SetSnapShotHdrNode(txn, resp.batch, resp.root, resp.layer)
		if err != nil {
			// should not return if err invalid
			utils.DebugTrace(ndm.logger, err)
			continue
		}
		for i := 0; i < len(lvs); i++ {
			bhHeight, _ := utils.UnmarshalUint32(lvs[i].Key[0:4])
			if bhHeight <= syncToHeight {
				continue
			}
			if err := ndm.database.SetPendingHdrLeafKey(txn, lvs[i].Key, lvs[i].Value); err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
		}
		// remove the keys from the pending set in the database
		err = ndm.database.DeletePendingHdrNodeKey(txn, k.key[:])
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		// store new pending keys to db
		for _, kk := range pendingBatch {
			ok, err := ndm.database.ContainsSnapShotHdrNode(txn, kk)
			if err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
			if ok {
				continue
			}
			if err := ndm.database.SetPendingHdrNodeKey(txn, kk, newLayer); err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
		}
	}
	return nil
}

func (ndm *SnapShotManager) dlHdrNodes(txn *badger.Txn, height uint32, stateRoot []byte, hdrRoot []byte) error {
	if len(ndm.currentDLs) >= maxDLCount {
		return nil
	}
	// iterate the pending keys in database and start downloads for each pending until the limit is reached
	iter := ndm.database.GetPendingHdrNodeKeysIter(txn)
	defer iter.Close()
	for {
		dlroot, dllayer, isDone, err := iter.Next()
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		if isDone {
			break
		}
		nk, err := newNodeKey(dlroot)
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		if ndm.currentDLs[nk] {
			continue
		}
		ndm.currentWg.Add(1)
		ndm.currentDLs[nk] = true
		go ndm.downloadWithRetryHdrNode(ndm.currentCtx, ndm.currentWg, height, dllayer, dlroot)
		if len(ndm.currentDLs) >= maxDLCount {
			break
		}
	}
	return nil
}

func (ndm *SnapShotManager) downloadWithRetryHdrNode(ctx context.Context, wg *sync.WaitGroup, height uint32, layer int, root []byte) {
	defer ndm.currentWg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// continue doing retries
		}
		var resp []byte
		err := func() error {
			subCtx, cf := context.WithTimeout(ctx, constants.MsgTimeout)
			defer cf()
			tmp, err := ndm.requestBus.RequestP2PGetSnapShotHdrNode(subCtx, root)
			if err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
			resp = tmp
			return nil
		}()
		if err != nil {
			continue
		}
		if len(resp) == 0 {
			continue
		}
		nr := &nodeResponse{
			layer: layer,
			root:  root,
			batch: resp,
		}
		//    store to the cache
		ndm.hcache.insert(height, nr)
		return
	}
}

func (ndm *SnapShotManager) syncStateNodes(txn *badger.Txn, height uint32, stateRoot []byte, hdrRoot []byte) error {
	// get keys from node cache and use those keys to store elements from the
	// node cache into the database as well as to get the leaf keys and store
	// those into the database as well
	nodeKeys := ndm.ncache.getNodeKeys(height, maxNumber)
	//for each key
	for _, k := range nodeKeys {
		resp, err := ndm.ncache.pop(height, k.key[:])
		if err != nil {
			return err
		}
		delete(ndm.currentDLs, k)
		// store all of those nodes into the database and get new pending keys
		pendingBatch, newLayer, lvs, err := ndm.appHandler.StoreSnapShotNode(txn, resp.batch, resp.root, resp.layer)
		if err != nil {
			// should not return if err invalid
			utils.DebugTrace(ndm.logger, err)
			continue
		}
		for i := 0; i < len(lvs); i++ {
			if _, err := ndm.appHandler.GetSnapShotStateData(txn, utils.CopySlice(lvs[i].Key)); err != nil {
				if err := ndm.database.SetPendingLeafKey(txn, lvs[i].Key, lvs[i].Value); err != nil {
					utils.DebugTrace(ndm.logger, err)
					return err
				}
			}
		}
		// remove the keys from the pending set in the database
		err = ndm.database.DeletePendingNodeKey(txn, k.key[:])
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		// store new pending keys to db
		for _, kk := range pendingBatch {
			if err := ndm.database.SetPendingNodeKey(txn, kk, newLayer); err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
		}
	}
	return nil
}

func (ndm *SnapShotManager) dlStateNodes(txn *badger.Txn, height uint32, stateRoot []byte, hdrRoot []byte) error {
	if len(ndm.currentDLs) >= maxDLCount {
		return nil
	}
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
			break
		}
		nk, err := newNodeKey(dlroot)
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		if ndm.currentDLs[nk] {
			continue
		}
		ndm.currentWg.Add(1)
		ndm.currentDLs[nk] = true
		go ndm.downloadWithRetryStateNode(ndm.currentCtx, ndm.currentWg, height, dllayer, dlroot)
		if len(ndm.currentDLs) >= maxDLCount {
			break
		}
	}
	return nil
}

func (ndm *SnapShotManager) downloadWithRetryStateNode(ctx context.Context, wg *sync.WaitGroup, height uint32, layer int, root []byte) {
	defer ndm.currentWg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// continue doing retries
		}
		var resp []byte
		err := func() error {
			subCtx, cf := context.WithTimeout(ctx, constants.MsgTimeout)
			defer cf()
			tmp, err := ndm.requestBus.RequestP2PGetSnapShotNode(subCtx, height, root)
			if err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
			resp = tmp
			return nil
		}()
		if err != nil {
			continue
		}
		if len(resp) == 0 {
			continue
		}
		nr := &nodeResponse{
			layer: layer,
			root:  root,
			batch: resp,
		}
		//    store to the cache
		ndm.ncache.insert(height, nr)
		return
	}
}

func (ndm *SnapShotManager) syncStateLeaves(txn *badger.Txn, height uint32, stateRoot []byte, hdrRoot []byte) error {
	// get keys from state cache use those to store the state data into the db
	leafKeys := ndm.nscache.getLeafKeys(height, maxNumber)
	// loop through LeafNode keys and retrieve from stateCache;
	// store data before deleting from database.
	for _, k := range leafKeys {
		resp, err := ndm.nscache.pop(height, k.key[:])
		if err != nil {
			return err
		}
		delete(ndm.currentDLs, k)
		// store key-value state data
		err = ndm.appHandler.StoreSnapShotStateData(txn, resp.key, resp.value, resp.data)
		if err != nil {
			// should not return if err invalid
			utils.DebugTrace(ndm.logger, err)
			continue
		}
		// remove the keys from the pending set in the database
		err = ndm.database.DeletePendingLeafKey(txn, k.key[:])
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
	}
	return nil
}

func (ndm *SnapShotManager) dlStateLeaves(txn *badger.Txn, height uint32, stateRoot []byte, hdrRoot []byte) error {
	if len(ndm.currentDLs) >= maxDLCount {
		return nil
	}
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
			break
		}
		nk, err := newNodeKey(dlkey)
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		if ndm.currentDLs[nk] {
			continue
		}
		ndm.currentWg.Add(1)
		ndm.currentDLs[nk] = true
		go ndm.downloadWithRetryStateLeaf(ndm.currentCtx, ndm.currentWg, height, dlkey, dlvalue)
		if len(ndm.currentDLs) >= maxDLCount {
			break
		}
	}
	return nil
}

func (ndm *SnapShotManager) downloadWithRetryStateLeaf(ctx context.Context, wg *sync.WaitGroup, height uint32, key []byte, value []byte) {
	defer ndm.currentWg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// continue doing retries
		}
		var resp []byte
		err := func() error {
			subCtx, cf := context.WithTimeout(ctx, constants.MsgTimeout)
			defer cf()
			tmp, err := ndm.requestBus.RequestP2PGetSnapShotStateData(subCtx, key)
			if err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
			resp = tmp
			return nil
		}()
		if err != nil {
			continue
		}
		if len(resp) == 0 {
			continue
		}
		sr := &stateResponse{
			key:   utils.CopySlice(key),
			value: utils.CopySlice(value),
			data:  utils.CopySlice(resp),
		}
		//    store to the cache
		ndm.nscache.insert(height, sr)
		return
	}
}

func (ndm *SnapShotManager) syncHdrLeaves(txn *badger.Txn, height uint32, stateRoot []byte, hdrRoot []byte) error {
	// get keys from state cache use those to store the state data into the db
	leafKeys := ndm.hscache.getLeafKeys(height, maxNumber)
	// loop through LeafNode keys and retrieve from stateCache;
	// store data before deleting from database.
	for _, k := range leafKeys {
		resp, err := ndm.hscache.pop(height, k.key[:])
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
		delete(ndm.currentDLs, k)
		// store key-value state data
		blockHeight, _ := utils.UnmarshalUint32(resp.key[0:4])
		err = ndm.database.SetStagedBlockHeader(txn, blockHeight, resp.data)
		if err != nil {
			// should not return if err invalid
			utils.DebugTrace(ndm.logger, err)
			continue
		}
		// remove the keys from the pending set in the database
		err = ndm.database.DeletePendingHdrLeafKey(txn, k.key[:])
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
	}
	return nil
}

func (ndm *SnapShotManager) dlHdrLeaves(txn *badger.Txn, height uint32, stateRoot []byte, hdrRoot []byte) error {
	if len(ndm.currentDLs) >= maxDLCount {
		return nil
	}
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
		if ndm.currentDLs[nk] {
			continue
		}
		ndm.currentWg.Add(1)
		ndm.currentDLs[nk] = true
		go ndm.downloadWithRetryHdrLeaf(ndm.currentCtx, ndm.currentWg, height, dlkey, dlvalue)
		if len(ndm.currentDLs) >= maxDLCount {
			break
		}
	}
	return nil
}

func (ndm *SnapShotManager) downloadWithRetryHdrLeaf(ctx context.Context, wg *sync.WaitGroup, height uint32, key []byte, value []byte) {
	defer ndm.currentWg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// continue doing retries
		}
		blockHeight, _ := utils.UnmarshalUint32(key[0:4])
		var resp []*objs.BlockHeader
		err := func() error {
			subCtx, cf := context.WithTimeout(ctx, constants.MsgTimeout)
			defer cf()
			tmp, err := ndm.requestBus.RequestP2PGetBlockHeaders(subCtx, []uint32{blockHeight})
			if err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
			resp = tmp
			return nil
		}()
		if err != nil {
			continue
		}
		if len(resp) != 1 {
			continue
		}
		if resp[0].BClaims.Height != blockHeight {
			continue
		}
		bhash, err := resp[0].BlockHash()
		if err != nil {
			continue
		}
		if !bytes.Equal(bhash, value) {
			continue
		}
		bhBytes, err := resp[0].MarshalBinary()
		if err != nil {
			continue
		}
		sr := &stateResponse{
			key:   utils.CopySlice(key),
			value: utils.CopySlice(value),
			data:  bhBytes,
		}
		//    store to the cache
		ndm.hscache.insert(height, sr)
		return
	}
}

func (ndm *SnapShotManager) syncStagedBlockHeaders(txn *badger.Txn, snapShotHeight uint32, syncToHeight uint32, stateRoot []byte, hdrRoot []byte) error {
	bhdrs := []*objs.BlockHeader{}
	dropHeights := []uint32{}
	err := func() error {
		iter := ndm.database.GetStagedBlockHeaderKeyIter(txn)
		defer iter.Close()
		for {
			bhHeight, bh, isDone, err := iter.Next()
			if err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
			if isDone {
				break
			}
			if bhHeight <= syncToHeight {
				dropHeights = append(dropHeights, bhHeight)
				if len(bhdrs)+len(dropHeights) >= maxNumber {
					break
				}
				continue
			}
			if len(bhdrs) == 0 {
				if bhHeight == syncToHeight+1 {
					bhdrs = append(bhdrs, bh)
				} else {
					break
				}
			} else {
				if bhHeight == bhdrs[len(bhdrs)-1].BClaims.Height+1 {
					bhdrs = append(bhdrs, bh)
				} else {
					break
				}
			}
			if len(bhdrs)+len(dropHeights) >= maxNumber {
				break
			}
		}
		return nil
	}()
	if err != nil {
		return err
	}
	for i := 0; i < len(dropHeights); i++ {
		err = ndm.database.DeleteStagedBlockHeaderKey(txn, dropHeights[i])
		if err != nil {
			utils.DebugTrace(ndm.logger, err)
			return err
		}
	}
	for i := 0; i < len(bhdrs); i++ {
		if bhdrs[i].BClaims.Height < snapShotHeight {
			err := ndm.database.SetCommittedBlockHeaderFastSync(txn, bhdrs[i])
			if err != nil {
				// should not return if err invalid
				utils.DebugTrace(ndm.logger, err)
				return err
			}
			// remove the keys from the pending set in the database
			err = ndm.database.DeleteStagedBlockHeaderKey(txn, bhdrs[i].BClaims.Height)
			if err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
		}
		if bhdrs[i].BClaims.Height == snapShotHeight && len(bhdrs) == 1 {
			err := ndm.database.SetCommittedBlockHeaderFastSync(txn, bhdrs[i])
			if err != nil {
				// should not return if err invalid
				utils.DebugTrace(ndm.logger, err)
				return err
			}
			// remove the keys from the pending set in the database
			err = ndm.database.DeleteStagedBlockHeaderKey(txn, bhdrs[i].BClaims.Height)
			if err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
			if err := ndm.database.UpdateHeaderTrieRootFastSync(txn, bhdrs[i]); err != nil {
				utils.DebugTrace(ndm.logger, err)
				return err
			}
		}
	}
	return nil
}
