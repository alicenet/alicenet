package interfaces

import (
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
)

type AdminHandler interface {
	AddPrivateKey([]byte, constants.CurveSpec) error
	AddSnapshot(*objs.BlockHeader, bool) error
	AddValidatorSet(*objs.ValidatorSet) error
	RegisterSnapshotCallback(func(*objs.BlockHeader) error)
	SetSynchronized(v bool)
}
