package blockheader

import (
	capnp "github.com/MadBase/go-capnproto2/v2"
	mdefs "github.com/alicenet/alicenet/consensus/objs/capn"
	"github.com/alicenet/alicenet/errorz"
)

// Marshal will marshal the BlockHeader object.
func Marshal(v mdefs.BlockHeader) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	return raw, nil
}

// Unmarshal will unmarshal the BlockHeader object.
func Unmarshal(data []byte) (mdefs.BlockHeader, error) {
	var err error
	fn := func() (mdefs.BlockHeader, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootBlockHeader(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.BlockHeader{}, err
	}
	return obj, nil
}

// Validate will validate the BlockHeader object
func Validate(p mdefs.BlockHeader) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("blockheader capn obj is not valid")
	}
	if !p.HasSigGroup() {
		return errorz.ErrInvalid{}.New("blockheader capn obj does not have SigGroup")
	}
	if !p.HasTxHshLst() {
		return errorz.ErrInvalid{}.New("blockheader capn obj does not have TxHshLst")
	}
	if !p.HasBClaims() {
		return errorz.ErrInvalid{}.New("blockheader capn obj does not have BClaims")
	}
	return nil
}
