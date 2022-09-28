package wrapper

import (
	"math/big"
	"testing"

	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/dynamics/mocks"
	"github.com/stretchr/testify/assert"
)

func TestStorageGetMaxBytesFailsWhenNotInitialised(t *testing.T) {
	t.Parallel()
	s := make([]*Storage, 1)
	fee, err := s[0].GetMaxBlockSize()
	assert.Error(t, err)
	assert.Zero(t, fee)
}

func TestStorageGetDataStoreEpochFeeFailsWhenNotInitialised(t *testing.T) {
	t.Parallel()
	s := make([]*Storage, 1)
	fee, err := s[0].GetDataStoreFee()
	assert.Error(t, err)
	assert.Nil(t, fee)
}

func TestStorageGetValueStoreFeeFailsWhenNotInitialised(t *testing.T) {
	t.Parallel()
	s := make([]*Storage, 1)
	fee, err := s[0].GetValueStoreFee()
	assert.Error(t, err)
	assert.Nil(t, fee)
}

func TestStorageGetMinTxFeeFailsWhenNotInitialised(t *testing.T) {
	t.Parallel()
	s := make([]*Storage, 1)
	fee, err := s[0].GetMinScaledTransactionFee()
	assert.Error(t, err)
	assert.Nil(t, fee)
}

func TestStorageWithoutStorageGetterGetMaxBytesFails(t *testing.T) {
	t.Parallel()
	s := &Storage{}
	fee, err := s.GetMaxBlockSize()
	assert.Error(t, err)
	assert.Zero(t, fee)
}

func TestStorageWithoutStorageGetterGetDataStoreEpochFeeFails(t *testing.T) {
	t.Parallel()
	s := &Storage{}
	fee, err := s.GetDataStoreFee()
	assert.Error(t, err)
	assert.Nil(t, fee)
}

func TestStorageWithoutStorageGetterGetValueStoreFeeFails(t *testing.T) {
	t.Parallel()
	s := &Storage{}
	fee, err := s.GetValueStoreFee()
	assert.Error(t, err)
	assert.Nil(t, fee)
}

func TestStorageWithoutStorageGetterGetMinTxFeeFails(t *testing.T) {
	t.Parallel()
	s := &Storage{}
	fee, err := s.GetMinScaledTransactionFee()
	assert.Error(t, err)
	assert.Nil(t, fee)
}

func TestStorageInitialisedReturnsExpectedMaxBytes(t *testing.T) {
	t.Parallel()
	msg := mocks.NewMockStorageGetter()
	s := NewStorage(msg)
	expectedMaxBytes := uint32(123)
	msg.GetMaxBlockSizeFunc.SetDefaultReturn(expectedMaxBytes)

	fee, err := s.GetMaxBlockSize()
	assert.NoError(t, err)
	assert.Equal(t, expectedMaxBytes, fee)
}

func TestStorageInitialisedReturnsExpectedDataStoreEpochFee(t *testing.T) {
	t.Parallel()
	msg := mocks.NewMockStorageGetter()
	s := NewStorage(msg)
	expectedFee := big.NewInt(123)
	expectedFeeUint256 := &uint256.Uint256{}
	_, err := expectedFeeUint256.FromBigInt(expectedFee)
	assert.NoError(t, err)
	msg.GetDataStoreFeeFunc.SetDefaultReturn(expectedFee)

	fee, err := s.GetDataStoreFee()
	assert.NoError(t, err)
	assert.Equal(t, expectedFeeUint256, fee)
}

func TestStorageInitialisedReturnsExpectedValueStoreFee(t *testing.T) {
	t.Parallel()
	msg := mocks.NewMockStorageGetter()
	s := NewStorage(msg)
	expectedFee := big.NewInt(123)
	expectedFeeUint256 := &uint256.Uint256{}
	_, err := expectedFeeUint256.FromBigInt(expectedFee)
	assert.NoError(t, err)

	msg.GetValueStoreFeeFunc.SetDefaultReturn(expectedFee)

	fee, err := s.GetValueStoreFee()
	assert.NoError(t, err)
	assert.Equal(t, expectedFeeUint256, fee)
}

func TestStorageInitialisedReturnsExpectedMinTxFee(t *testing.T) {
	t.Parallel()
	msg := mocks.NewMockStorageGetter()
	s := NewStorage(msg)
	expectedFee := big.NewInt(123)
	expectedFeeUint256 := &uint256.Uint256{}
	_, err := expectedFeeUint256.FromBigInt(expectedFee)
	assert.NoError(t, err)

	msg.GetMinScaledTransactionFeeFunc.SetDefaultReturn(expectedFee)

	fee, err := s.GetMinScaledTransactionFee()
	assert.NoError(t, err)
	assert.Equal(t, expectedFeeUint256, fee)
}
