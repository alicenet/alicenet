package dman

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"

	"github.com/MadBase/MadNet/consensus/appmock"
	"github.com/MadBase/MadNet/consensus/db"

	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/consensus/request"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

type pTXRFunc func(context.Context, [][]byte) ([][]byte, error)
type mTXRFunc func(context.Context, [][]byte) ([][]byte, error)
type bHRFunc func(context.Context, []uint32) ([]*objs.BlockHeader, error)
type uTXFunc func([]byte) (interfaces.Transaction, error)

const (
	heightDropLag       = 2
	downloadWorkerCount = 4
)

type txResult struct {
	logger     *logrus.Logger
	appHandler appmock.Application
	txs        []interfaces.Transaction
	txHashes   map[string]bool
}

func (t *txResult) init(txHashes [][]byte) {
	t.txs = []interfaces.Transaction{}
	t.txHashes = make(map[string]bool)
	for _, txHash := range txHashes {
		t.txHashes[string(txHash)] = false
	}
}

func (t *txResult) missing() [][]byte {
	var missing [][]byte
	for txHash, haveIt := range t.txHashes {
		if !haveIt {
			missing = append(missing, utils.CopySlice([]byte(txHash)))
		}
	}
	return missing
}

func (t *txResult) add(tx interfaces.Transaction) error {
	txHash, err := tx.TxHash()
	if err != nil {
		return err
	}
	haveIt, shouldHaveIt := t.txHashes[string(txHash)]
	if !haveIt && shouldHaveIt {
		t.txs = append(t.txs, tx)
		t.txHashes[string(txHash)] = true
	}
	return nil
}

func (t *txResult) addMany(txs []interfaces.Transaction) error {
	var err error
	for i := 0; i < len(txs); i++ {
		e := t.add(txs[i])
		if e != nil {
			err = e
			utils.DebugTrace(t.logger, err)
		}
	}
	return err
}

func (t *txResult) addRaw(txb []byte) error {
	tx, err := t.appHandler.UnmarshalTx(utils.CopySlice(txb))
	if err != nil {
		utils.DebugTrace(t.logger, err)
		return err
	}
	return t.add(tx)
}

type DMan struct {
	downloadActor *RootActor
	database      *db.Database
	appHandler    appmock.Application
	bnVal         *crypto.BNGroupValidator
	logger        *logrus.Logger
}

func (dm *DMan) Init(database *db.Database, app appmock.Application, reqBus *request.Client) error {
	dm.logger = logging.GetLogger(constants.LoggerDMan)
	dm.database = database
	dm.appHandler = app
	dm.bnVal = &crypto.BNGroupValidator{}
	dm.downloadActor = &RootActor{}
	dm.downloadActor.Init(database, dm.logger, reqBus.RequestP2PGetPendingTx, reqBus.RequestP2PGetMinedTxs, reqBus.RequestP2PGetBlockHeaders, dm.appHandler.UnmarshalTx)
	dm.downloadActor.txc = &txCache{
		app: dm.appHandler,
	}
	err := dm.downloadActor.txc.Init()
	if err != nil {
		utils.DebugTrace(dm.logger, err)
		return err
	}
	dm.downloadActor.bhc = &bHCache{}
	err = dm.downloadActor.bhc.Init()
	if err != nil {
		utils.DebugTrace(dm.logger, err)
		return err
	}
	return nil
}

func (dm *DMan) Start() {
	dm.downloadActor.Start()
}

func (dm *DMan) Close() {
	dm.downloadActor.wg.Wait()
}

func (dm *DMan) FlushCacheToDisk(txn *badger.Txn, height uint32) error {
	return dm.downloadActor.FlushCacheToDisk(txn, height)
}

