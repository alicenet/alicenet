package accusation

import (
	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
)

func detectNonExistentUTXOConsumption(rs *objs.RoundState, lrs *lstate.RoundStates) (tasks.Task, bool) {

	return nil, false
}
