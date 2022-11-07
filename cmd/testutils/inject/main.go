package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	aobjs "github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/handlers"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/localrpc"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

type fees struct {
	minTxFee      *uint256.Uint256
	valueStoreFee *uint256.Uint256
	dataStoreFee  *uint256.Uint256
}

func sleepWithContext(ctx context.Context, sleepTime time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(sleepTime):
	}
	return nil
}

func setupRPCClient(ctx context.Context, rpcNodeList []string, idx int) (*localrpc.Client, error) {
	idx2 := idx % len(rpcNodeList)
	client := &localrpc.Client{Address: rpcNodeList[idx2], TimeOut: constants.MsgTimeout}
	if err := client.Connect(ctx); err != nil {
		return nil, err
	}
	return client, nil
}

func setupTestingSigner(i int) (aobjs.Signer, []byte, error) {
	privk := crypto.Hasher([]byte(strconv.Itoa(i)))
	if i%2 == 0 {
		return setupSecpSigner(privk)
	}
	return setupBNSigner(privk)
}

func setupBNSigner(privk []byte) (*crypto.BNSigner, []byte, error) {
	signer := &crypto.BNSigner{}
	err := signer.SetPrivk(privk)
	if err != nil {
		return nil, nil, err
	}
	pubk, err := signer.Pubkey()
	if err != nil {
		return nil, nil, err
	}
	acct := crypto.GetAccount(pubk)
	return signer, acct, nil
}

func setupHexSigner(privk string) (*crypto.Secp256k1Signer, []byte, error) {
	privkb, err := hex.DecodeString(privk)
	if err != nil {
		return nil, nil, err
	}
	return setupSecpSigner(privkb)
}

func setupSecpSigner(privk []byte) (*crypto.Secp256k1Signer, []byte, error) {
	signer := &crypto.Secp256k1Signer{}
	if err := signer.SetPrivk(privk); err != nil {
		return nil, nil, err
	}
	pubk, err := signer.Pubkey()
	if err != nil {
		return nil, nil, err
	}
	acct := crypto.GetAccount(pubk)
	return signer, acct, nil
}

func getCurveSpec(s aobjs.Signer) constants.CurveSpec {
	curveSpec := constants.CurveSpec(0)
	switch s.(type) {
	case *crypto.Secp256k1Signer:
		curveSpec = constants.CurveSecp256k1
	case *crypto.BNSigner:
		curveSpec = constants.CurveBN256Eth
	default:
		panic("invalid signer type")
	}
	return curveSpec
}

type worker struct {
	f      *funder
	logger *logrus.Entry
	client *localrpc.Client
	signer aobjs.Signer
	acct   []byte
	idx    int
}

func (w *worker) run(ctx context.Context) {
	defer w.f.wg.Done()
	for {
		err := sleepWithContext(ctx, time.Second*time.Duration(1+w.idx%10))
		if err != nil {
			w.logger.Errorf("exiting context finished: %v", w.idx, err)
			return
		}
		w.logger.Tracef("gettingFunding", w.idx)
		u, v, err := w.f.blockingGetFunding(ctx, w.client, getCurveSpec(w.signer), w.acct, uint256.One())
		if err != nil {
			w.logger.Errorf("error at blockingGetFunding: %v", w.idx, err)
			return
		}
		w.logger.Tracef("gotFunding: %v", w.idx, v)
		tx, err := w.f.setupTransaction(w.signer, w.acct, v, u, nil)
		if err != nil {
			w.logger.Errorf("error at setupTransaction: %v", w.idx, err)
			return
		}
		w.logger.Tracef("sending Tx:", w.idx)
		err = w.f.blockingSendTx(ctx, w.client, tx)
		if err != nil {
			w.logger.Errorf("error at blockingSendTx: %v", w.idx, err)
			return
		}
	}
}

type funder struct {
	ctx          context.Context
	wg           sync.WaitGroup
	logger       *logrus.Entry
	signer       *crypto.Secp256k1Signer
	client       *localrpc.Client
	chainID      uint32
	acct         []byte
	ethAcct      accounts.Account
	ethClient    *ethereum.Client
	ethContracts layer1.AllSmartContracts
	ethTxWatcher *transaction.FrontWatcher
	numChildren  int
	children     []*worker
	rpcNodeList  []string
	baseIdx      int
}

