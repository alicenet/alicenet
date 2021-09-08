package objs

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/application/objs/tfpreimage"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "zombiezen.com/go/capnproto2"
)

// TFPreImage is a txfee preimage
type TFPreImage struct {
	ChainID  uint32
	TXOutIdx uint32
	Fee      *uint256.Uint256
	//
	preHash []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// TFPreImage object
func (b *TFPreImage) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	bc, err := tfpreimage.Unmarshal(data)
	if err != nil {
		return err
	}
	return b.UnmarshalCapn(bc)
}

// MarshalBinary takes the TFPreImage object and returns the canonical
// byte slice
func (b *TFPreImage) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	bc, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	return tfpreimage.Marshal(bc)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *TFPreImage) UnmarshalCapn(bc mdefs.TFPreImage) error {
	if err := tfpreimage.Validate(bc); err != nil {
		return err
	}
	b.ChainID = bc.ChainID()
	u32array := [8]uint32{}
	u32array[0] = bc.Fee0()
	u32array[1] = bc.Fee1()
	u32array[2] = bc.Fee2()
	u32array[3] = bc.Fee3()
	u32array[4] = bc.Fee4()
	u32array[5] = bc.Fee5()
	u32array[6] = bc.Fee6()
	u32array[7] = bc.Fee7()
	fObj := &uint256.Uint256{}
	err := fObj.FromUint32Array(u32array)
	if err != nil {
		return err
	}
	b.Fee = fObj
	b.TXOutIdx = bc.TXOutIdx()
	return nil
}

// MarshalCapn marshals the object into its capnproto definition
func (b *TFPreImage) MarshalCapn(seg *capnp.Segment) (mdefs.TFPreImage, error) {
	if b == nil {
		return mdefs.TFPreImage{}, errorz.ErrInvalid{}.New("not initialized")
	}
	var bc mdefs.TFPreImage
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bc, err
		}
		tmp, err := mdefs.NewRootTFPreImage(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	} else {
		tmp, err := mdefs.NewTFPreImage(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	}
	bc.SetChainID(b.ChainID)
	u32array, err := b.Fee.ToUint32Array()
	if err != nil {
		return bc, err
	}
	bc.SetFee0(u32array[0])
	bc.SetFee1(u32array[1])
	bc.SetFee2(u32array[2])
	bc.SetFee3(u32array[3])
	bc.SetFee4(u32array[4])
	bc.SetFee5(u32array[5])
	bc.SetFee6(u32array[6])
	bc.SetFee7(u32array[7])
	bc.SetTXOutIdx(b.TXOutIdx)
	return bc, nil
}

// PreHash calculates the PreHash of the object
func (b *TFPreImage) PreHash() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	if b.preHash != nil {
		return utils.CopySlice(b.preHash), nil
	}
	msg, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	hsh := crypto.Hasher(msg)
	b.preHash = hsh
	return utils.CopySlice(b.preHash), nil
}
