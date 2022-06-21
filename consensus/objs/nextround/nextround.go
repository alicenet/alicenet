package nextround

import (
	capnp "github.com/MadBase/go-capnproto2/v2"
	mdefs "github.com/alicenet/alicenet/consensus/objs/capn"
	"github.com/alicenet/alicenet/errorz"
)

// Marshal will marshal the NextRound object.
func Marshal(v mdefs.NextRound) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	return raw, nil
}

// Unmarshal will unmarshal the NextRound object.
func Unmarshal(data []byte) (mdefs.NextRound, error) {
	var err error
	fn := func() (mdefs.NextRound, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootNextRound(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.NextRound{}, err
	}
	return obj, nil
}

// Validate will validate the NextRound object
func Validate(p mdefs.NextRound) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("nextround capn obj is not valid")
	}
	if !p.HasNRClaims() {
		return errorz.ErrInvalid{}.New("nextround capn obj does not have NRClaims")
	}
	if !p.HasSignature() {
		return errorz.ErrInvalid{}.New("nextround capn obj does not have Signature")
	}
	return nil
}