func createNewFunder(
	ctx context.Context,
	logger *logrus.Entry,
	ethClient *ethereum.Client,
	ethAccount accounts.Account,
	ethWatcher *transaction.FrontWatcher,
	ethContracts layer1.AllSmartContracts,
	nodeList []string,
) (*funder, error) {

	logger.Info("Funder setting up signing")
	signer, acct, err := setupHexSigner(tests.TestAdminPrivateKey)
	if err != nil {
		logger.Errorf("Funder error at setupHexSigner: %v", err)
		return nil, err
	}
	logger.Info("Funder setting up client")
	client, err := setupRPCClient(ctx, nodeList, 0)
	if err != nil {
		logger.Errorf("Funder error at setupClient: %v", err)
		return nil, err
	}
	blockHeader, err := client.GetBlockHeader(ctx, 1)
	if err != nil {
		panic(fmt.Sprintf("Failed to get block header and chain id: %v", err))
	}
	chainID := blockHeader.BClaims.ChainID

	return &funder{
		ctx:          ctx,
		wg:           sync.WaitGroup{},
		logger:       logger,
		signer:       signer,
		client:       client,
		chainID:      chainID,
		acct:         acct,
		ethAcct:      ethAccount,
		ethClient:    ethClient,
		ethContracts: ethContracts,
		ethTxWatcher: ethWatcher,
		numChildren:  1,
		children:     []*worker{},
		rpcNodeList:  nodeList,
		baseIdx:      100,
	}, nil
}

func (f *funder) init(numChildren int) error {
	children, err := f.setupChildren(f.ctx, numChildren, f.baseIdx)
	if err != nil {
		f.logger.Errorf("Funder error at setupChildren: %v", err)
		return err
	}
	f.children = children
	f.logger.Info("Funder setting up funding")
	totalFunding, err := new(uint256.Uint256).FromUint64(uint64(len(children)))
	if err != nil {
		panic(err)
	}
	utxos, value, err := f.blockingGetFunding(f.ctx, f.client, getCurveSpec(f.signer), f.acct, totalFunding)
	if err != nil {
		f.logger.Errorf("Funder error at blockingGetFunding: %v", err)
		return err
	}
	f.logger.Info("Funder setting up tx")
	tx, err := f.setupTransaction(f.signer, f.acct, value, utxos, f.children)
	if err != nil {
		f.logger.Errorf("Funder error at setupTransaction: %v", err)
		return err
	}
	f.logger.Info("Funder setting up blockingSendTx")
	if err := f.blockingSendTx(f.ctx, f.client, tx); err != nil {
		f.logger.Errorf("Funder error at blockingSendTx: %v", err)
		return err
	}
	f.logger.Info("Funder starting children")
	for _, c := range f.children {
		f.wg.Add(1)
		go c.run(f.ctx)
	}
	f.logger.Info("Funder waiting on close")
	<-f.ctx.Done()
	f.wg.Wait()
	return nil
}

func (f *funder) setupChildren(ctx context.Context, numChildren int, baseIdx int) ([]*worker, error) {
	workers := []*worker{}
	for i := 0; i < numChildren; i++ {
		client, err := setupRPCClient(ctx, f.rpcNodeList, baseIdx+i)
		if err != nil {
			return nil, err
		}
		signer, acct, err := setupTestingSigner(baseIdx + i)
		if err != nil {
			return nil, err
		}
		logger := f.logger.WithFields(*&logrus.Fields{
			"worker": baseIdx + i,
		})
		c := &worker{
			f:      f,
			logger: logger,
			signer: signer,
			acct:   acct,
			client: client,
			idx:    baseIdx + i,
		}
		workers = append(workers, c)
	}
	return workers, nil
}

func (f *funder) getTxFees() (*fees, error) {
	feesString, err := f.client.GetTxFees(f.ctx)
	if err != nil {
		return nil, err
	}
	if len(feesString) != 3 {
		panic("invalid fee response")
	}
	minTxFee := new(uint256.Uint256)
	err = minTxFee.UnmarshalString(feesString[0])
	if err != nil {
		panic(fmt.Sprintf("failed to decode minTx fee %v", err))
	}
	vsFee := new(uint256.Uint256)
	err = vsFee.UnmarshalString(feesString[1])
	if err != nil {
		panic(fmt.Sprintf("failed to decode valueStoreTx fee %v", err))
	}
	dsFee := new(uint256.Uint256)
	err = dsFee.UnmarshalString(feesString[1])
	if err != nil {
		panic(fmt.Sprintf("failed to decode dataStoreTx fee %v", err))
	}
	return &fees{minTxFee, vsFee, dsFee}, nil
}

