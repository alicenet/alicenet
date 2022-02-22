package monitor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/config"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
	Close()
	GetStatus() <-chan string
}

type monitor struct {
	sync.RWMutex
	adminHandler   interfaces.AdminHandler
	depositHandler interfaces.DepositHandler
	eth            interfaces.Ethereum
	eventMap       *objects.EventMap
	db             *db.Database
	cdb            *db.Database
	tickInterval   time.Duration
	timeout        time.Duration
	logger         *logrus.Entry
	cancelChan     chan bool
	statusChan     chan string
	TypeRegistry   *objects.TypeRegistry
	State          *objects.MonitorState
	wg             *sync.WaitGroup
	batchSize      uint64
}

// NewMonitor creates a new Monitor
func NewMonitor(cdb *db.Database,
	db *db.Database,
	adminHandler interfaces.AdminHandler,
	depositHandler interfaces.DepositHandler,
	eth interfaces.Ethereum,
	tickInterval time.Duration,
	timeout time.Duration,
	batchSize uint64) (*monitor, error) {

	logger := logging.GetLogger("monitor").WithFields(logrus.Fields{
		"Interval": tickInterval.String(),
		"Timeout":  timeout.String(),
	})

	// Type registry is used to bidirectionally map a type name string to it's reflect.Type
	// -- This lets us use a wrapper class and unmarshal something where we don't know its type
	//    in advance.
	tr := &objects.TypeRegistry{}

	tr.RegisterInstanceType(&dkgtasks.CompletionTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeShareDistributionTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeMissingShareDistributionTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeMissingKeySharesTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeMissingGPKjTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeGPKjTask{})
	tr.RegisterInstanceType(&dkgtasks.GPKjSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.KeyshareSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.MPKSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.PlaceHolder{})
	tr.RegisterInstanceType(&dkgtasks.RegisterTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeMissingRegistrationTask{})
	tr.RegisterInstanceType(&dkgtasks.ShareDistributionTask{})

	eventMap := objects.NewEventMap()
	err := SetupEventMap(eventMap, cdb, adminHandler, depositHandler)
	if err != nil {
		return nil, err
	}

	wg := new(sync.WaitGroup)

	adminHandler.RegisterSnapshotCallback(func(bh *objs.BlockHeader) error {
		ctx, cf := context.WithTimeout(context.Background(), timeout)
		defer cf()

		logger.Info("Entering snapshot callback")
		return PersistSnapshot(ctx, wg, eth, logger, bh)
	})

	schedule := objects.NewSequentialSchedule(tr, adminHandler)
	dkgState := objects.NewDkgState(eth.GetDefaultAccount())
	State := objects.NewMonitorState(dkgState, schedule)

	return &monitor{
		adminHandler:   adminHandler,
		depositHandler: depositHandler,
		eth:            eth,
		eventMap:       eventMap,
		cdb:            cdb,
		db:             db,
		TypeRegistry:   tr,
		logger:         logger,
		tickInterval:   tickInterval,
		timeout:        timeout,
		cancelChan:     make(chan bool, 1),
		statusChan:     make(chan string, 1),
		State:          State,
		wg:             wg,
		batchSize:      batchSize,
	}, nil

}

