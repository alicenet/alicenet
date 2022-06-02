package events

import (
	"fmt"
	dkgObjects "github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/objects"

	ethereumInterfaces "github.com/MadBase/MadNet/blockchain/ethereum/interfaces"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// ProcessValueUpdated handles a dynamic value updating coming from our smart contract
func ProcessValueUpdated(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, log types.Log,
	cdb *db.Database) error {

	logger.Info("ProcessValueUpdated() ...")

	dkgState := &dkgObjects.DkgState{}
	var err error
	err = cdb.View(func(txn *badger.Txn) error {
		dkgState, err = dkgObjects.LoadEthDkgState(txn, logger)
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
