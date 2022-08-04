package accusation

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"runtime"
	"sync"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/layer1/executor"
	"github.com/alicenet/alicenet/layer1/executor/marshaller"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

// a function that returns an Accusation interface object when found, and a bool indicating if an accusation has been found (true) or not (false)
type detector = func(rs *objs.RoundState, lrs *lstate.RoundStates, db *db.Database) (tasks.Task, bool)

// rsCacheStruct caches a validator's roundState height, round and hash to avoid checking accusations unless anything changes
type rsCacheStruct struct {
	height uint32
	round  uint32
	rsHash []byte
}

// DidChange returns true if the local round state has changed since the last time it was checked
func (r *rsCacheStruct) DidChange(rs *objs.RoundState) (bool, error) {
	rsHash, err := rs.Hash()
	if err != nil {
		return false, err
	}

	return r.height != rs.RCert.RClaims.Height ||
			r.round != rs.RCert.RClaims.Round ||
			!bytes.Equal(r.rsHash, rsHash),
		nil
}

// Manager is responsible for checking validators' roundStates for malicious behavior and accuse them for such.
// It does so by processing each roundState through a pipeline of detetor functions until either an accusation is found or the pipeline is exhausted.
// If an accusation is found, it is sent to the Scheduler/TaskManager to be processed further, e.g., invoke accusation smart contracts for these purposes.
// The AccusationManager is responsible for persisting the accusations it creates, retrying persistence,
// sending accusations to the Scheduler/TaskManager, and even retrying these actions if they fail.
// The AccusationManager is executed by the Synchronizer through the Poll() function, which reads the local round states and sends it to a work queue so that
// workers can process it, offloading the synchronizer loop.
type Manager struct {
	logger                        *logrus.Logger
	detectionPipeline             []detector                           // the pipeline of detector functions
	database                      *db.Database                         // the database to store detected accusations
	sstore                        *lstate.Store                        // the state store to get round states from
	rsCache                       map[string]*rsCacheStruct            // cache of validator's roundState height, round and hash to avoid checking accusations unless anything changes
	rsCacheLock                   sync.RWMutex                         // this is currently being used by workers when interacting with rsCache
	workQ                         chan *lstate.RoundStates             // queue where new roundStates are pushed to be checked for malicious behavior by workers
	accusationQ                   chan tasks.Task                      // queue where identified accusations are pushed by workers to be further processed
	unpersistedCreatedAccusations []tasks.Task                         // newly found accusations that where not persisted into DB
	runningAccusations            map[string]*executor.HandlerResponse // accusations scheduled for execution and not yet completed. accusation.id -> response
	wg                            *sync.WaitGroup                      // wait group to wait for workers to stop
	ctx                           context.Context                      // the context to use for the task handler and go routines
	cancelCtx                     context.CancelFunc                   // the cancel function to cancel the context
	taskHandler                   executor.TaskHandler                 // the task handler to schedule accusation tasks against the smart contracts
}

// NewManager creates a new *Manager
func NewManager(database *db.Database, sstore *lstate.Store, taskHandler executor.TaskHandler, logger *logrus.Logger) *Manager {
	detectors := make([]detector, 0)
	detectors = append(detectors, detectMultipleProposal)

	m := &Manager{}
	m.detectionPipeline = detectors
	m.database = database
	m.logger = logger
	m.sstore = sstore
	m.rsCache = make(map[string]*rsCacheStruct)
	m.workQ = make(chan *lstate.RoundStates, 1)
	m.accusationQ = make(chan tasks.Task, 1)
	m.unpersistedCreatedAccusations = make([]tasks.Task, 0)
	m.runningAccusations = make(map[string]*executor.HandlerResponse)
	m.wg = &sync.WaitGroup{}
	m.taskHandler = taskHandler
	m.ctx, m.cancelCtx = context.WithCancel(context.Background())

	return m
}

// StartWorkers starts the workers that process the work queue
func (m *Manager) StartWorkers() {
	cpuCores := runtime.NumCPU()
	for i := 0; i < cpuCores; i++ {
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			m.runWorker()
		}()
	}
}

// StopWorkers stops the workers that process the work queue
func (m *Manager) StopWorkers() {
	m.cancelCtx()
	m.wg.Wait()
	m.logger.Warn("Accusation manager stopped")
}

// runWorker is the main worker function to processes workQ roundStates
func (m *Manager) runWorker() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case lrs := <-m.workQ:
			_, err := m.processLRS(lrs)
			if err != nil {
				m.logger.Errorf("failed to process LRS: %v", err)
			}
		}
	}
}

