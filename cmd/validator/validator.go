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
	mncrypto "github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/localrpc"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/peering"
	"github.com/MadBase/MadNet/proto"
	"github.com/MadBase/MadNet/rbus"
	"github.com/MadBase/MadNet/status"
	mnutils "github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts/keystore"
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

func initEthereumConnection(logger *logrus.Logger) (blockchain.Ethereum, *keystore.Key, []byte) {
	// Ethereum connection setup
	logger.Infof("Connecting to Ethereum...")
	eth, err := blockchain.NewEthereumEndpoint(
		config.Configuration.Ethereum.Endpoint,
		config.Configuration.Ethereum.Keystore,
		config.Configuration.Ethereum.Passcodes,
		config.Configuration.Ethereum.DefaultAccount,
		config.Configuration.Ethereum.Timeout,
		config.Configuration.Ethereum.RetryCount,
		config.Configuration.Ethereum.RetryDelay,
		config.Configuration.Ethereum.FinalityDelay)

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

	registryAddress := common.HexToAddress(config.Configuration.Ethereum.RegistryAddress)
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

	return eth, keys, publicKey
}

// Setup the peer manager:
// Peer manager owns the raw TCP connections of the p2p system, responsible for turning on and off p2p connections
// 	- gossip protocol
// 	- way to access methods on a remote peer (another node)
// 	- validators/miners, those who care about voting and consensus
func initPeerManager(consGossipHandlers *gossip.Handlers, consReqHandler *request.Handler) *peering.PeerManager {
	p2pDispatch := proto.NewP2PDispatch()

	peerManager, err := peering.NewPeerManager(
		proto.NewGeneratedP2PServer(p2pDispatch),
		uint32(config.Configuration.Chain.ID),
		config.Configuration.Transport.PeerLimitMin,
		config.Configuration.Transport.PeerLimitMax,
		config.Configuration.Transport.FirewallMode,
		config.Configuration.Transport.FirewallHost,
		config.Configuration.Transport.P2PListeningAddress,
		config.Configuration.Transport.PrivateKey,
		config.Configuration.Transport.UPnP)
	if err != nil {
		panic(err)
	}
	p2pDispatch.RegisterP2PGetPeers(peerManager)
	p2pDispatch.RegisterP2PGossipTransaction(consGossipHandlers)
	p2pDispatch.RegisterP2PGossipProposal(consGossipHandlers)
	p2pDispatch.RegisterP2PGossipPreVote(consGossipHandlers)
	p2pDispatch.RegisterP2PGossipPreVoteNil(consGossipHandlers)
	p2pDispatch.RegisterP2PGossipPreCommit(consGossipHandlers)
	p2pDispatch.RegisterP2PGossipPreCommitNil(consGossipHandlers)
	p2pDispatch.RegisterP2PGossipNextRound(consGossipHandlers)
	p2pDispatch.RegisterP2PGossipNextHeight(consGossipHandlers)
	p2pDispatch.RegisterP2PGossipBlockHeader(consGossipHandlers)
	p2pDispatch.RegisterP2PGetBlockHeaders(consReqHandler)
	p2pDispatch.RegisterP2PGetMinedTxs(consReqHandler)
	p2pDispatch.RegisterP2PGetPendingTxs(consReqHandler)
	p2pDispatch.RegisterP2PGetSnapShotNode(consReqHandler)
	p2pDispatch.RegisterP2PGetSnapShotStateData(consReqHandler)
	p2pDispatch.RegisterP2PGetSnapShotHdrNode(consReqHandler)

	return peerManager
}