func (f *funder) setupTransaction(
	signer aobjs.Signer,
	ownerAcct []byte,
	consumedValue *uint256.Uint256,
	consumedUtxos aobjs.Vout,
	recipients []*worker,
) (*aobjs.Tx, error) {
	fees, err := f.getTxFees()
	if err != nil {
		return nil, err
	}
	tx := &aobjs.Tx{
		Vin:  aobjs.Vin{},
		Vout: aobjs.Vout{},
		Fee:  fees.minTxFee.Clone(),
	}
	chainID := f.chainID
	for _, utxo := range consumedUtxos {
		txIn, err := utxo.MakeTxIn()
		if err != nil {
			return nil, err
		}
		tx.Vin = append(tx.Vin, txIn)
	}
	// We include txFee here!
	valueOut := fees.minTxFee.Clone()
	for _, r := range recipients {
		value := uint256.One()
		newOwner := &aobjs.ValueStoreOwner{}
		newOwner.New(r.acct, getCurveSpec(r.signer))
		newValueStore := &aobjs.ValueStore{
			VSPreImage: &aobjs.VSPreImage{
				ChainID:  chainID,
				Value:    value.Clone(),
				Owner:    newOwner,
				TXOutIdx: 0, //place holder for now
				Fee:      fees.valueStoreFee.Clone(),
			},
			TxHash: make([]byte, constants.HashLen),
		}
		valuePlusFee := &uint256.Uint256{}
		_, err := valuePlusFee.Add(value, fees.valueStoreFee.Clone())
		if err != nil {
			panic(err)
		}
		_, err = valueOut.Add(valueOut, valuePlusFee)
		if err != nil {
			panic(err)
		}

		newUTXO := &aobjs.TXOut{}
		err = newUTXO.NewValueStore(newValueStore)
		if err != nil {
			return nil, err
		}
		tx.Vout = append(tx.Vout, newUTXO)
	}
	if consumedValue.Gt(valueOut) {
		diff, err := new(uint256.Uint256).Sub(consumedValue, valueOut)
		if err != nil {
			panic(err)
		}
		diff, err = new(uint256.Uint256).Sub(diff, fees.valueStoreFee.Clone())
		if err != nil {
			panic(err)
		}
		newOwner := &aobjs.ValueStoreOwner{}
		newOwner.New(ownerAcct, getCurveSpec(signer))
		newValueStore := &aobjs.ValueStore{
			VSPreImage: &aobjs.VSPreImage{
				ChainID:  chainID,
				Value:    diff,
				Owner:    newOwner,
				TXOutIdx: 0,
				Fee:      fees.valueStoreFee.Clone().Clone(),
			},
			TxHash: make([]byte, constants.HashLen),
		}
		newUTXO := &aobjs.TXOut{}
		newUTXO.NewValueStore(newValueStore)
		tx.Vout = append(tx.Vout, newUTXO)
	}
	err = tx.SetTxHash()
	if err != nil {
		panic(err)
	}
	for idx, consumedUtxo := range consumedUtxos {
		consumedVS, err := consumedUtxo.ValueStore()
		if err != nil {
			return nil, err
		}
		txIn := tx.Vin[idx]
		err = consumedVS.Sign(txIn, signer)
		if err != nil {
			return nil, err
		}
	}
	txb, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	f.logger.Infof("TX SIZE: %v", len(txb))
	return tx, nil
}

