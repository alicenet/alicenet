package interfaces

import (
	pb "github.com/MadBase/MadNet/proto"
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

// DiscoveryClient implements the Discovery client service from the
// protobuf definition.
type DiscoveryClient interface {
	Close() error
	NodeAddr() NodeAddr
	pb.DiscoveryClient
}

// DiscoveryServer implements the Discovery server service from the
// protobuf definition.
type DiscoveryServer interface {
	pb.DiscoveryServer
}

// BootNodeServer implements the Discovery server service from the
// protobuf definition.
type BootNodeServer interface {
	pb.BootNodeServer
}
