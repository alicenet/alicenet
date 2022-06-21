package ovstate

import (
	capnp "github.com/MadBase/go-capnproto2/v2"
	mdefs "github.com/alicenet/alicenet/consensus/objs/capn"
	"github.com/alicenet/alicenet/errorz"
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
	return nil
}
