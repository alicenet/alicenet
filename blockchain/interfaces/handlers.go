package interfaces

import (
	"math/big"

	aobjs "github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/dgraph-io/badger/v2"
)

type AdminHandler interface {
	AddPrivateKey([]byte, constants.CurveSpec) error
	AddSnapshot(header *objs.BlockHeader, safeToProceedConsensus bool) error
	AddValidatorSet(*objs.ValidatorSet) error
	RegisterSnapshotCallback(func(*objs.BlockHeader) error)
	SetSynchronized(v bool)
	IsSynchronized() bool
}

type DepositHandler interface {
	Add(*badger.Txn, uint32, []byte, *big.Int, *aobjs.Owner) error
}
