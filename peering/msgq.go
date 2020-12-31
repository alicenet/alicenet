package peering

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/minio/highwayhash"
	"github.com/sirupsen/logrus"
)

var internalKey []byte

func init() {
	keytmp, err := utils.RandomBytes(constants.HashLen)
	if err != nil {
		panic(err)
	}
	internalKey = keytmp
}

type node struct {
	next *node
	prev *node
	name string
}

func (n *node) Next() *node {
	return n.next
}

func (n *node) Prev() *node {
	return n.next
}

func (n *node) InsertAfter(new *node) {
	if n.prev == n {
		n.prev = new
		n.next = new
		return
	}
	new.prev = n
	new.next = n.next
	n.next.prev = new
	n.next = new
}

func (n *node) Pop() {
	if n.prev == n {
		return
	}
	n.prev.next = n.next
	n.next.prev = n.prev
	n.next = n
	n.prev = n
}

func (n *node) Name() string {
	return n.name
}

func newNode(name string) *node {
	n := &node{}
	n.next = n
	n.prev = n
	n.name = name
	return n
}

type linkedList struct {
	head  *node
	tail  *node
	nodes map[string]*node
	max   int
}

func (ll *linkedList) Contains(name string) bool {
	_, ok := ll.nodes[name]
	return ok
}

func (ll *linkedList) Pop(name string) {
	popNode, ok := ll.nodes[name]
	if ok {
		if ll.head.Name() == name {
			newHead := ll.head.Next()
			newTail := ll.tail
			ll.head.Pop()
			ll.head = newHead
			ll.tail = newTail
			delete(ll.nodes, name)
			return
		}
		// if it is at tail, rotate the LL head and tail by one
		if ll.tail.Name() == name {
			newHead := ll.head
			newTail := ll.tail.Prev()
			ll.tail.Pop()
			ll.head = newHead
			ll.tail = newTail
			delete(ll.nodes, name)
			return
		}
		popNode.Pop()
		delete(ll.nodes, name)
	}
}

func (ll *linkedList) Push(names ...string) []string {
	for i := 0; i < len(names); i++ {
		// if the name is already in map, pop and push to head
		_, ok := ll.nodes[names[i]]
		if ok {
			// if it is already at head, continue to next
			if ll.head.Name() == names[i] {
				continue
			}
			// if it is at tail, rotate the LL head and tail by one
			if ll.tail.Name() == names[i] {
				ll.head = ll.tail
				ll.tail = ll.head.Prev()
				continue
			}
			// otherwise do a pop and insert to head
			n := ll.nodes[names[i]]
			n.Pop()
			ll.tail.InsertAfter(n)
			ll.head = n
			continue
		}
		// if the name is not already known, push to the head
		n := newNode(names[i])
		// if the list is empty, push to head and continue
		if ll.head == nil {
			ll.head = n
			ll.tail = n
			ll.nodes[names[i]] = n
			continue
		}
		// if the list is not empty push to head and update the head ref
		ll.head = n
		ll.tail.InsertAfter(n)
		ll.nodes[names[i]] = n
	} // all elements inserted after this scope exits
	// find any evictions to return
	result := []string{}
	// while the length of the LL is too long, pop the tail
	// and store in the list of evictions
	for len(ll.nodes) > ll.max {
		oldTail := ll.tail
		newTail := oldTail.Prev()
		oldTail.Pop()
		ll.tail = newTail
		result = append(result, oldTail.Name())
		delete(ll.nodes, oldTail.Name())
	}
	return result
}

func newLinkedList(maxSize int) *linkedList {
	return &linkedList{
		nodes: make(map[string]*node),
		max:   maxSize,
	}
}

type task struct {
	ctx     context.Context
	cancel  func()
	name    string
	fn      func(context.Context, interfaces.PeerLease) error
	cleanAt time.Time
}

type msgQueue struct {
	sync.RWMutex
	peer      interfaces.Peer
	lru       *linkedList
	next      chan *task
	cleanChan chan *task
	drainOnce sync.Once
	tasks     map[string]*task
	key       []byte
	draining  bool
	logger    *logrus.Logger
}

func newMsgQueue(max int, wc int, peer interfaces.Peer) (*msgQueue, error) {
	q := &msgQueue{
		lru:       newLinkedList(max),
		peer:      peer,
		next:      make(chan *task, max),
		cleanChan: make(chan *task, max+1+wc),
		tasks:     make(map[string]*task),
		key:       internalKey,
		logger:    logging.GetLogger(constants.LoggerPeerMan),
	}
	for i := 0; i < wc; i++ {
		go q.worker()
	}
	for i := 0; i < wc/2; i++ {
		go q.cleanupWorker()
	}
	return q, nil
}

func (mq *msgQueue) hash(msg []byte) string {
	hsh := highwayhash.Sum128(msg, mq.key)
	return string(hsh[:])
}

func (mq *msgQueue) Contains(msg []byte) bool {
	name := mq.hash(msg)
	mq.Lock()
	defer mq.Unlock()
	if mq.draining {
		return false
	}
	return mq.lru.Contains(name)
}

