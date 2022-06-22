package appmock

import (
	"bytes"
	"errors"
	appObjs "github.com/alicenet/alicenet/application/objs"
	trie "github.com/alicenet/alicenet/badgerTrie"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/logging"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

var _ interfaces.Application = (*MockApplication)(nil)

//MockApplication is the the receiver for TxHandler interface
type MockApplication struct {
	logger     *logrus.Logger
	validValue *objs.Proposal
	MissingTxn bool
	txs        []interfaces.Transaction
	leafs      []trie.LeafNode
	shouldFail bool
}

const (
	notImpl = "not impllemented"
)

// SetNextValidValue is defined on the interface object
func (m *MockApplication) SetNextValidValue(vv *objs.Proposal) {
	m.validValue = vv
}

// SetLeafs is defined on the interface object
func (m *MockApplication) SetLeafs(leafs []trie.LeafNode) {
	m.leafs = leafs
}

// SetTxs is defined on the interface object
func (m *MockApplication) SetTxs(txs []*appObjs.Tx) {
	m.txs = make([]interfaces.Transaction, len(txs))
	for i := 0; i < len(txs); i++ {
		m.txs[i] = txs[i]
	}
}

// SetNextValidValue is defined on the interface object
func (m *MockApplication) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

// New returns a mocked Application
func New() *MockApplication {
	return &MockApplication{logging.GetLogger("test"), nil, false, nil, nil, false}
}

// ApplyState is defined on the interface object
func (m *MockApplication) ApplyState(*badger.Txn, uint32, uint32, []interfaces.Transaction) ([]byte, error) {
	return nil, nil
}

//GetValidProposal is defined on the interface object
func (m *MockApplication) GetValidProposal(txn *badger.Txn, chainID, height, maxBytes uint32) ([]interfaces.Transaction, []byte, error) {
	if chainID == 7777 {
		return nil, nil, errors.New("error")
	}
	if chainID == 8888 {
		tx := &appObjs.Tx{
			Fee: nil,
		}
		txs := []interfaces.Transaction{tx}

		return txs, nil, nil
	}
	return m.txs, m.validValue.PClaims.BClaims.StateRoot, nil
}

// PendingTxAdd is defined on the interface object
func (m *MockApplication) PendingTxAdd(txn *badger.Txn, chainID uint32, height uint32, txs []interfaces.Transaction) error {
	if chainID == 7777 {
		return errorz.ErrInvalid{}.New("already  mined")
	}
	return nil
}

//IsValid is defined on the interface object
func (m *MockApplication) IsValid(txn *badger.Txn, chainID uint32, height uint32, stateHash []byte, _ []interfaces.Transaction) (bool, error) {
	if chainID == 7777 {
		return false, errorz.ErrInvalid{}.New("")
	}

	if chainID == 8888 {
		return false, errors.New("error")
	}

	if chainID == 9999 {
		return false, nil
	}

	if m.MissingTxn {
		return false, errorz.ErrMissingTransactions
	}
	return true, nil
}

// MinedTxGet is defined on the interface object
func (m *MockApplication) MinedTxGet(*badger.Txn, [][]byte) ([]interfaces.Transaction, [][]byte, error) {
	return nil, nil, nil
}

// PendingTxGet is defined on the interface object
func (m *MockApplication) PendingTxGet(txn *badger.Txn, height uint32, txhashes [][]byte) ([]interfaces.Transaction, [][]byte, error) {
	return nil, nil, nil
}

//PendingTxContains is defined on the interface object
func (m *MockApplication) PendingTxContains(txn *badger.Txn, height uint32, txHashes [][]byte) ([][]byte, error) {
	return nil, nil
}

// UnmarshalTx is defined on the interface object
func (m *MockApplication) UnmarshalTx(v []byte) (interfaces.Transaction, error) {
	return &MockTransaction{v}, nil
}

// StoreSnapShotNode is defined on the interface object
func (m *MockApplication) StoreSnapShotNode(txn *badger.Txn, batch []byte, root []byte, layer int) ([][]byte, int, []trie.LeafNode, error) {
	if m.shouldFail {
		return nil, 0, nil, errors.New("error")
	}
	return [][]byte{{123}}, 0, m.leafs, nil
}

// GetSnapShotNode is defined on the interface object
func (m *MockApplication) GetSnapShotNode(txn *badger.Txn, height uint32, key []byte) ([]byte, error) {
	panic(notImpl)
}

// StoreSnapShotStateData is defined on the interface object
func (m *MockApplication) StoreSnapShotStateData(txn *badger.Txn, key []byte, value []byte, data []byte) error {
	if m.shouldFail {
		return errors.New("error")
	}
	return nil
}

// GetSnapShotStateData is defined on the interface object
func (m *MockApplication) GetSnapShotStateData(txn *badger.Txn, key []byte) ([]byte, error) {
	if bytes.Equal(make([]byte, constants.HashLen), key) {
		return nil, errors.New("error")
	}

	if bytes.Equal([]byte{123}, key) {
		return nil, badger.ErrKeyNotFound
	}

	return make([]byte, constants.HashLen), nil
}

// FinalizeSnapShotRoot is defined on the interface object
func (m *MockApplication) FinalizeSnapShotRoot(txn *badger.Txn, root []byte, height uint32) error {
	panic(notImpl)
}

// BeginSnapShotSync is defined on the interface object
func (m *MockApplication) BeginSnapShotSync(txn *badger.Txn) error {
	if m.shouldFail {
		return errors.New("error")
	}
	return nil
}

// FinalizeSync is defined on the interface object
func (m *MockApplication) FinalizeSync(txn *badger.Txn) error {
	panic(notImpl)
}

// MockTransaction is defined on the interface object
type MockTransaction struct {
	V []byte
}

// TxHash is defined on the interface object
func (m *MockTransaction) TxHash() ([]byte, error) {
	return crypto.Hasher(m.V), nil
}

//MarshalBinary is defined on the interface object
func (m *MockTransaction) MarshalBinary() ([]byte, error) {
	return m.V, nil
}

//XXXIsTx is defined on the interface object
func (m *MockTransaction) XXXIsTx() {}
