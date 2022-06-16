package main

import (
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	aobjs "github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/localrpc"
	"github.com/MadBase/MadNet/utils"
)

var numEpochs uint32 = 1

var errClosing = errors.New("closing")

type worker struct {
	f      *funder
	client *localrpc.Client
	signer aobjs.Signer
	acct   []byte
	idx    int
}

func (w *worker) run(ctx context.Context) {
	defer w.f.cf()
	defer w.f.wg.Done()
	time.Sleep(time.Second * time.Duration(1+w.idx%10))
	for {
		select {
		case <-w.f.ctx.Done():
			return
		default:
		}
		fmt.Printf("Worker:%v gettingFunding\n", w.idx)
		u, v, err := w.f.blockingGetFunding(ctx, w.client, w.f.getCurveSpec(w.signer), w.acct, uint256.One())
		if err != nil {
			fmt.Printf("Worker%v error at blockingGetFunding: %v\n", w.idx, err)
			return
		}
		fmt.Printf("Worker:%v gotFunding: %v\n", w.idx, v)
		select {
		case <-w.f.ctx.Done():
			return
		default:
		}
		tx, err := w.f.setupTransaction(w.signer, w.acct, v, u, nil)
		if err != nil {
			fmt.Printf("Worker%v error at setupTransaction: %v\n", w.idx, err)
			return
		}
		select {
		case <-w.f.ctx.Done():
			return
		default:
		}
		fmt.Printf("Worker:%v sending Tx:\n", w.idx)
		err = w.f.blockingSendTx(ctx, w.client, tx)
		if err != nil {
			fmt.Printf("Worker:%v error at blockingSendTx: %v\n", w.idx, err)
			return
		}
	}
}

type funder struct {
	ctx         context.Context
	cf          func()
	wg          sync.WaitGroup
	signer      *crypto.Secp256k1Signer
	client      *localrpc.Client
	acct        []byte
	numChildren int
	children    []*worker
	nodeList    []string
	baseIdx     int
}

func (f *funder) doSpam(privk string, numChildren int, nodeList []string, baseIdx int) error {
	f.numChildren = numChildren
	f.baseIdx = baseIdx
	f.nodeList = nodeList
	ctx := context.Background()
	subCtx, cf := context.WithCancel(ctx)
	f.ctx = subCtx
	f.cf = cf
	defer f.cf()
	f.wg = sync.WaitGroup{}
	fmt.Printf("Funder setting up signing\n")
	signer, acct, err := f.setupHexSigner(privk)
	if err != nil {
		fmt.Printf("Funder error at setupHexSigner: %v\n", err)
		return err
	}
	f.signer = signer
	f.acct = acct
	fmt.Printf("Funder setting up client\n")
	client, err := f.setupClient(ctx, 0)
	if err != nil {
		fmt.Printf("Funder error at setupClient: %v\n", err)
		return err
	}
	f.client = client
	fmt.Printf("Funder setting up children\n")
	children, err := f.setupChildren(ctx, f.numChildren, f.baseIdx)
	if err != nil {
		fmt.Printf("Funder error at setupChildren: %v\n", err)
		return err
	}
	f.children = children
	fmt.Printf("Funder setting up funding\n")
	//numStuff := &uint256.Uint256{}
	numStuff, err := new(uint256.Uint256).FromUint64(uint64(len(children)))
	if err != nil {
		panic(err)
	}
	//utxos, value, err := f.blockingGetFunding(f.client, f.getCurveSpec(f.signer), f.acct, uint64(len(children)))
	utxos, value, err := f.blockingGetFunding(ctx, f.client, f.getCurveSpec(f.signer), f.acct, numStuff)
	if err != nil {
		fmt.Printf("Funder error at blockingGetFunding: %v\n", err)
		return err
	}
	fmt.Printf("Funder setting up tx\n")
	tx, err := f.setupTransaction(f.signer, f.acct, value, utxos, f.children)
	if err != nil {
		fmt.Printf("Funder error at setupTransaction: %v\n", err)
		return err
	}
	fmt.Printf("Funder setting up blockingSendTx\n")
	if err := f.blockingSendTx(ctx, f.client, tx); err != nil {
		fmt.Printf("Funder error at blockingSendTx: %v\n", err)
		return err
	}
	return nil
}

