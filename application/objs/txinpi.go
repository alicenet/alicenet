package objs

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/application/objs/txinpreimage"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// TXInPreImage is the tx input preimage
type TXInPreImage struct {
	ChainID        uint32
	ConsumedTxIdx  uint32
	ConsumedTxHash []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// TXInPreImage object
func (b *TXInPreImage) UnmarshalBinary(data []byte) error {
	bc, err := txinpreimage.Unmarshal(data)
	if err != nil {
		return err
	}
	return b.UnmarshalCapn(bc)
}

// MarshalBinary takes the TXInPreImage object and returns the canonical
// byte slice
func (b *TXInPreImage) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("txinpi.marshalBinary; txinpi not initialized")
	}
	if b.ChainID == 0 {
		return nil, errorz.ErrInvalid{}.New("txinpi.marshalBinary; txinpi.chainID is zero")
	}
	if len(b.ConsumedTxHash) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("txinpi.marshalBinary; txinpi.consumedTxHash has incorrect length")
	}
	bc, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	return txinpreimage.Marshal(bc)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *TXInPreImage) UnmarshalCapn(bc mdefs.TXInPreImage) error {
	if err := txinpreimage.Validate(bc); err != nil {
		return err
	}
	b.ChainID = bc.ChainID()
	b.ConsumedTxIdx = bc.ConsumedTxIdx()
	b.ConsumedTxHash = utils.CopySlice(bc.ConsumedTxHash())
	return nil
}

// MarshalCapn marshals the object into its capnproto definition
func (b *TXInPreImage) MarshalCapn(seg *capnp.Segment) (mdefs.TXInPreImage, error) {
	if b == nil {
		return mdefs.TXInPreImage{}, errorz.ErrInvalid{}.New("txinpi.marshalCapn; txinpi not initialized")
	}
	var bc mdefs.TXInPreImage
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bc, err
		}
		tmp, err := mdefs.NewRootTXInPreImage(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	} else {
		tmp, err := mdefs.NewTXInPreImage(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	}
	if err := bc.SetConsumedTxHash(utils.CopySlice(b.ConsumedTxHash)); err != nil {
		return bc, err
	}
	bc.SetChainID(b.ChainID)
	bc.SetConsumedTxIdx(b.ConsumedTxIdx)
	return bc, nil
}

// PreHash returns the PreHash of the object
func (b *TXInPreImage) PreHash() ([]byte, error) {
	msg, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	hsh := crypto.Hasher(msg)
	return hsh, nil
}

// UTXOID returns the UTXOID of the object
func (b *TXInPreImage) UTXOID() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("txinpi.utxoID; txinpi not initialized")
	}
	if b.ChainID == 0 {
		return nil, errorz.ErrInvalid{}.New("txinpi.utxoID; txinpi.chainID is zero")
	}
	return MakeUTXOID(utils.CopySlice(b.ConsumedTxHash), b.ConsumedTxIdx), nil
}

// IsDeposit returns true if TXInPreImage is a deposit; otherwise, returns false.
func (b *TXInPreImage) IsDeposit() bool {
	if b == nil || b.ChainID == 0 {
		return false
	}
	return b.ConsumedTxIdx == constants.MaxUint32
}
