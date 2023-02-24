package request

import (
	"context"

	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"

	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/middleware"
	"github.com/alicenet/alicenet/proto"
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
	protoStatus, ok := args.Get(0).(*proto.StatusResponse)
	if !ok {
		return nil, args.Error(1)
	}
	return protoStatus, args.Error(1)
}

func (p2p *P2PClientMock) GetBlockHeaders(ctx context.Context, in *proto.GetBlockHeadersRequest, opts ...grpc.CallOption) (*proto.GetBlockHeadersResponse, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoResponse, ok := args.Get(0).(*proto.GetBlockHeadersResponse)
	if !ok {
		panic("Unable to extract proto GetBlockHeadersResponse")
	}
	return protoResponse, args.Error(1)
}

func (p2p *P2PClientMock) GetMinedTxs(ctx context.Context, in *proto.GetMinedTxsRequest, opts ...grpc.CallOption) (*proto.GetMinedTxsResponse, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoResponse, ok := args.Get(0).(*proto.GetMinedTxsResponse)
	if !ok {
		panic("Unable to extract proto GetMinedTxsResponse")
	}
	return protoResponse, args.Error(1)
}

func (p2p *P2PClientMock) GetPendingTxs(ctx context.Context, in *proto.GetPendingTxsRequest, opts ...grpc.CallOption) (*proto.GetPendingTxsResponse, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoResponse, ok := args.Get(0).(*proto.GetPendingTxsResponse)
	if !ok {
		panic("Unable to extract proto GetMinedTxsRequest")
	}
	return protoResponse, args.Error(1)
}

func (p2p *P2PClientMock) GetSnapShotNode(ctx context.Context, in *proto.GetSnapShotNodeRequest, opts ...grpc.CallOption) (*proto.GetSnapShotNodeResponse, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoResponse, ok := args.Get(0).(*proto.GetSnapShotNodeResponse)
	if !ok {
		panic("Unable to extract proto GetSnapShotNodeResponse")
	}
	return protoResponse, args.Error(1)
}

func (p2p *P2PClientMock) GetSnapShotStateData(ctx context.Context, in *proto.GetSnapShotStateDataRequest, opts ...grpc.CallOption) (*proto.GetSnapShotStateDataResponse, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoResponse, ok := args.Get(0).(*proto.GetSnapShotStateDataResponse)
	if !ok {
		panic("Unable to extract proto GetSnapShotStateDataResponse")
	}
	return protoResponse, args.Error(1)
}

func (p2p *P2PClientMock) GetSnapShotHdrNode(ctx context.Context, in *proto.GetSnapShotHdrNodeRequest, opts ...grpc.CallOption) (*proto.GetSnapShotHdrNodeResponse, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoResponse, ok := args.Get(0).(*proto.GetSnapShotHdrNodeResponse)
	if !ok {
		panic("Unable to extract proto GetSnapShotHdrNodeResponse")
	}
	return protoResponse, args.Error(1)
}

func (p2p *P2PClientMock) GossipTransaction(ctx context.Context, in *proto.GossipTransactionMessage, opts ...grpc.CallOption) (*proto.GossipTransactionAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoAck, ok := args.Get(0).(*proto.GossipTransactionAck)
	if !ok {
		panic("Unable to extract proto GossipTransactionAck")
	}
	return protoAck, args.Error(1)
}

func (p2p *P2PClientMock) GossipProposal(ctx context.Context, in *proto.GossipProposalMessage, opts ...grpc.CallOption) (*proto.GossipProposalAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoAck, ok := args.Get(0).(*proto.GossipProposalAck)
	if !ok {
		panic("Unable to extract GossipProposalAck")
	}
	return protoAck, args.Error(1)
}

func (p2p *P2PClientMock) GossipPreVote(ctx context.Context, in *proto.GossipPreVoteMessage, opts ...grpc.CallOption) (*proto.GossipPreVoteAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoAck, ok := args.Get(0).(*proto.GossipPreVoteAck)
	if !ok {
		panic("Unable to extract GossipPreVoteAck")
	}
	return protoAck, args.Error(1)
}

func (p2p *P2PClientMock) GossipPreVoteNil(ctx context.Context, in *proto.GossipPreVoteNilMessage, opts ...grpc.CallOption) (*proto.GossipPreVoteNilAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoAck, ok := args.Get(0).(*proto.GossipPreVoteNilAck)
	if !ok {
		panic("Unable to extract GossipPreVoteNilAck")
	}
	return protoAck, args.Error(1)
}

func (p2p *P2PClientMock) GossipPreCommit(ctx context.Context, in *proto.GossipPreCommitMessage, opts ...grpc.CallOption) (*proto.GossipPreCommitAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoAck, ok := args.Get(0).(*proto.GossipPreCommitAck)
	if !ok {
		panic("Unable to extract GossipPreCommitAck")
	}
	return protoAck, args.Error(1)
}

func (p2p *P2PClientMock) GossipPreCommitNil(ctx context.Context, in *proto.GossipPreCommitNilMessage, opts ...grpc.CallOption) (*proto.GossipPreCommitNilAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoAck, ok := args.Get(0).(*proto.GossipPreCommitNilAck)
	if !ok {
		panic("Unable to extract GossipPreCommitNilAck")
	}
	return protoAck, args.Error(1)
}

func (p2p *P2PClientMock) GossipNextRound(ctx context.Context, in *proto.GossipNextRoundMessage, opts ...grpc.CallOption) (*proto.GossipNextRoundAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoAck, ok := args.Get(0).(*proto.GossipNextRoundAck)
	if !ok {
		panic("Unable to extract GossipNextRoundAck")
	}
	return protoAck, args.Error(1)
}

func (p2p *P2PClientMock) GossipNextHeight(ctx context.Context, in *proto.GossipNextHeightMessage, opts ...grpc.CallOption) (*proto.GossipNextHeightAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoAck, ok := args.Get(0).(*proto.GossipNextHeightAck)
	if !ok {
		panic("Unable to extract GossipNextHeightAck")
	}
	return protoAck, args.Error(1)
}

func (p2p *P2PClientMock) GossipBlockHeader(ctx context.Context, in *proto.GossipBlockHeaderMessage, opts ...grpc.CallOption) (*proto.GossipBlockHeaderAck, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoAck, ok := args.Get(0).(*proto.GossipBlockHeaderAck)
	if !ok {
		panic("Unable to extract GossipBlockHeaderAck")
	}
	return protoAck, args.Error(1)
}

func (p2p *P2PClientMock) GetPeers(ctx context.Context, in *proto.GetPeersRequest, opts ...grpc.CallOption) (*proto.GetPeersResponse, error) {
	middleware.SetPeer(p2p, opts...)
	args := p2p.Called(ctx, in, opts)
	protoResponse, ok := args.Get(0).(*proto.GetPeersResponse)
	if !ok {
		panic("Unable to extract GetPeersResponse")
	}
	return protoResponse, args.Error(1)
}
