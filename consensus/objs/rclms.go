package objs

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/rclaims"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	gUtils "github.com/MadBase/MadNet/utils"
	capnp "zombiezen.com/go/capnproto2"
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
	bh, err := rclaims.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *RClaims) UnmarshalCapn(bh mdefs.RClaims) error {
	err := rclaims.Validate(bh)
	if err != nil {
		return err
	}
	b.ChainID = bh.ChainID()
	b.Height = bh.Height()
	b.Round = bh.Round()
	b.PrevBlock = gUtils.CopySlice(bh.PrevBlock())
	if len(b.PrevBlock) != constants.HashLen {
		return errorz.ErrInvalid{}.New("rclaims bad prevb len")
	}
	if b.Height < 1 {
		return errorz.ErrInvalid{}.New("rclaims bad height")
	}
	if b.Round < 1 {
		return errorz.ErrInvalid{}.New("rclaims bad round")
	}
	if b.ChainID < 1 {
		return errorz.ErrInvalid{}.New("rclaims bad cid")
	}
	if b.Round > constants.DEADBLOCKROUND {
		return errorz.ErrInvalid{}.New("rclaims round too big")
	}
	return nil
}

// MarshalBinary takes the RClaims object and returns the canonical
// byte slice
func (b *RClaims) MarshalBinary() ([]byte, error) {
	if b == nil || b.ChainID == 0 || b.Round == 0 || b.Height == 0 {
		return nil, errorz.ErrInvalid{}.New("not initialized")
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
		return mdefs.RClaims{}, errorz.ErrInvalid{}.New("not initialized")
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
