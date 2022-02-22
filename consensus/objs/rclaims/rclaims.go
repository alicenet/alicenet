package rclaims

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the RClaims object.
func Marshal(v mdefs.RClaims) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	return raw, nil
}

// Unmarshal will unmarshal the RClaims object.
func Unmarshal(data []byte) (mdefs.RClaims, error) {
	var err error
	fn := func() (mdefs.RClaims, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootRClaims(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.RClaims{}, err
	}
	return obj, nil
}

// Validate will validate the RClaims object
func Validate(p mdefs.RClaims) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("rclaims capn obj is not valid")
	}
	if !p.HasPrevBlock() {
		return errorz.ErrInvalid{}.New("rclaims capn obj does not have PrevBlock")
	}
	if err := utils.ValidateHash(p.PrevBlock()); err != nil {
		return err
	}
	return nil
}
