package peering

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/alicenet/alicenet/interfaces"
	"github.com/sirupsen/logrus"
)

var _ interfaces.RPCListener = (*Listener)(nil)

// Listener allows a P2PConn to be converted into a net.Conn through the
// Accept method. This allows the injection of P2PConn objects into the
// grpc server through registering the Listener as the listener used by
// a grpc server.
type Listener struct {
	addr       net.Addr
	listenConn chan interfaces.P2PConn
	quit       chan struct{}
	log        *logrus.Logger
	closeOnce  sync.Once
}

// NewConnection allows a P2PConn to be injected into the Listener
// such that the goroutine calling Accept will recv the P2PConn as
// a net.Conn
func (rpcl *Listener) NewConnection(conn interfaces.P2PConn) error {
	select {
	case rpcl.listenConn <- conn:
		return nil
	case <-time.After(time.Second * 5):
		return errors.New("submission of new connection to RPC server timed out")
	case <-rpcl.quit:
		return errors.New("rpc server listener closed")
	}
}

// Addr allows Listener to implement net.Listener interface
func (rpcl *Listener) Addr() net.Addr {
	return rpcl.addr
}

// Close allows Listener to implement net.Listener interface
// Close also closes the Listener Accept method and raises an
// error to the caller of Accept.
func (rpcl *Listener) Close() error {
	fn := func() {
		close(rpcl.quit)
	}
	rpcl.closeOnce.Do(fn)
	return nil
}

// Accept allows Listener to implement net.Listener interface
// When called, Accept returns a net.Conn connection to the caller
// as new connections arrive.
func (rpcl *Listener) Accept() (net.Conn, error) {
	for {
		select {
		case p2pconn := <-rpcl.listenConn:
			return p2pconn, nil
		case <-rpcl.quit:
			return nil, errors.New("rpc server listener closed")
		}
	}
}

// NewListener returns a new Listener that conforms to the RPCListener
// interface.
func NewListener(logger *logrus.Logger, addr net.Addr) interfaces.RPCListener {
	return &Listener{
		listenConn: make(chan interfaces.P2PConn),
		addr:       addr,
		quit:       make(chan struct{}),
		log:        logger,
	}
}
