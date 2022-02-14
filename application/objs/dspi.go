package objs

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/application/objs/dspreimage"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// DSPreImage is a DataStore preimage
type DSPreImage struct {
	ChainID  uint32
	Index    []byte
	IssuedAt uint32
	Deposit  *uint256.Uint256
	RawData  []byte
	TXOutIdx uint32
	Owner    *DataStoreOwner
	Fee      *uint256.Uint256
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// DSPreImage object
func (b *DSPreImage) UnmarshalBinary(data []byte) error {
	bc, err := dspreimage.Unmarshal(data)
	if err != nil {
		return err
	}
	return b.UnmarshalCapn(bc)
}

// MarshalBinary takes the DSPreImage object and returns the canonical
// byte slice
func (b *DSPreImage) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("dspi.marshalBinary; dspi not initialized")
	}
	bc, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	return dspreimage.Marshal(bc)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *DSPreImage) UnmarshalCapn(bc mdefs.DSPreImage) error {
	if err := dspreimage.Validate(bc); err != nil {
		return err
	}
	b.ChainID = bc.ChainID()
	b.Index = utils.CopySlice(bc.Index())
	b.IssuedAt = bc.IssuedAt()
	u32array := [8]uint32{}
	u32array[0] = bc.Deposit()
	u32array[1] = bc.Deposit1()
	u32array[2] = bc.Deposit2()
	u32array[3] = bc.Deposit3()
	u32array[4] = bc.Deposit4()
	u32array[5] = bc.Deposit5()
	u32array[6] = bc.Deposit6()
	u32array[7] = bc.Deposit7()
	dObj := &uint256.Uint256{}
	err := dObj.FromUint32Array(u32array)
	if err != nil {
		return err
	}
	b.Deposit = dObj
	b.RawData = utils.CopySlice(bc.RawData())
	b.TXOutIdx = bc.TXOutIdx()
	owner := &DataStoreOwner{}
	if err := owner.UnmarshalBinary(bc.Owner()); err != nil {
		return err
	}
	b.Owner = owner
	fObj := &uint256.Uint256{}
	u32array[0] = bc.Fee0()
	u32array[1] = bc.Fee1()
	u32array[2] = bc.Fee2()
	u32array[3] = bc.Fee3()
	u32array[4] = bc.Fee4()
	u32array[5] = bc.Fee5()
	u32array[6] = bc.Fee6()
	u32array[7] = bc.Fee7()
	err = fObj.FromUint32Array(u32array)
	if err != nil {
		return err
	}
	b.Fee = fObj
	// protects against zero bytes errors in equations
	return b.ValidateDeposit()
}

// MarshalCapn marshals the object into its capnproto definition
func (b *DSPreImage) MarshalCapn(seg *capnp.Segment) (mdefs.DSPreImage, error) {
	if b == nil {
		return mdefs.DSPreImage{}, errorz.ErrInvalid{}.New("dspi.marshalCapn; dspi not initialized")
	}
	var bc mdefs.DSPreImage
	if err := b.ValidateDeposit(); err != nil {
		return bc, err
	}
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bc, err
		}
		tmp, err := mdefs.NewRootDSPreImage(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	} else {
		tmp, err := mdefs.NewDSPreImage(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	}
	if err := bc.SetIndex(utils.CopySlice(b.Index)); err != nil {
		return bc, err
	}
	if err := bc.SetRawData(utils.CopySlice(b.RawData)); err != nil {
		return bc, err
	}
	owner, err := b.Owner.MarshalBinary()
	if err != nil {
		return bc, err
	}
	if err := bc.SetOwner(owner); err != nil {
		return bc, err
	}
	bc.SetChainID(b.ChainID)
	bc.SetIssuedAt(b.IssuedAt)
	u32array, err := b.Deposit.ToUint32Array()
	if err != nil {
		return bc, err
	}
	bc.SetDeposit(u32array[0])
	bc.SetDeposit1(u32array[1])
	bc.SetDeposit2(u32array[2])
	bc.SetDeposit3(u32array[3])
	bc.SetDeposit4(u32array[4])
	bc.SetDeposit5(u32array[5])
	bc.SetDeposit6(u32array[6])
	bc.SetDeposit7(u32array[7])
	u32array, err = b.Fee.ToUint32Array()
	if err != nil {
		return bc, err
	}
	bc.SetFee0(u32array[0])
	bc.SetFee1(u32array[1])
	bc.SetFee2(u32array[2])
	bc.SetFee3(u32array[3])
	bc.SetFee4(u32array[4])
	bc.SetFee5(u32array[5])
	bc.SetFee6(u32array[6])
	bc.SetFee7(u32array[7])
	bc.SetTXOutIdx(b.TXOutIdx)
	return bc, nil
}

// PreHash returns the PreHash of the object
func (b *DSPreImage) PreHash() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("dspi.preHash; dspi not initialized")
	}
	msg, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	hsh := crypto.Hasher(msg)
	return hsh, nil
}