// Poll reads and sends the local round states to a work queue so that
// workers can process it, offloading the Synchronizer loop.
func (m *Manager) Poll() error {
	// load local RoundStates
	var lrs *lstate.RoundStates
	err := m.database.View(func(txn *badger.Txn) error {
		rss, err := m.sstore.LoadLocalState(txn)
		if err != nil {
			return err
		}
		lrs = rss
		return nil
	})
	if err != nil {
		return err
	}

	// should only look into blocks after height 2 because height 1 is the genesis block
	if lrs.Height() < 2 {
		return nil
	}

	// send this lstate.RoundStates to be processed by workers
	m.workQ <- lrs

	// receive accusations from workers and send them to the Scheduler/TaskManager to be processed further.
	// this could be done in a separate goroutine
	select {
	case <-m.ctx.Done():
		m.logger.Debug("AccusationManager is now closing")
		return nil
	case acc := <-m.accusationQ:
		// an accusation has been formed and it needs to be sent to the smart contracts
		m.logger.Debugf("Got an accusation from a worker: %#v", acc)
		m.unpersistedCreatedAccusations = append(m.unpersistedCreatedAccusations, acc)
	default:
		//m.logger.Debug("AccusationManager did not find an accusation")
	}

	m.persistCreatedAccusations()
	err = m.scheduleAccusations()
	if err != nil {
		m.logger.Warnf("AccusationManager failed to schedule accusations: %v", err)
		return err
	}

	return m.handleCompletedAccusations()
}

// persistCreatedAccusations persists the newly created accusations. If it
// fails to persist into the DB, it will retry persisting again later.
func (m *Manager) persistCreatedAccusations() {
	if len(m.unpersistedCreatedAccusations) > 0 {
		persistedIdx := make([]int, 0)
		for i, acc := range m.unpersistedCreatedAccusations {
			// persist the accusation into the database
			err := m.database.Update(func(txn *badger.Txn) error {
				// todo: put this marshalling into a separate function
				var id [32]byte
				idBin, err := hex.DecodeString(acc.GetId())
				if err != nil {
					return err
				}
				copy(id[:], idBin)
				data, err := marshaller.GobMarshalBinary(acc)
				if err != nil {
					return err
				}
				return m.database.SetAccusationRaw(txn, id, data)
			})
			if err != nil {
				m.logger.Errorf("AccusationManager failed to save accusation into DB: %v", err)
				continue
			}

			persistedIdx = append(persistedIdx, i)
		}

		// delete persistedIdx from m.unpersistedCreatedAccusations, iterating from the end
		for i := len(persistedIdx) - 1; i >= 0; i-- {
			idx := persistedIdx[i]
			m.unpersistedCreatedAccusations = append(m.unpersistedCreatedAccusations[:idx], m.unpersistedCreatedAccusations[idx+1:]...)
		}
	}
}

