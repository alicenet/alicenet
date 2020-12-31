package rbus

import (
	"errors"
	"sync"
	"time"
)

// NewRBus returns a Request Bus as an interface.
func NewRBus() Rbus {
	return &rbus{
		channelMsg:   make(map[string]chan Request),
		channelGuard: make(map[string]bool),
	}
}

// Rbus is a request bus interface for inter service communication.
// The actual implementation is in this file as well and may be
// constructed by invoking NewRBus()
//
// This request bus is a many-to-one message bus. IE: Many
// services can talk to a single registered service by name.
// The service that calls Register() should listen to the returned
// channel in a loop and launch a handler for each request.
// The bus is safe against races and deadlocks under a few
// assumptions. In order use this bus in a safe manner, the
// code that uses this bus must follow this contract. The
// description follows:
//
// A request that is sent BEFORE a service has been registered MUST
// use a timeout OR MUST gracefully handle an ErrUnknownService
// response.
//
// A request that is sent BEFORE a service has been registered AND
// DOES use a timeout, MUST check the response in a select statement
// that also checks for a timeout.
//
// If a program may have a message sent to a service that does not
// exist before the message is sent, the program SHOULD assume that
// a timeout is an uncaught exception and unwind. This recommendation
// is to prevent a bad service name from silently causing a memory
// leak.
//
// If a service sends a request with a timeout, it MUST
// check the timeout to prevent memory leaks, but MAY
// continue operation IFF the service is guaranteed to have
// existed before any message could have been sent.
//
// The service does not have to check for timeouts before
// writing a response. The service can pre-emptively quit
// processing a response if a time out occurs before processing
// starts or at any time during. Once a timeout occurs, there
// is no guarantee if the requestee will get a timeout or the
// response. This is because a select statement returns a random
// item if more than one is available.
//
// If a timeout occurs before a message may be placed on the
// bus, the message will never enter the bus. This is to
// provide backpressure for an overloaded service. The
// backpressure limit is set by the request channel capacity.
//
// If a requester calls Request() using a timeout of zero,
// no timeout signal will ever be fired. Specifically, the
// timeout channel will not be constructed. This means if a
// timeout of zero is used, any go routine that calls Timeout()
// outside of a Go routine MUST do so in a select statement.
// This requirement MUST be observed to prevent deadlocks
// from waiting for a non-existent channel.
//
// The consuming service MUST always call Timeout() in a select
// statement if the service does call timeout. Once again, the
// service does not have to listen for a timeout, but if it does
// it must do it in a select.
type Rbus interface {
	Register(serviceName string, capacity uint16) (<-chan Request, error)
	Request(serviceName string, timeOut time.Duration, request interface{}) (Response, error)
}

// ErrReservation is an error raised if a name is reserved twice with the
// message bus
var ErrReservation = errors.New("double reservation requested for channel name")

// ErrUnknownService is the response a service gets if it requests a handler
// that does not exist yet
var ErrUnknownService = errors.New("the requested service does not exist yet")

// Response is the interface handed back to the caller of the request bus
type Response interface {
	Timeout() <-chan struct{}
	Response() <-chan interface{}
}

// Request is the interface handed to the handler of a request from the bus
type Request interface {
	Timeout() <-chan struct{}
	Request() interface{}
	Respond(interface{})
}

// Rmsg is a message bus request object
// it allows a caller to pass a request
// to a remote handler for processing.
// When the handler is finished working
// on the request, the handler may pass
// back a response on the Rchan. The
// TimeoutChan is a channel that fires
// after a timeout chosen by the requesting
// party. The handler may still write the
// response, but the caller will not get it.
type rmsg struct {
	respondOnce sync.Once
	signal      chan struct{}
	request     interface{}
	rchan       chan interface{}
	done        chan struct{}
}

func (rm *rmsg) start(to time.Duration) {
	select {
	case <-time.After(to):
		close(rm.signal)
	case <-rm.done:
		return
	}
}

// Request is a function that will return the request object to the handler
func (rm *rmsg) Request() interface{} {
	return rm.request
}

// Respond is a function that will give a response back to the requester
func (rm *rmsg) Respond(obj interface{}) {
	rm.respondOnce.Do(func() {
		close(rm.done)
		rm.rchan <- obj
	})
}

// Timeout allows the requester/handler to be notified of a timeout signal
// as a channel.
func (rm *rmsg) Timeout() <-chan struct{} {
	return rm.signal
}

// Response returns a channel that carries the response to the
// caller when it is available
func (rm *rmsg) Response() <-chan interface{} {
	return rm.rchan
}

type rbus struct {
	sync.Mutex
	channelMsg   map[string]chan Request
	channelGuard map[string]bool
}

// Register allows a service to register itself for receiving messages
func (rb *rbus) Register(cname string, cap uint16) (<-chan Request, error) {
	rb.Lock()
	defer rb.Unlock()
	if rb.channelGuard[cname] {
		return nil, ErrReservation
	}
	rb.channelGuard[cname] = true
	msgChan := make(chan Request, int(cap))
	rb.channelMsg[cname] = msgChan
	return msgChan, nil
}

// Request allows a service to request from a remote named service a response
// the requester will get back an object that has a Timeout method and a
// Response method. The Timeout method signals a timeout has occurred based
// on the specified timeout. The Response channel will return a channel that
// carries the response back to the caller when the response is available.
func (rb *rbus) Request(cname string, to time.Duration, request interface{}) (Response, error) {
	if int64(to) > 0 {
		rm := &rmsg{
			respondOnce: sync.Once{},
			signal:      make(chan struct{}),
			done:        make(chan struct{}),
			request:     request,
			rchan:       make(chan interface{}),
		}
		go rm.start(to)
		go func() {
			select {
			case rb.channelMsg[cname] <- rm:
				return
			case <-time.After(to):
				return
			}
		}()
		return rm, nil
	}
	if !rb.channelGuard[cname] {
		return nil, ErrUnknownService
	}
	rm := &rmsg{
		respondOnce: sync.Once{},
		done:        make(chan struct{}),
		signal:      make(chan struct{}),
		request:     request,
		rchan:       make(chan interface{}),
	}
	rb.channelMsg[cname] <- rm
	return rm, nil
}
