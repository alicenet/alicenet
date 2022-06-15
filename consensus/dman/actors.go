package dman

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/middleware"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const backoffAmount = 1 // 1 ms backoff per
const retryMax = 6      // equates to approx 4 seconds

// RootActor spawns top level actor types for download manager.
// This system allows the synchronously run consensus algorithm to request the
// download of tx state and blocks from remote peers as a background task.
// The system keeps a record of all pending downloads to prevent double entry
// and stores all state requested into a hot cache that is flushed to disk
// by the synchronous code. This system will retry failed requests until
// the height lag is raised to a point that invalidates the given
// request.
type RootActor struct {
	sync.Mutex
	ba        *blockActor
	wg        *sync.WaitGroup
	closeChan chan struct{}
	closeOnce sync.Once
	dispatchQ chan DownloadRequest
	database  databaseView
	txc       *txCache
	bhc       *bHCache
	logger    *logrus.Logger
	reqs      map[string]bool
}

func (a *RootActor) Init(logger *logrus.Logger, proxy typeProxyIface) error {
	a.txc = &txCache{}
	err := a.txc.Init(proxy)
	if err != nil {
		utils.DebugTrace(logger, err)
		return err
	}
	a.bhc = &bHCache{}
	err = a.bhc.Init()
	if err != nil {
		utils.DebugTrace(logger, err)
		return err
	}
	a.closeOnce = sync.Once{}
	a.reqs = make(map[string]bool)
	a.wg = new(sync.WaitGroup)
	a.closeChan = make(chan struct{})
	a.dispatchQ = make(chan DownloadRequest)
	a.logger = logger
	a.ba = &blockActor{}
	a.database = proxy
	return a.ba.init(a.dispatchQ, logger, proxy, a.closeChan)
}

func (a *RootActor) Start() {
	go a.ba.start()
}

func (a *RootActor) Close() {
	a.closeOnce.Do(func() {
		close(a.closeChan)
		a.wg.Wait()
	})
}

// TODO verify blockheader cache is being cleaned
func (a *RootActor) FlushCacheToDisk(txn *badger.Txn, height uint32) error {
	txList, txHashList := a.txc.GetHeight(height + 1)
	for i := 0; i < len(txList); i++ {
		txb, err := txList[i].MarshalBinary()
		if err != nil {
			return err
		}
		if err := a.database.SetTxCacheItem(txn, height, []byte(txHashList[i]), txb); err != nil {
			return err
		}
	}
	return nil
}

// CleanCache flushes all items older than 5 blocks from cache
func (a *RootActor) CleanCache(txn *badger.Txn, height uint32) error {
	if height > 10 {
		dropKeys := a.bhc.DropBeforeHeight(height - 5)
		dropKeys = append(dropKeys, a.txc.DropBeforeHeight(height-5)...)
		for i := 0; i < len(dropKeys); i++ {
			delete(a.reqs, dropKeys[i])
		}
	}
	return nil
}

// DownloadPendingTx downloads txs that are pending from remote peers
func (a *RootActor) DownloadPendingTx(height, round uint32, txHash []byte) {
	req := NewTxDownloadRequest(txHash, PendingTxRequest, height, round)
	a.download(req, false)
}

// DownloadPendingTx downloads txs that are mined from remote peers
func (a *RootActor) DownloadMinedTx(height, round uint32, txHash []byte) {
	req := NewTxDownloadRequest(txHash, MinedTxRequest, height, round)
	a.download(req, false)
}

func (a *RootActor) DownloadTx(height, round uint32, txHash []byte) {
	req := NewTxDownloadRequest(txHash, PendingAndMinedTxRequest, height, round)
	a.download(req, false)
}

// DownloadBlockHeader downloads block headers from remote peers
func (a *RootActor) DownloadBlockHeader(height, round uint32) {
	req := NewBlockHeaderDownloadRequest(height, round, BlockHeaderRequest)
	a.download(req, false)
}

func (a *RootActor) download(b DownloadRequest, retry bool) {
	select {
	case <-a.closeChan:
		return
	default:
		a.wg.Add(1)
		go a.doDownload(b, retry)
	}
}

