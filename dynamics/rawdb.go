package dynamics

import (
	"github.com/dgraph-io/badger/v2"
)

type rawDataBase interface {
	GetValue(txn *badger.Txn, key []byte) ([]byte, error)
	SetValue(txn *badger.Txn, key, value []byte) error
	Update(func(txn *badger.Txn) error) error
	View(func(txn *badger.Txn) error) error
}