func (dm *DMan) AddTxs(txn *badger.Txn, height uint32, txs []interfaces.Transaction) error {
	for i := 0; i < len(txs); i++ {
		tx := txs[i]
		txHash, err := tx.TxHash()
		if err != nil {
			utils.DebugTrace(dm.logger, err)
			return err
		}
		txb, err := tx.MarshalBinary()
		if err != nil {
			utils.DebugTrace(dm.logger, err)
			return err
		}
		if err := dm.database.SetTxCacheItem(txn, height, utils.CopySlice(txHash), utils.CopySlice(txb)); err != nil {
			utils.DebugTrace(dm.logger, err)
			return err
		}
	}
	return nil
}

func (dm *DMan) GetTxs(txn *badger.Txn, height, round uint32, txLst [][]byte) ([]interfaces.Transaction, [][]byte, error) {
	result := &txResult{appHandler: dm.appHandler, logger: dm.logger}
	result.init(txLst)
	missing := result.missing()
	// get from the database
	for i := 0; i < len(missing); i++ {
		txHash := utils.CopySlice(missing[i])
		txb, err := dm.database.GetTxCacheItem(txn, height, txHash)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				utils.DebugTrace(dm.logger, err)
				return nil, nil, err
			}
			continue
		}
		if err := result.addRaw(txb); err != nil {
			return nil, nil, err
		}
	}
	missing = result.missing()
	// get from the pending store
	found, _, err := dm.appHandler.PendingTxGet(txn, height, missing)
	if err != nil {
		var e *errorz.ErrInvalid
		if err != errorz.ErrMissingTransactions && !errors.As(err, &e) && err != badger.ErrKeyNotFound {
			utils.DebugTrace(dm.logger, err)
			return nil, nil, err
		}
	}
	if err := result.addMany(found); err != nil {
		utils.DebugTrace(dm.logger, err)
		return nil, nil, err
	}
	missing = result.missing()
	if len(missing) > 0 {
		dm.DownloadTxs(height, round, missing)
	}
	missing = result.missing()
	return result.txs, missing, nil
}

// SyncOneBH syncs one blockheader and its transactions
// the initialization of prevBH from SyncToBH implies SyncToBH must be updated to
// the canonical bh before we begin unless we are syncing from a height gt the
// canonical bh
func (dm *DMan) SyncOneBH(txn *badger.Txn, syncToBH *objs.BlockHeader, validatorSet *objs.ValidatorSet) ([]interfaces.Transaction, *objs.BlockHeader, bool, error) {
	targetHeight := syncToBH.BClaims.Height + 1
	bhCache, inCache := dm.downloadActor.bhc.Get(targetHeight)
	if !inCache {
		dm.downloadActor.DownloadBlockHeader(targetHeight, 1)
		return nil, nil, false, nil
		//return nil, nil, errorz.ErrInvalid{}.New("block header was not in the bh cache")
	}
	// check the chainID of bh
	if bhCache.BClaims.ChainID != syncToBH.BClaims.ChainID {
		dm.downloadActor.bhc.Del(targetHeight)
		return nil, nil, false, errorz.ErrInvalid{}.New("Wrong chainID")
	}
	// check the height of the bh
	if bhCache.BClaims.Height != targetHeight {
		dm.downloadActor.bhc.Del(targetHeight)
		return nil, nil, false, errorz.ErrInvalid{}.New("Wrong block height")
	}
	// get prev bh
	prevBHsh, err := syncToBH.BlockHash()
	if err != nil {
		dm.downloadActor.bhc.Del(targetHeight)
		utils.DebugTrace(dm.logger, err)
		return nil, nil, false, err
	}
	// compare to prevBlock from bh
	if !bytes.Equal(bhCache.BClaims.PrevBlock, prevBHsh) {
		dm.downloadActor.bhc.Del(targetHeight)
		return nil, nil, false, errorz.ErrInvalid{}.New("BlockHash does not match previous!")
	}
	txs, _, err := dm.GetTxs(txn, targetHeight, 1, bhCache.TxHshLst)
	if err != nil {
		utils.DebugTrace(dm.logger, err)
		return nil, nil, false, err
	}
	// verify the signature and group key
	if err := bhCache.ValidateSignatures(dm.bnVal); err != nil {
		utils.DebugTrace(dm.logger, err)
		return nil, nil, false, errorz.ErrInvalid{}.New(err.Error())
	}
	if !bytes.Equal(bhCache.GroupKey, validatorSet.GroupKey) {
		return nil, nil, false, errorz.ErrInvalid{}.New("group key does not match expected")
	}
	if err := dm.database.SetCommittedBlockHeader(txn, bhCache); err != nil {
		return nil, nil, false, err
	}
	if syncToBH.BClaims.Height > 10 {
		if err := dm.database.TxCacheDropBefore(txn, syncToBH.BClaims.Height-5, 1000); err != nil {
			utils.DebugTrace(dm.logger, err)
			return nil, nil, false, err
		}
	}

	return txs, bhCache, true, nil

}

