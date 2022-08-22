/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

// The Package Trie implements a sparse merkle trie.
// TODO CRITIAL: BUILD TOOL TO WALK TRIE AND FIND MINIMUM DB COMMITMENT
// HEIGHT OF ALL NODES IN TRIE. FURTHER BUILD TOOL THAT WILL INCREMENTALLY
// PRUNE ALL NODES IN A SPECIFIED TRIE (BASED ON THE ROOT HASH) WHOSE COMMITMENT
// HEIGHT IS LESS THAN A THRESHOLD VALUE PASSED IN AS AN ARGUMENT

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

// mresult is used to contain the result of goroutines and is sent through a channel.
type mresult struct {
	update []byte
	// flag if a node was deleted and a shortcut node maybe has to move up the tree
	deleted bool
	err     error
}

// SMT is a sparse Merkle tree.
type SMT struct {
	db *cacheDB
	// Root is the current root of the smt.
	Root []byte
	// prevRoot is the root before the last update
	prevRoot []byte
	// lock is for the whole struct
	lock sync.RWMutex
	// hash is the hash function used in the trie
	hash func(data ...[]byte) []byte
	// TrieHeight is the number if bits in a key
	TrieHeight int
	// loadDbMux is a lock to protect concurrent db reads in shared txn
	loadDbMux sync.Mutex
}

func NewSMTForHeight(txn *badger.Txn, height uint32, hash func(data ...[]byte) []byte, prefixFunc func() []byte) (*SMT, error) {
	s := &SMT{
		hash:       hash,
		TrieHeight: 256,
	}
	s.db = &cacheDB{
		updatedNodes: make(map[Hash][][]byte),
		prefixFunc:   prefixFunc,
	}
	root, err := s.db.getRootForHeightDB(txn, height)
	if err != nil {
		return nil, err
	}
	s.Root = utils.CopySlice(root)
	return s, nil
}

// NewSMT creates a new SMT given a keySize and a hash function.
func NewSMT(root []byte, hash func(data ...[]byte) []byte, prefixFunc func() []byte) *SMT {
	s := &SMT{
		hash: hash,
		// hash any string to get output length
		TrieHeight: 256,
	}
	s.db = &cacheDB{
		updatedNodes: make(map[Hash][][]byte),
		prefixFunc:   prefixFunc,
	}
	s.Root = utils.CopySlice(root)
	return s
}

// Update adds a sorted list of keys and their values to the trie
// If Update is called multiple times, only the state after the last update
// is committed.
// When calling Update multiple times without commit, make sure the
// values of different keys are unique(hash contains the key for example)
// otherwise some subtree may get overwritten with the wrong hash.
func (s *SMT) Update(txn *badger.Txn, keys, values [][]byte) ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(keys) != len(values) {
		return nil, errors.New("keys and values are not of the same length")
	}
	keySet := make(map[string]bool)
	for i := 0; i < len(keys); i++ {
		keySet[string(keys[i])] = true
		if len(keys[i]) != constants.HashLen {
			return nil, errors.New("each key should be the length of the hash value")
		}
		if len(values[i]) != constants.HashLen {
			return nil, errors.New("each value should be the length of the hash value")
		}
	}
	if len(keySet) != len(keys) {
		return nil, errors.New("duplicate input key")
	}
	for i := 0; i < len(keys); i++ {
		if bytes.Equal(values[i], DefaultLeaf) {
			v, err := s.get(txn, s.Root, keys[i], nil, 0, s.TrieHeight)
			if err != nil {
				return nil, err
			}
			if v == nil {
				return nil, errors.New("missing")
			}
		} else {
			v, err := s.get(txn, s.Root, keys[i], nil, 0, s.TrieHeight)
			if err != nil {
				return nil, err
			}
			if v != nil {
				if !bytes.Equal(v, values[i]) {
					return nil, errors.New("conflicting key/value in trie")
				}
			}
		}
	}
	ch := make(chan mresult, 1)
	s.update(txn, s.Root, keys, values, nil, 0, s.TrieHeight, ch)
	result := <-ch
	if result.err != nil {
		return nil, result.err
	}
	if len(result.update) != 0 {
		s.Root = result.update[:constants.HashLen]
	} else {
		s.Root = nil
	}
	return s.Root, nil
}

