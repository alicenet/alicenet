package localrpc

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	aobjs "github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	pb "github.com/MadBase/MadNet/proto"
	"google.golang.org/grpc"
)

// Client is a wrapper around the gRPC local state server. This wrapper
// abstracts all types back to the native type system rather than
// passing around the protobufs themselves
// All methods take in a context. Iif the provided context does not have a
// Deadline set on it, constants.MsgTimeout will be set as the timeout for
// a given request. This timeout is four seconds. The server will return
// timeout after three seconds to ensure that the client is informed about the
// success of a request with a large probability of success.
type Client struct {
	sync.Mutex
	closeChan   chan struct{}
	closeOnce   sync.Once
	Address     string
	TimeOut     time.Duration
	conn        *grpc.ClientConn
	client      pb.LocalStateClient
	wg          sync.WaitGroup
	isConnected bool
}

// Connect establishes communication between the client and the server
func (lrpc *Client) Connect(ctx context.Context) error {
	err := func() error {
		fmt.Println("connecting")
		lrpc.Lock()
		defer lrpc.Unlock()
		if lrpc.isConnected {
			return errors.New("already connected")
		}
		if lrpc.TimeOut == 0 {
			lrpc.TimeOut = constants.MsgTimeout
		}
		// Set up a connection to the server.
		conn, err := grpc.Dial(lrpc.Address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(lrpc.TimeOut))
		if err != nil {
			return err
		}
		lrpc.conn = conn
		lrpc.client = pb.NewLocalStateClient(conn)
		lrpc.isConnected = true
		lrpc.closeChan = make(chan struct{})
		lrpc.closeOnce = sync.Once{}
		return nil
	}()
	if err != nil {
		return err
	}
	fmt.Println("pinging server")
	_, err = lrpc.GetBlockNumber(ctx)
	if err != nil {
		fmt.Printf("error pinging server: %v", err)
		return err
	}
	fmt.Println("connected")
	return nil
}

// Close termnates the client connection pool
// Close may be called more than once.
// A client should not be used after close or an error will be returned.
func (lrpc *Client) Close() error {
	var err error
	lrpc.closeOnce.Do(func() {
		lrpc.Lock()
		defer lrpc.Unlock()
		if !lrpc.isConnected {
			lrpc.closeChan = make(chan struct{})
		}
		close(lrpc.closeChan)
		err = lrpc.conn.Close()
		lrpc.wg.Wait()
	})
	return err
}

func (lrpc *Client) entrancyGuard() error {
	lrpc.Lock()
	defer lrpc.Unlock()
	if !lrpc.isConnected {
		return errors.New("connection closed")
	}
	select {
	case <-lrpc.closeChan:
		return errors.New("closing")
	default:
		lrpc.wg.Add(1)
		return nil
	}
}

func (lrpc *Client) contextGuard(ctx context.Context) (context.Context, func()) {
	if _, ok := ctx.Deadline(); !ok {
		return context.WithTimeout(ctx, lrpc.TimeOut)
	}
	return ctx, func() {}
}

// GetBlockHeader allows a caller to request a BlockHeader by height
func (lrpc *Client) GetBlockHeader(ctx context.Context, height uint32) (*objs.BlockHeader, error) {
	if err := lrpc.entrancyGuard(); err != nil {
		return nil, err
	}
	defer lrpc.wg.Done()
	subCtx, cleanup := lrpc.contextGuard(ctx)
	defer cleanup()

	request := &pb.BlockHeaderRequest{
		Height: height,
	}
	resp, err := lrpc.client.GetBlockHeader(subCtx, request)
	if err != nil {
		return nil, err
	}
	data := resp.BlockHeader
	bh, err := ReverseTranslateBlockHeader(data)
	if err != nil {
		return nil, err
	}
	return bh, nil
}

// GetBlockNumber returns the current block number
func (lrpc *Client) GetBlockNumber(ctx context.Context) (uint32, error) {
	if err := lrpc.entrancyGuard(); err != nil {
		return 0, err
	}
	defer lrpc.wg.Done()
	subCtx, cleanup := lrpc.contextGuard(ctx)
	defer cleanup()

	request := &pb.BlockNumberRequest{}
	resp, err := lrpc.client.GetBlockNumber(subCtx, request)
	if err != nil {
		return 0, err
	}
	data := resp.BlockHeight
	return data, nil
}

