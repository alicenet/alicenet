package brontide

import (
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto/secp256k1"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/types"
	"github.com/MadBase/MadNet/utils"
)

// ErrReject is an error raised if a connection is rejected
var ErrReject = errors.New("unable to accept connection")

// ErrBrontideClose is an error raised if a connection is closed
var ErrBrontideClose = errors.New("brontide connection closed")

type connCloseWrapper struct {
	net.Conn
	closeFN func() error
}

func (c *connCloseWrapper) Close() error {
	return c.closeFN()
}

// Listener is an implementation of a net.Conn which executes an authenticated
// key exchange and message encryption protocol dubbed "Machine" after
// initial connection acceptance. See the Machine struct for additional
// details w.r.t the handshake and encryption scheme used within the
// connection.
type Listener struct {
	sync.Mutex
	logger *logrus.Logger

	localStatic *secp256k1.PrivateKey

	tcp *net.TCPListener

	conns chan maybeConn
	quit  chan struct{}

	// connection limiting
	numConnectionsbyIP     map[string]int
	numConnectionsbyPubkey map[string]int
	totalLimit             int
	originLimit            int
	pubkeyLimit            int

	// handshaking
	chainID      types.ChainIdentifier
	port         int
	protoVersion types.ProtoVersion
}

// NewListener returns a new net.Listener which enforces the Brontide scheme
// during both initial connection establishment and state transfer.
func NewListener(localStatic *secp256k1.PrivateKey, host string, port int, protoVersion types.ProtoVersion, chainID types.ChainIdentifier, totalLimit int, pubkeyLimit int, originLimit int) (*Listener, error) {
	listenAddr := net.JoinHostPort(host, strconv.Itoa(port))

	addr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return nil, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	brontideListener := &Listener{
		localStatic:            localStatic,
		tcp:                    l,
		logger:                 logging.GetLogger(constants.LoggerTransport),
		conns:                  make(chan maybeConn),
		quit:                   make(chan struct{}),
		numConnectionsbyIP:     make(map[string]int),
		numConnectionsbyPubkey: make(map[string]int),
		totalLimit:             totalLimit,
		originLimit:            originLimit,
		pubkeyLimit:            pubkeyLimit,
		port:                   port,
		protoVersion:           protoVersion,
		chainID:                chainID,
	}

	go brontideListener.listen()

	return brontideListener, nil
}

// rejectedConnErr is a helper function that prepends the remote address of the
// failed connection attempt to the original error message.
func rejectedConnErr(err error, remoteAddr string) error {
	return ErrReject
}

// listen accepts connection from the underlying tcp conn, then performs
// the brontinde handshake procedure asynchronously. A maximum of
// defaultHandshakes will be active at any given time.
//
// NOTE: This method must be run as a goroutine.
func (l *Listener) listen() {
	for {
		select {
		case <-l.quit:
			return
		default:
		}

		conn, err := l.tcp.Accept()
		if err != nil {
			select {
			case <-l.quit:
				return
			default:
				l.rejectConn(err)
				continue
			}
		}

		go l.preHandshake(conn)
	}
}

// preHandshake limits by connections by remote host
func (l *Listener) preHandshake(conn net.Conn) {
	l.Lock()
	defer l.Unlock()

	// bypass origin limiting if not a tcp conn
	addr, ok := conn.RemoteAddr().(*net.TCPAddr)
	if !ok {
		go l.doHandshake(conn)
		return
	}

	// guard logic for total connections
	if len(l.numConnectionsbyIP) >= l.totalLimit {
		err := conn.Close()
		if err != nil {
			utils.DebugTrace(l.logger, err)
		}
		return
	}

	// get host for origin limit tracking
	host := addr.IP.String()
	// guard logic for origin limit tracking
	if l.numConnectionsbyIP[host] >= l.originLimit {
		err := conn.Close()
		if err != nil {
			utils.DebugTrace(l.logger, err)
		}
		return
	}

	// increment host counter
	l.numConnectionsbyIP[host]++
	// create cleanup fn closure
	closeFn := func() error {
		l.Lock()
		defer l.Unlock()
		// decrement the origin counter
		if l.numConnectionsbyIP[host] > 0 {
			l.numConnectionsbyIP[host]--
			if l.numConnectionsbyIP[host] == 0 {
				delete(l.numConnectionsbyIP, host)
			}
		}
		return conn.Close()
	}
	// create wrapper with passed in closure
	c := &connCloseWrapper{conn, closeFn}
	// hand off wrapper as the conn
	go l.doHandshake(c)
}

func (l *Listener) postHandshake(conn *Conn) {
	l.Lock()
	defer l.Unlock()

	// get pubkey for limit pubkey tracking
	pubk := string(conn.RemotePub().SerializeCompressed())

	// guard logic for pubkey limit tracking
	if l.numConnectionsbyPubkey[pubk] >= l.pubkeyLimit {
		err := conn.Close()
		if err != nil {
			utils.DebugTrace(l.logger, err)
		}
		return
	}

	// increment host counter
	l.numConnectionsbyPubkey[pubk]++

	//if guard logic passes create cleanup fn closure
	conn.wrapClose(func() error {
		l.Lock()
		defer l.Unlock()
		// decrement the origin counter if the total is gt zero
		// this is a protection against any unseen race
		if l.numConnectionsbyPubkey[pubk] > 0 {
			l.numConnectionsbyPubkey[pubk]--
			if l.numConnectionsbyPubkey[pubk] == 0 {
				delete(l.numConnectionsbyPubkey, pubk)
			}
		}
		return nil
	})

	// hand off wrapper as the conn
	go l.acceptConn(conn)
}