// update adds and deletes a sorted list of keys and their values to the trie.
// Adding and deleting can be simultaneous.
// To delete, set the value to DefaultLeaf.
// It returns the root of the updated tree.
func (s *SMT) update(txn *badger.Txn, root []byte, keys, values, batch [][]byte, iBatch, height int, ch chan<- mresult) {
	if len(keys) == 0 || len(values) == 0 {
		err := errors.New("length of keys or values should not be zero")
		ch <- mresult{nil, false, err}
		return
	}
	if height == 0 {
		if bytes.Equal(DefaultLeaf, values[0]) {
			// Delete the key-value from the trie if it is being set to DefaultLeaf
			// The value will be set to [] in batch by maybeMoveupShortcut or interiorHash
			s.deleteOldNode(root, height, false)
			ch <- mresult{nil, true, nil}
		} else {
			// create a new shortcut batch.
			// simply storing the value will make it hard to move up the
			// shortcut in case of sibling deletion
			batch = make([][]byte, 31)
			node := s.leafHash(keys[0], values[0], root, batch, 0, height)
			ch <- mresult{node, false, nil}
		}
		return
	}

	// Load the node to update
	batch, iBatch, lnode, rnode, isShortcut, err := s.loadChildren(txn, root, height, iBatch, batch)
	if err != nil {
		ch <- mresult{nil, false, err}
		return
	}
	// Check if the keys are updating the shortcut node
	if isShortcut {
		keys, values = s.maybeAddShortcutToKV(keys, values, lnode[:constants.HashLen], rnode[:constants.HashLen])
		if iBatch == 0 {
			// shortcut is moving so it's root will change
			s.deleteOldNode(root, height, false)
		}
		// The shortcut node was added to keys and values so consider this subtree default.
		lnode, rnode = nil, nil
		// update in the batch (set key, value to default so the next loadChildren is correct)
		batch[2*iBatch+1] = nil
		batch[2*iBatch+2] = nil
		if len(keys) == 0 {
			// Set true so that a potential sibling shortcut may move up.
			ch <- mresult{nil, true, nil}
			return
		}
	}
	// Store shortcut node
	if (len(lnode) == 0) && (len(rnode) == 0) && (len(keys) == 1) {
		// We are adding 1 key to an empty subtree so store it as a shortcut
		if bytes.Equal(DefaultLeaf, values[0]) {
			ch <- mresult{nil, true, nil}
		} else {
			node := s.leafHash(keys[0], values[0], root, batch, iBatch, height)
			ch <- mresult{node, false, nil}
		}
		return
	}

	// Split the keys array so each branch can be updated in parallel
	lkeys, rkeys := s.splitKeys(keys, s.TrieHeight-height)
	splitIndex := len(lkeys)
	lvalues, rvalues := values[:splitIndex], values[splitIndex:]

	switch {
	case len(lkeys) == 0 && len(rkeys) > 0:
		s.updateRight(txn, lnode, rnode, root, keys, values, batch, iBatch, height, ch)
	case len(lkeys) > 0 && len(rkeys) == 0:
		s.updateLeft(txn, lnode, rnode, root, keys, values, batch, iBatch, height, ch)
	default:
		s.updateParallel(txn, lnode, rnode, root, lkeys, rkeys, lvalues, rvalues, batch, iBatch, height, ch)
	}
}

// updateRight updates the right side of the tree
func (s *SMT) updateRight(txn *badger.Txn, lnode, rnode, root []byte, keys, values, batch [][]byte, iBatch, height int, ch chan<- mresult) {
	// all the keys go in the right subtree
	newch := make(chan mresult, 1)
	s.update(txn, rnode, keys, values, batch, 2*iBatch+2, height-1, newch)
	result := <-newch
	if result.err != nil {
		ch <- mresult{nil, false, result.err}
		return
	}
	// Move up a shortcut node if necessary.
	if result.deleted {
		if s.maybeMoveUpShortcut(txn, lnode, result.update, root, batch, iBatch, height, ch) {
			return
		}
	}
	node := s.interiorHash(lnode, result.update, root, batch, iBatch, height)
	ch <- mresult{node, false, nil}
}

// updateLeft updates the left side of the tree
func (s *SMT) updateLeft(txn *badger.Txn, lnode, rnode, root []byte, keys, values, batch [][]byte, iBatch, height int, ch chan<- mresult) {
	// all the keys go in the left subtree
	newch := make(chan mresult, 1)
	s.update(txn, lnode, keys, values, batch, 2*iBatch+1, height-1, newch)
	result := <-newch
	if result.err != nil {
		ch <- mresult{nil, false, result.err}
		return
	}
	// Move up a shortcut node if necessary.
	if result.deleted {
		if s.maybeMoveUpShortcut(txn, result.update, rnode, root, batch, iBatch, height, ch) {
			return
		}
	}
	node := s.interiorHash(result.update, rnode, root, batch, iBatch, height)
	ch <- mresult{node, false, nil}
}

// updateParallel updates both sides of the trie simultaneously
func (s *SMT) updateParallel(txn *badger.Txn, lnode, rnode, root []byte, lkeys, rkeys, lvalues, rvalues, batch [][]byte, iBatch, height int, ch chan<- mresult) {
	lch := make(chan mresult, 1)
	rch := make(chan mresult, 1)
	go s.update(txn, lnode, lkeys, lvalues, batch, 2*iBatch+1, height-1, lch)
	go s.update(txn, rnode, rkeys, rvalues, batch, 2*iBatch+2, height-1, rch)
	lresult := <-lch
	rresult := <-rch
	if lresult.err != nil {
		ch <- mresult{nil, false, lresult.err}
		return
	}
	if rresult.err != nil {
		ch <- mresult{nil, false, rresult.err}
		return
	}

	// Move up a shortcut node if it's sibling is default
	if lresult.deleted || rresult.deleted {
		if s.maybeMoveUpShortcut(txn, lresult.update, rresult.update, root, batch, iBatch, height, ch) {
			return
		}
	}
	node := s.interiorHash(lresult.update, rresult.update, root, batch, iBatch, height)
	ch <- mresult{node, false, nil}
}

