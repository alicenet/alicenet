package accusation

import (
	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/consensus/objs"
)

func detectDoubleSpend(rs *objs.RoundState, lrs *lstate.RoundStates) (objs.Accusation, bool) {

	return nil, false
}

// assert detectDoubleSpend is of type detector
var _ detector = detectDoubleSpend