// GetEpochNumber returns the current epoch number
func (lrpc *Client) GetEpochNumber(ctx context.Context) (uint32, error) {
	if err := lrpc.entrancyGuard(); err != nil {
		return 0, err
	}
	defer lrpc.wg.Done()
	subCtx, cleanup := lrpc.contextGuard(ctx)
	defer cleanup()

	request := &pb.EpochNumberRequest{}
	resp, err := lrpc.client.GetEpochNumber(subCtx, request)
	if err != nil {
		return 0, err
	}
	data := resp.Epoch
	return data, nil
}

// SendTransaction allows the caller to inject a tx into the pending tx pool
func (lrpc *Client) SendTransaction(ctx context.Context, tx *aobjs.Tx) ([]byte, error) {
	if err := lrpc.entrancyGuard(); err != nil {
		return nil, err
	}
	defer lrpc.wg.Done()
	subCtx, cleanup := lrpc.contextGuard(ctx)
	defer cleanup()

	txb, err := ForwardTranslateTx(tx)
	if err != nil {
		return nil, err
	}
	request := &pb.TransactionData{Tx: txb}
	resp, err := lrpc.client.SendTransaction(subCtx, request)
	if err != nil {
		return nil, err
	}
	data := resp.TxHash
	return hex.DecodeString(data)
}

// GetValueForOwner allows a caller to receive a list of UTXOs that are
// controlled by the named account
func (lrpc *Client) GetValueForOwner(ctx context.Context, curveSpec constants.CurveSpec, account []byte, minValue *uint256.Uint256) ([][]byte, *uint256.Uint256, error) {
	if err := lrpc.entrancyGuard(); err != nil {
		return nil, nil, err
	}
	defer lrpc.wg.Done()
	subCtx, cleanup := lrpc.contextGuard(ctx)
	defer cleanup()

	o := ForwardTranslateByte(account)

	minValueString, err := minValue.MarshalString()
	if err != nil {
		return nil, nil, err
	}
	request := &pb.GetValueRequest{Account: o, CurveSpec: uint32(curveSpec), Minvalue: minValueString}
	resp, err := lrpc.client.GetValueForOwner(subCtx, request)
	if err != nil {
		return nil, nil, err
	}
	data := resp.UTXOIDs
	vString := resp.TotalValue
	v := &uint256.Uint256{}
	err = v.UnmarshalString(vString)
	if err != nil {
		return nil, nil, err
	}
	d, err := ReverseTranslateByteSlice(data)
	if err != nil {
		return nil, nil, err
	}
	return d, v, nil
}

// GetUTXO allows the caller to request UTXOs by ID
func (lrpc *Client) GetUTXO(ctx context.Context, utxoIDs [][]byte) (aobjs.Vout, error) {
	if err := lrpc.entrancyGuard(); err != nil {
		return nil, err
	}
	defer lrpc.wg.Done()
	subCtx, cleanup := lrpc.contextGuard(ctx)
	defer cleanup()

	d, err := ForwardTranslateByteSlice(utxoIDs)
	if err != nil {
		return nil, err
	}
	request := &pb.UTXORequest{UTXOIDs: d}
	resp, err := lrpc.client.GetUTXO(subCtx, request)
	if err != nil {
		return nil, err
	}
	utxos := []*aobjs.TXOut{}
	for _, utxo := range resp.UTXOs {
		u, err := ReverseTranslateTXOut(utxo)
		if err != nil {
			return nil, err
		}
		utxos = append(utxos, u)
	}
	return utxos, nil
}

