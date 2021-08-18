package ovstate

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/errorz"
	capnp "zombiezen.com/go/capnproto2"
)

// Marshal will marshal the OwnValidatingState object.
func Marshal(v mdefs.OwnValidatingState) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	return raw, nil
}

// Unmarshal will unmarshal OwnValidatingState.
func Unmarshal(data []byte) (mdefs.OwnValidatingState, error) {
	var err error
	fn := func() (mdefs.OwnValidatingState, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootOwnValidatingState(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.OwnValidatingState{}, err
	}
	return obj, nil
}

// Validate will validate the OwnValidatingState object
func Validate(p mdefs.OwnValidatingState) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("ovs capn obj is not valid")
	}
	if !p.HasVAddr() {
		return errorz.ErrInvalid{}.New("ovs capn obj does not have VAddr")
	}
	if !p.HasGroupKey() {
		return errorz.ErrInvalid{}.New("ovs capn obj does not have GroupKey")
	}
	return nil
}
