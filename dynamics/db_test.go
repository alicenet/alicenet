package dynamics

import (
	"bytes"
	"testing"

	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

type MockRawDB struct {
	rawDB map[string]string
}

func (m *MockRawDB) GetValue(txn *badger.Txn, key []byte) ([]byte, error) {
	strValue, ok := m.rawDB[string(key)]
	if !ok {
		return nil, ErrKeyNotPresent
	}
	value := []byte(strValue)
	return value, nil
}

func (m *MockRawDB) SetValue(txn *badger.Txn, key []byte, value []byte) error {
	strKey := string(key)
	strValue := string(value)
	m.rawDB[strKey] = strValue
	return nil
}

func (m *MockRawDB) DeleteValue(key []byte) error {
	strKey := string(key)
	_, ok := m.rawDB[strKey]
	if !ok {
		return ErrKeyNotPresent
	}
	delete(m.rawDB, strKey)
	return nil
}

func (m *MockRawDB) View(fn func(txn *badger.Txn) error) error {
	return fn(nil)
}

func (m *MockRawDB) Update(fn func(txn *badger.Txn) error) error {
	return fn(nil)
}

func TestMock(t *testing.T) {
	key := []byte("Key")
	value := []byte("Key")

	m := &MockRawDB{}
	m.rawDB = make(map[string]string)

	_, err := m.GetValue(nil, key)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	err = m.SetValue(nil, key, value)
	if err != nil {
		t.Fatal(err)
	}

	retValue, err := m.GetValue(nil, key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retValue, value) {
		t.Fatal("values do not match")
	}

	err = m.DeleteValue(key)
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.GetValue(nil, key)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func newLogger() *logrus.Logger {
	logger := logrus.New()
	return logger
}

func initializeDB() *Database {
	logger := newLogger()
	db := &Database{}
	db.logger = logger
	mock := &MockRawDB{}
	mock.rawDB = make(map[string]string)
	db.rawDB = mock
	return db
}

func TestGetSetNode(t *testing.T) {
	db := initializeDB()

	node := &Node{}
	err := db.SetNode(nil, node)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	node.prevEpoch = 1
	node.thisEpoch = 1
	node.nextEpoch = 1
	node.dynamicValues = &DynamicValues{}
	err = db.SetNode(nil, node)
	if err != nil {
		t.Fatal(err)
	}
	nodeBytes, err := node.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	// Should raise error
	_, err = db.GetNode(nil, 0)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	node2, err := db.GetNode(nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	node2Bytes, err := node2.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(nodeBytes, node2Bytes) {
		t.Fatal("nodes do not match")
	}
}

func TestGetSetLinkedList(t *testing.T) {
	db := initializeDB()

	ll := &LinkedList{}
	err := db.SetLinkedList(nil, ll)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	ll.currentValue = 1
	err = db.SetLinkedList(nil, ll)
	if err != nil {
		t.Fatal(err)
	}
	llBytes := ll.Marshal()

	ll2, err := db.GetLinkedList(nil)
	if err != nil {
		t.Fatal(err)
	}
	ll2Bytes := ll2.Marshal()
	if !bytes.Equal(llBytes, ll2Bytes) {
		t.Fatal("LinkedLists do not match")
	}
}
