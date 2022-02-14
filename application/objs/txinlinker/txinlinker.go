package txinlinker

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the TXInLinker object.
func Marshal(v mdefs.TXInLinker) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	out := utils.CopySlice(raw)
	return out, nil
}

// Unmarshal will unmarshal the TXInLinker object.
func Unmarshal(data []byte) (mdefs.TXInLinker, error) {
	var err error
	fn := func() (mdefs.TXInLinker, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		dataCopy := utils.CopySlice(data)
		msg := &capnp.Message{Arena: capnp.SingleSegment(dataCopy)}
		obj, tmp := mdefs.ReadRootTXInLinker(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.TXInLinker{}, err
	}
	return obj, nil
}

// Validate will validate the TXInLinker object
func Validate(v mdefs.TXInLinker) error {
	if !v.HasTXInPreImage() {
		return errorz.ErrInvalid{}.New("txinlinker capn obj does not have TXInPreImage")
	}
	if !v.HasTxHash() {
		return errorz.ErrInvalid{}.New("txinlinker capn obj does not have TxHash;")
	}
	if len(v.TxHash()) != constants.HashLen {
		return errorz.ErrInvalid{}.New("txinlinker capn obj is not valid: invalid TxHash; incorrect byte length")
	}
	return nil
}
