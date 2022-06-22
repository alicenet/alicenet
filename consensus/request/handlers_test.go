package request

import (
	"context"
	"errors"
	"math/big"
	"strconv"
	"testing"

	appObjs "github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	bn256 "github.com/alicenet/alicenet/crypto/bn256/cloudflare"
	"github.com/alicenet/alicenet/dynamics"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/proto"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type HandlerMock struct {
	mock.Mock
}

func (h *HandlerMock) PendingTxGet(txn *badger.Txn, height uint32, txHash [][]byte) ([]interfaces.Transaction, [][]byte, error) {
	args := h.Called(txn, height, txHash)
	return args.Get(0).([]interfaces.Transaction), args.Get(1).([][]byte), args.Error(2)
}
func (h *HandlerMock) MinedTxGet(txn *badger.Txn, txsHashes [][]byte) ([]interfaces.Transaction, [][]byte, error) {
	args := h.Called(txn, txsHashes)
	return args.Get(0).([]interfaces.Transaction), args.Get(1).([][]byte), args.Error(2)
}
func (h *HandlerMock) GetSnapShotNode(txn *badger.Txn, height uint32, key []byte) ([]byte, error) {
	args := h.Called(txn, height, key)
	return args.Get(0).([]byte), args.Error(1)
}
func (h *HandlerMock) GetSnapShotStateData(txn *badger.Txn, key []byte) ([]byte, error) {
	args := h.Called(txn, key)
	return args.Get(0).([]byte), args.Error(1)
}

type mockRawDB struct {
	rawDB map[string]string
}

func (m *mockRawDB) GetValue(txn *badger.Txn, key []byte) ([]byte, error) {
	strValue, ok := m.rawDB[string(key)]
	if !ok {
		return nil, errors.New("key not present")
	}
	value := []byte(strValue)
	return value, nil
}

func (m *mockRawDB) SetValue(txn *badger.Txn, key []byte, value []byte) error {
	strValue := string(value)
	m.rawDB[string(key)] = strValue
	return nil
}

func (m *mockRawDB) DeleteValue(key []byte) error {
	strKey := string(key)
	_, ok := m.rawDB[strKey]
	if !ok {
		return errors.New("key not present")
	}
	delete(m.rawDB, strKey)
	return nil
}

func (m *mockRawDB) View(fn func(txn *badger.Txn) error) error {
	return fn(nil)
}

func (m *mockRawDB) Update(fn func(txn *badger.Txn) error) error {
	return fn(nil)
}

func initHandler(t *testing.T, done <-chan struct{}) *Handler {
	rawDb, err := utils.OpenBadger(done, "", true)
	assert.Nil(t, err)
	database := &db.Database{}
	database.Init(rawDb)

	appMock := &HandlerMock{}
	logger := logrus.New()

	mockRawDB := &mockRawDB{}
	mockRawDB.rawDB = make(map[string]string)
	storage := &dynamics.Storage{}
	err = storage.Init(mockRawDB, logger)
	if err != nil {
		panic(err)
	}
	storage.Start()

	handler := &Handler{}
	handler.Init(database, appMock, storage)

	return handler
}

func TestHandler_HandleP2PStatus_Ok(t *testing.T) {
	ctx := context.Background()
	hndlr := initHandler(t, ctx.Done())
	os := createOwnState(t)
	err := hndlr.database.Update(func(txn *badger.Txn) error {
		err := hndlr.database.SetOwnState(txn, os)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		return nil
	})
	assert.Nil(t, err)

	resp, err := hndlr.HandleP2PStatus(ctx, &proto.StatusRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint32(0x1), resp.MaxBlockHeightSeen)
	assert.Equal(t, uint32(0x1), resp.SyncToBlockHeight)
}

