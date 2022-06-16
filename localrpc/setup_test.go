package localrpc

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/MadBase/MadNet/application"
	"github.com/MadBase/MadNet/application/deposit"
	"github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/application/utxohandler"
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
	"github.com/MadBase/MadNet/crypto"
	mncrypto "github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/dynamics"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/peering"
	"github.com/MadBase/MadNet/proto"
	pb "github.com/MadBase/MadNet/proto"
	"github.com/MadBase/MadNet/status"
	"github.com/MadBase/MadNet/utils"
	mnutils "github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/spf13/viper"
)

var timeout time.Duration = time.Second * 10
var account []byte
var signer *crypto.Secp256k1Signer
var pubKey []byte
var tx *objs.Tx
var err error
var srpc *Handlers
var lrpc *Client
var tx1, tx2, tx3 *pb.TransactionData
var tx1Hash, tx2Hash, tx3Hash []byte
var consumedTx1Hash, consumedTx2Hash, consumedTx3Hash []byte
var utxoTx1IDs, utxoTx2IDs, utxoTx3IDs [][]byte
var tx1Value, tx2Value, tx3Value uint64
var tx1Signature, tx2Signature, tx3Signature []byte

// var stateDB *db.Database
var ctx context.Context
var app *application.Application
var consGossipHandlers *gossip.Handlers
var peerManager *peering.PeerManager
var consGossipClient *gossip.Client
var consDlManager *dman.DMan
var localStateHandler *Handlers
var consDB *db.Database
var consSync *consensus.Synchronizer
var storage *dynamics.Storage

func TestMain(m *testing.M) {

	loadSettings("validator.toml")
	signer, pubKey = getSignerData()
	validatorNode()

	srpc = &Handlers{
		ctx:         ctx,
		cancelCtx:   nil,
		database:    consDB,
		sstore:      nil,
		AppHandler:  app,
		GossipBus:   consGossipHandlers,
		Storage:     storage,
		logger:      nil,
		ethAcct:     nil,
		EthPubk:     nil,
		safeHandler: func() bool { return true },
		safecount:   1,
	}
	srpc.Init(srpc.database, srpc.AppHandler, srpc.GossipBus, srpc.EthPubk, srpc.safeHandler, srpc.Storage)

	go srpc.Start()
	defer srpc.Stop()

	lrpc = &Client{
		Mutex:       sync.Mutex{},
		closeChan:   nil,
		closeOnce:   sync.Once{},
		Address:     config.Configuration.Transport.LocalStateListeningAddress,
		TimeOut:     timeout,
		conn:        nil,
		client:      nil,
		wg:          sync.WaitGroup{},
		isConnected: false,
	}

	go func() {
		err := lrpc.Connect(ctx)
		if err != nil {
			panic(err)
		}
	}()
	defer func() {
		err := lrpc.Close()
		if err != nil {
			panic(err)
		}
	}()

	localStateServer := initLocalStateServer(srpc)
	go localStateServer.Serve()
	defer func() {
		err := localStateServer.Close()
		if err != nil {
			panic(err)
		}
	}()

	go storage.Start()

	consSync.Start()

	time.Sleep(1 * time.Second)
	tx1Value = 6
	tx2Value = 8
	tx3Value = 10

	utxoTx1IDs, account, consumedTx1Hash = insertTestUTXO(tx1Value)
	utxoTx2IDs, account, consumedTx2Hash = insertTestUTXO(tx2Value)
	utxoTx3IDs, account, consumedTx3Hash = insertTestUTXO(tx3Value)

	tx1, tx1Hash, tx1Signature = getTransactionRequest(consumedTx1Hash, mncrypto.GetAccount(pubKey), tx1Value)
	tx2, tx2Hash, tx2Signature = getTransactionRequest(consumedTx2Hash, mncrypto.GetAccount(pubKey), tx2Value)
	tx3, tx3Hash, tx3Signature = getTransactionRequest(consumedTx3Hash, mncrypto.GetAccount(pubKey), tx3Value)

	//Start tests after validator is running
	exitVal := m.Run()

	os.Exit(exitVal)
}