// deleteOldNode deletes an old node that has been updated
func (s *SMT) deleteOldNode(root []byte, height int, movingUp bool) {
	//var node Hash
	//copy(node[:], root)
	//if movingUp { // CHANGED WE NEVER USE ATOMIC UPDATE, SO THIS IS ALWAYS TRUE
	// dont delete old nodes with atomic updated except when
	// moving up a shortcut, we dont record every single move
	//s.db.updatedMux.Lock()
	//delete(s.db.updatedNodes, node)
	//s.db.updatedMux.Unlock()
	//}
}

// splitKeys divides the array of keys into 2 so they can update left and right branches in parallel
func (s *SMT) splitKeys(keys [][]byte, height int) ([][]byte, [][]byte) {
	for i, key := range keys {
		if bitIsSet(key, height) {
			return keys[:i], keys[i:]
		}
	}
	return keys, nil
}

// maybeMoveUpShortcut moves up a shortcut if it's sibling node is default
func (s *SMT) maybeMoveUpShortcut(txn *badger.Txn, left, right, root []byte, batch [][]byte, iBatch, height int, ch chan<- mresult) bool {
	if len(left) == 0 && len(right) == 0 {
		// Both update and sibling are deleted subtrees
		if iBatch == 0 {
			// If the deleted subtrees are at the root, then delete it.
			s.deleteOldNode(root, height, true)
		} else {
			batch[2*iBatch+1] = nil
			batch[2*iBatch+2] = nil
		}
		ch <- mresult{nil, true, nil}
		return true
	} else if len(left) == 0 {
		// If right is a shortcut move it up
		if right[constants.HashLen] == 1 {
			s.moveUpShortcut(txn, right, root, batch, iBatch, 2*iBatch+2, height, ch)
			return true
		}
	} else if len(right) == 0 {
		// If left is a shortcut move it up
		if left[constants.HashLen] == 1 {
			s.moveUpShortcut(txn, left, root, batch, iBatch, 2*iBatch+1, height, ch)
			return true
		}
	}
	return false
}

func (s *SMT) moveUpShortcut(txn *badger.Txn, shortcut, root []byte, batch [][]byte, iBatch, iShortcut, height int, ch chan<- mresult) {
	// it doesn't matter if atomic update is true or false since the batch node is modified
	_, _, shortcutKey, shortcutVal, _, err := s.loadChildren(txn, shortcut, height-1, iShortcut, batch)
	if err != nil {
		ch <- mresult{nil, false, err}
		return
	}
	// when moving up the shortcut, it's hash will change because height is +1
	newShortcut := s.hash(shortcutKey[:constants.HashLen], shortcutVal[:constants.HashLen], []byte{byte(height)})
	newShortcut = append(newShortcut, byte(1))

	if iBatch == 0 {
		// Modify batch to a shortcut batch
		batch[0] = []byte{1}
		batch[2*iBatch+1] = shortcutKey
		batch[2*iBatch+2] = shortcutVal
		batch[2*iShortcut+1] = nil
		batch[2*iShortcut+2] = nil
		// cache and updatedNodes deleted by store node
		s.storeNode(batch, newShortcut, root, height)
	} else if (height-1)%4 == 0 {
		// move up shortcut and delete old batch
		batch[2*iBatch+1] = shortcutKey
		batch[2*iBatch+2] = shortcutVal
		// set true so that AtomicUpdate can also delete a node moving up
		// otherwise every nodes moved up is recorded
		s.deleteOldNode(shortcut, height, true)
	} else {
		//move up shortcut
		batch[2*iBatch+1] = shortcutKey
		batch[2*iBatch+2] = shortcutVal
		batch[2*iShortcut+1] = nil
		batch[2*iShortcut+2] = nil
	}
	// Return the left sibling node to move it up
	ch <- mresult{newShortcut, true, nil}
}

// maybeAddShortcutToKV adds a shortcut key to the keys array to be updated.
// this is used when a subtree containing a shortcut node is being updated
func (s *SMT) maybeAddShortcutToKV(keys, values [][]byte, shortcutKey, shortcutVal []byte) ([][]byte, [][]byte) {
	newKeys := make([][]byte, 0, len(keys)+1)
	newVals := make([][]byte, 0, len(keys)+1)

	if bytes.Compare(shortcutKey, keys[0]) < 0 {
		newKeys = append(newKeys, shortcutKey)
		newKeys = append(newKeys, keys...)
		newVals = append(newVals, shortcutVal)
		newVals = append(newVals, values...)
	} else if bytes.Compare(shortcutKey, keys[len(keys)-1]) > 0 {
		newKeys = append(newKeys, keys...)
		newKeys = append(newKeys, shortcutKey)
		newVals = append(newVals, values...)
		newVals = append(newVals, shortcutVal)
	} else {
		higher := false
		for i, key := range keys {
			if bytes.Equal(shortcutKey, key) {
				if !bytes.Equal(DefaultLeaf, values[i]) {
					// Do nothing if the shortcut is simply updated
					return keys, values
				}
				// Delete shortcut if it is updated to DefaultLeaf
				newKeys = append(newKeys, keys[:i]...)
				newKeys = append(newKeys, keys[i+1:]...)
				newVals = append(newVals, values[:i]...)
				newVals = append(newVals, values[i+1:]...)
			}
			if !higher && bytes.Compare(shortcutKey, key) > 0 {
				higher = true
				continue
			}
			if higher && bytes.Compare(shortcutKey, key) < 0 {
				// insert shortcut in slices
				newKeys = append(newKeys, keys[:i]...)
				newKeys = append(newKeys, shortcutKey)
				newKeys = append(newKeys, keys[i:]...)
				newVals = append(newVals, values[:i]...)
				newVals = append(newVals, shortcutVal)
				newVals = append(newVals, values[i:]...)
				break
			}
		}
	}
	return newKeys, newVals
}

