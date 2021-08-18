package gossip

import (
	"context"
	"sync"
	"time"

	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/lstate"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/logging"
	pb "github.com/MadBase/MadNet/proto"
	"github.com/MadBase/MadNet/utils"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"

	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const maxRetryCount = 12
const backOffAmount = 1
const backOffJitter = float64(.1)

type appClient interface {
	GetTxsForGossip(txnState *badger.Txn, currentHeight uint32) ([]interfaces.Transaction, error)
	UnmarshalTx([]byte) (interfaces.Transaction, error)
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
}

// Init sets ups all subscriptions. This MUST be run at least once.
// It has no effect if run more than once.
func (mb *Client) Init(database *db.Database, client pb.P2PClient, app appClient) error {
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
			go mb.gossipTransaction(v)
			return nil
		},
	)

	pgfn := func(v []byte) error {
		go mb.gossipProposal(v)
		return nil
	}
	mb.database.SubscribeBroadcastProposal(mb.ctx, pgfn)

	pvgfn := func(v []byte) error {
		go mb.gossipPreVote(v)
		return nil
	}
	mb.database.SubscribeBroadcastPreVote(mb.ctx, pvgfn)

	pvngfn := func(v []byte) error {
		go mb.gossipPreVoteNil(v)
		return nil
	}
	mb.database.SubscribeBroadcastPreVoteNil(mb.ctx, pvngfn)

	pcgfn := func(v []byte) error {
		go mb.gossipPreCommit(v)
		return nil
	}
	mb.database.SubscribeBroadcastPreCommit(mb.ctx, pcgfn)

	pcngfn := func(v []byte) error {
		go mb.gossipPreCommitNil(v)
		return nil
	}
	mb.database.SubscribeBroadcastPreCommitNil(mb.ctx, pcngfn)

	nrgfn := func(v []byte) error {
		go mb.gossipNextRound(v)
		return nil
	}
	mb.database.SubscribeBroadcastNextRound(mb.ctx, nrgfn)

	nhgfn := func(v []byte) error {
		go mb.gossipNextHeight(v)
		return nil
	}
	mb.database.SubscribeBroadcastNextHeight(mb.ctx, nhgfn)

	bhgfn := func(v []byte) error {
		go mb.gossipBlockHeader(v)
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
		b, err := p.MarshalBinary()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		go mb.gossipProposal(b)
	}
	if pv != nil {
		b, err := pv.MarshalBinary()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		go mb.gossipPreVote(b)
	}
	if pvn != nil {
		b, err := pvn.MarshalBinary()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		go mb.gossipPreVoteNil(b)
	}
	if pc != nil {
		b, err := pc.MarshalBinary()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		go mb.gossipPreCommit(b)
	}
	if pcn != nil {
		b, err := pcn.MarshalBinary()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		go mb.gossipPreCommitNil(b)
	}
	if nr != nil {
		b, err := nr.MarshalBinary()
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		go mb.gossipNextRound(b)
	}
	if nh != nil {
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
		Transaction: utils.CopySlice(transaction),
	}
	opts := []grpc.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
		grpc_retry.WithMax(maxRetryCount),
	}
	_, err := mb.client.GossipTransaction(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) gossipProposal(proposal []byte) {
	msg := &pb.GossipProposalMessage{
		Proposal: proposal,
	}
	opts := []grpc.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
		grpc_retry.WithMax(maxRetryCount),
	}
	_, err := mb.client.GossipProposal(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) inSync() bool {
	var isSync bool
	var err error
	mb.database.View(func(txn *badger.Txn) error {
		isSync, err = mb.sstore.IsSync(txn)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return false
	}
	return isSync
}

func (mb *Client) gossipPreVote(preVote []byte) {
	if !mb.inSync() {
		return
	}
	msg := &pb.GossipPreVoteMessage{
		PreVote: preVote,
	}
	opts := []grpc.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
		grpc_retry.WithMax(maxRetryCount),
	}
	_, err := mb.client.GossipPreVote(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) gossipPreVoteNil(preVoteNil []byte) {
	if !mb.inSync() {
		return
	}
	msg := &pb.GossipPreVoteNilMessage{
		PreVoteNil: preVoteNil,
	}
	opts := []grpc.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
		grpc_retry.WithMax(maxRetryCount),
	}
	_, err := mb.client.GossipPreVoteNil(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) gossipPreCommit(preCommit []byte) {
	if !mb.inSync() {
		return
	}
	msg := &pb.GossipPreCommitMessage{
		PreCommit: preCommit,
	}
	opts := []grpc.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
		grpc_retry.WithMax(maxRetryCount),
	}
	_, err := mb.client.GossipPreCommit(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) gossipPreCommitNil(preCommitNil []byte) {
	if !mb.inSync() {
		return
	}
	msg := &pb.GossipPreCommitNilMessage{
		PreCommitNil: preCommitNil,
	}
	opts := []grpc.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
		grpc_retry.WithMax(maxRetryCount),
	}
	_, err := mb.client.GossipPreCommitNil(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) gossipNextRound(nextRound []byte) {
	if !mb.inSync() {
		return
	}
	msg := &pb.GossipNextRoundMessage{
		NextRound: nextRound,
	}
	opts := []grpc.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
		grpc_retry.WithMax(maxRetryCount),
	}
	_, err := mb.client.GossipNextRound(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) gossipNextHeight(nextHeight []byte) {
	if !mb.inSync() {
		return
	}
	msg := &pb.GossipNextHeightMessage{
		NextHeight: nextHeight,
	}
	opts := []grpc.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
		grpc_retry.WithMax(maxRetryCount),
	}
	_, err := mb.client.GossipNextHeight(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}

func (mb *Client) gossipBlockHeader(blockHeader []byte) {
	if !mb.inSync() {
		return
	}
	msg := &pb.GossipBlockHeaderMessage{
		BlockHeader: blockHeader,
	}
	opts := []grpc.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(backOffAmount*time.Millisecond, backOffJitter)),
		grpc_retry.WithMax(maxRetryCount),
	}
	_, err := mb.client.GossipBlockHeader(context.Background(), msg, opts...)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
	}
}
