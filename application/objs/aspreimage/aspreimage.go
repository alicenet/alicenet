package aspreimage

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the ASPreImage object.
func Marshal(v mdefs.ASPreImage) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	out := utils.CopySlice(raw)
	return out, nil
}

// Unmarshal will unmarshal the ASPreImage object.
func Unmarshal(data []byte) (mdefs.ASPreImage, error) {
	var err error
	fn := func() (mdefs.ASPreImage, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		dataCopy := utils.CopySlice(data)
		msg := &capnp.Message{Arena: capnp.SingleSegment(dataCopy)}
		obj, tmp := mdefs.ReadRootASPreImage(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.ASPreImage{}, err
	}
	return obj, nil
}

// Validate will validate the ASPreImage object
func Validate(v mdefs.ASPreImage) error {
	if v.ChainID() < 1 {
		return errorz.ErrInvalid{}.New("aspreimage capn obj is not valid: invalid ChainID")
	}
	if !v.HasOwner() {
		return errorz.ErrInvalid{}.New("aspreimage capn obj does not have Owner")
	}
	if len(v.Owner()) == 0 {
		return errorz.ErrInvalid{}.New("aspreimage capn obj is not valid: invalid Owner; zero byte length")
	}
	if v.IssuedAt() < 1 {
		return errorz.ErrInvalid{}.New("aspreimage capn obj is not valid: invalid IssuedAt")
	}
	if v.Exp() < 1 {
		return errorz.ErrInvalid{}.New("aspreimage capn obj is not valid: invalid Exp")
	}
	if v.Exp() <= v.IssuedAt() {
		return errorz.ErrInvalid{}.New("aspreimage capn obj is not valid: Exp <= IssuedAt")
	}
	if int(v.TXOutIdx()) >= constants.MaxTxVectorLength {
		return errorz.ErrInvalid{}.New("aspreimage capn obj is not valid: output index is too large")
	}
	return nil
}
