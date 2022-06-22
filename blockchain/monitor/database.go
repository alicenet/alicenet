package monitor

import (
	"encoding/json"
	"fmt"

	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

func getStateKey() []byte {
	return []byte("monitorStateKey")
}

// Database describes required functionality for monitor persistence
type Database interface {
	FindState() (*objects.MonitorState, error)
	UpdateState(state *objects.MonitorState) error
}

type monitorDB struct {
	database *db.Database
	logger   *logrus.Entry
}

// NewDatabase initializes a new monitor database
func NewDatabase(db *db.Database) Database {
	logger := logging.GetLogger("monitor").WithField("Component", "database")
	return &monitorDB{
		logger:   logger,
		database: db}
}

func (mon *monitorDB) FindState() (*objects.MonitorState, error) {

	state := &objects.MonitorState{}

	if err := mon.database.View(func(txn *badger.Txn) error {
		keyLabel := fmt.Sprintf("%x", getStateKey())
		mon.logger.WithField("Key", keyLabel).Infof("Looking up state")
		rawData, err := utils.GetValue(txn, getStateKey())
		if err != nil {
			return err
		}
		err = json.Unmarshal(rawData, state)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return state, nil
}

func (mon *monitorDB) UpdateState(state *objects.MonitorState) error {

	rawData, err := json.Marshal(state)
	if err != nil {
		return err
	}

	err = mon.database.Update(func(txn *badger.Txn) error {
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

	if err := mon.database.Sync(); err != nil {
		mon.logger.Error("Failed to set sync")
		return err
	}

	return nil
}
