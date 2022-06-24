//go:build flakes

package dman

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/logging"
	"github.com/dgraph-io/badger/v2"
	"google.golang.org/grpc"
)

type testingProxyCall int

func (tpc testingProxyCall) String() string {
	switch tpc {
	case pendingTxCall:
		return "pendingTxCall"
	case minedTxCall:
		return "minedTxCall"
	case blockHeaderCall:
		return "blockHeaderCall"
	case unmarshalTxCall:
		return "unmarshalTxCall"
	default:
		panic("")
	}
}

const (
	pendingTxCall testingProxyCall = iota + 1
	minedTxCall
	blockHeaderCall
	unmarshalTxCall
)

type testingTransaction struct {
	hash []byte
}

func (t *testingTransaction) setHashForTest(h []byte) *testingTransaction {
	t.hash = h
	return t
}

func (t *testingTransaction) TxHash() ([]byte, error) {
	if t.hash == nil {
		return make([]byte, constants.HashLen), nil
	}
	return t.hash, nil
}

func (t *testingTransaction) MarshalBinary() ([]byte, error) {
	if t.hash == nil {
		return make([]byte, constants.HashLen), nil
	}
	return t.hash, nil
}

func (t *testingTransaction) XXXIsTx() {}

type testingProxy struct {
	sync.Mutex
	callIndex     int
	expectedCalls []testingProxyCall
	returns       [][]interface{}
	skipCallCheck bool
}

func (trb *testingProxy) checkExpect() error {
	if trb.callIndex != len(trb.expectedCalls) {
		return fmt.Errorf("Missing calls: %v", trb.expectedCalls[trb.callIndex:])
	}
	return nil
}

func (trb *testingProxy) expect(trbc []testingProxyCall, rtypes [][]interface{}) {
	trb.expectedCalls = trbc
	trb.returns = rtypes
}

func (dv *testingProxy) SetTxCacheItem(txn *badger.Txn, height uint32, txHash []byte, tx []byte) error {
	panic("")
}
func (dv *testingProxy) GetTxCacheItem(txn *badger.Txn, height uint32, txHash []byte) ([]byte, error) {
	panic("")
}
func (dv *testingProxy) SetCommittedBlockHeader(txn *badger.Txn, v *objs.BlockHeader) error {
	panic("")
}
func (dv *testingProxy) TxCacheDropBefore(txn *badger.Txn, beforeHeight uint32, maxKeys int) error {
	panic("")
}

func (trb *testingProxy) RequestP2PGetPendingTx(ctx context.Context, txHashes [][]byte, opts ...grpc.CallOption) ([][]byte, error) {
	defer func() {
		trb.callIndex++
	}()
	cType := pendingTxCall
	trb.Lock()
	defer trb.Unlock()
	if trb.callIndex == len(trb.expectedCalls) {
		panic(fmt.Sprintf("got unexpected call of type %s : expected calls %v", cType, trb.expectedCalls))
	}
	if trb.expectedCalls[trb.callIndex] != cType {
		panic(fmt.Sprintf("got unexpected call of type %s at index %v : expected calls %v", cType, trb.callIndex, trb.expectedCalls))
	}
	if ctx == nil {
		panic(fmt.Sprintf("ctx was nil in test mock object of call type %s", cType))
	}
	returnTuple := trb.returns[trb.callIndex]
	tx := returnTuple[0].([][]byte)
	err, ok := returnTuple[1].(error)
	if ok {
		return tx, err
	}
	return tx, nil
}

func (trb *testingProxy) RequestP2PGetMinedTxs(ctx context.Context, txHashes [][]byte, opts ...grpc.CallOption) ([][]byte, error) {
	defer func() {
		trb.callIndex++
	}()
	cType := minedTxCall
	trb.Lock()
	defer trb.Unlock()
	if trb.callIndex == len(trb.expectedCalls) {
		panic(fmt.Sprintf("got unexpected call of type %s : expected calls %v", cType, trb.expectedCalls))
	}
	if trb.expectedCalls[trb.callIndex] != cType {
		panic(fmt.Sprintf("got unexpected call of type %s at index %v : expected calls %v", cType, trb.callIndex, trb.expectedCalls))
	}
	if ctx == nil {
		panic(fmt.Sprintf("ctx was nil in test mock object of call type %s", cType))
	}
	returnTuple := trb.returns[trb.callIndex]
	tx := returnTuple[0].([][]byte)
	err, ok := returnTuple[1].(error)
	if ok {
		return tx, err
	}
	return tx, nil
}

func (trb *testingProxy) RequestP2PGetBlockHeaders(ctx context.Context, blockNums []uint32, opts ...grpc.CallOption) ([]*objs.BlockHeader, error) {
	defer func() {
		trb.callIndex++
	}()
	cType := blockHeaderCall
	trb.Lock()
	defer trb.Unlock()
	if trb.callIndex == len(trb.expectedCalls) {
		panic(fmt.Sprintf("got unexpected call of type %s : expected calls %v", cType, trb.expectedCalls))
	}
	if trb.expectedCalls[trb.callIndex] != cType {
		panic(fmt.Sprintf("got unexpcted call of type %s at index %v : expected calls %v", cType, trb.callIndex, trb.expectedCalls))
	}
	if ctx == nil {
		panic(fmt.Sprintf("ctx was nil in test mock object of call type %s", cType))
	}
	returnTuple := trb.returns[trb.callIndex]
	bh := returnTuple[0].([]*objs.BlockHeader)
	err, ok := returnTuple[1].(error)
	if ok {
		return bh, err
	}
	return bh, nil
}

