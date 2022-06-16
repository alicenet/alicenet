package transport

import (
	"net"
	"sync"

	"github.com/MadBase/MadNet/config"
	"github.com/MadBase/MadNet/crypto/secp256k1"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/transport/brontide"
	"github.com/MadBase/MadNet/types"
	"github.com/MadBase/MadNet/utils"
	"github.com/sirupsen/logrus"
)

var _ interfaces.P2PTransport = (*P2PTransport)(nil)

// This is the required value for the network string used in this package due to
// the design of brontide.
const (
	tcpNetwork   string             = "tcp"
	protoVersion types.ProtoVersion = 1
)

// P2PTransport wraps the brontide library in native types.
type P2PTransport struct {
	// This is the logger for the transport
	logger *logrus.Logger
	// This stores the listener address of the local node.
	localNodeAddr interfaces.NodeAddr
	// This is the private key used during encryption and authentication.
	localPrivateKey *secp256k1.PrivateKey
	// This is the brontide listener.
	listener *brontide.Listener
	// This is the quit notification channel for Accept loops.
	closeChan chan struct{}
	// this is the sync once used to protect the close methods
	closeOnce sync.Once
}

// Close will close all loops in this object and any
// subListeners. This method will also cause an error to be
// raised on both the Accept() and AcceptFailures() methods
// for any listening loops that are above this transport
// object. This method is safe to be called multiple times
// from different goroutines.
func (pt *P2PTransport) Close() error {
	fn := func() {
		close(pt.closeChan)
		go pt.listener.Close()
		pt.logger.Info("P2PTransport closed.")
	}
	go pt.closeOnce.Do(fn)
	return nil
}

// NodeAddr returns the listener address of the transport.
// This will not be the public ip address of the node,
// but rather this will be the interface the listener
// is bound to.
func (pt *P2PTransport) NodeAddr() interfaces.NodeAddr {
	return pt.localNodeAddr
}

// Dial will dial a remote peer at the specified address with the given
// protocol.
func (pt *P2PTransport) Dial(addr interfaces.NodeAddr, protocol types.Protocol) (interfaces.P2PConn, error) {
	// convert to raw type for access to non-interface methods
	remoteAddr := addr.(*NodeAddr)
	// convert p2pAddr into the expected format for brontide
	btcAddr := remoteAddr.toBTCNetAddr()
	// run the authentication and encryption handshake
	bconn, err := brontide.Dial(pt.localPrivateKey,
		protocol,
		protoVersion,
		pt.localNodeAddr.ChainID(),
		pt.localNodeAddr.Port(),
		btcAddr,
		func(network string, address string) (net.Conn, error) {
			return net.Dial(network, addr.String())
		})
	if err != nil {
		return nil, err
	}
	// convert from brontide connection into P2PConn
	return &P2PConn{
		nodeAddr: &NodeAddr{
			host:     addr.Host(),
			port:     bconn.P2PPort,
			chainID:  pt.localNodeAddr.ChainID(),
			identity: bconn.RemotePub(),
		},
		Conn:         bconn,
		logger:       pt.logger,
		initiator:    types.SelfInitiatedConnection,
		protocol:     bconn.Protocol,
		protoVersion: bconn.Version,
		cleanupfn:    func() {},
		closeChan:    bconn.CloseChan(),
	}, nil
}

// This method handles type conversions.
func (pt *P2PTransport) handleConnection(bconn *brontide.Conn) interfaces.P2PConn {
	// form a p2PAddr for the remote peer
	host, _, err := net.SplitHostPort(bconn.RemoteAddr().String())
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		if err := bconn.Close(); err != nil {
			utils.DebugTrace(pt.logger, err)
		}
		return nil
	}
	// turn the brontide conn into a p2PConn
	return &P2PConn{
		nodeAddr: &NodeAddr{
			host:     host,
			port:     bconn.P2PPort,
			chainID:  pt.localNodeAddr.ChainID(),
			identity: bconn.RemotePub(),
		},
		Conn:         bconn,
		logger:       pt.logger,
		initiator:    types.PeerInitiatedConnection,
		protocol:     bconn.Protocol,
		protoVersion: bconn.Version,
		cleanupfn:    func() {},
		closeChan:    bconn.CloseChan(),
	}
}

// Accept method MUST be called after the transport is started.
// This method will return connections that remote peers open
// to this node.
func (pt *P2PTransport) Accept() (interfaces.P2PConn, error) {
	for {
		select {
		case <-pt.closeChan:
			return nil, ErrListenerClosed
		default:
		}
		conn, err := pt.listener.Accept()
		if err != nil {
			if err == brontide.ErrBrontideClose {
				err2 := pt.Close()
				if err2 != nil {
					utils.DebugTrace(pt.logger, err2)
				}
				return nil, ErrListenerClosed
			}
			if err == brontide.ErrReject {
				continue
			}
			utils.DebugTrace(pt.logger, err)
			return nil, err
		}
		if conn == nil {
			continue
		}
		return pt.handleConnection(conn), nil
	}
}

// NewP2PTransport returns a transport object. This object is both a server
// and a client.
func NewP2PTransport(logger *logrus.Logger, cid types.ChainIdentifier, privateKeyHex string, port int, host string) (interfaces.P2PTransport, error) {
	localPrivateKey, err := deserializeTransportPrivateKey(privateKeyHex)
	if err != nil {
		return nil, err
	}
	localPublicKey := publicKeyFromPrivateKey(localPrivateKey)
	localNodeAddr := &NodeAddr{
		host:     host,
		port:     port,
		identity: localPublicKey,
		chainID:  cid,
	}

	var mc int
	var mp int
	if config.Configuration.Transport.OriginLimit <= 0 {
		mc = 3
	} else {
		mc = config.Configuration.Transport.OriginLimit
	}
	if config.Configuration.Transport.PeerLimitMax <= 0 {
		mp = 16
	} else {
		mp = config.Configuration.Transport.PeerLimitMax
	}

	listener, err := brontide.NewListener(localPrivateKey, host, port, protoVersion, cid, mp, 1, mc)
	if err != nil {
		return nil, err
	}

	transport := &P2PTransport{
		logger:          logger,
		localNodeAddr:   localNodeAddr,
		localPrivateKey: localPrivateKey,
		listener:        listener,
		closeChan:       make(chan struct{}),
	}
	return transport, nil
}
