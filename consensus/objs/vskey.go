package objs

import (
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
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
		return errorz.ErrInvalid{}.New("ValidatorSetKey.UnmarshalBinary; vsk not initialized")
	}
	if len(data) != 6 {
		return errorz.ErrInvalid{}.New("ValidatorSetKey.UnmarshalBinary; incorrect state length")
	}
	b.Prefix = utils.CopySlice(data[0:2])
	nb, _ := utils.UnmarshalUint32(data[2:6])
	if nb == 0 {
		return errorz.ErrInvalid{}.New("ValidatorSetKey.UnmarshalBinary; NotBefore is zero")
	}
	b.NotBefore = nb
	return nil
}

// MarshalBinary takes the ValidatorSetKey object and returns the canonical
// byte slice
func (b *ValidatorSetKey) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("ValidatorSetKey.MarshalBinary; vsk not initialized")
	}
	if len(b.Prefix) != 2 {
		return nil, errorz.ErrInvalid{}.New("ValidatorSetKey.MarshalBinary; incorrect prefix length")
	}
	if b.NotBefore == 0 {
		return nil, errorz.ErrInvalid{}.New("ValidatorSetKey.MarshalBinary; NotBefore is zero")
	}
	key := []byte{}
	Prefix := utils.CopySlice(b.Prefix)
	nb := utils.MarshalUint32(b.NotBefore)
	key = append(key, Prefix...)
	key = append(key, nb...)
	return key, nil
}
