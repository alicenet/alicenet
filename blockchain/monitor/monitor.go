package monitor

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/config"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/logging"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

var (
	// ErrUnknownRequest a service was invoked but couldn't figure out which
	ErrUnknownRequest = errors.New("unknown request")

	// ErrUnknownResponse only used when response to a service is not of the expected type
	ErrUnknownResponse = errors.New("response isn't in expected form")
)

// Monitor describes required functionality to monitor Ethereum
type Monitor interface {
	Start() error
	Stop()
	StartEventLoop(accounts.Account) (chan<- bool, error)
	GetStatus() <-chan string
}

type monitor struct {
	adminHandler   interfaces.AdminHandler
	depositHandler interfaces.DepositHandler
	eth            interfaces.Ethereum
	consensusDB    *db.Database
	monitorDB      Database
	tickInterval   time.Duration
	timeout        time.Duration
	logger         *logrus.Logger
	cancelChan     chan bool
	statusChan     chan string
	typeRegistry   *objects.TypeRegistry
	state          *objects.MonitorState
}

// NewMonitor creates a new Monitor
func NewMonitor(consensusDB *db.Database, adminHandler interfaces.AdminHandler, depositHandler interfaces.DepositHandler, eth interfaces.Ethereum, tickInterval time.Duration, timeout time.Duration) (Monitor, error) {

	logger := logging.GetLogger("monitor")

	rand.Seed(time.Now().UnixNano())

	// Type registry is used to bidirectionally map a type name string to it's reflect.Type
	// -- This lets us use a wrapper class and unmarshal something where we don't know its type
	//    in advance.
	tr := &objects.TypeRegistry{}

	tr.RegisterInstanceType(&dkgtasks.CompletionTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeTask{})
	tr.RegisterInstanceType(&dkgtasks.GPKJDisputeTask{})
	tr.RegisterInstanceType(&dkgtasks.GPKSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.KeyshareSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.MPKSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.PlaceHolder{})
	tr.RegisterInstanceType(&dkgtasks.RegisterTask{})
	tr.RegisterInstanceType(&dkgtasks.ShareDistributionTask{})

	eventMap := objects.NewEventMap()
	err := SetupEventMap(eventMap, consensusDB, adminHandler, depositHandler)
	if err != nil {
		return nil, err
	}

	schedule := NewSequentialSchedule(tr, adminHandler)
	dkgState := objects.NewDkgState(eth.GetDefaultAccount())
	state := objects.NewMonitorState(dkgState, schedule)

	return &monitor{
		adminHandler:   adminHandler,
		depositHandler: depositHandler,
		eth:            eth,
		consensusDB:    consensusDB,
		typeRegistry:   tr,
		logger:         logger,
		tickInterval:   tickInterval,
		timeout:        timeout,
		cancelChan:     make(chan bool, 1),
		statusChan:     make(chan string, 1),
		state:          state,
	}, nil

}

func (mon *monitor) GetStatus() <-chan string {
	return mon.statusChan
}

func (mon *monitor) Start() error {

	mon.logger.WithFields(logrus.Fields{
		"s": "f",
	}).Info("Starting event loop")

	return nil
}

func (mon *monitor) Stop() {
	mon.cancelChan <- true
}

// StartEventLoop starts the event loop
func (mon *monitor) StartEventLoop(acct accounts.Account) (chan<- bool, error) {

	logger := mon.logger

	// Load or create initial state
	logger.Info(strings.Repeat("-", 80))
	initialState, err := mon.database.FindState()
	if err != nil {
		logger.Warnf("could not find previous state: %v", err)
		if err != badger.ErrKeyNotFound {
			return nil, err
		}

		logger.Info("Setting initial state to defaults...")
		startingBlock := config.Configuration.Ethereum.StartingBlock
		schedule := NewSequentialSchedule(mon.typeRegistry, mon.adminHandler)
		dkgState := objects.NewDkgState(acct)

		initialState = objects.NewMonitorState(dkgState, schedule)
		initialState.HighestBlockFinalized = uint64(startingBlock)
		initialState.HighestBlockProcessed = uint64(startingBlock)
	}

	initialState.InSync = false
	logger.Info("Current state:")
	logger.Infof("...Ethereum in sync: %v", initialState.EthereumInSync)
	logger.Infof("...Highest block finalized: %v", initialState.HighestBlockFinalized)
	logger.Infof("...Highest block processed: %v", initialState.HighestBlockProcessed)
	logger.Infof("...Monitor tick interval: %v", mon.tickInterval.String())
	logger.Info(strings.Repeat("-", 80))

	cancelChan := make(chan bool)
	go mon.eventLoop(initialState, cancelChan)

	return cancelChan, nil
}

func (mon *monitor) eventLoop(state *objects.MonitorState, cancelChan <-chan bool) error {

	done := false
	for !done {
		select {
		case done = <-cancelChan:
			mon.logger.Warnf("Received cancel request for event loop.")
		case tick := <-time.After(500 * time.Millisecond):
			err := mon.eventLoopTick(state, tick, done)
			if err != nil {
				mon.logger.Warnf("Error occurred: %v", err)
			}
		}
	}

	return nil
}

