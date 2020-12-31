package monitor

import (
	"bytes"
	"context"
	"encoding/gob"

	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

var stateKey = []byte("monitorStateKey")

// Database describes required functionality for monitor persistence
type Database interface {
	FindState() (*State, error)
	UpdateState(state *State) error
}

type monitorDB struct {
	database *badger.DB
	logger   *logrus.Logger
}

// NewDatabase initializes a new monitor database
func NewDatabase(ctx context.Context, directoryName string, inMemory bool) Database {

	logger := logging.GetLogger("monitor_db")

	logger.Infof("Opening badger DB... In-Memory:%v Directory:%v", inMemory, directoryName)

	opts := badger.DefaultOptions(directoryName).WithInMemory(inMemory)
	opts.Logger = logger

	db, err := badger.Open(opts)
	if err != nil {
		logger.Panicf("Could not open database: %v", err)
	}

	go func() {
		defer db.Close()
		<-ctx.Done()
	}()

	return &monitorDB{
		logger:   logger,
		database: db}
}

func NewDatabaseFromExisting(db *badger.DB) Database {
logger := logging.GetLogger("monitor_db")
return &monitorDB{
  logger:   logger,
  database: db}
}

func (mon *monitorDB) FindState() (*State, error) {

	state := &State{}

	fn := func(txn *badger.Txn) error {
		data, err := utils.GetValue(txn, stateKey)
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		err = dec.Decode(state)
		if err != nil {
			return err
		}
		return nil
	}

	err := mon.database.View(fn)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (mon *monitorDB) UpdateState(state *State) error {

	buf := &bytes.Buffer{}

	enc := gob.NewEncoder(buf)
	err := enc.Encode(state)
	if err != nil {
		return err
	}

	fn := func(txn *badger.Txn) error {
		return utils.SetValue(txn, stateKey, buf.Bytes())
	}

	return mon.database.Update(fn)
}
