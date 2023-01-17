package node

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alicenet/alicenet/application"
	"github.com/alicenet/alicenet/application/deposit"
	"github.com/alicenet/alicenet/bridge/bindings"
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
	ethEvents "github.com/alicenet/alicenet/layer1/chains/ethereum/events"
	polyEvents "github.com/alicenet/alicenet/layer1/chains/polygon/events"
	"github.com/alicenet/alicenet/layer1/evm"
	"github.com/alicenet/alicenet/layer1/executor"
	"github.com/alicenet/alicenet/layer1/handlers"
	"github.com/alicenet/alicenet/layer1/monitor"
	"github.com/alicenet/alicenet/layer1/monitor/objects"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/localrpc"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/peering"
	"github.com/alicenet/alicenet/proto"
	"github.com/alicenet/alicenet/status"
	aUtils "github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Command is the cobra.Command specifically for running as a node.
var Command = cobra.Command{
	Use:   "node",
	Short: "Starts a node",
	Long:  "Runs a AliceNet node in mining or non-mining mode",
	Run:   validatorNode,
}

func initEVMConnection(conf config.EthereumConfig,
	logger *logrus.Logger,
) (layer1.Client, *mncrypto.Secp256k1Signer, []byte) {
	// Ethereum connection setup
	logger.Infof("Connecting to EVM chain...")
	eth, err := evm.NewClient(
		logger,
		conf.Endpoint,
		conf.Keystore,
		conf.PassCodes,
		conf.DefaultAccount,
		false,
		constants.EthereumFinalityDelay,
		conf.TxMaxGasFeeAllowedInGwei,
		conf.EndpointMinimumPeers)
	if err != nil {
		logger.Fatalf("NewEthereumEndpoint(...) failed: %v", err)
		panic(err)
	}
	// Load the ethereum state
	if !eth.IsAccessible() {
		logger.Fatal("Ethereum endpoint not accessible...")
		panic(err)
	}

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
// Provides functionality to access methods on a remote peer (validators, miners, those who care about voting and consensus).
func initPeerManager(
	consGossipHandlers *gossip.Handlers,
	consReqHandler *request.Handler,
) *peering.PeerManager {
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

// Setup the localstate RPC server, a more REST-like API, used by e.g. wallet users (or anything that's not a node).
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
	db, err := aUtils.OpenBadger(ctx.Done(), path, inMemory)
	if err != nil {
		panic(err)
	}
	return db
}

func initSmartContractsHandler(
	ethClient layer1.Client,
	ethFactoryAddress common.Address,
	polygonClient layer1.Client,
	polygonFactoryAddress common.Address,
	logger *logrus.Logger) layer1.AllSmartContracts {
	// Initialize and find all the contracts
	contractsHandler := handlers.NewAllSmartContractsHandle(
		ethClient,
		ethFactoryAddress,
		polygonClient,
		polygonFactoryAddress,
	)

	utils.LogStatus(logger.WithField("Component", "validator"), ethClient, contractsHandler)
	return contractsHandler
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

	ethClient, ethSecp256k1Signer, ethPublicKey := initEVMConnection(config.Configuration.Ethereum, logger)
	defer ethClient.Close()

	//polygonClient, polygonContractsHandler, polygonSecp256k1Signer, polygonPublicKey := initEVMConnection(config.Configuration.Polygon, logger)
	polygonClient, _, _ := initEVMConnection(config.Configuration.Polygon, logger)
	defer polygonClient.Close()

	allContractsHandler := initSmartContractsHandler(
		ethClient,
		common.HexToAddress(config.Configuration.Ethereum.FactoryAddress),
		polygonClient,
		common.HexToAddress(config.Configuration.Polygon.FactoryAddress),
		logger)

	currentEpoch, latestVersion, err := getCurrentEpochAndCanonicalVersion(
		nodeCtx,
		ethClient,
		allContractsHandler,
		logger,
	)
	if err != nil {
		panic(err)
	}

	newMajorIsGreater, _, _, localVersion, err := aUtils.CompareCanonicalVersion(latestVersion)
	if err != nil {
		panic(err)
	}

	logger.Infof(
		"Local AliceNet Node Version %d.%d.%d",
		localVersion.Major,
		localVersion.Minor,
		localVersion.Patch,
	)
	if newMajorIsGreater && currentEpoch >= latestVersion.ExecutionEpoch {
		logger.Fatalf(
			"CRITICAL: Exiting! Your Node Version %d.%d.%d is lower than the latest required version %d.%d.%d! Please update your node!",
			localVersion.Major,
			localVersion.Minor,
			localVersion.Patch,
			latestVersion.Major,
			latestVersion.Minor,
			latestVersion.Patch,
		)
	}

	// Initialize consensus db: stores all state the consensus mechanism requires to work
	rawConsensusDb := initDatabase(
		nodeCtx,
		config.Configuration.Chain.StateDbPath,
		config.Configuration.Chain.StateDbInMemory,
	)
	defer rawConsensusDb.Close()

	// Initialize transaction pool db: contains transactions that have not been mined (and thus are to be gossiped)
	rawTxPoolDb := initDatabase(
		nodeCtx,
		config.Configuration.Chain.TransactionDbPath,
		config.Configuration.Chain.TransactionDbInMemory,
	)
	defer rawTxPoolDb.Close()

	// Initialize monitor database: tracks what ETH block number we're on (tracking deposits)
	rawMonitorDb := initDatabase(
		nodeCtx,
		config.Configuration.Chain.MonitorDbPath,
		config.Configuration.Chain.MonitorDbInMemory,
	)
	defer rawMonitorDb.Close()

	// giving some time to services finish their work on the databases to avoid
	// panic when closing the databases
	defer func() { <-time.After(15 * time.Second) }()

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
	consAdminHandlers.Init(
		chainID,
		consDB,
		mncrypto.Hasher([]byte(config.Configuration.Validator.SymmetricKey)),
		app,
		ethPublicKey,
		storage,
	)
	consLSEngine.Init(
		consDB,
		consDlManager,
		app,
		ethSecp256k1Signer,
		consAdminHandlers,
		ethPublicKey,
		consReqClient,
		storage,
	)

	// Setup monitor
	monDB.Init(rawMonitorDb)

	// Layer 1 transaction watcher
	ethTxWatcher := transaction.WatcherFromNetwork(
		ethClient,
		monDB,
		config.Configuration.Ethereum.TxMetricsDisplay,
		constants.TxPollingTime,
	)
	defer ethTxWatcher.Close()
	polygonTxWatcher := transaction.WatcherFromNetwork(polygonClient, monDB, config.Configuration.Polygon.TxMetricsDisplay, constants.TxPollingTime)
	defer ethTxWatcher.Close()

	// Setup tasks scheduler
	ethTasksHandler, err := executor.NewTaskHandler(
		monDB,
		ethClient, allContractsHandler,
		consAdminHandlers, ethTxWatcher)
	if err != nil {
		panic(err)
	}
	polygonTasksHandler, err := executor.NewTaskHandler(monDB, polygonClient, allContractsHandler,
		consAdminHandlers,
		polygonTxWatcher,
	)
	if err != nil {
		panic(err)
	}

	// setup monitor
	ethereumEventMap := objects.NewEventMap()
	err = ethEvents.SetupEventMap(ethereumEventMap, consDB, monDB, consAdminHandlers, appDepositHandler, ethTasksHandler, func() {}, uint32(config.Configuration.Chain.ID))
	if err != nil {
		logger.Fatalf("Error creating ethereum event map: %v", err)
	}

	monitorInterval := constants.MonitorInterval
	ethBatchSize := config.Configuration.Ethereum.ProcessingBlockBatchSize
	ethMon, err := monitor.NewMonitor(
		consDB,
		monDB,
		consAdminHandlers,
		appDepositHandler,
		ethClient,
		allContractsHandler, allContractsHandler.EthereumContracts().GetAllAddresses(),
		monitorInterval, ethBatchSize, uint32(config.Configuration.Chain.ID), ethTasksHandler, ethereumEventMap)
	if err != nil {
		panic(err)
	}

	polygonEventMap := objects.NewEventMap()
	// todo: update SetupPolygonEventMap() to only listen to necessary events
	err = polyEvents.SetupEventMap(polygonEventMap, consDB, monDB, consAdminHandlers, appDepositHandler, ethTasksHandler, func() {}, uint32(config.Configuration.Chain.ID))
	if err != nil {
		logger.Fatalf("Error creating polygon event map: %v", err)
	}

	polygonBatchSize := config.Configuration.Polygon.ProcessingBlockBatchSize
	polygonMon, err := monitor.NewMonitor(consDB, monDB, consAdminHandlers, appDepositHandler, polygonClient, allContractsHandler, allContractsHandler.EthereumContracts().GetAllAddresses(),
		monitorInterval,
		polygonBatchSize,
		uint32(config.Configuration.Chain.ID),
		polygonTasksHandler,
		polygonEventMap)
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

	consSync.Init(
		consDB,
		mDB,
		tDB,
		consGossipClient,
		consGossipHandlers,
		consTxPool,
		consLSEngine,
		app,
		consAdminHandlers,
		peerManager,
		storage,
	)
	localStateHandler.Init(consDB, app, consGossipHandlers, ethPublicKey, consSync.Safe, storage)
	statusLogger.Init(consLSEngine, peerManager, consAdminHandlers, ethMon)

	//////////////////////////////////////////////////////////////////////////////
	//LAUNCH ALL SERVICE GOROUTINES///////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////

	go statusLogger.Run()
	defer statusLogger.Close()

	ethTasksHandler.Start()
	defer ethTasksHandler.Close()

	err = ethMon.Start()
	if err != nil {
		panic(err)
	}
	defer ethMon.Close()

	err = polygonMon.Start()
	if err != nil {
		panic(err)
	}
	defer polygonMon.Close()

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
	case <-ethMon.CloseChan():
	case <-polygonMon.CloseChan():
	case <-ethTasksHandler.CloseChan():
	case <-polygonTasksHandler.CloseChan():
	case <-signals:
	}
	go countSignals(logger, 5, signals)

	defer consSync.Stop()
	defer func() { logger.Warning("Starting graceful unwind of core processes.") }()
}

