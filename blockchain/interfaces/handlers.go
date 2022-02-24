package interfaces

import (
	"math/big"

	aobjs "github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/dgraph-io/badger/v2"
)

type AdminHandler interface {
	AddPrivateKey([]byte, constants.CurveSpec) error
	AddSnapshot(header *objs.BlockHeader, safeToProceedConsensus bool) error
	AddValidatorSet(*objs.ValidatorSet) error
	RegisterSnapshotCallback(func(*objs.BlockHeader) error)
	SetSynchronized(v bool)
}

type DepositHandler interface {
	Add(*badger.Txn, uint32, []byte, *big.Int, *aobjs.Owner) error
}
