package request

import (
	"context"
	"errors"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/dynamics"
	"github.com/alicenet/alicenet/errorz"
	pb "github.com/alicenet/alicenet/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"strconv"
	"testing"
)

func TestClient_RequestP2PGetSnapShotNode_Ok(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetSnapShotNodeResponse{
		Node: make([]byte, constants.HashLen),
	}
	p2pClientMock.On("GetSnapShotNode", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	node, err := client.RequestP2PGetSnapShotNode(context.TODO(), 1, make([]byte, constants.HashLen), opts...)
	assert.Nil(t, err)
	assert.NotNil(t, node)
}

func TestClient_RequestP2PGetSnapShotNode_Error1(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetSnapShotNodeResponse{
		Node: make([]byte, constants.HashLen),
	}
	p2pClientMock.On("GetSnapShotNode", mock.Anything, mock.Anything, mock.Anything).Return(resp, errors.New("internal error"))

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	node, err := client.RequestP2PGetSnapShotNode(context.TODO(), 1, make([]byte, constants.HashLen), opts...)
	assert.NotNil(t, err)
	assert.Nil(t, node)
}

func TestClient_RequestP2PGetSnapShotNode_Error2(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetSnapShotNodeResponse{
		Node: nil,
	}
	p2pClientMock.On("GetSnapShotNode", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	node, err := client.RequestP2PGetSnapShotNode(context.TODO(), 1, make([]byte, constants.HashLen), opts...)
	assert.True(t, errors.Is(err, errorz.ErrBadResponse))
	assert.Nil(t, node)
}

func TestClient_RequestP2PGetSnapShotHdrNode_Ok(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetSnapShotHdrNodeResponse{
		Node: make([]byte, constants.HashLen),
	}
	p2pClientMock.On("GetSnapShotHdrNode", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	node, err := client.RequestP2PGetSnapShotHdrNode(context.TODO(), make([]byte, constants.HashLen), opts...)
	assert.Nil(t, err)
	assert.NotNil(t, node)
}

func TestClient_RequestP2PGetSnapShotHdrNode_Error1(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetSnapShotHdrNodeResponse{
		Node: make([]byte, constants.HashLen),
	}
	p2pClientMock.On("GetSnapShotHdrNode", mock.Anything, mock.Anything, mock.Anything).Return(resp, errors.New("internal error"))

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	node, err := client.RequestP2PGetSnapShotHdrNode(context.TODO(), make([]byte, constants.HashLen), opts...)
	assert.NotNil(t, err)
	assert.Nil(t, node)
}

func TestClient_RequestP2PGetSnapShotHdrNode_Error2(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetSnapShotHdrNodeResponse{
		Node: nil,
	}
	p2pClientMock.On("GetSnapShotHdrNode", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	node, err := client.RequestP2PGetSnapShotHdrNode(context.TODO(), make([]byte, constants.HashLen), opts...)
	assert.True(t, errors.Is(err, errorz.ErrBadResponse))
	assert.Nil(t, node)
}

func TestClient_RequestP2PGetBlockHeaders_Ok(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	bh1 := createBlockHeader(t, 1)
	bts1, err := bh1.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	bh2 := createBlockHeader(t, 2)
	bts2, err := bh2.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	resp := &pb.GetBlockHeadersResponse{
		BlockHeaders: [][]byte{bts1, bts2},
	}
	p2pClientMock.On("GetBlockHeaders", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}
	blocksLength := 2

	hdrs, err := client.RequestP2PGetBlockHeaders(context.TODO(), make([]uint32, blocksLength), opts...)
	assert.Nil(t, err)
	assert.NotNil(t, hdrs)
	assert.Len(t, hdrs, blocksLength)
}

func TestClient_RequestP2PGetBlockHeaders_Error1(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetBlockHeadersResponse{}
	p2pClientMock.On("GetBlockHeaders", mock.Anything, mock.Anything, mock.Anything).Return(resp, errors.New("internal error"))

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}
	blocksLength := 2

	hdrs, err := client.RequestP2PGetBlockHeaders(context.TODO(), make([]uint32, blocksLength), opts...)
	assert.NotNil(t, err)
	assert.Nil(t, hdrs)
}

func TestClient_RequestP2PGetBlockHeaders_Error2(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	bh1 := createBlockHeader(t, 1)
	bts1, err := bh1.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	bh2 := createBlockHeader(t, 2)
	bts2, err := bh2.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	resp := &pb.GetBlockHeadersResponse{
		BlockHeaders: [][]byte{bts1, bts2},
	}
	p2pClientMock.On("GetBlockHeaders", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}
	blocksLength := 1

	hdrs, err := client.RequestP2PGetBlockHeaders(context.TODO(), make([]uint32, blocksLength), opts...)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, errorz.ErrBadResponse))
	assert.Nil(t, hdrs)
}

func TestClient_RequestP2PGetPendingTx_Ok(t *testing.T) {
	txsLength := 1
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetPendingTxsResponse{
		Txs: make([][]byte, txsLength),
	}
	p2pClientMock.On("GetPendingTxs", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	txs, err := client.RequestP2PGetPendingTx(context.TODO(), make([][]byte, txsLength), opts...)
	assert.Nil(t, err)
	assert.NotNil(t, txs)
	assert.Len(t, txs, txsLength)
}

func TestClient_RequestP2PGetPendingTx_Error1(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetPendingTxsResponse{
		Txs: nil,
	}
	p2pClientMock.On("GetPendingTxs", mock.Anything, mock.Anything, mock.Anything).Return(resp, errors.New("internal error"))

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	txs, err := client.RequestP2PGetPendingTx(context.TODO(), make([][]byte, 1), opts...)
	assert.NotNil(t, err)
	assert.Nil(t, txs)
}

func TestClient_RequestP2PGetPendingTx_Error2(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetPendingTxsResponse{
		Txs: nil,
	}
	p2pClientMock.On("GetPendingTxs", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	txs, err := client.RequestP2PGetPendingTx(context.TODO(), make([][]byte, 1), opts...)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, errorz.ErrBadResponse))
	assert.Nil(t, txs)
}

func TestClient_RequestP2PGetPendingTx_Error3(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetPendingTxsResponse{
		Txs: make([][]byte, 0),
	}
	p2pClientMock.On("GetPendingTxs", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	txs, err := client.RequestP2PGetPendingTx(context.TODO(), make([][]byte, 1), opts...)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, errorz.ErrBadResponse))
	assert.Nil(t, txs)
}

