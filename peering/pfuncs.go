package peering

import (
	"context"

	"github.com/MadBase/MadNet/interfaces"
)

// GossipFunc is a function to be passed in that allows the
// local node to gossip outbound messages.
// Use peerManager.Gossip
type GossipFunc func(fn func(interfaces.PeerLease) error)

// PeerLeaseFunc is a function that takes in a context and returns
// a PeerLease object. The PeerLease object represents a peer with
// an active P2P connection. This object may be used to invoke
// requests against a remote peer.
type PeerLeaseFunc func(ctx context.Context) (interfaces.PeerLease, error)
