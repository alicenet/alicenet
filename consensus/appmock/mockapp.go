package appmock

import (
	trie "github.com/MadBase/MadNet/badgerTrie"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/logging"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

var _ Application = (*MockApplication)(nil)

//MockApplication is the the receiver for TxHandler interface
type MockApplication struct {
	logger     *logrus.Logger
	validValue *objs.Proposal
	MissingTxn bool
}

const (
	notImpl = "not impllemented"
)

// SetNextValidValue is defined on the interface object
func (m *MockApplication) SetNextValidValue(vv *objs.Proposal) {
	m.validValue = vv
}

// New returns a mocked Application
func New() *MockApplication {
	return &MockApplication{logging.GetLogger("MockApp"), nil, false}
}

// ApplyState is defined on the interface object
func (m *MockApplication) ApplyState(*badger.Txn, uint32, uint32, []interfaces.Transaction) ([]byte, error) {
	return nil, nil
}

//GetValidProposal is defined on the interface object
func (m *MockApplication) GetValidProposal(txn *badger.Txn, chainID, height, maxBytes uint32) ([]interfaces.Transaction, []byte, error) {
	return nil, m.validValue.PClaims.BClaims.StateRoot, nil
}

// PendingTxAdd is defined on the interface object
func (m *MockApplication) PendingTxAdd(*badger.Txn, uint32, uint32, []interfaces.Transaction) error {
	return nil
}

//IsValid is defined on the interface object
func (m *MockApplication) IsValid(txn *badger.Txn, chainID uint32, height uint32, stateHash []byte, _ []interfaces.Transaction) (bool, error) {
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
	panic(notImpl)
}

// GetSnapShotNode is defined on the interface object
func (m *MockApplication) GetSnapShotNode(txn *badger.Txn, height uint32, key []byte) ([]byte, error) {
	panic(notImpl)
}

// StoreSnapShotStateData is defined on the interface object
func (m *MockApplication) StoreSnapShotStateData(txn *badger.Txn, key []byte, value []byte, data []byte) error {
	panic(notImpl)
}

// GetSnapShotStateData is defined on the interface object
func (m *MockApplication) GetSnapShotStateData(txn *badger.Txn, key []byte) ([]byte, error) {
	panic(notImpl)
}

// FinalizeSnapShotRoot is defined on the interface object
func (m *MockApplication) FinalizeSnapShotRoot(txn *badger.Txn, root []byte, height uint32) error {
	panic(notImpl)
}

// BeginSnapShotSync is defined on the interface object
func (m *MockApplication) BeginSnapShotSync(txn *badger.Txn) error {
	panic(notImpl)
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