func (dm *DMan) DownloadTxs(height, round uint32, txHshLst [][]byte) {
	for i := 0; i < len(txHshLst); i++ {
		txHsh := txHshLst[i]
		dm.downloadActor.DownloadTx(height, round, txHsh)
	}
}

type DownloadRequest interface {
	DownloadType() DownloadType
	IsRequest() bool
	RequestHeight() uint32
	RequestRound() uint32
	ResponseChan() chan DownloadResponse
	Identifier() string
}

type DownloadResponse interface {
	DownloadType() DownloadType
	IsResponse() bool
	RequestHeight() uint32
	RequestRound() uint32
}

type DownloadType int

const (
	PendingTxRequest DownloadType = iota + 1
	MinedTxRequest
	PendingAndMinedTxRequest
	BlockHeaderRequest
)

type TxDownloadRequest struct {
	TxHash       []byte       `json:"tx_hash,omitempty"`
	Dtype        DownloadType `json:"dtype,omitempty"`
	Height       uint32       `json:"height,omitempty"`
	Round        uint32       `json:"round,omitempty"`
	responseChan chan DownloadResponse
}

func (r *TxDownloadRequest) DownloadType() DownloadType {
	return r.Dtype
}

func (r *TxDownloadRequest) IsRequest() bool {
	return true
}

func (r *TxDownloadRequest) RequestHeight() uint32 {
	return r.Height
}

func (r *TxDownloadRequest) RequestRound() uint32 {
	return r.Round
}

func (r *TxDownloadRequest) ResponseChan() chan DownloadResponse {
	return r.responseChan
}

func (r *TxDownloadRequest) Identifier() string {
	return string(r.TxHash)
}

func NewTxDownloadRequest(txHash []byte, downloadType DownloadType, height, round uint32) *TxDownloadRequest {
	responseChan := make(chan DownloadResponse, 1)
	return &TxDownloadRequest{
		responseChan: responseChan,
		Dtype:        downloadType,
		TxHash:       utils.CopySlice(txHash),
		Height:       height,
		Round:        round,
	}
}

type TxDownloadResponse struct {
	TxHash []byte       `json:"tx_hash,omitempty"`
	Dtype  DownloadType `json:"dtype,omitempty"`
	Tx     interfaces.Transaction
	Err    error  `json:"err,omitempty"`
	Height uint32 `json:"height,omitempty"`
	Round  uint32 `json:"round,omitempty"`
}

func (r *TxDownloadResponse) DownloadType() DownloadType {
	return r.Dtype
}

func (r *TxDownloadResponse) IsResponse() bool {
	return true
}

func (r *TxDownloadResponse) RequestHeight() uint32 {
	return r.Height
}

func (r *TxDownloadResponse) RequestRound() uint32 {
	return r.Round
}

func NewTxDownloadResponse(req *TxDownloadRequest, tx interfaces.Transaction, dlt DownloadType, err error) *TxDownloadResponse {
	return &TxDownloadResponse{
		Dtype:  dlt,
		TxHash: utils.CopySlice(req.TxHash),
		Tx:     tx,
		Err:    err,
		Height: req.Height,
		Round:  req.Round,
	}
}

