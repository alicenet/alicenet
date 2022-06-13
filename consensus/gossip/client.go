package gossip

import (
	"context"
	"sync"
	"time"

	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/lstate"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/dynamics"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/middleware"
	pb "github.com/MadBase/MadNet/proto"
	"github.com/MadBase/MadNet/utils"

	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type appClient interface {
	GetTxsForGossip(txnState *badger.Txn, currentHeight uint32) ([]interfaces.Transaction, error)
	UnmarshalTx([]byte) (interfaces.Transaction, error)
}

type mutexBool struct {
	sync.RWMutex
	value bool
}

func (mb *mutexBool) Set(v bool) {
	mb.Lock()
	defer mb.Unlock()
	mb.value = v
}
func (mb *mutexBool) Get() bool {
	mb.RLock()
	defer mb.RUnlock()
	return mb.value
}

type mutexUint32 struct {
	sync.RWMutex
	value uint32
}

func (mb *mutexUint32) Set(v uint32) {
	mb.Lock()
	defer mb.Unlock()
	mb.value = v
}
func (mb *mutexUint32) Get() uint32 {
	mb.RLock()
	defer mb.RUnlock()
	return mb.value
}

// Client handles outbound gossip
type Client struct {
	sync.Mutex
	wg       sync.WaitGroup
	client   pb.P2PClient
	database *db.Database
	sstore   *lstate.Store

	ctx       context.Context
	cancelCtx func()

	gossipTimeout time.Duration
	logger        *logrus.Logger
	lastHeight    uint32
	lastRound     uint32
	app           appClient
	storage       dynamics.StorageGetter

	inSync      *mutexBool
	isValidator *mutexBool
}

// Init sets ups all subscriptions. This MUST be run at least once.
// It has no effect if run more than once.
func (mb *Client) Init(database *db.Database, client pb.P2PClient, app appClient, storage dynamics.StorageGetter) {
	background := context.Background()
	ctx, cf := context.WithCancel(background)
	mb.logger = logging.GetLogger(constants.LoggerGossipBus)
	mb.cancelCtx = cf
	mb.ctx = ctx
	mb.wg = sync.WaitGroup{}
	mb.database = database
	mb.client = client
	mb.app = app
	mb.sstore = &lstate.Store{}
	mb.inSync = &mutexBool{}
	mb.isValidator = &mutexBool{}
	mb.storage = storage
	mb.sstore.Init(database)
	mb.gossipTimeout = constants.MsgTimeout
}

// Close will stop the gossip bus such that it can not be started again
func (mb *Client) Close() {
	mb.cancelCtx()
	mb.wg.Wait()
}

// Done blocks until the service has an exit
func (mb *Client) Done() <-chan struct{} {
	return mb.ctx.Done()
}

// Start will start the service
func (mb *Client) Start() error {
	mb.database.SubscribeBroadcastTransaction(
		mb.ctx,
		func(v []byte) error {
			opts := []grpc.CallOption{
				middleware.WithNoBlocking(),
			}
			go mb.gossipTransaction(v, opts...)
			return nil
		},
	)

	pgfn := func(v []byte) error {
		opts := []grpc.CallOption{
			// middleware.WithNoBlocking(),
		}
		go mb.gossipProposal(v, opts...)
		return nil
	}
	mb.database.SubscribeBroadcastProposal(mb.ctx, pgfn)

	pvgfn := func(v []byte) error {
		opts := []grpc.CallOption{
			// middleware.WithNoBlocking(),
		}
		go mb.gossipPreVote(v, opts...)
		return nil
	}
	mb.database.SubscribeBroadcastPreVote(mb.ctx, pvgfn)

	pvngfn := func(v []byte) error {
		opts := []grpc.CallOption{
			// middleware.WithNoBlocking(),
		}
		go mb.gossipPreVoteNil(v, opts...)
		return nil
	}
	mb.database.SubscribeBroadcastPreVoteNil(mb.ctx, pvngfn)

	pcgfn := func(v []byte) error {
		opts := []grpc.CallOption{
			// middleware.WithNoBlocking(),
		}
		go mb.gossipPreCommit(v, opts...)
		return nil
	}
	mb.database.SubscribeBroadcastPreCommit(mb.ctx, pcgfn)

	pcngfn := func(v []byte) error {
		opts := []grpc.CallOption{
			// middleware.WithNoBlocking(),
		}
		go mb.gossipPreCommitNil(v, opts...)
		return nil
	}
	mb.database.SubscribeBroadcastPreCommitNil(mb.ctx, pcngfn)

	nrgfn := func(v []byte) error {
		opts := []grpc.CallOption{
			// middleware.WithNoBlocking(),
		}
		go mb.gossipNextRound(v, opts...)
		return nil
	}
	mb.database.SubscribeBroadcastNextRound(mb.ctx, nrgfn)

	nhgfn := func(v []byte) error {
		opts := []grpc.CallOption{
			// middleware.WithNoBlocking(),
		}
		go mb.gossipNextHeight(v, opts...)
		return nil
	}
	mb.database.SubscribeBroadcastNextHeight(mb.ctx, nhgfn)

	bhgfn := func(v []byte) error {
		opts := []grpc.CallOption{
			// middleware.WithNoBlocking(),
		}
		go mb.gossipBlockHeader(v, opts...)
		return nil
	}
	mb.database.SubscribeBroadcastBlockHeader(mb.ctx, bhgfn)
	return nil
}

func (mb *Client) getReGossipTxs(txn *badger.Txn, height uint32) ([][]byte, error) {
	txns, err := mb.app.GetTxsForGossip(txn, height)
	if err != nil {
		return nil, err
	}
	txout := [][]byte{}
	for i := 0; i < len(txns); i++ {
		tx := txns[i]
		txb, err := tx.MarshalBinary()
		if err != nil {
			return nil, err
		}
		txout = append(txout, txb)
	}
	return txout, nil
}

// ReGossip performs the reGossip logic
func (mb *Client) ReGossip() error {
	var isValidator bool
	var height uint32
	var round uint32
	ok := func() bool {
		mb.Lock()
		defer mb.Unlock()
		select {
		case <-mb.Done():
			return false
		default:
			return true
		}
	}()
	if !ok {
		return nil
	}
	txs := [][]byte{}
	var bh *objs.BlockHeader
	var p *objs.Proposal
	var pv *objs.PreVote
	var pvn *objs.PreVoteNil
	var pc *objs.PreCommit
	var pcn *objs.PreCommitNil
	var nr *objs.NextRound
	var nh *objs.NextHeight
	err := mb.database.View(func(txn *badger.Txn) error {
		var err error
		var isSync bool
		isValidator, isSync, _, height, round, err = mb.sstore.GetDropData(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		mb.isValidator.Set(isValidator)
		mb.inSync.Set(isSync)
		if !isValidator {
			txs, err = mb.getReGossipTxs(txn, height)
			if err != nil {
				utils.DebugTrace(mb.logger, err)
				return err
			}
		}
		bh, err = mb.sstore.GetSyncToBH(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		p, pv, pvn, pc, pcn, nr, nh, err = mb.sstore.GetGossipValues(txn)
		if err != nil {
			mb.logger.Error(err)
			return err
		}
		return err
	})
	if err != nil {
		mb.logger.Error(err)
		return err
	}

	if mb.lastHeight != height {
		mb.lastRound = round
		mb.lastHeight = height
	}

	if mb.lastHeight == height {
		if mb.lastRound != round {
			mb.lastRound = round
		}
	}

	if bh != nil {
		bhBytes, err := bh.MarshalBinary()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
		} else {
			opts := []grpc.CallOption{
				middleware.WithNoBlocking(),
			}
			go mb.gossipBlockHeader(bhBytes, opts...)
		}
	}

	if p != nil {
		skip := false
		if mb.lastHeight > p.PClaims.RCert.RClaims.Height {
			skip = true
		}
		if mb.lastHeight == p.PClaims.RCert.RClaims.Height {
			if mb.lastRound > p.PClaims.RCert.RClaims.Round {
				skip = true
			}
		}
		if !skip {
			b, err := p.MarshalBinary()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
			} else {
				opts := []grpc.CallOption{
					middleware.WithNoBlocking(),
				}
				mb.logger.Debugf("GossipProposal: H:%v R:%v LH:%v LR:%v", p.PClaims.BClaims.Height, p.PClaims.RCert.RClaims.Round, mb.lastHeight, mb.lastRound)
				go mb.gossipProposal(b, opts...)
			}
		}
	}
	if pv != nil {
		skip := false
		if mb.lastHeight > pv.Proposal.PClaims.RCert.RClaims.Height {
			skip = true
		}
		if mb.lastHeight == pv.Proposal.PClaims.RCert.RClaims.Height {
			if mb.lastRound > pv.Proposal.PClaims.RCert.RClaims.Round {
				skip = true
			}
		}
		if pc != nil {
			if pc.Proposal.PClaims.RCert.RClaims.Height == pv.Proposal.PClaims.RCert.RClaims.Height {
				if pc.Proposal.PClaims.RCert.RClaims.Round == pv.Proposal.PClaims.RCert.RClaims.Round {
					skip = true
				}
			}
		}
		if !skip {
			b, err := pv.MarshalBinary()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
			} else {
				opts := []grpc.CallOption{
					middleware.WithNoBlocking(),
				}
				mb.logger.Debugf("GossipPreVote: H:%v R:%v LH:%v LR:%v", pv.Proposal.PClaims.BClaims.Height, pv.Proposal.PClaims.RCert.RClaims.Round, mb.lastHeight, mb.lastRound)
				go mb.gossipPreVote(b, opts...)
			}
		}
	}
	if pvn != nil {
		skip := false
		if mb.lastHeight > pvn.RCert.RClaims.Height {
			skip = true
		}
		if mb.lastHeight == pvn.RCert.RClaims.Height {
			if mb.lastRound > pvn.RCert.RClaims.Round {
				skip = true
			}
		}
		if !skip {
			b, err := pvn.MarshalBinary()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
			} else {
				opts := []grpc.CallOption{
					middleware.WithNoBlocking(),
				}
				mb.logger.Debugf("GossipPreVoteNil: H:%v R:%v LH:%v LR:%v", pvn.RCert.RClaims.Height, pvn.RCert.RClaims.Round, mb.lastHeight, mb.lastRound)
				go mb.gossipPreVoteNil(b, opts...)
			}
		}
	}
	if pc != nil {
		skip := false
		if mb.lastHeight > pc.Proposal.PClaims.RCert.RClaims.Height {
			skip = true
		}
		if pcn != nil {
			if pcn.RCert.RClaims.Height == pc.Proposal.PClaims.RCert.RClaims.Height {
				if pcn.RCert.RClaims.Round > pc.Proposal.PClaims.RCert.RClaims.Round {
					skip = true
				}
			}
		}
		if !skip {
			b, err := pc.MarshalBinary()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
			} else {
				opts := []grpc.CallOption{
					middleware.WithNoBlocking(),
				}
				mb.logger.Debugf("GossipPreCommit: H:%v R:%v LH:%v LR:%v", pc.Proposal.PClaims.BClaims.Height, pc.Proposal.PClaims.RCert.RClaims.Round, mb.lastHeight, mb.lastRound)
				go mb.gossipPreCommit(b, opts...)
			}
		}
	}
	if pcn != nil {
		skip := false
		if mb.lastHeight > pcn.RCert.RClaims.Height {
			skip = true
		}
		if mb.lastHeight == pcn.RCert.RClaims.Height {
			if mb.lastRound > pcn.RCert.RClaims.Round {
				skip = true
			}
		}
		if !skip {
			b, err := pcn.MarshalBinary()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
			} else {
				opts := []grpc.CallOption{
					middleware.WithNoBlocking(),
				}
				mb.logger.Debugf("GossipPreCommitNil: H:%v R:%v LH:%v LR:%v", pcn.RCert.RClaims.Height, pcn.RCert.RClaims.Round, mb.lastHeight, mb.lastRound)
				go mb.gossipPreCommitNil(b, opts...)
			}
		}
	}
	if nr != nil {
		skip := false
		if mb.lastHeight > nr.NRClaims.RCert.RClaims.Height {
			skip = true
		}
		if mb.lastHeight == nr.NRClaims.RCert.RClaims.Height {
			if mb.lastRound > nr.NRClaims.RCert.RClaims.Round+1 {
				skip = true
			}
		}
		if !skip {
			b, err := nr.MarshalBinary()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
			} else {
				opts := []grpc.CallOption{
					middleware.WithNoBlocking(),
				}
				mb.logger.Debugf("GossipNextRound: H:%v R:%v LH:%v LR:%v", nr.NRClaims.RCert.RClaims.Height, nr.NRClaims.RCert.RClaims.Round, mb.lastHeight, mb.lastRound)
				go mb.gossipNextRound(b, opts...)
			}
		}
	}
	if nh != nil {
		skip := false
		if mb.lastHeight > nh.NHClaims.Proposal.PClaims.RCert.RClaims.Height {
			skip = true
		}
		if !skip {
			b, err := nh.MarshalBinary()
			if err != nil {
				utils.DebugTrace(mb.logger, err)
			} else {
				opts := []grpc.CallOption{
					middleware.WithNoBlocking(),
				}
				mb.logger.Debugf("GossipNextHeight: H:%v R:%v LH:%v LR:%v", nh.NHClaims.Proposal.PClaims.RCert.RClaims.Height, nh.NHClaims.Proposal.PClaims.RCert.RClaims.Round, mb.lastHeight, mb.lastRound)
				go mb.gossipNextHeight(b, opts...)
			}
		}
	}

	if isValidator {
		for i := 0; i < len(txs); i++ {
			tx := txs[i]
			go mb.gossipTransaction(tx)
		}
	}

	return nil
}