func TestClient_RequestP2PGetMinedTxs_Ok(t *testing.T) {
	txsLength := 1
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetMinedTxsResponse{
		Txs: make([][]byte, txsLength),
	}
	p2pClientMock.On("GetMinedTxs", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	txs, err := client.RequestP2PGetMinedTxs(context.TODO(), make([][]byte, txsLength), opts...)
	assert.Nil(t, err)
	assert.NotNil(t, txs)
	assert.Len(t, txs, txsLength)
}

func TestClient_RequestP2PGetMinedTxs_Error1(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetMinedTxsResponse{
		Txs: nil,
	}
	p2pClientMock.On("GetMinedTxs", mock.Anything, mock.Anything, mock.Anything).Return(resp, errors.New("internal error"))

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	txs, err := client.RequestP2PGetMinedTxs(context.TODO(), make([][]byte, 1), opts...)
	assert.NotNil(t, err)
	assert.Nil(t, txs)
}

func TestClient_RequestP2PGetMinedTxs_Error2(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetMinedTxsResponse{
		Txs: nil,
	}
	p2pClientMock.On("GetMinedTxs", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	txs, err := client.RequestP2PGetMinedTxs(context.TODO(), make([][]byte, 1), opts...)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, errorz.ErrBadResponse))
	assert.Nil(t, txs)
}

func TestClient_RequestP2PGetSnapShotStateData_Ok(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetSnapShotStateDataResponse{
		Data: make([]byte, constants.HashLen),
	}
	p2pClientMock.On("GetSnapShotStateData", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	data, err := client.RequestP2PGetSnapShotStateData(context.TODO(), make([]byte, constants.HashLen), opts...)
	assert.Nil(t, err)
	assert.NotNil(t, data)
}

func TestClient_RequestP2PGetSnapShotStateData_Error1(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetSnapShotStateDataResponse{
		Data: nil,
	}
	p2pClientMock.On("GetSnapShotStateData", mock.Anything, mock.Anything, mock.Anything).Return(resp, errors.New("internal error"))

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	data, err := client.RequestP2PGetSnapShotStateData(context.TODO(), make([]byte, constants.HashLen), opts...)
	assert.NotNil(t, err)
	assert.Nil(t, data)
}

func TestClient_RequestP2PGetSnapShotStateData_Error2(t *testing.T) {
	p2pClientMock := &P2PClientMock{}
	resp := &pb.GetSnapShotStateDataResponse{
		Data: nil,
	}
	p2pClientMock.On("GetSnapShotStateData", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	client := &Client{}
	client.Init(p2pClientMock, &dynamics.Storage{})
	opts := []grpc.CallOption{}

	data, err := client.RequestP2PGetSnapShotStateData(context.TODO(), make([]byte, constants.HashLen), opts...)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, errorz.ErrBadResponse))
	assert.Nil(t, data)
}

func createBlockHeader(t *testing.T, length int) *objs.BlockHeader {
	bclaimsList, txHashListList, err := generateChain(length)
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

	return &objs.BlockHeader{
		BClaims:  bclaims,
		SigGroup: sig,
		TxHshLst: txHashListList[0],
	}
}

func generateChain(length int) ([]*objs.BClaims, [][][]byte, error) {
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
		Height:     1,
		TxCount:    1,
		PrevBlock:  crypto.Hasher([]byte("foo")),
		TxRoot:     txRoot,
		StateRoot:  crypto.Hasher([]byte("")),
		HeaderRoot: crypto.Hasher([]byte("")),
	}
	chain = append(chain, bclaims)
	for i := 1; i < length; i++ {
		bhsh, err := chain[i-1].BlockHash()
		if err != nil {
			return nil, nil, err
		}
		txhash := crypto.Hasher([]byte(strconv.Itoa(i)))
		txHshLst := [][]byte{txhash}
		txRoot, err := objs.MakeTxRoot(txHshLst)
		if err != nil {
			return nil, nil, err
		}
		txHashes = append(txHashes, txHshLst)
		bclaims := &objs.BClaims{
			ChainID:    1,
			Height:     uint32(len(chain) + 1),
			TxCount:    1,
			PrevBlock:  bhsh,
			TxRoot:     txRoot,
			StateRoot:  chain[i-1].StateRoot,
			HeaderRoot: chain[i-1].HeaderRoot,
		}
		chain = append(chain, bclaims)
	}
	return chain, txHashes, nil
}