func (f *funder) init(privk string, numChildren int, nodeList []string, baseIdx int) error {
	f.numChildren = numChildren
	f.baseIdx = baseIdx
	f.nodeList = nodeList
	ctx := context.Background()
	subCtx, cf := context.WithCancel(ctx)
	f.ctx = subCtx
	f.cf = cf
	defer f.cf()
	f.wg = sync.WaitGroup{}
	fmt.Printf("Funder setting up signing\n")
	signer, acct, err := f.setupHexSigner(privk)
	if err != nil {
		fmt.Printf("Funder error at setupHexSigner: %v\n", err)
		return err
	}
	f.signer = signer
	f.acct = acct
	fmt.Printf("Funder setting up client\n")
	client, err := f.setupClient(ctx, 0)
	if err != nil {
		fmt.Printf("Funder error at setupClient: %v\n", err)
		return err
	}
	f.client = client
	fmt.Printf("Funder setting up children\n")
	children, err := f.setupChildren(ctx, f.numChildren, f.baseIdx)
	if err != nil {
		fmt.Printf("Funder error at setupChildren: %v\n", err)
		return err
	}
	f.children = children
	fmt.Printf("Funder setting up funding\n")
	//numStuff := &uint256.Uint256{}
	numStuff, err := new(uint256.Uint256).FromUint64(uint64(len(children)))
	if err != nil {
		panic(err)
	}
	//utxos, value, err := f.blockingGetFunding(f.client, f.getCurveSpec(f.signer), f.acct, uint64(len(children)))
	utxos, value, err := f.blockingGetFunding(ctx, f.client, f.getCurveSpec(f.signer), f.acct, numStuff)
	if err != nil {
		fmt.Printf("Funder error at blockingGetFunding: %v\n", err)
		return err
	}
	fmt.Printf("Funder setting up tx\n")
	tx, err := f.setupTransaction(f.signer, f.acct, value, utxos, f.children)
	if err != nil {
		fmt.Printf("Funder error at setupTransaction: %v\n", err)
		return err
	}
	fmt.Printf("Funder setting up blockingSendTx\n")
	if err := f.blockingSendTx(ctx, f.client, tx); err != nil {
		fmt.Printf("Funder error at blockingSendTx: %v\n", err)
		return err
	}
	fmt.Printf("Funder starting children\n")
	for _, c := range f.children {
		f.wg.Add(1)
		go c.run(ctx)
	}
	fmt.Printf("Funder waiting on close\n")
	<-f.ctx.Done()
	f.wg.Wait()
	return nil
}

func (f *funder) setupClient(ctx context.Context, idx int) (*localrpc.Client, error) {
	idx2 := idx % len(f.nodeList)
	client := &localrpc.Client{Address: f.nodeList[idx2], TimeOut: constants.MsgTimeout}
	if err := client.Connect(ctx); err != nil {
		return nil, err
	}
	return client, nil
}

func (f *funder) setupTestingSigner(i int) (aobjs.Signer, []byte, error) {
	privk := crypto.Hasher([]byte(strconv.Itoa(i)))
	if i%2 == 0 {
		return f.setupSecpSigner(privk)
	}
	return f.setupBNSigner(privk)
}

