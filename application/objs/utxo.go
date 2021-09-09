package objs

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/application/objs/utxo"
	"github.com/MadBase/MadNet/application/wrapper"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "zombiezen.com/go/capnproto2"
)

// TXOut is a UTXO object
type TXOut struct {
	dataStore  *DataStore
	valueStore *ValueStore
	atomicSwap *AtomicSwap
	txFee      *TxFee
	// not part of serialized object below this line
	hasDataStore  bool
	hasValueStore bool
	hasAtomicSwap bool
	hasTxFee      bool
}

// CreateValueStore makes a new ValueStore
func (b *TXOut) CreateValueStore(chainID uint32, value *uint256.Uint256, fee *uint256.Uint256, acct []byte, curveSpec constants.CurveSpec, txHash []byte) error {
	vs := &ValueStore{}
	err := vs.New(chainID, value, fee, acct, curveSpec, txHash)
	if err != nil {
		return err
	}
	return b.NewValueStore(vs)
}

// CreateValueStoreFromDeposit makes a new ValueStore from a deposit
func (b *TXOut) CreateValueStoreFromDeposit(chainID uint32, value *uint256.Uint256, acct []byte, nonce []byte) error {
	vs := &ValueStore{}
	err := vs.NewFromDeposit(chainID, value, acct, nonce)
	if err != nil {
		return err
	}
	return b.NewValueStore(vs)
}

// NewDataStore makes a TXOut object which with the specified DataStore
func (b *TXOut) NewDataStore(v *DataStore) error {
	b.hasDataStore = true
	b.hasValueStore = false
	b.hasAtomicSwap = false
	b.hasTxFee = false
	b.dataStore = v
	b.atomicSwap = nil
	b.valueStore = nil
	b.txFee = nil
	return nil
}

// NewValueStore makes a TXOut object which with the specified ValueStore
func (b *TXOut) NewValueStore(v *ValueStore) error {
	b.hasDataStore = false
	b.hasValueStore = true
	b.hasAtomicSwap = false
	b.hasTxFee = false
	b.dataStore = nil
	b.valueStore = v
	b.atomicSwap = nil
	b.txFee = nil
	return nil
}

// NewAtomicSwap makes a TXOut object which with the specified AtomicSwap
func (b *TXOut) NewAtomicSwap(v *AtomicSwap) error {
	b.hasDataStore = false
	b.hasValueStore = false
	b.hasAtomicSwap = true
	b.hasTxFee = false
	b.dataStore = nil
	b.valueStore = nil
	b.atomicSwap = v
	b.txFee = nil
	return nil
}

// NewTxFee makes a TXOut object which with the specified TxFee
func (b *TXOut) NewTxFee(v *TxFee) error {
	b.hasDataStore = false
	b.hasValueStore = false
	b.hasAtomicSwap = false
	b.hasTxFee = true
	b.dataStore = nil
	b.valueStore = nil
	b.atomicSwap = nil
	b.txFee = v
	return nil
}

// HasDataStore specifies if the TXOut object has a DataStore
func (b *TXOut) HasDataStore() bool {
	if b == nil {
		return false
	}
	return b.hasDataStore
}

// HasValueStore specifies if the TXOut object has a ValueStore
func (b *TXOut) HasValueStore() bool {
	if b == nil {
		return false
	}
	return b.hasValueStore
}

// HasAtomicSwap specifies if the TXOut object has an AtomicSwap
func (b *TXOut) HasAtomicSwap() bool {
	if b == nil {
		return false
	}
	return b.hasAtomicSwap
}

// HasTxFee specifies if the TXOut object has a TxFee
func (b *TXOut) HasTxFee() bool {
	if b == nil {
		return false
	}
	return b.hasTxFee
}

// DataStore returns the DataStore of the TXOut object if it exists
func (b *TXOut) DataStore() (*DataStore, error) {
	if b.HasDataStore() {
		return b.dataStore, nil
	}
	return nil, errorz.ErrInvalid{}.New("object does not have a DataStore")
}

// ValueStore returns the ValueStore of the TXOut object if it exists
func (b *TXOut) ValueStore() (*ValueStore, error) {
	if b.HasValueStore() {
		return b.valueStore, nil
	}
	return nil, errorz.ErrInvalid{}.New("object does not have a ValueStore")
}

// AtomicSwap returns the AtomicSwap of the TXOut object if it exists
func (b *TXOut) AtomicSwap() (*AtomicSwap, error) {
	if b.HasAtomicSwap() {
		return b.atomicSwap, nil
	}
	return nil, errorz.ErrInvalid{}.New("object does not have an AtomicSwap")
}

