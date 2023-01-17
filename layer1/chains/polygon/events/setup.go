package events

import (
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/layer1/executor"
	monInterfaces "github.com/alicenet/alicenet/layer1/monitor/interfaces"
	"github.com/alicenet/alicenet/layer1/monitor/objects"
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
	// snapshots
	return nil
}
