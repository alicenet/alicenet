package peering

import (
	"net"
	"sync"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/logging"
	pb "github.com/alicenet/alicenet/proto"
	"github.com/alicenet/alicenet/types"
	"github.com/alicenet/alicenet/utils"
	"github.com/sirupsen/logrus"
)

// MuxHandler allows a P2PMuxConn to be converted into a bidirectional grpc
// client and server connection. The server side connection is injected into an
// exiting grpc server running the P2P service and the client is bound against
// a the P2P service.
type MuxHandler struct {
	closeOnce sync.Once
	ch        *clientHandler
	sh        *ServerHandler
	logger    *logrus.Logger
}

// Close will shutdown the server handler.
func (rpcm *MuxHandler) Close() error {
	fn := func() {
		rpcm.ch.Close()
		err := rpcm.sh.Close()
		if err != nil {
			utils.DebugTrace(rpcm.logger, err)
		}
	}
	rpcm.closeOnce.Do(fn)
	return nil
}

// HandleConnection binds the P2PMuxConn to a grpc client and server.
// Internally this method uses the Initiator() method to determine if it should
// run the client or server side handshake. The only object returned is the
// P2PClient. The server side connection is handed off to the grpc server.
// Both the client and the server side connections may be shut down using the
// original P2PMuxConn Close method.
func (rpcm *MuxHandler) HandleConnection(conn interfaces.P2PMuxConn) (interfaces.P2PClient, error) {
	switch conn.Initiator() {
	case types.SelfInitiatedConnection:
		return rpcm.gRPCclientHandler(conn)
	case types.PeerInitiatedConnection:
		return rpcm.gRPCserverHandler(conn)
	default:
		panic("Unknown initiator in RPCHandshake")
	}
}

func (rpcm *MuxHandler) gRPCserverHandler(conn interfaces.P2PMuxConn) (interfaces.P2PClient, error) {
	//bind a client
	rpcclientconn, err := rpcm.ch.HandleConnection(conn.ClientConn())
	if err != nil {
		utils.DebugTrace(rpcm.logger, err)
		return nil, err
	}
	client := pb.NewP2PClient(rpcclientconn)
	//submit connection to server
	err = rpcm.sh.HandleConnection(conn.ServerConn())
	if err != nil {
		utils.DebugTrace(rpcm.logger, err)
		return nil, err
	}
	c := &p2PClient{
		logger:       logging.GetLogger(constants.LoggerPeerMan),
		P2PClientRaw: client,
		nodeAddr:     conn.NodeAddr(),
		conn:         conn,
	}
	return c, nil
}

func (rpcm *MuxHandler) gRPCclientHandler(conn interfaces.P2PMuxConn) (interfaces.P2PClient, error) {
	//submit connection to server
	err := rpcm.sh.HandleConnection(conn.ServerConn())
	if err != nil {
		utils.DebugTrace(rpcm.logger, err)
		return nil, err
	}
	//bind a client
	rpcclientconn, err := rpcm.ch.HandleConnection(conn.ClientConn())
	if err != nil {
		utils.DebugTrace(rpcm.logger, err)
		return nil, err
	}
	client := pb.NewP2PClient(rpcclientconn)
	c := &p2PClient{
		logger:       logging.GetLogger(constants.LoggerPeerMan),
		P2PClientRaw: client,
		nodeAddr:     conn.NodeAddr(),
		conn:         conn,
	}
	return c, nil
}

// NewMuxServerHandler creates a new multiplexed grpc tunneling system for
// P2PMuxConn objects.
func NewMuxServerHandler(logger *logrus.Logger, addr net.Addr, service interfaces.P2PServer) *MuxHandler {
	sh := newP2PServerHandler(logger, addr, service)
	ch := newClientHandler()
	return &MuxHandler{
		ch:     ch,
		sh:     sh,
		logger: logger,
	}
}
