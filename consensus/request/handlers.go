package request

import (
	"context"
	"sync"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/dynamics"

	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/logging"
	pb "github.com/MadBase/MadNet/proto"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

var _ pb.P2PGetMinedTxsHandler = (*Handler)(nil)
var _ pb.P2PGetPendingTxsHandler = (*Handler)(nil)
var _ pb.P2PGetBlockHeadersHandler = (*Handler)(nil)
var _ pb.P2PGetSnapShotNodeHandler = (*Handler)(nil)
var _ pb.P2PGetSnapShotStateDataHandler = (*Handler)(nil)

type appHandler interface {
	PendingTxGet(txn *badger.Txn, height uint32, txHash [][]byte) ([]interfaces.Transaction, [][]byte, error)
	MinedTxGet(*badger.Txn, [][]byte) ([]interfaces.Transaction, [][]byte, error)
	GetSnapShotNode(txn *badger.Txn, height uint32, key []byte) ([]byte, error)
	GetSnapShotStateData(txn *badger.Txn, key []byte) ([]byte, error)
}

// Handler serves incoming requests and handles routing of outgoing
// requests for state from the consensus system.
type Handler struct {
	wg        sync.WaitGroup
	ctx       context.Context
	cancelCtx func()
	database  *db.Database
	logger    *logrus.Logger
	app       appHandler
	storage   dynamics.StorageGetter
}

// Init initializes the object
func (rb *Handler) Init(database *db.Database, app appHandler, storage dynamics.StorageGetter) {
	rb.logger = logging.GetLogger(constants.LoggerConsensus)
	background := context.Background()
	ctx, cf := context.WithCancel(background)
	rb.ctx = ctx
	rb.cancelCtx = cf
	rb.wg = sync.WaitGroup{}
	rb.app = app
	rb.database = database
	rb.storage = storage
}

// Done will trInger when both of the gossip busses have stopped
func (rb *Handler) Done() <-chan struct{} {
	rb.wg.Wait()
	return rb.ctx.Done()
}

// Start will start the gossip busses
func (rb *Handler) Start() {
	//do nothing
}

// Exit will kill the gossip busses
func (rb *Handler) Exit() {
	rb.cancelCtx()
}

//HandleP2PStatus serves status message from P2P protocol
func (rb *Handler) HandleP2PStatus(ctx context.Context, r *pb.StatusRequest) (*pb.StatusResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		rb.wg.Add(1)
		defer rb.wg.Done()
	}

	var resp pb.StatusResponse

	err := rb.database.View(func(txn *badger.Txn) error {
		os, err := rb.database.GetOwnState(txn)
		if err != nil {
			return err
		}
		resp.SyncToBlockHeight = os.SyncToBH.BClaims.Height
		resp.MaxBlockHeightSeen = os.MaxBHSeen.BClaims.Height
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &resp, nil
}

//HandleP2PGetBlockHeaders serves block headers
func (rb *Handler) HandleP2PGetBlockHeaders(ctx context.Context, r *pb.GetBlockHeadersRequest) (*pb.GetBlockHeadersResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		rb.wg.Add(1)
		defer rb.wg.Done()
	}
	hdrs := [][]byte{}
	byteCount := 0
	err := rb.database.View(func(txn *badger.Txn) error {
		for _, blknum := range r.BlockNumbers {
			hdrbytes, err := rb.database.GetCommittedBlockHeaderRaw(txn, blknum)
			if err != nil {
				return err
			}
			if len(hdrbytes)+byteCount < int(rb.storage.GetMaxBytes()) {
				byteCount = byteCount + len(hdrbytes)
				hdrs = append(hdrs, hdrbytes)
			} else {
				break
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := &pb.GetBlockHeadersResponse{
		BlockHeaders: hdrs,
	}
	return resp, nil
}

// HandleP2PGetPendingTxs serves pending txs
func (rb *Handler) HandleP2PGetPendingTxs(ctx context.Context, r *pb.GetPendingTxsRequest) (*pb.GetPendingTxsResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		rb.wg.Add(1)
		defer rb.wg.Done()
	}
	var txs [][]byte
	err := rb.database.View(func(txn *badger.Txn) error {
		var err error
		txi, _, err := rb.app.PendingTxGet(txn, 1, r.TxHashes)
		if err != nil {
			return err
		}
		for _, tx := range txi {
			txb, err := tx.MarshalBinary()
			if err != nil {
				return err
			}
			txs = append(txs, txb)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := &pb.GetPendingTxsResponse{
		Txs: txs,
	}
	return resp, nil
}

// HandleP2PGetMinedTxs returns the mined transactions from a specific block
// If the block number it is not known, it returns an empty byte slice.
func (rb *Handler) HandleP2PGetMinedTxs(ctx context.Context, r *pb.GetMinedTxsRequest) (*pb.GetMinedTxsResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		rb.wg.Add(1)
		defer rb.wg.Done()
	}
	var txs [][]byte
	err := rb.database.View(func(txn *badger.Txn) error {
		txilst, _, err := rb.app.MinedTxGet(txn, r.TxHashes)
		if err != nil {
			return err
		}
		txb := [][]byte{}
		for _, tx := range txilst {
			t, err := tx.MarshalBinary()
			if err != nil {
				return err
			}
			txb = append(txb, t)
		}
		txs = txb
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := &pb.GetMinedTxsResponse{
		Txs: txs,
	}
	return resp, nil
}

// HandleP2PGetSnapShotHdrNode serves nodes of the Header Trie to the caller
func (rb *Handler) HandleP2PGetSnapShotNode(ctx context.Context, r *pb.GetSnapShotNodeRequest) (*pb.GetSnapShotNodeResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var node []byte
	err := rb.database.View(func(txn *badger.Txn) error {
		tmp, err := rb.app.GetSnapShotNode(txn, r.Height, r.NodeHash)
		if err != nil {
			return err
		}
		node = tmp
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := &pb.GetSnapShotNodeResponse{
		Node: node,
	}
	return resp, nil
}

// HandleP2PGetSnapShotHdrNode serves nodes of the State Trie to the caller
func (rb *Handler) HandleP2PGetSnapShotHdrNode(ctx context.Context, r *pb.GetSnapShotHdrNodeRequest) (*pb.GetSnapShotHdrNodeResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var node []byte
	err := rb.database.View(func(txn *badger.Txn) error {
		tmp, err := rb.database.GetSnapShotHdrNode(txn, r.NodeHash)
		if err != nil {
			return err
		}
		node = tmp
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := &pb.GetSnapShotHdrNodeResponse{
		Node: node,
	}
	return resp, nil
}

// HandleP2PGetSnapShotStateData serves UTXOs based on State Trie hash state
func (rb *Handler) HandleP2PGetSnapShotStateData(ctx context.Context, r *pb.GetSnapShotStateDataRequest) (*pb.GetSnapShotStateDataResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		rb.wg.Add(1)
		defer rb.wg.Done()
	}
	var data []byte
	err := rb.database.View(func(txn *badger.Txn) error {
		tmp, err := rb.app.GetSnapShotStateData(txn, r.Key)
		if err != nil {
			return err
		}
		data = tmp
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := &pb.GetSnapShotStateDataResponse{
		Data: data,
	}
	return resp, nil
}
