package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/MadBase/MadNet/application/deposit"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/config"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

//
// type eventProcessor struct {
// 	name      string
// 	processor func(*objects.MonitorState, types.Log) error
// }

// Services just a bundle of requirements common for monitoring functionality
type Services struct {
	logger            *logrus.Logger
	eth               interfaces.Ethereum
	consensusDb       *db.Database
	dph               *deposit.Handler
	ah                interfaces.AdminHandler
	contractAddresses []common.Address
	batchSize         int
	eventMap          *objects.EventMap
}

// NewServices creates a new Services struct
func NewServices(eth interfaces.Ethereum, db *db.Database, dph *deposit.Handler, ah interfaces.AdminHandler, batchSize int) *Services {

	c := eth.Contracts()

	contractAddresses := []common.Address{
		c.DepositAddress(), c.EthdkgAddress(), c.RegistryAddress(),
		c.StakingTokenAddress(), c.UtilityTokenAddress(), c.ValidatorsAddress(),
		c.GovernorAddress()}

	serviceLogger := logging.GetLogger("services")

	svcs := &Services{
		ah:                ah,
		batchSize:         batchSize,
		consensusDb:       db,
		contractAddresses: contractAddresses,
		dph:               dph,
		eth:               eth,
		eventMap:          objects.NewEventMap(),
		logger:            serviceLogger,
		// taskManager:       tasks.NewManager(),
	}

	// Register handlers for known events, if this failed we really can't continue
	if err := SetupEventMap(svcs.eventMap, svcs.consensusDb, dph, ah); err != nil {
		panic(err)
	}

	ah.RegisterSnapshotCallback(svcs.PersistSnapshot) // HUNTER: moved out of main func and into constructor

	return svcs
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

// EndpointInSync Checks if our endpoint is good to use
// -- This function is different. Because we need to be aware of errors, state is always updated
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

	inSync = !syncing

	peerCount64, err := eth.GetPeerCount(ctx)
	if err != nil {
		return inSync, peerCount, err
	}
	peerCount = uint32(peerCount64)

	// TODO Remove direct reference to config. Specific values should be passed in.
	if inSync && peerCount >= uint32(config.Configuration.Ethereum.EndpointMinimumPeers) {
		inSync = true
	}

	return inSync, peerCount, err
}

// UpdateProgress updates what we know of Ethereum chain height
func (svcs *Services) UpdateProgress(ctx context.Context, state *objects.MonitorState) error {
	height, err := svcs.eth.GetFinalizedHeight(ctx)
	if err != nil {
		return err
	}

	// Only updated single attribute so no need to copy -- Not sure if check is required
	state.HighestBlockFinalized = height
	return nil
}

// PersistSnapshot records the given block header on Ethereum and increments epoch
// TODO Returning an error kills the main loop, retry forever instead
func (svcs *Services) PersistSnapshot(blockHeader *objs.BlockHeader) error {

	eth := svcs.eth
	c := eth.Contracts()
	logger := svcs.logger

	// pull out the block claims
	bclaims := blockHeader.BClaims
	rawBclaims, err := bclaims.MarshalBinary()
	if err != nil {
		logger.Errorf("Could not extract BClaims from BlockHeader: %v", err)
		return nil //CAN NOT RETURN ERROR OR SUBSCRIPTION IS LOST!
	}

	// pull out the sig
	rawSigGroup := blockHeader.SigGroup

	// Do the mechanics
	txnOpts, err := svcs.eth.GetTransactionOpts(context.Background(), svcs.eth.GetDefaultAccount())
	if err != nil {
		logger.Errorf("Could not create transaction for snapshot: %v", err)
		return nil //CAN NOT RETURN ERROR OR SUBSCRIPTION IS LOST!
	}

	txn, err := c.Validators().Snapshot(txnOpts, rawSigGroup, rawBclaims)
	if err != nil {
		logger.Errorf("Failed to take snapshot: %v", err)
		return nil //CAN NOT RETURN ERROR OR SUBSCRIPTION IS LOST!
	}

	toCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	rcpt, err := eth.Queue().QueueAndWait(toCtx, txn)
	if err != nil {
		logger.Errorf("Failed to retrieve snapshot receipt: %v", err)
		return nil //CAN NOT RETURN ERROR OR SUBSCRIPTION IS LOST!
	}

	if rcpt == nil {
		logger.Warnf("No receipt from snapshot")
	} else {
		if rcpt.Status != uint64(1) {
			logger.Errorf("Snapshot receipt shows failure.")
			return nil //CAN NOT RETURN ERROR OR SUBSCRIPTION IS LOST!
		}
	}

	return nil
}
