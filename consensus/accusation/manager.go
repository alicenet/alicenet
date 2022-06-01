package accusation

import (
	"errors"
	"sync"
	"time"

	"github.com/MadBase/MadNet/blockchain/interfaces"
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
	adminHandler interfaces.AdminHandler
	// isSynchronized *remoteVar
}

func NewManager(adminHandler interfaces.AdminHandler, database *db.Database, logger *logrus.Logger) *Manager {
	detectorLogics := make([]detectorLogic, 0)
	detectorLogics = append(detectorLogics, detectMultipleProposal)
	detectorLogics = append(detectorLogics, detectDoubleSpend)

	detector := NewDetector(nil, detectorLogics)
	m := &Manager{
		detector:     detector,
		database:     database,
		logger:       logger,
		adminHandler: adminHandler,
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
	var height uint32 = 1
	var round uint32 = 1
	for {
		time.Sleep(10 * time.Millisecond)

		// only poll data if the node is synchronized
		// todo: can we get this without requiring the adminHandler here? maybe a remoteVar()?
		if !m.adminHandler.IsSynchronized() {
			m.logger.Infof("AccusationManager: admin.Handler is not synchronized, skipping round state poll")
			continue
		}

		// fetch round states from DB
		rss, nextHeight, nextRound, err := m.fetchRoundStates(height, round)
		if err != nil {
			m.logger.Infof("AccusationManager could not poll roundStates: %v", err)
			continue
		}

		if len(rss) > 0 {

			m.logger.Infof("AccusationManager: polling %d roundStates at height %d and round %d", len(rss), height, round)

			for _, rs := range rss {
				// send round states to detector to be processed
				m.detector.HandleRoundState(rs)
			}
		}

		height = nextHeight
		round = nextRound
	}
}

func (m *Manager) fetchRoundStates(height uint32, round uint32) ([]*objs.RoundState, uint32, uint32, error) {
	roundStates := make([]*objs.RoundState, 0)

	err := m.database.View(func(txn *badger.Txn) error {
		// todo: load current polling height and round from DB, to resume operations in case of a node restart
		//rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Round
		// hdr, err := m.database.GetCommittedBlockHeader(txn, height)
		// if err != nil {
		// 	m.logger.Errorf("AccusationManager could not fetch committed block header: %v", err)
		// 	return err
		// }
		// m.logger.Infof("AccusationManager: fetching broadcast header at height %d", hdr.BClaims.Height)

		vs, err := m.database.GetValidatorSet(txn, height)
		if err != nil {
			//m.logger.Errorf("AccusationManager could not fetch validator set: %t %v", err, err)
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil
			}
			return err
		}

		for _, validator := range vs.Validators {
			// todo: check if it's better to get historic or current round states
			rs, err := m.database.GetHistoricRoundState(txn, validator.VAddr, height, round)
			if err != nil {
				m.logger.Errorf("AccusationManager could not fetch historic round state: %v", err)
				return err
			}

			roundStates = append(roundStates, rs)
		}

		return nil
	})

	if err != nil {
		return nil, height, round, err
	}

	// todo: save current polling height and round from DB, to resume operations in case of a node restart

	// compute next height and round
	if round >= 4 {
		round = 1
		height++
	} else {
		round++
	}

	return roundStates, height, round, nil
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
