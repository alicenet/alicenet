package localrpc

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/MadBase/MadNet/application"
	"github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/gossip"
	"github.com/MadBase/MadNet/consensus/lstate"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/dynamics"
	"github.com/MadBase/MadNet/logging"
	pb "github.com/MadBase/MadNet/proto"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

var _ pb.LocalStateGetBlockHeaderHandler = (*Handlers)(nil)
var _ pb.LocalStateGetPendingTransactionHandler = (*Handlers)(nil)
var _ pb.LocalStateGetRoundStateForValidatorHandler = (*Handlers)(nil)
var _ pb.LocalStateGetValidatorSetHandler = (*Handlers)(nil)
var _ pb.LocalStateGetBlockNumberHandler = (*Handlers)(nil)
var _ pb.LocalStateGetChainIDHandler = (*Handlers)(nil)
var _ pb.LocalStateGetEpochNumberHandler = (*Handlers)(nil)
var _ pb.LocalStateSendTransactionHandler = (*Handlers)(nil)
var _ pb.LocalStateGetDataHandler = (*Handlers)(nil)
var _ pb.LocalStateGetMinedTransactionHandler = (*Handlers)(nil)
var _ pb.LocalStateGetValueForOwnerHandler = (*Handlers)(nil)
var _ pb.LocalStateIterateNameSpaceHandler = (*Handlers)(nil)
var _ pb.LocalStateGetUTXOHandler = (*Handlers)(nil)

func (srpc *Handlers) notReady() error {
	if srpc.safe() {
		return nil
	}

	select {
	case <-srpc.ctx.Done():
		return errors.New("closing")
	case <-time.After(1 * time.Second):
		return errors.New("not in sync - unsafe to serve requests at this time")
	}
}

// Handlers is the server side of the local RPC system. Handlers dispatches
// requests to other systems for processing.
type Handlers struct {
	ctx       context.Context
	cancelCtx func()

	database *db.Database

	sstore *lstate.Store

	AppHandler *application.Application
	GossipBus  *gossip.Handlers
	Storage    dynamics.StorageGetter
	logger     *logrus.Logger

	ethAcct []byte
	EthPubk []byte

	safeHandler func() bool
	safecount   uint32
}

// Init will initialize the Consensus Engine and all sub modules
func (srpc *Handlers) Init(database *db.Database, app *application.Application, gh *gossip.Handlers, pubk []byte, safe func() bool, storage dynamics.StorageGetter) {
	background := context.Background()
	ctx, cf := context.WithCancel(background)
	srpc.cancelCtx = cf
	srpc.ctx = ctx
	srpc.logger = logging.GetLogger(constants.LoggerLocalRPC)
	srpc.database = database
	srpc.AppHandler = app
	srpc.GossipBus = gh
	srpc.Storage = storage
	srpc.EthPubk = pubk
	srpc.sstore = &lstate.Store{}
	srpc.sstore.Init(database)
	if len(srpc.EthPubk) > 0 {
		srpc.ethAcct = crypto.GetAccount(srpc.EthPubk)
	}
	srpc.safeHandler = safe
}

func (srpc *Handlers) Start() {
	srpc.SafeMonitor()
}

func (srpc *Handlers) Stop() {
	srpc.cancelCtx()
}

func (srpc *Handlers) safe() bool {
	return srpc.safecount != 0
}

