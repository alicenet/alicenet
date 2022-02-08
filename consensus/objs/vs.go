package objs

import (
	"bytes"

	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/vset"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// ValidatorSet ...
type ValidatorSet struct {
	Validators []*Validator
	GroupKey   []byte
	NotBefore  uint32
	// Not Part of actual object below this line
	ValidatorVAddrMap      map[string]int
	ValidatorGroupShareMap map[string]int
	ValidatorVAddrSet      map[string]bool
	ValidatorGroupShareSet map[string]bool
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// ValidatorSet object
func (b *ValidatorSet) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("ValidatorSet.UnmarshalBinary; vs not initialized")
	}
	bh, err := vset.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *ValidatorSet) UnmarshalCapn(bh mdefs.ValidatorSet) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("ValidatorSet.UnmarshalCapn; vs not initialized")
	}
	err := vset.Validate(bh)
	if err != nil {
		return err
	}
	b.NotBefore = bh.NotBefore()
	b.GroupKey = utils.CopySlice(bh.GroupKey())
	valList, err := bh.Validators()
	if err != nil {
		return err
	}
	for i := 0; i < valList.Len(); i++ {
		v := &Validator{}
		err := v.UnmarshalCapn(valList.At(i))
		if err != nil {
			return err
		}
		b.Validators = append(b.Validators, v)
	}
	b.ValidatorVAddrMap = make(map[string]int)
	b.ValidatorVAddrSet = make(map[string]bool)
	b.ValidatorGroupShareMap = make(map[string]int)
	b.ValidatorGroupShareSet = make(map[string]bool)
	for idx, v := range b.Validators {
		b.ValidatorVAddrMap[string(v.VAddr)] = idx
		b.ValidatorVAddrSet[string(v.VAddr)] = true
		b.ValidatorGroupShareMap[string(v.GroupShare)] = idx
		b.ValidatorGroupShareSet[string(v.GroupShare)] = true
	}
	return nil
}

// MarshalBinary takes the ValidatorSet object and returns the canonical
// byte slice
func (b *ValidatorSet) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("ValidatorSet.MarshalBinary; vs not initialized")
	}
	if b.NotBefore == 0 {
		return nil, errorz.ErrInvalid{}.New("ValidatorSet.MarshalBinary; NotBefore is zero")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return vset.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *ValidatorSet) MarshalCapn(seg *capnp.Segment) (mdefs.ValidatorSet, error) {
	if b == nil {
		return mdefs.ValidatorSet{}, errorz.ErrInvalid{}.New("ValidatorSet.MarshalCapn; vs not initialized")
	}
	var bh mdefs.ValidatorSet
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootValidatorSet(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewValidatorSet(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	}
	valList, err := bh.NewValidators(int32(len(b.Validators)))
	if err != nil {
		return bh, err
	}
	for i, val := range b.Validators {
		capv, err := val.MarshalCapn(seg)
		if err != nil {
			return bh, err
		}
		err = valList.Set(i, capv)
		if err != nil {
			return mdefs.ValidatorSet{}, err
		}
	}
	if err := bh.SetGroupKey(b.GroupKey); err != nil {
		return bh, err
	}
	bh.SetNotBefore(b.NotBefore)
	return bh, nil
}

func (b *ValidatorSet) IsGroupShareValidator(groupShare []byte) bool {
	return b.ValidatorGroupShareSet[string(groupShare)]
}

func (b *ValidatorSet) GroupShareIdx(groupShare []byte) (bool, int) {
	ok := b.IsGroupShareValidator(groupShare)
	idx := b.ValidatorGroupShareMap[string(groupShare)]
	return ok, idx
}

func (b *ValidatorSet) IsVAddrValidator(addr []byte) bool {
	addrCopy := utils.CopySlice(addr)
	return b.ValidatorVAddrSet[string(addrCopy)]
}

func (b *ValidatorSet) VAddrIdx(vAddr []byte) (bool, int) {
	ok := b.IsVAddrValidator(vAddr)
	idx := b.ValidatorVAddrMap[string(vAddr)]
	return ok, idx
}

func (b *ValidatorSet) IsValidTriplet(vAddr []byte, groupShare []byte, groupKey []byte) bool {
	if !bytes.Equal(groupKey, b.GroupKey) {
		return false
	}
	vOK, vIdx := b.VAddrIdx(vAddr)
	if !vOK {
		return false
	}
	gOK, gIdx := b.GroupShareIdx(groupShare)
	if !gOK {
		return false
	}
	if gIdx != vIdx {
		return false
	}
	return true
}

func (b *ValidatorSet) IsValidTuple(vAddr []byte, groupKey []byte) bool {
	if !bytes.Equal(groupKey, b.GroupKey) {
		return false
	}
	return b.ValidatorVAddrSet[string(vAddr)]
}
