package blockchain

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/MadBase/MadNet/logging"
	geth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

//
var (
	ErrUnknownRequest = errors.New("Unknown request type")
)

//
type Request struct {
	name   string
	txn    *types.Transaction
	group  int
	respch chan *Response
	to     time.Duration
}

//
type Response struct {
	message string
	err     error
	rcpt    *types.Receipt
	rcpts   []*types.Receipt
}

// TransactionProfile // TODO calculate updated values for each receipt receipt
type TransactionProfile struct {
	AverageGas   uint64
	MinimumGas   uint64
	MaximumGas   uint64
	TotalCount   uint64
	TotalGas     uint64
	TotalSuccess uint64
}

// Behind is the struct used while monitoring Ethereum transactions
type Behind struct {
	sync.Mutex
	waitingTxns    []common.Hash                       // Just a list of transactions whose receipts we're looking for
	readyTxns      map[common.Hash]*types.Receipt      // All the transaction -> receipt pairs we know of
	selectors      map[common.Hash]FuncSelector        // Maps a transaction to it's function selector
	groups         map[int][]common.Hash               // A group is just an ID and a list of transactions
	names          map[FuncSelector]string             // Map function selector's to their names
	aggregates     map[FuncSelector]TransactionProfile //
	client         GethClient                          // An interface with the Geth functionality we need
	knownSelectors *SelectorMap                        //
	logger         *logrus.Logger                      //
	reqch          <-chan *Request                     //
}

func (b *Behind) Loop() {

	done := false
	for !done {
		select {
		case req, ok := <-b.reqch:
			// Some sort of request came in
			if !ok {
				b.logger.Debugf("command channel closed")
				b.status(nil)
				done = true
				break
			}

			b.logger.Debugf("received request: %v channel open: %v", req.name, ok)

			var handler func(*Request) *Response
			switch req.name {
			case "queue":
				handler = b.queue
			case "status":
				handler = b.status
			case "wait":
				handler = b.wait
			default:
				handler = b.unknown
			}

			go b.process(req, handler)

		case <-time.After(500 * time.Millisecond):
			// No request, so let's do some work
			b.logger.Debugf("tick")
			b.collectReceipts()
		}
	}
	b.logger.Infof("buh bye")
}

func (b *Behind) collectReceipts() {
	ctx, cf := context.WithTimeout(context.Background(), 1*time.Second)
	defer cf()

	b.Lock()
	defer b.Unlock()

	n := len(b.waitingTxns)
	if n < 1 {
		return
	}

	// loop over transactions in need of receipts while building new list
	remainingTxns := make([]common.Hash, 0, n)
	for _, txn := range b.waitingTxns {
		rcpt, err := b.client.TransactionReceipt(ctx, txn)
		if err == geth.NotFound || (err == nil && rcpt == nil) {
			b.logger.Debugf("receipt not found: %v", txn.Hex())
		} else if err != nil {
			b.logger.Errorf("error getting receipt: %v", txn)
		} else if rcpt != nil {
			b.readyTxns[txn] = rcpt

			var profile TransactionProfile
			var selector [4]byte
			var present bool

			if selector, present = b.selectors[txn]; present {
				profile = b.aggregates[selector]
			} else {
				profile = TransactionProfile{}
			}

			// Update transaction profile
			profile.AverageGas = (profile.AverageGas*profile.TotalCount + rcpt.GasUsed) / (profile.TotalCount + 1)
			if profile.MaximumGas < rcpt.GasUsed {
				profile.MaximumGas = rcpt.GasUsed
			}
			if profile.MinimumGas == 0 || profile.MinimumGas > rcpt.GasUsed {
				profile.MinimumGas = rcpt.GasUsed
			}
			profile.TotalCount++
			profile.TotalGas += rcpt.GasUsed
			if rcpt.Status == uint64(1) {
				profile.TotalSuccess++
			}

			b.aggregates[selector] = profile
			b.logger.Infof("found receipt for transaction: %v profile:%+v", rcpt.TxHash.Hex(), profile)
		}

		if _, present := b.readyTxns[txn]; !present {
			remainingTxns = append(remainingTxns, txn)
		}
	}

	b.logger.Debugf("started with %v txns, %v are remaining", len(b.waitingTxns), len(remainingTxns))
	b.waitingTxns = remainingTxns
}

func (b *Behind) process(req *Request, handler func(req *Request) *Response) {

	b.logger.Infof("processing request...")

	resp := handler(req)

	b.logger.Debugf("response channel present: %v", req.respch != nil)
	if req.respch != nil {
		req.respch <- resp
		close(req.respch)
	}
	b.logger.Infof("...done processing request")
}

func (b *Behind) queue(req *Request) *Response {

	b.Lock()
	defer b.Unlock()

	txnHash := req.txn.Hash()

	selector := ExtractSelector(req.txn.Data())
	b.logger.Infof("queueing selector:%x", selector)

	b.selectors[txnHash] = selector
	b.waitingTxns = append(b.waitingTxns, txnHash)

	if _, present := b.groups[req.group]; !present {
		b.groups[req.group] = make([]common.Hash, 0, 10)
	}
	b.groups[req.group] = append(b.groups[req.group], txnHash)

	if b.logger.Level <= logrus.DebugLevel {
		b.logger.Debugf("total completed: %v pending:%v", len(b.readyTxns), len(b.waitingTxns))
		for k, v := range b.groups {
			b.logger.Debugf("txn group:%v txn count:%v", k, len(v))
		}
	}

	return &Response{message: "queued transaction"}
}

func (b *Behind) status(req *Request) *Response {
	b.Lock()
	defer b.Unlock()

	for selector, profile := range b.aggregates {
		b.logger.Infof("function selector:%x signature:%v profile:%+v", selector, b.knownSelectors.Signature(selector), profile)
	}

	return &Response{message: "status check"}
}

