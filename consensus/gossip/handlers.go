package gossip

import (
	"context"
	"fmt"
	"time"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/dynamics"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/logging"
	pb "github.com/alicenet/alicenet/proto"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

var _ pb.P2PGossipProposalHandler = (*Handlers)(nil)
var _ pb.P2PGossipPreVoteHandler = (*Handlers)(nil)
var _ pb.P2PGossipPreVoteNilHandler = (*Handlers)(nil)
var _ pb.P2PGossipPreCommitHandler = (*Handlers)(nil)
var _ pb.P2PGossipPreCommitNilHandler = (*Handlers)(nil)
var _ pb.P2PGossipNextRoundHandler = (*Handlers)(nil)
var _ pb.P2PGossipNextHeightHandler = (*Handlers)(nil)
var _ pb.P2PGossipBlockHeaderHandler = (*Handlers)(nil)
var _ pb.P2PGossipTransactionHandler = (*Handlers)(nil)

type appHandler interface {
	PendingTxAdd(txn *badger.Txn, chainID uint32, height uint32, tx []interfaces.Transaction) error
	UnmarshalTx([]byte) (interfaces.Transaction, error)
}

// Handlers consumes gossip and updates local state
type Handlers struct {
	client pb.P2PClient

	database  *db.Database
	shandlers *lstate.Handlers
	sstore    *lstate.Store

	ctx       context.Context
	cancelCtx func()
	logger    *logrus.Logger

	// channels acts as per message queues with validation occurring before
	// the message is queued up
	app     appHandler
	storage dynamics.StorageGetter

	height      *mutexUint32
	chainID     *mutexUint32
	isSync      *mutexBool
	isValidator *mutexBool
	ReceiveLock chan interfaces.Lockable
}

func (mb *Handlers) getLock(ctx context.Context) (interfaces.Lockable, bool) {
	select {
	case <-ctx.Done():
		return nil, false
	case lock := <-mb.ReceiveLock:
		return lock, true
	case <-mb.ctx.Done():
		return nil, false
	}
}

// Init will initialize the gossip consumer
// it must be run at least once and will have no
// effect if run more than once
func (mb *Handlers) Init(chainID uint32, database *db.Database, client pb.P2PClient, app appHandler, handlers *lstate.Handlers, storage dynamics.StorageGetter) {
	mb.logger = logging.GetLogger(constants.LoggerGossipBus)
	mb.client = client
	mb.app = app
	mb.database = database
	mb.shandlers = handlers
	background := context.Background()
	ctx, cf := context.WithCancel(background)
	mb.cancelCtx = cf
	mb.ctx = ctx
	mb.storage = storage
	mb.height = &mutexUint32{}
	mb.chainID = &mutexUint32{value: chainID}
	mb.isSync = &mutexBool{}
	mb.ReceiveLock = make(chan interfaces.Lockable)
	mb.isValidator = &mutexBool{}
	mb.sstore = &lstate.Store{}
	mb.sstore.Init(database)
}

// Close will shut down the gossip system such that it can not be
// restarted
func (mb *Handlers) Close() {
	mb.cancelCtx()
}

// Done blocks until the service has an exit
func (mb *Handlers) Done() <-chan struct{} {
	return mb.ctx.Done()
}

func (mb *Handlers) Start() {
	heightSet := false
	for {
		<-time.After(3 * time.Second)
		if !heightSet {
			height := mb.height.Get()
			if height == 0 {
				mb.height.Set(1)
			}
		}
		height, chainID, isSync, isValidator, err := mb.heightAndSync()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			continue
		}
		if !heightSet {
			mb.chainID.Set(chainID)
		}
		mb.height.Set(height)
		mb.isSync.Set(isSync)
		mb.isValidator.Set(isValidator)
		heightSet = true
	}
}

func (mb *Handlers) heightAndSync() (uint32, uint32, bool, bool, error) {
	var height uint32
	var chainID uint32
	var maxHeight uint32
	var isValidator bool
	err := mb.database.View(func(txn *badger.Txn) error {
		os, err := mb.database.GetOwnState(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		chainID = os.SyncToBH.BClaims.ChainID
		isValidator, _, _, _, _, err = mb.sstore.GetDropData(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		height = os.SyncToBH.BClaims.Height + 1
		return nil
	})
	return maxHeight, chainID, maxHeight-height <= 2, isValidator, err
}

// HandleP2PGossipTransaction adds a transaction to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipTransaction(ctx context.Context, msg *pb.GossipTransactionMessage) (*pb.GossipTransactionAck, error) {
	select {
	case <-mb.ctx.Done():
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	default:
	}
	ack := &pb.GossipTransactionAck{}
	tx := msg.Transaction
	// isSync := mb.isSync.Get()
	// if !isSync {
	// 	return ack, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	// }
	chainID := mb.chainID.Get()
	height := mb.height.Get()
	mutex, ok := mb.getLock(ctx)
	if !ok {
		return ack, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	}
	mutex.Lock()
	defer mutex.Unlock()
	err := mb.database.Update(func(txn *badger.Txn) error {
		tx, err := mb.app.UnmarshalTx(tx)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return status.Error(codes.InvalidArgument, err.Error())
		}
		err = mb.app.PendingTxAdd(txn, chainID, height, []interfaces.Transaction{tx})
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return status.Error(codes.InvalidArgument, err.Error())
		}
		return nil
	})
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return ack, err
	}
	return ack, nil
}

