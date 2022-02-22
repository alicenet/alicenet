package pclaims

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/errorz"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the PClaims object.
func Marshal(v mdefs.PClaims) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	return raw, nil
}

// Unmarshal will unmarshal the PClaims object.
func Unmarshal(data []byte) (mdefs.PClaims, error) {
	var err error
	fn := func() (mdefs.PClaims, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootPClaims(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.PClaims{}, err
	}
	return obj, nil
}

// Validate will validate the PClaims object
func Validate(p mdefs.PClaims) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("pclaims capn obj is not valid")
	}
	if !p.HasBClaims() {
		return errorz.ErrInvalid{}.New("pclaims capn obj does not have BClaims")
	}
	if !p.HasRCert() {
		return errorz.ErrInvalid{}.New("pclaims capn obj does not have RCert")
	}
	return nil
}