func (a *RootActor) doDownload(b DownloadRequest, retry bool) {
	defer a.wg.Done()
	switch b.DownloadType() {
	case PendingTxRequest, MinedTxRequest:
		ok := func() bool {
			a.Lock()
			defer a.Unlock()
			if a.txc.Contains([]byte(b.Identifier())) {
				return false
			}
			if _, exists := a.reqs[b.Identifier()]; exists {
				return retry
			}
			a.reqs[b.Identifier()] = true
			return true
		}()
		if !ok {
			return
		}
		select {
		case <-a.closeChan:
			return
		case a.dispatchQ <- b:
			a.await(b)
		}
	case PendingAndMinedTxRequest:
		ok := func() bool {
			a.Lock()
			defer a.Unlock()
			if a.txc.Contains([]byte(b.Identifier())) {
				return false
			}
			if _, exists := a.reqs[b.Identifier()]; exists {
				return false
			}
			a.reqs[b.Identifier()] = true
			return true
		}()
		if !ok {
			return
		}
		bc0 := &TxDownloadRequest{
			TxHash:       []byte(b.Identifier()),
			Dtype:        PendingTxRequest,
			Height:       b.RequestHeight(),
			Round:        b.RequestRound(),
			responseChan: make(chan DownloadResponse, 1),
		}
		bc1 := &TxDownloadRequest{
			TxHash:       []byte(b.Identifier()),
			Dtype:        MinedTxRequest,
			Height:       b.RequestHeight(),
			Round:        b.RequestRound(),
			responseChan: make(chan DownloadResponse, 1),
		}
		ptxc := make(chan struct{})
		mtxc := make(chan struct{})
		go func() {
			defer close(ptxc)
			select {
			case <-a.closeChan:
				return
			case a.dispatchQ <- bc0:
				a.await(bc0)
			}
		}()
		go func() {
			defer close(mtxc)
			select {
			case <-a.closeChan:
				return
			case a.dispatchQ <- bc1:
				a.await(bc1)
			}
		}()
		select {
		case <-a.closeChan:
			return
		case <-ptxc:
			select {
			case <-a.closeChan:
				return
			case <-mtxc:
			}
		case <-mtxc:
			select {
			case <-a.closeChan:
				return
			case <-ptxc:
			}
		}
	case BlockHeaderRequest:
		ok := func() bool {
			a.Lock()
			defer a.Unlock()
			if a.bhc.Contains(b.RequestHeight()) {
				return false
			}
			if _, exists := a.reqs[b.Identifier()]; exists {
				return retry
			}
			a.reqs[b.Identifier()] = true
			return true
		}()
		if !ok {
			return
		}
		select {
		case <-a.closeChan:
			return
		case a.dispatchQ <- b:
			a.await(b)
		}
	default:
		panic(b.DownloadType())
	}
}

func (a *RootActor) await(req DownloadRequest) {
	select {
	case <-a.closeChan:
		return
	case resp := <-req.ResponseChan():
		if resp == nil {
			return
		}
		switch resp.DownloadType() {
		case PendingTxRequest, MinedTxRequest:
			r := resp.(*TxDownloadResponse)
			if r.Err != nil {
				exists := a.txc.Contains(r.TxHash)
				if !exists {
					utils.DebugTrace(a.logger, r.Err)
					defer a.download(req, true)
				}
				return
			}
			ok := func() bool {
				if err := a.txc.Add(resp.RequestHeight(), r.Tx); err != nil {
					utils.DebugTrace(a.logger, err)
					return false
				}
				return true
			}()
			if ok {
				return
			}
			defer a.download(req, true)
		case BlockHeaderRequest:
			r := resp.(*BlockHeaderDownloadResponse)
			if r.Err != nil {
				utils.DebugTrace(a.logger, r.Err)
				defer a.download(req, true)
				return
			}
			ok := func() bool {
				if err := a.bhc.Add(r.BH); err != nil {
					utils.DebugTrace(a.logger, err)
					return false
				}
				return true
			}()
			if ok {
				return
			}
			defer a.download(req, true)
		default:
			panic(req.DownloadType())
		}
	}
}

// block actor does height based filter drop on requests
type blockActor struct {
	sync.RWMutex
	ra            *downloadActor
	WorkQ         chan DownloadRequest
	dispatchQ     chan DownloadRequest
	CurrentHeight uint32
	Logger        *logrus.Logger
	closeChan     chan struct{}
}

