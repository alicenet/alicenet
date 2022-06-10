package peering

import (
	"context"
	"errors"
	"time"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/middleware"
	pb "github.com/MadBase/MadNet/proto"
	"github.com/MadBase/MadNet/utils"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type StatusRequest struct {
	ctx   context.Context
	req   *pb.StatusRequest
	opts  []grpc.CallOption
	rChan chan *StatusResponse
}

type StatusResponse struct {
	resp *pb.StatusResponse
	err  error
}

type GetBlockHeadersRequest struct {
	ctx   context.Context
	req   *pb.GetBlockHeadersRequest
	opts  []grpc.CallOption
	rChan chan *GetBlockHeadersResponse
}

type GetBlockHeadersResponse struct {
	resp *pb.GetBlockHeadersResponse
	err  error
}

type GetMinedTxsRequest struct {
	ctx   context.Context
	req   *pb.GetMinedTxsRequest
	opts  []grpc.CallOption
	rChan chan *GetMinedTxsResponse
}

type GetMinedTxsResponse struct {
	resp *pb.GetMinedTxsResponse
	err  error
}

type GetPendingTxsRequest struct {
	ctx   context.Context
	req   *pb.GetPendingTxsRequest
	opts  []grpc.CallOption
	rChan chan *GetPendingTxsResponse
}

type GetPendingTxsResponse struct {
	resp *pb.GetPendingTxsResponse
	err  error
}

type GetSnapShotNodeRequest struct {
	ctx   context.Context
	req   *pb.GetSnapShotNodeRequest
	opts  []grpc.CallOption
	rChan chan *GetSnapShotNodeResponse
}

type GetSnapShotNodeResponse struct {
	resp *pb.GetSnapShotNodeResponse
	err  error
}

type GetSnapShotStateDataRequest struct {
	ctx   context.Context
	req   *pb.GetSnapShotStateDataRequest
	opts  []grpc.CallOption
	rChan chan *GetSnapShotStateDataResponse
}

type GetSnapShotStateDataResponse struct {
	resp *pb.GetSnapShotStateDataResponse
	err  error
}

type GetSnapShotHdrNodeRequest struct {
	ctx   context.Context
	req   *pb.GetSnapShotHdrNodeRequest
	opts  []grpc.CallOption
	rChan chan *GetSnapShotHdrNodeResponse
}

type GetSnapShotHdrNodeResponse struct {
	resp *pb.GetSnapShotHdrNodeResponse
	err  error
}

type GossipTransactionMessage struct {
	ctx  context.Context
	req  *pb.GossipTransactionMessage
	opts []grpc.CallOption
}

//nolint:structcheck,unused
type GossipTransactionAck struct {
	resp *pb.GossipTransactionAck
	err  error
}

type GossipProposalMessage struct {
	ctx  context.Context
	req  *pb.GossipProposalMessage
	opts []grpc.CallOption
}

//nolint:structcheck,unused
type GossipProposalAck struct {
	resp *pb.GossipProposalAck
	err  error
}

type GossipPreVoteMessage struct {
	ctx  context.Context
	req  *pb.GossipPreVoteMessage
	opts []grpc.CallOption
}

//nolint:structcheck,unused
type GossipPreVoteAck struct {
	resp *pb.GossipPreVoteAck
	err  error
}

type GossipPreVoteNilMessage struct {
	ctx  context.Context
	req  *pb.GossipPreVoteNilMessage
	opts []grpc.CallOption
}

//nolint:structcheck,unused
type GossipPreVoteNilAck struct {
	resp *pb.GossipPreVoteNilAck
	err  error
}

type GossipPreCommitMessage struct {
	ctx  context.Context
	req  *pb.GossipPreCommitMessage
	opts []grpc.CallOption
}

//nolint:structcheck,unused
type GossipPreCommitAck struct {
	resp *pb.GossipPreCommitAck
	err  error
}

type GossipPreCommitNilMessage struct {
	ctx  context.Context
	req  *pb.GossipPreCommitNilMessage
	opts []grpc.CallOption
}

//nolint:structcheck,unused
type GossipPreCommitNilAck struct {
	resp *pb.GossipPreCommitNilAck
	err  error
}

type GossipNextRoundMessage struct {
	ctx  context.Context
	req  *pb.GossipNextRoundMessage
	opts []grpc.CallOption
}

//nolint:structcheck,unused
type GossipNextRoundAck struct {
	resp *pb.GossipNextRoundAck
	err  error
}

type GossipNextHeightMessage struct {
	ctx  context.Context
	req  *pb.GossipNextHeightMessage
	opts []grpc.CallOption
}

//nolint:structcheck,unused
type GossipNextHeightAck struct {
	resp *pb.GossipNextHeightAck
	err  error
}

type GossipBlockHeaderMessage struct {
	ctx  context.Context
	req  *pb.GossipBlockHeaderMessage
	opts []grpc.CallOption
}

//nolint:structcheck,unused
type GossipBlockHeaderAck struct {
	resp *pb.GossipBlockHeaderAck
	err  error
}

type GetPeersRequest struct {
	ctx   context.Context
	req   *pb.GetPeersRequest
	opts  []grpc.CallOption
	rChan chan *GetPeersResponse
}

type GetPeersResponse struct {
	resp *pb.GetPeersResponse
	err  error
}

// NewP2PBus binds a peer to the common work sharing and broadcast channels of
// the peer system.
func newP2PBus(client interfaces.P2PClient, reqChan <-chan interface{}, gossipChan <-chan interface{}, gossipTxChan <-chan interface{}, closeChan <-chan struct{}, reqCount int, gossipCount int, gossipTxCount int, cleanup func()) *P2PBus {
	p2p := &P2PBus{
		client:            client,
		reqChan:           reqChan,
		gossipChan:        gossipChan,
		gossipTxChan:      gossipTxChan,
		closeChan:         closeChan,
		maxRequestWorkers: reqCount,
		metricChan:        make(chan error, reqCount),
		workerKillChan:    make(chan struct{}),
		logger:            logging.GetLogger(constants.LoggerPeerMan),
		cleanup:           cleanup,
	}
	p2p.numWorkers++
	go p2p.reqWorker()
	p2p.numWorkers++
	go p2p.reqWorker()
	for i := 0; i < gossipCount; i++ {
		go p2p.gossipWorker()
	}
	for i := 0; i < gossipTxCount; i++ {
		go p2p.gossipTxWorker()
	}
	go p2p.workerOversight()
	go p2p.cleaner()
	return p2p
}

type p2PBus struct {
	interfaces.P2PClient
	*P2PBus
}

func (p2p *p2PBus) Feedback(amount int) {
	var err error
	if amount < 0 {
		amount = amount * (-1)
		err = context.DeadlineExceeded
	}
	go func() {
		for i := 0; i < amount; i++ {
			select {
			case p2p.metricChan <- err:
				continue
			case <-p2p.closeChan:
				return
			}
		}
	}()
}

type P2PBus struct {
	client            interfaces.P2PClient
	reqChan           <-chan interface{}
	gossipChan        <-chan interface{}
	gossipTxChan      <-chan interface{}
	closeChan         <-chan struct{}
	maxRequestWorkers int
	metricChan        chan error
	workerKillChan    chan struct{}
	errMetric         int
	numWorkers        int
	backoff           int
	logger            *logrus.Logger
	cleanup           func()
}

func (p2p *P2PBus) cleaner() {
	<-p2p.closeChan
	p2p.cleanup()
	<-time.After(6 * constants.MsgTimeout)
}

// TODO: add additional logic that allows better introspection
func (p2p *P2PBus) workerOversight() {
	for {
		select {
		case <-p2p.closeChan:
			return
		case err := <-p2p.metricChan:
			if err == nil {
				p2p.errMetric++
				if p2p.errMetric >= p2p.numWorkers*2 {
					if p2p.backoff > 0 {
						p2p.backoff--
					}
					p2p.errMetric = 0
					if p2p.numWorkers < p2p.maxRequestWorkers {
						p2p.numWorkers++
						p2p.logger.Debugf("Increasing peer worker count for peer %v to %v", p2p.client.NodeAddr(), p2p.numWorkers)
						go p2p.reqWorker()
					}
				}
			}
			if err == context.DeadlineExceeded {
				if p2p.backoff == 10 {
					p2p.logger.Debugf("Peer %v disconnecting on maximum backoff", p2p.client.NodeAddr())
					go p2p.client.Close()
					continue
				}
				p2p.errMetric--
				if p2p.errMetric <= -p2p.numWorkers*2 {
					p2p.errMetric = 0
					if p2p.numWorkers > 0 {
						p2p.numWorkers--
						p2p.logger.Debugf("Decreasing peer worker count for peer %v to %v", p2p.client.NodeAddr(), p2p.numWorkers)
						select {
						case <-p2p.closeChan:
							return
						case p2p.workerKillChan <- struct{}{}:
							if p2p.numWorkers == 0 {
								if p2p.backoff < 10 {
									p2p.backoff++
								}
								backoff := p2p.backoffCalc()
								p2p.logger.Debugf("Waiting backoff %v seconds for peer %v", backoff, p2p.client.NodeAddr())
								select {
								case <-time.After(time.Duration(backoff) * time.Second):
									p2p.numWorkers++
									go p2p.reqWorker()
								case <-p2p.closeChan:
									return
								}
							}
						}
					}
				}
			}
		}
	}
}

func (p2p *P2PBus) backoffCalc() int {
	return p2p.backoff * 2
}

func (p2p *P2PBus) reqWorker() {
	p2p.logger.Debugf("Starting request worker for peer %v", p2p.client.NodeAddr())
	for {
		select {
		case <-p2p.closeChan:
			return
		case msg := <-p2p.reqChan:
			p2p.dispatch(msg)
		case <-p2p.workerKillChan:
			return
		}
	}
}

func (p2p *P2PBus) gossipWorker() {
	p2p.logger.Debugf("Starting gossip worker for peer %v", p2p.client.NodeAddr())
	for {
		select {
		case <-p2p.closeChan:
			return
		case msg := <-p2p.gossipChan:
			p2p.dispatch(msg)
		}
	}
}

func (p2p *P2PBus) gossipTxWorker() {
	p2p.logger.Debugf("Starting gossipTx worker for peer %v", p2p.client.NodeAddr())
	for {
		select {
		case <-p2p.closeChan:
			return
		case msg := <-p2p.gossipTxChan:
			p2p.dispatch(msg)
		}
	}
}

func (p2p *P2PBus) dispatch(obj interface{}) {
	switch req := obj.(type) {
	case *StatusRequest:
		middleware.SetPeer(&p2PBus{p2p.client, p2p}, req.opts...)
		ctx, cf := context.WithTimeout(req.ctx, constants.MsgTimeout)
		defer cf()
		r, err := p2p.client.Status(ctx, req.req, req.opts...)
		req.rChan <- &StatusResponse{r, err}
		select {
		case p2p.metricChan <- err:
			return
		case <-p2p.closeChan:
			return
		}
	case *GetBlockHeadersRequest:
		middleware.SetPeer(&p2PBus{p2p.client, p2p}, req.opts...)
		ctx, cf := context.WithTimeout(req.ctx, constants.MsgTimeout)
		defer cf()
		r, err := p2p.client.GetBlockHeaders(ctx, req.req, req.opts...)
		req.rChan <- &GetBlockHeadersResponse{r, err}
		select {
		case p2p.metricChan <- err:
			return
		case <-p2p.closeChan:
			return
		}
	case *GetMinedTxsRequest:
		middleware.SetPeer(&p2PBus{p2p.client, p2p}, req.opts...)
		ctx, cf := context.WithTimeout(req.ctx, constants.MsgTimeout)
		defer cf()
		r, err := p2p.client.GetMinedTxs(ctx, req.req, req.opts...)
		req.rChan <- &GetMinedTxsResponse{r, err}
		select {
		case p2p.metricChan <- err:
			return
		case <-p2p.closeChan:
			return
		}
	case *GetPendingTxsRequest:
		middleware.SetPeer(&p2PBus{p2p.client, p2p}, req.opts...)
		ctx, cf := context.WithTimeout(req.ctx, constants.MsgTimeout)
		defer cf()
		r, err := p2p.client.GetPendingTxs(ctx, req.req, req.opts...)
		req.rChan <- &GetPendingTxsResponse{r, err}
		select {
		case p2p.metricChan <- err:
			return
		case <-p2p.closeChan:
			return
		}
	case *GetSnapShotNodeRequest:
		middleware.SetPeer(&p2PBus{p2p.client, p2p}, req.opts...)
		ctx, cf := context.WithTimeout(req.ctx, constants.MsgTimeout)
		defer cf()
		r, err := p2p.client.GetSnapShotNode(ctx, req.req, req.opts...)
		req.rChan <- &GetSnapShotNodeResponse{r, err}
		select {
		case p2p.metricChan <- err:
			return
		case <-p2p.closeChan:
			return
		}
	case *GetSnapShotStateDataRequest:
		middleware.SetPeer(&p2PBus{p2p.client, p2p}, req.opts...)
		ctx, cf := context.WithTimeout(req.ctx, constants.MsgTimeout)
		defer cf()
		r, err := p2p.client.GetSnapShotStateData(ctx, req.req, req.opts...)
		req.rChan <- &GetSnapShotStateDataResponse{r, err}
		select {
		case p2p.metricChan <- err:
			return
		case <-p2p.closeChan:
			return
		}
	case *GetSnapShotHdrNodeRequest:
		middleware.SetPeer(&p2PBus{p2p.client, p2p}, req.opts...)
		ctx, cf := context.WithTimeout(req.ctx, constants.MsgTimeout)
		defer cf()
		r, err := p2p.client.GetSnapShotHdrNode(ctx, req.req, req.opts...)
		req.rChan <- &GetSnapShotHdrNodeResponse{r, err}
		select {
		case p2p.metricChan <- err:
			return
		case <-p2p.closeChan:
			return
		}
	case *GossipTransactionMessage:
		opts := []grpc.CallOption{
			grpc_retry.WithPerRetryTimeout(constants.MsgTimeout),
			grpc_retry.WithMax(3),
		}
		ctx, cf := context.WithTimeout(req.ctx, 3*constants.MsgTimeout)
		defer cf()
		_, err := p2p.client.GossipTransaction(ctx, req.req, opts...)
		if err != nil {
			utils.DebugTrace(p2p.logger, err)
		}
	case *GossipProposalMessage:
		opts := []grpc.CallOption{
			grpc_retry.WithPerRetryTimeout(constants.MsgTimeout),
			grpc_retry.WithMax(3),
		}
		ctx, cf := context.WithTimeout(req.ctx, 3*constants.MsgTimeout)
		defer cf()
		_, err := p2p.client.GossipProposal(ctx, req.req, opts...)
		if err != nil {
			utils.DebugTrace(p2p.logger, err)
		}
	case *GossipPreVoteMessage:
		opts := []grpc.CallOption{
			grpc_retry.WithPerRetryTimeout(constants.MsgTimeout),
			grpc_retry.WithMax(3),
		}
		ctx, cf := context.WithTimeout(req.ctx, 3*constants.MsgTimeout)
		defer cf()
		_, err := p2p.client.GossipPreVote(ctx, req.req, opts...)
		if err != nil {
			utils.DebugTrace(p2p.logger, err)
		}
	case *GossipPreVoteNilMessage:
		opts := []grpc.CallOption{
			grpc_retry.WithPerRetryTimeout(constants.MsgTimeout),
			grpc_retry.WithMax(3),
		}
		ctx, cf := context.WithTimeout(req.ctx, 3*constants.MsgTimeout)
		defer cf()
		_, err := p2p.client.GossipPreVoteNil(ctx, req.req, opts...)
		if err != nil {
			utils.DebugTrace(p2p.logger, err)
		}
	case *GossipPreCommitMessage:
		opts := []grpc.CallOption{
			grpc_retry.WithPerRetryTimeout(constants.MsgTimeout),
			grpc_retry.WithMax(3),
		}
		ctx, cf := context.WithTimeout(req.ctx, 3*constants.MsgTimeout)
		defer cf()
		_, err := p2p.client.GossipPreCommit(ctx, req.req, opts...)
		if err != nil {
			utils.DebugTrace(p2p.logger, err)
		}
	case *GossipPreCommitNilMessage:
		opts := []grpc.CallOption{
			grpc_retry.WithPerRetryTimeout(constants.MsgTimeout),
			grpc_retry.WithMax(3),
		}
		ctx, cf := context.WithTimeout(req.ctx, 3*constants.MsgTimeout)
		defer cf()
		_, err := p2p.client.GossipPreCommitNil(ctx, req.req, opts...)
		if err != nil {
			utils.DebugTrace(p2p.logger, err)
		}
	case *GossipNextRoundMessage:
		opts := []grpc.CallOption{
			grpc_retry.WithPerRetryTimeout(constants.MsgTimeout),
			grpc_retry.WithMax(3),
		}
		ctx, cf := context.WithTimeout(req.ctx, 3*constants.MsgTimeout)
		defer cf()
		_, err := p2p.client.GossipNextRound(ctx, req.req, opts...)
		if err != nil {
			utils.DebugTrace(p2p.logger, err)
		}
	case *GossipNextHeightMessage:
		opts := []grpc.CallOption{
			grpc_retry.WithPerRetryTimeout(constants.MsgTimeout),
			grpc_retry.WithMax(3),
		}
		ctx, cf := context.WithTimeout(req.ctx, 3*constants.MsgTimeout)
		defer cf()
		_, err := p2p.client.GossipNextHeight(ctx, req.req, opts...)
		if err != nil {
			utils.DebugTrace(p2p.logger, err)
		}
	case *GossipBlockHeaderMessage:
		opts := []grpc.CallOption{
			grpc_retry.WithPerRetryTimeout(constants.MsgTimeout),
			grpc_retry.WithMax(3),
		}
		ctx, cf := context.WithTimeout(req.ctx, 3*constants.MsgTimeout)
		defer cf()
		_, err := p2p.client.GossipBlockHeader(ctx, req.req, opts...)
		if err != nil {
			utils.DebugTrace(p2p.logger, err)
		}
	case *GetPeersRequest:
		ctx, cf := context.WithTimeout(req.ctx, constants.MsgTimeout)
		defer cf()
		r, err := p2p.client.GetPeers(ctx, req.req)
		req.rChan <- &GetPeersResponse{r, err}
	}
}

type P2PClient struct {
	reqChan      chan interface{}
	gossipChan   chan interface{}
	gossipTxChan chan interface{}
}

func (p2p *P2PClient) Status(ctx context.Context, in *pb.StatusRequest, opts ...grpc.CallOption) (*pb.StatusResponse, error) {
	rchan := make(chan *StatusResponse, 1)
	req := &StatusRequest{ctx, in, opts, rchan}
	select {
	case p2p.reqChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, middleware.ErrWouldBlock
		} else {
			p2p.reqChan <- req
		}
	}

	r := <-rchan
	return r.resp, r.err
}

