package monitor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/snapshots"
	"github.com/alicenet/alicenet/layer1/executor/tasks/snapshots/state"
	"github.com/alicenet/alicenet/layer1/monitor/events"
	"github.com/alicenet/alicenet/layer1/monitor/interfaces"
	"github.com/alicenet/alicenet/layer1/monitor/objects"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
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
	eth            layer1.Client
	contracts      layer1.Contracts
	eventMap       *objects.EventMap
	db             *db.Database
	cdb            *db.Database
	tickInterval   time.Duration
	timeout        time.Duration
	logger         *logrus.Entry
	cancelChan     chan bool
	statusChan     chan string
	State          *objects.MonitorState
	batchSize      uint64

	//for communication with the TasksScheduler
	taskRequestChan chan<- tasks.TaskRequest
}

// NewMonitor creates a new Monitor
func NewMonitor(cdb *db.Database,
	monDB *db.Database,
	adminHandler interfaces.AdminHandler,
	depositHandler interfaces.DepositHandler,
	eth layer1.Client,
	contracts layer1.Contracts,
	tickInterval time.Duration,
	batchSize uint64,
	taskRequestChan chan<- tasks.TaskRequest,
) (*monitor, error) {

	logger := logging.GetLogger("monitor").WithFields(logrus.Fields{
		"Interval": tickInterval.String(),
		"Timeout":  constants.MonitorTimeout.String(),
	})

	eventMap := objects.NewEventMap()
	err := events.SetupEventMap(eventMap, cdb, monDB, adminHandler, depositHandler, taskRequestChan)
	if err != nil {
		return nil, err
	}

	State := objects.NewMonitorState()

	adminHandler.RegisterSnapshotCallback(func(bh *objs.BlockHeader) error {
		logger.Info("Entering snapshot callback")
		return PersistSnapshot(eth, bh, taskRequestChan, monDB)
	})

	return &monitor{
		adminHandler:    adminHandler,
		depositHandler:  depositHandler,
		eth:             eth,
		contracts:       contracts,
		eventMap:        eventMap,
		cdb:             cdb,
		db:              monDB,
		logger:          logger,
		tickInterval:    tickInterval,
		timeout:         constants.MonitorTimeout,
		cancelChan:      make(chan bool, 1),
		statusChan:      make(chan string, 1),
		State:           State,
		batchSize:       batchSize,
		taskRequestChan: taskRequestChan,
	}, nil

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
	err := mon.State.LoadState(mon.db)
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

	mon.cancelChan = make(chan bool)
	go mon.eventLoop(logger, mon.cancelChan)
	return nil
}

