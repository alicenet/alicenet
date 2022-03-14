package dkgtasks

import (
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"math/rand"
	"strings"
	"time"
)

type ExecutionData struct {
	Start   uint64
	End     uint64
	State   *objects.DkgState
	Success bool
	TxOpts  *TxOpts
}

type TxOpts struct {
	TxHashes     []common.Hash
	Nonce        *big.Int
	GasFeeCap    *big.Int
	GasTipCap    *big.Int
	MinedInBlock uint64
}

func (d *ExecutionData) Clear() {
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

func NewDkgTask(state *objects.DkgState, start uint64, end uint64) *ExecutionData {
	return &ExecutionData{
		State:   state,
		Start:   start,
		End:     end,
		Success: false,
		TxOpts:  &TxOpts{TxHashes: make([]common.Hash, 0)},
	}
}

func GetRandomBool() bool {
	rand.Seed(time.Now().UnixNano())
	min := 1
	max := 5

	return rand.Intn(max-min+1)+min == 3
}
