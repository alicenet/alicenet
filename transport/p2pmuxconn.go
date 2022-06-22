package transport

import (
	"sync"

	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/types"
	"github.com/hashicorp/yamux"
)

var _ interfaces.P2PMuxConn = (*P2PMuxConn)(nil)

// P2PMuxConn object implements the P2PMuxConn interface.
// This object allows access to the multiplexed streams
// that are built on top of a single P2PConn.
type P2PMuxConn struct {
	closeOnce  sync.Once
	baseConn   interfaces.P2PConn
	session    *yamux.Session
	clientConn interfaces.P2PConn
	serverConn interfaces.P2PConn
	initiator  types.P2PInitiator
	nodeAddr   interfaces.NodeAddr
	closeChan  <-chan struct{}
}

// CloseChan returns a channel that will close when the connection begins
// closure
func (pmc *P2PMuxConn) CloseChan() <-chan struct{} {
	return pmc.closeChan
}

func (pmc *P2PMuxConn) monitor() {
	select {
	case <-pmc.session.CloseChan():
		pmc.Close()
	case <-pmc.closeChan:
		pmc.Close()
	}
}

// Close closes the multiplexed connection
func (pmc *P2PMuxConn) Close() error {
	pmc.closeOnce.Do(func() {
		go pmc.baseConn.Close()
		go pmc.serverConn.Close()
		go pmc.clientConn.Close()
		go pmc.session.Close()
	})
	return nil
}

// ServerConn gives access to server side of multiplexed connection
func (pmc *P2PMuxConn) ServerConn() interfaces.P2PConn {
	return pmc.serverConn
}

// ClientConn gives access to client side of multiplexed connection
func (pmc *P2PMuxConn) ClientConn() interfaces.P2PConn {
	return pmc.clientConn
}

// Initiator defines if this is a locally initiated or peer initiated connection
func (pmc *P2PMuxConn) Initiator() types.P2PInitiator {
	return pmc.initiator
}

// NodeAddr returns the address of the connection as a properly formatted
// string for use in dialing a peer.
func (pmc *P2PMuxConn) NodeAddr() interfaces.NodeAddr {
	return pmc.nodeAddr
}
