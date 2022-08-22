package bootnode

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/peering"
	pb "github.com/alicenet/alicenet/proto"
	"github.com/alicenet/alicenet/transport"
	"github.com/alicenet/alicenet/types"
	lru "github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/peer"
)

// Command is the cobra.Command specifically for running as an edge node, i.e. not a validator or relay
var Command = cobra.Command{
	Use:   "bootnode",
	Short: "Starts a bootnode",
	Long:  "Boot nodes do nothing put seed the peer table",
	Run:   bootNode}

func extractPort(addr string) (uint32, error) {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, err
	}
	return uint32(port), nil
}

func bootNode(cmd *cobra.Command, args []string) {
	logger := logging.GetLogger(cmd.Name())

	logger.Infof("Bootnode started with args: %v", args)
	privateKeyHex := config.Configuration.Transport.PrivateKey
	cid := types.ChainIdentifier(config.Configuration.Chain.ID)
	listenerAddr := config.Configuration.BootNode.ListeningAddress
	p2pPort, err := extractPort(config.Configuration.BootNode.ListeningAddress)
	if err != nil {
		panic(err)
	}
	host, _, err := net.SplitHostPort(listenerAddr)
	if err != nil {
		panic(err)
	}
	// Establish P2P listener
	xport, err := transport.NewP2PTransport(logger, cid, privateKeyHex, int(p2pPort), host)
	if err != nil {
		logger.Panic(err)
	}
	defer xport.Close()

	// Register a boot node server
	cacheSize := config.Configuration.BootNode.CacheSize
	cache, err := lru.New(cacheSize)
	if err != nil {
		panic(err)
	}
	srvr := &Server{nodes: cache, log: logger}
	handler := peering.NewBootNodeServerHandler(logger, xport.NodeAddr(), srvr)
	defer handler.Close()

	localP2PAddr := xport.NodeAddr()
	logger.Infof("Starting bootnode with address: %s", localP2PAddr.P2PAddr())

	// Kick-off event loop
	acceptLoop(logger, xport, handler)
}

// Server implements the bootnode protocol
type Server struct {
	log   *logrus.Logger
	nodes *lru.Cache
}

// KnownNodes returns a set of recently seen peers when the bootnode is connected to
//goland:noinspection GoUnusedParameter
func (bn *Server) KnownNodes(ctx context.Context, r *pb.BootNodeRequest) (*pb.BootNodeResponse, error) {
	// get the identity of the caller
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, nil
	}
	// convert to p2p types
	caller := p.Addr.(interfaces.NodeAddr)
	callerAddr := caller.P2PAddr()
	callerIdent := caller.Identity()
	// defer the addition of the caller to the cache
	defer bn.nodes.ContainsOrAdd(callerIdent, callerAddr)
	// get the list of known identities from the cache
	identList := bn.nodes.Keys()
	bn.log.Debugf("Serving bootnode request to %s with %d known nodes", callerAddr, len(identList))
	// create a var to store output in
	var returnList []string
	nmap := make(map[string]bool)
	// add the caller to the known map
	nmap[callerIdent] = true
	// for each known identity
	for i := 0; i < len(identList); i++ {
		// convet to a string
		identif := identList[i]
		ident, ok := identif.(string)
		if !ok {
			// should never happen
			bn.log.Fatal("Bootnode ident type cast failed - must shut down")
			continue
		}
		// if the ident is already being tracked, continue
		// this ensures the caller is not added to the list
		if nmap[ident] {
			continue
		}
		addrif, ok := bn.nodes.Get(ident)
		if !ok {
			// should never happen
			bn.log.Fatal("Bootnode addr type cast failed - must shut down")
			continue
		}
		// convert the addr to a string
		addr := addrif.(string)
		// append to the return list
		returnList = append(returnList, addr)
		// set the ident in the known map
		nmap[ident] = true
	}
	resp := &pb.BootNodeResponse{
		Peers: returnList,
	}
	bn.log.Debugf("Sending response to bootnode request from %s with %s as known nodes", callerAddr, returnList)
	return resp, nil
}

func forceCleanup(conn interfaces.P2PConn) {
	select {
	case <-time.After(time.Second * 10):
		conn.Close()
		return
	case <-conn.CloseChan():
		return
	}
}

func acceptLoop(log *logrus.Logger, transport interfaces.P2PTransport, handler *peering.ServerHandler) {
	for {
		conn, err := transport.Accept()
		if err != nil {
			log.Error(err)
			return
		}
		// bind the connection to serve the request
		go handler.HandleConnection(conn) //nolint:errcheck
		// force drop the connection after 10 seconds
		go forceCleanup(conn)
	}
}
