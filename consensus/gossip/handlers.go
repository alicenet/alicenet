package gossip

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"

	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/lstate"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/logging"
	pb "github.com/MadBase/MadNet/proto"
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

	wg        sync.WaitGroup
	database  *db.Database
	shandlers *lstate.Handlers

	ctx       context.Context
	cancelCtx func()
	logger    *logrus.Logger

	// channels acts as per message queues with validation occurring before
	// the message is queued up
	iNtxChan  chan *transactionMsg
	iNpChan   chan *proposalMsg
	iNpvChan  chan *preVoteMsg
	iNpvnChan chan *preVoteNilMsg
	iNpcChan  chan *preCommitMsg
	iNpcnChan chan *preCommitNilMsg
	iNnrChan  chan *nextRoundMsg
	iNnhChan  chan *nextHeightMsg
	iNbhChan  chan *blockHeaderMsg

	app appHandler
}

// Init will initialize the gossip consumer
// it must be run at least once and will have no
// effect if run more than once
func (mb *Handlers) Init(database *db.Database, client pb.P2PClient, app appHandler, handlers *lstate.Handlers) error {
	mb.logger = logging.GetLogger(constants.LoggerGossipBus)
	mb.client = client
	mb.app = app
	mb.database = database
	mb.shandlers = handlers
	mb.iNtxChan = make(chan *transactionMsg)
	mb.iNpChan = make(chan *proposalMsg)
	mb.iNpvChan = make(chan *preVoteMsg)
	mb.iNpvnChan = make(chan *preVoteNilMsg)
	mb.iNpcChan = make(chan *preCommitMsg)
	mb.iNpcnChan = make(chan *preCommitNilMsg)
	mb.iNnrChan = make(chan *nextRoundMsg)
	mb.iNnhChan = make(chan *nextHeightMsg)
	mb.iNbhChan = make(chan *blockHeaderMsg)
	background := context.Background()
	ctx, cf := context.WithCancel(background)
	mb.cancelCtx = cf
	mb.ctx = ctx
	mb.wg = sync.WaitGroup{}
	return nil
}

// Close will shut down the gossip system such that it can not be
// restarted
func (mb *Handlers) Close() {
	mb.cancelCtx()
}

// Done blocks until the service has an exit
func (mb *Handlers) Done() <-chan struct{} {
	mb.wg.Wait()
	return mb.ctx.Done()
}

func (mb *Handlers) sendErr(ctx context.Context, err error, eC chan<- error) bool {
	select {
	case eC <- err:
		return true
	case <-ctx.Done():
		eC <- ctx.Err()
		return true
	case <-mb.ctx.Done():
		return true
	}
}

func (mb *Handlers) handleStateUpdateErrors(ctx context.Context, err error, eC chan<- error) error {
	etestStale := &errorz.ErrStale{}
	if errors.As(err, &etestStale) {
		mb.sendErr(ctx, err, eC)
		return nil
	}
	etestInvalid := &errorz.ErrInvalid{}
	if errors.As(err, &etestInvalid) {
		mb.sendErr(ctx, err, eC)
		return nil
	}
	mb.sendErr(ctx, nil, eC)
	return err
}

func (mb *Handlers) shouldDrop(ctx context.Context, cf func(), eC chan<- error) (bool, error) {
	_, isSync, err := mb.heightAndSync()
	if err != nil {
		return true, err
	}
	if !isSync {
		cf()
		select {
		case <-ctx.Done():
			eC <- ctx.Err()
			return true, nil
		case <-mb.ctx.Done():
			eC <- errorz.ErrClosing
			return true, nil
		}
	}
	return false, nil
}

func (mb *Handlers) isDone(ctx context.Context, eC chan<- error) bool {
	select {
	case <-ctx.Done():
		eC <- ctx.Err()
		return true
	case <-mb.ctx.Done():
		eC <- errorz.ErrClosing
		return true
	default:
		return false
	}
}