func (p2p *P2PClient) GetBlockHeaders(ctx context.Context, in *pb.GetBlockHeadersRequest, opts ...grpc.CallOption) (*pb.GetBlockHeadersResponse, error) {
	rchan := make(chan *GetBlockHeadersResponse, 1)
	req := &GetBlockHeadersRequest{ctx, in, opts, rchan}
	select {
	case p2p.reqChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, middleware.ErrWouldBlock
		} else {
			p2p.reqChan <- req
		}
	}

	r := <-rchan
	return r.resp, r.err
}

func (p2p *P2PClient) GetMinedTxs(ctx context.Context, in *pb.GetMinedTxsRequest, opts ...grpc.CallOption) (*pb.GetMinedTxsResponse, error) {
	rchan := make(chan *GetMinedTxsResponse, 1)
	req := &GetMinedTxsRequest{ctx, in, opts, rchan}
	select {
	case p2p.reqChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, middleware.ErrWouldBlock
		} else {
			p2p.reqChan <- req
		}
	}

	r := <-rchan
	return r.resp, r.err
}

func (p2p *P2PClient) GetPendingTxs(ctx context.Context, in *pb.GetPendingTxsRequest, opts ...grpc.CallOption) (*pb.GetPendingTxsResponse, error) {
	rchan := make(chan *GetPendingTxsResponse, 1)
	req := &GetPendingTxsRequest{ctx, in, opts, rchan}
	select {
	case p2p.reqChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, middleware.ErrWouldBlock
		} else {
			p2p.reqChan <- req
		}
	}

	r := <-rchan
	return r.resp, r.err
}

