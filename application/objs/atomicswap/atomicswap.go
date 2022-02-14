package atomicswap

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the AtomicSwap object.
func Marshal(v mdefs.AtomicSwap) ([]byte, error) {
	return nil, errorz.ErrInvalid{}.New("AtomicSwap not activated")
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	out := utils.CopySlice(raw)
	return out, nil
}

// Unmarshal will unmarshal the AtomicSwap object.
func Unmarshal(data []byte) (mdefs.AtomicSwap, error) {
	return mdefs.AtomicSwap{}, errorz.ErrInvalid{}.New("AtomicSwap not activated")
	var err error
	fn := func() (mdefs.AtomicSwap, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		dataCopy := utils.CopySlice(data)
		msg := &capnp.Message{Arena: capnp.SingleSegment(dataCopy)}
		obj, tmp := mdefs.ReadRootAtomicSwap(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.AtomicSwap{}, err
	}
	return obj, nil
}

// Validate will validate the AtomicSwap object
func Validate(v mdefs.AtomicSwap) error {
	return errorz.ErrInvalid{}.New("AtomicSwap not activated")
	if !v.HasASPreImage() {
		return errorz.ErrInvalid{}.New("atomicswap capn obj does not have ASPreImage")
	}
	if !v.HasTxHash() {
		return errorz.ErrInvalid{}.New("atomicswap capn obj does not have TxHash")
	}
	if len(v.TxHash()) != constants.HashLen {
		return errorz.ErrInvalid{}.New("atomicswap capn obj is not valid: invalid TxHash; incorrect byte length")
	}
	return nil
}
