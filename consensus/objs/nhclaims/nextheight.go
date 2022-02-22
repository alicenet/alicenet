package nhclaims

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/errorz"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the NHClaims object.
func Marshal(v mdefs.NHClaims) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	return raw, nil
}

// Unmarshal will unmarshal the NHClaims object.
func Unmarshal(data []byte) (mdefs.NHClaims, error) {
	var err error
	fn := func() (mdefs.NHClaims, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootNHClaims(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.NHClaims{}, err
	}
	return obj, nil
}

// Validate will validate the NHClaimsobject
func Validate(p mdefs.NHClaims) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("nhclaims capn obj is not valid")
	}
	if !p.HasProposal() {
		return errorz.ErrInvalid{}.New("nhclaims capn obj does not have Proposal")
	}
	if !p.HasSigShare() {
		return errorz.ErrInvalid{}.New("nhclaims capn obj does not have SigShare")
	}
	return nil
}
