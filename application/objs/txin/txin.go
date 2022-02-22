package txin

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the TXIn object.
func Marshal(v mdefs.TXIn) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	out := utils.CopySlice(raw)
	return out, nil
}

// Unmarshal will unmarshal the TXIn object.
func Unmarshal(data []byte) (mdefs.TXIn, error) {
	var err error
	fn := func() (mdefs.TXIn, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		dataCopy := utils.CopySlice(data)
		msg := &capnp.Message{Arena: capnp.SingleSegment(dataCopy)}
		obj, tmp := mdefs.ReadRootTXIn(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.TXIn{}, err
	}
	return obj, nil
}

// Validate will validate the TXIn object
func Validate(v mdefs.TXIn) error {
	if !v.HasTXInLinker() {
		return errorz.ErrInvalid{}.New("txin capn obj does not have TXInLinker")
	}
	if !v.HasSignature() {
		return errorz.ErrInvalid{}.New("txin capn obj does not have Signature")
	}
	if len(v.Signature()) == 0 {
		return errorz.ErrInvalid{}.New("txin capn obj is not valid; invalid Signature; zero byte length")
	}
	return nil
}
