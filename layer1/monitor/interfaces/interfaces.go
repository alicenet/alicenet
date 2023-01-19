package interfaces

import (
	"math/big"

	"github.com/dgraph-io/badger/v2"

	aobjs "github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
)

type SnapshotCallbackRegistration = func(bh *objs.BlockHeader, numOfValidators, validatorIndex int) error

type AdminHandler interface {
	AddPrivateKey([]byte, constants.CurveSpec) error
	AddSnapshot(header *objs.BlockHeader, safeToProceedConsensus bool) error
	UpdateDynamicStorage(epoch uint32, rawDynamics []byte) error
	AddValidatorSet(*objs.ValidatorSet) error
	RegisterSnapshotCallback([]SnapshotCallbackRegistration)
	SetSynchronized(v bool)
	IsSynchronized() bool
}

type DepositHandler interface {
	Add(*badger.Txn, uint32, []byte, *big.Int, *aobjs.Owner) error
}

type AdminClient interface {
	SetAdminHandler(AdminHandler)
}
