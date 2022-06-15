package interfaces

// Peer is an element of the peer tree.
// This interface allows inspection of both the peer and
// the peer meta state.
type Peer interface {
	NodeAddr() NodeAddr
	CloseChan() <-chan struct{}
	P2PClient() (P2PClient, error)
}
