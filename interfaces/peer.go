package interfaces

import "context"

// Peer is an element of the peer tree.
// This interface allows inspection of both the peer and
// the peer meta data.
type Peer interface {
	NodeAddr() NodeAddr
	CloseChan() <-chan struct{}
	Do(func(PeerLease) error)
	P2PClient() (P2PClient, error)
	PreventGossipConsensus([]byte)
	PreventGossipTx([]byte)
}

// PeerLease allows a service to obtain a Peer for sending data via closures
type PeerLease interface {
	P2PClient() (P2PClient, error)
	Do(func(PeerLease) error)
}

// PeerSubscription allows a service to maintain an in sync copy of all active
// peers for the purpose of requesting data or tracking sent/rcvd msgs
type PeerSubscription interface {
	CloseChan() <-chan struct{}
	PeerLease(ctx context.Context) (PeerLease, error)
	RequestLease(ctx context.Context, hsh []byte) (PeerLease, error)
	PreventGossipTx(addr NodeAddr, hsh []byte)
	PreventGossipConsensus(addr NodeAddr, hsh []byte)
	GossipConsensus(hsh []byte, fn func(context.Context, PeerLease) error)
	GossipTx(hsh []byte, fn func(context.Context, PeerLease) error)
}