// loadChildren looks for the children of a node.
// if the node is not stored in cache, it will be loaded from db.
func (s *SMT) loadChildren(txn *badger.Txn, root []byte, height, iBatch int, batch [][]byte) ([][]byte, int, []byte, []byte, bool, error) {
	isShortcut := false
	if height%4 == 0 {
		if len(root) == 0 {
			// create a new default batch
			batch = make([][]byte, 31)
			batch[0] = []byte{0}
		} else {
			var err error
			batch, err = s.loadBatch(txn, root)
			if err != nil {
				return nil, 0, nil, nil, false, err
			}
		}
		iBatch = 0
		if batch[0][0] == 1 {
			isShortcut = true
		}
	} else {
		if len(batch[iBatch]) != 0 && batch[iBatch][constants.HashLen] == 1 {
			isShortcut = true
		}
	}
	return batch, iBatch, batch[2*iBatch+1], batch[2*iBatch+2], isShortcut, nil
}

// loadBatch fetches a batch of nodes in cache or db
func (s *SMT) loadBatch(txn *badger.Txn, root []byte) ([][]byte, error) {
	var node Hash
	copy(node[:], root)
	// checking updated nodes is useful if get() or update() is called twice in a row without db commit
	s.db.updatedMux.RLock()
	val, exists := s.db.updatedNodes[node]
	s.db.updatedMux.RUnlock()
	if exists {
		newVal := make([][]byte, 31)
		copy(newVal, val)
		return newVal, nil
	}
	s.loadDbMux.Lock()
	dbval, err := s.db.getNodeDB(txn, root[:constants.HashLen])
	s.loadDbMux.Unlock()
	if err != nil {
		return nil, err
	}
	nodeSize := len(dbval)
	if nodeSize != 0 {
		return s.parseBatch(dbval)
	}
	return nil, fmt.Errorf("the trie node %x is unavailable in the disk db, db may be corrupted", root)
}

// parseBatch decodes the byte data into a slice of nodes and bitmap
func (s *SMT) parseBatch(val []byte) ([][]byte, error) {
	batch := make([][]byte, 31)
	if len(val) == 0 {
		return nil, errors.New("length of input value should be higher than zero")
	}
	if len(val) < 4 {
		return nil, errors.New("length of input value should be at least 4 to contain the bitmap")
	}
	// fmt.Println("length of val is", len(val))
	bitmap := val[:4]
	// check if the batch root is a shortcut
	if bitIsSet(val, 31) {
		if len(val) < 70 {
			return nil, fmt.Errorf("length of input value should be seventy-one for a shortcut: got %v %x", len(val), val)
		}
		batch[0] = []byte{1}
		batch[1] = val[4 : 4+33]
		batch[2] = val[4+33 : 4+33*2]
	} else {
		j := 0
		maxVal := 0
		for i := 1; i <= 30; i++ {
			if bitIsSet(bitmap, i-1) {
				j++
			}
		}
		maxVal = 4 + 33*j
		if len(val) < maxVal {
			return nil, errors.New("length of input value is lower than it should be")
		}
		batch[0] = []byte{0}
		j = 0
		for i := 1; i <= 30; i++ {
			if bitIsSet(bitmap, i-1) {
				batch[i] = val[4+33*j : 4+33*(j+1)]
				j++
			}
		}
	}
	return batch, nil
}