func (mq *msgQueue) Prevent(msg []byte) {
	name := mq.hash(msg)
	mq.Lock()
	defer mq.Unlock()
	if mq.draining {
		return
	}
	if mq.lru.Contains(name) {
		return
	}
	evicted := mq.lru.Push(name)
	for i := 0; i < len(evicted); i++ {
		t, ok := mq.tasks[evicted[i]]
		if ok {
			t.cancel()
			delete(mq.tasks, t.name)
		}
	}
	baseCtx := context.Background()
	ctx, cf := context.WithCancel(baseCtx)
	t := &task{
		ctx:     ctx,
		cancel:  cf,
		name:    name,
		cleanAt: time.Now().Add(constants.MsgTimeout * 5),
	}
	mq.cleanChan <- t
}

func (mq *msgQueue) Add(msg []byte, fn func(ctx context.Context, peer interfaces.PeerLease) error) {
	name := mq.hash(msg)
	mq.Lock()
	defer mq.Unlock()
	if mq.draining {
		return
	}
	if mq.lru.Contains(name) {
		return
	}
	evicted := mq.lru.Push(name)
	for i := 0; i < len(evicted); i++ {
		t, ok := mq.tasks[evicted[i]]
		if ok {
			t.cancel()
			delete(mq.tasks, t.name)
		}
	}
	baseCtx := context.Background()
	ctx, cf := context.WithTimeout(baseCtx, constants.MsgTimeout*4)
	t := &task{
		ctx:     ctx,
		cancel:  cf,
		name:    name,
		fn:      fn,
		cleanAt: time.Now().Add(constants.MsgTimeout * 5),
	}
	mq.cleanChan <- t
	mq.tasks[name] = t
	select {
	case mq.next <- t:
	case <-mq.peer.CloseChan():
	}
}

func (mq *msgQueue) cleanupWorker() {
	for {
		select {
		case t := <-mq.cleanChan:
			now := time.Now()
			if now.After(t.cleanAt) {
				mq.cleanup(t)
				continue
			}
			time.Sleep(t.cleanAt.Sub(now))
			mq.cleanup(t)
		case <-mq.peer.CloseChan():
			mq.drain()
			return
		}
	}
}

func (mq *msgQueue) cleanup(t *task) {
	mq.Lock()
	defer mq.Unlock()
	if mq.lru.Contains(t.name) {
		mq.lru.Pop(t.name)
	}
	t2, ok := mq.tasks[t.name]
	if ok {
		if t2 == t {
			t.cancel()
			delete(mq.tasks, t.name)
		}
	}
}

func (mq *msgQueue) drain() {
	mq.drainOnce.Do(func() {
		mq.Lock()
		mq.draining = true
		mq.Unlock()
		for {
			select {
			case t := <-mq.cleanChan:
				t.cancel()
				mq.Lock()
				if _, ok := mq.tasks[t.name]; ok {
					delete(mq.tasks, t.name)
					mq.lru.Pop(t.name)
				}
				mq.Unlock()
			default:
				mq.Lock()
				if len(mq.tasks) != 0 {
					mq.Unlock()
					continue
				}
				mq.Unlock()
				return
			}
		}
	})
}

func (mq *msgQueue) worker() {
	for {
		var currentTask *task
		select {
		case <-mq.peer.CloseChan():
			return
		case t := <-mq.next:
			select {
			case <-t.ctx.Done():
				mq.Lock()
				if _, ok := mq.tasks[t.name]; ok {
					delete(mq.tasks, t.name)
				}
				mq.Unlock()
				continue
			default:
				currentTask = t
			}
		}
		if currentTask != nil {
			mq.sendWithRetry(currentTask)
			mq.Lock()
			_, ok := mq.tasks[currentTask.name]
			if ok {
				delete(mq.tasks, currentTask.name)
			}
			mq.Unlock()
		}
	}
}

func (mq *msgQueue) sendWithRetry(t *task) {
	defer t.cancel()
	backoffCount := 0
	backoff := time.Duration(50 * time.Millisecond)
	maxBackoff := constants.MsgTimeout

	for {
		select {
		case <-mq.peer.CloseChan():
			return
		case <-t.ctx.Done():
			return
		default:
			if backoffCount > 0 {
				// exponential backoff with jitter
				jitter := time.Duration(uint8(time.Now().UnixNano()) % 128)
				jitterMS := jitter * time.Millisecond
				backoff = (backoff * 2) + jitterMS
				if backoff > maxBackoff {
					backoff = maxBackoff - jitterMS
				}
				select {
				case <-t.ctx.Done():
					return
				case <-time.After(backoff):
				}
			} else {
				backoffCount++
			}
			err := t.fn(t.ctx, mq.peer)
			if err != nil {
				if strings.Contains(err.Error(), "invalid") {
					return
				}
				utils.DebugTrace(mq.logger, err)
				continue
			}
			return
		}
	}
}
