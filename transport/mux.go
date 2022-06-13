package transport

import (
	"context"
	"sync"
	"time"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/types"
	"github.com/hashicorp/yamux"
	"github.com/sirupsen/logrus"
)

var _ interfaces.P2PMux = (*P2PMux)(nil)

var muxConfig *yamux.Config
var mlog *logrus.Logger

func init() {
	m := &P2PMux{}
	muxConfig = m.defaultConfig()
	mlog = logging.GetLogger(constants.LoggerTransport)
}

// P2PMux implements the multiplexing handshake protocol for P2PMuxConn
// construction.
type P2PMux struct {
	sync.Mutex
}

func (pmx *P2PMux) defaultConfig() *yamux.Config {
	// Define the configuration for the yamux multiplexer
	yamuxConfig := &yamux.Config{
		// AcceptBacklog is used to limit how many streams may be
		// waiting an accept.
		AcceptBacklog: 1,
		// EnableKeepalive is used to do periodic keep alive
		// messages using a ping.
		EnableKeepAlive: true,
		// KeepAliveInterval is how often to perform the keep alive
		KeepAliveInterval: time.Second * 5,
		// ConnectionWriteTimeout is meant to be a "safety valve" timeout after
		// which we will suspect a problem with the underlying connection and
		// close it. This is only applied to writes, where's there's generally
		// an expectation that things will move along quickly.
		ConnectionWriteTimeout: time.Second * 10,
		// MaxStreamWindowSize is used to control the maximum
		// window size that we allow for a stream.
		MaxStreamWindowSize: 262144,
		// LogOutput is an io.Writer that will be forwarded on to a logrus Logger at
		// the level specified ('WarnLevel'). Since it's just a writer we lose
		// fine grain levels but Yamux only logs at Warn and Error levels.
		LogOutput: logging.GetLogWriter(constants.LoggerYamux, logrus.DebugLevel),
	}
	return yamuxConfig
}

// HandleConnection runs the multiplexing protocol on the provided P2PConn and
// returns a multipled client and server P2PConn pair.
func (pmx *P2PMux) HandleConnection(ctx context.Context, conn interfaces.P2PConn) (interfaces.P2PMuxConn, error) {
	switch conn.Initiator() {
	case types.SelfInitiatedConnection:
		return pmx.clientMux(ctx, conn)
	case types.PeerInitiatedConnection:
		return pmx.serverMux(ctx, conn)
	default:
		panic("Unknown initiator type.")
	}
}

type muxresult struct {
	conn interfaces.P2PMuxConn
	err  error
}

func (pmx *P2PMux) serverMux(ctx context.Context, conn interfaces.P2PConn) (interfaces.P2PMuxConn, error) {
	rc := make(chan *muxresult)
	fn := func() {
		defer close(rc)
		session, err := yamux.Server(conn, muxConfig)
		if err != nil {
			conn.Close()
			rc <- &muxresult{nil, err}
			return
		}
		serverConn, err := session.Accept()
		if err != nil {
			conn.Close()
			rc <- &muxresult{nil, err}
			return
		}
		clientConn, err := session.Open()
		if err != nil {
			conn.Close()
			rc <- &muxresult{nil, err}
			return
		}
		clientp2pconn := &P2PConn{
			Conn:         clientConn,
			logger:       mlog,
			nodeAddr:     conn.NodeAddr(),
			protocol:     conn.Protocol(),
			protoVersion: conn.ProtoVersion(),
			initiator:    conn.Initiator(),
			session:      session,
			closeChan:    conn.CloseChan(),
			cleanupfn:    func() {},
		}
		serverp2pconn := &P2PConn{
			Conn:         serverConn,
			logger:       mlog,
			nodeAddr:     conn.NodeAddr(),
			protocol:     conn.Protocol(),
			protoVersion: conn.ProtoVersion(),
			initiator:    conn.Initiator(),
			session:      session,
			closeChan:    conn.CloseChan(),
			cleanupfn:    func() {},
		}
		muxconn := &P2PMuxConn{
			baseConn:   conn,
			session:    session,
			initiator:  conn.Initiator(),
			clientConn: clientp2pconn,
			serverConn: serverp2pconn,
			nodeAddr:   conn.NodeAddr(),
			closeChan:  conn.CloseChan(),
		}
		go muxconn.monitor()
		rc <- &muxresult{muxconn, nil}
	}
	drain := func() {
		result := <-rc
		if result.err == nil {
			if result.conn != nil {
				result.conn.Close()
			}
		}
	}
	go fn()
	select {
	case result := <-rc:
		return result.conn, result.err
	case <-ctx.Done():
		go drain()
		return nil, ErrHandshakeTimeout
	case <-conn.CloseChan():
		go drain()
		return nil, ErrHandshakeTimeout
	}
}

func (pmx *P2PMux) clientMux(ctx context.Context, conn interfaces.P2PConn) (interfaces.P2PMuxConn, error) {
	rc := make(chan *muxresult)
	fn := func() {
		defer close(rc)
		session, err := yamux.Client(conn, muxConfig)
		if err != nil {
			conn.Close()
			rc <- &muxresult{nil, err}
			return
		}
		clientConn, err := session.Open()
		if err != nil {
			conn.Close()
			rc <- &muxresult{nil, err}
			return
		}
		serverConn, err := session.Accept()
		if err != nil {
			conn.Close()
			rc <- &muxresult{nil, err}
			return
		}
		clientp2pconn := &P2PConn{
			Conn:         clientConn,
			initiator:    conn.Initiator(),
			logger:       mlog,
			protoVersion: conn.ProtoVersion(),
			nodeAddr:     conn.NodeAddr(),
			session:      session,
			closeChan:    conn.CloseChan(),
			cleanupfn:    func() {},
		}
		serverp2pconn := &P2PConn{
			Conn:         serverConn,
			protoVersion: conn.ProtoVersion(),
			logger:       mlog,
			initiator:    conn.Initiator(),
			nodeAddr:     conn.NodeAddr(),
			session:      session,
			closeChan:    conn.CloseChan(),
			cleanupfn:    func() {},
		}
		muxconn := &P2PMuxConn{
			baseConn:   conn,
			session:    session,
			clientConn: clientp2pconn,
			serverConn: serverp2pconn,
			initiator:  conn.Initiator(),
			nodeAddr:   conn.NodeAddr(),
			closeChan:  conn.CloseChan(),
		}
		go muxconn.monitor()
		rc <- &muxresult{muxconn, nil}
	}
	drain := func() {
		result := <-rc
		if result.err == nil {
			if result.conn != nil {
				result.conn.Close()
			}
		}
	}
	go fn()
	select {
	case result := <-rc:
		return result.conn, result.err
	case <-ctx.Done():
		go drain()
		return nil, ErrHandshakeTimeout
	case <-conn.CloseChan():
		go drain()
		return nil, ErrHandshakeTimeout
	}
}