// Setup the localstate RPC server, used by e.g. wallet users (or anything that's not a node)
func initLocalStateServer(localStateHandler *localrpc.Handlers) *localrpc.Handler {
	localStateDispatch := proto.NewLocalStateDispatch()
	localStateServer, err := localrpc.NewStateServerHandler(
		logging.GetLogger(constants.LoggerTransport),
		config.Configuration.Transport.LocalStateListeningAddress,
		proto.NewGeneratedLocalStateServer(localStateDispatch),
	)
	if err != nil {
		panic(err)
	}
	localStateDispatch.RegisterLocalStateGetBlockNumber(localStateHandler)
	localStateDispatch.RegisterLocalStateGetEpochNumber(localStateHandler)
	localStateDispatch.RegisterLocalStateGetBlockHeader(localStateHandler)
	localStateDispatch.RegisterLocalStateGetChainID(localStateHandler)
	localStateDispatch.RegisterLocalStateSendTransaction(localStateHandler)
	localStateDispatch.RegisterLocalStateGetValueForOwner(localStateHandler)
	localStateDispatch.RegisterLocalStateGetUTXO(localStateHandler)
	localStateDispatch.RegisterLocalStateGetMinedTransaction(localStateHandler)
	localStateDispatch.RegisterLocalStateGetPendingTransaction(localStateHandler)
	localStateDispatch.RegisterLocalStateGetRoundStateForValidator(localStateHandler)
	localStateDispatch.RegisterLocalStateGetValidatorSet(localStateHandler)
	localStateDispatch.RegisterLocalStateIterateNameSpace(localStateHandler)
	localStateDispatch.RegisterLocalStateGetData(localStateHandler)
	localStateDispatch.RegisterLocalStateGetTxBlockNumber(localStateHandler)

	return localStateServer
}

func initDatabase(ctx *context.Context, path string, inMemory bool) *badger.DB {
	db, err := mnutils.OpenBadger(
		(*ctx).Done(),
		path,
		inMemory,
	)
	if err != nil {
		panic(err)
	}
	return db
}

