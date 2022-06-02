package peering

import (
	"net"
	"sync"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/interfaces"
	pb "github.com/MadBase/MadNet/proto"
	"github.com/MadBase/MadNet/utils"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// ServerHandler binds a RPCListener to a grpc server such that the connections
// injected into the listener will be bound against the grpc server.
type ServerHandler struct {
	listener  interfaces.RPCListener
	server    *grpc.Server
	logger    *logrus.Logger
	closeOnce sync.Once
	name      string
}

// Close will shutdown the server handler.
func (rpch *ServerHandler) Close() error {
	fn := func() {
		rpch.logger.Debugf("Stopping %s server", rpch.name)
		go rpch.server.GracefulStop()
		rpch.logger.Debugf("Stop of server complete for %s", rpch.name)
		err := rpch.listener.Close()
		if err != nil {
			utils.DebugTrace(rpch.logger, err)
		}
	}
	rpch.closeOnce.Do(fn)
	return nil
}

// HandleConnection will inject the provided P2PConn into the NewConnection
// method of the bound RPCListener for consumption by the grpc server.
func (rpch *ServerHandler) HandleConnection(conn interfaces.P2PConn) error {
	err := rpch.listener.NewConnection(conn)
	if err != nil {
		utils.DebugTrace(rpch.logger, err)
		err2 := conn.Close()
		if err2 != nil {
			utils.DebugTrace(rpch.logger, err2)
		}
		return err
	}
	return nil
}

func (rpch *ServerHandler) serve() {
	defer rpch.Close()
	err := rpch.server.Serve(rpch.listener)
	if err != nil {
		utils.DebugTrace(rpch.logger, err)
	}
}

// NewP2PDiscoveryServerHandler returns a RPC ServerHandler for the Discovery
// Service.
func NewP2PDiscoveryServerHandler(logger *logrus.Logger, addr net.Addr, service interfaces.P2PDiscoveryServer) *ServerHandler {
	srvr := grpc.NewServer(grpc.ConnectionTimeout(constants.SrvrMsgTimeout), grpc.MaxConcurrentStreams(constants.MaxConcurrentStreams), grpc.NumStreamWorkers(constants.DiscoStreamWorkers), grpc.ReadBufferSize(constants.ReadBufferSize))
	pb.RegisterP2PDiscoveryServer(srvr, service)
	handler := &ServerHandler{
		listener: NewListener(logger, addr),
		server:   srvr,
		logger:   logger,
		name:     "DiscoveryServerHandler",
	}
	go handler.serve()
	return handler
}

// NewP2PServerHandler returns a RPC ServerHandler for the Pz2P Service.
func newP2PServerHandler(logger *logrus.Logger, addr net.Addr, service interfaces.P2PServer) *ServerHandler {
	srvr := grpc.NewServer(grpc.ConnectionTimeout(constants.SrvrMsgTimeout)) //, grpc.MaxConcurrentStreams(constants.P2PMaxConcurrentStreams), grpc.NumStreamWorkers(constants.P2PStreamWorkers)) //, grpc.ReadBufferSize(constants.ReadBufferSize))
	pb.RegisterP2PServer(srvr, service)
	handler := &ServerHandler{
		listener: NewListener(logger, addr),
		server:   srvr,
		logger:   logger,
		name:     "P2PServerHandler",
	}
	go handler.serve()
	return handler
}

// NewBootNodeServerHandler returns a RPC ServerHandler for the BootNode
// Service.
func NewBootNodeServerHandler(logger *logrus.Logger, addr net.Addr, service interfaces.BootNodeServer) *ServerHandler {
	srvr := grpc.NewServer(grpc.ConnectionTimeout(constants.SrvrMsgTimeout), grpc.MaxConcurrentStreams(constants.MaxConcurrentStreams), grpc.NumStreamWorkers(constants.P2PStreamWorkers), grpc.ReadBufferSize(constants.ReadBufferSize))
	pb.RegisterBootNodeServer(srvr, service)
	handler := &ServerHandler{
		listener: NewDiscoveryListener(logger, addr),
		server:   srvr,
		logger:   logger,
		name:     "BootNodeServerHandler",
	}
	go handler.serve()
	return handler
}
