package gossip

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/lstate"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/logging"
	pb "github.com/MadBase/MadNet/proto"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

type appClient interface {
	GetTxsForGossip(txnState *badger.Txn, currentHeight uint32) ([]interfaces.Transaction, error)
	UnmarshalTx([]byte) (interfaces.Transaction, error)
}

// Client handles outbound gossip
type Client struct {
	sync.Mutex
	wg       sync.WaitGroup
	peerSub  interfaces.PeerSubscription
	database *db.Database
	sstore   *lstate.Store

	ctx       context.Context
	cancelCtx func()

	gossipTimeout time.Duration
	logger        *logrus.Logger
	lastHeight    uint32
	lastRound     uint32
	app           appClient
}

// Init sets ups all subscriptions. This MUST be run at least once.
// It has no effect if run more than once.
func (mb *Client) Init(database *db.Database, peerSub interfaces.PeerSubscription, app appClient) error {
	background := context.Background()
	ctx, cf := context.WithCancel(background)
	mb.logger = logging.GetLogger(constants.LoggerGossipBus)
	mb.cancelCtx = cf
	mb.ctx = ctx
	mb.wg = sync.WaitGroup{}
	mb.database = database
	mb.peerSub = peerSub
	mb.app = app
	mb.sstore = &lstate.Store{}
	err := mb.sstore.Init(database)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return err
	}
	mb.gossipTimeout = constants.MsgTimeout
	return nil
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
			mb.gossipTransaction(v)
			return nil
		},
	)

	pgfn := func(v []byte) error {
		mb.gossipProposal(v)
		return nil
	}
	mb.database.SubscribeBroadcastProposal(mb.ctx, pgfn)

	pvgfn := func(v []byte) error {
		mb.gossipPreVote(v)
		return nil
	}
	mb.database.SubscribeBroadcastPreVote(mb.ctx, pvgfn)

	pvngfn := func(v []byte) error {
		mb.gossipPreVoteNil(v)
		return nil
	}
	mb.database.SubscribeBroadcastPreVoteNil(mb.ctx, pvngfn)

	pcgfn := func(v []byte) error {
		mb.gossipPreCommit(v)
		return nil
	}
	mb.database.SubscribeBroadcastPreCommit(mb.ctx, pcgfn)

	pcngfn := func(v []byte) error {
		mb.gossipPreCommitNil(v)
		return nil
	}
	mb.database.SubscribeBroadcastPreCommitNil(mb.ctx, pcngfn)

	nrgfn := func(v []byte) error {
		mb.gossipNextRound(v)
		return nil
	}
	mb.database.SubscribeBroadcastNextRound(mb.ctx, nrgfn)

	nhgfn := func(v []byte) error {
		mb.gossipNextHeight(v)
		return nil
	}
	mb.database.SubscribeBroadcastNextHeight(mb.ctx, nhgfn)

	bhgfn := func(v []byte) error {
		mb.gossipBlockHeader(v)
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
	var isValidator, isSync bool
	var height uint32
	var round uint32
	ok := func() bool {
		mb.Lock()
		defer mb.Unlock()
		select {
		case <-mb.Done():
			return false
		default:
			mb.wg.Add(1)
			defer mb.wg.Done()
			return true
		}
	}()
	if !ok {
		return nil
	}
	err := mb.database.View(func(txn *badger.Txn) error {
		var err error
		isValidator, isSync, _, height, round, err = mb.sstore.GetDropData(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		if !isSync {
			return nil
		}
		txs, err := mb.getReGossipTxs(txn, height)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		for i := 0; i < len(txs); i++ {
			tx := txs[i]
			go mb.gossipTransaction(tx)
		}

		if mb.lastHeight != height || mb.lastRound != round {
			mb.lastHeight = height
			mb.lastRound = round
			return nil
		}
		bh, err := mb.sstore.GetSyncToBH(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		bhBytes, err := bh.MarshalBinary()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		go mb.gossipBlockHeader(bhBytes)
		if !isValidator {
			return nil
		}
		return err
	})
	if err != nil {
		mb.logger.Error(err)
		return err
	}
	if !isSync {
		return nil
	}
	p, pv, pvn, pc, pcn, nr, nh, err := mb.sstore.GetGossipValues()
	if err != nil {
		mb.logger.Error(err)
		return err
	}
	if p != nil {
		err := mb.gossipPreValidate(p)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		b, err := p.MarshalBinary()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		go mb.gossipProposal(b)
	}
	if pv != nil {
		err := mb.gossipPreValidate(pv)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		b, err := pv.MarshalBinary()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		go mb.gossipPreVote(b)
	}
	if pvn != nil {
		err := mb.gossipPreValidate(pvn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		b, err := pvn.MarshalBinary()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		go mb.gossipPreVoteNil(b)
	}
	if pc != nil {
		err := mb.gossipPreValidate(pc)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		b, err := pc.MarshalBinary()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		go mb.gossipPreCommit(b)
	}
	if pcn != nil {
		err := mb.gossipPreValidate(pcn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		b, err := pcn.MarshalBinary()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		go mb.gossipPreCommitNil(b)
	}
	if nr != nil {
		err := mb.gossipPreValidate(nr)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		b, err := nr.MarshalBinary()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		go mb.gossipNextRound(b)
	}
	if nh != nil {
		err := mb.gossipPreValidate(nh)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		b, err := nh.MarshalBinary()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		go mb.gossipNextHeight(b)
	}
	return nil
}

func (mb *Client) gossipTransaction(transaction []byte) {
	msg := &pb.GossipTransactionMessage{
		Transaction: transaction,
	}
	txIf, err := mb.app.UnmarshalTx(transaction)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return
	}
	hsh, err := txIf.TxHash()
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return
	}
	fn := func(ctx context.Context, peer interfaces.PeerLease) error {
		client, err := peer.P2PClient()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		subCtx, cancelFunc := context.WithTimeout(ctx, constants.MsgTimeout)
		defer cancelFunc()
		_, err = client.GossipTransaction(subCtx, msg)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	}
	mb.peerSub.GossipTx(hsh, fn)
}

func (mb *Client) gossipProposal(proposal []byte) {
	var isSync bool
	var err error
	mb.database.View(func(txn *badger.Txn) error {
		isSync, err = mb.sstore.IsSync(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	})
	if !isSync {
		return
	}
	msg := &pb.GossipProposalMessage{
		Proposal: proposal,
	}
	fn := func(ctx context.Context, peer interfaces.PeerLease) error {
		client, err := peer.P2PClient()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}

		subCtx, cancelFunc := context.WithTimeout(ctx, constants.MsgTimeout)
		defer cancelFunc()
		_, err = client.GossipProposal(subCtx, msg)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	}
	mb.peerSub.GossipConsensus(proposal, fn)
}

func (mb *Client) gossipPreVote(preVote []byte) {
	var isSync bool
	var err error
	mb.database.View(func(txn *badger.Txn) error {
		isSync, err = mb.sstore.IsSync(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	})
	if !isSync {
		return
	}
	msg := &pb.GossipPreVoteMessage{
		PreVote: preVote,
	}
	fn := func(ctx context.Context, peer interfaces.PeerLease) error {
		client, err := peer.P2PClient()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}

		subCtx, cancelFunc := context.WithTimeout(ctx, constants.MsgTimeout)
		defer cancelFunc()
		_, err = client.GossipPreVote(subCtx, msg)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	}
	mb.peerSub.GossipConsensus(preVote, fn)
}

func (mb *Client) gossipPreVoteNil(preVoteNil []byte) {
	var isSync bool
	var err error
	mb.database.View(func(txn *badger.Txn) error {
		isSync, err = mb.sstore.IsSync(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	})
	if !isSync {
		return
	}
	msg := &pb.GossipPreVoteNilMessage{
		PreVoteNil: preVoteNil,
	}
	fn := func(ctx context.Context, peer interfaces.PeerLease) error {
		client, err := peer.P2PClient()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}

		subCtx, cancelFunc := context.WithTimeout(ctx, constants.MsgTimeout)
		defer cancelFunc()
		_, err = client.GossipPreVoteNil(subCtx, msg)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	}
	mb.peerSub.GossipConsensus(preVoteNil, fn)
}

func (mb *Client) gossipPreCommit(preCommit []byte) {
	var isSync bool
	var err error
	mb.database.View(func(txn *badger.Txn) error {
		isSync, err = mb.sstore.IsSync(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	})
	if !isSync {
		return
	}
	msg := &pb.GossipPreCommitMessage{
		PreCommit: preCommit,
	}
	fn := func(ctx context.Context, peer interfaces.PeerLease) error {
		client, err := peer.P2PClient()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}

		subCtx, cancelFunc := context.WithTimeout(ctx, constants.MsgTimeout)
		defer cancelFunc()
		_, err = client.GossipPreCommit(subCtx, msg)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	}
	mb.peerSub.GossipConsensus(preCommit, fn)
}

func (mb *Client) gossipPreCommitNil(preCommitNil []byte) {
	var isSync bool
	var err error
	mb.database.View(func(txn *badger.Txn) error {
		isSync, err = mb.sstore.IsSync(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	})
	if !isSync {
		return
	}
	msg := &pb.GossipPreCommitNilMessage{
		PreCommitNil: preCommitNil,
	}
	fn := func(ctx context.Context, peer interfaces.PeerLease) error {
		client, err := peer.P2PClient()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		subCtx, cancelFunc := context.WithTimeout(ctx, constants.MsgTimeout)
		defer cancelFunc()
		_, err = client.GossipPreCommitNil(subCtx, msg)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	}
	mb.peerSub.GossipConsensus(preCommitNil, fn)
}

func (mb *Client) gossipNextRound(nextRound []byte) {
	var isSync bool
	var err error
	mb.database.View(func(txn *badger.Txn) error {
		isSync, err = mb.sstore.IsSync(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	})
	if !isSync {
		return
	}
	msg := &pb.GossipNextRoundMessage{
		NextRound: nextRound,
	}
	fn := func(ctx context.Context, peer interfaces.PeerLease) error {
		client, err := peer.P2PClient()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		subCtx, cancelFunc := context.WithTimeout(ctx, constants.MsgTimeout)
		defer cancelFunc()
		_, err = client.GossipNextRound(subCtx, msg)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	}
	mb.peerSub.GossipConsensus(nextRound, fn)
}

func (mb *Client) gossipNextHeight(nextHeight []byte) {
	var isSync bool
	var err error
	mb.database.View(func(txn *badger.Txn) error {
		isSync, err = mb.sstore.IsSync(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	})
	if !isSync {
		return
	}
	msg := &pb.GossipNextHeightMessage{
		NextHeight: nextHeight,
	}
	fn := func(ctx context.Context, peer interfaces.PeerLease) error {
		client, err := peer.P2PClient()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		subCtx, cancelFunc := context.WithTimeout(ctx, constants.MsgTimeout)
		defer cancelFunc()
		_, err = client.GossipNextHeight(subCtx, msg)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	}
	mb.peerSub.GossipConsensus(nextHeight, fn)
}

func (mb *Client) gossipBlockHeader(blockHeader []byte) {
	var isSync bool
	var err error
	mb.database.View(func(txn *badger.Txn) error {
		isSync, err = mb.sstore.IsSync(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	})
	if !isSync {
		return
	}
	msg := &pb.GossipBlockHeaderMessage{
		BlockHeader: blockHeader,
	}
	bh := objs.BlockHeader{}
	err = bh.UnmarshalBinary(blockHeader)
	if err != nil {
		return
	}
	hsh := utils.MarshalUint32(bh.BClaims.Height)
	fn := func(ctx context.Context, peer interfaces.PeerLease) error {
		bh := &objs.BlockHeader{}
		err := bh.UnmarshalBinary(blockHeader)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		client, err := peer.P2PClient()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}

		subCtx, cancelFunc := context.WithTimeout(ctx, constants.MsgTimeout)
		defer cancelFunc()
		_, err = client.GossipBlockHeader(subCtx, msg)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		return nil
	}
	mb.peerSub.GossipConsensus(hsh, fn)
}

// gossipPreValidate ensures that no Height 1 object is ever gossiped
func (mb *Client) gossipPreValidate(any interface{}) error {
	var rcHeight uint32
	switch v := any.(type) {
	case *objs.Proposal:
		rc := v.PClaims.RCert.RClaims
		rcHeight = rc.Height
	case *objs.PreVote:
		rc := v.Proposal.PClaims.RCert.RClaims
		rcHeight = rc.Height
	case *objs.PreVoteNil:
		rc := v.RCert.RClaims
		rcHeight = rc.Height
	case *objs.PreCommit:
		rc := v.Proposal.PClaims.RCert.RClaims
		rcHeight = rc.Height
	case *objs.PreCommitNil:
		rc := v.RCert.RClaims
		rcHeight = rc.Height
	case *objs.NextRound:
		rc := v.NRClaims.RCert.RClaims
		rcHeight = rc.Height
	case *objs.NextHeight:
		rc := v.NHClaims.Proposal.PClaims.RCert.RClaims
		rcHeight = rc.Height
	default:
		panic(fmt.Sprintf("undefined type for getting RCert Height and Round: %T", v))
	}
	if rcHeight == 1 || rcHeight == 0 {
		// This should never be gossiped; Height 1 block already set
		// and Height 0 objects should not exist
		return errorz.ErrInvalid{}.New("No Height 1 objects should ever be gossiped")
	}
	return nil
}