func TestHandler_HandleP2PStatus_Error(t *testing.T) {
	ctx := context.Background()
	hndlr := initHandler(t, ctx.Done())

	resp, err := hndlr.HandleP2PStatus(ctx, &proto.StatusRequest{})
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestHandler_HandleP2PGetBlockHeaders_Ok(t *testing.T) {
	ctx := context.Background()
	hndlr := initHandler(t, ctx.Done())
	bh := createBlockHeader(t, 1)
	err := hndlr.database.Update(func(txn *badger.Txn) error {
		err := hndlr.database.SetCommittedBlockHeader(txn, bh)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Shouldn't have raised error: %v", err)
		return
	}

	resp, err := hndlr.HandleP2PGetBlockHeaders(ctx, &proto.GetBlockHeadersRequest{BlockNumbers: []uint32{1}})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestHandler_HandleP2PGetBlockHeaders_Error(t *testing.T) {
	ctx := context.Background()
	hndlr := initHandler(t, ctx.Done())

	resp, err := hndlr.HandleP2PGetBlockHeaders(ctx, &proto.GetBlockHeadersRequest{BlockNumbers: []uint32{1}})
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestHandler_HandleP2PGetPendingTxs_Ok(t *testing.T) {
	ctx := context.Background()
	hndlr := initHandler(t, ctx.Done())

	appMock := &HandlerMock{}
	result := []interfaces.Transaction{createTx(t)}
	appMock.On("PendingTxGet", mock.Anything, mock.Anything, mock.Anything).Return(result, [][]byte{}, nil)
	hndlr.app = appMock

	resp, err := hndlr.HandleP2PGetPendingTxs(ctx, &proto.GetPendingTxsRequest{TxHashes: [][]byte{}})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestHandler_HandleP2PGetPendingTxs_Error(t *testing.T) {
	ctx := context.Background()
	hndlr := initHandler(t, ctx.Done())

	appMock := &HandlerMock{}
	appMock.On("PendingTxGet", mock.Anything, mock.Anything, mock.Anything).Return([]interfaces.Transaction{}, [][]byte{}, errors.New("key not found"))
	hndlr.app = appMock

	resp, err := hndlr.HandleP2PGetPendingTxs(ctx, &proto.GetPendingTxsRequest{TxHashes: [][]byte{}})
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestHandler_HandleP2PGetMinedTxs_Ok(t *testing.T) {
	ctx := context.Background()
	hndlr := initHandler(t, ctx.Done())

	appMock := &HandlerMock{}
	result := []interfaces.Transaction{createTx(t)}
	appMock.On("MinedTxGet", mock.Anything, mock.Anything, mock.Anything).Return(result, [][]byte{}, nil)
	hndlr.app = appMock

	resp, err := hndlr.HandleP2PGetMinedTxs(ctx, &proto.GetMinedTxsRequest{TxHashes: [][]byte{}})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestHandler_HandleP2PGetMinedTxs_Error(t *testing.T) {
	ctx := context.Background()
	hndlr := initHandler(t, ctx.Done())

	appMock := &HandlerMock{}
	appMock.On("MinedTxGet", mock.Anything, mock.Anything, mock.Anything).Return([]interfaces.Transaction{}, [][]byte{}, errors.New("key not found"))
	hndlr.app = appMock

	resp, err := hndlr.HandleP2PGetMinedTxs(ctx, &proto.GetMinedTxsRequest{TxHashes: [][]byte{}})
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestHandler_HandleP2PGetSnapShotNodes_Ok(t *testing.T) {
	ctx := context.Background()
	hndlr := initHandler(t, ctx.Done())

	appMock := &HandlerMock{}
	appMock.On("GetSnapShotNode", mock.Anything, mock.Anything, mock.Anything).Return(make([]byte, constants.HashLen), nil)
	hndlr.app = appMock

	resp, err := hndlr.HandleP2PGetSnapShotNode(ctx, &proto.GetSnapShotNodeRequest{Height: 1, NodeHash: make([]byte, constants.HashLen)})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestHandler_HandleP2PGetSnapShotNodes_Error(t *testing.T) {
	ctx := context.Background()
	hndlr := initHandler(t, ctx.Done())

	appMock := &HandlerMock{}
	appMock.On("GetSnapShotNode", mock.Anything, mock.Anything, mock.Anything).Return(make([]byte, constants.HashLen), errors.New("key not found"))
	hndlr.app = appMock

	resp, err := hndlr.HandleP2PGetSnapShotNode(ctx, &proto.GetSnapShotNodeRequest{Height: 1, NodeHash: make([]byte, constants.HashLen)})
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestHandler_HandleP2PGetSnapShotHdrNode_Ok(t *testing.T) {
	ctx := context.Background()
	hndlr := initHandler(t, ctx.Done())

	resp, err := hndlr.HandleP2PGetSnapShotHdrNode(ctx, &proto.GetSnapShotHdrNodeRequest{NodeHash: make([]byte, 0)})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestHandler_HandleP2PGetSnapShotStateData_Ok(t *testing.T) {
	ctx := context.Background()
	hndlr := initHandler(t, ctx.Done())

	appMock := &HandlerMock{}
	appMock.On("GetSnapShotStateData", mock.Anything, mock.Anything).Return(make([]byte, constants.HashLen), nil)
	hndlr.app = appMock

	resp, err := hndlr.HandleP2PGetSnapShotStateData(ctx, &proto.GetSnapShotStateDataRequest{Key: make([]byte, constants.HashLen)})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestHandler_HandleP2PGetSnapShotStateData_Error(t *testing.T) {
	ctx := context.Background()
	hndlr := initHandler(t, ctx.Done())

	appMock := &HandlerMock{}
	appMock.On("GetSnapShotStateData", mock.Anything, mock.Anything).Return(make([]byte, constants.HashLen), errors.New("key not found"))
	hndlr.app = appMock

	resp, err := hndlr.HandleP2PGetSnapShotStateData(ctx, &proto.GetSnapShotStateDataRequest{Key: make([]byte, constants.HashLen)})
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func createOwnState(t *testing.T) *objs.OwnState {
	secret1 := big.NewInt(100)
	secret2 := big.NewInt(101)
	secret3 := big.NewInt(102)
	secret4 := big.NewInt(103)

	big1 := big.NewInt(1)
	big2 := big.NewInt(2)

	privCoefs1 := []*big.Int{secret1, big1, big2}
	privCoefs2 := []*big.Int{secret2, big1, big2}
	privCoefs3 := []*big.Int{secret3, big1, big2}
	privCoefs4 := []*big.Int{secret4, big1, big2}

	share1to1 := bn256.PrivatePolyEval(privCoefs1, 1)
	share2to1 := bn256.PrivatePolyEval(privCoefs2, 1)
	share3to1 := bn256.PrivatePolyEval(privCoefs3, 1)
	share4to1 := bn256.PrivatePolyEval(privCoefs4, 1)

	listOfSS1 := []*big.Int{share1to1, share2to1, share3to1, share4to1}
	gsk1 := bn256.GenerateGroupSecretKeyPortion(listOfSS1)
	gpk1 := new(bn256.G2).ScalarBaseMult(gsk1)

	secpSigner := &crypto.Secp256k1Signer{}
	err := secpSigner.SetPrivk(crypto.Hasher(gpk1.Marshal()))
	if err != nil {
		panic(err)
	}
	secpKey, err := secpSigner.Pubkey()
	if err != nil {
		panic(err)
	}

	//BlockHeader
	bh := createBlockHeader(t, 1)

	return &objs.OwnState{
		VAddr:             secpKey,
		SyncToBH:          bh,
		MaxBHSeen:         bh,
		CanonicalSnapShot: bh,
		PendingSnapShot:   bh,
	}
}

func createTx(t *testing.T) *appObjs.Tx {
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	consumedUTXOs := appObjs.Vout{}
	consumedUTXO, vs := makeVS(t, ownerSigner, 1)
	consumedUTXOs = append(consumedUTXOs, consumedUTXO)
	err := consumedUTXOs.SetTxOutIdx()
	if err != nil {
		t.Fatal(err)
	}

	txsIn, err := consumedUTXOs.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	generatedUTXOs := appObjs.Vout{}
	generatedUTXO, _ := makeVS(t, ownerSigner, 0)
	generatedUTXOs = append(generatedUTXOs, generatedUTXO)

	err = generatedUTXOs.SetTxOutIdx()
	if err != nil {
		t.Fatal(err)
	}

	tx := &appObjs.Tx{
		Vin:  txsIn,
		Vout: generatedUTXOs,
		Fee:  uint256.Zero(),
	}

	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}

	err = vs.Sign(tx.Vin[0], ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	return tx
}

func makeVS(t *testing.T, ownerSigner appObjs.Signer, i int) (*appObjs.TXOut, *appObjs.ValueStore) {
	cid := uint32(2)
	val := uint256.One()

	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &appObjs.ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	fee := new(uint256.Uint256)
	vsp := &appObjs.VSPreImage{
		ChainID: cid,
		Value:   val,
		Owner:   owner,
		Fee:     fee.Clone(),
	}
	var txHash []byte
	if i == 0 {
		txHash = make([]byte, constants.HashLen)
	} else {
		txHash = crypto.Hasher([]byte(strconv.Itoa(i)))
	}
	vs := &appObjs.ValueStore{
		VSPreImage: vsp,
		TxHash:     txHash,
	}
	vs2 := &appObjs.ValueStore{}
	vsBytes, err := vs.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = vs2.UnmarshalBinary(vsBytes)
	if err != nil {
		t.Fatal(err)
	}
	utxInputs := &appObjs.TXOut{}
	err = utxInputs.NewValueStore(vs)
	if err != nil {
		t.Fatal(err)
	}
	return utxInputs, vs
}
