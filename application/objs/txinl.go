package objs

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/application/objs/txinlinker"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// TXInLinker ...
type TXInLinker struct {
	TXInPreImage *TXInPreImage
	TxHash       []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// TXInLinker object
func (b *TXInLinker) UnmarshalBinary(data []byte) error {
	bc, err := txinlinker.Unmarshal(data)
	if err != nil {
		return err
	}
	return b.UnmarshalCapn(bc)
}

// MarshalBinary takes the TXInLinker object and returns the canonical
// byte slice
func (b *TXInLinker) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("txinl.marshalBinary; txinl not initialized")
	}
	bc, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	return txinlinker.Marshal(bc)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *TXInLinker) UnmarshalCapn(bc mdefs.TXInLinker) error {
	if err := txinlinker.Validate(bc); err != nil {
		return err
	}
	b.TXInPreImage = &TXInPreImage{}
	if err := b.TXInPreImage.UnmarshalCapn(bc.TXInPreImage()); err != nil {
		return err
	}
	b.TxHash = utils.CopySlice(bc.TxHash())
	return nil
}

// MarshalCapn marshals the object into its capnproto definition
func (b *TXInLinker) MarshalCapn(seg *capnp.Segment) (mdefs.TXInLinker, error) {
	if b == nil {
		return mdefs.TXInLinker{}, errorz.ErrInvalid{}.New("txinl.marshalCapn; txinl not initialized")
	}
	var bc mdefs.TXInLinker
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bc, err
		}
		tmp, err := mdefs.NewRootTXInLinker(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	} else {
		tmp, err := mdefs.NewTXInLinker(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	}
	seg = bc.Struct.Segment()
	bt, err := b.TXInPreImage.MarshalCapn(seg)
	if err != nil {
		return bc, err
	}
	if err := bc.SetTXInPreImage(bt); err != nil {
		return bc, err
	}
	if err := bc.SetTxHash(utils.CopySlice(b.TxHash)); err != nil {
		return bc, err
	}
	return bc, nil
}

// PreHash returns the PreHash of the object
func (b *TXInLinker) PreHash() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("txinl.preHash; txinl not initialized")
	}
	return b.TXInPreImage.PreHash()
}

// UTXOID returns the UTXOID of the object
func (b *TXInLinker) UTXOID() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("txinl.utxoID; txinl not initialized")
	}
	return b.TXInPreImage.UTXOID()
}

// IsDeposit returns true if TXInLinker is a deposit; otherwise, returns false.
func (b *TXInLinker) IsDeposit() bool {
	if b == nil {
		return false
	}
	return b.TXInPreImage.IsDeposit()
}

// ChainID returns the chain ID
func (b *TXInLinker) ChainID() (uint32, error) {
	if b == nil {
		return 0, errorz.ErrInvalid{}.New("txinl.chainID; txinl not initialized")
	}
	if b.TXInPreImage == nil {
		return 0, errorz.ErrInvalid{}.New("txinl.chainID; txinpi not initialized")
	}
	if b.TXInPreImage.ChainID == 0 {
		return 0, errorz.ErrInvalid{}.New("txinl.chainID; txinpi.chainID is zero")
	}
	return b.TXInPreImage.ChainID, nil
}

// SetTxHash sets TxHash
func (b *TXInLinker) SetTxHash(txHash []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("txinl.setTxHash; txinl not initialized")
	}
	if b.TXInPreImage == nil {
		return errorz.ErrInvalid{}.New("txinl.setTxHash; txinpi not initialized")
	}
	if len(txHash) != constants.HashLen {
		return errorz.ErrInvalid{}.New("txinl.setTxHash; txhash has incorrect length")
	}
	b.TxHash = utils.CopySlice(txHash)
	return nil
}

// ConsumedTxIdx returns the consumed TxIdx
func (b *TXInLinker) ConsumedTxIdx() (uint32, error) {
	if b == nil {
		return 0, errorz.ErrInvalid{}.New("txinl.consumedTxIdx; txinl not initialized")
	}
	if b.TXInPreImage == nil {
		return 0, errorz.ErrInvalid{}.New("txinl.consumedTxIdx; txinpi not initialized")
	}
	return b.TXInPreImage.ConsumedTxIdx, nil
}

// ConsumedTxHash returns the consumed TxHash
func (b *TXInLinker) ConsumedTxHash() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("txinl.consumedTxHash; txinl not initialized")
	}
	if b.TXInPreImage == nil {
		return nil, errorz.ErrInvalid{}.New("txinl.consumedTxHash; txinpi not initialized")
	}
	if len(b.TXInPreImage.ConsumedTxHash) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("txinl.consumedTxHash; txinpi.txhash has incorrect length")
	}
	return utils.CopySlice(b.TXInPreImage.ConsumedTxHash), nil
}