func validatorNode() {
	publicKey := pubKey
	// create execution context for application
	ctx = context.Background()
	nodeCtx := ctx
	// defer cf()

	// setup logger for program assembly operations
	logger := logging.GetLogger("validator")
	logger.Infof("Starting node with configuration %v", viper.AllSettings())
	defer func() { logger.Warning("Goodbye.") }()

	chainID := uint32(config.Configuration.Chain.ID)
	// batchSize := config.Configuration.Monitor.BatchSize

	// No need of eth connection to test localrpc
	// eth, keys, publicKey := initEthereumConnection(logger)

	// Initialize consensus db: stores all state the consensus mechanism requires to work
	rawConsensusDb := initDatabase(nodeCtx, config.Configuration.Chain.StateDbPath, config.Configuration.Chain.StateDbInMemory)
	// Need to keep DBs open after this function
	// defer rawConsensusDb.Close()

	// Initialize transaction pool db: contains transcations that have not been mined (and thus are to be gossiped)
	rawTxPoolDb := initDatabase(nodeCtx, config.Configuration.Chain.TransactionDbPath, config.Configuration.Chain.TransactionDbInMemory)
	// defer rawTxPoolDb.Close()

	// Initialize monitor database: tracks what ETH block number we're on (tracking deposits)
	rawMonitorDb := initDatabase(nodeCtx, config.Configuration.Chain.MonitorDbPath, config.Configuration.Chain.MonitorDbInMemory)
	// defer rawMonitorDb.Close()

	/////////////////////////////////////////////////////////////////////////////
	// INITIALIZE ALL SERVICE OBJECTS ///////////////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////

	consDB = &db.Database{}
	monDB := &db.Database{}

	// app maintains the UTXO set of the MadNet blockchain (module is separate from consensus e.d.)
	app = &application.Application{}
	appDepositHandler := &deposit.Handler{} // watches ETH blockchain about deposits

	// consDlManager is used to retrieve transactions or block headers (to verify validity for proposal vote)
	consDlManager = &dman.DMan{}

	// gossip system (e.g. I gossip the block header, I request the transactions, to drive what the next request should be)
	consGossipHandlers = &gossip.Handlers{}
	consGossipClient = &gossip.Client{}

	// link between ETH net and our internal logic, relays important ETH events (e.g. snapshot) into our system
	consAdminHandlers := &admin.Handlers{}

	// consensus p2p comm
	consReqClient := &request.Client{}
	consReqHandler := &request.Handler{}

	// core of consensus algorithm: where outside stake relies, how gossip ends up, how state modifications occur
	consLSEngine := &lstate.Engine{}
	consLSHandler := &lstate.Handlers{}

	// synchronizes execution context, makes sure everything synchronizes with the ctx system - throughout modules
	consSync = &consensus.Synchronizer{}

	// define storage to dynamic values
	storage = &dynamics.Storage{}

	// account signer for ETH accounts
	secp256k1Signer := &crypto.Secp256k1Signer{}

	// stdout logger
	statusLogger := &status.Logger{}

	peerManager = initPeerManager(consGossipHandlers, consReqHandler)

	// Initialize the consensus engine signer
	// if err := secp256k1Signer.SetPrivk(crypto.FromECDSA(keys.PrivateKey)); err != nil {
	// 	panic(err)
	// }

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

	localStateHandler = &Handlers{}

	// Initialize consensus
	consReqClient.Init(peerManager.P2PClient(), storage)
	consReqHandler.Init(consDB, app, storage)
	consDlManager.Init(consDB, app, consReqClient)
	consLSHandler.Init(consDB, consDlManager)
	consGossipHandlers.Init(chainID, consDB, peerManager.P2PClient(), app, consLSHandler, storage)
	consGossipClient.Init(consDB, peerManager.P2PClient(), app, storage)
	consAdminHandlers.Init(chainID, consDB, crypto.Hasher([]byte(config.Configuration.Validator.SymmetricKey)), app, publicKey, storage)
	consLSEngine.Init(consDB, consDlManager, app, secp256k1Signer, consAdminHandlers, publicKey, consReqClient, storage)

	// Setup monitor
	monDB.Init(rawMonitorDb)
	// No need of mon for localrpc testing
	// monitorInterval := config.Configuration.Monitor.Interval
	// monitorTimeout := config.Configuration.Monitor.Timeout
	// mon, err := monitor.NewMonitor(consDB, monDB, consAdminHandlers, appDepositHandler, nil, monitorInterval, monitorTimeout, uint64(batchSize))
	// if err != nil {
	// 	panic(err)
	// }

	var tDB, mDB *badger.DB = nil, nil
	if config.Configuration.Chain.TransactionDbInMemory {
		// prevent value log GC on in memory by setting to nil - this will cause syncronizer to bypass GC on these databases
		tDB = rawTxPoolDb
	}
	if config.Configuration.Chain.MonitorDbInMemory {
		mDB = rawMonitorDb
	}

	consSync.Init(consDB, mDB, tDB, consGossipClient, consGossipHandlers, consTxPool, consLSEngine, app, consAdminHandlers, peerManager, storage)
	localStateHandler.Init(consDB, app, consGossipHandlers, publicKey, consSync.Safe, storage)
	statusLogger.Init(consLSEngine, peerManager, consAdminHandlers, nil)

	// No need of signal management for localrpc testing
	//////////////////////////////////////////////////////////////////////////////
	//LAUNCH ALL SERVICE GOROUTINES///////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////

	/* 	select {
	   	case <-peerManager.CloseChan():
	   	case <-consSync.CloseChan():
	   	case <-signals:
	   	}
	   	// go countSignals(logger, 5, signals)

	   	defer consSync.Stop()
	   	defer func() { logger.Warning("Starting graceful unwind of core processes.") }() */
}

