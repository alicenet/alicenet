package dynamics

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
)

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
	node.rawStorage = &RawStorage{}
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

	ll.epochLastUpdated = 1
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
