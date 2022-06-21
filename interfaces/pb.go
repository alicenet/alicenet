package interfaces

import (
	pb "github.com/alicenet/alicenet/proto"
)

// StateServer implements the State server service from the protobuf definition.
type StateServer interface {
	pb.LocalStateServer
}

// P2PClientRaw implements the P2P client service from the protobuf definition.
type P2PClientRaw interface {
	pb.P2PClient
}

// P2PClient implements the P2P client service from the protobuf definition
// and extends it for connection management.
type P2PClient interface {
	Close() error
	NodeAddr() NodeAddr
	CloseChan() <-chan struct{}
	pb.P2PClient
}

// P2PServer implements the P2P server service from the protobuf definition.
type P2PServer interface {
	pb.P2PServer
}

// P2PDiscoveryClient implements the Discovery client service from the
// protobuf definition.
type P2PDiscoveryClient interface {
	Close() error
	NodeAddr() NodeAddr
	pb.P2PDiscoveryClient
}

// P2PDiscoveryServer implements the Discovery server service from the
// protobuf definition.
type P2PDiscoveryServer interface {
	pb.P2PDiscoveryServer
}

// BootNodeServer implements the Discovery server service from the
// protobuf definition.
type BootNodeServer interface {
	pb.BootNodeServer
}
