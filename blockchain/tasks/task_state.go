package tasks

import (
	"context"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"
	"sync"
)

type Task struct {
	sync.RWMutex
	Id             string
	Start          uint64
	End            uint64
	State          interfaces.ITaskState
	Success        bool
	Ctx            context.Context
	Cf             context.CancelFunc
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

func (to *TxOpts) GetHexTxsHashes() string {
	var hashes strings.Builder
	for _, txHash := range to.TxHashes {
		hashes.WriteString(txHash.Hex())
		hashes.WriteString(" ")
	}
	return hashes.String()
}

func NewTask(state interfaces.ITaskState, start uint64, end uint64) *Task {
	ctx, cf := context.WithCancel(context.Background())

	return &Task{
		State:   state,
		Start:   start,
		End:     end,
		Success: false,
		Ctx:     ctx,
		Cf:      cf,
		TxOpts:  &TxOpts{TxHashes: make([]common.Hash, 0)},
	}
}

func (t *Task) WithContext(ctx context.Context, cancel context.CancelFunc) *Task {
	t.Ctx = ctx
	t.Cf = cancel
	return t
}

func (t *Task) ClearTxData() {
	t.Lock()
	defer t.Unlock()
	t.TxOpts = &TxOpts{
		TxHashes: make([]common.Hash, 0),
	}
}

func (t *Task) GetStart() uint64 {
	t.RLock()
	defer t.RUnlock()
	return t.Start
}

func (t *Task) GetEnd() uint64 {
	t.RLock()
	defer t.RUnlock()
	return t.End
}

func (t *Task) SetId(id string) {
	t.Lock()
	defer t.Unlock()
	t.Id = id
}

func (t *Task) Close() {
	t.Lock()
	defer t.Unlock()
	if t.Cf != nil {
		t.Cf()
	}
}
