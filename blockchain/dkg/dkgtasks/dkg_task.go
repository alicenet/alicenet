package dkgtasks

import (
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"math/rand"
	"time"
)

type DkgTask struct {
	Start      uint64
	End        uint64
	State      *objects.DkgState
	Success    bool
	TxReplOpts *TxReplOpts
}

type TxReplOpts struct {
	TxHash    common.Hash
	Nonce     *big.Int
	GasFeeCap *big.Int
	GasTipCap *big.Int
}

type DkgTaskIfase interface {
	GetDkgTask() *DkgTask
	SetDkgTask(*DkgTask)
}

func (d *DkgTask) Clear() {
	var emptyHash [32]byte
	d.TxReplOpts.TxHash = emptyHash
	d.TxReplOpts.Nonce = nil
	d.TxReplOpts.GasFeeCap = nil
	d.TxReplOpts.GasTipCap = nil
}

func NewDkgTask(state *objects.DkgState, start uint64, end uint64) *DkgTask {
	return &DkgTask{
		State:      state,
		Start:      start,
		End:        end,
		Success:    false,
		TxReplOpts: &TxReplOpts{},
	}
}

func GetRandomBool() bool {
	rand.Seed(time.Now().UnixNano())
	min := 1
	max := 5

	return rand.Intn(max-min+1)+min == 3
}
