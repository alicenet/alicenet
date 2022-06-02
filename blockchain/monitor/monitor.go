package monitor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/snapshots"
	monInterfaces "github.com/MadBase/MadNet/blockchain/monitor/interfaces"
	"strings"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/monitor/events"

	ethereumInterfaces "github.com/MadBase/MadNet/blockchain/ethereum/interfaces"
	"github.com/MadBase/MadNet/blockchain/monitor/objects"
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

func getStateKey() []byte {
	return []byte("monitorStateKey")
}

// Monitor describes required functionality to monitor Ethereum
type Monitor interface {
	Start() error
	Close()
	GetStatus() <-chan string
}

type monitor struct {
	sync.RWMutex
	adminHandler   monInterfaces.IAdminHandler
	depositHandler monInterfaces.IDepositHandler
	eth            ethereumInterfaces.IEthereum
	eventMap       *objects.EventMap
	db             *db.Database
	cdb            *db.Database
	tickInterval   time.Duration
	timeout        time.Duration
	logger         *logrus.Entry
	cancelChan     chan bool
	statusChan     chan string
	State          *objects.MonitorState
	wg             *sync.WaitGroup
	batchSize      uint64

	//for communication with the TasksScheduler
	taskRequestChan chan<- interfaces.ITask
	taskKillChan    chan<- string
}

// NewMonitor creates a new Monitor
func NewMonitor(cdb *db.Database,
	db *db.Database,
	adminHandler monInterfaces.IAdminHandler,
	depositHandler monInterfaces.IDepositHandler,
	eth ethereumInterfaces.IEthereum,
	tickInterval time.Duration,
	timeout time.Duration,
	batchSize uint64,
	taskRequestChan chan<- interfaces.ITask,
	taskKillChan chan<- string) (*monitor, error) {

	logger := logging.GetLogger("monitor").WithFields(logrus.Fields{
		"Interval": tickInterval.String(),
		"Timeout":  timeout.String(),
	})

	eventMap := objects.NewEventMap()
	err := events.SetupEventMap(eventMap, cdb, adminHandler, depositHandler, taskRequestChan, taskKillChan)
	if err != nil {
		return nil, err
	}

	wg := new(sync.WaitGroup)
	State := objects.NewMonitorState()

	//TODO: What to do with this after adding the new TasksScheduler???
	adminHandler.RegisterSnapshotCallback(func(bh *objs.BlockHeader) error {
		ctx, cf := context.WithTimeout(context.Background(), timeout)
		defer cf()

		logger.Info("Entering snapshot callback")
		return PersistSnapshot(eth, bh, taskRequestChan, ctx, cf)
	})

	return &monitor{
		adminHandler:    adminHandler,
		depositHandler:  depositHandler,
		eth:             eth,
		eventMap:        eventMap,
		cdb:             cdb,
		db:              db,
		logger:          logger,
		tickInterval:    tickInterval,
		timeout:         timeout,
		cancelChan:      make(chan bool, 1),
		statusChan:      make(chan string, 1),
		State:           State,
		wg:              wg,
		batchSize:       batchSize,
		taskRequestChan: taskRequestChan,
		taskKillChan:    taskKillChan,
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

	mon.cancelChan = make(chan bool)
	mon.wg.Add(1)
	go mon.eventLoop(mon.wg, logger, mon.cancelChan)
	return nil
}

func (mon *monitor) eventLoop(wg *sync.WaitGroup, logger *logrus.Entry, cancelChan <-chan bool) {

	defer wg.Done()
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
				logger.Errorf("Failed to run value log GC: %v", err)
			}
			gcTimer = time.After(time.Second * constants.MonDBGCFreq)
		case <-cancelChan:
			mon.logger.Warnf("Received cancel request for event loop.")
			cf()
			return
		case tick := <-time.After(tock):
			mon.logger.WithTime(tick).Debug("Tick")

			oldMonitorState := mon.State.Clone()

			persistMonitorCB := func() {
				err := mon.PersistState()
				if err != nil {
					logger.Errorf("Failed to persist State after MonitorTick(...): %v", err)
				}
			}

			if err := MonitorTick(ctx, cf, wg, mon.eth, mon.State, mon.logger, mon.eventMap, mon.adminHandler, mon.batchSize, persistMonitorCB); err != nil {
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
func MonitorTick(ctx context.Context, cf context.CancelFunc, wg *sync.WaitGroup, eth ethereumInterfaces.IEthereum, monitorState *objects.MonitorState, logger *logrus.Entry,
	eventMap *objects.EventMap, adminHandler monInterfaces.IAdminHandler, batchSize uint64, persistMonitorCB func()) error {

	defer cf()
	logger = logger.WithFields(logrus.Fields{
		"Method":         "MonitorTick",
		"EndpointInSync": monitorState.EndpointInSync,
		"EthereumInSync": monitorState.EthereumInSync})

	c := eth.Contracts()
	addresses := []common.Address{c.EthdkgAddress(), c.SnapshotsAddress(), c.BTokenAddress()}

	// 1. Check if our Ethereum endpoint is sync with sufficient peers
	inSync, peerCount, err := EndpointInSync(ctx, eth, logger)
	ethInSyncBefore := monitorState.EthereumInSync
	monitorState.EndpointInSync = inSync
	bmax := utils.Max(monitorState.HighestBlockFinalized, monitorState.HighestBlockProcessed)
	bmin := utils.Min(monitorState.HighestBlockFinalized, monitorState.HighestBlockProcessed)
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
		logs := logsList[i]

		currentBlock, err = ProcessEvents(eth, monitorState, logs, logger, currentBlock, eventMap)
		var forceExit bool
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

func ProcessEvents(eth ethereumInterfaces.IEthereum, monitorState *objects.MonitorState, logs []types.Log, logger *logrus.Entry, currentBlock uint64, eventMap *objects.EventMap) (uint64, error) {
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
func PersistSnapshot(eth ethereumInterfaces.IEthereum, bh *objs.BlockHeader, taskRequestChan chan<- interfaces.ITask, ctx context.Context, cancel context.CancelFunc) error {
	if bh == nil {
		return errors.New("invalid blockHeader for snapshot")
	}
	snapshotTask := snapshots.NewSnapshotTask(eth.GetDefaultAccount(), bh, 0, 0, ctx, cancel)
	taskRequestChan <- snapshotTask

	return nil
}

// EndpointInSync Checks if our endpoint is good to use
// -- This function is different. Because we need to be aware of errors, State is always updated
func EndpointInSync(ctx context.Context, eth ethereumInterfaces.IEthereum, logger *logrus.Entry) (bool, uint32, error) {

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
	eth     ethereumInterfaces.IEthereum
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
			es.Unlock()
		}()
	}
}

func getLogsConcurrentWithSort(ctx context.Context, addresses []common.Address, eth ethereumInterfaces.IEthereum, processed uint64, lastBlock uint64) ([][]types.Log, error) {
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
