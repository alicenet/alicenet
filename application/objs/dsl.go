package objs

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/application/objs/dslinker"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "zombiezen.com/go/capnproto2"
)

// DSLinker links the DSPreImage to the DataStore
type DSLinker struct {
	DSPreImage *DSPreImage
	TxHash     []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// DSLinker object
func (b *DSLinker) UnmarshalBinary(data []byte) error {
	bc, err := dslinker.Unmarshal(data)
	if err != nil {
		return err
	}
	return b.UnmarshalCapn(bc)
}

// MarshalBinary takes the DSLinker object and returns the canonical
// byte slice
func (b *DSLinker) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	bc, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	return dslinker.Marshal(bc)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *DSLinker) UnmarshalCapn(bc mdefs.DSLinker) error {
	if err := dslinker.Validate(bc); err != nil {
		return err
	}
	dSPreImage := &DSPreImage{}
	if err := dSPreImage.UnmarshalCapn(bc.DSPreImage()); err != nil {
		return err
	}
	b.DSPreImage = dSPreImage
	b.TxHash = utils.CopySlice(bc.TxHash())
	return nil
}

// MarshalCapn marshals the object into its capnproto definition
func (b *DSLinker) MarshalCapn(seg *capnp.Segment) (mdefs.DSLinker, error) {
	if b == nil {
		return mdefs.DSLinker{}, errorz.ErrInvalid{}.New("not initialized")
	}
	var bc mdefs.DSLinker
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bc, err
		}
		tmp, err := mdefs.NewRootDSLinker(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	} else {
		tmp, err := mdefs.NewDSLinker(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	}
	seg = bc.Struct.Segment()
	bt, err := b.DSPreImage.MarshalCapn(seg)
	if err != nil {
		return bc, err
	}
	if err := bc.SetDSPreImage(bt); err != nil {
		return bc, err
	}
	if err := bc.SetTxHash(utils.CopySlice(b.TxHash)); err != nil {
		return bc, err
	}
	return bc, nil
}

// PreHash returns the PreHash of the object
func (b *DSLinker) PreHash() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSPreImage.PreHash()
}

// IssuedAt returns the IssuedAt of the object
func (b *DSLinker) IssuedAt() (uint32, error) {
	if b == nil || b.DSPreImage == nil || b.DSPreImage.IssuedAt == 0 {
		return 0, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSPreImage.IssuedAt, nil
}

// ChainID returns the ChainID of the object
func (b *DSLinker) ChainID() (uint32, error) {
	if b == nil || b.DSPreImage == nil || b.DSPreImage.ChainID == 0 {
		return 0, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSPreImage.ChainID, nil
}

// Index returns the Index of the object
func (b *DSLinker) Index() ([]byte, error) {
	if b == nil || b.DSPreImage == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	if len(b.DSPreImage.Index) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("Invalid Index")
	}
	return utils.CopySlice(b.DSPreImage.Index), nil
}

// Owner returns the Owner field of the sub object
func (b *DSLinker) Owner() (*DataStoreOwner, error) {
	if b == nil || b.DSPreImage == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	if err := b.DSPreImage.Owner.Validate(); err != nil {
		return nil, errorz.ErrInvalid{}.New("DataStoreOwner invalid")
	}
	return b.DSPreImage.Owner, nil
}

// RawData returns the RawData field of the sub object
func (b *DSLinker) RawData() ([]byte, error) {
	if b == nil || b.DSPreImage == nil || len(b.DSPreImage.RawData) == 0 {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	return utils.CopySlice(b.DSPreImage.RawData), nil
}

// UTXOID returns the UTXOID of the object
func (b *DSLinker) UTXOID() ([]byte, error) {
	if b == nil || b.DSPreImage == nil || len(b.TxHash) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	return MakeUTXOID(b.TxHash, b.DSPreImage.TXOutIdx), nil
}

// TXOutIdx returns the TXOutIdx of the object
func (b *DSLinker) TXOutIdx() (uint32, error) {
	if b == nil || b.DSPreImage == nil {
		return 0, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSPreImage.TXOutIdx, nil
}

// SetTXOutIdx sets the TXOutIdx of the object
func (b *DSLinker) SetTXOutIdx(idx uint32) error {
	if b == nil || b.DSPreImage == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	b.DSPreImage.TXOutIdx = idx
	return nil
}

// EpochOfExpiration returns the epoch in which the datastore may be garbage
// collected
func (b *DSLinker) EpochOfExpiration() (uint32, error) {
	if b == nil {
		return 0, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSPreImage.EpochOfExpiration()
}

// RemainingValue returns remaining value at the time of consumption
func (b *DSLinker) RemainingValue(currentHeight uint32) (*uint256.Uint256, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSPreImage.RemainingValue(currentHeight)
}

// Value returns the value stored in the object at the time of creation
func (b *DSLinker) Value() (*uint256.Uint256, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSPreImage.Value()
}

// ValidateSignature validates the signature of the datastore at the time of
// consumption
func (b *DSLinker) ValidateSignature(currentHeight uint32, msg []byte, sig *DataStoreSignature) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSPreImage.ValidateSignature(currentHeight, msg, sig)
}

// ValidatePreSignature validates the signature of the datastore at the time of
// creation
func (b *DSLinker) ValidatePreSignature(msg []byte, sig *DataStoreSignature) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSPreImage.ValidatePreSignature(msg, sig)
}
