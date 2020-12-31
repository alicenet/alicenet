package peering

import (
	"context"
	"errors"
	"sync"

	"github.com/MadBase/MadNet/interfaces"
	"github.com/sirupsen/logrus"
)

// PeerSubscription allows a remote service to maintain a reference to the
// active peer set. This reference will be kept in sync with the local copy of
// the active peer set and allows external services to perform P2P RPC with
// the active peers set.
type PeerSubscription struct {
	sync.RWMutex
	log        *logrus.Logger
	clientChan chan interfaces.P2PClient
	closeChan  chan struct{}
	actives    *activePeerStore
	closeOnce  sync.Once
}

// CloseChan returns a channel that will be closed when the subscription
// *BEGINS* shutdown
func (p *PeerSubscription) CloseChan() <-chan struct{} {
	return p.closeChan
}

func (p *PeerSubscription) close() {
	p.closeOnce.Do(func() {
		close(p.closeChan)
	})
}

func (p *PeerSubscription) listen() {
	for {
		select {
		case c := <-p.clientChan:
			func() {
				p.Lock()
				defer p.Unlock()
				p.actives.add(c)
			}()
		case <-p.closeChan:
			return
		}
	}
}

func (p *PeerSubscription) add(c interfaces.P2PClient) {
	select {
	case p.clientChan <- c:
		return
	case <-p.closeChan:
		return
	}
}

// RequestLease returns an active peer that has sent the requested msg
func (p *PeerSubscription) RequestLease(ctx context.Context, msg []byte) (interfaces.PeerLease, error) {
	p.RLock()
	defer p.RUnlock()
	peers, ok := p.actives.getPeers()
	if ok {
		for i := 0; i < len(peers); i++ {
			client, ok := peers[i].(*p2PClient)
			if ok {
				if client.Contains(msg) {
					return client, nil
				}
			}
		}
	}
	return p.PeerLease(ctx)
}

// PeerLease returns a random active peer
func (p *PeerSubscription) PeerLease(ctx context.Context) (interfaces.PeerLease, error) {
	p.RLock()
	defer p.RUnlock()
	peers, ok := p.actives.getPeers()
	if ok {
		index, err := randomElement(len(peers))
		if err != nil {
			p.log.Debugf("Error in PeerSubscription.PeerLease at randomElement: %v", err)
			return nil, err
		}
		client, ok := peers[index].(*p2PClient)
		if ok {
			return client, nil
		}
	}
	return nil, errors.New("p2pclient is nil")
}

// GossipTx allows a service to Gossip a transaction all active peers
func (p *PeerSubscription) GossipTx(hsh []byte, fn func(context.Context, interfaces.PeerLease) error) {
	select {
	case <-p.CloseChan():
		fn(nil, &peerFail{})
	default:
	}
	peers, ok := p.actives.getPeers()
	if !ok {
		fn(nil, &peerFail{})
		return
	}
	for _, peer := range peers {
		client := peer.(*p2PClient)
		go func(c *p2PClient) {
			c.GossipTx(hsh, fn)
		}(client)
	}
}

// GossipConsensus allows a service to Gossip a consensus message to all
// active peers
func (p *PeerSubscription) GossipConsensus(hsh []byte, fn func(context.Context, interfaces.PeerLease) error) {
	select {
	case <-p.CloseChan():
		fn(nil, &peerFail{})
	default:
	}
	peers, ok := p.actives.getPeers()
	if !ok {
		fn(nil, &peerFail{})
		return
	}
	for _, peer := range peers {
		client := peer.(*p2PClient)
		go func(c *p2PClient) {
			c.GossipConsensus(hsh, fn)
		}(client)
	}
}

// PreventGossipConsensus allows a peer to be marked as having knowledge of a
// message. This will prevent the peer from being spammed by stale gossip as
// well as allow the local node to request any associated data directly from the
// providing peer.
func (p *PeerSubscription) PreventGossipConsensus(addr interfaces.NodeAddr, hsh []byte) {
	p.RLock()
	defer p.RUnlock()
	peer, ok := p.actives.get(addr)
	if ok {
		client, ok := peer.(*p2PClient)
		if ok {
			client.PreventGossipConsensus(hsh)
		}
	}
}

// PreventGossipTx allows a peer to be marked as having knowledge of a
// message. This will prevent the peer from being spammed by stale gossip as
// well as allow the local node to request any associated data directly from the
// providing peer.
func (p *PeerSubscription) PreventGossipTx(addr interfaces.NodeAddr, hsh []byte) {
	p.RLock()
	defer p.RUnlock()
	peer, ok := p.actives.get(addr)
	if ok {
		client, ok := peer.(*p2PClient)
		if ok {
			client.PreventGossipTx(hsh)
		}
	}
}

type peerFail struct{}

func (p *peerFail) P2PClient() (interfaces.P2PClient, error) {
	return nil, errors.New("unable to create peer lease")
}

func (p *peerFail) Do(fn func(interfaces.PeerLease) error) {
	fn(p)
}