// check ||| take in height with respect to height of tree in general
// take in root as an arg instead of over-writing batch[0]
// make a bool flag as last arg and have this flag specify classic mode vs leaf height hash mode
func (s *SMT) verifyBatch(batch [][]byte, idx, subtreeHeight, height int, root []byte, useUniformLeafHeight bool) ([]byte, bool) {
	lidx := idx*2 + 1
	ridx := idx*2 + 2

	llidx := lidx*2 + 1
	lridx := lidx*2 + 2

	rlidx := ridx*2 + 1
	rridx := ridx*2 + 2

	var res []byte
	equiv := false

	if !bytes.Equal(batch[0], []byte{0}) {
		if len(batch) < 3 {
			return s.hash([]byte{}), false
		}
		for i := 3; i < len(batch); i++ {
			if len(batch[i]) != 0 {
				return s.hash([]byte{}), false
			}
		}
		res = s.hash(batch[1][:constants.HashLen], batch[2][:constants.HashLen], []byte{byte(height + subtreeHeight - 1)})
		return res, bytes.Equal(res, root)
	}

	if subtreeHeight > 0 {
		var lres, rres []byte
		var eq bool
		// if left child has children, call the function to go lower
		if llidx < 32 && lridx < 32 {
			if len(batch[llidx]) != 0 || len(batch[lridx]) != 0 {
				// fmt.Println("launching vb for lc")
				lres, eq = s.verifyBatch(batch, lidx, subtreeHeight-1, height, root, useUniformLeafHeight)
				if !eq {
					return nil, false
				}
			}
		}

		// if right child has children, call the function to go lower
		if rlidx < 32 && rridx < 32 {
			if len(batch[rlidx]) != 0 || len(batch[rridx]) != 0 {
				// fmt.Println("launching vb for rc")
				rres, eq = s.verifyBatch(batch, ridx, subtreeHeight-1, height, root, useUniformLeafHeight)
				if !eq {
					return nil, false
				}
			}
		}

		// if children are leaves then get hash and return
		if len(lres) == 0 && len(rres) == 0 {

			if batch[idx][32] == 1 && idx < 15 {
				res = s.hash(batch[lidx][:constants.HashLen], batch[ridx][:constants.HashLen], []byte{byte(height + subtreeHeight)})

			} else if len(batch[ridx]) > 0 && len(batch[lidx]) > 0 {
				res = s.hash(batch[lidx][:constants.HashLen], batch[ridx][:constants.HashLen])
			} else if len(batch[ridx]) > 0 {
				res = s.hash(DefaultLeaf, batch[ridx][:constants.HashLen])
			} else if len(batch[lidx]) > 0 {
				res = s.hash(batch[lidx][:constants.HashLen], DefaultLeaf)
			}

		} else if len(lres) > 0 && len(rres) > 0 { // hashing our way back up the tree
			res = s.hash(lres, rres)
		} else if len(lres) > 0 {
			res = s.hash(lres, DefaultLeaf)
		} else if len(rres) > 0 {
			res = s.hash(DefaultLeaf, rres)
		}

		if idx == 0 {
			equiv = bytes.Equal(res, root)
		} else {
			equiv = bytes.Equal(res, batch[idx][:constants.HashLen])
		}
	}

	return res, equiv
}

func (s *SMT) getInteriorNodesNext(batch [][]byte, idx, subtreeHeight, height int, root []byte) ([][]byte, []byte, bool) {

	if !bytes.Equal(batch[0], []byte{0}) {
		return [][]byte{}, root, true
	}

	lidx := idx*2 + 1
	ridx := idx*2 + 2

	llidx := lidx*2 + 1
	lridx := lidx*2 + 2

	rlidx := ridx*2 + 1
	rridx := ridx*2 + 2

	var res []byte

	var unfinishedLeaves [][]byte

	if subtreeHeight > 0 {
		var lres, rres []byte
		var eq bool
		// if left child has children, call the function to go lower
		if llidx < 32 && lridx < 32 {
			if len(batch[llidx]) != 0 || len(batch[lridx]) != 0 {
				var temp [][]byte
				temp, lres, eq = s.getInteriorNodesNext(batch, lidx, subtreeHeight-1, height, root)
				if !eq {
					return nil, nil, false
				}
				unfinishedLeaves = append(unfinishedLeaves, temp...)
			}
		}

		// if right child has children, call the function to go lower
		if rlidx < 32 && rridx < 32 {
			if len(batch[rlidx]) != 0 || len(batch[rridx]) != 0 {
				var temp [][]byte
				temp, rres, eq = s.getInteriorNodesNext(batch, ridx, subtreeHeight-1, height, root)
				if !eq {
					return nil, nil, false
				}
				unfinishedLeaves = append(unfinishedLeaves, temp...)
			}
		}

		// if children are leaves then get hash and return
		if len(lres) == 0 && len(rres) == 0 {

			if batch[idx][len(batch[idx])-1] == 1 && idx < 15 {
				res = s.hash(batch[lidx][:constants.HashLen], batch[ridx][:constants.HashLen], []byte{byte(height + subtreeHeight)})
			} else if len(batch[ridx]) > 0 && len(batch[lidx]) > 0 {
				// add both leaves to list
				unfinishedLeaves = append(unfinishedLeaves, batch[lidx][:constants.HashLen])
				unfinishedLeaves = append(unfinishedLeaves, batch[ridx][:constants.HashLen])
				res = s.hash(batch[lidx][:constants.HashLen], batch[ridx][:constants.HashLen])
			} else if len(batch[ridx]) > 0 {
				// add batch[ridx][:HashLength] to returned list
				unfinishedLeaves = append(unfinishedLeaves, batch[ridx][:constants.HashLen])
				res = s.hash(DefaultLeaf, batch[ridx][:constants.HashLen])
			} else if len(batch[lidx]) > 0 {
				// add batch[lidx][:HashLength] to returned list
				unfinishedLeaves = append(unfinishedLeaves, batch[lidx][:constants.HashLen])
				res = s.hash(batch[lidx][:constants.HashLen], DefaultLeaf)
			}

		} else if len(lres) > 0 && len(rres) > 0 { // hashing our way back up the tree
			res = s.hash(lres, rres)
		} else if len(lres) > 0 {
			res = s.hash(lres, DefaultLeaf)
		} else if len(rres) > 0 {
			res = s.hash(DefaultLeaf, rres)
		}

	}

	return unfinishedLeaves, res, true
}