func (a *blockActor) init(workQ chan DownloadRequest, logger *logrus.Logger, reqBus typeProxyIface, closeChan chan struct{}) error {
	a.WorkQ = workQ
	a.dispatchQ = make(chan DownloadRequest)
	a.ra = &downloadActor{}
	a.Logger = logger
	a.closeChan = closeChan
	return a.ra.init(a.dispatchQ, logger, reqBus, a.closeChan)
}

func (a *blockActor) start() {
	go a.run()
	a.ra.start()
}

func (a *blockActor) updateHeight(newHeight uint32) {
	a.Lock()
	defer a.Unlock()
	if newHeight > a.CurrentHeight {
		a.CurrentHeight = newHeight
	}
}

func (a *blockActor) getHeight() uint32 {
	a.RLock()
	defer a.RUnlock()
	return a.CurrentHeight
}

func (a *blockActor) run() {
	for {
		select {
		case <-a.closeChan:
			return
		case req := <-a.WorkQ:
			ok := func() bool {
				a.Lock()
				defer a.Unlock()
				if req.RequestHeight()+heightDropLag < a.CurrentHeight {
					close(req.ResponseChan())
					return false
				}
				return true
			}()
			if !ok {
				continue
			}
			go a.await(req)
		}
	}
}

func (a *blockActor) await(req DownloadRequest) {
	var subReq DownloadRequest
	switch req.DownloadType() {
	case PendingTxRequest, MinedTxRequest:
		reqTyped := req.(*TxDownloadRequest)
		subReq = NewTxDownloadRequest(reqTyped.TxHash, reqTyped.Dtype, reqTyped.Height, reqTyped.Round)
		select {
		case <-a.closeChan:
			return
		case a.dispatchQ <- subReq:
		}
	case BlockHeaderRequest:
		reqTyped := req.(*BlockHeaderDownloadRequest)
		subReq = NewBlockHeaderDownloadRequest(reqTyped.Height, reqTyped.Round, reqTyped.Dtype)
		select {
		case <-a.closeChan:
			return
		case a.dispatchQ <- subReq:
		}
	default:
		panic(fmt.Sprintf("req download type not found: %v", req.DownloadType()))
	}
	select {
	case <-a.closeChan:
		return
	case resp := <-subReq.ResponseChan():
		if resp == nil {
			close(req.ResponseChan())
			return
		}
		ok := func() bool {
			a.RLock()
			defer a.RUnlock()
			if a.CurrentHeight > heightDropLag+1 {
				return resp.RequestHeight() >= a.CurrentHeight-heightDropLag
			}
			return true
		}()
		if !ok {
			close(req.ResponseChan())
			return
		}
		select {
		case <-a.closeChan:
			return
		case req.ResponseChan() <- resp:
		}
	}
}

// download actor does download dispatch based on request type to worker pools
type downloadActor struct {
	WorkQ            chan DownloadRequest
	PendingDispatchQ chan *TxDownloadRequest
	MinedDispatchQ   chan *TxDownloadRequest
	BlockDispatchQ   chan *BlockHeaderDownloadRequest
	ptxa             *pendingDownloadActor
	mtxa             *minedDownloadActor
	bha              *blockHeaderDownloadActor
	Logger           *logrus.Logger
	closeChan        chan struct{}
}

func (a *downloadActor) init(workQ chan DownloadRequest, logger *logrus.Logger, reqBus typeProxyIface, closeChan chan struct{}) error {
	a.PendingDispatchQ = make(chan *TxDownloadRequest, downloadWorkerCountPendingTx)
	a.MinedDispatchQ = make(chan *TxDownloadRequest, downloadWorkerCountMinedTx)
	a.BlockDispatchQ = make(chan *BlockHeaderDownloadRequest, downloadWorkerCountBlockHeader)
	a.WorkQ = workQ
	a.ptxa = &pendingDownloadActor{}
	a.Logger = logger
	a.closeChan = closeChan
	if err := a.ptxa.init(a.PendingDispatchQ, logger, reqBus, closeChan); err != nil {
		return err
	}
	a.mtxa = &minedDownloadActor{}
	if err := a.mtxa.init(a.MinedDispatchQ, logger, reqBus, closeChan); err != nil {
		return err
	}
	a.bha = &blockHeaderDownloadActor{}
	if err := a.bha.init(a.BlockDispatchQ, logger, reqBus, closeChan); err != nil {
		return err
	}
	return nil
}

