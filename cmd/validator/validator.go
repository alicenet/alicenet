package validator

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	_ "net/http/pprof"

	"github.com/MadBase/MadNet/application"
	"github.com/MadBase/MadNet/application/deposit"
	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/monitor"
	"github.com/MadBase/MadNet/cmd/utils"
	"github.com/MadBase/MadNet/config"
	"github.com/MadBase/MadNet/consensus"
	"github.com/MadBase/MadNet/consensus/admin"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/dman"
	"github.com/MadBase/MadNet/consensus/evidence"
	"github.com/MadBase/MadNet/consensus/gossip"
	"github.com/MadBase/MadNet/consensus/lstate"
	"github.com/MadBase/MadNet/consensus/request"
	"github.com/MadBase/MadNet/constants"
	hashlib "github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/localrpc"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/peering"
	"github.com/MadBase/MadNet/proto"
	"github.com/MadBase/MadNet/rbus"
	"github.com/MadBase/MadNet/status"
	mnutils "github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

// Command is the cobra.Command specifically for running as a node
var Command = cobra.Command{
	Use:   "validator",
	Short: "Starts a node",
	Long:  "Runs a MadNet node in mining or non-mining mode",
	Run:   validatorNode}

func validatorNode(cmd *cobra.Command, args []string) {
	// 	go func() {
	// 		log.Println(http.ListenAndServe("localhost:6060", nil))
	// 	}()

	//////////////////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////
	//SETUP LOGGING AND CONTEXT///////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////

	// create execution context for application
	ctx := context.Background()
	nodeCtx, cf := context.WithCancel(ctx)
	defer cf()

	// setup logger for program assembly operations
	logger := logging.GetLogger(cmd.Name())
	logger.Infof("Starting node with args %v", args)
	defer func() { logger.Warning("Goodbye.") }()

	//////////////////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////
	//INITIALIZE LOCAL CONFIG VARS////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////

	// these assignments are performed to keep an inventory of required
	// config values
	stateDbPath := config.Configuration.Chain.StateDbPath
	stateDbInMemory := config.Configuration.Chain.StateDbInMemory

	transactionDbPath := config.Configuration.Chain.TransactionDbPath
	transactionDbInMemory := config.Configuration.Chain.TransactionDbInMemory

	monitorDbPath := config.Configuration.Chain.MonitorDbPath
	monitorDbInMemory := config.Configuration.Chain.MonitorDbInMemory

	ethEndpoint := config.Configuration.Ethereum.Endpoint
	ethKeystore := config.Configuration.Ethereum.Keystore
	ethPasscodes := config.Configuration.Ethereum.Passcodes
	ethDefaultAccount := config.Configuration.Ethereum.DefaultAccount
	ethTimeout := config.Configuration.Ethereum.Timeout
	ethRetryCount := config.Configuration.Ethereum.RetryCount
	ethRetryDelay := config.Configuration.Ethereum.RetryDelay
	ethFinalityDelay := config.Configuration.Ethereum.FinalityDelay

	batchSize := config.Configuration.Monitor.BatchSize
	registryAddress := common.HexToAddress(config.Configuration.Ethereum.RegistryAddress)

	oneHour := 1 * time.Hour // TODO:ANTHONY - SHOULD THIS BE MOVED TO CONFIG?
	monitorInterval := config.Configuration.Monitor.Interval

	chainID := uint32(config.Configuration.Chain.ID)

	symK := hashlib.Hasher([]byte(config.Configuration.Validator.SymmetricKey))

	peerLimitMin := config.Configuration.Transport.PeerLimitMin
	peerLimitMax := config.Configuration.Transport.PeerLimitMax
	firewallMode := config.Configuration.Transport.FirewallMode
	firewallHost := config.Configuration.Transport.FirewallHost
	p2PListeningAddress := config.Configuration.Transport.P2PListeningAddress
	xportPrivateKey := config.Configuration.Transport.PrivateKey
	upnp := config.Configuration.Transport.UPnP

	lStateListenAddr := config.Configuration.Transport.LocalStateListeningAddress

	//////////////////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////
	//INITIALIZE ETHEREUM MONITORING//////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////

	// Ethereum connection setup
	logger.Infof("Connecting to Ethereum...")
	eth, err := blockchain.NewEthereumEndpoint(
		ethEndpoint,
		ethKeystore,
		ethPasscodes,
		ethDefaultAccount,
		ethTimeout,
		ethRetryCount,
		ethRetryDelay,
		ethFinalityDelay)
	if err != nil {
		logger.Fatalf("NewEthereumEndpoint(...) failed: %v", err)
		panic(err)
	}
	// Load the ethereum state
	if !eth.IsEthereumAccessible() {
		logger.Fatal("Ethereum endpoint not accessible...")
		panic(err)
	}
	logger.Infof("Looking up smart contracts on Ethereum...")
	// Find all the contracts
	if err := eth.Contracts().LookupContracts(registryAddress); err != nil {
		logger.Fatalf("Can't find contract registry: %v", err)
		panic(err)
	}
	utils.LogStatus(logger, eth)

	// Load accounts
	acct := eth.GetDefaultAccount()
	if err := eth.UnlockAccount(acct); err != nil {
		logger.Fatalf("Could not unlock account: %v", err)
		panic(err)
	}
	keys, err := eth.GetAccountKeys(acct.Address)
	if err != nil {
		logger.Fatalf("Could not get GetAccountKeys: %v", err)
		panic(err)
	}
	publicKey := crypto.FromECDSAPub(&keys.PrivateKey.PublicKey)
	logger.Infof("Account: %v Public Key: 0x%x", acct.Address.Hex(), publicKey)

	//////////////////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////
	//INITIALIZE DATABASE OBJECTS/////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////

	// Open consensus state db
	stateDb, err := mnutils.OpenBadger(
		nodeCtx.Done(),
		stateDbPath,
		stateDbInMemory,
	)
	if err != nil {
		panic(err)
	}
	defer stateDb.Close()

	// Open transaction pool db
	txnDb, err := mnutils.OpenBadger(
		nodeCtx.Done(),
		transactionDbPath,
		transactionDbInMemory,
	)
	if err != nil {
		panic(err)
	}
	defer txnDb.Close()

	// Open monitor database
	rawMonDb, err := mnutils.OpenBadger(
		nodeCtx.Done(),
		monitorDbPath,
		monitorDbInMemory,
	)
	if err != nil {
		panic(err)
	}
	defer rawMonDb.Close()
	monitorDb := monitor.NewDatabaseFromExisting(rawMonDb)

	//////////////////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////
	//CONSTRUCT CONSENSUS AND APPLICATION OBJECTS/////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////
	inboundRPCDispatch := proto.NewInboundRPCDispatch()
	stateRPCDispatch := proto.NewLocalStateDispatch()
	conDB := &db.Database{}
	pool := &evidence.Pool{}
	app := &application.Application{}
	ah := &admin.Handlers{}
	rbusHandlers := &request.Handler{}
	rbusClient := &request.Client{}
	lstateHandlers := &lstate.Handlers{}
	dman := &dman.DMan{}
	gh := &gossip.Handlers{}
	gc := &gossip.Client{}
	stateHandler := &lstate.Engine{}
	statusLogger := &status.Logger{}
	cesigner := &hashlib.Secp256k1Signer{}
	dph := &deposit.Handler{}
	sync := &consensus.Synchronizer{}
	stateRPCHandler := &localrpc.Handlers{}

	//////////////////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////
	//INITIALIZE CONSENSUS AND APPLICATION OBJECTS////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////

	// Initialize consensus database
	if err := conDB.Init(stateDb); err != nil {
		panic(err)
	}

	// Setup the peer manager
	peerManager, err := peering.NewPeerManager(
		proto.NewGeneratedP2PServer(inboundRPCDispatch),
		chainID,
		peerLimitMin,
		peerLimitMax,
		firewallMode,
		firewallHost,
		p2PListeningAddress,
		xportPrivateKey,
		upnp,
	)
	if err != nil {
		panic(err)
	}

	// Setup the local RPC server
	stateRPC, err := localrpc.NewStateServerHandler(
		logging.GetLogger(constants.LoggerTransport),
		lStateListenAddr,
		proto.NewGeneratedLocalStateServer(stateRPCDispatch),
	)
	if err != nil {
		panic(err)
	}

	// Initialize deposit handler
	if err := dph.Init(); err != nil {
		panic(err)
	}

	// Initialize the request bus client
	if err := rbusClient.Init(peerManager.P2PClient()); err != nil {
		panic(err)
	}

	// Initialize the consensus engine signer
	if err := cesigner.SetPrivk(crypto.FromECDSA(keys.PrivateKey)); err != nil {
		panic(err)
	}

	// Initialize the evidence pool
	if err := pool.Init(conDB); err != nil {
		panic(err)
	}

	// Initialize the app logic
	if err := app.Init(conDB, txnDb, dph); err != nil {
		panic(err)
	}

	// Initialize the request bus handler
	if err := rbusHandlers.Init(conDB, app); err != nil {
		panic(err)
	}

	// Initialize the download manager
	if err := dman.Init(conDB, app, rbusClient); err != nil {
		panic(err)
	}

	// Initialize the state handlers
	if err := lstateHandlers.Init(conDB, dman); err != nil {
		panic(err)
	}

	// Initialize the gossip bus handler
	if err := gh.Init(conDB, peerManager.P2PClient(), app, lstateHandlers); err != nil {
		panic(err)
	}

	// Initialize the gossip bus client
	if err := gc.Init(conDB, peerManager.P2PClient(), app); err != nil {
		panic(err)
	}

	// Initialize admin handler
	if err := ah.Init(chainID, conDB, symK, app, publicKey); err != nil {
		panic(err)
	}

	// Initialize the consensus engine
	if err := stateHandler.Init(conDB, dman, app, cesigner, ah, publicKey, rbusClient); err != nil {
		panic(err)
	}

	// Setup Request Bus Services
	svcs := monitor.NewServices(eth, conDB, dph, ah, batchSize, chainID)

	// Setup Request Bus
	mb, err := monitor.NewBus(rbus.NewRBus(), svcs)
	if err != nil {
		panic(err)
	}

	// Setup monitor
	mon, err := monitor.NewMonitor(monitorDb, mb, monitorInterval, oneHour)
	if err != nil {
		panic(err)
	}

	// Setup the synchronizer but prevent value log GC on in memory by setting
	// to nil - this will cause syncronizer to bypass GC on these databases
	var tDB, mDB *badger.DB
	if transactionDbInMemory {
		tDB = txnDb
	} else {
		tDB = nil
	}
	if monitorDbInMemory {
		mDB = rawMonDb
	} else {
		mDB = nil
	}
	if err := sync.Init(conDB, mDB, tDB, gc, gh, pool, stateHandler, app, ah, peerManager); err != nil {
		panic(err)
	}

	// Setup the local RPC server handler
	if err := stateRPCHandler.Init(conDB, app, gh, publicKey, sync.Safe); err != nil {
		panic(err)
	}

	// Initialize status logger
	if err := statusLogger.Init(stateHandler, peerManager, ah, mon); err != nil {
		panic(err)
	}

	// Register the inboundRPC handlers with the dispatch class
	inboundRPCDispatch.RegisterP2PGetPeers(peerManager)
	inboundRPCDispatch.RegisterP2PGossipTransaction(gh)
	inboundRPCDispatch.RegisterP2PGossipProposal(gh)
	inboundRPCDispatch.RegisterP2PGossipPreVote(gh)
	inboundRPCDispatch.RegisterP2PGossipPreVoteNil(gh)
	inboundRPCDispatch.RegisterP2PGossipPreCommit(gh)
	inboundRPCDispatch.RegisterP2PGossipPreCommitNil(gh)
	inboundRPCDispatch.RegisterP2PGossipNextRound(gh)
	inboundRPCDispatch.RegisterP2PGossipNextHeight(gh)
	inboundRPCDispatch.RegisterP2PGossipBlockHeader(gh)
	inboundRPCDispatch.RegisterP2PGetBlockHeaders(rbusHandlers)
	inboundRPCDispatch.RegisterP2PGetMinedTxs(rbusHandlers)
	inboundRPCDispatch.RegisterP2PGetPendingTxs(rbusHandlers)
	inboundRPCDispatch.RegisterP2PGetSnapShotNode(rbusHandlers)
	inboundRPCDispatch.RegisterP2PGetSnapShotStateData(rbusHandlers)
	inboundRPCDispatch.RegisterP2PGetSnapShotHdrNode(rbusHandlers)

	// Register the localState handlers with the dispatch class
	stateRPCDispatch.RegisterLocalStateGetBlockNumber(stateRPCHandler)
	stateRPCDispatch.RegisterLocalStateGetEpochNumber(stateRPCHandler)
	stateRPCDispatch.RegisterLocalStateGetBlockHeader(stateRPCHandler)
	stateRPCDispatch.RegisterLocalStateGetChainID(stateRPCHandler)
	stateRPCDispatch.RegisterLocalStateSendTransaction(stateRPCHandler)
	stateRPCDispatch.RegisterLocalStateGetValueForOwner(stateRPCHandler)
	stateRPCDispatch.RegisterLocalStateGetUTXO(stateRPCHandler)
	stateRPCDispatch.RegisterLocalStateGetMinedTransaction(stateRPCHandler)
	stateRPCDispatch.RegisterLocalStateGetPendingTransaction(stateRPCHandler)
	stateRPCDispatch.RegisterLocalStateGetRoundStateForValidator(stateRPCHandler)
	stateRPCDispatch.RegisterLocalStateGetValidatorSet(stateRPCHandler)
	stateRPCDispatch.RegisterLocalStateIterateNameSpace(stateRPCHandler)
	stateRPCDispatch.RegisterLocalStateGetData(stateRPCHandler)
	stateRPCDispatch.RegisterLocalStateGetTxBlockNumber(stateRPCHandler)

	//////////////////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////
	//LAUNCH ALL SERVICE GOROUTINES///////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////
	defer func() { os.Exit(0) }()
	defer func() { logger.Warning("Graceful unwind of core process complete.") }()

	go statusLogger.Run()
	defer statusLogger.Close()

	monitorCancelChan, err := mon.StartEventLoop()
	if err != nil {
		panic(err)
	}
	defer func() { monitorCancelChan <- true }()

	mb.StartLoop()
	defer mb.StopLoop()

	go peerManager.Start()
	defer peerManager.Close()

	go gc.Start()
	defer gc.Close()
	defer gh.Close()

	go dman.Start()
	defer dman.Close()

	go stateRPC.Serve()
	defer stateRPC.Close()

	go stateRPCHandler.Start()
	defer stateRPCHandler.Stop()

	//////////////////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////
	//SETUP SHUTDOWN MONITORING///////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	sync.Start()
	select {
	case <-peerManager.CloseChan():
	case <-sync.CloseChan():
	case <-signals:
	}
	go countSignals(logger, 5, signals)

	defer sync.Stop()
	defer func() { logger.Warning("Starting graceful unwind of core processes.") }()
}

// countSignals will cause a forced exit on repeated Ctrl+C commands
// this is a convient escape from a deadlock during shutdown
func countSignals(logger *logrus.Logger, num int, c chan os.Signal) {
	<-c
	for count := 0; count < num; count++ {
		logger.Warnf("Send Ctrl+C %v more times to force shutdown without waiting for services.\n", num-count)
		<-c
	}
	os.Exit(1)
}
