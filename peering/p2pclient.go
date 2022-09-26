package peering

import (
	"github.com/sirupsen/logrus"

	"github.com/alicenet/alicenet/interfaces"
)

type p2PClient struct {
	logger *logrus.Logger
	interfaces.P2PClientRaw
	nodeAddr interfaces.NodeAddr
	conn     interfaces.P2PMuxConn
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