func (a *downloadActor) start() {
	go a.run()
	a.ptxa.start()
	a.mtxa.start()
	a.bha.start()
}

func (a *downloadActor) run() {
	for {
		select {
		case <-a.closeChan:
			return
		case req := <-a.WorkQ:
			switch req.DownloadType() {
			case PendingTxRequest:
				select {
				case a.PendingDispatchQ <- req.(*TxDownloadRequest):
				default:
					a.ptxa.start()
					a.PendingDispatchQ <- req.(*TxDownloadRequest)
				}
			case MinedTxRequest:
				select {
				case a.MinedDispatchQ <- req.(*TxDownloadRequest):
				default:
					a.mtxa.start()
					a.MinedDispatchQ <- req.(*TxDownloadRequest)
				}
			case BlockHeaderRequest:
				select {
				case a.BlockDispatchQ <- req.(*BlockHeaderDownloadRequest):
				default:
					a.bha.start()
					a.BlockDispatchQ <- req.(*BlockHeaderDownloadRequest)
				}
			default:
				panic(req.DownloadType())
			}
		}
	}
}

type minedDownloadActor struct {
	sync.Mutex
	numWorkers int
	WorkQ      chan *TxDownloadRequest
	reqBus     typeProxyIface
	Logger     *logrus.Logger
	closeChan  chan struct{}
}

func (a *minedDownloadActor) init(workQ chan *TxDownloadRequest, logger *logrus.Logger, reqBus typeProxyIface, closeChan chan struct{}) error {
	a.WorkQ = workQ
	a.reqBus = reqBus
	a.Logger = logger
	a.closeChan = closeChan
	return nil
}

func (a *minedDownloadActor) start() {
	a.Lock()
	defer a.Unlock()
	for i := 0; i < 2; i++ {
		if a.numWorkers < downloadWorkerCountMinedTx {
			a.numWorkers++
			go a.run()
		}
	}
}

func (a *minedDownloadActor) run() {
	for {
		select {
		case <-time.After(10 * time.Second):
			a.Lock()
			if a.numWorkers > 1 {
				a.numWorkers--
				a.Unlock()
				return
			}
			a.Unlock()
		case <-a.closeChan:
			return
		case reqOrig := <-a.WorkQ:
			tx, err := func(req *TxDownloadRequest) (interfaces.Transaction, error) {
				opts := []grpc.CallOption{
					grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backoffAmount*time.Millisecond, .1)),
					grpc_retry.WithMax(retryMax),
				}
				peerOpt := middleware.NewPeerInterceptor()
				newOpts := append(opts, peerOpt)
				txLst, err := a.reqBus.RequestP2PGetMinedTxs(context.Background(), [][]byte{req.TxHash}, newOpts...)
				if err != nil {
					utils.DebugTrace(a.Logger, err)
					return nil, errorz.ErrInvalid{}.New(err.Error())
				}
				peer := peerOpt.Peer()
				if len(txLst) != 1 {
					peer.Feedback(-3)
					return nil, errorz.ErrInvalid{}.New("Downloaded more than 1 txn when only should have 1")
				}
				tx, err := a.reqBus.UnmarshalTx(utils.CopySlice(txLst[0]))
				if err != nil {
					peer.Feedback(-2)
					utils.DebugTrace(a.Logger, err)
					return nil, errorz.ErrInvalid{}.New(err.Error())
				}
				return tx, nil
			}(reqOrig)
			reqOrig.ResponseChan() <- NewTxDownloadResponse(reqOrig, tx, MinedTxRequest, err)
		}
	}
}

type pendingDownloadActor struct {
	sync.Mutex
	numWorkers int
	WorkQ      chan *TxDownloadRequest
	reqBus     typeProxyIface
	Logger     *logrus.Logger
	closeChan  chan struct{}
}

func (a *pendingDownloadActor) init(workQ chan *TxDownloadRequest, logger *logrus.Logger, reqBus typeProxyIface, closeChan chan struct{}) error {
	a.WorkQ = workQ
	a.reqBus = reqBus
	a.Logger = logger
	a.closeChan = closeChan
	return nil
}

