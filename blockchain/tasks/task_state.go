package tasks

import (
	"math/big"
	"strings"
	"sync"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/ethereum/go-ethereum/common"
)

type Task struct {
	sync.RWMutex
	Start          uint64
	End            uint64
	State          interfaces.ITaskState
	Success        bool
	StartBlockHash common.Hash
	TxOpts         *TxOpts
}

var _ interfaces.ITaskExecutionData = &Task{}

type TxOpts struct {
	TxHashes     []common.Hash
	Nonce        *big.Int
	GasFeeCap    *big.Int
	GasTipCap    *big.Int
	MinedInBlock uint64
}

func (t *TxOpts) GetHexTxsHashes() string {
	var hashes strings.Builder
	for _, txHash := range t.TxHashes {
		hashes.WriteString(txHash.Hex())
		hashes.WriteString(" ")
	}
	return hashes.String()
}

func NewTask(state interfaces.ITaskState, start uint64, end uint64) *Task {
	return &Task{
		State:   state,
		Start:   start,
		End:     end,
		Success: false,
		TxOpts:  &TxOpts{TxHashes: make([]common.Hash, 0)},
	}
}

func (d *Task) ClearTxData() {
	d.Lock()
	defer d.Unlock()
	d.TxOpts = &TxOpts{
		TxHashes: make([]common.Hash, 0),
	}
}

func (d *Task) GetStart() uint64 {
	d.Lock()
	defer d.Unlock()
	return d.Start
}

func (d *Task) GetEnd() uint64 {
	d.Lock()
	defer d.Unlock()
	return d.End
}