func (mb *Client) gossipTransaction(transaction []byte, opts ...grpc.CallOption) {
	if len(transaction) == 0 {
		return
	}
	if !mb.inSync.Get() {
		return
	}
	if mb.isValidator.Get() {
		return
	}
	mb.logger.Debug("gossipTransaction")
	msg := &pb.GossipTransactionMessage{
		Transaction: utils.CopySlice(transaction),
	}
	opts = append(opts, []grpc.CallOption{
		middleware.WithNoBlocking(),
	}...)
	_, err := mb.client.GossipTransaction(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) gossipProposal(proposal []byte, opts ...grpc.CallOption) {
	if !mb.inSync.Get() {
		return
	}
	mb.logger.Debug("gossipProposal")
	msg := &pb.GossipProposalMessage{
		Proposal: proposal,
	}
	_, err := mb.client.GossipProposal(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) gossipPreVote(preVote []byte, opts ...grpc.CallOption) {
	if !mb.inSync.Get() {
		return
	}
	mb.logger.Debug("gossipPreVote")
	msg := &pb.GossipPreVoteMessage{
		PreVote: preVote,
	}
	_, err := mb.client.GossipPreVote(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) gossipPreVoteNil(preVoteNil []byte, opts ...grpc.CallOption) {
	if !mb.inSync.Get() {
		return
	}
	mb.logger.Debug("gossipPreVoteNil")
	msg := &pb.GossipPreVoteNilMessage{
		PreVoteNil: preVoteNil,
	}
	_, err := mb.client.GossipPreVoteNil(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) gossipPreCommit(preCommit []byte, opts ...grpc.CallOption) {
	if !mb.inSync.Get() {
		return
	}
	mb.logger.Debug("gossipPreCommit")
	msg := &pb.GossipPreCommitMessage{
		PreCommit: preCommit,
	}
	_, err := mb.client.GossipPreCommit(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) gossipPreCommitNil(preCommitNil []byte, opts ...grpc.CallOption) {
	if !mb.inSync.Get() {
		return
	}
	mb.logger.Debug("gossipPreCommitNil")
	msg := &pb.GossipPreCommitNilMessage{
		PreCommitNil: preCommitNil,
	}
	_, err := mb.client.GossipPreCommitNil(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) gossipNextRound(nextRound []byte, opts ...grpc.CallOption) {
	if !mb.inSync.Get() {
		return
	}
	mb.logger.Debug("gossipNextRound")
	msg := &pb.GossipNextRoundMessage{
		NextRound: nextRound,
	}
	_, err := mb.client.GossipNextRound(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) gossipNextHeight(nextHeight []byte, opts ...grpc.CallOption) {
	if !mb.inSync.Get() {
		return
	}
	mb.logger.Debug("gossipNextHeight")
	msg := &pb.GossipNextHeightMessage{
		NextHeight: nextHeight,
	}
	_, err := mb.client.GossipNextHeight(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) gossipBlockHeader(blockHeader []byte, opts ...grpc.CallOption) {
	if !mb.inSync.Get() {
		return
	}
	mb.logger.Debug("gossipBlockHeader")
	msg := &pb.GossipBlockHeaderMessage{
		BlockHeader: blockHeader,
	}
	_, err := mb.client.GossipBlockHeader(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}
