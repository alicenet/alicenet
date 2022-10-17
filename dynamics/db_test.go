package dynamics

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

func NewTestRawDB() *badger.DB {
	logging.GetLogger(constants.LoggerBadger).SetOutput(ioutil.Discard)
	db, err := utils.OpenBadger(context.Background().Done(), "", true)
	if err != nil {
		panic(err)
	}
	return db
}

func NewTestDB() *db.Database {
	db := &db.Database{}
	db.Init(NewTestRawDB())
	return db
}

func TestMock(t *testing.T) {
	t.Parallel()
	key := []byte("Key")
	value := []byte("Key")

	m := NewTestDB()

	m.DB().Update(func(txn *badger.Txn) error {
		_, err := m.GetValue(txn, key)
		if !errors.Is(err, ErrKeyNotPresent) {
			t.Fatalf("should have failed: %v", err)
		}

		err = m.SetValue(txn, key, value)
		if err != nil {
			t.Fatal(err)
		}

		retValue, err := m.GetValue(txn, key)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(retValue, value) {
			t.Fatal("values do not match")
		}
		return nil
	})

}

func newLogger() *logrus.Logger {
	logger := logrus.New()
	return logger
}

func initializeDB() *Database {
	logger := newLogger()
	db := &Database{}
	db.logger = logger
	db.rawDB = NewTestDB()
	return db
}

func TestGetSetNode(t *testing.T) {
	t.Parallel()
	db := initializeDB()

	node := &Node{}
	db.rawDB.Update(func(txn *badger.Txn) error {
		err := db.SetNode(txn, node)
		if err == nil {
			t.Fatal("Should have raised error (1)")
		}

		node.prevEpoch = 0
		node.thisEpoch = 1
		node.nextEpoch = 0
		_, node.dynamicValues = GetStandardDynamicValue()
		err = db.SetNode(txn, node)
		if err != nil {
			t.Fatal(err)
		}
		nodeBytes, err := node.Marshal()
		if err != nil {
			t.Fatal(err)
		}

		// Should raise error
		_, err = db.GetNode(txn, 0)
		if err == nil {
			t.Fatal("Should have raised error (1)")
		}

		node2, err := db.GetNode(txn, 1)
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
		return nil
	})
}

func TestGetSetLinkedList(t *testing.T) {
	t.Parallel()
	db := initializeDB()

	ll := &LinkedList{}
	db.rawDB.Update(func(txn *badger.Txn) error {
		err := db.SetLinkedList(txn, ll)
		if err == nil {
			t.Fatal("Should have raised error (1)")
		}

		ll.currentValue = 1
		ll.tail = 2
		err = db.SetLinkedList(txn, ll)
		if err != nil {
			t.Fatal(err)
		}
		llBytes := ll.Marshal()

		ll2, err := db.GetLinkedList(txn)
		if err != nil {
			t.Fatal(err)
		}
		ll2Bytes := ll2.Marshal()
		if !bytes.Equal(llBytes, ll2Bytes) {
			t.Fatal("LinkedLists do not match")
		}
		return nil
	})
}
