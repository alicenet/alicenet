package txqueue

import (
	"container/heap"

	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/utils"
)

// Item is an item storing tx information for the tx heap
type Item struct {
	// txhash is the transaction hash
	txhash []byte
	// value is the scaled transaction fee per cost
	value *uint256.Uint256
	// utxoIDs is the list of utxoIDs consumed by this transaction
	utxoIDs [][]byte
	// isCleanup stores whether the tx is a cleanup tx
	isCleanup bool
	// index is the index in the heap
	index int
}

// TxHash returns a copy of the txhash for the item
func (ti *Item) TxHash() []byte {
	return utils.CopySlice(ti.txhash)
}

// UTXOIDs returns a copy of all the utxoIDs which are consumed by the item
func (ti *Item) UTXOIDs() [][]byte {
	utxoIDsCopy := [][]byte{}
	for k := 0; k < len(ti.utxoIDs); k++ {
		utxoIDsCopy = append(utxoIDsCopy, utils.CopySlice(ti.utxoIDs[k]))
	}
	return utxoIDsCopy
}

// TxHeap is a slice of TxItems
type TxHeap []*Item

var _ heap.Interface = (*TxHeap)(nil)

func (txh TxHeap) Len() int {
	return len(txh)
}

func (txh TxHeap) Less(i, j int) bool {
	return txh[i].value.Lt(txh[j].value)
}

func (txh TxHeap) Swap(i, j int) {
	txh[i], txh[j] = txh[j], txh[i]
	txh[i].index = i
	txh[j].index = j
}

// Push adds another item to the heap
func (txh *TxHeap) Push(x interface{}) {
	n := len(*txh)
	item := x.(*Item)
	item.index = n
	*txh = append(*txh, item)
}

// Pop returns the highest priority item from the heap and removes it
func (txh *TxHeap) Pop() interface{} {
	old := *txh
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*txh = old[0 : n-1]
	return item
}