func (srpc *Handlers) SafeMonitor() {
	for {
		select {
		case <-srpc.ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
		if !srpc.safeHandler() {
			if srpc.safecount > 0 {
				srpc.safecount--
			}
		} else {
			if srpc.safecount <= 6 {
				srpc.safecount++
			}
		}
		if srpc.safecount > 7 { //todo:HUNTER MOVE INTO SYNCHRONIZER FOR ERROR PROPOGATION
			panic("localRPC handler impossible state")
		}
	}
}

// HandleLocalStateGetPendingTransaction handles the get pending tx request
func (srpc *Handlers) HandleLocalStateGetPendingTransaction(ctx context.Context, req *pb.PendingTransactionRequest) (*pb.PendingTransactionResponse, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	srpc.logger.Debugf("HandleLocalStateGetPendingTransaction: %v", req)

	var height uint32
	var tx *objs.Tx
	err := srpc.database.View(func(txn *badger.Txn) error {
		os, err := srpc.database.GetOwnState(txn)
		if err != nil {
			return err
		}
		height = os.SyncToBH.BClaims.Height
		txHash, err := ReverseTranslateByte(req.TxHash)
		if err != nil {
			return err
		}
		if len(txHash) != 32 {
			return fmt.Errorf("invalid length for TxHash: %v", len(req.TxHash))
		}
		txi, missing, err := srpc.AppHandler.PendingTxGet(txn, height, [][]byte{txHash})
		if err != nil {
			return err
		}
		if len(missing) == 0 && len(txi) == 1 {
			tmp, ok := txi[0].(*objs.Tx)
			if !ok {
				return errors.New("server fault - state invalid for requested value")
			}
			tx = tmp
		} else {
			return fmt.Errorf("unknown transaction: %s", req.TxHash)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	objc, err := ForwardTranslateTx(tx)
	if err != nil {
		return nil, err
	}
	return &pb.PendingTransactionResponse{Tx: objc}, nil
}

func (srpc *Handlers) HandleLocalStateGetChainID(ctx context.Context, req *pb.ChainIDRequest) (*pb.ChainIDResponse, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	srpc.logger.Debugf("HandleLocalStateGetChainID: %v", req)
	var chainID uint32
	err := srpc.database.View(func(txn *badger.Txn) error {
		os, err := srpc.database.GetOwnState(txn)
		if err != nil {
			return err
		}
		chainID = os.SyncToBH.BClaims.ChainID
		return nil
	})
	if err != nil {
		return nil, err
	}
	result := &pb.ChainIDResponse{
		ChainID: chainID,
	}
	return result, nil
}

func (srpc *Handlers) HandleLocalStateSendTransaction(ctx context.Context, req *pb.TransactionData) (*pb.TransactionDetails, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	srpc.logger.Debugf("HandleLocalStateSendTransaction: %v", req)
	ntx, err := ReverseTranslateTx(req.Tx)
	if err != nil {
		return nil, err
	}
	txb, err := ntx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = ntx.UnmarshalBinary(txb)
	if err != nil {
		return nil, err
	}
	_, err = srpc.GossipBus.HandleP2PGossipTransaction(ctx, &pb.GossipTransactionMessage{Transaction: txb})
	if err != nil {
		return nil, err
	}
	txHash, err := ntx.TxHash()
	if err != nil {
		return nil, err
	}
	result := &pb.TransactionDetails{TxHash: hex.EncodeToString(txHash)}
	return result, nil
}

// HandleLocalStateGetValueForOwner ...
func (srpc *Handlers) HandleLocalStateGetValueForOwner(ctx context.Context, req *pb.GetValueRequest) (*pb.GetValueResponse, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	srpc.logger.Debugf("HandleLocalStateGetValueForOwner: %v", req)
	minValue := uint256.Max()
	if req.Minvalue != "" {
		err := minValue.UnmarshalString(req.Minvalue)
		if err != nil {
			return nil, err
		}
	}

	accountH := req.Account

	account, err := ReverseTranslateByte(accountH)
	if err != nil {
		return nil, err
	}
	if len(account) != 20 {
		return nil, fmt.Errorf("invalid length (%v) for Account:%s", len(req.Account), req.Account)
	}
	var utxoIDs [][]byte
	var value *uint256.Uint256
	var paginationToken *objs.PaginationToken
	var height uint32
	err = srpc.database.View(func(txn *badger.Txn) error {
		tmp, v, pt, err := srpc.AppHandler.GetValueForOwner(txn, constants.CurveSpec(req.CurveSpec), account, minValue, req.PaginationToken)
		if err != nil {
			return err
		}
		utxoIDs = tmp
		value = v
		paginationToken = pt

		os, err := srpc.database.GetOwnState(txn)
		if err != nil {
			return err
		}
		height = os.SyncToBH.BClaims.Height

		return nil
	})

	if err != nil {
		return nil, err
	}

	out, err := ForwardTranslateByteSlice(utxoIDs)
	if err != nil {
		return nil, err
	}

	valueString, err := value.MarshalString()
	if err != nil {
		return nil, err
	}

	var ptBytesRet []byte
	if paginationToken != nil {
		ptBytesRet, err = paginationToken.MarshalBinary()
		if err != nil {
			return nil, err
		}
	}

	result := &pb.GetValueResponse{TotalValue: valueString, UTXOIDs: out, PaginationToken: ptBytesRet, BlockHeight: height}
	return result, nil
}

func (srpc *Handlers) HandleLocalStateGetBlockNumber(ctx context.Context, req *pb.BlockNumberRequest) (*pb.BlockNumberResponse, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	srpc.logger.Debugf("HandleLocalStateGetBlockNumber: %v", req)
	var height uint32
	err := srpc.database.View(func(txn *badger.Txn) error {
		os, err := srpc.database.GetOwnState(txn)
		if err != nil {
			return err
		}
		height = os.SyncToBH.BClaims.Height
		return nil
	})
	if err != nil {
		return nil, err
	}
	result := &pb.BlockNumberResponse{BlockHeight: height}
	return result, nil
}

func (srpc *Handlers) HandleLocalStateGetBlockHeader(ctx context.Context, req *pb.BlockHeaderRequest) (*pb.BlockHeaderResponse, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	srpc.logger.Debugf("HandleLocalStateGetBlockHeader: %v", req)
	if req.Height == 0 {
		return nil, errors.New("height cannot be zero")
	}
	var bh *pb.BlockHeader
	err := srpc.database.View(func(txn *badger.Txn) error {
		bhh, err := srpc.database.GetCommittedBlockHeader(txn, req.Height)
		if err != nil {
			return err
		}
		tmp, err := ForwardTranslateBlockHeader(bhh)
		if err != nil {
			return err
		}
		bh = tmp
		return nil
	})
	if err != nil {
		return nil, err
	}
	result := &pb.BlockHeaderResponse{BlockHeader: bh}
	return result, nil
}

func (srpc *Handlers) HandleLocalStateGetRoundStateForValidator(ctx context.Context, req *pb.RoundStateForValidatorRequest) (*pb.RoundStateForValidatorResponse, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	srpc.logger.Debugf("HandleLocalStateGetRoundStateForValidator: %v", req)
	return nil, errors.New("not implemented")
	/*
	 */
}

func (srpc *Handlers) HandleLocalStateGetValidatorSet(ctx context.Context, req *pb.ValidatorSetRequest) (*pb.ValidatorSetResponse, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	srpc.logger.Debugf("HandleLocalStateGetValidatorSet: %v", req)
	return nil, errors.New("not implemented")
	/*
	     var vs []byte
	   	err := srpc.database.View(func(txn *badger.Txn) error {
	   		tmp, err := srpc.database.GetValidatorSet(txn, req.Height)
	   		if err != nil {
	   			return err
	   		}
	   		vsb, err := tmp.MarshalBinary()
	   		if err != nil {
	   			return err
	   		}
	   		vs = vsb
	   		return nil
	   	})
	   	if err != nil {
	   		return nil, err
	   	}
	   	result := &pb.ValidatorSetResponse{} //ValidatorSet: vs}
	   	return result, nil
	*/
}

func (srpc *Handlers) HandleLocalStateGetEpochNumber(ctx context.Context, req *pb.EpochNumberRequest) (*pb.EpochNumberResponse, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	srpc.logger.Debugf("HandleLocalStateGetEpochNumber: %v", req)
	var en uint32
	err := srpc.database.View(func(txn *badger.Txn) error {
		os, err := srpc.database.GetOwnState(txn)
		if err != nil {
			return err
		}
		en = utils.Epoch(os.SyncToBH.BClaims.Height)
		return nil
	})
	if err != nil {
		return nil, err
	}
	result := &pb.EpochNumberResponse{Epoch: en}
	return result, nil
}

// HandleLocalStateGetData ...
func (srpc *Handlers) HandleLocalStateGetData(ctx context.Context, req *pb.GetDataRequest) (*pb.GetDataResponse, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	var data []byte
	srpc.logger.Debugf("HandleLocalStateGetData: %v", req)

	err := srpc.database.View(func(txn *badger.Txn) error {
		account, err := ReverseTranslateByte(req.Account)
		if err != nil {
			return err
		}
		if len(account) != 20 {
			return fmt.Errorf("invalid length (%v) for Account:%s", len(req.Account), req.Account)
		}
		index, err := ReverseTranslateByte(req.Index)
		if err != nil {
			return err
		}
		if len(index) != 32 {
			return fmt.Errorf("invalid length (%v) for Index:%s", len(req.Index), req.Index)
		}
		tmp, err := srpc.AppHandler.UTXOGetData(txn, constants.CurveSpec(req.CurveSpec), account, index)
		if err != nil {
			return err
		}
		data = tmp
		return nil
	})
	if err != nil {
		return nil, err
	}
	d := ForwardTranslateByte(data)

	result := &pb.GetDataResponse{Rawdata: d}
	return result, nil
}

// HandleLocalStateIterateNameSpace ...
func (srpc *Handlers) HandleLocalStateIterateNameSpace(ctx context.Context, req *pb.IterateNameSpaceRequest) (*pb.IterateNameSpaceResponse, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	result := []*pb.IterateNameSpaceResponse_Result{}
	srpc.logger.Debugf("HandleLocalStateIterateNameSpace: %v", req)

	if req.Number > 256 {
		return nil, fmt.Errorf("number is not allowed to be greater than 256; got %v", req.Number)
	}

	err := srpc.database.View(func(txn *badger.Txn) error {
		a, err := ReverseTranslateByte(req.Account)
		if err != nil {
			return err
		}
		if len(a) != 20 {
			return fmt.Errorf("invalid length (%v) for account:%s", len(req.Account), req.Account)
		}
		si, err := ReverseTranslateByte(req.StartIndex)
		if err != nil {
			return err
		}
		if len(si) > 0 {
			if len(si) != 32 {
				return fmt.Errorf("StartIndex must be empty or valid; invalid length (%v) for StartIndex:%s", len(req.StartIndex), req.StartIndex)
			}
		}
		n := req.Number
		if n > 256 {
			n = 256
		}
		height := uint32(0)
		err = srpc.database.View(func(txn *badger.Txn) error {
			os, err := srpc.database.GetOwnState(txn)
			if err != nil {
				return err
			}
			// do not return the ds if it will expire in the next 1 blocks
			// todo: find a better rational for cutoff of return
			height = os.SyncToBH.BClaims.Height + 1
			return nil
		})
		if err != nil {
			return err
		}
		tmp, err := srpc.AppHandler.PaginateDataByOwner(txn, constants.CurveSpec(req.CurveSpec), a, height, int(n), si)
		if err != nil {
			return err
		}
		for i := 0; i < len(tmp); i++ {
			tmpUTXOID := ForwardTranslateByte(tmp[i].UTXOID)

			tmpIndex := ForwardTranslateByte(tmp[i].Index)

			itm := &pb.IterateNameSpaceResponse_Result{
				UTXOID: tmpUTXOID,
				Index:  tmpIndex,
			}
			result = append(result, itm)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := &pb.IterateNameSpaceResponse{Results: result}
	return resp, nil
}

// HandleLocalStateGetUTXO ...
func (srpc *Handlers) HandleLocalStateGetUTXO(ctx context.Context, req *pb.UTXORequest) (*pb.UTXOResponse, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	srpc.logger.Debugf("HandleLocalStateGetUTXO: %v", req)
	for i := 0; i < len(req.UTXOIDs); i++ {
		if len(req.UTXOIDs[i]) != 64 {
			return nil, fmt.Errorf("invalid length (%v) for utxoID@index %v out of %v; %s", len(req.UTXOIDs[i]), i, len(req.UTXOIDs), req.UTXOIDs)
		}
	}
	d, err := ReverseTranslateByteSlice(req.UTXOIDs)
	if err != nil {
		return nil, err
	}

	var utxos []*objs.TXOut
	err = srpc.database.View(func(txn *badger.Txn) error {
		tmp, err := srpc.AppHandler.UTXOGet(txn, d)
		if err != nil {
			return err
		}
		utxos = tmp
		return nil
	})
	if err != nil {
		return nil, err
	}
	result := []*pb.TXOut{}
	for _, u := range utxos {
		u2, err := ForwardTranslateTXOut(u)
		if err != nil {
			return nil, err
		}
		result = append(result, u2)
	}
	out := &pb.UTXOResponse{UTXOs: result}
	return out, nil
}

// HandleLocalStateGetMinedTransaction ...
func (srpc *Handlers) HandleLocalStateGetMinedTransaction(ctx context.Context, req *pb.MinedTransactionRequest) (*pb.MinedTransactionResponse, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	srpc.logger.Debugf("HandleLocalStateGetMinedTransaction: %v", req)
	d, err := ReverseTranslateByte(req.TxHash)
	if err != nil {
		return nil, err
	}
	if len(d) != 32 {
		return nil, fmt.Errorf("invalid length for TxHash: %v", len(req.TxHash))
	}
	var tx *objs.Tx
	err = srpc.database.View(func(txn *badger.Txn) error {
		txi, missing, err := srpc.AppHandler.MinedTxGet(txn, [][]byte{d})
		if err != nil {
			return err
		}
		if len(missing) == 0 && len(txi) == 1 {
			tmp, ok := txi[0].(*objs.Tx)
			if !ok {
				return errors.New("server fault - state invalid for requested value")
			}
			tx = tmp
		} else {
			return fmt.Errorf("unknown transaction: %s", req.TxHash)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	txOut, err := ForwardTranslateTx(tx)
	if err != nil {
		return nil, err
	}
	result := &pb.MinedTransactionResponse{Tx: txOut}
	return result, nil
}

// HandleLocalStateGetTransactionStatus ...
func (srpc *Handlers) HandleLocalStateGetTransactionStatus(ctx context.Context, req *pb.TransactionStatusRequest) (*pb.TransactionStatusResponse, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	srpc.logger.Debugf("HandleLocalStateGetTransactionStatus: %v", req)
	txHash, err := ReverseTranslateByte(req.TxHash)
	if err != nil {
		return nil, err
	}
	if len(txHash) != 32 {
		return nil, fmt.Errorf("invalid length for TxHash: %v", len(req.TxHash))
	}

	var tx *objs.Tx
	var isMined bool
	err = srpc.database.View(func(txn *badger.Txn) error {
		txi, missing, err := srpc.AppHandler.MinedTxGet(txn, [][]byte{txHash})
		if err == nil && len(missing) == 0 && len(txi) == 1 {
			tmp, ok := txi[0].(*objs.Tx)
			if ok {
				tx = tmp
				isMined = true
				return nil
			}
		}

		os, err2 := srpc.database.GetOwnState(txn)
		if err2 != nil {
			return err2
		}
		txi, missing, err2 = srpc.AppHandler.PendingTxGet(txn, os.SyncToBH.BClaims.Height, [][]byte{txHash})
		if err2 != nil {
			if err != nil {
				return fmt.Errorf("%v\n%v", err, err2)
			}
			return err2
		}
		if len(missing) == 0 && len(txi) == 1 {
			tmp, ok := txi[0].(*objs.Tx)
			if ok {
				tx = tmp
				isMined = false
				return nil
			}
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("unknown transaction: %s", req.TxHash)
	})
	if err != nil {
		return nil, err
	}

	result := &pb.TransactionStatusResponse{IsMined: isMined}
	if req.ReturnTx {
		txOut, err := ForwardTranslateTx(tx)
		if err != nil {
			return nil, err
		}
		result.Tx = txOut
	}
	return result, nil
}

func (srpc *Handlers) HandleLocalStateGetTxBlockNumber(ctx context.Context, req *pb.TxBlockNumberRequest) (*pb.TxBlockNumberResponse, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	var height uint32
	srpc.logger.Debugf("HandleLocalStateGetTxBlockNumber: %v", req)

	err := srpc.database.View(func(txn *badger.Txn) error {
		d, err := ReverseTranslateByte(req.TxHash)
		if err != nil {
			return err
		}
		if len(d) != 32 {
			return fmt.Errorf("invalid length (%v) for TxHash:%s", len(req.TxHash), req.TxHash)
		}
		tmp, err := srpc.AppHandler.GetHeightForTx(txn, d)
		if err != nil {
			return err
		}
		height = tmp
		return nil
	})
	if err != nil {
		return nil, err
	}
	result := &pb.TxBlockNumberResponse{BlockHeight: height}
	return result, nil
}

func (srpc *Handlers) HandleLocalStateGetFees(ctx context.Context, req *pb.FeeRequest) (*pb.FeeResponse, error) {
	if err := srpc.notReady(); err != nil {
		return nil, err
	}

	sg := srpc.Storage
	txFee := sg.GetMinTxFee()
	txfs, err := bigIntToString(txFee)
	if err != nil {
		return nil, err
	}

	vsFee := sg.GetValueStoreFee()
	vsfs, err := bigIntToString(vsFee)
	if err != nil {
		return nil, err
	}
	dsFee := sg.GetDataStoreEpochFee()
	dsfs, err := bigIntToString(dsFee)
	if err != nil {
		return nil, err
	}
	asFee := sg.GetAtomicSwapFee()
	asfs, err := bigIntToString(asFee)
	if err != nil {
		return nil, err
	}
	result := &pb.FeeResponse{
		MinTxFee:      txfs,
		ValueStoreFee: vsfs,
		DataStoreFee:  dsfs,
		AtomicSwapFee: asfs,
	}

	return result, nil
}

func bigIntToString(b *big.Int) (string, error) {
	bu, err := new(uint256.Uint256).FromBigInt(b)
	if err != nil {
		return "", err
	}
	bs, err := bu.MarshalString()
	if err != nil {
		return "", err
	}
	return bs, nil
}
