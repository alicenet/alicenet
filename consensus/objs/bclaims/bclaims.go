package bclaims

import (
	capnp "github.com/MadBase/go-capnproto2/v2"
	mdefs "github.com/alicenet/alicenet/consensus/objs/capn"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
)

// Marshal will marshal the BClaims object.
func Marshal(v mdefs.BClaims) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	return raw, nil
}

// Unmarshal will unmarshal the BClaims object.
func Unmarshal(data []byte) (mdefs.BClaims, error) {
	var err error
	fn := func() (mdefs.BClaims, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootBClaims(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.BClaims{}, err
	}
	return obj, nil
}

// Validate will validate the BClaims object
func Validate(p mdefs.BClaims) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("bclaims capn obj is not valid")
	}
	if !p.HasStateRoot() {
		return errorz.ErrInvalid{}.New("bclaims capn obj does not have StateRoot")
	}
	if err := utils.ValidateHash(p.StateRoot()); err != nil {
		return err
	}
	if !p.HasPrevBlock() {
		return errorz.ErrInvalid{}.New("bclaims capn obj does not have PrevBlock")
	}
	if err := utils.ValidateHash(p.PrevBlock()); err != nil {
		return err
	}
	if !p.HasTxRoot() {
		return errorz.ErrInvalid{}.New("bclaims capn obj does not have TxRoot")
	}
	if err := utils.ValidateHash(p.TxRoot()); err != nil {
		return err
	}
	if !p.HasHeaderRoot() {
		return errorz.ErrInvalid{}.New("bclaims capn obj does not have HeaderRoot")
	}
	if err := utils.ValidateHash(p.HeaderRoot()); err != nil {
		return err
	}
	return nil
}