func (mb *Handlers) heightAndSync() (uint32, bool, error) {
	var height uint32
	var maxHeight uint32
	err := mb.database.View(func(txn *badger.Txn) error {
		os, err := mb.database.GetOwnState(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		maxHeight = os.MaxBHSeen.BClaims.Height + 1
		height = os.SyncToBH.BClaims.Height + 1
		return nil
	})
	return height, maxHeight-height <= 1, err
}

// UpdateStateFromGossip stores messages from a remote node
func (mb *Handlers) UpdateStateFromGossip(forceExit <-chan struct{}, mut interfaces.Lockable, safe func() bool) error {
	for {
		runtime.Gosched()
		select {
		case <-forceExit:
			return errorz.ErrClosing
		default:
		}
		if !safe() {
			select {
			case <-time.After(200 * time.Millisecond):
				continue
			case <-forceExit:
				return errorz.ErrClosing
			}
		}
		select {
		case <-forceExit:
			return errorz.ErrClosing
		case <-mb.ctx.Done():
			return nil
		case obj := <-mb.iNpChan:
			drop, err := mb.shouldDrop(obj.ctx, obj.cancel, obj.errC)
			if err != nil {
				return err
			}
			if drop {
				continue
			}
			mut.Lock()
			err = mb.handleStateUpdateErrors(obj.ctx, mb.shandlers.AddProposal(obj.msg), obj.errC)
			mut.Unlock()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
				return err
			}
		case obj := <-mb.iNpvChan:
			drop, err := mb.shouldDrop(obj.ctx, obj.cancel, obj.errC)
			if err != nil {
				return err
			}
			if drop {
				continue
			}
			mut.Lock()
			err = mb.handleStateUpdateErrors(obj.ctx, mb.shandlers.AddPreVote(obj.msg), obj.errC)
			mut.Unlock()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
				return err
			}
		case obj := <-mb.iNpvnChan:
			drop, err := mb.shouldDrop(obj.ctx, obj.cancel, obj.errC)
			if err != nil {
				return err
			}
			if drop {
				continue
			}
			mut.Lock()
			err = mb.handleStateUpdateErrors(obj.ctx, mb.shandlers.AddPreVoteNil(obj.msg), obj.errC)
			mut.Unlock()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
				return err
			}
		case obj := <-mb.iNpcChan:
			drop, err := mb.shouldDrop(obj.ctx, obj.cancel, obj.errC)
			if err != nil {
				return err
			}
			if drop {
				continue
			}
			mut.Lock()
			err = mb.handleStateUpdateErrors(obj.ctx, mb.shandlers.AddPreCommit(obj.msg), obj.errC)
			mut.Unlock()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
				return err
			}
		case obj := <-mb.iNpcnChan:
			drop, err := mb.shouldDrop(obj.ctx, obj.cancel, obj.errC)
			if err != nil {
				return err
			}
			if drop {
				continue
			}
			mut.Lock()
			err = mb.handleStateUpdateErrors(obj.ctx, mb.shandlers.AddPreCommitNil(obj.msg), obj.errC)
			mut.Unlock()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
				return err
			}
		case obj := <-mb.iNnrChan:
			drop, err := mb.shouldDrop(obj.ctx, obj.cancel, obj.errC)
			if err != nil {
				return err
			}
			if drop {
				continue
			}
			mut.Lock()
			err = mb.handleStateUpdateErrors(obj.ctx, mb.shandlers.AddNextRound(obj.msg), obj.errC)
			mut.Unlock()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
				return err
			}
		case obj := <-mb.iNnhChan:
			drop, err := mb.shouldDrop(obj.ctx, obj.cancel, obj.errC)
			if err != nil {
				return err
			}
			if drop {
				continue
			}
			mut.Lock()
			err = mb.handleStateUpdateErrors(obj.ctx, mb.shandlers.AddNextHeight(obj.msg), obj.errC)
			mut.Unlock()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
				return err
			}
		case obj := <-mb.iNbhChan:
			if mb.isDone(obj.ctx, obj.errC) {
				continue
			}
			mut.Lock()
			err := mb.handleStateUpdateErrors(obj.ctx, mb.shandlers.AddBlockHeader(obj.msg), obj.errC)
			mut.Unlock()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
				return err
			}
		case obj := <-mb.iNtxChan:
			drop, err := mb.shouldDrop(obj.ctx, obj.cancel, obj.errC)
			if err != nil {
				return err
			}
			if drop {
				continue
			}
			mut.Lock()
			err = mb.handleStateUpdateErrors(obj.ctx, mb.updateTxsFromGossip(obj.msg), obj.errC)
			mut.Unlock()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
				return err
			}
		}
	}
}