// HandleP2PGossipProposal adds a proposal to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipProposal(ctx context.Context, msg *pb.GossipProposalMessage) (*pb.GossipProposalAck, error) {
	select {
	case <-mb.ctx.Done():
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	default:
	}
	ack := &pb.GossipProposalAck{}
	if !mb.isValidator.Get() {
		return ack, nil
	}
	rawmsg := msg.Proposal
	obj := &objs.Proposal{}
	err := obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	mutex, ok := mb.getLock(ctx)
	if !ok {
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	}
	mutex.Lock()
	defer mutex.Unlock()
	if err := mb.shandlers.AddProposal(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return ack, err
	}
	return ack, nil
}

// HandleP2PGossipPreVote adds a preVote to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipPreVote(ctx context.Context, msg *pb.GossipPreVoteMessage) (*pb.GossipPreVoteAck, error) {
	select {
	case <-mb.ctx.Done():
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	default:
	}
	ack := &pb.GossipPreVoteAck{}
	if !mb.isValidator.Get() {
		return ack, nil
	}
	rawmsg := msg.PreVote
	obj := &objs.PreVote{}
	err := obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return ack, err
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	mutex, ok := mb.getLock(ctx)
	if !ok {
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	}
	mutex.Lock()
	defer mutex.Unlock()
	if err := mb.shandlers.AddPreVote(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return ack, nil
}

// HandleP2PGossipPreVoteNil adds a preVoteNil to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipPreVoteNil(ctx context.Context, msg *pb.GossipPreVoteNilMessage) (*pb.GossipPreVoteNilAck, error) {
	select {
	case <-mb.ctx.Done():
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	default:
	}
	ack := &pb.GossipPreVoteNilAck{}
	if !mb.isValidator.Get() {
		return ack, nil
	}
	rawmsg := msg.PreVoteNil
	obj := &objs.PreVoteNil{}
	err := obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	mutex, ok := mb.getLock(ctx)
	if !ok {
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	}
	mutex.Lock()
	defer mutex.Unlock()
	if err := mb.shandlers.AddPreVoteNil(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return ack, nil
}

// HandleP2PGossipPreCommit adds a preCommit to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipPreCommit(ctx context.Context, msg *pb.GossipPreCommitMessage) (*pb.GossipPreCommitAck, error) {
	select {
	case <-mb.ctx.Done():
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	default:
	}
	ack := &pb.GossipPreCommitAck{}
	if !mb.isValidator.Get() {
		return ack, nil
	}
	rawmsg := msg.PreCommit
	obj := &objs.PreCommit{}
	err := obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	mutex, ok := mb.getLock(ctx)
	if !ok {
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	}
	mutex.Lock()
	defer mutex.Unlock()
	if err := mb.shandlers.AddPreCommit(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return ack, nil
}

// HandleP2PGossipPreCommitNil adds a preCommitNil to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipPreCommitNil(ctx context.Context, msg *pb.GossipPreCommitNilMessage) (*pb.GossipPreCommitNilAck, error) {
	select {
	case <-mb.ctx.Done():
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	default:
	}
	ack := &pb.GossipPreCommitNilAck{}
	if !mb.isValidator.Get() {
		return ack, nil
	}
	rawmsg := msg.PreCommitNil
	obj := &objs.PreCommitNil{}
	err := obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	mutex, ok := mb.getLock(ctx)
	if !ok {
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	}
	mutex.Lock()
	defer mutex.Unlock()
	if err := mb.shandlers.AddPreCommitNil(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return ack, nil
}

// HandleP2PGossipNextRound adds a nextRound to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipNextRound(ctx context.Context, msg *pb.GossipNextRoundMessage) (*pb.GossipNextRoundAck, error) {
	select {
	case <-mb.ctx.Done():
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	default:
	}
	ack := &pb.GossipNextRoundAck{}
	if !mb.isValidator.Get() {
		return ack, nil
	}
	rawmsg := msg.NextRound
	obj := &objs.NextRound{}
	err := obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	mutex, ok := mb.getLock(ctx)
	if !ok {
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	}
	mutex.Lock()
	defer mutex.Unlock()
	if err := mb.shandlers.AddNextRound(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return ack, nil
}

// HandleP2PGossipNextHeight adds a nextHeight to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipNextHeight(ctx context.Context, msg *pb.GossipNextHeightMessage) (*pb.GossipNextHeightAck, error) {
	select {
	case <-mb.ctx.Done():
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	default:
	}
	ack := &pb.GossipNextHeightAck{}
	if !mb.isValidator.Get() {
		return ack, nil
	}
	rawmsg := msg.NextHeight
	obj := &objs.NextHeight{}
	err := obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	mutex, ok := mb.getLock(ctx)
	if !ok {
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	}
	mutex.Lock()
	defer mutex.Unlock()
	if err := mb.shandlers.AddNextHeight(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return ack, nil
}

// HandleP2PGossipBlockHeader adds a nextHeight to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipBlockHeader(ctx context.Context, msg *pb.GossipBlockHeaderMessage) (*pb.GossipBlockHeaderAck, error) {
	select {
	case <-mb.ctx.Done():
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	default:
	}
	ack := &pb.GossipBlockHeaderAck{}
	rawmsg := msg.BlockHeader
	obj := &objs.BlockHeader{}
	err := obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, fmt.Errorf("BlockHeight:%d | SigGroup:%x | %q", obj.BClaims.Height, obj.SigGroup, err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	mutex, ok := mb.getLock(ctx)
	if !ok {
		return nil, status.Error(codes.Canceled, errorz.ErrClosing.Error())
	}
	mutex.Lock()
	defer mutex.Unlock()
	if err := mb.shandlers.AddBlockHeader(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return ack, nil
}
