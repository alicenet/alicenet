package events

import (
	"fmt"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/objects"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// ProcessValueUpdated handles a dynamic value updating coming from our smart contract
func ProcessValueUpdated(eth interfaces.Ethereum, logger *logrus.Entry, log types.Log,
	cdb *db.Database) error {

	logger.Info("ProcessValueUpdated() ...")

	dkgState := &objects.DkgState{}
	var err error
	err = cdb.View(func(txn *badger.Txn) error {
		dkgState, err = objects.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	if !dkgState.IsValidator {
		return nil
	}

	event, err := eth.Contracts().Governance().ParseValueUpdated(log)
	if err != nil {
		return err
	}

	logger = logger.WithFields(logrus.Fields{
		"Epoch": event.Epoch.Uint64(),
		"Key":   event.Key.String(),
		"Value": fmt.Sprintf("0x%x", event.Value),
	})

	logger.Infof("Value updated")

	logger.Warnf("Dropping dynamic value on the floor")
	return nil
}
