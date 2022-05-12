package main

import (
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"sync"
	"time"

	aobjs "github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/localrpc"
)

var errClosing = errors.New("closing")

type worker struct {
	f      *funder
	client *localrpc.Client
	signer aobjs.Signer
	acct   []byte
	idx    int
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

	children, err := f.setupChildren(ctx, numChildren, baseIdx)
	if err != nil {
		panic(err)
	}
	f.children = children

	numChildren256 := &uint256.Uint256{}
	numChildren256, err = numChildren256.FromUint64(uint64(numChildren))
	if err != nil {
		panic(err)
	}

	//utxos, value, err := f.blockingGetFunding(f.client, f.getCurveSpec(f.signer), f.acct, uint64(len(children)))
	utxos, value, err := f.blockingGetFunding(ctx, f.client, f.getCurveSpec(f.signer), f.acct, numChildren256)
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

func (f *funder) setupClient(ctx context.Context, idx int) (*localrpc.Client, error) {
	idx2 := idx % len(f.nodeList)
	client := &localrpc.Client{Address: f.nodeList[idx2], TimeOut: constants.MsgTimeout}
	if err := client.Connect(ctx); err != nil {
		return nil, err
	}
	return client, nil
}

func (f *funder) setupTestingSigner(i int) (aobjs.Signer, []byte, error) {
	privk, err := hex.DecodeString("6aea45ee1273170fb525da34015e4f20ba39fe792f486ba74020bcacc9badfc1")
	if err != nil {
		panic(err)
	}
	return f.setupSecpSigner(privk)
}

//nolint:unused
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
	tx := &aobjs.Tx{
		Vin:  aobjs.Vin{},
		Vout: aobjs.Vout{},
	}
	chainID := uint32(0)
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
	valueOut := uint256.Zero()
	for _, r := range recipients {
		_, err := valueOut.Add(valueOut, uint256.One())
		if err != nil {
			return nil, err
		}
		newOwner := &aobjs.ValueStoreOwner{}
		newOwner.New(r.acct, f.getCurveSpec(r.signer))
		newValueStore := &aobjs.ValueStore{
			VSPreImage: &aobjs.VSPreImage{
				ChainID:  chainID,
				Value:    uint256.One(),
				Owner:    newOwner,
				TXOutIdx: 0,
			},
			TxHash: make([]byte, constants.HashLen),
		}
		newUTXO := &aobjs.TXOut{}
		err = newUTXO.NewValueStore(newValueStore)
		if err != nil {
			return nil, err
		}

		tx.Vout = append(tx.Vout, newUTXO)
	}
	if consumedValue.Gt(valueOut) {
		diff, err := new(uint256.Uint256).Sub(consumedValue.Clone(), valueOut.Clone())
		if err != nil {
			panic(err)
		}
		newOwner := &aobjs.ValueStoreOwner{}
		newOwner.New(ownerAcct, f.getCurveSpec(signer))
		newValueStore := &aobjs.ValueStore{
			VSPreImage: &aobjs.VSPreImage{
				ChainID: chainID,
				//Value:    consumedValue - valueOut,
				Value:    diff,
				Owner:    newOwner,
				TXOutIdx: 0,
			},
			TxHash: make([]byte, constants.HashLen),
		}
		newUTXO := &aobjs.TXOut{}
		err = newUTXO.NewValueStore(newValueStore)
		if err != nil {
			return nil, err
		}
		tx.Vout = append(tx.Vout, newUTXO)
	}
	err := tx.SetTxHash()
	if err != nil {
		return nil, err
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
		client, err := f.setupClient(ctx, baseIdx)
		if err != nil {
			return nil, err
		}
		signer, acct, err := f.setupTestingSigner(baseIdx)
		if err != nil {
			return nil, err
		}
		c := &worker{
			f:      f,
			signer: signer,
			acct:   acct,
			client: client,
			idx:    baseIdx,
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

			var collectedUtxos []*aobjs.TXOut
			for i := range utxos {
				utxo := utxos[i]

				vs, err := utxo.ValueStore()
				if err != nil {
					panic(err)
				}

				v, err := vs.Value()
				if err != nil {
					panic(err)
				}

				if v.Gte(uint256.One()) {
					collectedUtxos = append(collectedUtxos, utxo)
				}
			}

			return collectedUtxos, totalValue, nil
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

func main() {
	nPtr := flag.Int("n", 100, "Number workers.")
	bPtr := flag.Int("b", 100, "Base privk offset for workers. This should not overlap such that another test group is in same range.")
	flag.Parse()
	privk := "6aea45ee1273170fb525da34015e4f20ba39fe792f486ba74020bcacc9badfc1"
	nodeList := []string{"127.0.0.1:8887", "127.0.0.1:8888"}

	f := &funder{}

	numChildren := *nPtr
	baseIdx := *bPtr
	if err := f.init(privk, numChildren, nodeList, baseIdx); err != nil {
		panic(err)
	}
}
