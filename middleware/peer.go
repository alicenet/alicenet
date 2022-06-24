package middleware

import (
	"github.com/alicenet/alicenet/interfaces"
	"google.golang.org/grpc"
)

// PeerClient is an extension of the interfaces.P2PClient
// to include the Feedback method for use by Get type requests
// on the P2P methods.
type PeerClient interface {
	interfaces.P2PClient
	Feedback(int)
}

//nolint:unused,deadcode
type peerClient struct {
	PeerClient
}

// PeerCallOption allows a caller of a P2P Get method to receive a reference
// to the P2PClient that handled the call for the purpose of feedback
type PeerCallOption struct {
	*grpc.EmptyCallOption
	Peer func() PeerClient
}

// CallOption returns the PeerCallOption as a grpc.CallOption
func (opt *PeerCallOption) CallOption() grpc.CallOption {
	return opt
}

func (opt *PeerCallOption) setPeer(p PeerClient) {
	opt.Peer = func() PeerClient {
		return p
	}
}

// NewPeerInterceptor is a function that builds a grpc.CallOption that will
// return a peer ref to the caller
func NewPeerInterceptor() *PeerCallOption {
	opt := &PeerCallOption{EmptyCallOption: &grpc.EmptyCallOption{}, Peer: nil}
	return opt
}

// SetPeer allows the P2PBus to set the peer reference on a grpc.CallOption
func SetPeer(p PeerClient, opts ...grpc.CallOption) {
	for i := 0; i < len(opts); i++ {
		po, ok := opts[i].(*PeerCallOption)
		if ok {
			po.setPeer(p)
		}
	}
}
