package ipc

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/MadBase/MadNet/config"
	"github.com/MadBase/MadNet/test/testutils"
	"github.com/sirupsen/logrus/hooks/test"
)

type Msg struct {
	Bytes json.RawMessage
	Err   error
}

func (m *Msg) String() string {
	return `{Bytes:` + string(m.Bytes) + `, Err:` + m.Err.Error() + `}`
}

func newTestDialer(socketFile string) (func() []Msg, func(b []byte) error, func()) {
	msgs := make([]Msg, 0)
	mu := sync.Mutex{}

	var conn *net.UnixConn

	go func() {
		testutils.WaitUntil(func() bool {
			var err error
			conn, err = net.DialUnix("unix", nil, &net.UnixAddr{Name: socketFile, Net: "unix"})
			return err == nil
		})

		dec := json.NewDecoder(conn)
		for {
			var r json.RawMessage
			err := dec.Decode(&r)
			if err == io.EOF {
				break
			}

			mu.Lock()
			msgs = append(msgs, Msg{r, err})
			mu.Unlock()
		}
	}()

	return func() []Msg {
			mu.Lock()
			ret := make([]Msg, len(msgs))
			copy(ret, msgs)
			mu.Unlock()
			return ret
		}, func(b []byte) error {
			if conn == nil {
				return fmt.Errorf("no conn")
			}
			_, err := conn.Write(b)
			return err
		}, func() {
			if conn != nil {
				conn.Close()
			}
		}
}

func TestA(t *testing.T) {
	config.Configuration.Firewalld.Enabled = true

	socketFile := testutils.SocketFileName()
	defer os.Remove(socketFile)

	s := NewServer(socketFile)
	l, _ := test.NewNullLogger()
	s.logger = l

	getMsgs, write, end := newTestDialer(socketFile)
	defer end()

	errchan := testutils.TestAsync(func() error {
		defer s.Close()

		var err error
		testutils.WaitUntil(func() bool {
			err = s.Push(PeersUpdate{Addrs: []string{"11.22.33.44:55", "33.44.55.66:77", "55.66.77.88:99"}, Seq: 0})
			return err != ErrNoConnection
		})
		if err != ErrNotSubscribed {
			return fmt.Errorf("expected ErrNotSubscribed instead got: %v", err)
		}

		time.Sleep(100 * time.Millisecond)
		msgs := getMsgs()

		if len(msgs) != 0 {
			return fmt.Errorf("Expected 0 messages, instead got %v", msgs)
		}

		write([]byte(`{"jsonrpc":"2.0","id":"sub","method":"subscribe"}`))
		time.Sleep(100 * time.Millisecond)
		err = s.Push(PeersUpdate{Addrs: []string{"11.22.33.44:55", "33.44.55.66:77", "55.66.77.88:99"}, Seq: 0})
		if err != nil {
			return fmt.Errorf("push err: %T %v\n", err, err)
		}

		testutils.WaitUntil(func() bool { msgs = getMsgs(); return len(msgs) >= 1 })
		if len(msgs) != 1 {
			return fmt.Errorf("Expected 1 messages, instead got %v", msgs)
		}
		if msgs[0].Err != nil || string(msgs[0].Bytes) != `{"jsonrpc":"2.0","id":"sub","result":{"Addrs":["11.22.33.44:55","33.44.55.66:77","55.66.77.88:99"],"Seq":0}}` {
			return fmt.Errorf("Message 1 does not match expectation %v", msgs[0].Bytes)
		}

		return nil
	})

	err := s.Start()
	if err != nil {
		t.Fatal("Unexpected err from server:", err)
	}

	err = <-errchan
	if err != nil {
		t.Fatal("Unexpected err:", err)
	}

}
