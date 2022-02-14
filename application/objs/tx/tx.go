package tx

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the Tx object.
func Marshal(v mdefs.Tx) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	out := utils.CopySlice(raw)
	return out, nil
}

// Unmarshal will unmarshal the Tx object.
func Unmarshal(data []byte) (mdefs.Tx, error) {
	var err error
	fn := func() (mdefs.Tx, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		dataCopy := utils.CopySlice(data)
		msg := &capnp.Message{Arena: capnp.SingleSegment(dataCopy)}
		obj, tmp := mdefs.ReadRootTx(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.Tx{}, err
	}
	return obj, nil
}

// Validate will validate the Tx object
func Validate(v mdefs.Tx) error {
	if !v.HasVin() {
		return errorz.ErrInvalid{}.New("tx capn obj does not have Vin")
	}
	vin, err := v.Vin()
	if err != nil {
		return err
	}
	if vin.Len() == 0 {
		return errorz.ErrInvalid{}.New("tx capn obj is not valid; invalid Vin; zero length object")
	}
	if vin.Len() > constants.MaxTxVectorLength {
		return errorz.ErrInvalid{}.New("tx capn obj is not valid; invalid Vin; length object too large")
	}
	if !v.HasVout() {
		return errorz.ErrInvalid{}.New("tx capn obj does not have Vout")
	}
	vout, err := v.Vout()
	if err != nil {
		return err
	}
	if vout.Len() == 0 {
		return errorz.ErrInvalid{}.New("tx capn obj is not valid; invalid Vout; zero length object")
	}
	if vout.Len() > constants.MaxTxVectorLength {
		return errorz.ErrInvalid{}.New("tx capn obj is not valid; invalid Vout; length object too large")
	}
	return nil
}
