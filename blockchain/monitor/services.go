package monitor

import (
	"context"
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
	"github.com/ethereum/go-ethereum/core/types"
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
	taskManager       tasks.Manager
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
		taskManager:       tasks.NewManager(),
	}

	// Register handlers for known events, if this failed we really can't continue
	if err := SetupEventMap(svcs.eventMap, svcs.consensusDb, dph, ah); err != nil {
		panic(err)
	}

	ah.RegisterSnapshotCallback(svcs.PersistSnapshot) // HUNTER: moved out of main func and into constructor

	return svcs
}

// WatchEthereum checks state of Ethereum and processes interesting conditions
func (svcs *Services) WatchEthereum(state *objects.MonitorState) error {
	logger := svcs.logger
	eth := svcs.eth

	ctx, cancelFunc := eth.GetTimeoutContext()
	defer cancelFunc()

	// This is making sure Ethereum endpoint has synced and has peers
	// -- This doesn't care if _we_ are insync with Ethereum
	err := svcs.EndpointInSync(ctx, state)
	if err != nil {
		logger.Warnf("Failed checking if endpoint is synchronized: %v", err)
		state.CommunicationFailures++
		if state.CommunicationFailures >= uint32(svcs.eth.RetryCount()) {
			state.InSync = false
			svcs.ah.SetSynchronized(false)
		}
		return nil
	}
	state.CommunicationFailures = 0

	// If Ethereum is not in synced, it isn't an error but we can't go on
	if !state.EthereumInSync {
		s := state.Diff(state)
		if len(s) > 0 {
			logger.Warnf("...Ethereum endpoint not ready %s", s)
		}
		return nil
	}

	err = svcs.UpdateProgress(ctx, state)
	if err != nil {
		return err
	}

	// Decide what events to look for
	firstBlock := state.HighestBlockProcessed + 1
	lastBlock := state.HighestBlockProcessed + uint64(svcs.batchSize) // Be optimistic

	// Make sure we weren't too optimistic...
	finalizedHeight, err := eth.GetFinalizedHeight(ctx)
	if err != nil {
		return err
	}

	// This could happen if finality delay is too small
	if state.HighestBlockProcessed > finalizedHeight {
		logger.Warnf("Chain height shrank. Processed %v blocks but only %v are finalized.", state.HighestBlockProcessed, finalizedHeight)
		return nil
	}

	// Don't process anything past the finalized height
	if lastBlock > finalizedHeight {
		lastBlock = finalizedHeight
	}

	// No need to look for events if we're caught up
	if lastBlock >= firstBlock {

		logsByBlock := make(map[uint64][]types.Log)

		// Grab all the events in range
		logs, err := svcs.eth.GetEvents(ctx, firstBlock, lastBlock, svcs.contractAddresses)
		if err != nil {
			return err
		}

		// Find the blocks with events
		for _, log := range logs {
			bn := log.BlockNumber
			if la, ok := logsByBlock[bn]; ok {
				logsByBlock[bn] = append(la, log)
			} else {
				logsByBlock[bn] = []types.Log{log}
			}
		}

		// Interesting blocks can change based on an event, so we need to look at all blocks in range in order
		for block := firstBlock; block <= lastBlock; block++ {

			logEntry := logger.WithField("Block", block)

			// If current block has any events, we process all of them
			if logs, present := logsByBlock[block]; present {
				for _, log := range logs {

					eventSelector := log.Topics[0].String()

					logEntry = logEntry.WithField("EventSelector", eventSelector)

					ei, ok := svcs.eventMap.Lookup(eventSelector)
					if ok {

						logEntry = logEntry.WithField("EventName", ei.Name)

						logEntry.Info("Found event handler")

						if ei.Processor != nil {
							err := ei.Processor(eth, logEntry, state, log)
							if err != nil {
								logEntry.Errorf("Event handler failed: %v", err)
							}
						}
					} else {
						logEntry.Info("No event handler found")
					}
				}
			}

			// Check if any tasks are scheduled
			uuid, err := state.Schedule.Find(block)
			if err == nil {
				task, _ := state.Schedule.Retrieve(uuid)
				log := logEntry.WithField("TaskID", uuid.String())

				svcs.taskManager.StartTask(log, eth, task)

				state.Schedule.Remove(uuid)
			}

			state.HighestBlockProcessed = lastBlock
		}

		if lastBlock < finalizedHeight {
			state.InSync = false
			svcs.ah.SetSynchronized(false)
		} else {
			state.InSync = true
			svcs.ah.SetSynchronized(true)
		}

	}

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

// AbortETHDKG does the required cleanup to stop a round of ETHDKG
// func AbortETHDKG(ethdkg *EthDKGState) {
// 	handlers := []tasks.TaskHandler{
// 		ethdkg.RegistrationTH,
// 		ethdkg.ShareDistributionTH,
// 		ethdkg.DisputeTH,
// 		ethdkg.KeyShareSubmissionTH,
// 		ethdkg.MPKSubmissionTH,
// 		ethdkg.GPKJSubmissionTH,
// 		ethdkg.GPKJGroupAccusationTH,
// 		ethdkg.CompleteTH}

// 	// We need to cancel any handler that might be running
// 	for _, handler := range handlers {
// 		if handler != nil {
// 			handler.Cancel()
// 		}
// 	}

// 	// Erase the schedule
// 	ethdkg.Schedule = &EthDKGSchedule{}
// }

// ETHDKGInProgress indicates if ETHDKG is currently running
// func ETHDKGInProgress(ethdkg *EthDKGState, currentBlock uint64) bool {
// 	if ethdkg == nil {
// 		return false
// 	}

// 	return currentBlock >= ethdkg.Schedule.RegistrationStart &&
// 		currentBlock <= ethdkg.Schedule.CompleteEnd
// }
