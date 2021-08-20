package objs

import (
	"github.com/MadBase/MadNet/consensus/objs/bclaims"
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	capnp "zombiezen.com/go/capnproto2"
)

// BClaims ...
type BClaims struct {
	ChainID    uint32
	Height     uint32
	TxCount    uint32
	PrevBlock  []byte
	TxRoot     []byte
	StateRoot  []byte
	HeaderRoot []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// BClaims object
func (b *BClaims) UnmarshalBinary(data []byte) error {
	bc, err := bclaims.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bc.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bc)
}

// MarshalBinary takes the BClaims object and returns the canonical
// byte slice
func (b *BClaims) MarshalBinary() ([]byte, error) {
	if b == nil || b.ChainID == 0 || b.Height == 0 {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	bc, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bc.Struct.Segment().Message().Reset(nil)
	return bclaims.Marshal(bc)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *BClaims) UnmarshalCapn(bc mdefs.BClaims) error {
	err := bclaims.Validate(bc)
	if err != nil {
		return err
	}
	b.ChainID = bc.ChainID()
	b.Height = bc.Height()
	b.TxCount = bc.TxCount()
	b.PrevBlock = bc.PrevBlock()
	b.HeaderRoot = bc.HeaderRoot()
	b.StateRoot = bc.StateRoot()
	b.TxRoot = bc.TxRoot()
	if len(b.PrevBlock) != constants.HashLen {
		return errorz.ErrInvalid{}.New("capn bclaims prevblock bad len")
	}
	if b.Height < 1 {
		return errorz.ErrInvalid{}.New("capn bclaims bad height")
	}
	if b.ChainID < 1 {
		return errorz.ErrInvalid{}.New("capn bclaims bad cid")
	}
	return nil
}

// MarshalCapn marshals the object into its capnproto definition
func (b *BClaims) MarshalCapn(seg *capnp.Segment) (mdefs.BClaims, error) {
	if b == nil {
		return mdefs.BClaims{}, errorz.ErrInvalid{}.New("not initialized")
	}
	var bc mdefs.BClaims
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bc, err
		}
		tmp, err := mdefs.NewRootBClaims(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	} else {
		tmp, err := mdefs.NewBClaims(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	}
	bc.SetChainID(b.ChainID)
	bc.SetHeight(b.Height)
	bc.SetTxCount(b.TxCount)
	err := bc.SetPrevBlock(b.PrevBlock[:])
	if err != nil {
		return bc, err
	}
	err = bc.SetHeaderRoot(b.HeaderRoot[:])
	if err != nil {
		return bc, err
	}
	err = bc.SetStateRoot(b.StateRoot[:])
	if err != nil {
		return bc, err
	}
	err = bc.SetTxRoot(b.TxRoot[:])
	if err != nil {
		return bc, err
	}
	return bc, nil
}

// BlockHash returns the BlockHash of BClaims
func (b *BClaims) BlockHash() ([]byte, error) {
	can, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return crypto.Hasher(can), nil
}
