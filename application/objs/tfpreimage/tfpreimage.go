package tfpreimage

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "zombiezen.com/go/capnproto2"
)

// Marshal will marshal the TFPreImage object.
func Marshal(v mdefs.TFPreImage) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	out := utils.CopySlice(raw)
	return out, nil
}

// Unmarshal will unmarshal the TFPreImage object.
func Unmarshal(data []byte) (mdefs.TFPreImage, error) {
	var err error
	fn := func() (mdefs.TFPreImage, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		dataCopy := utils.CopySlice(data)
		msg := &capnp.Message{Arena: capnp.SingleSegment(dataCopy)}
		obj, tmp := mdefs.ReadRootTFPreImage(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.TFPreImage{}, err
	}
	return obj, nil
}

// Validate will validate the TFPreImage object
func Validate(v mdefs.TFPreImage) error {
	if v.ChainID() < 1 {
		return errorz.ErrInvalid{}.New("tfpreimage capn obj is not valid; invalid ChainID")
	}
	if int(v.TXOutIdx()) >= constants.MaxTxVectorLength {
		return errorz.ErrInvalid{}.New("tfpreimage capn obj is not valid: output index is too large")
	}
	return nil
}
