package evidence

import (
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/lstate"
	"github.com/MadBase/MadNet/constants"
	"github.com/dgraph-io/badger/v2"
)

const defaultMax = 2000

// Pool cleans up stale records
type Pool struct {
	database *db.Database
	store    *lstate.Store
	max      int
}

// NewPool backed by database.
func NewPool(database *db.Database) *Pool {
	return &Pool{
		database: database,
		store:    lstate.New(database),
		max:      defaultMax,
	}
}

// Cleanup the evidence pool.
func (p *Pool) Cleanup() error {
	return p.database.Update(func(txn *badger.Txn) error {
		_, _, _, height, _, err := p.store.GetDropData(txn)
		if err != nil {
			return err
		}
		if height > constants.EpochLength*5 {
			dropHeight := height - constants.EpochLength*4
			return p.database.DeleteBeforeHistoricRoundState(txn, dropHeight, p.max)
		}
		return nil
	})
}
