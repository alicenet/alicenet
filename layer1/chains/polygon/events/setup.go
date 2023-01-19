package events

import (
	"strings"

	mbindings "github.com/alicenet/alicenet/bridge/bindings/multichain"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/executor"
	monInterfaces "github.com/alicenet/alicenet/layer1/monitor/interfaces"
	"github.com/alicenet/alicenet/layer1/monitor/objects"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

func SetupEventMap(
	em *objects.EventMap,
	cdb, monDB *db.Database,
	adminHandler monInterfaces.AdminHandler,
	depositHandler monInterfaces.DepositHandler,
	taskHandler executor.TaskHandler,
	exitFunc func(),
	chainID uint32,
) error {

	// LightSnapshots.SnapshotTaken
	snapshotsEvents := GetLightSnapshotEvents()
	snapshotTakenEvent, ok := snapshotsEvents["SnapshotTaken"]
	if !ok {
		panic("could not find event LightSnapshots.SnapshotTaken")
	}

	if err := em.Register(snapshotTakenEvent.ID.String(), snapshotTakenEvent.Name,
		func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return ProcessLightSnapshotTaken(contracts, logger, log, adminHandler, taskHandler)
		}); err != nil {
		return err
	}
	return nil
}

func GetLightSnapshotEvents() map[string]abi.Event {
	lightSnapshotsABI, err := abi.JSON(strings.NewReader(mbindings.LightSnapshotsMetaData.ABI))
	if err != nil {
		panic(err)
	}

	return lightSnapshotsABI.Events
}
