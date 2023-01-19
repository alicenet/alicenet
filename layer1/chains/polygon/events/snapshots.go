package events

import (
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/executor"
	monInterfaces "github.com/alicenet/alicenet/layer1/monitor/interfaces"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// ProcessLightSnapshotTaken handles receiving snapshots
func ProcessLightSnapshotTaken(
	contracts layer1.AllSmartContracts,
	logger *logrus.Entry,
	log types.Log,
	adminHandler monInterfaces.AdminHandler,
	taskHandler executor.TaskHandler,
) error {
	logger.Info("ProcessLightSnapshotTaken() ...")

	c := contracts.PolygonContracts()

	event, err := c.LightSnapshots().ParseSnapshotTaken(log)
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
		"IsSafeToProceedConsensus": safeToProceedConsensus})
	logger.Info("LightSnapshot taken")

	return nil
}
