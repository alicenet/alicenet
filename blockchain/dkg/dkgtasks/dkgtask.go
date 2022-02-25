package dkgtasks

import (
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type DkgTask struct {
	Start   uint64
	End     uint64
	State   *objects.DkgState
	Success bool
	TxOpts  *bind.TransactOpts
	TxHash  common.Hash
}

type DkgTaskIfase interface {
	GetDkgTask() *DkgTask
	SetDkgTask(*DkgTask)
}
