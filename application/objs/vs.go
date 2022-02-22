package objs

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/application/objs/valuestore"
	"github.com/MadBase/MadNet/application/wrapper"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// ValueStore stores value in a UTXO
type ValueStore struct {
	VSPreImage *VSPreImage
	TxHash     []byte
	//
	utxoID []byte
}

// New creates a new ValueStore
func (b *ValueStore) New(chainID uint32, value *uint256.Uint256, fee *uint256.Uint256, acct []byte, curveSpec constants.CurveSpec, txHash []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("vs.new: vs not initialized")
	}
	if value == nil {
		return errorz.ErrInvalid{}.New("vs.new: value is nil")
	}
	if value.IsZero() {
		return errorz.ErrInvalid{}.New("vs.new: value is zero")
	}
	if fee == nil {
		return errorz.ErrInvalid{}.New("vs.new: fee is nil")
	}
	vsowner := &ValueStoreOwner{}
	vsowner.New(acct, curveSpec)
	if err := vsowner.Validate(); err != nil {
		return err
	}
	if chainID == 0 {
		return errorz.ErrInvalid{}.New("vs.new: chainID is zero")
	}
	if len(txHash) != constants.HashLen {
		return errorz.ErrInvalid{}.New("vs.new: invalid txHash; incorrect txhash length")
	}
	vsp := &VSPreImage{
		ChainID:  chainID,
		Value:    value.Clone(),
		TXOutIdx: constants.MaxUint32,
		Owner:    vsowner,
		Fee:      fee.Clone(),
	}
	b.VSPreImage = vsp
	b.TxHash = utils.CopySlice(txHash)
	return nil
}

// NewFromDeposit creates a new ValueStore from a deposit event
func (b *ValueStore) NewFromDeposit(chainID uint32, value *uint256.Uint256, acct []byte, nonce []byte) error {
	vsowner := &ValueStoreOwner{}
	vsowner.New(acct, constants.CurveSecp256k1)
	if err := vsowner.Validate(); err != nil {
		return err
	}
	if chainID == 0 {
		return errorz.ErrInvalid{}.New("vs.newFromDeposit: chainID is zero")
	}
	if len(nonce) != constants.HashLen {
		return errorz.ErrInvalid{}.New("vs.newFromDeposit: invalid nonce; incorrect nonce length")
	}
	if value == nil {
		return errorz.ErrInvalid{}.New("vs.newFromDeposit: value is nil")
	}
	if value.IsZero() {
		return errorz.ErrInvalid{}.New("vs.newFromDeposit: value is zero")
	}
	vsp := &VSPreImage{
		ChainID:  chainID,
		Value:    value,
		TXOutIdx: constants.MaxUint32,
		Owner:    vsowner,
		Fee:      uint256.Zero(),
	}
	b.VSPreImage = vsp
	b.TxHash = utils.CopySlice(nonce)
	return nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// ValueStore object
func (b *ValueStore) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("vs.unmarshalBinary: vs not initialized")
	}
	bc, err := valuestore.Unmarshal(data)
	if err != nil {
		return err
	}
	return b.UnmarshalCapn(bc)
}

// MarshalBinary takes the ValueStore object and returns the canonical
// byte slice
func (b *ValueStore) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("vs.marshalBinary: vs not initialized")
	}
	bc, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	return valuestore.Marshal(bc)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *ValueStore) UnmarshalCapn(bc mdefs.ValueStore) error {
	if err := valuestore.Validate(bc); err != nil {
		return err
	}
	b.VSPreImage = &VSPreImage{}
	if err := b.VSPreImage.UnmarshalCapn(bc.VSPreImage()); err != nil {
		return err
	}
	b.TxHash = utils.CopySlice(bc.TxHash())
	return nil
}

// MarshalCapn marshals the object into its capnproto definition
func (b *ValueStore) MarshalCapn(seg *capnp.Segment) (mdefs.ValueStore, error) {
	if b == nil {
		return mdefs.ValueStore{}, errorz.ErrInvalid{}.New("vs.marshalCapn: vs not initialized")
	}
	var bc mdefs.ValueStore
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bc, err
		}
		tmp, err := mdefs.NewRootValueStore(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	} else {
		tmp, err := mdefs.NewValueStore(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	}
	seg = bc.Struct.Segment()
	bt, err := b.VSPreImage.MarshalCapn(seg)
	if err != nil {
		return bc, err
	}
	if err := bc.SetVSPreImage(bt); err != nil {
		return bc, err
	}
	if err := bc.SetTxHash(utils.CopySlice(b.TxHash)); err != nil {
		return bc, err
	}
	return bc, nil
}