func (a *pendingDownloadActor) start() {
	a.Lock()
	defer a.Unlock()
	for i := 0; i < 2; i++ {
		if a.numWorkers < downloadWorkerCountPendingTx {
			a.numWorkers++
			go a.run()
		}
	}
}

func (a *pendingDownloadActor) run() {
	for {
		select {
		case <-time.After(10 * time.Second):
			a.Lock()
			if a.numWorkers > 1 {
				a.numWorkers--
				a.Unlock()
				return
			}
			a.Unlock()
		case <-a.closeChan:
			return
		case reqOrig := <-a.WorkQ:
			tx, err := func(req *TxDownloadRequest) (interfaces.Transaction, error) {
				opts := []grpc.CallOption{
					grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backoffAmount*time.Millisecond, .1)),
					grpc_retry.WithMax(retryMax),
				}
				peerOpt := middleware.NewPeerInterceptor()
				newOpts := append(opts, peerOpt)
				txLst, err := a.reqBus.RequestP2PGetPendingTx(context.Background(), [][]byte{req.TxHash}, newOpts...)
				if err != nil {
					utils.DebugTrace(a.Logger, err)
					return nil, errorz.ErrInvalid{}.New(err.Error())
				}
				peer := peerOpt.Peer()
				if len(txLst) != 1 {
					peer.Feedback(-3)
					return nil, errorz.ErrInvalid{}.New("Downloaded more than 1 txn when only should have 1")
				}
				tx, err := a.reqBus.UnmarshalTx(utils.CopySlice(txLst[0]))
				if err != nil {
					peer.Feedback(-2)
					utils.DebugTrace(a.Logger, err)
					return nil, errorz.ErrInvalid{}.New(err.Error())
				}
				return tx, nil
			}(reqOrig)
			reqOrig.ResponseChan() <- NewTxDownloadResponse(reqOrig, tx, PendingTxRequest, err)
		}
	}
}

type blockHeaderDownloadActor struct {
	sync.Mutex
	numWorkers int
	WorkQ      chan *BlockHeaderDownloadRequest
	reqBus     typeProxyIface
	Logger     *logrus.Logger
	closeChan  chan struct{}
}

func (a *blockHeaderDownloadActor) init(workQ chan *BlockHeaderDownloadRequest, logger *logrus.Logger, reqBus typeProxyIface, closeChan chan struct{}) error {
	a.WorkQ = workQ
	a.reqBus = reqBus
	a.Logger = logger
	a.closeChan = closeChan
	return nil
}

func (a *blockHeaderDownloadActor) start() {
	a.Lock()
	defer a.Unlock()
	for i := 0; i < 2; i++ {
		if a.numWorkers < downloadWorkerCountBlockHeader {
			a.numWorkers++
			go a.run()
		}
	}
}

func (a *blockHeaderDownloadActor) run() {
	for {
		select {
		case <-time.After(10 * time.Second):
			a.Lock()
			if a.numWorkers > 1 {
				a.numWorkers--
				a.Unlock()
				return
			}
			a.Unlock()
		case <-a.closeChan:
			return
		case reqOrig := <-a.WorkQ:
			bh, err := func(req *BlockHeaderDownloadRequest) (*objs.BlockHeader, error) {
				opts := []grpc.CallOption{
					grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backoffAmount*time.Millisecond, .1)),
					grpc_retry.WithMax(retryMax),
				}
				peerOpt := middleware.NewPeerInterceptor()
				newOpts := append(opts, peerOpt)
				bhLst, err := a.reqBus.RequestP2PGetBlockHeaders(context.Background(), []uint32{req.Height}, newOpts...)
				if err != nil {
					utils.DebugTrace(a.Logger, err)
					return nil, errorz.ErrInvalid{}.New(err.Error())
				}
				peer := peerOpt.Peer()
				if len(bhLst) != 1 {
					peer.Feedback(-3)
					return nil, errorz.ErrInvalid{}.New("Downloaded more than 1 block header when only should have 1")
				}
				return bhLst[0], nil
			}(reqOrig)
			reqOrig.ResponseChan() <- NewBlockHeaderDownloadResponse(reqOrig, bh, BlockHeaderRequest, err)
		}
	}
}
