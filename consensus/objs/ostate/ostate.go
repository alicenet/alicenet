package ostate

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the OwnState object.
func Marshal(v mdefs.OwnState) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	out := utils.CopySlice(raw)
	return out, nil
}

// Unmarshal will unmarshal the OwnState object.
func Unmarshal(data []byte) (mdefs.OwnState, error) {
	var err error
	fn := func() (mdefs.OwnState, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootOwnState(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.OwnState{}, err
	}
	return obj, nil
}

// Validate will validate the OwnState object
func Validate(p mdefs.OwnState) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("ownstate capn obj is not valid")
	}
	if !p.HasSyncToBH() {
		return errorz.ErrInvalid{}.New("ownstate capn obj does not have SyncToBH")
	}
	if !p.HasMaxBHSeen() {
		return errorz.ErrInvalid{}.New("ownstate capn obj does not have MaxBHSeen")
	}
	if !p.HasPendingSnapShot() {
		return errorz.ErrInvalid{}.New("ownstate capn obj does not have PendingSnapShot")
	}
	if !p.HasCanonicalSnapShot() {
		return errorz.ErrInvalid{}.New("ownstate capn obj does not have CanonicalSnapShot")
	}
	return nil
}
