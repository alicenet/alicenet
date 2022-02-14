package vset

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/errorz"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the ValidatorSet object.
func Marshal(v mdefs.ValidatorSet) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	return raw, nil
}

// Unmarshal will unmarshal the ValidatorSet object.
func Unmarshal(data []byte) (mdefs.ValidatorSet, error) {
	var err error
	fn := func() (mdefs.ValidatorSet, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootValidatorSet(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.ValidatorSet{}, err
	}
	return obj, nil
}

// Validate will validate the ValidatorSet object
func Validate(p mdefs.ValidatorSet) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("validatorset capn obj is not valid")
	}
	if !p.HasGroupKey() {
		return errorz.ErrInvalid{}.New("validatorset capn obj does not have GroupKey")
	}
	if !p.HasValidators() {
		return errorz.ErrInvalid{}.New("validatorset capn obj does not have Validators")
	}
	return nil
}
