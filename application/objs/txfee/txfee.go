package txfee

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "zombiezen.com/go/capnproto2"
)

// Marshal will marshal the TxFee object.
func Marshal(v mdefs.TxFee) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	out := utils.CopySlice(raw)
	return out, nil
}

// Unmarshal will unmarshal the TxFee object.
func Unmarshal(data []byte) (mdefs.TxFee, error) {
	var err error
	fn := func() (mdefs.TxFee, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		dataCopy := utils.CopySlice(data)
		msg := &capnp.Message{Arena: capnp.SingleSegment(dataCopy)}
		obj, tmp := mdefs.ReadRootTxFee(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.TxFee{}, err
	}
	return obj, nil
}

// Validate will validate the TxFee object
func Validate(v mdefs.TxFee) error {
	if !v.HasTFPreImage() {
		return errorz.ErrInvalid{}.New("txfee capn obj does not have TFPreImage")
	}
	if !v.HasTxHash() {
		return errorz.ErrInvalid{}.New("txfee capn obj does not have TxHash")
	}
	if len(v.TxHash()) != constants.HashLen {
		return errorz.ErrInvalid{}.New("txfee capn obj is not valid: invalid TxHash; incorrect byte length")
	}
	return nil
}
