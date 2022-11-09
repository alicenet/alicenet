package transport

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/crypto/secp256k1"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/types"
	"github.com/alicenet/alicenet/utils"
	"github.com/lightningnetwork/lnd/brontide"
)

var _ interfaces.P2PTransport = (*P2PTransport)(nil)

// This is the required value for the network string used in this package due to
// the design of brontide.
const (
	tcpNetwork   string             = "tcp"
	protoVersion types.ProtoVersion = 1

	// handshakeReadTimeout is a read timeout that will be enforced when
	// waiting for state payloads during the various acts of Brontide. If
	// the remote party fails to deliver the proper payload within this
	// time frame, then we'll fail the connection.
	handshakeReadTimeout = time.Second * 5
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

	// connection limiting
	mutex                  sync.Mutex
	numConnectionsbyIP     map[string]int
	numConnectionsbyPubkey map[string]int
	totalLimit             int
	originLimit            int
	pubkeyLimit            int
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
	btcAddr, err := remoteAddr.toBTCNetAddr()
	if err != nil {
		return nil, err
	}
	// run the authentication and encryption handshake
	privateKey := convertCryptoPrivateKey2KeychainPrivateKey(pt.localPrivateKey)
	bconn, err := brontide.Dial(privateKey,
		btcAddr,
		10*time.Second,
		func(network, address string, timeout time.Duration) (net.Conn, error) {
			return net.Dial(network, addr.String())
		})
	if err != nil {
		return nil, err
	}

	// custom handshakes
	if err := selfInitiatedChainIdentifierHandshake(bconn, pt.localNodeAddr.ChainID()); err != nil {
		bconn.Close()
		return nil, err
	}

	if err := bconn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		bconn.Close()
		return nil, err
	}
	remoteP2PPort, err := selfInitiatedPortHandshake(bconn, pt.localNodeAddr.Port())
	if err != nil {
		bconn.Close()
		return nil, err
	}

	if err := bconn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		bconn.Close()
		return nil, err
	}
	remoteVersion, err := selfInitiatedVersionHandshake(bconn, protoVersion)
	if err != nil {
		bconn.Close()
		return nil, err
	}

	if err := writeUint32(bconn, uint32(protocol)); err != nil {
		bconn.Close()
		return nil, err
	}

	// We'll reset the deadline as it's no longer critical beyond the
	// initial handshake.
	err = bconn.SetReadDeadline(time.Time{})
	if err != nil {
		bconn.Close()
		return nil, err
	}

	publicKey, err := convertDecredSecpPubKey2CryptoSecpPubKey(bconn.RemotePub())
	if err != nil {
		return nil, err
	}

	closeChan := make(chan struct{})
	return &P2PConn{
		nodeAddr: &NodeAddr{
			host:     addr.Host(),
			port:     remoteP2PPort,
			chainID:  pt.localNodeAddr.ChainID(),
			identity: publicKey,
		},
		Conn:         bconn,
		logger:       pt.logger,
		initiator:    types.SelfInitiatedConnection,
		protocol:     protocol,
		protoVersion: types.ProtoVersion(remoteVersion),
		cleanupfn:    func() { close(closeChan) },
		closeChan:    closeChan,
	}, nil
}

// doAliceNetPreHandshake in order to administrate the connection limits and cleanups
func (pt *P2PTransport) doAliceNetPreHandshake(bconn *brontide.Conn) interfaces.P2PConn {
	pt.mutex.Lock()

	// bypass origin limiting if not a tcp conn
	addr, ok := bconn.RemoteAddr().(*net.TCPAddr)
	if !ok {
		pt.mutex.Unlock()
		return pt.doAliceNetHandshake(bconn, func() {})
	}

	// guard logic for total connections
	if len(pt.numConnectionsbyIP) >= pt.totalLimit {
		err := bconn.Close()
		if err != nil {
			utils.DebugTrace(pt.logger, err)
		}
		pt.mutex.Unlock()
		return nil
	}

	// get host for origin limit tracking
	host := addr.IP.String()
	// guard logic for origin limit tracking
	if pt.numConnectionsbyIP[host] >= pt.originLimit {
		err := bconn.Close()
		if err != nil {
			utils.DebugTrace(pt.logger, err)
		}
		pt.mutex.Unlock()
		return nil
	}

	// increment host counter
	pt.numConnectionsbyIP[host]++
	// create cleanup fn closure
	closeFn := func() {
		pt.mutex.Lock()
		defer pt.mutex.Unlock()
		// decrement the origin counter
		if pt.numConnectionsbyIP[host] > 0 {
			pt.numConnectionsbyIP[host]--
			if pt.numConnectionsbyIP[host] == 0 {
				delete(pt.numConnectionsbyIP, host)
			}
		}
	}

	pt.mutex.Unlock()

	// hand off wrapper as the conn
	return pt.doAliceNetHandshake(bconn, closeFn)
}

