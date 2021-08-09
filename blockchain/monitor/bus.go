package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
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
	logger            *logrus.Entry
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
		logger:            logging.GetLogger("monitor").WithField("Component", "bus"),
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

	// setup
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	eth := bus.svcs.eth
	monitorState := request.Request().(*objects.MonitorState)

	// TODO Fix this garbage
	logger := logging.GetLogger("services").WithField("Field", "Foo")
	var eventMap *objects.EventMap
	var adminHandler interfaces.AdminHandler
	var schedule interfaces.Schedule

	err := MonitorTick(ctx, wg, eth, monitorState, logger, eventMap, schedule, adminHandler)
	if err != nil {
		request.Respond(err)
	}
	request.Respond(monitorState) // TODO Is this required? I only care about error or not, state mutation just happens
	return nil
}
