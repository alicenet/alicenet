package dynamics

import (
	"sync"

	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

type rawDataBase interface {
	GetValue(txn *badger.Txn, key []byte) ([]byte, error)
	SetValue(txn *badger.Txn, key []byte, value []byte) error
	Update(func(txn *badger.Txn) error) error
	View(func(txn *badger.Txn) error) error
}

// Database is an abstraction for object storage
type Database struct {
	sync.Mutex
	rawDB  rawDataBase
	logger *logrus.Logger
}

// SetNode stores Node in the database
func (db *Database) SetNode(txn *badger.Txn, node *Node) error {
	err := node.Validate()
	if err != nil {
		return err
	}
	nodeKey, err := makeNodeKey(node.thisEpoch)
	if err != nil {
		return err
	}
	key, err := nodeKey.Marshal()
	if err != nil {
		return err
	}
	nodeBytes, err := node.Marshal()
	if err != nil {
		return err
	}
	err = db.rawDB.SetValue(txn, key, nodeBytes)
	if err != nil {
		return err
	}
	return nil
}

// GetNode retrieves Node from the database
func (db *Database) GetNode(txn *badger.Txn, epoch uint32) (*Node, error) {
	nodeKey, err := makeNodeKey(epoch)
	if err != nil {
		return nil, err
	}
	key, err := nodeKey.Marshal()
	if err != nil {
		return nil, err
	}
	v, err := db.rawDB.GetValue(txn, key)
	if err != nil {
		return nil, err
	}
	node := &Node{}
	err = node.Unmarshal(v)
	if err != nil {
		return nil, err
	}
	err = node.Validate()
	if err != nil {
		return nil, err
	}
	return node, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// SetLinkedList saves LinkedList to the database
func (db *Database) SetLinkedList(txn *badger.Txn, ll *LinkedList) error {
	if !ll.IsValid() {
		return ErrInvalid
	}
	value := ll.Marshal()
	llKey := dbprefix.PrefixStorageLinkedListKey()
	err := db.rawDB.SetValue(txn, llKey, value)
	if err != nil {
		return err
	}
	return nil
}

// GetLinkedList retrieves LinkedList from the database
func (db *Database) GetLinkedList(txn *badger.Txn) (*LinkedList, error) {
	llKey := dbprefix.PrefixStorageLinkedListKey()
	v, err := db.rawDB.GetValue(txn, llKey)
	if err != nil {
		return nil, err
	}
	ll := &LinkedList{}
	err = ll.Unmarshal(v)
	if err != nil {
		return nil, err
	}
	if !ll.IsValid() {
		return nil, ErrInvalid
	}
	return ll, nil
}
