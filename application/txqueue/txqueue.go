package txqueue

import (
	"errors"

	"github.com/aalpar/deheap"

	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/utils"
)

/*
Here we describe an overview of the desired system.

We want to have transaction selection algorithm which
attempts to populate a block by prioritizing fee-dense transactions.
To enable this to be performed quickly, we want to have a prepared
list of transactions we can choose from.
All these transactions should be high value, easily added in order
of decreasing value, and are independent (that is, all possible subsets
will form a valid block).

Ordering the transactions in order of decreasing fee density
may be performed by using a heap.
One challenge is that this must be updated each time a new blockheader
is received; thus, after receiving a new block,
we must be able to *quickly* go through and purge the heap of all transactions
which contain consumed utxos.
In order to do this, we need to keep track of where each transaction
is currently in the queue;
thus, we also have two maps:
 *	one which maps a transaction hash to its heap value;
 *	another which maps utxo to transaction which consumes it.

We also allow the ability to clear the entire queue.
Presently, the thought is that this will be performed after a regular
block proposal; it may be determined that there are additional times
when this may be done.
One challenge is that we want to ensure that the queue is not cleared
too close to when a standard proposal will be performed.
We want to ensure that the queue is essentially full even after dropping
transactions from the blockheader just received.
*/

// TxQueue is the struct which will store the ordering of transactions
// in priority of txfee per cost
type TxQueue struct {
	// txheap is the Heap of transactions prioritized by
	// transaction fee per cost;
	// it is a list of TxItems
	txheap TxHeap
	// refmap is a map which maps transaction hashes (in string form)
	// to the corresponding Item to allow easy deletion
	refmap map[string]*Item
	// utxoIDs is a map which maps utxoIDs to the transaction which contains it.
	// These utxoIDs are correspond to the utxos which would be consumed
	// by the transaction it maps to.
	utxoIDs map[string]string
	// queueSize is the size of TxQueue. No additional transactions are
	// added if the total number reaches this level.
	queueSize int
}

// Init initializes the TxQueue
func (tq *TxQueue) Init(queueSize int) error {
	//tq.queueThresholdFrac = 0.75
	if err := tq.SetQueueSize(queueSize); err != nil {
		return err
	}
	tq.ClearTxQueue()
	return nil
}

// SetQueueSize sets the queue size of TxQueue.
// Note that this *does not* remove any elements
func (tq *TxQueue) SetQueueSize(queueSize int) error {
	if queueSize <= 0 {
		return errors.New("TxQueue.SetQueueSize: queueSize <= 0")
	}
	tq.queueSize = queueSize
	return nil
}

// QueueSize returns the size of TxQueue
func (tq *TxQueue) QueueSize() int {
	return tq.queueSize
}

// Add adds a tx to the TxQueue.
// If there are no consumedUTXO conflicts, we add to queue and exit;
// if there are consumedUTXO conflicts, we check values and replace only if
// value is greater.
// Here, value specifies the feeCostRatio of the corresponding tx.
// At the end, txs are removed until TxQueue is no longer full.
func (tq *TxQueue) Add(txhash []byte, value *uint256.Uint256, utxoIDs [][]byte, isCleanup bool) (bool, error) {
	if value == nil {
		return false, errors.New("TxQueue.Add: value is nil")
	}
	if len(utxoIDs) == 0 {
		return false, errors.New("TxQueue.Add: len(uxtoIDs) == 0")
	}
	if tq.Contains(txhash) {
		// We do not add a tx that is already present
		return false, nil
	}
	if tq.IsFull() {
		// The queue is full, so we do not attempt to add the tx to the queue
		// if it is below the minimum value.
		minValue, err := tq.MinValue()
		if err != nil {
			return false, err
		}
		if value.Lt(minValue) {
			return false, nil
		}
	}
	utxoIDsCopy := [][]byte{}
	for k := 0; k < len(utxoIDs); k++ {
		utxoID := utils.CopySlice(utxoIDs[k])
		utxoIDsCopy = append(utxoIDsCopy, utxoID)
	}
	conflictingTxHashes, conflict := tq.ConflictingUTXOIDs(utxoIDsCopy)
	if conflict {
		// There are conflicts with utxo sets.
		// Find all txs with conflicts and their corresponding values.
		// If current value is larger than all of them, then we drop all
		// conflicts and add this one.
		// If value is less than any, we exit and do not add tx.
		maxValue := uint256.Zero()
		for k := 0; k < len(conflictingTxHashes); k++ {
			txhash := utils.CopySlice(conflictingTxHashes[k])
			item, ok := tq.refmap[string(txhash)]
			if !ok {
				continue
			}
			if item.value.Gt(maxValue) {
				maxValue.Set(item.value)
			}
		}
		if value.Lte(maxValue) {
			// The value (feeCostRatio) of the potential tx we want to add
			// is below the value of a conflicting tx;
			// thus, we do not add it to the queue.
			return false, nil
		}
		tq.BulkDrop(conflictingTxHashes)
	}
	txString := string(txhash)
	for k := 0; k < len(utxoIDs); k++ {
		utxoID := utils.CopySlice(utxoIDs[k])
		tq.utxoIDs[string(utxoID)] = txString
	}
	item := &Item{
		txhash:    utils.CopySlice(txhash),
		value:     value.Clone(),
		utxoIDs:   utxoIDsCopy,
		isCleanup: isCleanup,
	}
	tq.refmap[txString] = item
	deheap.Push(&tq.txheap, item)
	// Remove items until not overflowing
	for tq.overflowed() {
		if _, err := tq.PopMin(); err != nil {
			return false, err
		}
	}
	return true, nil
}

