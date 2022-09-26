package trie

import (
	"bytes"
	"fmt"

	"github.com/alicenet/alicenet/errorz"
	"github.com/dgraph-io/badger/v2"
)

func GetNodeDB(txn *badger.Txn, prefix []byte, key []byte) ([]byte, error) {
	key = convNilToBytes(key)
	var node Hash
	copy(node[:], key)
	v, err := getNodeDB(txn, prefix, key)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (s *SMT) StoreSnapShotNode(txn *badger.Txn, batch []byte, root []byte, layer int) ([][]byte, int, []LeafNode, error) {
	pbatch, err := s.parseBatch(batch)
	if err != nil {
		return nil, 0, nil, err
	}
	if layer == 0 {
		subBatch, lvs, err := s.storeFastSyncRoot(txn, pbatch, root)
		if err != nil {
			return nil, 0, nil, err
		}
		return subBatch, layer + 1, lvs, nil
	}
	return s.storeFastSyncNonRoot(txn, pbatch, root, layer)
}

func (s *SMT) verifyBatchEasy(batch [][]byte, root []byte, layer int) ([]byte, bool) {
	if !bytes.Equal(batch[0], []byte{0}) {
		// batch is a shortcut node
		return s.verifyBatch(batch, 0, 1, 256-4*layer, root, false)
	}
	// batch is a regular node
	return s.verifyBatch(batch, 0, 4, 256-4*(layer+1), root, false)
}

func (s *SMT) getInteriorNodesEasy(batch [][]byte, root []byte, layer int) ([][]byte, bool) {
	if !bytes.Equal(batch[0], []byte{0}) {
		// batch is a shortcut node
		subBatch, _, ok := s.getInteriorNodesNext(batch, 0, 1, 256-4*layer, root)
		return subBatch, ok
	}
	// batch is a regular node
	subBatch, _, ok := s.getInteriorNodesNext(batch, 0, 4, 256-4*(layer+1), root)
	return subBatch, ok
}

func (s *SMT) storeFastSyncRoot(txn *badger.Txn, batch [][]byte, root []byte) ([][]byte, []LeafNode, error) {
	_, ok := s.verifyBatchEasy(batch, root, 0)
	if !ok {
		return nil, nil, errorz.ErrInvalid{}.New("Error in smt.StoreFastSyncRoot at s.verifyBatchEasy(batch, root, 0)")
	}
	err := s.db.setNodeDB(txn, root, s.db.serializeBatch(batch))
	if err != nil {
		return nil, nil, err
	}
	subBatch, ok := s.getInteriorNodesEasy(batch, root, 0)
	if !ok {
		return nil, nil, errorz.ErrInvalid{}.New("Error in smt.StoreFastSyncRoot at s.getInteriorNodesEasy(batch, root, 0)")
	}
	lvs := s.getFinalLeafNodes(batch, 0)
	return subBatch, lvs, nil
}

func (s *SMT) storeFastSyncNonRoot(txn *badger.Txn, batch [][]byte, root []byte, layer int) ([][]byte, int, []LeafNode, error) {
	if layer <= 0 {
		return nil, 0, nil, errorz.ErrInvalid{}.New("Error in smt.StoreFastSyncNonRoot at s.storeFastSyncNonRoot: invalid layer")
	}
	_, ok := s.verifyBatchEasy(batch, root, layer)
	if !ok {
		return nil, 0, nil, errorz.ErrInvalid{}.New(fmt.Sprintf("Error in smt.StoreFastSyncNonRoot at s.verifyBatchEasy(batch, root, layer): layer:%v", layer))
	}
	err := s.db.setNodeDB(txn, root, s.db.serializeBatch(batch))
	if err != nil {
		return nil, 0, nil, err
	}
	subBatch, ok := s.getInteriorNodesEasy(batch, root, layer)
	if !ok {
		return nil, 0, nil, errorz.ErrInvalid{}.New("Error in smt.StoreFastSyncNonRoot at s.getInteriorNodesEasy(batch, root, layer)")
	}
	lvs := s.getFinalLeafNodes(batch, 0)
	return subBatch, layer + 1, lvs, nil
}

func (s *SMT) FinalizeSnapShotRoot(txn *badger.Txn, root []byte, height uint32) error {
	err := s.db.setCommitHeightDB(txn, height)
	if err != nil {
		return err
	}
	err = s.db.setRootForHeightDB(txn, height, root)
	if err != nil {
		return err
	}

	return nil
}
