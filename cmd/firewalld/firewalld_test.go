package firewalld

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"path"
	"sync"
	"testing"

	"github.com/MadBase/MadNet/cmd/firewalld/lib"
	"github.com/MadBase/MadNet/cmd/firewalld/mock"
	"github.com/MadBase/MadNet/test/testutils"
	"github.com/sirupsen/logrus/hooks/test"
)

var logger, _ = test.NewNullLogger()

func newTestServer(address string) (getMsg func() []mock.Msg, write func([]byte) error, close func()) {
	rcv := make([]mock.Msg, 0)
	mu := sync.Mutex{}
	rcvAppend := func(m mock.Msg) {
		mu.Lock()
		rcv = append(rcv, m)
		mu.Unlock()
	}

	var lis *net.UnixListener
	var conn *net.UnixConn
	go func() {

		lis, err := net.ListenUnix("unix", &net.UnixAddr{Name: address, Net: "unix"})
		if err != nil {
			rcvAppend(mock.Msg{Bytes: nil, Err: err})
			return
		}
		defer lis.Close()

		conn, err = lis.AcceptUnix()
		if err != nil {
			rcvAppend(mock.Msg{Bytes: nil, Err: err})
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
			rcvAppend(mock.Msg{Bytes: bytes, Err: err})
		}
	}()

	return func() []mock.Msg {
			mu.Lock()
			ret := make([]mock.Msg, len(rcv))
			copy(ret, rcv)
			mu.Unlock()
			return ret
		},
		func(b []byte) error {
			if conn == nil {
				return fmt.Errorf("no conn")
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

func TestSocket(t *testing.T) {
	dir := t.TempDir()
	socketFile := path.Join(dir, "socket")
	getMsgs, write, end := newTestServer(socketFile)
	im := mock.NewImplementation()

	errchan := testutils.TestAsync(func() error {
		defer end()

		testutils.WaitUntil(func() bool { return len(getMsgs()) > 0 })

		msgs := getMsgs()
		if msgs[0].Err != nil || string(msgs[0].Bytes) != `{"jsonrpc":"2.0","id":"sub","method":"subscribe"}` {
			return fmt.Errorf("Unexpected messages: %v", msgs)
		}

		err := write([]byte(`{"jsonrpc":"2.0","id":"sub","result":{"Addrs":["11.22.33.44:5555","22.33.44.55:6666","33.44.55.66:7777"],"Seq":0}}`))
		if err != nil {
			return fmt.Errorf("Error writing: %v", err)
		}
		var calls mock.Calls
		testutils.WaitUntil(func() bool { calls = im.Calls(); return len(calls.Update) >= 1 })

		if calls.Get != 1 {
			return fmt.Errorf("Expected 1 get call, instead got %v", calls.Get)
		}
		if len(calls.Update) != 1 {
			return fmt.Errorf("Expected 1 update call, instead got %v", len(calls.Update))
		}
		if !calls.Update[0].ToAdd.Equal(lib.NewAddresSet([]string{"11.22.33.44:5555", "22.33.44.55:6666", "33.44.55.66:7777"})) ||
			!calls.Update[0].ToDelete.Equal(lib.NewAddresSet([]string{})) {
			return fmt.Errorf("Update call 1 does not match expectation %v", calls.Update)
		}

		im.GetRet = mock.GetRet{Ret: lib.NewAddresSet([]string{"22.33.44.55:6666"})}
		err = write([]byte(`{"jsonrpc":"2.0","id":"sub","result":{"Addrs":["11.22.33.44:5555","22.33.44.55:6666","33.44.55.66:7777"],"Seq":0}}`))
		if err != nil {
			return fmt.Errorf("Error writing: %v", err)
		}
		testutils.WaitUntil(func() bool { calls = im.Calls(); return len(calls.Update) >= 2 })

		if calls.Get != 2 {
			return fmt.Errorf("Expected 2 get call, instead got %v", calls.Get)
		}
		if len(calls.Update) != 2 {
			return fmt.Errorf("Expected 2 update call, instead got %v", len(calls.Update))
		}
		if !calls.Update[1].ToAdd.Equal(lib.NewAddresSet([]string{"11.22.33.44:5555", "33.44.55.66:7777"})) ||
			!calls.Update[1].ToDelete.Equal(lib.NewAddresSet([]string{})) {
			return fmt.Errorf("Update call 2 does not matchh expectation %v", calls.Update[1])
		}

		return nil
	})

	var err error
	testutils.WaitUntil(func() bool {
		err = startConnection(logger, socketFile, im)
		return err != ErrNoConn
	})

	if err != io.EOF {
		t.Fatal("Expected EOF, instead got err:", err)
	}

	err = <-errchan
	if err != nil {
		t.Fatal("Unexpected err:", err)
	}
}
