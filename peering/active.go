package peering

import (
	"sync"

	"github.com/MadBase/MadNet/interfaces"
)

// activePeerStore stores active peers
type activePeerStore struct {
	sync.RWMutex
	// canClose tracks the close permission of the store.
	// this should only ever be set on the peer store of the
	// peer manager itself
	canClose bool
	// store tracks active connections
	store map[string]interfaces.P2PClient
	// pid map provides probabilistic protection against cleanup races
	pid map[string]uint64
	// closeChan causes a shutdown of service
	closeChan chan struct{}
	// protect double closure
	closeOnce sync.Once
}

func (ps *activePeerStore) close() {
	ps.closeOnce.Do(func() {
		close(ps.closeChan)
	})
}

// cleanup the store on closure of peer connections
func (ps *activePeerStore) onExit(pid uint64, c interfaces.P2PClient) {
	select {
	case <-c.CloseChan():
		ps.Lock()
		defer ps.Unlock()
		if ps.pid[c.NodeAddr().Identity()] == pid {
			delete(ps.store, c.NodeAddr().Identity())
			delete(ps.pid, c.NodeAddr().Identity())
		}
		return
	case <-ps.closeChan:
		if ps.canClose {
			c.Close()
		}
		return
	}
}

// add a peer to the store
func (ps *activePeerStore) add(c interfaces.P2PClient) {
	ps.Lock()
	defer ps.Unlock()
	stale, ok := ps.store[c.NodeAddr().Identity()]
	if ok {
		select {
		case <-stale.CloseChan():
			delete(ps.store, stale.NodeAddr().Identity())
			delete(ps.pid, stale.NodeAddr().Identity())
		default:
			if ps.canClose {
				c.Close()
			}
			return
		}
	}
	pid := makePid()
	ps.store[c.NodeAddr().Identity()] = c
	ps.pid[c.NodeAddr().Identity()] = pid
	go ps.onExit(pid, c)
}

// remove a peer from the store
func (ps *activePeerStore) del(c interfaces.NodeAddr) {
	ps.Lock()
	defer ps.Unlock()
	obj, ok := ps.store[c.Identity()]
	if ok {
		if ps.canClose {
			obj.Close()
		}
		delete(ps.store, c.Identity())
		delete(ps.pid, c.Identity())
	}
}

//nolint:unused,deadcode
func (ps *activePeerStore) get(c interfaces.NodeAddr) (interfaces.P2PClient, bool) {
	ps.RLock()
	defer ps.RUnlock()
	cc, ok := ps.store[c.Identity()]
	return cc, ok
}

func (ps *activePeerStore) contains(c interfaces.NodeAddr) bool {
	ps.Lock()
	defer ps.Unlock()
	_, ok := ps.store[c.Identity()]
	return ok
}

func (ps *activePeerStore) len() int {
	ps.RLock()
	defer ps.RUnlock()
	return len(ps.store)
}

// random returns a random active peer
func (ps *activePeerStore) random() (string, bool) {
	ps.RLock()
	defer ps.RUnlock()
	if len(ps.store) == 0 {
		return "", false
	}
	i := 0
	index, err := randomElement(len(ps.store))
	if err != nil {
		return "", false
	}
	for _, v := range ps.store {
		if i == index {
			return v.NodeAddr().P2PAddr(), true
		}
		i++
	}
	panic("unreachable")
}

// random returns a random active peer
//nolint:unused,deadcode
func (ps *activePeerStore) randomClient() (interfaces.P2PClient, bool) {
	ps.RLock()
	defer ps.RUnlock()
	if len(ps.store) == 0 {
		return nil, false
	}
	i := 0
	index, err := randomElement(len(ps.store))
	if err != nil {
		return nil, false
	}
	for _, v := range ps.store {
		if i == index {
			return v, true
		}
		i++
	}
	panic("unreachable")
}

// getPeers returns the set of active peers
//nolint:unused,deadcode
func (ps *activePeerStore) getPeers() ([]interfaces.P2PClient, bool) {
	ps.RLock()
	defer ps.RUnlock()
	if len(ps.store) == 0 {
		return nil, false
	}
	result := []interfaces.P2PClient{}
	for _, v := range ps.store {
		result = append(result, v)
	}
	return result, true
}