//nolint:unused
func (s *SMT) getUnsyncedNodes(txn *badger.Txn, batch [][]byte, idx, subtreeHeight, height int, root []byte) ([][]byte, []byte) {
	lidx := idx*2 + 1
	ridx := idx*2 + 2

	llidx := lidx*2 + 1
	lridx := lidx*2 + 2

	rlidx := ridx*2 + 1
	rridx := ridx*2 + 2

	var res []byte

	var unsyncedNodes [][]byte

	if subtreeHeight > 0 {
		var lres, rres []byte
		// if left child has children, call the function to go lower
		if llidx < 32 && lridx < 32 {
			if len(batch[llidx]) != 0 || len(batch[lridx]) != 0 {
				_, lres = s.getUnsyncedNodes(txn, batch, lidx, subtreeHeight-1, height, root)
			}
		}

		// if right child has children, call the function to go lower
		if rlidx < 32 && rridx < 32 {
			if len(batch[rlidx]) != 0 || len(batch[rridx]) != 0 {
				_, rres = s.getUnsyncedNodes(txn, batch, ridx, subtreeHeight-1, height, root)
			}
		}

		// if children are leaves then get hash and return
		if len(lres) == 0 && len(rres) == 0 {
			if batch[idx][32] == 1 && idx < 16 {
				res = s.hash(batch[lidx][:constants.HashLen], batch[ridx][:constants.HashLen], []byte{byte(height + subtreeHeight)})
				// should we check a node here
			} else if len(batch[ridx]) > 0 && len(batch[lidx]) > 0 {
				// both leaves are child nodes, so possibly add both to the list of unsynced nodes
				s.loadDbMux.Lock()
				_, err := s.db.getNodeDB(txn, batch[lidx][:constants.HashLen])
				s.loadDbMux.Unlock()
				if err == badger.ErrKeyNotFound {
					unsyncedNodes = append(unsyncedNodes, batch[lidx][:constants.HashLen])
				}
				s.loadDbMux.Lock()
				_, err = s.db.getNodeDB(txn, batch[ridx][:constants.HashLen])
				s.loadDbMux.Unlock()
				if err == badger.ErrKeyNotFound {
					unsyncedNodes = append(unsyncedNodes, batch[ridx][:constants.HashLen])
				}
				res = s.hash(batch[lidx][:constants.HashLen], batch[ridx][:constants.HashLen])
			} else if len(batch[ridx]) > 0 {
				// possibly add batch[ridx][:HashLength] to the list
				s.loadDbMux.Lock()
				_, err := s.db.getNodeDB(txn, batch[ridx][:constants.HashLen])
				s.loadDbMux.Unlock()
				if err == badger.ErrKeyNotFound {
					unsyncedNodes = append(unsyncedNodes, batch[ridx][:constants.HashLen])
				}
				res = s.hash(DefaultLeaf, batch[ridx][:constants.HashLen])
			} else if len(batch[lidx]) > 0 {
				// possibly add batch[lidx][:HashLength] to the list
				s.loadDbMux.Lock()
				_, err := s.db.getNodeDB(txn, batch[lidx][:constants.HashLen])
				s.loadDbMux.Unlock()
				if err == badger.ErrKeyNotFound {
					unsyncedNodes = append(unsyncedNodes, batch[lidx][:constants.HashLen])
				}
				res = s.hash(batch[lidx][:constants.HashLen], DefaultLeaf)
			}

		} else if len(lres) > 0 && len(rres) > 0 { // hashing our way back up the tree
			res = s.hash(lres, rres)
		} else if len(lres) > 0 {
			res = s.hash(lres, DefaultLeaf)
		} else if len(rres) > 0 {
			res = s.hash(DefaultLeaf, rres)
		}

	}

	return unsyncedNodes, res
}

type LeafNode struct {
	Key   []byte
	Value []byte
}

func (s *SMT) getFinalLeafNodes(batch [][]byte, idx int) []LeafNode {
	lidx := idx*2 + 1
	ridx := idx*2 + 2

	llidx := lidx*2 + 1
	lridx := lidx*2 + 2

	rlidx := ridx*2 + 1
	rridx := ridx*2 + 2

	var finalLeafNodes []LeafNode

	if batch[idx][len(batch[idx])-1] == 1 && idx < 15 {
		// found a final leaf node within this tree, so adding it to the output slice
		ln := LeafNode{batch[lidx][:constants.HashLen], batch[ridx][:constants.HashLen]}
		finalLeafNodes = append(finalLeafNodes, ln)
		return finalLeafNodes
	}

	// if left child has children, call the function to go lower
	if llidx < 32 && lridx < 32 && len(batch) > llidx && len(batch) > lridx {
		if len(batch[llidx]) != 0 || len(batch[lridx]) != 0 {
			newFLN := s.getFinalLeafNodes(batch, lidx)
			finalLeafNodes = append(finalLeafNodes, newFLN...)
		}
	}

	// if right child has children, call the function to go lower
	if rlidx < 32 && rridx < 32 && len(batch) > rlidx && len(batch) > rridx {
		if len(batch[rlidx]) != 0 || len(batch[rridx]) != 0 {
			newFLN := s.getFinalLeafNodes(batch, ridx)
			finalLeafNodes = append(finalLeafNodes, newFLN...)
		}
	}

	return finalLeafNodes
}