func (f *funder) setupDataStoreTransaction(
	ctx context.Context,
	signer aobjs.Signer,
	ownerAcct []byte,
	msg string,
	ind string,
	numEpochs uint32,
) (*aobjs.Tx, error) {
	fees, err := f.getTxFees()
	if err != nil {
		return nil, err
	}
	index := crypto.Hasher([]byte(ind))
	deposit, err := aobjs.BaseDepositEquation(uint32(len(msg)), numEpochs)
	if err != nil {
		return nil, err
	}
	consumedUtxos, consumedValue, err := f.blockingGetFunding(
		ctx,
		f.client,
		getCurveSpec(f.signer),
		f.acct,
		deposit,
	)
	if err != nil {
		panic(err)
	}
	if consumedValue.Lt(deposit) {
		f.logger.Infof("ACCOUNT DOES NOT HAVE ENOUGH FUNDING: REQUIRES:%v    HAS:%v", deposit, consumedValue)
		os.Exit(1)
	}
	tx := &aobjs.Tx{
		Vin:  aobjs.Vin{},
		Vout: aobjs.Vout{},
		Fee:  fees.minTxFee.Clone(),
	}
	for _, utxo := range consumedUtxos {
		txIn, err := utxo.MakeTxIn()
		if err != nil {
			return nil, err
		}
		tx.Vin = append(tx.Vin, txIn)
	}
	f.logger.Infof("THE LEN OF VIN %v ", len(tx.Vin))
	valueOut := uint256.Zero()

	en, err := f.client.GetEpochNumber(ctx)
	if err != nil {
		return nil, err
	}
	f.logger.Infof("REQUIRED DEPOSIT %v ", deposit)
	_, err = valueOut.Add(valueOut, deposit)
	if err != nil {
		panic(err)
	}
	newOwner := &aobjs.DataStoreOwner{}
	newOwner.New(ownerAcct, getCurveSpec(signer))
	newDataStore := &aobjs.DataStore{
		DSLinker: &aobjs.DSLinker{
			DSPreImage: &aobjs.DSPreImage{
				ChainID:  f.chainID,
				Index:    index,
				IssuedAt: en,
				Deposit:  deposit,
				RawData:  []byte(msg),
				TXOutIdx: 0,
				Owner:    newOwner,
				Fee:      fees.dataStoreFee.Clone(),
			},
			TxHash: make([]byte, constants.HashLen),
		},
	}
	eoe, err := newDataStore.EpochOfExpiration()
	if err != nil {
		return nil, err
	}
	f.logger.Infof("Consumed Next:%v    ValueOut:%v", consumedValue, valueOut)
	f.logger.Infof("DS:  index:%x    deposit:%v    EpochOfExpire:%v    msg:%s", index, deposit, eoe, msg)
	newUTXO := &aobjs.TXOut{}
	err = newUTXO.NewDataStore(newDataStore)
	if err != nil {
		panic(err)
	}
	tx.Vout = append(tx.Vout, newUTXO)

	f.logger.Infof("Consumed Next:%v    ValueOut:%v", consumedValue, valueOut)
	if consumedValue.Gt(valueOut) {
		diff, err := new(uint256.Uint256).Sub(consumedValue.Clone(), valueOut.Clone())
		if err != nil {
			panic(err)
		}
		newOwner := &aobjs.ValueStoreOwner{}
		newOwner.New(ownerAcct, getCurveSpec(signer))
		newValueStore := &aobjs.ValueStore{
			VSPreImage: &aobjs.VSPreImage{
				ChainID:  f.chainID,
				Value:    diff,
				Owner:    newOwner,
				TXOutIdx: 0,
				Fee:      fees.dataStoreFee.Clone(),
			},
			TxHash: make([]byte, constants.HashLen),
		}
		newUTXO := &aobjs.TXOut{}
		err = newUTXO.NewValueStore(newValueStore)
		if err != nil {
			panic(err)
		}
		tx.Vout = append(tx.Vout, newUTXO)
		//valueOut += diff
		_, err = valueOut.Add(valueOut, diff)
		if err != nil {
			panic(err)
		}
	}
	f.logger.Infof("Consumed Next:%v    ValueOut:%v", consumedValue, valueOut)
	err = tx.Vout.SetTxOutIdx()
	if err != nil {
		panic(err)
	}
	f.logger.Infof("Consumed Next:%v    ValueOut:%v", consumedValue, valueOut)
	err = tx.SetTxHash()
	if err != nil {
		panic(err)
	}
	for _, newUtxo := range tx.Vout {
		switch {
		case newUtxo.HasDataStore():
			ds, err := newUtxo.DataStore()
			if err != nil {
				return nil, err
			}
			if err := ds.PreSign(signer); err != nil {
				return nil, err
			}
		default:
			continue
		}
	}
	for idx, consumedUtxo := range consumedUtxos {
		switch {
		case consumedUtxo.HasValueStore():
			consumedVS, err := consumedUtxo.ValueStore()
			if err != nil {
				return nil, err
			}
			txIn := tx.Vin[idx]
			err = consumedVS.Sign(txIn, signer)
			if err != nil {
				panic(err)
			}
		case consumedUtxo.HasDataStore():
			consumedDS, err := consumedUtxo.DataStore()
			if err != nil {
				return nil, err
			}
			txIn := tx.Vin[idx]
			err = consumedDS.Sign(txIn, signer)
			if err != nil {
				panic(err)
			}
		}
	}
	f.logger.Infof("Consumed Next:%v    ValueOut:%v", consumedValue, valueOut)
	txb, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	f.logger.Infof("TX SIZE: %v", len(txb))
	return tx, nil
}