// TxFee returns the TxFee of the TXOut object if it exists
func (b *TXOut) TxFee() (*TxFee, error) {
	if b.HasTxFee() {
		return b.txFee, nil
	}
	return nil, errorz.ErrInvalid{}.New("object does not have a TxFee")
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// TXOut object
func (b *TXOut) UnmarshalBinary(data []byte) error {
	bc, err := utxo.Unmarshal(data)
	if err != nil {
		return err
	}
	return b.UnmarshalCapn(bc)
}

// MarshalBinary takes the TXOut object and returns the canonical
// byte slice
func (b *TXOut) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	bc, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	return utxo.Marshal(bc)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *TXOut) UnmarshalCapn(bc mdefs.TXOut) error {
	switch {
	case bc.HasDataStore():
		cObj, err := bc.DataStore()
		if err != nil {
			return err
		}
		obj := &DataStore{}
		err = obj.UnmarshalCapn(cObj)
		if err != nil {
			return err
		}
		b.dataStore = obj
		b.hasDataStore = true
		b.valueStore = nil
		b.hasValueStore = false
		b.atomicSwap = nil
		b.hasAtomicSwap = false
		b.txFee = nil
		b.hasTxFee = false
	case bc.HasValueStore():
		cObj, err := bc.ValueStore()
		if err != nil {
			return err
		}
		obj := &ValueStore{}
		err = obj.UnmarshalCapn(cObj)
		if err != nil {
			return err
		}
		b.dataStore = nil
		b.hasDataStore = false
		b.valueStore = obj
		b.hasValueStore = true
		b.atomicSwap = nil
		b.hasAtomicSwap = false
		b.txFee = nil
		b.hasTxFee = false
	case bc.HasAtomicSwap():
		cObj, err := bc.AtomicSwap()
		if err != nil {
			return err
		}
		obj := &AtomicSwap{}
		err = obj.UnmarshalCapn(cObj)
		if err != nil {
			return err
		}
		b.dataStore = nil
		b.hasDataStore = false
		b.valueStore = nil
		b.hasValueStore = false
		b.atomicSwap = obj
		b.hasAtomicSwap = true
		b.txFee = nil
		b.hasTxFee = false
	case bc.HasTxFee():
		cObj, err := bc.TxFee()
		if err != nil {
			return err
		}
		obj := &TxFee{}
		err = obj.UnmarshalCapn(cObj)
		if err != nil {
			return err
		}
		b.dataStore = nil
		b.hasDataStore = false
		b.valueStore = nil
		b.hasValueStore = false
		b.atomicSwap = nil
		b.hasAtomicSwap = false
		b.txFee = obj
		b.hasTxFee = true
	default:
		return errorz.ErrInvalid{}.New("TXOut type not defined in UnmarshalCapn")
	}
	return nil
}

// MarshalCapn marshals the object into its capnproto definition
func (b *TXOut) MarshalCapn(seg *capnp.Segment) (mdefs.TXOut, error) {
	if b == nil {
		return mdefs.TXOut{}, errorz.ErrInvalid{}.New("not initialized")
	}
	var bc mdefs.TXOut
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bc, err
		}
		tmp, err := mdefs.NewRootTXOut(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	} else {
		tmp, err := mdefs.NewTXOut(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	}
	seg = bc.Struct.Segment()
	switch {
	case b.hasDataStore:
		ds, err := b.dataStore.MarshalCapn(seg)
		if err != nil {
			return bc, err
		}
		if err := bc.SetDataStore(ds); err != nil {
			return bc, err
		}
	case b.hasValueStore:
		vs, err := b.valueStore.MarshalCapn(seg)
		if err != nil {
			return bc, err
		}
		if err := bc.SetValueStore(vs); err != nil {
			return bc, err
		}
	case b.hasAtomicSwap:
		as, err := b.atomicSwap.MarshalCapn(seg)
		if err != nil {
			return bc, err
		}
		if err := bc.SetAtomicSwap(as); err != nil {
			return bc, err
		}
	case b.hasTxFee:
		tf, err := b.txFee.MarshalCapn(seg)
		if err != nil {
			return bc, err
		}
		if err := bc.SetTxFee(tf); err != nil {
			return bc, err
		}
	default:
		return mdefs.TXOut{}, errorz.ErrInvalid{}.New("TXOut type not defined in MarshalCapn")
	}
	return bc, nil
}

// PreHash returns the PreHash of the object
func (b *TXOut) PreHash() ([]byte, error) {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		return obj.PreHash()
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		return obj.PreHash()
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		return obj.PreHash()
	case b.HasTxFee():
		obj, _ := b.TxFee()
		return obj.PreHash()
	default:
		return nil, errorz.ErrInvalid{}.New("TXOut type not defined in PreHash")
	}
}

// UTXOID returns the UTXOID of the object
func (b *TXOut) UTXOID() ([]byte, error) {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		return obj.UTXOID()
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		return obj.UTXOID()
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		return obj.UTXOID()
	case b.HasTxFee():
		obj, _ := b.TxFee()
		return obj.UTXOID()
	default:
		return nil, errorz.ErrInvalid{}.New("TXOut type not defined in UTXOID")
	}
}