// RemainingValue returns remaining value at the time of consumption
func (b *DSPreImage) RemainingValue(currentHeight uint32) (*uint256.Uint256, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("dspi.remainingValue; dspi not initialized")
	}
	if b.IssuedAt == 0 {
		return nil, errorz.ErrInvalid{}.New("dspi.remainingValue; dspi.issuedAt is zero")
	}
	if b.Deposit == nil {
		return nil, errorz.ErrInvalid{}.New("dspi.remainingValue; dspi.deposit not initialized")
	}
	if len(b.RawData) == 0 {
		return nil, errorz.ErrInvalid{}.New("dspi.remainingValue; dspi.rawData has length zero")
	}
	epochFinal := utils.Epoch(currentHeight)
	epochInitial := b.IssuedAt
	if epochFinal < epochInitial {
		epochFinal = epochInitial
	}
	deposit := b.Deposit
	dataSize := uint32(len(b.RawData))
	result, err := RewardDepositEquation(deposit, dataSize, epochInitial, epochFinal)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Value returns the value stored in the object at the time of creation
func (b *DSPreImage) Value() (*uint256.Uint256, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("dspi.value; dspi not initialized")
	}
	if b.Deposit == nil {
		return nil, errorz.ErrInvalid{}.New("dspi.value; dspi.deposit not initialized")
	}
	if b.Deposit.IsZero() {
		return nil, errorz.ErrInvalid{}.New("dspi.value; dspi.deposit is zero")
	}
	return b.Deposit.Clone(), nil
}

// ValidatePreSignature validates the signature of the datastore at the time of
// creation
func (b *DSPreImage) ValidatePreSignature(msg []byte, sig *DataStoreSignature) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("dspi.validatePreSignature; dspi not initialized")
	}
	return b.Owner.ValidateSignature(msg, sig, false)
}

// ValidateSignature validates the signature of the datastore at the time of
// consumption
func (b *DSPreImage) ValidateSignature(currentHeight uint32, msg []byte, sig *DataStoreSignature) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("dspi.validateSignature; dspi not initialized")
	}
	isExpired, err := b.IsExpired(currentHeight)
	if err != nil {
		return err
	}
	return b.Owner.ValidateSignature(msg, sig, isExpired)
}

// IsExpired returns true if the datastore is free for garbage collection
func (b *DSPreImage) IsExpired(currentHeight uint32) (bool, error) {
	if b == nil {
		return true, errorz.ErrInvalid{}.New("dspi.isExpired; dspi not initialized")
	}
	eoe, err := b.EpochOfExpiration()
	if err != nil {
		return true, err
	}
	if utils.Epoch(currentHeight) >= eoe {
		return true, nil
	}
	return false, nil
}

// EpochOfExpiration returns the epoch in which the datastore may be garbage
// collected
func (b *DSPreImage) EpochOfExpiration() (uint32, error) {
	if b == nil {
		return 0, errorz.ErrInvalid{}.New("dspi.epochOfExpiration; dspi not initialized")
	}
	if b.Deposit == nil {
		return 0, errorz.ErrInvalid{}.New("dspi.epochOfExpiration; dspi.deposit not initialized")
	}
	if b.Deposit.IsZero() {
		return 0, errorz.ErrInvalid{}.New("dspi.epochOfExpiration; dspi.deposit is zero")
	}
	if len(b.RawData) == 0 {
		return 0, errorz.ErrInvalid{}.New("dspi.epochOfExpiration; dspi.rawData has length zero")
	}
	dataSize := uint32(len(b.RawData))
	numEpochs, err := NumEpochsEquation(dataSize, b.Deposit)
	if err != nil {
		return 0, err
	}
	eoe := b.IssuedAt + numEpochs + 1
	return eoe, nil
}

// ValidateDeposit validates the deposit.
// Validating the Fee portion of DSPreImage does *not* happen here,
// as these values may change each epoch.
// Furthermore, the Fee is validated elsewhere.
func (b *DSPreImage) ValidateDeposit() error {
	if b == nil {
		return errorz.ErrInvalid{}.New("dspi.validateDeposit; dspi not initialized")
	}
	if b.Deposit == nil {
		return errorz.ErrInvalid{}.New("dspi.validateDeposit; dspi.deposit not initialized")
	}
	if b.Deposit.IsZero() {
		return errorz.ErrInvalid{}.New("dspi.validateDeposit; dspi.deposit is zero")
	}
	if b.Fee == nil {
		return errorz.ErrInvalid{}.New("dspi.validateDeposit; dspi.fee not initialized")
	}
	if b.ChainID == 0 {
		return errorz.ErrInvalid{}.New("dspi.validateDeposit; dspi.chainID is zero")
	}
	if len(b.Index) != constants.HashLen {
		return errorz.ErrInvalid{}.New("dspi.validateDeposit; dspi.index has invalid length")
	}
	if b.IssuedAt == 0 {
		return errorz.ErrInvalid{}.New("dspi.validateDeposit; dspi.issuedAt is zero")
	}
	dataSize := uint32(len(b.RawData))
	if dataSize == 0 {
		return errorz.ErrInvalid{}.New("dspi.validateDeposit; dspi.rawData has length zero")
	}
	if dataSize > constants.MaxDataStoreSize {
		return errorz.ErrInvalid{}.New("dspi.validateDeposit; dspi.rawData is too large")
	}
	numEpochs, err := NumEpochsEquation(dataSize, b.Deposit)
	if err != nil {
		return err
	}
	if numEpochs == 0 {
		return errorz.ErrInvalid{}.New("dspi.validateDeposit; invalid deposit: storing for zero epochs")
	}
	deposit, err := BaseDepositEquation(dataSize, numEpochs)
	if err != nil {
		return err
	}
	if !deposit.Eq(b.Deposit) {
		return errorz.ErrInvalid{}.New("dspi.validateDeposit; invalid deposit: does not match computed value")
	}
	err = b.Owner.Validate()
	if err != nil {
		return err
	}
	return nil
}

