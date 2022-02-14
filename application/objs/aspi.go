package objs

import (
	"github.com/MadBase/MadNet/application/objs/aspreimage"
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// ASPreImage holds the values required for an AtomicSwap object
type ASPreImage struct {
	ChainID  uint32
	Value    *uint256.Uint256
	TXOutIdx uint32
	IssuedAt uint32
	Exp      uint32
	Owner    *AtomicSwapOwner
	Fee      *uint256.Uint256
	//
	preHash []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// ASPreImage object
func (b *ASPreImage) UnmarshalBinary(data []byte) error {
	bc, err := aspreimage.Unmarshal(data)
	if err != nil {
		return err
	}
	return b.UnmarshalCapn(bc)
}

// MarshalBinary takes the ASPreImage object and returns the canonical
// byte slice
func (b *ASPreImage) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("aspi.marshalBinary; aspi not initialized")
	}
	bc, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	return aspreimage.Marshal(bc)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *ASPreImage) UnmarshalCapn(bc mdefs.ASPreImage) error {
	if err := aspreimage.Validate(bc); err != nil {
		return err
	}
	owner := &AtomicSwapOwner{}
	if err := owner.UnmarshalBinary(bc.Owner()); err != nil {
		return err
	}
	b.Owner = owner
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
	b.IssuedAt = bc.IssuedAt()
	b.Exp = bc.Exp()
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
func (b *ASPreImage) MarshalCapn(seg *capnp.Segment) (mdefs.ASPreImage, error) {
	if b == nil {
		return mdefs.ASPreImage{}, errorz.ErrInvalid{}.New("aspi.marshalCapn; aspi not initialized")
	}
	var bc mdefs.ASPreImage
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bc, err
		}
		tmp, err := mdefs.NewRootASPreImage(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	} else {
		tmp, err := mdefs.NewASPreImage(seg)
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
		return bc, nil
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
	bc.SetIssuedAt(b.IssuedAt)
	bc.SetExp(b.Exp)
	return bc, nil
}

// PreHash returns the PreHash of the object
func (b *ASPreImage) PreHash() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("aspi.preHash; aspi not initialized")
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

// IsExpired returns true if the current epoch is greater than exp
func (b *ASPreImage) IsExpired(currentHeight uint32) (bool, error) {
	if b == nil {
		return true, errorz.ErrInvalid{}.New("aspi.isExpired; aspi not initialized")
	}
	currentEpoch := utils.Epoch(currentHeight)
	if currentEpoch >= b.Exp {
		return true, nil
	}
	return false, nil
}

// ValidateSignature validates the signature for ASPreImage
func (b *ASPreImage) ValidateSignature(currentHeight uint32, msg []byte, sig *AtomicSwapSignature) error {
	isExpired, err := b.IsExpired(currentHeight)
	if err != nil {
		return err
	}
	return b.Owner.ValidateSignature(msg, sig, isExpired)
}

// SignAsPrimary signs the ASPreImage as the primary account owner
func (b *ASPreImage) SignAsPrimary(msg []byte, signer *crypto.Secp256k1Signer, hashKey []byte) (*AtomicSwapSignature, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("aspi.signAsPrimary; aspi not initialized")
	}
	return b.Owner.SignAsPrimary(msg, signer, hashKey)
}

// SignAsAlternate signs the ASPreImage as the alternate account owner
func (b *ASPreImage) SignAsAlternate(msg []byte, signer *crypto.Secp256k1Signer, hashKey []byte) (*AtomicSwapSignature, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("aspi.signAsAlternate; aspi not initialized")
	}
	return b.Owner.SignAsAlternate(msg, signer, hashKey)
}
