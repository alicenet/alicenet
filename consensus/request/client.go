package request

import (
	"context"

	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/logging"
	pb "github.com/MadBase/MadNet/proto"
	"github.com/MadBase/MadNet/utils"
	"github.com/sirupsen/logrus"
)

// Client serves incoming requests and handles routing of outgoing
// requests for data from the consensus system.
type Client struct {
	peerSub  interfaces.PeerSubscription
	logger   *logrus.Logger
	secpVal  *crypto.Secp256k1Validator
	groupVal *crypto.BNGroupValidator
}

// Init initializes the object
func (rb *Client) Init(peerSub interfaces.PeerSubscription) error {
	rb.logger = logging.GetLogger(constants.LoggerConsensus)
	rb.peerSub = peerSub
	rb.groupVal = &crypto.BNGroupValidator{}
	rb.secpVal = &crypto.Secp256k1Validator{}
	return nil
}

func (rb *Client) RequestP2PGetSnapShotNode(ctx context.Context, height uint32, key []byte) ([]byte, error) {
	req := &pb.GetSnapShotNodeRequest{
		Height:   height,
		NodeHash: key,
	}
	var node []byte
	peerLease, err := rb.peerSub.PeerLease(ctx)
	if err != nil {
		if err == ctx.Err() {
			return nil, err
		}
		return nil, errorz.ErrClosing
	}
	var reqErr error

	fn := func(peer interfaces.PeerLease) error {
		client, err := peer.P2PClient()
		if err != nil {
			return err
		}
		resp, err := client.GetSnapShotNode(ctx, req)
		if err != nil {
			utils.DebugTrace(rb.logger, err)
			return err
		}
		node = resp.Node
		return nil
	}
	if reqErr != nil {
		return nil, err
	}
	peerLease.Do(fn)
	if node == nil {
		return nil, errorz.ErrBadResponse
	}
	return node, nil
}

func (rb *Client) RequestP2PGetSnapShotHdrNode(ctx context.Context, key []byte) ([]byte, error) {
	req := &pb.GetSnapShotHdrNodeRequest{
		NodeHash: key,
	}
	var node []byte
	peerLease, err := rb.peerSub.PeerLease(ctx)
	if err != nil {
		utils.DebugTrace(rb.logger, err)
		if err == ctx.Err() {
			return nil, err
		}
		return nil, errorz.ErrClosing
	}
	var reqErr error

	fn := func(peer interfaces.PeerLease) error {
		client, err := peer.P2PClient()
		if err != nil {
			utils.DebugTrace(rb.logger, err)
			return err
		}
		resp, err := client.GetSnapShotHdrNode(ctx, req)
		if err != nil {
			utils.DebugTrace(rb.logger, err)
			return err
		}
		node = resp.Node
		return nil
	}
	if reqErr != nil {
		return nil, err
	}
	peerLease.Do(fn)
	if node == nil {
		utils.DebugTrace(rb.logger, err)
		return nil, errorz.ErrBadResponse
	}
	return node, nil
}

