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
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// a function that returns an Accusation interface object when found, and a bool indicating if an accusation has been found (true) or not (false)
type detector = func(rs *objs.RoundState) (objs.Accusation, bool)

// rsCacheStruct caches a validator's roundState height, round and hash to avoid checking for accusations unless anything changes
type rsCacheStruct struct {
	height uint32
	round  uint32
	rsHash []byte
}

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

// Manager polls validators' roundStates and checks for possible accusation conditions
type Manager struct {
	// this is currently being used by workers when interacting with rsCache
	sync.Mutex
	processingPipeline []detector
	database           *db.Database
	sstore             *lstate.Store
	logger             *logrus.Logger
	rsCache            map[string]*rsCacheStruct

	// queue where new roundStates are pushed to be checked for malicious behavior by workers
	workQ chan *lstate.RoundStates
	// queue where identified accusations are pushed by workers to be further processed
	accusationQ chan objs.Accusation
	// newly found accusations that where not persisted into DB
	unpersistedCreatedAccusations []objs.Accusation
	// accusations that scheduled for execution and not yet persisted as such
	unpersistedScheduledAccusations []objs.Accusation
	closeChan                       chan struct{}
}

// NewManager creates a new *Manager
func NewManager(database *db.Database, sstore *lstate.Store, logger *logrus.Logger) *Manager {
	detectorLogics := make([]detector, 0)
	detectorLogics = append(detectorLogics, detectMultipleProposal)
	detectorLogics = append(detectorLogics, detectDoubleSpend)

	workQ := make(chan *lstate.RoundStates, 1)
	accusationQ := make(chan objs.Accusation, 1)
	closeChan := make(chan struct{}, 1)

	m := &Manager{}
	err := m.Init(database, sstore, logger, detectorLogics, workQ, accusationQ, closeChan)
	if err != nil {
		panic(fmt.Errorf("failed to initialize Accusation.Manager: %v", err))
	}

	return m
}

func (m *Manager) Init(
	database *db.Database,
	sstore *lstate.Store,
	logger *logrus.Logger,
	detectorLogics []detector,
	workQ chan *lstate.RoundStates,
	accusationQ chan objs.Accusation,
	closeChan chan struct{}) error {

	m.processingPipeline = detectorLogics
	m.database = database
	m.logger = logger
	m.sstore = sstore
	m.rsCache = make(map[string]*rsCacheStruct)
	m.workQ = workQ
	m.accusationQ = accusationQ
	m.closeChan = closeChan
	m.unpersistedCreatedAccusations = make([]objs.Accusation, 0)
	m.unpersistedScheduledAccusations = make([]objs.Accusation, 0)

	return nil
}

func (m *Manager) StartWorkers() {
	cpuCores := runtime.NumCPU()
	for i := 0; i < cpuCores; i++ {
		go m.runWorker()
	}
}

func (m *Manager) StopWorkers() {
	m.closeChan <- struct{}{}
}

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
		acc.SetUUID(uuid.New())
		m.logger.Debugf("Got an accusation from a worker: %#v", acc)
		m.unpersistedCreatedAccusations = append(m.unpersistedCreatedAccusations, acc)
	default:
		//m.logger.Debug("AccusationManager did not find an accusation")
	}

	m.persistCreatedAccusations()
	return m.scheduleAccusations()
}

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

			m.processRoundState(rs)

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

// processRoundState checks if there is an accusation for a certain roundState and if so, sends it for further processing
func (m *Manager) processRoundState(rs *objs.RoundState) {
	for _, detector := range m.processingPipeline {
		accusation, found := detector(rs)
		if found {
			m.accusationQ <- accusation
			break
		}
	}
}
