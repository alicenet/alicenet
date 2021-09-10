package objs

import (
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
)

// PendingHdrLeafKey ...
type PendingHdrLeafKey struct {
	Prefix []byte
	Key    []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// PendingHdrLeafKey object
func (b *PendingHdrLeafKey) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	if len(data) != 34 {
		return errorz.ErrInvalid{}.New("Invalid data for PendingHdrLeafKey unmarshalling")
	}
	b.Prefix = utils.CopySlice(data[0:2])
	b.Key = utils.CopySlice(data[2:])
	return nil
}

// MarshalBinary takes the PendingHdrLeafKey object and returns the canonical
// byte slice
func (b *PendingHdrLeafKey) MarshalBinary() ([]byte, error) {
	if b == nil || len(b.Prefix) != 2 || len(b.Key) != 32 {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	key := []byte{}
	key = append(key, utils.CopySlice(b.Prefix)...)
	key = append(key, utils.CopySlice(b.Key)...)
	return key, nil
}
