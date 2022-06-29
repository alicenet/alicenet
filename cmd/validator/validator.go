package validator

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	"github.com/alicenet/alicenet/application"
	"github.com/alicenet/alicenet/application/deposit"
	"github.com/alicenet/alicenet/cmd/utils"
	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/consensus"
	"github.com/alicenet/alicenet/consensus/admin"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/dman"
	"github.com/alicenet/alicenet/consensus/evidence"
	"github.com/alicenet/alicenet/consensus/gossip"
	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/consensus/request"
	"github.com/alicenet/alicenet/constants"
	mncrypto "github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/dynamics"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/executor"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/monitor"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/localrpc"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/peering"
	"github.com/alicenet/alicenet/proto"
	"github.com/alicenet/alicenet/status"
	mnutils "github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Command is the cobra.Command specifically for running as a node
var Command = cobra.Command{
	Use:   "validator",
	Short: "Starts a node",
	Long:  "Runs a AliceNet node in mining or non-mining mode",
	Run:   validatorNode}

func initEthereumConnection(logger *logrus.Logger) (layer1.Client, *mncrypto.Secp256k1Signer, []byte) {
	// Ethereum connection setup
	logger.Infof("Connecting to Ethereum...")
	eth, err := ethereum.NewClient(
		config.Configuration.Ethereum.Endpoint,
		config.Configuration.Ethereum.Keystore,
		config.Configuration.Ethereum.PassCodes,
		config.Configuration.Ethereum.DefaultAccount,
		false,
		constants.EthereumFinalityDelay,
		config.Configuration.Ethereum.TxMaxGasFeeAllowedInGwei,
		config.Configuration.Ethereum.EndpointMinimumPeers)

	if err != nil {
		logger.Fatalf("NewEthereumEndpoint(...) failed: %v", err)
		panic(err)
	}
	// Load the ethereum state
	if !eth.IsAccessible() {
		logger.Fatal("Ethereum endpoint not accessible...")
		panic(err)
	}
	logger.Infof("Looking up smart contracts on Ethereum...")
	// Initialize and find all the contracts
	ethereum.NewContracts(eth, common.HexToAddress(config.Configuration.Ethereum.FactoryAddress))

	utils.LogStatus(logger.WithField("Component", "validator"), eth)

	secp256k1, err := eth.CreateSecp256k1Signer()
	if err != nil {
		panic(fmt.Sprintf("Failed to create secp global signer: %v", err))
	}
	pubKey, err := secp256k1.Pubkey()
	if err != nil {
		panic(fmt.Sprintf("Failed to get public key from secp256 signer: %v", err))
	}
	logger.Infof("Account: %v Public Key: 0x%x", eth.GetDefaultAccount().Address.Hex(), pubKey)

	return eth, secp256k1, pubKey
}

// Setup the peer manager:
// Peer manager owns the raw TCP connections of the p2p system
// Runs the gossip protocol
// Provides functionality to access methods on a remote peer (validators, miners, those who care about voting and consensus)
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

// Setup the localstate RPC server, a more REST-like API, used by e.g. wallet users (or anything that's not a node)
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
	localStateDispatch.RegisterLocalStateGetTransactionStatus(localStateHandler)
	localStateDispatch.RegisterLocalStateGetMinedTransaction(localStateHandler)
	localStateDispatch.RegisterLocalStateGetPendingTransaction(localStateHandler)
	localStateDispatch.RegisterLocalStateGetRoundStateForValidator(localStateHandler)
	localStateDispatch.RegisterLocalStateGetValidatorSet(localStateHandler)
	localStateDispatch.RegisterLocalStateIterateNameSpace(localStateHandler)
	localStateDispatch.RegisterLocalStateGetData(localStateHandler)
	localStateDispatch.RegisterLocalStateGetTxBlockNumber(localStateHandler)
	localStateDispatch.RegisterLocalStateGetFees(localStateHandler)

	return localStateServer
}

func initDatabase(ctx context.Context, path string, inMemory bool) *badger.DB {
	db, err := mnutils.OpenBadger(ctx.Done(), path, inMemory)
	if err != nil {
		panic(err)
	}
	return db
}

