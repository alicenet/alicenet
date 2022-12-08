package localrpc

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/spf13/viper"

	"github.com/alicenet/alicenet/application"
	"github.com/alicenet/alicenet/application/deposit"
	"github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/application/utxohandler"
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
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/dynamics"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/peering"
	"github.com/alicenet/alicenet/proto"
	"github.com/alicenet/alicenet/status"
	"github.com/alicenet/alicenet/utils"
)

var (
	timeout                                           time.Duration = time.Second * 10
	account                                           []byte
	signer                                            *crypto.Secp256k1Signer
	pubKey                                            []byte
	tx                                                *objs.Tx
	err                                               error
	srpc                                              *Handlers
	lrpc                                              *Client
	tx1, tx2, tx3                                     *proto.TransactionData
	tx1Hash, tx2Hash, tx3Hash                         []byte
	consumedTx1Hash, consumedTx2Hash, consumedTx3Hash []byte
	utxoTx1IDs, utxoTx2IDs, utxoTx3IDs                [][]byte
	tx1Signature, tx2Signature, tx3Signature          []byte
)

// var stateDB *db.Database.
var ctx context.Context

var (
	app                *application.Application
	consGossipHandlers *gossip.Handlers
	peerManager        *peering.PeerManager
	consGossipClient   *gossip.Client
	consDlManager      *dman.DMan
	localStateHandler  *Handlers
	consDB             *db.Database
	consSync           *consensus.Synchronizer
	storage            *dynamics.Storage
)

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

	consSync.Start()

	time.Sleep(1 * time.Second)
	fee := storage.GetValueStoreFee()

	utxoTx1IDs, account, consumedTx1Hash = insertTestUTXO(getTxValues(1), fee)
	utxoTx2IDs, account, consumedTx2Hash = insertTestUTXO(getTxValues(2), fee)
	utxoTx3IDs, account, consumedTx3Hash = insertTestUTXO(getTxValues(3), fee)

	tx1, tx1Hash, tx1Signature = getTransactionRequest(consumedTx1Hash, crypto.GetAccount(pubKey), getTxValues(1))
	tx2, tx2Hash, tx2Signature = getTransactionRequest(consumedTx2Hash, crypto.GetAccount(pubKey), getTxValues(2))
	tx3, tx3Hash, tx3Signature = getTransactionRequest(consumedTx3Hash, crypto.GetAccount(pubKey), getTxValues(3))

	// Start tests after validator is running
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

	// Initialize transaction pool db: contains transactions that have not been mined (and thus are to be gossiped)
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

	// app maintains the UTXO set of the AliceNet blockchain (module is separate from consensus e.d.)
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

	if storage.DynamicValues == nil {
		// add the node to the db in case is not already present
		err = consDB.Update(func(txn *badger.Txn) error {
			// dynamics with fees
			data, err := hex.DecodeString("00000fa000000bb800000bb8002dc6c00000000000000bb80000000000000bb800000000000000000000000000000fa0")
			if err != nil {
				panic(err)
			}
			return storage.ChangeDynamicValues(txn, 1, data)
		})
		if err != nil {
			panic(err)
		}
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

}

func getValueStoreFee() *uint256.Uint256 {
	value, err := new(uint256.Uint256).FromUint64(storage.GetValueStoreFee().Uint64())
	if err != nil {
		panic(err)
	}
	return value
}

func getMinTxFee() *uint256.Uint256 {
	value, err := new(uint256.Uint256).FromUint64(storage.GetMinScaledTransactionFee().Uint64())
	if err != nil {
		panic(err)
	}
	return value
}