// PreHash calculates the PreHash of the object
func (b *ValueStore) PreHash() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("vs.preHash: vs not initialized")
	}
	return b.VSPreImage.PreHash()
}

// UTXOID calculates the UTXOID of the object
func (b *ValueStore) UTXOID() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("vs.utxoID: vs not initialized")
	}
	if b.VSPreImage == nil {
		return nil, errorz.ErrInvalid{}.New("vs.utxoID: vspi not initialized")
	}
	if len(b.TxHash) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("vs.utxoID: vs.txhash has incorrect length")
	}
	if b.utxoID != nil {
		return utils.CopySlice(b.utxoID), nil
	}
	b.utxoID = MakeUTXOID(b.TxHash, b.VSPreImage.TXOutIdx)
	return utils.CopySlice(b.utxoID), nil
}

// TxOutIdx returns the TxOutIdx of the object
func (b *ValueStore) TxOutIdx() (uint32, error) {
	if b == nil {
		return 0, errorz.ErrInvalid{}.New("vs.txOutIdx: vs not initialized")
	}
	if b.VSPreImage == nil {
		return 0, errorz.ErrInvalid{}.New("vs.txOutIdx: vspi not initialized")
	}
	return b.VSPreImage.TXOutIdx, nil
}

// SetTxOutIdx sets the TxOutIdx of the object
func (b *ValueStore) SetTxOutIdx(idx uint32) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("vs.setTxOutIdx: vs not initialized")
	}
	if b.VSPreImage == nil {
		return errorz.ErrInvalid{}.New("vs.setTxOutIdx: vspi not initialized")
	}
	b.VSPreImage.TXOutIdx = idx
	return nil
}

// SetTxHash sets the TxHash of the object
func (b *ValueStore) SetTxHash(txHash []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("vs.setTxHash: vs not initialized")
	}
	if b.VSPreImage == nil {
		return errorz.ErrInvalid{}.New("vs.setTxHash: vspi not initialized")
	}
	if len(txHash) != constants.HashLen {
		return errorz.ErrInvalid{}.New("vs.setTxHash: invalid hash length")
	}
	b.TxHash = utils.CopySlice(txHash)
	return nil
}

// ChainID returns the ChainID of the object
func (b *ValueStore) ChainID() (uint32, error) {
	if b == nil {
		return 0, errorz.ErrInvalid{}.New("vs.chainID: vs not initialized")
	}
	if b.VSPreImage == nil {
		return 0, errorz.ErrInvalid{}.New("vs.chainID: vspi not initialized")
	}
	if b.VSPreImage.ChainID == 0 {
		return 0, errorz.ErrInvalid{}.New("vs.chainID: chainID is zero")
	}
	return b.VSPreImage.ChainID, nil
}

// Value returns the Value of the object
func (b *ValueStore) Value() (*uint256.Uint256, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("vs.value: vs not initialized")
	}
	if b.VSPreImage == nil {
		return nil, errorz.ErrInvalid{}.New("vs.value: vspi not initialized")
	}
	if b.VSPreImage.Value == nil {
		return nil, errorz.ErrInvalid{}.New("vs.value: vspi.value not initialized")
	}
	if b.VSPreImage.Value.IsZero() {
		return nil, errorz.ErrInvalid{}.New("vs.value: vspi.value is zero")
	}
	return b.VSPreImage.Value.Clone(), nil
}

// Fee returns the Fee of the object
func (b *ValueStore) Fee() (*uint256.Uint256, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("vs.fee: vs not initialized")
	}
	if b.VSPreImage == nil {
		return nil, errorz.ErrInvalid{}.New("vs.fee: vspi not initialized")
	}
	if b.VSPreImage.Fee == nil {
		return nil, errorz.ErrInvalid{}.New("vs.fee: vspi.fee not initialized")
	}
	return b.VSPreImage.Fee.Clone(), nil
}

