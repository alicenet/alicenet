package proposal

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/errorz"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the Proposal object.
func Marshal(v mdefs.Proposal) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	return raw, nil
}

// Unmarshal will unmarshal the Proposal object.
func Unmarshal(data []byte) (mdefs.Proposal, error) {
	var err error
	fn := func() (mdefs.Proposal, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootProposal(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.Proposal{}, err
	}
	return obj, nil
}

// Validate will validate the Proposal object
func Validate(p mdefs.Proposal) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("proposal capn obj is not valid")
	}
	if !p.HasPClaims() {
		return errorz.ErrInvalid{}.New("proposal capn obj does not have PClaims")
	}
	if !p.HasSignature() {
		return errorz.ErrInvalid{}.New("proposal capn obj does not have Signature")
	}
	return nil
}
