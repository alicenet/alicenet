/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

// The Package Trie implements a sparse merkle trie.

import (
	"bytes"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
)

// Get fetches the value of a key by going down the current trie root.
func (s *SMT) Get(txn *badger.Txn, key []byte) ([]byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.get(txn, s.Root, key, nil, 0, s.TrieHeight)
}

// get fetches the value of a key in a given trie
// defined by root
func (s *SMT) get(txn *badger.Txn, root []byte, key []byte, batch [][]byte, iBatch, height int) ([]byte, error) {
	if len(root) == 0 {
		return nil, nil
	}

	// Fetch the children of the node
	batch, iBatch, lnode, rnode, isShortcut, err := s.loadChildren(txn, root, height, iBatch, batch)
	if err != nil {
		return nil, err
	}
	if isShortcut {
		if bytes.Equal(lnode[:constants.HashLen], key) {
			return rnode[:constants.HashLen], nil
		}
		return nil, nil
	}
	if bitIsSet(key, s.TrieHeight-height) {
		return s.get(txn, rnode, key, batch, 2*iBatch+2, height-1)
	}
	return s.get(txn, lnode, key, batch, 2*iBatch+1, height-1)
}

// Commit stores the updated nodes to disk
// Commit should be called for every block
func (s *SMT) Commit(txn *badger.Txn, height uint32) ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	err := s.db.commit(txn, s, height)
	if err != nil {
		return nil, err
	}
	s.db.updatedNodes = make(map[Hash][][]byte)
	s.db.nodesToRevert = [][]byte{}
	s.prevRoot = s.Root
	return utils.CopySlice(s.Root), nil
}

// Discard rolls back the changes made by previous updates made without commit
func (s *SMT) Discard() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(s.db.updatedNodes) > 0 {
		s.Root = s.prevRoot
		s.db.updatedNodes = make(map[Hash][][]byte)
		s.db.nodesToRevert = [][]byte{}
	}
}

// SnapShot allows the state of the database to be copied into a different
// prefix. This function will return that copy as an SMT object to the caller.
func (s *SMT) SnapShot(txn *badger.Txn, snapShotPrefix func() []byte) (*SMT, error) {
	err := s.db.snapShot(txn, s, snapShotPrefix)
	if err != nil {
		return nil, err
	}
	newsmt := NewSMT(nil, s.hash, snapShotPrefix)
	rootCopy := make([]byte, len(s.Root))
	copy(rootCopy[:], s.Root)
	prevRootCopy := make([]byte, len(s.prevRoot))
	copy(prevRootCopy[:], s.prevRoot)
	newsmt.Root = rootCopy
	newsmt.prevRoot = prevRootCopy
	return newsmt, nil
}

// Drop will drop all state associated with this trie from the database.
func (s *SMT) Drop(bDB *badger.DB) error {
	return s.db.drop(bDB)
}

// Height returns the number of times commit has been called on this trie
func (s *SMT) Height(txn *badger.Txn) (uint32, error) {
	return s.db.getCommitHeightDB(txn)
}