func getTransactionRequest(ConsumedTxHash []byte, account []byte, val uint64) (tx_ *pb.TransactionData, TxHash []byte, TXSignature []byte) {
	pubKey, _ := signer.Pubkey()
	value_, _ := new(uint256.Uint256).FromUint64(1)
	txin := &objs.TXIn{
		TXInLinker: &objs.TXInLinker{
			TXInPreImage: &objs.TXInPreImage{
				ChainID:        chainID,
				ConsumedTxIdx:  0,
				ConsumedTxHash: ConsumedTxHash,
			},
			TxHash: make([]byte, 32),
		},
	}
	v := &objs.ValueStore{
		VSPreImage: &objs.VSPreImage{
			TXOutIdx: 0,
			Value:    value_,
			ChainID:  chainID,
			Owner: &objs.ValueStoreOwner{
				SVA:       objs.ValueStoreSVA,
				CurveSpec: constants.CurveSecp256k1,
				Account:   account,
			},
			Fee: vsFee,
		},
		TxHash: make([]byte, 32),
	}
	tx = &objs.Tx{}
	tx.Vin = []*objs.TXIn{txin}
	newValueStore = &objs.ValueStore{
		VSPreImage: &objs.VSPreImage{
			ChainID:  chainID,
			Value:    vsValue,
			TXOutIdx: 0,
			Fee:      vsFee,
			Owner: &objs.ValueStoreOwner{
				SVA:       objs.ValueStoreSVA,
				CurveSpec: constants.CurveSecp256k1,
				Account:   crypto.GetAccount(pubKey)},
		},
		TxHash: make([]byte, constants.HashLen),
	}
	newUTXO := &objs.TXOut{}
	err = newUTXO.NewValueStore(newValueStore)
	tx.Vout = append(tx.Vout, newUTXO)
	tx.Fee, _ = new(uint256.Uint256).FromUint64(val - 2)
	err = tx.SetTxHash()
	err = v.Sign(tx.Vin[0], signer)
	hash, _ = tx.TxHash()
	signature := txin.Signature
	/* 	fmt.Printf("Hash %x \n", hash)
	   	fmt.Printf("Signature %x \n", signature) */
	transactionData := &pb.TransactionData{
		Tx: &pb.Tx{
			Vin: []*pb.TXIn{
				{
					TXInLinker: &pb.TXInLinker{
						TXInPreImage: &pb.TXInPreImage{
							ChainID:        txin.TXInLinker.TXInPreImage.ChainID,
							ConsumedTxIdx:  txin.TXInLinker.TXInPreImage.ConsumedTxIdx,
							ConsumedTxHash: hex.EncodeToString(txin.TXInLinker.TXInPreImage.ConsumedTxHash),
						},
						TxHash: hex.EncodeToString(hash),
					},
					Signature: hex.EncodeToString(signature),
				},
			},
			Vout: []*pb.TXOut{
				{
					Utxo: &pb.TXOut_ValueStore{
						ValueStore: &pb.ValueStore{
							VSPreImage: &pb.VSPreImage{
								ChainID:  newValueStore.VSPreImage.ChainID,
								TXOutIdx: newValueStore.VSPreImage.TXOutIdx,
								Value:    newValueStore.VSPreImage.Value.String(),
								Owner:    "0101" + hex.EncodeToString(account),
								Fee:      newValueStore.VSPreImage.Fee.String(),
							},
							TxHash: hex.EncodeToString(hash),
						},
					},
				},
			},
			Fee: tx.Fee.String(),
		},
	}
	/* 	fmt.Println(transactionData)
	 */return transactionData, hash, signature
}

func getSignerData() (*crypto.Secp256k1Signer, []byte) {
	signer := &crypto.Secp256k1Signer{}
	err := signer.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		panic(err)
	}
	pubKey, _ := signer.Pubkey()
	return signer, pubKey
}

