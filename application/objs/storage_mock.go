package objs

import (
	"math/big"
	"time"

	"github.com/alicenet/alicenet/application/wrapper"
	"github.com/alicenet/alicenet/dynamics"
	"github.com/dgraph-io/badger/v2"
)

func MakeMockStorageGetter() *MockStorageGetter {
	maxBytes := uint32(0)
	dataStoreEpochFee := new(big.Int)
	atomicSwapFee := new(big.Int)
	valueStoreFee := new(big.Int)
	minTxFee := new(big.Int)

	msg := &MockStorageGetter{
		maxBytes:          maxBytes,
		dataStoreEpochFee: dataStoreEpochFee,
		valueStoreFee:     valueStoreFee,
		atomicSwapFee:     atomicSwapFee,
		minTxFee:          minTxFee,
	}
	return msg
}

func MakeStorage(msg dynamics.StorageGetter) *wrapper.Storage {
	storage := wrapper.NewStorage(msg)
	return storage
}

type MockStorageGetter struct {
	maxBytes          uint32
	dataStoreEpochFee *big.Int
	valueStoreFee     *big.Int
	atomicSwapFee     *big.Int
	minTxFee          *big.Int
	// maxTxVectorLength int
}

func (msg *MockStorageGetter) GetMaxBytes() uint32 {
	return msg.maxBytes
}

func (msg *MockStorageGetter) SetMaxBytes(value uint32) {
	msg.maxBytes = value
}

func (msg *MockStorageGetter) GetMaxProposalSize() uint32 {
	return msg.maxBytes
}

func (msg *MockStorageGetter) GetProposalStepTimeout() time.Duration {
	return time.Duration(0)
}
func (msg *MockStorageGetter) GetPreVoteStepTimeout() time.Duration {
	return time.Duration(0)
}
func (msg *MockStorageGetter) GetPreCommitStepTimeout() time.Duration {
	return time.Duration(0)
}
func (msg *MockStorageGetter) GetDeadBlockRoundNextRoundTimeout() time.Duration {
	return time.Duration(0)
}
func (msg *MockStorageGetter) GetDownloadTimeout() time.Duration {
	return time.Duration(0)
}
func (msg *MockStorageGetter) GetSrvrMsgTimeout() time.Duration {
	return time.Duration(0)
}
func (msg *MockStorageGetter) GetMsgTimeout() time.Duration {
	return time.Duration(0)
}
func (msg *MockStorageGetter) GetMaxTxVectorLength() int {
	return 128
}

func (msg *MockStorageGetter) UpdateStorage(txn *badger.Txn, update dynamics.Updater) error {
	return nil
}
func (msg *MockStorageGetter) LoadStorage(txn *badger.Txn, epoch uint32) error {
	return nil
}

func (msg *MockStorageGetter) GetDataStoreEpochFee() *big.Int {
	return msg.dataStoreEpochFee
}

func (msg *MockStorageGetter) SetDataStoreEpochFee(value *big.Int) {
	if value == nil {
		panic("invalid value")
	}
	msg.dataStoreEpochFee.Set(value)
}

func (msg *MockStorageGetter) GetDataStoreValidVersion() uint32 {
	return 0
}

func (msg *MockStorageGetter) GetValueStoreFee() *big.Int {
	return msg.valueStoreFee
}

func (msg *MockStorageGetter) SetValueStoreFee(value *big.Int) {
	if value == nil {
		panic("invalid value")
	}
	msg.valueStoreFee.Set(value)
}

func (msg *MockStorageGetter) GetValueStoreValidVersion() uint32 {
	return 0
}

func (msg *MockStorageGetter) GetAtomicSwapFee() *big.Int {
	return msg.atomicSwapFee
}

func (msg *MockStorageGetter) SetAtomicSwapFee(value *big.Int) {
	if value == nil {
		panic("invalid value")
	}
	msg.atomicSwapFee.Set(value)
}

func (msg *MockStorageGetter) GetAtomicSwapValidStopEpoch() uint32 {
	return 0
}

func (msg *MockStorageGetter) GetMinTxFee() *big.Int {
	return msg.minTxFee
}

func (msg *MockStorageGetter) SetMinTxFee(value *big.Int) {
	if value == nil {
		panic("invalid value")
	}
	msg.minTxFee.Set(value)
}

func (msg *MockStorageGetter) GetTxValidVersion() uint32 {
	return 0
}
