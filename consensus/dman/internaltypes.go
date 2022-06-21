package dman

import (
	"context"
	"strconv"

	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type DownloadType int

const (
	PendingTxRequest DownloadType = iota + 1
	MinedTxRequest
	PendingAndMinedTxRequest
	BlockHeaderRequest
)

const (
	heightDropLag                  = 5
	downloadWorkerCountMinedTx     = 512
	downloadWorkerCountPendingTx   = 512
	downloadWorkerCountBlockHeader = 256
)

type reqBusView interface {
	RequestP2PGetPendingTx(ctx context.Context, txHashes [][]byte, opts ...grpc.CallOption) ([][]byte, error)
	RequestP2PGetMinedTxs(ctx context.Context, txHashes [][]byte, opts ...grpc.CallOption) ([][]byte, error)
	RequestP2PGetBlockHeaders(ctx context.Context, blockNums []uint32, opts ...grpc.CallOption) ([]*objs.BlockHeader, error)
}

type txMarshaller interface {
	UnmarshalTx([]byte) (interfaces.Transaction, error)
}

type typeProxyIface interface {
	reqBusView
	txMarshaller
	databaseView
}

type databaseView interface {
	SetTxCacheItem(txn *badger.Txn, height uint32, txHash []byte, tx []byte) error
	GetTxCacheItem(txn *badger.Txn, height uint32, txHash []byte) ([]byte, error)
	SetCommittedBlockHeader(txn *badger.Txn, v *objs.BlockHeader) error
	TxCacheDropBefore(txn *badger.Txn, beforeHeight uint32, maxKeys int) error
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

type typeProxy struct {
	interfaces.Application
	reqBusView
	databaseView
}

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

type txResult struct {
	logger     *logrus.Logger
	appHandler interfaces.Application
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
