package datastore

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the DataStore object.
func Marshal(v mdefs.DataStore) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	out := utils.CopySlice(raw)
	return out, nil
}

// Unmarshal will unmarshal the DataStore object.
func Unmarshal(data []byte) (mdefs.DataStore, error) {
	var err error
	fn := func() (mdefs.DataStore, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		dataCopy := utils.CopySlice(data)
		msg := &capnp.Message{Arena: capnp.SingleSegment(dataCopy)}
		obj, tmp := mdefs.ReadRootDataStore(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.DataStore{}, err
	}
	return obj, nil
}

// Validate will validate the DataStore object
func Validate(v mdefs.DataStore) error {
	if !v.HasDSLinker() {
		return errorz.ErrInvalid{}.New("datastore capn obj does not have DSLinker")
	}
	if !v.HasSignature() {
		return errorz.ErrInvalid{}.New("datastore capn obj does not have Signature")
	}
	if len(v.Signature()) == 0 {
		return errorz.ErrInvalid{}.New("datastore capn obj is not valid; invalid Signature; zero byte length")
	}
	return nil
}
