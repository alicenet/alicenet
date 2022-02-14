package vspreimage

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the VSPreImage object.
func Marshal(v mdefs.VSPreImage) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	out := utils.CopySlice(raw)
	return out, nil
}

// Unmarshal will unmarshal the VSPreImage object.
func Unmarshal(data []byte) (mdefs.VSPreImage, error) {
	var err error
	fn := func() (mdefs.VSPreImage, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		dataCopy := utils.CopySlice(data)
		msg := &capnp.Message{Arena: capnp.SingleSegment(dataCopy)}
		obj, tmp := mdefs.ReadRootVSPreImage(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.VSPreImage{}, err
	}
	return obj, nil
}

// Validate will validate the VSPreImage object
func Validate(v mdefs.VSPreImage) error {
	if v.ChainID() < 1 {
		return errorz.ErrInvalid{}.New("vspreimage capn obj is not valid; invalid ChainID")
	}
	if !v.HasOwner() {
		return errorz.ErrInvalid{}.New("vspreimage capn obj does not have Owner")
	}
	if len(v.Owner()) == 0 {
		return errorz.ErrInvalid{}.New("vspreimage capn obj is not valid: invalid Owner; zero byte length")
	}
	if v.Value()|v.Value1()|v.Value2()|v.Value3()|v.Value4()|v.Value5()|v.Value6()|v.Value7() == 0 {
		return errorz.ErrInvalid{}.New("vspreimage capn obj is not valid; no value")
	}
	if v.TXOutIdx() != constants.MaxUint32 {
		if int(v.TXOutIdx()) >= constants.MaxTxVectorLength {
			return errorz.ErrInvalid{}.New("vspreimage capn obj is not valid: output index is too large")
		}
	}
	return nil
}
