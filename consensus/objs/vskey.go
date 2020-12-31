package objs

import (
	"github.com/MadBase/MadNet/errorz"
	gUtils "github.com/MadBase/MadNet/utils"
)

// ValidatorSetKey ...
type ValidatorSetKey struct {
	Prefix    []byte
	NotBefore uint32
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// ValidatorSetKey object
func (b *ValidatorSetKey) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	if len(data) != 6 {
		return errorz.ErrInvalid{}.New("data incorrect length for unmarshalling ValidatorSetKey")
	}
	b.Prefix = gUtils.CopySlice(data[0:2])
	nb, _ := gUtils.UnmarshalUint32(data[2:6])
	if nb == 0 {
		return errorz.ErrInvalid{}.New("invalid NotBefore for unmarshalling ValidatorSetKey; NotBefore == 0")
	}
	b.NotBefore = nb
	return nil
}

// MarshalBinary takes the ValidatorSetKey object and returns the canonical
// byte slice
func (b *ValidatorSetKey) MarshalBinary() ([]byte, error) {
	if b == nil || len(b.Prefix) != 2 || b.NotBefore == 0 {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	key := []byte{}
	Prefix := gUtils.CopySlice(b.Prefix)
	nb := gUtils.MarshalUint32(b.NotBefore)
	key = append(key, Prefix...)
	key = append(key, nb...)
	return key, nil
}
