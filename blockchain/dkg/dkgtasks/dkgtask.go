package dkgtasks

import (
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type DkgTask struct {
	Start             uint64
	End               uint64
	State             *objects.DkgState
	Success           bool
	TxReplacementOpts *TxReplacementOpts
}

type TxReplacementOpts struct {
	Nonce     *big.Int
	GasFeeCap *big.Int
	GasTipCap *big.Int
	TxHash    common.Hash
}

type DkgTaskIfase interface {
	GetDkgTask() *DkgTask
	SetDkgTask(*DkgTask)
}

func (t *TxReplacementOpts) Clean() {
	t.Nonce = nil
	t.GasFeeCap = nil
	t.GasTipCap = nil
	var emptyHash [32]byte
	t.TxHash = emptyHash
}