func (b *Behind) wait(req *Request) *Response {

	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second) // TODO context or duration has to be passed in
	defer cf()

	resp := &Response{message: "status check"}
	done := false

	check := func() {
		b.Lock()
		defer b.Unlock()

		if req.txn != nil {
			// waiting for a specific transaction to complete
			txn := req.txn.Hash()
			if rcpt, present := b.readyTxns[txn]; present {
				resp.rcpt = rcpt
				// delete(b.readyTxns, txn) // TODO Add an explicit purge/cleanup
				done = true
			} else {
				b.logger.Infof("rcpt not ready yet for %v", txn.Hex())
			}
		} else {
			// waiting for a group of transactions to complete
			allPresent := true
			for _, txn := range b.groups[req.group] {
				if _, present := b.readyTxns[txn]; !present {
					allPresent = false
				}
			}

			if allPresent {
				resp.rcpts = make([]*types.Receipt, 0, len(b.groups[req.group]))
				for _, txn := range b.groups[req.group] {
					resp.rcpts = append(resp.rcpts, b.readyTxns[txn])
					// delete(b.readyTxns, txn) // TODO Add an explicit purge/cleanup
				}
				// delete(b.groups, req.group) // TODO Add an explicit purge/cleanup
				done = true
			}
		}
	}

	check()
	for !done {
		select {
		case <-time.After(500 * time.Millisecond):
			check()
		case <-ctx.Done():
			done = true
			resp.err = ctx.Err()
		}
	}

	return resp
}

func (b *Behind) unknown(req *Request) *Response {
	return &Response{err: ErrUnknownRequest}
}

type TxnQueue struct {
	backend *Behind
	logger  *logrus.Logger
	reqch   chan<- *Request
}

func NewTxnQueue(client GethClient, sm *SelectorMap) *TxnQueue {
	reqch := make(chan *Request, 10)

	b := &Behind{
		reqch:          reqch,
		client:         client,
		logger:         logging.GetLogger("behind"),
		waitingTxns:    make([]common.Hash, 0, 20),
		readyTxns:      make(map[common.Hash]*types.Receipt),
		selectors:      make(map[common.Hash]FuncSelector),
		aggregates:     make(map[FuncSelector]TransactionProfile),
		knownSelectors: sm,
		groups:         make(map[int][]common.Hash)}

	q := &TxnQueue{
		reqch:   reqch,
		backend: b,
		logger:  logging.GetLogger("infront")}

	return q
}

func (f *TxnQueue) StartLoop() {
	go f.backend.Loop()
}

func (f *TxnQueue) QueueTransaction(ctx context.Context, txn *types.Transaction) {
	txn.Data()
	f.logger.Infof("queue...")
	req := &Request{name: "queue", txn: txn} // no response channel because I don't want to wait
	f.requestWait(ctx, req)
	f.logger.Infof("...done queueing")
}

func (f *TxnQueue) QueueGroupTransaction(ctx context.Context, grp int, txn *types.Transaction) {
	f.logger.Infof("queue...")
	req := &Request{name: "queue", txn: txn, group: grp} // no response channel because I don't want to wait
	f.requestWait(ctx, req)
	f.logger.Infof("...done queueing")
}

func (f *TxnQueue) QueueAndWait(ctx context.Context, txn *types.Transaction) (*types.Receipt, error) {
	f.QueueTransaction(ctx, txn)
	return f.WaitTransaction(ctx, txn)
}

func (f *TxnQueue) WaitTransaction(ctx context.Context, txn *types.Transaction) (*types.Receipt, error) {
	f.logger.Infof("waiting...")
	req := &Request{name: "wait", txn: txn, respch: make(chan *Response)}
	resp := f.requestWait(ctx, req)
	f.logger.Infof("...done waiting")

	if resp.err != nil {
		return nil, resp.err
	}

	return resp.rcpt, nil
}

func (f *TxnQueue) WaitGroupTransactions(ctx context.Context, grp int) ([]*types.Receipt, error) {
	f.logger.Infof("waiting...")
	req := &Request{name: "wait", group: grp, respch: make(chan *Response)}
	resp := f.requestWait(ctx, req)
	f.logger.Infof("...done waiting")

	if resp.err != nil {
		return nil, resp.err
	}

	return resp.rcpts, nil
}

func (f *TxnQueue) requestWait(ctx context.Context, req *Request) *Response {
	f.reqch <- req

	logReciept := func(message string, rcpt *types.Receipt) {
		f.logger.Debugf("response message: %v txn: %v block: %v gas: %v status: %v",
			message,
			rcpt.TxHash.Hex(),
			rcpt.BlockNumber.String(),
			rcpt.GasUsed,
			rcpt.Status)
	}

	if req.respch != nil {
		select {
		case resp, ok := <-req.respch:
			if !ok {
				f.logger.Errorf("response channel closed")
			} else if resp != nil {
				if resp.err != nil {
					f.logger.Infof("response error: %v", resp.err.Error())
				}

				if resp.rcpt != nil {
					logReciept(resp.message, resp.rcpt)
				} else if resp.rcpts != nil {
					for _, rcpt := range resp.rcpts {
						logReciept(resp.message, rcpt)
					}
				}

			} else {
				f.logger.Warn("no response or error")
			}
			return resp

		case <-ctx.Done():
			f.logger.Infof("context cancelled: %v", ctx.Err())
			return &Response{err: ctx.Err()}
		}
	}

	return nil // no response channel, so no meaningful response
}

func (f *TxnQueue) Close() {
	f.logger.Debugf("closing request channel...")
	close(f.reqch)
}