// doAliceNetHandshake with additional custom handshakes and keep track of connections and limits
func (pt *P2PTransport) doAliceNetHandshake(bconn *brontide.Conn, closeFn func()) interfaces.P2PConn {
	// The following handshakes extend brontide to perform additional information
	// exchange.
	if err := bconn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		closeFn()
		return nil
	}
	if err := peerInitiatedChainIdentifierHandshake(bconn, pt.localNodeAddr.ChainID()); err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		closeFn()
		return nil
	}

	if err := bconn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		closeFn()
		return nil
	}
	remoteP2PPort, err := peerInitiatedPortHandshake(bconn, pt.localNodeAddr.Port())
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		closeFn()
		return nil
	}

	if err := bconn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		closeFn()
		return nil
	}
	remoteVersion, err := peerInitiatedVersionHandshake(bconn, protoVersion)
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		closeFn()
		return nil
	}

	if err := bconn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		closeFn()
		return nil
	}
	protocol, err := readUint32(bconn)
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		closeFn()
		return nil
	}

	// reset read deadline
	if err := bconn.SetReadDeadline(time.Time{}); err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		closeFn()
		return nil
	}

	// form a p2PAddr for the remote peer
	host, _, err := net.SplitHostPort(bconn.RemoteAddr().String())
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		if err := bconn.Close(); err != nil {
			utils.DebugTrace(pt.logger, err)
		}
		closeFn()
		return nil
	}

	publicKey, err := convertDecredSecpPubKey2CryptoSecpPubKey(bconn.RemotePub())
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		if err := bconn.Close(); err != nil {
			utils.DebugTrace(pt.logger, err)
		}
		closeFn()
		return nil
	}

	// post handshake

	// get pubkey for limit pubkey tracking
	pubk := string(bconn.RemotePub().SerializeCompressed())

	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	// guard logic for pubkey limit tracking
	if pt.numConnectionsbyPubkey[pubk] >= pt.pubkeyLimit {
		err := bconn.Close()
		if err != nil {
			utils.DebugTrace(pt.logger, err)
		}
		closeFn()
		return nil
	}

	// increment host counter
	pt.numConnectionsbyPubkey[pubk]++

	// if guard logic passes create cleanup fn closure
	closeChan := make(chan struct{})
	cleanupFn := func() {
		pt.mutex.Lock()

		// decrement the origin counter if the total is gt zero
		// this is a protection against any unseen race
		if pt.numConnectionsbyPubkey[pubk] > 0 {
			pt.numConnectionsbyPubkey[pubk]--
			if pt.numConnectionsbyPubkey[pubk] == 0 {
				delete(pt.numConnectionsbyPubkey, pubk)
			}
		}

		pt.mutex.Unlock()

		closeFn()
		close(closeChan)
	}

	// turn the brontide conn into a p2PConn
	return &P2PConn{
		nodeAddr: &NodeAddr{
			host:     host,
			port:     remoteP2PPort,
			chainID:  pt.localNodeAddr.ChainID(),
			identity: publicKey,
		},
		Conn:         bconn,
		logger:       pt.logger,
		initiator:    types.PeerInitiatedConnection,
		protocol:     types.Protocol(protocol),
		protoVersion: types.ProtoVersion(remoteVersion),
		cleanupfn:    cleanupFn,
		closeChan:    closeChan,
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
			err2 := pt.Close()
			if err2 != nil {
				utils.DebugTrace(pt.logger, err2)
			}
			return nil, err
		}
		if conn == nil {
			continue
		}

		bconn := conn.(*brontide.Conn)

		return pt.doAliceNetPreHandshake(bconn), nil
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

	keychainPrivateKey := convertCryptoPrivateKey2KeychainPrivateKey(localPrivateKey)

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

	listener, err := brontide.NewListener(keychainPrivateKey, fmt.Sprintf("%v:%v", host, port))
	if err != nil {
		return nil, err
	}

	transport := &P2PTransport{
		logger:                 logger,
		localNodeAddr:          localNodeAddr,
		localPrivateKey:        localPrivateKey,
		listener:               listener,
		closeChan:              make(chan struct{}),
		totalLimit:             mp,
		pubkeyLimit:            1,
		originLimit:            mc,
		numConnectionsbyIP:     make(map[string]int),
		numConnectionsbyPubkey: make(map[string]int),
	}
	return transport, nil
}
