package events

import (
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/snapshots"
	monInterfaces "github.com/alicenet/alicenet/layer1/monitor/interfaces"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// ProcessSnapshotTaken handles receiving snapshots
func ProcessSnapshotTaken(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, log types.Log, adminHandler monInterfaces.AdminHandler, taskRequestChan chan<- tasks.TaskRequest) error {
	logger = logger.WithField("method", ProcessSnapshotTaken)
	logger.Info("Processing snapshots...")

	c := contracts.EthereumContracts()

	event, err := c.Snapshots().ParseSnapshotTaken(log)
	if err != nil {
		return err
	}

	epoch := event.Epoch
	safeToProceedConsensus := event.IsSafeToProceedConsensus

	logger = logger.WithFields(logrus.Fields{
		"ChainID":                  event.ChainId,
		"Epoch":                    epoch,
		"Height":                   event.Height,
		"Validator":                event.Validator.Hex(),
		"IsSafeToProceedConsensus": event.IsSafeToProceedConsensus})
	logger.Info("Snapshot taken")

	// put it back together
	bclaims := &objs.BClaims{
		ChainID:    event.BClaims.ChainId,
		Height:     event.BClaims.Height,
		TxCount:    event.BClaims.TxCount,
		PrevBlock:  event.BClaims.PrevBlock[:],
		TxRoot:     event.BClaims.TxRoot[:],
		StateRoot:  event.BClaims.StateRoot[:],
		HeaderRoot: event.BClaims.HeaderRoot[:],
	}

	header := &objs.BlockHeader{}
	header.BClaims = bclaims
	header.SigGroup = event.SignatureRaw
	header.TxHshLst = [][]byte{}

	// send the reconstituted header to a handler
	logger.Debug("invoking adminHandler.AddSnapshot")
	err = adminHandler.AddSnapshot(header, safeToProceedConsensus)
	if err != nil {
		return err
	}

	// kill any task that might still be trying to do this snapshot
	taskRequestChan <- tasks.NewKillTaskRequest(&snapshots.SnapshotTask{})

	return nil
}