// leafHash returns the hash of key_value_byte(height) concatenated, stores it in the updatedNodes and maybe in liveCache.
// leafHash is never called for a default value. Default value should not be stored.
func (s *SMT) leafHash(key, value, oldRoot []byte, batch [][]byte, iBatch, height int) []byte {
	// byte(height) is here for 2 reasons.
	// 1- to prevent potential problems with merkle proofs where if an account
	// has the same address as a node, it would be possible to prove a
	// different value for the account.
	// 2- when accounts are added to the trie, accounts on their path get pushed down the tree
	// with them. if an old account changes position from a shortcut batch to another
	// shortcut batch of different height, if would be deleted when reverting.
	h := s.hash(key, value, []byte{byte(height)})
	h = append(h, byte(1)) // byte(1) is a flag for the shortcut
	batch[2*iBatch+2] = append(value, byte(2))
	batch[2*iBatch+1] = append(key, byte(2))
	if height%4 == 0 {
		batch[0] = []byte{1} // byte(1) is a flag for the shortcut batch
		s.storeNode(batch, h, oldRoot, height)
	}
	return h
}

// storeNode stores a batch and deletes the old node from cache
func (s *SMT) storeNode(batch [][]byte, h, oldRoot []byte, height int) {
	if !bytes.Equal(h, oldRoot) {
		var node Hash
		copy(node[:], h)
		// record new node
		s.db.updatedMux.Lock()
		s.db.updatedNodes[node] = batch
		s.db.updatedMux.Unlock()
		// Cache the shortcut node if it's height is over CacheHeightLimit
		s.deleteOldNode(oldRoot, height, false)
	}
}

// interiorHash hashes 2 children to get the parent hash and stores it in the updatedNodes and maybe in liveCache.
func (s *SMT) interiorHash(left, right, oldRoot []byte, batch [][]byte, iBatch, height int) []byte {
	var h []byte
	// left and right cannot both be default. It is handled by maybeMoveUpShortcut()
	if len(left) == 0 {
		h = s.hash(DefaultLeaf, right[:constants.HashLen])
	} else if len(right) == 0 {
		h = s.hash(left[:constants.HashLen], DefaultLeaf)
	} else {
		h = s.hash(left[:constants.HashLen], right[:constants.HashLen])
	}
	h = append(h, byte(0))
	batch[2*iBatch+2] = right
	batch[2*iBatch+1] = left
	if height%4 == 0 {
		batch[0] = []byte{0}
		s.storeNode(batch, h, oldRoot, height)
	}
	return h
}

// ErrCBDone should be passed up from the callback to walkNodes when the
// iteration will stop.
var ErrCBDone = errors.New("cb done with iteration")

type result struct {
	key         []byte
	value       []byte
	unsyncedKey []byte
}

