package objs

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/rclaims"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// RClaims ...
type RClaims struct {
	ChainID   uint32
	Height    uint32
	Round     uint32
	PrevBlock []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// RClaims object
func (b *RClaims) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("RClaims.UnmarshalBinary; rclaims not initialized")
	}
	bh, err := rclaims.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *RClaims) UnmarshalCapn(bh mdefs.RClaims) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("RClaims.UnmarshalCapn; rclaims not initialized")
	}
	err := rclaims.Validate(bh)
	if err != nil {
		return err
	}
	b.ChainID = bh.ChainID()
	b.Height = bh.Height()
	b.Round = bh.Round()
	b.PrevBlock = utils.CopySlice(bh.PrevBlock())
	if b.Height < 1 {
		return errorz.ErrInvalid{}.New("RClaims.UnmarshalCapn; height is zero")
	}
	if b.Round < 1 {
		return errorz.ErrInvalid{}.New("RClaims.UnmarshalCapn; round is zero")
	}
	if b.ChainID < 1 {
		return errorz.ErrInvalid{}.New("RClaims.UnmarshalCapn; chainID is zero")
	}
	if b.Round > constants.DEADBLOCKROUND {
		return errorz.ErrInvalid{}.New("RClaims.UnmarshalCapn; round > DBR")
	}
	return nil
}

// MarshalBinary takes the RClaims object and returns the canonical
// byte slice
func (b *RClaims) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("RClaims.MarshalBinary; rclaims not initialized")
	}
	if b.ChainID == 0 {
		return nil, errorz.ErrInvalid{}.New("RClaims.MarshalBinary; chainID is zero")
	}
	if b.Round == 0 {
		return nil, errorz.ErrInvalid{}.New("RClaims.MarshalBinary; round is zero")
	}
	if b.Round > constants.DEADBLOCKROUND {
		return nil, errorz.ErrInvalid{}.New("RClaims.MarshalBinary; round > DBR")
	}
	if b.Height == 0 {
		return nil, errorz.ErrInvalid{}.New("RClaims.MarshalBinary; height is zero")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return rclaims.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *RClaims) MarshalCapn(seg *capnp.Segment) (mdefs.RClaims, error) {
	if b == nil {
		return mdefs.RClaims{}, errorz.ErrInvalid{}.New("RClaims.MarshalCapn; rclaims not initialized")
	}
	var bh mdefs.RClaims
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootRClaims(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewRClaims(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	}
	bh.SetChainID(b.ChainID)
	bh.SetHeight(b.Height)
	bh.SetRound(b.Round)
	err := bh.SetPrevBlock(b.PrevBlock[:])
	if err != nil {
		return mdefs.RClaims{}, err
	}
	return bh, nil
}