// ValuePlusFee returns the Value of the object with the associated fee
func (b *ValueStore) ValuePlusFee() (*uint256.Uint256, error) {
	value, err := b.Value()
	if err != nil {
		return nil, err
	}
	fee, err := b.Fee()
	if err != nil {
		return nil, err
	}
	total, err := new(uint256.Uint256).Add(value, fee)
	if err != nil {
		return nil, err
	}
	return total, nil
}

// IsDeposit returns true if the object is a deposit
func (b *ValueStore) IsDeposit() bool {
	if b == nil || b.VSPreImage == nil {
		return false
	}
	return b.VSPreImage.TXOutIdx == constants.MaxUint32
}

// Owner returns the ValueStoreOwner of the ValueStore
func (b *ValueStore) Owner() (*ValueStoreOwner, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("vs.owner: vs not initialized")
	}
	if b.VSPreImage == nil {
		return nil, errorz.ErrInvalid{}.New("vs.owner: vspi not initialized")
	}
	if err := b.VSPreImage.Owner.Validate(); err != nil {
		return nil, errorz.ErrInvalid{}.New("vs.owner: ValueStoreOwner invalid")
	}
	return b.VSPreImage.Owner, nil
}

// GenericOwner returns the Owner of the ValueStore
func (b *ValueStore) GenericOwner() (*Owner, error) {
	vso, err := b.Owner()
	if err != nil {
		return nil, err
	}
	onr := &Owner{}
	if err := onr.NewFromValueStoreOwner(vso); err != nil {
		return nil, err
	}
	return onr, nil
}

// Sign generates the signature for a ValueStore at the time of consumption
func (b *ValueStore) Sign(txIn *TXIn, s Signer) error {
	if txIn == nil {
		return errorz.ErrInvalid{}.New("vs.sign: txin not initialized")
	}
	msg, err := txIn.TXInLinker.MarshalBinary()
	if err != nil {
		return err
	}
	owner, err := b.Owner()
	if err != nil {
		return err
	}
	sig, err := owner.Sign(msg, s)
	if err != nil {
		return err
	}
	sigb, err := sig.MarshalBinary()
	if err != nil {
		return err
	}
	txIn.Signature = sigb
	return nil
}

// ValidateFee validates the fee of the object at the time of creation
func (b *ValueStore) ValidateFee(storage *wrapper.Storage) error {
	fee, err := b.Fee()
	if err != nil {
		return err
	}
	if b.IsDeposit() {
		if !fee.IsZero() {
			return errorz.ErrInvalid{}.New("vs.validateFee: invalid fee; deposits should have fee equal zero")
		}
		return nil
	}
	feeTrue, err := storage.GetValueStoreFee()
	if err != nil {
		return err
	}
	if fee.Cmp(feeTrue) != 0 {
		return errorz.ErrInvalid{}.New("vs.validateFee: invalid fee")
	}
	return nil
}

// ValidateSignature validates the signature of the ValueStore at the time of
// consumption
func (b *ValueStore) ValidateSignature(txIn *TXIn) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("vs.validateSignature: vs not initialized")
	}
	if txIn == nil {
		return errorz.ErrInvalid{}.New("vs.validateSignature: txin not initialized")
	}
	msg, err := txIn.TXInLinker.MarshalBinary()
	if err != nil {
		return err
	}
	sig := &ValueStoreSignature{}
	if err := sig.UnmarshalBinary(txIn.Signature); err != nil {
		return err
	}
	return b.VSPreImage.ValidateSignature(msg, sig)
}

// MakeTxIn constructs a TXIn object for the current object
func (b *ValueStore) MakeTxIn() (*TXIn, error) {
	txOutIdx, err := b.TxOutIdx()
	if err != nil {
		return nil, err
	}
	cid, err := b.ChainID()
	if err != nil {
		return nil, err
	}
	if len(b.TxHash) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("vs.makeTxIn: invalid TxHash")
	}
	return &TXIn{
		TXInLinker: &TXInLinker{
			TXInPreImage: &TXInPreImage{
				ConsumedTxIdx:  txOutIdx,
				ConsumedTxHash: utils.CopySlice(b.TxHash),
				ChainID:        cid,
			},
		},
	}, nil
}