func (mb *Handlers) updateTxsFromGossip(tx []byte) error {
	err := mb.database.Update(func(txn *badger.Txn) error {
		os, err := mb.database.GetOwnState(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		chainID := os.SyncToBH.BClaims.ChainID
		height := os.SyncToBH.BClaims.Height + 1
		tx, err := mb.app.UnmarshalTx(tx)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		err = mb.app.PendingTxAdd(txn, chainID, height, []interfaces.Transaction{tx})
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// UpdateBlocksFromGossip stores blocks only from a remote node
func (mb *Handlers) UpdateBlocksFromGossip() error {
	for {
		select {
		case obj := <-mb.iNbhChan:
			return mb.handleStateUpdateErrors(obj.ctx, mb.shandlers.AddBlockHeader(obj.msg), obj.errC)
		case <-mb.ctx.Done():
			return nil
		}
	}
}

func (mb *Handlers) preventGossip(ctx context.Context, rawmsg []byte, isTx bool, isBh bool) {
	// peerAddr, ok := peer.FromContext(ctx)
	// if !ok {
	// 	return
	// }
	// nodeAddr, ok := peerAddr.Addr.(interfaces.NodeAddr)
	// if !ok {
	// 	return
	// }
	// if isTx {
	// 	txIf, err := mb.app.UnmarshalTx(rawmsg)
	// 	if err != nil {
	// 		return
	// 	}
	// hsh, err := txIf.TxHash()
	// if err != nil {
	// 	return
	// }
	//mb.peerSub.PreventGossipTx(nodeAddr, hsh)
	// return
	// }
	// if isBh {
	// 	bh := objs.BlockHeader{}
	// 	err := bh.UnmarshalBinary(rawmsg)
	// 	if err != nil {
	// 		return
	// 	}
	//hsh := utils.MarshalUint32(bh.BClaims.Height)
	//mb.peerSub.PreventGossipConsensus(nodeAddr, hsh)
	// return
	// }
	//mb.peerSub.PreventGossipConsensus(nodeAddr, rawmsg)
}

func (mb *Handlers) setupHandler(pctx context.Context) (context.Context, func(), chan error, error) {
	ctx, cf := context.WithTimeout(pctx, constants.SrvrMsgTimeout)
	select {
	case <-mb.ctx.Done():
		cf()
		return nil, func() {}, nil, errorz.ErrClosing
	default:
		mb.wg.Add(1)
	}
	return ctx, cf, make(chan error, 4), nil
}

// HandleP2PGossipTransaction adds a transaction to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipTransaction(pctx context.Context, msg *pb.GossipTransactionMessage) (*pb.GossipTransactionAck, error) {
	ctx, cf, eC, err := mb.setupHandler(pctx)
	if err != nil {
		return nil, err
	}
	defer mb.wg.Done()
	defer cf()
	rawmsg := msg.Transaction
	// // mb.preventGossip(ctx, rawmsg, true, false)
	mobj := &transactionMsg{ctx, cf, rawmsg, eC}
	select {
	case mb.iNtxChan <- mobj:
		return &pb.GossipTransactionAck{}, <-mobj.errC
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-mb.ctx.Done():
		return nil, errorz.ErrClosing
	}
}

// HandleP2PGossipProposal adds a proposal to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipProposal(pctx context.Context, msg *pb.GossipProposalMessage) (*pb.GossipProposalAck, error) {
	ctx, cf, eC, err := mb.setupHandler(pctx)
	if err != nil {
		return nil, err
	}
	defer mb.wg.Done()
	defer cf()
	rawmsg := msg.Proposal
	obj := &objs.Proposal{}
	err = obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipProposalAck{}, err
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipProposalAck{}, err
	}
	// mb.preventGossip(ctx, rawmsg, false, false)
	mobj := &proposalMsg{ctx, cf, obj, eC}
	select {
	case mb.iNpChan <- mobj:
		return &pb.GossipProposalAck{}, <-mobj.errC
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-mb.ctx.Done():
		return nil, errorz.ErrClosing
	}
}

// HandleP2PGossipPreVote adds a preVote to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipPreVote(pctx context.Context, msg *pb.GossipPreVoteMessage) (*pb.GossipPreVoteAck, error) {
	ctx, cf, eC, err := mb.setupHandler(pctx)
	if err != nil {
		return nil, err
	}
	defer mb.wg.Done()
	defer cf()
	rawmsg := msg.PreVote
	obj := &objs.PreVote{}
	err = obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipPreVoteAck{}, err
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipPreVoteAck{}, err
	}
	// mb.preventGossip(ctx, rawmsg, false, false)
	mobj := &preVoteMsg{ctx, cf, obj, eC}
	select {
	case mb.iNpvChan <- mobj:
		return &pb.GossipPreVoteAck{}, <-mobj.errC
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-mb.ctx.Done():
		return nil, errorz.ErrClosing
	}
}

// HandleP2PGossipPreVoteNil adds a preVoteNil to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipPreVoteNil(pctx context.Context, msg *pb.GossipPreVoteNilMessage) (*pb.GossipPreVoteNilAck, error) {
	ctx, cf, eC, err := mb.setupHandler(pctx)
	if err != nil {
		return nil, err
	}
	defer mb.wg.Done()
	defer cf()
	rawmsg := msg.PreVoteNil
	obj := &objs.PreVoteNil{}
	err = obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipPreVoteNilAck{}, err
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipPreVoteNilAck{}, err
	}
	// mb.preventGossip(ctx, rawmsg, false, false)
	mobj := &preVoteNilMsg{ctx, cf, obj, eC}
	select {
	case mb.iNpvnChan <- mobj:
		return &pb.GossipPreVoteNilAck{}, <-mobj.errC
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-mb.ctx.Done():
		return nil, errorz.ErrClosing
	}
}

