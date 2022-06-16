package localrpc

import (
	"context"
	"embed"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/interfaces"
	pb "github.com/MadBase/MadNet/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

// Handler binds a Listener to a grpc server and sets up a RESTful API based on
// swagger. The Swagger API is a translation of the gRPC API. See the swagger
// docs for further details.
type Handler struct {
	cf         func()
	listener   net.Listener
	grpcServer *grpc.Server
	server     *http.Server
	log        *logrus.Logger
	closeOnce  sync.Once
}

//go:embed swagger/*
var swagger embed.FS

// Close will shutdown the server handler.
func (rpch *Handler) Close() error {
	fn := func() {
		rpch.cf()
		rpch.grpcServer.Stop()
		rpch.server.Close()
		rpch.listener.Close()
	}
	rpch.closeOnce.Do(fn)
	return nil
}

func (rpch *Handler) Serve() {
	defer rpch.Close()
	if err := rpch.server.Serve(rpch.listener); err != nil {
		rpch.log.Warn(err)
	}
}

// NewStateServerHandler returns a RPC ServerHandler for the BootNode
// Service.
func NewStateServerHandler(logger *logrus.Logger, addr string, service interfaces.StateServer) (*Handler, error) {
	// create the grpc server
	grpcServer := grpc.NewServer(grpc.MaxConcurrentStreams(constants.MaxConcurrentStreams), grpc.NumStreamWorkers(constants.LocalRPCMaxWorkers), grpc.ReadBufferSize(constants.ReadBufferSize))
	pb.RegisterLocalStateServer(grpcServer, service)

	// make a server mux
	mux := http.NewServeMux()

	// setup the swagger-ui fileserver using the embedded assets
	fileServer := http.FileServer(http.FS(swagger))

	// register the swagger fs handler with the server
	prefix := "/swagger/"
	mux.Handle(prefix, fileServer)
	// add redirect to file server
	mux.HandleFunc("/swagger.json", serveSwagger)

	// make a new grpc runtime mux
	gwmux := runtime.NewServeMux()

	// make grpc handle cors request
	cmux := cors.Default().Handler(gwmux)

	// assign the cors enabled grpc mux as a handler for the default route of sever mux
	mux.Handle("/", cmux)

	//create a context for grpc
	ctx := context.Background()
	subCtx, cf := context.WithCancel(ctx)
	// register state handler against http server
	err := pb.RegisterLocalStateHandlerServer(subCtx, gwmux, service)
	if err != nil {
		cf()
		return nil, err
	}

	// create the http server for the mux
	srv := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(grpcHandlerFunc(grpcServer, mux), &http2.Server{}),
	}

	// setup the listener
	lis, err := net.Listen("tcp", addr) // todo allow unix sockets
	if err != nil {
		cf()
		return nil, err
	}

	// build the handler object
	handler := &Handler{
		cf:         cf,
		listener:   lis,
		server:     srv,
		grpcServer: grpcServer,
		log:        logger,
	}

	// return the handler
	return handler, nil
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func serveSwagger(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "swagger/madnet.swagger.json", http.StatusMovedPermanently)
}