func (trb *testingProxy) UnmarshalTx(tx []byte) (interfaces.Transaction, error) {
	if trb.skipCallCheck {
		return &testingTransaction{}, nil
	}
	defer func() {
		trb.callIndex++
	}()
	cType := unmarshalTxCall
	trb.Lock()
	defer trb.Unlock()
	if trb.callIndex == len(trb.expectedCalls) {
		panic(fmt.Sprintf("got unexpected call of type %s : expected calls %v", cType, trb.expectedCalls))
	}
	if trb.expectedCalls[trb.callIndex] != cType {
		panic(fmt.Sprintf("got unexpected call of type %s at index %v : expected calls %v", cType, trb.callIndex, trb.expectedCalls))
	}
	if tx == nil {
		panic(fmt.Sprintf("tx was nil in test mock object of call type %s", cType))
	}
	returnTuple := trb.returns[trb.callIndex]
	txi := returnTuple[0].(*testingTransaction)
	err, ok := returnTuple[1].(error)
	if ok {
		return txi, err
	}
	return txi, nil
}

func TestRootActor_download(t *testing.T) {
	trb := &testingProxy{}

	type args struct {
		b     DownloadRequest
		check bool
	}
	tests := []struct {
		name         string
		args         args
		proxyCalls   []testingProxyCall
		proxyReturns [][]interface{}
		after        func(ra *RootActor) error
	}{
		{
			"BadBlock",
			args{b: NewBlockHeaderDownloadRequest(10000, 1, BlockHeaderRequest)},
			[]testingProxyCall{blockHeaderCall},
			append([][]interface{}{}, []interface{}{[]*objs.BlockHeader{new(objs.BlockHeader)}, nil}),
			func(ra *RootActor) error {
				if ra.bhc.Contains(1) {
					return errors.New("had one in it")
				}
				return nil
			},
		},
		{
			"GoodBlock",
			args{b: NewBlockHeaderDownloadRequest(100, 1, BlockHeaderRequest), check: true},
			[]testingProxyCall{blockHeaderCall},
			append([][]interface{}{}, []interface{}{makeGoodBlock(t), nil}),
			func(ra *RootActor) error {
				if !ra.bhc.Contains(100) {
					return errors.New("missing one in it")
				}
				return nil
			},
		},
		{
			"GoodPendingTx",
			args{b: NewTxDownloadRequest(make([]byte, constants.HashLen), PendingTxRequest, 1, 1)},
			[]testingProxyCall{pendingTxCall, unmarshalTxCall, unmarshalTxCall},
			append([][]interface{}{}, []interface{}{[][]byte{make([]byte, constants.HashLen)}, nil}, []interface{}{&testingTransaction{}, nil}, []interface{}{&testingTransaction{}, nil}),
			func(ra *RootActor) error {
				if !ra.txc.Contains(make([]byte, constants.HashLen)) {
					return errors.New("missing one in it")
				}
				return nil
			},
		},
		{
			"GoodMinedTx",
			args{b: NewTxDownloadRequest(make([]byte, constants.HashLen), MinedTxRequest, 1, 1)},
			[]testingProxyCall{minedTxCall, unmarshalTxCall, unmarshalTxCall},
			append([][]interface{}{}, []interface{}{[][]byte{make([]byte, constants.HashLen)}, nil}, []interface{}{&testingTransaction{}, nil}, []interface{}{&testingTransaction{}, nil}),
			func(ra *RootActor) error {
				if !ra.txc.Contains(make([]byte, constants.HashLen)) {
					return errors.New("missing one in it")
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		ra := &RootActor{}
		t.Run(tt.name, func(t *testing.T) {
			func() {
				tlog := logging.GetLogger("Test")
				err := ra.Init(tlog, trb)
				if err != nil {
					t.Fatal(err)
				}
				ra.Start()
				defer ra.Close()
				trb.expect(tt.proxyCalls, tt.proxyReturns)
				ra.download(tt.args.b, false)
			}()
		})
		if tt.args.check {
			if err := trb.checkExpect(); err != nil {
				t.Fatal(err)
			}
		}
		if err := tt.after(ra); err != nil {
			t.Fatal(err)
		}
		trb.callIndex = 0
	}
}

func makeGoodBlock(t *testing.T) []*objs.BlockHeader {
	bclaimsList, txHashListList, err := generateChain()
	if err != nil {
		t.Fatal(err)
	}
	bclaims := bclaimsList[0]
	bhsh, err := bclaims.BlockHash()
	if err != nil {
		t.Fatal(err)
	}
	gk := crypto.BNGroupSigner{}
	err = gk.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	sig, err := gk.Sign(bhsh)
	if err != nil {
		t.Fatal(err)
	}
	bh := &objs.BlockHeader{
		BClaims:  bclaims,
		SigGroup: sig,
		TxHshLst: txHashListList[0],
	}
	return []*objs.BlockHeader{bh}
}

func generateChain() ([]*objs.BClaims, [][][]byte, error) {
	chain := []*objs.BClaims{}
	txHashes := [][][]byte{}
	txhash := crypto.Hasher([]byte(strconv.Itoa(1)))
	txHshLst := [][]byte{txhash}
	txRoot, err := objs.MakeTxRoot(txHshLst)
	if err != nil {
		return nil, nil, err
	}
	txHashes = append(txHashes, txHshLst)
	bclaims := &objs.BClaims{
		ChainID:    1,
		Height:     100,
		TxCount:    1,
		PrevBlock:  crypto.Hasher([]byte("foo")),
		TxRoot:     txRoot,
		StateRoot:  crypto.Hasher([]byte("")),
		HeaderRoot: crypto.Hasher([]byte("")),
	}
	chain = append(chain, bclaims)
	return chain, txHashes, nil
}