func (mon *monitor) eventLoopTick(state *objects.MonitorState, tick time.Time, shutdownRequested bool) error {

	logger := mon.logger

	// Make a backup of state to monitor for changes
	originalState := state.Clone()

	// Every tick we request events be processed and we require it doesn't overlap with the next
	resp, err := mon.bus.Request(SvcWatchEthereum, mon.timeout, state)
	if err != nil {
		logger.Warnf("Could not request SvcWatchEthereum: %v", err)
		return err
	}

	select {
	case responseValue := <-resp.Response():
		switch value := responseValue.(type) {
		case *objects.MonitorState:
			diff := originalState.Diff(state)
			if len(diff) > 0 {
				select {
				case mon.statusChan <- fmt.Sprintf("State \xce\x94 %v", diff):
				default:
				}
				mon.database.UpdateState(value)
			}
			return nil
		case error:
			logger.Warnf("SvcWatchEthereum() : %v", value)
		default:
			logger.Errorf("SvcWatchEthereum() invalid return type: %v", value)
		}
	case to := <-resp.Timeout():
		logger.Warnf("SvcWatchEthereum() : Timeout %v", to)
	}

	return nil
}

// MonitorTick using existing monitorState and incrementally updates it based on current state of Ethereum endpoint
func MonitorTick(ctx context.Context, wg *sync.WaitGroup, eth interfaces.Ethereum, monitorState *objects.MonitorState, logger *logrus.Entry,
	eventMap *objects.EventMap, schedule interfaces.Schedule, adminHandler interfaces.AdminHandler) error {

	c := eth.Contracts()

	addresses := []common.Address{c.ValidatorsAddress(), c.DepositAddress(), c.EthdkgAddress(), c.GovernorAddress()}

	// 1. Check if our Ethereum endpoint is sync with sufficient peers
	inSync, peerCount, err := EndpointInSync(ctx, eth, logger)
	if err != nil {
		monitorState.CommunicationFailures++

		logger.WithField("CommunicationFailures", monitorState.CommunicationFailures).
			WithField("Error", err).
			Warn("EndpointInSync() Failed")

		if monitorState.CommunicationFailures >= uint32(eth.RetryCount()) {
			monitorState.InSync = false
			adminHandler.SetSynchronized(false)
		}
		return nil
	}
	monitorState.CommunicationFailures = 0
	monitorState.PeerCount = peerCount
	monitorState.InSync = inSync

	if peerCount < uint32(config.Configuration.Ethereum.EndpointMinimumPeers) {
		return nil
	}

	// 2. Check what the latest finalized block number is
	finalized, err := eth.GetFinalizedHeight(ctx)
	if err != nil {
		return err
	}

	// 3. Grab up to the next _batch size_ unprocessed block(s)
	processed := monitorState.HighestBlockProcessed

	lastBlock := uint64(0)
	remaining := finalized - processed
	if remaining <= 1000 {
		lastBlock = processed + remaining
	} else {
		lastBlock = processed + 1000
	}

	for currentBlock := processed + 1; currentBlock <= lastBlock; currentBlock++ {

		logEntry := logger.WithField("Block", currentBlock)

		logs, err := eth.GetEvents(ctx, currentBlock, currentBlock, addresses)
		if err != nil {
			return err
		}

		// Check all the logs for an event we want to process
		for _, log := range logs {

			eventID := log.Topics[0].String()
			logEntry := logEntry.WithField("EventID", eventID)

			info, present := eventMap.Lookup(eventID)
			if present {
				logEntry = logEntry.WithField("Event", info.Name)
				if info.Processor != nil {
					err := info.Processor(eth, logEntry, monitorState, log)
					if err != nil {
						logEntry.Errorf("Failed processing event: %v", err)
						return err
					}
				} else {
					logEntry.Info("No processor configured.")
				}

			} else {
				logEntry.Debug("Found unkown event")
			}

		}

		// Check if any tasks are scheduled
		logEntry.Info("Looking for scheduled task")
		uuid, err := schedule.Find(currentBlock)
		if err == nil {
			task, _ := schedule.Retrieve(uuid)
			log := logEntry.WithField("TaskID", uuid.String())

			var wg sync.WaitGroup

			wg.Add(1)
			tasks.StartTask(log, &wg, eth, task)

			schedule.Remove(uuid)
		} else if err == ErrNothingScheduled {
			logEntry.Debug("No tasks scheduled")
		} else {
			logEntry.Warnf("Error retrieving scheduled task: %v", err)
		}

		processed = currentBlock
	}

	// Only after batch is processed do we update monitor state
	logger.Debugf("Block Processed %d -> %d, Finalized %d -> %d",
		monitorState.HighestBlockProcessed, processed,
		monitorState.HighestBlockFinalized, finalized)
	monitorState.HighestBlockFinalized = finalized
	monitorState.HighestBlockProcessed = processed

	return nil
}
