package firewalld

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"

	"github.com/MadBase/MadNet/cmd/firewalld/lib"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus/hooks/test"
)

var logger, _ = test.NewNullLogger()

type msg struct {
	bytes json.RawMessage
	err   error
}

type updateCall struct {
	toAdd    lib.AddressSet
	toDelete lib.AddressSet
}
type getRet struct {
	ret lib.AddressSet
	err error
}
type mockImplementation struct {
	getCalls    int
	updateCalls []updateCall
	getRet      getRet
	updateRet   error
}

func newMockImplementation() *mockImplementation {
	return &mockImplementation{updateCalls: make([]updateCall, 0)}
}
func (mi *mockImplementation) GetAllowedAddresses() (lib.AddressSet, error) {
	mi.getCalls++
	return mi.getRet.ret, mi.getRet.err
}
func (mi *mockImplementation) UpdateAllowedAddresses(toAdd lib.AddressSet, toDelete lib.AddressSet) error {
	mi.updateCalls = append(mi.updateCalls, updateCall{toAdd, toDelete})
	return mi.updateRet
}

func newAddress() string {
	rand.Seed(time.Now().Unix())
	return "/tmp/madnet-firewalld-test-" + fmt.Sprint(rand.Int())
}

func newTestListener(address string) (getMsg func() []msg, write func([]byte) error, close func()) {
	received := make([]msg, 0)

	var lis *net.UnixListener
	var conn *net.UnixConn
	go func() {

		lis, err := net.ListenUnix("unix", &net.UnixAddr{Name: address, Net: "unix"})
		if err != nil {
			received = append(received, msg{nil, err})
			return
		}
		defer lis.Close()

		conn, err = lis.AcceptUnix()
		if err != nil {
			received = append(received, msg{nil, err})
			return
		}
		defer conn.Close()

		dec := json.NewDecoder(conn)

		for {
			var bytes json.RawMessage
			err := dec.Decode(&bytes)
			if err == io.EOF {
				break
			}
			received = append(received, msg{bytes, err})
		}
	}()

	return func() []msg {
			return received
		},
		func(b []byte) error {
			if conn == nil {
				return errors.Errorf("no conn")
			}
			_, err := conn.Write(b)
			return err
		}, func() {
			if lis != nil {
				lis.Close()
			}
			if conn != nil {
				conn.Close()
			}
		}
}

func waitUntil(f func() bool) {
	end := time.Now().Add(time.Second)
	for time.Now().Before(end) {
		time.Sleep(time.Millisecond)
		if f() {
			time.Sleep(time.Millisecond)
			return
		}
	}
	panic("timeout")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func TestAA(t *testing.T) {
	socketFile := newAddress()
	getMsgs, _, end := newTestListener(socketFile)
	im := newMockImplementation()

	waitUntil(func() bool { return fileExists(socketFile) })

	c := make(chan error)
	go func() {
		defer end()
		defer close(c)

		waitUntil(func() bool { return len(getMsgs()) > 0 })

		msgs := getMsgs()
		if msgs[0].err != nil || string(msgs[0].bytes) != `{"jsonrpc":"2.0","id":"sub","method":"subscribe"}` {
			c <- fmt.Errorf("Unexpected messages: %v", msgs)
			return
		}
	}()

	time.Sleep(2 * time.Second)
	err := startConnection(logger, socketFile, im)
	if err != io.EOF {
		t.Fatal("Expected EOF, instead got err:", err)
	}

	err = <-c
	if err != nil {
		t.Fatal("Unexpected err ", err)
	}
}
