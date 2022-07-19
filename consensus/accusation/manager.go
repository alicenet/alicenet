package accusation

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

// a function that returns an Accusation interface object when found, and a bool indicating if an accusation has been found (true) or not (false)
type detector = func(rs *objs.RoundState, lrs *lstate.RoundStates) (objs.Accusation, bool)

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
	sync.Mutex                                    // this is currently being used by workers when interacting with rsCache
	detectionPipeline               []detector    // the pipeline of detector functions
	database                        *db.Database  // the database to store detected accusations
	sstore                          *lstate.Store // the state store to get round states from
	logger                          *logrus.Logger
	rsCache                         map[string]*rsCacheStruct // cache of validator's roundState height, round and hash to avoid checking accusations unless anything changes
	workQ                           chan *lstate.RoundStates  // queue where new roundStates are pushed to be checked for malicious behavior by workers
	accusationQ                     chan objs.Accusation      // queue where identified accusations are pushed by workers to be further processed
	unpersistedCreatedAccusations   []objs.Accusation         // newly found accusations that where not persisted into DB
	unpersistedScheduledAccusations []objs.Accusation         // accusations that scheduled for execution and not yet persisted as such
	closeChan                       chan struct{}             // channel to signal to workers to stop
	wg                              *sync.WaitGroup           // wait group to wait for workers to stop
}

// NewManager creates a new *Manager
func NewManager(database *db.Database, sstore *lstate.Store, logger *logrus.Logger) *Manager {
	detectors := make([]detector, 0)

	m := &Manager{}
	m.detectionPipeline = detectors
	m.database = database
	m.logger = logger
	m.sstore = sstore
	m.rsCache = make(map[string]*rsCacheStruct)
	m.workQ = make(chan *lstate.RoundStates, 1)
	m.accusationQ = make(chan objs.Accusation, 1)
	m.closeChan = make(chan struct{}, 1)
	m.unpersistedCreatedAccusations = make([]objs.Accusation, 0)
	m.unpersistedScheduledAccusations = make([]objs.Accusation, 0)
	m.wg = &sync.WaitGroup{}

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
	close(m.closeChan)
	m.wg.Wait()
	m.logger.Warn("Accusation manager stopped")
}

// runWorker is the worker function that processes the work queue
func (m *Manager) runWorker() {
	for {
		select {
		case <-m.closeChan:
			return
		case lrs := <-m.workQ:
			_, err := m.processLRS(lrs)
			if err != nil {
				m.logger.Errorf("failed to process LRS: %v", err)
			}
		}
	}
}

// Poll reads the local round states and sends it to a work queue so that
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
	case <-m.closeChan:
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
	return m.scheduleAccusations()
}

// persistCreatedAccusations persists the newly created accusations. If it
// fails to persist into the DB, it will retry persisting again later.
func (m *Manager) persistCreatedAccusations() {
	if len(m.unpersistedCreatedAccusations) > 0 {
		persistedIdx := make([]int, 0)
		for i, acc := range m.unpersistedCreatedAccusations {
			// persist the accusation into the database
			acc.SetState(objs.Persisted)
			acc.SetPersistenceTimestamp(uint64(time.Now().Unix()))
			err := m.database.Update(func(txn *badger.Txn) error {
				return m.database.SetAccusation(txn, acc)
			})
			if err != nil {
				m.logger.Errorf("AccusationManager failed to save accusation into DB: %v", err)
				continue
			}

			persistedIdx = append(persistedIdx, i)
		}

		// delete persistedIdx from m.unpersistedAccusations, iterating from the end
		for i := len(persistedIdx) - 1; i >= 0; i-- {
			idx := persistedIdx[i]
			m.unpersistedCreatedAccusations = append(m.unpersistedCreatedAccusations[:idx], m.unpersistedCreatedAccusations[idx+1:]...)
		}
	}
}

// scheduleAccusations schedules the accusations that are not yet scheduled in the Task Manager.
func (m *Manager) scheduleAccusations() error {
	// schedule Persisted but not ScheduledForExecution accusations
	var unscheduledAccusations []objs.Accusation
	var err error

	err = m.database.View(func(txn *badger.Txn) error {
		unscheduledAccusations, err = m.database.GetPersistedButUnscheduledAccusations(txn)
		return err
	})
	if err != nil {
		return err
	}

	for _, acc := range unscheduledAccusations {
		m.logger.Debugf("Scheduling accusation %#v", acc)
		/*
			// boilerplate code until we have the real scheduler from Egor, Gabriele and Leonardo
			if err := m.scheduler.AddTask(acc); err != nil {
				m.logger.Errorf("AccusationManager failed to add accusation to scheduler: %v", err)
				// if this is a retryable error, then retry the accusation later on
				// continue

				// otherwise return the error which will stop the Synchronizer loop, causing the node to exit
				// return err
			} else {
				m.logger.Debugf("AccusationManager successfully scheduled accusation for execution: %#v", acc)
				acc.SetState(objs.ScheduledForExecution)
				m.unpersistedScheduledAccusations = append(m.unpersistedScheduledAccusations, acc)
			}
		*/

		// todo: delete this placeholder code down here after the real scheduler is implemented
		acc.SetState(objs.ScheduledForExecution)
		m.unpersistedScheduledAccusations = append(m.unpersistedScheduledAccusations, acc)
	}

	if len(m.unpersistedScheduledAccusations) > 0 {
		// persist scheduled accusations
		persistedIdx := make([]int, 0)
		for i, acc := range m.unpersistedScheduledAccusations {
			err := m.database.Update(func(txn *badger.Txn) error {
				return m.database.SetAccusation(txn, acc)
			})

			if err != nil {
				m.logger.Errorf("AccusationManager failed to save scheduled accusation into DB: %v", err)
				continue
			}

			persistedIdx = append(persistedIdx, i)
		}

		// delete persistedIdx from m.unpersistedScheduledAccusations, iterating from the end
		for i := len(persistedIdx) - 1; i >= 0; i-- {
			idx := persistedIdx[i]
			m.unpersistedScheduledAccusations = append(m.unpersistedScheduledAccusations[:idx], m.unpersistedScheduledAccusations[idx+1:]...)
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

		valAddress := fmt.Sprintf("0x%x", v.VAddr)
		updated := false
		currentValidators[valAddress] = true

		m.Lock()
		rsCacheEntry, isCached := m.rsCache[valAddress]
		m.Unlock()

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
			m.Lock()
			m.rsCache[valAddress] = rsCacheEntry
			m.Unlock()
		}
	}

	// remove validators from cache that are not in the current validatorSet,
	// ensuring the cache is not growing indefinitely with old validators
	m.Lock()
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
	m.Unlock()

	return hadUpdates, nil
}

// findAccusation checks if there is an accusation for a certain roundState and if so, sends it for further processing.
func (m *Manager) findAccusation(rs *objs.RoundState, lrs *lstate.RoundStates) {
	for _, detector := range m.detectionPipeline {
		accusation, found := detector(rs, lrs)
		if found {
			m.accusationQ <- accusation
			break
		}
	}
}
