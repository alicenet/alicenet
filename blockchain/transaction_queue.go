package blockchain

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

//
var (
	ErrUnknownRequest = errors.New("unknown request type")
)

// Internal struct to keep track of transactions that are being monitoring
type TransactionInfo struct {
	ctx                  context.Context         // ctx used for calling the monitoring a certain tx
	txn                  *types.Transaction      // Transaction object
	txConfirmationBlocks uint64                  // number of blocks that we should wait to consider a tx valid (return the receipt)
	selector             interfaces.FuncSelector // 4 bytes that identify the function being called by the tx
	functionSignature    string                  // function signature as we see on the smart contracts
	commitedHeight       uint64                  // ethereum height where we first added the tx to be watched. Mainly used to see if a tx was dropped
	isStale              bool                    // flag to keep track if a tx is sitting too long in the txPool without being mined
}

// Internal struct to keep track of what blocks we already checked during monitoring
type BlockInfo struct {
	height uint64      // block height
	hash   common.Hash // block header hash
}

// Compare if 2 blockInfo structs are equal by comparing the height and block hash
func (a *BlockInfo) Equal(b *BlockInfo) bool {
	return bytes.Equal(a.hash[:], b.hash[:]) && a.height == b.height
}

// Type to do request against the tx receipt monitoring system
type Request struct {
	command              string             // command to be executed by the monitoring system
	ctx                  context.Context    // tx ctx used for tx monitoring cancelation
	txn                  *types.Transaction // the transaction that should watched
	txConfirmationBlocks uint64             // number of ethereum blocks that we should wait until we return the receipt
	txGroup              uint64             // Id to group a tx with other tx. Only once all tx in a group are completed we return the receipts of all txs in that group
	respCh               chan *Response     // channel where the response will be send
}

// Response of the monitoring system
type Response struct {
	message string           // string containing information about the request+
	err     error            // response error that happened during processing
	rcpt    *types.Receipt   // tx receipt after txConfirmationBlocks of a tx that was not queued in txGroup
	rcpts   []*types.Receipt // list of receipts of tx into a txGroup after all tx have respected their txConfirmationBlocks
}

// TransactionProfile to keep track of gas metrics in the overall system
type TransactionProfile struct {
	AverageGas   uint64
	MinimumGas   uint64
	MaximumGas   uint64
	TotalCount   uint64
	TotalGas     uint64
	TotalSuccess uint64
}

// Backend structs used to monitor Ethereum transactions and retrieve their receipts
type Behind struct {
	sync.RWMutex
	lastProcessedBlock  *BlockInfo                                     // Last ethereum block that we checked for receipts
	waitingTxns         map[common.Hash]*TransactionInfo               // Just a MAP of transactions whose receipts we're looking for
	readyTxns           map[common.Hash]*types.Receipt                 // All the transaction -> receipt pairs we know of
	groups              map[uint64][]*TransactionInfo                  // A group is just an ID and a list of transactions
	aggregates          map[interfaces.FuncSelector]TransactionProfile // Struct to keep track of the gas metrics used by the system
	client              interfaces.GethClient                          // An interface with the Geth functionality we need
	knownSelectors      interfaces.SelectorMap                         //
	logger              *logrus.Entry                                  // Logger to log messages
	reqCh               <-chan *Request                                // Channel used to send request to this backend service
	timeout             time.Duration                                  // How long will we wait for a receipt
	txNotFoundMaxBlocks uint64                                         // How many blocks we should wait for removing a tx from the waiting list in case we don't find in the ethereum chain
	pollingTime         time.Duration                                  // time which we should poll the ethereum node to check for new blocks
}