// ConflictingUTXOIDs determines if the proposed additional utxos conflict
// with the present set of utxos in the TxQueue;
// returns total list of corresponding txhashes where there are conflicts.
func (tq *TxQueue) ConflictingUTXOIDs(utxoIDs [][]byte) ([][]byte, bool) {
	conflictingTxHashes := [][]byte{}
	for k := 0; k < len(utxoIDs); k++ {
		utxoID := utils.CopySlice(utxoIDs[k])
		txhashString, ok := tq.utxoIDs[string(utxoID)]
		if ok {
			conflictingTxHashes = append(conflictingTxHashes, []byte(txhashString))
		}
	}
	return conflictingTxHashes, len(conflictingTxHashes) > 0
}

// PopMax returns the txhash of the highest valued item in the TxQueue
func (tq *TxQueue) PopMax() (*Item, error) {
	if tq.IsEmpty() {
		return nil, errors.New("TxQueue.PopMax: queue is empty")
	}
	// Pop max item from TxHeap
	item := deheap.PopMax(&tq.txheap).(*Item)
	// Drop references to item in reference maps
	tq.dropReferences(item)
	return item, nil
}

// PopMin returns the txhash of the lowest valued item in the TxQueue
func (tq *TxQueue) PopMin() (*Item, error) {
	if tq.IsEmpty() {
		return nil, errors.New("TxQueue.PopMin: queue is empty")
	}
	// Pop min item from TxHeap
	item := deheap.Pop(&tq.txheap).(*Item)
	// Drop references to item in reference maps
	tq.dropReferences(item)
	return item, nil
}

// MinValue returns the minimum value in TxQueue
func (tq *TxQueue) MinValue() (*uint256.Uint256, error) {
	if tq.IsEmpty() {
		return nil, errors.New("TxQueue.MinValue: queue is empty")
	}
	minValue := new(uint256.Uint256)
	err := minValue.Set(tq.txheap[0].value)
	if err != nil {
		return nil, err
	}
	return minValue, nil
}

// Contains checks if tx is present in queue
func (tq *TxQueue) Contains(txhash []byte) bool {
	_, ok := tq.refmap[string(txhash)]
	return ok
}

// dropReferences drops all references to an item in the maps
func (tq *TxQueue) dropReferences(item *Item) {
	if item == nil {
		return
	}
	// Remove utxoIDs from utxoID map
	for k := 0; k < len(item.utxoIDs); k++ {
		utxoID := utils.CopySlice(item.utxoIDs[k])
		delete(tq.utxoIDs, string(utxoID))
	}
	// Remove txhash from txhash map
	delete(tq.refmap, string(item.txhash))
}

// BulkDrop drops a collection of txs from the TxQueue
func (tq *TxQueue) BulkDrop(txhashes [][]byte) {
	for k := 0; k < len(txhashes); k++ {
		txhash := utils.CopySlice(txhashes[k])
		tq.Drop(txhash)
	}
}

// Drop drops a tx from the TxQueue and all associated utxoIDs
func (tq *TxQueue) Drop(txhash []byte) error {
	item, ok := tq.refmap[string(txhash)]
	if !ok {
		return nil
	}
	tq.dropReferences(item)
	// Remove element from TxHeap
	_ = deheap.Remove(&tq.txheap, item.index)
	return nil
}

// DeleteMined drops all txs which include the listed utxoIDs
func (tq *TxQueue) DeleteMined(utxoIDs [][]byte) {
	// Remove all utxoIDs and the txs which include them
	for k := 0; k < len(utxoIDs); k++ {
		utxoID := utils.CopySlice(utxoIDs[k])
		refTxHash, ok := tq.utxoIDs[string(utxoID)]
		if !ok {
			continue
		}
		// Drop the tx which contains it
		tq.Drop([]byte(refTxHash))
	}
}

// ClearTxQueue clears the tx queue.
func (tq *TxQueue) ClearTxQueue() {
	tq.txheap = make(TxHeap, 0, tq.queueSize)
	tq.refmap = make(map[string]*Item, tq.queueSize)
	tq.utxoIDs = make(map[string]string, tq.queueSize)
}

// IsFull returns true if queue is at capacity
func (tq *TxQueue) IsFull() bool {
	return len(tq.refmap) >= tq.queueSize
}

// overflowed returns true if queue is above capacity
func (tq *TxQueue) overflowed() bool {
	return len(tq.refmap) >= tq.queueSize+1
}

// IsEmpty returns true if queue is empty
func (tq *TxQueue) IsEmpty() bool {
	return len(tq.refmap) == 0
}