func validatorNode(cmd *cobra.Command, args []string) {

	// setup logger for program assembly operations
	logger := logging.GetLogger(cmd.Name())
	logger.Infof("Starting node with args %v", args)
	defer func() { logger.Warning("Graceful unwind of core process complete.") }()

	// create execution context for application
	ctx := context.Background()
	nodeCtx, cf := context.WithCancel(ctx)
	defer cf()

	chainID := uint32(config.Configuration.Chain.ID)
	batchSize := config.Configuration.Monitor.BatchSize

	eth, secp256k1Signer, publicKey := initEthereumConnection(logger)
	defer eth.Close()

	// Initialize consensus db: stores all state the consensus mechanism requires to work
	rawConsensusDb := initDatabase(nodeCtx, config.Configuration.Chain.StateDbPath, config.Configuration.Chain.StateDbInMemory)
	defer rawConsensusDb.Close()

	// Initialize transaction pool db: contains transactions that have not been mined (and thus are to be gossiped)
	rawTxPoolDb := initDatabase(nodeCtx, config.Configuration.Chain.TransactionDbPath, config.Configuration.Chain.TransactionDbInMemory)
	defer rawTxPoolDb.Close()

	// Initialize monitor database: tracks what ETH block number we're on (tracking deposits)
	rawMonitorDb := initDatabase(nodeCtx, config.Configuration.Chain.MonitorDbPath, config.Configuration.Chain.MonitorDbInMemory)
	defer rawMonitorDb.Close()

	// giving some time to services finish their work on the databases to avoid
	// panic when closing the databases
	defer func() { <-time.After(3 * time.Second) }()

	/////////////////////////////////////////////////////////////////////////////
	// INITIALIZE ALL SERVICE OBJECTS ///////////////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////

	consDB := &db.Database{}
	monDB := &db.Database{}

	// app maintains the UTXO set of the AliceNet blockchain (module is separate from consensus e.d.)
	app := &application.Application{}
	appDepositHandler := &deposit.Handler{} // watches ETH blockchain about deposits

	// consDlManager is used to retrieve transactions or block headers (to verify validity for proposal vote)
	consDlManager := &dman.DMan{}

	// gossip system (e.g. I gossip the block header, I request the transactions, to drive what the next request should be)
	consGossipHandlers := &gossip.Handlers{}
	consGossipClient := &gossip.Client{}

	// link between ETH net and our internal logic, relays important ETH events (e.g. snapshot) into our system
	consAdminHandlers := &admin.Handlers{}

	// consensus p2p comm
	consReqClient := &request.Client{}
	consReqHandler := &request.Handler{}

	// core of consensus algorithm: where outside stake relies, how gossip ends up, how state modifications occur
	consLSEngine := &lstate.Engine{}
	consLSHandler := &lstate.Handlers{}

	// synchronizes execution context, makes sure everything synchronizes with the ctx system - throughout modules
	consSync := &consensus.Synchronizer{}

	// define storage to dynamic values
	storage := &dynamics.Storage{}

	// stdout logger
	statusLogger := &status.Logger{}

	peerManager := initPeerManager(consGossipHandlers, consReqHandler)

	consDB.Init(rawConsensusDb)

	// consTxPool takes old state from consensusDB, used as evidence for what was done (new blocks, consensus, voting)
	consTxPool := evidence.NewPool(consDB)

	appDepositHandler.Init()
	if err := app.Init(consDB, rawTxPoolDb, appDepositHandler, storage); err != nil {
		panic(err)
	}

	// Initialize storage
	if err := storage.Init(consDB, logger); err != nil {
		panic(err)
	}

	localStateHandler := &localrpc.Handlers{}
	localStateServer := initLocalStateServer(localStateHandler)

	// Initialize consensus
	consReqClient.Init(peerManager.P2PClient(), storage)
	consReqHandler.Init(consDB, app, storage)
	consDlManager.Init(consDB, app, consReqClient)
	consLSHandler.Init(consDB, consDlManager)
	consGossipHandlers.Init(chainID, consDB, peerManager.P2PClient(), app, consLSHandler, storage)
	consGossipClient.Init(consDB, peerManager.P2PClient(), app, storage)
	consAdminHandlers.Init(chainID, consDB, mncrypto.Hasher([]byte(config.Configuration.Validator.SymmetricKey)), app, publicKey, storage)
	consLSEngine.Init(consDB, consDlManager, app, secp256k1Signer, consAdminHandlers, publicKey, consReqClient, storage)

	// Setup monitor
	monDB.Init(rawMonitorDb)

	// Layer 1 transaction watcher
	txWatcher := transaction.WatcherFromNetwork(eth, monDB, config.Configuration.Ethereum.TxMetricsDisplay, constants.TxPollingTime)
	defer txWatcher.Close()

	// Setup tasks scheduler
	taskRequestChan := make(chan tasks.TaskRequest, constants.TaskSchedulerBufferSize)
	defer close(taskRequestChan)

	tasksScheduler, err := executor.NewTasksScheduler(monDB, eth, consAdminHandlers, taskRequestChan, txWatcher)
	if err != nil {
		panic(err)
	}

	monitorInterval := config.Configuration.Monitor.Interval
	mon, err := monitor.NewMonitor(consDB, monDB, consAdminHandlers, appDepositHandler, eth, monitorInterval, uint64(batchSize), taskRequestChan)
	if err != nil {
		panic(err)
	}

	var tDB, mDB *badger.DB = nil, nil
	if config.Configuration.Chain.TransactionDbInMemory {
		// prevent value log GC on in memory by setting to nil - this will cause synchronizer to bypass GC on these databases
		tDB = rawTxPoolDb
	}
	if config.Configuration.Chain.MonitorDbInMemory {
		mDB = rawMonitorDb
	}

	consSync.Init(consDB, mDB, tDB, consGossipClient, consGossipHandlers, consTxPool, consLSEngine, app, consAdminHandlers, peerManager, storage)
	localStateHandler.Init(consDB, app, consGossipHandlers, publicKey, consSync.Safe, storage)
	statusLogger.Init(consLSEngine, peerManager, consAdminHandlers, mon)

	//////////////////////////////////////////////////////////////////////////////
	//LAUNCH ALL SERVICE GOROUTINES///////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////

	go storage.Start()

	go statusLogger.Run()
	defer statusLogger.Close()

	err = tasksScheduler.Start()
	if err != nil {
		panic(err)
	}
	defer tasksScheduler.Close()

	err = mon.Start()
	if err != nil {
		panic(err)
	}
	defer mon.Close()

	go peerManager.Start()
	defer peerManager.Close()

	go consGossipClient.Start() //nolint:errcheck
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

	signals := make(chan os.Signal, 1)
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
// this is a convenient escape from a deadlock during shutdown
func countSignals(logger *logrus.Logger, num int, c chan os.Signal) {
	<-c
	for count := 0; count < num; count++ {
		logger.Warnf("Send Ctrl+C %v more times to force shutdown without waiting for services.\n", num-count)
		<-c
	}
	os.Exit(1)
}
