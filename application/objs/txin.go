package objs

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/application/objs/txin"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// TXIn is a tx input object that acts as a reference to a UTXO
type TXIn struct {
	TXInLinker *TXInLinker
	Signature  []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// TXIn object
func (b *TXIn) UnmarshalBinary(data []byte) error {
	bc, err := txin.Unmarshal(data)
	if err != nil {
		return err
	}
	return b.UnmarshalCapn(bc)
}

// MarshalBinary takes the TXIn object and returns the canonical
// byte slice
func (b *TXIn) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("txin.marshalBinary; txin not initialized")
	}
	bc, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	return txin.Marshal(bc)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *TXIn) UnmarshalCapn(bc mdefs.TXIn) error {
	if err := txin.Validate(bc); err != nil {
		return err
	}
	b.TXInLinker = &TXInLinker{}
	if err := b.TXInLinker.UnmarshalCapn(bc.TXInLinker()); err != nil {
		return err
	}
	b.Signature = utils.CopySlice(bc.Signature())
	return nil
}

// MarshalCapn marshals the object into its capnproto definition
func (b *TXIn) MarshalCapn(seg *capnp.Segment) (mdefs.TXIn, error) {
	if b == nil {
		return mdefs.TXIn{}, errorz.ErrInvalid{}.New("txin.marshalCapn; txin not initialized")
	}
	var bc mdefs.TXIn
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bc, err
		}
		tmp, err := mdefs.NewRootTXIn(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	} else {
		tmp, err := mdefs.NewTXIn(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	}
	seg = bc.Struct.Segment()
	bt, err := b.TXInLinker.MarshalCapn(seg)
	if err != nil {
		return bc, err
	}
	if err := bc.SetTXInLinker(bt); err != nil {
		return bc, err
	}
	if err := bc.SetSignature(utils.CopySlice(b.Signature)); err != nil {
		return bc, err
	}
	return bc, nil
}

// PreHash returns the PreHash of the object
func (b *TXIn) PreHash() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("txin.preHash; txin not initialized")
	}
	return b.TXInLinker.PreHash()
}

// UTXOID returns the UTXOID of the object
func (b *TXIn) UTXOID() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("txin.utxoID; txin not initialized")
	}
	return b.TXInLinker.UTXOID()
}

// IsDeposit returns true if TXIn is a deposit; otherwise, returns false.
func (b *TXIn) IsDeposit() bool {
	if b == nil {
		return false
	}
	return b.TXInLinker.IsDeposit()
}

// TxHash returns the TxHash of the TXIn object
func (b *TXIn) TxHash() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("txin.txhash; txin not initialized")
	}
	if b.TXInLinker == nil {
		return nil, errorz.ErrInvalid{}.New("txin.txhash; txinl not initialized")
	}
	if len(b.TXInLinker.TxHash) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("txin.txhash; txinl.txhash has incorrect length")
	}
	return utils.CopySlice(b.TXInLinker.TxHash), nil
}

// SetTxHash sets the TxHash of the TXIn object
func (b *TXIn) SetTxHash(txHash []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("txin.setTxHash; txin not initialized")
	}
	if b.TXInLinker == nil {
		return errorz.ErrInvalid{}.New("txin.setTxHash; txinl not initialized")
	}
	if len(txHash) != constants.HashLen {
		return errorz.ErrInvalid{}.New("txin.setTxHash; txhash has incorrect length")
	}
	return b.TXInLinker.SetTxHash(txHash)
}

// ChainID returns the chain ID
func (b *TXIn) ChainID() (uint32, error) {
	if b == nil {
		return 0, errorz.ErrInvalid{}.New("txin.chainID; txin not initialized")
	}
	return b.TXInLinker.ChainID()
}

// ConsumedTxIdx returns the consumed TxIdx
func (b *TXIn) ConsumedTxIdx() (uint32, error) {
	if b == nil {
		return 0, errorz.ErrInvalid{}.New("txin.consumedTxIdx; txin not initialized")
	}
	return b.TXInLinker.ConsumedTxIdx()
}

// ConsumedTxHash returns the consumed TxHash
func (b *TXIn) ConsumedTxHash() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("txin.consumedTxHash; txin not initialized")
	}
	return b.TXInLinker.ConsumedTxHash()
}
