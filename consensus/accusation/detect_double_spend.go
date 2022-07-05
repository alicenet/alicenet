package accusation

import "github.com/alicenet/alicenet/consensus/objs"

func detectDoubleSpend(rs *objs.RoundState) (objs.Accusation, bool) {

	return nil, false
}

// assert detectDoubleSpend is of type detector
var _ detector = detectDoubleSpend
