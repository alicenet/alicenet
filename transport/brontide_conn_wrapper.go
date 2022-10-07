package transport

import (
	"github.com/alicenet/alicenet/crypto/secp256k1"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/types"
	"github.com/lightningnetwork/lnd/brontide"
	"github.com/lightningnetwork/lnd/lnwire"
	"net"
	"sync"
	"time"
)

type BrontideConnWrapper struct {
	conn      *brontide.Conn
	closeOnce sync.Once

	P2PPort  int
	Protocol types.Protocol
	Version  types.ProtoVersion

	closeFn   func() error
	closeChan chan struct{}

	publicKey *secp256k1.PublicKey
}

var _ net.Conn = (*BrontideConnWrapper)(nil)

func (bw *BrontideConnWrapper) wrapClose(newClose func() error) {
	bw.closeFn = newClose
}

// Close closes the connection.  Any blocked Read or Write operations will be
// unblocked and return errors.
//
// Part of the net.Conn interface.
func (bw *BrontideConnWrapper) Close() error {
	bw.closeOnce.Do(func() {
		close(bw.closeChan)
		go bw.conn.Close() //nolint:errcheck
		go bw.closeFn()    //nolint:errcheck
	})
	return nil
}

// CloseChan returns a channel that is closed when the connection closes.
func (bw *BrontideConnWrapper) CloseChan() <-chan struct{} {
	return bw.closeChan
}

func Dial(addr interfaces.NodeAddr, btcAddr *lnwire.NetAddress, localPrivateKey *secp256k1.PrivateKey, chainID types.ChainIdentifier, port int, protocol types.Protocol) (*BrontideConnWrapper, error) {
	// run the authentication and encryption handshake
	privateKey := convertCryptoPrivateKey2KeychainPrivateKey(localPrivateKey)
	conn, err := brontide.Dial(privateKey,
		btcAddr,
		10*time.Second,
		func(network, address string, timeout time.Duration) (net.Conn, error) {
			return net.Dial(network, addr.String())
		})
	if err != nil {
		return nil, err
	}

	if err := selfInitiatedChainIdentifierHandshake(conn, chainID); err != nil {
		conn.Close()
		return nil, err
	}

	if err := conn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		conn.Close()
		return nil, err
	}
	remoteP2PPort, err := selfInitiatedPortHandshake(conn, port)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err := conn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		conn.Close()
		return nil, err
	}
	remoteVersion, err := selfInitiatedVersionHandshake(conn, protoVersion)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err := writeUint32(conn, uint32(protocol)); err != nil {
		conn.Close()
		return nil, err
	}

	// We'll reset the deadline as it's no longer critical beyond the
	// initial handshake.
	err = conn.SetReadDeadline(time.Time{})
	if err != nil {
		conn.Close()
		return nil, err
	}

	publicKey, err := convertDecredSecpPubKey2CryptoSecpPubKey(conn.RemotePub())
	if err != nil {
		return nil, err
	}

	return &BrontideConnWrapper{
		conn:      conn,
		P2PPort:   remoteP2PPort,
		Version:   types.ProtoVersion(remoteVersion),
		Protocol:  protocol,
		closeFn:   conn.Close,
		closeChan: make(chan struct{}),
		closeOnce: sync.Once{},
		publicKey: publicKey,
	}, nil
}

// RemoteAddr See docs for net.Conn.
func (bw *BrontideConnWrapper) RemoteAddr() net.Addr {
	return bw.conn.RemoteAddr()
}

func (bw *BrontideConnWrapper) LocalAddr() net.Addr {
	return bw.conn.LocalAddr()
}

func (bw *BrontideConnWrapper) Read(b []byte) (n int, err error) {
	return bw.conn.Read(b)
}

func (bw *BrontideConnWrapper) SetDeadline(t time.Time) error {
	return bw.conn.SetDeadline(t)
}

func (bw *BrontideConnWrapper) SetWriteDeadline(t time.Time) error {
	return bw.conn.SetWriteDeadline(t)
}

func (bw *BrontideConnWrapper) Write(b []byte) (n int, err error) {
	return bw.conn.Write(b)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (bw *BrontideConnWrapper) SetReadDeadline(t time.Time) error {
	return bw.conn.SetReadDeadline(t)
}