func (p2p *P2PClient) GetSnapShotNode(ctx context.Context, in *pb.GetSnapShotNodeRequest, opts ...grpc.CallOption) (*pb.GetSnapShotNodeResponse, error) {
	rchan := make(chan *GetSnapShotNodeResponse, 1)
	req := &GetSnapShotNodeRequest{ctx, in, opts, rchan}
	select {
	case p2p.reqChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, middleware.ErrWouldBlock
		} else {
			p2p.reqChan <- req
		}
	}

	r := <-rchan
	return r.resp, r.err
}

func (p2p *P2PClient) GetSnapShotStateData(ctx context.Context, in *pb.GetSnapShotStateDataRequest, opts ...grpc.CallOption) (*pb.GetSnapShotStateDataResponse, error) {
	rchan := make(chan *GetSnapShotStateDataResponse, 1)
	req := &GetSnapShotStateDataRequest{ctx, in, opts, rchan}
	select {
	case p2p.reqChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, middleware.ErrWouldBlock
		} else {
			p2p.reqChan <- req
		}
	}

	r := <-rchan
	return r.resp, r.err
}

func (p2p *P2PClient) GetSnapShotHdrNode(ctx context.Context, in *pb.GetSnapShotHdrNodeRequest, opts ...grpc.CallOption) (*pb.GetSnapShotHdrNodeResponse, error) {
	rchan := make(chan *GetSnapShotHdrNodeResponse, 1)
	req := &GetSnapShotHdrNodeRequest{ctx, in, opts, rchan}
	select {
	case p2p.reqChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, middleware.ErrWouldBlock
		} else {
			p2p.reqChan <- req
		}
	}

	r := <-rchan
	return r.resp, r.err
}

