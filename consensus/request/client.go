package request

import (
	"context"

	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/dynamics"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/middleware"
	pb "github.com/MadBase/MadNet/proto"
	"github.com/MadBase/MadNet/utils"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Client serves incoming requests and handles routing of outgoing
// requests for state from the consensus system.
type Client struct {
	client   pb.P2PClient
	logger   *logrus.Logger
	secpVal  *crypto.Secp256k1Validator
	groupVal *crypto.BNGroupValidator
	storage  dynamics.StorageGetter
}

// Init initializes the object
func (rb *Client) Init(client pb.P2PClient, storage dynamics.StorageGetter) {
	rb.logger = logging.GetLogger(constants.LoggerConsensus)
	rb.client = client
	rb.groupVal = &crypto.BNGroupValidator{}
	rb.secpVal = &crypto.Secp256k1Validator{}
	rb.storage = storage
}

// RequestP2PGetSnapShotNode implements the client for the P2P method
// GetSnapShotNode
func (rb *Client) RequestP2PGetSnapShotNode(ctx context.Context, height uint32, key []byte, opts ...grpc.CallOption) ([]byte, error) {
	req := &pb.GetSnapShotNodeRequest{
		Height:   height,
		NodeHash: key,
	}
	peerOpt := middleware.NewPeerInterceptor()
	newOpts := append(opts, peerOpt)
	resp, err := rb.client.GetSnapShotNode(ctx, req, newOpts...)
	if err != nil {
		return nil, err
	}
	peer := peerOpt.Peer()
	if resp.Node == nil {
		peer.Feedback(-2)
		return nil, errorz.ErrBadResponse
	}
	return resp.Node, nil
}

// RequestP2PGetSnapShotHdrNode implements the client for the P2P method
// GetSnapShotHdrNode
func (rb *Client) RequestP2PGetSnapShotHdrNode(ctx context.Context, key []byte, opts ...grpc.CallOption) ([]byte, error) {
	req := &pb.GetSnapShotHdrNodeRequest{
		NodeHash: key,
	}
	peerOpt := middleware.NewPeerInterceptor()
	newOpts := append(opts, peerOpt)
	resp, err := rb.client.GetSnapShotHdrNode(ctx, req, newOpts...)
	if err != nil {
		return nil, err
	}
	peer := peerOpt.Peer()
	if resp.Node == nil {
		peer.Feedback(-2)
		return nil, errorz.ErrBadResponse
	}
	return resp.Node, nil
}

// RequestP2PGetBlockHeaders implements the client for the P2P method
// GetBlockHeaders
func (rb *Client) RequestP2PGetBlockHeaders(ctx context.Context, blockNums []uint32, opts ...grpc.CallOption) ([]*objs.BlockHeader, error) {
	req := &pb.GetBlockHeadersRequest{
		BlockNumbers: blockNums,
	}
	peerOpt := middleware.NewPeerInterceptor()
	newOpts := append(opts, peerOpt)
	resp, err := rb.client.GetBlockHeaders(ctx, req, newOpts...)
	if err != nil {
		return nil, err
	}
	peer := peerOpt.Peer()
	hdrs := []*objs.BlockHeader{}
	if len(resp.BlockHeaders) > len(blockNums) {
		peer.Feedback(-2)
		return nil, errorz.ErrBadResponse
	}
	for _, hdrbytes := range resp.BlockHeaders {
		hdr := &objs.BlockHeader{}
		err := hdr.UnmarshalBinary(utils.CopySlice(hdrbytes))
		if err != nil {
			peer.Feedback(-2)
			return nil, errorz.ErrBadResponse
		}
		hdrs = append(hdrs, hdr)
	}
	if hdrs == nil {
		utils.DebugTrace(rb.logger, err)
		peer.Feedback(-2)
		return nil, errorz.ErrBadResponse
	}
	return hdrs, nil
}

// RequestP2PGetPendingTx implements the client for the P2P method
// GetPendingTx
func (rb *Client) RequestP2PGetPendingTx(ctx context.Context, txHashes [][]byte, opts ...grpc.CallOption) ([][]byte, error) {
	req := &pb.GetPendingTxsRequest{
		TxHashes: txHashes,
	}
	peerOpt := middleware.NewPeerInterceptor()
	newOpts := append(opts, peerOpt)
	resp, err := rb.client.GetPendingTxs(ctx, req, newOpts...)
	if err != nil {
		return nil, err
	}
	peer := peerOpt.Peer()
	if resp.Txs == nil {
		peer.Feedback(-2)
		return nil, errorz.ErrBadResponse
	}
	if len(resp.Txs) == 0 {
		peer.Feedback(-2)
		utils.DebugTrace(rb.logger, err)
		return nil, errorz.ErrBadResponse
	}
	return resp.Txs, err
}

// RequestP2PGetMinedTxs implements the client for the P2P method
// GetMinedTxs
func (rb *Client) RequestP2PGetMinedTxs(ctx context.Context, txHashes [][]byte, opts ...grpc.CallOption) ([][]byte, error) {
	req := &pb.GetMinedTxsRequest{
		TxHashes: txHashes,
	}
	peerOpt := middleware.NewPeerInterceptor()
	newOpts := append(opts, peerOpt)
	resp, err := rb.client.GetMinedTxs(ctx, req, newOpts...)
	if err != nil {
		return nil, err
	}
	peer := peerOpt.Peer()
	if resp.Txs == nil {
		peer.Feedback(-2)
		return nil, errorz.ErrBadResponse
	}
	return resp.Txs, err
}

// RequestP2PGetSnapShotStateData implements the client for the P2P method
// GetSnapShotStateData
func (rb *Client) RequestP2PGetSnapShotStateData(ctx context.Context, key []byte, opts ...grpc.CallOption) ([]byte, error) {
	req := &pb.GetSnapShotStateDataRequest{
		Key: key,
	}
	peerOpt := middleware.NewPeerInterceptor()
	newOpts := append(opts, peerOpt)
	resp, err := rb.client.GetSnapShotStateData(ctx, req, newOpts...)
	if err != nil {
		return nil, err
	}
	peer := peerOpt.Peer()
	if resp.Data == nil {
		peer.Feedback(-2)
		utils.DebugTrace(rb.logger, err)
		return nil, errorz.ErrBadResponse
	}
	return resp.Data, err
}
