package dkgtasks

import (
	"math/big"
	"strings"

	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/ethereum/go-ethereum/common"
)

type ExecutionData struct {
	Start          uint64
	End            uint64
	State          *objects.DkgState
	Success        bool
	StartBlockHash common.Hash
	TxOpts         *TxOpts
}

type TxOpts struct {
	TxHashes     []common.Hash
	Nonce        *big.Int
	GasFeeCap    *big.Int
	GasTipCap    *big.Int
	MinedInBlock uint64
}

func (d *ExecutionData) Clear() {
	d.State.Lock()
	defer d.State.Unlock()
	d.TxOpts = &TxOpts{
		TxHashes: make([]common.Hash, 0),
	}
}

func (t *TxOpts) GetHexTxsHashes() string {
	var hashes strings.Builder
	for _, txHash := range t.TxHashes {
		hashes.WriteString(txHash.Hex())
		hashes.WriteString(" ")
	}
	return hashes.String()
}

func NewExecutionData(state *objects.DkgState, start uint64, end uint64) *ExecutionData {
	return &ExecutionData{
		State:   state,
		Start:   start,
		End:     end,
		Success: false,
		TxOpts:  &TxOpts{TxHashes: make([]common.Hash, 0)},
	}
}
