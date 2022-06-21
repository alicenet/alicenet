package transport

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/types"
	"github.com/sirupsen/logrus"
)

var testWaitForClose = time.Second * 6

func TestMux(t *testing.T) {
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

	transport2, err := NewP2PTransport(logger, testCID, nodePrivKey2Hex, t2Port, t2Host)
	if err != nil {
		t.Fatal(err)
	}
	defer transport2.Close()
	nodeAddr2 := transport2.NodeAddr()

	mux := &P2PMux{}

	connDialChan := make(chan interfaces.P2PMuxConn)
	asyncDial := func(tp interfaces.P2PTransport, addr interfaces.NodeAddr) {
		conn, err := tp.Dial(addr, 1)
		if err != nil {
			t.Error(err)
			connDialChan <- nil
			return
		}
		if conn.Initiator() != types.SelfInitiatedConnection {
			t.Logf("Bad initiator: %d", conn.Initiator())
			t.Fail()
		}
		muxconn, err := mux.HandleConnection(context.TODO(), conn)
		if err != nil {
			t.Error(err)
			connDialChan <- nil
			return
		}
		if muxconn.Initiator() != types.SelfInitiatedConnection {
			t.Logf("Bad initiator: %d", conn.Initiator())
			t.Fail()
		}
		connDialChan <- muxconn
	}

	connAcceptChan := make(chan interfaces.P2PMuxConn)
	asyncAccept := func(tp interfaces.P2PTransport) {
		conn, err := tp.Accept()
		if err != nil {
			t.Error(err)
			connAcceptChan <- nil
			return
		}
		if conn.Initiator() != types.PeerInitiatedConnection {
			t.Logf("Bad initiator: %d", conn.Initiator())
			t.Fail()
		}
		muxconn, err := mux.HandleConnection(context.TODO(), conn)
		if err != nil {
			t.Error(err)
			connAcceptChan <- nil
			return
		}
		if muxconn.Initiator() != types.PeerInitiatedConnection {
			t.Logf("Bad initiator: got %d wanted %d", muxconn.Initiator(), types.PeerInitiatedConnection)
			t.Fail()
		}
		connAcceptChan <- muxconn
	}

	go asyncDial(transport1, nodeAddr2)
	go asyncAccept(transport2)

	t1mc := <-connDialChan
	if t1mc == nil {
		t.Fatal("t1 nil")
	}
	t2mc := <-connAcceptChan
	if t2mc == nil {
		t.Fatal("t2 nil")
	}

	mut1 := sync.Mutex{}
	asyncSend1 := func(wg *sync.WaitGroup, msg []byte) {
		defer wg.Done()
		mut1.Lock()
		defer mut1.Unlock()
		_, err := t1mc.ClientConn().Write(msg)
		if err != nil {
			t.Errorf("Error in asyncSend1 at Write: %v", err)
			return
		}
		b := make([]byte, len(msg))
		_, err = io.ReadFull(t2mc.ServerConn(), b)
		if err != nil {
			t.Errorf("Error in asyncSend1 at ReadFull: %v", err)
			return
		}
		if !bytes.Equal(msg, b) {
			t.Errorf("Error in asyncSend1: %s vs %s", msg, b)
		}
	}

	asyncSend2 := func(wg *sync.WaitGroup, msg []byte) {
		defer wg.Done()
		mut1.Lock()
		defer mut1.Unlock()
		_, err := t2mc.ClientConn().Write(msg)
		if err != nil {
			t.Errorf("Error in asyncSend2 at Write: %v", err)
			return
		}
		b := make([]byte, len(msg))
		_, err = io.ReadFull(t1mc.ServerConn(), b)
		if err != nil {
			t.Errorf("Error in asyncSend2 at ReadFull: %v", err)
			return
		}
		if !bytes.Equal(msg, b) {
			t.Errorf("Error in asyncSend2: %s vs %s", msg, b)
		}
	}

	// check async sends and recvs
	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go asyncSend1(wg, []byte(strings.Join([]string{"hi1", strconv.Itoa(i)}, ":")))
		wg.Add(1)
		go asyncSend2(wg, []byte(strings.Join([]string{"hi2", strconv.Itoa(i)}, ":")))
	}
	wg.Wait()
	// check close on transport close
	err = t1mc.Close()
	if err != nil {
		t.Fatal(err)
	}

	// check close on clientconn
	closeWatch1 := func(wg *sync.WaitGroup) {
		defer wg.Done()
		select {
		case <-t2mc.CloseChan():
		case <-time.After(testWaitForClose):
			t.Error("Error closing t2mc")
		}
	}

	closeWatch2 := func(wg *sync.WaitGroup) {
		defer wg.Done()
		select {
		case <-t2mc.CloseChan():
		case <-time.After(testWaitForClose):
			t.Error("Error closing t2mc")
		}
	}

	wg.Add(1)
	go closeWatch1(wg)
	wg.Add(1)
	go closeWatch2(wg)
	wg.Wait()

}

