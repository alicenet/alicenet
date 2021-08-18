package objs

import (
	"bytes"

	"github.com/MadBase/MadNet/errorz"

	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/pclaims"
	capnp "zombiezen.com/go/capnproto2"
)

// PClaims ...
type PClaims struct {
	BClaims *BClaims
	RCert   *RCert
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// PClaims object
func (b *PClaims) UnmarshalBinary(data []byte) error {
	bh, err := pclaims.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *PClaims) UnmarshalCapn(bh mdefs.PClaims) error {
	b.RCert = &RCert{}
	b.BClaims = &BClaims{}
	err := pclaims.Validate(bh)
	if err != nil {
		return err
	}
	err = b.RCert.UnmarshalCapn(bh.RCert())
	if err != nil {
		return err
	}
	err = b.BClaims.UnmarshalCapn(bh.BClaims())
	if err != nil {
		return err
	}
	if b.RCert.RClaims.ChainID != b.BClaims.ChainID {
		return errorz.ErrInvalid{}.New("pclaims cid != cid")
	}
	if b.RCert.RClaims.Height != b.BClaims.Height {
		return errorz.ErrInvalid{}.New("pclaims h != h")
	}
	if !bytes.Equal(b.RCert.RClaims.PrevBlock[:], b.BClaims.PrevBlock[:]) {
		return errorz.ErrInvalid{}.New("pclaims prevb != prevb")
	}
	return nil
}

// MarshalBinary takes the PClaims object and returns the canonical
// byte slice
func (b *PClaims) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return pclaims.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *PClaims) MarshalCapn(seg *capnp.Segment) (mdefs.PClaims, error) {
	if b == nil {
		return mdefs.PClaims{}, errorz.ErrInvalid{}.New("not initialized")
	}
	var bh mdefs.PClaims
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootPClaims(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewPClaims(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	}
	bc, err := b.BClaims.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	err = bh.SetBClaims(bc)
	if err != nil {
		return mdefs.PClaims{}, err
	}
	rc, err := b.RCert.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	err = bh.SetRCert(rc)
	if err != nil {
		return mdefs.PClaims{}, err
	}
	return bh, nil
}
