package txinpreimage

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the TXInPreImage object.
func Marshal(v mdefs.TXInPreImage) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	out := utils.CopySlice(raw)
	return out, nil
}

// Unmarshal will unmarshal the TXInPreImage object.
func Unmarshal(data []byte) (mdefs.TXInPreImage, error) {
	var err error
	fn := func() (mdefs.TXInPreImage, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		dataCopy := utils.CopySlice(data)
		msg := &capnp.Message{Arena: capnp.SingleSegment(dataCopy)}
		obj, tmp := mdefs.ReadRootTXInPreImage(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.TXInPreImage{}, err
	}
	return obj, nil
}

// Validate will validate the TXInPreImage object
func Validate(v mdefs.TXInPreImage) error {
	if v.ChainID() < 1 {
		return errorz.ErrInvalid{}.New("txinpreimage capn obj is not valid: invalid ChainID")
	}
	if !v.HasConsumedTxHash() {
		return errorz.ErrInvalid{}.New("txinpreimage capn obj does not have ConsumedTxHash")
	}
	if len(v.ConsumedTxHash()) != constants.HashLen {
		return errorz.ErrInvalid{}.New("txinpreimage capn obj is not valid: invalid ConsumedTxHash; incorrect byte length")
	}
	return nil
}
