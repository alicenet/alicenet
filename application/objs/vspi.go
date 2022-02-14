package objs

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/application/objs/vspreimage"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// VSPreImage is a value store preimage
type VSPreImage struct {
	ChainID  uint32
	Value    *uint256.Uint256
	TXOutIdx uint32
	Owner    *ValueStoreOwner
	Fee      *uint256.Uint256
	//
	preHash []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// VSPreImage object
func (b *VSPreImage) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("vspi.unmarshalBinary: vspi not initialized")
	}
	bc, err := vspreimage.Unmarshal(data)
	if err != nil {
		return err
	}
	return b.UnmarshalCapn(bc)
}

// MarshalBinary takes the VSPreImage object and returns the canonical
// byte slice
func (b *VSPreImage) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("vspi.marshalBinary: vspi not initialized")
	}
	bc, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	return vspreimage.Marshal(bc)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *VSPreImage) UnmarshalCapn(bc mdefs.VSPreImage) error {
	if err := vspreimage.Validate(bc); err != nil {
		return err
	}
	b.ChainID = bc.ChainID()
	u32array := [8]uint32{}
	u32array[0] = bc.Value()
	u32array[1] = bc.Value1()
	u32array[2] = bc.Value2()
	u32array[3] = bc.Value3()
	u32array[4] = bc.Value4()
	u32array[5] = bc.Value5()
	u32array[6] = bc.Value6()
	u32array[7] = bc.Value7()
	vObj := &uint256.Uint256{}
	err := vObj.FromUint32Array(u32array)
	if err != nil {
		return err
	}
	b.Value = vObj
	b.TXOutIdx = bc.TXOutIdx()

	owner := &ValueStoreOwner{}
	if err := owner.UnmarshalBinary(bc.Owner()); err != nil {
		return err
	}
	b.Owner = owner
	fObj := &uint256.Uint256{}
	u32array[0] = bc.Fee0()
	u32array[1] = bc.Fee1()
	u32array[2] = bc.Fee2()
	u32array[3] = bc.Fee3()
	u32array[4] = bc.Fee4()
	u32array[5] = bc.Fee5()
	u32array[6] = bc.Fee6()
	u32array[7] = bc.Fee7()
	err = fObj.FromUint32Array(u32array)
	if err != nil {
		return err
	}
	b.Fee = fObj
	return nil
}

// MarshalCapn marshals the object into its capnproto definition
func (b *VSPreImage) MarshalCapn(seg *capnp.Segment) (mdefs.VSPreImage, error) {
	if b == nil {
		return mdefs.VSPreImage{}, errorz.ErrInvalid{}.New("vspi.marshalCapn: vspi not initialized")
	}
	var bc mdefs.VSPreImage
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bc, err
		}
		tmp, err := mdefs.NewRootVSPreImage(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	} else {
		tmp, err := mdefs.NewVSPreImage(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	}
	owner, err := b.Owner.MarshalBinary()
	if err != nil {
		return bc, err
	}
	if err := bc.SetOwner(owner); err != nil {
		return bc, err
	}
	bc.SetChainID(b.ChainID)
	u32array, err := b.Value.ToUint32Array()
	if err != nil {
		return bc, err
	}
	bc.SetValue(u32array[0])
	bc.SetValue1(u32array[1])
	bc.SetValue2(u32array[2])
	bc.SetValue3(u32array[3])
	bc.SetValue4(u32array[4])
	bc.SetValue5(u32array[5])
	bc.SetValue6(u32array[6])
	bc.SetValue7(u32array[7])
	u32array, err = b.Fee.ToUint32Array()
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
func (b *VSPreImage) PreHash() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("vspi.preHash: vspi not initialized")
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

// ValidateSignature validates the signature for VSPreImage
func (b *VSPreImage) ValidateSignature(msg []byte, sig *ValueStoreSignature) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("vspi.validateSignature: vspi not initialized")
	}
	return b.Owner.ValidateSignature(msg, sig)
}