func (p2p *P2PClient) GetPeers(ctx context.Context, in *pb.GetPeersRequest, opts ...grpc.CallOption) (*pb.GetPeersResponse, error) {
	rchan := make(chan *GetPeersResponse, 1)
	req := &GetPeersRequest{ctx, in, opts, rchan}
	select {
	case p2p.reqChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, middleware.ErrWouldBlock
		} else {
			p2p.reqChan <- req
		}
	}

	r := <-rchan
	return r.resp, r.err
}

func (p2p *P2PClient) GossipTransaction(ctx context.Context, in *pb.GossipTransactionMessage, opts ...grpc.CallOption) (*pb.GossipTransactionAck, error) {
	req := &GossipTransactionMessage{ctx, in, opts}
	select {
	case p2p.gossipTxChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, ErrWouldBlock
		} else {
			p2p.gossipTxChan <- req
		}
	}

	return &pb.GossipTransactionAck{}, nil
}

func (p2p *P2PClient) GossipProposal(ctx context.Context, in *pb.GossipProposalMessage, opts ...grpc.CallOption) (*pb.GossipProposalAck, error) {
	req := &GossipProposalMessage{ctx, in, opts}
	select {
	case p2p.gossipChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, ErrWouldBlock
		} else {
			p2p.gossipChan <- req
		}
	}

	return &pb.GossipProposalAck{}, nil
}

