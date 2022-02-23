package dkgtasks

import (
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"time"
)

type CallOptions struct {
	TxOpts                    *bind.TransactOpts
	TxHash                    common.Hash
	TxFeePercentageToIncrease *big.Int
	TxMaxFeeThreshold         *big.Int
	TxCheckFrequency          time.Duration
	TxTimeoutForReplacement   time.Duration
}

type DkgTask struct {
	Start       uint64
	End         uint64
	State       *objects.DkgState
	Success     bool
	CallOptions CallOptions
}

type DkgTaskIfase interface {
	GetDkgTask() *DkgTask
	SetDkgTask(*DkgTask)
}