func validatorNode(cmd *cobra.Command, args []string) {
	// create execution context for application
	ctx := context.Background()
	nodeCtx, cf := context.WithCancel(ctx)
	defer cf()

	// setup logger for program assembly operations
	logger := logging.GetLogger(cmd.Name())
	logger.Infof("Starting node with args %v", args)
	defer func() { logger.Warning("Goodbye.") }()

	batchSize := config.Configuration.Monitor.BatchSize

	monitorInterval := config.Configuration.Monitor.Interval

	chainID := uint32(config.Configuration.Chain.ID)

	eth, keys, publicKey := initEthereumConnection(logger)

	// Initialize consensus db: stores all state the consensus mechanism requires to work
	rawConsensusDb := initDatabase(&nodeCtx, config.Configuration.Chain.StateDbPath, config.Configuration.Chain.StateDbInMemory)
	defer rawConsensusDb.Close()

	// Initialize transaction pool db: contains transcations that have not been mined (and thus are to be gossiped)
	rawTxPoolDb := initDatabase(&nodeCtx, config.Configuration.Chain.TransactionDbPath, config.Configuration.Chain.TransactionDbInMemory)
	defer rawTxPoolDb.Close()

	// Initialize monitor database: tracks what ETH block number we're on (tracking deposits)
	rawMonitorDb := initDatabase(&nodeCtx, config.Configuration.Chain.MonitorDbPath, config.Configuration.Chain.MonitorDbInMemory)
	defer rawMonitorDb.Close()
	monitorDb := monitor.NewDatabaseFromExisting(rawMonitorDb)

	consDB := &db.Database{}
	consTxPool := &evidence.Pool{}
	app := &application.Application{}
	consAdminHandlers := &admin.Handlers{}
	consReqHandler := &request.Handler{}
	consReqClient := &request.Client{}
	consLSHandler := &lstate.Handlers{}
	consDlManager := &dman.DMan{}
	consGossipHandlers := &gossip.Handlers{}
	consGossipClient := &gossip.Client{}
	consLSEngine := &lstate.Engine{}
	statusLogger := &status.Logger{}
	secp256k1Signer := &mncrypto.Secp256k1Signer{}
	appDepositHandler := &deposit.Handler{}
	consSync := &consensus.Synchronizer{}
	localStateHandler := &localrpc.Handlers{}

	// Initialize consensus database
	if err := consDB.Init(rawConsensusDb); err != nil {
		panic(err)
	}

	peerManager := initPeerManager(consGossipHandlers, consReqHandler)
	localStateServer := initLocalStateServer(localStateHandler)

	// Initialize deposit handler
	if err := appDepositHandler.Init(); err != nil {
		panic(err)
	}

	// Initialize the request bus client
	if err := consReqClient.Init(peerManager.P2PClient()); err != nil {
		panic(err)
	}

	// Initialize the consensus engine signer
	if err := secp256k1Signer.SetPrivk(crypto.FromECDSA(keys.PrivateKey)); err != nil {
		panic(err)
	}

	// Initialize the evidence pool
	if err := consTxPool.Init(consDB); err != nil {
		panic(err)
	}

	// Initialize the app logic
	if err := app.Init(consDB, rawTxPoolDb, appDepositHandler); err != nil {
		panic(err)
	}

	// Initialize the request bus handler
	if err := consReqHandler.Init(consDB, app); err != nil {
		panic(err)
	}

	// Initialize the download manager
	if err := consDlManager.Init(consDB, app, consReqClient); err != nil {
		panic(err)
	}

	// Initialize the state handlers
	if err := consLSHandler.Init(consDB, consDlManager); err != nil {
		panic(err)
	}

	// Initialize the gossip bus handler
	if err := consGossipHandlers.Init(consDB, peerManager.P2PClient(), app, consLSHandler); err != nil {
		panic(err)
	}

	// Initialize the gossip bus client
	if err := consGossipClient.Init(consDB, peerManager.P2PClient(), app); err != nil {
		panic(err)
	}

	// Initialize admin handler
	if err := consAdminHandlers.Init(chainID, consDB, mncrypto.Hasher([]byte(config.Configuration.Validator.SymmetricKey)), app, publicKey); err != nil {
		panic(err)
	}

	// Initialize the consensus engine
	if err := consLSEngine.Init(consDB, consDlManager, app, secp256k1Signer, consAdminHandlers, publicKey, consReqClient); err != nil {
		panic(err)
	}

	// Setup Request Bus Services
	svcs := monitor.NewServices(eth, consDB, appDepositHandler, consAdminHandlers, batchSize, chainID)

	// Setup Request Bus
	monitorBus, err := monitor.NewBus(rbus.NewRBus(), svcs)
	if err != nil {
		panic(err)
	}

	// Setup monitor
	oneHour := 1 * time.Hour // TODO:ANTHONY - SHOULD THIS BE MOVED TO CONFIG?
	mon, err := monitor.NewMonitor(monitorDb, monitorBus, monitorInterval, oneHour)
	if err != nil {
		panic(err)
	}

	var tDB, mDB *badger.DB = nil, nil
	if config.Configuration.Chain.TransactionDbInMemory {
		// prevent value log GC on in memory by setting to nil - this will cause syncronizer to bypass GC on these databases
		tDB = rawTxPoolDb
	}
	if config.Configuration.Chain.MonitorDbInMemory {
		mDB = rawMonitorDb
	}
	if err := consSync.Init(consDB, mDB, tDB, consGossipClient, consGossipHandlers, consTxPool, consLSEngine, app, consAdminHandlers, peerManager); err != nil {
		panic(err)
	}

	// Setup the local RPC server handler
	if err := localStateHandler.Init(consDB, app, consGossipHandlers, publicKey, consSync.Safe); err != nil {
		panic(err)
	}

	// Initialize status logger
	if err := statusLogger.Init(consLSEngine, peerManager, consAdminHandlers, mon); err != nil {
		panic(err)
	}

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

	monitorBus.StartLoop()
	defer monitorBus.StopLoop()

	go peerManager.Start()
	defer peerManager.Close()

	go consGossipClient.Start()
	defer consGossipClient.Close()

	go consDlManager.Start()
	defer consDlManager.Close()

	go localStateServer.Serve()
	defer localStateServer.Close()

	go localStateHandler.Start()
	defer localStateHandler.Stop()

	go consGossipHandlers.Start()
	defer consGossipHandlers.Close()

	//////////////////////////////////////////////////////////////////////////////
	//SETUP SHUTDOWN MONITORING///////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	consSync.Start()
	select {
	case <-peerManager.CloseChan():
	case <-consSync.CloseChan():
	case <-signals:
	}
	go countSignals(logger, 5, signals)

	defer consSync.Stop()
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