func (f *funder) mintALCBDepositOnEthereum() {
	txnOpts, err := f.ethClient.GetTransactionOpts(f.ctx, f.ethAcct)
	if err != nil {
		panic(fmt.Errorf("failed to get transaction options: %v", err))
	}
	// 1_000_000 ALCB (10**18)
	depositAmount, ok := new(big.Int).SetString("1000000000000000000000000", 10)
	if !ok {
		panic("Could not generate deposit amount")
	}
	txn, err := f.ethContracts.EthereumContracts().BToken().VirtualMintDeposit(
		txnOpts,
		1,
		f.ethAcct.Address,
		depositAmount,
	)
	if err != nil {
		panic(fmt.Errorf("Could not send deposit amount to ethereum %v", err))
	}
	f.ethTxWatcher.SubscribeAndWait(f.ctx, txn, nil)
}

func (f *funder) blockingGetFunding(ctx context.Context, client *localrpc.Client, curveSpec constants.CurveSpec, acct []byte, value *uint256.Uint256) (aobjs.Vout, *uint256.Uint256, error) {
	for {
		select {
		case <-f.ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(1 * time.Second):
			utxoIDs, totalValue, err := client.GetValueForOwner(ctx, curveSpec, acct, value)
			if err != nil {
				f.logger.Errorf("Getting fund err: %v", err)
				continue
			}
			utxos, err := client.GetUTXO(ctx, utxoIDs)
			if err != nil {
				f.logger.Errorf("Getting UTXO err: %v", err)
				continue
			}
			if len(utxos) == 0 {
				continue
			}
			return utxos, totalValue, nil
		}
	}
}

func (f *funder) blockingSendTx(ctx context.Context, client *localrpc.Client, tx *aobjs.Tx) error {
	sent := false
	for {
		for i := 0; i < 3; i++ {
			select {
			case <-f.ctx.Done():
				return ctx.Err()
			case <-time.After(1 * time.Second):
				if !sent {
					_, err := client.SendTransaction(ctx, tx)
					if err != nil {
						f.logger.Errorf("Sending Tx: %x got err: %v", tx.Vin[0].TXInLinker.TxHash, err)
						continue
					}
					f.logger.Infof("Sending Tx: %x", tx.Vin[0].TXInLinker.TxHash)
					sent = true
				}
				err := sleepWithContext(ctx, 1*time.Second)
				if err != nil {
					return err
				}
				_, err = client.GetMinedTransaction(ctx, tx.Vin[0].TXInLinker.TxHash)
				if err == nil {
					return nil
				}
				f.logger.Errorf("Checking mined Tx: %x got err: %v", tx.Vin[0].TXInLinker.TxHash, err)
			}
		}
		err := sleepWithContext(ctx, 10*time.Second)
		if err != nil {
			return err
		}
		_, err = client.GetPendingTransaction(ctx, tx.Vin[0].TXInLinker.TxHash)
		if err == nil {
			continue
		}
		f.logger.Errorf("Pending Tx: %x got err: %v", tx.Vin[0].TXInLinker.TxHash, err)
		_, err = client.GetMinedTransaction(ctx, tx.Vin[0].TXInLinker.TxHash)
		if err == nil {
			return nil
		}
		f.logger.Errorf("Checking mined Tx 2: %x got err: %v", tx.Vin[0].TXInLinker.TxHash, err)
	}
}

