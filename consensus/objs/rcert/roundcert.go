package rcert

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/errorz"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the RCert object.
func Marshal(v mdefs.RCert) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	return raw, nil
}

// Unmarshal will unmarshal the RCert object.
func Unmarshal(data []byte) (mdefs.RCert, error) {
	var err error
	fn := func() (mdefs.RCert, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootRCert(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.RCert{}, err
	}
	return obj, nil
}

// Validate will validate the RCert object
func Validate(p mdefs.RCert) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("rcert capn obj is not valid")
	}
	if !p.HasRClaims() {
		return errorz.ErrInvalid{}.New("rcert capn obj does not have RClaims")
	}
	if !p.HasSigGroup() {
		return errorz.ErrInvalid{}.New("rcert capn obj does not have SigGroup")
	}
	return nil
}
