package rbus

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestRBus(t *testing.T) {
	// Setup the bus
	rb := NewRBus()
	// test delayed send of message to help with startup races
	// this message is sent before the handler even exists, but
	// the handler still gets it anyway due to a long timeout
	go testTimeoutDelayed(t, rb)
	// test instant failure on deadlock condition
	// this ia a handler that has no timeout set and also
	// calls a service that does not exist as of yet
	// this is either a deadlock or a memory leak if the
	// wrong service name was called, so it just errors out
	_, err := rb.Request(SNAME, 0, 1*time.Second)
	if err == nil {
		t.Error("should have gotten a failure but did not")
	}
	if err != ErrUnknownService {
		t.Error(err)
	}
	// setup the handler and register the service with
	// the request bus
	stopFunc := setupHandler(t, rb)
	// test the prevention of double registration
	// it is an undefined behavior to have two listeners
	// in a system that may need a dedicated consumer
	// so we prevent double registration
	_, err = rb.Register(SNAME, 0)
	if err == nil {
		t.Error("double registration did not fail")
	}
	if err != ErrReservation {
		t.Error("did not get back correct error for double registration")
	}
	// coordinate the cleanup of the service handler
	defer stopFunc()
	// test a short timeout that should fail on timeout
	go testBlockingShort(t, rb)
	// this is a test for making sure the test handler
	// is doing what it should. Has nothing to do with the
	// actual request bus code.
	go testErr(t, rb, 32)
	// test the int response for mixups of messages
	// also check backpressure timeouts.
	// A backpressure timeout happens when the listener
	// drops before the responder has a chance to begin
	// processing. This can be caught by checking the
	// timeout before write to save on long computation
	// see the complex handler for details.
	for i := 0; i < 1000; i++ {
		go testTimeoutBackpressure(t, rb)
		go testInt(t, rb, i)
	}
	// test the timeout logic using a long delay
	go testTimeoutDelayed(t, rb)
	// test the blocking mode of operation with no timeout
	// this is only safe if the remote service is guaranteed
	// to exist before write
	// this is also used to give all previous tests time
	// to finish before exit
	testBlockingLong(t, rb)
}

const (
	SNAME = "foo"
)

type testHandler struct {
	sync.WaitGroup
	closeOnce sync.Once
	closeChan chan struct{}
	rchan     <-chan Request
}

func (th *testHandler) Start() {
	counter := 0
	func() {
		for {
			select {
			case <-th.closeChan:
				return
			case req := <-th.rchan:
				if counter%2 == 0 {
					counter++
					th.Add(1)
					go th.handleComplex(req)
				} else {
					counter++
					th.Add(1)
					go th.handleSimple(req)
				}
			}
		}
	}()
	th.Wait()
}

func (th *testHandler) Stop() {
	th.closeOnce.Do(func() {
		close(th.closeChan)
	})
}

func (th *testHandler) handleSimple(req Request) {
	defer th.Done()
	switch v := req.Request().(type) {
	case int:
		v++
		req.Respond(v)
	default:
		th.Add(1)
		go th.handleComplex(req)
	}
}

func (th *testHandler) handleComplex(req Request) {
	defer th.Done()
	subChan := make(chan interface{}) // this could go to BC handler
	go th.typeDispatch(req.Request(), subChan)
	select {
	case <-req.Timeout():
		return
	case response := <-subChan:
		req.Respond(response)
	case <-th.closeChan:
		req.Respond(errors.New("this should never happen"))
		return
	}
}

func (th *testHandler) typeDispatch(obj interface{}, rchan chan<- interface{}) {
	switch v := obj.(type) {
	case int:
		v++
		rchan <- v
	case time.Duration:
		time.Sleep(v)
		rchan <- time.Now()
	default:
		rchan <- errors.New("bad type")
	}
}

func setupHandler(t *testing.T, rb Rbus) func() {
	rchan, err := rb.Register(SNAME, 3)
	if err != nil {
		t.Error(err)
	}
	th := testHandler{
		closeChan: make(chan struct{}),
		rchan:     rchan,
	}
	go th.Start()
	return th.Stop
}

func testInt(t *testing.T, rb Rbus, rv int) {
	resp, err := rb.Request(SNAME, 1*time.Second, rv)
	if err != nil {
		t.Error(err)
	}
	select {
	case <-resp.Timeout():
		t.Error("got an unexpected timeout")
	case r := <-resp.Response():
		switch v := r.(type) {
		case int:
			if v == rv+1 {
				return
			}
			t.Errorf("got back a number that was invalid: sent:%d got:%d\n", rv, v)
		case error:
			t.Errorf("should not get back an error, error was %s", v)
		default:
			t.Errorf("got back invalid type! %T", v)
		}
	}
}

func testErr(t *testing.T, rb Rbus, rv uint32) {
	resp, err := rb.Request(SNAME, 1*time.Second, rv)
	if err != nil {
		t.Error(err)
	}
	select {
	case <-resp.Timeout():
		t.Error("got an unexpected timeout")
	case r := <-resp.Response():
		switch v := r.(type) {
		case error:
			return
		default:
			t.Errorf("got back invalid type! %T", v)
		}
	}
}

func testTimeoutDelayed(t *testing.T, rb Rbus) {
	resp, err := rb.Request(SNAME, 3*time.Second, 6*time.Second)
	if err != nil {
		t.Error(err)
	}
	select {
	case <-resp.Timeout():
		return
	case r := <-resp.Response():
		switch v := r.(type) {
		case error:
			t.Errorf("got back an unexpected error: %s", v)
		default:
			t.Errorf("got back invalid type! %T", v)
		}
	}
}

func testTimeoutBackpressure(t *testing.T, rb Rbus) {
	resp, err := rb.Request(SNAME, 1*time.Nanosecond, 1*time.Second)
	if err != nil {
		t.Error(err)
	}
	select {
	case <-resp.Timeout():
		return
	case r := <-resp.Response():
		switch v := r.(type) {
		case error:
			t.Errorf("got back an unexpected error: %s", v)
		default:
			t.Errorf("got back invalid type! %T", v)
		}
	}
}

func testBlockingShort(t *testing.T, rb Rbus) {
	resp, err := rb.Request(SNAME, 0, 0)
	if err != nil {
		t.Error(err)
	}
	select {
	case <-resp.Timeout():
		return
	case r := <-resp.Response():
		switch v := r.(type) {
		case int:
			return
		case error:
			t.Errorf("got back an unexpected error: %s", v)
		default:
			t.Errorf("got back invalid type! %T", v)
		}
	}
}

func testBlockingLong(t *testing.T, rb Rbus) {
	resp, err := rb.Request(SNAME, 0, time.Second*4)
	if err != nil {
		t.Error(err)
	}
	select {
	case <-resp.Timeout():
		t.Error("impossible timeout!")
	case r := <-resp.Response():
		switch v := r.(type) {
		case time.Time:
			return
		default:
			t.Errorf("got back invalid type! %T", v)
		}
	}
}
