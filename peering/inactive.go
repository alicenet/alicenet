package peering

import (
	"fmt"
	"sync"
	"time"

	"github.com/alicenet/alicenet/interfaces"
)

// inactivePeerStore tracks inactive peers
type inactivePeerStore struct {
	sync.RWMutex
	// map of peer p2p addr and time last added
	store    map[string]interfaces.NodeAddr
	cooldown map[string]uint64
	// close broadcast channel
	closeChan chan struct{}
	// protect double closure
	closeOnce sync.Once
}

// close the store
func (ps *inactivePeerStore) close() {
	ps.closeOnce.Do(func() {
		close(ps.closeChan)
	})
}

// add a peer to the store
func (ps *inactivePeerStore) add(c interfaces.NodeAddr) {
	ps.Lock()
	defer ps.Unlock()
	_, ok := ps.cooldown[c.Identity()]
	if !ok {
		ps.store[c.Identity()] = c
	}
}

// delete a peer
func (ps *inactivePeerStore) del(c interfaces.NodeAddr) {
	ps.Lock()
	defer ps.Unlock()
	_, ok := ps.store[c.Identity()]
	if ok {
		delete(ps.store, c.Identity())
		ps.backoff(c)
	}
}

// delete a peer
func (ps *inactivePeerStore) backoff(c interfaces.NodeAddr) {
	pid := makePid()
	ps.cooldown[c.Identity()] = pid
	go func() {
		// Random delay between 20 and 40 seconds
		// Note: 20 > 0 so no error will occur
		delay, _ := randomElement(20)
		delay = delay + 20
		time.Sleep(time.Second * time.Duration(delay))
		ps.Lock()
		defer ps.Unlock()
		pidLater := ps.cooldown[c.Identity()]
		if pidLater == pid {
			delete(ps.cooldown, c.Identity())
		}
	}()
}

// get a random peer and remove it from the store
// use this method for getting peers to dial
func (ps *inactivePeerStore) randomPop() (interfaces.NodeAddr, bool) {
	ps.Lock()
	defer ps.Unlock()
	if len(ps.store) == 0 {
		return nil, false
	}
	i := 0
	index, err := randomElement(len(ps.store))
	if err != nil {
		return nil, false
	}
	for k, v := range ps.store {
		if i == index {
			delete(ps.store, k)
			ps.backoff(v)
			return v, true
		}
		i++
	}
	panic(fmt.Sprintf("unreachable index with %d nodes: %d", len(ps.store), index))
}

// get a random peer. intended to provide random peers when a remote
// peer performs a discovery dial against the local node
func (ps *inactivePeerStore) random() (string, bool) {
	ps.RLock()
	defer ps.RUnlock()
	if len(ps.store) == 0 {
		return "", false
	}
	index, err := randomElement(len(ps.store))
	if err != nil {
		return "", false
	}
	i := 0
	for _, v := range ps.store {
		if i == index {
			return v.P2PAddr(), true
		}
		i++
	}
	panic(fmt.Sprintf("unreachable index with %d nodes: %d", len(ps.store), index))
}

func (ps *inactivePeerStore) len() int {
	ps.RLock()
	defer ps.RUnlock()
	return len(ps.store)
}