func loadSettings(configFile string) {
	file, _ := os.Open(configFile)
	bs, _ := ioutil.ReadAll(file)
	reader := bytes.NewReader(bs)
	viper.SetConfigType("toml")
	err := viper.ReadConfig(reader)
	if err != nil {
		panic(err)
	}
	//StateDbPath is in configuration but StateDB on the file, hence fixing
	viper.Set("chain.StateDbPath", viper.GetString("chain.StateDB"))
	viper.Set("chain.MonitorDbPath", viper.GetString("chain.MonitorDB"))
	err = viper.Unmarshal(&config.Configuration)
	if err != nil {
		panic(err)
	}
}

func initDatabase(ctx context.Context, path string, inMemory bool) *badger.DB {
	db, err := mnutils.OpenBadger(ctx.Done(), path, inMemory)
	if err != nil {
		panic(err)
	}
	return db
}

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

func initLocalStateServer(localStateHandler *Handlers) *Handler {
	localStateDispatch := proto.NewLocalStateDispatch()
	localStateServer, err := NewStateServerHandler(
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

func insertTestUTXO(value_ uint64) ([][]byte, []byte, []byte) {
	accountAddress := crypto.GetAccount(pubKey)
	owner := &objs.ValueStoreOwner{
		SVA:       objs.ValueStoreSVA,
		CurveSpec: constants.CurveSecp256k1,
		Account:   accountAddress,
	}
	hndlr := utxohandler.NewUTXOHandler(consDB.DB())
	err = hndlr.Init(1)
	if err != nil {
		fmt.Printf("could not create utxo handler %v \n", err)
	}
	value, _ := new(uint256.Uint256).FromUint64(value_)
	vs := &objs.ValueStore{
		VSPreImage: &objs.VSPreImage{
			TXOutIdx: constants.MaxUint32,
			Value:    value,
			ChainID:  uint32(config.Configuration.Chain.ID),
			Owner:    owner,
		},
		TxHash: utils.ForceSliceToLength([]byte(strconv.Itoa(1)), constants.HashLen),
	}
	utxoDep := &objs.TXOut{}
	err := utxoDep.NewValueStore(vs)
	if err != nil {
		fmt.Printf("could not create utxo %v \n", err)
	}
	tx, consumedTxHash := makeTxs(signer, vs)
	utxoIDs, _ := tx.GeneratedUTXOID()
	err = consDB.Update(func(txn *badger.Txn) error {
		_, err := hndlr.ApplyState(txn, []*objs.Tx{tx}, 2)
		if err != nil {
			fmt.Printf("Could not validate %v \n", err)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Could not update DB %v \n", err)
	}
	return utxoIDs, accountAddress, consumedTxHash
}

func makeTxs(s objs.Signer, v *objs.ValueStore) (*objs.Tx, []byte) {
	txIn, err := v.MakeTxIn()
	if err != nil {
		fmt.Printf("Could not make tx in %v \n", err)
	}
	value, err := v.Value()
	if err != nil {
		fmt.Printf("Could not get value %v \n", err)
	}
	chainID, err := txIn.ChainID()
	if err != nil {
		fmt.Printf("Could not get chain id %v \n", err)
	}
	pubkey, err := s.Pubkey()
	if err != nil {
		fmt.Printf("Could not get pubkey %v \n", err)
	}
	tx := &objs.Tx{}
	tx.Vin = []*objs.TXIn{txIn}
	newValueStore := &objs.ValueStore{
		VSPreImage: &objs.VSPreImage{
			ChainID:  chainID,
			Value:    value,
			TXOutIdx: 0,
			Fee:      new(uint256.Uint256).SetZero(),
			Owner: &objs.ValueStoreOwner{
				SVA:       objs.ValueStoreSVA,
				CurveSpec: constants.CurveSecp256k1,
				Account:   crypto.GetAccount(pubkey),
			},
		},
		TxHash: make([]byte, constants.HashLen),
	}
	newUTXO := &objs.TXOut{}
	err = newUTXO.NewValueStore(newValueStore)
	if err != nil {
		fmt.Printf("Could not create utxo %v \n", err)
	}
	tx.Vout = append(tx.Vout, newUTXO)
	tx.Fee = uint256.Zero()
	err = tx.SetTxHash()
	if err != nil {
		fmt.Printf("Could not set tx hash %v \n", err)
	}
	err = v.Sign(tx.Vin[0], s)
	if err != nil {
		fmt.Printf("Could not create Txs %v \n", err)
	}
	return tx, tx.Vin[0].TXInLinker.TxHash
}
