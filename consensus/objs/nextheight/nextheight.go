package nextheight

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/errorz"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the NextHeight object.
func Marshal(v mdefs.NextHeight) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	return raw, nil
}

// Unmarshal will unmarshal the NextHeight object.
func Unmarshal(data []byte) (mdefs.NextHeight, error) {
	var err error
	fn := func() (mdefs.NextHeight, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootNextHeight(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.NextHeight{}, err
	}
	return obj, nil
}

// Validate will validate the NextHeight object
func Validate(p mdefs.NextHeight) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("nextheight capn obj is not valid")
	}
	if !p.HasNHClaims() {
		return errorz.ErrInvalid{}.New("nextheight capn obj does not have NHClaims")
	}
	if !p.HasPreCommits() {
		return errorz.ErrInvalid{}.New("nextheight capn obj does not have PreCommits")
	}
	if !p.HasSignature() {
		return errorz.ErrInvalid{}.New("nextheight capn obj does not have Signature")
	}
	return nil
}
