package validator

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/errorz"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the Validator object.
func Marshal(v mdefs.Validator) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	return raw, nil
}

// Unmarshal will unmarshal the Validator object.
func Unmarshal(data []byte) (mdefs.Validator, error) {
	var err error
	fn := func() (mdefs.Validator, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootValidator(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.Validator{}, err
	}
	return obj, nil
}

// Validate will validate the Validator object
func Validate(p mdefs.Validator) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("validator capn obj is not valid")
	}
	if !p.HasGroupShare() {
		return errorz.ErrInvalid{}.New("validator capn obj does not have GroupShare")
	}
	if !p.HasVAddr() {
		return errorz.ErrInvalid{}.New("validator capn obj does not have VAddr")
	}
	return nil
}
