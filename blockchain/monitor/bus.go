package monitor

import (
	"sync"
	"time"

	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/rbus"
	"github.com/sirupsen/logrus"
)

// Bus basic functionality required for monitor system
type Bus interface {
	StartLoop() (chan<- bool, error)
	StopLoop() error
	Register(serviceName string, capacity uint16) (<-chan rbus.Request, error)
	Request(serviceName string, timeOut time.Duration, request interface{}) (rbus.Response, error)
}

type monitorBus struct {
	sync.WaitGroup
	svcs              *Services
	rbus              rbus.Rbus
	logger            *logrus.Logger
	processEventsChan <-chan rbus.Request
	cancelChan        chan bool
}

// These are pseudo-constants for registering services on bus
var (
	// SvcProcessEvents Uses the current state to process more events
	SvcWatchEthereum = "SvcWatchEthereum"

	// SvcEndpointInSync Checks if the Ethereum endpoint is ready for use
	SvcEndpointInSync = "SvcEndpointInSync"

	// SvcGetEvents Checks known contracts for recently emitted events
	SvcGetEvents = "SvcGetEvents"

	// SvcGetValidators Calls contract to get current list of validators
	SvcGetValidators = "SvcGetValidators"

	// SvcGetSnapShot Calls contract to get snapshots
	SvcGetSnapShot = "SvcGetSnapShot"
)

// NewBus setups an rbus for monitoring services
func NewBus(rb rbus.Rbus, svcs *Services) (Bus, error) {

	// Each exposed service gets it's own channel
	processEventsChan, err := rb.Register(SvcWatchEthereum, 5)
	if err != nil {
		return nil, err
	}

	return &monitorBus{
		logger:            logging.GetLogger("monitor_bus"),
		svcs:              svcs,
		rbus:              rb,
		processEventsChan: processEventsChan}, nil
}

// StartLoop spawns a gofunc to look for services requests
func (bus *monitorBus) StartLoop() (chan<- bool, error) {
	bus.cancelChan = make(chan bool, 10)

	go func() {
		cancelled := false
		for !cancelled {
			select {

			case cancelled = <-bus.cancelChan:
				bus.logger.Infof("Received cancel message: %v", cancelled)

			case processEventsRequest := <-bus.processEventsChan:
				err := bus.SvcProcessEvents(processEventsRequest)
				if err != nil {
					bus.logger.Warnf("Failed to process events: %v", err)
				}
			}
		}
		bus.logger.Infof("Shutting down...")
		bus.Wait()
	}()

	return bus.cancelChan, nil
}

// StopLoop sends a message on the cancel channel to exit monitoring loop
func (bus *monitorBus) StopLoop() error {
	bus.cancelChan <- true
	return nil
}

// Register wraps rbus.Register to make the API more consistent
func (bus *monitorBus) Register(serviceName string, capacity uint16) (<-chan rbus.Request, error) {
	return bus.rbus.Register(serviceName, capacity)
}

// Request wraps rbus.Request to make the API more consistent
func (bus *monitorBus) Request(serviceName string, timeOut time.Duration, request interface{}) (rbus.Response, error) {
	return bus.rbus.Request(serviceName, timeOut, request)
}

// SvcProcessEvents Exposes the ProcessEvents method on the bus
func (bus *monitorBus) SvcProcessEvents(request rbus.Request) error {
	var state = request.Request().(*State)
	err := bus.svcs.WatchEthereum(state)
	if err != nil {
		request.Respond(err)
	}
	request.Respond(state) // TODO Is this required? I only care about error or not, state mutation just happens
	return nil
}