type BlockHeaderDownloadRequest struct {
	Height       uint32       `json:"height,omitempty"`
	Dtype        DownloadType `json:"dtype,omitempty"`
	Round        uint32       `json:"round,omitempty"`
	responseChan chan DownloadResponse
}

func (r *BlockHeaderDownloadRequest) DownloadType() DownloadType {
	return r.Dtype
}

func (r *BlockHeaderDownloadRequest) IsRequest() bool {
	return true
}

func (r *BlockHeaderDownloadRequest) RequestHeight() uint32 {
	return r.Height
}

func (r *BlockHeaderDownloadRequest) RequestRound() uint32 {
	return r.Round
}

func (r *BlockHeaderDownloadRequest) ResponseChan() chan DownloadResponse {
	return r.responseChan
}

func (r *BlockHeaderDownloadRequest) Identifier() string {
	return strconv.Itoa(int(r.Height))
}

func NewBlockHeaderDownloadRequest(height uint32, round uint32, downloadType DownloadType) *BlockHeaderDownloadRequest {
	responseChan := make(chan DownloadResponse, 1)
	return &BlockHeaderDownloadRequest{
		responseChan: responseChan,
		Dtype:        downloadType,
		Height:       height,
		Round:        round,
	}
}

type BlockHeaderDownloadResponse struct {
	Height uint32       `json:"height,omitempty"`
	Dtype  DownloadType `json:"dtype,omitempty"`
	BH     *objs.BlockHeader
	Err    error  `json:"err,omitempty"`
	Round  uint32 `json:"round,omitempty"`
}

func (r *BlockHeaderDownloadResponse) IsResponse() bool {
	return true
}

func (r *BlockHeaderDownloadResponse) DownloadType() DownloadType {
	return r.Dtype
}

func (r *BlockHeaderDownloadResponse) RequestHeight() uint32 {
	return r.Height
}

func (r *BlockHeaderDownloadResponse) RequestRound() uint32 {
	return r.Round
}

func NewBlockHeaderDownloadResponse(req *BlockHeaderDownloadRequest, bh *objs.BlockHeader, dlt DownloadType, err error) *BlockHeaderDownloadResponse {
	return &BlockHeaderDownloadResponse{
		Dtype:  req.Dtype,
		Height: req.Height,
		Round:  req.Round,
		BH:     bh,
		Err:    err,
	}
}

// Root Actor spawns top level actor types
type RootActor struct {
	sync.Mutex
	ba        *BlockActor
	wg        *sync.WaitGroup
	closeChan chan struct{}
	dispatchQ chan DownloadRequest
	database  *db.Database
	txc       *txCache
	bhc       *bHCache
	logger    *logrus.Logger
	reqs      map[string]bool
}

func (a *RootActor) Init(database *db.Database, logger *logrus.Logger, ptxrf pTXRFunc, mtxrf mTXRFunc, bhrf bHRFunc, umtxf uTXFunc) error {
	a.reqs = make(map[string]bool)
	a.wg = new(sync.WaitGroup)
	a.closeChan = make(chan struct{})
	a.dispatchQ = make(chan DownloadRequest)
	a.logger = logger
	a.ba = &BlockActor{}
	a.database = database
	return a.ba.Init(a.wg, a.closeChan, a.dispatchQ, logger, ptxrf, mtxrf, bhrf, umtxf)
}

func (a *RootActor) Start() {
	a.wg.Add(1)
	go a.ba.Start()
}

// TODO verify blockheader cache is being cleaned
func (a *RootActor) FlushCacheToDisk(txn *badger.Txn, height uint32) error {
	txList, txHashList := a.txc.GetHeight(height)
	for i := 0; i < len(txList); i++ {
		txb, err := txList[i].MarshalBinary()
		if err != nil {
			return err
		}
		if err := a.database.SetTxCacheItem(txn, height, []byte(txHashList[i]), txb); err != nil {
			return err
		}
	}
	dropKeys := a.bhc.DropBeforeHeight(height)
	dropKeys = append(dropKeys, a.txc.DropBeforeHeight(height)...)
	for i := 0; i < len(dropKeys); i++ {
		delete(a.reqs, dropKeys[i])
	}
	return nil
}