// doHandshake asynchronously performs the brontide handshake, so that it does
// not block the main accept loop. This prevents peers that delay writing to the
// connection from block other connection attempts.
func (l *Listener) doHandshake(conn net.Conn) {
	select {
	case <-l.quit:
		return
	default:
	}

	remoteAddr := conn.RemoteAddr().String()

	brontideConn := &Conn{
		conn:      conn,
		noise:     NewBrontideMachine(false, l.localStatic, nil),
		closeChan: make(chan struct{}),
		closeOnce: sync.Once{},
		closeFn:   conn.Close,
	}

	// We'll ensure that we get ActOne from the remote peer in a timely
	// manner. If they don't respond within 1s, then we'll kill the
	// connection.
	if err := conn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}

	// Attempt to carry out the first act of the handshake protocol. If the
	// connecting node doesn't know our long-term static public key, then
	// this portion will fail with a non-nil error.
	var actOne [ActOneSize]byte
	if _, err := io.ReadFull(conn, actOne[:]); err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}
	if err := brontideConn.noise.RecvActOne(actOne); err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}

	// Next, progress the handshake processes by sending over our ephemeral
	// key for the session along with an authenticating tag.
	actTwo, err := brontideConn.noise.GenActTwo()
	if err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}
	if _, err := conn.Write(actTwo[:]); err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}

	select {
	case <-l.quit:
		return
	default:
	}

	// We'll ensure that we get ActTwo from the remote peer in a timely
	// manner. If they don't respond within 1 second, then we'll kill the
	// connection.
	if err := conn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}

	// Finally, finish the handshake processes by reading and decrypting
	// the connection peer's static public key. If this succeeds then both
	// sides have mutually authenticated each other.
	var actThree [ActThreeSize]byte
	if _, err := io.ReadFull(conn, actThree[:]); err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}
	if err := brontideConn.noise.RecvActThree(actThree); err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}

	select {
	case <-l.quit:
		return
	default:
	}

	// The following handshakes extend brontide to perform additional information
	// exchange.
	if err := conn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}
	if err := peerInitiatedChainIdentifierHandshake(brontideConn, l.chainID); err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}

	select {
	case <-l.quit:
		return
	default:
	}

	if err := conn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}
	remoteP2PPort, err := peerInitiatedPortHandshake(brontideConn, l.port)
	if err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}

	select {
	case <-l.quit:
		return
	default:
	}

	if err := conn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}
	remoteVersion, err := peerInitiatedVersionHandshake(brontideConn, l.protoVersion)
	if err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}

	select {
	case <-l.quit:
		return
	default:
	}

	if err := conn.SetReadDeadline(time.Now().Add(handshakeReadTimeout)); err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}
	protocol, err := readUint32(brontideConn)
	if err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}

	// reset read deadline
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		utils.DebugTrace(l.logger, err)
		err2 := brontideConn.Close()
		if err2 != nil {
			utils.DebugTrace(l.logger, err2)
		}
		l.rejectConn(rejectedConnErr(err, remoteAddr))
		return
	}

	brontideConn.P2PPort = remoteP2PPort
	brontideConn.Version = types.ProtoVersion(remoteVersion)
	brontideConn.Protocol = types.Protocol(protocol)

	go l.postHandshake(brontideConn)
}

// maybeConn holds either a brontide connection or an error returned from the
// handshake.
type maybeConn struct {
	conn *Conn
	err  error
}

// acceptConn returns a connection that successfully performed a handshake.
func (l *Listener) acceptConn(conn *Conn) {
	select {
	case l.conns <- maybeConn{conn: conn}:
	case <-l.quit:
	}
}

// rejectConn logs any errors encountered during connection or handshake.
func (l *Listener) rejectConn(err error) {
	select {
	case l.conns <- maybeConn{err: err}:
		utils.DebugTrace(l.logger, err)
	case <-l.quit:
	}
}

// Accept waits for and returns the next connection to the listener. All
// incoming connections are authenticated via the three act Brontide
// key-exchange scheme. This function will fail with a non-nil error in the
// case that either the handshake breaks down, or the remote peer doesn't know
// our static public key.
//
// Part of the net.Listener interface.
func (l *Listener) Accept() (*Conn, error) {
	select {
	case result := <-l.conns:
		return result.conn, result.err
	case <-l.quit:
		return nil, ErrBrontideClose
	}
}

// Close closes the listener.  Any blocked Accept operations will be unblocked
// and return errors.
//
// Part of the net.Listener interface.
func (l *Listener) Close() error {
	select {
	case <-l.quit:
	default:
		close(l.quit)
	}
	return l.tcp.Close()
}

// Addr returns the listener's network address.
//
// Part of the net.Listener interface.
func (l *Listener) Addr() net.Addr {
	return l.tcp.Addr()
}
