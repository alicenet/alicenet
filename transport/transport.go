package transport

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

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
		//protocol,
		//protoVersion,
		//pt.localNodeAddr.ChainID(),
		//pt.localNodeAddr.Port(),
		btcAddr,
		10*time.Second,
		func(network, address string, timeout time.Duration) (net.Conn, error) {
			return net.Dial(network, addr.String())
		})
	if err != nil {
		return nil, err
	}

	// todo: do other handshakes here
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

	// b.P2PPort = remoteP2PPort
	//b.Version = types.ProtoVersion(remoteVersion)
	//b.Protocol = protocol

	publicKey, err := convertDecredSecpPubKey2CryptoSecpPubKey(bconn.RemotePub())
	if err != nil {
		return nil, err
	}

	// todo: move it to a more appropriate place
	closeChan := make(chan struct{})
	go func() {
		pt.logger.Warn("waiting connection close from custom Dial.CloseChan 1")
		<-closeChan
		pt.logger.Warn("closing peer connection with custom Dial.CloseChan 2")
		//bconn.Close()
		//pt.logger.Warn("closing peer connection with custom Dial.CloseChan 3")
	}()

	return &P2PConn{
		nodeAddr: &NodeAddr{
			host:     addr.Host(),
			port:     remoteP2PPort,
			chainID:  pt.localNodeAddr.ChainID(),
			identity: publicKey,
		},
		conn:         bconn,
		logger:       pt.logger,
		initiator:    types.SelfInitiatedConnection,
		protocol:     protocol,
		protoVersion: types.ProtoVersion(remoteVersion),
		cleanupfn:    func() { close(closeChan) },
		closeChan:    closeChan,
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

	publicKey, err := convertDecredSecpPubKey2CryptoSecpPubKey(bconn.RemotePub())
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		if err := bconn.Close(); err != nil {
			utils.DebugTrace(pt.logger, err)
		}
		return nil
	}

	// todo: do other handshakes here

	// The following handshakes extend brontide to perform additional information
	// exchange.
	if err := bconn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		return nil
	}
	if err := peerInitiatedChainIdentifierHandshake(bconn, pt.localNodeAddr.ChainID()); err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		return nil
	}

	if err := bconn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		return nil
	}
	remoteP2PPort, err := peerInitiatedPortHandshake(bconn, pt.localNodeAddr.Port())
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		return nil
	}

	if err := bconn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		return nil
	}
	remoteVersion, err := peerInitiatedVersionHandshake(bconn, protoVersion)
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		return nil
	}

	if err := bconn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		return nil
	}
	protocol, err := readUint32(bconn)
	if err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		return nil
	}

	// reset read deadline
	if err := bconn.SetReadDeadline(time.Time{}); err != nil {
		utils.DebugTrace(pt.logger, err)
		err2 := bconn.Close()
		if err2 != nil {
			utils.DebugTrace(pt.logger, err2)
		}
		return nil
	}

	// brontideConn.P2PPort = remoteP2PPort
	// brontideConn.Version = types.ProtoVersion(remoteVersion)
	// brontideConn.Protocol = types.Protocol(protocol)

	// // get pubkey for limit pubkey tracking
	// pubk := string(bconn.RemotePub().SerializeCompressed())

	// // guard logic for pubkey limit tracking
	// if pt.numConnectionsbyPubkey[pubk] >= pt.pubkeyLimit {
	// 	err := bconn.Close()
	// 	if err != nil {
	// 		utils.DebugTrace(pt.logger, err)
	// 	}
	// 	return nil
	// }

	// // increment host counter
	// pt.numConnectionsbyPubkey[pubk]++

	// // if guard logic passes create cleanup fn closure
	// bconn.wrapClose(func() error {
	// 	pt.Lock()
	// 	defer pt.Unlock()
	// 	// decrement the origin counter if the total is gt zero
	// 	// this is a protection against any unseen race
	// 	if pt.numConnectionsbyPubkey[pubk] > 0 {
	// 		pt.numConnectionsbyPubkey[pubk]--
	// 		if pt.numConnectionsbyPubkey[pubk] == 0 {
	// 			delete(pt.numConnectionsbyPubkey, pubk)
	// 		}
	// 	}
	// 	return nil
	// })

	// todo: move it to a more appropriate place
	closeChan := make(chan struct{})
	go func() {
		pt.logger.Warn("waiting connection close from custom handleConnection.CloseChan 1")
		<-closeChan
		pt.logger.Warn("closing peer connection with custom handleConnection.CloseChan 2")
		//bconn.Close()
		//pt.logger.Warn("closing peer connection with custom handleConnection.CloseChan 3")
	}()

	// turn the brontide conn into a p2PConn
	return &P2PConn{
		nodeAddr: &NodeAddr{
			host:     host,
			port:     remoteP2PPort,
			chainID:  pt.localNodeAddr.ChainID(),
			identity: publicKey,
		},
		conn:         bconn,
		logger:       pt.logger,
		initiator:    types.PeerInitiatedConnection,
		protocol:     types.Protocol(protocol),
		protoVersion: types.ProtoVersion(remoteVersion),
		cleanupfn:    func() { close(closeChan) },
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

		return pt.handleConnection(bconn), nil
	}
}

// NewP2PTransport returns a transport object. This object is both a server
// and a client.
func NewP2PTransport(logger *logrus.Logger, cid types.ChainIdentifier, privateKeyHex string, port int, host string) (interfaces.P2PTransport, error) {
	// privateKeyBytes, err := hexutil.Decode(privateKeyHex)
	// if err != nil {
	// 	return nil, err
	// }

	//privateKey, publicKey := btcec.PrivKeyFromBytes(privateKeyBytes)

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

	// var mc int
	// var mp int
	// if config.Configuration.Transport.OriginLimit <= 0 {
	// 	mc = 3
	// } else {
	// 	mc = config.Configuration.Transport.OriginLimit
	// }
	// if config.Configuration.Transport.PeerLimitMax <= 0 {
	// 	mp = 16
	// } else {
	// 	mp = config.Configuration.Transport.PeerLimitMax
	// }

	listener, err := brontide.NewListener(keychainPrivateKey, fmt.Sprintf("%v:%v", host, port))
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