func (mon *monitor) eventLoop(logger *logrus.Entry, cancelChan <-chan bool) {

	gcTimer := time.After(time.Second * constants.MonDBGCFreq)
	for {
		ctx, cf := context.WithTimeout(context.Background(), mon.timeout)
		tock := mon.tickInterval
		bmax := utils.Max(mon.State.HighestBlockFinalized, mon.State.HighestBlockProcessed)
		bmin := utils.Min(mon.State.HighestBlockFinalized, mon.State.HighestBlockProcessed)
		if !(bmax-bmin < mon.batchSize) {
			tock = time.Millisecond * 100
		}
		select {
		case <-gcTimer:
			err := mon.db.DB().RunValueLogGC(constants.BadgerDiscardRatio)
			if err != nil {
				logger.Debugf("Failed to reclaim any space during garbage collection: %v", err)
			}
			gcTimer = time.After(time.Second * constants.MonDBGCFreq)
		case <-cancelChan:
			mon.logger.Warnf("Received cancel request for event loop.")
			cf()
			return
		case tick := <-time.After(tock):
			mon.logger.WithTime(tick).Debug("Tick")

			oldMonitorState := mon.State.Clone()

			if err := MonitorTick(ctx, cf, mon.eth, mon.State, mon.logger, mon.eventMap, mon.adminHandler, mon.batchSize, mon.contracts); err != nil {
				logger.Errorf("Failed MonitorTick(...): %v", err)
			}

			diff, shouldWrite := oldMonitorState.Diff(mon.State)

			if shouldWrite {
				if err := mon.State.PersistState(mon.db); err != nil {
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
	m.State.RLock()
	defer m.State.RUnlock()
	rawData, err := json.Marshal(m.State)

	if err != nil {
		return nil, fmt.Errorf("could not marshal state: %v", err)
	}

	return rawData, nil
}

func (m *monitor) UnmarshalJSON(raw []byte) error {
	err := json.Unmarshal(raw, m.State)
	return err
}

// MonitorTick using existing monitorState and incrementally updates it based on current State of Ethereum endpoint
func MonitorTick(ctx context.Context, cf context.CancelFunc, eth layer1.Client, monitorState *objects.MonitorState, logger *logrus.Entry,
	eventMap *objects.EventMap, adminHandler interfaces.AdminHandler, batchSize uint64, contracts layer1.Contracts) error {

	defer cf()
	logger = logger.WithFields(logrus.Fields{
		"EndpointInSync": monitorState.EndpointInSync,
		"EthereumInSync": monitorState.EthereumInSync})

	addresses := contracts.GetAllAddresses()

	// 1. Check if our Ethereum endpoint is sync with sufficient peers
	inSync, peerCount, err := eth.EndpointInSync(ctx)
	ethInSyncBefore := monitorState.EthereumInSync
	monitorState.EndpointInSync = inSync
	bmax := utils.Max(monitorState.HighestBlockFinalized, monitorState.HighestBlockProcessed)
	bmin := utils.Min(monitorState.HighestBlockFinalized, monitorState.HighestBlockProcessed)
	monitorState.EthereumInSync = bmax-bmin < 2 && monitorState.EndpointInSync && monitorState.IsInitialized
	if ethInSyncBefore != monitorState.EthereumInSync {
		adminHandler.SetSynchronized(monitorState.EthereumInSync)
	}
	if err != nil {
		monitorState.CommunicationFailures++

		logger.WithField("CommunicationFailures", monitorState.CommunicationFailures).
			WithField("Error", err).
			Warn("EndpointInSync() Failed")

		if monitorState.CommunicationFailures >= uint32(constants.MonitorRetryCount) {
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
	monitorState.IsInitialized = true

	// 3. Grab up to the next _batch size_ unprocessed block(s)
	processed := monitorState.HighestBlockProcessed
	if processed >= finalized {
		logger.Debugf("Processed block %d is higher than finalized block %d", processed, finalized)
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
		logs := logsList[i]

		var forceExit bool
		currentBlock, err = ProcessEvents(eth, monitorState, logs, logger, currentBlock, eventMap)
		if err != nil {
			if !errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			forceExit = true
		}

		processed = currentBlock
		if forceExit {
			break
		}
	}

	// Only after batch is processed do we update monitor State
	monitorState.HighestBlockProcessed = processed
	return nil
}

func ProcessEvents(eth layer1.Client, monitorState *objects.MonitorState, logs []types.Log, logger *logrus.Entry, currentBlock uint64, eventMap *objects.EventMap) (uint64, error) {
	logEntry := logger.WithField("Block", currentBlock)

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
					return currentBlock - 1, err
				}
			} else {
				panic(fmt.Errorf("no processor configured for %v", info.Name))
			}
		}
	}

	return currentBlock, nil
}

// PersistSnapshot should be registered as a callback and be kicked off automatically by badger when appropriate
func PersistSnapshot(eth layer1.Client, bh *objs.BlockHeader, taskRequestChan chan<- tasks.TaskRequest, monDB *db.Database) error {
	if bh == nil {
		return errors.New("invalid blockHeader for snapshot")
	}

	snapshotState := &state.SnapshotState{
		Account:     eth.GetDefaultAccount(),
		BlockHeader: bh,
	}

	err := state.SaveSnapshotState(monDB, snapshotState)
	if err != nil {
		return err
	}

	// kill any snapshot task that might be running
	taskRequestChan <- tasks.NewKillTaskRequest(&snapshots.SnapshotTask{})

	taskRequestChan <- tasks.NewScheduleTaskRequest(snapshots.NewSnapshotTask(0, 0, uint64(bh.BClaims.Height)))

	return nil
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
	eth     layer1.Client
}

func (es *eventSorter) Start(num uint64) {
	for i := uint64(0); i < num; i++ {
		es.wg.Add(1)
		go es.worker()
	}
	es.wg.Wait()
}

func (es *eventSorter) worker() {
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
			es.Unlock()
		}()
	}
}

func getLogsConcurrentWithSort(ctx context.Context, addresses []common.Address, eth layer1.Client, processed uint64, lastBlock uint64) ([][]types.Log, error) {
	numworkers := utils.Max(utils.Min((utils.Max(lastBlock, processed)-utils.Min(lastBlock, processed))/4, 128), 1)
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