// ChainID returns the ChainID of the object
func (b *TXOut) ChainID() (uint32, error) {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		return obj.ChainID()
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		return obj.ChainID()
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		return obj.ChainID()
	case b.HasTxFee():
		obj, _ := b.TxFee()
		return obj.ChainID()
	default:
		return 0, errorz.ErrInvalid{}.New("TXOut type not defined for ChainID")
	}
}

// TXOutIdx returns the TXOutIdx of the object
func (b *TXOut) TXOutIdx() (uint32, error) {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		return obj.TXOutIdx()
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		return obj.TXOutIdx()
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		return obj.TXOutIdx()
	case b.HasTxFee():
		obj, _ := b.TxFee()
		return obj.TXOutIdx()
	default:
		return 0, errorz.ErrInvalid{}.New("TXOut type not defined in TXOutIdx")
	}
}

// SetTXOutIdx sets the TXOutIdx of the object
func (b *TXOut) SetTXOutIdx(idx uint32) error {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		return obj.SetTXOutIdx(idx)
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		return obj.SetTXOutIdx(idx)
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		return obj.SetTXOutIdx(idx)
	case b.HasTxFee():
		obj, _ := b.TxFee()
		return obj.SetTXOutIdx(idx)
	default:
		return errorz.ErrInvalid{}.New("TXOut type not defined in SetTXOutIdx")
	}
}

// TxHash returns the txHash from the object
func (b *TXOut) TxHash() ([]byte, error) {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		txHash, err := obj.TxHash()
		if err != nil {
			return nil, err
		}
		return utils.CopySlice(txHash), nil
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		if obj == nil || len(obj.TxHash) != constants.HashLen {
			return nil, errorz.ErrInvalid{}.New("not initialized")
		}
		return utils.CopySlice(obj.TxHash), nil
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		if obj == nil || len(obj.TxHash) != constants.HashLen {
			return nil, errorz.ErrInvalid{}.New("not initialized")
		}
		return utils.CopySlice(obj.TxHash), nil
	case b.HasTxFee():
		obj, _ := b.TxFee()
		if obj == nil || len(obj.TxHash) != constants.HashLen {
			return nil, errorz.ErrInvalid{}.New("not initialized")
		}
		return utils.CopySlice(obj.TxHash), nil
	default:
		return nil, errorz.ErrInvalid{}.New("TXOut type not defined in TxHash")
	}
}

// SetTxHash sets the txHash of the object
func (b *TXOut) SetTxHash(txHash []byte) error {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		return obj.SetTxHash(utils.CopySlice(txHash))
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		return obj.SetTxHash(utils.CopySlice(txHash))
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		return obj.SetTxHash(utils.CopySlice(txHash))
	case b.HasTxFee():
		obj, _ := b.TxFee()
		return obj.SetTxHash(utils.CopySlice(txHash))
	default:
		return errorz.ErrInvalid{}.New("TXOut type not defined in SetTxHash")
	}
}

// IsExpired returns true if the utxo has expired
func (b *TXOut) IsExpired(currentHeight uint32) (bool, error) {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		return obj.IsExpired(currentHeight)
	case b.HasValueStore():
		return false, nil
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		return obj.IsExpired(currentHeight)
	default:
		return false, errorz.ErrInvalid{}.New("TXOut type not defined in IsExpired")
	}
}

// RemainingValue returns the remaining value after discount
func (b *TXOut) RemainingValue(currentHeight uint32) (*uint256.Uint256, error) {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		return obj.RemainingValue(currentHeight)
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		return obj.Value()
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		return obj.Value()
	default:
		return nil, errorz.ErrInvalid{}.New("TXOut type not defined in RemainingValue")
	}
}

// MakeTxIn returns a TXIn for the object
func (b *TXOut) MakeTxIn() (*TXIn, error) {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		return obj.MakeTxIn()
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		return obj.MakeTxIn()
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		return obj.MakeTxIn()
	default:
		return nil, errorz.ErrInvalid{}.New("TXOut type not defined in MakeTxIn")
	}
}

// Value returns the Value of the object
func (b *TXOut) Value() (*uint256.Uint256, error) {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		return obj.Value()
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		return obj.Value()
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		return obj.Value()
	default:
		return nil, errorz.ErrInvalid{}.New("TXOut type not defined in Value")
	}
}

// ValuePlusFee returns the Value of the object plus the associated fee
func (b *TXOut) ValuePlusFee() (*uint256.Uint256, error) {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		return obj.ValuePlusFee()
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		return obj.ValuePlusFee()
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		return obj.ValuePlusFee()
	case b.HasTxFee():
		// We return the fee because ValuePlusFee is called to calculate
		// the total value out of the Tx; this must include TxFee.
		obj, _ := b.TxFee()
		return obj.Fee()
	default:
		return nil, errorz.ErrInvalid{}.New("TXOut type not defined in ValuePlusFee")
	}
}