func main() {
	modePtr := flag.Uint(
		"s",
		0,
		"0 for Spam mode (default mode) which sends a lot of dataStore and valueStore txs and"+
			"1 for DataStore mode which inserts mode only send 3 data store transactions)",
	)
	workerNumberPtr := flag.Int("w", 100, "Number workers to send txs in the spam mode.")
	dataPtr := flag.String("m", "", "Data to write to state store in case dataStore mode.")
	dataIndexPtr := flag.String("i", "", "Index of data to write to state store in case dataStore mode.")
	dataDurationPtr := flag.Uint("e", 10, "Number of epochs that a data store will be stored.")
	ethereumEndPointPtr := flag.String(
		"e",
		"http://127.0.0.1:8545",
		"Endpoint to connect with the layer 1 server. If not provided, defaults to 127.0.0.1:8545",
	)
	factoryAddressPtr := flag.String(
		"f",
		"0x0b1f9c2b7bed6db83295c7b5158e3806d67ec5bc",
		"AliceNet Factory Address. If not provided, defaults to 0x0b1f9c2b7bed6db83295c7b5158e3806d67ec5bc",
	)
	flag.Parse()

	if !strings.Contains(*ethereumEndPointPtr, "https://") && !strings.Contains(*ethereumEndPointPtr, "http://") {
		panic("Incorrect endpoint. Endpoint should start with 'http://' or 'https://'")
	}
	if *modePtr > 1 {
		panic("Invalid mode, only 0 or 1 allowed!")
	}
	var logger *logrus.Entry
	if *modePtr == 0 {
		logger = logging.GetLogger("test").WithFields(logrus.Fields{
			"workers":        *workerNumberPtr,
			"factoryAddress": *factoryAddressPtr,
			"ethereum":       *ethereumEndPointPtr,
		})
	} else {
		logger = logging.GetLogger("test").WithFields(logrus.Fields{
			"dataStoreIndex": *dataIndexPtr,
			"initial data":   *dataPtr,
		})
	}

	logger.Logger.SetLevel(logrus.DebugLevel)

	spamMode := *modePtr == 0
	nodeList := []string{
		"127.0.0.1:8887",
		"127.0.0.1:9884",
		"127.0.0.1:9885",
		"127.0.0.1:9886",
	}

	mainCtx, cf := context.WithCancel(context.Background())
	defer cf()

	tempDir, err := os.MkdirTemp("", "spammerdir")
	if err != nil {
		panic(fmt.Errorf("failed to create tmp dir: %v", err))
	}
	defer os.RemoveAll(tempDir)

	recoverDB := mocks.NewTestDB()
	defer recoverDB.DB().Close()

	keyStorePath, passCodePath, accounts := tests.CreateAccounts(tempDir, 1)
	eth, err := ethereum.NewClient(
		*ethereumEndPointPtr,
		keyStorePath,
		passCodePath,
		accounts[0].Address.String(),
		false,
		2,
		500,
		0,
	)
	if err != nil {
		panic(fmt.Errorf("failed to create ethereum client: %v", err))
	}
	defer eth.Close()
	contracts := handlers.NewAllSmartContractsHandle(eth, common.HexToAddress(*factoryAddressPtr))

	watcher := transaction.WatcherFromNetwork(eth, recoverDB, false, constants.TxPollingTime)
	defer watcher.Close()

	funder, err := createNewFunder(mainCtx, logger, eth, accounts[0], watcher, contracts, nodeList)
	if err != nil {
		panic(err)
	}

	if spamMode {
		logger.Info("Running in Spam Mode")
		tx, err := funder.setupDataStoreTransaction(
			mainCtx,
			funder.signer,
			funder.acct,
			*dataPtr,
			*dataIndexPtr,
			uint32(*dataDurationPtr),
		)
		if err != nil {
			panic(err)
		}
		err = funder.blockingSendTx(mainCtx, funder.client, tx)
		if err != nil {
			panic(err)
		}
	} else {
		tx, err := funder.setupDataStoreTransaction(
			mainCtx,
			funder.signer,
			funder.acct,
			*dataPtr,
			*dataIndexPtr,
			uint32(*dataDurationPtr),
		)
		if err != nil {
			panic(err)
		}
		err = funder.blockingSendTx(mainCtx, funder.client, tx)
		if err != nil {
			panic(err)
		}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-signals:
		cf()
	}
	logger.Info("Exiting ...")
}
