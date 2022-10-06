package transport

import (
	"net"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/sirupsen/logrus"

	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/types"
	"github.com/alicenet/alicenet/utils"
)

var (
	_ net.Conn           = (*P2PConn)(nil)
	_ interfaces.P2PConn = (*P2PConn)(nil)
)

// The P2PConn type implements the net.Conn and the P2PConn interfaces.
// This is a concrete instantiation of the P2PConn object that all other
// objects of this system should accept as the connection object for network
// communications.
type P2PConn struct {
	conn         net.Conn
	logger       *logrus.Logger
	protocol     types.Protocol
	protoVersion types.ProtoVersion
	initiator    types.P2PInitiator
	nodeAddr     interfaces.NodeAddr
	closeOnce    sync.Once
	closeChan    <-chan struct{}
	cleanupfn    func()
	session      *yamux.Session
}

// implementation of net.Conn interface

func (pc *P2PConn) LocalAddr() net.Addr {
	return pc.conn.LocalAddr()
}

func (pc *P2PConn) Read(b []byte) (n int, err error) {
	return pc.conn.Read(b)
}

func (pc *P2PConn) SetDeadline(t time.Time) error {
	return pc.conn.SetDeadline(t)
}

func (pc *P2PConn) SetWriteDeadline(t time.Time) error {
	return pc.conn.SetWriteDeadline(t)
}

func (pc *P2PConn) Write(b []byte) (n int, err error) {
	return pc.conn.Write(b)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (pc *P2PConn) SetReadDeadline(t time.Time) error {
	return pc.conn.SetReadDeadline(t)
}

// CloseChan closes channel.
func (pc *P2PConn) CloseChan() <-chan struct{} {
	return pc.closeChan
}

// Close closes the connection.
func (pc *P2PConn) Close() error {
	pc.closeOnce.Do(pc.close)
	return nil
}

func (pc *P2PConn) close() {
	pc.cleanupfn()
	err := pc.conn.Close()
	if err != nil {
		utils.DebugTrace(pc.logger, err)
	}
}

// RemoteAddr See docs for net.Conn.
func (pc *P2PConn) RemoteAddr() net.Addr {
	return pc.nodeAddr
}

// Initiator defines if this is a locally initiated or peer initiated connection.
func (pc *P2PConn) Initiator() types.P2PInitiator {
	return pc.initiator
}

// RemoteP2PAddr returns the connections remote net.Addr as a P2PAddr.
func (pc *P2PConn) RemoteP2PAddr() interfaces.NodeAddr {
	return pc.nodeAddr
}

// ChainID identifies the chain this connection is expecting it's
// peers to also be a member of.
func (pc *P2PConn) ChainID() types.ChainIdentifier {
	return pc.nodeAddr.ChainID()
}

// Identity returns the hex string representation of the public key of the
// remote peer. This is a unique identifier to each node.
func (pc *P2PConn) Identity() string {
	return pc.nodeAddr.Identity()
}

// NodeAddr returns the address of the peer.
func (pc *P2PConn) NodeAddr() interfaces.NodeAddr {
	return pc.nodeAddr
}

// Protocol returns the protocol being used.
func (pc *P2PConn) Protocol() types.Protocol {
	return pc.protocol
}

// ProtoVersion returns the protocol version being used.
// This is not used at this time.
func (pc *P2PConn) ProtoVersion() types.ProtoVersion {
	return pc.protoVersion
}
