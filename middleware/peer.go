package middleware

import (
	"github.com/MadBase/MadNet/interfaces"
	"google.golang.org/grpc"
)

type PeerClient interface {
	interfaces.P2PClient
	Feedback(int)
}

type peerClient struct {
	PeerClient
}

type PeerCallOption struct {
	*grpc.EmptyCallOption
	Peer func() PeerClient
}

func (opt *PeerCallOption) CallOption() grpc.CallOption {
	return opt
}

func (opt *PeerCallOption) setPeer(p PeerClient) {
	opt.Peer = func() PeerClient {
		return p
	}
}

func NewPeerInterceptor() *PeerCallOption {
	opt := &PeerCallOption{EmptyCallOption: &grpc.EmptyCallOption{}, Peer: nil}
	return opt
}

func SetPeer(p PeerClient, opts ...grpc.CallOption) {
	for i := 0; i < len(opts); i++ {
		po, ok := opts[i].(*PeerCallOption)
		if ok {
			po.setPeer(p)
		}
	}
}