func (b *Behind) Loop() {

	done := false
	pollingTime := b.pollingTime
	for !done {
		select {
		case req, ok := <-b.reqCh:
			// Some sort of request came in
			if !ok {
				b.logger.Debugf("command channel closed")
				b.status(nil)
				done = true
				break
			}

			b.logger.Debugf("received request: %v channel open: %v", req.command, ok)

			var handler func(*Request) *Response
			switch req.command {
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
		case <-time.After(pollingTime):
			b.collectReceipts()
		}
	}
	b.logger.Debug("finished")
}

func (b *Behind) collectReceipts() {
	b.Lock()
	defer b.Unlock()

	// If there's no tx to be monitored just return
	if len(b.waitingTxns) == 0 {
		return
	}

	// this was set to one hour in an earlier version. This had
	// the potential to starve other services waiting for the lock.
	// Since this will be performed approx every 1 second,
	// there is no need to wait more than what is the worst possible read
	// speed that would allow the node to remain in sync if a significant
	// number of requests took this long - 6 seconds is pretty long
	// delay, but we only expect this to take 250 Milliseconds per request
	// in a normal setting. Hence the actual runtime is likely very short.
	ctx, cf := context.WithTimeout(context.Background(), 6*time.Second)
	defer cf()

	blockHeader, err := b.client.HeaderByNumber(ctx, nil)
	if err != nil {
		b.logger.Errorf("error getting latest block number from ethereum node: %v", err)
		return
	}
	blockInfo := &BlockInfo{
		blockHeader.Number.Uint64(),
		blockHeader.Hash(),
	}

	// we already processed the last block we just return
	if blockInfo.Equal(b.lastProcessedBlock) {
		return
	}

	var expiredTxs map[common.Hash]*TransactionInfo
	for txn, txnInfo := range b.waitingTxns {
		select {
		case <-txnInfo.ctx.Done():
			// the go-routine who wanted this information has stopped caring. This
			// most likely indicates a failure, and cancellation of polling
			// prevents a memory leak so we add this to txs that will be removed after the processing
			b.logger.Tracef("context for tx %v is finished, marking it to be excluded!", txn.Hex())
			expiredTxs[txn] = txnInfo
		default:
			// if this is the first time seeing a tx, mark
			if txnInfo.commitedHeight == 0 {
				txnInfo.commitedHeight = blockInfo.height
			}
			_, isPending, err := b.client.TransactionByHash(ctx, txn)
			if err != nil {
				// if we couldn't locate a tx after 100 blocks, probably means that it was dropped
				if err == ethereum.NotFound && (blockInfo.height-txnInfo.commitedHeight >= 100) {
					b.logger.Tracef("could not find tx and 100 blocks have passed, marking tx %v to be excluded!", txn.Hex(), err)
					expiredTxs[txn] = txnInfo
				}
				b.logger.Tracef("error getting tx: %v: %v", txn, err)
				continue
			}

			if isPending {
				if blockInfo.height-txnInfo.commitedHeight >= 10 {
					txnInfo.isStale = true

				}
				continue
			}

			// context is good on the tx level object, so check for receipt
			rcpt, err := b.client.TransactionReceipt(ctx, txn)
			if err != nil {
				// couldn't find the receipt, checking if the tx in the txPool (pending)
				if err == ethereum.NotFound {
					b.logger.Errorf("error getting receipt: %v: %v", txn, err)
				}

				b.logger.Errorf("error getting receipt for tx: %v error: %v", txn.Hex(), err)
				return
			}

			// This only happens using a SimulatedBackend during tests
			if isTestRun() {
				if rcpt == nil {
					b.logger.Debugf("receipt is nil: %v", txn.Hex())
					return
				}
			}

			b.readyTxns[txn] = rcpt

			var profile TransactionProfile
			var selector [4]byte
			var sig string
			var present bool

			if selector, present = b.selectors[txn]; present {
				profile = b.aggregates[selector]
				sig = b.knownSelectors.Signature(selector)
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
		}
	}

	// Cleaning expired transactions
	for _, txn := range expiredTxs {
		delete(b.waitingTxns, txn.txn.Hash())
	}
	b.lastProcessedBlock = blockInfo
}

func (b *Behind) process(req *Request, handler func(req *Request) *Response) {

	b.logger.Debugf("processing request for command: %v", req.command)

	resp := handler(req)

	b.logger.Debugf("response channel present: %v", req.respCh != nil)
	if req.respCh != nil {
		req.respCh <- resp
	}
	b.logger.Debug("...done processing request")
}

func (b *Behind) queue(req *Request) *Response {

	if req.txn != nil {
		return &Response{
			message: "",
			err:     errors.New("invalid request: txn not sent!"),
			rcpt:    &types.Receipt{},
			rcpts:   []*types.Receipt{},
		}
	}

	b.Lock()
	defer b.Unlock()

	txnHash := req.txn.Hash()

	selector := ExtractSelector(req.txn.Data())

	//todo: replace this with a generated mapping from the bindings
	sig := b.knownSelectors.Signature(selector)

	logEntry := b.logger.WithField("Transaction", txnHash).
		WithField("Function", sig).
		WithField("Selector", fmt.Sprintf("%x", selector))

	ctx := context.Background()
	if req.ctx != nil {
		ctx = req.ctx
	}

	txInfo := &TransactionInfo{
		ctx:                  ctx,
		txn:                  req.txn,
		txConfirmationBlocks: req.txConfirmationBlocks,
		selector:             selector,
		functionSignature:    sig,
	}

	b.waitingTxns[txnHash] = txInfo

	if req.txGroup > 0 {
		if _, present := b.groups[req.txGroup]; !present {
			b.groups[req.txGroup] = make([]*TransactionInfo, 0, 10)
		}
		b.groups[req.txGroup] = append(b.groups[req.txGroup], txInfo)
		logEntry.Debugf("Transaction queued under the group: %v", req.txGroup)
	} else {
		logEntry.Debug("Transaction queued")
	}

	return &Response{message: "queued transaction"}
}

func (b *Behind) status(req *Request) *Response {
	b.RLock()
	defer b.RUnlock()

	b.logger.WithField("Completed", len(b.readyTxns)).
		WithField("Pending", len(b.waitingTxns)).
		Info("Transaction counts")

	for selector, profile := range b.aggregates {
		sig := b.knownSelectors.Signature(selector)
		b.logger.WithField("Selector", fmt.Sprintf("%x", selector)).
			WithField("Function", sig).
			WithField("Profile", fmt.Sprintf("%+v", profile)).
			Info("Status")
	}

	return &Response{message: "status check"}
}

func (b *Behind) wait(req *Request) *Response {

	if req.txn == nil && req.txGroup == 0 {
		return &Response{
			message: "",
			err:     errors.New("invalid request: txn or txGroup should sent for waiting!"),
			rcpt:    &types.Receipt{},
			rcpts:   []*types.Receipt{},
		}
	}

	// ensure the request has a valid context
	parentCtx := context.Background()
	if req.ctx != nil {
		parentCtx = context.Background()
	}

	// derive the context from a parent context that comes from caller
	// the default behavior is to wait a VERY long time to prevent
	// tx re-submission, but for a certain class of actions this is
	// not valid behavior. This is most important during EthDKG
	ctx, cf := context.WithTimeout(parentCtx, b.timeout)
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
				delete(b.readyTxns, txn)
				done = true
			} else {
				b.logger.Debugf("Receipt not ready yet for %v", txn.Hex())
			}
		} else {
			// waiting for a group of transactions to complete
			allPresent := true
			for _, txnInfo := range b.groups[req.txGroup] {
				if _, present := b.readyTxns[txnInfo.txn.Hash()]; !present {
					allPresent = false
				}
			}

			if allPresent {
				resp.rcpts = make([]*types.Receipt, 0, len(b.groups[req.txGroup]))
				for _, txnInfo := range b.groups[req.txGroup] {
					resp.rcpts = append(resp.rcpts, b.readyTxns[txnInfo.txn.Hash()])
					delete(b.readyTxns, txnInfo.txn.Hash())
				}
				delete(b.groups, req.txGroup)
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

type TxnQueueDetail struct {
	backend *Behind
	logger  *logrus.Entry
	reqCh   chan<- *Request
}

func NewTxnQueue(client interfaces.GethClient, sm interfaces.SelectorMap, to time.Duration) *TxnQueueDetail {
	reqch := make(chan *Request, 100)

	b := &Behind{
		reqCh:              reqch,
		client:             client,
		logger:             logging.GetLogger("ethereum").WithField("Component", "behind"),
		waitingTxns:        make(map[common.Hash]*TransactionInfo),
		readyTxns:          make(map[common.Hash]*types.Receipt),
		aggregates:         make(map[interfaces.FuncSelector]TransactionProfile),
		knownSelectors:     sm,
		timeout:            to,
		groups:             make(map[uint64][]*TransactionInfo),
		lastProcessedBlock: &BlockInfo{0, common.HexToHash("")},
		pollingTime:        1 * time.Second,
	}

	q := &TxnQueueDetail{
		reqCh:   reqch,
		backend: b,
		logger:  logging.GetLogger("ethereum").WithField("Component", "infront"),
	}
	return q
}

func (f *TxnQueueDetail) StartLoop() {
	go f.backend.Loop()
}

func (f *TxnQueueDetail) QueueTransaction(ctx context.Context, txn *types.Transaction, confirmationBlocks uint64) (*Response, error) {
	f.logger.WithField("Txn", txn.Hash()).Debug("Queueing")
	respChannel := make(chan *Response, 1)
	defer close(respChannel)
	req := &Request{ctx: ctx, command: "queue", txn: txn, respCh: respChannel, txConfirmationBlocks: confirmationBlocks}
	response, err := f.requestWait(ctx, req)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (f *TxnQueueDetail) QueueGroupTransaction(ctx context.Context, groupId uint64, txn *types.Transaction, confirmationBlocks uint64) (*Response, error) {
	if groupId == 0 {
		return nil, errors.New("groupTx should be greater than 0!")
	}
	respChannel := make(chan *Response, 1)
	defer close(respChannel)
	req := &Request{ctx: ctx, command: "queue", txn: txn, txGroup: groupId, respCh: respChannel, txConfirmationBlocks: confirmationBlocks}
	response, err := f.requestWait(ctx, req)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (f *TxnQueueDetail) QueueAndWait(ctx context.Context, txn *types.Transaction, confirmationBlocks uint64) (*types.Receipt, error) {
	f.QueueTransaction(ctx, txn, confirmationBlocks)
	return f.WaitTransaction(ctx, txn)
}

func (f *TxnQueueDetail) WaitTransaction(ctx context.Context, txn *types.Transaction) (*types.Receipt, error) {
	respChannel := make(chan *Response, 1)
	defer close(respChannel)
	req := &Request{ctx: ctx, command: "wait", txn: txn, respCh: respChannel}
	resp, err := f.requestWait(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.rcpt == nil {
		return nil, errors.New(fmt.Sprintf("no receipts were found for txn: %v", txn.Hash()))
	}

	return resp.rcpt, nil
}

func (f *TxnQueueDetail) WaitGroupTransactions(ctx context.Context, groupId uint64) ([]*types.Receipt, error) {
	if groupId == 0 {
		return nil, errors.New("groupTx should be greater than 0!")
	}
	respChannel := make(chan *Response, 1)
	defer close(respChannel)
	req := &Request{ctx: ctx, command: "wait", txGroup: groupId, respCh: respChannel}
	resp, err := f.requestWait(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.rcpts == nil {
		return nil, errors.New(fmt.Sprintf("no receipts were found for group: %v", groupId))
	}

	return resp.rcpts, nil
}

func (f *TxnQueueDetail) Status(ctx context.Context) error {
	req := &Request{ctx: ctx, command: "status", respCh: make(chan *Response, 1)}
	_, err := f.requestWait(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (f *TxnQueueDetail) Close() {
	f.logger.Debug("closing request channel...")
	close(f.reqCh)
}

func (f *TxnQueueDetail) requestWait(ctx context.Context, req *Request) (*Response, error) {

	if req.respCh == nil {
		return nil, errors.New("invalid request, request doesn't containt a response channel")
	}
	f.reqCh <- req

	logReceipt := func(message string, rcpt *types.Receipt) {
		f.logger.WithFields(logrus.Fields{
			"Message":     message,
			"Transaction": rcpt.TxHash.Hex(),
			"Block":       rcpt.BlockNumber.String(),
			"GasUsed":     rcpt.GasUsed,
			"Status":      rcpt.Status,
		}).Debugf("Received response")
	}

	select {
	case resp, ok := <-req.respCh:
		if !ok {
			return nil, errors.New("response channel closed")
		}
		if resp == nil {
			return nil, errors.New("no response or error")
		}
		if resp.err != nil {
			return nil, resp.err
		}

		if resp.rcpt != nil {
			logReceipt(resp.message, resp.rcpt)
		} else if resp.rcpts != nil {
			for _, rcpt := range resp.rcpts {
				logReceipt(resp.message, rcpt)
			}
		}
		return resp, nil

	case <-ctx.Done():
		f.logger.Infof("context cancelled: %v", ctx.Err())
		return nil, ctx.Err()
	}
}

func isTestRun() bool {
	return flag.Lookup("test.v") != nil
}
