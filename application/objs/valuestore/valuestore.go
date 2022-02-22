package valuestore

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the ValueStore object.
func Marshal(v mdefs.ValueStore) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	out := utils.CopySlice(raw)
	return out, nil
}

// Unmarshal will unmarshal the ValueStore object.
func Unmarshal(data []byte) (mdefs.ValueStore, error) {
	var err error
	fn := func() (mdefs.ValueStore, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		dataCopy := utils.CopySlice(data)
		msg := &capnp.Message{Arena: capnp.SingleSegment(dataCopy)}
		obj, tmp := mdefs.ReadRootValueStore(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.ValueStore{}, err
	}
	return obj, nil
}

// Validate will validate the ValueStore object
func Validate(v mdefs.ValueStore) error {
	if !v.HasVSPreImage() {
		return errorz.ErrInvalid{}.New("valuestore capn obj does not have VSPreImage")
	}
	if !v.HasTxHash() {
		return errorz.ErrInvalid{}.New("valuestore capn obj does not have TxHash")
	}
	if len(v.TxHash()) != constants.HashLen {
		return errorz.ErrInvalid{}.New("valuestore capn obj is not valid: invalid TxHash; incorrect byte length")
	}
	return nil
}
