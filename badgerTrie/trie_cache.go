/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"sync"

	"github.com/dgraph-io/badger/v2"
)

type cacheDB struct {
	// updatedNodes that will be flushed to disk
	updatedNodes map[Hash][][]byte
	// updatedMux is a lock for updatedNodes
	updatedMux sync.RWMutex
	// nodesToRevert will be deleted from db
	nodesToRevert [][]byte
	// prefixFunc appends a prefix to all database key
	// writes and reads to prevent collisions with other
	// state stored in the db
	prefixFunc func() []byte
}

// commit stores the updated nodes to disk.
func (db *cacheDB) commit(txn *badger.Txn, s *SMT, height uint32) error {
	for key, batch := range db.updatedNodes {
		var node []byte
		err := db.setNodeDB(txn, append(node, key[:]...), db.serializeBatch(batch))
		if err != nil {
			return err
		}
	}
	err := db.setCommitHeightDB(txn, height)
	if err != nil {
		return err
	}
	err = db.setRootForHeightDB(txn, height, s.Root)
	if err != nil {
		return err
	}
	return nil
}

func (db *cacheDB) drop(externalDB *badger.DB) error {
	return externalDB.DropPrefix(db.prefixFunc())
}

func (db *cacheDB) serializeBatch(batch [][]byte) []byte {
	serialized := make([]byte, 4)
	if batch[0][0] == 1 {
		// the batch node is a shortcut
		bitSet(serialized, 31)
	}
	for i := 1; i < 31; i++ {
		if len(batch[i]) != 0 {
			bitSet(serialized, i-1)
			serialized = append(serialized, batch[i]...)
		}
	}
	return serialized
}

func (db *cacheDB) snapShot(txn *badger.Txn, s *SMT, snapShotPrefix func() []byte) error {
	nodeMap := make(map[Hash][]byte)
	keep := func(k []byte, v []byte) error {
		var node Hash
		copy(node[:], k)
		nodeMap[node] = v
		return nil
	}

	height, err := db.getCommitHeightDB(txn)
	if err != nil {
		return err
	}
	err = s.walkNodes(txn, s.Root, keep)
	if err != nil {
		return err
	}
	for k, v := range nodeMap {
		err := setNodeDB(txn, snapShotPrefix(), k[:], v)
		if err != nil {
			return err
		}
	}
	err = setRootForHeightDB(txn, snapShotPrefix(), height, s.Root)
	if err != nil {
		return err
	}
	err = setCommitHeightDB(txn, snapShotPrefix(), height)
	if err != nil {
		return err
	}
	return nil
}
