package utils

import (
	"bytes"
	"os"
	"testing"

	"github.com/dgraph-io/badger/v2"
)

func TestDBValue(t *testing.T) {
	dir, err := os.MkdirTemp("", "dbtools-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	////////////////////////////////////////
	valueSize := 32
	value := make([]byte, valueSize)
	value[0] = 1
	value[valueSize-1] = 1
	keySize := 32
	key := make([]byte, keySize)
	key[0] = 255
	key[keySize-1] = 255
	err = db.Update(func(txn *badger.Txn) error {
		_, err := GetValue(txn, key)
		if err == nil {
			t.Fatal("Should have raised error (1)")
		}
		err = SetValue(txn, key, value)
		if err != nil {
			t.Fatal(err)
		}
		valueTest, err := GetValue(txn, key)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(value, valueTest) {
			t.Fatal("value returned does not match true value")
		}
		err = DeleteValue(txn, key)
		if err != nil {
			t.Fatal(err)
		}
		_, err = GetValue(txn, key)
		if err == nil {
			t.Fatal("Should have raised error (2)")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDBInt64(t *testing.T) {
	dir, err := os.MkdirTemp("", "dbtools-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	////////////////////////////////////////
	var value int64 = 9223372036854775807
	keySize := 32
	key := make([]byte, keySize)
	key[0] = 255
	key[keySize-1] = 255
	err = db.Update(func(txn *badger.Txn) error {
		_, err := GetInt64(txn, key)
		if err == nil {
			t.Fatal("Should have raised error (1)")
		}
		err = SetInt64(txn, key, value)
		if err != nil {
			t.Fatal(err)
		}
		valueTest, err := GetInt64(txn, key)
		if err != nil {
			t.Fatal(err)
		}
		if value != valueTest {
			t.Fatal("value returned does not match true value")
		}
		err = DeleteValue(txn, key)
		if err != nil {
			t.Fatal(err)
		}
		_, err = GetInt64(txn, key)
		if err == nil {
			t.Fatal("Should have raised error (2)")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
