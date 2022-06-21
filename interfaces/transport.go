package interfaces

import (
	"context"
	"net"

	"github.com/alicenet/alicenet/types"
)

// The NodeAddr interface implements net.Addr and other functionality needed for
// the identification of peer to peer clients. NodeAddr implements and extends
// the net.addr interface.
type NodeAddr interface {
	net.Addr
	Identity() string
	P2PAddr() string
	ChainID() types.ChainIdentifier
	Host() string
	Port() int
}

// P2PConn is the connection interface for the transport.
// This interface implements the net.Conn interface and
// extends the interface to allow introspection of the remote
// node identity as well as if this connection is a locally
// created connection through dial or if this connection is
// a remote initiated connection where the remote peer dialed
// the local node.
type P2PConn interface {
	net.Conn
	Initiator() types.P2PInitiator
	NodeAddr() NodeAddr
	Protocol() types.Protocol
	ProtoVersion() types.ProtoVersion
	CloseChan() <-chan struct{}
}

// P2PMuxConn is a multiplexed P2PConn as is returned by the P2PMuxTransport.
type P2PMuxConn interface {
	Initiator() types.P2PInitiator
	ClientConn() P2PConn
	ServerConn() P2PConn
	NodeAddr() NodeAddr
	CloseChan() <-chan struct{}
	Close() error
}

// P2PTransport is the interface that defines what the peer to peer
// transport object must conform to. This interface allows
// failed remote connections to be returned through the
// AcceptFailures() method. See ConnFailure interface for
// more information. The Accept() method will return new
// incoming connections from remote peers that have completed
// the initial handshake. The Dial() method allows remote peers
// to be connected to, by the local node.
type P2PTransport interface {
	NodeAddr() NodeAddr
	Accept() (P2PConn, error)
	Dial(NodeAddr, types.Protocol) (P2PConn, error)
	Close() error
}

// P2PMux handles the handshake protocol of the multiplexing protocol.
type P2PMux interface {
	HandleConnection(context.Context, P2PConn) (P2PMuxConn, error)
}

// RPCListener binds a P2PConn to a grpc server through connection injection.
// To inject a connection into the grpc server, you may pass the connection
// to NewConnection. This will make the connection available on the Accept
// method of the listener, which is intended to be consumed by the
// grpc.Server.Serve loop.
type RPCListener interface {
	net.Listener
	NewConnection(P2PConn) error
}