// countSignals will cause a forced exit on repeated Ctrl+C commands
// this is a convenient escape from a deadlock during shutdown.
func countSignals(logger *logrus.Logger, num int, c chan os.Signal) {
	<-c
	for count := 0; count < num; count++ {
		logger.Warnf(
			"Send Ctrl+C %v more times to force shutdown without waiting for services.\n",
			num-count,
		)
		<-c
	}
	os.Exit(1)
}

func getCurrentEpochAndCanonicalVersion(
	ctx context.Context,
	eth layer1.Client,
	contractsHandler layer1.AllSmartContracts,
	logger *logrus.Logger,
) (uint32, bindings.CanonicalVersion, error) {
	retryCount := 5
	retryDelay := 1 * time.Second

	logEntry := logger.WithField("Method", "getCurrentEpochAndCanonicalVersion")

	var callOpts *bind.CallOpts
	var err error
	var latestVersion bindings.CanonicalVersion
	var currentEpoch *big.Int
	for i := 0; i < retryCount; i++ {
		callOpts, err = eth.GetCallOpts(ctx, eth.GetDefaultAccount())
		if err != nil {
			logEntry.Errorf("Received and error during GetCallOpts: %v", err)
			<-time.After(retryDelay)
			continue
		}
		break
	}
	if err != nil {
		return 0, latestVersion, err
	}

	for i := 0; i < retryCount; i++ {
		latestVersion, err = contractsHandler.EthereumContracts().
			Dynamics().
			GetLatestAliceNetVersion(callOpts)
		if err != nil {
			logEntry.Errorf("Received and error during GetLatestAliceNetVersion: %v", err)
			<-time.After(retryDelay)
			continue
		}
		break
	}
	if err != nil {
		return 0, latestVersion, err
	}
	logEntry.Infof(
		"Current Canonical AliceNet Node Version %d.%d.%d",
		latestVersion.Major,
		latestVersion.Minor,
		latestVersion.Patch,
	)

	for i := 0; i < retryCount; i++ {
		currentEpoch, err = contractsHandler.EthereumContracts().Snapshots().GetEpoch(callOpts)
		if err != nil {
			logEntry.Errorf("Received and error during GetEpoch: %v", err)
			<-time.After(retryDelay)
			continue
		}
		break
	}
	if err != nil {
		return 0, latestVersion, err
	}

	return uint32(currentEpoch.Uint64()), latestVersion, err
}
