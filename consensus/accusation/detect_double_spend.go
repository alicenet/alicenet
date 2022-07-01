package accusation

import "github.com/MadBase/MadNet/consensus/objs"

func detectDoubleSpend(rs *objs.RoundState) (objs.Accusation, bool) {

	return nil, false
}

// assert detectDoubleSpend is of type detector
var _ detector = detectDoubleSpend
