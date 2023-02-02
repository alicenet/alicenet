package main

import (
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	aobjs "github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/evm"
	"github.com/alicenet/alicenet/layer1/handlers"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/localrpc"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

const (
	sentValuePerChild   = uint64(5_000000000000000000)
	ethDepositAmount    = "1000000000000000000000000"
	dataStoreSizeOffset = int(constants.MaxDataStoreSize / 200000)
	runesBytes          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-=*+;,/.^~][{}\"'\\|_"
)

type ErrNotEnoughFundsForDataStore struct {
	account       []byte
	valueOut      *uint256.Uint256
	consumedValue *uint256.Uint256
}

func (e *ErrNotEnoughFundsForDataStore) Error() string {
	return fmt.Sprintf("account %x does not have enough funding: requires:%v has:%v", e.account, e.valueOut, e.consumedValue)
}

type fees struct {
	minTxFee      *uint256.Uint256
	valueStoreFee *uint256.Uint256
	dataStoreFee  *uint256.Uint256
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = runesBytes[rand.Intn(len(runesBytes))]
	}
	return string(b)
}

func uint256ToDecimalString(v *uint256.Uint256) string {
	bigInt, err := v.ToBigInt()
	if err != nil {
		panic(err)
	}
	return bigInt.String()
}

func computeFinalDataStoreFee(dataStoreFee *uint256.Uint256, numEpochs32 uint32) *uint256.Uint256 {
	numEpochs, err := new(uint256.Uint256).FromUint64(uint64(numEpochs32))
	if err != nil {
		panic(err)
	}
	totalEpochs, err := new(uint256.Uint256).Add(uint256.Two(), numEpochs)
	if err != nil {
		panic(err)
	}
	totalFeeUint256, err := new(uint256.Uint256).Mul(dataStoreFee, totalEpochs)
	if err != nil {
		panic(err)
	}
	return totalFeeUint256
}

func sleepWithContext(ctx context.Context, sleepTime time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(sleepTime):
	}
	return nil
}

// setupRPCClient tries to connect to a random node passed in the list of `rpcNodeList`. In case it
// fails to connect to the chosen node, it tries the next until all nodes are tried. This function
// panics if it's not able to connect to any of the nodes.
func setupRPCClient(ctx context.Context, logger *logrus.Entry, rpcNodeList []string, idx int) ([]*localrpc.Client, error) {
	nodes := []*localrpc.Client{}
	// try to connect to any of available nodes
	for i := 0; i < len(rpcNodeList); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		logger.Tracef("trying to connect with: %v", rpcNodeList[i])
		client := &localrpc.Client{Address: rpcNodeList[i], TimeOut: constants.MsgTimeout}
		if err := client.Connect(ctx); err != nil {
			logger.WithError(err).Errorf("failed to connect with: %v", rpcNodeList[i])
			continue
		}
		nodes = append(nodes, client)
	}
	if len(nodes) == 0 {
		return nil, errors.New(fmt.Sprintf("failed to connect to any of the nodes: %v", rpcNodeList))
	}
	return nodes, nil
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

func computeValuePerUser(fees *fees, consumedValue *uint256.Uint256, numUsers uint64) *uint256.Uint256 {
	valuePerRecipient := &uint256.Uint256{}
	_, err := valuePerRecipient.Sub(consumedValue, fees.minTxFee.Clone())
	if err != nil {
		panic(err)
	}
	numUsersBigInt, err := new(uint256.Uint256).FromUint64(numUsers)
	if err != nil {
		panic(err)
	}
	totalVsFees, err := new(uint256.Uint256).Mul(fees.valueStoreFee.Clone(), numUsersBigInt)
	if err != nil {
		panic(err)
	}
	_, err = valuePerRecipient.Sub(valuePerRecipient, totalVsFees)
	if err != nil {
		panic(err)
	}
	_, err = valuePerRecipient.Div(valuePerRecipient, numUsersBigInt)
	if err != nil {
		panic(err)
	}
	return valuePerRecipient
}

