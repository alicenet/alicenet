package dman

import (
	"context"
	"fmt"
	"sync"

	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

// Root Actor spawns top level actor types
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
	return a.ba.init(a.dispatchQ, logger, proxy)
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

func (a *RootActor) DownloadPendingTx(height, round uint32, txHash []byte) {
	req := NewTxDownloadRequest(txHash, PendingTxRequest, height, round)
	a.download(req, false)
}

func (a *RootActor) DownloadMinedTx(height, round uint32, txHash []byte) {
	req := NewTxDownloadRequest(txHash, MinedTxRequest, height, round)
	a.download(req, false)
}

func (a *RootActor) DownloadTx(height, round uint32, txHash []byte) {
	req := NewTxDownloadRequest(txHash, PendingAndMinedTxRequest, height, round)
	a.download(req, false)
}

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
				if retry {
					return true
				}
				return false
			}
			a.reqs[b.Identifier()] = true
			return true
		}()
		if !ok {
			return
		}
		a.dispatchQ <- b
		a.await(b)
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
			responseChan: make(chan DownloadResponse),
		}
		bc1 := &TxDownloadRequest{
			TxHash:       []byte(b.Identifier()),
			Dtype:        MinedTxRequest,
			Height:       b.RequestHeight(),
			Round:        b.RequestRound(),
			responseChan: make(chan DownloadResponse),
		}
		ptxc := make(chan struct{})
		mtxc := make(chan struct{})
		go func() {
			defer close(ptxc)
			a.dispatchQ <- bc0
			a.await(bc0)
		}()
		go func() {
			defer close(mtxc)
			a.dispatchQ <- bc1
			a.await(bc1)
		}()
		select {
		case <-ptxc:
			<-mtxc
		case <-mtxc:
			<-ptxc
		}
	case BlockHeaderRequest:
		ok := func() bool {
			a.Lock()
			defer a.Unlock()
			if a.bhc.Contains(b.RequestHeight()) {
				return false
			}
			if _, exists := a.reqs[b.Identifier()]; exists {
				if retry {
					return true
				}
				return false
			}
			a.reqs[b.Identifier()] = true
			return true
		}()
		if !ok {
			return
		}
		a.dispatchQ <- b
		a.await(b)
	default:
		panic(b.DownloadType())
	}
}

func (a *RootActor) await(req DownloadRequest) {
	resp := <-req.ResponseChan()
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

type blockActor struct {
	sync.RWMutex
	ra            *downloadActor
	WorkQ         chan DownloadRequest
	DisptachQ     chan DownloadRequest
	CurrentHeight uint32
	Logger        *logrus.Logger
}

func (a *blockActor) init(workQ chan DownloadRequest, logger *logrus.Logger, reqBus typeProxyIface) error {
	a.WorkQ = workQ
	a.DisptachQ = make(chan DownloadRequest)
	a.ra = &downloadActor{}
	a.Logger = logger
	return a.ra.init(a.DisptachQ, logger, reqBus)
}

func (a *blockActor) start() {
	go a.run()
	a.ra.start()
}

func (a *blockActor) run() {
	for {
		req := <-a.WorkQ
		ok := func() bool {
			a.Lock()
			defer a.Unlock()
			if req.RequestHeight()+heightDropLag < a.CurrentHeight {
				close(req.ResponseChan())
				return false
			}
			if req.RequestHeight() > a.CurrentHeight {
				a.CurrentHeight = req.RequestHeight()
			}
			return true
		}()
		if !ok {
			continue
		}
		go a.await(req)
	}
}

func (a *blockActor) await(req DownloadRequest) {
	var subReq DownloadRequest
	switch req.DownloadType() {
	case PendingTxRequest, MinedTxRequest:
		reqTyped := req.(*TxDownloadRequest)
		subReq = NewTxDownloadRequest(reqTyped.TxHash, reqTyped.Dtype, reqTyped.Height, reqTyped.Round)
		a.DisptachQ <- subReq
	case BlockHeaderRequest:
		reqTyped := req.(*BlockHeaderDownloadRequest)
		subReq = NewBlockHeaderDownloadRequest(reqTyped.Height, reqTyped.Round, reqTyped.Dtype)
		a.DisptachQ <- subReq
	default:
		panic(fmt.Sprintf("req download type not found: %v", req.DownloadType()))
	}
	resp := <-subReq.ResponseChan()
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
	req.ResponseChan() <- resp
}

type downloadActor struct {
	WorkQ            chan DownloadRequest
	PendingDispatchQ chan *TxDownloadRequest
	MinedDispatchQ   chan *TxDownloadRequest
	BlockDispatchQ   chan *BlockHeaderDownloadRequest
	ptxa             *pendingDownloadActor
	mtxa             *minedDownloadActor
	bha              *blockHeaderDownloadActor
	Logger           *logrus.Logger
}

func (a *downloadActor) init(workQ chan DownloadRequest, logger *logrus.Logger, reqBus typeProxyIface) error {
	a.PendingDispatchQ = make(chan *TxDownloadRequest, downloadWorkerCount)
	a.MinedDispatchQ = make(chan *TxDownloadRequest, downloadWorkerCount)
	a.BlockDispatchQ = make(chan *BlockHeaderDownloadRequest, downloadWorkerCount)
	a.WorkQ = workQ
	a.ptxa = &pendingDownloadActor{}
	a.Logger = logger
	if err := a.ptxa.init(a.PendingDispatchQ, logger, reqBus); err != nil {
		return err
	}
	a.mtxa = &minedDownloadActor{}
	if err := a.mtxa.init(a.MinedDispatchQ, logger, reqBus); err != nil {
		return err
	}
	a.bha = &blockHeaderDownloadActor{}
	if err := a.bha.init(a.BlockDispatchQ, logger, reqBus); err != nil {
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
		case req := <-a.WorkQ:
			switch req.DownloadType() {
			case PendingTxRequest:
				a.PendingDispatchQ <- req.(*TxDownloadRequest)
			case MinedTxRequest:
				a.MinedDispatchQ <- req.(*TxDownloadRequest)
			case BlockHeaderRequest:
				a.BlockDispatchQ <- req.(*BlockHeaderDownloadRequest)
			default:
				panic(req.DownloadType())
			}
		}
	}
}