// walkNodes will call cb with the hash of all super nodes in the sub tree
// of root in a synchronous manner. Thus this function is safe to use with
// a callback that changes the database in a badger transaction.
// The callback may return ErrCBDone when it does not want the nodes to
// continue to cause callbacks. No further callbacks will be invoked after
// this error is returned and the function will return nil.
func (s *SMT) walkNodes(txn *badger.Txn, root []byte, cb func([]byte, []byte) error) error {

	// create the parent wait group for all sub routines
	wg := &sync.WaitGroup{}
	// create a cancel channel to signal child routines to stop work
	// a callback channel to collect results from
	// an error channel to collect errors from child routines
	// a wait group channel to signal the wait group has finished
	//     this last channel allows us to use the wait group in an async manner
	cancelChan := make(chan struct{}, 1)
	cbChan := make(chan result, 10000)
	errChan := make(chan error, 10000)
	wgChan := make(chan struct{})
	// the cleanup func will drain the callback channel and the
	// error channel and discard all data collected
	// this is to allow cleanup of channels for an error raised
	cleanup := func() {
		// THE CANCEL CHANNEL MUST BE CLOSED BEFORE CALLING OR THIS
		// WILL LEAK GO ROUTINES!
		// wait for the waitgroups to finish
		wg.Wait()
		// defer cleanup of the callback and error channels
		defer close(errChan)
		defer close(cbChan)
		// drain the channels
		// this is safe because the wait group is finished
		// thus as soon as both channels are empty, we may
		// return without fear of a race
		for {
			select {
			case <-cbChan:
				// drain cbChan
			case <-errChan:
				// drain errChan
			default:
				// channels drained
				return
			}
		}
	}
	// create a function that will signal the wait group as done
	wgAsync := func() {
		// register cleanup of wait group channel
		defer close(wgChan)
		// block until wait group is finished
		wg.Wait()
	}
	// register the cleanup function
	defer cleanup()

	s.loadDbMux.Lock()
	dbval, err := s.db.getNodeDB(txn, root)
	s.loadDbMux.Unlock()
	if err != nil {
		return err
	}
	var batch [][]byte
	if len(dbval) != 0 {
		var err error
		batch, err = s.parseBatch(dbval)
		if err != nil {
			// if an error occurs here, the wait group is empty and all
			// cleanup will pass through on default cases
			// return the error to caller
			return err
		}
	}
	// send back the root for cb
	err = cb(root, dbval)
	if err != nil {
		return err
	}
	// chek if this is a shortcut
	isShortcut := false
	if batch[0][0] == 1 {
		isShortcut = true
	}
	// if this node is a shortcut, it has no children
	// we may return with no action
	if isShortcut {
		return nil
		// done going down this side
	}
	// for each of the leaf nodes in this super node check the
	// node type flag
	// 00 indicates internal and thus must be walked further down
	// 01 indicates the next node is a shortcut and thus does not
	//    need to be traced -  the callback can just be invoked
	for i := 15; i < len(batch); i++ {
		if len(batch[i]) != 0 {
			flag := batch[i][len(batch[i])-1]
			if flag == 0 {
				// leads to new root of another supernode, walk down it
				// first increment the wait group
				wg.Add(1)
				go s.walk(txn, cbChan, errChan, cancelChan, wg, batch[i][:constants.HashLen], s.TrieHeight-4)
			}
			if flag == 1 {
				s.loadDbMux.Lock()
				dbval, err := s.db.getNodeDB(txn, batch[i][:constants.HashLen])
				s.loadDbMux.Unlock()
				if err != nil {
					return err
				}
				// leads to shortcut
				err = cb(batch[i][:constants.HashLen], dbval)
				if err != nil {
					return err
				}
			}
		}
	}
	// launch the routine that waits for the parent wait group
	// if any sub routines were launched this will wait for them
	// to finish and then close the wgChan
	go wgAsync()
	// collect the results from cbChan and invoke the callback until
	// an error occurs or the wait group signals done
	// once the WaitGroup signals done, drain the channels and return
	for {
		select {
		case err := <-errChan: // an error occurred in a child
			// kill all children
			close(cancelChan)
			// return the error
			// channels will be cleaned up and drained on exit
			return err
		case res := <-cbChan: // got data for callback
			// trigger the callback with data
			err := cb(res.key, res.value)
			if err != nil { // an error occurred
				// kill all children
				close(cancelChan)
				// if the error was not expected, return it
				if err != ErrCBDone {
					return err
				}
				// the callback triggered the error
				return nil
			}
		case <-wgChan: // waitgroup is done
			// there still may be data in the pipes, so drain them
			for {
				select {
				case err := <-errChan: // error see above
					close(cancelChan)
					return err
				case res := <-cbChan: // data see above
					err := cb(res.key, res.value)
					if err != nil {
						close(cancelChan)
						if err != ErrCBDone {
							return err
						}
						return nil
					}
				default: // pipes are empty
					return nil
				}
			}
		}
	}
}

// invoked by walkNodes to walk a sub node of the root
func (s *SMT) walk(txn *badger.Txn, cbChan chan<- result, errChan chan<- error, cancelChan <-chan struct{}, wg *sync.WaitGroup, root []byte, height int) {
	// register the cleanup of the wait group for this routine
	defer wg.Done()
	select {
	case <-cancelChan:
		return
	default:
		// continue
	}
	// get value from db
	s.loadDbMux.Lock()
	dbval, err := s.db.getNodeDB(txn, root)
	s.loadDbMux.Unlock()
	if err != nil {
		errChan <- err
		return
	}
	var batch [][]byte
	if len(dbval) != 0 {
		var err error
		batch, err = s.parseBatch(dbval)
		if err != nil {
			errChan <- err
			return
		}
	}
	// send root on chan
	cbChan <- result{root, dbval, nil}
	// walk the leaf nodes of the supernode
	for i := 15; i < len(batch); i++ {
		if len(batch[i]) != 0 {
			flag := batch[i][len(batch[i])-1]
			if flag == 0 { // iterior node with child
				// add a new worker to the local wait group
				wg.Add(1)
				// spawn a new walker for the child node
				go s.walk(txn, cbChan, errChan, cancelChan, wg, batch[i][:constants.HashLen], height-4)
			}
			if flag == 1 { // shortcut
				s.loadDbMux.Lock()
				dbval, err := s.db.getNodeDB(txn, batch[i][:constants.HashLen])
				s.loadDbMux.Unlock()
				// should probably check if key was not found here and return the unsynced key somehow
				if err == badger.ErrKeyNotFound {
					cbChan <- result{nil, nil, batch[i][:constants.HashLen]}
				}
				if err != nil {
					errChan <- err
					return
				}
				cbChan <- result{batch[i][:constants.HashLen], dbval, nil}
			}
		}
	}
}
