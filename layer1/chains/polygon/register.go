package polygon

import (
	psnapshots "github.com/alicenet/alicenet/layer1/chains/polygon/tasks/snapshots"
	"github.com/alicenet/alicenet/layer1/executor/marshaller"
)

// PolygonTaskRegistry all the Tasks we can handle in the request.
// If you want to create a new task register its instance type here.
var PolygonTaskRegistry = func(tr *marshaller.TypeRegistry) *marshaller.TypeRegistry {
	tr.RegisterInstanceType(&psnapshots.SnapshotTask{})
	return tr
}
