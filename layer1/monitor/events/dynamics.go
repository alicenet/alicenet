package events

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/layer1"
)

// ProcessValueUpdated handles a dynamic value updating coming from our smart contract.
func ProcessValueUpdated(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, log types.Log, monDB *db.Database) error {
	logger.Info("ProcessValueUpdated() ...")

	event, err := contracts.EthereumContracts().Governance().ParseValueUpdated(log)
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
