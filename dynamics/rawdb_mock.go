package dynamics

import "github.com/dgraph-io/badger/v2"

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