type worker struct {
	f       *funder
	logger  *logrus.Entry
	client  *localrpc.Client
	signer  aobjs.Signer
	acct    []byte
	balance *uint256.Uint256
	idx     int
}

func (w *worker) run(ctx context.Context) {
	defer w.f.wg.Done()
	isToSendDataStore := w.idx%3 == 0
	for {
		err := sleepWithContext(w.f.ctx, time.Second*time.Duration(1+w.idx%10))
		if err != nil {
			w.logger.Errorf("exiting context finished: %v", err)
			return
		}
		w.logger.Trace("gettingFunding")
		utxos, totalValue, err := w.f.blockingGetFunding(w.client, getCurveSpec(w.signer), w.acct, w.balance)
		if err != nil {
			w.logger.Errorf("error at blockingGetFunding: %v", err)
			continue
		}
		w.logger.Tracef("gotFunding: %v", uint256ToDecimalString(totalValue))
		// sending the ALCB back to the funder
		funder := []*worker{{
			signer: w.f.signer,
			acct:   w.f.acct,
		}}
		var tx *aobjs.Tx
		if isToSendDataStore {
			// sending a data store with random data
			tx, err = w.f.createDataStoreTx(w.logger, w.signer, w.acct, totalValue, utxos, "", "", 0)
			if err != nil {
				w.logger.Errorf("error at setupTransaction dataStore: %v", err)
				// if we don't have enough ALCB to cover a datastore we fallback to a value store tx
				expectError := &ErrNotEnoughFundsForDataStore{}
				if errors.As(err, &expectError) {
					isToSendDataStore = false
				} else {
					continue
				}
			}
		}

		if tx == nil {
			tx, err = w.f.createValueStoreTx(w.logger, w.signer, w.acct, totalValue, utxos, nil, funder)
			if err != nil {
				w.logger.Errorf("error at setupTransaction valueStore: %v", err)
				continue
			}
		}
		err = w.f.blockingSendTx(w.client, w.logger, tx, w.f.sendTxAllClients)
		if err != nil {
			w.logger.Errorf("error at blockingSendTx: %v", err)
		}
		return
	}
}

type funder struct {
	ctx              context.Context
	wg               sync.WaitGroup
	logger           *logrus.Entry
	signer           *crypto.Secp256k1Signer
	mainClient       *localrpc.Client
	chainID          uint32
	acct             []byte
	ethAcct          accounts.Account
	ethClient        *evm.Client
	ethContracts     layer1.AllSmartContracts
	ethTxWatcher     *transaction.FrontWatcher
	clientList       []*localrpc.Client
	sendTxAllClients bool
}

func createNewFunder(
	ctx context.Context,
	logger *logrus.Entry,
	ethClient *evm.Client,
	ethAccount accounts.Account,
	ethWatcher *transaction.FrontWatcher,
	ethContracts layer1.AllSmartContracts,
	nodeList []string,
	sendTxAllNodes bool,
) (*funder, error) {
	logger.Info("Funder setting up signing")
	signer, acct, err := setupHexSigner(tests.TestAdminPrivateKey)
	if err != nil {
		logger.Errorf("Funder error at setupHexSigner: %v", err)
		return nil, err
	}
	logger.Info("Funder setting up client")
	nodes, err := setupRPCClient(ctx, logger, nodeList, 0)
	if err != nil {
		return nil, err
	}
	blockHeader, err := nodes[0].GetBlockHeader(ctx, 1)
	if err != nil {
		logger.Fatalf("Failed to get block header and chain id: %v", err)
	}
	chainID := blockHeader.BClaims.ChainID

	return &funder{
		ctx:              ctx,
		wg:               sync.WaitGroup{},
		logger:           logger,
		signer:           signer,
		mainClient:       nodes[0],
		chainID:          chainID,
		acct:             acct,
		ethAcct:          ethAccount,
		ethClient:        ethClient,
		ethContracts:     ethContracts,
		ethTxWatcher:     ethWatcher,
		clientList:       nodes,
		sendTxAllClients: sendTxAllNodes,
	}, nil
}