func (rb *Client) RequestP2PGetBlockHeaders(ctx context.Context, blockNums []uint32) ([]*objs.BlockHeader, error) {
	req := &pb.GetBlockHeadersRequest{
		BlockNumbers: blockNums,
	}
	var hdrs []*objs.BlockHeader
	hsh := utils.MarshalUint32(blockNums[0])
	peerLease, err := rb.peerSub.RequestLease(ctx, hsh)
	if err != nil {
		utils.DebugTrace(rb.logger, err)
		if err == ctx.Err() {
			return nil, err
		}
		return nil, errorz.ErrClosing
	}
	byteCount := 0
	var reqErr error

	fn := func(peer interfaces.PeerLease) error {
		client, err := peer.P2PClient()
		if err != nil {
			utils.DebugTrace(rb.logger, err)
			return err
		}
		resp, err := client.GetBlockHeaders(ctx, req)
		if err != nil {
			utils.DebugTrace(rb.logger, err)
			return err
		}
		tmpHdrs := []*objs.BlockHeader{}
		if len(resp.BlockHeaders) > len(blockNums) {
			reqErr = errorz.ErrBadResponse
			return errorz.ErrInvalid{}.New("too many headers")
		}
		for _, hdrbytes := range resp.BlockHeaders {
			byteCount = byteCount + len(utils.CopySlice(hdrbytes))
			if byteCount > constants.MaxBytes {
				reqErr = errorz.ErrBadResponse
				return errorz.ErrInvalid{}.New("too big of hdr msg")
			}
			hdr := &objs.BlockHeader{}
			err := hdr.UnmarshalBinary(utils.CopySlice(hdrbytes))
			if err != nil {
				reqErr = errorz.ErrBadResponse
				return err
			}
			if err := hdr.ValidateSignatures(rb.groupVal); err != nil {
				reqErr = errorz.ErrBadResponse
				return errorz.ErrInvalid{}.New("bad signatures")
			}
			tmpHdrs = append(tmpHdrs, hdr)
		}
		hdrs = tmpHdrs
		return nil
	}
	if reqErr != nil {
		utils.DebugTrace(rb.logger, err)
		return nil, reqErr
	}
	peerLease.Do(fn)
	if hdrs == nil {
		utils.DebugTrace(rb.logger, err)
		return nil, errorz.ErrBadResponse
	}
	return hdrs, nil
}

func (rb *Client) RequestP2PGetPendingTx(ctx context.Context, txHashes [][]byte) ([][]byte, error) {
	req := &pb.GetPendingTxsRequest{
		TxHashes: txHashes,
	}
	var transactions [][]byte
	peerLease, err := rb.peerSub.RequestLease(ctx, txHashes[0])
	if err != nil {
		if err == ctx.Err() {
			return nil, err
		}
		return nil, errorz.ErrClosing
	}
	var reqErr error

	fn := func(peer interfaces.PeerLease) error {
		client, err := peer.P2PClient()
		if err != nil {
			return err
		}
		resp, err := client.GetPendingTxs(ctx, req)
		if err != nil {
			return err
		}
		transactions = resp.Txs
		return nil
	}
	if reqErr != nil {
		return nil, err
	}
	peerLease.Do(fn)
	if transactions == nil {
		return nil, errorz.ErrBadResponse
	}
	return transactions, nil
}

func (rb *Client) RequestP2PGetMinedTxs(ctx context.Context, txHashes [][]byte) ([][]byte, error) {
	req := &pb.GetMinedTxsRequest{
		TxHashes: txHashes,
	}
	var transactions [][]byte
	peerLease, err := rb.peerSub.PeerLease(ctx)
	if err != nil {
		if err == ctx.Err() {
			return nil, err
		}
		return nil, errorz.ErrClosing
	}
	var reqErr error

	fn := func(peer interfaces.PeerLease) error {
		client, err := peer.P2PClient()
		if err != nil {
			return err
		}
		resp, err := client.GetMinedTxs(ctx, req)
		if err != nil {
			return err
		}
		transactions = resp.Txs
		return nil
	}
	if reqErr != nil {
		return nil, err
	}
	peerLease.Do(fn)
	if transactions == nil {
		return nil, errorz.ErrBadResponse
	}
	return transactions, nil
}

func (rb *Client) RequestP2PGetSnapShotStateData(ctx context.Context, key []byte) ([]byte, error) {
	req := &pb.GetSnapShotStateDataRequest{
		Key: key,
	}
	var leaf []byte
	peerLease, err := rb.peerSub.PeerLease(ctx)
	if err != nil {
		if err == ctx.Err() {
			return nil, err
		}
		return nil, errorz.ErrClosing
	}
	var reqErr error

	fn := func(peer interfaces.PeerLease) error {
		client, err := peer.P2PClient()
		if err != nil {
			return err
		}
		resp, err := client.GetSnapShotStateData(ctx, req)
		if err != nil {
			return err
		}
		leaf = resp.Data
		return nil
	}
	if reqErr != nil {
		return nil, err
	}
	peerLease.Do(fn)
	if leaf == nil {
		return nil, errorz.ErrBadResponse
	}
	return leaf, nil
}
