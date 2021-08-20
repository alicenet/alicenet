package evidence

import (
	"context"

	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/lstate"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/logging"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

// Pool cleans up stale records
// Will also drive accusations in future
type Pool struct {
	database *db.Database
	sstore   *lstate.Store

	ctx       context.Context
	cancelCtx func()
	logger    *logrus.Logger
	maxnum    int
}

// Init will start the in and out gossip busses
func (ep *Pool) Init(database *db.Database) error {
	ep.logger = logging.GetLogger(constants.LoggerConsensus)
	ep.database = database
	ep.sstore = &lstate.Store{}
	err := ep.sstore.Init(database)
	if err != nil {
		ep.logger.Debugf("Error in Pool.Init at ep.sstore.Init: %v", err)
		return err
	}

	ep.maxnum = 2000
	background := context.Background()
	ctx, cf := context.WithCancel(background)
	ep.cancelCtx = cf
	ep.ctx = ctx
	return nil
}

// Done will trInger when both of the gossip busses have stopped
func (ep *Pool) Done() <-chan struct{} {
	return ep.ctx.Done()
}

// Cleanup is the run function for the pool cleanup logic
func (ep *Pool) Cleanup() error {
	ep.database.Update(func(txn *badger.Txn) error {
		_, _, _, height, _, err := ep.sstore.GetDropData(txn)
		if err != nil {
			return err
		}
		if height > constants.EpochLength*5 {
			dropHeight := height - constants.EpochLength*4
			return ep.database.DeleteBeforeHistoricRoundState(txn, dropHeight, ep.maxnum)
		}
		return nil
	})
	return nil
}

// Exit will kill the service
func (ep *Pool) Exit() {
	ep.cancelCtx()
}
