package transport

import (
	"io"
	"net"
	"strconv"
	"testing"

	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	testCID     types.ChainIdentifier = 13
	testCIDFail types.ChainIdentifier = 31
	t1Host      string                = "127.0.0.1"
	t1Port      int                   = 3000
	t2Host      string                = "127.0.0.1"
	t2Port      int                   = 4000
)

func TestTransportsuccess(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	nodePrivKey1, err := newTransportPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	nodePrivKey1Hex := serializeTransportPrivateKey(nodePrivKey1)

	nodePrivKey2, err := newTransportPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	nodePrivKey2Hex := serializeTransportPrivateKey(nodePrivKey2)

	transport1, err := NewP2PTransport(logger, testCID, nodePrivKey1Hex, t1Port, t1Host)
	if err != nil {
		t.Fatal(err)
	}
	defer transport1.Close()
	nodeAddr1 := transport1.NodeAddr()

	transport2, err := NewP2PTransport(logger, testCID, nodePrivKey2Hex, t2Port, t2Host)
	if err != nil {
		t.Fatal(err)
	}
	defer transport2.Close()

	t1addrtest := transport1.NodeAddr().String()
	assert.Equal(t, net.JoinHostPort(t1Host, strconv.Itoa(t1Port)), t1addrtest, "Transport Addr method does not return address as expected.")

	complete1 := make(chan struct{})
	complete2 := make(chan struct{})
	complete3 := make(chan struct{})

	go dialer(t, transport2, nodeAddr1, complete1) //nolint:govet,staticcheck
	go acceptWithResp(t, transport1, complete2)    //nolint:govet,staticcheck
	go accept(t, transport2, complete3)

	<-complete1
	err = transport1.Close()
	if err != nil {
		t.Fatal(err)
	}
	<-complete2
	<-complete3
}

func dialer(t *testing.T, transport interfaces.P2PTransport, addr interfaces.NodeAddr, complete chan struct{}) {
	defer close(complete)
	defer transport.Close()
	conn, err := transport.Dial(addr, 1)
	if err != nil {
		t.Error(err)
		return
	}
	str := "test"
	_, err = conn.Write([]byte(str))
	if err != nil {
		t.Error(err)
		panic(err)
	}
	buf := make([]byte, 4)
	_, err2 := io.ReadFull(conn, buf)
	if err2 != nil {
		t.Error(err)
	}
	assert.Equal(t, str, string(buf), "Recv'd message does not match sent.")
	err2 = transport.Close()
	if err2 != nil {
		t.Error(err)
	}
}

func accept(t *testing.T, transport interfaces.P2PTransport, complete chan struct{}) {
	defer close(complete)
	_, err := transport.Accept()
	if err != nil {
		return
	}
	t.Error("Got a connection that should not have occurred.")
}

func acceptWithResp(t *testing.T, transport interfaces.P2PTransport, complete chan struct{}) {
	defer close(complete)
	conn, err := transport.Accept()
	if err != nil {
		t.Error(err)
		return
	}
	buf := make([]byte, 4)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		t.Error(err)
	}
	_, err = conn.Write(buf)
	if err != nil {
		t.Error(err)
		panic(err)
	}
}