func (f *funder) sendDataStores(numDataStores uint32, msg string, index string, numEpochs uint32) {
	isDepositDone := false
	minimumValue, err := new(uint256.Uint256).FromUint64(sentValuePerChild)
	if err != nil {
		f.logger.Fatalf("couldn't convert uint256 in datastore mode: %v", err)
	}
	for i := uint32(0); i < numDataStores; i++ {
		for {
			select {
			case <-f.ctx.Done():
				return
			default:
			}
			if !isDepositDone {
				err := f.mintALCBDepositOnEthereum()
				if err != nil {
					f.logger.Errorf("error at mintALCBDepositOnEthereum: %v", err)
					continue
				}
				isDepositDone = true
			}
			utxos, totalValue, err := f.blockingGetFunding(f.mainClient, getCurveSpec(f.signer), f.acct, minimumValue)
			if err != nil {
				f.logger.Errorf("error at blockingGetFunding: %v", err)
				continue
			}
			if msg != "" {
				msg += fmt.Sprintf("-myIndex:%d", i)
			}
			if index != "" {
				index += fmt.Sprintf("-myIndex:%d", i)
			}
			tx, err := f.createDataStoreTx(f.logger, f.signer, f.acct, totalValue, utxos, msg, index, numEpochs)
			if err != nil {
				f.logger.Fatalf("failed to create data store tx: %v", err)
			}
			err = f.blockingSendTx(f.mainClient, f.logger, tx, f.sendTxAllClients)
			if err != nil {
				f.logger.Fatalf("failed to send tx %v", err)
			}
			break
		}
	}
}

func (f *funder) startSpamming(numWorkers int) {
	baseIdx := 0

	sentValuePerChildBigInt, err := new(uint256.Uint256).FromUint64(sentValuePerChild)
	if err != nil {
		f.logger.Fatalf("failed to convert uint256-2 %v", err)
	}
	// sending the double per worker (1 ALCB as value and the rest to make sure that we cover fees)
	fundingPerChild, err := new(uint256.Uint256).FromUint64(sentValuePerChild * 2)
	if err != nil {
		f.logger.Fatalf("failed to convert uint256-3 %v", err)
	}

	for {
		select {
		case <-f.ctx.Done():
			return
		default:
		}

		children, err := f.setupChildren(f.ctx, numWorkers, baseIdx)
		if err != nil {
			f.logger.Fatalf("Funder error at setupChildren: %v", err)
		}
		f.logger.Info("Funder setting up funding")
		chunkSize := 50
		if len(children) > chunkSize {
			chuckSizeUint256, err := new(uint256.Uint256).FromUint64(uint64(chunkSize))
			if err != nil {
				f.logger.Fatalf("failed to convert uint256-1 %v", err)
			}
			totalFunding, err := new(uint256.Uint256).Mul(fundingPerChild, chuckSizeUint256)
			if err != nil {
				f.logger.Fatalf("failed to convert uint256-4 %v", err)
			}
			for i := 0; i <= len(children)/chunkSize; i++ {
				//batchSize := chunkSize
				start := i * chunkSize
				end := (i + 1) * chunkSize
				if end > len(children) {
					end = len(children)
				}

				utxos, value, err := f.blockingGetFunding(f.mainClient, getCurveSpec(f.signer), f.acct, totalFunding)
				if err != nil {
					f.logger.Errorf("error at blockingGetFunding1: %v", err)
					continue
				}
				for {
					select {
					case <-f.ctx.Done():
						return
					case <-time.After(time.Second):
					}
					if value.Gte(totalFunding) {
						break
					}
					f.logger.Info("Not enough funds, creating deposit on ethereum")
					err := f.mintALCBDepositOnEthereum()
					if err != nil {
						f.logger.Errorf("error at mintALCBDepositOnEthereum: %v", err)
						continue
					}
					utxos, value, err = f.blockingGetFunding(f.mainClient, getCurveSpec(f.signer), f.acct, totalFunding)
					if err != nil {
						f.logger.Errorf("error at blockingGetFunding2: %v", err)
						continue
					}
				}
				f.logger.Infof("Funder setting up tx with %v utxos and total value: %v", len(utxos), uint256ToDecimalString(value))

				tx, err := f.createValueStoreTx(f.logger, f.signer, f.acct, value, utxos, sentValuePerChildBigInt, children[start:end])
				if err != nil {
					f.logger.Errorf("Funder error at setupTransaction: %v", err)
					continue
				}
				f.logger.Info("Funder setting up blockingSendTx")
				if err := f.blockingSendTx(f.mainClient, f.logger, tx, false); err != nil {
					f.logger.Errorf("Funder error at blockingSendTx: %v", err)
					continue
				}
			}
		}

		f.logger.Info("Funder starting children")
		for _, c := range children {
			f.wg.Add(1)
			go c.run(f.ctx)
		}
		f.logger.Info("Funder waiting for workers to send txs")
		f.wg.Wait()
		baseIdx += numWorkers
		f.logger.Info("All workers finished sending txs!")
	}
}