type minedDownloadActor struct {
	WorkQ  chan *TxDownloadRequest
	reqBus typeProxyIface
	Logger *logrus.Logger
}

func (a *minedDownloadActor) init(workQ chan *TxDownloadRequest, logger *logrus.Logger, reqBus typeProxyIface) error {
	a.WorkQ = workQ
	a.reqBus = reqBus
	a.Logger = logger
	return nil
}

func (a *minedDownloadActor) start() {
	for i := 0; i < downloadWorkerCount; i++ {
		go a.run()
	}
}

func (a *minedDownloadActor) run() {
	for {
		reqOrig := <-a.WorkQ
		tx, err := func(req *TxDownloadRequest) (interfaces.Transaction, error) {
			ctx := context.Background()
			subCtx, cf := context.WithTimeout(ctx, constants.MsgTimeout)
			defer cf()
			txLst, err := a.reqBus.RequestP2PGetMinedTxs(subCtx, [][]byte{req.TxHash})
			if err != nil {
				utils.DebugTrace(a.Logger, err)
				return nil, errorz.ErrInvalid{}.New(err.Error())
			}
			if len(txLst) != 1 {
				return nil, errorz.ErrInvalid{}.New("Downloaded more than 1 txn when only should have 1")
			}
			tx, err := a.reqBus.UnmarshalTx(utils.CopySlice(txLst[0]))
			if err != nil {
				utils.DebugTrace(a.Logger, err)
				return nil, errorz.ErrInvalid{}.New(err.Error())
			}
			return tx, nil
		}(reqOrig)
		reqOrig.ResponseChan() <- NewTxDownloadResponse(reqOrig, tx, MinedTxRequest, err)
	}
}

type pendingDownloadActor struct {
	WorkQ  chan *TxDownloadRequest
	reqBus typeProxyIface
	Logger *logrus.Logger
}

func (a *pendingDownloadActor) init(workQ chan *TxDownloadRequest, logger *logrus.Logger, reqBus typeProxyIface) error {
	a.WorkQ = workQ
	a.reqBus = reqBus
	a.Logger = logger
	return nil
}

func (a *pendingDownloadActor) start() {
	for i := 0; i < downloadWorkerCount; i++ {
		go a.run()
	}
}

func (a *pendingDownloadActor) run() {
	for {
		reqOrig := <-a.WorkQ
		tx, err := func(req *TxDownloadRequest) (interfaces.Transaction, error) {
			ctx := context.Background()
			subCtx, cf := context.WithTimeout(ctx, constants.MsgTimeout)
			defer cf()
			txLst, err := a.reqBus.RequestP2PGetPendingTx(subCtx, [][]byte{req.TxHash})
			if err != nil {
				utils.DebugTrace(a.Logger, err)
				return nil, errorz.ErrInvalid{}.New(err.Error())
			}
			if len(txLst) != 1 {
				return nil, errorz.ErrInvalid{}.New("Downloaded more than 1 txn when only should have 1")
			}
			tx, err := a.reqBus.UnmarshalTx(utils.CopySlice(txLst[0]))
			if err != nil {
				utils.DebugTrace(a.Logger, err)
				return nil, errorz.ErrInvalid{}.New(err.Error())
			}
			return tx, nil
		}(reqOrig)
		reqOrig.ResponseChan() <- NewTxDownloadResponse(reqOrig, tx, PendingTxRequest, err)
	}
}

type blockHeaderDownloadActor struct {
	WorkQ  chan *BlockHeaderDownloadRequest
	reqBus typeProxyIface
	Logger *logrus.Logger
}

func (a *blockHeaderDownloadActor) init(workQ chan *BlockHeaderDownloadRequest, logger *logrus.Logger, reqBus typeProxyIface) error {
	a.WorkQ = workQ
	a.reqBus = reqBus
	a.Logger = logger
	return nil
}

func (a *blockHeaderDownloadActor) start() {
	for i := 0; i < downloadWorkerCount; i++ {
		go a.run()
	}
}

func (a *blockHeaderDownloadActor) run() {
	for {
		reqOrig := <-a.WorkQ
		bh, err := func(req *BlockHeaderDownloadRequest) (*objs.BlockHeader, error) {
			ctx := context.Background()
			subCtx, cf := context.WithTimeout(ctx, constants.MsgTimeout)
			defer cf()
			bhLst, err := a.reqBus.RequestP2PGetBlockHeaders(subCtx, []uint32{req.Height})
			if err != nil {
				utils.DebugTrace(a.Logger, err)
				return nil, errorz.ErrInvalid{}.New(err.Error())
			}
			if len(bhLst) != 1 {
				return nil, errorz.ErrInvalid{}.New("Downloaded more than 1 block header when only should have 1")
			}
			return bhLst[0], nil
		}(reqOrig)
		reqOrig.ResponseChan() <- NewBlockHeaderDownloadResponse(reqOrig, bh, BlockHeaderRequest, err)
	}
}