func (a *RootActor) DownloadPendingTx(height, round uint32, txHash []byte) {
	req := NewTxDownloadRequest(txHash, PendingTxRequest, height, round)
	a.download(req)
}

func (a *RootActor) DownloadMinedTx(height, round uint32, txHash []byte) {
	req := NewTxDownloadRequest(txHash, MinedTxRequest, height, round)
	a.download(req)
}

func (a *RootActor) DownloadTx(height, round uint32, txHash []byte) {
	req := NewTxDownloadRequest(txHash, PendingAndMinedTxRequest, height, round)
	a.download(req)
}

func (a *RootActor) DownloadBlockHeader(height, round uint32) {
	req := NewBlockHeaderDownloadRequest(height, round, BlockHeaderRequest)
	a.download(req)
}

func (a *RootActor) download(b DownloadRequest) {
	switch b.DownloadType() {
	case PendingAndMinedTxRequest, PendingTxRequest, MinedTxRequest:
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
	case BlockHeaderRequest:
		ok := func() bool {
			a.Lock()
			defer a.Unlock()
			if a.bhc.Contains(b.RequestHeight()) {
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
	default:
		panic(b.DownloadType())
	}
	select {
	case a.dispatchQ <- b:
		a.wg.Add(1)
		go a.await(b)
	case <-a.closeChan:
		return
	}
}

func (a *RootActor) await(req DownloadRequest) {
	defer a.wg.Done()
	select {
	case resp := <-req.ResponseChan():
		if resp == nil {
			return
		}
		a.processAwait(req, resp)
	case <-a.closeChan:
		return
	}
}

func (a *RootActor) processAwait(req DownloadRequest, resp DownloadResponse) {
	switch resp.DownloadType() {
	case PendingTxRequest, MinedTxRequest, PendingAndMinedTxRequest:
		r := resp.(*TxDownloadResponse)
		if r.Err != nil {
			utils.DebugTrace(a.logger, r.Err)
			a.download(req)
			return
		}
		ok := func() bool {
			if err := a.txc.Add(resp.RequestHeight(), r.Tx); err != nil {
				utils.DebugTrace(a.logger, err)
				return true
			}
			return false
		}()
		if ok {
			return
		}
		a.download(req)
	case BlockHeaderRequest:
		r := resp.(*BlockHeaderDownloadResponse)
		if r.Err != nil {
			utils.DebugTrace(a.logger, r.Err)
			a.download(req)
			return
		}
		ok := func() bool {
			if err := a.bhc.Add(r.BH); err != nil {
				utils.DebugTrace(a.logger, err)
				return true
			}
			return false
		}()
		if ok {
			return
		}
		a.download(req)
	default:
		panic(req.DownloadType())
	}
}

type BlockActor struct {
	sync.RWMutex
	//ra            *RoundActor
	ra            *DownloadActor
	wg            *sync.WaitGroup
	CloseChan     chan struct{}
	WorkQ         chan DownloadRequest
	DisptachQ     chan DownloadRequest
	CurrentHeight uint32
	Logger        *logrus.Logger
}

func (a *BlockActor) Init(wg *sync.WaitGroup, closeChan chan struct{}, workQ chan DownloadRequest, logger *logrus.Logger, ptxrf pTXRFunc, mtxrf mTXRFunc, bhrf bHRFunc, umtxf uTXFunc) error {
	a.wg = wg
	a.CloseChan = closeChan
	a.WorkQ = workQ
	a.DisptachQ = make(chan DownloadRequest)
	//a.ra = &RoundActor{}
	a.ra = &DownloadActor{}
	a.Logger = logger
	return a.ra.Init(wg, closeChan, a.DisptachQ, logger, ptxrf, mtxrf, bhrf, umtxf)
}

func (a *BlockActor) Start() {
	a.wg.Add(1)
	go a.run()
	a.ra.Start()
}

func (a *BlockActor) run() {
	defer a.wg.Done()
	for {
		select {
		case req := <-a.WorkQ:
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
			go a.Await(req)
		case <-a.CloseChan:
			return
		}
	}
}

func (a *BlockActor) Await(req DownloadRequest) {
	var subReq DownloadRequest
	switch req.DownloadType() {
	case PendingTxRequest, MinedTxRequest, PendingAndMinedTxRequest:
		reqTyped := req.(*TxDownloadRequest)
		subReq = NewTxDownloadRequest(reqTyped.TxHash, reqTyped.Dtype, reqTyped.Height, reqTyped.Round)
		select {
		case a.DisptachQ <- subReq:
		case <-a.CloseChan:
			return
		}
	case BlockHeaderRequest:
		reqTyped := req.(*BlockHeaderDownloadRequest)
		subReq = NewBlockHeaderDownloadRequest(reqTyped.Height, reqTyped.Round, reqTyped.Dtype)
		select {
		case a.DisptachQ <- subReq:
		case <-a.CloseChan:
			return
		}
	default:
		panic(fmt.Sprintf("req download type not found: %v", req.DownloadType()))
	}
	select {
	case resp := <-subReq.ResponseChan():
		if resp == nil {
			close(req.ResponseChan())
			return
		}
		ok := func() bool {
			a.RLock()
			defer a.RUnlock()
			return resp.RequestHeight() >= a.CurrentHeight-heightDropLag
		}()
		if !ok {
			close(req.ResponseChan())
			return
		}
		select {
		case req.ResponseChan() <- resp:
			return
		case <-a.CloseChan:
			return
		}
	case <-a.CloseChan:
		return
	}
}

type DownloadActor struct {
	wg               *sync.WaitGroup
	CloseChan        chan struct{}
	WorkQ            chan DownloadRequest
	PendingDispatchQ chan *TxDownloadRequest
	MinedDispatchQ   chan *TxDownloadRequest
	BlockDispatchQ   chan *BlockHeaderDownloadRequest
	ptxa             *PendingDownloadActor
	mtxa             *MinedDownloadActor
	bha              *BlockHeaderDownloadActor
	Logger           *logrus.Logger
}

func (a *DownloadActor) Init(wg *sync.WaitGroup, closeChan chan struct{}, workQ chan DownloadRequest, logger *logrus.Logger, ptxrf pTXRFunc, mtxrf mTXRFunc, bhrf bHRFunc, umtxf uTXFunc) error {
	a.PendingDispatchQ = make(chan *TxDownloadRequest, downloadWorkerCount)
	a.MinedDispatchQ = make(chan *TxDownloadRequest, downloadWorkerCount)
	a.BlockDispatchQ = make(chan *BlockHeaderDownloadRequest, downloadWorkerCount)
	a.wg = wg
	a.CloseChan = closeChan
	a.WorkQ = workQ
	a.ptxa = &PendingDownloadActor{}
	a.Logger = logger
	if err := a.ptxa.Init(wg, closeChan, a.PendingDispatchQ, logger, ptxrf, umtxf); err != nil {
		return err
	}
	a.mtxa = &MinedDownloadActor{}
	if err := a.mtxa.Init(wg, closeChan, a.MinedDispatchQ, logger, mtxrf, umtxf); err != nil {
		return err
	}
	a.bha = &BlockHeaderDownloadActor{}
	if err := a.bha.Init(wg, closeChan, a.BlockDispatchQ, logger, bhrf); err != nil {
		return err
	}
	return nil
}

func (a *DownloadActor) Start() {
	a.wg.Add(1)
	go a.run()
	a.ptxa.Start()
	a.mtxa.Start()
	a.bha.Start()
}

func (a *DownloadActor) run() {
	defer a.wg.Done()
	for {
		select {
		case req := <-a.WorkQ:
			switch req.DownloadType() {
			case PendingTxRequest:
				select {
				case a.PendingDispatchQ <- req.(*TxDownloadRequest):
				case <-a.CloseChan:
					return
				}
			case MinedTxRequest:
				select {
				case a.MinedDispatchQ <- req.(*TxDownloadRequest):
				case <-a.CloseChan:
					return
				}
			case PendingAndMinedTxRequest:
				select {
				case a.MinedDispatchQ <- req.(*TxDownloadRequest):
					select {
					case a.PendingDispatchQ <- req.(*TxDownloadRequest):
					case <-a.CloseChan:
						return
					}
				case a.PendingDispatchQ <- req.(*TxDownloadRequest):
					select {
					case a.MinedDispatchQ <- req.(*TxDownloadRequest):
					case <-a.CloseChan:
						return
					}
				case <-a.CloseChan:
					return
				}
			case BlockHeaderRequest:
				select {
				case a.BlockDispatchQ <- req.(*BlockHeaderDownloadRequest):
				case <-a.CloseChan:
					return
				}
			default:
				panic(req.DownloadType())
			}
		case <-a.CloseChan:
			return
		}
	}
}

type MinedDownloadActor struct {
	wg                    *sync.WaitGroup
	CloseChan             chan struct{}
	WorkQ                 chan *TxDownloadRequest
	RequestP2PGetMinedTxs func(context.Context, [][]byte) ([][]byte, error)
	UnmarshalTx           func([]byte) (interfaces.Transaction, error)
	Logger                *logrus.Logger
}

func (a *MinedDownloadActor) Init(wg *sync.WaitGroup, closeChan chan struct{}, workQ chan *TxDownloadRequest, logger *logrus.Logger, mtxrf mTXRFunc, umtxf uTXFunc) error {
	a.wg = wg
	a.CloseChan = closeChan
	a.WorkQ = workQ
	a.CloseChan = closeChan
	a.RequestP2PGetMinedTxs = mtxrf
	a.UnmarshalTx = umtxf
	a.Logger = logger
	return nil
}

func (a *MinedDownloadActor) Start() {
	for i := 0; i < downloadWorkerCount; i++ {
		a.wg.Add(1)
		go a.run()
	}
}

func (a *MinedDownloadActor) run() {
	defer a.wg.Done()
	for {
		select {
		case <-a.CloseChan:
			return
		case reqOrig := <-a.WorkQ:
			tx, err := func(req *TxDownloadRequest) (interfaces.Transaction, error) {
				ctx := context.Background()
				subCtx, cf := context.WithTimeout(ctx, constants.MsgTimeout)
				defer cf()
				txLst, err := a.RequestP2PGetMinedTxs(subCtx, [][]byte{req.TxHash})
				if err != nil {
					utils.DebugTrace(a.Logger, err)
					return nil, errorz.ErrInvalid{}.New(err.Error())
				}
				if len(txLst) != 1 {
					return nil, errorz.ErrInvalid{}.New("Downloaded more than 1 txn when only should have 1")
				}
				tx, err := a.UnmarshalTx(utils.CopySlice(txLst[0]))
				if err != nil {
					utils.DebugTrace(a.Logger, err)
					return nil, errorz.ErrInvalid{}.New(err.Error())
				}
				return tx, nil
			}(reqOrig)
			select {
			case reqOrig.ResponseChan() <- NewTxDownloadResponse(reqOrig, tx, MinedTxRequest, err):
				continue
			case <-a.CloseChan:
				return
			}
		}
	}
}

type PendingDownloadActor struct {
	wg                      *sync.WaitGroup
	CloseChan               chan struct{}
	WorkQ                   chan *TxDownloadRequest
	RequestP2PGetPendingTxs func(context.Context, [][]byte) ([][]byte, error)
	UnmarshalTx             func([]byte) (interfaces.Transaction, error)
	Logger                  *logrus.Logger
}

func (a *PendingDownloadActor) Init(wg *sync.WaitGroup, closeChan chan struct{}, workQ chan *TxDownloadRequest, logger *logrus.Logger, ptxrf pTXRFunc, umtxf uTXFunc) error {
	a.wg = wg
	a.CloseChan = closeChan
	a.WorkQ = workQ
	a.RequestP2PGetPendingTxs = ptxrf
	a.UnmarshalTx = umtxf
	a.Logger = logger
	return nil
}

func (a *PendingDownloadActor) Start() {
	for i := 0; i < downloadWorkerCount; i++ {
		a.wg.Add(1)
		go a.run()
	}
}

func (a *PendingDownloadActor) run() {
	defer a.wg.Done()
	for {
		select {
		case <-a.CloseChan:
			return
		case reqOrig := <-a.WorkQ:
			tx, err := func(req *TxDownloadRequest) (interfaces.Transaction, error) {
				ctx := context.Background()
				subCtx, cf := context.WithTimeout(ctx, constants.MsgTimeout)
				defer cf()
				txLst, err := a.RequestP2PGetPendingTxs(subCtx, [][]byte{req.TxHash})
				if err != nil {
					utils.DebugTrace(a.Logger, err)
					return nil, errorz.ErrInvalid{}.New(err.Error())
				}
				if len(txLst) != 1 {
					return nil, errorz.ErrInvalid{}.New("Downloaded more than 1 txn when only should have 1")
				}
				tx, err := a.UnmarshalTx(utils.CopySlice(txLst[0]))
				if err != nil {
					utils.DebugTrace(a.Logger, err)
					return nil, errorz.ErrInvalid{}.New(err.Error())
				}
				return tx, nil
			}(reqOrig)
			select {
			case reqOrig.ResponseChan() <- NewTxDownloadResponse(reqOrig, tx, PendingTxRequest, err):
				continue
			case <-a.CloseChan:
				return
			}
		}
	}
}

type BlockHeaderDownloadActor struct {
	wg                        *sync.WaitGroup
	CloseChan                 chan struct{}
	WorkQ                     chan *BlockHeaderDownloadRequest
	RequestP2PGetBlockHeaders func(context.Context, []uint32) ([]*objs.BlockHeader, error)
	Logger                    *logrus.Logger
}

func (a *BlockHeaderDownloadActor) Init(wg *sync.WaitGroup, closeChan chan struct{}, workQ chan *BlockHeaderDownloadRequest, logger *logrus.Logger, bhrf bHRFunc) error {
	a.wg = wg
	a.CloseChan = closeChan
	a.WorkQ = workQ
	a.RequestP2PGetBlockHeaders = bhrf
	a.Logger = logger
	return nil
}

func (a *BlockHeaderDownloadActor) Start() {
	for i := 0; i < downloadWorkerCount; i++ {
		a.wg.Add(1)
		go a.run()
	}
}

func (a *BlockHeaderDownloadActor) run() {
	defer a.wg.Done()
	for {
		select {
		case <-a.CloseChan:
			return
		case reqOrig := <-a.WorkQ:
			bh, err := func(req *BlockHeaderDownloadRequest) (*objs.BlockHeader, error) {
				ctx := context.Background()
				subCtx, cf := context.WithTimeout(ctx, constants.MsgTimeout)
				defer cf()
				bhLst, err := a.RequestP2PGetBlockHeaders(subCtx, []uint32{req.Height})
				if err != nil {
					utils.DebugTrace(a.Logger, err)
					return nil, errorz.ErrInvalid{}.New(err.Error())
				}
				if len(bhLst) != 1 {
					return nil, errorz.ErrInvalid{}.New("Downloaded more than 1 block header when only should have 1")
				}
				return bhLst[0], nil
			}(reqOrig)
			select {
			case reqOrig.ResponseChan() <- NewBlockHeaderDownloadResponse(reqOrig, bh, BlockHeaderRequest, err):
				continue
			case <-a.CloseChan:
				return
			}
		}
	}
}
