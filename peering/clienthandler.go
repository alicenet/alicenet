package peering

import (
	"context"
	"net"
	"time"

	"errors"

	"github.com/alicenet/alicenet/interfaces"
	"google.golang.org/grpc"
)

// ClientHandler is an object that allows a P2PConn to be converted into
// a *grpc.ClientConn. Once this conversion is accomplished, a grpc tunnel
// has been established over the P2PConn such that a grpc client service may
// be bound to the *grpc.ClientConn.
type clientHandler struct {
	closeChan chan struct{}
}

// Close will block further outbound dialing
func (rpcch *clientHandler) Close() {
	rpcch.closeChan = make(chan struct{})
	close(rpcch.closeChan)
}

// HandleConnection converts the P2PConn into a net.Conn, and injects the
// net.Conn into a grpc dialer using a closure. The returned connection from
// grpc.Dial is a ClientConn that may have grpc client services bound to the
// connection object.
func (rpcch *clientHandler) HandleConnection(p2pconn interfaces.P2PConn) (*grpc.ClientConn, error) {
	select {
	case <-rpcch.closeChan:
		p2pconn.Close()
		return nil, errors.New("closing")
	default:
	}
	// Setup contextDialer as a closure over connection
	contextDialer := func(ctx context.Context, a string) (net.Conn, error) {
		select {
		case <-ctx.Done():
			return nil, errors.New("context canceled")
		default:
			if p2pconn != nil {
				return p2pconn, nil
			}
			return nil, errors.New("connection is nil")
		}
	}
	conn, err := grpc.Dial(
		p2pconn.RemoteAddr().String(), // THIS WILL NEVER BE DIALED
		grpc.WithTimeout(time.Second*5),
		grpc.WithContextDialer(contextDialer),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithDisableRetry(),
		grpc.WithDisableHealthCheck(),
	)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, err
	}
	if conn == nil {
		return nil, errors.New("connection closed")
	}
	go func() {
		<-p2pconn.CloseChan()
		conn.Close()
	}()
	return conn, nil
}

// NewClientHandler creates a ClientHandler
func newClientHandler() *clientHandler {
	return &clientHandler{}
}
