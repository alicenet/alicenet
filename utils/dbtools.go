package utils

import (
	"github.com/dgraph-io/badger/v2"
)

// GetValue returns a value from the database that is safe for use.
func GetValue(txn *badger.Txn, key []byte) ([]byte, error) {
	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}
	return item.ValueCopy(nil)
}

// SetValue sets the value for key in the database that is safe for use.
func SetValue(txn *badger.Txn, key []byte, value []byte) error {
	return txn.Set(key, value)
}

// DeleteValue removes the value for key in the database that is safe for use.
func DeleteValue(txn *badger.Txn, key []byte) error {
	return txn.Delete(key)
}

// GetInt64 will retrieve an int64 value in the database
func GetInt64(txn *badger.Txn, key []byte) (int64, error) {
	v, err := GetValue(txn, key)
	if err != nil {
		return 0, err
	}
	return UnmarshalInt64(v)
}

// SetInt64 will set an int64 value in the database
func SetInt64(txn *badger.Txn, key []byte, v int64) error {
	vv := MarshalInt64(v)
	return SetValue(txn, key, vv)
}