func TestMux2(t *testing.T) {
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

	transport1, err := NewP2PTransport(logger, testCID, nodePrivKey1Hex, 3002, t1Host)
	if err != nil {
		t.Fatal(err)
	}
	defer transport1.Close()

	transport2, err := NewP2PTransport(logger, testCID, nodePrivKey2Hex, 4002, t2Host)
	if err != nil {
		t.Fatal(err)
	}
	defer transport2.Close()
	nodeAddr2 := transport2.NodeAddr()

	mux := &P2PMux{}

	connDialChan := make(chan interfaces.P2PMuxConn)
	asyncDial := func(tp interfaces.P2PTransport, addr interfaces.NodeAddr) {
		conn, err := tp.Dial(addr, 1)
		if err != nil {
			t.Error(err)
			connDialChan <- nil
			return
		}
		muxconn, err := mux.HandleConnection(context.TODO(), conn)
		if err != nil {
			t.Error(err)
			connDialChan <- nil
			return
		}
		connDialChan <- muxconn
	}

	connAcceptChan := make(chan interfaces.P2PMuxConn)
	asyncAccept := func(tp interfaces.P2PTransport) {
		conn, err := tp.Accept()
		if err != nil {
			t.Error(err)
			connAcceptChan <- nil
			return
		}
		muxconn, err := mux.HandleConnection(context.TODO(), conn)
		if err != nil {
			t.Error(err)
			connAcceptChan <- nil
			return
		}
		connAcceptChan <- muxconn
	}

	go asyncDial(transport1, nodeAddr2)
	go asyncAccept(transport2)

	t1mc := <-connDialChan
	if t1mc == nil {
		t.Fatal("t1 nil")
	}
	t2mc := <-connAcceptChan
	if t2mc == nil {
		t.Fatal("t2 nil")
	}

	// check close on transport close
	err = t1mc.Close()
	if err != nil {
		t.Fatal(err)
	}
	wg := &sync.WaitGroup{}
	closeWatch1 := func(wg *sync.WaitGroup) {
		defer wg.Done()
		select {
		case <-t2mc.CloseChan():
		case <-time.After(testWaitForClose):
			t.Error("Error closing t2mc")
		}
	}

	closeWatch2 := func(wg *sync.WaitGroup) {
		defer wg.Done()
		select {
		case <-t2mc.CloseChan():
		case <-time.After(testWaitForClose):
			t.Error("Error closing t2mc")
		}
	}

	wg.Add(1)
	go closeWatch1(wg)
	wg.Add(1)
	go closeWatch2(wg)
	wg.Wait()

}

func TestMux3(t *testing.T) {
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

	transport1, err := NewP2PTransport(logger, testCID, nodePrivKey1Hex, 3003, t1Host)
	if err != nil {
		t.Fatal(err)
	}
	defer transport1.Close()

	transport2, err := NewP2PTransport(logger, testCID, nodePrivKey2Hex, 4003, t2Host)
	if err != nil {
		t.Fatal(err)
	}
	defer transport2.Close()
	nodeAddr2 := transport2.NodeAddr()

	mux := &P2PMux{}

	connDialChan := make(chan interfaces.P2PMuxConn)
	asyncDial := func(tp interfaces.P2PTransport, addr interfaces.NodeAddr) {
		conn, err := tp.Dial(addr, 1)
		if err != nil {
			t.Error(err)
			connDialChan <- nil
			return
		}
		muxconn, err := mux.HandleConnection(context.TODO(), conn)
		if err != nil {
			t.Error(err)
			connDialChan <- nil
			return
		}
		connDialChan <- muxconn
	}

	connAcceptChan := make(chan interfaces.P2PMuxConn)
	asyncAccept := func(tp interfaces.P2PTransport) {
		conn, err := tp.Accept()
		if err != nil {
			t.Error(err)
			connAcceptChan <- nil
			return
		}
		muxconn, err := mux.HandleConnection(context.TODO(), conn)
		if err != nil {
			t.Error(err)
			connAcceptChan <- nil
			return
		}
		connAcceptChan <- muxconn
	}

	go asyncDial(transport1, nodeAddr2)
	go asyncAccept(transport2)

	t1mc := <-connDialChan
	if t1mc == nil {
		t.Fatal("t1 nil")
	}
	t2mc := <-connAcceptChan
	if t2mc == nil {
		t.Fatal("t2 nil")
	}

	// check close on base conn
	err = t1mc.Close()
	if err != nil {
		t.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	closeWatch1 := func(wg *sync.WaitGroup) {
		defer wg.Done()
		select {
		case <-t2mc.CloseChan():
		case <-time.After(testWaitForClose):
			t.Error("Error closing t2mc")
		}
	}

	closeWatch2 := func(wg *sync.WaitGroup) {
		defer wg.Done()
		select {
		case <-t2mc.CloseChan():
		case <-time.After(testWaitForClose):
			t.Error("Error closing t2mc")
		}
	}
	wg.Add(1)
	go closeWatch1(wg)
	wg.Add(1)
	go closeWatch2(wg)
	wg.Wait()

}