// ValidateFee validates the Fee of the object
func (b *TXOut) ValidateFee(storage *wrapper.Storage) error {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		return obj.ValidateFee(storage)
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		return obj.ValidateFee(storage)
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		return obj.ValidateFee(storage)
	case b.HasTxFee():
		// We cannot validate the total transaction fee here because
		// we must look at the *entire* Tx object to ensure at most one TxFee
		// UTXO is present
		return nil
	default:
		return errorz.ErrInvalid{}.New("TXOut type not defined in ValidateFee")
	}
}

// ValidatePreSignature validates the PreSignature of the object
func (b *TXOut) ValidatePreSignature() error {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		return obj.ValidatePreSignature()
	case b.HasValueStore():
		return nil
	case b.HasAtomicSwap():
		return nil
	case b.HasTxFee():
		return nil
	default:
		return errorz.ErrInvalid{}.New("TXOut type not defined in ValidatePreSignature")
	}
}

// ValidateSignature validates the signature of the txIn against the UTXO
func (b *TXOut) ValidateSignature(currentHeight uint32, txIn *TXIn) error {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		return obj.ValidateSignature(currentHeight, txIn)
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		return obj.ValidateSignature(txIn)
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		return obj.ValidateSignature(currentHeight, txIn)
	case b.HasTxFee():
		return nil
	default:
		return errorz.ErrInvalid{}.New("TXOut type not defined in ValidateSignature")
	}
}

// MustBeMinedBeforeHeight ...
func (b *TXOut) MustBeMinedBeforeHeight() (uint32, error) {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		iat, err := obj.IssuedAt()
		if err != nil {
			return 0, err
		}
		return (iat * constants.EpochLength) - 1, nil
	case b.HasValueStore():
		return constants.MaxUint32, nil
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		iat, err := obj.IssuedAt()
		if err != nil {
			return 0, err
		}
		return (iat * constants.EpochLength) - 1, nil
	case b.HasTxFee():
		return constants.MaxUint32, nil
	default:
		return 0, errorz.ErrInvalid{}.New("TXOut type not defined in MustBeMinedBeforeHeight")
	}
}

// CannotBeMinedBeforeHeight ...
func (b *TXOut) CannotBeMinedBeforeHeight() (uint32, error) {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		iat, err := obj.IssuedAt()
		if err != nil {
			return 0, err
		}
		return (iat-1)*constants.EpochLength + 1, nil
	case b.HasValueStore():
		return 1, nil
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		iat, err := obj.IssuedAt()
		if err != nil {
			return 0, err
		}
		return (iat-1)*constants.EpochLength + 1, nil
	case b.HasTxFee():
		return 1, nil
	default:
		return 0, errorz.ErrInvalid{}.New("TXOut type not defined in MustBeMinedBeforeHeight")
	}
}

// Account returns the account from the TXOut
func (b *TXOut) Account() ([]byte, error) {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		dso, err := obj.Owner()
		if err != nil {
			return nil, err
		}
		return utils.CopySlice(dso.Account), nil
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		vso, err := obj.Owner()
		if err != nil {
			return nil, err
		}
		return utils.CopySlice(vso.Account), nil
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		aso, err := obj.Owner()
		if err != nil {
			return nil, err
		}
		asoPrimaryAcct, err := aso.PrimaryAccount()
		if err != nil {
			return nil, err
		}
		return utils.CopySlice(asoPrimaryAcct), nil
	default:
		return nil, errorz.ErrInvalid{}.New("TXOut type not defined in Account")
	}
}

// GenericOwner returns the Owner from the TXOut
func (b *TXOut) GenericOwner() (*Owner, error) {
	switch {
	case b.HasDataStore():
		obj, _ := b.DataStore()
		onr, err := obj.GenericOwner()
		if err != nil {
			return nil, err
		}
		return onr, nil
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		onr, err := obj.GenericOwner()
		if err != nil {
			return nil, err
		}
		return onr, nil
	case b.HasAtomicSwap():
		obj, _ := b.AtomicSwap()
		onr, err := obj.GenericOwner()
		if err != nil {
			return nil, err
		}
		return onr, nil
	default:
		return nil, errorz.ErrInvalid{}.New("TXOut type not defined in GenericOwner")
	}
}

// IsDeposit returns true if it is a valid ValueStore with deposit.
// All other instances return false.
func (b *TXOut) IsDeposit() bool {
	if b == nil {
		return false
	}
	switch {
	case b.HasValueStore():
		obj, _ := b.ValueStore()
		return obj.IsDeposit()
	default:
		return false
	}
}