func getTransactionRequest(ConsumedTxHash, account []byte, val uint64) (tx_ *proto.TransactionData, TxHash, TXSignature []byte) {
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
			Fee: getValueStoreFee(),
		},
		TxHash: make([]byte, 32),
	}
	tx = &objs.Tx{}
	tx.Vin = []*objs.TXIn{txin}
	expectedOutput, err := new(uint256.Uint256).FromUint64(val)
	if err != nil {
		panic(err)
	}
	expectedOutput, err = new(uint256.Uint256).Sub(expectedOutput, getValueStoreFee())
	if err != nil {
		panic(err)
	}
	expectedOutput, err = new(uint256.Uint256).Sub(expectedOutput, getMinTxFee())
	if err != nil {
		panic(err)
	}
	newValueStore = &objs.ValueStore{
		VSPreImage: &objs.VSPreImage{
			ChainID:  chainID,
			Value:    expectedOutput,
			TXOutIdx: 0,
			Fee:      getValueStoreFee(),
			Owner: &objs.ValueStoreOwner{
				SVA:       objs.ValueStoreSVA,
				CurveSpec: constants.CurveSecp256k1,
				Account:   crypto.GetAccount(pubKey),
			},
		},
		TxHash: make([]byte, constants.HashLen),
	}
	newUTXO := &objs.TXOut{}
	if err := newUTXO.NewValueStore(newValueStore); err != nil {
		panic(err)
	}
	if tx.Vout = append(tx.Vout, newUTXO); tx.Vout == nil {
		panic(err)
	}
	tx.Fee = getMinTxFee()
	if err = tx.SetTxHash(); err != nil {
		panic(err)
	}
	if err = v.Sign(tx.Vin[0], signer); err != nil {
		panic(err)
	}
	hash, _ = tx.TxHash()
	signature := txin.Signature
	/* 	fmt.Printf("Hash %x \n", hash)
	   	fmt.Printf("Signature %x \n", signature) */
	transactionData := &proto.TransactionData{
		Tx: &proto.Tx{
			Vin: []*proto.TXIn{
				{
					TXInLinker: &proto.TXInLinker{
						TXInPreImage: &proto.TXInPreImage{
							ChainID:        txin.TXInLinker.TXInPreImage.ChainID,
							ConsumedTxIdx:  txin.TXInLinker.TXInPreImage.ConsumedTxIdx,
							ConsumedTxHash: hex.EncodeToString(txin.TXInLinker.TXInPreImage.ConsumedTxHash),
						},
						TxHash: hex.EncodeToString(hash),
					},
					Signature: hex.EncodeToString(signature),
				},
			},
			Vout: []*proto.TXOut{
				{
					Utxo: &proto.TXOut_ValueStore{
						ValueStore: &proto.ValueStore{
							VSPreImage: &proto.VSPreImage{
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
	return transactionData, hash, signature
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
	bs, _ := io.ReadAll(file)
	reader := bytes.NewReader(bs)
	viper.SetConfigType("toml")
	err := viper.ReadConfig(reader)
	if err != nil {
		panic(err)
	}
	// StateDbPath is in configuration but StateDB on the file, hence fixing
	viper.Set("chain.StateDbPath", viper.GetString("chain.StateDB"))
	viper.Set("chain.MonitorDbPath", viper.GetString("chain.MonitorDB"))
	err = viper.Unmarshal(&config.Configuration)
	if err != nil {
		panic(err)
	}
}

func initDatabase(ctx context.Context, path string, inMemory bool) *badger.DB {
	db, err := utils.OpenBadger(ctx.Done(), path, inMemory)
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

func insertTestUTXO(value_ uint64, fee_ *big.Int) ([][]byte, []byte, []byte) {
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
	value, err := new(uint256.Uint256).FromUint64(value_)
	if err != nil {
		panic(err)
	}
	fee, err := new(uint256.Uint256).FromBigInt(fee_)
	if err != nil {
		panic(err)
	}
	vs := &objs.ValueStore{
		VSPreImage: &objs.VSPreImage{
			TXOutIdx: constants.MaxUint32,
			Value:    value,
			ChainID:  uint32(config.Configuration.Chain.ID),
			Owner:    owner,
			Fee:      fee,
		},
		TxHash: utils.ForceSliceToLength([]byte(strconv.Itoa(1)), constants.HashLen),
	}
	utxoDep := &objs.TXOut{}
	err = utxoDep.NewValueStore(vs)
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
			Fee:      v.VSPreImage.Fee,
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

func getTxValues(index uint32) uint64 {
	switch index {
	case 1:
		return 8000
	case 2:
		return 12000
	case 3:
		return 16000
	default:
		panic(fmt.Sprintf("Invalid index: %d", index))
	}
}
