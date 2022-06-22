package minedtx

import (
	"math/big"
	"time"

	"github.com/alicenet/alicenet/application/wrapper"
	"github.com/alicenet/alicenet/dynamics"
	"github.com/dgraph-io/badger/v2"
)

func makeMockStorageGetter() *mockStorageGetter {
	maxBytes := uint32(0)
	dataStoreEpochFee := new(big.Int).SetInt64(0)
	atomicSwapFee := new(big.Int).SetInt64(0)
	valueStoreFee := new(big.Int).SetInt64(0)
	minTxFee := new(big.Int).SetInt64(0)

	msg := &mockStorageGetter{
		maxBytes:          maxBytes,
		dataStoreEpochFee: dataStoreEpochFee,
		valueStoreFee:     valueStoreFee,
		atomicSwapFee:     atomicSwapFee,
		minTxFee:          minTxFee,
	}
	return msg
}

func makeStorage(msg dynamics.StorageGetter) *wrapper.Storage {
	storage := wrapper.NewStorage(msg)
	return storage
}

type mockStorageGetter struct {
	maxBytes          uint32
	dataStoreEpochFee *big.Int
	valueStoreFee     *big.Int
	atomicSwapFee     *big.Int
	minTxFee          *big.Int
}

func (msg *mockStorageGetter) GetMaxBytes() uint32 {
	return msg.maxBytes
}

func (msg *mockStorageGetter) SetMaxBytes(value uint32) {
	msg.maxBytes = value
}

func (msg *mockStorageGetter) GetMaxProposalSize() uint32 {
	return msg.maxBytes
}

func (msg *mockStorageGetter) GetProposalStepTimeout() time.Duration {
	return time.Duration(0)
}
func (msg *mockStorageGetter) GetPreVoteStepTimeout() time.Duration {
	return time.Duration(0)
}
func (msg *mockStorageGetter) GetPreCommitStepTimeout() time.Duration {
	return time.Duration(0)
}
func (msg *mockStorageGetter) GetDeadBlockRoundNextRoundTimeout() time.Duration {
	return time.Duration(0)
}
func (msg *mockStorageGetter) GetDownloadTimeout() time.Duration {
	return time.Duration(0)
}
func (msg *mockStorageGetter) GetSrvrMsgTimeout() time.Duration {
	return time.Duration(0)
}
func (msg *mockStorageGetter) GetMsgTimeout() time.Duration {
	return time.Duration(0)
}
func (msg *mockStorageGetter) GetMaxTxVectorLength() int {
	return 128
}

func (msg *mockStorageGetter) UpdateStorage(txn *badger.Txn, update dynamics.Updater) error {
	return nil
}
func (msg *mockStorageGetter) LoadStorage(txn *badger.Txn, epoch uint32) error {
	return nil
}

func (msg *mockStorageGetter) GetDataStoreEpochFee() *big.Int {
	return msg.dataStoreEpochFee
}

func (msg *mockStorageGetter) SetDataStoreEpochFee(value *big.Int) {
	if value == nil {
		panic("invalid value")
	}
	msg.dataStoreEpochFee.Set(value)
}

func (msg *mockStorageGetter) GetDataStoreValidVersion() uint32 {
	return 0
}

func (msg *mockStorageGetter) GetValueStoreFee() *big.Int {
	return msg.valueStoreFee
}

func (msg *mockStorageGetter) SetValueStoreFee(value *big.Int) {
	if value == nil {
		panic("invalid value")
	}
	msg.valueStoreFee.Set(value)
}

func (msg *mockStorageGetter) GetValueStoreValidVersion() uint32 {
	return 0
}

func (msg *mockStorageGetter) GetAtomicSwapFee() *big.Int {
	return msg.atomicSwapFee
}

func (msg *mockStorageGetter) SetAtomicSwapFee(value *big.Int) {
	if value == nil {
		panic("invalid value")
	}
	msg.atomicSwapFee.Set(value)
}

func (msg *mockStorageGetter) GetAtomicSwapValidStopEpoch() uint32 {
	return 0
}

func (msg *mockStorageGetter) GetMinTxFee() *big.Int {
	return msg.minTxFee
}

func (msg *mockStorageGetter) SetMinTxFee(value *big.Int) {
	if value == nil {
		panic("invalid value")
	}
	msg.minTxFee.Set(value)
}

func (msg *mockStorageGetter) GetTxValidVersion() uint32 {
	return 0
}
