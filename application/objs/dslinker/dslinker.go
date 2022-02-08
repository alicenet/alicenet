package dslinker

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the DSLinker object.
func Marshal(v mdefs.DSLinker) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	out := utils.CopySlice(raw)
	return out, nil
}

// Unmarshal will unmarshal the DSLinker object.
func Unmarshal(data []byte) (mdefs.DSLinker, error) {
	var err error
	fn := func() (mdefs.DSLinker, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		dataCopy := utils.CopySlice(data)
		msg := &capnp.Message{Arena: capnp.SingleSegment(dataCopy)}
		obj, tmp := mdefs.ReadRootDSLinker(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.DSLinker{}, err
	}
	return obj, nil
}

// Validate will validate the DSLinker object
func Validate(v mdefs.DSLinker) error {
	if !v.HasDSPreImage() {
		return errorz.ErrInvalid{}.New("dslinker capn obj does not have DSPreImage")
	}
	if !v.HasTxHash() {
		return errorz.ErrInvalid{}.New("dslinker capn obj does not have TxHash")
	}
	if len(v.TxHash()) != constants.HashLen {
		return errorz.ErrInvalid{}.New("dslinker capn obj is invalid: invalid TxHash; incorrect byte length")
	}
	return nil
}