func (p2p *P2PClient) GossipPreVote(ctx context.Context, in *pb.GossipPreVoteMessage, opts ...grpc.CallOption) (*pb.GossipPreVoteAck, error) {
	req := &GossipPreVoteMessage{ctx, in, opts}
	select {
	case p2p.gossipChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, ErrWouldBlock
		} else {
			p2p.gossipChan <- req
		}
	}

	return &pb.GossipPreVoteAck{}, nil
}

func (p2p *P2PClient) GossipPreVoteNil(ctx context.Context, in *pb.GossipPreVoteNilMessage, opts ...grpc.CallOption) (*pb.GossipPreVoteNilAck, error) {
	req := &GossipPreVoteNilMessage{ctx, in, opts}
	select {
	case p2p.gossipChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, ErrWouldBlock
		} else {
			p2p.gossipChan <- req
		}
	}

	return &pb.GossipPreVoteNilAck{}, nil
}

func (p2p *P2PClient) GossipPreCommit(ctx context.Context, in *pb.GossipPreCommitMessage, opts ...grpc.CallOption) (*pb.GossipPreCommitAck, error) {
	req := &GossipPreCommitMessage{ctx, in, opts}
	select {
	case p2p.gossipChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, ErrWouldBlock
		} else {
			p2p.gossipChan <- req
		}
	}

	return &pb.GossipPreCommitAck{}, nil
}

