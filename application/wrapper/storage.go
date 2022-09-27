package wrapper

import (
	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/dynamics"
	"github.com/alicenet/alicenet/errorz"
)

// Storage wraps the dynamics.StorageGetter interface to make
// it easier to interact within application logic.
type Storage struct {
	storage dynamics.StorageGetter
}

// NewStorage creates a new storage struct which wraps
// the StorageGetter interface.
func NewStorage(storageInter dynamics.StorageGetter) *Storage {
	storage := &Storage{storage: storageInter}
	return storage
}

// GetMaxBytes returns MaxBytes
func (s *Storage) GetMaxBytes() (uint32, error) {
	if s == nil {
		return 0, errorz.ErrInvalid{}.New("storage.MaxBlockSize; struct not initialized")
	}
	if s.storage == nil {
		return 0, errorz.ErrInvalid{}.New("storage.MaxBlockSize; storage not initialized")
	}
	return s.storage.GetMaxBlockSize(), nil
}

// GetDataStoreEpochFee returns the per-epoch fee of DataStore
func (s *Storage) GetDataStoreEpochFee() (*uint256.Uint256, error) {
	if s == nil {
		return nil, errorz.ErrInvalid{}.New("storage.GetDataStoreEpochFee; struct not initialized")
	}
	if s.storage == nil {
		return nil, errorz.ErrInvalid{}.New("storage.GetDataStoreEpochFee; storage not initialized")
	}
	fee := s.storage.GetDataStoreFee()
	feeUint256 := &uint256.Uint256{}
	_, err := feeUint256.FromBigInt(fee)
	if err != nil {
		return nil, err
	}
	return feeUint256, nil
}

// GetValueStoreFee returns the fee of ValueStore.
func (s *Storage) GetValueStoreFee() (*uint256.Uint256, error) {
	if s == nil {
		return nil, errorz.ErrInvalid{}.New("storage.GetValueStoreFee; struct not initialized")
	}
	if s.storage == nil {
		return nil, errorz.ErrInvalid{}.New("storage.GetValueStoreFee; storage not initialized")
	}
	fee := s.storage.GetValueStoreFee()
	feeUint256 := &uint256.Uint256{}
	_, err := feeUint256.FromBigInt(fee)
	if err != nil {
		return nil, err
	}
	return feeUint256, nil
}

// GetMinTxFee returns the minimum TxFee.
func (s *Storage) GetMinTxFee() (*uint256.Uint256, error) {
	if s == nil {
		return nil, errorz.ErrInvalid{}.New("storage.GetMinTxFee; struct not initialized")
	}
	if s.storage == nil {
		return nil, errorz.ErrInvalid{}.New("storage.GetMinTxFee; storage not initialized")
	}
	fee := s.storage.GetMinScaledTransactionFee()
	feeUint256 := &uint256.Uint256{}
	_, err := feeUint256.FromBigInt(fee)
	if err != nil {
		return nil, err
	}
	return feeUint256, nil
}