// GetMinedTransaction allows a caller to see if a mined tx is known. Due to
// state pruning, transactions will only be stored for a maximum of four epochs.
// after this time, the transaction is no longer available but all UTXOs are.
func (lrpc *Client) GetMinedTransaction(ctx context.Context, txHash []byte) (*aobjs.Tx, error) {
	if err := lrpc.entrancyGuard(); err != nil {
		return nil, err
	}
	defer lrpc.wg.Done()
	subCtx, cleanup := lrpc.contextGuard(ctx)
	defer cleanup()

	txHash2 := ForwardTranslateByte(txHash)

	request := &pb.MinedTransactionRequest{TxHash: txHash2}
	resp, err := lrpc.client.GetMinedTransaction(subCtx, request)
	if err != nil {
		return nil, err
	}
	tx, err := ReverseTranslateTx(resp.Tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// GetPendingTransaction allows a caller to inspect the Pending Tx Pool to see
// if a tx is present
func (lrpc *Client) GetPendingTransaction(ctx context.Context, txHash []byte) (*aobjs.Tx, error) {
	if err := lrpc.entrancyGuard(); err != nil {
		return nil, err
	}
	defer lrpc.wg.Done()
	subCtx, cleanup := lrpc.contextGuard(ctx)
	defer cleanup()

	txHash2 := ForwardTranslateByte(txHash)

	request := &pb.PendingTransactionRequest{TxHash: txHash2}
	resp, err := lrpc.client.GetPendingTransaction(subCtx, request)
	if err != nil {
		return nil, err
	}
	tx, err := ReverseTranslateTx(resp.Tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// GetData returns only the state stored in a datastore
func (lrpc *Client) GetData(ctx context.Context, curveSpec constants.CurveSpec, account []byte, index []byte) ([]byte, error) {
	if err := lrpc.entrancyGuard(); err != nil {
		return nil, err
	}
	defer lrpc.wg.Done()
	subCtx, cleanup := lrpc.contextGuard(ctx)
	defer cleanup()

	o := ForwardTranslateByte(account)

	i := ForwardTranslateByte(index)

	request := &pb.GetDataRequest{Account: o, CurveSpec: uint32(curveSpec), Index: i}
	resp, err := lrpc.client.GetData(subCtx, request)
	if err != nil {
		return nil, err
	}
	do, err := ReverseTranslateByte(resp.Rawdata)
	if err != nil {
		return nil, err
	}
	return do, nil
}

// PaginateDataStoreUTXOByOwner allows an account namespace to be iterated
// so that all datastores may be observed. It also allows the query of a single
// datastore by setting the num param to 1 and the startIndex param to the
// target index of the desired datastore
// It returns a list of tuples of the form ( <utxoID>, <index> )
func (lrpc *Client) PaginateDataStoreUTXOByOwner(ctx context.Context, curveSpec constants.CurveSpec, account []byte, num uint8, startIndex []byte) ([]*aobjs.PaginationResponse, error) {
	if err := lrpc.entrancyGuard(); err != nil {
		return nil, err
	}
	defer lrpc.wg.Done()
	subCtx, cleanup := lrpc.contextGuard(ctx)
	defer cleanup()

	o := ForwardTranslateByte(account)

	i := ForwardTranslateByte(startIndex)

	request := &pb.IterateNameSpaceRequest{Account: o, CurveSpec: uint32(curveSpec), StartIndex: i, Number: uint32(num)}
	resp, err := lrpc.client.IterateNameSpace(subCtx, request)
	if err != nil {
		return nil, err
	}
	result := []*aobjs.PaginationResponse{}
	for i := 0; i < len(resp.Results); i++ {
		tmpUTXOID, err := ReverseTranslateByte(resp.Results[i].UTXOID)
		if err != nil {
			return nil, err
		}
		tmpIndex, err := ReverseTranslateByte(resp.Results[i].Index)
		if err != nil {
			return nil, err
		}
		tmp := &aobjs.PaginationResponse{
			UTXOID: tmpUTXOID,
			Index:  tmpIndex,
		}
		result = append(result, tmp)
	}
	return result, nil
}

// GetBlockHeightForTx returns the block height at which a tx was mined
func (lrpc *Client) GetBlockHeightForTx(ctx context.Context, txHash []byte) (uint32, error) {
	if err := lrpc.entrancyGuard(); err != nil {
		return 0, err
	}
	defer lrpc.wg.Done()
	subCtx, cleanup := lrpc.contextGuard(ctx)
	defer cleanup()

	txh := ForwardTranslateByte(txHash)

	request := &pb.TxBlockNumberRequest{TxHash: txh}
	resp, err := lrpc.client.GetTxBlockNumber(subCtx, request)
	if err != nil {
		return 0, err
	}
	return resp.BlockHeight, nil
}

// TODO: Not tested and may not work
func (lrpc *Client) GetTxFees(ctx context.Context) ([]string, error) {
	if err := lrpc.entrancyGuard(); err != nil {
		return nil, err
	}
	defer lrpc.wg.Done()
	subCtx, cleanup := lrpc.contextGuard(ctx)
	defer cleanup()

	request := &pb.FeeRequest{}
	response, err := lrpc.client.GetFees(subCtx, request)
	if err != nil {
		return nil, err
	}
	resp := []string{}

	re := regexp.MustCompile(`"[^"]+"`)
	newStrs := re.FindAllString(response.String(), -1)
	for _, s := range newStrs {
		s = s[1:]
		s = s[:len(s)-1]
		resp = append(resp, s)
	}

	return resp, nil
}