func (p2p *P2PClient) GossipPreCommitNil(ctx context.Context, in *pb.GossipPreCommitNilMessage, opts ...grpc.CallOption) (*pb.GossipPreCommitNilAck, error) {
	req := &GossipPreCommitNilMessage{ctx, in, opts}
	select {
	case p2p.gossipChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, ErrWouldBlock
		} else {
			p2p.gossipChan <- req
		}
	}

	return &pb.GossipPreCommitNilAck{}, nil
}

func (p2p *P2PClient) GossipNextRound(ctx context.Context, in *pb.GossipNextRoundMessage, opts ...grpc.CallOption) (*pb.GossipNextRoundAck, error) {
	req := &GossipNextRoundMessage{ctx, in, opts}
	select {
	case p2p.gossipChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, ErrWouldBlock
		} else {
			p2p.gossipChan <- req
		}
	}

	return &pb.GossipNextRoundAck{}, nil
}

func (p2p *P2PClient) GossipNextHeight(ctx context.Context, in *pb.GossipNextHeightMessage, opts ...grpc.CallOption) (*pb.GossipNextHeightAck, error) {
	req := &GossipNextHeightMessage{ctx, in, opts}
	select {
	case p2p.gossipChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, ErrWouldBlock
		} else {
			p2p.gossipChan <- req
		}
	}

	return &pb.GossipNextHeightAck{}, nil
}

func (p2p *P2PClient) GossipBlockHeader(ctx context.Context, in *pb.GossipBlockHeaderMessage, opts ...grpc.CallOption) (*pb.GossipBlockHeaderAck, error) {
	req := &GossipBlockHeaderMessage{ctx, in, opts}
	select {
	case p2p.gossipChan <- req:
	default:
		if !middleware.CanBlock(opts...) {
			return nil, ErrWouldBlock
		} else {
			p2p.gossipChan <- req
		}
	}

	return &pb.GossipBlockHeaderAck{}, nil
}

var ErrWouldBlock = errors.New("unable to broadcast due to blocking")
