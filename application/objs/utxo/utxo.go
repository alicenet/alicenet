package utxo

import (
	capnp "github.com/MadBase/go-capnproto2/v2"

	mdefs "github.com/alicenet/alicenet/application/objs/capn"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
)

// Marshal will marshal the TXOut object.
func Marshal(v mdefs.TXOut) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	out := utils.CopySlice(raw)
	return out, nil
}

// Unmarshal will unmarshal the TXOut object.
func Unmarshal(data []byte) (mdefs.TXOut, error) {
	var err error
	fn := func() (mdefs.TXOut, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		dataCopy := utils.CopySlice(data)
		msg := &capnp.Message{Arena: capnp.SingleSegment(dataCopy)}
		obj, tmp := mdefs.ReadRootTXOut(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.TXOut{}, err
	}
	return obj, nil
}