func (f *funder) setupBNSigner(privk []byte) (*crypto.BNSigner, []byte, error) {
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

func (f *funder) setupHexSigner(privk string) (*crypto.Secp256k1Signer, []byte, error) {
	privkb, err := hex.DecodeString(privk)
	if err != nil {
		return nil, nil, err
	}
	return f.setupSecpSigner(privkb)
}

func (f *funder) setupSecpSigner(privk []byte) (*crypto.Secp256k1Signer, []byte, error) {
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

func (f *funder) getCurveSpec(s aobjs.Signer) constants.CurveSpec {
	curveSpec := constants.CurveSpec(0)
	switch s.(type) {
	case *crypto.Secp256k1Signer:
		fmt.Println("secp")
		curveSpec = constants.CurveSecp256k1
	case *crypto.BNSigner:
		fmt.Println("bn")
		curveSpec = constants.CurveBN256Eth
	default:
		panic("invalid signer type")
	}
	return curveSpec
}

func (f *funder) setupTransaction(signer aobjs.Signer, ownerAcct []byte, consumedValue *uint256.Uint256, consumedUtxos aobjs.Vout, recipients []*worker) (*aobjs.Tx, error) {
	feesString, err := f.client.GetTxFees(f.ctx)
	if err != nil {
		panic(err)
	}
	if len(feesString) != 4 {
		panic("invalid fee response")
	}
	minTxFee := new(uint256.Uint256)
	vsFee := new(uint256.Uint256)
	//dsEpochFee := new(uint256.Uint256)
	//asFee := new(uint256.Uint256)
	err = minTxFee.UnmarshalString(feesString[0])
	if err != nil {
		panic(err)
	}
	err = vsFee.UnmarshalString(feesString[1])
	if err != nil {
		panic(err)
	}
	//err = dsEpochFee.UnmarshalString(feesString[2])
	//err = asFee.UnmarshalString(feesString[3])
	tx := &aobjs.Tx{
		Vin:  aobjs.Vin{},
		Vout: aobjs.Vout{},
		Fee:  minTxFee.Clone(),
	}
	chainID := uint32(42)
	for _, utxo := range consumedUtxos {
		consumedVS, err := utxo.ValueStore()
		if err != nil {
			return nil, err
		}
		chainID, err = consumedVS.ChainID()
		if err != nil {
			return nil, err
		}
		txIn, err := utxo.MakeTxIn()
		if err != nil {
			return nil, err
		}
		tx.Vin = append(tx.Vin, txIn)
	}
	// We include txFee here!
	valueOut := minTxFee.Clone()
	for _, r := range recipients {
		value := uint256.One()
		newOwner := &aobjs.ValueStoreOwner{}
		newOwner.New(r.acct, f.getCurveSpec(r.signer))
		newValueStore := &aobjs.ValueStore{
			VSPreImage: &aobjs.VSPreImage{
				ChainID:  chainID,
				Value:    value.Clone(),
				Owner:    newOwner,
				TXOutIdx: 0,
				Fee:      vsFee.Clone(),
			},
			TxHash: make([]byte, constants.HashLen),
		}
		valuePlusFee := &uint256.Uint256{}
		_, err := valuePlusFee.Add(value, vsFee)
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
		/*
			// Previous code; keep because we may need it
			newOwner := &aobjs.ValueStoreOwner{}
			newOwner.New(ownerAcct, f.getCurveSpec(signer))
			newValueStore := &aobjs.ValueStore{
				VSPreImage: &aobjs.VSPreImage{
					ChainID: chainID,
					//Value:    consumedValue - valueOut,
					Value:    diff,
					Owner:    newOwner,
					TXOutIdx: 0,
					Fee:      vsFee.Clone(),
				},
				TxHash: make([]byte, constants.HashLen),
			}
			newUTXO := &aobjs.TXOut{}
			newUTXO.NewValueStore(newValueStore)
			tx.Vout = append(tx.Vout, newUTXO)
		*/
		// Add difference to TxFee
		_, err = tx.Fee.Add(tx.Fee, diff)
		if err != nil {
			panic(err)
		}
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
	fmt.Printf("TX SIZE: %v\n", len(txb))
	return tx, nil
}

func (f *funder) setupChildren(ctx context.Context, numChildren int, baseIdx int) ([]*worker, error) {
	workers := []*worker{}
	for i := 0; i < numChildren; i++ {
		client, err := f.setupClient(ctx, baseIdx+i)
		if err != nil {
			return nil, err
		}
		signer, acct, err := f.setupTestingSigner(baseIdx + i)
		if err != nil {
			return nil, err
		}
		c := &worker{
			f:      f,
			signer: signer,
			acct:   acct,
			client: client,
			idx:    baseIdx + i,
		}
		workers = append(workers, c)
	}
	return workers, nil
}

func (f *funder) blockingGetFunding(ctx context.Context, client *localrpc.Client, curveSpec constants.CurveSpec, acct []byte, value *uint256.Uint256) (aobjs.Vout, *uint256.Uint256, error) {
	for {
		select {
		case <-f.ctx.Done():
			return nil, nil, errClosing
		case <-time.After(1 * time.Second):
			utxoIDs, totalValue, err := client.GetValueForOwner(ctx, curveSpec, acct, value)
			if err != nil {
				fmt.Printf("Getting fund err: %v\n", err)
				continue
			}
			utxos, err := client.GetUTXO(ctx, utxoIDs)
			if err != nil {
				fmt.Printf("Getting UTXO err: %v\n", err)
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
				return nil
			case <-time.After(1 * time.Second):
				if !sent {
					_, err := client.SendTransaction(ctx, tx)
					if err != nil {
						fmt.Printf("Sending Tx: %x got err: %v\n", tx.Vin[0].TXInLinker.TxHash, err)
						continue
					}
					if err == nil {
						fmt.Printf("Sending Tx: %x\n", tx.Vin[0].TXInLinker.TxHash)
						sent = true
					}
				}
				time.Sleep(1 * time.Second)
				_, err := client.GetMinedTransaction(ctx, tx.Vin[0].TXInLinker.TxHash)
				if err == nil {
					return nil
				}
			}
		}
		time.Sleep(10 * time.Second)
		_, err := client.GetPendingTransaction(ctx, tx.Vin[0].TXInLinker.TxHash)
		if err == nil {
			continue
		}
		_, err = client.GetMinedTransaction(ctx, tx.Vin[0].TXInLinker.TxHash)
		if err == nil {
			return nil
		}
		return nil
	}
}

func (f *funder) setupDataStoreMode(privk string, nodeList []string) error {
	f.nodeList = nodeList
	ctx := context.Background()
	subCtx, cf := context.WithCancel(ctx)
	f.ctx = subCtx
	f.cf = cf
	fmt.Printf("Funder setting up signing\n")
	signer, acct, err := f.setupHexSigner(privk)
	if err != nil {
		fmt.Printf("Funder error at setupHexSigner: %v\n", err)
		return err
	}
	f.signer = signer
	f.acct = acct
	fmt.Printf("Funder setting up client\n")
	client, err := f.setupClient(ctx, 0)
	if err != nil {
		fmt.Printf("Funder error at setupClient: %v\n", err)
		return err
	}
	f.client = client
	return nil
}

func (f *funder) setupDataStoreTransaction(ctx context.Context, signer aobjs.Signer, ownerAcct []byte, msg string, ind string) (*aobjs.Tx, error) {
	index := crypto.Hasher([]byte(ind))
	deposit, err := aobjs.BaseDepositEquation(uint32(len(msg)), numEpochs)
	if err != nil {
		return nil, err
	}
	consumedUtxos, consumedValue, err := f.blockingGetFunding(ctx, f.client, f.getCurveSpec(f.signer), f.acct, deposit)
	if err != nil {
		panic(err)
	}
	if consumedValue.Lt(deposit) {
		fmt.Printf("ACCOUNT DOES NOT HAVE ENOUGH FUNDING: REQUIRES:%v    HAS:%v\n", deposit, consumedValue)
		os.Exit(1)
	}
	tx := &aobjs.Tx{
		Vin:  aobjs.Vin{},
		Vout: aobjs.Vout{},
	}
	chainID := uint32(42)
	for _, utxo := range consumedUtxos {
		consumedVS, err := utxo.ValueStore()
		if err != nil {
			return nil, err
		}
		fmt.Println(consumedVS.Value())
		chainID, err = consumedVS.ChainID()
		if err != nil {
			return nil, err
		}
		txIn, err := utxo.MakeTxIn()
		if err != nil {
			return nil, err
		}
		tx.Vin = append(tx.Vin, txIn)
	}
	fmt.Printf("THE LEN OF VIN %v \n", len(tx.Vin))
	valueOut := uint256.Zero()
	{
		en, err := f.client.GetEpochNumber(ctx)
		if err != nil {
			return nil, err
		}
		fmt.Printf("REQUIRED DEPOSIT %v \n", deposit)
		//valueOut += deposit
		_, err = valueOut.Add(valueOut, deposit)
		if err != nil {
			return nil, err
		}
		newOwner := &aobjs.DataStoreOwner{}
		newOwner.New(ownerAcct, f.getCurveSpec(signer))
		newDataStore := &aobjs.DataStore{
			DSLinker: &aobjs.DSLinker{
				DSPreImage: &aobjs.DSPreImage{
					ChainID:  chainID,
					Index:    index,
					IssuedAt: en,
					Deposit:  deposit,
					RawData:  []byte(msg),
					TXOutIdx: 0,
					Owner:    newOwner,
					Fee:      new(uint256.Uint256).SetZero(),
				},
				TxHash: make([]byte, constants.HashLen),
			},
		}
		eoe, err := newDataStore.EpochOfExpiration()
		if err != nil {
			return nil, err
		}
		fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
		fmt.Printf("DS:  index:%x    deposit:%v    EpochOfExpire:%v    msg:%s\n", index, deposit, eoe, msg)
		newUTXO := &aobjs.TXOut{}
		err = newUTXO.NewDataStore(newDataStore)
		if err != nil {
			panic(err)
		}
		tx.Vout = append(tx.Vout, newUTXO)
	}
	fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
	if consumedValue.Gt(valueOut) {
		diff, err := new(uint256.Uint256).Sub(consumedValue.Clone(), valueOut.Clone())
		if err != nil {
			panic(err)
		}
		newOwner := &aobjs.ValueStoreOwner{}
		newOwner.New(ownerAcct, f.getCurveSpec(signer))
		newValueStore := &aobjs.ValueStore{
			VSPreImage: &aobjs.VSPreImage{
				ChainID:  chainID,
				Value:    diff,
				Owner:    newOwner,
				TXOutIdx: 0,
				Fee:      new(uint256.Uint256).SetZero(),
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
	fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
	err = tx.Vout.SetTxOutIdx()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
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
	fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
	txb, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	fmt.Printf("TX SIZE: %v\n", len(txb))
	return tx, nil
}

func (f *funder) setupDataStoreTransaction2(ctx context.Context, signer aobjs.Signer, ownerAcct []byte, msg string, ind string) (*aobjs.Tx, error) {
	index := crypto.Hasher([]byte(ind))
	deposit, err := aobjs.BaseDepositEquation(uint32(len(msg)), numEpochs)
	if err != nil {
		return nil, err
	}

	curveSpec := f.getCurveSpec(signer)

	var ds *aobjs.TXOut
	for {
		resp, err := f.client.PaginateDataStoreUTXOByOwner(ctx, curveSpec, ownerAcct, 1, utils.CopySlice(index))
		if err != nil {
			fmt.Println(err)
			time.Sleep(1 * time.Second)
			continue
		}
		utxoIDs := [][]byte{}
		for i := 0; i < len(resp); i++ {
			utxoIDs = append(utxoIDs, resp[i].UTXOID)
		}
		if len(utxoIDs) != 1 {
			time.Sleep(1 * time.Second)
			continue
		}

		resp2, err := f.client.GetUTXO(ctx, utxoIDs)
		if err != nil {
			continue
		}

		if len(resp2) != 1 {
			time.Sleep(1 * time.Second)
			continue
		} else {
			ds = resp2[0]
			break
		}
	}
	bn, err := f.client.GetBlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	v, err := ds.RemainingValue(bn) // + constants.EpochLength)
	if err != nil {
		return nil, err
	}
	depositClone := deposit.Clone()
	valueNeeded, err := depositClone.Sub(depositClone, v)
	if err != nil {
		panic(err)
	}

	consumedUtxos, consumedValue, err := f.blockingGetFunding(ctx, f.client, f.getCurveSpec(f.signer), f.acct, valueNeeded)
	if err != nil {
		panic(err)
	}
	if consumedValue.Lt(valueNeeded) {
		fmt.Printf("ACCOUNT DOES NOT HAVE ENOUGH FUNDING: REQUIRES:%v    HAS:%v\n", valueNeeded, consumedValue)
	}

	tx := &aobjs.Tx{
		Vin:  aobjs.Vin{},
		Vout: aobjs.Vout{},
	}
	chainID := uint32(42)
	for _, utxo := range consumedUtxos {
		consumedVS, err := utxo.ValueStore()
		if err != nil {
			return nil, err
		}
		fmt.Println(consumedVS.Value())
		chainID, err = consumedVS.ChainID()
		if err != nil {
			return nil, err
		}
		txIn, err := utxo.MakeTxIn()
		if err != nil {
			return nil, err
		}
		tx.Vin = append(tx.Vin, txIn)
	}

	dsTxIn, err := ds.MakeTxIn()
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	tx.Vin = append(tx.Vin, dsTxIn)
	consumedUtxos = append(consumedUtxos, ds)
	fmt.Printf("THE LEN OF VIN %v \n", len(tx.Vin))

	_, err = consumedValue.Add(consumedValue, v)
	if err != nil {
		panic(err)
	}
	valueOut := uint256.Zero()
	{
		en, err := f.client.GetEpochNumber(ctx)
		if err != nil {
			return nil, err
		}
		fmt.Printf("REQUIRED DEPOSIT %v \n", deposit)
		_, err = valueOut.Add(valueOut, deposit)
		if err != nil {
			panic(err)
		}
		newOwner := &aobjs.DataStoreOwner{}
		newOwner.New(ownerAcct, f.getCurveSpec(signer))
		newDataStore := &aobjs.DataStore{
			DSLinker: &aobjs.DSLinker{
				DSPreImage: &aobjs.DSPreImage{
					ChainID:  chainID,
					Index:    index,
					IssuedAt: en, //+ 1,
					Deposit:  deposit,
					RawData:  []byte(msg),
					TXOutIdx: 0,
					Owner:    newOwner,
					Fee:      new(uint256.Uint256).SetZero(),
				},
				TxHash: make([]byte, constants.HashLen),
			},
		}
		eoe, err := newDataStore.EpochOfExpiration()
		if err != nil {
			return nil, err
		}
		fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
		fmt.Printf("DS:  index:%x    deposit:%v    EpochOfExpire:%v    msg:%s\n", index, deposit, eoe, msg)
		newUTXO := &aobjs.TXOut{}
		err = newUTXO.NewDataStore(newDataStore)
		if err != nil {
			return nil, err
		}
		tx.Vout = append(tx.Vout, newUTXO)
	}
	fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
	if consumedValue.Gt(valueOut) {
		diff, err := new(uint256.Uint256).Sub(consumedValue.Clone(), valueOut.Clone())
		if err != nil {
			panic(err)
		}
		newOwner := &aobjs.ValueStoreOwner{}
		newOwner.New(ownerAcct, f.getCurveSpec(signer))
		newValueStore := &aobjs.ValueStore{
			VSPreImage: &aobjs.VSPreImage{
				ChainID:  chainID,
				Value:    diff,
				Owner:    newOwner,
				TXOutIdx: 0,
				Fee:      new(uint256.Uint256).SetZero(),
			},
			TxHash: make([]byte, constants.HashLen),
		}
		newUTXO := &aobjs.TXOut{}
		err = newUTXO.NewValueStore(newValueStore)
		if err != nil {
			return nil, err
		}
		tx.Vout = append(tx.Vout, newUTXO)
		_, err = valueOut.Add(valueOut, diff)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
	err = tx.Vout.SetTxOutIdx()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
	err = tx.SetTxHash()
	if err != nil {
		return nil, err
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
				return nil, err
			}
		case consumedUtxo.HasDataStore():
			consumedDS, err := consumedUtxo.DataStore()
			if err != nil {
				return nil, err
			}
			txIn := tx.Vin[idx]
			err = consumedDS.Sign(txIn, signer)
			if err != nil {
				return nil, err
			}
		}
	}
	fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
	_, err = f.client.GetBlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
	txb, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	fmt.Printf("TX SIZE: %v\n", len(txb))
	return tx, nil
}

func (f *funder) setupDataStoreTransaction3(ctx context.Context, signer aobjs.Signer, ownerAcct []byte, msg string, ind string) (*aobjs.Tx, error) {
	index := crypto.Hasher([]byte(ind))
	deposit, err := aobjs.BaseDepositEquation(uint32(len(msg)), numEpochs)
	if err != nil {
		return nil, err
	}

	curveSpec := f.getCurveSpec(signer)

	var ds *aobjs.TXOut
	for {
		resp, err := f.client.PaginateDataStoreUTXOByOwner(ctx, curveSpec, ownerAcct, 1, utils.CopySlice(index))
		if err != nil {
			fmt.Println(err)
			time.Sleep(1 * time.Second)
			continue
		}
		utxoIDs := [][]byte{}
		for i := 0; i < len(resp); i++ {
			utxoIDs = append(utxoIDs, resp[i].UTXOID)
		}
		if len(utxoIDs) != 1 {
			time.Sleep(1 * time.Second)
			continue
		}

		resp2, err := f.client.GetUTXO(ctx, utxoIDs)
		if err != nil {
			continue
		}

		if len(resp2) != 1 {
			time.Sleep(1 * time.Second)
			continue
		} else {
			ds = resp2[0]
			break
		}
	}
	bn, err := f.client.GetBlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	v, err := ds.RemainingValue(bn + constants.EpochLength)
	if err != nil {
		return nil, err
	}
	depositClone := deposit.Clone()
	valueNeeded, err := depositClone.Sub(depositClone, v)
	if err != nil {
		panic(err)
	}

	consumedUtxos, consumedValue, err := f.blockingGetFunding(ctx, f.client, f.getCurveSpec(f.signer), f.acct, valueNeeded)
	if err != nil {
		panic(err)
	}
	if consumedValue.Lt(valueNeeded) {
		fmt.Printf("ACCOUNT DOES NOT HAVE ENOUGH FUNDING: REQUIRES:%v    HAS:%v\n", valueNeeded, consumedValue)
	}

	tx := &aobjs.Tx{
		Vin:  aobjs.Vin{},
		Vout: aobjs.Vout{},
	}
	chainID := uint32(42)
	for _, utxo := range consumedUtxos {
		consumedVS, err := utxo.ValueStore()
		if err != nil {
			return nil, err
		}
		fmt.Println(consumedVS.Value())
		chainID, err = consumedVS.ChainID()
		if err != nil {
			return nil, err
		}
		txIn, err := utxo.MakeTxIn()
		if err != nil {
			return nil, err
		}
		tx.Vin = append(tx.Vin, txIn)
	}

	dsTxIn, err := ds.MakeTxIn()
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	tx.Vin = append(tx.Vin, dsTxIn)
	consumedUtxos = append(consumedUtxos, ds)
	fmt.Printf("THE LEN OF VIN %v \n", len(tx.Vin))

	_, err = consumedValue.Add(consumedValue, v)
	if err != nil {
		return nil, err
	}
	valueOut := uint256.Zero()
	{
		en, err := f.client.GetEpochNumber(ctx)
		if err != nil {
			return nil, err
		}
		fmt.Printf("REQUIRED DEPOSIT %v \n", deposit)
		_, err = valueOut.Add(valueOut, deposit)
		if err != nil {
			return nil, err
		}
		newOwner := &aobjs.DataStoreOwner{}
		newOwner.New(ownerAcct, f.getCurveSpec(signer))
		newDataStore := &aobjs.DataStore{
			DSLinker: &aobjs.DSLinker{
				DSPreImage: &aobjs.DSPreImage{
					ChainID:  chainID,
					Index:    index,
					IssuedAt: en + 1,
					Deposit:  deposit,
					RawData:  []byte(msg),
					TXOutIdx: 0,
					Owner:    newOwner,
					Fee:      new(uint256.Uint256).SetZero(),
				},
				TxHash: make([]byte, constants.HashLen),
			},
		}
		eoe, err := newDataStore.EpochOfExpiration()
		if err != nil {
			return nil, err
		}
		fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
		fmt.Printf("DS:  index:%x    deposit:%v    EpochOfExpire:%v    msg:%s\n", index, deposit, eoe, msg)
		newUTXO := &aobjs.TXOut{}
		err = newUTXO.NewDataStore(newDataStore)
		if err != nil {
			return nil, err
		}
		tx.Vout = append(tx.Vout, newUTXO)
	}
	fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
	if consumedValue.Gt(valueOut) {
		diff, err := new(uint256.Uint256).Sub(consumedValue.Clone(), valueOut.Clone())
		if err != nil {
			panic(err)
		}
		newOwner := &aobjs.ValueStoreOwner{}
		newOwner.New(ownerAcct, f.getCurveSpec(signer))
		newValueStore := &aobjs.ValueStore{
			VSPreImage: &aobjs.VSPreImage{
				ChainID:  chainID,
				Value:    diff,
				Owner:    newOwner,
				TXOutIdx: 0,
				Fee:      new(uint256.Uint256).SetZero(),
			},
			TxHash: make([]byte, constants.HashLen),
		}
		newUTXO := &aobjs.TXOut{}
		err = newUTXO.NewValueStore(newValueStore)
		if err != nil {
			return nil, err
		}
		tx.Vout = append(tx.Vout, newUTXO)
		_, err = valueOut.Add(valueOut, diff)
		if err != nil {
			return nil, err
		}
	}
	fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
	err = tx.Vout.SetTxOutIdx()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
	err = tx.SetTxHash()
	if err != nil {
		return nil, err
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
				return nil, err
			}
		case consumedUtxo.HasDataStore():
			consumedDS, err := consumedUtxo.DataStore()
			if err != nil {
				return nil, err
			}
			txIn := tx.Vin[idx]
			err = consumedDS.Sign(txIn, signer)
			if err != nil {
				return nil, err
			}
		}
	}
	fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
	_, err = f.client.GetBlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Consumed Next:%v    ValueOut:%v\n", consumedValue, valueOut)
	txb, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	fmt.Printf("TX SIZE: %v\n", len(txb))
	return tx, nil
}

func main() {
	dPtr := flag.Bool("d", false, "DataStore mode.")
	sPtr := flag.Bool("s", false, "Spam mode.")
	nPtr := flag.Int("n", 100, "Number workers.")
	bPtr := flag.Int("b", 100, "Base privk offset for workers. This should not overlap such that another test group is in same range.")
	mPtr := flag.String("m", "", "Data to write to state store.")
	iPtr := flag.String("i", "", "Index of data to write to state store.")
	flag.Parse()
	datastoreMode := *dPtr
	spamMode := *sPtr
	privk := "6aea45ee1273170fb525da34015e4f20ba39fe792f486ba74020bcacc9badfc1"
	nodeList := []string{"127.0.0.1:8887", "127.0.0.1:8888"}
	if spamMode {
		base := *bPtr
		num := 63
		for {
			f := &funder{}
			if err := f.doSpam(privk, num, nodeList, base); err != nil {
				panic(err)
			}
			base = base + num + 1
		}
	}
	f := &funder{}
	ctx := context.Background()
	if datastoreMode {
		fmt.Println("Running in DataStore Mode")
		if err := f.setupDataStoreMode(privk, nodeList); err != nil {
			panic(err)
		}
		tx, err := f.setupDataStoreTransaction(ctx, f.signer, f.acct, *mPtr, *iPtr)
		if err != nil {
			panic(err)
		}
		err = f.blockingSendTx(ctx, f.client, tx)
		if err != nil {
			panic(err)
		}
		time.Sleep(5 * time.Second)
		tx, err = f.setupDataStoreTransaction2(ctx, f.signer, f.acct, strings.Join([]string{*mPtr, "two"}, "-"), *iPtr)
		if err != nil {
			panic(err)
		}
		err = f.blockingSendTx(ctx, f.client, tx)
		if err != nil {
			panic(err)
		}
		time.Sleep(5 * time.Second)
		tx, err = f.setupDataStoreTransaction3(ctx, f.signer, f.acct, strings.Join([]string{*mPtr, "three"}, "-"), *iPtr)
		if err != nil {
			panic(err)
		}
		err = f.blockingSendTx(ctx, f.client, tx)
		if err != nil {
			panic(err)
		}
	} else {
		numChildren := *nPtr
		baseIdx := *bPtr
		if err := f.init(privk, numChildren, nodeList, baseIdx); err != nil {
			panic(err)
		}
	}
}
