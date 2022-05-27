package accusation

import (
	"fmt"
	"sync"
	"time"

	"github.com/MadBase/MadNet/consensus/admin"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

// Manager polls validators' roundStates and forwards them to a Detector. Also handles detected accusations.
type Manager struct {
	sync.Mutex
	detector     *Detector
	database     *db.Database
	logger       *logrus.Logger
	adminHandler *admin.Handlers
	// isSynchronized *remoteVar
}

func NewManager(database *db.Database, logger *logrus.Logger) *Manager {
	detectorLogics := make([]detectorLogic, 0)
	detectorLogics = append(detectorLogics, detectMultipleProposal)
	detectorLogics = append(detectorLogics, detectDoubleSpend)

	detector := NewDetector(nil, detectorLogics)
	m := &Manager{
		detector: detector,
		database: database,
		logger:   logger,
	}
	detector.manager = m

	return m
}

func (m *Manager) Start() {
	go m.run()
}

// Stop terminates the manager and its detector
func (m *Manager) Stop() {
	// todo: stop the manager. close a channel or something
}

func (m *Manager) run() {
	// poll validators' roundStates
	for {

		// only poll data if the node is synchronized
		if !m.adminHandler.IsSynchronized() {
			time.Sleep(1 * time.Second)
			continue
		}

		// fetch round states from DB
		rss, err := m.fetchRoundStates()
		if err != nil {
			panic(fmt.Sprintf("AccusationManager could not poll roundStates: %v", err))
		}

		for _, rs := range rss {
			// send round states to detector to be processed
			m.detector.HandleRoundState(rs)
		}

		time.Sleep(1 * time.Second)
	}
}

func (m *Manager) fetchRoundStates() ([]*objs.RoundState, error) {
	roundStates := make([]*objs.RoundState, 0)

	err := m.database.View(func(txn *badger.Txn) error {
		// todo: load current polling height and round from DB, to resume operations in case of a node restart
		vs, err := m.database.GetValidatorSet(txn, 0)
		if err != nil {
			return err
		}

		for _, validator := range vs.Validators {
			// todo: check if it's better to get historic or current round states
			rs, err := m.database.GetHistoricRoundState(txn, validator.VAddr, 0, 0)
			if err != nil {
				return err
			}

			roundStates = append(roundStates, rs)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// todo: save current polling height and round from DB, to resume operations in case of a node restart

	return roundStates, nil
}

// HandleAccusation receives an accusation, stores it in the DB and sends it to the ethereum smart contracts
func (m *Manager) HandleAccusation(accusation *Accusation) error {
	// todo: store accusation in DB

	if accusation == nil {
		panic("AccusationManager received nil accusation")
	} else {
		return (*accusation).SubmitToSmartContracts()
	}
}
