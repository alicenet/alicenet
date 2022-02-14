package prevotenil

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/errorz"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the PreVoteNil object.
func Marshal(v mdefs.PreVoteNil) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	return raw, nil
}

// Unmarshal will unmarshal the PreVoteNil object.
func Unmarshal(data []byte) (mdefs.PreVoteNil, error) {
	var err error
	fn := func() (mdefs.PreVoteNil, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootPreVoteNil(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.PreVoteNil{}, err
	}
	return obj, nil
}

// Validate will validate the PreVoteNil object
func Validate(p mdefs.PreVoteNil) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("pvn capn obj is not valid")
	}
	if !p.HasRCert() {
		return errorz.ErrInvalid{}.New("pvn capn obj does not have RCert")
	}
	if !p.HasSignature() {
		return errorz.ErrInvalid{}.New("pvn capn obj does not have Signature")
	}
	return nil
}
