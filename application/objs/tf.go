package objs

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/application/objs/txfee"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "zombiezen.com/go/capnproto2"
)

// TxFee stores the transaction fee in a Tx
type TxFee struct {
	TFPreImage *TFPreImage
	TxHash     []byte
	//
	utxoID []byte
}

// New creates a new TxFee; fees must always
func (b *TxFee) New(chainID uint32, fee *uint256.Uint256) error {
	if chainID == 0 {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	if fee == nil || fee.IsZero() {
		return errorz.ErrInvalid{}.New("Error in TxFee.New: fee is nil")
	}
	tfp := &TFPreImage{
		ChainID: chainID,
		Fee:     fee.Clone(),
	}
	b.TFPreImage = tfp
	return nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// TxFee object
func (b *TxFee) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	bc, err := txfee.Unmarshal(data)
	if err != nil {
		return err
	}
	return b.UnmarshalCapn(bc)
}

// MarshalBinary takes the ValueStore object and returns the canonical
// byte slice
func (b *TxFee) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	bc, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	return txfee.Marshal(bc)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *TxFee) UnmarshalCapn(bc mdefs.TxFee) error {
	if err := txfee.Validate(bc); err != nil {
		return err
	}
	b.TFPreImage = &TFPreImage{}
	if err := b.TFPreImage.UnmarshalCapn(bc.TFPreImage()); err != nil {
		return err
	}
	b.TxHash = utils.CopySlice(bc.TxHash())
	return nil
}

// MarshalCapn marshals the object into its capnproto definition
func (b *TxFee) MarshalCapn(seg *capnp.Segment) (mdefs.TxFee, error) {
	if b == nil {
		return mdefs.TxFee{}, errorz.ErrInvalid{}.New("not initialized")
	}
	var bc mdefs.TxFee
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bc, err
		}
		tmp, err := mdefs.NewRootTxFee(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	} else {
		tmp, err := mdefs.NewTxFee(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	}
	seg = bc.Struct.Segment()
	bt, err := b.TFPreImage.MarshalCapn(seg)
	if err != nil {
		return bc, err
	}
	if err := bc.SetTFPreImage(bt); err != nil {
		return bc, err
	}
	if err := bc.SetTxHash(utils.CopySlice(b.TxHash)); err != nil {
		return bc, err
	}
	return bc, nil
}

// PreHash calculates the PreHash of the object
func (b *TxFee) PreHash() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.TFPreImage.PreHash()
}

// UTXOID calculates the UTXOID of the object
func (b *TxFee) UTXOID() ([]byte, error) {
	if b == nil || b.TFPreImage == nil || len(b.TxHash) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	if b.utxoID != nil {
		return utils.CopySlice(b.utxoID), nil
	}
	b.utxoID = MakeUTXOID(b.TxHash, b.TFPreImage.TXOutIdx)
	return utils.CopySlice(b.utxoID), nil
}

// TXOutIdx returns the TXOutIdx of the object
func (b *TxFee) TXOutIdx() (uint32, error) {
	if b == nil || b.TFPreImage == nil {
		return 0, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.TFPreImage.TXOutIdx, nil
}

// SetTXOutIdx sets the TXOutIdx of the object
func (b *TxFee) SetTXOutIdx(idx uint32) error {
	if b == nil || b.TFPreImage == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	b.TFPreImage.TXOutIdx = idx
	return nil
}

// SetTxHash sets the TxHash of the object
func (b *TxFee) SetTxHash(txHash []byte) error {
	if b == nil || b.TFPreImage == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	if len(txHash) != constants.HashLen {
		return errorz.ErrInvalid{}.New("Invalid hash length")
	}
	b.TxHash = utils.CopySlice(txHash)
	return nil
}

// ChainID returns the ChainID of the object
func (b *TxFee) ChainID() (uint32, error) {
	if b == nil || b.TFPreImage == nil || b.TFPreImage.ChainID == 0 {
		return 0, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.TFPreImage.ChainID, nil
}

// Fee returns the Fee of the object; Fee should always be nonzero
func (b *TxFee) Fee() (*uint256.Uint256, error) {
	if b == nil || b.TFPreImage == nil || b.TFPreImage.Fee == nil || b.TFPreImage.Fee.IsZero() {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.TFPreImage.Fee.Clone(), nil
}