// HandleP2PGossipPreCommit adds a preCommit to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipPreCommit(pctx context.Context, msg *pb.GossipPreCommitMessage) (*pb.GossipPreCommitAck, error) {
	ctx, cf, eC, err := mb.setupHandler(pctx)
	if err != nil {
		return nil, err
	}
	defer mb.wg.Done()
	defer cf()
	rawmsg := msg.PreCommit
	obj := &objs.PreCommit{}
	err = obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipPreCommitAck{}, err
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipPreCommitAck{}, err
	}
	// mb.preventGossip(ctx, rawmsg, false, false)
	mobj := &preCommitMsg{ctx, cf, obj, eC}
	select {
	case mb.iNpcChan <- mobj:
		return &pb.GossipPreCommitAck{}, <-mobj.errC
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-mb.ctx.Done():
		return nil, errorz.ErrClosing
	}
}

// HandleP2PGossipPreCommitNil adds a preCommitNil to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipPreCommitNil(pctx context.Context, msg *pb.GossipPreCommitNilMessage) (*pb.GossipPreCommitNilAck, error) {
	ctx, cf, eC, err := mb.setupHandler(pctx)
	if err != nil {
		return nil, err
	}
	defer mb.wg.Done()
	defer cf()
	rawmsg := msg.PreCommitNil
	obj := &objs.PreCommitNil{}
	err = obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipPreCommitNilAck{}, err
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipPreCommitNilAck{}, err
	}
	// mb.preventGossip(ctx, rawmsg, false, false)
	mobj := &preCommitNilMsg{ctx, cf, obj, eC}
	select {
	case mb.iNpcnChan <- mobj:
		return &pb.GossipPreCommitNilAck{}, <-mobj.errC
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-mb.ctx.Done():
		return nil, errorz.ErrClosing
	}
}

// HandleP2PGossipNextRound adds a nextRound to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipNextRound(pctx context.Context, msg *pb.GossipNextRoundMessage) (*pb.GossipNextRoundAck, error) {
	ctx, cf, eC, err := mb.setupHandler(pctx)
	if err != nil {
		return nil, err
	}
	defer mb.wg.Done()
	defer cf()
	rawmsg := msg.NextRound
	obj := &objs.NextRound{}
	err = obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipNextRoundAck{}, err
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipNextRoundAck{}, err
	}
	// mb.preventGossip(ctx, rawmsg, false, false)
	mobj := &nextRoundMsg{ctx, cf, obj, eC}
	select {
	case mb.iNnrChan <- mobj:
		return &pb.GossipNextRoundAck{}, <-mobj.errC
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-mb.ctx.Done():
		return nil, errorz.ErrClosing
	}
}

// HandleP2PGossipNextHeight adds a nextHeight to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipNextHeight(pctx context.Context, msg *pb.GossipNextHeightMessage) (*pb.GossipNextHeightAck, error) {
	ctx, cf, eC, err := mb.setupHandler(pctx)
	if err != nil {
		return nil, err
	}
	defer mb.wg.Done()
	defer cf()
	rawmsg := msg.NextHeight
	obj := &objs.NextHeight{}
	err = obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipNextHeightAck{}, err
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipNextHeightAck{}, err
	}
	// mb.preventGossip(ctx, rawmsg, false, false)
	mobj := &nextHeightMsg{ctx, cf, obj, eC}
	select {
	case mb.iNnhChan <- mobj:
		return &pb.GossipNextHeightAck{}, <-mobj.errC
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-mb.ctx.Done():
		return nil, errorz.ErrClosing
	}
}

// HandleP2PGossipBlockHeader adds a nextHeight to the database
// This method should be invoked when a remote peer
// sends this type of object to the local node over
// the gossip protocol
func (mb *Handlers) HandleP2PGossipBlockHeader(pctx context.Context, msg *pb.GossipBlockHeaderMessage) (*pb.GossipBlockHeaderAck, error) {
	ctx, cf, eC, err := mb.setupHandler(pctx)
	if err != nil {
		return nil, err
	}
	defer mb.wg.Done()
	defer cf()
	rawmsg := msg.BlockHeader
	obj := &objs.BlockHeader{}
	err = obj.UnmarshalBinary(rawmsg)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipBlockHeaderAck{}, err
	}
	if err := mb.shandlers.PreValidate(obj); err != nil {
		utils.DebugTrace(mb.logger, err)
		return &pb.GossipBlockHeaderAck{}, err
	}
	// mb.preventGossip(ctx, rawmsg, false, true)
	mobj := &blockHeaderMsg{ctx, cf, obj, eC}
	select {
	case mb.iNbhChan <- mobj:
		return &pb.GossipBlockHeaderAck{}, <-mobj.errC
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-mb.ctx.Done():
		return nil, errorz.ErrClosing
	}
}