func (mon *monitor) LoadState() error {

	mon.Lock()
	defer mon.Unlock()

	if err := mon.db.View(func(txn *badger.Txn) error {
		keyLabel := fmt.Sprintf("%x", getStateKey())
		mon.logger.WithField("Key", keyLabel).Infof("Looking up state")
		rawData, err := utils.GetValue(txn, getStateKey())
		if err != nil {
			return err
		}
		// TODO: Cleanup loaded obj, this is a memory / storage leak
		err = json.Unmarshal(rawData, mon)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil

}

func (mon *monitor) PersistState() error {

	mon.Lock()
	defer mon.Unlock()

	rawData, err := json.Marshal(mon)
	if err != nil {
		return err
	}

	err = mon.db.Update(func(txn *badger.Txn) error {
		keyLabel := fmt.Sprintf("%x", getStateKey())
		mon.logger.WithField("Key", keyLabel).Infof("Saving state")
		if err := utils.SetValue(txn, getStateKey(), rawData); err != nil {
			mon.logger.Error("Failed to set Value")
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := mon.db.Sync(); err != nil {
		mon.logger.Error("Failed to set sync")
		return err
	}

	return nil
}

func (mon *monitor) GetStatus() <-chan string {
	return mon.statusChan
}

func (mon *monitor) Close() {
	mon.cancelChan <- true
}

// Start starts the event loop
func (mon *monitor) Start() error {

	logger := mon.logger

	// Load or create initial State
	logger.Info(strings.Repeat("-", 80))
	startingBlock := config.Configuration.Ethereum.StartingBlock
	err := mon.LoadState()
	if err != nil {
		logger.Warnf("could not find previous State: %v", err)
		if err != badger.ErrKeyNotFound {
			return err
		}

		logger.Info("Setting initial State to defaults...")

		mon.State.HighestBlockFinalized = startingBlock
		mon.State.HighestBlockProcessed = startingBlock
	}

	if startingBlock > mon.State.HighestBlockProcessed {
		logger.WithFields(logrus.Fields{
			"StartingBlock":         startingBlock,
			"HighestBlockProcessed": mon.State.HighestBlockProcessed}).
			Info("Overriding highest block processed due to config")
		mon.State.HighestBlockProcessed = startingBlock
	}

	if startingBlock > mon.State.HighestBlockFinalized {
		logger.WithFields(logrus.Fields{
			"StartingBlock":         startingBlock,
			"HighestBlockFinalized": mon.State.HighestBlockFinalized}).
			Info("Overriding highest block finalized due to config")
		mon.State.HighestBlockFinalized = startingBlock
	}

	mon.State.EndpointInSync = false
	logger.Info("Current State:")
	logger.Infof("...Ethereum in sync: %v", mon.State.EthereumInSync)
	logger.Infof("...Highest block finalized: %v", mon.State.HighestBlockFinalized)
	logger.Infof("...Highest block processed: %v", mon.State.HighestBlockProcessed)
	logger.Infof("...Monitor tick interval: %v", mon.tickInterval.String())
	logger.Info(strings.Repeat("-", 80))
	logger.Infof("Current Tasks: %v", len(mon.State.Schedule.Ranges))
	for id, block := range mon.State.Schedule.Ranges {
		taskName, _ := objects.GetNameType(block.Task)
		logger.Infof("...ID: %v Name: %v Between: %v and %v", id, taskName, block.Start, block.End)
	}
	logger.Info(strings.Repeat("-", 80))

	mon.cancelChan = make(chan bool)
	mon.wg.Add(1)
	go mon.eventLoop(mon.wg, logger, mon.cancelChan)
	return nil
}

func (mon *monitor) eventLoop(wg *sync.WaitGroup, logger *logrus.Entry, cancelChan <-chan bool) error {

	defer wg.Done()
	gcTimer := time.After(time.Second * constants.MonDBGCFreq)
	for {
		ctx, cf := context.WithTimeout(context.Background(), mon.timeout)
		tock := mon.tickInterval
		bmax := max(mon.State.HighestBlockFinalized, mon.State.HighestBlockProcessed)
		bmin := min(mon.State.HighestBlockFinalized, mon.State.HighestBlockProcessed)
		if !(bmax-bmin < mon.batchSize) {
			tock = time.Millisecond * 100
		}
		select {
		case <-gcTimer:
			mon.db.DB().RunValueLogGC(constants.BadgerDiscardRatio)
			gcTimer = time.After(time.Second * constants.MonDBGCFreq)
		case <-cancelChan:
			mon.logger.Warnf("Received cancel request for event loop.")
			cf()
			return nil
		case tick := <-time.After(tock):
			mon.logger.WithTime(tick).Debug("Tick")

			oldMonitorState := mon.State.Clone()

			if err := MonitorTick(ctx, cf, wg, mon.eth, mon.State, mon.logger, mon.eventMap, mon.adminHandler, mon.batchSize); err != nil {
				logger.Errorf("Failed MonitorTick(...): %v", err)
			}

			diff, shouldWrite := oldMonitorState.Diff(mon.State)

			if shouldWrite {
				if err := mon.PersistState(); err != nil {
					logger.Errorf("Failed to persist State after MonitorTick(...): %v", err)
				}
			}

			select {
			case mon.statusChan <- diff:
			default:
			}
		}
	}
}

func (m *monitor) MarshalJSON() ([]byte, error) {
	rawData, err := json.Marshal(m.State)

	if err != nil {
		return nil, fmt.Errorf("could not marshal state: %v", err)
	}

	return rawData, nil
}

func (m *monitor) UnmarshalJSON(raw []byte) error {
	err := json.Unmarshal(raw, m.State)
	if err != nil {
		if m.State.Schedule != nil {
			m.State.Schedule.Initialize(m.TypeRegistry, m.adminHandler)
			// TODO: VERIFY ALL TASKS SHOULD NOT BE RE-STARTED HERE
			return nil
		}
	}
	return err
}

// MonitorTick using existing monitorState and incrementally updates it based on current State of Ethereum endpoint
func MonitorTick(ctx context.Context, cf context.CancelFunc, wg *sync.WaitGroup, eth interfaces.Ethereum, monitorState *objects.MonitorState, logger *logrus.Entry,
	eventMap *objects.EventMap, adminHandler interfaces.AdminHandler, batchSize uint64) error {

	defer cf()
	logger = logger.WithFields(logrus.Fields{
		"Method":         "MonitorTick",
		"EndpointInSync": monitorState.EndpointInSync,
		"EthereumInSync": monitorState.EthereumInSync})

	c := eth.Contracts()
	addresses := []common.Address{c.ValidatorsAddress(), c.DepositAddress(), c.EthdkgAddress(), c.GovernorAddress()}

	// 1. Check if our Ethereum endpoint is sync with sufficient peers
	inSync, peerCount, err := EndpointInSync(ctx, eth, logger)
	ethInSyncBefore := monitorState.EthereumInSync
	monitorState.EndpointInSync = inSync
	bmax := max(monitorState.HighestBlockFinalized, monitorState.HighestBlockProcessed)
	bmin := min(monitorState.HighestBlockFinalized, monitorState.HighestBlockProcessed)
	monitorState.EthereumInSync = bmax-bmin < 2 && monitorState.EndpointInSync
	if ethInSyncBefore != monitorState.EthereumInSync {
		adminHandler.SetSynchronized(monitorState.EthereumInSync)
	}
	if err != nil {
		monitorState.CommunicationFailures++

		logger.WithField("CommunicationFailures", monitorState.CommunicationFailures).
			WithField("Error", err).
			Warn("EndpointInSync() Failed")

		if monitorState.CommunicationFailures >= uint32(eth.RetryCount()) {
			monitorState.EndpointInSync = false
		}
		return nil
	}

	if peerCount < uint32(config.Configuration.Ethereum.EndpointMinimumPeers) {
		return nil
	}

	// 2. Check what the latest finalized block number is
	finalized, err := eth.GetFinalizedHeight(ctx)
	if err != nil {
		return err
	}

	monitorState.CommunicationFailures = 0
	monitorState.PeerCount = peerCount
	monitorState.EndpointInSync = inSync
	monitorState.HighestBlockFinalized = finalized

	// 3. Grab up to the next _batch size_ unprocessed block(s)
	processed := monitorState.HighestBlockProcessed
	if processed >= finalized {
		return nil
	}

	lastBlock := uint64(0)
	remaining := finalized - processed
	if remaining <= batchSize {
		lastBlock = processed + remaining
	} else {
		lastBlock = processed + batchSize
	}

	logsList, err := getLogsConcurrentWithSort(ctx, addresses, eth, processed, lastBlock)
	if err != nil {
		return err
	}
	// set the current block initial value
	// this value is incremented at head of
	// each loop iteration, so it is initialized
	// as one less than the expected value at this
	// point
	currentBlock := processed

	for i := 0; i < len(logsList); i++ {
		currentBlock++
		logEntry := logger.WithField("Block", currentBlock)

		logs := logsList[i]

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
				logEntry.Debug("Found unknown event")
			}

		}

		// Check if any tasks are scheduled
		logEntry.Debug("Looking for scheduled task")
		uuid, err := monitorState.Schedule.Find(currentBlock)
		if err == nil {
			task, _ := monitorState.Schedule.Retrieve(uuid)

			taskName, _ := objects.GetNameType(task)

			log := logEntry.WithFields(logrus.Fields{
				"TaskID":   uuid.String(),
				"TaskName": taskName})

			tasks.StartTask(log, wg, eth, task, nil)

			monitorState.Schedule.Remove(uuid)
		} else if err == objects.ErrNothingScheduled {
			logEntry.Debug("No tasks scheduled")
		} else {
			logEntry.Warnf("Error retrieving scheduled task: %v", err)
		}
		processed = currentBlock
	}

	// Only after batch is processed do we update monitor State
	monitorState.HighestBlockProcessed = processed
	return nil
}

// PersistSnapshot should be registered as a callback and be kicked off automatically by badger when appropriate
func PersistSnapshot(ctx context.Context, wg *sync.WaitGroup, eth interfaces.Ethereum, logger *logrus.Entry, bh *objs.BlockHeader) error {

	task := tasks.NewSnapshotTask(eth.GetDefaultAccount())
	task.BlockHeader = bh

	tasks.StartTask(logger, wg, eth, task, nil)

	return nil
}

// EndpointInSync Checks if our endpoint is good to use
// -- This function is different. Because we need to be aware of errors, State is always updated
func EndpointInSync(ctx context.Context, eth interfaces.Ethereum, logger *logrus.Entry) (bool, uint32, error) {

	// Default to assuming everything is awful
	inSync := false
	peerCount := uint32(0)

	// Check if the endpoint is itself still syncing
	syncing, progress, err := eth.GetSyncProgress()
	if err != nil {
		logger.Warnf("Could not check if Ethereum endpoint it still syncing: %v", err)
		return inSync, peerCount, err
	}

	if syncing && progress != nil {
		logger.Debugf("Ethereum endpoint syncing... at block %v of %v.",
			progress.CurrentBlock, progress.HighestBlock)
	}

	peerCount64, err := eth.GetPeerCount(ctx)
	if err != nil {
		return inSync, peerCount, err
	}
	peerCount = uint32(peerCount64)

	// TODO Remove direct reference to config. Specific values should be passed in.
	if !syncing && peerCount >= uint32(config.Configuration.Ethereum.EndpointMinimumPeers) {
		inSync = true
	}

	return inSync, peerCount, err
}

// TODO: Remove from request hot path use memory cache
// persist worker group across execution iterations
type logWork struct {
	isLast    bool
	ctx       context.Context
	block     uint64
	addresses []common.Address
	logs      []types.Log
	err       error
}

type eventSorter struct {
	*sync.Mutex
	wg      *sync.WaitGroup
	pending chan *logWork
	done    map[uint64]*logWork
	eth     interfaces.Ethereum
}

func (es *eventSorter) Start(num uint64) {
	for i := uint64(0); i < num; i++ {
		es.wg.Add(1)
		go es.wrkr()
	}
	es.wg.Wait()
}

func (es *eventSorter) wrkr() {
	defer es.wg.Done()
	for {
		wrk, ok := <-es.pending
		if !ok {
			return
		}
		if wrk.isLast {
			close(es.pending)
			return
		}
		func() {
			for i := 0; i < 10; i++ {
				select {
				case <-wrk.ctx.Done():
					wrk.err = wrk.ctx.Err()
					es.Lock()
					es.done[wrk.block] = wrk
					es.Unlock()
					return
				default:
					logs, err := es.eth.GetEvents(wrk.ctx, wrk.block, wrk.block, wrk.addresses)
					if err == nil {
						wrk.logs = logs
						wrk.err = nil
						es.Lock()
						es.done[wrk.block] = wrk
						es.Unlock()
						return
					}
					select {
					case <-time.After(10 * time.Duration(i) * time.Millisecond):
						// continue trying
					case <-wrk.ctx.Done():
						wrk.err = wrk.ctx.Err()
						es.Lock()
						es.done[wrk.block] = wrk
						es.Unlock()
						return
					}
				}
			}
			wrk.err = errors.New("timeouts exhausted")
			es.Lock()
			es.done[wrk.block] = wrk
		}()
	}
}

func getLogsConcurrentWithSort(ctx context.Context, addresses []common.Address, eth interfaces.Ethereum, processed uint64, lastBlock uint64) ([][]types.Log, error) {
	numworkers := max(min((max(lastBlock, processed)-min(lastBlock, processed))/4, 128), 1)
	wc := make(chan *logWork, 3+numworkers)
	go func() {
		for currentBlock := processed + 1; currentBlock <= lastBlock; currentBlock++ {
			blk := currentBlock
			wc <- &logWork{false, ctx, blk, addresses, nil, nil}
		}
		wc <- &logWork{true, nil, 0, nil, nil, nil}
	}()

	es := &eventSorter{new(sync.Mutex), new(sync.WaitGroup), wc, make(map[uint64]*logWork), eth}
	es.Start(numworkers)

	la := [][]types.Log{}
	for currentBlock := processed + 1; currentBlock <= lastBlock; currentBlock++ {
		if es.done[currentBlock].err != nil {
			return la, nil
		}
		logsO, ok := es.done[currentBlock]
		if !ok {
			return la, nil
		}
		la = append(la, logsO.logs)
	}
	return la, nil
}

func max(a uint64, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func min(a uint64, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}