// scheduleAccusations schedules the accusations that are not yet scheduled in the Task Scheduler.
func (m *Manager) scheduleAccusations() error {
	var currentAccusations []tasks.Task

	// first retrieve all the current accusations from the database
	err := m.database.View(func(txn *badger.Txn) error {
		rawAccusations, err := m.database.GetAccusations(txn, nil)
		if err != nil {
			return err
		}

		for _, rawAcc := range rawAccusations {
			acc, err := marshaller.GobUnmarshalBinary(rawAcc)
			if err != nil {
				return err
			}
			currentAccusations = append(currentAccusations, acc)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// schedule existing accusations if they are not yet scheduled in the Task Scheduler
	for _, acc := range currentAccusations {
		//m.logger.Debugf("Scheduling accusation %#v", acc)

		// schedule if not yet scheduled
		if _, ok := m.runningAccusations[acc.GetId()]; !ok {
			m.logger.Debugf("Scheduling accusation %s", acc.GetId())
			resp, err := m.taskHandler.ScheduleTask(m.ctx, acc, acc.GetId())
			if err != nil {
				m.logger.Warnf("AccusationManager failed to schedule accusation: %v", err)
				continue
			}

			m.runningAccusations[acc.GetId()] = resp
		}
	}

	return nil
}

// handleCompletedAccusations checks for the completion of the accusations that are scheduled in the Task Scheduler.
// This function does not block while waiting for task completion. If an accusation task if completed, it is then deleted
// from the database.
func (m *Manager) handleCompletedAccusations() error {
	// check for completed accusations in m.runningAccusations ResponseHandler, and cleanup
	// the m.runningAccusations map accordingly as well as the database
	for accusationId, resp := range m.runningAccusations {
		if resp.IsReady() {
			err := resp.GetResponseBlocking(m.ctx)
			if err != nil {
				// todo: maybe check for an error indicating that the task was concluded because the accusation already took place, perhaps by another validator. or maybe in that case the task can just return OK.
				m.logger.Warnf("AccusationManager got error response for accusation task: %v", err)
				return err
			}

			// delete the response from the database
			err = m.database.Update(func(txn *badger.Txn) error {
				var id [32]byte
				copy(id[:], []byte(accusationId))
				return m.database.DeleteAccusation(txn, id)
			})
			if err != nil {
				m.logger.Warnf("AccusationManager failed to delete accusation from DB: %v", err)
				return err
			}

			// delete the completed accusation from the m.runningAccusations map
			delete(m.runningAccusations, accusationId)
		}
	}

	return nil
}

// processLRS processes the local state of the blockchain. This function
// is called by a worker. It returns a boolean indicating whether the
// local round state had updates or not, as well as an error.
func (m *Manager) processLRS(lrs *lstate.RoundStates) (bool, error) {
	// keep track of new validators to clear the cache from old validators
	currentValidators := make(map[string]bool)
	hadUpdates := false

	for _, v := range lrs.ValidatorSet.Validators {
		rs := lrs.GetRoundState(v.VAddr)
		if rs == nil {
			m.logger.Errorf("AccusationManager: could not get roundState for validator 0x%x", v.VAddr)
			continue
		}

		if rs.Proposal != nil {
			rs.Proposal.Proposer = utils.CopySlice(v.VAddr)
			rs.Proposal.GroupKey = utils.CopySlice(lrs.ValidatorSet.GroupKey)
		}
		if rs.ConflictingProposal != nil {
			rs.ConflictingProposal.Proposer = utils.CopySlice(v.VAddr)
			rs.ConflictingProposal.GroupKey = utils.CopySlice(lrs.ValidatorSet.GroupKey)
		}

		valAddress := fmt.Sprintf("0x%x", v.VAddr)
		updated := false
		currentValidators[valAddress] = true

		m.rsCacheLock.RLock()
		rsCacheEntry, isCached := m.rsCache[valAddress]
		m.rsCacheLock.RUnlock()

		if isCached {
			// validator exists in cache, let's check if there are changes in its roundState
			var err error
			updated, err = rsCacheEntry.DidChange(rs)
			if err != nil {
				return hadUpdates, err
			}
		} else {
			updated = true
			rsCacheEntry = &rsCacheStruct{
				// data will be populated down below
			}
		}

		if updated {
			hadUpdates = true
			// m.logger.WithFields(logrus.Fields{
			// 	"lrs.height":              lrs.Height(),
			// 	"lrs.round":               lrs.Round(),
			// 	"rs.RCert.RClaims.Height": rs.RCert.RClaims.Height,
			// 	"rs.RCert.RClaims.Round":  rs.RCert.RClaims.Round,
			// 	"vAddr":                   valAddress,
			// }).Debug("AccusationManager: processing roundState")

			m.findAccusation(rs, lrs)

			// update rsCache
			rsCacheEntry.height = rs.RCert.RClaims.Height
			rsCacheEntry.round = rs.RCert.RClaims.Round
			rsHash, err := rs.Hash()
			if err != nil {
				return hadUpdates, err
			}
			rsCacheEntry.rsHash = rsHash
			m.rsCacheLock.Lock()
			m.rsCache[valAddress] = rsCacheEntry
			m.rsCacheLock.Unlock()
		}
	}

	// remove validators from cache that are not in the current validatorSet,
	// ensuring the cache is not growing indefinitely with old validators
	m.rsCacheLock.Lock()
	toDelete := make([]string, 0)
	// iterate over the cache and keep track of validators not in the current validatorSet
	for vAddr := range m.rsCache {
		if _, ok := currentValidators[vAddr]; !ok {
			toDelete = append(toDelete, vAddr)
		}
	}

	// delete old validators from cache
	for _, vAddr := range toDelete {
		delete(m.rsCache, vAddr)
	}
	m.rsCacheLock.Unlock()

	return hadUpdates, nil
}

// findAccusation checks if there is an accusation for a certain roundState and if so, sends it for further processing.
func (m *Manager) findAccusation(rs *objs.RoundState, lrs *lstate.RoundStates) {
	for _, detector := range m.detectionPipeline {
		accusation, found := detector(rs, lrs, m.database)
		if found {
			m.accusationQ <- accusation
			break
		}
	}
}