// RewardDepositEquation allows a reward calculated for cleaning up an expired
// DataStore.
func RewardDepositEquation(deposit *uint256.Uint256, dataSize32 uint32, epochInitial uint32, epochFinal uint32) (*uint256.Uint256, error) {
	if deposit == nil {
		return nil, errorz.ErrInvalid{}.New("rewardDepositEquation: deposit is nil")
	}
	// These calculations will only be performed after a DataStore is validated,
	// so deposit and dataSize are correct.
	if epochFinal < epochInitial {
		return nil, errorz.ErrInvalid{}.New("rewardDepositEquation: epochFinal < epochInitial")
	}
	// This ensures
	//					epochDiff == epochFinal - epochInitial >= 0
	epochDiff := epochFinal - epochInitial
	dataSize, _ := new(uint256.Uint256).FromUint64(uint64(dataSize32))
	epochCost, _ := new(uint256.Uint256).Add(uint256.BaseDatasizeConst(), dataSize)
	numEpochs, err := NumEpochsEquation(dataSize32, deposit)
	if err != nil {
		return nil, err
	}
	if epochDiff > numEpochs {
		// The DataStore is now expired and the reward is the cost of an epoch
		return epochCost, nil
	}
	currentDeposit, err := BaseDepositEquation(dataSize32, epochDiff)
	if err != nil {
		return nil, err
	}
	tmp, err := new(uint256.Uint256).Sub(deposit, currentDeposit)
	if err != nil {
		return nil, err
	}
	tmp3, _ := new(uint256.Uint256).Mul(uint256.Two(), epochCost)
	remainder, err := new(uint256.Uint256).Add(tmp, tmp3)
	if err != nil {
		return nil, err
	}
	return remainder, nil
}

// BaseDepositEquation specifies a required deposit for a certain amount of
// data to persist for a specified number of epochs.
//
// The simple equation is
//
//		deposit = (dataSize + BaseDatasizeConst) * (2 + numEpochs)
func BaseDepositEquation(dataSize32 uint32, numEpochs32 uint32) (*uint256.Uint256, error) {
	if dataSize32 > constants.MaxDataStoreSize {
		// dataSize is too large so we do not perform any checks
		return nil, errorz.ErrInvalid{}.New("baseDepositEquation: dataSize is too large")
	}
	dataSize, _ := new(uint256.Uint256).FromUint64(uint64(dataSize32))
	epochCost, _ := new(uint256.Uint256).Add(dataSize, uint256.BaseDatasizeConst())
	numEpochs, _ := new(uint256.Uint256).FromUint64(uint64(numEpochs32))
	totalEpochs, _ := new(uint256.Uint256).Add(uint256.Two(), numEpochs)
	depositUint256, _ := new(uint256.Uint256).Mul(epochCost, totalEpochs)
	return depositUint256, nil
}

// NumEpochsEquation returns the number of epochs until expiration.
//
// The simple equation is
//
// 		numEpochs = (deposit / (dataSize + BaseDatasizeConst)) - 2
//
// We have additional checks to ensure there is no integer overflow.
func NumEpochsEquation(dataSize32 uint32, deposit *uint256.Uint256) (uint32, error) {
	if deposit == nil {
		return 0, errorz.ErrInvalid{}.New("numEpochsEquation: deposit is nil")
	}
	if dataSize32 > constants.MaxDataStoreSize {
		return 0, errorz.ErrInvalid{}.New("numEpochsEquation: dataSize is too large")
	}
	dataSize, _ := new(uint256.Uint256).FromUint64(uint64(dataSize32))
	totalDataSize, _ := new(uint256.Uint256).Add(dataSize, uint256.BaseDatasizeConst())
	tmp, err := new(uint256.Uint256).Div(deposit, totalDataSize)
	if err != nil {
		return 0, err
	}
	if tmp.Lt(uint256.Two()) {
		return 0, errorz.ErrInvalid{}.New("numEpochsEquation: invalid dataSize and deposit causing integer overflow")
	}
	// The above check ensures there is no integer overflow in this subtraction
	numEpochs, _ := new(uint256.Uint256).Sub(tmp, uint256.Two())
	numEpochs32, err := numEpochs.ToUint32()
	if err != nil {
		return 0, err
	}
	return numEpochs32, nil
}