func (f *funder) setupChildren(ctx context.Context, numChildren int, baseIdx int) ([]*worker, error) {
	workers := []*worker{}
	for i := 0; i < numChildren; i++ {
		client := f.clientList[i%len(f.clientList)]
		signer, acct, err := setupTestingSigner(baseIdx + i)
		if err != nil {
			return nil, err
		}
		logger := f.logger.WithFields(*&logrus.Fields{
			"workerID": i,
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
	feesString, err := f.mainClient.GetTxFees(f.ctx)
	if err != nil {
		return nil, err
	}
	if len(feesString) != 3 {
		f.logger.Fatal("invalid fee response")
	}
	minTxFee := new(uint256.Uint256)
	err = minTxFee.UnmarshalString(feesString[0])
	if err != nil {
		f.logger.Fatalf("failed to decode minTx fee %v", err)
	}
	vsFee := new(uint256.Uint256)
	err = vsFee.UnmarshalString(feesString[1])
	if err != nil {
		f.logger.Fatalf("failed to decode valueStoreTx fee %v", err)
	}
	dsFee := new(uint256.Uint256)
	err = dsFee.UnmarshalString(feesString[2])
	if err != nil {
		f.logger.Fatalf("failed to decode dataStoreTx fee %v", err)
	}
	return &fees{minTxFee, vsFee, dsFee}, nil
}

func (f *funder) createValueStoreTx(
	logger *logrus.Entry,
	signer aobjs.Signer,
	ownerAcct []byte,
	consumedValue *uint256.Uint256,
	consumedUtxos aobjs.Vout,
	valuePerRecipient *uint256.Uint256,
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
	valueOut := fees.minTxFee.Clone()
	if valuePerRecipient == nil {
		valuePerRecipient = computeValuePerUser(fees, consumedValue, uint64(len(recipients)))
	}
	for _, r := range recipients {
		newOwner := &aobjs.ValueStoreOwner{}
		newOwner.New(r.acct, getCurveSpec(r.signer))
		newValueStore := &aobjs.ValueStore{
			VSPreImage: &aobjs.VSPreImage{
				ChainID:  chainID,
				Value:    valuePerRecipient.Clone(),
				Owner:    newOwner,
				TXOutIdx: 0, //place holder for now
				Fee:      fees.valueStoreFee.Clone(),
			},
			TxHash: make([]byte, constants.HashLen),
		}
		valuePlusFee := &uint256.Uint256{}
		_, err := valuePlusFee.Add(valuePerRecipient.Clone(), fees.valueStoreFee.Clone())
		if err != nil {
			logger.Fatalf("failed to convert uint256-1: %v", err)
		}
		_, err = valueOut.Add(valueOut, valuePlusFee)
		if err != nil {
			logger.Fatalf("failed to convert uint256-2: %v", err)
		}

		newUTXO := &aobjs.TXOut{}
		err = newUTXO.NewValueStore(newValueStore)
		if err != nil {
			return nil, err
		}
		tx.Vout = append(tx.Vout, newUTXO)
		r.balance = valuePerRecipient
	}
	return f.finalizeTx(
		logger,
		tx,
		signer,
		consumedUtxos,
		fees,
		consumedValue,
		valueOut,
	)
}

func (f *funder) createDataStoreTx(
	logger *logrus.Entry,
	signer aobjs.Signer,
	ownerAcct []byte,
	consumedValue *uint256.Uint256,
	consumedUtxos aobjs.Vout,
	msg string,
	indexStr string,
	numEpochs uint32,
) (*aobjs.Tx, error) {

	if msg == "" {
		size := rand.Intn(dataStoreSizeOffset) + dataStoreSizeOffset
		msg = randStringBytes(int(size))
		logger.Trace("No data was passed, generating random data for datastore")
	}

	if indexStr == "" {
		indexStr = randStringBytes(20)
		logger.Trace("No index was passed, generating random index for datastore")
	}
	index := crypto.Hasher([]byte(indexStr))

	if numEpochs == 0 {
		numEpochs = uint32(rand.Int31n(10) + 2)
		logger.Tracef("No duration was passed, generating random duration (epochs): %v", numEpochs)
	}

	fees, err := f.getTxFees()
	if err != nil {
		return nil, err
	}
	deposit, err := aobjs.BaseDepositEquation(uint32(len(msg)), numEpochs)
	if err != nil {
		logger.Fatal(err)
	}
	valueOut := fees.minTxFee.Clone()
	currentEpoch, err := f.mainClient.GetEpochNumber(f.ctx)
	if err != nil {
		return nil, err
	}
	_, err = valueOut.Add(valueOut, deposit)
	if err != nil {
		logger.Fatal(err)
	}
	newOwner := &aobjs.DataStoreOwner{}
	newOwner.New(ownerAcct, getCurveSpec(signer))
	dataStoreFinalFee := computeFinalDataStoreFee(fees.dataStoreFee, numEpochs)
	_, err = valueOut.Add(dataStoreFinalFee, deposit)
	if err != nil {
		logger.Fatal(err)
	}
	if consumedValue.Lt(valueOut) {
		return nil, &ErrNotEnoughFundsForDataStore{ownerAcct, valueOut, consumedValue}
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
	logger.Tracef("The len of vin %v ", len(tx.Vin))
	newDataStore := &aobjs.DataStore{
		DSLinker: &aobjs.DSLinker{
			DSPreImage: &aobjs.DSPreImage{
				ChainID:  f.chainID,
				Index:    index,
				IssuedAt: currentEpoch,
				Deposit:  deposit,
				RawData:  []byte(msg),
				TXOutIdx: 0,
				Owner:    newOwner,
				Fee:      dataStoreFinalFee.Clone(),
			},
			TxHash: make([]byte, constants.HashLen),
		},
	}
	eoe, err := newDataStore.EpochOfExpiration()
	if err != nil {
		return nil, err
	}
	logger.Tracef("DS:index:%x  deposit:%v EpochOfExpire:%v", index, uint256ToDecimalString(deposit), eoe)
	newUTXO := &aobjs.TXOut{}
	err = newUTXO.NewDataStore(newDataStore)
	if err != nil {
		return nil, err
	}
	tx.Vout = append(tx.Vout, newUTXO)
	return f.finalizeTx(
		logger,
		tx,
		signer,
		consumedUtxos,
		fees,
		consumedValue,
		valueOut,
	)
}

func (f *funder) finalizeTx(
	logger *logrus.Entry,
	tx *aobjs.Tx,
	signer aobjs.Signer,
	consumedUtxos aobjs.Vout,
	fees *fees,
	consumedValue *uint256.Uint256,
	valueOut *uint256.Uint256,
) (*aobjs.Tx, error) {
	// if we have dust, check if dust is greater than fees, if yes send back as change to the funder.
	// Otherwise, burn as fee
	if consumedValue.Gt(valueOut) {
		diff, err := new(uint256.Uint256).Sub(consumedValue.Clone(), valueOut.Clone())
		if err != nil {
			logger.Fatal(err)
		}
		if diff.Lte(fees.valueStoreFee.Clone()) {
			newFee, err := new(uint256.Uint256).Add(fees.minTxFee.Clone(), diff)
			if err != nil {
				logger.Fatal(err)
			}
			tx.Fee = newFee
		} else {
			newOwner := &aobjs.ValueStoreOwner{}
			newOwner.New(f.acct, getCurveSpec(f.signer))
			newValueStore := &aobjs.ValueStore{
				VSPreImage: &aobjs.VSPreImage{
					ChainID:  f.chainID,
					Value:    diff,
					Owner:    newOwner,
					TXOutIdx: 0,
					Fee:      fees.valueStoreFee.Clone(),
				},
				TxHash: make([]byte, constants.HashLen),
			}
			newUTXO := &aobjs.TXOut{}
			err = newUTXO.NewValueStore(newValueStore)
			if err != nil {
				logger.Fatal(err)
			}
			tx.Vout = append(tx.Vout, newUTXO)
		}

		_, err = valueOut.Add(valueOut, diff)
		if err != nil {
			logger.Fatal(err)
		}
	}
	err := tx.SetTxHash()
	if err != nil {
		logger.Fatal(err)
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
				logger.Fatal(err)
			}
		case consumedUtxo.HasDataStore():
			consumedDS, err := consumedUtxo.DataStore()
			if err != nil {
				return nil, err
			}
			txIn := tx.Vin[idx]
			err = consumedDS.Sign(txIn, signer)
			if err != nil {
				logger.Fatal(err)
			}
		}
	}
	txHash, err := tx.TxHash()
	if err != nil {
		logger.Fatalf("Could not get txhash %v", err)
	}
	logger.Tracef("Consumed Value: %v  ValueOut: %v TxHash:%x", uint256ToDecimalString(consumedValue), uint256ToDecimalString(valueOut), txHash)
	txb, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	logger.Tracef("TxHash: %x, TX SIZE: %v", txHash, len(txb))
	return tx, nil
}

func (f *funder) mintALCBDepositOnEthereum() error {
	txnOpts, err := f.ethClient.GetTransactionOpts(f.ctx, f.ethAcct)
	if err != nil {
		return fmt.Errorf("failed to get transaction options: %v", err)
	}
	// 1_000_000 ALCB (10**18)
	depositAmount, ok := new(big.Int).SetString(ethDepositAmount, 10)
	if !ok {
		f.logger.Fatal("Could not generate deposit amount")
	}
	bTokenABI, err := abi.JSON(strings.NewReader(bindings.ALCBMetaData.ABI))
	if err != nil {
		return err
	}
	input, err := bTokenABI.Pack(
		"virtualMintDeposit",
		uint8(1),
		f.ethAcct.Address,
		depositAmount,
	)
	if err != nil {
		return err
	}
	txn, err := f.ethContracts.EthereumContracts().ContractFactory().CallAny(
		txnOpts,
		f.ethContracts.EthereumContracts().ALCBAddress(),
		big.NewInt(0),
		input,
	)
	if err != nil {
		return fmt.Errorf("could not send deposit amount to ethereum %v", err)
	}
	_, err = f.ethTxWatcher.SubscribeAndWait(f.ctx, txn, nil)
	if err != nil {
		return fmt.Errorf("failed to get receipt: %v", err)
	}
	return nil
}

func (f *funder) blockingGetFunding(client *localrpc.Client, curveSpec constants.CurveSpec, acct []byte, value *uint256.Uint256) (aobjs.Vout, *uint256.Uint256, error) {
	for {
		select {
		case <-f.ctx.Done():
			return nil, nil, f.ctx.Err()
		case <-time.After(1 * time.Second):
			utxoIDs, totalValue, err := client.GetValueForOwner(f.ctx, curveSpec, acct, value)
			if err != nil {
				f.logger.Errorf("Getting fund err: %v", err)
				continue
			}
			utxos, err := client.GetUTXO(f.ctx, utxoIDs)
			if err != nil {
				f.logger.Errorf("Getting UTXO err: %v", err)
				continue
			}
			return utxos, totalValue, nil
		}
	}
}

func (f *funder) blockingSendTx(client *localrpc.Client, logger *logrus.Entry, tx *aobjs.Tx, sendTxAllClients bool) error {
	sent := false
	for {
		for i := 0; i < 5; i++ {
			select {
			case <-f.ctx.Done():
				return f.ctx.Err()
			case <-time.After(1 * time.Second):
			}
			if !sent {
				if sendTxAllClients {
					sentTxCount := 0
					for _, iterClient := range f.clientList {
						_, err := iterClient.SendTransaction(f.ctx, tx)
						if err != nil {
							logger.Errorf("Sending Tx: %x got err: %v", tx.Vin[0].TXInLinker.TxHash, err)
							continue
						}
						logger.Tracef("Sending Tx: %x in node: %v", tx.Vin[0].TXInLinker.TxHash, iterClient.Address)
						sentTxCount++
					}
					if sentTxCount == 0 {
						continue
					}
					sent = true
				} else {
					_, err := client.SendTransaction(f.ctx, tx)
					if err != nil {
						logger.Errorf("Sending Tx: %x got err: %v", tx.Vin[0].TXInLinker.TxHash, err)
						continue
					}
					logger.Infof("Sending Tx: %x", tx.Vin[0].TXInLinker.TxHash)
					sent = true
				}

			}
			err := sleepWithContext(f.ctx, 7*time.Second)
			if err != nil {
				return err
			}
			_, err = client.GetMinedTransaction(f.ctx, tx.Vin[0].TXInLinker.TxHash)
			if err == nil {
				logger.Infof("Tx successfully mined: %x", tx.Vin[0].TXInLinker.TxHash)
				return nil
			}
		}
		err := sleepWithContext(f.ctx, 10*time.Second)
		if err != nil {
			return err
		}
		_, err = client.GetPendingTransaction(f.ctx, tx.Vin[0].TXInLinker.TxHash)
		if err == nil {
			continue
		}
		_, err2 := client.GetMinedTransaction(f.ctx, tx.Vin[0].TXInLinker.TxHash)
		if err2 == nil {
			logger.Infof("Tx successfully mined: %x", tx.Vin[0].TXInLinker.TxHash)
			return nil
		}
		return fmt.Errorf("Tx was not mined nor is in pending: %x got err1: %v and err2: %v", tx.Vin[0].TXInLinker.TxHash, err, err2)
	}
}

func main() {
	modePtr := flag.Uint(
		"mode",
		0,
		"0 for Spam mode (default mode) which sends a lot of dataStore and valueStore txs and"+
			"1 for DataStore mode which inserts mode only send 3 data store transactions)",
	)
	workerNumberPtr := flag.Int("workers", 10, "Number workers to send txs in the spam mode.")
	dataPtr := flag.String("data", "", "Data to write to state store in case dataStore mode.")
	dataIndexPtr := flag.String("index", "", "Index of data to write to state store in case dataStore mode.")
	dataDurationPtr := flag.Uint("duration", 0, "Number of epochs that a data store will be stored.")
	dataQuantityPtr := flag.Uint("amount", 5, "Number of dataStores to send in the dataStore mode.")
	ethereumEndPointPtr := flag.String(
		"endpoint",
		"http://127.0.0.1:8545",
		"Endpoint to connect with the layer 1 server. If not provided, defaults to 127.0.0.1:8545",
	)
	factoryAddressPtr := flag.String(
		"factory",
		"0x77D7c620E3d913AA78a71acffA006fc1Ae178b66",
		"AliceNet Factory Address. If not provided, defaults to 0x77D7c620E3d913AA78a71acffA006fc1Ae178b66",
	)
	sendTxToAllClientsPtr := flag.Bool(
		"sendTxToAllClients",
		false,
		"If a tx should be relayed to all clients at the same time. Can be used in both modes.",
	)
	verbosePtr := flag.Bool(
		"v",
		false,
		"Should execute with verbose terminal output.",
	)
	flag.Parse()

	logger := logging.GetLogger("test").WithFields(logrus.Fields{
		"factoryAddress": *factoryAddressPtr,
		"ethereum":       *ethereumEndPointPtr,
	})
	if *modePtr == 0 {
		logger = logging.GetLogger("test").WithFields(logrus.Fields{
			"workers": *workerNumberPtr,
		})
	} else {
		logger = logging.GetLogger("test").WithFields(logrus.Fields{
			"txQuantity": dataQuantityPtr,
		})
	}

	rand.Seed(time.Now().UnixNano())

	if *verbosePtr {
		logger.Logger.SetLevel(logrus.TraceLevel)
	}

	if !strings.Contains(*ethereumEndPointPtr, "https://") && !strings.Contains(*ethereumEndPointPtr, "http://") {
		logger.Fatalf("Incorrect endpoint. Endpoint should start with 'http://' or 'https://'")
	}
	if *modePtr > 1 {
		logger.Fatalf("Invalid mode, only 0 or 1 allowed!")
	}

	spamMode := *modePtr == 0
	// make sure to have nodes listening in these ports
	nodeList := []string{
		"127.0.0.1:9884",
	}

	mainCtx, cf := context.WithCancel(context.Background())
	defer cf()

	tempDir, err := os.MkdirTemp("", "spammerdir")
	if err != nil {
		logger.Fatalf("failed to create tmp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	recoverDB := mocks.NewTestDB()
	defer recoverDB.DB().Close()

	keyStorePath, passCodePath, accounts := tests.CreateAccounts(tempDir, 1)
	eth, err := evm.NewClient(logger.Logger,
		*ethereumEndPointPtr,
		keyStorePath,
		passCodePath,
		accounts[0].Address.String(),
		false,
		constants.EthereumFinalityDelay+1,
		500,
		0,
	)
	if err != nil {
		logger.Fatalf("failed to create ethereum client: %v", err)
	}
	defer eth.Close()
	contracts := handlers.NewAllSmartContractsHandle(eth, common.HexToAddress(*factoryAddressPtr))

	watcher := transaction.WatcherFromNetwork(eth, recoverDB, false, constants.TxPollingTime)
	defer watcher.Close()

	funder, err := createNewFunder(
		mainCtx,
		logger,
		eth,
		accounts[0],
		watcher,
		contracts,
		nodeList,
		*sendTxToAllClientsPtr,
	)
	if err != nil {
		logger.Fatalf("failed to create funder: %v", err)
	}

	// channel for closing the app
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	if spamMode {
		logger.Info("Running in Spam Mode")
		go funder.startSpamming(*workerNumberPtr)
	} else {
		logger.Logger.SetLevel(logrus.TraceLevel)
		logger.Info("Running in DataStore Insertion Mode")
		go funder.sendDataStores(uint32(*dataQuantityPtr), *dataPtr, *dataIndexPtr, uint32(*dataDurationPtr))
	}

	select {
	case <-signals:
		cf()
	}
	logger.Info("Exiting ...")
}
