package request

import (
	"context"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/middleware"
	"github.com/alicenet/alicenet/proto"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type P2PClientMock struct {
	mock.Mock
}

func (p2p *P2PClientMock) Close() error {
	return nil
}

func (p2p *P2PClientMock) NodeAddr() interfaces.NodeAddr {
	return nil
}

func (p2p *P2PClientMock) CloseChan() <-chan struct{} {
	return nil
}

func (p2p *P2PClientMock) Feedback(i int) {
}

func (p2p *P2PClientMock) Status(ctx context.Context, in *proto.StatusRequest, opts ...grpc.CallOption) (*proto.StatusResponse, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.StatusResponse), args.Error(1)
}
func (p2p *P2PClientMock) GetBlockHeaders(ctx context.Context, in *proto.GetBlockHeadersRequest, opts ...grpc.CallOption) (*proto.GetBlockHeadersResponse, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GetBlockHeadersResponse), args.Error(1)
}
func (p2p *P2PClientMock) GetMinedTxs(ctx context.Context, in *proto.GetMinedTxsRequest, opts ...grpc.CallOption) (*proto.GetMinedTxsResponse, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GetMinedTxsResponse), args.Error(1)
}
func (p2p *P2PClientMock) GetPendingTxs(ctx context.Context, in *proto.GetPendingTxsRequest, opts ...grpc.CallOption) (*proto.GetPendingTxsResponse, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GetPendingTxsResponse), args.Error(1)
}
func (p2p *P2PClientMock) GetSnapShotNode(ctx context.Context, in *proto.GetSnapShotNodeRequest, opts ...grpc.CallOption) (*proto.GetSnapShotNodeResponse, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GetSnapShotNodeResponse), args.Error(1)
}
func (p2p *P2PClientMock) GetSnapShotStateData(ctx context.Context, in *proto.GetSnapShotStateDataRequest, opts ...grpc.CallOption) (*proto.GetSnapShotStateDataResponse, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GetSnapShotStateDataResponse), args.Error(1)
}
func (p2p *P2PClientMock) GetSnapShotHdrNode(ctx context.Context, in *proto.GetSnapShotHdrNodeRequest, opts ...grpc.CallOption) (*proto.GetSnapShotHdrNodeResponse, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GetSnapShotHdrNodeResponse), args.Error(1)
}
func (p2p *P2PClientMock) GossipTransaction(ctx context.Context, in *proto.GossipTransactionMessage, opts ...grpc.CallOption) (*proto.GossipTransactionAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GossipTransactionAck), args.Error(1)
}
func (p2p *P2PClientMock) GossipProposal(ctx context.Context, in *proto.GossipProposalMessage, opts ...grpc.CallOption) (*proto.GossipProposalAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GossipProposalAck), args.Error(1)
}
func (p2p *P2PClientMock) GossipPreVote(ctx context.Context, in *proto.GossipPreVoteMessage, opts ...grpc.CallOption) (*proto.GossipPreVoteAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GossipPreVoteAck), args.Error(1)
}
func (p2p *P2PClientMock) GossipPreVoteNil(ctx context.Context, in *proto.GossipPreVoteNilMessage, opts ...grpc.CallOption) (*proto.GossipPreVoteNilAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GossipPreVoteNilAck), args.Error(1)
}
func (p2p *P2PClientMock) GossipPreCommit(ctx context.Context, in *proto.GossipPreCommitMessage, opts ...grpc.CallOption) (*proto.GossipPreCommitAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GossipPreCommitAck), args.Error(1)
}
func (p2p *P2PClientMock) GossipPreCommitNil(ctx context.Context, in *proto.GossipPreCommitNilMessage, opts ...grpc.CallOption) (*proto.GossipPreCommitNilAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GossipPreCommitNilAck), args.Error(1)
}
func (p2p *P2PClientMock) GossipNextRound(ctx context.Context, in *proto.GossipNextRoundMessage, opts ...grpc.CallOption) (*proto.GossipNextRoundAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GossipNextRoundAck), args.Error(1)
}
func (p2p *P2PClientMock) GossipNextHeight(ctx context.Context, in *proto.GossipNextHeightMessage, opts ...grpc.CallOption) (*proto.GossipNextHeightAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GossipNextHeightAck), args.Error(1)
}
func (p2p *P2PClientMock) GossipBlockHeader(ctx context.Context, in *proto.GossipBlockHeaderMessage, opts ...grpc.CallOption) (*proto.GossipBlockHeaderAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GossipBlockHeaderAck), args.Error(1)
}
func (p2p *P2PClientMock) GetPeers(ctx context.Context, in *proto.GetPeersRequest, opts ...grpc.CallOption) (*proto.GetPeersResponse, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	return args.Get(0).(*proto.GetPeersResponse), args.Error(1)
}
