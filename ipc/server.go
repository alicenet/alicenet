package ipc

import (
	"encoding/json"
	"io"
	"net"
	"os"
	"time"

	"github.com/MadBase/MadNet/config"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/logging"
	"github.com/sirupsen/logrus"
)

type Server struct {
	address string
	logger  *logrus.Logger

	conn *net.UnixConn
	lis  *net.UnixListener

	writes chan write
	pushId interface{}

	close  chan struct{}
	closed bool
}

type write struct {
	message   interface{}
	errorchan chan error
}

type PeersUpdate struct {
	Addrs []string
	Seq   uint
}

func NewServer(address string) *Server {
	return &Server{
		address: address,
		logger:  logging.GetLogger(constants.LoggerIPC),
		writes:  make(chan write, 2),
		close:   make(chan struct{}, 2),
	}
}

func (s *Server) Close() {
	s.closed = true
	if s.lis != nil {
		s.lis.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}
	s.close <- struct{}{}
}

func (s *Server) handle(msg interface{}) interface{} {
	switch data := msg.(type) {

	case []interface{}:

		responses := make([]interface{}, 0)
		for _, req := range data {
			response := s.handle(req)
			if response != nil {
				responses = append(responses, response)
			}
		}
		return responses

	case map[string]interface{}:

		method, ok := data["method"].(string)
		jsonrpc, _ := data["jsonrpc"].(string)
		if !ok || jsonrpc != "2.0" {
			return newInvalidRequestResponse(data["id"])
		}

		req := request{"2.0", data["id"], data["params"], method}

		fn := serverMethods[method]
		if fn == nil {
			return newMethodNotFoundResponse(req.Id)
		}

		result, err := fn(s, req)
		if data["id"] == nil || (result == nil && err == nil) {
			return nil
		}
		return response{"2.0", req.Id, result, err}

	default:
		return newInvalidRequestResponse(nil)
	}
}

func (s *Server) startWriter() {
	for {
		select {
		case w := <-s.writes:
			b, err := json.Marshal(w.message)
			if err != nil {
				if w.errorchan != nil {
					w.errorchan <- err
				} else {
					s.logger.Errorf("error marshalling response: %T %v\n", err, err)
				}
				continue
			}

			s.conn.SetWriteDeadline(time.Now().Add(time.Second))
			_, err = s.conn.Write(b)

			if err != nil {
				if _, ok := err.(*net.OpError); ok {
					err = ErrNoConnection
				}
				if w.errorchan != nil {
					w.errorchan <- err
				} else {
					s.logger.Errorf("write error: %T %v", err, err)
				}
				continue
			}

			if w.errorchan != nil {
				w.errorchan <- nil
			}

		case <-s.close:
			return
		}
	}
}

func (s *Server) Start() error {
	if !config.Configuration.Firewalld.Enabled {
		return nil
	}

	s.logger.Info("server started")
	os.Remove(s.address)

	var err error
	s.lis, err = net.ListenUnix("unix", &net.UnixAddr{Name: s.address, Net: "unix"})
	if err != nil {
		return err
	}

	go s.startWriter()

	for !s.closed {
		var err error
		s.conn, err = s.lis.AcceptUnix()
		if err != nil {
			continue
		}
		s.logger.Info("client connected")

		func() {
			defer func() {
				s.conn.Close()
				s.pushId = nil
			}()

			dec := json.NewDecoder(s.conn)
			var msg interface{}
			for {
				err := dec.Decode(&msg)

				if err == io.EOF {
					s.logger.Info("client disconnected")
					return
				}

				switch v := err.(type) {
				case nil:
					response := s.handle(msg)
					if response != nil {
						s.writes <- write{response, nil}
					}

				case *json.UnmarshalTypeError, *json.SyntaxError, *json.InvalidUnmarshalError:
					s.logger.Warnf("error unmarshalling request %v: %T %v\n", msg, err, err)
					s.writes <- write{newParseErrorResponse(err.Error()), nil}
					dec = json.NewDecoder(s.conn)

				case *net.OpError:
					if !s.closed {
						s.logger.Errorf("connection error", v.Unwrap())
					} else {
						s.logger.Info("server closed")
					}
					return

				default:
					s.logger.Errorf("fatal error, closing connection: %T %v\n", err, err)
					return
				}
			}
		}()
	}

	return nil
}

func (s *Server) Push(u PeersUpdate) (ret error) {
	if !config.Configuration.Firewalld.Enabled {
		return nil
	}

	if s.conn == nil {
		return ErrNoConnection
	}
	if s.pushId == nil {
		return ErrNotSubscribed
	}

	c := make(chan error, 1)

	s.writes <- write{response{"2.0", s.pushId, u, nil}, c}
	return <-c
}
