package peering

import (
	"context"

	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/utils"
	"github.com/sirupsen/logrus"
)

type p2PClient struct {
	logger *logrus.Logger
	interfaces.P2PClientRaw
	nodeAddr       interfaces.NodeAddr
	conn           interfaces.P2PMuxConn
	consensusQueue *msgQueue
	txQueue        *msgQueue
}

func (c *p2PClient) Close() error {
	return c.conn.Close()
}

func (c *p2PClient) CloseChan() <-chan struct{} {
	return c.conn.CloseChan()
}

func (c *p2PClient) NodeAddr() interfaces.NodeAddr {
	return c.nodeAddr
}

func (c *p2PClient) Do(fn func(interfaces.PeerLease) error) {
	err := fn(c)
	if err != nil {
		utils.DebugTrace(c.logger, err)
	}
}

func (c *p2PClient) Contains(msg []byte) bool {
	if c.consensusQueue.Contains(msg) {
		return true
	}
	if c.txQueue.Contains(msg) {
		return true
	}
	return false
}

func (c *p2PClient) PreventGossipConsensus(msg []byte) {
	c.consensusQueue.Prevent(msg)
}

func (c *p2PClient) PreventGossipTx(msg []byte) {
	c.txQueue.Prevent(msg)
}

func (c *p2PClient) GossipConsensus(hsh []byte, fn func(context.Context, interfaces.PeerLease) error) {
	c.consensusQueue.Add(hsh, fn)
}

func (c *p2PClient) GossipTx(hsh []byte, fn func(context.Context, interfaces.PeerLease) error) {
	c.txQueue.Add(hsh, fn)
}

func (c *p2PClient) P2PClient() (interfaces.P2PClient, error) {
	return c, nil
}
